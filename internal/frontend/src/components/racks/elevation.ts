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

// sharedTags are the tags every node in the rack has, other than the rack tag itself, so they can be shown once instead of on every faceplate. A rack of one node has none, since nothing would be left to tell it apart
export function sharedTags(hosts: Host[], rack: string): string[] {
  if (hosts.length < 2) return [];
  const [first, ...rest] = hosts;
  return [...new Set(first.tags ?? [])]
    .filter((t) => t !== `rack=${rack}` && rest.every((h) => h.tags?.includes(t)))
    .sort();
}
