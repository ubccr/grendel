import { type Lldp } from "@/client";
import { getV1NodesFindOptions } from "@/client/@tanstack/react-query.gen";
import { NodeLink } from "@/components/nodes/node-link";
import type { DataTableColumnDef } from "@/components/data-table/features";
import { DataTableColumnHeader } from "@/components/data-table/header";
import { Unset } from "@/components/detail-header";
import { kindLabel } from "@/components/switches/front-panel";
import { SwitchPortView } from "@/components/switches/port-view";
import { SwitchProblem } from "@/components/switches/switch-problem";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Skeleton } from "@/components/ui/skeleton";
import { lldpOptions, nodeOptions } from "@/lib/queries";
import { neighborKind, shortName, switchNos } from "@/lib/switch-ports";
import { useQuery, useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { Unplug } from "lucide-react";
import { useMemo } from "react";
import { z } from "zod";

export const Route = createFileRoute("/nodes/$node/lldp")({
  staticData: { crumb: "LLDP" },
  component: RouteComponent,
  pendingComponent: Pending,
  // port_name is also the table's Port filter, so picking a cage filters the table to it and clearing the filter lets the cage go
  validateSearch: z.object({ port_name: z.string().optional() }),
  // the switch is only asked once the tab is opened, not when its link is hovered
  preload: false,
  // waited on so the tab opens with the table, but a switch that can't be reached is a problem the tab explains, not a failed page
  loader: async ({ context: { queryClient }, params: { node } }) => {
    await queryClient.prefetchQuery(lldpOptions(node));
    const neighbors = neighborNodeset(queryClient.getQueryData(lldpOptions(node).queryKey) ?? []);
    if (neighbors) await queryClient.prefetchQuery(knownNodesOptions(neighbors));
  },
});

// neighborNodeset is the neighbors' names as a nodeset to look up in Grendel. Names a nodeset can't hold are left out
function neighborNodeset(rows: Lldp[]) {
  return [...new Set(rows.map((r) => shortName(r.system_name)))]
    .filter((n) => /^[A-Za-z0-9_-]+$/.test(n))
    .join(",");
}

const knownNodesOptions = (nodeset: string) => getV1NodesFindOptions({ query: { nodeset } });

const noRows: Lldp[] = [];

function Pending() {
  return (
    <div className="grid gap-4">
      <Skeleton className="h-[92px] w-full" />
      <p className="text-sm text-muted-foreground">Asking the switch for its LLDP neighbors…</p>
    </div>
  );
}

function RouteComponent() {
  const { node } = Route.useParams();
  const { data: rows = noRows, error: problem, isPending } = useQuery(lldpOptions(node));
  const { data: hosts } = useSuspenseQuery(nodeOptions(node));
  const host = hosts[0];
  const neighbors = neighborNodeset(rows);
  // which neighbors are nodes in Grendel. A failed lookup only costs the links, not the page
  const { data: found } = useQuery({ ...knownNodesOptions(neighbors), enabled: !!neighbors });
  const known = useMemo(() => new Set(found?.map((h) => h.name ?? "")), [found]);
  const kindOf = (row: Lldp) => neighborKind(row, known);

  if (isPending) return <Pending />;
  if (problem)
    return (
      <SwitchProblem
        what="LLDP neighbors"
        code={problem.type}
        title={problem.title}
        detail={problem.detail}
        address={host?.interfaces?.[0]?.ip?.split("/")[0]}
      />
    );
  if (rows.length === 0)
    return (
      <Empty>
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <Unplug />
          </EmptyMedia>
          <EmptyTitle>No LLDP neighbors</EmptyTitle>
          <EmptyDescription>
            The switch answered, but none of its ports hear a neighbor. Hosts only show up here
            while they run an LLDP agent such as lldpd.
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    );

  const columns: DataTableColumnDef<Lldp, any>[] = [
    {
      accessorKey: "system_name",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Neighbor" />,
      cell: ({ row }) => <Neighbor row={row.original} known={kindOf(row.original) === "node"} />,
    },
    ...otherColumns,
  ];

  return (
    <SwitchPortView
      nos={switchNos(host?.tags)}
      rows={rows}
      kindOf={kindOf}
      legend={(kind, n) =>
        kind === "node" ? (n === 1 ? "node in Grendel" : "nodes in Grendel") : "not in Grendel"
      }
      cageLabel={(number, here) =>
        here.length
          ? `Port ${number}, ${here.map((n) => `${n.system_name || n.chassis_id || "unnamed"} (${kindLabel[kindOf(n)].toLowerCase()})`).join(", ")}`
          : `Port ${number}, no LLDP neighbor`
      }
      emptyLegend={(n) => (n === 1 ? "port without a neighbor" : "ports without a neighbor")}
      columns={columns}
      getRowId={(row, i) => row.port_name ?? String(i)}
      initialVisibility={{ system_description: false }}
    />
  );
}

const otherColumns: DataTableColumnDef<Lldp, any>[] = [
  {
    accessorKey: "port_id",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Its port" />,
    cell: ({ row }) => <TheirPort row={row.original} />,
  },
  {
    accessorKey: "management_address",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Address" />,
    cell: ({ row }) =>
      row.original.management_address ? (
        <span className="font-mono text-sm">{row.original.management_address}</span>
      ) : (
        <Unset>none sent</Unset>
      ),
  },
  {
    accessorKey: "chassis_id",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Chassis ID" />,
    cell: ({ row }) => <Chassis row={row.original} />,
  },
  {
    accessorKey: "system_description",
    header: ({ column }) => <DataTableColumnHeader column={column} title="System description" />,
    cell: ({ row }) => (
      <span className="block max-w-md truncate" title={row.original.system_description}>
        {row.original.system_description}
      </span>
    ),
  },
];

function Neighbor({ row, known }: { row: Lldp; known: boolean }) {
  const name = row.system_name;
  return (
    <span className="inline-flex min-w-0">
      {known ? (
        <NodeLink node={shortName(name)} className="truncate font-mono text-sm font-medium">
          {name}
        </NodeLink>
      ) : name ? (
        <span className="truncate font-mono text-sm">{name}</span>
      ) : (
        <Unset>no name sent</Unset>
      )}
    </span>
  );
}

function TheirPort({ row }: { row: Lldp }) {
  const desc =
    row.port_description && row.port_description !== row.port_id ? row.port_description : undefined;
  if (!row.port_id && !desc) return <Unset>none sent</Unset>;
  return (
    <span className="inline-flex min-w-0 items-baseline gap-2">
      {row.port_id && <span className="truncate font-mono text-sm">{row.port_id}</span>}
      {desc && <span className="truncate text-sm text-muted-foreground">{desc}</span>}
    </span>
  );
}

function Chassis({ row }: { row: Lldp }) {
  if (row.chassis_id) return <span className="font-mono text-sm">{row.chassis_id}</span>;
  return (
    <Unset>{row.chassis_id_type ? `unreadable ${row.chassis_id_type} ID` : "none sent"}</Unset>
  );
}
