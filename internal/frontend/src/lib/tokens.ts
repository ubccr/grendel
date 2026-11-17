import type { GetRolesResponse } from "@/client/types.gen";

// NEVER is the expiry the API takes for a token that does not expire
export const NEVER = "infinite";

// the expiries offered before anything is typed, as the durations the API takes, which Go's time.ParseDuration reads
const presets: Record<string, string> = {
  "1h": "1 hour",
  "8h": "8 hours",
  "24h": "1 day",
  "720h": "30 days",
  "2160h": "90 days",
  "8760h": "1 year",
  [NEVER]: "Never",
};
export const expiryPresets = Object.keys(presets);

export const DEFAULT_EXPIRY = "720h";

const second = 1;
const minute = 60 * second;
const hour = 60 * minute;
const day = 24 * hour;

// the units an expiry may be written in, in seconds. The API only takes hours, minutes and seconds, so years, weeks and days are turned into hours
const units: [RegExp, number][] = [
  [/^y(ears?|rs?)?$/, 365 * day],
  [/^w(eeks?)?$/, 7 * day],
  [/^d(ays?)?$/, day],
  [/^h(ours?|rs?)?$/, hour],
  [/^m(in(ute)?s?)?$/, minute],
  [/^s(ec(ond)?s?)?$/, second],
];

// expirySeconds reads a length of time as seconds: what someone typed, such as 45d, 2 weeks or 1h30m, or a duration the API takes, which is the same grammar with fewer units. Anything else gives undefined
export function expirySeconds(text: string) {
  const input = text.trim().toLowerCase();
  if (!/^(\s*\d+\s*[a-z]+)+$/.test(input)) return undefined;
  let seconds = 0;
  for (const [, n, word] of input.matchAll(/(\d+)\s*([a-z]+)/g)) {
    const unit = units.find(([re]) => re.test(word));
    if (!unit) return undefined;
    seconds += Number(n) * unit[1];
  }
  return seconds;
}

// parseExpiry reads what someone typed as the expiry the API takes: a duration in hours, minutes and seconds, or never and infinite as no expiry. Anything else, or a duration of zero, gives undefined
export function parseExpiry(text: string) {
  const input = text.trim().toLowerCase();
  if (input === "never" || input === NEVER) return NEVER;
  const seconds = expirySeconds(input);
  if (!seconds) return undefined;
  const h = Math.floor(seconds / hour);
  const m = Math.floor((seconds % hour) / minute);
  const s = seconds % minute;
  return `${h ? `${h}h` : ""}${m ? `${m}m` : ""}${s ? `${s}s` : ""}`;
}

const plural = (n: number, unit: string) => `${n} ${unit}${n === 1 ? "" : "s"}`;

// expiryLabel names an expiry for people: a preset by its name, any other duration in days, hours, minutes and seconds, as 45 days or 1 hour 30 minutes
export function expiryLabel(value: string) {
  if (presets[value]) return presets[value];
  let rest = expirySeconds(value);
  if (rest === undefined) return value;
  const parts: string[] = [];
  for (const [size, unit] of [
    [day, "day"],
    [hour, "hour"],
    [minute, "minute"],
    [second, "second"],
  ] as const) {
    const n = Math.floor(rest / size);
    if (n) parts.push(plural(n, unit));
    rest %= size;
  }
  return parts.join(" ");
}

// expiryOptions is what the expiry list shows for what has been typed, given the expiry chosen. While the input shows the chosen expiry's name, nothing has been typed, so every preset is shown, along with the chosen expiry if it is not one. Otherwise the typed duration comes first when it reads as one, then the presets whose name matches
export function expiryOptions(query: string, chosen: string) {
  if (query.trim() === "" || query === expiryLabel(chosen)) {
    return expiryPresets.includes(chosen) ? expiryPresets : [chosen, ...expiryPresets];
  }
  const typed = parseExpiry(query);
  const q = query.trim().toLowerCase();
  const matches = expiryPresets.filter(
    (p) => p !== typed && expiryLabel(p).toLowerCase().includes(q),
  );
  return typed ? [typed, ...matches] : matches;
}

type Role = NonNullable<GetRolesResponse["roles"]>[number];

const key = (p: { method?: string; path?: string }) => `${p.method} ${p.path}`;

// tokenRoles is the roles a token for a user with userRole can carry: those holding no permission the user's own role lacks, which is what the server checks when it makes the token. The user's role comes first and the rest keep the server's order
export function tokenRoles(roles: Role[], userRole: string) {
  const own = roles.find((r) => r.name === userRole);
  if (!own) return [userRole];
  const allowed = new Set(own.permission_list?.map(key));
  const fits = roles
    .filter((r) => r.name && r.name !== userRole)
    .filter((r) => (r.permission_list ?? []).every((p) => allowed.has(key(p))))
    .map((r) => r.name!);
  return [userRole, ...fits];
}

export type TokenClaims = { username?: string; role?: string; exp?: number };

// tokenParts splits a token into its header, claims and signature, the three parts a JWT is made of
export function tokenParts(token: string) {
  const [header = "", claims = "", signature = ""] = token.split(".");
  return { header, claims, signature };
}

// tokenClaims reads who a token is for, its role and when it expires, as the server wrote them; a token it can't read gives no claims
export function tokenClaims(token: string): TokenClaims {
  try {
    const base64 = tokenParts(token).claims.replace(/-/g, "+").replace(/_/g, "/");
    const bytes = Uint8Array.from(atob(base64.padEnd(Math.ceil(base64.length / 4) * 4, "=")), (c) =>
      c.charCodeAt(0),
    );
    // the claims are UTF-8, which atob alone would garble for a name outside ASCII
    return JSON.parse(new TextDecoder().decode(bytes));
  } catch {
    return {};
  }
}
