import { cn } from "@/lib/utils";
import { Link, type LinkProps } from "@tanstack/react-router";
import type { ReactNode } from "react";

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

export function Unset({ children }: { children: ReactNode }) {
  return <span className="text-muted-foreground">{children}</span>;
}

// a tab never links to the root, and leaving it out lets the paths be checked against from
type Tab = { to: Exclude<LinkProps["to"], "/">; label: string };

// from is the route the tabs' paths are relative to, which supplies their params, such as a node's name
export function DetailTabs({
  label,
  tabs,
  from,
}: {
  label: string;
  tabs: readonly Tab[];
  from?: LinkProps["from"];
}) {
  return (
    <nav aria-label={label} className="-mb-px flex gap-1 overflow-x-auto">
      {tabs.map((t) => (
        <Link
          key={t.to}
          to={t.to}
          from={from}
          activeProps={{ className: "border-foreground text-foreground", "aria-current": "page" }}
          inactiveProps={{
            className:
              "border-transparent text-muted-foreground hover:text-foreground hover:border-border",
          }}
          className="rounded-t-sm border-b-2 px-3 pt-1 pb-2.5 text-sm font-medium whitespace-nowrap transition-colors focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
        >
          {t.label}
        </Link>
      ))}
    </nav>
  );
}
