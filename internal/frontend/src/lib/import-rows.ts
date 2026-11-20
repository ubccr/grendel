import type { NodeImportRequest, SwitchScanResponse } from "@/client";
import { parseTemplate, renderTemplate, nameProblem, type Template } from "@/lib/name-template";
import { portNumbers } from "@/lib/switch-ports";

export type ScanSwitch = NonNullable<SwitchScanResponse["switches"]>[number];
export type ScanPort = NonNullable<ScanSwitch["ports"]>[number];
export type ScanMac = NonNullable<ScanPort["macs"]>[number];
export type ImportRow = NonNullable<NodeImportRequest["rows"]>[number];

export type Settings = {
  name: string;
  fqdn: string;
  ifname: string;
  // the BMC port's MTU; empty leaves Grendel's default
  mtu: string;
  // the BMC port's VLAN
  portVlan: string;
  // each tag is a template, like rack=a{sw.1}
  tags: string[];
  // the port table's Port filter, which also decides which ports are imported: faceplate numbers like 1-40,45, or text in the port's name; empty is every port
  ports: string;
  start: number;
};

export const defaultSettings: Settings = {
  name: "",
  fqdn: "",
  ifname: "",
  mtu: "",
  portVlan: "",
  tags: [],
  ports: "",
  start: 1,
};

export type Override = {
  include?: boolean;
  mac?: string;
  name?: string;
};

export type Overrides = Record<string, Override>;

export type Status =
  | "ready"
  | "existing"
  | "pick"
  | "empty"
  | "unnamed"
  | "badmac"
  | "conflict"
  | "invalid"
  | "excluded";

export type Row = {
  key: string;
  sw: string;
  ifname: string;
  number?: number;
  macs: ScanMac[];
  mac?: ScanMac;
  name: string;
  nameOverridden: boolean;
  // macText is what the MAC field shows: what was typed, or the MAC in use
  macText: string;
  macOverridden: boolean;
  // mac is typed rather than learned by the switch
  macTyped: boolean;
  fqdn: string;
  tags: string[];
  include: boolean;
  // why a port is left out without anyone unticking it
  skipped?: string;
  // outside the Port filter, so not listed or imported
  hidden: boolean;
  status: Status;
  problem?: string;
  macProblem?: string;
};

export const rowKey = (sw: string, ifname: string) => `${sw}:${ifname}`;

// numberPorts is each scanned port's faceplate number, which {port} fills in
export const numberPorts = (sw: ScanSwitch) =>
  portNumbers(
    sw.nos,
    (sw.ports ?? []).map((p) => p.ifname ?? ""),
  );

// parsePortFilter reads a comma separated list where each part is a faceplate number, a range like 1-40, or text found in the port's name, ignoring case
export function parsePortFilter(text: string): {
  test: (number: number | undefined, ifname: string) => boolean;
} {
  const parts = text
    .split(",")
    .map((p) => p.trim())
    .filter(Boolean);
  if (!parts.length) return { test: () => true };
  const tests = parts.map((part): ((n: number | undefined, ifname: string) => boolean) => {
    const m = /^(\d+)\s*(?:-\s*(\d+))?$/.exec(part);
    if (!m) {
      const needle = part.toLowerCase();
      return (_, ifname) => ifname.toLowerCase().includes(needle);
    }
    const a = Number(m[1]);
    const b = m[2] ? Number(m[2]) : a;
    const [lo, hi] = [Math.min(a, b), Math.max(a, b)];
    return (n) => n !== undefined && n >= lo && n <= hi;
  });
  return { test: (n, ifname) => tests.some((t) => t(n, ifname)) };
}

export type Compiled = {
  name: Template;
  fqdn: Template;
  tags: Template[];
  ports: ReturnType<typeof parsePortFilter>;
};

export function compile(s: Settings): Compiled {
  return {
    name: parseTemplate(s.name),
    fqdn: parseTemplate(s.fqdn),
    tags: s.tags
      .map((t) => t.trim())
      .filter(Boolean)
      .map(parseTemplate),
    ports: parsePortFilter(s.ports),
  };
}

// normalizeMac reads a MAC written with colons, dashes, dots or nothing between the bytes, in either case, as Grendel stores it: lower case with colons. It returns undefined for anything else
export function normalizeMac(text: string): string | undefined {
  const hex = text
    .trim()
    .toLowerCase()
    .replace(/[:\-.\s]/g, "");
  if (!/^[0-9a-f]{12}$/.test(hex)) return undefined;
  return hex.match(/../g)!.join(":");
}

// autoPick chooses a MAC on a port that carries more than one when the choice is clear: the one Grendel already has as this port's BMC
function autoPick(macs: ScanMac[]) {
  const known = macs.filter((m) => m.owner && m.bmc);
  return known.length === 1 ? known[0] : undefined;
}

// buildRows turns the scanned switches into one row per port, in switch then faceplate order. {n} counts the ports being imported, empty ones included, so a BMC that hasn't come up doesn't shift every name after it
export function buildRows(
  switches: ScanSwitch[],
  s: Settings,
  c: Compiled,
  overrides: Overrides,
): Row[] {
  const rows: Row[] = [];
  let n = s.start;

  for (const sw of switches) {
    if (sw.error || !sw.name) continue;
    const ports = sw.ports ?? [];
    const numbers = numberPorts(sw);
    const ordered = [...ports].sort(
      (a, b) => (numbers.get(a.ifname ?? "") ?? 1e9) - (numbers.get(b.ifname ?? "") ?? 1e9),
    );

    for (const p of ordered) {
      const ifname = p.ifname ?? "";
      const key = rowKey(sw.name, ifname);
      const o = overrides[key] ?? {};
      const number = numbers.get(ifname);
      const macs = p.macs ?? [];
      const hidden = !c.ports.test(number, ifname);

      let skipped: string | undefined;
      if (hidden) skipped = "Outside the Port filter";
      else if (number === undefined) skipped = "No port number in its name";
      const include = !hidden && (o.include ?? !skipped);

      const row: Row = {
        key,
        sw: sw.name,
        ifname,
        number,
        macs,
        name: "",
        nameOverridden: o.name !== undefined,
        macText: "",
        macOverridden: o.mac !== undefined,
        macTyped: false,
        fqdn: "",
        tags: [],
        include,
        skipped,
        hidden,
        status: "excluded",
      };

      if (include) {
        const vars = { port: number, n, sw: sw.name };
        n++;

        const rendered = renderTemplate(c.name, vars);
        row.name = o.name ?? rendered.value;
        if (o.name === undefined && rendered.error) row.problem = rendered.error;

        if (o.mac === undefined) {
          row.mac = macs.length === 1 ? macs[0] : autoPick(macs);
        } else {
          // a value matching a learned MAC picks it; anything else is a MAC typed by hand, as for a BMC the switch hasn't seen
          const typed = normalizeMac(o.mac);
          row.mac = macs.find((m) => m.mac === typed) ?? (typed ? { mac: typed } : undefined);
          row.macTyped = !!typed && !macs.some((m) => m.mac === typed);
          if (o.mac.trim() && !typed) row.macProblem = `"${o.mac.trim()}" isn't a MAC address.`;
        }
        row.macText = o.mac ?? row.mac?.mac ?? "";

        if (row.name) {
          const withName = { ...vars, name: row.name };
          row.fqdn = c.fqdn.parts.length ? renderTemplate(c.fqdn, withName).value : "";
          row.tags = c.tags.map((t) => renderTemplate(t, withName).value).filter(Boolean);
        }
        row.status = status(row);
      } else {
        if (o.name !== undefined) row.name = o.name;
        row.macText = o.mac ?? (macs.length === 1 ? (macs[0].mac ?? "") : "");
      }
      rows.push(row);
    }
  }

  markDuplicates(rows);
  return rows;
}

function status(row: Row): Status {
  if (row.macProblem) {
    row.problem = row.macProblem;
    return "badmac";
  }
  if (!row.mac && row.macs.length === 0) return "empty";
  if (!row.mac) {
    row.problem ??=
      row.macs.length === 1
        ? "Type a MAC, or reset to use the one the switch learned."
        : `${row.macs.length} MACs on this port. Pick the BMC's.`;
    return "pick";
  }
  if (row.problem) return "invalid";
  // a port with no name and no template yet is waiting on the operator, not wrong
  if (!row.name && !row.nameOverridden) return "unnamed";
  const bad = nameProblem(row.name);
  if (bad) {
    row.problem = bad;
    return "invalid";
  }
  if (row.mac.owner && row.mac.owner !== row.name) {
    row.problem = `This MAC belongs to ${row.mac.owner}.`;
    return "conflict";
  }
  if (row.mac.owner === row.name && row.mac.bmc) return "existing";
  return "ready";
}

function markDuplicates(rows: Row[]) {
  const byName = new Map<string, Row[]>();
  for (const r of rows) {
    if (!r.include || !r.name || r.status === "empty") continue;
    byName.set(r.name, [...(byName.get(r.name) ?? []), r]);
  }
  for (const [name, same] of byName) {
    if (same.length < 2) continue;
    for (const r of same) {
      if (r.status === "invalid" || r.status === "pick" || r.status === "badmac") continue;
      r.status = "conflict";
      r.problem = `${same.length} ports are named ${name}.`;
    }
  }
}

// importable rows are the ones sent to the server: included, with a MAC and a usable name
export const importable = (r: Row) =>
  r.include && (r.status === "ready" || r.status === "existing");

export const needsAttention = (r: Row) =>
  r.include &&
  (r.status === "pick" ||
    r.status === "conflict" ||
    r.status === "invalid" ||
    r.status === "badmac");

export function tally(rows: Row[]) {
  const t = {
    total: rows.length,
    included: 0,
    ready: 0,
    existing: 0,
    attention: 0,
    empty: 0,
    unnamed: 0,
    excluded: 0,
  };
  for (const r of rows) {
    if (!r.include) {
      t.excluded++;
      continue;
    }
    t.included++;
    if (r.status === "ready") t.ready++;
    else if (r.status === "existing") t.existing++;
    else if (r.status === "empty") t.empty++;
    else if (r.status === "unnamed") t.unnamed++;
    else t.attention++;
  }
  return t;
}

// ouiChoices groups the MACs on ports that need one picked by vendor prefix, so a whole batch of shared-LOM ports can be settled at once: BMCs and host NICs usually come from different prefixes
export function ouiChoices(rows: Row[]) {
  const counts = new Map<string, number>();
  for (const r of rows) {
    if (!r.include || r.status !== "pick") continue;
    const seen = new Set(r.macs.map((m) => (m.mac ?? "").slice(0, 8)));
    for (const oui of seen) counts.set(oui, (counts.get(oui) ?? 0) + 1);
  }
  return [...counts.entries()].sort((a, b) => b[1] - a[1]).map(([oui, ports]) => ({ oui, ports }));
}

// pickByOui returns the overrides that pick, on every port waiting for a choice, the one MAC with that prefix
export function pickByOui(rows: Row[], oui: string, overrides: Overrides): Overrides {
  const next = { ...overrides };
  for (const r of rows) {
    if (!r.include || r.status !== "pick") continue;
    const match = r.macs.filter((m) => (m.mac ?? "").startsWith(oui));
    if (match.length === 1) next[r.key] = { ...next[r.key], mac: match[0].mac };
  }
  return next;
}

// fillNames returns the overrides that put pasted values, names by default or MACs, on the rows from start down, one line each, skipping rows that aren't included
export function fillNames(
  rows: Row[],
  startKey: string,
  names: string[],
  overrides: Overrides,
  field: "name" | "mac" = "name",
): Overrides {
  const next = { ...overrides };
  let i = rows.findIndex((r) => r.key === startKey);
  if (i < 0) return next;
  for (const value of names) {
    while (i < rows.length && !rows[i].include) i++;
    if (i >= rows.length) break;
    next[rows[i].key] = { ...next[rows[i].key], [field]: value };
    i++;
  }
  return next;
}

export function toImportRows(
  rows: Row[],
  s: Pick<Settings, "ifname" | "mtu" | "portVlan">,
): ImportRow[] {
  const mtu = Number(s.mtu) || undefined;
  return rows.filter(importable).map((r) => ({
    node: r.name,
    mac: r.mac!.mac,
    bmc: true,
    fqdn: r.fqdn || undefined,
    ifname: s.ifname || undefined,
    vlan: s.portVlan || undefined,
    mtu,
    tags: r.tags.length ? r.tags : undefined,
  }));
}
