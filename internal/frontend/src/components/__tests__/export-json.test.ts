import { describe, expect, it } from "vitest";
import { withoutIds } from "@/components/actions/export-json";

describe("withoutIds", () => {
  it("drops node, interface and bond ids and keeps everything else", () => {
    const node = {
      id: 7,
      uid: "2abc",
      name: "cpn-f01-24",
      tags: ["id"],
      interfaces: [{ id: 3, mac: "aa:bb", name: "eth0" }],
      bonds: [{ id: 4, name: "bond0", peers: ["eth0"] }],
    };
    expect(withoutIds([node])).toEqual([
      {
        name: "cpn-f01-24",
        tags: ["id"],
        interfaces: [{ mac: "aa:bb", name: "eth0" }],
        bonds: [{ name: "bond0", peers: ["eth0"] }],
      },
    ]);
    expect(node.id).toBe(7);
  });
  it("leaves a template named id in an image alone", () => {
    const image = { id: 1, uid: "x", name: "rocky", provision_templates: { id: "id.tmpl" } };
    expect(withoutIds([image])).toEqual([
      { name: "rocky", provision_templates: { id: "id.tmpl" } },
    ]);
  });
});
