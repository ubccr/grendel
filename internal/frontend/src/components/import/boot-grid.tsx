import { DataTable, type DataTableTree, type RowTone } from "@/components/data-table/data-table";
import { DataTableColumnHeader } from "@/components/data-table/header";
import type { DataTableColumnDef, DataTableInstance } from "@/components/data-table/features";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  bootNeedsAttention,
  choose,
  ruleLabel,
  unreachedGroup,
  type BootOverride,
  type BootOverrides,
  type BootRow,
  type Group,
  type Nic,
  type Rule,
  type Seen,
} from "@/lib/boot-nics";
import { fold } from "@/lib/nodeset";
import { patchOverride, withIncluded } from "@/lib/overrides";
import { cn } from "@/lib/utils";
import { LoaderCircle, RefreshCw, RotateCcw } from "lucide-react";
import { useMemo } from "react";
import { statusColumn, type StatusFilter } from "./grid";
import { BootStatusTag } from "./status";

const statusFilters: StatusFilter<BootRow>[] = [
  { value: "attention", label: "Needs attention", test: (r) => r.include && bootNeedsAttention(r) },
  { value: "ready", label: "Ready", test: (r) => r.include && r.status === "ready" },
  {
    value: "existing",
    label: "Already in Grendel",
    test: (r) => r.include && r.status === "existing",
  },
  { value: "empty", label: "No interfaces listed", test: (r) => r.status === "empty" },
  { value: "excluded", label: "Left out", test: (r) => r.status === "excluded" },
  { value: "edited", label: "Picked by hand", test: (r) => r.overridden },
];

// GridRow is a port layout with its nodes under it, or one of those nodes
type GridRow = { key: string; group?: Group; items?: GridRow[]; row?: BootRow };

type GridMeta = {
  rows: BootRow[];
  rules: Record<string, Rule>;
  byNode: Map<string, Nic[]>;
  seen: Map<string, Seen>;
  retrying: boolean;
  setRule: (group: string, rule: Rule) => void;
  override: (node: string, patch: Partial<BootOverride>) => void;
  retry: (nodes: string[]) => void;
};

const metaOf = (table: DataTableInstance<GridRow>) => table.options.meta as unknown as GridMeta;

// a node whose BMC didn't answer or listed no ports has nothing to pick from, so it can't be ticked
const eligible = (r: BootRow) => r.status !== "unreached" && r.nics.length > 0;

function tone(r: BootRow): RowTone | undefined {
  if (r.status === "conflict" || r.status === "invalid" || r.status === "unreached") return "error";
  if (r.status === "pick") return "warn";
  if (r.status === "ready" || r.status === "existing") return "ok";
  return undefined;
}

const tree: DataTableTree<GridRow> = {
  getSubRows: (r) => r.items,
  groupColumns: ["layout", "rule"],
};

const s = (n: number) => (n === 1 ? "" : "s");
const speed = (n: Nic) => (n.speed_mbps ? `${n.speed_mbps / 1000} Gb/s` : undefined);
const where = (seen?: Seen) => seen && `${seen.sw} ${seen.ifname}`;

const columns: DataTableColumnDef<GridRow, any>[] = [
  {
    id: "layout",
    enableColumnFilter: false,
    enableSorting: false,
    enableHiding: false,
    header: "Interface layout",
    cell: ({ row }) => {
      const g = row.original.group!;
      const nodes = g.nodes.length;
      if (g.key === unreachedGroup)
        return (
          <span className="text-sm whitespace-nowrap">
            <span className="tabular-nums">{nodes}</span> BMC{s(nodes)} didn't answer
          </span>
        );
      return (
        <span className="grid min-w-0 leading-tight" title={g.ids.join("\n")}>
          <span className="text-sm whitespace-nowrap">
            <span className="tabular-nums">{nodes}</span> node{s(nodes)} with{" "}
            {g.ids.length === 0
              ? "no interfaces listed"
              : `${g.ids.length} interface${s(g.ids.length)}`}
          </span>
          <span className="max-w-[28rem] truncate font-mono text-xs font-normal text-muted-foreground">
            {fold(g.nodes)}
          </span>
        </span>
      );
    },
  },
  {
    id: "rule",
    enableColumnFilter: false,
    enableSorting: false,
    enableHiding: false,
    header: () => <span className="sr-only">Rule</span>,
    cell: ({ row, table }) => {
      const m = metaOf(table);
      const g = row.original.group!;
      if (g.key === unreachedGroup)
        return (
          <div className="flex justify-end">
            <Button
              variant="outline"
              size="sm"
              disabled={m.retrying}
              onClick={() => m.retry(g.nodes)}
            >
              {m.retrying ? <LoaderCircle className="animate-spin" /> : <RefreshCw />}
              Ask again
            </Button>
          </div>
        );
      if (g.ids.length === 0) return null;
      return <RuleSelect group={g} meta={m} />;
    },
  },
  {
    id: "node",
    accessorFn: (g) => g.row?.node ?? "",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Node" />,
    enableSorting: false,
    enableHiding: false,
    cell: ({ row }) => (
      <span className="font-mono text-xs whitespace-nowrap">{row.original.row!.node}</span>
    ),
  },
  {
    id: "port",
    accessorFn: (g) => (g.row?.nic ? `${g.row.nic.id} ${g.row.nic.mac}` : ""),
    header: ({ column }) => <DataTableColumnHeader column={column} title="Interface" />,
    enableSorting: false,
    enableHiding: false,
    cell: ({ row, table }) => {
      const m = metaOf(table);
      const r = row.original.row!;
      if (r.status === "unreached" || r.status === "empty")
        return (
          <p className="max-w-[26rem] text-xs whitespace-normal text-muted-foreground">
            {r.problem}
          </p>
        );
      return (
        <div className="grid gap-1">
          <div className="flex items-center gap-1">
            <NicSelect
              row={r}
              seen={m.seen}
              onChange={(mac) => {
                // picking the port the group's rule chooses goes back to following the rule
                const auto = choose(r.nics, m.rules[r.group], m.seen);
                m.override(r.node, { mac: mac === auto?.mac ? undefined : mac });
              }}
            />
            {r.overridden && (
              <Button
                size="icon-sm"
                variant="ghost"
                title="Use the group's rule"
                onClick={() => m.override(r.node, { mac: undefined })}
              >
                <RotateCcw />
                <span className="sr-only">Use the group's rule for {r.node}</span>
              </Button>
            )}
          </div>
          {r.include && r.note && <p className="text-xs text-muted-foreground">{r.note}</p>}
        </div>
      );
    },
  },
  {
    id: "seen",
    accessorFn: (g) => where(g.row?.seen) ?? "",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Switch port" />,
    enableSorting: false,
    enableHiding: false,
    cell: ({ row }) => {
      const r = row.original.row!;
      return r.seen ? (
        <span className="font-mono text-xs whitespace-nowrap">{where(r.seen)}</span>
      ) : null;
    },
  },
  {
    id: "fqdn",
    accessorFn: (g) => g.row?.fqdn ?? "",
    header: ({ column }) => <DataTableColumnHeader column={column} title="FQDN" />,
    enableSorting: false,
    enableHiding: false,
    cell: ({ row }) => {
      const r = row.original.row!;
      return (
        <span
          className="block max-w-40 truncate font-mono text-xs text-muted-foreground"
          title={r.fqdn}
        >
          {r.include ? r.fqdn : ""}
        </span>
      );
    },
  },
  statusColumn<GridRow, BootRow>(statusFilters, (r) => {
    const note = r.status !== "unreached" && r.status !== "empty" ? r.problem : undefined;
    return (
      <>
        <BootStatusTag row={r} />
        {note && <p className="mt-0.5 max-w-64 text-xs text-muted-foreground">{note}</p>}
      </>
    );
  }),
];

const switchRule = "__switch";
const linkRule = "__link";

function RuleSelect({ group, meta }: { group: Group; meta: GridMeta }) {
  const rule = meta.rules[group.key];
  const settled = (r: Rule) =>
    group.nodes.filter((n) => choose(meta.byNode.get(n) ?? [], r, meta.seen)).length;
  const linked = (id: string) =>
    group.nodes.filter((n) => meta.byNode.get(n)?.some((x) => x.id === id && x.link === "LinkUp"))
      .length;
  const value = rule.kind === "switch" ? switchRule : rule.kind === "link" ? linkRule : rule.id;
  const count = settled(rule);
  const total = group.nodes.length;
  const choices = [
    ...(meta.seen.size > 0 ? [{ value: switchRule, rule: { kind: "switch" } as Rule }] : []),
    { value: linkRule, rule: { kind: "link" } as Rule },
  ];

  return (
    <div className="flex items-center justify-end gap-3 font-normal">
      <span className="text-xs text-muted-foreground">Boot from</span>
      <Select
        items={[
          ...choices.map((c) => ({ value: c.value, label: ruleLabel(c.rule) })),
          ...group.ids.map((id) => ({ value: id, label: id })),
        ]}
        value={value}
        onValueChange={(v) =>
          meta.setRule(
            group.key,
            v === switchRule
              ? { kind: "switch" }
              : v === linkRule
                ? { kind: "link" }
                : { kind: "id", id: v as string },
          )
        }
      >
        <SelectTrigger
          title={`The rule picks an interface on ${count} of the ${total} node${s(total)}`}
          className="h-8 w-80 bg-background font-normal"
          aria-label={`Rule for the ${total} node${s(total)} with these interfaces`}
        >
          <SelectValue />
          <span
            className={cn(
              "ml-auto text-xs tabular-nums",
              count < total ? "text-amber-700 dark:text-amber-300" : "text-muted-foreground",
            )}
          >
            {count} of {total}
          </span>
        </SelectTrigger>
        <SelectContent
          align="end"
          alignItemWithTrigger={false}
          className="w-auto min-w-(--anchor-width)"
        >
          {choices.map((c) => (
            <SelectItem key={c.value} value={c.value}>
              {ruleLabel(c.rule)}
              <span className="ml-auto pl-3 text-muted-foreground tabular-nums">
                {settled(c.rule)} of {total}
              </span>
            </SelectItem>
          ))}
          {group.ids.map((id) => (
            <SelectItem key={id} value={id} className="font-mono text-xs">
              {id}
              <span className="ml-auto pl-3 font-sans text-muted-foreground tabular-nums">
                {linked(id)} linked
              </span>
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
}

// Each option shows its link, speed and the switch port it was seen on, which are what tell the cabled port apart
function NicSelect({
  row,
  seen,
  onChange,
}: {
  row: BootRow;
  seen: Map<string, Seen>;
  onChange: (mac: string) => void;
}) {
  const items = row.nics.map((n) => ({ value: n.mac ?? "", label: `${n.id} ${n.mac}` }));
  return (
    <Select items={items} value={row.nic?.mac ?? null} onValueChange={(v) => onChange(v as string)}>
      <SelectTrigger
        size="sm"
        className={cn(
          "w-80 max-w-full font-mono text-xs",
          !row.nic && row.include && "border-amber-500/60 text-amber-700 dark:text-amber-300",
          !row.include && "text-muted-foreground",
        )}
        aria-label={`Interface for ${row.node}`}
      >
        <SelectValue placeholder="Pick an interface" />
      </SelectTrigger>
      <SelectContent
        align="start"
        alignItemWithTrigger={false}
        className="w-auto min-w-(--anchor-width)"
      >
        {row.nics.map((n) => {
          const at = n.mac ? seen.get(n.mac) : undefined;
          return (
            <SelectItem key={n.mac} value={n.mac ?? ""} className="font-mono text-xs">
              <span
                aria-hidden
                className="size-1.5 shrink-0 rounded-full"
                style={{
                  backgroundColor:
                    n.link === "LinkUp" ? "var(--node-ready)" : "var(--muted-foreground)",
                }}
              />
              {n.id}
              <span className="text-muted-foreground">{n.mac}</span>
              <span className="ml-auto flex gap-2 pl-3 font-sans text-muted-foreground">
                {row.known.includes(n.mac ?? "") && <span>in Grendel</span>}
                {at && <span className="text-sky-700 dark:text-sky-300">on {where(at)}</span>}
                {speed(n)}
              </span>
            </SelectItem>
          );
        })}
      </SelectContent>
    </Select>
  );
}

export function BootGrid({
  groups,
  rows,
  rules,
  byNode,
  seen,
  overrides,
  retrying,
  refreshing,
  onRule,
  onOverrides,
  onRetry,
  onRefresh,
}: {
  groups: Group[];
  rows: BootRow[];
  rules: Record<string, Rule>;
  byNode: Map<string, Nic[]>;
  seen: Map<string, Seen>;
  overrides: BootOverrides;
  retrying: boolean;
  refreshing: boolean;
  onRule: (group: string, rule: Rule) => void;
  onOverrides: (o: BootOverrides) => void;
  onRetry: (nodes: string[]) => void;
  onRefresh: () => void;
}) {
  const data = useMemo<GridRow[]>(() => {
    const byGroup = new Map<string, BootRow[]>();
    for (const r of rows) byGroup.set(r.group, [...(byGroup.get(r.group) ?? []), r]);
    const unreached = byGroup.get(unreachedGroup) ?? [];
    const all: Group[] = [
      ...(unreached.length
        ? [{ key: unreachedGroup, ids: [], nodes: unreached.map((r) => r.node) }]
        : []),
      ...groups,
    ];
    return all.map((g) => ({
      key: `group:${g.key}`,
      group: g,
      items: (byGroup.get(g.key) ?? []).map((row) => ({ key: row.node, row })),
    }));
  }, [groups, rows]);

  const override = (node: string, patch: Partial<BootOverride>) =>
    onOverrides(patchOverride(overrides, node, patch));

  const selection = {
    ids: useMemo(() => rows.filter((r) => r.include).map((r) => r.node), [rows]),
    onChange: (ids: string[]) =>
      onOverrides(
        withIncluded(
          overrides,
          rows.filter(eligible).map((r) => ({ key: r.node, byDefault: true })),
          new Set(ids),
        ),
      ),
  };

  const meta: GridMeta = {
    rows,
    rules,
    byNode,
    seen,
    retrying,
    setRule: onRule,
    override,
    retry: onRetry,
  };

  return (
    <DataTable
      columns={columns}
      data={data}
      tree={tree}
      selection={selection}
      canSelect={(g) => !!g.row && eligible(g.row)}
      getRowId={(r) => r.key}
      meta={meta}
      paginate={false}
      syncFilters={false}
      scrollHeight="min(75vh, 56rem)"
      refresh={onRefresh}
      progress={refreshing}
      getRowTone={(row) => (row.original.row ? tone(row.original.row) : undefined)}
      getRowProps={(row) => {
        const r = row.original.row;
        return r ? { className: cn(!r.include && "text-muted-foreground") } : {};
      }}
    />
  );
}
