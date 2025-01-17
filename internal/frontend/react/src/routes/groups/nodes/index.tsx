import { createFileRoute, Link } from "@tanstack/react-router";
import { useImageList } from "@/openapi/queries";
import { BootImage } from "@/openapi/requests";
import { ColumnDef } from "@tanstack/react-table";
import { DataTable } from "@/components/data-table/data-table";
import { DataTableColumnHeader } from "@/components/data-table/header";
import { Checkbox } from "@/components/ui/checkbox";

export const Route = createFileRoute("/groups/nodes/")({
    component: RouteComponent,
});

const columns: ColumnDef<BootImage>[] = [
    {
        id: "select",
        header: ({ table }) => (
            <Checkbox
                checked={table.getIsAllPageRowsSelected() || (table.getIsSomePageRowsSelected() && "indeterminate")}
                onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
                aria-label="Select all"
            />
        ),
        cell: ({ row }) => (
            <Checkbox
                checked={row.getIsSelected()}
                onCheckedChange={(value) => row.toggleSelected(!!value)}
                aria-label="Select row"
            />
        ),
    },
    {
        accessorKey: "name",
        header: ({ column }) => <DataTableColumnHeader column={column} title="Name" />,
        cell: ({ row }) => {
            const name = row.original.name;
            return (
                <Link to={`/images/${name}`} className="hover:underline">
                    {name}
                </Link>
            );
        },
    },
    {
        accessorKey: "arch",
        header: ({ column }) => <DataTableColumnHeader column={column} title="Arch" />,
    },
    {
        accessorKey: "kernels",
        header: ({ column }) => <DataTableColumnHeader column={column} title="Kernels" />,
    },
];

function RouteComponent() {
    const { data, isSuccess } = useImageList();

    return <div className="px-6">{isSuccess && data != undefined && <DataTable columns={columns} data={data} />}</div>;
}
