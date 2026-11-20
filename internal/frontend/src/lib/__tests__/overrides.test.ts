import { patchOverride, withIncluded } from "@/lib/overrides";
import { describe, expect, it } from "vitest";

type O = { include?: boolean; mac?: string };

describe("patchOverride", () => {
  it("merges a patch into the entry", () => {
    expect(patchOverride<O>({ a: { mac: "x" } }, "a", { include: false })).toEqual({
      a: { mac: "x", include: false },
    });
  });

  it("drops fields set to undefined and entries left empty", () => {
    expect(patchOverride<O>({ a: { mac: "x" }, b: { mac: "y" } }, "a", { mac: undefined })).toEqual(
      {
        b: { mac: "y" },
      },
    );
  });
});

describe("withIncluded", () => {
  const rows = [
    { key: "on", byDefault: true },
    { key: "off", byDefault: false },
  ];

  it("keeps include only where it differs from the row's default", () => {
    expect(withIncluded<O>({}, rows, new Set(["off"]))).toEqual({
      on: { include: false },
      off: { include: true },
    });
    expect(withIncluded<O>({ on: { include: false, mac: "m" } }, rows, new Set(["on"]))).toEqual({
      on: { mac: "m" },
    });
  });
});
