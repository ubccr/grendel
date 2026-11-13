import { describe, expect, it } from "vitest";
import { type CrumbPart, crumbsFrom } from "../app-breadcrumbs";

// match builds the part of a route match crumbsFrom reads
const match = (
  routeId: string,
  crumb?: CrumbPart | CrumbPart[],
  params: Record<string, string> = {},
) => ({ routeId, fullPath: routeId.replace(/\/$/, ""), params, staticData: { crumb } });

const nodeBios = (params = { node: "cpn-01" }) => [
  match("__root__"),
  match("/nodes", "Nodes"),
  match("/nodes/$node", { param: "node" }, params),
  match("/nodes/$node/bios", "BIOS", params),
];

describe("crumbsFrom", () => {
  it("shows nothing for routes without crumbs", () => {
    expect(crumbsFrom([match("__root__"), match("/")])).toEqual([]);
  });

  it("links each route's crumb to itself, with its params", () => {
    expect(crumbsFrom(nodeBios())).toEqual([
      { label: "Nodes", link: { to: "/nodes", params: {} } },
      {
        label: "cpn-01",
        name: true,
        link: { to: "/nodes/$node", params: { node: "cpn-01" } },
      },
      { label: "BIOS", link: { to: "/nodes/$node/bios", params: { node: "cpn-01" } } },
    ]);
  });

  it("keeps names as the router gives them", () => {
    expect(crumbsFrom(nodeBios({ node: "a%b c" }))[1].label).toBe("a%b c");
  });

  it("links only the last of a route's parts to itself", () => {
    const params = { nodes: "cpn-[01-04]" };
    const crumbs = crumbsFrom([
      match("/nodes", "Nodes"),
      match("/nodes/bios/$nodes", ["BIOS", { param: "nodes" }], params),
    ]);
    expect(crumbs[1]).toEqual({ label: "BIOS", link: undefined });
    expect(crumbs[2]).toEqual({
      label: "cpn-[01-04]",
      name: true,
      link: { to: "/nodes/bios/$nodes", params },
    });
  });

  it("links a part to the list a route isn't nested under", () => {
    expect(
      crumbsFrom([match("/add/node", [{ label: "Nodes", to: "/nodes" }, "Add node"])]),
    ).toEqual([
      { label: "Nodes", link: { to: "/nodes" } },
      { label: "Add node", link: { to: "/add/node", params: {} } },
    ]);
  });

  it("puts a node opened from a rack under that rack", () => {
    const labels = crumbsFrom(nodeBios(), "r12").map((c) => c.label);
    expect(labels).toEqual(["Racks", "r12", "cpn-01", "BIOS"]);
    expect(crumbsFrom(nodeBios(), "r12")[1].link).toEqual({
      to: "/racks/$rack",
      params: { rack: "r12" },
    });
  });
});
