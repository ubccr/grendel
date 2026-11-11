import { deleteV1NodesMutation } from "@/client/@tanstack/react-query.gen";
import ActionCard, { ConfirmDialog } from "@/components/actions/action-card";
import NodesetView from "@/components/nodeset-view";
import { Button } from "@/components/ui/button";
import { useMutation } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import { toast } from "sonner";

export default function NodesDeleteAction({ nodes }: { nodes: string }) {
  const { mutate, isPending } = useMutation(deleteV1NodesMutation());

  return (
    <ActionCard
      title="Delete"
      description="Remove the nodes from Grendel. Their BMCs are not touched."
      footer={
        <ConfirmDialog
          trigger={
            <Button variant="destructive" disabled={isPending}>
              {isPending && <LoaderCircle className="animate-spin" />}
              Delete
            </Button>
          }
          onConfirm={() =>
            mutate(
              { query: { nodeset: nodes } },
              {
                onSuccess: (data) => toast.success(data?.title, { description: data?.detail }),
                onError: (e) => toast.error(e.title, { description: e.detail }),
              },
            )
          }
        >
          These nodes will be permanently removed from Grendel:
          <NodesetView nodes={nodes} />
        </ConfirmDialog>
      }
    />
  );
}
