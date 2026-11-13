import type { ReactTable } from "@tanstack/react-table";
import type { RowData } from "@tanstack/react-table";
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { DataTableFeatures, DataTableRow } from "./features";

interface DataTablePaginationProps<TData extends RowData> {
  table: ReactTable<DataTableFeatures, TData>;
  tree?: boolean;
}

export function DataTablePagination<TData extends RowData>({
  table,
  tree,
}: DataTablePaginationProps<TData>) {
  // a tree counts the rows its groups hold, the jobs or entries, rather than the nodes holding them
  const counted = (r: DataTableRow<TData>) => !tree || r.depth > 0;
  const { pageIndex, pageSize } = table.state.pagination;
  const pageCount = Math.max(table.getPageCount(), 1);
  const selected = table.getSelectedRowModel().flatRows.filter(counted).length;
  const total = table.getFilteredRowModel().flatRows.filter(counted).length;
  const groups = tree ? table.getPrePaginatedRowModel().rows.length : 0;

  return (
    <div className="flex flex-wrap items-center justify-between gap-x-6 gap-y-2 px-2 text-sm text-muted-foreground tabular-nums">
      <div aria-live="polite">
        {selected > 0 ? (
          <span className="text-foreground">
            {selected} of {total} selected
          </span>
        ) : groups > 0 ? (
          <span>
            {total} {total === 1 ? "row" : "rows"} across {groups}{" "}
            {groups === 1 ? "group" : "groups"}
          </span>
        ) : (
          <span>
            {total} {total === 1 ? "row" : "rows"}
          </span>
        )}
      </div>
      <div className="flex items-center gap-6">
        <div className="flex items-center gap-2">
          <span>{groups > 0 ? "Groups per page" : "Rows per page"}</span>
          <Select value={`${pageSize}`} onValueChange={(value) => table.setPageSize(Number(value))}>
            <SelectTrigger className="h-8 w-[70px]" aria-label="Rows per page">
              <SelectValue placeholder={pageSize} />
            </SelectTrigger>
            <SelectContent side="top">
              {[10, 20, 30, 40, 50].map((size) => (
                <SelectItem key={size} value={`${size}`}>
                  {size}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div className="flex items-center gap-1">
          <span className="mr-2">
            Page {pageIndex + 1} of {pageCount}
          </span>
          <Button
            variant="ghost"
            size="icon"
            className="hidden size-8 lg:inline-flex"
            onClick={() => table.firstPage()}
            disabled={!table.getCanPreviousPage()}
          >
            <span className="sr-only">Go to first page</span>
            <ChevronsLeft />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="size-8"
            onClick={() => table.previousPage()}
            disabled={!table.getCanPreviousPage()}
          >
            <span className="sr-only">Go to previous page</span>
            <ChevronLeft />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="size-8"
            onClick={() => table.nextPage()}
            disabled={!table.getCanNextPage()}
          >
            <span className="sr-only">Go to next page</span>
            <ChevronRight />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="hidden size-8 lg:inline-flex"
            onClick={() => table.lastPage()}
            disabled={!table.getCanNextPage()}
          >
            <span className="sr-only">Go to last page</span>
            <ChevronsRight />
          </Button>
        </div>
      </div>
    </div>
  );
}
