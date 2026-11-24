import { ListEditor, MapEditor } from "@/components/settings/editors";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { type Setting, show } from "@/lib/settings";
import { cn } from "@/lib/utils";

// Control is the input for one setting, picked by its kind. Text inputs hand back what was typed, which the server reads the way it reads a config file.
export function Control({
  id,
  option: o,
  file,
  value,
  disabled,
  invalid,
  onChange,
}: {
  id: string;
  option: Setting;
  file: string;
  value: unknown;
  disabled: boolean;
  invalid: boolean;
  onChange: (value: unknown) => void;
}) {
  const fileKeys = o.file_keys ?? [];
  const text = show(value);
  const on = value === true || value === "true";

  switch (o.kind) {
    case "bool":
      return (
        <div className="flex h-9 items-center gap-2">
          <Switch
            id={id}
            checked={on}
            disabled={disabled}
            onCheckedChange={(checked) => onChange(checked)}
          />
          <span className="text-sm text-muted-foreground">{on ? "On" : "Off"}</span>
        </div>
      );
    case "enum": {
      const choices = (o.default ?? "") === "" ? ["", ...(o.enum ?? [])] : (o.enum ?? []);
      return (
        <Select
          value={text}
          disabled={disabled}
          onValueChange={(v) => onChange(v ?? "")}
          items={choices.map((c) => ({ value: c, label: c || "None" }))}
        >
          <SelectTrigger id={id} className="w-full" aria-invalid={invalid}>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {choices.map((c) => (
              <SelectItem key={c} value={c}>
                {c || "None"}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      );
    }
    case "string_list":
    case "int_list":
      return (
        <ListEditor
          id={id}
          value={Array.isArray(value) ? value : []}
          disabled={disabled}
          onChange={onChange}
          numeric={o.kind === "int_list"}
        />
      );
    case "group":
      return (
        <p className="flex min-h-9 items-center font-mono text-sm">
          {Array.isArray(value) && value.length > 0 ? (
            value.join(", ")
          ) : (
            <span className="font-sans text-muted-foreground">None yet</span>
          )}
        </p>
      );
    case "map":
      return (
        <MapEditor
          id={id}
          value={value && typeof value === "object" ? (value as Record<string, unknown>) : {}}
          disabled={disabled}
          onChange={onChange}
          fileKeys={fileKeys}
          fileName={file}
        />
      );
  }

  const numeric = o.kind === "int" || o.kind === "float";
  return (
    <div className="relative">
      <Input
        id={id}
        type={numeric ? "number" : "text"}
        step={o.kind === "float" ? "any" : undefined}
        min={o.min ?? undefined}
        max={o.max ?? undefined}
        className={cn("font-mono", o.unit && "pr-20", invalid && "border-destructive")}
        aria-invalid={invalid}
        disabled={disabled}
        title={o.kind === "path" ? text : undefined}
        placeholder={o.kind === "duration" ? "30s, 15m, 24h" : undefined}
        value={text}
        onChange={(e) => onChange(e.target.value)}
      />
      {o.unit && (
        <span className="pointer-events-none absolute top-2 right-3 text-sm text-muted-foreground">
          {o.unit}
        </span>
      )}
    </div>
  );
}
