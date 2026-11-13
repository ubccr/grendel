import { BmcResponseModelBmcJobList, getV1BmcJobs } from "@/client";
import ActionsSheet from "@/components/actions-sheet";
import JobsDeleteAction from "@/components/actions/jobs/delete";
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
import { Card, CardContent } from "@/components/ui/card";
import AuthRedirect from "@/lib/auth";
import { createFileRoute } from "@tanstack/react-router";
import { useMemo } from "react";

type ArrayElement<T> = T extends (infer U)[] ? U : never;
type Result = ArrayElement<NonNullable<BmcResponseModelBmcJobList["results"]>>;
type Job = ArrayElement<NonNullable<NonNullable<Result["data"]>["jobs"]>>;

type nodeJob = NodeTreeRow<Job>;

// columns are defined once, outside the component, so the table keeps the same column definitions across renders
const columns: DataTableColumnDef<nodeJob, any>[] = [
  ...nodeColumns<Job>(),
  {
    accessorKey: "id",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Job ID" />,
  },
  {
    accessorKey: "status",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Job Status" />,
  },
  {
    accessorKey: "state",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Job State" />,
  },
  {
    accessorKey: "percent_complete",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Percent Complete" />,
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

export const Route = createFileRoute("/nodes/jobs/$nodes")({
  component: RouteComponent,
  beforeLoad: AuthRedirect,
  loader: async ({ params: { nodes } }) => getV1BmcJobs({ query: { nodeset: nodes } }),
});

function RouteComponent() {
  const { data } = Route.useLoaderData();
  const { isFetching } = Route.useMatch();

  const nodeJobList = useMemo(
    () => nodeTree<Job>((data?.results ?? []).map((r) => ({ ...r, items: r.data?.jobs }))),
    [data],
  );

  const actions: DataTableActions<nodeJob> = ({ table }) => {
    const checked = new Map<string, string[]>();
    // the selection holds node rows too, when a whole node is checked, so take only the job rows under them
    const selected = table.getSelectedRowModel().flatRows.filter((v) => v.depth > 0);
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
          getRowId={nodeRowId}
          columns={columns}
          Actions={actions}
          selectable
          tree={nodeTreeTable}
          getRowTone={nodeRowTone}
          data={nodeJobList}
          progress={isFetching !== false}
        />
      </CardContent>
    </Card>
  );
}
