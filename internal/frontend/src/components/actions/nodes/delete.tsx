import { deleteV1NodesMutation } from "@/client/@tanstack/react-query.gen";
import { ConfirmAction } from "@/components/actions/action-card";
import NodesetView from "@/components/nodeset-view";

export default function NodesDeleteAction({ nodes }: { nodes: string }) {
  return (
    <ConfirmAction
      title="Delete"
      description="Remove the nodes from Grendel. Their BMCs are not touched."
      label="Delete"
      destructive
      mutation={deleteV1NodesMutation()}
      variables={{ query: { nodeset: nodes } }}
    >
      These nodes will be permanently removed from Grendel:
      <NodesetView nodes={nodes} />
    </ConfirmAction>
  );
}
