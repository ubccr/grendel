import { getV1BmcBios } from "@/client";
import BiosEditor from "@/components/bios/editor";
import { useFailureToast } from "@/components/bios/use-failure-toast";
import { useBiosRegistry } from "@/components/bios/use-registry";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import AuthRedirect from "@/lib/auth";
import { asValues, biosDiff, BiosValues, DiffRow } from "@/lib/bios";
import { parseNodeSet } from "@/lib/nodeset";
import { cn } from "@/lib/utils";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { CircleCheck, Clock, Cpu, GitCompare, Search, Server, ServerOff } from "lucide-react";
import { ReactNode, useMemo, useState } from "react";
import { z } from "zod";

export const Route = createFileRoute("/nodes/bios/$nodes")({
  staticData: { crumb: ["BIOS", { param: "nodes" }] },
  component: RouteComponent,
  beforeLoad: AuthRedirect,
  validateSearch: z.object({
    tab: z.enum(["diff", "edit"]).default("diff").catch("diff"),
  }),
  loader: ({ params: { nodes } }) => getV1BmcBios({ query: { nodeset: nodes } }),
});

function RouteComponent() {
  const router = useRouter();
  const { nodes } = Route.useParams();
  const { tab } = Route.useSearch();
  const navigate = Route.useNavigate();
  const { data } = Route.useLoaderData();

  const results = useMemo(() => data?.results ?? [], [data]);
  const reached = results.filter((r) => r.success && r.data);
  const reference = reached[0];
  const { registry, isLoading } = useBiosRegistry(reference?.node, reference?.data);
  useFailureToast(data?.run_id, results);

  // a staged value counts as the node's, since that is what it runs once the change applies
  const diff = useMemo(() => {
    const byNode: Record<string, BiosValues> = {};
    for (const r of results) {
      if (!r.success || !r.data) continue;
      byNode[r.node ?? ""] = { ...asValues(r.data.attributes), ...asValues(r.data.pending) };
    }
    return biosDiff(byNode);
  }, [results]);

  const builds = [
    ...new Set(
      reached.map((r) => [r.data?.model, r.data?.bios_version].filter(Boolean).join(", BIOS ")),
    ),
  ];
  const withPending = reached.filter((r) => Object.keys(r.data?.pending ?? {}).length > 0).length;
  const failed = results.length - reached.length;

  return (
    <Tabs
      value={tab}
      onValueChange={(v) => navigate({ search: { tab: v as "diff" | "edit" } })}
      className="grid gap-2"
    >
      <Card>
        <CardHeader className="flex flex-row flex-wrap items-start justify-between gap-4 space-y-0">
          <div className="min-w-0 space-y-1.5">
            <CardTitle>BIOS Settings</CardTitle>
            <CardDescription className="truncate font-mono">
              {parseNodeSet(nodes).toString()}
            </CardDescription>
          </div>
          {reference && (
            <TabsList>
              <TabsTrigger value="diff">Differences</TabsTrigger>
              <TabsTrigger value="edit">Edit</TabsTrigger>
            </TabsList>
          )}
        </CardHeader>
        <CardContent className="flex flex-wrap gap-2">
          <Stat icon={<Server />} tone={failed ? "warn" : undefined}>
            {reached.length} of {results.length} node(s) read
          </Stat>
          {builds.length > 0 && (
            <Stat
              icon={<Cpu />}
              tone={builds.length > 1 ? "warn" : undefined}
              title={
                builds.length > 1
                  ? `An attribute can mean something different, or be missing, across these: ${builds.join("; ")}`
                  : undefined
              }
            >
              {builds.length > 1 ? `${builds.length} models or BIOS versions` : builds[0]}
            </Stat>
          )}
          {reference && <Stat icon={<GitCompare />}>{diff.length} attribute(s) differ</Stat>}
          {withPending > 0 && (
            <Stat icon={<Clock />} tone="info">
              {withPending} node(s) with staged changes
            </Stat>
          )}
        </CardContent>
      </Card>

      {!reference ? (
        <Card>
          <CardContent className="pt-6">
            <Empty>
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <ServerOff />
                </EmptyMedia>
                <EmptyTitle>No BIOS settings to show</EmptyTitle>
                <EmptyDescription>None of the selected nodes could be read.</EmptyDescription>
              </EmptyHeader>
            </Empty>
          </CardContent>
        </Card>
      ) : (
        <>
          <TabsContent value="diff" className="mt-0">
            <Card>
              <CardContent className="p-4">
                <Differences
                  rows={diff}
                  nodeCount={reached.length}
                  names={new Map((registry ?? []).map((a) => [a.name ?? "", a.display_name ?? ""]))}
                />
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="edit" className="mt-0">
            <Card>
              <CardContent className="p-4">
                {isLoading ? (
                  <Skeleton className="h-96 w-full" />
                ) : (
                  <BiosEditor
                    key={JSON.stringify(reference.data?.pending)}
                    nodes={nodes}
                    values={asValues(reference.data?.attributes)}
                    pending={asValues(reference.data?.pending)}
                    registry={registry}
                    applyTimes={reference.data?.apply_times ?? []}
                    notice={
                      results.length > 1 && (
                        <p className="text-sm text-muted-foreground">
                          Showing the values on{" "}
                          <span className="font-mono text-foreground">{reference.node}</span>.
                          Changes are staged on every node, and a node where a value does not fit is
                          left unchanged and reported.
                        </p>
                      )
                    }
                    onApplied={() => router.invalidate()}
                  />
                )}
              </CardContent>
            </Card>
          </TabsContent>
        </>
      )}
    </Tabs>
  );
}

function Stat({
  icon,
  tone,
  title,
  children,
}: {
  icon: ReactNode;
  tone?: "warn" | "info";
  title?: string;
  children: ReactNode;
}) {
  return (
    <span
      title={title}
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full border px-3 py-1 text-xs [&>svg]:size-3.5 [&>svg]:shrink-0",
        !tone && "bg-muted/50 [&>svg]:text-muted-foreground",
        tone === "warn" && "border-amber-500/40 bg-amber-500/10 text-amber-700 dark:text-amber-300",
        tone === "info" && "border-sky-500/40 bg-sky-500/10 text-sky-700 dark:text-sky-300",
      )}
    >
      {icon}
      {children}
    </span>
  );
}

// Differences lists each attribute that differs, with the most common value first and the nodes that stray from it below
function Differences({
  rows,
  nodeCount,
  names,
}: {
  rows: DiffRow[];
  nodeCount: number;
  names: Map<string, string>;
}) {
  const [search, setSearch] = useState("");

  if (rows.length === 0) {
    return (
      <Empty>
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <CircleCheck />
          </EmptyMedia>
          <EmptyTitle>Every node matches</EmptyTitle>
          <EmptyDescription>
            All {nodeCount} node(s) read have the same BIOS settings, counting staged changes.
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    );
  }

  const query = search.trim().toLowerCase();
  const shown = query
    ? rows.filter((r) => `${r.name} ${names.get(r.name) ?? ""}`.toLowerCase().includes(query))
    : rows;

  return (
    <div className="grid gap-3">
      <div className="relative max-w-sm">
        <Search className="absolute top-2.5 left-2.5 size-4 text-muted-foreground" />
        <Input
          className="pl-8"
          placeholder="Filter attributes"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
      </div>

      <div className="overflow-hidden rounded-lg border">
        <div className="hidden grid-cols-[minmax(12rem,18rem)_minmax(8rem,14rem)_4rem_1fr] gap-4 border-b bg-muted/50 px-4 py-2 text-xs font-medium text-muted-foreground md:grid">
          <span>Attribute</span>
          <span>Value</span>
          <span className="text-right">Nodes</span>
          <span>Nodeset</span>
        </div>
        <div className="divide-y">
          {shown.map((row) => (
            <div
              key={row.name}
              className="grid gap-x-4 gap-y-1 px-4 py-3 md:grid-cols-[minmax(12rem,18rem)_1fr]"
            >
              <div className="min-w-0">
                <div className="truncate font-medium">{names.get(row.name) || row.name}</div>
                {names.get(row.name) && (
                  <div className="truncate font-mono text-xs text-muted-foreground">{row.name}</div>
                )}
              </div>
              <div className="grid gap-1">
                {row.groups.map((g, i) => (
                  <div
                    key={g.value}
                    className="grid grid-cols-[minmax(8rem,14rem)_4rem_1fr] items-center gap-4 text-sm"
                  >
                    <span
                      className={cn(
                        "w-fit max-w-full truncate rounded-md px-2 py-0.5 font-mono text-xs",
                        i === 0 ? "bg-muted" : "bg-amber-500/10 text-amber-700 dark:text-amber-300",
                      )}
                      title={g.value}
                    >
                      {g.value}
                    </span>
                    <span className="text-right text-muted-foreground tabular-nums">
                      {g.nodes.length}
                    </span>
                    <span
                      className="truncate font-mono text-xs text-muted-foreground"
                      title={g.nodeset}
                    >
                      {g.nodeset}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          ))}
          {shown.length === 0 && (
            <div className="p-6 text-center text-sm text-muted-foreground">
              No attributes match.
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
