import BiosEditor from "@/components/bios/editor";
import { useBiosRegistry } from "@/components/bios/use-registry";
import { ErrorPanel } from "@/components/error-panel";
import { Skeleton } from "@/components/ui/skeleton";
import { asValues } from "@/lib/bios";
import { biosOptions } from "@/lib/queries";
import { useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useEffect } from "react";
import { toast } from "sonner";

export const Route = createFileRoute("/nodes/$node/bios")({
  staticData: { crumb: "BIOS" },
  component: RouteComponent,
  // a BMC is only logged into once the tab is opened, not when its link is hovered
  preload: false,
  loader: ({ context: { queryClient }, params: { node } }) =>
    queryClient.ensureQueryData(biosOptions(node)),
});

function RouteComponent() {
  const { node } = Route.useParams();
  const { data: res } = useSuspenseQuery(biosOptions(node));
  const result = res.results?.[0];
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
    />
  );
}
