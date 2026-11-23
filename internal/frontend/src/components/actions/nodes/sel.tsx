import { postV1BmcSelClearMutation } from "@/client/@tanstack/react-query.gen";
import ActionCard, { ConfirmDialog } from "@/components/actions/action-card";
import NodesetView from "@/components/nodeset-view";
import { Button } from "@/components/ui/button";
import { useMutation } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { LoaderCircle } from "lucide-react";

export default function NodesSelAction({ nodes }: { nodes: string }) {
  const { mutate, isPending } = useMutation({
    ...postV1BmcSelClearMutation(),
    meta: { bmc: "Clear SEL" },
  });

  return (
    <ActionCard
      title="System Event Log"
      description="View or clear the BMC's system event log."
      footer={
        <>
          <ConfirmDialog
            trigger={
              <Button variant="secondary" disabled={isPending}>
                {isPending && <LoaderCircle className="animate-spin" />}
                Clear
              </Button>
            }
            onConfirm={() => mutate({ query: { nodeset: nodes } })}
          >
            Every entry in the system event log will be deleted on:
            <NodesetView nodes={nodes} />
          </ConfirmDialog>
          <Button
            render={<Link to="/nodes/sel/$nodes" params={{ nodes: nodes }} />}
            nativeButton={false}
          >
            Open
          </Button>
        </>
      }
    />
  );
}
