import { describe, expect, it } from "vitest";
import { frontPanel, neighborKind, shortName, switchNos } from "../switch-ports";

const ports = (...names: string[]) => names.map((port_name) => ({ port_name }));
const layout = (nos: string, ...names: string[]) =>
  frontPanel(nos, ports(...names))?.map((b) => ({
    slots: b.slots.map((s) => (s.neighbors.length ? s.number : -s.number)),
  }));

describe("frontPanel", () => {
  it("fills gaps with empty cages and pads each block to whole columns", () => {
    expect(layout("arista", "Ethernet1", "Ethernet2", "Ethernet4", "Ethernet5")).toEqual([
      { slots: [1, 2, -3, 4, 5, -6] },
    ]);
  });

  it("puts every lane of a breakout port in its cage", () => {
    const blocks = frontPanel("arista", ports("Ethernet47", "Ethernet49/1", "Ethernet49/2"))!;
    expect(blocks[0].slots.map((s) => [s.number, s.neighbors.length])).toEqual([
      [47, 1],
      [48, 0],
      [49, 2],
      [50, 0],
    ]);
  });

  it("starts a new block after a wide gap", () => {
    expect(layout("dellos10", "ethernet1/1/1", "ethernet1/1/2", "ethernet1/1/49")).toEqual([
      { slots: [1, 2] },
      { slots: [49, -50] },
    ]);
  });

  it("draws a switch of only breakout ports as one block", () => {
    expect(layout("arista", "Ethernet1/1", "Ethernet2/1", "Ethernet4/1")).toEqual([
      { slots: [1, 2, -3, 4] },
    ]);
  });

  it("gives up on ports spread over several units or blade slots", () => {
    expect(frontPanel("dellos10", ports("ethernet1/1/1", "ethernet1/71/1"))).toBeUndefined();
  });

  it("numbers SONiC ports by front panel position rather than serdes lane", () => {
    expect(layout("sonic", "Ethernet0", "Ethernet4", "Ethernet12")).toEqual([
      { slots: [1, 2, -3, 4] },
    ]);
  });

  it("gives up rather than guess when a name doesn't fit the switch type", () => {
    expect(frontPanel("arista", ports("Ethernet1", "ethernet1/1/2"))).toBeUndefined();
    expect(frontPanel(undefined, ports("Ethernet1"))).toBeUndefined();
    expect(frontPanel("arista", [])).toBeUndefined();
  });
});

describe("switchNos", () => {
  it("reads the switch type from the node's tags", () => {
    expect(switchNos(["switch", "sonic"])).toBe("sonic");
    expect(switchNos(["switch"])).toBeUndefined();
    expect(switchNos(null)).toBeUndefined();
  });
});

describe("neighborKind", () => {
  const nodes = new Set(["cpn-a01-01"]);

  it("matches nodes by short name", () => {
    expect(shortName("cpn-a01-01.ccr.example.org")).toBe("cpn-a01-01");
    expect(neighborKind({ system_name: "cpn-a01-01.ccr.example.org" }, nodes)).toBe("node");
  });

  it("calls anything Grendel doesn't have other, whatever it says it is", () => {
    expect(
      neighborKind(
        {
          system_name: "spine-01",
          port_id_type: "interfaceName",
          system_description: "Arista Networks EOS version 4.32",
        },
        nodes,
      ),
    ).toBe("other");
  });
});
