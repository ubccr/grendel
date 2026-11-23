import type { DataTableColumnDef, DataTableInstance } from "@/components/data-table/features";
import { DataTableColumnHeader } from "@/components/data-table/header";
import type { ReactNode } from "react";

export type StatusFilter<R> = { value: string; label: string; test: (r: R) => boolean };

// statusColumn is an import grid's status column: a choice filter of filters, each with how many of the grid's rows it matches, read from the table meta's rows so the counts follow edits
export function statusColumn<G extends { row?: R }, R extends { status: string }>(
  filters: StatusFilter<R>[],
  cell: (r: R, table: DataTableInstance<G>) => ReactNode,
): DataTableColumnDef<G, any> {
  return {
    id: "status",
    accessorFn: (g) => g.row?.status ?? "",
    enableSorting: false,
    enableHiding: false,
    header: ({ column, table }) => {
      const rows = (table.options.meta as { rows: R[] }).rows;
      return (
        <DataTableColumnHeader
          column={column}
          title="Status"
          options={filters.map((f) => ({
            value: f.value,
            label: f.label,
            count: rows.filter(f.test).length,
          }))}
        />
      );
    },
    filterFn: (row, _id, value) => {
      const r = row.original.row;
      const f = filters.find((x) => x.value === value);
      return !r || !f || f.test(r);
    },
    cell: ({ row, table }) => (
      <div className="whitespace-normal">{cell(row.original.row!, table)}</div>
    ),
  };
}
