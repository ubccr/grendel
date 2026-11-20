import { describe, expect, it } from "vitest";
import {
  bootRows,
  choose,
  defaultBootSettings,
  defaultRule,
  groups,
  layout,
  learned,
  toBootImport,
  unreachedGroup,
  type NodeNics,
  type Seen,
} from "../boot-nics";
import { parseTemplate } from "../name-template";

const nic = (id: string, mac: string, link = "LinkDown") => ({ id, mac, link });

const r650 = (node: string, upSlot = true): NodeNics => ({
  node,
  nics: [
    nic("NIC.Integrated.1-1-1", `b8:00:00:00:${node.slice(-2)}:01`),
    nic("NIC.Slot.3-1-1", `b8:00:00:00:${node.slice(-2)}:31`, upSlot ? "LinkUp" : "LinkDown"),
    nic("NIC.Slot.3-2-1", `b8:00:00:00:${node.slice(-2)}:32`),
  ],
});
const r760 = (node: string): NodeNics => ({
  node,
  nics: [nic("NIC.Embedded.1-1-1", `c4:00:00:00:${node.slice(-2)}:01`, "LinkUp")],
});

const none = new Map<string, Seen>();
const byNodeOf = (nodes: NodeNics[]) => new Map(nodes.map((n) => [n.node, n.nics]));

describe("boot NIC rules", () => {
  it("groups nodes by port layout, biggest first, leaving out unreached BMCs", () => {
    const g = groups([
      r650("n01"),
      r760("n02"),
      r650("n03"),
      { node: "n04", nics: [], error: "x" },
    ]);
    expect(g.map((x) => x.nodes)).toEqual([["n01", "n03"], ["n02"]]);
    expect(g[0].ids).toEqual(["NIC.Integrated.1-1-1", "NIC.Slot.3-1-1", "NIC.Slot.3-2-1"]);
  });

  it("picks by id, by link, or by what a switch has seen", () => {
    const n = r650("n01").nics;
    expect(choose(n, { kind: "id", id: "NIC.Slot.3-2-1" }, none)?.mac).toBe("b8:00:00:00:01:32");
    expect(choose(n, { kind: "link" }, none)?.id).toBe("NIC.Slot.3-1-1");
    const seen = new Map<string, Seen>([
      ["b8:00:00:00:01:01", { sw: "swe-a01", ifname: "Ethernet1" }],
    ]);
    expect(choose(n, { kind: "switch" }, seen)?.id).toBe("NIC.Integrated.1-1-1");
    seen.set("b8:00:00:00:01:32", { sw: "swe-a01", ifname: "Ethernet2" });
    expect(choose(n, { kind: "switch" }, seen)).toBeUndefined();
  });

  it("leaves out a MAC learned on more than one port rather than guess which is the server's", () => {
    const seen = learned([
      {
        name: "spine-01",
        nos: "arista",
        ports: [
          {
            ifname: "Ethernet1",
            macs: [{ mac: "b8:00:00:00:01:01" }, { mac: "b8:00:00:00:03:01" }],
          },
        ],
      },
      {
        name: "swe-a01",
        nos: "arista",
        ports: [
          { ifname: "Ethernet7", macs: [{ mac: "b8:00:00:00:01:01", owner: "n01" }] },
          { ifname: "Ethernet8", macs: [{ mac: "b8:00:00:00:02:01", owner: "n02" }] },
        ],
      },
    ]);
    expect(seen.get("b8:00:00:00:02:01")).toEqual({
      sw: "swe-a01",
      ifname: "Ethernet8",
      port: 8,
      owner: "n02",
    });
    expect(seen.has("b8:00:00:00:01:01")).toBe(false);
    expect(seen.get("b8:00:00:00:03:01")?.sw).toBe("spine-01");
  });

  it("defaults to the switch's port when it settles every node", () => {
    const nodes = [r650("n01"), r650("n03")];
    const seen = new Map<string, Seen>([
      ["b8:00:00:00:01:01", { sw: "s", ifname: "Ethernet1" }],
      ["b8:00:00:00:03:01", { sw: "s", ifname: "Ethernet3" }],
    ]);
    expect(defaultRule(groups(nodes)[0], byNodeOf(nodes), seen)).toEqual({ kind: "switch" });
  });

  it("defaults to the one linked port id when every node agrees", () => {
    const nodes = [r650("n01"), r650("n03")];
    expect(defaultRule(groups(nodes)[0], byNodeOf(nodes), none)).toEqual({
      kind: "id",
      id: "NIC.Slot.3-1-1",
    });
  });

  it("falls back to the first port when a node has no link", () => {
    const nodes = [r650("n01"), r650("n03", false)];
    expect(defaultRule(groups(nodes)[0], byNodeOf(nodes), none)).toEqual({
      kind: "id",
      id: "NIC.Integrated.1-1-1",
    });
  });

  it("builds rows with overrides, the ports already in Grendel and the FQDN", () => {
    const nodes = [r650("n01"), r650("n03")];
    const rules = { [layout(nodes[0].nics)]: { kind: "id" as const, id: "NIC.Slot.3-1-1" } };
    const rows = bootRows(
      nodes,
      [
        {
          name: "n01",
          interfaces: [
            { mac: "aa:aa:aa:aa:aa:aa", bmc: true },
            { mac: "b8:00:00:00:01:31", bmc: false },
          ],
        },
      ],
      rules,
      { n03: { mac: "b8:00:00:00:03:32" } },
      parseTemplate("{name}.cluster"),
      none,
    );
    expect(rows.map((r) => r.nic?.mac)).toEqual(["b8:00:00:00:01:31", "b8:00:00:00:03:32"]);
    expect(rows.map((r) => r.status)).toEqual(["existing", "ready"]);
    expect(rows[0].known).toEqual(["b8:00:00:00:01:31"]);
    expect(rows[1].problem).toBeUndefined();
    expect(rows[1].overridden).toBe(true);
    expect(toBootImport(rows, { ...defaultBootSettings, mtu: "9000", ifname: "eno1" })[0]).toEqual({
      node: "n01",
      mac: "b8:00:00:00:01:31",
      bmc: false,
      fqdn: "n01.cluster",
      ifname: "eno1",
      vlan: undefined,
      mtu: 9000,
    });
  });

  it("flags rows that can't be imported and leaves them out", () => {
    const nodes: NodeNics[] = [
      r650("n01"),
      r650("n02"),
      r650("n03"),
      { node: "n04", nics: [] },
      { node: "n05", nics: [], error: "timeout" },
    ];
    const seen = new Map<string, Seen>([
      ["b8:00:00:00:02:01", { sw: "swe-a01", ifname: "Ethernet2", port: 2, owner: "n09" }],
    ]);
    const rows = bootRows(
      nodes,
      [],
      { [layout(nodes[0].nics)]: { kind: "switch" } },
      { n03: { include: false } },
      parseTemplate("{name}-{sw}"),
      seen,
    );
    expect(rows.map((r) => r.status)).toEqual([
      "pick",
      "conflict",
      "excluded",
      "empty",
      "unreached",
    ]);
    expect(rows[1].fqdn).toBe("n02-swe-a01");
    expect(rows[4].group).toBe(unreachedGroup);
    expect(toBootImport(rows, defaultBootSettings)).toEqual([]);
  });

  it("adds a port after a node's boot port rather than replacing it", () => {
    const n = r760("n01");
    const rows = bootRows(
      [n],
      [{ name: "n01", interfaces: [{ mac: "aa:00:00:00:00:01", bmc: false }] }],
      { [layout(n.nics)]: { kind: "link" } },
      {},
      parseTemplate(""),
      none,
    );
    expect(rows[0].status).toBe("ready");
    expect(rows[0].note).toBe("Added after aa:00:00:00:00:01, which stays the boot interface.");
    expect(rows[0].problem).toBeUndefined();
  });

  it("reports a template field the port can't fill", () => {
    const rows = bootRows(
      [r760("n01")],
      [],
      { [layout(r760("n01").nics)]: { kind: "link" } },
      {},
      parseTemplate("{name}-{port}"),
      none,
    );
    expect(rows[0].status).toBe("invalid");
  });
});
