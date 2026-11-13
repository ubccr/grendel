import { ProvisionStatusPanel } from "@/components/nodes/provision-status";
import { createFileRoute, useLoaderData } from "@tanstack/react-router";

export const Route = createFileRoute("/nodes/$node/provision")({
  staticData: { crumb: "Provision" },
  component: RouteComponent,
});

function RouteComponent() {
  const { data } = useLoaderData({ from: "/nodes/$node" });
  const name = data?.[0]?.name;

  return name ? <ProvisionStatusPanel node={name} /> : null;
}
