import { getV1SwitchScan } from "@/client";
import type { DataTableColumnDef } from "@/components/data-table/features";
import { DataTableColumnHeader } from "@/components/data-table/header";
import { NodeLink } from "@/components/nodes/node-link";
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
import type { ScanSwitch } from "@/lib/import-rows";
import { switchNos, type NeighborKind } from "@/lib/switch-ports";
import { createFileRoute, useLoaderData } from "@tanstack/react-router";
import { TriangleAlert, Unplug } from "lucide-react";
import { useMemo } from "react";
import { z } from "zod";

type Problem = { type?: string; title?: string; detail?: string };

export const Route = createFileRoute("/nodes/$node/macs")({
  staticData: { crumb: "MAC table" },
  component: RouteComponent,
  pendingComponent: Pending,
  // port_name is also the table's Port filter, so picking a cage filters the table to it
  validateSearch: z.object({ port_name: z.string().optional() }),
  loader: async ({ params: { node } }) => {
    try {
      const res = await getV1SwitchScan({ query: { nodeset: node } });
      return { sw: res.data?.switches?.[0], problem: undefined };
    } catch (e) {
      return { sw: undefined, problem: e as Problem };
    }
  },
});

// Entry is one MAC the switch has learned, with the port it was learned on and the node holding it, if any
type Entry = {
  id: string;
  port_name: string;
  description?: string;
  mac: string;
  vlan: string;
  owner?: string;
};

const kindOf = (e: Entry): NeighborKind => (e.owner ? "node" : "other");

function entriesOf(sw: ScanSwitch | undefined): Entry[] {
  const out: Entry[] = [];
  for (const port of sw?.ports ?? [])
    for (const m of port.macs ?? [])
      out.push({
        id: `${port.ifname}|${m.vlan}|${m.mac}`,
        port_name: port.ifname ?? "",
        description: port.description,
        mac: m.mac ?? "",
        vlan: m.vlan ?? "",
        owner: m.owner || undefined,
      });
  return out;
}

const columns: DataTableColumnDef<Entry, any>[] = [
  {
    accessorKey: "mac",
    header: ({ column }) => <DataTableColumnHeader column={column} title="MAC" />,
    cell: ({ row }) => <span className="font-mono text-sm">{row.original.mac}</span>,
  },
  {
    accessorKey: "vlan",
    sortFn: "alphanumeric",
    header: ({ column }) => <DataTableColumnHeader column={column} title="VLAN" />,
    cell: ({ row }) => <span className="font-mono text-sm">{row.original.vlan}</span>,
  },
  {
    accessorKey: "owner",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Node" />,
    // a MAC no node holds leaves the cell blank, since the legend above already counts those
    cell: ({ row }) =>
      row.original.owner && (
        <NodeLink node={row.original.owner} className="truncate font-mono text-sm font-medium" />
      ),
  },
  {
    accessorKey: "description",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Port description" />,
    cell: ({ row }) => (
      <span className="block max-w-xs truncate" title={row.original.description}>
        {row.original.description}
      </span>
    ),
  },
];

const s = (n: number) => (n === 1 ? "" : "s");

function Pending() {
  return (
    <div className="grid gap-4">
      <Skeleton className="h-[92px] w-full" />
      <p className="text-sm text-muted-foreground">Reading the switch's MAC table…</p>
    </div>
  );
}

function RouteComponent() {
  const { sw, problem } = Route.useLoaderData();
  const { data: hosts } = useLoaderData({ from: "/nodes/$node" });
  const host = hosts?.[0];
  const entries = useMemo(() => entriesOf(sw), [sw]);

  if (problem || sw?.error)
    return (
      <SwitchProblem
        what="the MAC table"
        code={problem?.type ?? sw?.error_code}
        title={problem?.title}
        detail={problem?.detail ?? sw?.error}
        address={host?.interfaces?.[0]?.ip?.split("/")[0]}
      />
    );
  if (entries.length === 0)
    return (
      <Empty>
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <Unplug />
          </EmptyMedia>
          <EmptyTitle>No MACs learned</EmptyTitle>
          <EmptyDescription>
            The switch answered, but hasn't learned a MAC on any of its ports. A port learns one
            once the machine plugged into it sends traffic, such as a BMC asking for an address.
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    );

  return (
    <SwitchPortView
      nos={switchNos(host?.tags)}
      rows={entries}
      kindOf={kindOf}
      legend={(kind, n) =>
        kind === "node" ? `MAC${s(n)} on nodes in Grendel` : `MAC${s(n)} not in Grendel`
      }
      cageLabel={(number, rows) =>
        rows.length
          ? `Port ${number}, ${rows.length} MAC${s(rows.length)}, ${rows.filter((r) => !r.owner).length} not in Grendel`
          : `Port ${number}, no MAC learned`
      }
      emptyLegend={(n) => `port${s(n)} with no MAC learned`}
      columns={columns}
      getRowId={(e) => e.id}
      initialVisibility={{ description: false }}
      notice={
        (sw?.warnings?.length ?? 0) > 0 && (
          <p className="flex items-start gap-1.5 text-xs text-amber-700 dark:text-amber-300">
            <TriangleAlert className="mt-px size-3.5 shrink-0" />
            <span>
              Ports with no MAC learned may be missing and port descriptions blank, since the
              switch's interface list couldn't be read: {sw!.warnings!.join("; ")}
            </span>
          </p>
        )
      }
    />
  );
}
