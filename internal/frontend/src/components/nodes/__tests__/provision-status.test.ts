import {
  groupFacts,
  imageCheck,
  formatDuration,
  isFinished,
  provisionElapsed,
  stepStates,
  stepTimeLabels,
  stepTimes,
} from "@/components/nodes/provision-status";
import { describe, expect, it } from "vitest";

describe("provisionElapsed", () => {
  const started = "2026-11-11T10:00:00Z";

  it("runs to now while the provision is going", () => {
    const now = Date.parse("2026-11-11T10:02:05Z");
    expect(provisionElapsed({ state: "booting", started_at: started }, now)).toBe("2m 5s");
  });

  it("stops at completion", () => {
    const s = { state: "completed", started_at: started, completed_at: "2026-11-11T11:30:00Z" };
    expect(provisionElapsed(s, Date.parse("2026-11-12T00:00:00Z"))).toBe("1h 30m");
  });

  it("stops at the failure report", () => {
    const s = { state: "failed", started_at: started, updated_at: "2026-11-11T10:00:42Z" };
    expect(provisionElapsed(s, Date.parse("2026-11-12T00:00:00Z"))).toBe("42s");
  });

  it("is empty without a start", () => {
    expect(provisionElapsed({ state: "pxe" })).toBe("");
  });
});

describe("formatDuration", () => {
  it("never goes negative", () => {
    expect(formatDuration(-5000)).toBe("0s");
  });
});

describe("isFinished", () => {
  it("knows the end states", () => {
    expect(isFinished("completed")).toBe(true);
    expect(isFinished("failed")).toBe(true);
    expect(isFinished("booting")).toBe(false);
    expect(isFinished(undefined)).toBe(false);
  });
});

describe("stepStates", () => {
  const all = { pxe: "t1", ipxe: "t2", config: "t3", booting: "t4" };

  it("marks the steps before the current one done", () => {
    expect(stepStates({ state: "config", stages: all })).toEqual([
      "done",
      "done",
      "current",
      "pending",
      "pending",
    ]);
    expect(stepStates({ state: "pxe", stages: { pxe: "t1" } })).toEqual([
      "current",
      "pending",
      "pending",
      "pending",
      "pending",
    ]);
  });

  it("finishes every step on completion", () => {
    expect(stepStates({ state: "completed", stages: { ...all, completed: "t5" } })).toEqual([
      "done",
      "done",
      "done",
      "done",
      "done",
    ]);
  });

  it("fails the last step", () => {
    expect(stepStates({ state: "failed", stages: { ...all, failed: "t5" } })).toEqual([
      "done",
      "done",
      "done",
      "done",
      "failed",
    ]);
  });

  it("marks passed steps without a stage time as skipped", () => {
    const stages = { pxe: "t1", ipxe: "t2", completed: "t3" };
    expect(stepStates({ state: "completed", stages })).toEqual([
      "done",
      "done",
      "skipped",
      "skipped",
      "done",
    ]);
    expect(
      stepStates({ state: "booting", stages: { pxe: "t1", ipxe: "t2", booting: "t3" } }),
    ).toEqual(["done", "done", "skipped", "current", "pending"]);
  });

  it("skips every step for a node the complete endpoint finished", () => {
    expect(stepStates({ state: "completed", stages: { completed: "t1" } })).toEqual([
      "skipped",
      "skipped",
      "skipped",
      "skipped",
      "done",
    ]);
  });

  it("has nothing reached before a provision is recorded", () => {
    expect(stepStates({})).toEqual(["pending", "pending", "pending", "pending", "pending"]);
  });
});

describe("stepTimes", () => {
  it("uses the recorded stage times", () => {
    const s = {
      state: "booting",
      stages: { pxe: "t1", ipxe: "t2", config: "t3", booting: "t4" },
    };
    expect(stepTimes(s)).toEqual(["t1", "t2", "t3", "t4", undefined]);
  });

  it("leaves a stage the node skipped empty", () => {
    const s = { state: "config", stages: { ipxe: "t2", config: "t3" } };
    expect(stepTimes(s)).toEqual([undefined, "t2", "t3", undefined, undefined]);
  });

  it("puts the finish under the last step", () => {
    const s = { state: "failed", stages: { booting: "t4", failed: "t5" } };
    expect(stepTimes(s)[4]).toBe("t5");
    const c = { state: "completed", completed_at: "t6", stages: { completed: "t5" } };
    expect(stepTimes(c)[4]).toBe("t6");
  });

  it("has no finish time while the provision is going, even if an earlier one finished", () => {
    const s = { state: "booting", stages: { booting: "t4", completed: "t0" } };
    expect(stepTimes(s)[4]).toBeUndefined();
  });
});

describe("imageCheck", () => {
  it("flags a node running another version than it was sent", () => {
    const s = { boot_image_version: "2026.11.3", facts: { "os.image_version": "2026.10.1" } };
    expect(imageCheck(s)).toEqual({ running: "2026.10.1", sent: "2026.11.3", mismatch: true });
  });
  it("needs both versions to disagree", () => {
    expect(imageCheck({ facts: { "os.image_version": "2026.10.1" } }).mismatch).toBe(false);
    expect(imageCheck({ boot_image_version: "2026.11.3" }).mismatch).toBe(false);
    expect(
      imageCheck({ boot_image_version: "1", facts: { "os.image_version": "1" } }).mismatch,
    ).toBe(false);
  });
});

describe("groupFacts", () => {
  const facts = {
    "os.id": "rocky",
    "kernel.release": "5.14",
    "os.image_version": "2026",
    lone: "x",
  };
  it("groups by the first part of the key, in order", () => {
    expect(groupFacts(facts)).toEqual([
      ["kernel", [["release", "5.14"]]],
      ["lone", [["lone", "x"]]],
      [
        "os",
        [
          ["id", "rocky"],
          ["image_version", "2026"],
        ],
      ],
    ]);
  });
  it("filters on keys and values", () => {
    expect(groupFacts(facts, "ROCKY")).toEqual([["os", [["id", "rocky"]]]]);
    expect(groupFacts(facts, "kernel.")).toEqual([["kernel", [["release", "5.14"]]]]);
  });
});

describe("stepTimeLabels", () => {
  const now = Date.parse("2026-11-11T10:00:00Z");
  it("shows the first step as an age and later steps as offsets from it", () => {
    const times = [
      "2026-11-11T09:50:00Z",
      "2026-11-11T09:50:12Z",
      "2026-11-11T09:52:30Z",
      undefined,
      undefined,
    ];
    expect(stepTimeLabels(times, now)).toEqual(["10m ago", "+12s", "+2m 30s", "", ""]);
  });
  it("uses the first step with a time as the base", () => {
    const times = [undefined, "2026-11-11T09:59:00Z", "2026-11-11T10:00:00Z"];
    expect(stepTimeLabels(times, now)).toEqual(["", "1m ago", "+1m 0s"]);
  });
});
