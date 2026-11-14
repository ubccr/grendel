import { getV1GrendelEventsOptions } from "@/client/@tanstack/react-query.gen";
import { DataTable, type DataTableTree, type RowTone } from "@/components/data-table/data-table";
import type { DataTableColumnDef, DataTableRow } from "@/components/data-table/features";
import { DataTableColumnHeader } from "@/components/data-table/header";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { RunRing } from "@/components/run-ring";
import AuthRedirect from "@/lib/auth";
import { eventRows, type EventRow } from "@/lib/events";
import { relativeTime, severityOf, severityVariant, type Severity } from "@/lib/utils";
import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import { useMemo } from "react";

export const Route = createFileRoute("/events")({
  staticData: { crumb: "Events" },
  component: RouteComponent,
  beforeLoad: AuthRedirect,
});

// columns are defined once, outside the component, so the table keeps the same column definitions across renders
const columns: DataTableColumnDef<EventRow, any>[] = [
  {
    accessorKey: "message",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Event" />,
    cell: ({ getValue }) => (
      // a message can hold a long nodeset with no spaces, so it may break anywhere
      <span className="line-clamp-2 min-w-48 wrap-anywhere whitespace-normal" title={getValue()}>
        {getValue<string>()}
      </span>
    ),
  },
  {
    // the column sorts on the share of nodes that failed, so the worst runs come first sorting down and events with no nodes last, and filters on the words, such as "1 of 6 failed"
    id: "result",
    accessorFn: (row) => (row.total ? row.failed / row.total : -1),
    sortFn: "basic",
    filterFn: (row, _, value: string) =>
      row.original.result.toLowerCase().includes(String(value).toLowerCase()),
    header: ({ column }) => <DataTableColumnHeader column={column} title="Result" />,
    // the ring shows the share of nodes that succeeded and the count gives it exactly
    cell: ({ row }) => {
      const { succeeded, failed, total } = row.original;
      if (row.depth > 0 || total === 0) return null;
      return (
        <span
          className="flex items-center gap-2 whitespace-nowrap tabular-nums"
          title={row.original.result}
        >
          <RunRing succeeded={succeeded} failed={failed} />
          <span>
            {succeeded}/{total}
          </span>
        </span>
      );
    },
  },
  {
    accessorKey: "severity",
    // severities sort by how bad they are rather than by name
    sortFn: (a, b) => severityRank(a.original.severity) - severityRank(b.original.severity),
    header: ({ column }) => <DataTableColumnHeader column={column} title="Severity" />,
    cell: ({ getValue }) => (
      <Badge variant={severityVariant(getValue<string>())}>{getValue<string>()}</Badge>
    ),
  },
  {
    accessorKey: "user",
    header: ({ column }) => <DataTableColumnHeader column={column} title="User" />,
  },
  {
    accessorKey: "time",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Time" />,
    cell: ({ getValue }) => {
      const time = relativeTime(getValue<string>());
      return (
        <span className="whitespace-nowrap" title={time.title}>
          {time.text}
        </span>
      );
    },
  },
  {
    accessorKey: "run",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Run" />,
    cell: ({ getValue }) => <span className="font-mono text-xs">{getValue<string>()}</span>,
  },
  {
    accessorKey: "host",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Node" />,
    cell: ({ getValue }) => (
      <Link
        to="/nodes/$node/node"
        params={{ node: getValue<string>() }}
        className="hover:underline"
      >
        {getValue<string>()}
      </Link>
    ),
  },
  {
    accessorKey: "status",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Status" />,
    cell: ({ getValue }) => (
      <Badge variant={severityVariant(getValue<string>())}>{getValue<string>()}</Badge>
    ),
  },
];

// an event is a group, and the nodes a bmc action reached are the rows under it
const tree: DataTableTree<EventRow> = {
  getSubRows: (row) => row.items,
  groupColumns: ["message", "result", "severity", "user", "time", "run"],
};

const ranks: Record<Severity, number> = { success: 0, info: 1, warning: 2, error: 3 };

function severityRank(value?: string) {
  const severity = severityOf(value);
  return severity ? ranks[severity] : -1;
}

const tones: Partial<Record<Severity, RowTone>> = {
  success: "ok",
  warning: "warn",
  error: "error",
};

// an event's rail shows its severity, so a failed action stands out in the list; info events have none
function rowTone(row: DataTableRow<EventRow>) {
  const severity = severityOf(row.original.severity);
  return row.depth === 0 && severity ? tones[severity] : undefined;
}

function RouteComponent() {
  // the header's latest event polls this same query, so the table stays current without polling of its own
  const { data, isFetching, refetch } = useQuery(getV1GrendelEventsOptions());
  const rows = useMemo(() => eventRows(data ?? []), [data]);

  return (
    <Card>
      <CardContent className="p-2">
        <DataTable
          columns={columns}
          data={rows}
          tree={tree}
          getRowId={(row) => row.id}
          getRowTone={rowTone}
          initialSorting={[{ id: "time", desc: true }]}
          initialVisibility={{ run: false }}
          progress={isFetching}
          refresh={() => void refetch()}
        />
      </CardContent>
    </Card>
  );
}
