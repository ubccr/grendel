import { biosDiff, editable, inputKind, menuSection, parseInput, sameValue } from "@/lib/bios";
import { describe, expect, it } from "vitest";

describe("biosDiff", () => {
  const nodes = {
    "cpn-01": { Sriov: "Enabled", CStates: "Disabled", NumLock: "On" },
    "cpn-02": { Sriov: "Enabled", CStates: "Disabled", NumLock: "On" },
    "cpn-03": { Sriov: "Disabled", CStates: "Disabled", NumLock: "On" },
    "cpn-04": { Sriov: "Enabled", CStates: "Disabled" },
  };

  it("lists only attributes that differ, most common value first", () => {
    const rows = biosDiff(nodes);
    expect(rows.map((r) => r.name)).toEqual(["NumLock", "Sriov"]);
    expect(rows[1].groups.map((g) => [g.value, g.nodes])).toEqual([
      ["Enabled", ["cpn-01", "cpn-02", "cpn-04"]],
      ["Disabled", ["cpn-03"]],
    ]);
  });

  it("counts a missing attribute as its own value", () => {
    const groups = biosDiff(nodes)[0].groups;
    expect(groups[groups.length - 1]).toMatchObject({ value: "(absent)", nodes: ["cpn-04"] });
  });

  it("folds the nodes into a nodeset", () => {
    expect(biosDiff(nodes)[1].groups[0].nodeset).toBe("cpn-[01-02,04]");
  });

  it("limits the comparison to the named attributes", () => {
    expect(biosDiff(nodes, ["CStates"])).toEqual([]);
  });
});

describe("inputKind", () => {
  it("prefers the registry type", () => {
    expect(inputKind({ type: "Enumeration" }, "x")).toBe("enum");
    expect(inputKind({ type: "Password" }, null)).toBe("password");
  });

  it("falls back to the current value's type without a registry", () => {
    expect(inputKind(undefined, 3)).toBe("integer");
    expect(inputKind(undefined, false)).toBe("boolean");
    expect(inputKind(undefined, "On")).toBe("string");
  });
});

describe("editable", () => {
  it("refuses read only, immutable, grayed out and password attributes", () => {
    expect(editable(undefined)).toBe(true);
    expect(editable({ read_only: true })).toBe(false);
    expect(editable({ immutable: true })).toBe(false);
    expect(editable({ gray_out: true })).toBe(false);
    expect(editable({ type: "Password" })).toBe(false);
  });
});

describe("menuSection", () => {
  it("humanizes the top menu", () => {
    expect(menuSection("./ProcSettingsRef/Turbo")).toBe("Proc Settings");
    expect(menuSection("./SysSecurity")).toBe("Sys Security");
    expect(menuSection("./BIOSBootSettings")).toBe("BIOS Boot Settings");
    expect(menuSection(undefined)).toBe("Other");
  });
});

describe("values", () => {
  it("parses typed input", () => {
    expect(parseInput("integer", "4")).toBe(4);
    expect(parseInput("integer", "four")).toBe("four");
    expect(parseInput("boolean", "true")).toBe(true);
    expect(parseInput("enum", "Enabled")).toBe("Enabled");
  });

  it("compares numbers loosely", () => {
    expect(sameValue(4, 4)).toBe(true);
    expect(sameValue("4", 4)).toBe(true);
    expect(sameValue("On", "Off")).toBe(false);
    expect(sameValue(undefined, "On")).toBe(false);
  });
});
