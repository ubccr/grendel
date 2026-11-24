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
import { afterSave, type Draft, editable, isLocked, type Setting, summarize } from "@/lib/settings";
import { ArrowRight } from "lucide-react";
import type { ReactElement } from "react";

export function ReviewDialog({
  trigger,
  changes,
  draft,
  onConfirm,
}: {
  trigger: ReactElement;
  changes: Setting[];
  draft: Draft;
  onConfirm: () => void;
}) {
  const restart = changes.filter((o) => o.apply === "restart" && !isLocked(o));

  return (
    <Dialog>
      <DialogTrigger render={trigger} />
      <DialogContent className="flex max-h-[90dvh] flex-col gap-0 overflow-hidden p-0 sm:max-w-2xl">
        <DialogHeader className="gap-3 p-6 pb-4">
          <DialogTitle>Review settings</DialogTitle>
          <DialogDescription>
            {changes.length} change(s) will be saved to the database.
            {restart.length > 0 && ` ${restart.length} of them wait for a restart of grendeld.`}
          </DialogDescription>
        </DialogHeader>

        <ul className="min-h-0 flex-1 divide-y overflow-y-auto border-y">
          {changes.map((o) => (
            <li
              key={o.key}
              className="grid gap-2 px-6 py-3 sm:grid-cols-[1fr_auto] sm:items-center"
            >
              <div className="min-w-0">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="truncate font-medium">{o.summary}</span>
                  {isLocked(o) ? (
                    <Badge variant="secondary">database only</Badge>
                  ) : o.apply === "restart" ? (
                    <Badge variant="secondary">after restart</Badge>
                  ) : (
                    <Badge variant="outline">applies now</Badge>
                  )}
                </div>
                <div className="truncate font-mono text-xs text-muted-foreground">{o.key}</div>
              </div>
              <div className="flex items-center gap-2 font-mono text-sm">
                <span className="text-muted-foreground line-through decoration-muted-foreground/50">
                  {summarize(isLocked(o) ? o.stored : editable(o))}
                </span>
                <ArrowRight className="size-3.5 shrink-0 text-muted-foreground" />
                <span className="rounded-md bg-primary/10 px-2 py-0.5 font-medium text-foreground">
                  {draft[o.key ?? ""] === null
                    ? isLocked(o)
                      ? "removed"
                      : `${summarize(afterSave(o, draft))} (default)`
                    : summarize(afterSave(o, draft))}
                </span>
              </div>
            </li>
          ))}
        </ul>

        <DialogFooter className="gap-2 p-6 pt-4">
          <DialogClose render={<Button variant="outline" />}>Cancel</DialogClose>
          <DialogClose render={<Button onClick={onConfirm} />}>
            Save {changes.length} change(s)
          </DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
