import ActionCard from "@/components/actions/action-card";
import { Button } from "@/components/ui/button";
import { Link } from "@tanstack/react-router";

export default function NodesBootNicsAction({ nodes }: { nodes: string }) {
  return (
    <ActionCard
      title="Add interfaces"
      description="Read each node's network interfaces from its BMC, pick one, and add it with an address."
      footer={
        <Button
          render={<Link to="/add/node" search={{ tab: "boot", bootStep: "ports", nodes }} />}
          nativeButton={false}
        >
          Open
        </Button>
      }
    />
  );
}
