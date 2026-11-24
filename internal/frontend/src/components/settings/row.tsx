import { Control } from "@/components/settings/control";
import { AddEntry } from "@/components/settings/editors";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import {
  editable,
  fileName,
  hasStored,
  isLocked,
  isMerged,
  isSet,
  isWritable,
  overriddenBy,
  type Setting,
  summarize,
} from "@/lib/settings";
import { cn } from "@/lib/utils";
import { EyeOff, Lock, RotateCw, Undo2 } from "lucide-react";
import type { ReactNode } from "react";

export function SettingRow({
  option: o,
  configFile,
  canEdit,
  edit,
  refused,
  onChange,
  onUndo,
  taken,
  onAddEntry,
}: {
  option: Setting;
  configFile?: string;
  canEdit: boolean;
  edit: unknown;
  refused?: string;
  onChange: (value: unknown) => void;
  onUndo: () => void;
  // taken and onAddEntry serve a setting made of named entries: the names in use, and what adds one
  taken: string[];
  onAddEntry: (name: string) => void;
}) {
  const edited = edit !== undefined;
  const removing = edit === null;
  const file = fileName(configFile);
  const id = `setting-${o.key}`;
  const writable = isWritable(o);
  const locked = isLocked(o);
  const disabled = !canEdit || !writable || locked;
  const cleared = locked ? o.value : isMerged(o) ? {} : o.default;
  const shown = removing ? cleared : (edit ?? editable(o));
  // a value made of entries needs the row's whole width, so it sits under the description rather than beside it
  const stacked = o.kind === "map";

  return (
    <div
      className={cn(
        "relative grid gap-x-8 gap-y-3 px-6 py-4",
        !stacked && "lg:grid-cols-[minmax(0,1fr)_19rem]",
        edited && "bg-primary/5",
      )}
    >
      {edited && <span aria-hidden className="absolute inset-y-0 left-0 w-0.5 bg-primary" />}

      <div className="min-w-0">
        <div className="flex flex-wrap items-baseline gap-x-2.5 gap-y-0.5">
          <Label htmlFor={id} className="text-sm font-medium">
            {o.summary}
          </Label>
          <code className="font-mono text-xs text-muted-foreground">{o.key}</code>
        </div>
        <p className="mt-1 max-w-[70ch] text-sm text-muted-foreground">{o.help}</p>
        {(o.pending || o.deprecated || (o.apply === "restart" && writable)) && (
          <div className="mt-2 flex flex-wrap items-center gap-2 text-xs">
            {o.pending ? (
              <span className="flex items-center gap-1 text-sky-600 dark:text-sky-400">
                <RotateCw className="size-3" />
                Running with {summarize(o.running)} until grendeld restarts
              </span>
            ) : (
              o.apply === "restart" &&
              writable && (
                <span className="flex items-center gap-1 text-muted-foreground">
                  <RotateCw className="size-3" />
                  Applies after a restart
                </span>
              )
            )}
            {o.deprecated && (
              <span className="text-amber-700 dark:text-amber-400">Deprecated, {o.deprecated}</span>
            )}
          </div>
        )}
      </div>

      <div className={cn("grid content-start gap-1.5", stacked && "max-w-3xl")}>
        {o.secret ? (
          <div className="flex h-9 items-center gap-2 rounded-md border border-dashed px-3 text-sm text-muted-foreground">
            <EyeOff className="size-4" />
            {o.source === "database" ? "Generated" : isSet(o) ? "Set" : "Not set"}, never shown
          </div>
        ) : (
          <div className="flex items-start gap-1">
            <div className="min-w-0 flex-1">
              <Control
                id={id}
                option={o}
                file={file}
                value={shown}
                disabled={disabled}
                invalid={!!refused}
                onChange={onChange}
              />
            </div>
            {edited && (
              <Button
                variant="ghost"
                size="icon"
                onClick={onUndo}
                aria-label={`Undo the change to ${o.key}`}
              >
                <Undo2 />
              </Button>
            )}
          </div>
        )}

        {refused && (
          <p role="alert" className="text-sm text-destructive">
            {refused}
          </p>
        )}

        {o.kind === "group"
          ? canEdit && (
              <AddEntry id={id} label={o.entry_label ?? "entry"} taken={taken} onAdd={onAddEntry} />
            )
          : null}
        <Origin
          option={o}
          file={file}
          removing={removing}
          onRemove={
            canEdit && writable && hasStored(o) && !removing ? () => onChange(null) : undefined
          }
        />
      </div>
    </div>
  );
}

// Origin says in one line where a setting's value comes from and what, if anything, can be done about it
function Origin({
  option: o,
  file,
  removing,
  onRemove,
}: {
  option: Setting;
  file: string;
  removing: boolean;
  onRemove?: () => void;
}) {
  const by = overriddenBy(o, file);
  const outside = [file, o.env, o.flag && `--${o.flag}`].filter(Boolean).join(", ");

  // a setting made of entries has no one source: each entry says its own
  if (o.kind === "group") return null;

  const writable = isWritable(o);
  const locked = isLocked(o);
  let note: ReactNode = null;
  if (removing) note = "Saving removes it from the database";
  else if (!writable && !locked)
    note = o.secret
      ? o.source === "database"
        ? `Set your own in ${outside}`
        : `Secrets are set in ${outside}`
      : `Only set in ${outside}`;
  else if (!writable) note = "Not stored in the database";
  else if (locked && o.shadowed)
    note = (
      <>
        wins over <span className="font-mono">{summarize(o.stored)}</span> in the database
      </>
    );
  else if (locked) note = "wins over the database";
  else if (o.source === "inherited")
    note = (
      <>
        follows <span className="font-mono">{o.inherit_from}</span>
      </>
    );
  else if (hasStored(o) && !isMerged(o))
    note = (
      <>
        {o.inherit_from ? "otherwise " : "default "}
        <span className="font-mono">{summarize(o.default)}</span>
      </>
    );
  else if (o.file_entries_secret)
    note = `Anyone who can see settings can read these. Keep secrets in ${file}`;

  return (
    <div className="flex flex-wrap items-center gap-x-2 gap-y-1 text-xs text-muted-foreground">
      {/* most settings are at their default, so only the ones that are not get a marker */}
      {o.source !== "default" && o.source !== "inherited" && (
        <SourceMark source={o.source ?? ""} label={by || undefined} />
      )}
      {(o.file_keys ?? []).length > 0 && o.source !== "file" && (
        <SourceMark source="file" label={file} />
      )}
      {note && <span className="min-w-0">{note}</span>}
      {onRemove && (
        <button
          type="button"
          className="rounded-sm font-medium text-foreground underline underline-offset-2 hover:no-underline focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
          onClick={onRemove}
        >
          {locked
            ? "Remove it"
            : o.inherit_from
              ? "Follow it again"
              : isMerged(o)
                ? "Clear"
                : "Reset"}
        </button>
      )}
    </div>
  );
}

// SourceMark is the small marker for one of the places a value comes from. A lock means the value is set outside the database, a filled dot that the database holds it, and a hollow one that nothing sets it.
export function SourceMark({ source, label }: { source: string; label?: string }) {
  const outside = source === "flag" || source === "env" || source === "file";
  const stored = source === "database" || source === "generated";
  return (
    <span
      className={cn(
        "inline-flex shrink-0 items-center gap-1.5 rounded-sm border px-1.5 py-0.5 leading-none",
        outside && "border-amber-500/40 bg-amber-500/10 text-amber-700 dark:text-amber-300",
        stored && "border-primary/30 bg-primary/10 text-foreground",
      )}
    >
      {outside ? (
        <Lock className="size-3" />
      ) : (
        <span
          aria-hidden
          className={cn("size-1.5 rounded-full border border-current", stored && "bg-current")}
        />
      )}
      <span className={cn(outside && "font-mono")}>
        {label ?? (source === "inherited" ? "default" : source)}
      </span>
    </span>
  );
}
