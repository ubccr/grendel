import { getV1BmcSecureboot } from "@/client";
import { ErrorPanel } from "@/components/error-panel";
import KeysPanel from "@/components/secureboot/keys-panel";
import SecureBootState from "@/components/secureboot/state";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { ShieldQuestion } from "lucide-react";

export const Route = createFileRoute("/nodes/$node/secureboot")({
  staticData: { crumb: "Secure boot" },
  component: RouteComponent,
  // one request reads the state and the keys over a single login to the BMC
  loader: ({ params: { node } }) => getV1BmcSecureboot({ query: { nodeset: node, keys: true } }),
  staleTime: 30_000,
});

function RouteComponent() {
  const router = useRouter();
  const { node } = Route.useParams();
  const { data } = Route.useLoaderData();
  const result = data?.results?.[0];
  const refresh = () => router.invalidate();

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
      <SecureBootState node={node} status={result.data} onChanged={refresh} />
      <KeysPanel nodes={node} results={data?.results ?? []} onChanged={refresh} />
    </div>
  );
}
