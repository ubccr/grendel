import { GetRolesResponse, getV1Roles } from "@/client";
import ActionsSheet from "@/components/actions-sheet";
import RolesDeleteAction from "@/components/actions/roles/delete";
import { DataTable, DataTableActions } from "@/components/data-table/data-table";
import { DataTableColumnHeader } from "@/components/data-table/header";
import { Card, CardContent } from "@/components/ui/card";
import AuthRedirect from "@/lib/auth";
import { createFileRoute, Link } from "@tanstack/react-router";
import type { DataTableColumnDef } from "@/components/data-table/features";

export const Route = createFileRoute("/account/roles/")({
  component: RouteComponent,
  beforeLoad: AuthRedirect,
  loader: () => getV1Roles(),
});

// columns are defined once, outside the component, so the table keeps the same column definitions across renders
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
    accessorKey: "permission_length",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Permission Length" />,
    cell: ({ row }) => {
      return <span>{row.original.permission_list?.length}</span>;
    },
  },
  {
    accessorKey: "unassigned_permission_length",
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title="Unassigned Permission Length" />
    ),
    cell: ({ row }) => {
      return <span>{row.original.unassigned_permission_list?.length}</span>;
    },
  },
];

function RouteComponent() {
  const roles = Route.useLoaderData();
  const { isFetching } = Route.useMatch();

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
          data={roles.data?.roles ?? []}
          Actions={actions}
          add="/add/role"
          progress={isFetching !== false}
        />
      </CardContent>
    </Card>
  );
}
