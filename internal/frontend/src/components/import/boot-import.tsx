import type { GetV1BmcNicsResponse, NodeImportResponse } from "@/client";
import { getV1BmcNics } from "@/client";
import {
  getV1BmcNicsOptions,
  getV1BmcNicsQueryKey,
  getV1NodesQueryKey,
  getV1SwitchScanOptions,
} from "@/client/@tanstack/react-query.gen";
import { BootGrid } from "@/components/import/boot-grid";
import { BootSettingsPanel } from "@/components/import/boot-settings";
import { NodePicker, readable } from "@/components/import/node-picker";
import { useNodeList } from "@/components/nodes/node-table";
import { Review } from "@/components/import/review";
import { StepBar } from "@/components/import/step-bar";
import { Stepper } from "@/components/import/stepper";
import { SwitchPicker } from "@/components/import/switch-picker";
import { Button } from "@/components/ui/button";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { Skeleton } from "@/components/ui/skeleton";
import { useStoredJSON } from "@/hooks/use-stored-value";
import {
  bootNeedsAttention,
  bootRows,
  compileBoot,
  defaultBootSettings,
  defaultRule,
  groups as groupNodes,
  toBootImport,
  type BootOverrides,
  type NodeNics,
  type Rule,
  learned,
} from "@/lib/boot-nics";
import { expandOrEmpty, fold } from "@/lib/nodeset";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useRefreshPart } from "./use-refresh-part";
import { Link } from "@tanstack/react-router";
import {
  ArrowLeft,
  ArrowRight,
  ChevronRight,
  CircleCheck,
  Cpu,
  IdCard,
  LoaderCircle,
  RefreshCw,
  ScanSearch,
} from "lucide-react";
import { useMemo, useState, type ReactNode } from "react";

const settingsKey = "grendel.boot-nics.settings";

export const bootSteps = ["nodes", "ports", "review", "done"] as const;
export type BootStep = (typeof bootSteps)[number];

// the step, the nodes and the switches live in the URL, so a reload or a shared link lands in the same place, and the switch import can hand its nodes and switches over
type NicResult = NonNullable<GetV1BmcNicsResponse["results"]>[number];

export type BootSearch = { bootStep: BootStep; nodes: string; switches: string };

// BootImport is the interface wizard: read each node's interfaces from its BMC, pick one by a rule per interface layout, and store it with an address. Like SwitchImport, it hands its stepper and body to children to lay out
export function BootImport({
  search,
  onSearch,
  children,
}: {
  search: BootSearch;
  onSearch: (patch: Partial<BootSearch>, replace?: boolean) => void;
  children: (parts: { stepper?: ReactNode; body: ReactNode }) => ReactNode;
}) {
  const { bootStep: wanted, nodes, switches } = search;
  const queryClient = useQueryClient();

  const [settings, store] = useStoredJSON(settingsKey, defaultBootSettings);
  const fqdn = useMemo(() => compileBoot(settings), [settings]);

  const [draft, setDraft] = useState(() => expandOrEmpty(nodes));
  const [seenNodes, setSeenNodes] = useState(nodes);
  // a link elsewhere can change the nodes under the page, which the table then shows picked
  if (nodes !== seenNodes) {
    setSeenNodes(nodes);
    setDraft(expandOrEmpty(nodes));
  }
  const list = useNodeList();
  const picked = useMemo(() => {
    const want = new Set(draft);
    return (list.data ?? []).filter((h) => want.has(h.name ?? ""));
  }, [draft, list.data]);
  const toRead = picked.filter(readable);

  const [ruleEdits, setRuleEdits] = useState<Record<string, Rule>>({});
  const [overrides, setOverrides] = useState<BootOverrides>({});
  const [result, setResult] = useState<{ res: NodeImportResponse; nodes: string[] }>();

  const selectedSwitches = useMemo(() => expandOrEmpty(switches), [switches]);

  const step: BootStep =
    wanted !== "nodes" && !nodes ? "nodes" : wanted === "done" && !result ? "review" : wanted;
  const go = (to: BootStep, extra?: Partial<BootSearch>) => onSearch({ ...extra, bootStep: to });

  const nicsOptions = getV1BmcNicsOptions({ query: { nodeset: nodes } });
  const nics = useQuery({
    ...nicsOptions,
    enabled: step !== "nodes" && !!nodes,
    staleTime: Infinity,
    gcTime: Infinity,
  });
  // the same query as the switch import's, so switches it just scanned aren't read again
  const scan = useQuery({
    ...getV1SwitchScanOptions({ query: { nodeset: switches } }),
    enabled: step !== "nodes" && !!switches,
    staleTime: Infinity,
    gcTime: Infinity,
  });

  const read: NodeNics[] = useMemo(
    () =>
      (nics.data?.results ?? []).map((r) =>
        r.success
          ? { node: r.node ?? "", nics: r.data?.nics ?? [] }
          : { node: r.node ?? "", nics: [], error: r.message ?? "" },
      ),
    [nics.data],
  );
  const seen = useMemo(() => learned(scan.data?.switches ?? []), [scan.data]);
  const byNode = useMemo(() => new Map(read.map((n) => [n.node, n.nics])), [read]);
  const grouped = useMemo(() => groupNodes(read), [read]);
  const rules = useMemo(() => {
    const out: Record<string, Rule> = {};
    for (const g of grouped) out[g.key] = ruleEdits[g.key] ?? defaultRule(g, byNode, seen);
    return out;
  }, [grouped, ruleEdits, byNode, seen]);
  const rows = useMemo(
    () => bootRows(read, list.data ?? [], rules, overrides, fqdn, seen),
    [read, list.data, rules, overrides, fqdn, seen],
  );
  const importRows = useMemo(() => toBootImport(rows, settings), [rows, settings]);
  const attention = rows.filter((r) => r.include && bootNeedsAttention(r)).length;
  const changes = rows.filter((r) => r.include && r.status === "ready").length;
  const scanFailed = (scan.data?.switches ?? []).filter((s) => s.error);

  // asks the BMCs that didn't answer again, and puts what they say in place of their errors
  const { refresh: retry, busy: retryingNodes } = useRefreshPart<GetV1BmcNicsResponse, NicResult>({
    queryKey: nicsOptions.queryKey,
    fetch: async (nodeset) => (await getV1BmcNics({ query: { nodeset } })).data?.results ?? [],
    entries: (d) => d.results ?? [],
    withEntries: (d, results) => ({ ...d, results }),
    name: (r) => r.node,
  });
  const retrying = retryingNodes.size > 0;

  const stepList = [
    { id: "nodes", label: "Nodes" },
    { id: "ports", label: "Interfaces" },
    { id: "review", label: "Addresses and review" },
  ];

  const stepper = step !== "done" && (
    <Stepper
      steps={stepList}
      current={step}
      reachable={(id) =>
        id === "nodes" || (id === "ports" && !!nodes) || (id === "review" && importRows.length > 0)
      }
      onGo={(id) => go(id as BootStep)}
    />
  );

  // only the nodes whose BMCs can be asked are read, since one name the API can't find fails the whole request
  const start = () => {
    if (toRead.length === 0) return;
    const nodeset = fold(toRead.map((h) => h.name ?? ""));
    setRuleEdits({});
    setOverrides({});
    queryClient.removeQueries({
      queryKey: getV1BmcNicsQueryKey({ query: { nodeset } }),
    });
    go("ports", { nodes: nodeset });
  };

  const body = (
    <>
      {step === "nodes" && (
        <div className="grid gap-6">
          <NodePicker selected={draft} onChange={setDraft} />
          <SwitchSection
            selected={selectedSwitches}
            onChange={(names) => onSearch({ switches: fold(names) }, true)}
          />
          <StepBar>
            <p className="mr-auto text-sm text-muted-foreground">
              {picked.length === 0
                ? "Pick the nodes to add interfaces to."
                : `${toRead.length} BMC${toRead.length === 1 ? "" : "s"} to ask${
                    selectedSwitches.length
                      ? ` and ${selectedSwitches.length} switch${selectedSwitches.length === 1 ? "" : "es"} to scan`
                      : ""
                  }${
                    picked.length > toRead.length
                      ? `, leaving out ${picked.length - toRead.length} with no BMC address`
                      : ""
                  }`}
            </p>
            <Button disabled={toRead.length === 0} onClick={start}>
              <ScanSearch />
              Read interfaces
            </Button>
          </StepBar>
        </div>
      )}

      {step === "ports" &&
        (nics.isPending || (!!switches && scan.isPending) ? (
          <Reading nodes={nodes} switches={selectedSwitches.length} />
        ) : nics.isError ? (
          <div className="grid justify-items-start gap-3">
            <p className="text-sm text-destructive">
              {nics.error.detail ?? "Couldn't ask the BMCs."}
            </p>
            <Button variant="outline" onClick={() => nics.refetch()}>
              <RefreshCw />
              Try again
            </Button>
          </div>
        ) : (
          <>
            <BootSettingsPanel settings={settings} fqdn={fqdn} rows={rows} onChange={store} />
            {(scan.isError || scanFailed.length > 0) && (
              <p className="text-sm text-amber-700 dark:text-amber-300">
                {scan.isError
                  ? "The switches couldn't be scanned, so interfaces are picked without them."
                  : `Couldn't scan ${fold(scanFailed.map((s) => s.name ?? ""))}, so interfaces cabled to ${scanFailed.length === 1 ? "it" : "them"} aren't marked as seen.`}
              </p>
            )}
            <BootGrid
              groups={grouped}
              rows={rows}
              rules={rules}
              byNode={byNode}
              seen={seen}
              overrides={overrides}
              retrying={retrying}
              refreshing={nics.isFetching || scan.isFetching}
              onRule={(g, r) => setRuleEdits((e) => ({ ...e, [g]: r }))}
              onOverrides={setOverrides}
              onRetry={retry}
              onRefresh={() => {
                void nics.refetch();
                void list.refetch();
                if (switches) void scan.refetch();
              }}
            />
            <StepBar>
              <p className="mr-auto text-sm text-muted-foreground">
                <span className="font-medium text-foreground tabular-nums">{changes}</span> to set
                {importRows.length > changes && (
                  <>, {importRows.length - changes} already in Grendel</>
                )}
                {attention > 0 && (
                  <>
                    ,{" "}
                    <span className="text-amber-700 tabular-nums dark:text-amber-300">
                      {attention}
                    </span>{" "}
                    need attention
                  </>
                )}
              </p>
              <Button variant="ghost" onClick={() => go("nodes")}>
                <ArrowLeft />
                Nodes
              </Button>
              <Button disabled={importRows.length === 0} onClick={() => go("review")}>
                Addresses
                <ArrowRight />
              </Button>
            </StepBar>
          </>
        ))}

      {step === "review" && (
        <Review
          rows={importRows}
          bmc={false}
          back={
            <Button variant="ghost" onClick={() => go("ports")}>
              <ArrowLeft />
              Interfaces
            </Button>
          }
          onCommitted={(res, names) => {
            void queryClient.invalidateQueries({ queryKey: getV1NodesQueryKey() });
            setResult({ res, nodes: names });
            go("done");
          }}
        />
      )}

      {step === "done" && result && (
        <Done
          result={result}
          onAgain={() => {
            setResult(undefined);
            setOverrides({});
            setRuleEdits({});
            go("nodes", { nodes: "" });
          }}
        />
      )}
    </>
  );

  return children({ stepper: stepper || undefined, body });
}

// SwitchSection is the optional choice of switches to scan, closed until wanted unless switches are already chosen, as when the switch import hands them over
function SwitchSection({
  selected,
  onChange,
}: {
  selected: string[];
  onChange: (names: string[]) => void;
}) {
  const [open, setOpen] = useState(selected.length > 0);
  return (
    <Collapsible
      open={open}
      onOpenChange={setOpen}
      className="grid overflow-hidden rounded-lg border bg-muted/30"
    >
      <CollapsibleTrigger className="group flex w-full items-start gap-3 px-4 py-3 text-left outline-none hover:bg-muted/40 focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-inset data-panel-open:border-b">
        <ChevronRight className="mt-0.5 size-4 shrink-0 text-muted-foreground transition-transform group-data-panel-open:rotate-90 motion-reduce:transition-none" />
        <span className="grid gap-0.5">
          <span className="text-sm font-semibold tracking-tight">
            Switches the host interfaces are cabled to
            <span className="ml-2 text-xs font-normal text-muted-foreground">
              {selected.length ? fold(selected) : "optional"}
            </span>
          </span>
          <span className="text-xs text-muted-foreground">
            A port a switch has learned is the one that's cabled. That is a surer pick than link
            state, which a BMC often can't report while the host is off.
          </span>
        </span>
      </CollapsibleTrigger>
      <CollapsibleContent className="p-4">
        <SwitchPicker selected={selected} onChange={onChange} />
      </CollapsibleContent>
    </Collapsible>
  );
}

function Reading({ nodes, switches }: { nodes: string; switches: number }) {
  return (
    <div className="grid gap-3" aria-busy>
      <p className="flex items-center gap-2 text-sm text-muted-foreground">
        <LoaderCircle className="size-4 animate-spin" />
        Asking the BMCs of <span className="font-mono text-foreground">{nodes}</span> for their
        host's interfaces
        {switches > 0 && ` and scanning ${switches} switch${switches === 1 ? "" : "es"}`}…
      </p>
      <Skeleton className="h-14 w-full" />
      <Skeleton className="h-64 w-full" />
    </div>
  );
}

function Done({
  result,
  onAgain,
}: {
  result: { res: NodeImportResponse; nodes: string[] };
  onAgain: () => void;
}) {
  const s = result.res.summary ?? {};
  const nodes = fold(result.nodes);
  const next = [
    {
      to: "/nodes/bios/$nodes" as const,
      icon: <Cpu />,
      title: "Configure BIOS",
      text: "Turn on PXE for the boot port and set the boot order.",
    },
    {
      to: "/nodes/identity/$nodes" as const,
      icon: <IdCard />,
      title: "Set locations and asset tags",
      text: "Write each node's rack location to its BMC.",
    },
  ];

  return (
    <div className="grid gap-6">
      <div className="flex items-start gap-3">
        <CircleCheck className="mt-0.5 size-6 shrink-0 text-(--node-ready)" />
        <div className="grid gap-1">
          <h2 className="text-lg font-semibold tracking-tight">
            Saved interfaces for {result.nodes.length} node{result.nodes.length === 1 ? "" : "s"}
          </h2>
          <p className="text-sm text-muted-foreground">
            {[
              s.add && `${s.add} given an interface`,
              s.replace && `${s.replace} had an interface updated`,
              s.unchanged && `${s.unchanged} already as shown`,
            ]
              .filter(Boolean)
              .join(", ")}
            .
          </p>
          <p className="font-mono text-sm break-all">{nodes}</p>
        </div>
      </div>

      <div className="grid gap-2">
        <h3 className="text-sm font-medium">Next, for these nodes</h3>
        <ul className="grid gap-2 sm:grid-cols-2">
          {next.map((n) => (
            <li key={n.to}>
              <Link
                to={n.to}
                params={{ nodes }}
                className="flex h-full gap-3 rounded-lg border p-3 outline-none hover:bg-muted/50 focus-visible:ring-2 focus-visible:ring-ring [&>svg]:mt-0.5 [&>svg]:size-4 [&>svg]:shrink-0 [&>svg]:text-muted-foreground"
              >
                {n.icon}
                <span className="grid gap-0.5">
                  <span className="text-sm font-medium">{n.title}</span>
                  <span className="text-xs text-muted-foreground">{n.text}</span>
                </span>
              </Link>
            </li>
          ))}
        </ul>
      </div>

      <div>
        <Button variant="outline" onClick={onAgain}>
          Add more interfaces
        </Button>
      </div>
    </div>
  );
}
