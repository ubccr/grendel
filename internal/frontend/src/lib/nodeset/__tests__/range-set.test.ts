import { describe, expect, it } from "vitest";
import { ParseRangeSetError } from "../errors";
import { RangeSet } from "../range-set";

describe("RangeSet simple", () => {
  const tests: { test: string; result: string; length: number }[] = [
    { test: "0", result: "0", length: 1 },
    { test: "1", result: "1", length: 1 },
    { test: "0-2", result: "0-2", length: 3 },
    { test: "1-3", result: "1-3", length: 3 },
    { test: "1-3,4-6", result: "1-6", length: 6 },
    { test: "1-3,4-6,7-10", result: "1-10", length: 10 },
    { test: "0001-0010", result: "0001-0010", length: 10 },
  ];

  it.each(tests)("parses %s", ({ test, result, length }) => {
    const r = new RangeSet(test);
    expect(r.toString()).toBe(result);
    expect(r.len()).toBe(length);
  });
});

describe("RangeSet step", () => {
  const tests: { test: string; result: string; length: number }[] = [
    { test: "0-4/2", result: "0,2,4", length: 3 },
    { test: "1-4/2", result: "1,3", length: 2 },
    { test: "1-4/3", result: "1,4", length: 2 },
    { test: "1-4/4", result: "1", length: 1 },
  ];

  it.each(tests)("parses %s", ({ test, result, length }) => {
    const r = new RangeSet(test);
    expect(r.toString()).toBe(result);
    expect(r.len()).toBe(length);
  });
});

describe("RangeSet bad syntax", () => {
  const badSyntax = [
    "-",
    "A",
    "2-5/a",
    "3/2",
    "3-/2",
    "-3/2",
    "-/2",
    "4-a/2",
    "4-3/2",
    "4-5/-2",
    "4-2/-2",
    "004-002",
    "3-59/2,102a",
  ];

  it.each(badSyntax)("rejects %s", (syn) => {
    expect(() => new RangeSet(syn)).toThrow(ParseRangeSetError);
  });
});

describe("RangeSet equality", () => {
  it("empty sets are equal", () => {
    expect(new RangeSet("").equal(new RangeSet(""))).toBe(true);
  });
  it("different padding still compares by values", () => {
    expect(new RangeSet("2-5").equal(new RangeSet("1,2,3,4"))).toBe(false);
    expect(new RangeSet("1-5").equal(new RangeSet("1,2,3,4,5"))).toBe(true);
  });
});

describe("RangeSet intersection", () => {
  it("in-place intersection", () => {
    const r1 = new RangeSet("4-34");
    const r2 = new RangeSet("27-42");
    r1.inPlaceIntersection(r2);
    expect(r1.toString()).toBe("27-34");
    expect(r1.len()).toBe(8);
  });

  it("complex intersection", () => {
    const r1 = new RangeSet("2-450,654-700,800");
    const r2 = new RangeSet("500-502,690-820,830-840,900");
    r1.inPlaceIntersection(r2);
    expect(r1.toString()).toBe("690-700,800");
    expect(r1.len()).toBe(12);
  });

  it("copy intersection", () => {
    const r1 = new RangeSet("2-450,654-700,800");
    const r2 = new RangeSet("500-502,690-820,830-840,900");
    const r3 = r1.intersection(r2);
    expect(r3.toString()).toBe("690-700,800");
    expect(r3.len()).toBe(12);
  });

  it("empty intersection", () => {
    const r1 = new RangeSet("");
    const r2 = new RangeSet("500-502,690-820,830-840,900");
    const r3 = r1.intersection(r2);
    expect(r3.toString()).toBe("");
    expect(r3.len()).toBe(0);
  });
});

describe("RangeSet symmetric difference", () => {
  it("case 1", () => {
    const r1 = new RangeSet("4,7-33");
    const r2 = new RangeSet("8-34");
    r1.inPlaceSymmetricDifference(r2);
    expect(r1.toString()).toBe("4,7,34");
    expect(r1.len()).toBe(3);
  });

  it("copy sym diff", () => {
    const r1 = new RangeSet("4,7-33");
    const r2 = new RangeSet("8-34");
    const r3 = r1.symmetricDifference(r2);
    expect(r3.toString()).toBe("4,7,34");
    expect(r3.len()).toBe(3);
  });

  it("case 2", () => {
    const r1 = new RangeSet("5,7,10-12,33-50");
    const r2 = new RangeSet("8-34");
    r1.inPlaceSymmetricDifference(r2);
    expect(r1.toString()).toBe("5,7-9,13-32,35-50");
    expect(r1.len()).toBe(40);
  });

  it("disjoint sym diff is union", () => {
    const r1 = new RangeSet("8-30");
    const r2 = new RangeSet("31-40");
    r1.inPlaceSymmetricDifference(r2);
    expect(r1.toString()).toBe("8-40");
    expect(r1.len()).toBe(33);
  });

  it("identical sym diff is empty", () => {
    const r1 = new RangeSet("8-30");
    const r2 = new RangeSet("8-30");
    r1.inPlaceSymmetricDifference(r2);
    expect(r1.toString()).toBe("");
    expect(r1.len()).toBe(0);
  });
});

describe("RangeSet superset", () => {
  it("superset detection", () => {
    const r1 = new RangeSet("1-100,102,105-242,800");
    expect(r1.len()).toBe(240);

    const r2 = new RangeSet("3-98,140-199,800");
    expect(r2.len()).toBe(157);
    expect(r1.superset(r1)).toBe(true);
    expect(r1.superset(r2)).toBe(true);
    expect(r2.subset(r1)).toBe(true);

    const r3 = new RangeSet("3-98,140-199,243,800");
    expect(r3.len()).toBe(158);
    expect(r1.superset(r3)).toBe(false);
  });
});

describe("RangeSet iterator", () => {
  it("yields padded strings and ints sorted", () => {
    const rgs = new RangeSet("011,003,005-008,001,004");
    expect(rgs.strings()).toEqual(["001", "003", "004", "005", "006", "007", "008", "011"]);
    expect(rgs.ints()).toEqual([1, 3, 4, 5, 6, 7, 8, 11]);
  });
});
