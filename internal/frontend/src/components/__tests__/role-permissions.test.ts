import { describe, expect, it } from "vitest";
import { groupPermissions, kindOf, matches } from "@/components/roles/permissions";

describe("groupPermissions", () => {
  it("sorts into the catalog's groups and order, with unknown ones under Other", () => {
    const groups = groupPermissions([
      { method: "DELETE", path: "/v1/nodes" },
      { method: "GET", path: "/v1/zzz" },
      { method: "GET", path: "/v1/nodes" },
      { method: "GET", path: "/v1/images" },
      { method: "GET", path: "/v1/nodes" },
    ]);
    expect(groups.map((g) => g.name)).toEqual(["Nodes", "Boot images", "Other"]);
    expect(groups[0].entries.map((e) => e.label)).toEqual(["List nodes", "Delete nodes"]);
    expect(groups[2].entries[0]).toMatchObject({ label: "GET /v1/zzz", kind: "view" });
  });
});

describe("kindOf", () => {
  it("tells viewing, changing and deleting apart", () => {
    expect(kindOf("GET")).toBe("view");
    expect(kindOf("PATCH")).toBe("change");
    expect(kindOf("POST")).toBe("change");
    expect(kindOf("DELETE")).toBe("delete");
  });
});

describe("matches", () => {
  const e = {
    key: "",
    method: "PATCH",
    path: "/v1/nodes/tags/%",
    label: "Add and remove tags",
    kind: "change" as const,
  };
  it("matches the label, method or path", () => {
    expect(matches(e, "TAGS")).toBe(true);
    expect(matches(e, "patch")).toBe(true);
    expect(matches(e, " ")).toBe(true);
    expect(matches(e, "power")).toBe(false);
  });
});
