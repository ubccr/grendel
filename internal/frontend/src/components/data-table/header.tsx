import type { RowData } from "@tanstack/react-table";
import { ArrowDown, ArrowUp, ArrowUpDown, Info, X } from "lucide-react";
import { useEffect, useRef, useState } from "react";

import { TagCombobox } from "@/components/tag-combobox";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupInput,
} from "@/components/ui/input-group";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";
import type { DataTableColumn } from "./features";

interface DataTableColumnHeaderProps<TData extends RowData> {
  column: DataTableColumn<TData, any>;
  title: string;
  className?: string;
  // options turns the filter into a choice of these, for a column whose values are a few known kinds. Given here rather than in the column definition, they can carry counts that change without the columns being rebuilt
  options?: FilterOption[];
}

// The column name is the filter box's placeholder and gives way to what is typed. A column that cannot be filtered shows its name with the sort control beside it
export function DataTableColumnHeader<TData extends RowData>({
  column,
  title,
  className,
  options,
}: DataTableColumnHeaderProps<TData>) {
  const filterValue = column.getFilterValue();
  const sort = column.getCanSort() ? <SortButton column={column} title={title} /> : null;

  if (!column.getCanFilter()) {
    return (
      <div className={cn("flex h-8 items-center gap-1", className)}>
        <span>{title}</span>
        {sort}
      </div>
    );
  }

  if (options) {
    return (
      <div className={cn("flex items-center gap-1", className)}>
        <ChoiceFilter
          title={title}
          options={options}
          value={(filterValue as string) ?? ""}
          onValueChange={(v) => column.setFilterValue(v || undefined)}
        />
        {sort}
      </div>
    );
  }

  if (column.id === "tags") {
    return (
      <div className={cn("flex items-center gap-1", className)}>
        <TagsFilter
          title={title}
          // a url written by hand can hold one tag as plain text rather than a list
          value={
            Array.isArray(filterValue) ? filterValue : filterValue ? [String(filterValue)] : []
          }
          onValueChange={(val) => column.setFilterValue(val)}
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
      onValueChange={(v) => column.setFilterValue(v)}
      sort={sort}
    />
  );
}

export type FilterOption = { value: string; label: string; count?: number };

function ChoiceFilter({
  title,
  options,
  value,
  onValueChange,
}: {
  title: string;
  options: FilterOption[];
  value: string;
  onValueChange: (value: string) => void;
}) {
  const items = [{ value: "", label: title }, ...options];
  return (
    <Select items={items} value={value} onValueChange={(v) => onValueChange((v as string) ?? "")}>
      <SelectTrigger
        size="sm"
        aria-label={`Filter ${title}`}
        className={cn(
          "h-8 w-full min-w-28 bg-background font-normal",
          !value && "font-medium text-muted-foreground",
        )}
      >
        <SelectValue />
      </SelectTrigger>
      {/* sized to its options rather than the narrow header box, and dropped below it, so the counts line up in a column */}
      <SelectContent
        align="end"
        alignItemWithTrigger={false}
        className="w-auto min-w-(--anchor-width)"
      >
        <SelectItem value="">All</SelectItem>
        {options.map((o) => (
          <SelectItem key={o.value} value={o.value}>
            {o.label}
            {o.count !== undefined && (
              <span className="ml-auto pl-3 text-muted-foreground tabular-nums">{o.count}</span>
            )}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}

// FilterBox keeps what is typed in its own state and hands it to the table after a short pause, so typing never waits on the table filtering its rows. A change made elsewhere, such as clearing all filters, still reaches the box
export function FilterBox({
  title,
  hint,
  className,
  value,
  onValueChange,
  sort,
}: {
  title: string;
  // hint is shown on hover, for a box that reads more than plain text
  hint?: string;
  className?: string;
  value: string;
  onValueChange: (value: string) => void;
  sort?: React.ReactNode;
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
        title={hint}
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

// The button stays faint until the column is sorted, so the active sort stands out across a row of headers
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
      <TagCombobox
        placeholder={title}
        label={`Filter ${title}`}
        // match the other filter boxes in the header row, whose column name is a medium weight placeholder, with chips darker than the default so they stand out in the busy header
        className="min-h-8 py-1 **:data-[slot=combobox-chip]:bg-foreground/10 **:data-[slot=combobox-chip-input]:placeholder:font-medium dark:**:data-[slot=combobox-chip]:bg-black/40"
        value={value}
        onValueChange={onValueChange}
      />
      <Tooltip>
        <TooltipTrigger
          render={
            <Button
              type="button"
              size="icon"
              variant="ghost"
              className="size-6 text-muted-foreground"
            />
          }
        >
          <Info />
          <span className="sr-only">Tag filter syntax</span>
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
    </div>
  );
}
