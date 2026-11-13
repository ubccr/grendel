import { BmcResponseModelBmcFirmwareUpdates, getV1BmcFirmwareUpdates } from "@/client";
import ActionsSheet from "@/components/actions-sheet";
import FirmwareUpgradeAction from "@/components/actions/firmware/upgrade";
import NodesJobsAction from "@/components/actions/nodes/jobs";
import { DataTable, DataTableActions } from "@/components/data-table/data-table";
import { DataTableColumnHeader } from "@/components/data-table/header";
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
import { createFileRoute, Link } from "@tanstack/react-router";
import type { DataTableColumnDef, DataTableRow } from "@/components/data-table/features";
import { CircleAlert, OctagonAlert, TriangleAlert } from "lucide-react";

type NodeFirmware = NonNullable<BmcResponseModelBmcFirmwareUpdates["results"]>[number];

export const Route = createFileRoute("/nodes/dell-firmware/$nodes")({
  component: RouteComponent,
  beforeLoad: AuthRedirect,
  loader: async ({ params: { nodes } }) => getV1BmcFirmwareUpdates({ query: { nodeset: nodes } }),
});

// columns are defined once, outside the component, so the table keeps the same column definitions across renders
const columns: DataTableColumnDef<NodeFirmware, any>[] = [
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
];

function RouteComponent() {
  const { data } = Route.useLoaderData();
  const { isFetching } = Route.useMatch();

  const renderSubComponent = ({ row }: { row: DataTableRow<NodeFirmware> }) => {
    const data = row.original.data?.updates;
    return (
      <Table>
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

  const getRowCanExpand = (row: DataTableRow<NodeFirmware>) => {
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
          selectable
          getRowId={(result, i) => result.node ?? String(i)}
          columns={columns}
          renderSubComponent={renderSubComponent}
          getRowCanExpand={getRowCanExpand}
          getRowTone={(row) =>
            !row.original.success ? "error" : getRowCanExpand(row) ? "warn" : undefined
          }
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
