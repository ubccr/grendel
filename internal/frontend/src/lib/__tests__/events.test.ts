import { describe, expect, it } from "vitest";
import { eventRows, parseRun } from "../events";

describe("parseRun", () => {
  it("splits a bmc action into what was done, how it went and its run", () => {
    expect(parseRun("bmc indicator on: 1 succeeded, 0 failed of 1 (run 3M76vaSkVXQi)")).toEqual({
      message: "bmc indicator on",
      result: "1 of 1 succeeded",
      run: "3M76vaSkVXQi",
      succeeded: 1,
      failed: 0,
      total: 1,
    });
  });

  it("reports the failures when any node failed", () => {
    expect(parseRun("bmc power cycle: 3 succeeded, 1 failed of 4 (run abc)").result).toBe(
      "1 of 4 failed",
    );
  });

  it("leaves any other event as it is", () => {
    expect(parseRun("Successfully deleted node(s): cpn-a[01-02]")).toEqual({
      message: "Successfully deleted node(s): cpn-a[01-02]",
      result: "",
      run: "",
      succeeded: 0,
      failed: 0,
      total: 0,
    });
  });
});

describe("eventRows", () => {
  it("puts a row under an event for each node it reached, carrying the event's fields", () => {
    const [row] = eventRows([
      {
        Severity: "Warning",
        Time: "2026-11-14T20:51:29Z",
        User: "admin",
        Message: "bmc power on: 1 succeeded, 1 failed of 2 (run r1)",
        JobMessages: [
          { host: "cpn-01", status: "success", msg: "" },
          { host: "cpn-02", status: "error", msg: "" },
        ],
      },
    ]);
    expect(row.message).toBe("bmc power on");
    expect(row.items?.map((i) => [i.id, i.host, i.status, i.user])).toEqual([
      ["2026-11-14T20:51:29Z/cpn-01", "cpn-01", "success", "admin"],
      ["2026-11-14T20:51:29Z/cpn-02", "cpn-02", "error", "admin"],
    ]);
  });

  it("gives an event with no nodes no rows under it", () => {
    expect(
      eventRows([{ Severity: "Info", Message: "bmc console opened on cpn-01" }])[0].items,
    ).toEqual([]);
  });
});
