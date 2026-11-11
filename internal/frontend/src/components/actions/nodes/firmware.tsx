import ActionCard from "@/components/actions/action-card";
import { Button } from "@/components/ui/button";
import { Link } from "@tanstack/react-router";

export default function NodesFirmwareAction({ nodes }: { nodes: string }) {
  return (
    <ActionCard
      title="Firmware Updates"
      description="Check the Dell catalog for firmware updates, then install them."
      footer={
        <Button asChild>
          <Link to="/nodes/dell-firmware/$nodes" params={{ nodes: nodes }}>
            Open
          </Link>
        </Button>
      }
    />
  );
}
