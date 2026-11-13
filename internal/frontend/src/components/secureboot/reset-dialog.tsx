import { postV1BmcSecurebootResetMutation } from "@/client/@tanstack/react-query.gen";
import NodesetView from "@/components/nodeset-view";
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
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { toastBmcResponse } from "@/lib/bmc";
import { writable } from "@/lib/secureboot";
import { useMutation } from "@tanstack/react-query";
import { LoaderCircle, RotateCcw } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

const types = [
  {
    value: "to-default",
    label: "Restore the factory keys",
    detail:
      "Undoes any custom enrollment. This is the way back for a node that stopped booting because of its keys.",
    action: "Restore factory keys",
  },
  {
    value: "delete-all",
    label: "Delete every key",
    detail:
      "Leaves the node in setup mode, where secure boot enforces nothing until a new platform key is enrolled.",
    action: "Delete keys",
  },
  {
    value: "delete-pk",
    label: "Delete the platform key only",
    detail:
      "Also leaves the node in setup mode, but keeps KEK, db and dbx for when a new platform key is enrolled.",
    action: "Delete platform key",
  },
] as const;

type Type = (typeof types)[number]["value"];

// ResetDialog resets or empties the secure boot databases, every one or a single one
export default function ResetDialog({ nodes, onDone }: { nodes: string; onDone: () => void }) {
  const [open, setOpen] = useState(false);
  const [type, setType] = useState<Type>("to-default");
  const [scope, setScope] = useState("all");
  const { mutate, isPending } = useMutation(postV1BmcSecurebootResetMutation());
  const chosen = types.find((t) => t.value === type)!;
  const database = type === "delete-pk" || scope === "all" ? undefined : scope;

  const run = () =>
    mutate(
      { query: { nodeset: nodes }, body: { type, database } },
      {
        onSuccess: (data) => {
          toastBmcResponse(chosen.action, data);
          setOpen(false);
          onDone();
        },
        onError: (e) => toast.error(e.title, { description: e.detail }),
      },
    );

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button variant="outline" />}>
        <RotateCcw />
        Reset keys
      </DialogTrigger>
      <DialogContent className="max-h-dvh overflow-y-auto sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Reset secure boot keys</DialogTitle>
          <DialogDescription>Takes effect at the next boot.</DialogDescription>
        </DialogHeader>

        <RadioGroup value={type} onValueChange={(v) => setType(v as Type)} className="gap-3">
          {types.map((t) => (
            <Label
              key={t.value}
              htmlFor={`reset-${t.value}`}
              className="flex cursor-pointer items-start gap-3 rounded-lg border p-3 font-normal has-data-checked:border-foreground"
            >
              <RadioGroupItem id={`reset-${t.value}`} value={t.value} className="mt-0.5" />
              <span className="grid gap-1">
                <span className="font-medium">{t.label}</span>
                <span className="text-xs text-muted-foreground">{t.detail}</span>
              </span>
            </Label>
          ))}
        </RadioGroup>

        <div className="grid gap-1.5">
          <Label htmlFor="reset-scope">Databases</Label>
          <Select
            value={type === "delete-pk" ? "all" : scope}
            onValueChange={(v) => setScope(v ?? "all")}
            items={{
              all: "Every database",
              ...Object.fromEntries(writable.map((d) => [d, `Only ${d}`])),
            }}
            disabled={type === "delete-pk"}
          >
            <SelectTrigger id="reset-scope" className="w-48">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">Every database</SelectItem>
              {writable.map((d) => (
                <SelectItem key={d} value={d}>
                  Only {d}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="text-sm text-muted-foreground">
          Resets on:
          <NodesetView nodes={nodes} size="sm" />
        </div>

        <DialogFooter>
          <DialogClose render={<Button variant="secondary" />}>Cancel</DialogClose>
          <Button variant="destructive" disabled={isPending} onClick={run}>
            {isPending && <LoaderCircle className="animate-spin" />}
            {chosen.action}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
