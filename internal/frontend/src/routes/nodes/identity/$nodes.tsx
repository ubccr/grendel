import { getV1BmcIdentityDefaultsOptions } from "@/client/@tanstack/react-query.gen";
import AssetTagForm from "@/components/identity/asset-tag-form";
import LocationForm from "@/components/identity/location-form";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import AuthRedirect from "@/lib/auth";
import { expand, parseNodeSet } from "@/lib/nodeset";
import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useMemo } from "react";
import { z } from "zod";

export const Route = createFileRoute("/nodes/identity/$nodes")({
  component: RouteComponent,
  beforeLoad: AuthRedirect,
  validateSearch: z.object({
    tab: z.enum(["location", "asset-tag"]).default("location").catch("location"),
  }),
});

function RouteComponent() {
  const { nodes } = Route.useParams();
  const { tab } = Route.useSearch();
  const navigate = Route.useNavigate();
  const { data: defaults, isLoading } = useQuery(getV1BmcIdentityDefaultsOptions());

  const { folded, example } = useMemo(() => {
    try {
      return { folded: parseNodeSet(nodes).toString(), example: expand(nodes)[0] };
    } catch {
      return { folded: nodes, example: undefined };
    }
  }, [nodes]);

  return (
    <Tabs
      value={tab}
      onValueChange={(v) => navigate({ search: { tab: v as "location" | "asset-tag" } })}
      className="grid gap-2"
    >
      <Card>
        <CardHeader className="flex flex-row flex-wrap items-start justify-between gap-4 space-y-0">
          <div className="min-w-0 space-y-1.5">
            <CardTitle>Asset Tag & Location</CardTitle>
            <CardDescription className="truncate font-mono">{folded}</CardDescription>
          </div>
          <TabsList>
            <TabsTrigger value="location">Location</TabsTrigger>
            <TabsTrigger value="asset-tag">Asset tag</TabsTrigger>
          </TabsList>
        </CardHeader>
      </Card>

      <Card>
        <CardContent className="p-4">
          {isLoading ? (
            <Skeleton className="h-40 w-full" />
          ) : (
            <>
              <TabsContent value="location" className="mt-0">
                <LocationForm nodes={nodes} example={example} defaults={defaults} />
              </TabsContent>
              <TabsContent value="asset-tag" className="mt-0">
                <AssetTagForm
                  nodes={nodes}
                  example={example}
                  defaultTemplate={defaults?.asset_tag}
                />
              </TabsContent>
            </>
          )}
        </CardContent>
      </Card>
    </Tabs>
  );
}
