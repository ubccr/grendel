import { ProvisionStatusPanel } from "@/components/nodes/provision-status";
import { nodeOptions } from "@/lib/queries";
import { useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/nodes/$node/provision")({
  staticData: { crumb: "Provision" },
  component: RouteComponent,
});

function RouteComponent() {
  const { data } = useSuspenseQuery(nodeOptions(Route.useParams().node));
  const name = data[0]?.name;

  return name ? <ProvisionStatusPanel node={name} /> : null;
}
