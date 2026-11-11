import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { cn } from "@/lib/utils";
import { ArrowRight, CircleCheck, CircleX, PencilLine } from "lucide-react";
import { ReactNode } from "react";

export type PreviewRow = {
  node: string;
  serial?: string;
  current: string;
  next: string;
  changed: boolean;
  success: boolean;
  message: string;
};

// counts sums up a preview, so the apply button can say how many nodes it touches
export function counts(rows: PreviewRow[]) {
  return {
    change: rows.filter((r) => r.success && r.changed).length,
    same: rows.filter((r) => r.success && !r.changed).length,
    failed: rows.filter((r) => !r.success).length,
  };
}

// PreviewTable shows each node's value before and after, with the nodes that change first, so a run over a rack reads as what it is about to do
export default function PreviewTable({
  rows,
  applied,
  showSerial,
}: {
  rows: PreviewRow[];
  applied: boolean;
  showSerial?: boolean;
}) {
  const order = (r: PreviewRow) => (!r.success ? 0 : r.changed ? 1 : 2);
  const sorted = [...rows].sort((a, b) => order(a) - order(b));
  const c = counts(rows);

  return (
    <div className="grid gap-3">
      <div className="flex flex-wrap gap-2">
        <Chip tone="info" icon={<PencilLine />}>
          {c.change} {applied ? "changed" : "will change"}
        </Chip>
        <Chip icon={<CircleCheck />}>{c.same} already set</Chip>
        {c.failed > 0 && (
          <Chip tone="warn" icon={<CircleX />}>
            {c.failed} failed
          </Chip>
        )}
      </div>
      <div className="overflow-hidden rounded-lg border">
        <Table>
          <TableHeader className="bg-muted/50">
            <TableRow>
              <TableHead>Node</TableHead>
              {showSerial && <TableHead>Serial</TableHead>}
              <TableHead>Current</TableHead>
              <TableHead className="w-6" />
              <TableHead>{applied ? "Written" : "New"}</TableHead>
              <TableHead>Status</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {sorted.map((r) => (
              <TableRow key={r.node}>
                <TableCell className="font-mono text-xs">{r.node}</TableCell>
                {showSerial && (
                  <TableCell className="font-mono text-xs">{r.serial || "-"}</TableCell>
                )}
                {r.success ? (
                  <>
                    <TableCell
                      className={cn(
                        "font-mono text-xs text-muted-foreground",
                        r.changed && "line-through decoration-muted-foreground/50",
                      )}
                    >
                      {r.current || <span className="italic no-underline">empty</span>}
                    </TableCell>
                    <TableCell>
                      {r.changed && <ArrowRight className="size-3.5 text-muted-foreground" />}
                    </TableCell>
                    <TableCell className="font-mono text-xs">
                      <span
                        className={cn(
                          r.changed && "rounded-md bg-primary/10 px-2 py-0.5 font-medium",
                        )}
                      >
                        {r.next || <span className="text-muted-foreground italic">empty</span>}
                      </span>
                    </TableCell>
                    <TableCell>
                      <Badge variant={r.changed ? "default" : "secondary"}>
                        {r.changed ? (applied ? "changed" : "will change") : "already set"}
                      </Badge>
                    </TableCell>
                  </>
                ) : (
                  <TableCell colSpan={4} className="text-sm text-destructive">
                    {r.message}
                  </TableCell>
                )}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}

function Chip({
  icon,
  tone,
  children,
}: {
  icon: ReactNode;
  tone?: "warn" | "info";
  children: ReactNode;
}) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full border px-3 py-1 text-xs [&>svg]:size-3.5 [&>svg]:shrink-0",
        !tone && "bg-muted/50 [&>svg]:text-muted-foreground",
        tone === "warn" && "border-amber-500/40 bg-amber-500/10 text-amber-700 dark:text-amber-300",
        tone === "info" && "border-sky-500/40 bg-sky-500/10 text-sky-700 dark:text-sky-300",
      )}
    >
      {icon}
      {children}
    </span>
  );
}
