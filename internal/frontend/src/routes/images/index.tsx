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
    header: ({ column }) => <DataTableColumnHeader column={column} title="Version" />,
  },
  {
    accessorKey: "kernel",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Kernel" />,
    cell: ({ row }) => <span className="font-mono text-xs break-all">{row.original.kernel}</span>,
  },
  {
    // the paths joined, so a filter finds an image by any of its initrds
    id: "initrd",
    accessorFn: (image) => (image.initrd ?? []).join(" "),
    header: ({ column }) => <DataTableColumnHeader column={column} title="Initrds" />,
    cell: ({ row }) => (
      <div className="grid gap-0.5 font-mono text-xs">
        {row.original.initrd?.map((path, i) => (
          <span key={i} className="break-all">
            {path}
          </span>
        ))}
      </div>
    ),
  },
  {
    accessorKey: "cmdline",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Command line" />,
    // a command line can run to hundreds of characters, so it is cut to two lines with the whole of it on hover
    cell: ({ row }) => (
      <span
        className="line-clamp-2 max-w-md font-mono text-xs break-all"
        title={row.original.cmdline}
      >
        {row.original.cmdline}
      </span>
    ),
  },
  {
    id: "verify",
    accessorFn: (image) => (image.verify ? "on" : "off"),
    header: ({ column }) => <DataTableColumnHeader column={column} title="Verify" />,
    cell: ({ row }) =>
      row.original.verify ? "On" : <span className="text-muted-foreground">Off</span>,
  },
  {
    // each template as name=file, so a filter finds an image by a template's name or its file
    id: "provision_templates",
    accessorFn: (image) =>
      Object.entries(image.provision_templates ?? {})
        .map(([name, file]) => `${name}=${file}`)
        .join(" "),
    header: ({ column }) => <DataTableColumnHeader column={column} title="Templates" />,
    cell: ({ row }) => (
      <dl className="grid grid-cols-[auto_1fr] gap-x-2 gap-y-0.5 text-xs">
        {Object.entries(row.original.provision_templates ?? {}).map(([name, file]) => (
          <div key={name} className="contents">
            <dt className="text-muted-foreground">{name}</dt>
            <dd className="font-mono break-all">{file}</dd>
          </div>
        ))}
      </dl>
    ),
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
          initialVisibility={{ initrd: false, cmdline: false, provision_templates: false }}
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
