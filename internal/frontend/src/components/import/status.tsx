import type { BootRow, BootStatus } from "@/lib/boot-nics";
import type { Row, Status } from "@/lib/import-rows";
import { cn } from "@/lib/utils";
import {
  CircleAlert,
  CircleCheck,
  CircleDashed,
  CircleMinus,
  CircleX,
  Database,
  MousePointerClick,
  PencilLine,
  Unplug,
} from "lucide-react";
import type { ReactNode } from "react";

type Look = { label: string; icon: ReactNode; color?: string };

// the faceplate lights and the grid share one color per status, so a lit cage and its row read as the same thing
export const statusLook: Record<Status, Look> = {
  ready: { label: "Ready", icon: <CircleCheck />, color: "var(--node-ready)" },
  existing: { label: "Already in Grendel", icon: <Database />, color: "var(--node-blue)" },
  pick: { label: "Pick a MAC", icon: <MousePointerClick />, color: "var(--node-amber)" },
  conflict: { label: "Conflict", icon: <CircleX />, color: "var(--destructive)" },
  invalid: { label: "Fix the name", icon: <CircleAlert />, color: "var(--destructive)" },
  empty: { label: "No MAC learned", icon: <CircleDashed /> },
  unnamed: { label: "Needs a name", icon: <PencilLine /> },
  badmac: { label: "Fix the MAC", icon: <CircleAlert />, color: "var(--destructive)" },
  excluded: { label: "Left out", icon: <CircleMinus /> },
};

const bootLook: Record<BootStatus, Look> = {
  ready: { label: "Ready", icon: <CircleCheck />, color: "var(--node-ready)" },
  existing: { label: "Already in Grendel", icon: <Database />, color: "var(--node-blue)" },
  pick: { label: "Pick an interface", icon: <MousePointerClick />, color: "var(--node-amber)" },
  conflict: { label: "Conflict", icon: <CircleX />, color: "var(--destructive)" },
  invalid: { label: "Fix the FQDN", icon: <CircleAlert />, color: "var(--destructive)" },
  empty: { label: "No interfaces listed", icon: <CircleDashed /> },
  unreached: { label: "BMC didn't answer", icon: <Unplug />, color: "var(--destructive)" },
  excluded: { label: "Left out", icon: <CircleMinus /> },
};

export function StatusTag({ row, className }: { row: Row; className?: string }) {
  return (
    <Tag look={statusLook[row.status]} title={row.problem ?? row.skipped} className={className} />
  );
}

export function BootStatusTag({ row }: { row: BootRow }) {
  return <Tag look={bootLook[row.status]} title={row.problem} />;
}

function Tag({ look, title, className }: { look: Look; title?: string; className?: string }) {
  return (
    <span
      title={title}
      className={cn(
        "inline-flex items-center gap-1.5 text-xs whitespace-nowrap [&>svg]:size-3.5 [&>svg]:shrink-0",
        !look.color && "text-muted-foreground",
        className,
      )}
      style={
        look.color
          ? { color: `color-mix(in oklab, ${look.color} 85%, var(--foreground))` }
          : undefined
      }
    >
      {look.icon}
      {look.label}
    </span>
  );
}
