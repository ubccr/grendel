import { writable, type Writable } from "@/lib/secureboot";
import { cn } from "@/lib/utils";
import { ChevronDown, ChevronRight } from "lucide-react";

const roles: Record<Writable, { name: string; tone: string }> = {
  PK: { name: "Platform key", tone: "" },
  KEK: { name: "Key exchange keys", tone: "" },
  db: { name: "Allowed to boot", tone: "text-green-700 dark:text-green-400" },
  dbx: { name: "Revoked", tone: "text-red-700 dark:text-red-400" },
};

type Props = {
  // counts is how many keys each database holds
  counts: Record<string, number>;
  // unread are the databases with no count to show, still being read from the BMC or failed to be
  unread?: Partial<Record<Writable, "reading" | "failed">>;
  selected: Writable;
  onSelect: (db: Writable) => void;
  // unit is what a count counts, keys on one node or distinct keys across many
  unit?: string;
};

// The databases are drawn in the order authority flows: PK signs changes to KEK, KEK signs changes to db, and dbx sits beside db as the list that overrides it
export default function TrustChain({
  counts,
  unread = {},
  selected,
  onSelect,
  unit = "key",
}: Props) {
  const link = (db: Writable) => {
    const n = counts[db] ?? 0;
    const state = unread[db];
    const empty = !state && n === 0;
    const active = selected === db;
    return (
      <button
        key={db}
        type="button"
        aria-pressed={active}
        onClick={() => onSelect(db)}
        className={cn(
          "group flex min-w-0 flex-1 flex-col items-start rounded-lg border px-4 py-3 text-left transition-colors focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none",
          empty && "border-dashed",
          active ? "border-foreground bg-accent" : "hover:border-foreground/40 hover:bg-accent/50",
        )}
      >
        <span className={cn("text-2xl leading-none font-semibold tracking-tight", roles[db].tone)}>
          {db}
        </span>
        <span className="mt-1.5 text-sm">{roles[db].name}</span>
        <span className={cn("text-xs", empty ? "text-yellow-600" : "text-muted-foreground")}>
          {state === "reading"
            ? "Reading…"
            : state === "failed"
              ? "Could not be read"
              : empty
                ? db === "PK"
                  ? "None, so nothing is enforced"
                  : "Empty"
                : `${n} ${unit}${n === 1 ? "" : "s"}`}
        </span>
      </button>
    );
  };

  const arrow = (label: string) => (
    <div
      aria-hidden
      className="flex shrink-0 items-center justify-center gap-1 text-[11px] text-muted-foreground sm:flex-col sm:gap-0 sm:px-1"
    >
      <ChevronRight className="hidden size-4 sm:block" />
      <ChevronDown className="size-4 sm:hidden" />
      <span className="sm:-mt-0.5">{label}</span>
    </div>
  );

  const [pk, kek, db, dbx] = writable;
  return (
    <nav
      aria-label="Secure boot databases"
      className="flex flex-col gap-2 sm:flex-row sm:items-stretch"
    >
      {link(pk)}
      {arrow("signs")}
      {link(kek)}
      {arrow("signs")}
      <div className="flex min-w-0 flex-[2] gap-2">
        {link(db)}
        <div aria-hidden className="w-px self-stretch bg-border" />
        {link(dbx)}
      </div>
    </nav>
  );
}
