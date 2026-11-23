import { RadioGroupItem } from "@/components/ui/radio-group";
import { lastOctet } from "@/lib/addresses";
import { cn } from "@/lib/utils";
import { CircleCheck } from "lucide-react";

export type Choice = "lowest" | "after";

const quietLabel = "No node has it, and nothing answered a ping";

// skipped are the free addresses before the suggested one that answered a ping, named by last octet unless the range is wide. A quiet address needs no words, since its icon says nothing answered
export function Option({
  value,
  label,
  ip,
  prefix,
  skipped,
  pinged,
  pingError,
  wide,
  empty = "Nothing is free.",
}: {
  value: Choice;
  label: string;
  ip?: string;
  prefix?: number;
  skipped: string[];
  pinged?: boolean;
  pingError?: string;
  wide: boolean;
  empty?: string;
}) {
  let note: string | undefined;
  if (!ip && skipped.length)
    note = `The next ${skipped.length} free addresses all answered a ping, though no node has them.`;
  else if (!ip) note = empty;
  else if (!pinged) note = `Not pinged${pingError ? `: ${pingError}` : ""}.`;
  else if (skipped.length)
    note = `Skipped ${skipped.map((a) => (wide ? a : lastOctet(a))).join(", ")}, which answered a ping but no node has.`;

  return (
    <label
      className={cn(
        "flex items-center gap-3 rounded-md border px-3 py-2.5 transition-colors has-data-checked:border-foreground/40 has-data-checked:bg-muted/40",
        ip ? "cursor-pointer hover:bg-muted/40" : "opacity-60",
      )}
    >
      <RadioGroupItem value={value} disabled={!ip} />
      <span className="grid min-w-0 flex-1 gap-0.5">
        <span className="text-xs text-muted-foreground">{label}</span>
        <span className="font-mono text-base leading-tight">
          {ip ?? "None"}
          {ip && prefix !== undefined && <span className="text-muted-foreground">/{prefix}</span>}
        </span>
        {note && (
          <span
            className={cn("text-xs", skipped.length ? "text-foreground" : "text-muted-foreground")}
          >
            {note}
          </span>
        )}
      </span>
      {ip && pinged && (
        <span title={quietLabel} className="shrink-0">
          <CircleCheck aria-hidden className="size-5 text-(--node-ready)" />
          <span className="sr-only">{quietLabel}</span>
        </span>
      )}
    </label>
  );
}
