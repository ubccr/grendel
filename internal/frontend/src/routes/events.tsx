import {
  getV1GrendelEventsIdOptions,
  getV1GrendelEventsOptions,
} from "@/client/@tanstack/react-query.gen";
import { NodeLink } from "@/components/nodes/node-link";
import { DataTable, type RowTone } from "@/components/data-table/data-table";
import type { DataTableColumnDef, DataTableRow } from "@/components/data-table/features";
import { DataTableColumnHeader } from "@/components/data-table/header";
import { ErrorPanel } from "@/components/error-panel";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { RunRing } from "@/components/run-ring";
import AuthRedirect from "@/lib/auth";
import { eventRows, isRunning, nodeNote, nodeSeverity, type EventRow } from "@/lib/events";
import { relativeTime, severityOf, severityVariant, type Severity } from "@/lib/utils";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { LoaderCircle } from "lucide-react";
import { useMemo } from "react";

export const Route = createFileRoute("/events")({
  staticData: { crumb: "Events" },
  component: RouteComponent,
  beforeLoad: AuthRedirect,
});

// how often the list is fetched while an event is running and while none is
const LIVE_POLL_MS = 5_000;
const POLL_MS = 30_000;

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
      const { succeeded, failed, going, stalled, total } = row.original;
      if (total === 0) return null;
      return (
        <span
          className="flex items-center gap-2 whitespace-nowrap tabular-nums"
          title={row.original.result}
        >
          <RunRing succeeded={succeeded} failed={failed} going={going} stalled={stalled} />
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
    // an event still running has no outcome yet, so it says that it is running instead
    cell: ({ row, getValue }) =>
      row.original.state === "running" ? (
        <Badge variant="outline">running</Badge>
      ) : (
        <Badge variant={severityVariant(getValue<string>())}>{getValue<string>()}</Badge>
      ),
  },
  {
    accessorKey: "kind",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Kind" />,
  },
  {
    accessorKey: "state",
    header: ({ column }) => <DataTableColumnHeader column={column} title="State" />,
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
];

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
  return severity ? tones[severity] : undefined;
}

// an event opens to show its nodes when it has any
const hasNodes = (row: DataTableRow<EventRow>) => row.original.total > 0;

function RouteComponent() {
  const { data, isFetching, refetch } = useQuery({
    ...getV1GrendelEventsOptions(),
    refetchInterval: (q) => (q.state.data?.some(isRunning) ? LIVE_POLL_MS : POLL_MS),
  });
  const rows = useMemo(() => eventRows(data ?? []), [data]);

  return (
    <Card>
      <CardContent className="p-2">
        <DataTable
          columns={columns}
          data={rows}
          getRowId={(row) => row.id}
          getRowTone={rowTone}
          getRowCanExpand={hasNodes}
          renderSubComponent={({ row }) => <EventNodes event={row.original} />}
          initialSorting={[{ id: "time", desc: true }]}
          initialVisibility={{ run: false, kind: false, state: false }}
          progress={isFetching}
          refresh={() => void refetch()}
        />
      </CardContent>
    </Card>
  );
}

// EventNodes is how an opened event went on each of its nodes. They are fetched only while it is open, and again whenever the list shows the event changed, as every write to its nodes says when it was updated, so a running event's panel keeps up with no polling of its own and a finished one is fetched once
function EventNodes({ event }: { event: EventRow }) {
  const options = getV1GrendelEventsIdOptions({ path: { id: event.id } });
  const { data, error, isPending } = useQuery({
    ...options,
    queryKey: [{ ...options.queryKey[0], tags: [event.updated] }],
    // the nodes as last fetched stay up while the newer ones load
    placeholderData: keepPreviousData,
  });

  if (error) return <ErrorPanel title="Could not load the nodes" message={error.detail} />;
  if (isPending) {
    return (
      <div className="flex items-center gap-2 p-3 text-sm text-muted-foreground">
        <LoaderCircle className="size-4 animate-spin" /> Loading nodes
      </div>
    );
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Node</TableHead>
          <TableHead>Status</TableHead>
          <TableHead>Detail</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {data.nodes?.map((n) => (
          <TableRow key={n.node}>
            <TableCell>
              <NodeLink node={n.node ?? ""} />
            </TableCell>
            <TableCell>
              <Badge variant={severityVariant(nodeSeverity(n.state))}>{n.state}</Badge>
            </TableCell>
            <TableCell className="whitespace-normal text-muted-foreground">
              <span className="line-clamp-2 wrap-anywhere">{nodeNote(n)}</span>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
