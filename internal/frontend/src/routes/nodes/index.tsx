import { getV1Nodes, getV1NodesProvisionStatus, Host } from "@/client";
import { NodeLink } from "@/components/nodes/node-link";
import ActionsSheet from "@/components/actions-sheet";
import NodeActions from "@/components/actions/nodes";
import { DataTable, DataTableActions } from "@/components/data-table/data-table";
import { DataTableColumnHeader } from "@/components/data-table/header";
import ProvisionIcon from "@/components/nodes/provision-button";
import { ProvisionStatusBadge, statusesByNode } from "@/components/nodes/provision-status";
import TagsList from "@/components/tags";
import { Card, CardContent } from "@/components/ui/card";
import { matchesTags } from "@/lib/tags";
import AuthRedirect from "@/lib/auth";
import { stateLabels } from "@/lib/provision";
import { createFileRoute } from "@tanstack/react-router";
import type { DataTableColumnDef } from "@/components/data-table/features";
import { useMemo } from "react";

export const Route = createFileRoute("/nodes/")({
  component: TableComponent,
  beforeLoad: AuthRedirect,
  // the statuses load with the nodes so the table's refresh reloads both. A failure there only leaves the status column empty
  loader: async () => {
    const [nodes, statuses] = await Promise.all([getV1Nodes(), getV1NodesProvisionStatus()]);
    return { ...nodes, statuses: statuses.data ?? [] };
  },
});

function TableComponent() {
  const { data, statuses } = Route.useLoaderData();
  const { isFetching } = Route.useMatch();

  const statusByNode = useMemo(() => statusesByNode(statuses), [statuses]);

  // columns are memoized, so the table keeps the same column definitions across renders. They change only when the provision statuses do
  const columns = useMemo<DataTableColumnDef<Host, any>[]>(
    () => [
      {
        accessorKey: "name",
        header: ({ column }) => <DataTableColumnHeader column={column} title="Name" />,
        cell: ({ row }) => {
          const name = row.original?.name;
          return <NodeLink node={name ?? "unknown"}>{name}</NodeLink>;
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
        id: "provision_status",
        accessorFn: (row) => statusByNode.get(row.name ?? "")?.state ?? "",
        header: ({ column }) => <DataTableColumnHeader column={column} title="Status" />,
        cell: ({ row }) => (
          <ProvisionStatusBadge status={statusByNode.get(row.original?.name ?? "")} />
        ),
        // a whole state matches only that state, so pxe does not also list the nodes at ipxe; anything else typed matches part of a state
        filterFn: (row, columnId, filterValue: string) => {
          const state = row.getValue<string>(columnId);
          return Object.keys(stateLabels).includes(filterValue)
            ? state === filterValue
            : state.includes(filterValue);
        },
      },
      {
        accessorKey: "boot_image",
        header: ({ column }) => <DataTableColumnHeader column={column} title="Boot Image" />,
      },
      {
        accessorKey: "arch",
        header: ({ column }) => <DataTableColumnHeader column={column} title="Arch" />,
      },
      {
        accessorKey: "tags",
        header: ({ column }) => <DataTableColumnHeader column={column} title="Tags" />,
        cell: ({ row }) => {
          const tags = row.original?.tags;
          return <TagsList tags={tags ?? []} />;
        },
        filterFn: (row, _, filterValue: string[]) => matchesTags(row.original?.tags, filterValue),
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
    ],
    [statusByNode],
  );

  const actions: DataTableActions<Host> = ({ table }) => {
    const checked = table
      .getSelectedRowModel()
      .rows.map((v) => v.original.name)
      .join(",");
    const length = table.getSelectedRowModel().rows.length;
    return (
      <ActionsSheet type="node" checked={checked} length={length}>
        <NodeActions nodes={checked} />
      </ActionsSheet>
    );
  };
  return (
    <Card>
      <CardContent>
        <DataTable
          selectable
          // id is only optional on the type because a node being added has none yet; every listed node comes from the database and has one
          getRowId={(host) => String(host.id)}
          columns={columns}
          data={data ?? []}
          add={"/add/node"}
          Actions={actions}
          initialVisibility={{
            provision: false,
            arch: false,
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
