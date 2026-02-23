import { getV1Nodes, Host } from "@/client";
import ActionsSheet from "@/components/actions-sheet";
import NodeActions from "@/components/actions/nodes";
import { DataTable, DataTableActions } from "@/components/data-table/data-table";
import { DataTableColumnHeader } from "@/components/data-table/header";
import SelectableCheckbox from "@/components/data-table/selectableCheckbox";
import ProvisionIcon from "@/components/nodes/provision-button";
import TagsList from "@/components/tags";
import { Card, CardContent } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import AuthRedirect from "@/lib/auth";
import { createFileRoute, Link } from "@tanstack/react-router";
import { ColumnDef } from "@tanstack/react-table";
import { useState } from "react";

export const Route = createFileRoute("/nodes/")({
  component: TableComponent,
  beforeLoad: AuthRedirect,
  loader: () => getV1Nodes(),
});

function TableComponent() {
  const { data } = Route.useLoaderData();
  const { isFetching } = Route.useMatch();

  const [lastSelectedID, setLastSelectedID] = useState(0);

  const columns: ColumnDef<Host>[] = [
    {
      id: "select",
      header: ({ table }) => (
        <Checkbox
          checked={table.getIsAllRowsSelected()}
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
      accessorKey: "name",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Name" />,
      cell: ({ row }) => {
        const name = row.original?.name;
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
      accessorKey: "provision",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Provision" />,
      cell: ({ row }) => (
        <ProvisionIcon provision={row.original?.provision} name={row.original?.name} />
      ),
      filterFn: (row, columId, filterValue) => {
        const val = row.getValue(columId) as boolean;
        return `${val}`.includes(filterValue);
      },
    },
    {
      accessorKey: "boot_image",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Boot Image" />,
    },
    {
      accessorKey: "tags",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Tags" />,
      cell: ({ row }) => {
        const tags = row.original?.tags;
        return <TagsList tags={tags ?? []} />;
      },
      filterFn: (row, _, filterValue: string[]) => {
        const match = filterValue.filter((filterTag) => {
          if (row.original?.tags == undefined) {
            return false;
          }

          if (filterTag.startsWith("!")) {
            return !row.original?.tags.includes(filterTag.substring(1));
          } else if (filterTag.startsWith("=") || filterTag.endsWith("=")) {
            const match = row.original?.tags.filter((tag) => tag.includes(filterTag)) ?? [];
            return match.length > 0;
          } else if (filterTag.startsWith(":") || filterTag.endsWith(":")) {
            const match = row.original?.tags.filter((tag) => tag.includes(filterTag)) ?? [];
            return match.length > 0;
          }

          return row.original?.tags.includes(filterTag);
        });

        return match.length === filterValue.length;
      },
    },
    {
      accessorKey: "interfaces.ip",
      header: ({ column }) => <DataTableColumnHeader column={column} title="IP Addresses" />,
      cell: ({ row }) => {
        const data = row.original.interfaces?.map((iface) => iface?.ip);
        return <span>{data?.join(" ")}</span>;
      },
      filterFn: (row, _, filterValue: string) => {
        const data = row.original.interfaces?.map((iface) => iface?.ip);
        if (!data) return false;

        return data.join(" ").includes(filterValue);
      },
    },
    {
      accessorKey: "interfaces.fqdn",
      header: ({ column }) => <DataTableColumnHeader column={column} title="FQDNs" />,
      cell: ({ row }) => {
        const data = row.original.interfaces?.map((iface) => iface?.fqdn);
        return <span>{data?.join(" ")}</span>;
      },
      filterFn: (row, _, filterValue: string) => {
        const data = row.original.interfaces?.map((iface) => iface?.fqdn);
        if (!data) return false;

        return data.join(" ").includes(filterValue);
      },
    },
    {
      accessorKey: "interfaces.mac",
      header: ({ column }) => <DataTableColumnHeader column={column} title="MACs" />,
      cell: ({ row }) => {
        const data = row.original.interfaces?.map((iface) => iface?.mac);
        return <span>{data?.join(" ")}</span>;
      },
      filterFn: (row, _, filterValue: string) => {
        const data = row.original.interfaces?.map((iface) => iface?.mac);
        if (!data) return false;

        return data.join(" ").includes(filterValue);
      },
    },
  ];

  const actions: DataTableActions<Host> = ({ table }) => {
    const checked = table
      .getSelectedRowModel()
      .rows.map((v) => v.getAllCells()[1].getValue())
      .join(",");
    const length = table.getSelectedRowModel().rows.length;
    return (
      <ActionsSheet checked={checked} length={length}>
        <NodeActions nodes={checked} />
      </ActionsSheet>
    );
  };
  return (
    <Card>
      <CardContent>
        <DataTable
          columns={columns}
          data={data ?? []}
          add={"/add/node"}
          Actions={actions}
          initialVisibility={{
            interfaces_ip: false,
            interfaces_fqdn: false,
            interfaces_mac: false,
          }}
          progress={isFetching !== false}
        />
      </CardContent>
    </Card>
  );
}
