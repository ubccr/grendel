import { describe, expect, it } from "vitest";
import { contains, expand, fold } from "../index";
import { NodeSet } from "../node-set";

interface NodeSetCase {
  input: string;
  result: string;
  length: number;
}

const simpleCases: NodeSetCase[] = [
  { input: "cws-machin", result: "cws-machin", length: 1 },
  { input: "supercluster0", result: "supercluster0", length: 1 },
  { input: "0cluster", result: "0cluster", length: 1 },
  { input: "[0]cluster", result: "0cluster", length: 1 },
  { input: "cpn-d13-01", result: "cpn-d13-01", length: 1 },
  { input: "cpn-d13-[01]", result: "cpn-d13-01", length: 1 },
  { input: "cpn-d13-[01-10]", result: "cpn-d13-[01-10]", length: 10 },
  {
    input: "cpn-k[08-09]-[02-24/2]-[01-02]",
    result: "cpn-k[08-09]-[02,04,06,08,10,12,14,16,18,20,22,24]-[01-02]",
    length: 48,
  },
  { input: " tigrou2 , tigrou7 , tigrou[5,9-11] ", result: "tigrou[2,5,7,9-11]", length: 6 },
  { input: "clu-0-3", result: "clu-0-3", length: 1 },
  { input: "clu-0-[3-23]", result: "clu-0-[3-23]", length: 21 },
  { input: "cluster[0001-0100]", result: "cluster[0001-0100]", length: 100 },
  { input: "cluster[0034-8127]", result: "cluster[0034-8127]", length: 8094 },
  {
    input: "cluster[0001,0002,1555-1559]-ipmi",
    result: "cluster[0001-0002,1555-1559]-ipmi",
    length: 7,
  },
  {
    input: "cluster115,cluster116,cluster117,cluster130,cluster166",
    result: "cluster[115-117,130,166]",
    length: 5,
  },
  {
    input: "cluster115,cluster116,cluster117,cluster130,cluster[166-169],cluster170",
    result: "cluster[115-117,130,166-170]",
    length: 9,
  },
  { input: "srv-p24-09,srv-p24-12", result: "srv-p24-[09,12]", length: 2 },
  { input: "srv-p24-10,srv-p24-09", result: "srv-p24-[09-10]", length: 2 },
  { input: "node,cpn-v14-[05,10,13-18]", result: "cpn-v14-[05,10,13-18],node", length: 9 },
  {
    input:
      "cpn-q[06-09]-[36,35,32,31,28,27,17,16,13,12,09,08,05,04]-[01-02],cpn-q[06-09]-[20,23],cpn-q[07-08]-[39,40]-[01-02]",
    result:
      "cpn-q[06-09]-[04-05,08-09,12-13,16-17,27-28,31-32,35-36]-[01-02],cpn-q[07-08]-[39-40]-[01-02],cpn-q[06-09]-[20,23]",
    length: 128,
  },
  {
    input: "a3b2c0,a2b3c1,a2b4c1,a1b2c0,a1b2c1,a3b2c1,a2b5c1",
    result: "a[1,3]b2c[0-1],a2b[3-5]c1",
    length: 7,
  },
];

describe("NodeSet simple", () => {
  it.each(simpleCases)("$input -> $result", ({ input, result, length }) => {
    const ns = new NodeSet(input);
    expect(ns.toString()).toBe(result);
    expect(ns.len()).toBe(length);
  });
});

describe("NodeSet iterator", () => {
  it("yields the expected node names", () => {
    const expected: string[] = [];
    for (let i = 1; i < 11; i++) {
      expected.push(`cpn-d13-${String(i).padStart(2, "0")}`);
    }

    const ns = new NodeSet("cpn-d13-[01-10]");
    expect(ns.len()).toBe(10);

    const it = ns.iterator();
    const result: string[] = [];
    while (it.next()) result.push(it.value());

    expect(result).toEqual(expected);
  });

  it("supports the for-of protocol", () => {
    const ns = new NodeSet("cpn-d13-[01-03]");
    expect([...ns]).toEqual(["cpn-d13-01", "cpn-d13-02", "cpn-d13-03"]);
  });
});

describe("NodeSet JSON", () => {
  const tests: { test: string[]; result: string; length: number; roundtrip: string[] }[] = [
    { test: ["cws-machin"], result: "cws-machin", length: 1, roundtrip: ["cws-machin"] },
    {
      test: ["cpn-d13-[01-10]", "cpn-d14-[01-05]"],
      result: "cpn-d13-[01-10],cpn-d14-[01-05]",
      length: 15,
      roundtrip: ["cpn-d13-[01-10]", "cpn-d14-[01-05]"],
    },
    {
      test: ["node", "cpn-v14-[05,10,13-18]"],
      result: "cpn-v14-[05,10,13-18],node",
      length: 9,
      roundtrip: ["cpn-v14-[05,10,13-18]", "node"],
    },
  ];

  it.each(tests)("round-trips %j", ({ test, result, length, roundtrip }) => {
    const ns = NodeSet.fromJSON(test);
    expect(ns.toString()).toBe(result);
    expect(ns.len()).toBe(length);

    const round = JSON.parse(JSON.stringify(ns)) as string[];
    expect([...round].sort()).toEqual([...roundtrip].sort());
  });
});

describe("functional helpers", () => {
  it("expand returns the materialized node names", () => {
    expect(expand("cpn-d13-[01-03]")).toEqual(["cpn-d13-01", "cpn-d13-02", "cpn-d13-03"]);
  });

  it("fold compresses a flat list", () => {
    expect(fold(["cpn-d13-01", "cpn-d13-02", "cpn-d13-03"])).toBe("cpn-d13-[01-03]");
  });

  it("fold compresses a flat multivariate list", () => {
    expect(fold(["cpn-d13-01", "cpn-d13-02", "cpn-d14-01", "cpn-d14-02"])).toBe(
      "cpn-d[13-14]-[01-02]",
    );
  });

  it("contains finds a node", () => {
    expect(contains("cpn-d13-[01-10]", "cpn-d13-05")).toBe(true);
    expect(contains("cpn-d13-[01-10]", "cpn-d13-99")).toBe(false);
  });

  it("fold returns empty for empty input", () => {
    expect(fold([])).toBe("");
  });
});
