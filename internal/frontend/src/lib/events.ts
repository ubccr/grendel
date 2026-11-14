import type { Event } from "@/client";
import { severityOf, type Severity } from "@/lib/utils";

// EventRow is a row of the events table: an event, with a row under it for each node a bmc action reached. Each node's row carries its event's fields too, so a filter on an event column still finds the nodes under it
export type EventRow = {
  id: string;
  time: string;
  severity: string;
  user: string;
  // message is the event's text, less the run summary a bmc action adds
  message: string;
  result: string;
  run: string;
  // succeeded, failed and total count a bmc action's nodes; total is 0 for any other event
  succeeded: number;
  failed: number;
  total: number;
  host?: string;
  status?: string;
  items?: EventRow[];
};

// a bmc action's event reads "bmc power on: 3 succeeded, 1 failed of 4 (run <id>)", as runBmc in internal/api/bmc.go writes it
const runPattern = /^(.*): (\d+) succeeded, (\d+) failed of (\d+) \(run (\w+)\)$/;

type Run = Pick<EventRow, "message" | "result" | "run" | "succeeded" | "failed" | "total">;

// parseRun splits a bmc action's event into what was done, how it went and its run id. Any other event is all message, with no nodes counted
export function parseRun(message: string): Run {
  const m = runPattern.exec(message);
  if (!m) return { message, result: "", run: "", succeeded: 0, failed: 0, total: 0 };
  const [, action, ok, bad, all, run] = m;
  const [succeeded, failed, total] = [ok, bad, all].map(Number);
  const result = failed === 0 ? `${ok} of ${all} succeeded` : `${bad} of ${all} failed`;
  return { message: action, result, run, succeeded, failed, total };
}

// eventRows turns the api's events, newest first, into the events table's rows
export function eventRows(events: Event[]): EventRow[] {
  return events.map((e, i) => {
    const time = e.Time ?? "";
    const event = {
      id: time || String(i),
      time,
      severity: e.Severity ?? "",
      user: e.User ?? "",
      ...parseRun(e.Message ?? ""),
    };
    const items = (e.JobMessages ?? []).map((j): EventRow => ({
      ...event,
      id: `${event.id}/${j.host}`,
      host: j.host ?? "",
      status: j.status ?? "",
    }));
    return { ...event, items };
  });
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
