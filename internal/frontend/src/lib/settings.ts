import type { HttpError, Listing } from "@/client/types.gen";

export type SettingsSection = NonNullable<Listing["sections"]>[number];
export type Setting = NonNullable<SettingsSection["options"]>[number];
export type SettingField = NonNullable<Setting["fields"]>[number];

// a draft holds the unsaved edits by key: null removes the setting from the database, and anything else is its new value
export type Draft = Record<string, unknown>;

// layers are the places a value can come from, strongest first
export const layers = ["flag", "env", "file", "database", "default"] as const;

// a merged setting takes entries from both the config file and the database, so what is edited is the database's share
export function isMerged(o: Setting): boolean {
  return o.kind === "map";
}

function isList(o: Setting): boolean {
  return o.kind === "string_list" || o.kind === "int_list";
}

// isLocked means a flag, the environment or the config file sets the value, which the database cannot override
export function isLocked(o: Setting): boolean {
  if (o.kind === "map" || o.kind === "group") return false;
  return o.source === "flag" || o.source === "env" || o.source === "file";
}

export function isSet(o: Setting): boolean {
  return o.source !== "default" && o.source !== "inherited";
}

export function isWritable(o: Setting): boolean {
  return !o.secret && !o.external;
}

export function hasStored(o: Setting): boolean {
  return o.stored !== undefined && o.stored !== null;
}

// editable is the value an edit starts from: the database's own entries for a merged setting, and the value in use for the rest
export function editable(o: Setting): unknown {
  if (isMerged(o)) return o.stored ?? {};
  return o.value ?? (isList(o) ? [] : "");
}

function sortKeys(v: unknown): unknown {
  if (Array.isArray(v)) return v.map(sortKeys);
  if (v && typeof v === "object")
    return Object.fromEntries(
      Object.entries(v as Record<string, unknown>)
        .sort(([a], [b]) => a.localeCompare(b))
        .map(([k, x]) => [k, sortKeys(x)]),
    );
  return v;
}

// show writes a value the way it is typed: text and numbers as they are, and a list or a map as JSON
export function show(v: unknown): string {
  if (v === undefined || v === null) return "";
  return typeof v === "object" ? JSON.stringify(sortKeys(v)) : String(v);
}

// afterSave is the value a setting has once the draft is saved. Removing a setting a stronger layer overrides changes nothing that runs
export function afterSave(o: Setting, draft: Draft): unknown {
  const edit = draft[o.key ?? ""];
  if (edit === undefined) return editable(o);
  if (edit === null) {
    if (isMerged(o)) return {};
    return isLocked(o) ? o.value : o.default;
  }
  return edit;
}

// a typed 9000 and a stored 9000 are the same value, as are two maps with their keys in another order
export function isChange(o: Setting, edit: unknown): boolean {
  if (edit === null) return hasStored(o);
  return show(edit) !== show(editable(o));
}

// withEdit returns the draft with one edit made, dropping the edit when it leaves the setting as it is
export function withEdit(draft: Draft, o: Setting, edit: unknown): Draft {
  const next = { ...draft };
  if (isChange(o, edit)) next[o.key ?? ""] = edit;
  else delete next[o.key ?? ""];
  return next;
}

// withoutEntry drops the edits to one entry's fields. A saved entry is also marked for removal, which an unsaved one does not need.
export function withoutEntry(draft: Draft, entry: string, saved: boolean): Draft {
  const next = Object.fromEntries(
    Object.entries(draft).filter(([key]) => !key.startsWith(`${entry}.`)),
  );
  if (saved) next[entry] = null;
  return next;
}

// entryRemovals are the whole entries a draft removes, as settings of their own so each gets a line in the review
export function entryRemovals(all: Setting[], draft: Draft): Setting[] {
  return all
    .filter((g) => g.kind === "group")
    .flatMap((g) =>
      [...new Set(all.filter((o) => o.entry_of === g.key).map((o) => o.entry ?? ""))]
        .filter((name) => draft[`${g.key}.${name}`] === null)
        .map((name) => ({
          key: `${g.key}.${name}`,
          summary: `${g.entry_label} ${name}`,
          kind: "string",
          apply: g.apply,
          value: "kept",
          default: "removed",
        })),
    );
}

export function toPatch(draft: Draft): { set: Record<string, unknown>; unset: string[] } {
  const set: Record<string, unknown> = {};
  const unset: string[] = [];
  for (const [key, edit] of Object.entries(draft)) {
    if (edit === null) unset.push(key);
    else set[key] = edit;
  }
  return { set, unset };
}

// refusals maps a setting's key to why the server refused it
export function refusals(error: HttpError | null | undefined): Record<string, string> {
  const out: Record<string, string> = {};
  for (const item of error?.errors ?? []) {
    if (item.name) out[item.name] = item.reason ?? "refused";
  }
  return out;
}

export function fileName(path?: string): string {
  return path?.split("/").pop() || "grendel.toml";
}

// overriddenBy names what a locked setting is set by, the way an admin would look for it
export function overriddenBy(o: Setting, configFile?: string): string {
  switch (o.source) {
    case "flag":
      return `--${o.flag}`;
    case "env":
      return o.env ?? "the environment";
    case "file":
      return fileName(configFile);
    default:
      return "";
  }
}

export function matches(o: Setting, query: string): boolean {
  return `${o.key} ${o.summary} ${o.help} ${o.group}`
    .toLowerCase()
    .includes(query.trim().toLowerCase());
}

// summarize shortens a value for a line of a review, since a map does not fit
export function summarize(v: unknown): string {
  if (Array.isArray(v)) return v.length === 0 ? "none" : v.join(", ");
  if (v && typeof v === "object") {
    const keys = Object.keys(v);
    return keys.length === 0 ? "none" : keys.join(", ");
  }
  return show(v) || "empty";
}

export const entryName = /^[a-z0-9][a-z0-9_-]*$/;

// entryRows are the settings of an entry that does not exist yet, one per field, so a new subnet can be filled in like any other
export function entryRows(group: Setting, name: string): Setting[] {
  return (group.fields ?? []).map((f) => ({
    key: `${group.key}.${name}.${f.name}`,
    summary: f.summary,
    help: f.help,
    kind: f.kind,
    group: `${group.entry_label} ${name}`,
    apply: group.apply,
    source: "default",
    entry_of: group.key,
    entry: name,
    enum: [],
    fields: [],
    file_keys: [],
  }));
}

// groups splits a section's settings by their subheading, keeping the order they are listed in
export function groups(options: Setting[]): { name: string; options: Setting[] }[] {
  const out: { name: string; options: Setting[] }[] = [];
  for (const o of options) {
    const name = o.group ?? "";
    const last = out.find((g) => g.name === name);
    if (last) last.options.push(o);
    else out.push({ name, options: [o] });
  }
  return out;
}
