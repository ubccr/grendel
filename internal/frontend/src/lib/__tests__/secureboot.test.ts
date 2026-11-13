import { describe, expect, it } from "vitest";
import {
  coverage,
  derToPem,
  expiry,
  fingerprint,
  hashName,
  keysOf,
  readCertificateFile,
  splitPem,
} from "../secureboot";

// a tiny DER sequence is enough to exercise the encoding; the server is what parses certificates
const der = new Uint8Array([0x30, 0x03, 0x02, 0x01, 0x05]);

describe("pem", () => {
  it("round trips DER through PEM and finds every certificate in a bundle", () => {
    const pem = derToPem(der);
    expect(pem).toBe("-----BEGIN CERTIFICATE-----\nMAMCAQU=\n-----END CERTIFICATE-----\n");
    const bundle = `${pem}-----BEGIN PRIVATE KEY-----\nabc\n-----END PRIVATE KEY-----\n${pem}`;
    expect(splitPem(bundle)).toEqual([pem, pem]);
  });

  it("reads a DER file, and nothing from a file that is neither", () => {
    expect(readCertificateFile(der)).toEqual([derToPem(der)]);
    expect(readCertificateFile(new TextEncoder().encode("hello"))).toEqual([]);
  });

  it("fingerprints the DER inside the PEM", async () => {
    const fp = await fingerprint(derToPem(der));
    const sum = await crypto.subtle.digest("SHA-256", der);
    const hex = Array.from(new Uint8Array(sum), (b) => b.toString(16).padStart(2, "0")).join("");
    expect(fp).toBe(hex);
  });
});

describe("expiry", () => {
  const now = Date.parse("2026-11-13T00:00:00Z");
  it("flags expired and expiring certificates", () => {
    expect(expiry("2026-01-01T00:00:00Z", now)).toBe("expired");
    expect(expiry("2026-12-01T00:00:00Z", now)).toBe("soon");
    expect(expiry("2030-01-01T00:00:00Z", now)).toBe("ok");
    expect(expiry("", now)).toBe("unknown");
  });
});

describe("hashName", () => {
  it("names a signature type by its hash", () => {
    expect(hashName("EFI_CERT_SHA256_GUID")).toBe("SHA-256");
    expect(hashName("EFI_CERT_X509_SHA256_GUID")).toBe("X.509 SHA-256");
    expect(hashName("")).toBe("Hash");
  });
});

describe("coverage", () => {
  const db = (certs: object[], sigs: object[] = []) => ({
    keys: [
      { id: "dbx", certificates: [], signatures: sigs },
      { id: "db", certificates: certs, signatures: [] },
    ],
  });

  it("groups a key by fingerprint across nodes whatever id each gives it", () => {
    const rows = coverage([
      { node: "a", success: true, data: db([{ id: "3", fingerprint: "ff", subject: "Grendel" }]) },
      {
        node: "b",
        success: true,
        data: db([{ id: "9", fingerprint: "ff", subject: "Grendel" }], [{ id: "1", hash: "AB" }]),
      },
      { node: "c", success: false, message: "unreachable" },
    ]);
    expect(rows.map((r) => [r.db, r.fingerprint, r.nodes])).toEqual([
      ["db", "ff", ["a", "b"]],
      ["dbx", "ab", ["b"]],
    ]);
  });

  it("keeps a key with no fingerprint to its own node", () => {
    const rows = coverage([
      { node: "a", success: true, data: db([{ id: "1" }]) },
      { node: "b", success: true, data: db([{ id: "1" }]) },
    ]);
    expect(rows.map((r) => r.nodes)).toEqual([["a"], ["b"]]);
  });

  it("flattens certificates and hashes into keys", () => {
    const keys = keysOf(
      db([{ id: "1", fingerprint: "ff" }], [{ id: "2", hash: "AB", type: "t" }]).keys,
    );
    expect(keys.map((k) => [k.db, k.kind, k.fingerprint])).toEqual([
      ["dbx", "hash", "ab"],
      ["db", "cert", "ff"],
    ]);
  });
});
