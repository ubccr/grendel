import { describe, expect, test } from "vitest";
import { relativeTime, severityOf, severityVariant } from "../utils";

describe("relativeTime", () => {
  const now = Date.parse("2026-11-10T12:00:00Z");

  test.each([
    ["2026-11-10T11:55:00Z", "5 minutes ago"],
    ["2026-11-07T12:00:00Z", "3 days ago"],
    ["2026-11-10T11:59:58Z", "2 seconds ago"],
    ["2026-11-10T12:00:00Z", "now"],
    ["2026-11-09T12:00:00Z", "yesterday"],
    ["2025-11-10T12:00:00Z", "last year"],
  ])("%s", (value, want) => {
    expect(relativeTime(value, { now }).text).toBe(want);
  });

  test.each([
    ["2026-11-10T11:55:00Z", "5m ago"],
    ["2026-11-10T11:59:58Z", "2s ago"],
    ["2026-11-10T10:00:00Z", "2h ago"],
  ])("%s narrow", (value, want) => {
    expect(relativeTime(value, { now, narrow: true }).text).toBe(want);
  });

  test("unparseable value is returned as is", () => {
    expect(relativeTime("20240506123456.000000+000", { now }).text).toBe(
      "20240506123456.000000+000",
    );
  });
});

describe("severityOf", () => {
  test.each([
    ["Success", "success"],
    ["success", "success"],
    ["OK", "success"],
    ["Info", "info"],
    ["Warning", "warning"],
    ["Error", "error"],
    ["error", "error"],
    ["Critical", "error"],
    ["", undefined],
    [undefined, undefined],
    ["Enabled", undefined],
  ])("%s", (value, want) => {
    expect(severityOf(value)).toBe(want);
  });

  test("an unknown severity gets the plain badge", () => {
    expect(severityVariant("Unknown")).toBe("secondary");
    expect(severityVariant("Critical")).toBe("destructive");
  });
});
