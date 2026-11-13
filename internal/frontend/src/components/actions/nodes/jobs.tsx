import { JobsClearDialog } from "@/components/actions/jobs/delete";
import ActionCard from "@/components/actions/action-card";
import { Button } from "@/components/ui/button";
import { Link } from "@tanstack/react-router";

export default function NodesJobsAction({ nodes }: { nodes: string }) {
  return (
    <ActionCard
      title="Job Queue"
      description="View the BMC's job queue, or delete every job in it."
      footer={
        <>
          <JobsClearDialog allNodes={nodes} />
          <Button
            render={<Link to="/nodes/jobs/$nodes" params={{ nodes: nodes }} />}
            nativeButton={false}
          >
            Open
          </Button>
        </>
      }
    />
  );
}
