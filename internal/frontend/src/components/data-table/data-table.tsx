import {
  FlexRender,
  useTable,
  type Cell,
  type ColumnFiltersState,
  type ColumnVisibilityState,
  type DisplayColumnDef,
  type RowData,
  type SortingState,
} from "@tanstack/react-table";

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { cn } from "@/lib/utils";
import { Link, useNavigate, useRouter, useSearch } from "@tanstack/react-router";
import { ChevronRight, ChevronsDownUp, ChevronsUpDown, Plus, RefreshCw } from "lucide-react";
import React, {
  createContext,
  JSX,
  useContext,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
} from "react";
import { Badge } from "../ui/badge";
import { Button } from "../ui/button";
import { Checkbox } from "../ui/checkbox";
import { Progress } from "../ui/progress";
import {
  features,
  type DataTableColumnDef,
  type DataTableFeatures,
  type DataTableInstance,
  type DataTableRow,
} from "./features";
import { DataTableFilterContext } from "./header";
import { DataTablePagination } from "./pagination";
import { DataTableViewOptions } from "./view-options";

// RowTone colors the rail on the start edge of a group or expandable row, so a failed node stands out in a long list without reading each row
export type RowTone = "ok" | "warn" | "error";

// DataTableTree makes a table of groups, each a top level row with the rows it holds as sub rows, like a node and its jobs. groupColumns are the columns that describe a group: they get the first header row and lay out the group rows, every other column gets the second header row and lays out the rows under each group
export type DataTableTree<TData extends RowData> = {
  getSubRows: (row: TData) => TData[] | undefined;
  groupColumns: string[];
};

interface DataTableProps<TData extends RowData> {
  columns: DataTableColumnDef<TData, any>[];
  data: TData[];
  add?: string;
  Actions?: DataTableActions<TData>;
  renderSubComponent?: (props: { row: DataTableRow<TData> }) => React.ReactElement;
  getRowCanExpand?: (row: DataTableRow<TData>) => boolean;
  getRowTone?: (row: DataTableRow<TData>) => RowTone | undefined;
  initialVisibility?: ColumnVisibilityState;
  initialSorting?: SortingState;
  progress?: boolean;
  refresh?: () => void;
  // selectable adds a checkbox column, with shift click selecting every row shown between two clicks
  selectable?: boolean;
  tree?: DataTableTree<TData>;
  // getRowId names a row by what it is rather than where it is, for example a node by its id. Selection is kept by row id, and the default id is the row's position, so without this a selection would move to whatever rows took the selected positions after a refresh
  getRowId?: (row: TData, index: number) => string;
  // meta is handed to TanStack as the table's meta, so cells can read state that changes often, like the values of inputs in them, without the column definitions changing and remounting every cell
  meta?: object;
  // getRowProps adds attributes to a row's <tr>, such as hover handlers
  getRowProps?: (row: DataTableRow<TData>) => React.HTMLAttributes<HTMLTableRowElement>;
  // paginate false shows every row on one page, with no page controls
  paginate?: boolean;
  // scrollHeight caps the table's height, scrolling its body under a header that stays in view
  scrollHeight?: string;
  // syncFilters false keeps column filters out of the url, for a table on a page whose search parameters are its own
  syncFilters?: boolean;
  // renderGroupDetail draws a full width row under each group row of a tree, shown while the group is open or when it has no rows of its own
  renderGroupDetail?: (row: DataTableRow<TData>) => React.ReactNode;
  // reveal opens the groups holding the row with this id and scrolls it into view, each time at changes
  reveal?: { id: string; at: number };
  // selection makes the table selectable with the selection held by the parent, as row ids. It is kept through filtering and through rows not yet loaded, so a table used to pick rows can build its pick up across several filters
  selection?: { ids: string[]; onChange: (ids: string[]) => void };
  // canSelect decides which rows get a checkbox. In a tree only rows without sub rows are selected, and a group's box ticks the rows under it
  canSelect?: (row: TData) => boolean;
  // bulkSelect is which rows the select all and group boxes tick, for rows that can be ticked one at a time but shouldn't be swept up with the rest
  bulkSelect?: (row: TData) => boolean;
}

export type DataTableActions<TData extends RowData> = ({
  table,
}: {
  table: DataTableInstance<TData>;
}) => JSX.Element;

const SELECT_COLUMN = "select";
const EXPAND_COLUMN = "expand";
const LEAD_COLUMNS = new Set([SELECT_COLUMN, EXPAND_COLUMN]);

const toneRail: Record<RowTone, string> = {
  ok: "before:bg-emerald-500/70",
  warn: "before:bg-amber-500/80",
  error: "before:bg-red-500/80",
};

// interactive reports whether a click landed on a control inside a row, so clicking a link, checkbox or button in a group row does not also open or close it
function interactive(target: EventTarget | null) {
  return (
    target instanceof Element &&
    !!target.closest("a,button,input,label,[role=checkbox],[role=switch],[data-no-toggle]")
  );
}

// columnId is the id TanStack gives a column definition: its id, or its accessorKey with dots turned into underscores
function columnId<TData extends RowData>(def: DataTableColumnDef<TData, any>) {
  if (def.id) return def.id;
  return "accessorKey" in def ? String(def.accessorKey).replace(/\./g, "_") : undefined;
}

export function DataTable<TData extends RowData>({
  columns,
  data,
  add,
  Actions,
  renderSubComponent,
  getRowCanExpand,
  getRowTone,
  initialVisibility,
  initialSorting,
  progress,
  refresh,
  selectable,
  tree,
  getRowId,
  meta,
  getRowProps,
  paginate = true,
  scrollHeight,
  syncFilters = true,
  renderGroupDetail,
  reveal,
  selection,
  canSelect,
  bulkSelect,
}: DataTableProps<TData>) {
  const navigate = useNavigate();
  const router = useRouter();
  const search = useSearch({ strict: false }) as Record<string, unknown>;

  const hasDetail = renderSubComponent !== undefined;

  const checkboxes = selectable || selection !== undefined;
  const allColumns = useMemo<DataTableColumnDef<TData, any>[]>(
    () => [...(checkboxes ? [selectColumn] : []), ...(hasDetail ? [expandColumn] : []), ...columns],
    [columns, checkboxes, hasDetail],
  );

  // the filters the url holds, taking only the parameters that name a column, since a page can have search parameters of its own
  const urlFilters = useMemo(
    () =>
      Object.entries(syncFilters ? search : {})
        .filter(([id]) => allColumns.some((c) => columnId(c) === id))
        .map(([id, value]) => ({ id, value })) as ColumnFiltersState,
    [search, allColumns, syncFilters],
  );

  const selectedIds = selection?.ids;
  const rowSelection = useMemo(
    () =>
      selectedIds
        ? (Object.fromEntries(selectedIds.map((id) => [id, true])) as Record<string, true>)
        : undefined,
    [selectedIds],
  );
  const onSelection = selection?.onChange;

  const table = useTable({
    features,
    ...(rowSelection && {
      state: { rowSelection },
      onRowSelectionChange: (u: unknown) => {
        const next = (typeof u === "function" ? u(rowSelection) : u) as Record<string, true>;
        onSelection?.(Object.keys(next).filter((id) => next[id]));
      },
    }),
    columns: allColumns,
    data,
    getRowId,
    getSubRows: tree?.getSubRows,
    getRowCanExpand,
    enableRowSelection: (row) =>
      (tree === undefined || row.subRows.length === 0) && (canSelect?.(row.original) ?? true),
    meta: meta as never,
    // a filter on a group's rows keeps every group holding a match, rather than judging the group row itself
    filterFromLeafRows: tree !== undefined,
    // keep a group's rows on the same page as the group, rather than letting them spill onto the next one
    paginateExpandedRows: false,
    // keep groups open while filtering, so a filter on a job column shows the matching jobs rather than closing the nodes holding them
    autoResetExpanded: false,
    initialState: {
      sorting: initialSorting ?? [],
      columnVisibility: initialVisibility ?? {},
      // filters start from the url
      columnFilters: urlFilters,
      // groups start open, since the rows under a node are what the page is for. Tables with detail panels start closed, as each panel can be large
      expanded: tree ? true : {},
      pagination: { pageIndex: 0, pageSize: paginate ? 20 : Number.MAX_SAFE_INTEGER },
    },
  });

  // when the data changes, drop the selection of rows that are gone, such as nodes just deleted. Selected ids stay in the selection state after their rows go, where they would still count towards "some rows selected" and leave the select all box partly checked with nothing visibly selected
  const controlled = selection !== undefined;
  useEffect(() => {
    if (controlled) return;
    table.setRowSelection((prev) => {
      const rows = table.getCoreRowModel().rowsById;
      const kept = Object.keys(prev).filter((id) => id in rows);
      return kept.length === Object.keys(prev).length
        ? prev
        : Object.fromEntries(kept.map((id) => [id, prev[id]]));
    });
  }, [data, table, controlled]);

  // the url can also change under the table, as when a link elsewhere on the page leads back to it with other filters, and the table then takes its filters from it. The table's own changes are left alone: it already holds them, and the url trails behind while typing, so taking it then would put back an earlier value
  const ownChanges = useRef(0);
  const urlKey = JSON.stringify(urlFilters);
  const seenUrlKey = useRef(urlKey);
  useEffect(() => {
    if (urlKey === seenUrlKey.current) return;
    seenUrlKey.current = urlKey;
    if (ownChanges.current > 0) return;
    if (!controlled) table.resetRowSelection();
    table.setColumnFilters(JSON.parse(urlKey) as ColumnFiltersState);
  }, [urlKey, table, controlled]);
  const own = (navigated: Promise<void>) => {
    ownChanges.current++;
    void navigated.finally(() => ownChanges.current--);
  };

  // every filter change goes through here, so the filter lands in the url and the selection is cleared, since it could otherwise hold rows the filter now hides
  const setFilter = (id: string, value: unknown) => {
    if (syncFilters)
      own(navigate({ to: ".", search: (prev) => ({ ...prev, [id]: value }), replace: true }));
    if (!controlled) table.resetRowSelection();
    table.getColumn(id)?.setFilterValue(value);
  };

  const filters = table.state.columnFilters;
  const clearFilters = () => {
    if (syncFilters)
      own(
        navigate({
          to: ".",
          search: (prev) =>
            Object.fromEntries(
              Object.entries(prev as Record<string, unknown>).filter(
                ([k]) => !filters.some((f) => f.id === k),
              ),
            ),
          replace: true,
        }),
      );
    if (!controlled) table.resetRowSelection();
    table.resetColumnFilters(true);
  };

  const bodyRef = useRef<HTMLDivElement>(null);
  const revealId = reveal?.id;
  const revealAt = reveal?.at;
  // keyed on the id and time rather than the object, so a parent that builds a new reveal object each render doesn't scroll back to the row on every render
  // the table is read through a ref, as the instance can change between renders and would otherwise re-run the reveal; a layout effect runs before the effect below, so it always sees the current one
  const tableRef = useRef(table);
  useLayoutEffect(() => {
    tableRef.current = table;
  });
  useEffect(() => {
    if (revealId === undefined) return;
    // a row the filters hide can't be shown, so they are cleared first
    if (!tableRef.current.getFilteredRowModel().flatRows.some((r) => r.id === revealId))
      tableRef.current.resetColumnFilters(true);
    const row = tableRef.current.getCoreRowModel().rowsById[revealId] as
      | DataTableRow<TData>
      | undefined;
    for (let parent = row?.getParentRow(); parent; parent = parent.getParentRow())
      if (!parent.getIsExpanded()) parent.toggleExpanded(true);
    // the opened group draws on the next frame, so the row is looked for after it
    requestAnimationFrame(() =>
      bodyRef.current
        ?.querySelector(`[data-row-id="${CSS.escape(revealId)}"]`)
        ?.scrollIntoView({ block: "nearest", behavior: "smooth" }),
    );
  }, [revealId, revealAt]);

  const leaf = table.getVisibleLeafColumns();
  const headerById = new Map(table.getFlatHeaders().map((h) => [h.column.id, h]));
  const groupIds = new Set(tree?.groupColumns ?? []);
  // a tree's leading columns (selection, disclosure) are shared, the group's own columns lay out its group rows and the rest lay out the rows under it, so those rows start at the left edge rather than after the group's columns
  const lead = tree ? leaf.filter((c) => LEAD_COLUMNS.has(c.id)) : [];
  const groupCols = tree ? leaf.filter((c) => groupIds.has(c.id)) : [];
  const itemCols = tree ? leaf.filter((c) => !groupIds.has(c.id) && !LEAD_COLUMNS.has(c.id)) : leaf;
  const span = Math.max(groupCols.length, itemCols.length);
  const width = tree ? lead.length + span : leaf.length;
  // the last column of a band stretches over whatever the other band has beyond it, so both bands fill the row
  const stretch = (i: number, n: number) => (i === n - 1 ? span - n + 1 : undefined);

  const expandable = tree !== undefined || hasDetail;
  const allExpanded = table.getIsAllRowsExpanded();
  const rows = table.getRowModel().rows;

  const headCell = (
    id: string,
    opts: { colSpan?: number; rowSpan?: number; className?: string; guide?: boolean } = {},
  ) => {
    const header = headerById.get(id);
    if (!header) return null;
    const sorted = header.column.getIsSorted();
    return (
      <TableHead
        key={id}
        colSpan={opts.colSpan}
        rowSpan={opts.rowSpan}
        aria-sort={sorted === "asc" ? "ascending" : sorted === "desc" ? "descending" : undefined}
        className={cn("h-auto py-1.5", LEAD_COLUMNS.has(id) && "w-10", opts.className)}
      >
        {opts.guide && <TreeGuide last />}
        {header.isPlaceholder ? null : <FlexRender header={header} />}
      </TableHead>
    );
  };

  const railClass = (tone: RowTone | undefined) =>
    tone &&
    cn(
      "relative before:absolute before:inset-y-1 before:left-0 before:w-[3px] before:rounded-full",
      toneRail[tone],
    );

  const renderRow = (row: DataTableRow<TData>, i: number) => {
    const isGroup = tree !== undefined && row.depth === 0;
    const isChild = row.depth > 0;
    const lastChild = isChild && rows[i + 1]?.depth !== row.depth;
    const tone = getRowTone?.(row);
    const extra = getRowProps?.(row);
    const open = row.getIsExpanded() && row.getCanExpand();
    const rowClass = cn(
      // group rows sit on the card and the rows under them are tinted, so a node's rows read as one block set into the list
      isGroup && "font-medium",
      isGroup && row.getCanExpand() && "cursor-pointer",
      isGroup && open && "border-b-transparent",
      isChild && "border-b-border/50 bg-muted/30 motion-safe:animate-in motion-safe:fade-in-0",
      isChild && lastChild && "border-b-border",
      hasDetail && open && "border-b-transparent bg-muted/25",
    );
    const onClick =
      isGroup && row.getCanExpand()
        ? (e: React.MouseEvent) => !interactive(e.target) && row.toggleExpanded()
        : undefined;

    let body: React.ReactNode;
    if (tree) {
      const cells = new Map(row.getVisibleCells().map((c) => [c.column.id, c]));
      const band = isGroup ? groupCols : itemCols;
      body = (
        <>
          {lead.map((c, ci) => (
            <TableCell key={c.id} className={cn("w-10", ci === 0 && railClass(tone))}>
              <FlexRender cell={cells.get(c.id)!} />
            </TableCell>
          ))}
          {band.map((c, ci) => (
            <TableCell
              key={c.id}
              colSpan={stretch(ci, band.length)}
              className={cn(isChild && ci === 0 && "relative pl-9")}
            >
              {isChild && ci === 0 && <TreeGuide last={lastChild} />}
              {isGroup && ci === 0 ? (
                <GroupLabel cell={cells.get(c.id)!} />
              ) : (
                <FlexRender cell={cells.get(c.id)!} />
              )}
            </TableCell>
          ))}
        </>
      );
    } else {
      body = row.getVisibleCells().map((cell, ci) => (
        <TableCell key={cell.id} className={cn(ci === 0 && railClass(tone))}>
          <FlexRender cell={cell} />
        </TableCell>
      ));
    }

    return (
      <React.Fragment key={row.id}>
        <TableRow
          {...extra}
          data-row-id={row.id}
          data-state={row.getIsSelected() ? "selected" : undefined}
          data-expanded={open || undefined}
          onClick={onClick}
          className={cn(rowClass, extra?.className)}
        >
          {body}
        </TableRow>
        {isGroup && renderGroupDetail && (open || row.subRows.length === 0) && (
          <TableRow className="hover:bg-transparent">
            <TableCell colSpan={width} className="px-3 pt-0 pb-3 whitespace-normal">
              {renderGroupDetail(row)}
            </TableCell>
          </TableRow>
        )}
        {hasDetail && open && (
          <TableRow className="bg-muted/25 hover:bg-muted/25">
            <TableCell colSpan={width} className="relative pt-0 pb-3 pl-10">
              <span aria-hidden className="absolute top-0 bottom-3 left-[1.25rem] w-px bg-border" />
              <div className="rounded-md border bg-background motion-safe:animate-in motion-safe:fade-in-0 motion-safe:slide-in-from-top-1">
                {renderSubComponent({ row })}
              </div>
            </TableCell>
          </TableRow>
        )}
      </React.Fragment>
    );
  };

  return (
    <div className="grid gap-2">
      <div className="flex flex-wrap items-center justify-between gap-2 p-2">
        <div className="flex items-center gap-2">
          {expandable && (
            <Button
              variant="secondary"
              onClick={() => table.toggleAllRowsExpanded(!allExpanded)}
              disabled={!table.getCanSomeRowsExpand()}
            >
              {allExpanded ? <ChevronsDownUp /> : <ChevronsUpDown />}
              <span className="sr-only sm:not-sr-only">
                {allExpanded ? "Collapse all" : "Expand all"}
              </span>
            </Button>
          )}
        </div>
        <div className="flex flex-wrap justify-end gap-2">
          {add && (
            <Button variant="secondary" render={<Link to={add} />} nativeButton={false}>
              <Plus />
              <span className="sr-only sm:not-sr-only">Add</span>
            </Button>
          )}
          {/* use passed refresh func (tanstack query) or fallback to route cache invalidation*/}
          <Button variant="secondary" onClick={refresh ?? (() => router.invalidate())}>
            <RefreshCw className={cn(progress && "motion-safe:animate-spin")} />
            <span className="sr-only sm:not-sr-only">Refresh</span>
          </Button>
          <DataTableViewOptions table={table} />
          {!!Actions && <Actions table={table} />}
        </div>
      </div>
      <BulkSelectContext.Provider value={bulkSelect as BulkSelect}>
        <DataTableFilterContext.Provider value={setFilter}>
          <div
            ref={bodyRef}
            className={cn("rounded-md border", scrollHeight ? "overflow-auto" : "overflow-hidden")}
            style={scrollHeight ? { maxHeight: scrollHeight } : undefined}
          >
            <Table>
              <TableHeader
                className={cn(
                  scrollHeight && "sticky top-0 z-10 bg-background shadow-[0_1px_0_var(--border)]",
                )}
              >
                {tree ? (
                  <>
                    {/* the leading cells span both rows, so select all sits once beside both bands */}
                    {/* each header row takes the colour of the rows it heads: plain for the group rows, tinted for the rows under them */}
                    <TableRow className="border-b-border/50 hover:bg-transparent">
                      {lead.map((c) => headCell(c.id, { rowSpan: 2, className: "border-b" }))}
                      {groupCols.map((c, ci) =>
                        headCell(c.id, {
                          colSpan: stretch(ci, groupCols.length),
                          className: ci === 0 ? "pl-6" : undefined,
                        }),
                      )}
                    </TableRow>
                    {/* the row band's header hangs off the group band's with the same elbow its rows use, so it reads as the header for the rows under each group */}
                    <TableRow className="bg-muted/30 hover:bg-muted/30">
                      {itemCols.map((c, ci) =>
                        headCell(c.id, {
                          colSpan: stretch(ci, itemCols.length),
                          className: ci === 0 ? "relative pl-9" : undefined,
                          guide: ci === 0,
                        }),
                      )}
                    </TableRow>
                  </>
                ) : (
                  <TableRow className="bg-muted/40 hover:bg-muted/40">
                    {leaf.map((c) => headCell(c.id))}
                  </TableRow>
                )}
                {/* the loading bar's row takes no height and no border, the header's [&_tr]:border-b outranks a plain border-0, and the bar overlays the top edge of the body, so the header keeps a single bottom border whether or not anything is loading */}
                <TableRow className="border-0! hover:bg-transparent">
                  <TableCell className="relative h-0 p-0" colSpan={width}>
                    {progress && (
                      <Progress
                        value={null}
                        className="absolute inset-x-0 top-0 z-10 *:data-[slot=progress-track]:h-0.5 *:data-[slot=progress-track]:rounded-none"
                      />
                    )}
                  </TableCell>
                </TableRow>
              </TableHeader>
              <TableBody>
                {rows.length ? (
                  rows.map(renderRow)
                ) : (
                  <TableRow className="hover:bg-transparent">
                    <TableCell colSpan={width} className="h-28 text-center">
                      {filters.length > 0 ? (
                        <div className="grid justify-items-center gap-2">
                          <span className="text-muted-foreground">
                            No rows match these filters.
                          </span>
                          <Button variant="secondary" size="sm" onClick={clearFilters}>
                            Clear filters
                          </Button>
                        </div>
                      ) : (
                        <span className="text-muted-foreground">Nothing to show.</span>
                      )}
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
        </DataTableFilterContext.Provider>
      </BulkSelectContext.Provider>
      {paginate && <DataTablePagination table={table} tree={tree !== undefined} />}
    </div>
  );
}

type BulkSelect = ((row: unknown) => boolean) | undefined;

const BulkSelectContext = createContext<BulkSelect>(undefined);

// selectColumn is the checkbox column a selectable table gets
const selectColumn: DisplayColumnDef<DataTableFeatures, any> = {
  id: SELECT_COLUMN,
  enableHiding: false,
  enableSorting: false,
  enableColumnFilter: false,
  header: ({ table }) => (
    <BulkBox label="Select all" rows={table.getFilteredRowModel().flatRows} table={table} />
  ),
  cell: ({ row, table }) =>
    row.subRows.length > 0 ? (
      <BulkBox label="Select group" rows={row.getLeafRows()} table={table} group />
    ) : row.getCanSelect() ? (
      <Checkbox
        aria-label="Select row"
        checked={row.getIsSelected()}
        // without this a shift click also selects the page's text between the two rows
        onMouseDown={(e) => e.shiftKey && e.preventDefault()}
        // TanStack reads the new value from the target and a shift click from the event, then selects the range from the row clicked before, in the order rows are shown
        onClick={(e) =>
          row.getToggleSelectedHandler()({
            target: { checked: !row.getIsSelected() },
            shiftKey: e.shiftKey,
          })
        }
      />
    ) : null,
};

// BulkBox is the box over many rows: the header's over every row shown, a group's over its rows. It shows whether all, some or none of them are selected, and ticks the ones bulkSelect allows, or clears them all
function BulkBox<TData extends RowData>({
  label,
  rows,
  table,
  group,
}: {
  label: string;
  rows: DataTableRow<TData>[];
  table: DataTableInstance<TData>;
  // a group with nothing selectable under it shows no box, where the header's stays, disabled
  group?: boolean;
}) {
  const bulk = useContext(BulkSelectContext);
  const leaves = rows.filter((r) => r.subRows.length === 0 && r.getCanSelect());
  const eligible = bulk ? leaves.filter((r) => bulk(r.original)) : leaves;
  const all = eligible.length > 0 && eligible.every((r) => r.getIsSelected());
  const some = leaves.some((r) => r.getIsSelected());
  if (group && leaves.length === 0) return null;
  return (
    <Checkbox
      aria-label={label}
      checked={all}
      indeterminate={some && !all}
      disabled={eligible.length === 0}
      onClick={() =>
        table.setRowSelection((prev) => {
          const next = { ...prev };
          for (const r of all ? leaves : eligible) {
            if (all) delete next[r.id];
            else next[r.id] = true;
          }
          return next;
        })
      }
    />
  );
}

// expandColumn is the disclosure column a table with detail panels gets
const expandColumn: DisplayColumnDef<DataTableFeatures, any> = {
  id: EXPAND_COLUMN,
  enableHiding: false,
  enableSorting: false,
  enableColumnFilter: false,
  header: () => <span className="sr-only">Details</span>,
  cell: ({ row }) => (row.getCanExpand() ? <Disclosure row={row} /> : null),
};

// GroupLabel is the group row's lead cell: the disclosure, the group's own value and, at the cell's end, how many rows the group holds. An empty group's 0 is drawn faintly, so nodes with nothing to show recede, and the counts line up down the column whatever the length of the name
function GroupLabel<TData extends RowData>({
  cell,
}: {
  cell: Cell<DataTableFeatures, TData, unknown>;
}) {
  const row = cell.row;
  const count = row.subRows.length;
  return (
    <div className="flex items-center gap-1.5">
      {row.getCanExpand() ? <Disclosure row={row} /> : <span className="size-6 shrink-0" />}
      <FlexRender cell={cell} />
      <Badge
        variant={count ? "secondary" : "outline"}
        className={cn(
          "ml-auto rounded-sm tabular-nums",
          !count && "font-normal text-muted-foreground",
        )}
      >
        {count}
      </Badge>
    </div>
  );
}

// Disclosure is the chevron that opens a row. It is a real button so the row can be opened from the keyboard, and it turns rather than swapping icons so the change reads as the same control moving
function Disclosure<TData extends RowData>({ row }: { row: DataTableRow<TData> }) {
  const open = row.getIsExpanded();
  return (
    <Button
      variant="ghost"
      size="icon"
      className="size-6 shrink-0 text-muted-foreground hover:text-foreground"
      aria-expanded={open}
      aria-label={open ? "Collapse row" : "Expand row"}
      onClick={(e) => {
        e.stopPropagation();
        row.toggleExpanded();
      }}
    >
      <ChevronRight
        className={cn(
          "motion-safe:transition-transform motion-safe:duration-150",
          open && "rotate-90",
        )}
      />
    </Button>
  );
}

// TreeGuide is the line a group's rows hang from, centred under the group row's chevron, with an elbow on the last row
function TreeGuide({ last }: { last: boolean }) {
  return (
    <span aria-hidden className="pointer-events-none absolute inset-0">
      <span
        className={cn("absolute top-0 left-[1.25rem] w-px bg-border", last ? "h-1/2" : "bottom-0")}
      />
      <span className="absolute top-1/2 left-[1.25rem] h-px w-2.5 bg-border" />
    </span>
  );
}
