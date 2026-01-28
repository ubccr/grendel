import { getV1NodesFind } from "@/client";
import ActionsSheet from "@/components/actions-sheet";
import NodeActions from "@/components/actions/nodes";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import AuthRedirect from "@/lib/auth";
import { createFileRoute, Link, notFound, Outlet, useRouter } from "@tanstack/react-router";
import { RefreshCw } from "lucide-react";
import { useState } from "react";
import { z } from "zod";

export const Route = createFileRoute("/nodes/$node")({
  component: RouteComponent,
  validateSearch: z.object({
    tab: z.string().optional().catch("node"),
  }),
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

type NodeType = "Server" | "Switch" | "PDU";

function RouteComponent() {
  const router = useRouter();
  const nodeName = Route.useParams().node;
  const { data } = Route.useLoaderData();
  const [nodeType, setNodeType] = useState<NodeType>("Server");

  if (nodeType != "Switch" && data?.[0].tags?.includes("switch")) {
    setNodeType("Switch");
  }

  return (
    <Card>
      <CardContent>
        <div className="mb-2 grid grid-cols-2 gap-3 pt-2 sm:grid-cols-3">
          <div className="hidden sm:block"></div>
          <div className="sm:text-center">
            <div className="inline-flex h-9 items-center justify-center rounded-lg bg-muted p-1 text-muted-foreground">
              <Link
                to="/nodes/$node/node"
                from="/nodes/$node"
                activeProps={{
                  className: "shadow bg-background text-foreground",
                }}
                className="inline-flex items-center justify-center rounded-md px-3 py-1 text-sm font-medium whitespace-nowrap ring-offset-background transition-all focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:outline-none disabled:pointer-events-none disabled:opacity-50"
              >
                Node
              </Link>
              {nodeType == "Switch" && (
                <Link
                  to="/nodes/$node/lldp"
                  from="/nodes/$node"
                  activeProps={{
                    className: "shadow bg-background text-foreground",
                  }}
                  className="inline-flex items-center justify-center rounded-md px-3 py-1 text-sm font-medium whitespace-nowrap ring-offset-background transition-all focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:outline-none disabled:pointer-events-none disabled:opacity-50"
                >
                  LLDP
                </Link>
              )}
              {nodeType == "Server" && (
                <Link
                  to="/nodes/$node/redfish"
                  from="/nodes/$node"
                  activeProps={{
                    className: "shadow bg-background text-foreground",
                  }}
                  className="inline-flex items-center justify-center rounded-md px-3 py-1 text-sm font-medium whitespace-nowrap ring-offset-background transition-all focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:outline-none disabled:pointer-events-none disabled:opacity-50"
                >
                  Redfish
                </Link>
              )}
            </div>
          </div>

          <div>
            <div className="flex justify-end gap-2">
              <ActionsSheet checked={nodeName} length={1}>
                <NodeActions nodes={nodeName} />
              </ActionsSheet>
              <Button variant="secondary" type="button" onClick={() => router.invalidate()}>
                <RefreshCw />
                <span className="sr-only md:not-sr-only">Refresh</span>
              </Button>
            </div>
          </div>
        </div>
        <Outlet />
      </CardContent>
    </Card>
  );
}
