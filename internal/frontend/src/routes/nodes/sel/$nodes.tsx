import { BmcResponseModelBmcSelLog, getV1BmcSel } from "@/client";
import ActionsSheet from "@/components/actions-sheet";
import NodesSelAction from "@/components/actions/nodes/sel";
import { DataTable, DataTableActions } from "@/components/data-table/data-table";
import { DataTableColumnHeader } from "@/components/data-table/header";
import {
  nodeColumns,
  nodeRowId,
  nodeRowTone,
  nodeTree,
  nodeTreeTable,
  type NodeTreeRow,
} from "@/components/nodes/bmc-columns";
import type { DataTableColumnDef } from "@/components/data-table/features";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import AuthRedirect from "@/lib/auth";
import { logSeverityColor } from "@/lib/bmc";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useMemo } from "react";
import { z } from "zod";

type ArrayElement<T> = T extends (infer U)[] ? U : never;
type Result = ArrayElement<NonNullable<BmcResponseModelBmcSelLog["results"]>>;
type SelEntry = ArrayElement<NonNullable<NonNullable<Result["data"]>["entries"]>>;
type Entry = NodeTreeRow<SelEntry>;

const limits = ["25", "50", "100", "0"] as const;

// columns are defined once, outside the component, so the table keeps the same column definitions across renders
const columns: DataTableColumnDef<Entry, any>[] = [
  ...nodeColumns<SelEntry>(),
  {
    accessorKey: "created",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Time" />,
    cell: ({ getValue }) => {
      const created = getValue<string>();
      return (
        <span className="whitespace-nowrap">
          {created ? new Date(created).toLocaleString() : ""}
        </span>
      );
    },
  },
  {
    accessorKey: "severity",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Severity" />,
    cell: ({ getValue }) => {
      const severity = getValue<string>();
      if (!severity) return null;
      return (
        <Badge variant="secondary" className={`rounded-sm ${logSeverityColor(severity)}`}>
          {severity}
        </Badge>
      );
    },
  },
  {
    accessorKey: "message",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Event" />,
  },
  {
    accessorKey: "sensor_type",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Sensor" />,
  },
  {
    accessorKey: "message_id",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Message ID" />,
  },
];

export const Route = createFileRoute("/nodes/sel/$nodes")({
  component: RouteComponent,
  beforeLoad: AuthRedirect,
  validateSearch: z.object({
    limit: z.enum(limits).default("50").catch("50"),
  }),
  loaderDeps: ({ search: { limit } }) => ({ limit }),
  loader: async ({ params: { nodes }, deps: { limit } }) =>
    getV1BmcSel({ query: { nodeset: nodes, limit } }),
});

function RouteComponent() {
  const { data } = Route.useLoaderData();
  const { isFetching } = Route.useMatch();
  const { limit } = Route.useSearch();
  const navigate = useNavigate({ from: Route.fullPath });

  const entries = useMemo(
    () => nodeTree<SelEntry>((data?.results ?? []).map((r) => ({ ...r, items: r.data?.entries }))),
    [data],
  );

  const actions: DataTableActions<Entry> = ({ table }) => {
    const nodes = [
      ...new Set(
        table
          .getSelectedRowModel()
          .flatRows.map((v) => v.original.node)
          .filter(Boolean),
      ),
    ];
    return (
      <ActionsSheet type="node" checked={nodes.join(",")} length={nodes.length}>
        <div className="grid gap-4 sm:grid-cols-1">
          <NodesSelAction nodes={nodes.join(",")} />
        </div>
      </ActionsSheet>
    );
  };

  return (
    <div className="grid gap-2">
      <Card>
        <CardHeader className="flex flex-row flex-wrap items-center justify-between gap-4 space-y-0">
          <div className="space-y-1.5">
            <CardTitle>
              <Label htmlFor="sel-limit">Entries per node</Label>
            </CardTitle>
            <CardDescription>
              A limit takes the entries the BMC lists first, which are the newest on most BMCs. All
              reads the whole log and sorts it.
            </CardDescription>
          </div>
          <Select
            value={limit}
            onValueChange={(value) => navigate({ search: { limit: value as typeof limit } })}
          >
            <SelectTrigger id="sel-limit" className="w-[110px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {limits.map((l) => (
                <SelectItem key={l} value={l}>
                  {l === "0" ? "All" : l}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </CardHeader>
      </Card>
      <Card>
        <CardContent className="p-2">
          <DataTable
            getRowId={nodeRowId}
            columns={columns}
            Actions={actions}
            selectable
            tree={nodeTreeTable}
            getRowTone={nodeRowTone}
            initialVisibility={{ message_id: false }}
            data={entries}
            progress={isFetching !== false}
          />
        </CardContent>
      </Card>
    </div>
  );
}
