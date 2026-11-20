import { cn } from "@/lib/utils";
import type { ReactNode } from "react";

// StepBar floats a step's buttons at the bottom of the window, styled like the node form's SaveBar. The card around it must not clip overflow, or it can't stick
export function StepBar({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <div
      className={cn(
        "sticky bottom-3 z-20 flex flex-wrap items-center justify-end gap-3 rounded-lg border bg-background/95 px-4 py-2.5 shadow-lg backdrop-blur supports-[backdrop-filter]:bg-background/80",
        className,
      )}
    >
      {children}
    </div>
  );
}
