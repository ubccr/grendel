import { getV1BmcBios } from "@/client";
import BiosEditor from "@/components/bios/editor";
import { useFailureToast } from "@/components/bios/use-failure-toast";
import { useBiosRegistry } from "@/components/bios/use-registry";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Skeleton } from "@/components/ui/skeleton";
import { asValues } from "@/lib/bios";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { ServerOff } from "lucide-react";
import { useEffect } from "react";
import { toast } from "sonner";

export const Route = createFileRoute("/nodes/$node/bios")({
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
  useFailureToast(res?.run_id, res?.results ?? []);

  useEffect(() => {
    if (error)
      toast.warning("Could not read the attribute registry", { description: error.message });
  }, [error]);

  if (!bios) {
    return (
      <Empty>
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <ServerOff />
          </EmptyMedia>
          <EmptyTitle>No BIOS settings to show</EmptyTitle>
          <EmptyDescription>{result?.message || "The BMC could not be read."}</EmptyDescription>
        </EmptyHeader>
      </Empty>
    );
  }
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
