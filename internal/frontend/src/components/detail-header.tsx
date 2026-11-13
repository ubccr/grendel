import { cn } from "@/lib/utils";
import type { ReactNode } from "react";

// DetailHeader heads the page for one thing, a node or an image: its name, a row of the facts about it worth seeing at a glance, its actions, and the tabs of its views when it has more than one
export function DetailHeader({
  title,
  facts,
  actions,
  nav,
}: {
  title: string;
  facts?: ReactNode;
  actions?: ReactNode;
  nav?: ReactNode;
}) {
  return (
    <header className={cn("grid gap-5 border-b px-6 pt-5", !nav && "pb-5")}>
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="grid min-w-0 gap-5">
          <h1 className="font-mono text-3xl font-semibold tracking-tight break-all">{title}</h1>
          {facts && <dl className="flex flex-wrap gap-x-4 gap-y-3 pl-4">{facts}</dl>}
        </div>
        {actions && <div className="flex gap-2">{actions}</div>}
      </div>
      {nav}
    </header>
  );
}

// Fact is one labelled fact in a DetailHeader, divided from the one before it by a rule
export function Fact({
  label,
  children,
  className,
}: {
  label: string;
  children: ReactNode;
  className?: string;
}) {
  return (
    <div
      className={cn(
        "grid min-w-0 content-start gap-0.5 sm:border-l sm:pl-4 sm:first:border-l-0 sm:first:pl-0",
        className,
      )}
    >
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd className="truncate text-sm">{children}</dd>
    </div>
  );
}

// Unset is a fact's value when nothing is set, like a node with no image of its own
export function Unset({ children }: { children: ReactNode }) {
  return <span className="text-muted-foreground">{children}</span>;
}
