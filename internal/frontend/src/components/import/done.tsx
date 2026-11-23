import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { Link } from "@tanstack/react-router";
import { CircleCheck, Cpu, IdCard } from "lucide-react";
import type { ReactNode } from "react";

export const nextStepLink =
  "flex h-full gap-3 rounded-lg border p-3 outline-none hover:bg-muted/50 focus-visible:ring-2 focus-visible:ring-ring [&>svg]:mt-0.5 [&>svg]:size-4 [&>svg]:shrink-0 [&>svg]:text-muted-foreground";

export function NextStep({ icon, title, text }: { icon: ReactNode; title: string; text: string }) {
  return (
    <>
      {icon}
      <span className="grid gap-0.5">
        <span className="text-sm font-medium">{title}</span>
        <span className="text-xs text-muted-foreground">{text}</span>
      </span>
    </>
  );
}

export function Done({
  title,
  summary,
  nodes,
  first,
  again,
  onAgain,
}: {
  title: string;
  // summary is what the import did, as the parts of one sentence; the empty ones are left out
  summary: (string | false | 0 | undefined)[];
  // nodes is a nodeset
  nodes: string;
  // first is a link offered ahead of the ones every import ends with, as a NextStep inside a link styled with nextStepLink
  first?: ReactNode;
  again: string;
  onAgain: () => void;
}) {
  const next = [
    {
      to: "/nodes/bios/$nodes" as const,
      icon: <Cpu />,
      title: "Configure BIOS",
      text: "Turn on PXE for the boot port and set the boot order.",
    },
    {
      to: "/nodes/identity/$nodes" as const,
      icon: <IdCard />,
      title: "Set locations and asset tags",
      text: "Write each node's rack location to its BMC.",
    },
  ];

  return (
    <div className="grid gap-6">
      <div className="flex items-start gap-3">
        <CircleCheck className="mt-0.5 size-6 shrink-0 text-(--node-ready)" />
        <div className="grid gap-1">
          <h2 className="text-lg font-semibold tracking-tight">{title}</h2>
          <p className="text-sm text-muted-foreground">{summary.filter(Boolean).join(", ")}.</p>
          <p className="font-mono text-sm break-all">{nodes}</p>
        </div>
      </div>

      <div className="grid gap-2">
        <h3 className="text-sm font-medium">Next, for these nodes</h3>
        <ul className={cn("grid gap-2", first ? "sm:grid-cols-3" : "sm:grid-cols-2")}>
          {first && <li>{first}</li>}
          {next.map((n) => (
            <li key={n.to}>
              <Link to={n.to} params={{ nodes }} className={nextStepLink}>
                <NextStep icon={n.icon} title={n.title} text={n.text} />
              </Link>
            </li>
          ))}
        </ul>
      </div>

      <div>
        <Button variant="outline" onClick={onAgain}>
          {again}
        </Button>
      </div>
    </div>
  );
}
