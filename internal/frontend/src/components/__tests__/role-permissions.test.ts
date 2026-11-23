import { describe, expect, it } from "vitest";
import { groupPermissions, kindOf, matches } from "@/components/roles/permissions";

describe("groupPermissions", () => {
  it("sorts into the server's groups, labelled by summary, with undescribed ones under Other", () => {
    const groups = groupPermissions([
      {
        method: "DELETE",
        path: "/v1/nodes",
        summary: "Delete nodes",
        group: "Nodes",
        kind: "delete",
      },
      { method: "GET", path: "/v1/zzz" },
      { method: "GET", path: "/v1/nodes", summary: "List nodes", group: "Nodes", kind: "view" },
      {
        method: "POST",
        path: "/v1/bmc/sel/clear",
        summary: "Clear the system event log",
        group: "BMC",
        kind: "delete",
      },
      { method: "GET", path: "/v1/nodes", summary: "List nodes", group: "Nodes", kind: "view" },
    ]);
    expect(groups.map((g) => g.name)).toEqual(["BMC", "Nodes", "Other"]);
    expect(groups[0].entries[0]).toMatchObject({
      label: "Clear the system event log",
      kind: "delete",
    });
    expect(groups[1].entries.map((e) => e.label)).toEqual(["Delete nodes", "List nodes"]);
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
