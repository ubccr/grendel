import type { Host } from "@/client";
import { getV1NodesQueryKey } from "@/client/@tanstack/react-query.gen";
import { Unset } from "@/components/detail-header";
import { bmcNic, bootNic, LastProvision, ProvisionLight } from "@/components/nodes/plate";
import { Badge } from "@/components/ui/badge";
import { HoverCard, HoverCardContent, HoverCardTrigger } from "@/components/ui/hover-card";
import { Skeleton } from "@/components/ui/skeleton";
import { nodeOptions, provisionStatusOptions } from "@/lib/queries";
import { cn } from "@/lib/utils";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Link, type LinkProps } from "@tanstack/react-router";
import type { ReactNode } from "react";

const TAGS_SHOWN = 8;

type NodeTab = Extract<LinkProps["to"], `/nodes/$node/${string}`>;

export function NodeLink({
  node,
  to = "/nodes/$node/node",
  search,
  className,
  children,
}: {
  node: string;
  to?: NodeTab;
  // search is handed to the node view as is, such as the rack a node was opened from
  search?: Record<string, unknown>;
  className?: string;
  children?: ReactNode;
}) {
  return (
    <HoverCard>
      <HoverCardTrigger
        delay={400}
        closeDelay={100}
        render={
          <Link
            to={to}
            params={{ node }}
            search={search as never}
            className={cn(
              "rounded-sm underline-offset-4 outline-none hover:underline focus-visible:ring-2 focus-visible:ring-ring",
              className,
            )}
          />
        }
      >
        {children ?? node}
      </HoverCardTrigger>
      {/* the popup only mounts while open, so the summary's lookups wait until someone asks for it */}
      <HoverCardContent align="start" className="w-80 p-0">
        <NodeSummary node={node} />
      </HoverCardContent>
    </HoverCard>
  );
}

export function NodeSummary({ node }: { node: string }) {
  const queryClient = useQueryClient();
  const { data, isPending, isError } = useQuery({
    // the same read as the node's own page, which the link asks for ahead of a click, so the two are one request
    ...nodeOptions(node),
    // a page that already loaded every node shows the summary at once, and the lookup only refreshes it
    placeholderData: () => {
      const h = queryClient
        .getQueryData<Host[]>(getV1NodesQueryKey())
        ?.find((x) => x.name === node);
      return h ? [h] : undefined;
    },
  });
  const host = data?.[0];
  const isSwitch = !!host?.tags?.includes("switch");
  const { data: statuses } = useQuery({
    ...provisionStatusOptions({ nodeset: node }),
    enabled: !!host && !isSwitch,
  });
  const status = statuses?.[0];

  if (isPending && !host)
    return (
      <div className="grid gap-3 p-3">
        <Skeleton className="h-5 w-40" />
        <Skeleton className="h-16 w-full" />
      </div>
    );
  if (isError || !host)
    return <p className="p-3 text-sm text-muted-foreground">No node named {node} in Grendel.</p>;

  const boot = bootNic(host.interfaces);
  const bmc = bmcNic(host.interfaces);
  const tags = [...(host.tags ?? [])].sort();

  return (
    <div className="grid text-sm">
      <div className="grid gap-0.5 px-3 pt-3 pb-2.5">
        <p className="flex items-baseline justify-between gap-3">
          <span className="truncate font-mono text-base font-semibold tracking-tight">
            {host.name}
          </span>
          {!isSwitch && (
            <span className="inline-flex shrink-0 items-center gap-1.5 text-xs text-muted-foreground">
              <ProvisionLight on={host.provision} className="size-1.5" />
              Provision {host.provision ? "on" : "off"}
            </span>
          )}
        </p>
        <p className="text-xs text-muted-foreground">
          {isSwitch ? "Switch" : [host.arch, host.boot_image].filter(Boolean).join(", ") || "Node"}
        </p>
      </div>

      <dl className="grid grid-cols-[auto_minmax(0,1fr)] gap-x-4 gap-y-1.5 border-t px-3 py-2.5">
        <Row label={isSwitch ? "Management" : "Boot"}>
          {boot?.mac || boot?.ip ? (
            <span className="grid font-mono text-xs leading-snug">
              {boot.mac && <span className="truncate">{boot.mac}</span>}
              {boot.ip && <span className="truncate">{boot.ip}</span>}
            </span>
          ) : (
            <Unset>none</Unset>
          )}
        </Row>
        {!isSwitch && (
          <Row label="BMC">
            {bmc?.fqdn || bmc?.ip ? (
              <span className="grid font-mono text-xs leading-snug">
                {bmc.fqdn && <span className="truncate">{bmc.fqdn}</span>}
                {bmc.ip && <span className="truncate">{bmc.ip}</span>}
              </span>
            ) : (
              <Unset>none</Unset>
            )}
          </Row>
        )}
        {status?.state && (
          <Row label="Provisioned">
            <span className="text-xs">
              <LastProvision status={status} node={node} />
            </span>
          </Row>
        )}
      </dl>

      {tags.length > 0 && (
        <div className="flex flex-wrap gap-1 border-t px-3 py-2.5">
          {tags.slice(0, TAGS_SHOWN).map((t) => (
            <Badge key={t} variant="secondary" className="rounded-md">
              {t}
            </Badge>
          ))}
          {tags.length > TAGS_SHOWN && (
            <Badge variant="outline" className="rounded-md tabular-nums">
              +{tags.length - TAGS_SHOWN}
            </Badge>
          )}
        </div>
      )}
    </div>
  );
}

function Row({ label, children }: { label: string; children: ReactNode }) {
  return (
    <>
      <dt className="pt-px text-xs text-muted-foreground">{label}</dt>
      <dd className="min-w-0">{children}</dd>
    </>
  );
}
