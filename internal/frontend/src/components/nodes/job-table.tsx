import { useGetV1BmcJobs } from "@/openapi/queries";
import ActionsSheet from "../actions-sheet";
import JobActions from "./job-actions";
import { DataTable, DataTableActions } from "../data-table/data-table";
import { DataTableColumnHeader } from "../data-table/header";
import SelectableCheckbox from "../data-table/selectableCheckbox";
import { ColumnDef } from "@tanstack/react-table";
import { RedfishJob } from "@/openapi/requests";
import { useState } from "react";
import { Checkbox } from "../ui/checkbox";

export default function RedfishJobList({ nodes }: { nodes: string }) {
  const { data, isFetching } = useGetV1BmcJobs({
    query: { nodeset: nodes },
  });
  const [lastSelectedID, setLastSelectedID] = useState(0);

  const nodeData = data?.[0];

  type ArrayElement<T> = T extends (infer U)[] ? U : never;
  type Job = ArrayElement<RedfishJob["jobs"]>;
  const columns: ColumnDef<Job>[] = [
    {
      id: "select",
      header: ({ table }) => (
        <Checkbox
          checked={
            table.getIsAllPageRowsSelected() ||
            (table.getIsSomePageRowsSelected() && "indeterminate")
          }
          onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
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
    },
    {
      accessorKey: "Id",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="ID" />
      ),
    },
    {
      accessorKey: "JobState",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Job Status" />
      ),
    },
    {
      accessorKey: "PercentComplete",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Percent Complete" />
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
  const actions: DataTableActions<Job> = ({ table }) => {
    const checked = table
      .getSelectedRowModel()
      .rows.map((v) => v.getAllCells()[1].getValue())
      .join(",");
    const length = table.getSelectedRowModel().rows.length;
    return (
      <ActionsSheet checked={checked} length={length}>
        <JobActions jids={checked} nodes={nodeData?.name ?? ""} />
      </ActionsSheet>
    );
  };

  return (
    <div className="px-6">
      <DataTable
        columns={columns}
        data={nodeData?.jobs ?? []}
        Actions={actions}
        progress={isFetching}
      />
    </div>
  );
}
