import { useGetV1BmcUpgradeDellRepo } from "@/openapi/queries";
import { DataTable, DataTableActions } from "../../data-table/data-table";
import { RedfishDellUpgradeFirmware } from "@/openapi/requests";
import { ColumnDef, Row } from "@tanstack/react-table";
import { DataTableColumnHeader } from "../../data-table/header";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Button } from "../../ui/button";
import {
  ChevronDown,
  ChevronLeft,
  CircleAlert,
  OctagonAlert,
  TriangleAlert,
} from "lucide-react";
import { Checkbox } from "../../ui/checkbox";
import SelectableCheckbox from "../../data-table/selectableCheckbox";
import { useState } from "react";
import ActionsSheet from "../../actions-sheet";
import DellFirmwareUpgradeActions from "./actions";

import { Switch } from "../../ui/switch";

export default function DellFirmwareUpgrade({ nodes }: { nodes: string }) {
  const initialData = nodes.split(",").map((v) => {
    return { Name: v, Status: "pending" };
  });
  const query = useGetV1BmcUpgradeDellRepo(
    { query: { nodeset: nodes } },
    undefined,
    { initialData: initialData },
  );

  const [lastSelectedID, setLastSelectedID] = useState(0);

  const columns: ColumnDef<RedfishDellUpgradeFirmware>[] = [
    {
      id: "select",
      header: ({ table }) => (
        <Checkbox
          checked={
            table.getIsAllPageRowsSelected() ||
            (table.getIsSomePageRowsSelected() && "indeterminate")
          }
          onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
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
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Node" />
      ),
    },
    {
      accessorKey: "Status",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Status" />
      ),
    },
    {
      accessorKey: "Message",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Message" />
      ),
    },
    {
      accessorKey: "UpdateCount",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Updates" />
      ),
    },
    {
      accessorKey: "UpdateRebootType",
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Reboot Type" />
      ),
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

  const renderSubComponent = ({
    row,
  }: {
    row: Row<RedfishDellUpgradeFirmware>;
  }) => {
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
                <TableCell>
                  {displayCriticality(row?.Criticality ?? "")}
                </TableCell>
                <TableCell>{row?.DisplayName}</TableCell>
                <TableCell className="text-muted-foreground">
                  {row?.InstalledVersion}
                </TableCell>
                <TableCell>{row?.PackageVersion}</TableCell>
                <TableCell
                  className={
                    row?.RebootType == "NONE" ? "text-muted-foreground" : ""
                  }
                >
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
        <DellFirmwareUpgradeActions nodes={checked} />
      </ActionsSheet>
    );
  };

  return (
    <div>
      <DataTable
        columns={columns}
        renderSubComponent={renderSubComponent}
        getRowCanExpand={getRowCanExpand}
        Actions={actions}
        data={query?.data ?? []}
        progress={query.isFetching}
        refresh={() => query.refetch()}
      />
    </div>
  );
}

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
