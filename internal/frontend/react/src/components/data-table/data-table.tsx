import {
    ColumnDef,
    ColumnFiltersState,
    flexRender,
    getCoreRowModel,
    getFilteredRowModel,
    getPaginationRowModel,
    getSortedRowModel,
    SortingState,
    useReactTable,
    VisibilityState,
} from "@tanstack/react-table";

import { Table as TanstackTable } from "@tanstack/react-table";

import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { useState } from "react";
import { Input } from "../ui/input";
import { DataTablePagination } from "./pagination";
import { DataTableViewOptions } from "./view-options";
import { Button } from "../ui/button";
import { Info, Plus } from "lucide-react";
import { Link } from "@tanstack/react-router";
import { TagsInput } from "../tags-input";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "../ui/tooltip";

interface DataTableProps<TData, TValue> {
    columns: ColumnDef<TData, TValue>[];
    data: TData[];
    add?: string;
    Actions?: DataTableActions<TData>;
}

export type DataTableActions<TData> = ({ table }: { table: TanstackTable<TData> }) => JSX.Element;

export function DataTable<TData, TValue>({ columns, data, add, Actions }: DataTableProps<TData, TValue>) {
    const [sorting, setSorting] = useState<SortingState>([]);
    const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
    const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({});
    const [rowSelection, setRowSelection] = useState({});

    const table = useReactTable({
        data,
        columns,
        getCoreRowModel: getCoreRowModel(),
        getPaginationRowModel: getPaginationRowModel(),
        onSortingChange: setSorting,
        getSortedRowModel: getSortedRowModel(),
        onColumnFiltersChange: (e) => {
            table.setRowSelection({});
            setColumnFilters(e);
        },
        getFilteredRowModel: getFilteredRowModel(),
        onColumnVisibilityChange: setColumnVisibility,
        onRowSelectionChange: setRowSelection,
        enableRowSelection: true,
        // getRowId: (row) => row.name, // TODO: ?
        state: {
            sorting,
            columnFilters,
            columnVisibility,
            rowSelection,
        },
    });

    return (
        <div>
            <div className="flex items-center justify-between py-4">
                <div className="flex gap-2"></div>
                <div className="flex gap-2">
                    {add && (
                        <Button variant="outline" size="sm" asChild>
                            <Link to={add}>
                                <Plus />
                                <span className="sr-only sm:not-sr-only">Add</span>
                            </Link>
                        </Button>
                    )}
                    <DataTableViewOptions table={table} />
                    {!!Actions && <Actions table={table} />}
                </div>
            </div>
            <div className="rounded-md border">
                <Table>
                    <TableHeader>
                        {table.getHeaderGroups().map((headerGroup) => (
                            <TableRow key={headerGroup.id} className="*:p-2">
                                {headerGroup.headers.map((header) => {
                                    if (header.id == "select") {
                                        return <TableHead></TableHead>;
                                    } else if (header.id === "tags") {
                                        return (
                                            <div className="flex gap-2">
                                                <TagsInput
                                                    placeholder="Filter tags"
                                                    className="w-full"
                                                    value={
                                                        (table.getColumn("tags")?.getFilterValue() as string[]) ?? []
                                                    }
                                                    onValueChange={(val) =>
                                                        table.getColumn("tags")?.setFilterValue(val)
                                                    }
                                                />
                                                <Tooltip>
                                                    <TooltipProvider>
                                                        <TooltipTrigger>
                                                            <Button type="button" size="sm" variant="outline">
                                                                <Info />
                                                            </Button>
                                                        </TooltipTrigger>
                                                        <TooltipContent>
                                                            Valid search operators include: <br />
                                                            "example" filters for rows that contain example tag <br />
                                                            "!example" filters for rows that do not contain example tag{" "}
                                                            <br />
                                                            "key:" matches any value with key <br />
                                                            ":value" matches any key with value <br />
                                                        </TooltipContent>
                                                    </TooltipProvider>
                                                </Tooltip>
                                            </div>
                                        );
                                    }
                                    return (
                                        <TableHead key={header.id}>
                                            <Input
                                                placeholder={`Filter ${header.id}`}
                                                value={table.getColumn(header.id)?.getFilterValue() as string}
                                                onChange={(e) =>
                                                    table.getColumn(header.id)?.setFilterValue(e.target.value)
                                                }
                                            />
                                        </TableHead>
                                    );
                                })}
                            </TableRow>
                        ))}
                        {table.getHeaderGroups().map((headerGroup) => (
                            <TableRow key={headerGroup.id}>
                                {headerGroup.headers.map((header) => {
                                    return (
                                        <TableHead key={header.id}>
                                            {header.isPlaceholder
                                                ? null
                                                : flexRender(header.column.columnDef.header, header.getContext())}
                                        </TableHead>
                                    );
                                })}
                            </TableRow>
                        ))}
                    </TableHeader>
                    <TableBody>
                        {table.getRowModel().rows?.length ? (
                            table.getRowModel().rows.map((row) => (
                                <TableRow key={row.id} data-state={row.getIsSelected() && "selected"}>
                                    {row.getVisibleCells().map((cell) => (
                                        <TableCell key={cell.id}>
                                            {flexRender(cell.column.columnDef.cell, cell.getContext())}
                                        </TableCell>
                                    ))}
                                </TableRow>
                            ))
                        ) : (
                            <TableRow>
                                <TableCell colSpan={columns.length} className="h-24 text-center">
                                    No results.
                                </TableCell>
                            </TableRow>
                        )}
                    </TableBody>
                </Table>
            </div>
            <div className="py-4">
                <DataTablePagination table={table} />
            </div>
        </div>
    );
}
