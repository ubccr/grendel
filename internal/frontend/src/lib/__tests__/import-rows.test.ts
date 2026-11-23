import { describe, expect, it } from "vitest";
import {
  buildRows,
  compile,
  defaultSettings,
  fillNames,
  ouiChoices,
  parsePortFilter,
  pickByOui,
  tally,
  toImportRows,
  type ScanSwitch,
  type Settings,
} from "../import-rows";

const sw = (name: string, ports: ScanSwitch["ports"], nos = "arista"): ScanSwitch => ({
  name,
  nos,
  ports,
});
const mac = (m: string, extra: object = {}) => ({ mac: m, vlan: "3001", ...extra });

const a01 = sw("swe-a01", [
  { ifname: "Ethernet3", macs: [] },
  { ifname: "Ethernet1", macs: [mac("d0:8e:79:00:00:01")] },
  { ifname: "Ethernet2", macs: [mac("d0:8e:79:00:00:02"), mac("b8:ca:3a:00:00:02")] },
  { ifname: "Ethernet49/1", macs: [] },
]);
const a02 = sw("swe-a02", [{ ifname: "Ethernet1", macs: [mac("d0:8e:79:00:01:01")] }]);

const settings = (s: Partial<Settings>): Settings => ({
  ...defaultSettings,
  name: "cpn-a{sw.1}-{port:02}",
  ...s,
});
const rows = (switches: ScanSwitch[], s: Partial<Settings> = {}, o = {}) => {
  const st = settings(s);
  return buildRows(switches, st, compile(st), o);
};

describe("buildRows", () => {
  it("orders by faceplate number and names from the template", () => {
    const r = rows([a01, a02]);
    expect(r.map((x) => x.key)).toEqual([
      "swe-a01:Ethernet1",
      "swe-a01:Ethernet2",
      "swe-a01:Ethernet3",
      "swe-a01:Ethernet49/1",
      "swe-a02:Ethernet1",
    ]);
    expect(r.map((x) => x.name)).toEqual([
      "cpn-a01-01",
      "cpn-a01-02",
      "cpn-a01-03",
      "cpn-a01-49",
      "cpn-a02-01",
    ]);
  });

  it("gives each port a status", () => {
    const r = rows([a01]);
    expect(r.map((x) => x.status)).toEqual(["ready", "pick", "empty", "empty"]);
    expect(r[3].include).toBe(true);
  });

  it("picks the MAC already known as this node's BMC", () => {
    const known = sw("swe-a01", [
      {
        ifname: "Ethernet2",
        macs: [
          mac("d0:8e:79:00:00:02", { owner: "cpn-a01-02", bmc: true }),
          mac("b8:ca:3a:00:00:02"),
        ],
      },
    ]);
    const r = rows([known]);
    expect(r[0].status).toBe("existing");
  });

  it("flags a MAC held by another node", () => {
    const taken = sw("swe-a01", [
      { ifname: "Ethernet1", macs: [mac("d0:8e:79:00:00:01", { owner: "old-node", bmc: true })] },
    ]);
    const r = rows([taken]);
    expect(r[0].status).toBe("conflict");
    expect(r[0].problem).toMatch(/old-node/);
  });

  it("flags duplicate names", () => {
    const r = rows([a01, a02], { name: "cpn-{port:02}" });
    expect(r[0].status).toBe("conflict");
    expect(r[4].status).toBe("conflict");
    expect(r[0].problem).toMatch(/2 ports are named cpn-01/);
  });

  it("counts {n} over included ports, empty ones too", () => {
    const r = rows([a01, a02], { name: "node{n:02}", start: 10 });
    expect(r.filter((x) => x.include).map((x) => x.name)).toEqual([
      "node10",
      "node11",
      "node12",
      "node13",
      "node14",
    ]);
  });

  it("limits to the Port filter", () => {
    const r = rows([a01], { ports: "2-3" });
    expect(r.map((x) => x.include)).toEqual([false, true, true, false]);
    expect(r.map((x) => x.hidden)).toEqual([true, false, false, true]);
    expect(r[1].name).toBe(rows([a01], {})[1].name);
  });

  it("applies overrides", () => {
    const r = rows(
      [a01],
      {},
      {
        "swe-a01:Ethernet1": { name: "special" },
        "swe-a01:Ethernet2": { mac: "b8:ca:3a:00:00:02" },
        "swe-a01:Ethernet49/1": { include: true },
      },
    );
    expect(r[0].name).toBe("special");
    expect(r[0].nameOverridden).toBe(true);
    expect(r[1].status).toBe("ready");
    expect(r[1].mac?.mac).toBe("b8:ca:3a:00:00:02");
    expect(r[3].include).toBe(true);
  });

  it("waits for a name without calling it an error", () => {
    const r = rows([a01], { name: "" });
    expect(r[0].status).toBe("unnamed");
    expect(r[0].problem).toBeUndefined();
    // only the shared port, which still needs its MAC picked, asks for attention
    expect(tally(r).attention).toBe(1);
    expect(tally(r).unnamed).toBe(1);
    const typed = rows([a01], { name: "" }, { "swe-a01:Ethernet1": { name: "" } });
    expect(typed[0].status).toBe("invalid");
  });

  it("takes a MAC typed for a port the switch learned nothing on", () => {
    const r = rows([a01], {}, { "swe-a01:Ethernet3": { mac: "D0-8E-79-00-00-03" } });
    expect(r[2].status).toBe("ready");
    expect(r[2].mac?.mac).toBe("d0:8e:79:00:00:03");
    expect(r[2].macTyped).toBe(true);
    expect(r[2].macText).toBe("D0-8E-79-00-00-03");
    expect(toImportRows(r, { ifname: "", mtu: "", portVlan: "" }).map((x) => x.mac)).toContain(
      "d0:8e:79:00:00:03",
    );
  });

  it("types over a learned MAC, or picks one by typing it", () => {
    const over = rows([a01], {}, { "swe-a01:Ethernet1": { mac: "aabb.ccdd.eeff" } });
    expect(over[0].mac?.mac).toBe("aa:bb:cc:dd:ee:ff");
    expect(over[0].macTyped).toBe(true);
    const picked = rows([a01], {}, { "swe-a01:Ethernet2": { mac: "B8:CA:3A:00:00:02" } });
    expect(picked[1].status).toBe("ready");
    expect(picked[1].macTyped).toBe(false);
  });

  it("flags a MAC it can't read, and asks again for one that was emptied", () => {
    const bad = rows([a01], {}, { "swe-a01:Ethernet1": { mac: "d0:8e:79" } });
    expect(bad[0].status).toBe("badmac");
    expect(bad[0].problem).toMatch(/isn't a MAC address/);
    const cleared = rows([a01], {}, { "swe-a01:Ethernet1": { mac: "" } });
    expect(cleared[0].status).toBe("pick");
    expect(cleared[0].problem).toMatch(/reset to use the one the switch learned/);
  });

  it("fills pasted MACs down", () => {
    const r = rows([a01]);
    const o = fillNames(
      r,
      "swe-a01:Ethernet2",
      ["aa:00:00:00:00:02", "aa:00:00:00:00:03"],
      {},
      "mac",
    );
    const after = rows([a01], {}, o);
    expect(after[1].mac?.mac).toBe("aa:00:00:00:00:02");
    expect(after[2].status).toBe("ready");
  });

  it("explains a template it can't fill", () => {
    const r = rows([sw("core", [{ ifname: "Ethernet1", macs: [mac("d0:8e:79:00:00:01")] }])]);
    expect(r[0].status).toBe("invalid");
    expect(r[0].problem).toMatch(/no number 1/);
  });

  it("skips switches that failed to scan", () => {
    expect(rows([{ name: "swe-a03", nos: "arista", error: "timeout" }])).toEqual([]);
  });

  it("renders fqdn and tags from the name", () => {
    const r = rows([a02], { fqdn: "bmc-{name}.cluster", tags: ["a{sw.1}", "bmc-import"] });
    expect(r[0].fqdn).toBe("bmc-cpn-a02-01.cluster");
    expect(r[0].tags).toEqual(["a02", "bmc-import"]);
  });

  it("refuses a row whose FQDN or tag template can't be filled, rather than importing part of it", () => {
    const fqdn = rows([a02], { fqdn: "{name}.r{sw.2}.cluster" });
    expect(fqdn[0].status).toBe("invalid");
    expect(fqdn[0].problem).toMatch(/^FQDN: .*no number 2/);
    expect(fqdn[0].fqdn).toBe("");
    expect(toImportRows(fqdn, settings({}))).toEqual([]);

    const tag = rows([a02], { tags: ["rack=r{sw.2}", "bmc-import"] });
    expect(tag[0].status).toBe("invalid");
    expect(tag[0].problem).toMatch(/^Tag rack=r\{sw\.2\}: .*no number 2/);
    expect(tag[0].tags).toEqual([]);
  });

  it("refuses every row while the FQDN or a tag template has a syntax error", () => {
    const fqdn = rows([a02], { fqdn: "{nme}.cluster" });
    expect(fqdn[0].status).toBe("invalid");
    expect(fqdn[0].problem).toMatch(/^FQDN: .*isn't a field/);

    const tag = rows([a02], { tags: ["rack={sw.1"] });
    expect(tag[0].status).toBe("invalid");
    expect(tag[0].problem).toMatch(/^Tag rack=\{sw\.1: .*never closed/);
  });

  it("explains a port with several MACs by its MACs, not by its name template", () => {
    const r = rows([sw("core", [a01.ports![2]])]);
    expect(r[0].status).toBe("pick");
    expect(r[0].problem).toMatch(/2 MACs on this port/);
  });
});

describe("bulk helpers", () => {
  it("settles shared-LOM ports by vendor prefix", () => {
    const lom = sw("swe-a01", [
      { ifname: "Ethernet1", macs: [mac("d0:8e:79:00:00:01"), mac("b8:ca:3a:00:00:01")] },
      { ifname: "Ethernet2", macs: [mac("d0:8e:79:00:00:02"), mac("b8:ca:3a:00:00:02")] },
    ]);
    const r = rows([lom]);
    expect(ouiChoices(r)).toEqual([
      { oui: "d0:8e:79", ports: 2 },
      { oui: "b8:ca:3a", ports: 2 },
    ]);
    const o = pickByOui(r, "d0:8e:79", {});
    expect(rows([lom], {}, o).map((x) => x.mac?.mac)).toEqual([
      "d0:8e:79:00:00:01",
      "d0:8e:79:00:00:02",
    ]);
  });

  it("fills pasted values down only over the rows it is given", () => {
    const r = rows([a01]);
    const shown = r.filter((x) => x.key !== "swe-a01:Ethernet2");
    const o = fillNames(shown, "swe-a01:Ethernet1", ["n1", "n2"], {});
    expect(o).toEqual({ "swe-a01:Ethernet1": { name: "n1" }, "swe-a01:Ethernet3": { name: "n2" } });
  });

  it("fills pasted names down over included rows", () => {
    const left = { "swe-a01:Ethernet49/1": { include: false } };
    const o = fillNames(rows([a01, a02], {}, left), "swe-a01:Ethernet3", ["x", "y", "z"], left);
    expect(o).toEqual({
      "swe-a01:Ethernet3": { name: "x" },
      "swe-a01:Ethernet49/1": { include: false },
      "swe-a02:Ethernet1": { name: "y" },
    });
  });

  it("tallies and builds the request", () => {
    const r = rows([a01, a02], { fqdn: "bmc-{name}" });
    expect(tally(r)).toMatchObject({ included: 5, ready: 2, attention: 1, empty: 2, excluded: 0 });
    expect(toImportRows(r, { ifname: "bmc", mtu: "9000", portVlan: "3001" })).toEqual([
      {
        node: "cpn-a01-01",
        mac: "d0:8e:79:00:00:01",
        bmc: true,
        fqdn: "bmc-cpn-a01-01",
        ifname: "bmc",
        vlan: "3001",
        mtu: 9000,
        tags: undefined,
      },
      {
        node: "cpn-a02-01",
        mac: "d0:8e:79:00:01:01",
        bmc: true,
        fqdn: "bmc-cpn-a02-01",
        ifname: "bmc",
        vlan: "3001",
        mtu: 9000,
        tags: undefined,
      },
    ]);
    expect(toImportRows(r, { ifname: "", mtu: "", portVlan: "" })[0].mtu).toBeUndefined();
  });

  it("reads port filters", () => {
    const p = parsePortFilter("1-4, 9,12-10");
    expect([1, 4, 5, 9, 10, 11, 12, 13].map((n) => p.test(n, `Ethernet${n}`))).toEqual([
      true,
      true,
      false,
      true,
      true,
      true,
      true,
      false,
    ]);
    const named = parsePortFilter("ethernet4, 2");
    expect(named.test(40, "Ethernet40")).toBe(true);
    expect(named.test(2, "Ethernet2")).toBe(true);
    expect(named.test(3, "Ethernet3")).toBe(false);
    expect(parsePortFilter("1-40").test(undefined, "Management1")).toBe(false);
  });
});
