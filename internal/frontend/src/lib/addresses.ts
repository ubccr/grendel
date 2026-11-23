// the largest range the API searches
const MIN_RANGE_BITS = 16;

// rangeKey is where this browser remembers the range last used for each kind of port, since a rack's nodes usually go in the same ranges as the last one's
export const rangeKey = (bmc: boolean) => `grendel.address-range.${bmc ? "bmc" : "port"}`;

export function ipToInt(ip: string): number | undefined {
  const parts = ip.trim().split(".");
  if (parts.length !== 4 || parts.some((p) => !/^\d{1,3}$/.test(p) || +p > 255)) return undefined;
  return parts.reduce((n, p) => n * 256 + +p, 0);
}

export function intToIp(n: number): string {
  return [n >>> 24, (n >>> 16) & 255, (n >>> 8) & 255, n & 255].join(".");
}

function network(n: number, bits: number) {
  return bits === 0 ? 0 : (n & (-1 << (32 - bits))) >>> 0;
}

// parseRange keeps a range written out, like 10.17.25.0/24, to its network, and reads an address without a prefix, or its first three octets, as its /24. Anything else, or a range larger than a /16, gives undefined
export function parseRange(text: string): string | undefined {
  const [addr, bits, ...rest] = text.trim().split("/");
  if (rest.length || !addr) return undefined;
  const octets = addr.split(".");
  const n = ipToInt(octets.length === 3 ? `${addr}.0` : addr);
  if (n === undefined) return undefined;
  if (bits === undefined) return `${intToIp(network(n, 24))}/24`;
  if (!/^\d{1,2}$/.test(bits) || +bits > 32 || +bits < MIN_RANGE_BITS) return undefined;
  return `${intToIp(network(n, +bits))}/${+bits}`;
}

// rangeOf suggests a port's own subnet when that is from a /24 to a /30, and otherwise the /24 it sits in. A /31 or /32 is a point to point link or a single host rather than a subnet to fill, so it too gives its /24
export function rangeOf(ip?: string): string | undefined {
  const [addr, bits] = (ip ?? "").split("/");
  const n = ipToInt(addr ?? "");
  if (n === undefined) return undefined;
  const b = bits && /^\d{1,2}$/.test(bits) && +bits >= 24 && +bits <= 30 ? +bits : 24;
  return `${intToIp(network(n, b))}/${b}`;
}

// stepRange is the range of the same size just after a range, or just before it with dir -1, or undefined past either end of IPv4
export function stepRange(range: string, dir: 1 | -1): string | undefined {
  const [addr, bits] = range.split("/");
  const n = ipToInt(addr ?? "");
  const b = Number(bits);
  if (n === undefined || !Number.isInteger(b) || b < 0 || b > 32) return undefined;
  const size = 2 ** (32 - b);
  const next = network(n, b) + dir * size;
  if (next < 0 || next + size - 1 > 0xffffffff) return undefined;
  return `${intToIp(next)}/${b}`;
}

export function stripPrefix(ip?: string) {
  return (ip ?? "").split("/")[0]?.trim() ?? "";
}

// blockOf is the stretch of at most 256 addresses a map of the range shows: the whole range when it is a /24 or smaller, otherwise the /24 holding focus, or the range's first /24 without one
export function blockOf(range: string, focus?: string): { start: number; size: number } {
  const [addr, bits] = range.split("/");
  const n = ipToInt(addr ?? "") ?? 0;
  const b = +(bits ?? 24);
  if (b >= 24) return { start: network(n, b), size: 2 ** (32 - b) };
  const f = focus ? ipToInt(focus) : undefined;
  const inRange = f !== undefined && network(f, b) === network(n, b);
  return { start: network(inRange ? f : n, 24), size: 256 };
}

export function lastOctet(ip: string) {
  return `.${ip.split(".")[3] ?? ""}`;
}
