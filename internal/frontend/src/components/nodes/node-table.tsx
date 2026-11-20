import type { Host } from "@/client";
import { getV1NodesOptions } from "@/client/@tanstack/react-query.gen";
import type { DataTableColumnDef } from "@/components/data-table/features";
import { DataTableColumnHeader } from "@/components/data-table/header";
import TagsList from "@/components/tags";
import { matchesTags } from "@/lib/tags";
import { useQuery } from "@tanstack/react-query";
import { TriangleAlert } from "lucide-react";

// useNodeList is every node in Grendel, shared by the tables that pick nodes and the steps that read what was picked
export const useNodeList = () => useQuery({ ...getV1NodesOptions(), staleTime: 60_000 });

export const byName = (a: Host, b: Host) =>
  (a.name ?? "").localeCompare(b.name ?? "", undefined, { numeric: true });

// ReadyNote is a picker row's status: Ready, or the amber reason its node can't be used
export function ReadyNote({ problem }: { problem?: string }) {
  if (!problem) return <span className="text-muted-foreground">Ready</span>;
  return (
    <span className="flex items-center gap-1.5 whitespace-nowrap text-amber-700 dark:text-amber-300">
      <TriangleAlert className="size-3.5 shrink-0" />
      {problem}
    </span>
  );
}

export const nameColumn: DataTableColumnDef<Host, any> = {
  accessorKey: "name",
  header: ({ column }) => <DataTableColumnHeader column={column} title="Name" />,
  cell: ({ row }) => <span className="font-mono">{row.original.name}</span>,
};

export const tagsColumn: DataTableColumnDef<Host, any> = {
  accessorKey: "tags",
  header: ({ column }) => <DataTableColumnHeader column={column} title="Tags" />,
  cell: ({ row }) => <TagsList tags={row.original.tags ?? []} />,
  filterFn: (row, _, filterValue: string[]) => matchesTags(row.original.tags, filterValue),
};
