import { describe, expect, it } from "vitest";
import { RangeSetND } from "../range-set-nd";

describe("RangeSetND simple and fold", () => {
  const tests: { test: string[][]; result: string; length: number }[] = [
    { test: [["0-10"]], result: "0-10\n", length: 11 },
    { test: [["0-10/2", "01-02"]], result: "0,2,4,6,8,10; 01-02\n", length: 12 },
    { test: [["008-009", "0-10/2", "01-02"]], result: "008-009; 0,2,4,6,8,10; 01-02\n", length: 24 },
    { test: [["0-10"], ["40-60"]], result: "0-10,40-60\n", length: 32 },
    { test: [["0-2", "1-2"], ["10", "3-5"]], result: "0-2; 1-2\n10; 3-5\n", length: 9 },
    {
      test: [["0-10", "1-2"], ["5-15,40-60", "1-3"], ["0-4", "3"]],
      result: "0-15,40-60; 1-3\n",
      length: 111,
    },
    { test: [["0-10"], ["11-60"]], result: "0-60\n", length: 61 },
    { test: [["0-2", "1-2"], ["3", "1-2"]], result: "0-3; 1-2\n", length: 8 },
    { test: [["3", "1-3"], ["0-2", "1-2"]], result: "0-2; 1-2\n3; 1-3\n", length: 9 },
    { test: [["0-2", "1-2"], ["3", "1-3"]], result: "0-2; 1-2\n3; 1-3\n", length: 9 },
    {
      test: [["0-2", "1-2"], ["1-3", "1-3"]],
      result: "1-2; 1-3\n0,3; 1-2\n3; 3\n",
      length: 11,
    },
    {
      test: [["0-2", "1-2", "0-4"], ["3", "1-2", "0-5"]],
      result: "0-2; 1-2; 0-4\n3; 1-2; 0-5\n",
      length: 42,
    },
    {
      test: [["0-2", "1-2", "0-4"], ["1-3", "1-3", "0-4"]],
      result: "1-2; 1-3; 0-4\n0,3; 1-2; 0-4\n3; 3; 0-4\n",
      length: 55,
    },
    {
      test: [["0-100", "50-200"], ["2-101", "49"]],
      result: "0-100; 50-200\n2-101; 49\n",
      length: 15351,
    },
  ];

  it.each(tests)("folds %#", ({ test, result, length }) => {
    const nd = new RangeSetND(test);
    expect(nd.toString()).toBe(result);
    expect(nd.len()).toBe(length);
  });
});

describe("RangeSetND iterator on unfolded data", () => {
  it("iterates the cartesian product of a single vector", () => {
    const nd = new RangeSetND([["0-2", "1-2"]]);
    const it = nd.iterator();
    const vals: number[][] = [];
    while (it.next()) vals.push(it.intValue());
    expect(vals).toEqual([
      [0, 1],
      [0, 2],
      [1, 1],
      [1, 2],
      [2, 1],
      [2, 2],
    ]);
  });
});
