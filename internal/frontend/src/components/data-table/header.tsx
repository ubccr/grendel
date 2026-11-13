import type { RowData } from "@tanstack/react-table";
import { ArrowDown, ArrowUp, ArrowUpDown, Info, X } from "lucide-react";
import { createContext, useContext, useEffect, useRef, useState } from "react";

import { TagsInput } from "@/components/tags-input";
import { Button } from "@/components/ui/button";
import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupInput,
} from "@/components/ui/input-group";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";
import type { DataTableColumn } from "./features";

// DataTableFilterContext carries a DataTable's filter setter to its headers, which puts the filter in the url and clears the selection as well as filtering. A header outside a DataTable sets the column's filter directly
export const DataTableFilterContext = createContext<((id: string, value: unknown) => void) | null>(
  null,
);

interface DataTableColumnHeaderProps<TData extends RowData> {
  column: DataTableColumn<TData, any>;
  title: string;
  className?: string;
}

// DataTableColumnHeader is a column's whole header in one line: a filter box named for the column, with the sort control at its end. The name is the box's placeholder and gives way to what is typed, since the column being filtered is already known. A column that cannot be filtered shows its name with the sort control beside it
export function DataTableColumnHeader<TData extends RowData>({
  column,
  title,
  className,
}: DataTableColumnHeaderProps<TData>) {
  const filterValue = column.getFilterValue();
  const setFilter =
    useContext(DataTableFilterContext) ?? ((_: string, v: unknown) => column.setFilterValue(v));
  const sort = column.getCanSort() ? <SortButton column={column} title={title} /> : null;

  if (!column.getCanFilter()) {
    return (
      <div className={cn("flex h-8 items-center gap-1", className)}>
        <span>{title}</span>
        {sort}
      </div>
    );
  }

  if (column.id === "tags") {
    return (
      <div className={cn("flex items-center gap-1", className)}>
        <TagsFilter
          title={title}
          value={(filterValue as string[]) ?? []}
          onValueChange={(val) => setFilter(column.id, val)}
        />
        {sort}
      </div>
    );
  }

  return (
    <FilterBox
      title={title}
      className={className}
      value={(filterValue as string) ?? ""}
      onValueChange={(v) => setFilter(column.id, v)}
      sort={sort}
    />
  );
}

// FilterBox is a column's filter input. It keeps what is typed in its own state and hands it to the table after a short pause, so typing never waits on the table filtering and rendering its rows. A change made elsewhere, such as clearing all filters, still reaches the box
function FilterBox({
  title,
  className,
  value,
  onValueChange,
  sort,
}: {
  title: string;
  className?: string;
  value: string;
  onValueChange: (value: string) => void;
  sort: React.ReactNode;
}) {
  const [draft, setDraft] = useState(value);
  const sent = useRef(value);
  const timer = useRef<ReturnType<typeof setTimeout>>(undefined);

  // take a change that did not come from this box, but not the echo of one it sent
  useEffect(() => {
    if (value !== sent.current) {
      sent.current = value;
      setDraft(value);
    }
  }, [value]);

  useEffect(() => () => clearTimeout(timer.current), []);

  const send = (v: string, delay: number) => {
    setDraft(v);
    sent.current = v;
    clearTimeout(timer.current);
    timer.current = setTimeout(() => onValueChange(v), delay);
  };

  return (
    <InputGroup className={cn("h-8 min-w-28 bg-background font-normal", className)}>
      <InputGroupInput
        className="h-8 px-2.5 placeholder:font-medium placeholder:text-muted-foreground"
        placeholder={title}
        aria-label={`Filter ${title}`}
        value={draft}
        onChange={(e) => send(e.target.value, 150)}
      />
      <InputGroupAddon align="inline-end" className="gap-0.5">
        {draft && (
          <InputGroupButton
            size="icon-xs"
            aria-label={`Clear ${title} filter`}
            onClick={() => send("", 0)}
          >
            <X />
          </InputGroupButton>
        )}
        {sort}
      </InputGroupAddon>
    </InputGroup>
  );
}

// SortButton cycles the column through its sort orders and back to unsorted. It stays faint until the column is sorted, so the active sort is the one that stands out across a row of headers
function SortButton<TData extends RowData>({
  column,
  title,
}: {
  column: DataTableColumn<TData, any>;
  title: string;
}) {
  const sorted = column.getIsSorted();
  const next = column.getNextSortingOrder();
  const label =
    next === "asc"
      ? `Sort ${title} ascending`
      : next === "desc"
        ? `Sort ${title} descending`
        : `Stop sorting by ${title}`;
  return (
    <InputGroupButton
      size="icon-xs"
      aria-label={label}
      title={label}
      className={cn(!sorted && "text-muted-foreground/60 hover:text-foreground")}
      onClick={column.getToggleSortingHandler()}
    >
      {sorted === "asc" ? <ArrowUp /> : sorted === "desc" ? <ArrowDown /> : <ArrowUpDown />}
    </InputGroupButton>
  );
}

function TagsFilter({
  title,
  value,
  onValueChange,
}: {
  title: string;
  value: string[];
  onValueChange: (val: string[]) => void;
}) {
  return (
    <div className="flex w-full items-center gap-1">
      <TagsInput
        placeholder={title}
        className="w-full"
        value={value}
        onValueChange={onValueChange}
      />
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              type="button"
              size="icon"
              variant="ghost"
              className="size-6 text-muted-foreground"
            >
              <Info />
              <span className="sr-only">Tag filter syntax</span>
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            Valid search operators include: <br />
            "example" filters for rows that contain example tag <br />
            "!example" filters for rows that do not contain example tag <br />
            "key=" matches any value with key <br />
            "=value" matches any key with value <br />
            "namespace:" matches any key and value in a namespace <br />
            ":key" matches any namespace and value with that key <br />
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    </div>
  );
}
