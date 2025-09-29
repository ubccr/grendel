import { createFileRoute, Link } from "@tanstack/react-router";
import { BootImage } from "@/openapi/requests";
import { ColumnDef } from "@tanstack/react-table";
import {
  DataTable,
  DataTableActions,
} from "@/components/data-table/data-table";
import { DataTableColumnHeader } from "@/components/data-table/header";
import { Checkbox } from "@/components/ui/checkbox";
import ActionsSheet from "@/components/actions-sheet";
import ImageActions from "@/components/images/actions";
import { useEffect, useState } from "react";
import AuthRedirect from "@/auth";
import SelectableCheckbox from "@/components/data-table/selectableCheckbox";
import { Card, CardContent } from "@/components/ui/card";
import { useGetV1Images } from "@/openapi/queries";
import { toast } from "sonner";

export const Route = createFileRoute("/images/")({
  component: RouteComponent,
  beforeLoad: AuthRedirect,
});

function RouteComponent() {
  return (
    <div>
      <TableComponent />
    </div>
  );
}

function TableComponent() {
  const { data, error, isFetching } = useGetV1Images();
  const [lastSelectedID, setLastSelectedID] = useState(0);

  useEffect(() => {
    if (error) {
      toast.error(error.title, {
        description: error.detail,
      });
    }
  }, [error]);

  const columns: ColumnDef<BootImage>[] = [
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
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Name" />
      ),
      cell: ({ row }) => {
        const name = row.original.name;
        return (
          <Link
            to={"/images/$image"}
            params={{ image: name }}
            className="hover:underline"
          >
            {name}
          </Link>
        );
      },
    },
    {
      accessorKey: "kernel",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="kernel" />
      ),
    },
  ];

  const actions: DataTableActions<BootImage> = ({ table }) => {
    const checked = table
      .getSelectedRowModel()
      .rows.map((v) => v.getAllCells()[1].getValue())
      .join(",");
    return (
      <ActionsSheet
        checked={checked}
        length={table.getSelectedRowModel().rows.length}
      >
        <ImageActions images={checked} />
      </ActionsSheet>
    );
  };

  return (
    <Card>
      <CardContent>
        <DataTable
          columns={columns}
          data={data ?? []}
          add={"/add/image"}
          Actions={actions}
          progress={isFetching}
        />
      </CardContent>
    </Card>
  );
}
