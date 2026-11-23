import type { NodeImportResponse, SwitchScanResponse } from "@/client";
import { getV1SwitchScan } from "@/client";
import { getV1SwitchScanOptions } from "@/client/@tanstack/react-query.gen";
import { forgetRead } from "@/components/import/boot-import";
import { Done, NextStep, nextStepLink } from "@/components/import/done";
import { Naming } from "@/components/import/naming";
import { PortGrid } from "@/components/import/port-grid";
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
import { getRouteApi, Link } from "@tanstack/react-router";
import { ArrowLeft, ArrowRight, LoaderCircle, Network, Radar, RefreshCw } from "lucide-react";
import { useMemo, useState } from "react";
import { useRefreshPart } from "./use-refresh-part";

const settingsKey = "grendel.import.settings";

export const importSteps = ["switches", "map", "review", "done"] as const;
export type ImportStep = (typeof importSteps)[number];

const route = getRouteApi("/add/node");

// The step and switches are the add node page's search parameters, so Back and Forward move between steps
export function SwitchImport() {
  const { tab, step: wanted, switches } = route.useSearch();
  const navigate = route.useNavigate();
  const queryClient = useQueryClient();

  const [stored, store] = useStoredJSON(settingsKey, defaultSettings);
  // the Port filter belongs to this import alone and is never stored, since one left from an earlier import would silently leave ports out of the next
  const [ports, setPorts] = useState("");
  const settings = useMemo(() => ({ ...stored, ports }), [stored, ports]);
  const setSettings = (s: Settings) => store({ ...s, ports: "" });

  const [overrides, setOverrides] = useState<Overrides>({});
  const [result, setResult] = useState<{ res: NodeImportResponse; nodes: string[] }>();

  const selected = useMemo(() => expandOrEmpty(switches), [switches]);

  // a step that needs what an earlier one made falls back to it, as after a reload mid-import
  const step: ImportStep =
    wanted !== "switches" && selected.length === 0
      ? "switches"
      : wanted === "done" && !result
        ? "review"
        : wanted;
  const go = (to: ImportStep, extra?: { switches?: string }) =>
    navigate({ search: (s) => ({ ...s, ...extra, step: to }) });

  const scanOptions = getV1SwitchScanOptions({ query: { nodeset: switches } });
  const scan = useQuery({
    ...scanOptions,
    // the tab stays mounted behind the others, where a link changing its search parameters must not start a scan nobody is looking at
    enabled: tab === "switches" && step !== "switches" && !!switches,
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

  const summary = result?.res.summary ?? {};
  const imported = fold(result?.nodes ?? []);

  const stepList = [
    { id: "switches", label: "Switches" },
    { id: "map", label: "Ports and names" },
    { id: "review", label: "Addresses and review" },
  ];

  return (
    <>
      {step !== "done" && (
        // pulled up by the difference between the panel's gap and the gap under the tab list, so the steps sit midway between the tabs and the step itself
        <div className="-mb-2 justify-self-center">
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
        </div>
      )}

      {step === "switches" && (
        <SwitchPicker
          selected={selected}
          onChange={(names) =>
            navigate({ search: (s) => ({ ...s, switches: fold(names) }), replace: true })
          }
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
              rescanning={rescanning}
              scanning={scan.isFetching}
              onOverrides={setOverrides}
              onPorts={setPorts}
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
          needsAddresses={t.ready > 0}
          back={
            <Button variant="ghost" onClick={() => go("map")}>
              <ArrowLeft />
              Ports and names
            </Button>
          }
          onCommitted={(res, nodes) => {
            setResult({ res, nodes });
            go("done");
          }}
        />
      )}

      {step === "done" && result && (
        <Done
          title={`Imported ${result.nodes.length} node${result.nodes.length === 1 ? "" : "s"}`}
          summary={[
            summary.create && `${summary.create} created`,
            summary.add && `${summary.add} given a BMC port`,
            summary.replace && `${summary.replace} had their BMC port replaced`,
            summary.unchanged && `${summary.unchanged} already as shown`,
          ]}
          nodes={imported}
          first={
            <Link
              to="/add/node"
              search={(s) => ({
                ...s,
                tab: "boot",
                bootStep: "ports",
                nodes: imported,
                bootSwitches: switches,
              })}
              className={nextStepLink}
              onClick={() => forgetRead(queryClient, imported, switches)}
            >
              <NextStep
                icon={<Network />}
                title="Add interfaces"
                text="Read each node's interfaces from its BMC and add one to Grendel."
              />
            </Link>
          }
          again="Import more"
          onAgain={() => {
            setOverrides({});
            setResult(undefined);
            go("switches", { switches: "" });
          }}
        />
      )}
    </>
  );
}

function Scanning({ count }: { count: number }) {
  return (
    <div className="grid gap-3" aria-busy>
      <p className="flex items-center gap-2 text-sm text-muted-foreground">
        <LoaderCircle className="size-4 animate-spin" />
        Reading MAC tables and LLDP neighbors from {count} switch{count === 1 ? "" : "es"}…
      </p>
      {Array.from({ length: Math.min(count, 4) }, (_, i) => (
        <Skeleton key={i} className="h-24 w-full" />
      ))}
    </div>
  );
}
