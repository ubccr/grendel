import { createFileRoute, Link } from "@tanstack/react-router";
import { useHostList } from "@/openapi/queries";
import { Host } from "@/openapi/requests";
import { ColumnDef } from "@tanstack/react-table";
import { DataTable, DataTableActions } from "@/components/data-table/data-table";
import { DataTableColumnHeader } from "@/components/data-table/header";
import { Checkbox } from "@/components/ui/checkbox";
import ProvisionIcon from "@/components/nodes/provision-button";
import { useState } from "react";
import SelectableCheckbox from "@/components/data-table/selectableCheckbox";
import Actions from "@/components/data-table/actions";
import TagsList from "@/components/tags";

export const Route = createFileRoute("/nodes/")({
    component: RouteComponent,
});

function RouteComponent() {
    const { data, isSuccess } = useHostList();
    const [lastSelectedID, setLastSelectedID] = useState(0);

    const columns: ColumnDef<Host>[] = [
        {
            id: "select",
            header: ({ table }) => (
                <Checkbox
                    checked={table.getIsAllPageRowsSelected() || (table.getIsSomePageRowsSelected() && "indeterminate")}
                    onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
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
                    <Link to={`/nodes/${name}`} className="hover:underline">
                        {name}
                    </Link>
                );
            },
        },
        {
            accessorKey: "provision",
            header: ({ column }) => <DataTableColumnHeader column={column} title="Provision" />,
            cell: ({ row }) => <ProvisionIcon provision={row.original.provision} name={row.original.name} />,
        },
        {
            accessorKey: "boot_image",
            header: ({ column }) => <DataTableColumnHeader column={column} title="Boot Image" />,
        },
        {
            accessorKey: "tags",
            header: ({ column }) => <DataTableColumnHeader column={column} title="Tags" />,
            cell: ({ row }) => {
                const tags = row.original.tags;
                return <TagsList tags={tags} />;
            },
            filterFn: (row, _, filterValue: string[]) => {
                const match = filterValue.filter((filterTag) => {
                    if (row.original.tags == undefined) {
                        return false;
                    }

                    if (filterTag.startsWith("!")) {
                        return !row.original.tags.includes(filterTag.substring(1));
                    } else if (filterTag.startsWith(":") || filterTag.endsWith(":")) {
                        const match = row.original.tags.filter((tag) => tag.includes(filterTag)) ?? [];
                        return match.length > 0;
                    }

                    return row.original.tags.includes(filterTag);
                });

                return match.length === filterValue.length;
            },
        },
    ];

    const actions: DataTableActions<Host> = ({ table }) => (
        <Actions
            checked={table
                .getSelectedRowModel()
                .rows.map((v) => v.getAllCells()[1].getValue())
                .join(",")}
            length={table.getSelectedRowModel().rows.length}
        />
    );
    return (
        <div className="px-6">
            {isSuccess && data != undefined && (
                <DataTable columns={columns} data={data} add={"/add/node"} Actions={actions} />
            )}
        </div>
    );
}
