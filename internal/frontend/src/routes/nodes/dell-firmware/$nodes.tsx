import { getV1BmcUpgradeDellRepo, RedfishDellUpgradeFirmware } from "@/client";
import ActionsSheet from "@/components/actions-sheet";
import FirmwareUpgradeAction from "@/components/actions/firmware/upgrade";
import { DataTable, DataTableActions } from "@/components/data-table/data-table";
import { DataTableColumnHeader } from "@/components/data-table/header";
import SelectableCheckbox from "@/components/data-table/selectableCheckbox";
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
import { createFileRoute } from "@tanstack/react-router";
import { ColumnDef, Row } from "@tanstack/react-table";
import { ChevronDown, ChevronLeft, CircleAlert, OctagonAlert, TriangleAlert } from "lucide-react";
import { useState } from "react";

export const Route = createFileRoute("/nodes/dell-firmware/$nodes")({
  component: RouteComponent,
  beforeLoad: AuthRedirect,
  loader: async ({ params: { nodes } }) => getV1BmcUpgradeDellRepo({ query: { nodeset: nodes } }),
});

function RouteComponent() {
  const { data } = Route.useLoaderData();
  const { isFetching } = Route.useMatch();

  const [lastSelectedID, setLastSelectedID] = useState(0);
  const columns: ColumnDef<RedfishDellUpgradeFirmware>[] = [
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
      accessorKey: "Name",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Node" />,
    },
    {
      accessorKey: "Status",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Status" />,
    },
    {
      accessorKey: "Message",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Message" />,
    },
    {
      accessorKey: "UpdateCount",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Updates" />,
    },
    {
      accessorKey: "UpdateRebootType",
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

  const renderSubComponent = ({ row }: { row: Row<RedfishDellUpgradeFirmware> }) => {
    const data = row.original.UpdateList;
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
                <TableCell>{displayCriticality(row?.Criticality ?? "")}</TableCell>
                <TableCell>{row?.DisplayName}</TableCell>
                <TableCell className="text-muted-foreground">{row?.InstalledVersion}</TableCell>
                <TableCell>{row?.PackageVersion}</TableCell>
                <TableCell className={row?.RebootType == "NONE" ? "text-muted-foreground" : ""}>
                  {row?.RebootType}
                </TableCell>
              </TableRow>
            ))}
        </TableBody>
      </Table>
    );
  };

  const getRowCanExpand = (row: Row<RedfishDellUpgradeFirmware>) => {
    const len = row.original.UpdateList?.length ?? 0;
    if (len > 0) return true;
    return false;
  };

  const actions: DataTableActions<RedfishDellUpgradeFirmware> = ({ table }) => {
    const checked = table
      .getSelectedRowModel()
      .rows.map((v) => v.getAllCells()[1].getValue())
      .join(",");
    const length = table.getSelectedRowModel().rows.length;
    return (
      <ActionsSheet checked={checked} length={length}>
        <FirmwareUpgradeAction nodes={checked} />
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
          data={data ?? []}
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
