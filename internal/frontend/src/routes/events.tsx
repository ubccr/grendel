import type { Event } from "@/client";
import { getV1GrendelEvents } from "@/client";
import { DataTable } from "@/components/data-table/data-table";
import { DataTableColumnHeader } from "@/components/data-table/header";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import AuthRedirect from "@/lib/auth";
import { severityColor } from "@/lib/utils";
import { createFileRoute } from "@tanstack/react-router";
import type { ColumnDef, Row } from "@tanstack/react-table";
import { ChevronDown, ChevronLeft } from "lucide-react";

export const Route = createFileRoute("/events")({
  component: TableComponent,
  beforeLoad: AuthRedirect,
  loader: () => getV1GrendelEvents(),
});

function TableComponent() {
  const events = Route.useLoaderData();

  const columns: ColumnDef<Event>[] = [
    {
      accessorKey: "Severity",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Severity" />,
      cell: ({ row }) => (
        <Badge className={`rounded-sm ${severityColor(row.original.Severity)}`} variant="secondary">
          {row.original.Severity}
        </Badge>
      ),
    },
    {
      accessorKey: "Time",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Time" />,
      cell: ({ row }) => {
        const time = row.original?.Time;
        return <span>{new Date(time ?? "").toLocaleString()}</span>;
      },
    },
    {
      accessorKey: "User",
      header: ({ column }) => <DataTableColumnHeader column={column} title="User" />,
    },
    {
      accessorKey: "Message",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Message" />,
    },
    {
      id: "expand",
      enableColumnFilter: false,
      header: ({ column }) => <DataTableColumnHeader column={column} title="Expand" />,
      cell: ({ row }) => (
        <div>
          {row.getCanExpand() ? (
            <Button
              size="icon"
              variant="secondary"
              onClick={() => row.getCanExpand() && row.toggleExpanded()}
            >
              {row.getIsExpanded() ? <ChevronDown /> : <ChevronLeft />}
            </Button>
          ) : (
            <span></span>
          )}
        </div>
      ),
    },
  ];

  const renderSubComponent = ({ row }: { row: Row<Event> }) => {
    const data = row.original.JobMessages;
    return (
      <Table className="border p-2">
        <TableHeader>
          <TableRow>
            <TableHead>Status</TableHead>
            <TableHead>Host</TableHead>
            <TableHead>Message</TableHead>
            <TableHead>Error</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data?.map((row) => (
            <TableRow key={row.host}>
              <TableCell>
                <Badge variant="secondary" className={`rounded-sm ${severityColor(row.status)}`}>
                  {row.status}
                </Badge>
              </TableCell>
              <TableCell>{row.host}</TableCell>
              <TableCell>{row.msg}</TableCell>
              <TableCell>{row.redfish_error?.code && JSON.stringify(row.redfish_error)}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    );
  };

  const getRowCanExpand = (row: Row<Event>) => {
    if (row.original.JobMessages) return true;
    return false;
  };

  return (
    <Card>
      <CardContent>
        <DataTable
          columns={columns}
          data={events.data ?? []}
          // Actions={actions}
          renderSubComponent={renderSubComponent}
          getRowCanExpand={getRowCanExpand}
          // progress={isFetching}
          initialSorting={[{ id: "Time", desc: true }]}
        />
      </CardContent>
    </Card>
  );
}
