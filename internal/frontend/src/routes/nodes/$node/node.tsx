import NodeForm from "@/components/nodes/form";
import { createFileRoute, useLoaderData } from "@tanstack/react-router";

export const Route = createFileRoute("/nodes/$node/node")({
  component: RouteComponent,
});

function RouteComponent() {
  const { data } = useLoaderData({ from: "/nodes/$node" });

  return <NodeForm data={data?.[0]} />;
}
