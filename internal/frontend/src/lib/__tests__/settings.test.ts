import {
  afterSave,
  entryRemovals,
  entryRows,
  groups,
  hasStored,
  isLocked,
  isSet,
  isWritable,
  refusals,
  type Setting,
  show,
  summarize,
  toPatch,
  withEdit,
  withoutEntry,
} from "@/lib/settings";
import { describe, expect, it } from "vitest";

const ttl: Setting = { key: "dns.ttl", kind: "int", value: 300, default: 300, source: "default" };
const stored: Setting = { ...ttl, value: 600, stored: 600, source: "database" };
const locked: Setting = { ...stored, value: 900, source: "file", shadowed: true };
const vars: Setting = {
  key: "provision.extra_vars",
  kind: "map",
  value: { timezone: "UTC" },
  stored: { timezone: "UTC" },
  default: {},
  source: "database",
  file_keys: ["root_password"],
};
const subnets: Setting = {
  key: "dhcp.subnets",
  kind: "group",
  entry_label: "Subnet",
  apply: "live",
  fields: [
    { name: "gateway", summary: "Gateway", kind: "string" },
    { name: "dns", summary: "DNS servers", kind: "string_list" },
  ],
};

describe("state derived from where a value comes from", () => {
  it("locks a setting a stronger layer sets, but never one made of entries", () => {
    expect(isLocked(locked)).toBe(true);
    expect(isLocked({ ...ttl, source: "env" })).toBe(true);
    expect(isLocked(stored)).toBe(false);
    expect(isLocked({ ...vars, source: "file" })).toBe(false);
  });

  it("tells set from default, and stored from not", () => {
    expect(isSet(ttl)).toBe(false);
    expect(isSet({ ...ttl, source: "inherited" })).toBe(false);
    expect(isSet(stored)).toBe(true);
    expect(hasStored(ttl)).toBe(false);
    expect(hasStored({ ...ttl, stored: "" })).toBe(true);
  });

  it("keeps secrets and file-only settings out of the database", () => {
    expect(isWritable(ttl)).toBe(true);
    expect(isWritable({ ...ttl, secret: true })).toBe(false);
    expect(isWritable({ ...ttl, external: true })).toBe(false);
  });
});

describe("withEdit", () => {
  it("keeps an edit that changes the value", () => {
    expect(withEdit({}, ttl, "600")).toEqual({ "dns.ttl": "600" });
  });

  it("drops an edit typed back to the current value, which is a number where the edit is text", () => {
    expect(withEdit({ "dns.ttl": "600" }, ttl, "300")).toEqual({});
  });

  it("only removes a setting the database holds", () => {
    expect(withEdit({}, ttl, null)).toEqual({});
    expect(withEdit({}, stored, null)).toEqual({ "dns.ttl": null });
  });

  it("does not count a map with its keys in another order as changed", () => {
    const two = { ...vars, stored: { a: 1, b: 2 } };
    expect(withEdit({}, two, { b: 2, a: 1 })).toEqual({});
    expect(withEdit({}, two, { b: 3, a: 1 })).toEqual({ "provision.extra_vars": { b: 3, a: 1 } });
  });

  it("treats a list with nothing in it as an unset list", () => {
    const dns: Setting = { key: "dhcp.dns_servers", kind: "string_list", source: "default" };
    expect(withEdit({}, dns, [])).toEqual({});
    expect(withEdit({}, dns, ["10.0.0.2"])).toEqual({ "dhcp.dns_servers": ["10.0.0.2"] });
  });
});

describe("afterSave", () => {
  it("returns a removed setting to its default", () => {
    expect(afterSave(stored, { "dns.ttl": null })).toBe(300);
  });

  it("leaves an overridden setting running with the override", () => {
    expect(afterSave(locked, { "dns.ttl": null })).toBe(900);
  });

  it("edits only the database's entries of a merged setting", () => {
    expect(afterSave(vars, {})).toEqual({ timezone: "UTC" });
    expect(afterSave(vars, { "provision.extra_vars": null })).toEqual({});
  });
});

describe("toPatch", () => {
  it("splits values to store from keys to remove", () => {
    expect(toPatch({ "dns.ttl": "600", "dhcp.dns_servers": ["a"], "dns.forward": null })).toEqual({
      set: { "dns.ttl": "600", "dhcp.dns_servers": ["a"] },
      unset: ["dns.forward"],
    });
  });
});

describe("refusals", () => {
  it("maps each refused key to its reason", () => {
    expect(
      refusals({ errors: [{ name: "dns.ttl", reason: "must be at least 0" }, { reason: "x" }] }),
    ).toEqual({ "dns.ttl": "must be at least 0" });
    expect(refusals(undefined)).toEqual({});
  });
});

describe("show and summarize", () => {
  it("writes a value the way it is typed", () => {
    expect(show(300)).toBe("300");
    expect(show(true)).toBe("true");
    expect(show(undefined)).toBe("");
    expect(show({ b: 1, a: 2 })).toBe('{"a":2,"b":1}');
  });

  it("names the items of a list and the keys of a map", () => {
    expect(summarize({ timezone: "UTC", ntp: [] })).toBe("timezone, ntp");
    expect(summarize({})).toBe("none");
    expect(summarize(["a", "b"])).toBe("a, b");
    expect(summarize("")).toBe("empty");
    expect(summarize(600)).toBe("600");
  });
});

describe("entries", () => {
  it("gives a new entry one empty setting per field, keyed like a saved one", () => {
    const rows = entryRows(subnets, "20-compute");
    expect(rows.map((r) => r.key)).toEqual([
      "dhcp.subnets.20-compute.gateway",
      "dhcp.subnets.20-compute.dns",
    ]);
    expect(rows[0]).toMatchObject({
      group: "Subnet 20-compute",
      entry_of: "dhcp.subnets",
      entry: "20-compute",
    });
    expect(isWritable(rows[0])).toBe(true);
    expect(withEdit({}, rows[1], ["10.0.0.2"])).toEqual({
      "dhcp.subnets.20-compute.dns": ["10.0.0.2"],
    });
  });

  it("removes a saved entry whole and forgets an unsaved one", () => {
    const draft = { "dhcp.subnets.a.mtu": "9000", "dns.ttl": "600" };
    expect(withoutEntry(draft, "dhcp.subnets.a", true)).toEqual({
      "dns.ttl": "600",
      "dhcp.subnets.a": null,
    });
    expect(withoutEntry(draft, "dhcp.subnets.a", false)).toEqual({ "dns.ttl": "600" });
  });

  it("lists each removed entry once for the review", () => {
    const saved = entryRows(subnets, "a");
    const removals = entryRemovals([subnets, ...saved], { "dhcp.subnets.a": null });
    expect(removals.map((r) => [r.key, r.summary])).toEqual([["dhcp.subnets.a", "Subnet a"]]);
    expect(entryRemovals([subnets, ...saved], {})).toEqual([]);
  });
});

describe("groups", () => {
  it("keeps settings under their subheading in listed order", () => {
    const g = groups([
      { key: "a" },
      { key: "b", group: "gNMI" },
      { key: "c" },
      { key: "d", group: "gNMI" },
    ]);
    expect(g.map((x) => [x.name, x.options.map((o) => o.key)])).toEqual([
      ["", ["a", "c"]],
      ["gNMI", ["b", "d"]],
    ]);
  });
});
