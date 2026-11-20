import type { NodeImportResponse, SwitchScanResponse } from "@/client";
import { getV1SwitchScan } from "@/client";
import { getV1NodesQueryKey, getV1SwitchScanOptions } from "@/client/@tanstack/react-query.gen";
import { Naming } from "@/components/import/naming";
import { PortGrid, type Reveal } from "@/components/import/port-grid";
import { Review } from "@/components/import/review";
import { StepBar } from "@/components/import/step-bar";
import { Stepper } from "@/components/import/stepper";
import { SwitchPicker } from "@/components/import/switch-picker";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useStoredJSON } from "@/hooks/use-stored-value";
import {
  buildRows,
  compile,
  defaultSettings,
  tally,
  toImportRows,
  type Overrides,
  type ScanSwitch,
  type Settings,
} from "@/lib/import-rows";
import { expandOrEmpty, fold } from "@/lib/nodeset";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import {
  ArrowLeft,
  ArrowRight,
  CircleCheck,
  Cpu,
  IdCard,
  LoaderCircle,
  Network,
  Radar,
  RefreshCw,
} from "lucide-react";
import { useMemo, useState, type ReactNode } from "react";
import { useRefreshPart } from "./use-refresh-part";

const settingsKey = "grendel.import.settings";

export const importSteps = ["switches", "map", "review", "done"] as const;
export type ImportStep = (typeof importSteps)[number];

// the step and the chosen switches live in the URL of whichever page shows the wizard, so a reload or a shared link lands in the same place
export type ImportSearch = { step: ImportStep; switches: string };

// SwitchImport is the whole switch import wizard. It hands its stepper and its body to children to lay out, so a page can put the stepper in its own header
export function SwitchImport({
  search,
  onSearch,
  children,
}: {
  search: ImportSearch;
  onSearch: (patch: Partial<ImportSearch>, replace?: boolean) => void;
  children: (parts: { stepper?: ReactNode; body: ReactNode }) => ReactNode;
}) {
  const { step: wanted, switches } = search;
  const queryClient = useQueryClient();

  const [stored, store] = useStoredJSON(settingsKey, defaultSettings);
  // the Port filter belongs to this import alone and is never stored, since one left from an earlier import would silently leave ports out of the next
  const [ports, setPorts] = useState("");
  const settings = useMemo(() => ({ ...stored, ports }), [stored, ports]);
  const setSettings = (s: Settings) => store({ ...s, ports: "" });

  const [overrides, setOverrides] = useState<Overrides>({});
  const [focused, setFocused] = useState<string>();
  const [reveal, setReveal] = useState<Reveal>();
  const [result, setResult] = useState<{ res: NodeImportResponse; nodes: string[] }>();

  const selected = useMemo(() => expandOrEmpty(switches), [switches]);

  // a step that needs what an earlier one made falls back to it, as after a reload mid-import
  const step: ImportStep =
    wanted !== "switches" && selected.length === 0
      ? "switches"
      : wanted === "done" && !result
        ? "review"
        : wanted;
  const go = (to: ImportStep, extra?: { switches?: string }) => onSearch({ ...extra, step: to });

  const scanOptions = getV1SwitchScanOptions({ query: { nodeset: switches } });
  const scan = useQuery({
    ...scanOptions,
    enabled: step !== "switches" && !!switches,
    staleTime: Infinity,
    gcTime: Infinity,
  });

  const compiled = useMemo(() => compile(settings), [settings]);
  const scanned = scan.data?.switches ?? [];
  const rows = useMemo(
    () => buildRows(scanned, settings, compiled, overrides),
    [scanned, settings, compiled, overrides],
  );
  const t = tally(rows);
  const importRows = useMemo(() => toImportRows(rows, settings), [rows, settings]);

  const { refresh: rescan, busy: rescanning } = useRefreshPart<SwitchScanResponse, ScanSwitch>({
    queryKey: scanOptions.queryKey,
    fetch: async (nodeset) => (await getV1SwitchScan({ query: { nodeset } })).data?.switches ?? [],
    entries: (d) => d.switches ?? [],
    withEntries: (d, switches) => ({ ...d, switches }),
    name: (s) => s.name,
  });

  const stepList = [
    { id: "switches", label: "Switches" },
    { id: "map", label: "Ports and names" },
    { id: "review", label: "Addresses and review" },
  ];

  const stepper = step !== "done" && (
    <Stepper
      steps={stepList}
      current={step}
      reachable={(id) =>
        id === "switches" ||
        (id === "map" && selected.length > 0) ||
        (id === "review" && importRows.length > 0)
      }
      onGo={(id) => go(id as ImportStep)}
    />
  );

  const body = (
    <>
      {step === "switches" && (
        <SwitchPicker
          selected={selected}
          onChange={(names) => onSearch({ switches: fold(names) }, true)}
        >
          <StepBar>
            <p className="mr-auto text-sm text-muted-foreground">
              {selected.length === 0
                ? "Pick the switches the new nodes' BMCs are plugged into."
                : `${selected.length} switch${selected.length === 1 ? "" : "es"} selected`}
            </p>
            <Button
              disabled={selected.length === 0}
              onClick={() => {
                setOverrides({});
                queryClient.removeQueries({ queryKey: scanOptions.queryKey });
                go("map");
              }}
            >
              <Radar />
              Scan {selected.length > 1 ? `${selected.length} switches` : "switch"}
            </Button>
          </StepBar>
        </SwitchPicker>
      )}

      {step === "map" &&
        (scan.isPending ? (
          <Scanning count={selected.length} />
        ) : scan.isError ? (
          <div className="grid justify-items-start gap-3">
            <p className="text-sm text-destructive">{scan.error.detail ?? "The scan failed."}</p>
            <Button variant="outline" onClick={() => scan.refetch()}>
              <RefreshCw />
              Try again
            </Button>
          </div>
        ) : (
          <>
            <Naming settings={settings} compiled={compiled} rows={rows} onChange={setSettings} />
            <PortGrid
              switches={scanned}
              rows={rows}
              ports={ports}
              overrides={overrides}
              focused={focused}
              reveal={reveal}
              rescanning={rescanning}
              scanning={scan.isFetching}
              onOverrides={setOverrides}
              onPorts={setPorts}
              onFocus={setFocused}
              onReveal={(key) => {
                setFocused(key);
                setReveal({ key, at: Date.now() });
              }}
              onRescan={(name) => rescan([name])}
              onRescanAll={() => scan.refetch()}
            />
            <StepBar>
              <p className="mr-auto text-sm text-muted-foreground">
                <span className="font-medium text-foreground tabular-nums">
                  {importRows.length}
                </span>{" "}
                ready to import
                {t.attention > 0 && (
                  <>
                    ,{" "}
                    <span className="text-amber-700 tabular-nums dark:text-amber-300">
                      {t.attention}
                    </span>{" "}
                    need attention
                  </>
                )}
                {t.unnamed > 0 && <>, {t.unnamed} waiting for a name template</>}
                {t.empty > 0 && <>, {t.empty} with no MAC are skipped</>}
              </p>
              <Button variant="ghost" onClick={() => go("switches")}>
                <ArrowLeft />
                Switches
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
          bmc
          back={
            <Button variant="ghost" onClick={() => go("map")}>
              <ArrowLeft />
              Ports and names
            </Button>
          }
          onCommitted={(res, nodes) => {
            void queryClient.invalidateQueries({ queryKey: getV1NodesQueryKey() });
            setResult({ res, nodes });
            go("done");
          }}
        />
      )}

      {step === "done" && result && (
        <Done
          result={result}
          switches={switches}
          onAgain={() => {
            setOverrides({});
            setResult(undefined);
            go("switches", { switches: "" });
          }}
        />
      )}
    </>
  );

  return children({ stepper: stepper || undefined, body });
}

function Scanning({ count }: { count: number }) {
  return (
    <div className="grid gap-3" aria-busy>
      <p className="flex items-center gap-2 text-sm text-muted-foreground">
        <LoaderCircle className="size-4 animate-spin" />
        Reading MAC tables from {count} switch{count === 1 ? "" : "es"}…
      </p>
      {Array.from({ length: Math.min(count, 4) }, (_, i) => (
        <Skeleton key={i} className="h-24 w-full" />
      ))}
    </div>
  );
}

function Done({
  result,
  switches,
  onAgain,
}: {
  result: { res: NodeImportResponse; nodes: string[] };
  switches: string;
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
            Imported {result.nodes.length} node{result.nodes.length === 1 ? "" : "s"}
          </h2>
          <p className="text-sm text-muted-foreground">
            {[
              s.create && `${s.create} created`,
              s.add && `${s.add} given a BMC port`,
              s.replace && `${s.replace} had their BMC port replaced`,
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
        <ul className="grid gap-2 sm:grid-cols-3">
          <li>
            <Link
              to="/add/node"
              search={{ tab: "boot", bootStep: "ports", nodes, switches }}
              className="flex h-full gap-3 rounded-lg border p-3 outline-none hover:bg-muted/50 focus-visible:ring-2 focus-visible:ring-ring [&>svg]:mt-0.5 [&>svg]:size-4 [&>svg]:shrink-0 [&>svg]:text-muted-foreground"
            >
              <Network />
              <span className="grid gap-0.5">
                <span className="text-sm font-medium">Add interfaces</span>
                <span className="text-xs text-muted-foreground">
                  Read each node's interfaces from its BMC and add one to Grendel.
                </span>
              </span>
            </Link>
          </li>
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
          Import more
        </Button>
      </div>
    </div>
  );
}
