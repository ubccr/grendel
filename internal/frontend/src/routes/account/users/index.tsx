import UserActions from "@/components/account/user-actions";
import ActionsSheet from "@/components/actions-sheet";
import {
  DataTable,
  DataTableActions,
} from "@/components/data-table/data-table";
import { DataTableColumnHeader } from "@/components/data-table/header";
import SelectableCheckbox from "@/components/data-table/selectableCheckbox";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { useGetV1Users } from "@/openapi/queries";
import { User } from "@/openapi/requests";
import { createFileRoute } from "@tanstack/react-router";
import { ColumnDef } from "@tanstack/react-table";
import { useState } from "react";

export const Route = createFileRoute("/account/users/")({
  component: RouteComponent,
});

function RouteComponent() {
  const users = useGetV1Users();
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
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Username" />
      ),
      cell: ({ row }) => {
        return <span>{row.original.username}</span>;
      },
    },
    {
      accessorKey: "role",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Role" />
      ),
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
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Enabled" />
      ),
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
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Modified At" />
      ),
      cell: ({ row }) => {
        const date = new Date(row.original.modified_at ?? "");
        return <span>{date.toLocaleString()}</span>;
      },
    },
    {
      accessorKey: "created_at",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Created At" />
      ),
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
      <ActionsSheet
        checked={checked}
        length={table.getSelectedRowModel().rows.length}
      >
        <UserActions users={checked} />
      </ActionsSheet>
    );
  };
  return (
    <Card>
      <CardContent>
        <DataTable
          columns={columns}
          data={users.data ?? []}
          Actions={actions}
          progress={users.isFetching}
        />
      </CardContent>
    </Card>
  );
}
