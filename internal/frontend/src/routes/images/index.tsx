import type { BootImage } from "@/client";
import { getV1Images } from "@/client";
import ActionsSheet from "@/components/actions-sheet";
import ImagesDeleteAction from "@/components/actions/images/delete";
import ImagesExportAction from "@/components/actions/images/export";
import { DataTable, type DataTableActions } from "@/components/data-table/data-table";
import { DataTableColumnHeader } from "@/components/data-table/header";
import SelectableCheckbox from "@/components/data-table/selectableCheckbox";
import { Card, CardContent } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import AuthRedirect from "@/lib/auth";
import { createFileRoute, Link } from "@tanstack/react-router";
import type { ColumnDef } from "@tanstack/react-table";
import { useState } from "react";

export const Route = createFileRoute("/images/")({
  component: TableComponent,
  beforeLoad: AuthRedirect,
  loader: () => getV1Images(),
});

function TableComponent() {
  const images = Route.useLoaderData();
  const { isFetching } = Route.useMatch();

  const [lastSelectedID, setLastSelectedID] = useState(0);

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
      header: ({ column }) => <DataTableColumnHeader column={column} title="Name" />,
      cell: ({ row }) => {
        const name = row.original.name;
        return (
          <Link to={"/images/$image"} params={{ image: name }} className="hover:underline">
            {name}
          </Link>
        );
      },
    },
    {
      accessorKey: "kernel",
      header: ({ column }) => <DataTableColumnHeader column={column} title="kernel" />,
    },
  ];

  const actions: DataTableActions<BootImage> = ({ table }) => {
    const checked = table
      .getSelectedRowModel()
      .rows.map((v) => v.getAllCells()[1].getValue())
      .join(",");
    return (
      <ActionsSheet checked={checked} length={table.getSelectedRowModel().rows.length}>
        <div className="mt-4 grid gap-4 sm:grid-cols-2">
          <ImagesDeleteAction images={checked} />
          <ImagesExportAction images={checked} />
        </div>
      </ActionsSheet>
    );
  };

  return (
    <Card>
      <CardContent>
        <DataTable
          columns={columns}
          data={images.data ?? []}
          add={"/add/image"}
          Actions={actions}
          progress={isFetching !== false}
        />
      </CardContent>
    </Card>
  );
}
