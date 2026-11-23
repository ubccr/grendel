import { ProvisionStatus } from "@/client";
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
import { isFinished, provisionState, runningStates, stateLabel } from "@/lib/provision";
import { provisionStatusOptions } from "@/lib/queries";
import { cn, relativeTime } from "@/lib/utils";
import { useQuery } from "@tanstack/react-query";
import { Check, LoaderCircle, Minus, Rocket, X } from "lucide-react";
import { Link } from "@tanstack/react-router";
import { useNow } from "@/hooks/use-now";
import { Fragment, useState } from "react";

export { isFinished };

// LastProvisionCounts sums up how the last provisions of a group of nodes went, failures first, as "2 failed, 1 in progress, 18 completed"
export function LastProvisionCounts({ statuses }: { statuses: ProvisionStatus[] }) {
  const failed = statuses.filter((s) => s.state === "failed").length;
  const running = statuses.filter((s) => s.state && !isFinished(s.state)).length;
  const done = statuses.filter((s) => s.state === "completed").length;
  return (
    <>
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
    </>
  );
}

export function ProvisionStateBadge({ state }: { state?: string }) {
  if (!state) return null;
  const label = stateLabel(state);
  if (state === "failed")
    return (
      <Badge variant="destructive" className="whitespace-nowrap">
        {label}
      </Badge>
    );
  // tinted like the destructive badge, so the three outcomes read as one set
  if (state === "completed")
    return (
      <Badge className="bg-emerald-600/10 whitespace-nowrap text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-400">
        {label}
      </Badge>
    );
  return (
    <Badge className="bg-sky-600/10 whitespace-nowrap text-sky-700 dark:bg-sky-500/20 dark:text-sky-400">
      {label}
    </Badge>
  );
}

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

// how long ago a report was made. The server's clock can run ahead of the browser's, so a time in the future counts as now
const ago = (t: string, now: number) =>
  relativeTime(t, { now: Math.max(now, Date.parse(t) || 0), narrow: true }).text;

// the steps of the timeline: each state a provision is still going in, then one for however it finished
const steps = [...runningStates.map((s) => s.state), "done"] as const;
const stepLabels = [...runningStates.map((s) => s.short), "Done"];
const last = steps.length - 1;

export type StepState = "done" | "current" | "pending" | "failed" | "skipped";

// where each step stands for a provision. A step the node passed with no time recorded for it was skipped, such as config for a node that never fetched one, or every step before the end for a node completed by the complete endpoint
export function stepStates({ state, stages }: Pick<ProvisionStatus, "state" | "stages">) {
  const reached = isFinished(state) ? last : steps.indexOf(state as (typeof steps)[number]);
  return steps.map((step, i): StepState => {
    if (i > reached) return "pending";
    if (i === last) return state === "failed" ? "failed" : "done";
    if (i === reached) return "current";
    return stages?.[step as keyof typeof stages] ? "done" : "skipped";
  });
}

// when each step was reached, for the steps that were
export function stepTimes(s: ProvisionStatus): (string | undefined)[] {
  const states = stepStates(s);
  const stages = s.stages ?? {};
  return steps.map((step, i) => {
    if (states[i] === "pending") return undefined;
    if (step === "done") return s.completed_at ?? stages.completed ?? stages.failed;
    return stages[step];
  });
}

// the text under each step: the first step reached shows how long ago it was, every later one how long after that first step it came
export function stepTimeLabels(times: (string | undefined)[], now: number): string[] {
  const first = times.find((t) => t && !isNaN(Date.parse(t)));
  const base = first ? Date.parse(first) : NaN;
  return times.map((t) => {
    if (!t) return "";
    if (t === first) return ago(t, now);
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
  const states = stepStates(status);
  const times = stepTimes(status);
  const timeLabels = stepTimeLabels(times, now);

  return (
    <ol className="flex flex-col sm:flex-row" aria-label="Provision stages">
      {steps.map((step, i) => {
        const st = states[i];
        const label =
          i === last && isFinished(status.state)
            ? provisionState(status.state)?.short
            : stepLabels[i];
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

function headline(s: ProvisionStatus, now: number) {
  const took = provisionElapsed(s, now);
  if (s.state === "completed") return `Completed in ${took}`;
  if (s.state === "failed") return `Failed after ${took}`;
  return `${provisionState(s.state)?.doing ?? s.state} for ${took}`;
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

// a provision still going is followed closely. One that is over, or none yet, is still looked at now and then, so the next one is noticed without a refresh
const FOLLOW_MS = 15_000;
const WATCH_MS = 60_000;

// Each caller runs its own timer, so one place on a page follows the provision and the rest read what it fetched
export function useFollowedProvision(node: string) {
  return useQuery({
    ...provisionStatusOptions({ nodeset: node }),
    refetchInterval: (q) => {
      const state = q.state.data?.[0]?.state;
      return state && !isFinished(state) ? FOLLOW_MS : WATCH_MS;
    },
  });
}

export function ProvisionStatusPanel({ node }: { node: string }) {
  // kept up to date by the node's header, which is on every tab and follows the provision
  const { data, isLoading, error } = useQuery(provisionStatusOptions({ nodeset: node }));
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

export function ProvisionSummary({ status: s, node }: { status: ProvisionStatus; node: string }) {
  const now = useNow(1000);
  const states = stepStates(s);
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
            Last report {ago(s.updated_at, now)}
            {s.addr && ` from ${s.addr}`}
          </p>
        )}
      </div>
      <ol className="flex items-center gap-1" aria-label="Provision stages">
        {steps.map((step, i) => (
          <li key={step} className="flex items-center gap-1" title={stepLabels[i]}>
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
              {stepLabels[i]}: {states[i]}
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
