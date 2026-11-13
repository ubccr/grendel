// Permission is one API route a role may call. The path matches like SQL LIKE, so % stands for a path part such as a node name
export type Permission = { method?: string; path?: string };

export type Kind = "view" | "change" | "delete";

export type Entry = {
  key: string;
  method: string;
  path: string;
  label: string;
  kind: Kind;
};

export type Group = { name: string; entries: Entry[] };

// catalog names each permission by what it lets someone do, in the order the editor lists them. A permission missing here still shows, under Other, by its method and path
const catalog: [string, [string, string][]][] = [
  [
    "Nodes",
    [
      ["GET /v1/nodes", "List nodes"],
      ["GET /v1/nodes/find", "Find nodes"],
      ["POST /v1/nodes", "Add and edit nodes"],
      ["DELETE /v1/nodes", "Delete nodes"],
      ["GET /v1/nodes/tags", "List tags"],
      ["PATCH /v1/nodes/tags/%", "Add and remove tags"],
      ["PATCH /v1/nodes/image", "Change a node's boot image"],
    ],
  ],
  [
    "Provisioning",
    [
      ["PATCH /v1/nodes/provision", "Turn provisioning on or off"],
      ["GET /v1/nodes/provision-status", "See provision status"],
      ["GET /v1/nodes/token/%", "Create boot tokens"],
    ],
  ],
  [
    "Boot images",
    [
      ["GET /v1/images", "List images"],
      ["GET /v1/images/find", "Find images"],
      ["POST /v1/images", "Add and edit images"],
      ["DELETE /v1/images", "Delete images"],
    ],
  ],
  [
    "Power and console",
    [
      ["GET /v1/bmc/status", "See BMC status"],
      ["POST /v1/bmc/power", "Change power state"],
      ["POST /v1/bmc/indicator", "Turn the locator light on or off"],
      ["GET /v1/bmc/console", "Open the serial console"],
    ],
  ],
  [
    "Event logs and jobs",
    [
      ["GET /v1/bmc/sel", "Read the event log"],
      ["POST /v1/bmc/sel/clear", "Clear the event log"],
      ["GET /v1/bmc/jobs", "See BMC jobs"],
      ["POST /v1/bmc/jobs/clear", "Clear BMC jobs"],
    ],
  ],
  [
    "BIOS and firmware",
    [
      ["GET /v1/bmc/bios", "Read BIOS settings"],
      ["GET /v1/bmc/bios/registry", "Read the BIOS attribute list"],
      ["POST /v1/bmc/bios", "Stage BIOS changes"],
      ["POST /v1/bmc/bios/discard", "Discard staged BIOS changes"],
      ["GET /v1/bmc/firmware/updates", "Check for firmware updates"],
      ["POST /v1/bmc/firmware/install", "Install firmware"],
      ["POST /v1/bmc/configure/auto", "Auto-configure BMCs"],
      ["POST /v1/bmc/configure/import", "Import a BMC configuration"],
    ],
  ],
  [
    "Location and asset tags",
    [
      ["GET /v1/bmc/identity/defaults", "See default locations and asset tags"],
      ["POST /v1/bmc/location", "Set rack locations"],
      ["POST /v1/bmc/asset-tag", "Set asset tags"],
    ],
  ],
  [
    "Switches",
    [
      ["GET /v1/switch/%/lldp", "Read LLDP neighbors"],
      ["GET /v1/switch/%/hosttree", "Read the host tree"],
    ],
  ],
  [
    "Users and roles",
    [
      ["GET /v1/users", "List users"],
      ["POST /v1/users", "Add users"],
      ["PATCH /v1/users/%/enable", "Enable and disable users"],
      ["PATCH /v1/users/%/role", "Change a user's role"],
      ["DELETE /v1/users/%", "Delete users"],
      ["GET /v1/roles", "List roles"],
      ["POST /v1/roles", "Add roles"],
      ["PATCH /v1/roles", "Edit role permissions"],
      ["DELETE /v1/roles/%", "Delete roles"],
    ],
  ],
  [
    "Account and server",
    [
      ["PATCH /v1/auth/reset", "Change their own password"],
      ["POST /v1/auth/token", "Create API tokens"],
      ["GET /v1/grendel/version", "See the server version"],
      ["GET /v1/grendel/events", "Follow live events"],
      ["GET /v1/db/dump", "Download a database backup"],
      ["POST /v1/db/restore", "Restore a database backup"],
    ],
  ],
];

export function permissionKey(p: Permission) {
  return `${p.method ?? ""} ${p.path ?? ""}`;
}

export function kindOf(method: string): Kind {
  if (method === "GET") return "view";
  if (method === "DELETE") return "delete";
  return "change";
}

// groupPermissions sorts every permission the server knows into the catalog's groups and order, dropping groups with none, and puts unknown ones last under Other
export function groupPermissions(all: Permission[]): Group[] {
  const present = new Map<string, Entry>();
  for (const { method, path } of all)
    if (method && path) {
      const key = permissionKey({ method, path });
      present.set(key, { key, method, path, label: key, kind: kindOf(method) });
    }
  const take = ([key, label]: [string, string]) => {
    const e = present.get(key);
    present.delete(key);
    return e ? [{ ...e, label }] : [];
  };
  const groups = catalog.map(([name, items]) => ({ name, entries: items.flatMap(take) }));
  const other = [...present.values()].sort(
    (a, b) => a.path.localeCompare(b.path) || a.method.localeCompare(b.method),
  );
  return [...groups, { name: "Other", entries: other }].filter((g) => g.entries.length > 0);
}

// matches says whether an entry fits a filter typed by the person, by its label, method or path
export function matches(e: Entry, filter: string) {
  const f = filter.trim().toLowerCase();
  return !f || `${e.label} ${e.method} ${e.path}`.toLowerCase().includes(f);
}
