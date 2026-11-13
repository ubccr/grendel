import ActionCard from "@/components/actions/action-card";
import { Button } from "@/components/ui/button";
import { Link } from "@tanstack/react-router";

export default function NodesIdentityAction({ nodes }: { nodes: string }) {
  return (
    <ActionCard
      title="Asset Tag & Location"
      description="Set the rack location from the node names, and asset tags from a template or a CSV of serial numbers."
      footer={
        <Button
          render={<Link to="/nodes/identity/$nodes" params={{ nodes }} />}
          nativeButton={false}
        >
          Open
        </Button>
      }
    />
  );
}
