import { buildSummary, releaseUrl } from "@/lib/build";
import { describe, expect, it } from "vitest";

describe("releaseUrl", () => {
  it("links a release with or without the leading v", () => {
    const url = "https://github.com/ubccr/grendel/releases/tag/v0.3.0";
    expect(releaseUrl("0.3.0")).toBe(url);
    expect(releaseUrl("v0.3.0")).toBe(url);
  });

  it("has no page for a build that is not a release", () => {
    expect(releaseUrl("vDEV")).toBeUndefined();
    expect(releaseUrl("0.3.0-next")).toBeUndefined();
    expect(releaseUrl(undefined)).toBeUndefined();
  });
});

describe("buildSummary", () => {
  it("lists what the server reported", () => {
    expect(
      buildSummary({
        Version: "0.3.0",
        commit: "2233d8d0123456789abcdef",
        modified: true,
        go_version: "go1.25.0",
        schema: 20261123120000,
      }),
    ).toBe("Grendel 0.3.0, commit 2233d8d01234+changes, go1.25.0, schema 20261123120000");
  });

  it("leaves out what a build outside git does not know", () => {
    expect(buildSummary({ Version: "vDEV" })).toBe("Grendel vDEV");
  });
});
