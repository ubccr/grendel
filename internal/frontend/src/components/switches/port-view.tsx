import { DataTable } from "@/components/data-table/data-table";
import type { DataTableColumnDef } from "@/components/data-table/features";
import { DataTableColumnHeader } from "@/components/data-table/header";
import { Faceplate, kindColor, Led } from "@/components/switches/front-panel";
import { frontPanel, type NeighborKind } from "@/lib/switch-ports";
import { cn } from "@/lib/utils";
import { useNavigate, useSearch } from "@tanstack/react-router";
import { useMemo, useState, type ReactNode } from "react";

type PortRow = { port_name?: string };

const kinds: NeighborKind[] = ["node", "other"];

// cageKind is how a cage is lit: one row Grendel doesn't know turns it amber, so a port holding new hardware is never hidden by a known neighbor beside it on a breakout
function cageKind<T>(rows: T[], kindOf: (row: T) => NeighborKind): NeighborKind | undefined {
  if (rows.length === 0) return undefined;
  return rows.some((r) => kindOf(r) === "other") ? "other" : "node";
}

// SwitchPortView is a switch's ports as its faceplate, a legend that hides rows by kind, and a table with a row for each thing seen on a port. Picking a cage filters the table to its port through the table's own Port filter, which the route keeps in its port_name search parameter
export function SwitchPortView<T extends PortRow>({
  nos,
  rows,
  kindOf,
  legend,
  cageLabel,
  emptyLegend,
  columns,
  getRowId,
  initialVisibility,
  notice,
}: {
  nos?: string;
  rows: T[];
  kindOf: (row: T) => NeighborKind;
  // legend names a count of rows of a kind, like "3 MACs not in Grendel"
  legend: (kind: NeighborKind, n: number) => string;
  // cageLabel names what a cage holds, for its tooltip and screen readers
  cageLabel: (number: number, rows: T[]) => string;
  // emptyLegend names a count of cages with nothing in them
  emptyLegend: (n: number) => string;
  // columns follow the Port column
  columns: DataTableColumnDef<T, any>[];
  getRowId: (row: T, i: number) => string;
  initialVisibility?: Record<string, boolean>;
  // notice sits between the legend and the table, for warnings about what couldn't be read
  notice?: ReactNode;
}) {
  const { port_name: port } = useSearch({ strict: false }) as { port_name?: string };
  const navigate = useNavigate();
  const [hidden, setHidden] = useState<Set<NeighborKind>>(new Set());
  const blocks = useMemo(() => frontPanel(nos, rows), [nos, rows]);

  const select = (name?: string) =>
    navigate({
      to: ".",
      search: (prev) => ({ ...prev, port_name: name && name !== port ? name : undefined }),
      replace: true,
      resetScroll: false,
    });
  const toggle = (kind: NeighborKind) =>
    setHidden((prev) => {
      const next = new Set(prev);
      if (!next.delete(kind)) next.add(kind);
      return next;
    });

  const counts = { node: 0, other: 0 } satisfies Record<NeighborKind, number>;
  for (const r of rows) counts[kindOf(r)]++;
  const empty = blocks?.reduce((n, b) => n + b.slots.filter((s) => !s.neighbors.length).length, 0);
  const portNames = new Set(rows.map((r) => r.port_name));
  const shown = rows.filter((r) => !hidden.has(kindOf(r)));

  const portColumn: DataTableColumnDef<T, any> = {
    id: "port_name",
    accessorFn: (r) => r.port_name ?? "",
    sortFn: "alphanumeric",
    // a whole port name matches only that port, so Ethernet1 doesn't also list Ethernet10 to 19; anything else typed matches part of a name
    filterFn: (row, id, value: string) => {
      const name = row.getValue<string>(id);
      return portNames.has(value)
        ? name === value
        : name.toLowerCase().includes(value.toLowerCase());
    },
    header: ({ column }) => <DataTableColumnHeader column={column} title="Port" />,
    cell: ({ row }) => (
      <button
        type="button"
        onClick={() => select(row.original.port_name)}
        className={cn(
          "rounded-sm font-mono text-sm underline-offset-4 outline-none hover:underline focus-visible:ring-2 focus-visible:ring-ring",
          row.original.port_name === port && "text-sky-600 dark:text-sky-400",
        )}
      >
        {row.original.port_name}
      </button>
    ),
  };

  return (
    <div className="grid gap-4">
      {blocks && (
        <Faceplate
          blocks={blocks}
          light={(slot) => {
            const kind = cageKind(slot.neighbors, kindOf);
            return {
              label: cageLabel(slot.number, slot.neighbors),
              plugged: !!kind,
              inert: !kind,
              color: kind && kindColor[kind],
              dim: !!kind && hidden.has(kind),
            };
          }}
          selected={(slot) => slot.neighbors.some((r) => r.port_name === port)}
          onSelect={(slot) => select(slot.neighbors[0]?.port_name)}
        />
      )}

      <div className="flex flex-wrap items-center gap-x-1 gap-y-1 text-sm">
        {kinds.map((kind) => (
          <button
            key={kind}
            type="button"
            aria-pressed={!hidden.has(kind)}
            disabled={counts[kind] === 0}
            onClick={() => toggle(kind)}
            title={hidden.has(kind) ? "Show these" : "Hide these"}
            className={cn(
              "inline-flex items-center gap-2 rounded-md px-2 py-1 outline-none hover:bg-foreground/5 focus-visible:ring-2 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50",
              hidden.has(kind) &&
                "text-muted-foreground line-through decoration-muted-foreground/60",
            )}
          >
            <Led kind={kind} dim={hidden.has(kind)} className="size-2 rounded-full" />
            <span className="tabular-nums">{counts[kind]}</span> {legend(kind, counts[kind])}
          </button>
        ))}
        {empty !== undefined && empty > 0 && (
          <span className="inline-flex items-center gap-2 px-2 py-1 text-muted-foreground">
            <Led className="size-2 rounded-full" />
            <span className="tabular-nums">{empty}</span> {emptyLegend(empty)}
          </span>
        )}
      </div>

      {notice}

      <DataTable
        getRowId={getRowId}
        columns={[portColumn, ...columns]}
        data={shown}
        initialSorting={[{ id: "port_name", desc: false }]}
        initialVisibility={initialVisibility}
      />
    </div>
  );
}
