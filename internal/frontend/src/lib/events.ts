import type { Event } from "@/client";
import { runningStates, stateLabel } from "@/lib/provision";
import { severityOf, type Severity } from "@/lib/utils";

// Tally is an event's nodes by how it went on them. going are the nodes it is not over for, whether it has got to them yet or not
export type Tally = {
  total: number;
  succeeded: number;
  failed: number;
  going: number;
  stalled: number;
};

export function eventTally({ counts: c }: Pick<Event, "counts">): Tally {
  return {
    total: c.total,
    succeeded: c.succeeded,
    failed: c.failed,
    going: c.pending + c.running,
    stalled: c.stalled,
  };
}

// a tracked event is written while it is running and updated as its nodes report in
export function isRunning(e?: Pick<Event, "state">) {
  return e?.state === "running";
}

// eventTime is when an event counts as having happened: when it ended, or when it started while it is running
export function eventTime(e: Pick<Event, "ended_at" | "started_at">) {
  return e.ended_at ?? e.started_at;
}

// newestFirst orders events by when each counts as having happened, so one that ran for a while moves to the top when it ends. The API lists them by when they started
export function newestFirst<T extends Pick<Event, "ended_at" | "started_at">>(events: T[]): T[] {
  const at = (e: T) => Date.parse(eventTime(e)) || 0;
  return [...events].sort((a, b) => at(b) - at(a));
}

const nodes = (n: number) => `${n} ${n === 1 ? "node" : "nodes"}`;

// eventText says what an event did and to what. A provisioning is nodes arriving on their own, so it is said by how many there are, and once it is over by how many made it
export function eventText(e: Pick<Event, "kind" | "state" | "counts" | "summary" | "subject">) {
  if (e.kind === "provision") {
    const t = eventTally(e);
    if (isRunning(e)) return `Provisioning ${nodes(t.total)}`;
    if (t.succeeded === t.total) return `Provisioned ${nodes(t.total)}`;
    return `Provisioned ${t.succeeded} of ${nodes(t.total)}`;
  }
  return [e.summary, e.subject].filter(Boolean).join(" ");
}

// eventResult says how an event went across its nodes, such as "1 of 6 failed". An event with no nodes has none
export function eventResult(e: Pick<Event, "counts">) {
  const t = eventTally(e);
  if (t.total === 0) return "";
  if (t.going > 0) return `${t.succeeded + t.failed + t.stalled} of ${t.total} done`;
  if (t.failed > 0) return `${t.failed} of ${t.total} failed`;
  if (t.stalled > 0) return `${t.stalled} of ${t.total} stalled`;
  return `${t.succeeded} of ${t.total} succeeded`;
}

// EventPart is the nodes of an event at one point, with the colour that stands for it. state is the provision state the nodes list filters on, for the parts that have one
export type EventPart = {
  key: string;
  label: string;
  count: number;
  color: string;
  state?: string;
};

// a fainter blue the further a stage is from done, for up to four stages
const stageColors = [
  "bg-(--node-blue)",
  "bg-(--node-blue)/75",
  "bg-(--node-blue)/55",
  "bg-(--node-blue)/35",
];

// the stages of a provision, the furthest along first
const provisionStages: string[] = runningStates.map((s) => s.state).reverse();

// eventParts are an event's nodes by how far they got, the furthest first, so drawn in order they read as progress filling from the left: done, failed, each stage still running, not started, stalled. Parts with no nodes are left out
export function eventParts(e: Pick<Event, "kind" | "counts" | "stages">): EventPart[] {
  const c = e.counts;
  const provision = e.kind === "provision";
  const stages = e.stages ?? {};
  const names = provision ? provisionStages.filter((s) => s in stages) : Object.keys(stages).sort();
  const staged = names.reduce((n, s) => n + stages[s], 0);

  return [
    {
      key: "succeeded",
      label: provision ? stateLabel("completed") : "succeeded",
      count: c.succeeded,
      color: "bg-(--node-ready)",
      state: provision ? "completed" : undefined,
    },
    {
      key: "failed",
      label: "failed",
      count: c.failed,
      color: "bg-destructive",
      state: provision ? "failed" : undefined,
    },
    ...names.map((s, i) => ({
      key: `stage:${s}`,
      label: provision ? stateLabel(s) : s,
      count: stages[s],
      color: stageColors[Math.min(i, stageColors.length - 1)],
      state: provision ? s : undefined,
    })),
    {
      key: "running",
      label: "running",
      count: c.running - staged,
      color: "bg-(--node-blue)",
    },
    { key: "pending", label: "waiting", count: c.pending, color: "bg-muted-foreground/40" },
    { key: "stalled", label: "stalled", count: c.stalled, color: "bg-(--node-amber)" },
  ].filter((p) => p.count > 0);
}

export type EventRow = Tally & {
  id: string;
  time: string;
  // updated is when the event or any of its nodes last changed
  updated: string;
  kind: string;
  state: string;
  severity: string;
  user: string;
  message: string;
  result: string;
  run: string;
};

export function eventRows(events: Event[]): EventRow[] {
  return events.map((e) => ({
    id: e.id,
    time: eventTime(e),
    updated: e.updated_at,
    kind: e.kind,
    state: e.state,
    severity: e.severity,
    user: e.actor,
    message: eventText(e),
    result: eventResult(e),
    run: e.detail?.run ?? "",
    ...eventTally(e),
  }));
}

// EventNode is how an event went on one node, which the client inlines rather than naming
export type EventNode = NonNullable<Event["nodes"]>[number];

// nodeNote is what a node's row says beside its state: the stage it is at while it runs, or what it had to say once it finished
export function nodeNote(n: EventNode) {
  return [n.stage, n.message].filter(Boolean).join(": ");
}

export function nodeSeverity(state?: string): Severity | undefined {
  switch (state) {
    case "succeeded":
      return "success";
    case "failed":
      return "error";
    case "stalled":
      return "warning";
    case "running":
    case "pending":
      return "info";
  }
}

// severityDot is the colour of an event's severity in the node field's palette, for the dots and cells that stand for an event
export function severityDot(value?: string) {
  const severity = severityOf(value);
  return severity ? dots[severity] : "bg-muted-foreground/40";
}

const dots: Record<Severity, string> = {
  success: "bg-(--node-ready)",
  info: "bg-(--node-blue)",
  warning: "bg-(--node-amber)",
  error: "bg-destructive",
};
