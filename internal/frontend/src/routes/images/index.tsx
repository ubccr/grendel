import type { BootImage } from "@/client";
import { getV1Images } from "@/client";
import ActionsSheet from "@/components/actions-sheet";
import ImagesDeleteAction from "@/components/actions/images/delete";
import ImagesExportAction from "@/components/actions/images/export";
import { DataTable, type DataTableActions } from "@/components/data-table/data-table";
import { DataTableColumnHeader } from "@/components/data-table/header";
import { Card, CardContent } from "@/components/ui/card";
import AuthRedirect from "@/lib/auth";
import { createFileRoute, Link } from "@tanstack/react-router";
import type { DataTableColumnDef } from "@/components/data-table/features";

export const Route = createFileRoute("/images/")({
  component: TableComponent,
  beforeLoad: AuthRedirect,
  loader: () => getV1Images(),
});

// columns are defined once, outside the component, so the table keeps the same column definitions across renders
const columns: DataTableColumnDef<BootImage, any>[] = [
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
    accessorKey: "version",
    header: ({ column }) => <DataTableColumnHeader column={column} title="version" />,
  },
  {
    accessorKey: "kernel",
    header: ({ column }) => <DataTableColumnHeader column={column} title="kernel" />,
  },
];

function TableComponent() {
  const images = Route.useLoaderData();
  const { isFetching } = Route.useMatch();

  const actions: DataTableActions<BootImage> = ({ table }) => {
    const checked = table
      .getSelectedRowModel()
      .rows.map((v) => v.original.name)
      .join(",");
    return (
      <ActionsSheet checked={checked} length={table.getSelectedRowModel().rows.length}>
        <div className="grid gap-4 sm:grid-cols-2">
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
          selectable
          // id is only optional on the type because an image being added has none yet; every listed image comes from the database and has one
          getRowId={(image) => String(image.id)}
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
