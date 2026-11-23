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
import { ErrorPanel } from "@/components/error-panel";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { switchNos } from "@/lib/switch-ports";
import { Link } from "@tanstack/react-router";
import { Network, RefreshCw } from "lucide-react";
import { useMemo, type ReactNode } from "react";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";

const protocol: Record<string, string> = { arista: "gNMI", sonic: "gNMI", dellos10: "SNMP" };

const address = (h: Host) => h.interfaces?.[0]?.ip?.split("/")[0] ?? "";

function switchProblem(h: Host) {
  if (!switchNos(h.tags)) return "No arista, sonic or dellos10 tag";
  if (!address(h)) return "No management address";
}

const columns: DataTableColumnDef<Host, any>[] = [
  nameColumn,
  {
    id: "nos",
    accessorFn: (h) => switchNos(h.tags) ?? "",
    header: ({ column }) => <DataTableColumnHeader column={column} title="OS" />,
    cell: ({ row }) => {
      const nos = switchNos(row.original.tags);
      return nos ? (
        <span>
          {nos} <span className="text-muted-foreground">over {protocol[nos]}</span>
        </span>
      ) : null;
    },
  },
  {
    id: "address",
    accessorFn: address,
    header: ({ column }) => <DataTableColumnHeader column={column} title="Management address" />,
    cell: ({ row }) => <span className="font-mono">{address(row.original)}</span>,
  },
  tagsColumn,
  {
    id: "problem",
    accessorFn: (h) => switchProblem(h) ?? "Ready",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Status" />,
    cell: ({ row }) => <ReadyNote problem={switchProblem(row.original)} />,
  },
];

// children is the step's bar, drawn under the table
export function SwitchPicker({
  selected,
  onChange,
  children,
}: {
  selected: string[];
  onChange: (names: string[]) => void;
  children?: ReactNode;
}) {
  const { data, isLoading, isFetching, error, refetch } = useNodeList();
  const switches = useMemo(
    () => (data ?? []).filter((h) => h.tags?.includes("switch")).sort(byName),
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

  if (error)
    return (
      <div className="grid justify-items-center gap-3">
        <ErrorPanel title="Couldn't list the switches" message={error.detail} />
        <Button variant="outline" onClick={() => void refetch()}>
          <RefreshCw />
          Try again
        </Button>
      </div>
    );

  if (switches.length === 0)
    return (
      <Empty className="border">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <Network />
          </EmptyMedia>
          <EmptyTitle>No switches in Grendel</EmptyTitle>
          <EmptyDescription>
            Add each switch as a node tagged <code>switch</code> and <code>arista</code>,{" "}
            <code>sonic</code> or <code>dellos10</code>, with its management address on its first
            interface.{" "}
            <Link to="/add/node" className="underline underline-offset-4">
              Add a node
            </Link>
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    );

  return (
    <div className="grid gap-4">
      <DataTable
        selection={{ ids: selected, onChange }}
        getRowId={(h) => h.name ?? ""}
        columns={columns}
        data={switches}
        syncFilters={false}
        refresh={() => void refetch()}
        progress={isFetching}
        getRowTone={(row) => (switchProblem(row.original) ? "warn" : undefined)}
      />
      {children}
    </div>
  );
}
