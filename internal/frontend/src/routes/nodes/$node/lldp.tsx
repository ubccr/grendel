import { getV1SwitchNodesetLldp, Lldp } from "@/client";
import { DataTable } from "@/components/data-table/data-table";
import { DataTableColumnHeader } from "@/components/data-table/header";
import { createFileRoute } from "@tanstack/react-router";
import { ColumnDef } from "@tanstack/react-table";
export const Route = createFileRoute("/nodes/$node/lldp")({
  component: RouteComponent,
  loader: ({ params: { node } }) => getV1SwitchNodesetLldp({ path: { nodeset: node } }),
});

function RouteComponent() {
  const { data } = Route.useLoaderData();

  const columns: ColumnDef<Lldp>[] = [
    {
      accessorKey: "port_name",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Port Name" />,
    },
    {
      accessorKey: "chassis_id",
      header: ({ column }) => <DataTableColumnHeader column={column} title="System MAC Address" />,
    },
    {
      accessorKey: "system_name",
      header: ({ column }) => <DataTableColumnHeader column={column} title="System Name" />,
    },
    {
      accessorKey: "system_description",
      header: ({ column }) => <DataTableColumnHeader column={column} title="System Description" />,
    },
    {
      accessorKey: "port_id",
      header: ({ column }) => <DataTableColumnHeader column={column} title="port_id" />,
    },
  ];
  return (
    <div className="my-4">
      <DataTable columns={columns} data={data ?? []} />
    </div>
  );
}
