import { GetRolesResponse, getV1Roles } from "@/client";
import ActionsSheet from "@/components/actions-sheet";
import RolesDeleteAction from "@/components/actions/roles/delete";
import { DataTable, DataTableActions } from "@/components/data-table/data-table";
import { DataTableColumnHeader } from "@/components/data-table/header";
import SelectableCheckbox from "@/components/data-table/selectableCheckbox";
import { Card, CardContent } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import AuthRedirect from "@/lib/auth";
import { createFileRoute, Link } from "@tanstack/react-router";
import { ColumnDef } from "@tanstack/react-table";
import { useState } from "react";

export const Route = createFileRoute("/account/roles/")({
  component: RouteComponent,
  beforeLoad: AuthRedirect,
  loader: () => getV1Roles(),
});

function RouteComponent() {
  const roles = Route.useLoaderData();
  const { isFetching } = Route.useMatch();

  const [lastSelectedID, setLastSelectedID] = useState(0);

  const columns: ColumnDef<NonNullable<GetRolesResponse["roles"]>[number]>[] = [
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

  const actions: DataTableActions<NonNullable<GetRolesResponse["roles"]>[number]> = ({ table }) => {
    const checked = table
      .getSelectedRowModel()
      .rows.map((v) => v.getAllCells()[1].getValue())
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
