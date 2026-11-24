import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { entryName } from "@/lib/settings";
import { cn } from "@/lib/utils";
import { Lock, Plus, X } from "lucide-react";
import { useState } from "react";

type EditorProps<T> = {
  id: string;
  value: T;
  disabled?: boolean;
  onChange: (value: T) => void;
};

export function ListEditor({
  id,
  value,
  disabled,
  onChange,
  numeric,
}: EditorProps<(string | number)[]> & { numeric?: boolean }) {
  const items = value;
  const [entry, setEntry] = useState("");

  const add = () => {
    const parts = entry
      .split(/[\s,]+/)
      .filter(Boolean)
      .map((p) => (numeric ? Number(p) : p))
      .filter((p) => !numeric || Number.isInteger(p));
    if (parts.length > 0) onChange([...items, ...parts]);
    setEntry("");
  };

  return (
    <div className="grid gap-2">
      {items.length > 0 && (
        <ul className="flex flex-wrap gap-1.5">
          {items.map((item, i) => (
            <li
              key={`${item}-${i}`}
              className="flex items-center gap-1 rounded-md border bg-muted/50 py-0.5 pr-1 pl-2 font-mono text-sm"
            >
              {item}
              {!disabled && (
                <button
                  type="button"
                  className="rounded-sm p-0.5 text-muted-foreground hover:bg-muted hover:text-foreground"
                  aria-label={`Remove ${item}`}
                  onClick={() => onChange(items.filter((_, j) => j !== i))}
                >
                  <X className="size-3" />
                </button>
              )}
            </li>
          ))}
        </ul>
      )}
      {!disabled && (
        <div className="flex gap-1">
          <Input
            id={id}
            className="font-mono"
            inputMode={numeric ? "numeric" : undefined}
            placeholder={items.length === 0 ? "None. Type one and press Enter" : "Add another"}
            value={entry}
            onChange={(e) => setEntry(e.target.value)}
            onBlur={add}
            onKeyDown={(e) => {
              if (e.key === "Enter" || e.key === ",") {
                e.preventDefault();
                add();
              }
            }}
          />
        </div>
      )}
      {disabled && items.length === 0 && (
        <span className="text-sm text-muted-foreground">None</span>
      )}
    </div>
  );
}

// AddEntry names a new entry of a setting made of named entries, such as a subnet
export function AddEntry({
  id,
  label,
  taken,
  onAdd,
}: {
  id: string;
  label: string;
  // taken are the names already in use
  taken: string[];
  onAdd: (name: string) => void;
}) {
  const [name, setName] = useState("");
  const wanted = name.trim().toLowerCase();
  const valid = entryName.test(wanted) && !taken.includes(wanted);
  const add = () => {
    if (!valid) return;
    onAdd(wanted);
    setName("");
  };

  return (
    <div className="flex gap-2">
      <Input
        id={id}
        className="font-mono"
        placeholder="name, such as 20-compute"
        aria-label={`Name of the new ${label.toLowerCase()}`}
        value={name}
        onChange={(e) => setName(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === "Enter") {
            e.preventDefault();
            add();
          }
        }}
      />
      <Button variant="outline" disabled={!valid} onClick={add}>
        <Plus /> Add
      </Button>
    </div>
  );
}

// a value typed for a free-form entry is kept as JSON when it is a list, a number or true/false, and as text otherwise
function parseEntry(text: string): unknown {
  const t = text.trim();
  if (/^(\[|\{|-?\d|true$|false$)/.test(t)) {
    try {
      return JSON.parse(t);
    } catch {
      return text;
    }
  }
  return text;
}

function showEntry(v: unknown): string {
  return typeof v === "string" ? v : JSON.stringify(v);
}

export function MapEditor({
  id,
  value,
  disabled,
  onChange,
  fileKeys,
  fileName,
}: EditorProps<Record<string, unknown>> & { fileKeys: string[]; fileName: string }) {
  const entries = Object.entries(value);
  const [name, setName] = useState("");
  const update = (next: [string, unknown][]) => onChange(Object.fromEntries(next));

  const add = () => {
    const key = name.trim().toLowerCase();
    if (!key || entries.some(([k]) => k === key)) return;
    update([...entries, [key, ""]]);
    setName("");
  };

  return (
    <div className="grid gap-2">
      {fileKeys.map((key) => (
        <div key={key} className="grid grid-cols-[10rem_1fr] items-center gap-2">
          <span className="truncate font-mono text-sm">{key}</span>
          <span className="flex h-9 items-center gap-2 rounded-md border border-dashed px-3 text-sm text-muted-foreground">
            <Lock className="size-3.5" /> set in {fileName}, never shown
          </span>
        </div>
      ))}
      {entries.map(([key, v], i) => (
        <div key={key} className="grid grid-cols-[10rem_1fr_auto] items-center gap-2">
          <span className="truncate font-mono text-sm" title={key}>
            {key}
          </span>
          <Input
            className={cn("font-mono", fileKeys.includes(key) && "line-through opacity-60")}
            disabled={disabled}
            title={
              fileKeys.includes(key) ? `${fileName} sets ${key}, which wins over this` : undefined
            }
            value={showEntry(v)}
            onChange={(e) =>
              update(
                entries.map((entry, j) => (j === i ? [key, parseEntry(e.target.value)] : entry)),
              )
            }
          />
          {!disabled && (
            <Button
              variant="ghost"
              size="icon"
              aria-label={`Remove ${key}`}
              onClick={() => update(entries.filter((_, j) => j !== i))}
            >
              <X />
            </Button>
          )}
        </div>
      ))}
      {entries.length === 0 && fileKeys.length === 0 && (
        <p className="text-sm text-muted-foreground">No entries.</p>
      )}
      {!disabled && (
        <div className="flex gap-2">
          <Input
            id={id}
            className="max-w-40 font-mono"
            placeholder="name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") {
                e.preventDefault();
                add();
              }
            }}
          />
          <Button variant="outline" disabled={!name.trim()} onClick={add}>
            <Plus /> Add
          </Button>
        </div>
      )}
    </div>
  );
}
