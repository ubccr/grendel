import { describe, expect, it } from "vitest";
import { incompleteTemplates, renameTemplate } from "../images/form";
import { checkNode, emptyPort } from "../nodes/form";

describe("checkNode", () => {
  it("accepts a named node whose ports each have something set", () => {
    expect(
      checkNode({
        value: { name: "cpn-01", interfaces: [{ mac: "aa:bb" }, { ip: "10.0.0.1/24" }] },
      }),
    ).toBeUndefined();
  });

  it("refuses a node without a name on its name field", () => {
    expect(checkNode({ value: { name: "  " } })).toEqual({
      form: undefined,
      fields: { name: "Enter a name." },
    });
  });

  it("refuses an empty port as a form error", () => {
    expect(
      checkNode({
        value: { name: "cpn-01", interfaces: [{ mac: "aa:bb" }, { vlan: "10", mtu: 9000 }] },
      }),
    ).toEqual({
      form: "Fill in or remove the empty ports.",
      fields: {},
    });
  });
});

describe("emptyPort", () => {
  it("counts only the name, MAC, IP and FQDN, ignoring blanks", () => {
    expect(emptyPort({})).toBe(true);
    expect(emptyPort({ ifname: " ", vlan: "10", bmc: true })).toBe(true);
    expect(emptyPort({ fqdn: "cpn-01.example.org" })).toBe(false);
  });
});

describe("incompleteTemplates", () => {
  it("refuses a template missing its name or its file", () => {
    expect(incompleteTemplates({ value: { kickstart: "rocky.ks.tmpl" } })).toBeUndefined();
    expect(incompleteTemplates({ value: { "": "rocky.ks.tmpl" } })).toBeDefined();
    expect(incompleteTemplates({ value: { kickstart: " " } })).toBeDefined();
  });
});

describe("renameTemplate", () => {
  it("renames a template in place", () => {
    const next = renameTemplate({ ipxe: "a.tmpl", kickstart: "b.tmpl" }, "ipxe", "ipxe2");
    expect(Object.entries(next ?? {})).toEqual([
      ["ipxe2", "a.tmpl"],
      ["kickstart", "b.tmpl"],
    ]);
  });

  it("refuses a name another template already has", () => {
    expect(
      renameTemplate({ ipxe: "a.tmpl", kickstart: "b.tmpl" }, "ipxe", "kickstart"),
    ).toBeUndefined();
  });
});
