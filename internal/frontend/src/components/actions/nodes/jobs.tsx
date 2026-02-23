import { Button } from "@/components/ui/button";
import { Card, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Link } from "@tanstack/react-router";

export default function NodesJobsAction({ nodes }: { nodes: string }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>View Jobs</CardTitle>
      </CardHeader>
      <CardFooter className="flex gap-1">
        <Button asChild>
          <Link to="/nodes/jobs/$nodes" params={{ nodes: nodes }}>
            Open
          </Link>
        </Button>
      </CardFooter>
    </Card>
  );
}
