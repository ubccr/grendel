import type { BmcResponseModelBmcNicList, Host } from "@/client";
import { numberPorts, type ImportRow, type ScanSwitch } from "@/lib/import-rows";
import { parseTemplate, renderTemplate, type Template } from "@/lib/name-template";

type Result = NonNullable<BmcResponseModelBmcNicList["results"]>[number];
export type Nic = NonNullable<NonNullable<Result["data"]>["nics"]>[number];

export type NodeNics = { node: string; nics: Nic[]; error?: string };

// Seen is where a switch learned a MAC, and the node Grendel already has it on
export type Seen = { sw: string; ifname: string; port?: number; owner?: string };

// a rule picks one port on every node in a group: the one a scanned switch has learned, the first with link, or a Redfish port id, the same on every node of one model
export type Rule = { kind: "switch" } | { kind: "link" } | { kind: "id"; id: string };

export type Group = {
  key: string;
  ids: string[];
  nodes: string[];
};

export type BootStatus =
  | "ready"
  | "existing"
  | "pick"
  | "conflict"
  | "invalid"
  | "empty"
  | "unreached"
  | "excluded";

export type BootOverride = { include?: boolean; mac?: string };
export type BootOverrides = Record<string, BootOverride>;

export type BootRow = {
  node: string;
  group: string;
  nics: Nic[];
  nic?: Nic;
  seen?: Seen;
  overridden: boolean;
  include: boolean;
  // known is the MACs of the ports Grendel already has on the node, besides its BMC. A port picked from them is updated; any other is added after them
  known: string[];
  fqdn: string;
  status: BootStatus;
  problem?: string;
  // note is about a row that is fine, such as where the interface goes among the node's others
  note?: string;
};

export type BootSettings = { fqdn: string; ifname: string; mtu: string; vlan: string };
export const defaultBootSettings: BootSettings = { fqdn: "", ifname: "", mtu: "", vlan: "" };

// unreachedGroup holds the nodes whose BMC didn't answer, apart from the port layouts
export const unreachedGroup = "unreached";

// layout is the set of port ids a node reports, so nodes of the same model and cards fall in one group
export const layout = (nics: Nic[]) => [...new Set(nics.map((n) => n.id ?? ""))].sort().join(" ");

export function groups(nodes: NodeNics[]): Group[] {
  const out = new Map<string, Group>();
  for (const n of nodes) {
    if (n.error !== undefined) continue;
    const key = layout(n.nics);
    const g = out.get(key) ?? { key, ids: key ? key.split(" ") : [], nodes: [] };
    g.nodes.push(n.node);
    out.set(key, g);
  }
  return [...out.values()].sort((a, b) => b.nodes.length - a.nodes.length);
}

// learned maps every MAC the scanned switches hold to where they learned it, with the port numbered as on the faceplate so it can fill {port}. A MAC learned on more than one port, as on a server's own port and on an uplink between two scanned switches, is left out rather than guessed at
export function learned(switches: ScanSwitch[]): Map<string, Seen> {
  const out = new Map<string, Seen>();
  const twice = new Set<string>();
  for (const sw of switches) {
    if (sw.error || !sw.name) continue;
    const numbers = numberPorts(sw);
    for (const p of sw.ports ?? [])
      for (const m of p.macs ?? []) {
        if (!m.mac) continue;
        const here = { sw: sw.name, ifname: p.ifname ?? "" };
        const prior = out.get(m.mac);
        if (prior && (prior.sw !== here.sw || prior.ifname !== here.ifname)) twice.add(m.mac);
        out.set(m.mac, { ...here, port: numbers.get(here.ifname), owner: m.owner || undefined });
      }
  }
  for (const mac of twice) out.delete(mac);
  return out;
}

export function choose(nics: Nic[], rule: Rule, seen: Map<string, Seen>): Nic | undefined {
  if (rule.kind === "id") return nics.find((n) => n.id === rule.id);
  if (rule.kind === "switch") {
    const hits = nics.filter((n) => n.mac && seen.has(n.mac));
    return hits.length === 1 ? hits[0] : undefined;
  }
  return nics.find((n) => n.link === "LinkUp");
}

// defaultRule prefers the port a switch has seen when that settles every node in the group, then a port id every node has link on, then the first port with link. Otherwise it is the first port of the layout, which the operator is expected to change
export function defaultRule(
  group: Group,
  byNode: Map<string, Nic[]>,
  seen: Map<string, Seen>,
): Rule {
  const settles = (rule: Rule) =>
    group.nodes.map((n) => choose(byNode.get(n) ?? [], rule, seen)?.id);
  if (seen.size > 0 && settles({ kind: "switch" }).every(Boolean)) return { kind: "switch" };
  const linked = settles({ kind: "link" });
  if (linked.every(Boolean) && new Set(linked).size === 1) return { kind: "id", id: linked[0]! };
  if (linked.every(Boolean)) return { kind: "link" };
  return group.ids[0] ? { kind: "id", id: group.ids[0] } : { kind: "link" };
}

export function ruleLabel(rule: Rule) {
  return rule.kind === "switch"
    ? "The interface a switch has seen"
    : rule.kind === "link"
      ? "The first interface with link"
      : rule.id;
}

// bootPort is the port Grendel boots a node from: its first one that isn't the BMC's
export const bootPort = (h: Host) => h.interfaces?.find((i) => !i.bmc);

export function bootRows(
  nodes: NodeNics[],
  hosts: Host[],
  rules: Record<string, Rule>,
  overrides: BootOverrides,
  fqdn: Template,
  seen: Map<string, Seen>,
): BootRow[] {
  const known = new Map(
    hosts.map((h) => [
      h.name ?? "",
      (h.interfaces ?? []).filter((i) => !i.bmc && i.mac).map((i) => i.mac!),
    ]),
  );
  return nodes.map((n) => {
    const o = overrides[n.node] ?? {};
    const row: BootRow = {
      node: n.node,
      group: n.error !== undefined ? unreachedGroup : layout(n.nics),
      nics: n.nics,
      overridden: o.mac !== undefined,
      include: n.error === undefined && n.nics.length > 0 && o.include !== false,
      known: known.get(n.node) ?? [],
      fqdn: "",
      status: "excluded",
    };
    if (n.error !== undefined) {
      row.status = "unreached";
      row.problem = n.error || "The BMC didn't answer.";
      return row;
    }
    if (n.nics.length === 0) {
      row.status = "empty";
      row.problem = "The BMC listed no interfaces. Some only do once the host has been powered on.";
      return row;
    }
    const rule = rules[row.group];
    row.nic =
      o.mac !== undefined
        ? n.nics.find((x) => x.mac === o.mac)
        : rule
          ? choose(n.nics, rule, seen)
          : undefined;
    row.seen = row.nic?.mac ? seen.get(row.nic.mac) : undefined;
    if (!row.include) return row;

    if (!row.nic) {
      row.status = "pick";
      row.problem =
        rule?.kind === "switch"
          ? n.nics.some((x) => x.mac && seen.has(x.mac))
            ? "A switch has seen more than one of its interfaces. Pick one."
            : "No scanned switch has seen any of its interfaces. Pick one."
          : "No interface matches the group's rule. Pick one.";
      return row;
    }
    if (fqdn.parts.length || fqdn.error) {
      const r = renderTemplate(fqdn, {
        name: n.node,
        sw: row.seen?.sw,
        port: row.seen?.port,
      });
      if (r.error) {
        row.status = "invalid";
        row.problem =
          row.seen || fqdn.error
            ? r.error
            : [...fqdn.uses].some((u) => u === "port" || u.startsWith("sw"))
              ? "No scanned switch has seen this interface, so the FQDN's switch fields are empty."
              : r.error;
        return row;
      }
      row.fqdn = r.value;
    }
    if (row.seen?.owner && row.seen.owner !== n.node) {
      row.status = "conflict";
      row.problem = `This MAC is already on ${row.seen.owner}.`;
      return row;
    }
    row.status = row.known.includes(row.nic.mac ?? "") ? "existing" : "ready";
    if (row.status === "ready" && row.known.length)
      row.note = `Added after ${row.known[0]}, which stays the boot interface.`;
    return row;
  });
}

export const bootNeedsAttention = (r: BootRow) =>
  r.status === "pick" ||
  r.status === "conflict" ||
  r.status === "invalid" ||
  r.status === "unreached";

export function toBootImport(rows: BootRow[], s: BootSettings): ImportRow[] {
  const mtu = Number(s.mtu) || undefined;
  return rows
    .filter((r) => r.include && r.nic?.mac && (r.status === "ready" || r.status === "existing"))
    .map((r) => ({
      node: r.node,
      mac: r.nic!.mac,
      bmc: false,
      fqdn: r.fqdn || undefined,
      ifname: s.ifname || undefined,
      vlan: s.vlan || undefined,
      mtu,
    }));
}

export const compileBoot = (s: BootSettings) => parseTemplate(s.fqdn);
