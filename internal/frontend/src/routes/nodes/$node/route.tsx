import ActionsSheet from "@/components/actions-sheet";
import NodeActions from "@/components/actions/nodes";
import { DetailHeader, DetailTabs } from "@/components/detail-header";
import { NodeFacts } from "@/components/nodes/plate";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { nodeOptions, useRefresh } from "@/lib/queries";
import { useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute, notFound, Outlet, retainSearchParams } from "@tanstack/react-router";
import { RefreshCw } from "lucide-react";
import { z } from "zod";

export const Route = createFileRoute("/nodes/$node")({
  staticData: { crumb: { param: "node" } },
  component: RouteComponent,
  validateSearch: z.object({
    // from is the rack a node was opened from, which the breadcrumbs show in place of the node list; it is kept while moving between the node's tabs
    from: z.string().optional(),
  }),
  search: { middlewares: [retainSearchParams(["from"])] },
  loader: async ({ context: { queryClient }, params: { node } }) => {
    const hosts = await queryClient.ensureQueryData(nodeOptions(node));
    if (hosts.length != 1)
      throw notFound({
        data: { message: `Query '${node}' matched ${hosts.length} node(s)` },
      });
  },
});

const serverTabs = [
  { to: "/nodes/$node/node", label: "Node" },
  { to: "/nodes/$node/redfish", label: "Redfish" },
  { to: "/nodes/$node/bios", label: "BIOS" },
  { to: "/nodes/$node/secureboot", label: "Secure boot" },
  { to: "/nodes/$node/provision", label: "Provision" },
  { to: "/nodes/$node/console", label: "Console" },
] as const;

const switchTabs = [
  { to: "/nodes/$node/node", label: "Node" },
  { to: "/nodes/$node/lldp", label: "LLDP" },
  { to: "/nodes/$node/macs", label: "MAC table" },
] as const;

function RouteComponent() {
  const refresh = useRefresh();
  const nodeName = Route.useParams().node;
  const { data } = useSuspenseQuery(nodeOptions(nodeName));
  const host = data[0];
  const tabs = host?.tags?.includes("switch") ? switchTabs : serverTabs;

  return (
    <Card>
      <CardContent className="p-0">
        <DetailHeader
          title={host?.name ?? nodeName}
          facts={<NodeFacts host={host} name={nodeName} />}
          actions={
            <>
              <ActionsSheet type="node" checked={nodeName} length={1}>
                <NodeActions nodes={nodeName} />
              </ActionsSheet>
              <Button variant="secondary" type="button" onClick={() => void refresh()}>
                <RefreshCw />
                <span className="sr-only md:not-sr-only">Refresh</span>
              </Button>
            </>
          }
          nav={<DetailTabs label="Node views" tabs={tabs} from="/nodes/$node" />}
        />
        <div className="px-6 py-5">
          <Outlet />
        </div>
      </CardContent>
    </Card>
  );
}
