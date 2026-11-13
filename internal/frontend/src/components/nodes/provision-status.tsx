import { ProvisionStatus } from "@/client";
import { getV1NodesProvisionStatusOptions } from "@/client/@tanstack/react-query.gen";
import { Badge } from "@/components/ui/badge";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Input } from "@/components/ui/input";
import { ErrorPanel } from "@/components/error-panel";
import { cn } from "@/lib/utils";
import { useQuery } from "@tanstack/react-query";
import { Check, LoaderCircle, Minus, Rocket, X } from "lucide-react";
import { Link } from "@tanstack/react-router";
import { Fragment, useEffect, useState } from "react";

// match model.ProvisionState
const stateLabels: Record<string, string> = {
  pxe: "PXE boot",
  ipxe: "iPXE",
  config: "fetched config",
  booting: "booting",
  failed: "failed",
  completed: "completed",
};

export function isFinished(state?: string) {
  return state === "completed" || state === "failed";
}

export function ProvisionStateBadge({ state }: { state?: string }) {
  if (!state) return null;
  const label = stateLabels[state] ?? state;
  if (state === "failed")
    return (
      <Badge variant="destructive" className="whitespace-nowrap">
        {label}
      </Badge>
    );
  if (state === "completed")
    return <Badge className="bg-emerald-600 whitespace-nowrap hover:bg-emerald-600">{label}</Badge>;
  return (
    <Badge
      variant="outline"
      className="border-sky-500/50 whitespace-nowrap text-sky-600 dark:text-sky-400"
    >
      {label}
    </Badge>
  );
}

// the badge for a node in a list, with when it last changed and the beacon's message on hover
export function ProvisionStatusBadge({ status }: { status?: ProvisionStatus }) {
  if (!status) return null;
  const title = [
    status.updated_at && `updated ${new Date(status.updated_at).toLocaleString()}`,
    status.message,
  ]
    .filter(Boolean)
    .join("\n");
  return (
    <span className="inline-flex" title={title}>
      <ProvisionStateBadge state={status.state} />
    </span>
  );
}

// maps node name to status, for views that list many nodes
export function statusesByNode(statuses: ProvisionStatus[] | undefined) {
  const m = new Map<string, ProvisionStatus>();
  for (const s of statuses ?? []) if (s.node) m.set(s.node, s);
  return m;
}

// how long the provision took when it finished, or has taken so far
export function provisionElapsed(s: ProvisionStatus, now = Date.now()) {
  if (!s.started_at) return "";
  let end = now;
  if (s.completed_at) end = Date.parse(s.completed_at);
  else if (s.state === "failed" && s.updated_at) end = Date.parse(s.updated_at);
  return formatDuration(end - Date.parse(s.started_at));
}

export function formatDuration(ms: number) {
  const total = Math.max(0, Math.round(ms / 1000));
  const h = Math.floor(total / 3600);
  const m = Math.floor((total % 3600) / 60);
  const sec = total % 60;
  if (h > 0) return `${h}h ${m}m`;
  if (m > 0) return `${m}m ${sec}s`;
  return `${sec}s`;
}

function time(t?: string) {
  return t ? new Date(t).toLocaleString() : "";
}

// a relative age like "4s ago", in the browser's clock
export function formatAgo(t: string | undefined, now: number) {
  if (!t) return "";
  const then = Date.parse(t);
  if (isNaN(then)) return "";
  const sec = Math.max(0, Math.round((now - then) / 1000));
  if (sec < 1) return "just now";
  if (sec < 60) return `${sec}s ago`;
  const min = Math.floor(sec / 60);
  if (min < 60) return `${min}m ago`;
  const h = Math.floor(min / 60);
  if (h < 24) return `${h}h ago`;
  return `${Math.floor(h / 24)}d ago`;
}

// re-renders on an interval so relative times stay current
function useNow(ms: number) {
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), ms);
    return () => clearInterval(id);
  }, [ms]);
  return now;
}

// the stages a provision goes through, the last one is completed or failed
const steps = ["pxe", "ipxe", "config", "booting", "done"] as const;
const stepLabels: Record<(typeof steps)[number], string> = {
  pxe: "PXE",
  ipxe: "iPXE",
  config: "Config",
  booting: "Booting",
  done: "Done",
};

export type StepState = "done" | "current" | "pending" | "failed" | "skipped";

// where each step stands for a provision in state. A node completed by the complete endpoint skips straight to the end, so every step before it counts as done
export function stepStates(state?: string, stages?: ProvisionStatus["stages"]): StepState[] {
  let states: StepState[];
  if (state === "completed") states = steps.map(() => "done");
  else if (state === "failed")
    states = steps.map((_, i) => (i === steps.length - 1 ? "failed" : "done"));
  else {
    const cur = steps.indexOf(state as (typeof steps)[number]);
    states = steps.map((_, i) => (i < cur ? "done" : i === cur ? "current" : "pending"));
  }
  // with stage times recorded, a step the node passed without ever reaching was skipped, such as config for a node that never fetched one. Statuses saved before stage times existed can't tell, so they keep every passed step as done
  if (!stages || Object.keys(stages).length === 0) return states;
  return states.map((st, i) =>
    st === "done" && i < steps.length - 1 && !stages[steps[i] as keyof typeof stages]
      ? "skipped"
      : st,
  );
}

// when each step was reached. Statuses saved before stage times were recorded fall back to the start time for the first step reached and the last update for the step the node is on
export function stepTimes(s: ProvisionStatus): (string | undefined)[] {
  const states = stepStates(s.state);
  const stages = s.stages ?? {};
  const legacy = Object.keys(stages).length === 0;
  const first = states.findIndex((st) => st !== "pending");

  return steps.map((step, i) => {
    const st = states[i];
    if (st === "pending") return undefined;
    if (step === "done")
      return (
        s.completed_at ?? stages.completed ?? stages.failed ?? (legacy ? s.updated_at : undefined)
      );
    if (stages[step]) return stages[step];
    if (!legacy) return undefined;
    if (st === "current") return s.updated_at;
    if (i === first) return s.started_at;
    return undefined;
  });
}

// the text under each step: the first step reached shows how long ago it was, every later one how long after that first step it came
export function stepTimeLabels(times: (string | undefined)[], now: number): string[] {
  const first = times.find((t) => t && !isNaN(Date.parse(t)));
  const base = first ? Date.parse(first) : NaN;
  return times.map((t) => {
    if (!t) return "";
    if (t === first) return formatAgo(t, now);
    const at = Date.parse(t);
    if (isNaN(at)) return "";
    return `+${formatDuration(at - base)}`;
  });
}

function StepIcon({ state }: { state: StepState }) {
  const base = "flex size-6 shrink-0 items-center justify-center rounded-full";
  switch (state) {
    case "done":
      return (
        <span className={cn(base, "bg-emerald-600 text-white")}>
          <Check className="size-3.5" strokeWidth={3} />
        </span>
      );
    case "failed":
      return (
        <span className={cn(base, "bg-destructive text-white")}>
          <X className="size-3.5" strokeWidth={3} />
        </span>
      );
    case "current":
      return (
        <span className={cn(base, "border-2 border-sky-500 text-sky-500")}>
          <LoaderCircle className="size-3.5 motion-safe:animate-spin" />
        </span>
      );
    case "skipped":
      return (
        <span
          className={cn(
            base,
            "border-2 border-dashed border-muted-foreground/50 text-muted-foreground",
          )}
        >
          <Minus className="size-3.5" strokeWidth={3} />
        </span>
      );
    default:
      return <span className={cn(base, "border-2 border-border")} />;
  }
}

// a row of equal columns across the tab, and a vertical list on a phone
function Timeline({ status, now }: { status: ProvisionStatus; now: number }) {
  const states = stepStates(status.state, status.stages);
  const times = stepTimes(status);
  const timeLabels = stepTimeLabels(times, now);
  const last = steps.length - 1;

  return (
    <ol className="flex flex-col sm:flex-row" aria-label="Provision stages">
      {steps.map((step, i) => {
        const st = states[i];
        const label =
          i === last && status.state === "completed"
            ? "Completed"
            : i === last && status.state === "failed"
              ? "Failed"
              : stepLabels[step];
        return (
          <li
            key={step}
            className="relative flex gap-3 pb-6 last:pb-0 sm:flex-1 sm:flex-col sm:items-center sm:gap-2 sm:pb-0 sm:text-center"
          >
            {/* the connector runs to the next step, down on a phone and across otherwise, coloured once that step is reached */}
            {i < last && (
              <span
                aria-hidden
                className={cn(
                  "absolute top-8 bottom-2 left-3 w-0.5 -translate-x-1/2 rounded-full sm:top-3 sm:right-[calc(-50%+1.25rem)] sm:bottom-auto sm:left-[calc(50%+1.25rem)] sm:h-0.5 sm:w-auto sm:translate-x-0 sm:-translate-y-1/2",
                  states[i + 1] === "pending" ? "bg-border" : "bg-emerald-600",
                )}
              />
            )}
            <StepIcon state={st} />
            <div className="grid gap-0.5">
              <span
                className={cn(
                  "text-sm",
                  st === "pending" || st === "skipped" ? "text-muted-foreground" : "font-medium",
                  st === "failed" && "text-destructive",
                )}
                aria-current={st === "current" ? "step" : undefined}
              >
                {label}
              </span>
              {times[i] && (
                <time
                  dateTime={times[i]}
                  title={time(times[i])}
                  className="text-xs text-muted-foreground tabular-nums"
                >
                  {timeLabels[i]}
                </time>
              )}
              {st === "skipped" && <span className="text-xs text-muted-foreground">skipped</span>}
            </div>
          </li>
        );
      })}
    </ol>
  );
}

// what the node is doing in the stage it reached, read as "<doing> for 3m 2s"
const doing: Record<string, string> = {
  pxe: "Network booting",
  ipxe: "Loading the boot image",
  config: "Configuring the OS",
  booting: "Starting services",
};

function headline(s: ProvisionStatus, now: number) {
  const took = provisionElapsed(s, now);
  if (s.state === "completed") return `Completed in ${took}`;
  if (s.state === "failed") return `Failed after ${took}`;
  return `${doing[s.state ?? ""] ?? s.state} for ${took}`;
}

// match model.FactImageVersion
const factImageVersion = "os.image_version";

// what the node reports running against the image it was sent. Only a version both sides know can disagree
export function imageCheck(s: ProvisionStatus): {
  running?: string;
  sent?: string;
  mismatch: boolean;
} {
  const running = s.facts?.[factImageVersion];
  const sent = s.boot_image_version;
  return { running, sent, mismatch: !!running && !!sent && running !== sent };
}

function imageText(s: ProvisionStatus) {
  const { running, sent, mismatch } = imageCheck(s);
  const image = s.boot_image ? `${s.boot_image} ${sent ?? ""}`.trim() : sent;
  if (mismatch) return `Running ${running}, but it was sent ${image}.`;
  if (running) return `Running ${s.boot_image ? `${s.boot_image} ` : ""}${running}.`;
  return "";
}

// facts grouped by the part of the key before the first dot, groups and keys in order
export function groupFacts(facts: Record<string, string> | undefined, filter = "") {
  const f = filter.trim().toLowerCase();
  const groups = new Map<string, [string, string][]>();
  for (const key of Object.keys(facts ?? {}).sort()) {
    const value = facts![key];
    if (f && !key.includes(f) && !value.toLowerCase().includes(f)) continue;
    const dot = key.indexOf(".");
    const group = dot < 0 ? key : key.slice(0, dot);
    const name = dot < 0 ? key : key.slice(dot + 1);
    if (!groups.has(group)) groups.set(group, []);
    groups.get(group)!.push([name, value]);
  }
  return [...groups.entries()];
}

function Facts({ facts }: { facts: Record<string, string> }) {
  const [filter, setFilter] = useState("");
  const total = Object.keys(facts).length;
  const groups = groupFacts(facts, filter);

  return (
    <section className="grid gap-3">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h3 className="text-sm font-medium">Facts ({total})</h3>
        {total > 12 && (
          <Input
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            placeholder="Filter facts"
            aria-label="Filter facts"
            className="h-8 w-full sm:w-64"
          />
        )}
      </div>
      {groups.length === 0 ? (
        <p className="text-sm text-muted-foreground">No facts match “{filter}”.</p>
      ) : (
        <div className="grid gap-x-8 gap-y-6 sm:grid-cols-2 xl:grid-cols-3">
          {groups.map(([group, entries]) => (
            <div key={group} className="min-w-0">
              <h4 className="mb-1 border-b pb-1 font-mono text-sm font-medium">{group}</h4>
              <dl className="grid grid-cols-[minmax(0,auto)_minmax(0,1fr)] gap-x-4 gap-y-1 text-sm">
                {entries.map(([name, value]) => (
                  <Fragment key={name}>
                    <dt className="truncate text-muted-foreground" title={`${group}.${name}`}>
                      {name}
                    </dt>
                    <dd className="break-words">{value}</dd>
                  </Fragment>
                ))}
              </dl>
            </div>
          ))}
        </div>
      )}
    </section>
  );
}

function subtext(s: ProvisionStatus) {
  const parts: string[] = [];
  if (s.addr) parts.push(`Reported from ${s.addr}.`);
  if (s.state === "booting" && s.queued_jobs) parts.push(`${s.queued_jobs} jobs still queued.`);
  return parts.join(" ");
}

function FailedUnits({ units }: { units: NonNullable<ProvisionStatus["failed_units"]> }) {
  return (
    <section className="grid gap-2">
      <h3 className="text-sm font-medium">Failed units ({units.length})</h3>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Unit</TableHead>
            <TableHead className="hidden md:table-cell">Description</TableHead>
            <TableHead>State</TableHead>
            <TableHead>Result</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {units.map((u) => (
            <TableRow key={u.name}>
              <TableCell className="font-mono break-all">{u.name}</TableCell>
              <TableCell className="hidden text-muted-foreground md:table-cell">
                {u.description}
              </TableCell>
              <TableCell className="text-muted-foreground">{u.sub_state}</TableCell>
              <TableCell className="text-destructive">{u.result}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </section>
  );
}

export function ProvisionStatusPanel({ node }: { node: string }) {
  const { data, isLoading, error } = useQuery({
    ...getV1NodesProvisionStatusOptions({ query: { nodeset: node } }),
    // keep following a provision that is still going
    refetchInterval: (q) => (isFinished(q.state.data?.[0]?.state) ? false : 15_000),
  });
  const now = useNow(1000);
  const s = data?.[0];

  if (isLoading) return null;
  if (error)
    return (
      <ErrorPanel
        title="Could not read the provision status"
        message={error.detail || error.title}
        icon={Rocket}
      />
    );
  if (!s) {
    return (
      <Empty className="my-4">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <Rocket />
          </EmptyMedia>
          <EmptyTitle>No provision recorded</EmptyTitle>
          <EmptyDescription>
            Status is tracked from the node's next network boot. Run grendel-beacon in the node's OS
            to see how it comes up.
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    );
  }

  const failed = s.failed_units ?? [];
  const sub = subtext(s);
  const image = imageText(s);
  const { mismatch } = imageCheck(s);

  return (
    <div className="my-6 grid gap-8">
      <header className="grid gap-1">
        <h2
          className="text-xl font-semibold tracking-tight"
          title={s.boot_id ? `boot ${s.boot_id}` : undefined}
        >
          {s.node ?? node}
        </h2>
        <p className="text-sm text-muted-foreground">
          {headline(s, now)}.{sub && ` ${sub}`}
        </p>
        {image && (
          <p className={cn("text-sm", mismatch ? "text-destructive" : "text-muted-foreground")}>
            {image}
          </p>
        )}
      </header>

      <Timeline status={s} now={now} />

      {s.message &&
        (s.state === "failed" ? (
          <pre className="rounded-md bg-muted px-4 py-3 font-mono text-sm break-words whitespace-pre-wrap">
            {s.message}
          </pre>
        ) : (
          <p className="text-sm text-muted-foreground">{s.message}</p>
        ))}

      {failed.length > 0 && <FailedUnits units={failed} />}

      {s.facts && Object.keys(s.facts).length > 0 && <Facts facts={s.facts} />}
    </div>
  );
}

// ProvisionSummary is a node's last provision in brief, for a popover in a view of many nodes: where it got to, why it failed, and a way to the full tab
export function ProvisionSummary({ status: s, node }: { status: ProvisionStatus; node: string }) {
  const now = useNow(1000);
  const states = stepStates(s.state, s.stages);
  const failed = s.failed_units ?? [];
  const image = imageText(s);
  const { mismatch } = imageCheck(s);
  const shown = failed.slice(0, 5);

  return (
    <div className="grid gap-3 text-sm">
      <div className="grid gap-0.5">
        <p className="font-medium">{headline(s, now)}</p>
        {s.updated_at && (
          <p className="text-xs text-muted-foreground" title={time(s.updated_at)}>
            Last report {formatAgo(s.updated_at, now)}
            {s.addr && ` from ${s.addr}`}
          </p>
        )}
      </div>
      <ol className="flex items-center gap-1" aria-label="Provision stages">
        {steps.map((step, i) => (
          <li key={step} className="flex items-center gap-1" title={stepLabels[step]}>
            {i > 0 && (
              <span
                aria-hidden
                className={cn(
                  "h-0.5 w-3 rounded-full",
                  states[i] === "pending" ? "bg-border" : "bg-emerald-600",
                )}
              />
            )}
            <span className="inline-flex scale-75">
              <StepIcon state={states[i]} />
            </span>
            <span className="sr-only">
              {stepLabels[step]}: {states[i]}
            </span>
          </li>
        ))}
      </ol>
      {image && (
        <p className={cn("text-xs", mismatch ? "text-destructive" : "text-muted-foreground")}>
          {image}
        </p>
      )}
      {s.message &&
        (s.state === "failed" ? (
          <pre className="max-h-40 overflow-y-auto rounded-md bg-muted px-3 py-2 font-mono text-xs break-words whitespace-pre-wrap">
            {s.message}
          </pre>
        ) : (
          <p className="text-muted-foreground">{s.message}</p>
        ))}
      {shown.length > 0 && (
        <div className="grid gap-1">
          <p className="text-xs font-medium">Failed units ({failed.length})</p>
          <ul className="grid gap-0.5">
            {shown.map((u) => (
              <li key={u.name} className="flex justify-between gap-3 text-xs">
                <span className="truncate font-mono" title={u.description}>
                  {u.name}
                </span>
                <span className="shrink-0 text-destructive">{u.result}</span>
              </li>
            ))}
          </ul>
          {failed.length > shown.length && (
            <p className="text-xs text-muted-foreground">and {failed.length - shown.length} more</p>
          )}
        </div>
      )}
      <Link
        to="/nodes/$node/provision"
        params={{ node }}
        className="text-sm font-medium underline-offset-4 hover:underline"
      >
        Open the provision tab
      </Link>
    </div>
  );
}
