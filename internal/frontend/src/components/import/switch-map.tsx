import { Faceplate } from "@/components/switches/front-panel";
import { SwitchProblem } from "@/components/switches/switch-problem";
import { needsAttention, type Row, type ScanSwitch, type Status } from "@/lib/import-rows";
import { frontPanel } from "@/lib/switch-ports";
import { cn } from "@/lib/utils";
import { TriangleAlert } from "lucide-react";
import { statusLook } from "./status";

const shown: Status[] = [
  "ready",
  "existing",
  "pick",
  "conflict",
  "invalid",
  "badmac",
  "unnamed",
  "empty",
];

const legend: Record<Status, string> = {
  ready: "ready",
  existing: "already in Grendel",
  pick: "need a MAC picked",
  conflict: "conflicting",
  invalid: "to fix",
  empty: "with no MAC",
  unnamed: "waiting for a name",
  badmac: "with an unreadable MAC",
  excluded: "left out",
};

export function SwitchSummary({ rows }: { rows: Row[] }) {
  const counts = new Map<Status, number>();
  for (const r of rows) if (r.include) counts.set(r.status, (counts.get(r.status) ?? 0) + 1);
  return (
    <ul className="flex flex-wrap gap-x-3 gap-y-1 text-xs font-normal text-muted-foreground">
      {shown
        .filter((s) => counts.get(s))
        .map((s) => (
          <li key={s} className="flex items-center gap-1.5">
            <span
              aria-hidden
              className={cn(
                "size-2 rounded-[1px]",
                !statusLook[s].color && "bg-zinc-400/50 dark:bg-zinc-700",
              )}
              style={statusLook[s].color ? { backgroundColor: statusLook[s].color } : undefined}
            />
            <span className="text-foreground tabular-nums">{counts.get(s)}</span>
            {legend[s]}
          </li>
        ))}
    </ul>
  );
}

export function SwitchDetail({
  sw,
  rows,
  focused,
  onSelect,
}: {
  sw: ScanSwitch;
  rows: Row[];
  focused?: string;
  onSelect: (key: string) => void;
}) {
  if (sw.error)
    return (
      <SwitchProblem
        what="the MAC table"
        code={sw.error_code}
        detail={sw.error}
        address={sw.address}
      />
    );

  const blocks = frontPanel(
    sw.nos,
    rows.map((row) => ({ port_name: row.ifname, row })),
  );

  return (
    <div className="grid justify-items-center gap-2 px-4 py-4">
      {blocks ? (
        <Faceplate
          blocks={blocks}
          light={(slot) => {
            const lanes = slot.neighbors.map((n) => n.row);
            const row = lanes.find((r) => r.include) ?? lanes[0];
            if (!row)
              return { label: `Port ${slot.number}, not reported`, plugged: false, inert: true };
            const color = statusLook[row.status].color;
            const attention = needsAttention(row);
            return {
              label: `Port ${slot.number}${row.name ? `, ${row.name}` : ""}: ${statusLook[row.status].label}${row.problem ? `. ${row.problem}` : ""}`,
              plugged: row.macs.length > 0 || !!row.mac,
              color,
              dim: !row.include,
              ring: attention ? color : undefined,
            };
          }}
          selected={(slot) => slot.neighbors.some((n) => n.row.key === focused)}
          onSelect={(slot) => {
            const row = slot.neighbors[0]?.row;
            if (row) onSelect(row.key);
          }}
        />
      ) : (
        <p className="text-xs text-muted-foreground">
          The faceplate can't be drawn for these port names. Every port is still in the list below.
        </p>
      )}
      {(sw.warnings?.length ?? 0) > 0 && (
        <ul className="grid gap-1 text-xs text-amber-700 dark:text-amber-300">
          {sw.warnings!.map((w) => (
            <li key={w} className="flex items-start gap-1.5">
              <TriangleAlert className="mt-px size-3.5 shrink-0" />
              <span>
                Couldn't read {w}. Ports that have learned no MAC are missing from the list, so{" "}
                {"{n}"} doesn't count them.
              </span>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
