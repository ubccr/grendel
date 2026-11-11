import { postV1BmcJobsClearMutation } from "@/client/@tanstack/react-query.gen";
import ActionCard, { ConfirmDialog } from "@/components/actions/action-card";
import NodesetView from "@/components/nodeset-view";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { BmcResponse, toastBmcResponse } from "@/lib/bmc";
import { useMutation } from "@tanstack/react-query";
import { useRouter } from "@tanstack/react-router";
import { LoaderCircle } from "lucide-react";
import { toast } from "sonner";

// byJobs groups nodes that have the same jobs selected, since one clear request removes the same ids from every node it names
function byJobs(list: { [node: string]: string[] }) {
  const groups = new Map<string, { nodes: string[]; ids: string[] }>();
  for (const [node, ids] of Object.entries(list)) {
    if (ids.length === 0) continue;
    const sorted = [...ids].sort();
    const key = sorted.join(",");
    const group = groups.get(key) ?? { nodes: [], ids: sorted };
    group.nodes.push(node);
    groups.set(key, group);
  }
  return [...groups.values()];
}

type clearProps = { list?: { [key: string]: string[] }; allNodes?: string };

// JobsClearDialog clears the selected jobs in list, or with allNodes the whole queue on those nodes
export function JobsClearDialog({ list = {}, allNodes }: clearProps) {
  const router = useRouter();
  const { mutateAsync, isPending } = useMutation(postV1BmcJobsClearMutation());
  const hasSelection = Object.values(list).some((ids) => ids.length > 0);

  const clear = async () => {
    let total = 0;
    let succeeded = 0;
    let failed = 0;
    const failedNodes: string[] = [];
    const results: NonNullable<BmcResponse["results"]> = [];

    const requests = allNodes
      ? [{ query: { nodeset: allNodes }, body: { all: true } }]
      : byJobs(list).map((group) => ({
          query: { nodeset: group.nodes.join(",") },
          body: { ids: group.ids },
        }));

    try {
      for (const request of requests) {
        const res = await mutateAsync(request);
        total += res.summary?.total ?? 0;
        succeeded += res.summary?.succeeded ?? 0;
        failed += res.summary?.failed ?? 0;
        if (res.summary?.failed_nodes) failedNodes.push(res.summary.failed_nodes);
        results.push(...(res.results ?? []));
      }
    } catch (e) {
      const err = e as { title?: string; detail?: string };
      toast.error(err.title ?? "Failed to clear jobs", { description: err.detail });
      return;
    } finally {
      router.invalidate();
    }

    toastBmcResponse("Clear jobs", {
      summary: { total, succeeded, failed, failed_nodes: failedNodes.join(",") },
      results,
    });
  };

  return (
    <ConfirmDialog
      trigger={
        <Button variant="secondary" disabled={isPending || (!allNodes && !hasSelection)}>
          {isPending && <LoaderCircle className="animate-spin" />}
          Clear
        </Button>
      }
      onConfirm={clear}
    >
      {allNodes ? (
        <>
          Every job in the queue will be deleted on:
          <NodesetView nodes={allNodes} />
        </>
      ) : (
        <>
          The following nodes will have the listed jobs deleted:
          <ScrollArea className="max-h-[80dvh] overflow-scroll">
            <pre>{JSON.stringify(list, null, 4)}</pre>
          </ScrollArea>
        </>
      )}
    </ConfirmDialog>
  );
}

export default function JobsDeleteAction({ list }: { list: { [key: string]: string[] } }) {
  return (
    <ActionCard
      title="Delete Jobs"
      description="Delete the selected jobs from each node's queue."
      footer={<JobsClearDialog list={list} />}
    />
  );
}
