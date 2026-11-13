import ActionCard from "@/components/actions/action-card";
import { Button } from "@/components/ui/button";
import { Link } from "@tanstack/react-router";

export default function NodesBiosAction({ nodes }: { nodes: string }) {
  return (
    <ActionCard
      title="BIOS Settings"
      description="Compare BIOS settings across the nodes, and stage the same changes on all of them."
      footer={
        <Button render={<Link to="/nodes/bios/$nodes" params={{ nodes }} />} nativeButton={false}>
          Open
        </Button>
      }
    />
  );
}
