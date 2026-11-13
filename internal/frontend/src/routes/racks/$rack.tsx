import { getV1NodesFind, getV1NodesProvisionStatus, type ProvisionStatus } from "@/client";
import ActionsSheet from "@/components/actions-sheet";
import NodeActions from "@/components/actions/nodes";
import { DetailHeader, Fact, Unset } from "@/components/detail-header";
import { ErrorPanel } from "@/components/error-panel";
import { isFinished, statusesByNode } from "@/components/nodes/provision-status";
import { placeNodes, rackOrder, sharedTags } from "@/components/racks/elevation";
import { type Detail, RackElevation, TagList, UnplacedNodes } from "@/components/racks/rack-view";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import AuthRedirect from "@/lib/auth";
import { createFileRoute, Link, useRouter } from "@tanstack/react-router";
import { RefreshCw, Server, Settings2 } from "lucide-react";
import { type MouseEvent, useEffect, useMemo, useRef, useState } from "react";

export const Route = createFileRoute("/racks/$rack")({
  staticData: { crumb: { param: "rack" } },
  component: RackPage,
  beforeLoad: AuthRedirect,
  // the statuses load with the nodes so a refresh reloads both. The client throws on an error, so a failed status request is caught here and only leaves the badges off, as for a role without that permission
  loader: async ({ params: { rack } }) => {
    const tags = `rack=${rack}`;
    const [nodes, statuses] = await Promise.all([
      getV1NodesFind({ query: { tags } }),
      getV1NodesProvisionStatus({ query: { tags } }).catch(() => undefined),
    ]);
    return { ...nodes, statuses: statuses?.data ?? [] };
  },
});

const detailOptions: [Detail, string][] = [
  ["status", "Last provision"],
  ["image", "Boot image"],
  ["arch", "Arch"],
  ["firmware", "Firmware"],
  ["tags", "Tags"],
  ["bmc", "BMC link"],
];
const detailsKey = "grendel.rack.show";

// useDetails remembers which details the rack shows, per browser
function useDetails() {
  const [details, setDetails] = useState<Detail[]>(() => {
    try {
      const saved = JSON.parse(localStorage.getItem(detailsKey) ?? "null");
      if (Array.isArray(saved)) return saved;
    } catch {
      // storage can be blocked, the defaults do
    }
    return ["status", "tags", "bmc"];
  });
  useEffect(() => {
    try {
      localStorage.setItem(detailsKey, JSON.stringify(details));
    } catch {
      // as above
    }
  }, [details]);
  return [details, setDetails] as const;
}

// RackFacts sum up the rack: how full it is, how many nodes will provision, and how their last provisions went
function RackFacts({
  rack,
  nodes,
  used,
  height,
  provisioning,
  statuses,
  common,
}: {
  rack: string;
  nodes: number;
  used: number;
  height: number;
  provisioning: number;
  statuses: ProvisionStatus[];
  common: string[];
}) {
  const failed = statuses.filter((s) => s.state === "failed").length;
  const running = statuses.filter((s) => s.state && !isFinished(s.state)).length;
  const done = statuses.filter((s) => s.state === "completed").length;
  const tags = [`rack=${rack}`];

  return (
    <>
      <Fact label="Nodes">
        <Link to="/nodes" search={{ tags }} className="underline-offset-4 hover:underline">
          {nodes}
        </Link>
      </Fact>
      <Fact label="Units used">
        {used} of {height}
      </Fact>
      <Fact label="Provision">
        {provisioning ? (
          <Link
            to="/nodes"
            search={{ tags, provision: "true" }}
            className="underline-offset-4 hover:underline"
          >
            {provisioning} on
          </Link>
        ) : (
          <Unset>all off</Unset>
        )}
        {provisioning > 0 && provisioning < nodes && (
          <span className="text-muted-foreground">, {nodes - provisioning} off</span>
        )}
      </Fact>
      {statuses.length > 0 && (
        <Fact label="Last provision">
          {[
            failed > 0 && (
              <span key="f" className="text-destructive">
                {failed} failed
              </span>
            ),
            running > 0 && <span key="r">{running} in progress</span>,
            done > 0 && <span key="d">{done} completed</span>,
          ]
            .filter(Boolean)
            .flatMap((el, i) => (i ? [", ", el] : [el]))}
        </Fact>
      )}
      {common.length > 0 && (
        <Fact label="Shared tags">
          <TagList tags={common} max={6} label="Shared tags" className="justify-start py-0.5" />
        </Fact>
      )}
    </>
  );
}

function RackPage() {
  const router = useRouter();
  const rack = Route.useParams().rack;
  const { data, statuses } = Route.useLoaderData();
  const hosts = data ?? [];
  const placement = useMemo(() => placeNodes(hosts), [hosts]);
  const order = useMemo(() => rackOrder(placement), [placement]);
  const common = useMemo(() => sharedTags(hosts, rack), [hosts, rack]);
  const statusByNode = useMemo(() => statusesByNode(statuses), [statuses]);
  const [details, setDetails] = useDetails();
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const anchor = useRef<string | null>(null);

  // a shift-click selects or clears every node between it and the last one clicked, in the order the rack shows them
  const onSelect = (name: string, on: boolean, e: MouseEvent) => {
    const next = new Set(selected);
    const from = anchor.current ? order.indexOf(anchor.current) : -1;
    const to = order.indexOf(name);
    const range =
      e.shiftKey && from >= 0 && to >= 0
        ? order.slice(Math.min(from, to), Math.max(from, to) + 1)
        : [name];
    for (const n of range)
      if (on) next.add(n);
      else next.delete(n);
    anchor.current = name;
    setSelected(next);
  };

  if (hosts.length === 0)
    return (
      <Card>
        <CardContent className="p-6">
          <ErrorPanel
            title="Rack not found"
            message={`No nodes are tagged rack=${rack}. Tag the nodes in this rack to see them here.`}
            icon={Server}
          />
        </CardContent>
      </Card>
    );

  const checked = order.filter((n) => selected.has(n));
  const all = checked.length === order.length;
  const shared = {
    rack,
    common: new Set(common),
    statuses: statusByNode,
    details,
    selected,
    onSelect,
  };

  return (
    <Card>
      <CardContent className="p-0">
        <DetailHeader
          title={rack}
          facts={
            <RackFacts
              rack={rack}
              nodes={hosts.length}
              used={placement.units.size}
              height={placement.height}
              provisioning={hosts.filter((h) => h.provision).length}
              statuses={statuses}
              common={common}
            />
          }
          actions={
            <>
              <ActionsSheet type="node" checked={checked.join(",")} length={checked.length}>
                <NodeActions nodes={checked.join(",")} />
              </ActionsSheet>
              <Button variant="secondary" type="button" onClick={() => router.invalidate()}>
                <RefreshCw />
                <span className="sr-only md:not-sr-only">Refresh</span>
              </Button>
            </>
          }
        />
        <div className="mx-auto grid w-full max-w-6xl gap-4 px-6 py-5">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <label className="flex items-center gap-2 text-sm">
              <Checkbox
                checked={all}
                indeterminate={!all && checked.length > 0}
                onCheckedChange={() => setSelected(all ? new Set() : new Set(order))}
              />
              {checked.length ? (
                <span>
                  {checked.length} of {order.length} selected
                  <span className="text-muted-foreground">
                    {" "}
                    for Actions. Shift-click selects a range.
                  </span>
                </span>
              ) : (
                <span className="text-muted-foreground">Select all</span>
              )}
            </label>
            <DropdownMenu>
              <DropdownMenuTrigger render={<Button type="button" variant="secondary" size="sm" />}>
                <Settings2 />
                Show
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuGroup>
                  <DropdownMenuLabel>On each node</DropdownMenuLabel>
                  {detailOptions.map(([key, label]) => (
                    <DropdownMenuCheckboxItem
                      key={key}
                      checked={details.includes(key)}
                      onCheckedChange={(on) =>
                        setDetails(on ? [...details, key] : details.filter((d) => d !== key))
                      }
                    >
                      {label}
                    </DropdownMenuCheckboxItem>
                  ))}
                </DropdownMenuGroup>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
          <RackElevation placement={placement} {...shared} />
          <UnplacedNodes hosts={placement.unplaced} {...shared} />
        </div>
      </CardContent>
    </Card>
  );
}
