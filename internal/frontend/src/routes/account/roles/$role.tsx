import { getV1Roles } from "@/client";
import { getV1UsersOptions, patchV1RolesMutation } from "@/client/@tanstack/react-query.gen";
import ActionsSheet from "@/components/actions-sheet";
import RolesDeleteAction from "@/components/actions/roles/delete";
import { DetailHeader, Fact, Unset } from "@/components/detail-header";
import { ErrorPanel } from "@/components/error-panel";
import { SaveBar } from "@/components/form-layout";
import {
  type Entry,
  type Group,
  groupPermissions,
  type Kind,
  matches,
  type Permission,
  permissionKey,
} from "@/components/roles/permissions";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { useUser } from "@/hooks/user-provider";
import AuthRedirect from "@/lib/auth";
import { cn } from "@/lib/utils";
import { useMutation, useQuery } from "@tanstack/react-query";
import { createFileRoute, Link, useRouter } from "@tanstack/react-router";
import { Lock, Search, ShieldOff, TriangleAlert, UserRound } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

export const Route = createFileRoute("/account/roles/$role")({
  component: RouteComponent,
  beforeLoad: AuthRedirect,
  // the server answers an unknown role with an error, and the client throws on one, so it is caught here and shown in the page
  loader: ({ params: { role } }) =>
    getV1Roles({ query: { name: role } }).then(
      (res) => ({ roles: res.data?.roles, error: undefined }),
      (e: { detail?: string }) => ({ roles: undefined, error: e?.detail }),
    ),
});

// builtIn are the roles the server refuses to change, so their permissions show read only
const builtIn = new Set(["admin", "user", "read-only"]);

// access are the permissions this page needs: the loader lists the role and saving edits it
const access = ["GET /v1/roles", "PATCH /v1/roles"];

// kindStyle colors a permission by what it does, the same in its method, its tick and the header counts: sky to view, amber to change, red to delete
const kindStyle: Record<Kind, { text: string; tick: string; label: string }> = {
  view: { text: "text-sky-700 dark:text-sky-400", tick: "bg-sky-500", label: "View" },
  change: { text: "text-amber-700 dark:text-amber-400", tick: "bg-amber-500", label: "Change" },
  delete: { text: "text-red-700 dark:text-red-400", tick: "bg-red-500", label: "Delete" },
};

type Show = "all" | "granted" | "missing";

function RouteComponent() {
  const { roles, error } = Route.useLoaderData();
  const name = Route.useParams().role;
  const role = roles?.find((r) => r.name === name);

  if (!role)
    return (
      <Card>
        <CardContent className="p-6">
          <ErrorPanel
            title="Role not found"
            message={error || `No role is named ${name}.`}
            icon={ShieldOff}
          />
        </CardContent>
      </Card>
    );

  const saved = role.permission_list ?? [];
  // after a save the loader reloads, so saved catches up with granted and the save bar goes away
  return (
    <RoleEditor
      name={name}
      saved={saved}
      all={[...saved, ...(role.unassigned_permission_list ?? [])]}
    />
  );
}

function RoleEditor({
  name,
  saved,
  all,
}: {
  name: string;
  saved: Permission[];
  all: Permission[];
}) {
  const router = useRouter();
  const { user } = useUser();
  const { mutate, isPending } = useMutation(patchV1RolesMutation());
  const groups = groupPermissions(all);
  const entries = groups.flatMap((g) => g.entries);
  const initial = new Set(saved.map(permissionKey));
  const [granted, setGranted] = useState(initial);
  const [filter, setFilter] = useState("");
  const [show, setShow] = useState<Show>("all");
  const locked = builtIn.has(name);
  const mine = user?.role === name;

  const changed = entries.filter((e) => granted.has(e.key) !== initial.has(e.key)).length;
  const set = (keys: string[], on: boolean) => {
    const next = new Set(granted);
    for (const k of keys) {
      if (on) next.add(k);
      else next.delete(k);
    }
    setGranted(next);
  };

  // the show toggle goes by the saved permissions, so a row stays put while it is ticked or unticked
  const visible = (e: Entry) =>
    matches(e, filter) && (show === "all" || (show === "granted") === initial.has(e.key));
  const shown = groups
    .map((g) => ({ ...g, entries: g.entries.filter(visible) }))
    .filter((g) => g.entries.length > 0);

  // a person editing their own role can take away their own access to this page, which needs both listing and editing roles
  const lost = mine
    ? entries.filter((e) => access.includes(e.key) && initial.has(e.key) && !granted.has(e.key))
    : [];

  const save = () =>
    mutate(
      {
        body: {
          role: name,
          permission_list: entries
            .filter((e) => granted.has(e.key))
            .map((e) => ({ method: e.method, path: e.path })),
        },
      },
      {
        onSuccess: (res) => {
          toast.success("Saved permissions", { description: res?.detail });
          router.invalidate();
        },
        onError: (e) => toast.error(e.title, { description: e.detail }),
      },
    );

  return (
    <Card>
      <CardContent className="p-0">
        <DetailHeader
          title={name}
          facts={<RoleFacts name={name} entries={entries} granted={granted} />}
          actions={
            !locked && (
              <ActionsSheet checked={name} length={1}>
                <RolesDeleteAction roles={name} />
              </ActionsSheet>
            )
          }
        />
        <form
          className="mx-auto grid w-full max-w-6xl gap-6 px-6 py-5"
          onSubmit={(e) => {
            e.preventDefault();
            save();
          }}
        >
          {locked ? (
            <Notice icon={Lock}>
              {name} is a built-in role, so its permissions can't be changed. To choose your own,
              add a role.
              <Button
                size="sm"
                variant="secondary"
                className="ml-auto"
                render={<Link to="/add/role" />}
                nativeButton={false}
              >
                Add role
              </Button>
            </Notice>
          ) : lost.length > 0 ? (
            <Notice icon={TriangleAlert} tone="warn">
              You have this role. Saving without {lost.map((e) => `"${e.label}"`).join(" and ")}{" "}
              locks you out of this page.
            </Notice>
          ) : (
            mine && (
              <Notice icon={UserRound}>
                You have this role, so your own access changes when you save.
              </Notice>
            )
          )}

          <div className="flex flex-wrap items-center gap-3">
            <div className="relative w-full sm:w-72">
              <Search className="pointer-events-none absolute top-1/2 left-2.5 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                type="search"
                aria-label="Filter permissions"
                placeholder="Filter by name or path"
                className="pl-8"
                value={filter}
                onChange={(e) => setFilter(e.target.value)}
              />
            </div>
            <ToggleGroup
              variant="outline"
              size="sm"
              value={[show]}
              onValueChange={(v) => v[0] && setShow(v[0] as Show)}
              aria-label="Show permissions"
            >
              <ToggleGroupItem value="all">All</ToggleGroupItem>
              <ToggleGroupItem value="granted">Granted</ToggleGroupItem>
              <ToggleGroupItem value="missing">Not granted</ToggleGroupItem>
            </ToggleGroup>
            <Legend />
          </div>

          {shown.length === 0 ? (
            <p className="rounded-md border border-dashed px-4 py-8 text-center text-sm text-muted-foreground">
              {filter.trim()
                ? `No permissions match "${filter.trim()}".`
                : show === "granted"
                  ? "This role has no permissions yet. Show all to grant some."
                  : "This role has every permission."}
            </p>
          ) : (
            <div className="columns-1 gap-x-12 lg:columns-2">
              {shown.map((g) => (
                <PermissionGroup
                  key={g.name}
                  group={g}
                  full={groups.find((x) => x.name === g.name)!}
                  granted={granted}
                  initial={initial}
                  locked={locked}
                  onSet={set}
                />
              ))}
            </div>
          )}

          {changed > 0 && (
            <SaveBar
              noun="role"
              adding={false}
              pending={isPending}
              onDiscard={() => setGranted(initial)}
            />
          )}
        </form>
      </CardContent>
    </Card>
  );
}

// RoleFacts count what the role may do by kind, and who has it
function RoleFacts({
  name,
  entries,
  granted,
}: {
  name: string;
  entries: Entry[];
  granted: Set<string>;
}) {
  const { data: users } = useQuery({ ...getV1UsersOptions(), staleTime: 60_000, retry: false });
  const holders = users?.filter((u) => u.role === name).length;
  const count = (kind: Kind) => {
    const of = entries.filter((e) => e.kind === kind);
    return [of.filter((e) => granted.has(e.key)).length, of.length] as const;
  };

  return (
    <>
      <Fact label="Permissions">
        {entries.filter((e) => granted.has(e.key)).length} of {entries.length}
      </Fact>
      {(["view", "change", "delete"] as const).map((kind) => {
        const [n, of] = count(kind);
        return (
          <Fact key={kind} label={kindStyle[kind].label}>
            <span className={cn(n > 0 && kindStyle[kind].text)}>{n}</span> of {of}
          </Fact>
        );
      })}
      {holders !== undefined && (
        <Fact label="Users">
          {holders ? (
            <Link to="/account/users" className="underline-offset-4 hover:underline">
              {holders}
            </Link>
          ) : (
            <Unset>none</Unset>
          )}
        </Fact>
      )}
    </>
  );
}

// Ticks draw one mark per permission in a group, filled in its kind's color when granted, so a group's access reads at a glance
function Ticks({ entries, granted }: { entries: Entry[]; granted: Set<string> }) {
  return (
    <span aria-hidden className="flex items-center gap-[3px]">
      {entries.map((e) => (
        <span
          key={e.key}
          className={cn(
            "h-3 w-1.5 rounded-[1px] transition-colors",
            granted.has(e.key) ? kindStyle[e.kind].tick : "bg-muted-foreground/20",
          )}
        />
      ))}
    </span>
  );
}

function Legend() {
  return (
    <span className="ml-auto hidden items-center gap-3 text-xs text-muted-foreground md:flex">
      {(["view", "change", "delete"] as const).map((kind) => (
        <span key={kind} className="flex items-center gap-1.5">
          <span className={cn("h-3 w-1.5 rounded-[1px]", kindStyle[kind].tick)} />
          {kindStyle[kind].label}
        </span>
      ))}
    </span>
  );
}

function PermissionGroup({
  group,
  full,
  granted,
  initial,
  locked,
  onSet,
}: {
  group: Group;
  full: Group;
  granted: Set<string>;
  initial: Set<string>;
  locked: boolean;
  onSet: (keys: string[], on: boolean) => void;
}) {
  const keys = group.entries.map((e) => e.key);
  const on = keys.filter((k) => granted.has(k)).length;
  const total = full.entries.filter((e) => granted.has(e.key)).length;

  return (
    <section className="mb-8 break-inside-avoid" aria-label={group.name}>
      <div className="flex items-center gap-3 border-b px-2 pb-2">
        <Checkbox
          aria-label={`Grant every permission in ${group.name}`}
          disabled={locked}
          checked={on === keys.length}
          indeterminate={on > 0 && on < keys.length}
          onCheckedChange={() => onSet(keys, on !== keys.length)}
        />
        <h2 className="text-sm font-semibold">{group.name}</h2>
        <span className="ml-auto flex items-center gap-2.5">
          <Ticks entries={full.entries} granted={granted} />
          <span className="w-12 text-right text-xs text-muted-foreground tabular-nums">
            {total} of {full.entries.length}
          </span>
        </span>
      </div>
      <ul className="grid pt-1">
        {group.entries.map((e) => {
          const has = granted.has(e.key);
          const was = initial.has(e.key);
          return (
            <li key={e.key}>
              <label
                className={cn(
                  "flex items-start gap-3 rounded-md px-2 py-1.5",
                  !locked && "cursor-pointer hover:bg-muted/50",
                )}
              >
                <Checkbox
                  className="mt-0.5"
                  disabled={locked}
                  checked={has}
                  onCheckedChange={(v) => onSet([e.key], v === true)}
                />
                <span className="grid min-w-0 flex-1 gap-0.5">
                  <span className={cn("text-sm", !has && "text-muted-foreground")}>{e.label}</span>
                  <span className="truncate font-mono text-xs text-muted-foreground">
                    <span className={cn("inline-block w-14 font-medium", kindStyle[e.kind].text)}>
                      {e.method}
                    </span>
                    {e.path.replaceAll("%", "*")}
                  </span>
                </span>
                {has !== was && (
                  <span
                    className={cn(
                      "mt-0.5 shrink-0 text-xs",
                      has ? "text-emerald-700 dark:text-emerald-400" : "text-muted-foreground",
                    )}
                  >
                    {has ? "Added" : "Removed"}
                  </span>
                )}
              </label>
            </li>
          );
        })}
      </ul>
    </section>
  );
}

function Notice({
  icon: Icon,
  tone,
  children,
}: {
  icon: typeof Lock;
  tone?: "warn";
  children: React.ReactNode;
}) {
  return (
    <div
      role={tone === "warn" ? "alert" : undefined}
      className={cn(
        "flex flex-wrap items-center gap-x-3 gap-y-2 rounded-md border px-4 py-3 text-sm",
        tone === "warn"
          ? "border-amber-500/40 bg-amber-500/10"
          : "bg-muted/40 text-muted-foreground",
      )}
    >
      <Icon className={cn("size-4 shrink-0", tone === "warn" && "text-amber-600")} />
      {children}
    </div>
  );
}
