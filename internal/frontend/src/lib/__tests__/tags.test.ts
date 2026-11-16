import { describe, expect, it } from "vitest";
import { matchesTags } from "../tags";

describe("matchesTags", () => {
  const tags = ["rack=u25", "gpu", "grendel:brand=dell"];

  it("needs every filter to pass", () => {
    expect(matchesTags(tags, ["rack=u25", "gpu"])).toBe(true);
    expect(matchesTags(tags, ["rack=u25", "ib"])).toBe(false);
    expect(matchesTags(tags, [])).toBe(true);
  });

  it("reads ! as without, and a trailing = or : as any tag holding it", () => {
    expect(matchesTags(tags, ["!ib"])).toBe(true);
    expect(matchesTags(tags, ["!gpu"])).toBe(false);
    expect(matchesTags(tags, ["rack="])).toBe(true);
    expect(matchesTags(tags, ["grendel:"])).toBe(true);
    expect(matchesTags(tags, ["row="])).toBe(false);
  });

  it("passes no filter for a node without tags", () => {
    expect(matchesTags(undefined, ["gpu"])).toBe(false);
  });
});
