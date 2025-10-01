import { useGetV1SwitchNodesetLldp } from "@/openapi/queries";
import { LLDP } from "@/openapi/requests";
import { DataTableColumnHeader } from "../data-table/header";
import { ColumnDef } from "@tanstack/react-table";
import { DataTable } from "../data-table/data-table";

export default function LldpTable({ node }: { node: string }) {
  const query_lldp = useGetV1SwitchNodesetLldp(
    { path: { nodeset: node } },
    undefined,
    { staleTime: 5 * 60 * 1000 },
  );

  const columns: ColumnDef<LLDP>[] = [
    {
      accessorKey: "port_name",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Port Name" />
      ),
    },
    {
      accessorKey: "chassis_id",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="System MAC Address" />
      ),
    },
    {
      accessorKey: "system_name",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="System Name" />
      ),
    },
    {
      accessorKey: "system_description",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="System Description" />
      ),
    },
    {
      accessorKey: "port_id",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="port_id" />
      ),
    },
  ];

  return (
    <div className="my-4">
      <DataTable
        columns={columns}
        data={query_lldp.data ?? []}
        progress={query_lldp.isFetching}
      />
    </div>
  );
}
