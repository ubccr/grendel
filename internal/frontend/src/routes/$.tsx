import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/$")({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <div className="flex justify-center p-4">
      <Card className="w-56">
        <CardHeader>
          <CardTitle>404: Not Found</CardTitle>
          <Separator />
          <CardDescription>
            How did you end up here? Who sent you!?!
          </CardDescription>
        </CardHeader>
      </Card>
    </div>
  );
}
