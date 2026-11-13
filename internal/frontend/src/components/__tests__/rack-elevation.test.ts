import { describe, expect, it } from "vitest";
import {
  hallRow,
  placeNodes,
  rackOrder,
  rackRows,
  rackUnit,
  rackTags,
} from "@/components/racks/elevation";

const h = (name: string) => ({ name });

describe("rackUnit", () => {
  it("reads the third part of the name", () => {
    expect(rackUnit("cpn-f01-24")).toBe(24);
    expect(rackUnit("cpn-b05-08-01")).toBe(8);
  });
  it("refuses names without a unit", () => {
    expect(rackUnit("leaf-fern-a")).toBeUndefined();
    expect(rackUnit("pdu-f01-r")).toBeUndefined();
    expect(rackUnit("test-1")).toBeUndefined();
    expect(rackUnit("cpn-f01-00")).toBeUndefined();
    expect(rackUnit(undefined)).toBeUndefined();
  });
});

describe("placeNodes", () => {
  it("puts sleds in one unit, sorted, and keeps the unplaced", () => {
    const p = placeNodes([h("cpn-b05-08-02"), h("cpn-b05-08-01"), h("sw-core")]);
    expect(p.units.get(8)?.map((x) => x.name)).toEqual(["cpn-b05-08-01", "cpn-b05-08-02"]);
    expect(p.unplaced.map((x) => x.name)).toEqual(["sw-core"]);
    expect(p.height).toBe(42);
  });
  it("grows the rack for a node above 42", () => {
    expect(placeNodes([h("cpn-x-48")]).height).toBe(48);
  });
});

describe("rackRows", () => {
  it("collapses long empty runs and keeps short ones", () => {
    const p = placeNodes([h("a-r-42"), h("a-r-39"), h("a-r-03")]);
    const rows = rackRows(p);
    expect(rows[0]).toMatchObject({ kind: "unit", u: 42 });
    expect(rows[1]).toMatchObject({ kind: "unit", u: 41, hosts: [] });
    expect(rows[2]).toMatchObject({ kind: "unit", u: 40, hosts: [] });
    expect(rows[3]).toMatchObject({ kind: "unit", u: 39 });
    expect(rows[4]).toEqual({ kind: "gap", top: 38, bottom: 4 });
    expect(rows[5]).toMatchObject({ kind: "unit", u: 3 });
    expect(rows.slice(6)).toMatchObject([{ u: 2 }, { u: 1 }]);
  });
});

describe("rackOrder", () => {
  it("lists top down, then the unplaced", () => {
    const p = placeNodes([h("a-r-03"), h("sw"), h("a-r-20-2"), h("a-r-20-1")]);
    expect(rackOrder(p)).toEqual(["a-r-20-1", "a-r-20-2", "a-r-03", "sw"]);
  });
});

describe("rackTags", () => {
  it("splits the tags on every node from those on some, without the rack tag", () => {
    const hosts = [
      { name: "a", tags: ["rack=f01", "fern", "nvme", "gpu", "ib"] },
      { name: "b", tags: ["nvme", "rack=f01", "fern", "gpu"] },
      { name: "c", tags: ["rack=f01", "fern", "nvme", "ib", "gpu"] },
      { name: "d", tags: ["rack=f01", "fern", "nvme", "solo"] },
    ];
    expect(rackTags(hosts, "f01")).toEqual({
      every: ["fern", "nvme"],
      some: [
        ["gpu", 3],
        ["ib", 2],
      ],
    });
  });
  it("counts a tag a node repeats once", () => {
    const hosts = [
      { name: "a", tags: ["gpu", "gpu"] },
      { name: "b", tags: [] },
      { name: "c", tags: ["gpu"] },
    ];
    expect(rackTags(hosts, "f01").some).toEqual([["gpu", 2]]);
  });
  it("has none for a single node", () => {
    expect(rackTags([{ name: "a", tags: ["fern"] }], "f01")).toEqual({ every: [], some: [] });
  });
});

describe("hallRow", () => {
  it("drops the trailing number", () => {
    expect(hallRow("f01")).toBe("f");
    expect(hallRow("rack-12")).toBe("rack");
    expect(hallRow("b05_3")).toBe("b05");
  });
  it("gives no row when the name is all number or has none", () => {
    expect(hallRow("12")).toBe("");
    expect(hallRow("west")).toBe("");
  });
});
