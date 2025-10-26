import { useGetV1BmcJobs } from "@/openapi/queries";
import ActionsSheet from "../actions-sheet";
import JobActions from "./job-actions";
import { DataTable, DataTableActions } from "../data-table/data-table";
import { DataTableColumnHeader } from "../data-table/header";
import SelectableCheckbox from "../data-table/selectableCheckbox";
import { ColumnDef } from "@tanstack/react-table";
import { RedfishJob } from "@/openapi/requests";
import { useEffect, useState } from "react";
import { Checkbox } from "../ui/checkbox";
import { Switch } from "../ui/switch";
import { Badge } from "../ui/badge";

type ArrayElement<T> = T extends (infer U)[] ? U : never;
type Job = ArrayElement<RedfishJob["jobs"]>;

type nodeJob = Job & {
  Node: string;
};

export default function RedfishJobList({ nodes }: { nodes: string }) {
  const { data, isFetching, refetch } = useGetV1BmcJobs({
    query: { nodeset: nodes },
  });
  const [lastSelectedID, setLastSelectedID] = useState(0);
  const [nodeJobList, setNodeJobList] = useState<Array<nodeJob>>([]);

  useEffect(() => {
    const list: Array<nodeJob> = [];

    data?.forEach((d) => {
      d.jobs?.forEach((v) => {
        list.push({
          Node: d.name ?? "",
          ...v,
        });
      });
    });

    setNodeJobList(list);
  }, [data]);

  const columns: ColumnDef<nodeJob>[] = [
    {
      id: "select",
      header: ({ table }) => (
        <Checkbox
          checked={
            table.getIsAllRowsSelected() ||
            (table.getIsSomePageRowsSelected() && "indeterminate")
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
          checked={
            row.getIsAllSubRowsSelected() ||
            (row.getIsSomeSelected() && "indeterminate")
          }
          onCheckedChange={(value) => row.toggleSelected(!!value)}
          aria-label="Select all in group"
        />
      ),
    },
    {
      accessorKey: "Node",
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
    },
    {
      accessorKey: "Id",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="ID" />
      ),
      aggregationFn: "count",
      aggregatedCell: ({ getValue }) => (
        <Badge variant="secondary">{getValue<number>()}</Badge>
      ),
    },
    {
      accessorKey: "JobStatus",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Job Status" />
      ),
      aggregationFn: "unique",
    },
    {
      accessorKey: "JobState",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Job State" />
      ),
      aggregationFn: "unique",
    },
    {
      accessorKey: "PercentComplete",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Percent Complete" />
      ),
      aggregationFn: "min",
      aggregatedCell: ({ getValue }) => (
        <span>Minimum: {getValue<number>()}</span>
      ),
    },
    {
      accessorKey: "Name",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Name" />
      ),
    },
    {
      accessorKey: "Messages",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Messages" />
      ),
      cell: ({ row }) => {
        return (
          <span>{row.original?.Messages?.map((v) => v.Message).join(",")}</span>
        );
      },
    },
  ];

  const actions: DataTableActions<nodeJob> = ({ table }) => {
    const checked = new Map<string, string[]>();
    table.getSelectedRowModel().rows.forEach((v) => {
      const node = v.getAllCells()[0].getValue<string>();
      const jid = v.getAllCells()[2].getValue<string>();
      checked.set(node, [jid, ...(checked.get(node) ?? [])]);
    });

    const length = table.getSelectedRowModel().rows.length;
    return (
      <ActionsSheet checked={""} length={length}>
        <JobActions checked={checked} />
      </ActionsSheet>
    );
  };

  return (
    <div className="px-6">
      <DataTable
        columns={columns}
        Actions={actions}
        initialGrouping={["Node"]}
        data={nodeJobList}
        progress={isFetching}
        refresh={() => refetch()}
      />
    </div>
  );
}
