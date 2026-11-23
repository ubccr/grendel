import { postV1BmcSecurebootMutation } from "@/client/@tanstack/react-query.gen";
import { ConfirmDialog } from "@/components/actions/action-card";
import NodesetView from "@/components/nodeset-view";
import { Button } from "@/components/ui/button";
import { useMutation } from "@tanstack/react-query";
import { LoaderCircle, ShieldCheck, ShieldOff } from "lucide-react";

type Props = {
  nodes: string;
  enable: boolean;
  variant?: "default" | "secondary" | "outline";
};

// Turning secure boot off asks first, since it drops the protection silently
export default function SecureBootToggle({ nodes, enable, variant = "default" }: Props) {
  const { mutate, isPending } = useMutation({
    ...postV1BmcSecurebootMutation(),
    meta: { bmc: enable ? "Turn on secure boot" : "Turn off secure boot" },
  });
  const label = enable ? "Turn on" : "Turn off";

  const run = () => mutate({ query: { nodeset: nodes }, body: { enabled: enable } });

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
