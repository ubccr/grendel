import type { Lldp } from "@/client";

// a neighbor is a node when Grendel has a node by its name, switches included, and other when it doesn't
export type NeighborKind = "node" | "other";

// Slot is one cage on the front panel. A breakout cage carries several ports, one per lane, so neighbors can hold more than one entry
export type Slot<T = Lldp> = {
  number: number;
  neighbors: T[];
};

// Block is a run of cages printed together on the faceplate. Every cage is drawn the same size, since a port's name doesn't say what kind of cage it is
export type Block<T = Lldp> = {
  slots: Slot<T>[];
};

type Port = { port_name?: string };

// unit is the stack unit or blade slot a port sits on; ports on different ones aren't on one faceplate
type Parsed = { number: number; unit?: string };

// a gap this wide between neighbors starts a new block rather than drawing a run of empty cages
const BLOCK_GAP = 8;

const parsers: Record<string, (name: string) => Parsed | undefined> = {
  arista: (name) => {
    const m = /^Ethernet(\d+)(\/\d+)?$/.exec(name);
    return m ? { number: Number(m[1]) } : undefined;
  },
  dellos10: (name) => {
    const m = /^ethernet(\d+\/\d+)\/(\d+)(:\d+)?$/.exec(name);
    return m ? { number: Number(m[2]), unit: m[1] } : undefined;
  },
  // SONiC numbers ports by their first serdes lane, so a 4 lane switch counts Ethernet0, Ethernet4, Ethernet8
  sonic: (name) => {
    const m = /^Ethernet(\d+)$/.exec(name);
    return m ? { number: Number(m[1]) } : undefined;
  },
};

export function switchNos(tags: string[] | null | undefined) {
  return Object.keys(parsers).find((nos) => tags?.includes(nos));
}

function gcd(a: number, b: number): number {
  return b === 0 ? a : gcd(b, a % b);
}

// parsePorts reads each port name as the cage number printed on the faceplate, or returns undefined when the switch type is unknown, a name doesn't fit it, or the ports span several units or blade slots
function parsePorts<T extends Port>(nos: string | undefined, rows: T[]) {
  const parse = nos ? parsers[nos] : undefined;
  if (!parse || rows.length === 0) return undefined;

  const parsed: { row: T; port: Parsed }[] = [];
  for (const row of rows) {
    const port = parse(row.port_name ?? "");
    if (!port) return undefined;
    parsed.push({ row, port });
  }
  if (new Set(parsed.map((p) => p.port.unit)).size > 1) return undefined;

  // the step is taken from every port on the switch, so SONiC numbering only comes out right when rows covers all of them
  if (nos === "sonic") {
    const step = parsed.reduce((g, { port }) => gcd(g, port.number), 0) || 1;
    for (const p of parsed) p.port.number = p.port.number / step + 1;
  }
  return parsed;
}

// portNumbers maps each port name to its faceplate number. A name the switch type can't read falls back to the last number in it, so Ethernet1/7 is still 7
export function portNumbers(nos: string | undefined, names: string[]): Map<string, number> {
  const out = new Map<string, number>();
  const parsed = parsePorts(
    nos,
    names.map((port_name) => ({ port_name })),
  );
  if (parsed) {
    for (const { row, port } of parsed) out.set(row.port_name, port.number);
    return out;
  }
  for (const name of names) {
    const m = /(\d+)(?!.*\d)/.exec(name);
    if (m) out.set(name, Number(m[1]));
  }
  return out;
}

// frontPanel lays the ports with neighbors out as cages, odd numbers on the top row and even below as switches print them, filling the gaps between neighbors with empty cages. It returns undefined when the ports can't be placed, since a guessed panel would put neighbors on the wrong cages
export function frontPanel<T extends Port = Lldp>(
  nos: string | undefined,
  rows: T[],
): Block<T>[] | undefined {
  const parsed = parsePorts(nos, rows);
  if (!parsed) return undefined;

  const cages = new Map<number, T[]>();
  for (const { row, port } of parsed)
    cages.set(port.number, [...(cages.get(port.number) ?? []), row]);

  const numbers = [...cages.keys()].sort((a, b) => a - b);
  const blocks: Block<T>[] = [];
  let start = numbers[0];
  for (let i = 0; i < numbers.length; i++) {
    const n = numbers[i];
    const next = numbers[i + 1];
    if (next !== undefined && next - n <= BLOCK_GAP) continue;

    // blocks start on an odd number and end on an even one, so each column holds a top and bottom pair
    const tail = blocks[blocks.length - 1]?.slots;
    const previous = tail ? tail[tail.length - 1].number : -1;
    const first = Math.max(start % 2 === 0 ? start - 1 : start, previous + 1, 0);
    const last = n % 2 === 0 ? n : n + 1;
    const slots: Slot<T>[] = [];
    for (let k = first; k <= last; k++) slots.push({ number: k, neighbors: cages.get(k) ?? [] });
    blocks.push({ slots });
    start = next;
  }
  return blocks;
}

// shortName is how a neighbor is matched to a node: hosts often advertise their fully qualified name, while Grendel names nodes by the short one
export function shortName(systemName: string | undefined) {
  return (systemName ?? "").split(".")[0];
}

export function neighborKind(row: Lldp, nodes: Set<string>): NeighborKind {
  return nodes.has(shortName(row.system_name)) ? "node" : "other";
}
