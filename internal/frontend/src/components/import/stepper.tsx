import { cn } from "@/lib/utils";
import { Check } from "lucide-react";

export type Step = { id: string; label: string };

export function Stepper({
  steps,
  current,
  reachable,
  onGo,
}: {
  steps: Step[];
  current: string;
  reachable: (id: string) => boolean;
  onGo: (id: string) => void;
}) {
  const at = steps.findIndex((s) => s.id === current);
  return (
    <ol className="flex flex-wrap items-stretch gap-x-1 gap-y-2" aria-label="Import steps">
      {steps.map((step, i) => {
        const done = i < at;
        const active = i === at;
        const canGo = !active && reachable(step.id);
        return (
          <li key={step.id} className="flex items-center gap-1">
            {i > 0 && (
              <span
                aria-hidden
                className={cn("h-px w-6 bg-border", done || active ? "bg-foreground/40" : "")}
              />
            )}
            <button
              type="button"
              disabled={!canGo}
              aria-current={active ? "step" : undefined}
              onClick={() => onGo(step.id)}
              className={cn(
                "flex items-center gap-2.5 rounded-md px-2 py-1.5 text-left outline-none focus-visible:ring-2 focus-visible:ring-ring",
                canGo && "hover:bg-muted",
                !canGo && !active && "cursor-default",
              )}
            >
              <span
                className={cn(
                  "grid size-6 shrink-0 place-items-center rounded-full border text-xs font-medium tabular-nums",
                  active && "border-foreground bg-foreground text-background",
                  done && "border-foreground/40 text-foreground",
                  !active && !done && "text-muted-foreground",
                )}
              >
                {done ? <Check className="size-3.5" /> : i + 1}
              </span>
              <span className={cn("text-sm", active ? "font-medium" : "text-muted-foreground")}>
                {step.label}
              </span>
            </button>
          </li>
        );
      })}
    </ol>
  );
}
