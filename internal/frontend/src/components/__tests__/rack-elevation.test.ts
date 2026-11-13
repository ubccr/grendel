import { describe, expect, it } from "vitest";
import {
  placeNodes,
  rackOrder,
  rackRows,
  rackUnit,
  sharedTags,
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

describe("sharedTags", () => {
  it("finds the tags on every node, without the rack tag", () => {
    const hosts = [
      { name: "a", tags: ["rack=f01", "fern", "nvme", "gpu"] },
      { name: "b", tags: ["nvme", "rack=f01", "fern"] },
    ];
    expect(sharedTags(hosts, "f01")).toEqual(["fern", "nvme"]);
  });
  it("has none for a single node", () => {
    expect(sharedTags([{ name: "a", tags: ["fern"] }], "f01")).toEqual([]);
  });
});
