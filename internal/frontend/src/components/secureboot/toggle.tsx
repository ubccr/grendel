import { postV1BmcSecurebootMutation } from "@/client/@tanstack/react-query.gen";
import { ConfirmDialog } from "@/components/actions/action-card";
import NodesetView from "@/components/nodeset-view";
import { Button } from "@/components/ui/button";
import { toastBmcResponse } from "@/lib/bmc";
import { useMutation } from "@tanstack/react-query";
import { LoaderCircle, ShieldCheck, ShieldOff } from "lucide-react";
import { toast } from "sonner";

type Props = {
  nodes: string;
  enable: boolean;
  variant?: "default" | "secondary" | "outline";
  onDone: () => void;
};

// SecureBootToggle turns secure boot on or off from the next boot. Turning it off asks first, since it drops the protection silently
export default function SecureBootToggle({ nodes, enable, variant = "default", onDone }: Props) {
  const { mutate, isPending } = useMutation(postV1BmcSecurebootMutation());
  const label = enable ? "Turn on" : "Turn off";

  const run = () =>
    mutate(
      { query: { nodeset: nodes }, body: { enabled: enable } },
      {
        onSuccess: (data) => {
          toastBmcResponse(enable ? "Turn on secure boot" : "Turn off secure boot", data);
          onDone();
        },
        onError: (e) => toast.error(e.title, { description: e.detail }),
      },
    );

  const button = (
    <Button variant={variant} disabled={isPending} onClick={enable ? run : undefined}>
      {isPending ? (
        <LoaderCircle className="animate-spin" />
      ) : enable ? (
        <ShieldCheck />
      ) : (
        <ShieldOff />
      )}
      {label}
    </Button>
  );

  if (enable) return button;
  return (
    <ConfirmDialog trigger={button} onConfirm={run}>
      From their next boot, these nodes will run boot loaders and kernels without checking their
      signatures:
      <NodesetView nodes={nodes} />
    </ConfirmDialog>
  );
}
