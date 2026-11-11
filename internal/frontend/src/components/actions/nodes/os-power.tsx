import { postV1BmcPowerMutation } from "@/client/@tanstack/react-query.gen";
import ActionCard, { ConfirmDialog } from "@/components/actions/action-card";
import NodesetView from "@/components/nodeset-view";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { toastBmcResponse } from "@/lib/bmc";
import { useMutation } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

const powerOptions = {
  ForceRestart: "Power Cycle",
  On: "On",
  ForceOff: "Off",
  GracefulRestart: "Graceful Restart",
};

const bootOptions = {
  None: "None",
  Pxe: "PXE",
  BiosSetup: "BIOS",
  Utilities: "Utilities",
  Diags: "Diagnostics",
  Usb: "USB",
  Hdd: "HDD",
};

export default function NodesOsPowerAction({ nodes }: { nodes: string }) {
  const [powerOption, setPowerOption] = useState("");
  const [bootOption, setBootOption] = useState("None");
  const { mutate, isPending } = useMutation(postV1BmcPowerMutation());

  return (
    <ActionCard
      title="Host Power"
      description="Power the nodes on, off or cycle them, optionally booting a one time target."
      footer={
        <ConfirmDialog
          trigger={
            <Button disabled={isPending || !powerOption}>
              {isPending && <LoaderCircle className="animate-spin" />}
              Run
            </Button>
          }
          onConfirm={() =>
            mutate(
              {
                query: { nodeset: nodes },
                body: {
                  target: "host",
                  action: powerOption,
                  boot_override: bootOption === "None" ? "" : bootOption,
                },
              },
              {
                onSuccess: (data) => toastBmcResponse(`Power ${powerOption}`, data),
                onError: (e) => toast.error(e.title, { description: e.detail }),
              },
            )
          }
        >
          Power: {powerOptions[powerOption as keyof typeof powerOptions]}, boot:{" "}
          {bootOptions[bootOption as keyof typeof bootOptions]}
          <NodesetView nodes={nodes} />
        </ConfirmDialog>
      }
    >
      <div className="grid grid-cols-2 gap-2">
        <div className="grid gap-1.5">
          <Label htmlFor="power-option">Power</Label>
          <Select value={powerOption} onValueChange={setPowerOption}>
            <SelectTrigger id="power-option">
              <SelectValue placeholder="Action" />
            </SelectTrigger>
            <SelectContent>
              {Object.entries(powerOptions).map(([value, label]) => (
                <SelectItem key={value} value={value}>
                  {label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div className="grid gap-1.5">
          <Label htmlFor="boot-option">Boot</Label>
          <Select value={bootOption} onValueChange={setBootOption}>
            <SelectTrigger id="boot-option">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {Object.entries(bootOptions).map(([value, label]) => (
                <SelectItem key={value} value={value}>
                  {label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>
    </ActionCard>
  );
}
