import { describe, expect, test } from "vitest";
import { relativeTime } from "../utils";

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
    expect(relativeTime(value, now).text).toBe(want);
  });

  test("unparseable value is returned as is", () => {
    expect(relativeTime("20240506123456.000000+000", now).text).toBe("20240506123456.000000+000");
  });
});
