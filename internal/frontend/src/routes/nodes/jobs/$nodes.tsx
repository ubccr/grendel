import { BmcResponseModelBmcJobList, getV1BmcJobs } from "@/client";
import ActionsSheet from "@/components/actions-sheet";
import JobsDeleteAction from "@/components/actions/jobs/delete";
import { DataTable, DataTableActions } from "@/components/data-table/data-table";
import { DataTableColumnHeader } from "@/components/data-table/header";
import { countWith, nodeColumns, NodeRow, rowsPerNode } from "@/components/nodes/bmc-columns";
import SelectableCheckbox from "@/components/data-table/selectableCheckbox";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Switch } from "@/components/ui/switch";
import AuthRedirect from "@/lib/auth";
import { createFileRoute, Link } from "@tanstack/react-router";
import { ColumnDef } from "@tanstack/react-table";
import { useMemo, useState } from "react";

type ArrayElement<T> = T extends (infer U)[] ? U : never;
type Result = ArrayElement<NonNullable<BmcResponseModelBmcJobList["results"]>>;
type Job = ArrayElement<NonNullable<NonNullable<Result["data"]>["jobs"]>>;

type nodeJob = NodeRow & Partial<Job>;

export const Route = createFileRoute("/nodes/jobs/$nodes")({
  component: RouteComponent,
  beforeLoad: AuthRedirect,
  loader: async ({ params: { nodes } }) => getV1BmcJobs({ query: { nodeset: nodes } }),
});

function RouteComponent() {
  const { data } = Route.useLoaderData();
  const { isFetching } = Route.useMatch();

  const [lastSelectedID, setLastSelectedID] = useState(0);
  const nodeJobList = useMemo(
    () => rowsPerNode((data?.results ?? []).map((r) => ({ ...r, items: r.data?.jobs }))),
    [data],
  );

  const columns: ColumnDef<nodeJob>[] = [
    {
      id: "select",
      header: ({ table }) => (
        <Checkbox
          checked={
            table.getIsAllRowsSelected() || (table.getIsSomePageRowsSelected() && "indeterminate")
          }
          onCheckedChange={(value) => table.toggleAllRowsSelected(!!value)}
          aria-label="Select all"
        />
      ),
      cell: ({ row, table }) => (
        <SelectableCheckbox
          row={row}
          table={table}
          lastSelectedID={lastSelectedID}
          setLastSelectedID={setLastSelectedID}
        />
      ),
      aggregatedCell: ({ row }) => (
        <Checkbox
          checked={row.getIsAllSubRowsSelected() || (row.getIsSomeSelected() && "indeterminate")}
          onCheckedChange={(value) => row.toggleSelected(!!value)}
          aria-label="Select all in group"
        />
      ),
    },
    {
      accessorKey: "node",
      header: ({ table }) => (
        <div className="flex justify-between gap-2">
          <span>Node</span>
          <Switch
            aria-description="expand all"
            checked={table.getIsAllRowsExpanded()}
            onCheckedChange={(value) => {
              table.toggleAllRowsExpanded(!!value);
            }}
          />
        </div>
      ),
      cell: ({ row }) => {
        const node = row.original?.node;
        return (
          <Link
            to={"/nodes/$node/node"}
            params={{ node: node ?? "unknown" }}
            className="hover:underline"
          >
            {node}
          </Link>
        );
      },
    },
    ...nodeColumns<nodeJob>(),
    {
      accessorKey: "id",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Job ID" />,
      aggregationFn: countWith<nodeJob>("id"),
      aggregatedCell: ({ getValue }) => <Badge variant="secondary">{getValue<number>()}</Badge>,
    },
    {
      accessorKey: "status",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Job Status" />,
      aggregationFn: "unique",
    },
    {
      accessorKey: "state",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Job State" />,
      aggregationFn: "unique",
    },
    {
      accessorKey: "percent_complete",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Percent Complete" />,
      aggregationFn: "min",
      aggregatedCell: ({ getValue }) => <span>Minimum: {getValue<number>()}</span>,
    },
    {
      accessorKey: "name",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Job Name" />,
    },
    {
      accessorKey: "message",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Job Message" />,
    },
  ];

  const actions: DataTableActions<nodeJob> = ({ table }) => {
    const checked = new Map<string, string[]>();
    const selected = table.getSelectedRowModel().rows.filter((v) => !v.getIsGrouped());
    selected.forEach((v) => {
      const { node, id } = v.original;
      if (!id) return;
      checked.set(node, [id, ...(checked.get(node) ?? [])]);
    });

    const length = selected.length;
    return (
      <ActionsSheet checked={""} length={length}>
        <JobsDeleteAction list={Object.fromEntries(checked)} />
      </ActionsSheet>
    );
  };

  return (
    <Card>
      <CardContent className="p-2">
        <DataTable
          columns={columns}
          Actions={actions}
          initialGrouping={["node"]}
          data={nodeJobList}
          progress={isFetching !== false}
        />
      </CardContent>
    </Card>
  );
}
