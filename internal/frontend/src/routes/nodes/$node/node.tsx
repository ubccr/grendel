import NodeForm from "@/components/nodes/form";
import { nodeOptions } from "@/lib/queries";
import { useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/nodes/$node/node")({
  staticData: { crumb: "Node" },
  component: RouteComponent,
});

function RouteComponent() {
  const { data } = useSuspenseQuery(nodeOptions(Route.useParams().node));

  return <NodeForm data={data[0]} />;
}
