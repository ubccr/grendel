import { getV1Users, User } from "@/client";
import ActionsSheet from "@/components/actions-sheet";
import UsersDeleteAction from "@/components/actions/users/delete";
import UsersEnabledAction from "@/components/actions/users/enabled";
import UsersRoleAction from "@/components/actions/users/role";
import { DataTable, DataTableActions } from "@/components/data-table/data-table";
import { DataTableColumnHeader } from "@/components/data-table/header";
import SelectableCheckbox from "@/components/data-table/selectableCheckbox";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import AuthRedirect from "@/lib/auth";
import { createFileRoute } from "@tanstack/react-router";
import { ColumnDef } from "@tanstack/react-table";
import { useState } from "react";

export const Route = createFileRoute("/account/users/")({
  component: RouteComponent,
  beforeLoad: AuthRedirect,
  loader: () => getV1Users(),
});

function RouteComponent() {
  const { isFetching } = Route.useMatch();
  const { data } = Route.useLoaderData();

  const [lastSelectedID, setLastSelectedID] = useState(0);

  const columns: ColumnDef<User>[] = [
    {
      id: "select",
      header: ({ table }) => (
        <Checkbox
          checked={table.getIsAllRowsSelected()}
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
    },
    {
      accessorKey: "username",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Username" />,
      cell: ({ row }) => {
        return <span>{row.original.username}</span>;
      },
    },
    {
      accessorKey: "role",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Role" />,
      cell: ({ row }) => {
        return (
          <Badge variant="secondary" className="rounded-sm">
            {row.original.role}
          </Badge>
        );
      },
    },
    {
      accessorKey: "enabled",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Enabled" />,
      cell: ({ row }) => {
        return (
          <Badge variant="secondary" className="rounded-sm">
            {row.original.enabled?.toString()}
          </Badge>
        );
      },
    },
    {
      accessorKey: "modified_at",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Modified At" />,
      cell: ({ row }) => {
        const date = new Date(row.original.modified_at ?? "");
        return <span>{date.toLocaleString()}</span>;
      },
    },
    {
      accessorKey: "created_at",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Created At" />,
      cell: ({ row }) => {
        const date = new Date(row.original.created_at ?? "");
        return <span>{date.toLocaleString()}</span>;
      },
    },
  ];

  const actions: DataTableActions<User> = ({ table }) => {
    const checked = table
      .getSelectedRowModel()
      .rows.map((v) => v.getAllCells()[1].getValue())
      .join(",");
    return (
      <ActionsSheet checked={checked} length={table.getSelectedRowModel().rows.length}>
        <div className="mt-4 grid gap-4 sm:grid-cols-2">
          <UsersDeleteAction users={checked} />
          <UsersRoleAction users={checked} />
          <UsersEnabledAction users={checked} />
        </div>
      </ActionsSheet>
    );
  };
  return (
    <Card>
      <CardContent>
        <DataTable
          columns={columns}
          data={data ?? []}
          Actions={actions}
          progress={isFetching !== false}
        />
      </CardContent>
    </Card>
  );
}
