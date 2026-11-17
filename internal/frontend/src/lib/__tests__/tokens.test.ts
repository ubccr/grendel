import { describe, expect, it } from "vitest";
import {
  expirySeconds,
  expiryLabel,
  expiryOptions,
  parseExpiry,
  tokenClaims,
  tokenParts,
  tokenRoles,
} from "../tokens";

describe("parseExpiry", () => {
  it("passes durations the server takes through", () => {
    expect(parseExpiry("12h")).toBe("12h");
    expect(parseExpiry("1h30m")).toBe("1h30m");
  });

  it("turns days and weeks into hours", () => {
    expect(parseExpiry("45d")).toBe("1080h");
    expect(parseExpiry("2 weeks")).toBe("336h");
    expect(parseExpiry("1d 6h")).toBe("30h");
  });

  it("carries minutes and seconds over into larger units", () => {
    expect(parseExpiry("90m")).toBe("1h30m");
    expect(parseExpiry("3601s")).toBe("1h1s");
  });

  it("reads never as no expiry", () => {
    expect(parseExpiry("Never")).toBe("infinite");
    expect(parseExpiry("infinite")).toBe("infinite");
  });

  it("refuses what it can't read, and a duration of nothing", () => {
    expect(parseExpiry("soon")).toBeUndefined();
    expect(parseExpiry("12")).toBeUndefined();
    expect(parseExpiry("3 fortnights")).toBeUndefined();
    expect(parseExpiry("1.5h")).toBeUndefined();
    expect(parseExpiry("0h")).toBeUndefined();
    expect(parseExpiry("")).toBeUndefined();
  });
});

describe("expiryLabel", () => {
  it("names presets", () => {
    expect(expiryLabel("8760h")).toBe("1 year");
    expect(expiryLabel("infinite")).toBe("Never");
  });

  it("spells out other durations", () => {
    expect(expiryLabel("1080h")).toBe("45 days");
    expect(expiryLabel("1h30m")).toBe("1 hour 30 minutes");
    expect(expiryLabel("30h")).toBe("1 day 6 hours");
  });
});

describe("expiryOptions", () => {
  it("shows every preset while the input shows the chosen expiry", () => {
    expect(expiryOptions("30 days", "720h")).toHaveLength(7);
    expect(expiryOptions("", "720h")).toHaveLength(7);
  });

  it("keeps a chosen custom expiry in the list", () => {
    expect(expiryOptions("45 days", "1080h")[0]).toBe("1080h");
    expect(expiryOptions("45 days", "1080h")).toHaveLength(8);
  });

  it("offers what was typed first, then matching presets", () => {
    expect(expiryOptions("45d", "720h")).toEqual(["1080h"]);
    expect(expiryOptions("day", "720h")).toEqual(["24h", "720h", "2160h"]);
    expect(expiryOptions("4", "720h")).toEqual([]);
  });

  it("doesn't offer a typed duration twice when it is a preset", () => {
    expect(expiryOptions("30d", "1h")).toEqual(["720h"]);
  });
});

describe("labels", () => {
  it("read back as the expiry they name", () => {
    for (const v of ["1h", "8h", "24h", "720h", "8760h", "infinite", "1080h", "1h30m", "30h1s"]) {
      expect(parseExpiry(expiryLabel(v))).toBe(v);
    }
  });
});

describe("expirySeconds", () => {
  it("reads the durations the API takes", () => {
    expect(expirySeconds("1h")).toBe(3600);
    expect(expirySeconds("1h30m")).toBe(5400);
    expect(expirySeconds("90s")).toBe(90);
  });

  it("reads days, weeks and years too", () => {
    expect(expirySeconds("2d")).toBe(172_800);
    expect(expirySeconds("1 week")).toBe(604_800);
    expect(expirySeconds("1y")).toBe(31_536_000);
  });

  it("refuses what it can't read", () => {
    expect(expirySeconds("1.5h")).toBeUndefined();
    expect(expirySeconds("")).toBeUndefined();
    expect(expirySeconds("infinite")).toBeUndefined();
  });
});

const perm = (path: string) => ({ method: "GET", path });

describe("tokenRoles", () => {
  const roles = [
    { name: "admin", permission_list: [perm("/a"), perm("/b"), perm("/c")] },
    { name: "user", permission_list: [perm("/a"), perm("/b")] },
    { name: "read-only", permission_list: [perm("/a")] },
    { name: "odd", permission_list: [perm("/a"), perm("/z")] },
  ];

  it("offers the user's role first, then the roles inside it", () => {
    expect(tokenRoles(roles, "user")).toEqual(["user", "read-only"]);
    expect(tokenRoles(roles, "admin")).toEqual(["admin", "user", "read-only"]);
  });

  it("leaves out a role holding a permission the user lacks", () => {
    expect(tokenRoles(roles, "admin")).not.toContain("odd");
  });

  it("falls back to the user's role when the roles don't include it", () => {
    expect(tokenRoles([], "user")).toEqual(["user"]);
  });
});

// a token signed by the server for admin with role user, expiring at 1800000000
const token =
  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE4MDAwMDAwMDAsInJvbGUiOiJ1c2VyIiwidXNlcm5hbWUiOiJhZG1pbiJ9.c2lnbmF0dXJl";

describe("tokenClaims", () => {
  it("reads a name outside ASCII", () => {
    const claims = btoa(String.fromCharCode(...new TextEncoder().encode('{"username":"zoë"}')));
    expect(tokenClaims(`h.${claims.replace(/=+$/, "")}.s`).username).toBe("zoë");
  });

  it("reads the claims the server wrote", () => {
    expect(tokenClaims(token)).toEqual({ exp: 1800000000, role: "user", username: "admin" });
  });

  it("gives no claims for a token it can't read", () => {
    expect(tokenClaims("not a token")).toEqual({});
  });
});

describe("tokenParts", () => {
  it("splits a token at its dots", () => {
    expect(tokenParts("a.b.c")).toEqual({ header: "a", claims: "b", signature: "c" });
  });
});
