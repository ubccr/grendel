import { getV1Nodes, getV1NodesProvisionStatus, type Host, type ProvisionStatus } from "@/client";
import {
  isFinished,
  LastProvisionCounts,
  statusesByNode,
} from "@/components/nodes/provision-status";
import { hallRow, type Placement, placeNodes, rackTags } from "@/components/racks/elevation";
import { TagList } from "@/components/racks/rack-view";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Input } from "@/components/ui/input";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import AuthRedirect from "@/lib/auth";
import { parseNodeSet } from "@/lib/nodeset";
import { cn } from "@/lib/utils";
import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";
import { Copy, SearchX, SquareMenu } from "lucide-react";
import {
  type ComponentProps,
  type MouseEvent,
  type ReactNode,
  type RefObject,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
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
  // the statuses load with the nodes so the racks can light failed and running provisions. A failed status request only leaves those lights off, as for a role without that permission
  loader: async () => {
    const [nodes, statuses] = await Promise.all([
      getV1Nodes(),
      getV1NodesProvisionStatus().catch(() => undefined),
    ]);
    return { ...nodes, statuses: statuses?.data ?? [] };
  },
});

type Rack = {
  name: string;
  hosts: Host[];
  placement: Placement;
};

type Tone = "on" | "off" | "running" | "failed";

// a failed provision outranks the provision switch, since it is the one that needs someone to look at it
function toneOf(host: Host, status?: ProvisionStatus): Tone {
  if (status?.state === "failed") return "failed";
  if (status?.state && !isFinished(status.state)) return "running";
  return host.provision ? "on" : "off";
}

const toneClass: Record<Tone, string> = {
  on: "bg-(--node-ready)",
  off: "bg-zinc-400 dark:bg-zinc-500",
  running: "bg-(--node-blue)",
  failed: "bg-destructive",
};

const toneLabel: Record<Tone, string> = {
  on: "Provision on",
  off: "Provision off",
  running: "Provisioning",
  failed: "Last provision failed",
};

const byName = new Intl.Collator(undefined, { numeric: true }).compare;

// racksOf gathers the nodes tagged rack=<name> into their racks
function racksOf(hosts: Host[]): Rack[] {
  const groups = new Map<string, Host[]>();
  for (const h of hosts) {
    const tag = h.tags?.find((t) => t.startsWith("rack="));
    if (!tag) continue;
    const name = tag.slice("rack=".length);
    groups.set(name, [...(groups.get(name) ?? []), h]);
  }
  return [...groups.entries()]
    .sort(([a], [b]) => byName(a, b))
    .map(([name, hosts]) => ({ name, hosts, placement: placeNodes(hosts) }));
}

// hallRows splits the racks into the rows of the hall they stand in, the racks with no row last
function hallRows(racks: Rack[]): [string, Rack[]][] {
  const rows = new Map<string, Rack[]>();
  for (const r of racks) {
    const key = hallRow(r.name);
    rows.set(key, [...(rows.get(key) ?? []), r]);
  }
  return [...rows.entries()].sort(([a], [b]) => (!a ? 1 : !b ? -1 : byName(a, b)));
}

function RouteComponent() {
  const { data, statuses } = Route.useLoaderData();
  const { name = "", tag = "" } = Route.useSearch();
  const navigate = useNavigate();

  const racks = useMemo(() => racksOf(data ?? []), [data]);
  const statusByNode = useMemo(() => statusesByNode(statuses), [statuses]);

  const setFilter = (key: "name" | "tag", value: string) =>
    navigate({
      to: "/racks",
      search: (prev) => ({ ...prev, [key]: value || undefined }),
      replace: true,
    });

  // the filters ignore case. A node matches the tag filter when one of its tags contains it, and a rack is shown when any of its nodes match
  const nameQuery = name.toLowerCase();
  const tagQuery = tag.toLowerCase();
  const matches = (h: Host) =>
    !tagQuery || !!h.tags?.some((t) => t.toLowerCase().includes(tagQuery));
  const shown = racks.filter(
    (r) => r.name.toLowerCase().includes(nameQuery) && r.hosts.some(matches),
  );
  const rows = hallRows(shown);

  const nodeCount = racks.reduce((n, r) => n + r.hosts.length, 0);
  const failed = racks.reduce(
    (n, r) => n + r.hosts.filter((h) => statusByNode.get(h.name ?? "")?.state === "failed").length,
    0,
  );

  if (racks.length === 0)
    return (
      <Card>
        <CardContent>
          <Empty>
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <SquareMenu />
              </EmptyMedia>
              <EmptyTitle>No racks yet</EmptyTitle>
              <EmptyDescription>
                Tag each node with the rack it&apos;s in, like rack=f01, and the rack appears here.
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        </CardContent>
      </Card>
    );

  return (
    <Card>
      <CardContent className="p-0">
        <header className="grid gap-4 border-b px-6 py-5 lg:grid-cols-[1fr_auto] lg:items-end">
          <div className="grid gap-1.5">
            <h1 className="text-3xl font-semibold tracking-tight">Racks</h1>
            <p className="text-sm text-muted-foreground">
              {nodeCount.toLocaleString()} {nodeCount === 1 ? "node" : "nodes"} in {racks.length}{" "}
              {racks.length === 1 ? "rack" : "racks"}
              {failed > 0 && (
                <>
                  ,{" "}
                  <span className="text-destructive">
                    {failed} failed {failed === 1 ? "its" : "their"} last provision
                  </span>
                </>
              )}
              .
            </p>
          </div>
          <div className="grid gap-2 sm:grid-cols-2 lg:w-[28rem]">
            <FilterInput
              aria-label="Filter racks by name"
              value={name}
              onValueChange={(v) => setFilter("name", v)}
              placeholder="Rack name"
            />
            <FilterInput
              aria-label="Light nodes with a tag"
              value={tag}
              onValueChange={(v) => setFilter("tag", v)}
              placeholder="Tag, like gpu"
            />
          </div>
        </header>

        <div className="grid gap-8 px-6 py-6">
          <Legend />
          {rows.length === 0 ? (
            <Empty className="py-10">
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <SearchX />
                </EmptyMedia>
                <EmptyTitle>No racks match</EmptyTitle>
                <EmptyDescription>
                  {name && tag
                    ? `No rack named like “${name}” has a node tagged like “${tag}”.`
                    : name
                      ? `No rack is named like “${name}”.`
                      : `No node in a rack has a tag like “${tag}”.`}
                </EmptyDescription>
              </EmptyHeader>
              <EmptyContent>
                <Button
                  variant="secondary"
                  onClick={() => navigate({ to: "/racks", search: {}, replace: true })}
                >
                  Clear filters
                </Button>
              </EmptyContent>
            </Empty>
          ) : (
            rows.map(([row, racks]) => (
              <section
                key={row}
                aria-label={row ? `Row ${row}` : "Other racks"}
                className="grid gap-3"
              >
                {/* racks with no row need no heading when they are the only ones, they are the whole page */}
                {(rows.length > 1 || row) && (
                  <h2 className="flex items-baseline gap-2 border-b pb-2 text-sm">
                    {row ? (
                      <>
                        <span className="text-muted-foreground">Row</span>
                        <span className="font-mono text-base font-semibold">{row}</span>
                      </>
                    ) : (
                      <span className="font-semibold">Other racks</span>
                    )}
                    <span className="ml-auto text-xs text-muted-foreground tabular-nums">
                      {racks.length} {racks.length === 1 ? "rack" : "racks"}
                    </span>
                  </h2>
                )}
                <ul className="flex flex-wrap items-end gap-x-2 gap-y-5">
                  {racks.map((r) => (
                    <li key={r.name}>
                      <MiniRack
                        rack={r}
                        statuses={statusByNode}
                        matches={tagQuery ? matches : undefined}
                      />
                    </li>
                  ))}
                </ul>
              </section>
            ))
          )}
        </div>
      </CardContent>
    </Card>
  );
}

// Legend names the colours a node is lit in
function Legend() {
  return (
    <div className="flex flex-wrap items-center gap-x-5 gap-y-2 text-xs text-muted-foreground">
      {(Object.keys(toneLabel) as Tone[]).map((t) => (
        <span key={t} className="flex items-center gap-1.5">
          <span className={cn("h-1 w-4 rounded-[1px]", toneClass[t])} />
          {toneLabel[t]}
        </span>
      ))}
    </div>
  );
}

// MiniRack draws a rack the way it stands in the hall, top unit first, each node a lit slot in the unit its name gives and sleds side by side. Hovering it shows the rack's facts and a click opens the rack. A tap or a key opens the facts instead, with focus inside them, since touch has no hover and a keyboard can't otherwise reach the card's links
function MiniRack({
  rack,
  statuses,
  matches,
}: {
  rack: Rack;
  statuses: Map<string, ProvisionStatus>;
  matches?: (h: Host) => boolean;
}) {
  const { placement, hosts, name } = rack;
  const units = Array.from({ length: placement.height }, (_, i) => placement.height - i);
  const lit = matches ? hosts.filter(matches).length : hosts.length;
  const open = useRef<HTMLAnchorElement>(null);

  // only a mouse or pen click follows the link. A tap or Enter, whose click has another or no pointer type, is kept from it so the popover opens instead
  const onClick = (e: MouseEvent) => {
    const type = (e.nativeEvent as PointerEvent).pointerType;
    if (type !== "mouse" && type !== "pen") e.preventDefault();
  };

  return (
    <Popover>
      <PopoverTrigger
        openOnHover
        delay={250}
        closeDelay={100}
        nativeButton={false}
        onClick={onClick}
        render={
          <Link
            to="/racks/$rack"
            params={{ rack: name }}
            aria-label={`Rack ${name}, ${hosts.length} nodes`}
            className="group grid w-16 justify-items-center gap-1.5 rounded-md px-1 pt-1 pb-1.5 hover:bg-foreground/5 focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
          />
        }
      >
        <div
          aria-hidden
          className="w-12 rounded-[3px] border-x-[3px] border-y-[5px] border-zinc-300 bg-zinc-100 transition-colors group-hover:border-zinc-400 dark:border-zinc-700 dark:bg-zinc-900 dark:group-hover:border-zinc-500"
        >
          <div className="grid gap-px px-[3px] py-px">
            {units.map((u) => {
              const list = placement.units.get(u);
              if (!list?.length) return <span key={u} className="h-[3px]" />;
              return (
                <span key={u} className="flex h-[3px] gap-px">
                  {list.map((h) => (
                    <span
                      key={h.name}
                      className={cn(
                        "flex-1 rounded-[0.5px]",
                        toneClass[toneOf(h, statuses.get(h.name ?? ""))],
                        matches && !matches(h) && "opacity-15",
                      )}
                    />
                  ))}
                </span>
              );
            })}
          </div>
        </div>
        <span
          title={name}
          className="max-w-full truncate font-mono text-xs font-medium underline-offset-4 group-hover:underline"
        >
          {name}
        </span>
        <span className="-mt-1 text-[11px] text-muted-foreground tabular-nums">
          {matches ? `${lit} of ${hosts.length}` : hosts.length}
        </span>
      </PopoverTrigger>
      <PopoverContent
        side="right"
        align="start"
        initialFocus={(type) => (type === "mouse" ? false : open.current)}
        className="w-80 max-w-[calc(100vw-2rem)]"
      >
        <RackSummary rack={rack} statuses={statuses} openRef={open} />
      </PopoverContent>
    </Popover>
  );
}

// RackSummary is what the popover tells about a rack: how full it is, how its nodes are set to provision and how their last provisions went, the tags on all or some of them, its nodeset to copy, and a link to open it
function RackSummary({
  rack,
  statuses,
  openRef,
}: {
  rack: Rack;
  statuses: Map<string, ProvisionStatus>;
  openRef: RefObject<HTMLAnchorElement | null>;
}) {
  const { name, hosts, placement } = rack;
  const tags = [`rack=${name}`];
  const on = hosts.filter((h) => h.provision).length;
  const last = hosts.flatMap((h) => statuses.get(h.name ?? "") ?? []);
  const { every, some } = rackTags(hosts, name);
  const nodeset = parseNodeSet(
    hosts
      .map((h) => h.name ?? "")
      .filter(Boolean)
      .join(","),
  ).toString();
  const link = "underline-offset-4 hover:underline";

  // grid-cols-1 sizes the column from the card rather than the content, so the unbroken nodeset truncates instead of pushing past the card's edge
  return (
    <div className="grid grid-cols-1 gap-3">
      <div className="flex items-baseline justify-between gap-2">
        <Link
          to="/nodes"
          search={{ tags }}
          className="font-mono text-base font-semibold underline-offset-4 hover:underline"
        >
          {name}
        </Link>
        <span className="text-xs text-muted-foreground tabular-nums">
          {placement.units.size} of {placement.height} units used
        </span>
      </div>
      <dl className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-1.5 text-sm">
        <Row label="Nodes">
          <Link to="/nodes" search={{ tags }} className={link}>
            {hosts.length}
          </Link>
          {placement.unplaced.length > 0 && (
            <span className="text-muted-foreground">, {placement.unplaced.length} not placed</span>
          )}
        </Row>
        <Row label="Provision">
          {on > 0 && (
            <Link to="/nodes" search={{ tags, provision: "true" }} className={link}>
              {on} on
            </Link>
          )}
          {on > 0 && on < hosts.length && ", "}
          {on < hosts.length && (
            <Link to="/nodes" search={{ tags, provision: "false" }} className={link}>
              {hosts.length - on} off
            </Link>
          )}
        </Row>
        {last.length > 0 && (
          <Row label="Last provision">
            <LastProvisionCounts statuses={last} />
          </Row>
        )}
      </dl>
      {every.length > 0 && (
        <div className="grid gap-1">
          <span className="text-xs text-muted-foreground">On every node</span>
          <TagList
            linked
            tags={every}
            max={6}
            label="Shared tags"
            className="justify-start *:max-w-full"
          />
        </div>
      )}
      {some.length > 0 && (
        <div className="grid gap-1">
          <span className="text-xs text-muted-foreground">On some nodes</span>
          {/* a rack can have many of these, so the list scrolls past a few lines rather than growing the card, and a long tag truncates within its chip */}
          <ul className="-m-0.5 flex max-h-28 flex-wrap gap-1 overflow-y-auto p-0.5">
            {some.map(([tag, n]) => (
              <li key={tag} className="max-w-full min-w-0">
                <Link
                  to="/nodes"
                  search={{ tags: [...tags, tag] }}
                  title={`${tag} is on ${n} of ${hosts.length} nodes`}
                  className="flex h-5 max-w-full items-center gap-1.5 rounded-sm bg-secondary px-1.5 text-xs text-secondary-foreground hover:bg-secondary/70 focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
                >
                  <span className="truncate">{tag}</span>
                  <span className="shrink-0 text-muted-foreground tabular-nums">{n}</span>
                </Link>
              </li>
            ))}
          </ul>
        </div>
      )}
      <div className="grid gap-2 border-t pt-3">
        <code title={nodeset} className="truncate text-xs text-muted-foreground">
          {nodeset}
        </code>
        <div className="flex justify-end gap-2">
          <Button size="sm" variant="secondary" onClick={() => copy(nodeset)}>
            <Copy />
            Copy nodeset
          </Button>
          <Button
            size="sm"
            nativeButton={false}
            render={<Link ref={openRef} to="/racks/$rack" params={{ rack: name }} />}
          >
            Open rack
          </Button>
        </div>
      </div>
    </div>
  );
}

// copy puts the nodeset on the clipboard and says whether it got there, since the browser can refuse, as on a page not served over https
async function copy(nodeset: string) {
  try {
    await navigator.clipboard.writeText(nodeset);
    toast.success("Copied nodeset");
  } catch {
    toast.error("Could not copy the nodeset");
  }
}

function Row({ label, children }: { label: string; children: ReactNode }) {
  return (
    <>
      <dt className="text-muted-foreground">{label}</dt>
      <dd className="min-w-0">{children}</dd>
    </>
  );
}

// FilterInput edits one search param. It keeps its own text so typing never waits on the router, whose updates land in a transition, and takes the param's value only while it isn't focused, so a link, the back button or Clear filters still reach it
function FilterInput({
  value,
  onValueChange,
  ...props
}: Omit<ComponentProps<typeof Input>, "value" | "onChange"> & {
  value: string;
  onValueChange: (value: string) => void;
}) {
  const ref = useRef<HTMLInputElement>(null);
  const [text, setText] = useState(value);
  useEffect(() => {
    if (document.activeElement !== ref.current) setText(value);
  }, [value]);
  return (
    <Input
      ref={ref}
      {...props}
      value={text}
      onChange={(e) => {
        setText(e.currentTarget.value);
        onValueChange(e.currentTarget.value);
      }}
    />
  );
}
