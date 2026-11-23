import { describe, expect, it } from "vitest";
import { nameProblem, parseTemplate, renderTemplate, switchDigits } from "../name-template";

const render = (src: string, vars: Parameters<typeof renderTemplate>[1]) =>
  renderTemplate(parseTemplate(src), vars);

describe("name templates", () => {
  it("fills in port, switch digits and padding", () => {
    expect(render("cpn-a{sw.1}-{port:02}", { port: 7, sw: "swe-a01" }).value).toBe("cpn-a01-07");
    expect(render("cpn-{sw}-{port}", { port: 12, sw: "swe-a01.cluster" }).value).toBe(
      "cpn-swe-a01-12",
    );
  });

  it("does arithmetic on numbers, rounding division up", () => {
    expect(render("{port+2}", { port: 1 }).value).toBe("3");
    expect(render("{port-1:02}", { port: 1 }).value).toBe("00");
    expect(render("{port*2}", { port: 4 }).value).toBe("8");
    expect(render("{port/2}", { port: 1 }).value).toBe("1");
    expect(render("{port/2}", { port: 2 }).value).toBe("1");
    expect(render("{port/2}", { port: 3 }).value).toBe("2");
    expect(render("r{sw.2+10}", { sw: "swe-a01-r05" }).value).toBe("r15");
  });

  it("refuses arithmetic that goes below 0 rather than writing a minus into the name", () => {
    expect(render("cpn-{port-2}", { port: 2 }).value).toBe("cpn-0");
    expect(render("cpn-{port-2}", { port: 1 }).error).toMatch(/\{port-2\} is below 0/);
    expect(render("r{sw.1-5}", { sw: "swe-a01" }).error).toMatch(/\{sw\.1-5\} is below 0/);
  });

  it("keeps switch digits as written unless asked to do arithmetic", () => {
    expect(render("{sw.1}", { sw: "sw-007" }).value).toBe("007");
    expect(render("{sw.1+0}", { sw: "sw-007" }).value).toBe("7");
    expect(switchDigits("swe-a01-r12.cluster.local")).toEqual(["01", "12"]);
  });

  it("swaps text in the name, as for a BMC named after its node", () => {
    expect(render("{name/cpn/bmc}.cluster", { name: "cpn-blah" }).value).toBe("bmc-blah.cluster");
    expect(render("{name/cpn-/}", { name: "cpn-a01-07" }).value).toBe("a01-07");
    expect(render("{name/cpn/bmc}", { name: "gpu-special" }).value).toBe("gpu-special");
    expect(render("{name/a/b}", { name: "a-a" }).value).toBe("b-a");
    expect(render("{sw/swe/cpn}-{port}", { sw: "swe-a01", port: 3 }).value).toBe("cpn-a01-3");
    expect(parseTemplate("{port/cpn/bmc}").error).toBeDefined();
    expect(render("{port/2}", { port: 3 }).value).toBe("2");
  });

  it("counts with {n}", () => {
    expect(render("node{n:03}", { n: 5 }).value).toBe("node005");
  });

  it("builds on the node name", () => {
    expect(render("bmc-{name}.cluster", { name: "cpn-a01-07" }).value).toBe(
      "bmc-cpn-a01-07.cluster",
    );
    expect(render("bmc-{name}", {}).error).toMatch(/empty/);
  });

  it("points at the mistake", () => {
    expect(parseTemplate("cpn-{rack}").error).toEqual({
      message: expect.stringMatching(/isn't a field/),
      at: 4,
    });
    expect(parseTemplate("cpn-{port").error?.at).toBe(4);
    expect(parseTemplate("cpn}").error?.at).toBe(3);
    expect(parseTemplate("{sw+1}").error?.message).toMatch(/can't take arithmetic/);
    expect(parseTemplate("{port.1}").error?.message).toMatch(/Only \{sw\}/);
    expect(parseTemplate("{port/0}").error?.message).toMatch(/divide by 0/);
  });

  it("explains a switch name without the digits asked for", () => {
    expect(render("{sw.2}", { sw: "swe-a01" }).error).toMatch(/no number 2/);
  });

  it("lists the fields it uses", () => {
    expect([...parseTemplate("{sw.1}-{port}-{n}").uses]).toEqual(["sw.1", "port", "n"]);
  });

  it("refuses names DNS and nodesets can't hold", () => {
    expect(nameProblem("cpn-a01-07")).toBeUndefined();
    expect(nameProblem("cpn a01")).toBeDefined();
    expect(nameProblem("cpn[1-2]")).toBeDefined();
    expect(nameProblem("-cpn")).toBeDefined();
    expect(nameProblem("")).toBeDefined();
  });
});
