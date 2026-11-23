import { DataTable, type DataTableTree, type RowTone } from "@/components/data-table/data-table";
import { DataTableColumnHeader, FilterBox } from "@/components/data-table/header";
import type { DataTableColumnDef, DataTableInstance } from "@/components/data-table/features";
import { cellInput } from "@/components/form-layout";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import {
  fillNames,
  needsAttention,
  ouiChoices,
  pickByOui,
  type Override,
  type Overrides,
  type Row,
  type ScanSwitch,
  type Status,
} from "@/lib/import-rows";
import { patchOverride, withIncluded } from "@/lib/overrides";
import { cn } from "@/lib/utils";
import { LoaderCircle, RefreshCw, RotateCcw } from "lucide-react";
import { useMemo, useState, type ClipboardEvent } from "react";
import { statusColumn, type StatusFilter } from "./grid";
import { StatusTag } from "./status";
import { SwitchDetail, SwitchSummary } from "./switch-map";

const statusFilters: StatusFilter<Row>[] = [
  { value: "attention", label: "Needs attention", test: needsAttention },
  { value: "ready", label: "Ready", test: (r) => r.include && r.status === "ready" },
  {
    value: "existing",
    label: "Already in Grendel",
    test: (r) => r.include && r.status === "existing",
  },
  { value: "unnamed", label: "Needs a name", test: (r) => r.include && r.status === "unnamed" },
  { value: "empty", label: "No MAC", test: (r) => r.include && r.status === "empty" },
  { value: "excluded", label: "Left out", test: (r) => !r.include },
  { value: "edited", label: "Edited", test: (r) => r.nameOverridden || r.macOverridden },
];

// GridRow is a switch with its ports under it, or one of those ports
type GridRow = { key: string; sw?: ScanSwitch; all?: Row[]; items?: GridRow[]; row?: Row };

// GridMeta is what the cells read from the table's meta: the state that changes as the operator types, kept out of the column definitions so the inputs in them aren't remounted on every keystroke
type GridMeta = {
  // rows is every port the Port filter lets through
  rows: Row[];
  ports: string;
  setPorts: (ports: string) => void;
  overrides: Overrides;
  focused?: string;
  rescanning: ReadonlySet<string>;
  override: (key: string, patch: Partial<Override>) => void;
  setOverrides: (o: Overrides) => void;
  focus: (key: string) => void;
  rescan: (sw: string) => void;
};

const metaOf = (table: DataTableInstance<GridRow>) => table.options.meta as unknown as GridMeta;

// shownRows are the ports the table is showing, in its order, leaving out the ones a filter or a closed switch hides
const shownRows = (table: DataTableInstance<GridRow>) =>
  table.getRowModel().rows.flatMap((r) => r.original.row ?? []);

const attention: Status[] = ["conflict", "invalid", "badmac"];

function tone(r: Row): RowTone | undefined {
  if (!r.include) return undefined;
  if (attention.includes(r.status)) return "error";
  if (r.status === "pick" || r.status === "unnamed") return "warn";
  if (r.status === "ready" || r.status === "existing") return "ok";
  return undefined;
}

const tree: DataTableTree<GridRow> = {
  getSubRows: (r) => r.items,
  groupColumns: ["switch", "summary"],
};

const columns: DataTableColumnDef<GridRow, any>[] = [
  {
    id: "switch",
    enableColumnFilter: false,
    header: "Switch",
    enableSorting: false,
    enableHiding: false,
    cell: ({ row }) => (
      <span className="font-mono text-sm whitespace-nowrap">
        {row.original.sw?.name}
        <span className="ml-2 font-sans text-xs font-normal text-muted-foreground">
          {row.original.sw?.nos}
        </span>
      </span>
    ),
  },
  {
    id: "summary",
    enableColumnFilter: false,
    header: () => <span className="sr-only">Summary</span>,
    enableSorting: false,
    enableHiding: false,
    cell: ({ row, table }) => {
      const m = metaOf(table);
      const name = row.original.sw?.name ?? "";
      const busy = m.rescanning.has(name);
      return (
        <div className="flex items-center gap-3">
          {!row.original.sw?.error && <SwitchSummary rows={row.original.all ?? []} />}
          <Button
            variant="ghost"
            size="sm"
            className="ml-auto font-normal text-muted-foreground"
            disabled={busy}
            onClick={() => m.rescan(name)}
          >
            {busy ? <LoaderCircle className="animate-spin" /> : <RefreshCw />}
            Scan again
          </Button>
        </div>
      );
    },
  },
  {
    id: "port",
    accessorFn: (g) => g.row?.ifname ?? "",
    // the Port filter is a setting rather than a column filter, since it also decides what is imported, so clearing the table's filters or revealing a port leaves it alone
    enableColumnFilter: false,
    header: ({ table }) => {
      const m = metaOf(table);
      return (
        <FilterBox
          title="Port"
          hint="Faceplate numbers like 1-40,45, or text in the port's name. Ports it hides aren't imported."
          value={m.ports}
          onValueChange={m.setPorts}
        />
      );
    },
    enableSorting: false,
    enableHiding: false,
    cell: ({ row, table }) => {
      const r = row.original.row!;
      return (
        <span
          className={cn(
            "-mx-1.5 rounded px-1.5 py-0.5 font-mono text-xs whitespace-nowrap",
            metaOf(table).focused === r.key && "bg-sky-500/20 ring-1 ring-sky-500",
          )}
        >
          {r.ifname}
        </span>
      );
    },
  },
  {
    id: "mac",
    accessorFn: (g) => g.row?.macText ?? "",
    header: ({ column }) => <DataTableColumnHeader column={column} title="MAC" />,
    enableSorting: false,
    enableHiding: false,
    cell: ({ row, table }) => {
      const m = metaOf(table);
      const r = row.original.row!;
      return (
        <Macs
          row={r}
          onSet={(mac) => m.override(r.key, { mac })}
          onFocus={() => m.focus(r.key)}
          onPasteLines={(lines) =>
            m.setOverrides(fillNames(shownRows(table), r.key, lines, m.overrides, "mac"))
          }
        />
      );
    },
  },
  {
    id: "name",
    accessorFn: (g) => g.row?.name ?? "",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Node name" />,
    enableSorting: false,
    enableHiding: false,
    cell: ({ row, table }) => {
      const m = metaOf(table);
      const r = row.original.row!;
      return (
        <div className="relative min-w-48">
          <Input
            aria-label={`Name for ${r.sw} ${r.ifname}`}
            value={r.name}
            disabled={!r.include}
            spellCheck={false}
            aria-invalid={
              r.include && (r.status === "invalid" || r.status === "conflict") && !!r.problem
            }
            className={cn(cellInput, "pr-8", r.nameOverridden && "bg-sky-500/10")}
            onFocus={() => m.focus(r.key)}
            onChange={(e) => m.override(r.key, { name: e.target.value })}
            onPaste={(e) => {
              const lines = pasteLines(e);
              if (lines) m.setOverrides(fillNames(shownRows(table), r.key, lines, m.overrides));
            }}
          />
          {r.nameOverridden && (
            <ResetButton
              title="Use the template's name"
              label={`Use the template's name for ${r.sw} ${r.ifname}`}
              onClick={() => m.override(r.key, { name: undefined })}
            />
          )}
        </div>
      );
    },
  },
  {
    id: "fqdn",
    accessorFn: (g) => g.row?.fqdn ?? "",
    header: ({ column }) => <DataTableColumnHeader column={column} title="FQDN" />,
    enableSorting: false,
    cell: ({ row }) => {
      const r = row.original.row!;
      return (
        <span
          className="block max-w-56 truncate font-mono text-xs text-muted-foreground"
          title={r.fqdn}
        >
          {r.include ? r.fqdn : ""}
        </span>
      );
    },
  },
  statusColumn<GridRow, Row>(statusFilters, (r, table) => {
    return (
      <>
        <StatusTag row={r} />
        {(r.problem || (!r.include && r.skipped)) && (
          <p className="mt-0.5 max-w-64 text-xs text-muted-foreground">
            {r.include ? r.problem : r.skipped}
          </p>
        )}
        {!r.include && r.skipped && (
          <Button
            variant="link"
            size="xs"
            className="h-auto p-0"
            onClick={() => metaOf(table).override(r.key, { include: true })}
          >
            Import it anyway
          </Button>
        )}
      </>
    );
  }),
];

export function PortGrid({
  switches,
  rows,
  ports,
  overrides,
  rescanning,
  scanning,
  onOverrides,
  onPorts,
  onRescan,
  onRescanAll,
}: {
  switches: ScanSwitch[];
  rows: Row[];
  ports: string;
  overrides: Overrides;
  rescanning: ReadonlySet<string>;
  scanning: boolean;
  onOverrides: (o: Overrides) => void;
  onPorts: (ports: string) => void;
  onRescan: (sw: string) => void;
  onRescanAll: () => void;
}) {
  const [focused, setFocused] = useState<string>();
  // reveal is a cage clicked on a faceplate, whose row is brought into view. at tells a second click on the same cage from the first
  const [reveal, setReveal] = useState<{ id: string; at: number }>();
  const ouis = useMemo(() => ouiChoices(rows), [rows]);

  const data = useMemo<GridRow[]>(
    () =>
      switches.map((sw) => {
        const all = rows.filter((r) => r.sw === sw.name);
        return {
          key: `switch:${sw.name}`,
          sw,
          all,
          items: all.filter((row) => !row.hidden).map((row) => ({ key: row.key, row })),
        };
      }),
    [switches, rows],
  );

  const override = (key: string, patch: Partial<Override>) =>
    onOverrides(patchOverride(overrides, key, patch));

  const listed = useMemo(() => rows.filter((r) => !r.hidden), [rows]);
  const selection = {
    ids: useMemo(() => listed.filter((r) => r.include).map((r) => r.key), [listed]),
    onChange: (ids: string[]) =>
      onOverrides(
        withIncluded(
          overrides,
          listed.map((r) => ({ key: r.key, byDefault: !r.skipped })),
          new Set(ids),
        ),
      ),
  };

  const meta: GridMeta = {
    rows: listed,
    ports,
    setPorts: onPorts,
    overrides,
    focused,
    rescanning,
    override,
    setOverrides: onOverrides,
    focus: setFocused,
    rescan: onRescan,
  };

  return (
    <div className="grid gap-3">
      {ouis.length > 1 && (
        <div className="flex flex-wrap items-center gap-1.5 text-xs text-muted-foreground">
          <span>Shared ports: use the MAC starting</span>
          {ouis.map(({ oui, ports }) => (
            <Button
              key={oui}
              size="xs"
              variant="outline"
              className="font-mono"
              title={`On ${ports} ports waiting for a choice, pick the MAC starting ${oui}`}
              onClick={() => onOverrides(pickByOui(rows, oui, overrides))}
            >
              {oui}
              <span className="font-sans text-muted-foreground tabular-nums">{ports}</span>
            </Button>
          ))}
        </div>
      )}

      <DataTable
        columns={columns}
        data={data}
        tree={tree}
        selection={selection}
        // a port with no number in its name has no checkbox until it is asked for from its status, so ticking a whole switch or every port shown never sweeps it in
        canSelect={(g) => !!g.row && (!g.row.skipped || g.row.include)}
        getRowId={(r) => r.key}
        meta={meta}
        paginate={false}
        syncFilters={false}
        scrollHeight="min(75vh, 56rem)"
        refresh={onRescanAll}
        progress={scanning}
        reveal={reveal}
        getRowTone={(row) => (row.original.row ? tone(row.original.row) : undefined)}
        getRowProps={(row) => {
          const r = row.original.row;
          return r
            ? {
                // a moving pointer focuses the row under it; rows sliding under a still pointer, as while a revealed row scrolls into view, send moves with no movement and are ignored
                onMouseMove: (e) =>
                  (e.movementX !== 0 || e.movementY !== 0) &&
                  focused !== r.key &&
                  setFocused(r.key),
                className: cn(!r.include && "text-muted-foreground"),
              }
            : {};
        }}
        renderGroupDetail={(row) =>
          row.original.sw && (
            <SwitchDetail
              sw={row.original.sw}
              rows={row.original.all ?? []}
              focused={focused}
              onSelect={(key) => {
                setFocused(key);
                setReveal({ id: key, at: Date.now() });
              }}
            />
          )
        }
      />
      <p className="text-xs text-muted-foreground">
        Paste a column of names or MACs into a field to fill it and the rows below. Shift-click a
        checkbox to include or leave out a run of ports.
      </p>
    </div>
  );
}

function ResetButton({
  title,
  label,
  onClick,
}: {
  title: string;
  label: string;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      title={title}
      onClick={onClick}
      className="absolute top-1/2 right-1 grid size-6 -translate-y-1/2 place-items-center rounded text-muted-foreground outline-none hover:bg-muted hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring"
    >
      <RotateCcw className="size-3.5" />
      <span className="sr-only">{label}</span>
    </button>
  );
}

// pasteLines returns the lines of a multi-line paste, or nothing for a single value, which pastes into the field as usual
function pasteLines(e: ClipboardEvent<HTMLInputElement>) {
  const lines = e.clipboardData
    .getData("text")
    .split(/\r?\n/)
    .map((l) => l.trim())
    .filter(Boolean);
  if (lines.length < 2) return undefined;
  e.preventDefault();
  return lines;
}

function Macs({
  row,
  onSet,
  onFocus,
  onPasteLines,
}: {
  row: Row;
  onSet: (mac: string | undefined) => void;
  onFocus: () => void;
  onPasteLines: (lines: string[]) => void;
}) {
  // the VLAN only tells the MACs on a port apart when they differ, so it is left off otherwise to keep the choices on one line
  const sameVlan = new Set(row.macs.map((m) => m.vlan)).size === 1;
  const hint = row.macTyped
    ? "typed"
    : row.mac?.owner
      ? `on ${row.mac.owner}${row.mac.bmc ? " BMC" : ""}`
      : row.macs.length === 1 && row.mac?.vlan
        ? `VLAN ${row.mac.vlan}`
        : undefined;

  return (
    <div className="grid gap-1">
      <div className="relative w-52">
        <Input
          aria-label={`MAC for ${row.sw} ${row.ifname}`}
          value={row.macText}
          placeholder={row.macs.length ? "Pick or type one" : "Type the MAC"}
          disabled={!row.include}
          spellCheck={false}
          aria-invalid={row.include && row.status === "badmac"}
          className={cn(cellInput, "pr-8 text-xs", row.macOverridden && "bg-sky-500/10")}
          onFocus={onFocus}
          onChange={(e) => onSet(e.target.value)}
          onPaste={(e) => {
            const lines = pasteLines(e);
            if (lines) onPasteLines(lines);
          }}
        />
        {row.macOverridden && (
          <ResetButton
            title="Use the MAC the switch learned"
            label={`Use the learned MAC for ${row.sw} ${row.ifname}`}
            onClick={() => onSet(undefined)}
          />
        )}
      </div>
      {row.macs.length > 1 ? (
        // a toggle group gives arrow key focus and one pressed item at a time; pressing the picked MAC again clears the pick
        <ToggleGroup
          aria-label={`Learned MACs on ${row.ifname}`}
          variant="outline"
          size="sm"
          spacing={1}
          className="flex-wrap"
          disabled={!row.include}
          value={row.mac?.mac ? [row.mac.mac] : []}
          onValueChange={(v) => onSet(v[0])}
        >
          {row.macs.map((m) => (
            <ToggleGroupItem
              key={m.mac}
              value={m.mac ?? ""}
              className="h-auto px-1.5 py-0.5 text-left"
            >
              <MacText mac={m} vlan={!sameVlan} />
            </ToggleGroupItem>
          ))}
        </ToggleGroup>
      ) : (
        hint && <span className="px-2 text-[11px] leading-none text-muted-foreground">{hint}</span>
      )}
    </div>
  );
}

function MacText({ mac, vlan = true }: { mac: Row["macs"][number]; vlan?: boolean }) {
  return (
    <span className="grid leading-tight">
      <span className="font-mono text-xs">{mac.mac}</span>
      {(mac.owner || (vlan && mac.vlan)) && (
        <span className="text-[11px] text-muted-foreground">
          {mac.owner ? `on ${mac.owner}${mac.bmc ? " BMC" : ""}` : `VLAN ${mac.vlan}`}
        </span>
      )}
    </span>
  );
}
