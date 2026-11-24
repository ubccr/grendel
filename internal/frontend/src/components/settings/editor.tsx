import { patchV1SettingsMutation } from "@/client/@tanstack/react-query.gen";
import type { Listing } from "@/client/types.gen";
import { ReviewDialog } from "@/components/settings/review";
import { SettingRow, SourceMark } from "@/components/settings/row";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { errorToast } from "@/lib/query-client";
import {
  type Draft,
  editable,
  entryRemovals,
  entryRows,
  fileName,
  groups,
  isSet,
  layers,
  matches,
  refusals,
  type Setting,
  type SettingsSection,
  toPatch,
  withEdit,
  withoutEntry,
} from "@/lib/settings";
import { cn } from "@/lib/utils";
import { useMutation } from "@tanstack/react-query";
import { LoaderCircle, RotateCw, Search, TriangleAlert } from "lucide-react";
import { type ReactNode, useState } from "react";
import { toast } from "sonner";

// canEdit is false for an account that may read the settings but not change them
export default function SettingsEditor({
  listing,
  canEdit,
}: {
  listing: Listing;
  canEdit: boolean;
}) {
  const [search, setSearch] = useState("");
  const [section, setSection] = useState<string>();
  const [changedOnly, setChangedOnly] = useState(false);
  const [pendingOnly, setPendingOnly] = useState(false);
  const [draft, setDraft] = useState<Draft>({});
  const [refused, setRefused] = useState<Record<string, string>>({});
  // added are the entries named here but not saved yet, by the setting they belong to
  const [added, setAdded] = useState<Record<string, string[]>>({});

  // a refusal is shown on its row, so the toast is kept for failures that name no setting
  const save = useMutation({ ...patchV1SettingsMutation(), meta: { quietError: true } });

  // an entry that is new gets a row per field after the ones the server already lists
  const sections = (listing.sections ?? []).map((s) => {
    const options = s.options ?? [];
    const rows = options.flatMap((o) =>
      o.kind === "group" ? (added[o.key ?? ""] ?? []).flatMap((name) => entryRows(o, name)) : [],
    );
    return { ...s, options: [...options, ...rows] };
  });
  const all = sections.flatMap((s) => s.options);
  const query = search.trim();

  const keep = (o: Setting) => (!changedOnly || isSet(o)) && (!pendingOnly || !!o.pending);
  const current = sections.find((s) => s.id === section) ?? sections[0];
  const shown = (query ? sections : [current])
    .map((s) => ({ ...s, options: s.options.filter((o) => keep(o) && matches(o, query)) }))
    .filter((s) => s.options.length > 0);

  const changes = [...all.filter((o) => (o.key ?? "") in draft), ...entryRemovals(all, draft)];
  const pending = all.filter((o) => o.pending);
  const unsaved = changes.length > 0 || Object.values(added).some((names) => names.length > 0);

  const edit = (o: Setting, value: unknown) => {
    setDraft((prev) => withEdit(prev, o, value));
    setRefused((prev) => {
      const { [o.key ?? ""]: _, ...rest } = prev;
      return rest;
    });
  };
  const addEntry = (group: string, name: string) =>
    setAdded((prev) => ({ ...prev, [group]: [...(prev[group] ?? []), name] }));
  // an entry that was never saved is simply dropped, and a saved one is marked for removal until the save
  const removeEntry = (group: string, name: string) => {
    const saved = !(added[group] ?? []).includes(name);
    setAdded((prev) => ({ ...prev, [group]: (prev[group] ?? []).filter((n) => n !== name) }));
    setDraft((prev) => withoutEntry(prev, `${group}.${name}`, saved));
  };
  const keepEntry = (group: string, name: string) =>
    setDraft((prev) => {
      const { [`${group}.${name}`]: _, ...rest } = prev;
      return rest;
    });
  const discard = () => {
    setDraft({});
    setAdded({});
    setRefused({});
  };

  const submit = () =>
    save.mutate(
      { body: toPatch(draft) },
      {
        onSuccess: () => {
          toast.success(`Saved ${changes.length} setting(s)`);
          discard();
        },
        onError: (error) => {
          const reasons = refusals(error);
          setRefused(reasons);
          if (Object.keys(reasons).length === 0) errorToast(error);
          else
            toast.error("Nothing was saved", {
              description: "Fix the marked settings and save again.",
            });
        },
      },
    );

  return (
    <div className="grid">
      {(!listing.persistent || pending.length > 0) && (
        <div className="grid gap-2 border-b px-6 py-4">
          {!listing.persistent && (
            <Notice>
              grendeld is running on an in-memory database, so settings saved here are lost when it
              stops. Start it with a database path to keep them.
            </Notice>
          )}
          {pending.length > 0 && (
            <Notice icon={<RotateCw className="mt-0.5 size-4 shrink-0" />}>
              {pending.length} saved setting(s) take effect when grendeld restarts:{" "}
              <span className="font-mono">{pending.map((o) => o.key).join(", ")}</span>
            </Notice>
          )}
        </div>
      )}

      <div className="flex flex-wrap items-center gap-x-6 gap-y-3 border-b px-6 py-3">
        <div className="relative min-w-56 flex-1">
          <Search className="pointer-events-none absolute top-1/2 left-2.5 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            type="search"
            aria-label="Search settings"
            className="pl-8"
            placeholder="Search every section"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <Filter
          id="settings-changed"
          label="Changed from default"
          checked={changedOnly}
          onChange={setChangedOnly}
        />
        <Filter
          id="settings-pending"
          label="Waiting for restart"
          checked={pendingOnly}
          onChange={setPendingOnly}
        />
      </div>

      <div className="grid md:grid-cols-[13rem_minmax(0,1fr)]">
        <SectionNav
          sections={sections}
          current={query ? undefined : current?.id}
          draft={draft}
          configFile={listing.config_file}
          onSelect={(id) => {
            setSection(id);
            setSearch("");
          }}
        />

        <div className="min-w-0 md:border-l">
          {shown.length === 0 && (
            <p className="p-10 text-center text-sm text-muted-foreground">No settings match.</p>
          )}
          {shown.map((s) => (
            <section
              key={s.id}
              aria-labelledby={`section-${s.id}`}
              className="border-b last:border-b-0"
            >
              <header className="px-6 pt-5 pb-3">
                <h2 id={`section-${s.id}`} className="text-lg font-semibold tracking-tight">
                  {s.title}
                </h2>
                <p className="max-w-[70ch] text-sm text-muted-foreground">{s.description}</p>
              </header>
              {groups(s.options).map((g) => {
                // a heading over the fields of one entry, such as a subnet, is also where the entry is removed
                const { entry_of: owner = "", entry = "", source } = g.options[0];
                const leaving = owner !== "" && draft[`${owner}.${entry}`] === null;
                return (
                  <div key={g.name}>
                    {g.name && (
                      <GroupHeading
                        name={g.name}
                        leaving={leaving}
                        fromFile={owner !== "" && source === "file"}
                        configFile={listing.config_file}
                        onToggle={
                          owner && canEdit && source !== "file"
                            ? () => (leaving ? keepEntry(owner, entry) : removeEntry(owner, entry))
                            : undefined
                        }
                      />
                    )}
                    <div className={cn("divide-y", leaving && "pointer-events-none opacity-50")}>
                      {g.options.map((o) => (
                        <SettingRow
                          key={o.key}
                          option={o}
                          configFile={listing.config_file}
                          canEdit={canEdit && !leaving}
                          edit={draft[o.key ?? ""]}
                          refused={refused[o.key ?? ""]}
                          onChange={(v) => edit(o, v)}
                          onUndo={() => edit(o, editable(o))}
                          taken={all.filter((x) => x.entry_of === o.key).map((x) => x.entry ?? "")}
                          onAddEntry={(name) => addEntry(o.key ?? "", name)}
                        />
                      ))}
                    </div>
                  </div>
                );
              })}
            </section>
          ))}
        </div>
      </div>

      {canEdit && (
        <div className="sticky bottom-0 z-10 flex flex-wrap items-center justify-end gap-3 rounded-b-xl border-t bg-card px-6 py-3">
          <span className="mr-auto text-sm text-muted-foreground">
            {changes.length === 0 ? "No unsaved changes" : `${changes.length} unsaved change(s)`}
          </span>
          <Button variant="ghost" disabled={!unsaved} onClick={discard}>
            Discard
          </Button>
          <ReviewDialog
            trigger={
              <Button disabled={changes.length === 0 || save.isPending}>
                {save.isPending && <LoaderCircle className="animate-spin" />}
                Review and save
              </Button>
            }
            changes={changes}
            draft={draft}
            onConfirm={submit}
          />
        </div>
      )}
    </div>
  );
}

function SectionNav({
  sections,
  current,
  draft,
  configFile,
  onSelect,
}: {
  sections: SettingsSection[];
  // current is left out while a search shows every section
  current?: string;
  draft: Draft;
  configFile?: string;
  onSelect: (id?: string) => void;
}) {
  return (
    <nav
      aria-label="Sections"
      className="content-start border-b p-3 md:sticky md:top-0 md:self-start md:border-b-0"
    >
      <ul className={cn("grid gap-0.5", !current && "opacity-50")}>
        {sections.map((s) => {
          const edited = (s.options ?? []).filter((o) => (o.key ?? "") in draft).length;
          const active = s.id === current;
          return (
            <li key={s.id}>
              <button
                type="button"
                aria-current={active ? "true" : undefined}
                onClick={() => onSelect(s.id)}
                className={cn(
                  "flex w-full items-center justify-between rounded-md px-3 py-1.5 text-left text-sm text-muted-foreground hover:bg-muted hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none",
                  active && "bg-muted font-medium text-foreground",
                )}
              >
                <span className="truncate">{s.title}</span>
                {edited > 0 && (
                  <span className="rounded-full bg-primary px-1.5 text-xs text-primary-foreground tabular-nums">
                    {edited}
                  </span>
                )}
              </button>
            </li>
          );
        })}
      </ul>
      <Precedence configFile={configFile} />
    </nav>
  );
}

function GroupHeading({
  name,
  leaving,
  fromFile,
  configFile,
  onToggle,
}: {
  name: string;
  // leaving means the entry under the heading is removed by the next save
  leaving: boolean;
  fromFile: boolean;
  configFile?: string;
  // onToggle removes the entry, or keeps one marked for removal, and is left out when it cannot be removed here
  onToggle?: () => void;
}) {
  return (
    <h3 className="flex min-h-9 items-center gap-3 border-y bg-muted/40 px-6 py-1 text-sm font-medium">
      <span className={cn(leaving && "line-through opacity-60")}>{name}</span>
      {fromFile && <SourceMark source="file" label={fileName(configFile)} />}
      {leaving && (
        <span className="text-xs font-normal text-muted-foreground">Removed when you save</span>
      )}
      {onToggle && (
        <Button variant="ghost" size="sm" className="ml-auto h-7" onClick={onToggle}>
          {leaving ? "Keep" : "Remove"}
        </Button>
      )}
    </h3>
  );
}

// Precedence is the key to the source markers, listed strongest first
function Precedence({ configFile }: { configFile?: string }) {
  return (
    <div className="mt-4 grid gap-2 border-t px-3 pt-4 text-xs text-muted-foreground">
      <p className="font-medium text-foreground">Legend</p>
      <p>Listed in order of precedence</p>
      <ol className="grid justify-items-start gap-1.5">
        {layers.map((layer) => (
          <li key={layer}>
            <SourceMark
              source={layer}
              label={
                layer === "file"
                  ? fileName(configFile)
                  : layer === "flag"
                    ? "command line"
                    : layer === "env"
                      ? "GRENDEL_*"
                      : undefined
              }
            />
          </li>
        ))}
      </ol>
    </div>
  );
}

function Filter({
  id,
  label,
  checked,
  onChange,
}: {
  id: string;
  label: string;
  checked: boolean;
  onChange: (v: boolean) => void;
}) {
  return (
    <div className="flex items-center gap-2">
      <Switch id={id} checked={checked} onCheckedChange={onChange} />
      <Label htmlFor={id}>{label}</Label>
    </div>
  );
}

function Notice({ icon, children }: { icon?: ReactNode; children: ReactNode }) {
  return (
    <div className="flex items-start gap-2 rounded-md border border-amber-500/40 bg-amber-500/10 px-3 py-2 text-sm text-amber-700 dark:text-amber-300">
      {icon ?? <TriangleAlert className="mt-0.5 size-4 shrink-0" />}
      <span>{children}</span>
    </div>
  );
}
