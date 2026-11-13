import {
  columnFilteringFeature,
  columnVisibilityFeature,
  createExpandedRowModel,
  createFilteredRowModel,
  createPaginatedRowModel,
  createSortedRowModel,
  filterFn_arrIncludes,
  filterFn_equals,
  filterFn_includesString,
  filterFn_inNumberRange,
  filterFn_weakEquals,
  rowExpandingFeature,
  rowPaginationFeature,
  rowSelectionFeature,
  rowSortingFeature,
  sortFn_alphanumeric,
  sortFn_basic,
  sortFn_datetime,
  sortFn_text,
  tableFeatures,
  type CellData,
  type Column,
  type ColumnDef,
  type Row,
  type RowData,
  type Table,
} from "@tanstack/react-table";

// features is every table feature the shared DataTable uses. TanStack Table v9 only gives a table the APIs of the features registered here, and the filter and sort names columns may use are the keys of the registries below. The filter and sort functions are the ones a column's "auto" setting picks from. It is defined once, outside any component, so its reference stays stable
export const features = tableFeatures({
  columnFilteringFeature,
  rowSortingFeature,
  columnVisibilityFeature,
  rowSelectionFeature,
  rowPaginationFeature,
  rowExpandingFeature,
  filteredRowModel: createFilteredRowModel(),
  sortedRowModel: createSortedRowModel(),
  expandedRowModel: createExpandedRowModel(),
  paginatedRowModel: createPaginatedRowModel(),
  filterFns: {
    includesString: filterFn_includesString,
    inNumberRange: filterFn_inNumberRange,
    equals: filterFn_equals,
    weakEquals: filterFn_weakEquals,
    arrIncludes: filterFn_arrIncludes,
  },
  sortFns: {
    alphanumeric: sortFn_alphanumeric,
    text: sortFn_text,
    datetime: sortFn_datetime,
    basic: sortFn_basic,
  },
});

export type DataTableFeatures = typeof features;
export type DataTableColumnDef<
  TData extends RowData,
  TValue extends CellData = CellData,
> = ColumnDef<DataTableFeatures, TData, TValue>;
export type DataTableRow<TData extends RowData> = Row<DataTableFeatures, TData>;
export type DataTableInstance<TData extends RowData> = Table<DataTableFeatures, TData>;
export type DataTableColumn<TData extends RowData, TValue extends CellData = CellData> = Column<
  DataTableFeatures,
  TData,
  TValue
>;
