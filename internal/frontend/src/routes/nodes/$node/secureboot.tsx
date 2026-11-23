import { ErrorPanel } from "@/components/error-panel";
import KeysPanel from "@/components/secureboot/keys-panel";
import SecureBootState from "@/components/secureboot/state";
import { secureBootOptions } from "@/lib/queries";
import { useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { ShieldQuestion } from "lucide-react";

export const Route = createFileRoute("/nodes/$node/secureboot")({
  staticData: { crumb: "Secure boot" },
  component: RouteComponent,
  // a BMC is only logged into once the tab is opened, not when its link is hovered
  preload: false,
  loader: ({ context: { queryClient }, params: { node } }) =>
    queryClient.ensureQueryData(secureBootOptions(node)),
});

function RouteComponent() {
  const { node } = Route.useParams();
  const { data } = useSuspenseQuery(secureBootOptions(node));
  const result = data.results?.[0];

  if (!result?.success || !result.data)
    return (
      <ErrorPanel
        icon={ShieldQuestion}
        title="Could not read secure boot"
        message={result?.message}
      />
    );

  return (
    <div className="grid gap-6">
      <SecureBootState node={node} status={result.data} />
      <KeysPanel nodes={node} />
    </div>
  );
}
