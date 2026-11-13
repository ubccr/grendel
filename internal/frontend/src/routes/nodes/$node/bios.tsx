import { getV1BmcBios } from "@/client";
import BiosEditor from "@/components/bios/editor";
import { useBiosRegistry } from "@/components/bios/use-registry";
import { ErrorPanel } from "@/components/error-panel";
import { Skeleton } from "@/components/ui/skeleton";
import { asValues } from "@/lib/bios";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { useEffect } from "react";
import { toast } from "sonner";

export const Route = createFileRoute("/nodes/$node/bios")({
  staticData: { crumb: "BIOS" },
  component: RouteComponent,
  loader: ({ params: { node } }) => getV1BmcBios({ query: { nodeset: node } }),
  staleTime: 30_000,
});

function RouteComponent() {
  const router = useRouter();
  const { node } = Route.useParams();
  const { data: res } = Route.useLoaderData();
  const result = res?.results?.[0];
  const bios = result?.success ? result.data : undefined;
  const { registry, isLoading, error } = useBiosRegistry(node, bios);

  useEffect(() => {
    if (error)
      toast.warning("Could not read the attribute registry", { description: error.message });
  }, [error]);

  if (!bios) return <ErrorPanel title="Could not read the BIOS" message={result?.message} />;
  if (isLoading) return <Skeleton className="h-96 w-full" />;

  return (
    <BiosEditor
      // a new reading of the node starts the editor over, so edits are never compared against stale values
      key={JSON.stringify(bios.pending)}
      nodes={node}
      values={asValues(bios.attributes)}
      pending={asValues(bios.pending)}
      registry={registry}
      applyTimes={bios.apply_times ?? []}
      onApplied={() => router.invalidate()}
    />
  );
}
