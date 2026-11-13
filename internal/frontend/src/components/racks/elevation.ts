import type { Host } from "@/client";

// RACK_UNITS is the height a rack is drawn at unless a node sits higher
export const RACK_UNITS = 42;

// rackUnit reads the U a node sits in from its name, the third part of names like cpn-f01-24 or cpn-b05-08-01, the same part the BMC location template defaults to
export function rackUnit(name?: string): number | undefined {
  const part = name?.split("-")[2];
  if (!part || !/^\d+$/.test(part)) return undefined;
  const u = parseInt(part, 10);
  return u > 0 ? u : undefined;
}

export type Placement = {
  // units maps a U to the nodes in it, more than one for sleds sharing a chassis, sorted by name
  units: Map<number, Host[]>;
  // unplaced are the rack's nodes whose name gives no U
  unplaced: Host[];
  height: number;
};

export function placeNodes(hosts: Host[]): Placement {
  const units = new Map<number, Host[]>();
  const unplaced: Host[] = [];
  for (const h of hosts) {
    const u = rackUnit(h.name);
    if (u === undefined) unplaced.push(h);
    else units.set(u, [...(units.get(u) ?? []), h]);
  }
  for (const list of units.values())
    list.sort((a, b) => (a.name ?? "").localeCompare(b.name ?? ""));
  unplaced.sort((a, b) => (a.name ?? "").localeCompare(b.name ?? ""));
  const height = Math.max(RACK_UNITS, ...units.keys());
  return { units, unplaced, height };
}

export type RackRow =
  | { kind: "unit"; u: number; hosts: Host[] }
  | { kind: "gap"; top: number; bottom: number };

// rackRows lists the rack from the top U down. A run of at least minGap empty units becomes one gap row so a sparse rack still fits on a screen
export function rackRows(p: Placement, minGap = 3): RackRow[] {
  const rows: RackRow[] = [];
  let run: number[] = [];
  const flush = () => {
    if (run.length >= minGap) rows.push({ kind: "gap", top: run[0], bottom: run[run.length - 1] });
    else for (const u of run) rows.push({ kind: "unit", u, hosts: [] });
    run = [];
  };
  for (let u = p.height; u >= 1; u--) {
    const hosts = p.units.get(u);
    if (hosts?.length) {
      flush();
      rows.push({ kind: "unit", u, hosts });
    } else run.push(u);
  }
  flush();
  return rows;
}

// rackOrder is every node in the order the rack shows them, top to bottom, then the unplaced ones, so shift-click can select a range
export function rackOrder(p: Placement): string[] {
  return [...p.units.entries()]
    .sort(([a], [b]) => b - a)
    .flatMap(([, hosts]) => hosts)
    .concat(p.unplaced)
    .map((h) => h.name ?? "")
    .filter(Boolean);
}

// RackTags sorts the tags of a rack's nodes, other than the rack tag itself, by how many nodes have them
export type RackTags = {
  // every are the tags every node has, so they can be shown once instead of on every faceplate. A rack of one node has none, since nothing would be left to tell it apart
  every: string[];
  // some are the tags on more than one node but not all, with how many have each, most common first. A tag on one node says nothing about the rack
  some: [string, number][];
};

const byName = new Intl.Collator(undefined, { numeric: true }).compare;

export function rackTags(hosts: Host[], rack: string): RackTags {
  const counts = new Map<string, number>();
  for (const h of hosts) for (const t of new Set(h.tags)) counts.set(t, (counts.get(t) ?? 0) + 1);
  counts.delete(`rack=${rack}`);
  const every: string[] = [];
  const some: [string, number][] = [];
  if (hosts.length < 2) return { every, some };
  for (const [t, n] of counts) {
    if (n === hosts.length) every.push(t);
    else if (n > 1) some.push([t, n]);
  }
  return { every: every.sort(), some: some.sort(([a, x], [b, y]) => y - x || byName(a, b)) };
}

// hallRow is the row of the data hall a rack stands in, its name less the trailing number, so f01 and f12 share row f. A name with no trailing number, or nothing before it, has no row and gives ""
export function hallRow(rack: string): string {
  return rack.match(/^(.*?)[-_]?\d+$/)?.[1] ?? "";
}
