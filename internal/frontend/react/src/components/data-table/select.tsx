import { Table } from "@tanstack/react-table";

import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuLabel,
    DropdownMenuRadioGroup,
    DropdownMenuRadioItem,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Separator } from "@/components/ui/separator";

interface DataTableSelectProps<TData> {
    table: Table<TData>;
}

export function DataTableSelect<TData>({ table }: DataTableSelectProps<TData>) {
    return (
        <DropdownMenu>
            <DropdownMenuTrigger>Select</DropdownMenuTrigger>
            <DropdownMenuContent>
                <DropdownMenuLabel>Select</DropdownMenuLabel>
                <Separator />
                <DropdownMenuRadioGroup onValueChange={(value) => changeValue(value, table)}>
                    {/* <DropdownMenuRadioItem value="none">None</DropdownMenuRadioItem> */}
                    <DropdownMenuRadioItem value="page">Page</DropdownMenuRadioItem>
                    <DropdownMenuRadioItem value="all">All</DropdownMenuRadioItem>
                </DropdownMenuRadioGroup>
                {/* <DropdownMenuCheckboxItem
                    checked={table.getIsAllPageRowsSelected() || (table.getIsSomePageRowsSelected() && "indeterminate")}
                    onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}>
                    Page
                </DropdownMenuCheckboxItem>
                <DropdownMenuCheckboxItem
                    checked={table.getIsAllRowsSelected()}
                    onCheckedChange={(value) => table.toggleAllRowsSelected(!!value)}>
                    All
                </DropdownMenuCheckboxItem> */}
            </DropdownMenuContent>
        </DropdownMenu>
    );
}

function changeValue<TData>(value: string, table: Table<TData>) {
    switch (value) {
        // case "none":
        //     table.toggleAllRowsSelected(false)
        //     break
        case "page":
            table.toggleAllPageRowsSelected(true);
            break;
        case "all":
            table.toggleAllRowsSelected(true);
            break;
    }
}
