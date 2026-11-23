import { NodeRanges } from "@/components/nodes/address-finder/node-ranges";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useStoredValue } from "@/hooks/use-stored-value";
import { parseRange, rangeKey } from "@/lib/addresses";
import { Radar } from "lucide-react";
import { useId, useState } from "react";

export function RangePicker({
  bmc,
  value,
  onChange,
}: {
  bmc: boolean;
  value: string;
  onChange: (range: string) => void;
}) {
  const id = useId();
  const [, remember] = useStoredValue(rangeKey(bmc));
  const [open, setOpen] = useState(false);
  const bad = value.trim() !== "" && !parseRange(value);

  const pick = (r: string) => {
    onChange(r);
    remember(r);
  };

  return (
    <div className="grid content-start gap-1.5">
      <Label htmlFor={id}>{bmc ? "BMC address range" : "Boot address range"}</Label>
      <div className="flex gap-1.5">
        <Input
          id={id}
          value={value}
          placeholder="10.17.25.0/24"
          spellCheck={false}
          aria-invalid={bad}
          aria-describedby={`${id}-note`}
          className="h-10 font-mono text-[15px]"
          onChange={(e) => onChange(e.target.value)}
          onBlur={() => parseRange(value) && remember(value.trim())}
        />
        <Dialog open={open} onOpenChange={setOpen}>
          <DialogTrigger render={<Button variant="outline" className="h-10" />}>
            <Radar />
            Find
          </DialogTrigger>
          <DialogContent className="top-[8vh] max-h-[84vh] translate-y-0 content-start overflow-y-auto p-5 sm:max-w-md">
            <DialogTitle>Ranges other nodes use</DialogTitle>
            <DialogDescription>
              Pick nodes like the new ones, such as the rack next door, to use the range their{" "}
              {bmc ? "BMCs" : "interfaces"} are in.
            </DialogDescription>
            <NodeRanges
              bmc={bmc}
              current={parseRange(value) ?? undefined}
              onPick={(r) => {
                pick(r);
                setOpen(false);
              }}
            />
          </DialogContent>
        </Dialog>
      </div>
      <p id={`${id}-note`} className="text-xs text-muted-foreground">
        {bad ? (
          <span className="text-destructive">
            Enter a range like 10.17.25.0/24, no larger than a /16.
          </span>
        ) : (
          "Addresses are given in order, skipping ones any node holds."
        )}
      </p>
    </div>
  );
}
