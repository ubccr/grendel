// Permission is one API route a role may call. The path matches like SQL LIKE, so % stands for a path part such as a node name
export type Kind = "view" | "change" | "delete";

export type Permission = {
  method?: string;
  path?: string;
  summary?: string;
  group?: string;
  kind?: Kind;
};

export type Entry = {
  key: string;
  method: string;
  path: string;
  label: string;
  kind: Kind;
};

export type Group = { name: string; entries: Entry[] };

export function permissionKey(p: Permission) {
  return `${p.method ?? ""} ${p.path ?? ""}`;
}

// kindOf is the kind of a permission the server did not describe
export function kindOf(method: string): Kind {
  if (method === "GET") return "view";
  if (method === "DELETE") return "delete";
  return "change";
}

// A permission the server does not describe shows under Other, by its method and path
export function groupPermissions(all: Permission[]): Group[] {
  const groups = new Map<string, Map<string, Entry>>();
  for (const { method, path, summary, group, kind } of all) {
    if (!method || !path) continue;
    const key = permissionKey({ method, path });
    const name = group || "Other";
    const entries = groups.get(name) ?? new Map<string, Entry>();
    entries.set(key, { key, method, path, label: summary || key, kind: kind ?? kindOf(method) });
    groups.set(name, entries);
  }
  return [...groups]
    .map(([name, entries]) => ({
      name,
      entries: [...entries.values()].sort(
        (a, b) => a.path.localeCompare(b.path) || a.method.localeCompare(b.method),
      ),
    }))
    .sort(
      (a, b) =>
        Number(a.name === "Other") - Number(b.name === "Other") || a.name.localeCompare(b.name),
    );
}

export function matches(e: Entry, filter: string) {
  const f = filter.trim().toLowerCase();
  return !f || `${e.label} ${e.method} ${e.path}`.toLowerCase().includes(f);
}
