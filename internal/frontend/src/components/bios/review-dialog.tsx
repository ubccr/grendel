import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { BiosAttribute, BiosValue, showValue } from "@/lib/bios";
import { parseNodeSet } from "@/lib/nodeset";
import { cn } from "@/lib/utils";
import { ArrowRight, Clock, Server, TriangleAlert } from "lucide-react";
import { ReactNode } from "react";

export type BiosChange = { name: string; attr?: BiosAttribute; from: BiosValue; to: BiosValue };

type Props = {
  trigger: ReactNode;
  changes: BiosChange[];
  nodes: string;
  applyTime: string;
  applyTimeLabel?: string;
  onConfirm: () => void;
};

// ReviewDialog lists what is about to be staged, and where and when it applies, before anything is sent
export default function ReviewDialog({
  trigger,
  changes,
  nodes,
  applyTime,
  applyTimeLabel,
  onConfirm,
}: Props) {
  const nodeset = parseNodeSet(nodes);
  const count = nodeset.len();
  const folded = nodeset.toString();
  const when = applyTime
    ? (applyTimeLabel ?? applyTime).toLowerCase()
    : "at the BMC's default time";
  const unique = changes.filter((c) => c.attr?.system_unique);
  const reboot = changes.some((c) => c.attr?.reset_required);

  return (
    <Dialog>
      <DialogTrigger asChild>{trigger}</DialogTrigger>
      <DialogContent className="flex max-h-[90dvh] flex-col gap-0 overflow-hidden p-0 sm:max-w-2xl">
        <DialogHeader className="gap-3 p-6 pb-4">
          <DialogTitle>Review BIOS changes</DialogTitle>
          <DialogDescription>
            {changes.length} change(s) will be staged on {count} node(s), and applied {when}.
          </DialogDescription>
          <div className="flex flex-wrap gap-2 pt-1">
            <Chip icon={<Clock />}>{applyTimeLabel ?? (applyTime || "BMC default")}</Chip>
            <Chip icon={<Server />} title={folded}>
              <span className="max-w-72 truncate font-mono">{folded}</span>
            </Chip>
          </div>
        </DialogHeader>

        <ul className="min-h-0 flex-1 divide-y overflow-y-auto border-y">
          {changes.map((c) => (
            <li
              key={c.name}
              className="grid gap-2 px-6 py-3 sm:grid-cols-[1fr_auto] sm:items-center"
            >
              <div className="min-w-0">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="truncate font-medium">{c.attr?.display_name || c.name}</span>
                  {c.attr?.reset_required && <Badge variant="secondary">needs reboot</Badge>}
                  {c.attr?.system_unique && <Badge variant="secondary">per system</Badge>}
                </div>
                {c.attr?.display_name && (
                  <div className="truncate font-mono text-xs text-muted-foreground">{c.name}</div>
                )}
              </div>
              <div className="flex items-center gap-2 font-mono text-sm">
                <span className="text-muted-foreground line-through decoration-muted-foreground/50">
                  {display(c.attr, c.from)}
                </span>
                <ArrowRight className="size-3.5 shrink-0 text-muted-foreground" />
                <span className="rounded-md bg-primary/10 px-2 py-0.5 font-medium text-foreground">
                  {display(c.attr, c.to)}
                </span>
              </div>
            </li>
          ))}
        </ul>

        {(applyTime === "Immediate" || unique.length > 0 || (reboot && !applyTime)) && (
          <div className="grid gap-2 px-6 pt-4">
            {applyTime === "Immediate" && (
              <Notice>The BMC may reboot the hosts right away to apply these changes.</Notice>
            )}
            {reboot && !applyTime && (
              <Notice>
                Some of these changes need a reboot, and the BMC's default apply time decides when
                that happens.
              </Notice>
            )}
            {count > 1 && unique.length > 0 && (
              <Notice>
                {unique.map((c) => c.attr?.display_name || c.name).join(", ")} is meant to differ
                per system, but every node will get the same value.
              </Notice>
            )}
          </div>
        )}

        <DialogFooter className="gap-2 p-6 pt-4">
          <DialogClose asChild>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
          <DialogClose asChild>
            <Button onClick={onConfirm}>Stage {changes.length} change(s)</Button>
          </DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

// display shows an enumeration value by its registry display name
function display(attr: BiosAttribute | undefined, value: BiosValue) {
  const match = attr?.values?.find((v) => v.name === value);
  return match?.display_name || showValue(value);
}

function Chip({ icon, title, children }: { icon: ReactNode; title?: string; children: ReactNode }) {
  return (
    <span
      title={title}
      className={cn(
        "inline-flex min-w-0 items-center gap-1.5 rounded-full border bg-muted/50 px-2.5 py-1 text-xs",
        "[&>svg]:size-3.5 [&>svg]:shrink-0 [&>svg]:text-muted-foreground",
      )}
    >
      {icon}
      {children}
    </span>
  );
}

function Notice({ children }: { children: ReactNode }) {
  return (
    <div className="flex items-start gap-2 rounded-md border border-amber-500/40 bg-amber-500/10 px-3 py-2 text-sm text-amber-700 dark:text-amber-300">
      <TriangleAlert className="mt-0.5 size-4 shrink-0" />
      <span>{children}</span>
    </div>
  );
}
