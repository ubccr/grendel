import type { Host } from "@/client";
import { DataTable } from "@/components/data-table/data-table";
import type { DataTableColumnDef } from "@/components/data-table/features";
import { DataTableColumnHeader } from "@/components/data-table/header";
import {
  byName,
  nameColumn,
  ReadyNote,
  tagsColumn,
  useNodeList,
} from "@/components/nodes/node-table";
import { Skeleton } from "@/components/ui/skeleton";
import { bootPort } from "@/lib/boot-nics";
import { useMemo, type ReactNode } from "react";

const bmcAddress = (h: Host) => h.interfaces?.find((i) => i.bmc && i.ip)?.ip?.split("/")[0] ?? "";

// readable is whether Grendel can ask the node's BMC for its interfaces
export const readable = (h: Host) => !!bmcAddress(h);

const others = (h: Host) => (h.interfaces ?? []).filter((i) => !i.bmc);

const columns: DataTableColumnDef<Host, any>[] = [
  nameColumn,
  {
    id: "bmc",
    accessorFn: bmcAddress,
    header: ({ column }) => <DataTableColumnHeader column={column} title="BMC address" />,
    cell: ({ row }) => <span className="font-mono">{bmcAddress(row.original)}</span>,
  },
  {
    id: "interfaces",
    accessorFn: (h) =>
      others(h)
        .map((i) => `${i.mac} ${i.ip}`)
        .join(" "),
    header: ({ column }) => <DataTableColumnHeader column={column} title="Interfaces" />,
    cell: ({ row }) => {
      const list = others(row.original);
      const boot = bootPort(row.original);
      if (list.length === 0) return <span className="text-muted-foreground">None</span>;
      return (
        <span className="grid leading-tight">
          <span className="font-mono">{boot?.mac}</span>
          <span className="text-xs text-muted-foreground">
            {boot?.ip?.split("/")[0] ?? "No address"}
            {list.length > 1 && `, and ${list.length - 1} more`}
          </span>
        </span>
      );
    },
  },
  tagsColumn,
  {
    id: "problem",
    accessorFn: (h) => (readable(h) ? "Ready" : "No BMC address"),
    header: ({ column }) => <DataTableColumnHeader column={column} title="Status" />,
    cell: ({ row }) => (
      <ReadyNote problem={readable(row.original) ? undefined : "No BMC address"} />
    ),
  },
];

// children is the step's bar, drawn under the table
export function NodePicker({
  selected,
  onChange,
  children,
}: {
  selected: string[];
  onChange: (names: string[]) => void;
  children?: ReactNode;
}) {
  const { data, isLoading, isFetching, refetch } = useNodeList();
  const nodes = useMemo(
    () => (data ?? []).filter((h) => !h.tags?.includes("switch")).sort(byName),
    [data],
  );

  if (isLoading)
    return (
      <div className="grid gap-2">
        {[0, 1, 2].map((i) => (
          <Skeleton key={i} className="h-11 w-full" />
        ))}
      </div>
    );

  return (
    <div className="grid gap-4">
      <DataTable
        selection={{ ids: selected, onChange }}
        getRowId={(h) => h.name ?? ""}
        columns={columns}
        data={nodes}
        syncFilters={false}
        refresh={() => void refetch()}
        progress={isFetching}
        getRowTone={(row) => (readable(row.original) ? undefined : "warn")}
      />
      {children}
    </div>
  );
}
