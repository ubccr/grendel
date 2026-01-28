import { Button } from "@/components/ui/button";
import { Card, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Link } from "@tanstack/react-router";

export default function NodesFirmwareAction({ nodes }: { nodes: string }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Dell Firmware Upgrade</CardTitle>
      </CardHeader>
      <CardFooter>
        <Button asChild>
          <Link to="/nodes/dell-firmware/$nodes" params={{ nodes: nodes }}>
            Open
          </Link>
        </Button>
      </CardFooter>
    </Card>
  );
}
