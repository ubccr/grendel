import { CodeBlockCopyButton } from "@/components/ui/code-block";
import { X } from "lucide-react";
import type { ReactNode } from "react";

// Problem is why one thing a node is sent can't be: what failed, and the error as the server gave it
export function Problem({
  title,
  error,
  children,
}: {
  title: string;
  error: string;
  children?: ReactNode;
}) {
  return (
    <section className="overflow-hidden rounded-lg border border-destructive/30">
      <header className="flex items-start gap-3 border-b border-destructive/20 bg-destructive/5 px-4 py-3">
        <span className="mt-0.5 flex size-5 shrink-0 items-center justify-center rounded-full bg-destructive text-white">
          <X className="size-3" strokeWidth={3} />
        </span>
        <h3 className="text-sm font-medium">{title}</h3>
      </header>

      <div className="grid gap-4 px-4 py-4">
        <div className="grid gap-1.5">
          <div className="flex items-center justify-between gap-2">
            <span className="text-xs text-muted-foreground">Error from the server</span>
            <CodeBlockCopyButton value={error} />
          </div>
          <pre className="font-mono text-xs leading-relaxed break-words whitespace-pre-wrap">
            {error}
          </pre>
        </div>

        {children}
      </div>
    </section>
  );
}
