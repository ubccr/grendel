import { describe, expect, it } from "vitest";
import {
  blockOf,
  intToIp,
  ipToInt,
  lastOctet,
  parseRange,
  rangeOf,
  stepRange,
  stripPrefix,
} from "../addresses";

describe("ipToInt", () => {
  it("reads dotted addresses", () => {
    expect(ipToInt("10.17.25.7")).toBe(0x0a111907);
    expect(ipToInt("255.255.255.255")).toBe(0xffffffff);
    expect(intToIp(0xffffffff)).toBe("255.255.255.255");
  });

  it("refuses anything else", () => {
    for (const s of ["", "10.17.25", "10.17.25.256", "10.17.25.x", "1.2.3.4.5"]) {
      expect(ipToInt(s)).toBeUndefined();
    }
  });
});

describe("parseRange", () => {
  it("keeps a range to its network", () => {
    expect(parseRange("10.17.25.0/24")).toBe("10.17.25.0/24");
    expect(parseRange(" 10.17.25.77/26 ")).toBe("10.17.25.64/26");
    expect(parseRange("10.17.200.1/16")).toBe("10.17.0.0/16");
  });

  it("reads an address, or three octets, as its /24", () => {
    expect(parseRange("10.17.25.7")).toBe("10.17.25.0/24");
    expect(parseRange("10.17.25")).toBe("10.17.25.0/24");
  });

  it("refuses ranges larger than a /16, and what isn't a range", () => {
    for (const s of ["10.0.0.0/8", "10.17.25.0/33", "10.17", "", "10.17.25.0/", "fd00::/64"]) {
      expect(parseRange(s)).toBeUndefined();
    }
  });
});

describe("rangeOf", () => {
  it("searches a rack's /24 of a larger subnet", () => {
    expect(rangeOf("10.17.25.7/18")).toBe("10.17.25.0/24");
    expect(rangeOf("10.17.25.7")).toBe("10.17.25.0/24");
  });

  it("searches a small subnet whole", () => {
    expect(rangeOf("10.17.45.70/25")).toBe("10.17.45.0/25");
  });

  it("searches the /24 of a single host or point to point address", () => {
    expect(rangeOf("10.17.0.1/32")).toBe("10.17.0.0/24");
    expect(rangeOf("10.17.0.1/31")).toBe("10.17.0.0/24");
  });

  it("has nothing to suggest without an address", () => {
    expect(rangeOf("")).toBeUndefined();
    expect(rangeOf(undefined)).toBeUndefined();
  });
});

describe("blockOf", () => {
  it("shows a small range whole", () => {
    expect(blockOf("10.17.25.64/26")).toEqual({ start: ipToInt("10.17.25.64"), size: 64 });
  });

  it("shows the /24 of a larger range holding the focus", () => {
    expect(blockOf("10.17.0.0/18", "10.17.30.9")).toEqual({
      start: ipToInt("10.17.30.0"),
      size: 256,
    });
    expect(blockOf("10.17.0.0/18", "10.99.0.1")).toEqual({
      start: ipToInt("10.17.0.0"),
      size: 256,
    });
  });
});

describe("stripPrefix and lastOctet", () => {
  it("name addresses briefly", () => {
    expect(stripPrefix("10.17.25.7/18")).toBe("10.17.25.7");
    expect(lastOctet("10.17.25.214")).toBe(".214");
  });
});

describe("stepRange", () => {
  it("moves to the range of the same size after or before", () => {
    expect(stepRange("10.17.25.0/24", 1)).toBe("10.17.26.0/24");
    expect(stepRange("10.17.25.0/24", -1)).toBe("10.17.24.0/24");
    expect(stepRange("10.17.25.64/26", 1)).toBe("10.17.25.128/26");
    expect(stepRange("10.17.255.0/24", 1)).toBe("10.18.0.0/24");
  });

  it("stops at the ends of IPv4", () => {
    expect(stepRange("0.0.0.0/24", -1)).toBeUndefined();
    expect(stepRange("255.255.255.0/24", 1)).toBeUndefined();
  });
});
