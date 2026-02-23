import { getV1Nodes } from "@/client";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/animate-ui/primitives/radix/collapsible";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Input } from "@/components/ui/input";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import AuthRedirect from "@/lib/auth";
import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";
import { ChevronsUpDown, Copy, Info, SquareMenu, Zap, ZapOff } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import z from "zod";

const rackSearchSchema = z.object({
  name: z.string().optional(),
  tag: z.string().optional(),
});

export const Route = createFileRoute("/racks/")({
  component: RouteComponent,
  beforeLoad: AuthRedirect,
  validateSearch: rackSearchSchema,
  loader: async () => {
    const res = await getV1Nodes();
    if (!res.data) {
      return res;
    }
    const racks: Map<string, RackData> = new Map();
    res.data.forEach((node) => {
      const rackTag = node.tags?.find((v) => v.includes("rack="));
      if (!rackTag) {
        return;
      }
      const rack = rackTag.split("=")[1];
      const rd: RackData = racks.get(rack) ?? {
        Rack: rack,
        Size: 0,
        Tags: new Map(),
        Provision: 0,
        Unprovision: 0,
        Nodeset: [],
      };
      rd.Size = rd.Size + 1;
      if (node.name) rd.Nodeset.push(node.name);
      if (node.provision) rd.Provision = rd.Provision + 1;
      else rd.Unprovision = rd.Unprovision + 1;

      node.tags?.forEach((tag) => rd.Tags.set(tag, (rd.Tags.get(tag) ?? 0) + 1));

      racks.set(rack, rd);
    });
    return { ...res, data: racks };
  },
});

type RackData = {
  Rack: string;
  Size: number;
  Tags: Map<string, number>;
  Provision: number;
  Unprovision: number;
  Nodeset: Array<string>;
};

function RouteComponent() {
  const { data } = Route.useLoaderData();
  const { name, tag } = Route.useSearch();
  const navigate = useNavigate();

  const [searchName, setSearchName] = useState(name ?? "");
  const [searchTags, setSearchTags] = useState(tag ?? "");

  return (
    <div className="grid gap-y-2">
      <div>
        <Card>
          <CardHeader>
            <CardTitle className="flex justify-between">
              <div>Filters</div>
              <Tooltip>
                <TooltipProvider>
                  <TooltipTrigger asChild>
                    <Button type="button" size="icon" variant="secondary">
                      <Info />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>
                    Tags that are only on a single node per rack will not be displayed, however can
                    still be filtered. <br />
                    Tags with a count above them are only found on that number of nodes in a rack.
                  </TooltipContent>
                </TooltipProvider>
              </Tooltip>
            </CardTitle>
          </CardHeader>
          <CardContent className="grid gap-2 md:grid-cols-2">
            <Input
              value={searchName}
              onChange={(e) => {
                const value = e.currentTarget.value;
                navigate({
                  to: "/racks",
                  search: (prev) => ({ ...prev, name: value }),
                  replace: true,
                });
                setSearchName(value);
              }}
              placeholder="Name"
            />
            <Input
              value={searchTags}
              onChange={(e) => {
                const value = e.currentTarget.value;
                navigate({
                  to: "/racks",
                  search: (prev) => ({ ...prev, tag: value }),
                  replace: true,
                });
                setSearchTags(value);
              }}
              placeholder="Tag"
            />
          </CardContent>
        </Card>
      </div>
      {data && data?.size > 0 ? (
        Array.from(data.entries())
          .sort()
          .filter(([key]) => key.includes(searchName))
          .filter(([, value]) => Array.from(value.Tags).find(([key]) => key.includes(searchTags)))
          .map(([key, value]) => (
            <Card key={key}>
              <CardContent className="p-3">
                <div className="mb-4 flex justify-between gap-6">
                  <div className="whitespace-nowrap">
                    <Link className="text-2xl" to="/racks/$rack" params={{ rack: key }}>
                      Rack {key}
                    </Link>
                    <CardDescription>Node count: {value.Size}</CardDescription>
                  </div>
                  <Collapsible>
                    <div className="flex justify-end gap-2 align-middle">
                      <h4 className="my-auto text-sm font-semibold text-foreground">Nodeset</h4>
                      <CollapsibleTrigger asChild>
                        <Button variant="secondary" size="icon">
                          <ChevronsUpDown />
                          <span className="sr-only">Toggle Nodeset</span>
                        </Button>
                      </CollapsibleTrigger>
                      <Button
                        size="icon"
                        variant="secondary"
                        onClick={() => {
                          navigator.clipboard.writeText(value.Nodeset.sort().join(","));
                          toast.success("Successfully copied item(s)");
                        }}
                      >
                        <Copy />
                      </Button>
                    </div>
                    <CollapsibleContent>
                      <span className="text-xs text-muted-foreground">
                        {value.Nodeset.sort().join(",")}
                      </span>
                    </CollapsibleContent>
                  </Collapsible>
                </div>
                <div className="flex justify-between">
                  <div className="flex flex-wrap items-center gap-2">
                    {Array.from(value.Tags)
                      .sort()
                      .map(([tag, count]) => {
                        if (count < 2 || tag === `rack=${value.Rack}`) return;

                        return (
                          <Link key={tag} to="/nodes" search={{ tags: [`rack=${key}`, tag] }}>
                            <Badge className="relative">
                              {tag}
                              {count !== value.Size && (
                                <span className="absolute -top-1.5 -right-2 flex min-w-4">
                                  <span className="relative inline-flex size-full justify-center rounded-full bg-sky-400 p-0.5 text-[10px] text-black">
                                    {count}
                                  </span>
                                </span>
                              )}
                            </Badge>
                          </Link>
                        );
                      })}
                  </div>
                  <div className="flex gap-2">
                    <Button className="relative" variant="secondary" size="icon" asChild>
                      <Link
                        to="/nodes"
                        search={{
                          provision: "true",
                          tags: [`rack=${key}`],
                        }}
                      >
                        <Zap className="text-green-600" />
                        <span className="absolute -top-1.5 -right-2 flex min-w-4">
                          <span className="relative inline-flex size-full justify-center rounded-full bg-sky-400 p-0.5 text-[10px] text-black">
                            {value.Provision}
                          </span>
                        </span>
                      </Link>
                    </Button>
                    <Button className="relative" variant="secondary" size="icon" asChild>
                      <Link
                        to="/nodes"
                        search={{
                          provision: "false",
                          tags: [`rack=${key}`],
                        }}
                      >
                        <ZapOff className="text-red-600" />
                        <span className="absolute -top-1.5 -right-2 flex min-w-4">
                          <span className="relative inline-flex size-full justify-center rounded-full bg-sky-400 p-0.5 text-[10px] text-black">
                            {value.Unprovision}
                          </span>
                        </span>
                      </Link>
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))
      ) : (
        <Card>
          <CardContent>
            <Empty>
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <SquareMenu />
                </EmptyMedia>
                <EmptyTitle>No Racks Found</EmptyTitle>
                <EmptyDescription>
                  Try tagging nodes in the same rack with "rack=a01"
                </EmptyDescription>
              </EmptyHeader>
            </Empty>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
