import { BmcResponseModelBmcFirmwareUpdates, getV1BmcFirmwareUpdates } from "@/client";
import ActionsSheet from "@/components/actions-sheet";
import FirmwareUpgradeAction from "@/components/actions/firmware/upgrade";
import NodesJobsAction from "@/components/actions/nodes/jobs";
import { DataTable, DataTableActions } from "@/components/data-table/data-table";
import { DataTableColumnHeader } from "@/components/data-table/header";
import SelectableCheckbox from "@/components/data-table/selectableCheckbox";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Switch } from "@/components/ui/switch";
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
import { createFileRoute, Link } from "@tanstack/react-router";
import { ColumnDef, Row } from "@tanstack/react-table";
import { ChevronDown, ChevronLeft, CircleAlert, OctagonAlert, TriangleAlert } from "lucide-react";
import { useState } from "react";

type NodeFirmware = NonNullable<BmcResponseModelBmcFirmwareUpdates["results"]>[number];

export const Route = createFileRoute("/nodes/dell-firmware/$nodes")({
  component: RouteComponent,
  beforeLoad: AuthRedirect,
  loader: async ({ params: { nodes } }) => getV1BmcFirmwareUpdates({ query: { nodeset: nodes } }),
});

function RouteComponent() {
  const { data } = Route.useLoaderData();
  const { isFetching } = Route.useMatch();

  const [lastSelectedID, setLastSelectedID] = useState(0);
  const columns: ColumnDef<NodeFirmware>[] = [
    {
      id: "select",
      header: ({ table }) => (
        <Checkbox
          checked={
            table.getIsAllRowsSelected() || (table.getIsSomeRowsSelected() && "indeterminate")
          }
          onCheckedChange={(value) => table.toggleAllRowsSelected(!!value)}
          aria-label="Select all"
        />
      ),
      cell: ({ row, table }) => (
        <SelectableCheckbox
          row={row}
          table={table}
          lastSelectedID={lastSelectedID}
          setLastSelectedID={setLastSelectedID}
        />
      ),
    },
    {
      accessorKey: "node",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Node" />,
      cell: ({ row }) => {
        const name = row.original?.node;
        return (
          <Link
            to={"/nodes/$node/node"}
            params={{ node: name ?? "unknown" }}
            className="hover:underline"
          >
            {name}
          </Link>
        );
      },
    },
    {
      id: "status",
      accessorFn: (row) => (row.success ? "success" : "failed"),
      header: ({ column }) => <DataTableColumnHeader column={column} title="Status" />,
      cell: ({ row, getValue }) => (
        <Badge
          variant="secondary"
          className={`rounded-sm ${severityColor(row.original.success ? "success" : "error")}`}
        >
          {getValue<string>()}
        </Badge>
      ),
    },
    {
      accessorKey: "message",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Message" />,
    },
    {
      id: "updates",
      accessorFn: (row) => row.data?.updates?.length ?? 0,
      header: ({ column }) => <DataTableColumnHeader column={column} title="Updates" />,
    },
    {
      id: "reboot_type",
      accessorFn: (row) => row.data?.reboot_type ?? "",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Reboot Type" />,
    },
    {
      id: "expand",
      enableColumnFilter: false,
      header: ({ table }) => (
        <Switch
          aria-description="expand all"
          checked={table.getIsAllRowsExpanded()}
          onCheckedChange={(value) => {
            table.toggleAllRowsExpanded(!!value);
          }}
        />
      ),
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

  const renderSubComponent = ({ row }: { row: Row<NodeFirmware> }) => {
    const data = row.original.data?.updates;
    return (
      <Table className="border p-2">
        <TableHeader>
          <TableRow>
            <TableHead>Criticality</TableHead>
            <TableHead>Name</TableHead>
            <TableHead>Installed Version</TableHead>
            <TableHead>Package Version</TableHead>
            <TableHead>Reboot Type</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data &&
            data.map((row, i) => (
              <TableRow key={i}>
                <TableCell>{displayCriticality(row?.criticality ?? "")}</TableCell>
                <TableCell>{row?.display_name}</TableCell>
                <TableCell className="text-muted-foreground">{row?.installed_version}</TableCell>
                <TableCell>{row?.package_version}</TableCell>
                <TableCell className={row?.reboot_type == "NONE" ? "text-muted-foreground" : ""}>
                  {row?.reboot_type}
                </TableCell>
              </TableRow>
            ))}
        </TableBody>
      </Table>
    );
  };

  const getRowCanExpand = (row: Row<NodeFirmware>) => {
    const len = row.original.data?.updates?.length ?? 0;
    if (len > 0) return true;
    return false;
  };

  const actions: DataTableActions<NodeFirmware> = ({ table }) => {
    const checked = table
      .getSelectedRowModel()
      .rows.map((v) => v.original.node)
      .join(",");
    const length = table.getSelectedRowModel().rows.length;
    return (
      <ActionsSheet type="node" checked={checked} length={length}>
        <div className="grid gap-4 sm:grid-cols-1">
          <FirmwareUpgradeAction nodes={checked} />
          <NodesJobsAction nodes={checked} />
        </div>
      </ActionsSheet>
    );
  };

  return (
    <Card>
      <CardContent className="p-2">
        <DataTable
          columns={columns}
          renderSubComponent={renderSubComponent}
          getRowCanExpand={getRowCanExpand}
          Actions={actions}
          data={data?.results ?? []}
          progress={isFetching !== false}
        />
      </CardContent>
    </Card>
  );
}

// TODO: move into component
function displayCriticality(criticality: string) {
  let icon = <></>;
  switch (criticality) {
    case "1":
      icon = <TriangleAlert className="text-yellow-500" />;
      break;
    case "2":
      icon = <OctagonAlert className="text-red-500" />;
      break;
    case "3":
      icon = <CircleAlert className="text-blue-500" />;
      break;
  }

  return <span>{icon}</span>;
}
