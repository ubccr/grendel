import { getV1NodesFind } from "@/client";
import ActionsSheet from "@/components/actions-sheet";
import NodeActions from "@/components/actions/nodes";
import { DetailHeader } from "@/components/detail-header";
import { NodeFacts } from "@/components/nodes/plate";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import AuthRedirect from "@/lib/auth";
import {
  createFileRoute,
  Link,
  notFound,
  Outlet,
  retainSearchParams,
  useRouter,
} from "@tanstack/react-router";
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
  beforeLoad: AuthRedirect,
  loader: async ({ params: { node } }) => {
    const res = await getV1NodesFind({ query: { nodeset: node } });
    if (res.data && res.data.length != 1)
      throw notFound({
        data: { message: `Query '${node}' matched ${res.data.length} node(s)` },
      });
    else return res;
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
] as const;

function RouteComponent() {
  const router = useRouter();
  const nodeName = Route.useParams().node;
  const { data } = Route.useLoaderData();
  const host = data?.[0];
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
              <Button variant="secondary" type="button" onClick={() => router.invalidate()}>
                <RefreshCw />
                <span className="sr-only md:not-sr-only">Refresh</span>
              </Button>
            </>
          }
          nav={
            <nav aria-label="Node views" className="-mb-px flex gap-1 overflow-x-auto">
              {tabs.map((t) => (
                <Link
                  key={t.to}
                  to={t.to}
                  from="/nodes/$node"
                  activeProps={{
                    className: "border-foreground text-foreground",
                    "aria-current": "page",
                  }}
                  inactiveProps={{
                    className:
                      "border-transparent text-muted-foreground hover:text-foreground hover:border-border",
                  }}
                  className="rounded-t-sm border-b-2 px-3 pt-1 pb-2.5 text-sm font-medium whitespace-nowrap transition-colors focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
                >
                  {t.label}
                </Link>
              ))}
            </nav>
          }
        />
        <div className="px-6 py-5">
          <Outlet />
        </div>
      </CardContent>
    </Card>
  );
}
