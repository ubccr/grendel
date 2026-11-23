import { ArchToggle } from "@/components/boot/arch-toggle";
import { DataTable, type RowTone } from "@/components/data-table/data-table";
import type { DataTableColumnDef, DataTableRow } from "@/components/data-table/features";
import { DataTableColumnHeader, type FilterOption } from "@/components/data-table/header";
import { ErrorPanel } from "@/components/error-panel";
import { NodeLink } from "@/components/nodes/node-link";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { archChoices, arches, groupProblems, problems, type NodeCheck } from "@/lib/boot-check";
import { fold } from "@/lib/nodeset";
import { bootCheckOptions } from "@/lib/queries";
import { cn } from "@/lib/utils";
import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import { FileX2 } from "lucide-react";
import { useMemo } from "react";
import { z } from "zod";

export const Route = createFileRoute("/nodes/bootcheck/$nodes")({
  staticData: { crumb: ["Boot check", { param: "nodes" }] },
  component: RouteComponent,
  validateSearch: z.object({ arch: z.enum(archChoices).optional() }),
});

type Row = NodeCheck & { id: string; problems: string[]; result: "Ready" | "Failed" };

const resultOptions: FilterOption[] = ["Ready", "Failed"].map((v) => ({ value: v, label: v }));

const passed = (items: { error?: string }[] | undefined) =>
  items?.length ? `${items.filter((i) => !i.error).length}/${items.length}` : "";

// the node's own tab can only be opened as one arch, which is the row's when it is one the tab offers
const tabArch = (arch?: string) => arches.find((a) => a === arch);

const columns: DataTableColumnDef<Row, any>[] = [
  {
    accessorKey: "name",
    sortFn: "alphanumeric",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Node" />,
    cell: ({ row: { original: n } }) => (
      <NodeLink
        node={n.name ?? ""}
        to="/nodes/$node/bootcheck"
        search={{ arch: tabArch(n.arch) }}
        className="font-mono text-sm font-medium"
      />
    ),
  },
  {
    accessorKey: "arch",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Arch" />,
    cell: ({ row }) => <span className="font-mono text-xs">{row.original.arch}</span>,
  },
  {
    accessorKey: "image",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Image" />,
    cell: ({ row: { original: n } }) =>
      n.image && (
        <Link
          to="/images/$image"
          params={{ image: n.image }}
          className="underline-offset-4 hover:underline"
        >
          {n.image}
        </Link>
      ),
  },
  {
    accessorKey: "result",
    filterFn: "equals",
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title="Result" options={resultOptions} />
    ),
    cell: ({ row: { original: n } }) => (
      <span className={cn("font-medium", n.ok ? "text-emerald-600" : "text-destructive")}>
        {n.result}
      </span>
    ),
  },
  {
    id: "files",
    accessorFn: (n) => passed(n.files),
    enableColumnFilter: false,
    header: ({ column }) => <DataTableColumnHeader column={column} title="Files" />,
    cell: ({ getValue }) => <span className="tabular-nums">{getValue()}</span>,
  },
  {
    id: "templates",
    accessorFn: (n) => passed(n.templates),
    enableColumnFilter: false,
    header: ({ column }) => <DataTableColumnHeader column={column} title="Templates" />,
    cell: ({ getValue }) => <span className="tabular-nums">{getValue()}</span>,
  },
  {
    id: "problems",
    accessorFn: (n) => n.problems.join("\n"),
    enableSorting: false,
    header: ({ column }) => <DataTableColumnHeader column={column} title="Problems" />,
    cell: ({ row: { original: n } }) =>
      n.problems.length > 0 && (
        <ul className="grid max-w-xl gap-1 text-xs break-words whitespace-normal">
          {n.problems.map((p) => (
            <li key={p}>{p}</li>
          ))}
        </ul>
      ),
  },
];

const rowTone = (row: DataTableRow<Row>): RowTone => (row.original.ok ? "ok" : "error");

const s = (n: number) => (n === 1 ? "" : "s");

function RouteComponent() {
  const { nodes } = Route.useParams();
  const arch = Route.useSearch().arch ?? "x86_64";
  const navigate = Route.useNavigate();
  const { data, error, isPending, isFetching, refetch } = useQuery(bootCheckOptions(nodes, arch));

  const rows = useMemo(
    () =>
      (data?.nodes ?? []).map((n): Row => ({
        ...n,
        id: `${n.name}/${n.arch}`,
        problems: problems(n),
        result: n.ok ? "Ready" : "Failed",
      })),
    [data],
  );
  const groups = useMemo(() => groupProblems(data?.nodes ?? []), [data]);

  const failedNodes = new Set(rows.filter((r) => !r.ok).map((r) => r.name)).size;
  const total = new Set(rows.map((r) => r.name)).size;

  return (
    <div className="grid gap-4">
      <Card>
        <CardHeader className="flex flex-row flex-wrap items-start justify-between gap-4 space-y-0">
          <div className="space-y-1.5">
            <CardTitle>Boot check</CardTitle>
            <CardDescription>
              {isPending
                ? "Reading the boot files and templates from the server's disk…"
                : error
                  ? "The check did not run."
                  : failedNodes === 0
                    ? `${total === 1 ? "The node" : `All ${total} nodes`} can be sent every boot file and template.`
                    : `${failedNodes} of ${total} node${s(total)} would not boot, for ${groups.length} reason${s(groups.length)}.`}
            </CardDescription>
          </div>
          <ArchToggle
            options={archChoices}
            value={arch}
            onChange={(a) => void navigate({ search: { arch: a }, replace: true })}
          />
        </CardHeader>
        {groups.length > 0 && (
          <CardContent>
            <ul className="grid gap-2">
              {groups.map((g) => {
                const folded = fold(g.nodes);
                return (
                  <li
                    key={g.problem}
                    className="grid gap-1 rounded-lg border border-destructive/30 px-4 py-3"
                  >
                    <p className="text-sm break-words">{g.problem}</p>
                    <p className="truncate text-xs text-muted-foreground" title={folded}>
                      <span className="tabular-nums">
                        {g.nodes.length} node{s(g.nodes.length)}
                      </span>
                      <span className="ml-2 font-mono">{folded}</span>
                    </p>
                  </li>
                );
              })}
            </ul>
          </CardContent>
        )}
      </Card>

      <Card>
        <CardContent className="p-2">
          {isPending ? (
            <Skeleton className="h-64 w-full" />
          ) : error ? (
            <ErrorPanel
              title="Could not check the boot"
              message={error.detail || error.title}
              icon={FileX2}
            />
          ) : (
            <DataTable
              columns={columns}
              data={rows}
              getRowId={(r) => r.id}
              getRowTone={rowTone}
              progress={isFetching}
              refresh={() => void refetch()}
              syncFilters={false}
            />
          )}
        </CardContent>
      </Card>
    </div>
  );
}
