import { createFileRoute } from "@tanstack/react-router";

import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";

export const Route = createFileRoute("/")({
  component: Index,
});

function Index() {
  return (
    <div className="flex justify-center">
      <Card>
        <CardHeader className="text-center">
          <CardTitle className="text-8xl">Grendel</CardTitle>
          <Separator />
          <CardDescription>
            New to Grendel? Checkout our{" "}
            <a
              className="text-primary"
              href="https://grendel.readthedocs.io/en/latest/"
              target="_blank"
            >
              Docs
            </a>
          </CardDescription>
        </CardHeader>
      </Card>
    </div>
  );
}
