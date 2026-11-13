import ActionCard from "@/components/actions/action-card";
import { Button } from "@/components/ui/button";
import { Link } from "@tanstack/react-router";

export default function NodesSecureBootAction({ nodes }: { nodes: string }) {
  return (
    <ActionCard
      title="Secure Boot"
      description="Turn UEFI secure boot on or off, and see which nodes have each key."
      footer={
        <Button
          render={<Link to="/nodes/secureboot/$nodes" params={{ nodes }} />}
          nativeButton={false}
        >
          Open
        </Button>
      }
    />
  );
}
