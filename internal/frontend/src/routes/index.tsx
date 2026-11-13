import { Card, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/")({
  component: Index,
});

function Index() {
  return (
    <div className="flex justify-center">
      {/* the card needs a width of its own, because a Nova CardHeader is a size container and can't take its width from the title */}
      <Card className="w-full max-w-lg">
        <CardHeader className="text-center">
          <CardTitle className="text-8xl">Grendel</CardTitle>
          <Separator />
          <CardDescription>
            New to Grendel? Checkout our{" "}
            <a
              className="text-primary"
              href="https://grendel.readthedocs.io/en/latest/"
              target="_blank"
              rel="noreferrer"
            >
              Docs
            </a>
          </CardDescription>
        </CardHeader>
      </Card>
    </div>
  );
}
