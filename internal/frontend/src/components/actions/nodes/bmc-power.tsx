import { postV1BmcPowerMutation } from "@/client/@tanstack/react-query.gen";
import ActionCard, { ConfirmDialog } from "@/components/actions/action-card";
import NodesetView from "@/components/nodeset-view";
import { Button } from "@/components/ui/button";
import { toastBmcResponse } from "@/lib/bmc";
import { useMutation } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import { toast } from "sonner";

export default function NodesBmcPowerAction({ nodes }: { nodes: string }) {
  const { mutate, isPending } = useMutation(postV1BmcPowerMutation());

  return (
    <ActionCard
      title="BMC Reboot"
      description="Restart the BMCs. The hosts keep running, but the BMCs are unreachable for a few minutes."
      footer={
        <ConfirmDialog
          trigger={
            <Button disabled={isPending}>
              {isPending && <LoaderCircle className="animate-spin" />}
              Reboot
            </Button>
          }
          onConfirm={() =>
            mutate(
              { query: { nodeset: nodes }, body: { target: "bmc" } },
              {
                onSuccess: (data) => toastBmcResponse("BMC reboot", data),
                onError: (e) => toast.error(e.title, { description: e.detail }),
              },
            )
          }
        >
          These BMCs will be unavailable while they reboot:
          <NodesetView nodes={nodes} />
        </ConfirmDialog>
      }
    />
  );
}
