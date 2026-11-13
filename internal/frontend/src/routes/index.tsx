import { Card, CardDescription, CardHeader } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import "@fontsource/unifrakturcook/latin-700.css";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/")({
  staticData: { nodeField: "full" },
  component: Index,
});

function Index() {
  return (
    <div className="flex flex-1 items-center justify-center px-2 pb-[8vh]">
      {/* the card needs a width of its own, because a Nova CardHeader is a size container and can't take its width from the title; it floats over the node field like the sidebar and header */}
      <Card className="w-full max-w-lg">
        <CardHeader className="gap-4 px-8 py-4 text-center">
          <h1 className="font-[UnifrakturCook] text-7xl leading-none text-foreground sm:text-8xl">
            Grendel
          </h1>
          <Separator />
          <CardDescription>
            New to Grendel?{" "}
            <a
              className="font-medium text-foreground underline underline-offset-4 hover:text-primary"
              href="https://grendel.readthedocs.io/en/latest/"
              target="_blank"
              rel="noreferrer"
            >
              Read the docs
            </a>
          </CardDescription>
        </CardHeader>
      </Card>
    </div>
  );
}
