import { GetRolesResponse } from "@/client";
import { getV1RolesOptions } from "@/client/@tanstack/react-query.gen";
import ActionsSheet from "@/components/actions-sheet";
import RolesDeleteAction from "@/components/actions/roles/delete";
import { DataTable, DataTableActions } from "@/components/data-table/data-table";
import { DataTableColumnHeader } from "@/components/data-table/header";
import { Card, CardContent } from "@/components/ui/card";
import { useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import type { DataTableColumnDef } from "@/components/data-table/features";

export const Route = createFileRoute("/account/roles/")({
  component: RouteComponent,
  loader: ({ context: { queryClient } }) => queryClient.ensureQueryData(getV1RolesOptions()),
});

const columns: DataTableColumnDef<NonNullable<GetRolesResponse["roles"]>[number], any>[] = [
  {
    accessorKey: "name",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Name" />,
    cell: ({ row }) => {
      return (
        <Link to="/account/roles/$role" params={{ role: row.original.name ?? "" }}>
          {row.original.name}
        </Link>
      );
    },
  },
  {
    id: "permission_length",
    accessorFn: (role) => role.permission_list?.length ?? 0,
    header: ({ column }) => <DataTableColumnHeader column={column} title="Permission Length" />,
  },
  {
    id: "unassigned_permission_length",
    accessorFn: (role) => role.unassigned_permission_list?.length ?? 0,
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title="Unassigned Permission Length" />
    ),
  },
];

function RouteComponent() {
  const { data, isFetching } = useSuspenseQuery(getV1RolesOptions());

  const actions: DataTableActions<NonNullable<GetRolesResponse["roles"]>[number]> = ({ table }) => {
    const checked = table
      .getSelectedRowModel()
      .rows.map((v) => v.original.name)
      .join(",");
    return (
      <ActionsSheet checked={checked} length={table.getSelectedRowModel().rows.length}>
        <RolesDeleteAction roles={checked} />
      </ActionsSheet>
    );
  };
  return (
    <Card>
      <CardContent>
        <DataTable
          selectable
          getRowId={(role, i) => role.name ?? String(i)}
          columns={columns}
          data={data.roles ?? []}
          Actions={actions}
          add="/add/role"
          progress={isFetching}
        />
      </CardContent>
    </Card>
  );
}
