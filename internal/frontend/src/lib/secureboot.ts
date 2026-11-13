import type { BmcResponseModelBmcSecureBoot, BmcSecureBootKeyBody } from "@/client";

export type SecureBootResult = NonNullable<BmcResponseModelBmcSecureBoot["results"]>[number];
type Database = NonNullable<NonNullable<SecureBootResult["data"]>["keys"]>[number];

// writable are the databases Grendel changes, in the order trust flows: PK governs KEK, which governs db and dbx
export const writable = ["PK", "KEK", "db", "dbx"] as const;
export type Writable = (typeof writable)[number];

export type SignatureType = NonNullable<BmcSecureBootKeyBody["signature_type"]>;

// hashBytes is how long each signature type's data is, matching what the server accepts. Its keys are the API's enum, so a type added there fails to compile here until it is given a length
export const hashBytes: Record<SignatureType, number> = {
  EFI_CERT_SHA256_GUID: 32,
  EFI_CERT_SHA384_GUID: 48,
  EFI_CERT_SHA512_GUID: 64,
  EFI_CERT_X509_SHA256_GUID: 48,
};

export const signatureTypes = Object.keys(hashBytes) as SignatureType[];

// Key is one certificate or hash in one database, shaped the same whichever it is so a list can hold both
export type Key = {
  db: string;
  kind: "cert" | "hash";
  id: string;
  // fingerprint is a certificate's SHA-256, or a hash itself, which is the same on every node and is what removing a key takes. Empty when the bmc reports neither
  fingerprint: string;
  subject: string;
  issuer: string;
  notAfter: string;
  type: string;
};

// keysOf flattens one node's databases into keys, in the order the databases are listed
export function keysOf(databases: Database[] | undefined): Key[] {
  return (databases ?? []).flatMap((d) => {
    const db = d.id ?? "";
    const certs = (d.certificates ?? []).map((c): Key => ({
      db,
      kind: "cert",
      id: c.id ?? "",
      fingerprint: c.fingerprint ?? "",
      subject: c.subject ?? "",
      issuer: c.issuer ?? "",
      notAfter: c.not_after ?? "",
      type: "",
    }));
    const sigs = (d.signatures ?? []).map((s): Key => ({
      db,
      kind: "hash",
      id: s.id ?? "",
      fingerprint: (s.hash ?? "").toLowerCase(),
      subject: "",
      issuer: "",
      notAfter: "",
      type: s.type ?? "",
    }));
    return [...certs, ...sigs];
  });
}

export type Expiry = "expired" | "soon" | "ok" | "unknown";

const soonMs = 90 * 24 * 60 * 60 * 1000;

// expiry says whether a certificate has run out, or does within 90 days. UEFI firmware does not check expiry, so this is a prompt to replace a key, not a sign it stopped working
export function expiry(notAfter: string, now = Date.now()): Expiry {
  const t = Date.parse(notAfter);
  if (!notAfter || Number.isNaN(t)) return "unknown";
  if (t < now) return "expired";
  if (t - now < soonMs) return "soon";
  return "ok";
}

// hashName names a signature type by its hash, so EFI_CERT_SHA256_GUID reads as SHA-256
export function hashName(type: string) {
  const m = /^EFI_CERT_(X509_)?SHA(\d+)_GUID$/.exec(type);
  if (!m) return type || "Hash";
  return `${m[1] ? "X.509 " : ""}SHA-${m[2]}`;
}

// Coverage is one key across a nodeset: which nodes hold it and the id each one gives it
export type Coverage = Key & { nodes: string[] };

// coverage groups the keys of every node that answered by database and fingerprint, so a key enrolled on some nodes but not others shows as one row. A key with no fingerprint cannot be matched across nodes, and stays a row of its own
export function coverage(results: SecureBootResult[]): Coverage[] {
  const groups = new Map<string, Coverage>();
  for (const r of results) {
    if (!r.success) continue;
    const node = r.node ?? "";
    for (const k of keysOf(r.data?.keys)) {
      const id = k.fingerprint ? `${k.db}/${k.fingerprint}` : `${k.db}/${node}/${k.id}`;
      const group = groups.get(id);
      if (group) group.nodes.push(node);
      else groups.set(id, { ...k, nodes: [node] });
    }
  }
  const order = (db: string) => {
    const i = writable.indexOf(db as Writable);
    return i === -1 ? writable.length : i;
  };
  return [...groups.values()].sort(
    (a, b) =>
      order(a.db) - order(b.db) ||
      a.db.localeCompare(b.db) ||
      a.kind.localeCompare(b.kind) ||
      (a.subject || a.fingerprint).localeCompare(b.subject || b.fingerprint),
  );
}

const pemBlock = /-----BEGIN CERTIFICATE-----([\s\S]*?)-----END CERTIFICATE-----/g;

// splitPem finds every certificate in a PEM text, skipping keys and anything else in it
export function splitPem(text: string): string[] {
  return [...text.matchAll(pemBlock)].map((m) => toPem(m[1].replace(/\s+/g, "")));
}

function toPem(base64: string) {
  const lines = base64.match(/.{1,64}/g) ?? [];
  return `-----BEGIN CERTIFICATE-----\n${lines.join("\n")}\n-----END CERTIFICATE-----\n`;
}

// derToPem wraps a DER certificate as PEM, which is what the server takes
export function derToPem(der: Uint8Array) {
  let binary = "";
  for (const b of der) binary += String.fromCharCode(b);
  return toPem(btoa(binary));
}

// readCertificateFile reads a PEM file of one or more certificates, or a single DER certificate, as PEM
export function readCertificateFile(bytes: Uint8Array): string[] {
  const text = new TextDecoder().decode(bytes);
  const pems = splitPem(text);
  if (pems.length > 0) return pems;
  // a DER certificate is an ASN.1 sequence, which starts with 0x30
  return bytes.length > 0 && bytes[0] === 0x30 ? [derToPem(bytes)] : [];
}

// fingerprint is a PEM certificate's SHA-256, the same as openssl prints. Undefined where the browser has no WebCrypto, which it only offers over https
export async function fingerprint(pem: string): Promise<string | undefined> {
  if (!globalThis.crypto?.subtle) return undefined;
  const body = pem.replace(/-----(BEGIN|END) CERTIFICATE-----/g, "").replace(/\s+/g, "");
  const der = Uint8Array.from(atob(body), (c) => c.charCodeAt(0));
  const sum = await crypto.subtle.digest("SHA-256", der);
  return [...new Uint8Array(sum)].map((b) => b.toString(16).padStart(2, "0")).join("");
}

// normalizeHex drops the colons and spaces a hash is often printed with
export function normalizeHex(value: string) {
  return value.replace(/[\s:-]/g, "").toLowerCase();
}

// groupFingerprint shows a fingerprint in fours, which is easier to compare by eye than one run of hex
export function groupFingerprint(fp: string) {
  return fp.match(/.{1,4}/g)?.join(" ") ?? fp;
}
