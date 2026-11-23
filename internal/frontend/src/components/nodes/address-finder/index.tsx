import { cellInput } from "@/components/form-layout";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogTrigger } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import { Radar } from "lucide-react";
import { useRef, useState } from "react";
import { Finder } from "./finder";

export type Port = { bmc: boolean; others: string[] };

// port gives whether this port is a BMC, which decides the ranges suggested first, and the addresses on the node's other ports, which count as taken though they aren't saved yet. It is read when the finder opens, as the ports' list doesn't render again when a port's fields change
export function AddressInput({
  value,
  onChange,
  invalid,
  port,
  label,
}: {
  value: string;
  onChange: (v: string) => void;
  invalid?: boolean;
  port: () => Port;
  label: string;
}) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [open, setOpen] = useState<Port>();

  return (
    <div className="relative">
      <Input
        ref={inputRef}
        aria-invalid={invalid}
        className={cn(cellInput, "pr-8")}
        value={value}
        onChange={(e) => onChange(e.target.value)}
      />
      <Dialog open={!!open} onOpenChange={(o) => setOpen(o ? port() : undefined)}>
        <DialogTrigger
          render={
            <Button
              type="button"
              size="icon"
              variant="ghost"
              className="absolute top-1/2 right-0.5 size-7 -translate-y-1/2 text-muted-foreground hover:text-foreground"
            />
          }
        >
          <Radar />
          <span className="sr-only">Find a free address for {label}</span>
        </DialogTrigger>
        {/* a fixed height held near the top of the screen, so the dialog stays put whatever it shows and anything longer scrolls inside it */}
        <DialogContent
          className="top-[8vh] h-[min(38rem,84vh)] translate-y-0 content-start overflow-y-auto p-5 sm:max-w-md"
          finalFocus={inputRef}
        >
          {open && (
            <Finder
              value={value}
              port={open}
              label={label}
              onPick={(ip) => {
                onChange(ip);
                setOpen(undefined);
              }}
            />
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
