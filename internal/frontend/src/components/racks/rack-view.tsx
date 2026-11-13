import type { Host, ProvisionStatus } from "@/client";
import { patchV1NodesProvisionMutation } from "@/client/@tanstack/react-query.gen";
import { bmcNic } from "@/components/nodes/plate";
import { ProvisionStateBadge, ProvisionSummary } from "@/components/nodes/provision-status";
import { type Placement, rackRows } from "@/components/racks/elevation";
import { Badge } from "@/components/ui/badge";
import { Checkbox } from "@/components/ui/checkbox";
import { HoverCard, HoverCardContent, HoverCardTrigger } from "@/components/ui/hover-card";
import { cn } from "@/lib/utils";
import { useMutation } from "@tanstack/react-query";
import { Link, useRouter } from "@tanstack/react-router";
import { ExternalLink } from "lucide-react";
import { type MouseEvent, useState } from "react";
import { toast } from "sonner";

export type Detail = "status" | "image" | "arch" | "firmware" | "tags" | "bmc";

type Shared = {
  rack: string;
  // common are the tags every node in the rack has, shown once in the header instead of on each faceplate
  common: Set<string>;
  statuses: Map<string, ProvisionStatus>;
  details: Detail[];
  selected: Set<string>;
  onSelect: (name: string, on: boolean, e: MouseEvent) => void;
};

// Rail is one upright of the rack: the U number stamped beside three cage-nut holes, as on a real rail
function Rail({
  label,
  side,
  holes = true,
}: {
  label: string;
  side: "left" | "right";
  holes?: boolean;
}) {
  return (
    <div
      aria-hidden
      className={cn(
        "flex items-center gap-1.5 border-t border-zinc-300/80 bg-zinc-200 px-1.5 font-mono text-[11px] text-zinc-500 tabular-nums dark:border-zinc-700/60 dark:bg-zinc-800 dark:text-zinc-400",
        side === "left" ? "justify-end" : "flex-row-reverse justify-end max-sm:hidden",
      )}
    >
      <span className="whitespace-nowrap">{label}</span>
      {holes && (
        <span className="grid gap-[3px]">
          {[0, 1, 2].map((i) => (
            <span
              key={i}
              className="size-[5px] rounded-[1px] bg-zinc-400/70 shadow-[inset_0_1px_1px_rgba(0,0,0,0.35)] dark:bg-zinc-950"
            />
          ))}
        </span>
      )}
    </div>
  );
}

// ProvisionLed shows whether a node is set to provision and turns it on or off when pressed
function ProvisionLed({ host }: { host: Host }) {
  const router = useRouter();
  const { mutateAsync } = useMutation(patchV1NodesProvisionMutation());
  // busy lasts until the page has reloaded, so the light doesn't flash back to the old state between the change and the reload
  const [busy, setBusy] = useState(false);
  const on = !!host.provision;
  const toggle = async () => {
    setBusy(true);
    try {
      await mutateAsync({ query: { nodeset: host.name ?? "" }, body: { provision: !on } });
      await router.invalidate();
    } catch (e) {
      const err = e as { title?: string; detail?: string };
      toast.error(err.title ?? "Could not change provision", { description: err.detail });
    } finally {
      setBusy(false);
    }
  };
  return (
    <button
      type="button"
      aria-pressed={on}
      aria-label={`Provision ${host.name}`}
      title={
        on ? "Provision is on. Press to turn it off." : "Provision is off. Press to turn it on."
      }
      disabled={busy}
      onClick={toggle}
      className="grid size-7 shrink-0 place-items-center rounded-md hover:bg-foreground/5 focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
    >
      <span
        className={cn(
          "size-2.5 rounded-full transition-colors",
          on
            ? "bg-emerald-500 shadow-[0_0_6px_1px] shadow-emerald-500/70"
            : "bg-zinc-300 ring-1 ring-zinc-400/60 ring-inset dark:bg-zinc-700",
          busy && "animate-pulse",
        )}
      />
    </button>
  );
}

// StatusCard is a node's provision state badge. Hovering or focusing it shows a summary of its last provision, and clicking it opens the node's provision tab
function StatusCard({ status, node }: { status?: ProvisionStatus; node: string }) {
  if (!status?.state) return null;
  return (
    <HoverCard>
      <HoverCardTrigger
        delay={150}
        closeDelay={100}
        render={
          <Link
            to="/nodes/$node/provision"
            params={{ node }}
            aria-label={`Last provision of ${node}: ${status.state}`}
            className="inline-flex rounded-md focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
          />
        }
      >
        <ProvisionStateBadge state={status.state} />
      </HoverCardTrigger>
      <HoverCardContent align="start" className="w-80">
        <ProvisionSummary status={status} node={node} />
      </HoverCardContent>
    </HoverCard>
  );
}

// Faceplate is one node's front in the rack: selection, provision light and last provision on the left, the name in the middle like an asset label, and the chosen details on the right
function Faceplate({
  host,
  rack,
  common,
  statuses,
  details,
  selected,
  onSelect,
}: Shared & { host: Host }) {
  const name = host.name ?? "";
  const isSelected = selected.has(name);
  const bmc = bmcNic(host.interfaces);
  const bmcHost = bmc?.fqdn || bmc?.ip?.split("/")[0];
  const tags = (host.tags ?? []).filter((t) => t !== `rack=${rack}` && !common.has(t)).sort();

  return (
    <div
      className={cn(
        "flex min-w-0 flex-1 flex-wrap items-center gap-x-3 gap-y-1.5 px-2 py-1.5 sm:grid sm:grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)]",
        isSelected && "bg-sky-500/10",
      )}
    >
      <div className="flex min-w-0 flex-wrap items-center gap-x-1.5 gap-y-1">
        <Checkbox
          aria-label={`Select ${name}`}
          checked={isSelected}
          onClick={(e) => {
            e.preventDefault();
            onSelect(name, !isSelected, e);
          }}
        />
        <ProvisionLed host={host} />
        {details.includes("status") && <StatusCard status={statuses.get(name)} node={name} />}
        {details.includes("image") && host.boot_image && (
          <Link
            to="/images/$image"
            params={{ image: host.boot_image }}
            title="Boot image"
            className="ml-1.5 truncate text-xs text-muted-foreground hover:text-foreground"
          >
            {host.boot_image}
          </Link>
        )}
        {details.includes("arch") && host.arch && (
          <span title="Arch" className="ml-1.5 font-mono text-xs text-muted-foreground">
            {host.arch}
          </span>
        )}
        {details.includes("firmware") && host.firmware && (
          <span title="Firmware" className="ml-1.5 font-mono text-xs text-muted-foreground">
            {host.firmware}
          </span>
        )}
      </div>
      <Link
        to="/nodes/$node/node"
        params={{ node: name }}
        className="truncate text-center font-mono text-sm font-medium underline-offset-4 hover:underline"
      >
        {name}
      </Link>
      <div className="ml-auto flex min-w-0 flex-wrap items-center justify-end gap-x-3 gap-y-1">
        {details.includes("tags") && tags.length > 0 && (
          <TagList
            tags={tags}
            max={3}
            label={`Tags on ${name}`}
            note={
              common.size > 0
                ? `Plus the ${common.size === 1 ? "tag" : `${common.size} tags`} every node in the rack has.`
                : undefined
            }
          />
        )}
        {details.includes("bmc") && bmcHost && (
          <a
            href={`https://${bmcHost}`}
            target="_blank"
            rel="noreferrer"
            title={`https://${bmcHost}`}
            aria-label={`Open the BMC of ${name}`}
            className="grid size-7 place-items-center rounded-md text-muted-foreground hover:bg-foreground/5 hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
          >
            <ExternalLink className="size-3.5" />
          </a>
        )}
      </div>
    </div>
  );
}

// TagList shows the first max tags and folds the rest into a +N chip, which opens every tag in a hover card when hovered or focused, so a node with many tags keeps to its row
export function TagList({
  tags,
  max,
  label,
  note,
  className,
}: {
  tags: string[];
  max: number;
  label: string;
  note?: string;
  className?: string;
}) {
  const shown = tags.length > max ? tags.slice(0, max - 1) : tags;
  const hidden = tags.length - shown.length;
  const badge = "rounded-sm px-1.5 py-0 font-normal";
  return (
    <div className={cn("flex min-w-0 flex-wrap items-center justify-end gap-1", className)}>
      {shown.map((t) => (
        <Badge key={t} variant="secondary" className={badge}>
          {t}
        </Badge>
      ))}
      {hidden > 0 && (
        <HoverCard>
          <HoverCardTrigger
            delay={150}
            closeDelay={100}
            render={
              <button
                type="button"
                aria-label={`${label}: ${hidden} more`}
                className="rounded-sm focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
              />
            }
          >
            <Badge variant="outline" className={cn(badge, "tabular-nums")}>
              +{hidden}
            </Badge>
          </HoverCardTrigger>
          <HoverCardContent align="end" className="grid w-80 gap-2">
            <p className="text-sm font-medium">
              {label} ({tags.length})
            </p>
            <div className="flex flex-wrap gap-1">
              {tags.map((t) => (
                <Badge key={t} variant="secondary" className={badge}>
                  {t}
                </Badge>
              ))}
            </div>
            {note && <p className="text-xs text-muted-foreground">{note}</p>}
          </HoverCardContent>
        </HoverCard>
      )}
    </div>
  );
}

// Ear is the mounting flange at either end of a chassis, with its thumbscrew
function Ear() {
  return (
    <span
      aria-hidden
      className="grid w-2.5 shrink-0 place-items-center self-stretch bg-zinc-300/70 dark:bg-zinc-700/70"
    >
      <span className="size-1.5 rounded-full bg-zinc-400 dark:bg-zinc-500" />
    </span>
  );
}

const row = "grid grid-cols-[2.75rem_1fr] sm:grid-cols-[3.25rem_1fr_3.25rem]";

// RackElevation draws the rack from its top unit down, each node in the U its name gives, sleds sharing a U side by side
export function RackElevation(props: Shared & { placement: Placement }) {
  const rows = rackRows(props.placement);
  return (
    <div
      role="list"
      aria-label={`Rack ${props.rack}, top to bottom`}
      className="overflow-hidden rounded-md border-x-[6px] border-y-[10px] border-zinc-300 bg-zinc-100/60 dark:border-zinc-700 dark:bg-zinc-900/60"
    >
      {rows.map((r) => {
        if (r.kind === "gap") {
          const n = r.top - r.bottom + 1;
          const label = `${r.top}–${r.bottom}`;
          return (
            <div key={`gap-${r.top}`} role="listitem" className={cn(row, "h-8")}>
              <Rail label={label} side="left" holes={false} />
              <div className="flex items-center border-t border-dashed border-zinc-300 px-3 text-xs text-muted-foreground dark:border-zinc-700">
                {n} empty units
              </div>
              <Rail label={label} side="right" holes={false} />
            </div>
          );
        }
        const label = String(r.u).padStart(2, "0");
        return (
          <div
            key={r.u}
            role="listitem"
            aria-label={`U${r.u}`}
            className={cn(row, r.hosts.length ? "min-h-12" : "h-8")}
          >
            <Rail label={label} side="left" />
            {r.hosts.length ? (
              <div className="flex min-w-0 border-t border-zinc-300 bg-card shadow-[inset_0_-1px_0_rgba(0,0,0,0.04)] dark:border-zinc-700">
                <Ear />
                <div className="flex min-w-0 flex-1 divide-x divide-dashed max-md:flex-col max-md:divide-x-0 max-md:divide-y">
                  {r.hosts.map((h) => (
                    <Faceplate key={h.name} host={h} {...props} />
                  ))}
                </div>
                <Ear />
              </div>
            ) : (
              <div className="border-t border-dashed border-zinc-300 dark:border-zinc-700" />
            )}
            <Rail label={label} side="right" />
          </div>
        );
      })}
    </div>
  );
}

// UnplacedNodes lists the rack's nodes whose name doesn't say which U they are in, so they aren't lost from the view
export function UnplacedNodes(props: Shared & { hosts: Host[] }) {
  if (props.hosts.length === 0) return null;
  return (
    <section className="grid gap-2">
      <div>
        <h2 className="text-sm font-semibold">Not placed</h2>
        <p className="text-sm text-muted-foreground">
          Tagged rack={props.rack}, but the third part of the name isn&apos;t a rack unit.
        </p>
      </div>
      <div className="divide-y rounded-md border bg-card">
        {props.hosts.map((h) => (
          <Faceplate key={h.name} host={h} {...props} />
        ))}
      </div>
    </section>
  );
}
