import type { Event } from "@/client";
import { describe, expect, it } from "vitest";
import {
  type EventNode,
  eventParts,
  eventResult,
  eventRows,
  eventTally,
  eventText,
  eventTime,
  nodeNote,
} from "../events";

const noNodes = { total: 0, pending: 0, running: 0, succeeded: 0, failed: 0, stalled: 0 };

function event(
  over: Partial<Omit<Event, "counts">> & { counts?: Partial<Event["counts"]> },
): Event {
  return {
    id: "e1",
    kind: "audit",
    action: "test",
    summary: "",
    actor: "admin",
    state: "done",
    severity: "success",
    started_at: "2026-11-16T12:00:00Z",
    updated_at: "2026-11-16T12:00:00Z",
    ...over,
    counts: { ...noNodes, ...over.counts },
  };
}

const provisioning = event({
  id: "p1",
  kind: "provision",
  summary: "Provisioning",
  actor: "system",
  state: "running",
  severity: "info",
  counts: { total: 9, succeeded: 3, failed: 1, running: 4, stalled: 1 },
  stages: { pxe: 1, booting: 2, config: 1 },
});

describe("eventTally", () => {
  it("counts the nodes not started with those running as still going", () => {
    expect(
      eventTally(event({ counts: { total: 5, pending: 2, running: 1, succeeded: 2 } })),
    ).toEqual({
      total: 5,
      succeeded: 2,
      failed: 0,
      going: 3,
      stalled: 0,
    });
  });

  it("is all zero for an event with no nodes", () => {
    expect(eventTally(event({})).total).toBe(0);
  });
});

describe("eventText", () => {
  it("joins what was done with what it was done to", () => {
    expect(eventText(event({ summary: "Saved nodes", subject: "cpn-[01-02]" }))).toBe(
      "Saved nodes cpn-[01-02]",
    );
    expect(eventText(event({ kind: "bmc", summary: "bmc power On" }))).toBe("bmc power On");
  });

  it("says a provisioning by its nodes, and how many made it once it is over", () => {
    expect(eventText(provisioning)).toBe("Provisioning 9 nodes");
    const done = { ...provisioning, state: "done" as const };
    expect(eventText(done)).toBe("Provisioned 3 of 9 nodes");
    expect(eventText({ ...done, counts: { ...noNodes, total: 1, succeeded: 1 } })).toBe(
      "Provisioned 1 node",
    );
  });
});

describe("eventResult", () => {
  it("says how far a running event has got", () => {
    expect(eventResult(provisioning)).toBe("5 of 9 done");
  });

  it("reports the failures of a finished one first, then the stalled", () => {
    expect(eventResult(event({ counts: { total: 4, succeeded: 3, failed: 1 } }))).toBe(
      "1 of 4 failed",
    );
    expect(eventResult(event({ counts: { total: 4, succeeded: 3, stalled: 1 } }))).toBe(
      "1 of 4 stalled",
    );
    expect(eventResult(event({ counts: { total: 4, succeeded: 4 } }))).toBe("4 of 4 succeeded");
  });

  it("is empty for an event with no nodes", () => {
    expect(eventResult(event({}))).toBe("");
  });
});

describe("eventTime", () => {
  it("is when the event ended, or when it started while it runs", () => {
    expect(eventTime(provisioning)).toBe("2026-11-16T12:00:00Z");
    expect(eventTime({ ...provisioning, ended_at: "2026-11-16T12:09:00Z" })).toBe(
      "2026-11-16T12:09:00Z",
    );
  });
});

describe("eventParts", () => {
  it("orders a provisioning from the furthest along, with the state the nodes list filters on", () => {
    expect(eventParts(provisioning).map((p) => [p.key, p.count, p.state])).toEqual([
      ["succeeded", 3, "completed"],
      ["failed", 1, "failed"],
      ["stage:booting", 2, "booting"],
      ["stage:config", 1, "config"],
      ["stage:pxe", 1, "pxe"],
      ["stalled", 1, undefined],
    ]);
  });

  it("names a provisioning's parts as the provision tab does", () => {
    const labels = eventParts(provisioning).map((p) => p.label);
    expect(labels).toEqual([
      "completed",
      "failed",
      "booting",
      "fetched config",
      "PXE boot",
      "stalled",
    ]);
  });

  it("takes any other event's stages as they come, with what has none as running or waiting", () => {
    const parts = eventParts(
      event({
        kind: "firmware",
        counts: { total: 6, pending: 1, running: 4, succeeded: 1 },
        stages: { scheduled: 1, downloading: 2 },
      }),
    );
    expect(parts.map((p) => [p.label, p.count, p.state])).toEqual([
      ["succeeded", 1, undefined],
      ["downloading", 2, undefined],
      ["scheduled", 1, undefined],
      ["running", 1, undefined],
      ["waiting", 1, undefined],
    ]);
  });
});

describe("eventRows", () => {
  it("makes a row of each event, with how it went across its nodes", () => {
    const [row] = eventRows([
      event({
        kind: "bmc",
        summary: "bmc power on",
        severity: "warning",
        ended_at: "2026-11-14T20:51:29Z",
        detail: { run: "r1" },
        counts: { total: 2, succeeded: 1, failed: 1 },
      }),
    ]);
    expect(row).toMatchObject({
      id: "e1",
      message: "bmc power on",
      result: "1 of 2 failed",
      run: "r1",
      user: "admin",
      time: "2026-11-14T20:51:29Z",
      total: 2,
      failed: 1,
    });
  });
});

// fixtures name only the fields a test is about
const row = (n: Partial<EventNode>) => n as EventNode;

describe("nodeNote", () => {
  it("is the stage a node is at, with what it said", () => {
    expect(nodeNote(row({ node: "cpn-01", state: "running", stage: "booting" }))).toBe("booting");
    expect(
      nodeNote(
        row({ node: "cpn-02", state: "failed", stage: "exception", message: "JID_1 timed out" }),
      ),
    ).toBe("exception: JID_1 timed out");
    expect(nodeNote(row({ node: "cpn-03", state: "succeeded" }))).toBe("");
  });
});
