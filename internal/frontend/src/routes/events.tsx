import AuthRedirect from "@/auth";
import { DataTable } from "@/components/data-table/data-table";
import { DataTableColumnHeader } from "@/components/data-table/header";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { severityColor } from "@/lib/utils";
import { useGetV1GrendelEvents } from "@/openapi/queries";
import { Event } from "@/openapi/requests";
import { createFileRoute } from "@tanstack/react-router";
import { ColumnDef, Row } from "@tanstack/react-table";
import { ChevronDown, ChevronLeft } from "lucide-react";

export const Route = createFileRoute("/events")({
  component: RouteComponent,
  beforeLoad: AuthRedirect,
});

function RouteComponent() {
  const { data, isSuccess } = useGetV1GrendelEvents();

  const columns: ColumnDef<Event>[] = [
    {
      accessorKey: "Severity",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Severity" />
      ),
      cell: ({ row }) => (
        <Badge
          className={"rounded-sm " + severityColor(row.original.Severity)}
          variant="outline"
        >
          {row.original.Severity}
        </Badge>
      ),
    },
    {
      accessorKey: "Time",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Time" />
      ),
      cell: ({ row }) => {
        const time = row.original?.Time;
        return <span>{new Date(time ?? "").toLocaleString()}</span>;
      },
    },
    {
      accessorKey: "User",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="User" />
      ),
    },
    {
      accessorKey: "Message",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Message" />
      ),
    },
    {
      id: "expand",
      enableColumnFilter: false,
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Expand" />
      ),
      cell: ({ row }) => (
        <div>
          {row.getCanExpand() ? (
            <Button
              size="sm"
              variant="outline"
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
      <Table className="p-2 border">
        <TableHeader>
          <TableRow>
            <TableHead>Status</TableHead>
            <TableHead>Host</TableHead>
            <TableHead>Message</TableHead>
            <TableHead>Error</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data &&
            data.map((row, i) => (
              <TableRow key={i}>
                <TableCell>
                  <Badge
                    variant="outline"
                    className={"rounded-sm " + severityColor(row.status)}
                  >
                    {row.status}
                  </Badge>
                </TableCell>
                <TableCell>{row.host}</TableCell>
                <TableCell>{row.msg}</TableCell>
                <TableCell>
                  {row.redfish_error?.code && JSON.stringify(row.redfish_error)}
                </TableCell>
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

  // const actions: DataTableActions<Event> = ({ table }) => {
  //   const checked = table
  //     .getSelectedRowModel()
  //     .rows.map((v) => v.getAllCells()[1].getValue())
  //     .join(",");
  //   const length = table.getSelectedRowModel().rows.length;
  //   return (
  //     <ActionsSheet checked={checked} length={length}>
  //       <NodeActions nodes={checked} length={length} />
  //     </ActionsSheet>
  //   );
  // };
  return (
    <div className="px-6">
      {isSuccess && (
        <DataTable
          columns={columns}
          data={data ?? []}
          // Actions={actions}
          renderSubComponent={renderSubComponent}
          getRowCanExpand={getRowCanExpand}
        />
      )}
    </div>
  );
}
