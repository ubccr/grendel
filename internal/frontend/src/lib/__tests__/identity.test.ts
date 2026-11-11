import { assetCsv, formatLocation, guessColumns, parseCsv } from "@/lib/identity";
import { describe, expect, it } from "vitest";

describe("formatLocation", () => {
  it("puts rack and unit first", () => {
    expect(formatLocation({ rack: "f01", rack_unit: 4, datacenter: "VTC", room: "Suite 3" })).toBe(
      "f01 U4 (VTC/Suite 3)",
    );
    expect(formatLocation({ datacenter: "VTC" })).toBe("VTC");
    expect(formatLocation({ rack: "f01", rack_unit: null })).toBe("f01");
    expect(formatLocation(undefined)).toBe("");
  });
});

describe("parseCsv", () => {
  it("handles quotes, commas and line endings", () => {
    expect(parseCsv('a,"b, c"\r\n"say ""hi""",d\n')).toEqual([
      ["a", "b, c"],
      ['say "hi"', "d"],
    ]);
  });
  it("drops a byte order mark", () => {
    expect(parseCsv("\ufeffSerial,Tag")).toEqual([["Serial", "Tag"]]);
  });
});

describe("assetCsv", () => {
  it("finds columns by header", () => {
    const rows = parseCsv("rack,Asset Tag,Service_Tag\nf01,UB-1,ABC123\n");
    const g = guessColumns(rows);
    expect(g).toEqual({ header: true, serialIndex: 2, tagIndex: 1 });
    expect(assetCsv(rows, g.header, g.serialIndex, g.tagIndex).bySerial).toEqual({
      ABC123: "UB-1",
    });
  });

  it("reads a file with no header as serial then tag", () => {
    const rows = parseCsv("ABC123,UB-1\nDEF456,UB-2");
    const g = guessColumns(rows);
    expect(g.header).toBe(false);
    expect(assetCsv(rows, g.header, g.serialIndex, g.tagIndex).bySerial).toEqual({
      ABC123: "UB-1",
      DEF456: "UB-2",
    });
  });

  it("asks for columns the header does not name", () => {
    const rows = parseCsv("serial,inventory\nABC123,UB-1");
    const g = guessColumns(rows);
    expect(assetCsv(rows, g.header, g.serialIndex, g.tagIndex).error).toMatch(/Pick/);
    expect(assetCsv(rows, true, 0, 1).bySerial).toEqual({ ABC123: "UB-1" });
  });

  it("rejects a duplicate serial", () => {
    const rows = parseCsv("ABC123,UB-1\nABC123,UB-2");
    expect(assetCsv(rows, false, 0, 1).error).toMatch(/listed twice/);
  });
});
