import type { NextAddressResponse } from "@/client";
import { blockOf, intToIp, lastOctet } from "@/lib/addresses";
import { cn } from "@/lib/utils";

type Cell = "free" | "used" | "here" | "skipped" | "reserved" | "chosen";

const cellClass: Record<Cell, string> = {
  free: "bg-muted-foreground/15",
  used: "bg-foreground/60",
  here: "bg-(--node-blue)",
  skipped: "bg-(--node-amber)",
  reserved: "border border-dashed border-muted-foreground/40",
  chosen: "bg-(--node-ready) ring-2 ring-(--node-ready)/45",
};

// only the addresses checked for the suggestions are pinged, so amber marks those skipped for answering, not every address in the range that would
const legend: { cell: Cell; label: string }[] = [
  { cell: "used", label: "Node" },
  { cell: "here", label: "This node's other ports" },
  { cell: "skipped", label: "Skipped, answered a ping" },
  { cell: "chosen", label: "Chosen" },
];

// RangeMap draws up to a /24 of the range as a grid, a row for each 32 addresses, so how full it is and where the gaps are shows at a glance. It is a picture of what the summary and the suggestions say, so it is hidden from screen readers in favour of them. others are the addresses on the node's other ports
export function RangeMap({
  data,
  others,
  chosen,
}: {
  data: NextAddressResponse;
  others: string[];
  chosen?: string;
}) {
  const { start, size } = blockOf(data.range ?? "", chosen);
  const nodes = new Map((data.used ?? []).map((u) => [u.ip, u.node]));
  const here = new Set(others);
  const reserved = new Set(data.reserved);
  const skipped = new Set(
    [...(data.lowest ?? []), ...(data.after_highest ?? [])]
      .filter((s) => s.answered)
      .map((s) => s.ip),
  );

  // the first state that applies, in this order, colours the cell
  const states: [Cell, (ip: string) => boolean][] = [
    ["chosen", (ip) => ip === chosen],
    ["skipped", (ip) => skipped.has(ip)],
    ["here", (ip) => here.has(ip)],
    ["used", (ip) => nodes.has(ip)],
    ["reserved", (ip) => reserved.has(ip)],
  ];
  const cells = Array.from({ length: size }, (_, i) => {
    const ip = intToIp(start + i);
    const cell = states.find(([, applies]) => applies(ip))?.[0] ?? "free";
    return { ip, cell, node: nodes.get(ip) };
  });
  const rows = Array.from({ length: Math.ceil(size / 32) }, (_, r) =>
    cells.slice(r * 32, r * 32 + 32),
  );
  const shown = new Set(cells.map((c) => c.cell));
  const block = `${intToIp(start)}/${32 - Math.log2(size)}`;

  return (
    <figure className="grid gap-2">
      <figcaption className="flex items-baseline justify-between gap-3 text-xs">
        <span className="font-mono">
          {block}
          {block !== data.range && <span className="text-muted-foreground"> of {data.range}</span>}
        </span>
        <span className="text-muted-foreground tabular-nums">{data.free ?? 0} free</span>
      </figcaption>
      <div aria-hidden className="grid gap-[3px]">
        {rows.map((row, r) => (
          <div key={r} className="grid grid-cols-[1.75rem_1fr] items-center gap-1.5">
            <span className="text-right font-mono text-[10px] leading-none text-muted-foreground tabular-nums">
              {lastOctet(row[0]!.ip)}
            </span>
            <div className="grid grid-cols-[repeat(32,minmax(0,1fr))] gap-[2px]">
              {row.map((c) => (
                <span
                  key={c.ip}
                  title={c.node ? `${c.ip} ${c.node}` : c.ip}
                  className={cn(
                    "aspect-square rounded-[2px] transition-colors motion-reduce:transition-none",
                    cellClass[c.cell],
                  )}
                />
              ))}
            </div>
          </div>
        ))}
      </div>
      <ul aria-hidden className="flex flex-wrap gap-x-3 gap-y-1 text-xs text-muted-foreground">
        {legend
          .filter((l) => shown.has(l.cell))
          .map((l) => (
            <li key={l.cell} className="flex items-center gap-1.5">
              <span className={cn("size-2.5 rounded-[2px]", cellClass[l.cell])} />
              {l.label}
            </li>
          ))}
      </ul>
    </figure>
  );
}
