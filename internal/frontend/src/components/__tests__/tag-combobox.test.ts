import { describe, expect, it } from "vitest";
import { groupTags } from "../tag-combobox";

describe("groupTags", () => {
  it("groups key=value tags under their key, sorted, with key only tags last", () => {
    expect(groupTags(["rack=f01", "gpu", "brand=dell", "rack=b05", "dell"])).toEqual([
      { value: "brand", items: ["brand=dell"] },
      { value: "rack", items: ["rack=b05", "rack=f01"] },
      { value: "", items: ["dell", "gpu"] },
    ]);
  });

  it("groups a namespaced tag under its namespace, even before an =", () => {
    expect(groupTags(["grendel:brand=dell", "grendel:managed", "note=a=b"])).toEqual([
      { value: "grendel", items: ["grendel:brand=dell", "grendel:managed"] },
      { value: "note", items: ["note=a=b"] },
    ]);
  });

  it("treats a colon after the = as part of the value", () => {
    expect(groupTags(["url=http://host"])).toEqual([{ value: "url", items: ["url=http://host"] }]);
  });

  it("puts a tag starting with = with the key only tags", () => {
    expect(groupTags(["=value"])).toEqual([{ value: "", items: ["=value"] }]);
  });
});
