import type { ReactNode } from "react";

// AccountPanel lays out one tab of the account page: a titled form on the left and, beside it on a wide screen or under it on a narrow one, what goes with it, a SideNote or a result
export function AccountPanel({
  title,
  description,
  children,
  aside,
}: {
  title: string;
  description: ReactNode;
  children: ReactNode;
  aside: ReactNode;
}) {
  return (
    <div className="grid grid-cols-1 gap-x-12 gap-y-10 lg:grid-cols-[minmax(0,26rem)_minmax(0,1fr)]">
      <section className="grid content-start gap-5">
        <div className="grid gap-1">
          <h2 className="text-base font-semibold tracking-tight">{title}</h2>
          <p className="text-sm text-pretty text-muted-foreground">{description}</p>
        </div>
        {children}
      </section>
      {aside}
    </div>
  );
}

// SideNote is what to know before using the form beside it, set apart by a rule
export function SideNote({ title, children }: { title: string; children: ReactNode }) {
  return (
    <aside className="grid content-start gap-4 border-l pl-6 text-sm text-pretty text-muted-foreground max-lg:border-t max-lg:border-l-0 max-lg:pt-6 max-lg:pl-0">
      <h2 className="text-base font-semibold tracking-tight text-foreground">{title}</h2>
      {children}
    </aside>
  );
}
