import { postV1BmcConfigureAutoMutation } from "@/client/@tanstack/react-query.gen";
import ActionCard, { ConfirmDialog } from "@/components/actions/action-card";
import NodesetView from "@/components/nodeset-view";
import { Button } from "@/components/ui/button";
import { toastBmcResponse } from "@/lib/bmc";
import { useMutation } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import { toast } from "sonner";

export default function NodesAutoConfigAction({ nodes }: { nodes: string }) {
  const { mutate, isPending } = useMutation(postV1BmcConfigureAutoMutation());

  return (
    <ActionCard
      title="BMC Auto Configure"
      description="Set the iDRAC to auto configure once, from DHCP, on the node's next boot."
      footer={
        <ConfirmDialog
          trigger={
            <Button disabled={isPending}>
              {isPending && <LoaderCircle className="animate-spin" />}
              Enable
            </Button>
          }
          onConfirm={() =>
            mutate(
              { query: { nodeset: nodes } },
              {
                onSuccess: (data) => toastBmcResponse("Auto configure", data),
                onError: (e) => toast.error(e.title, { description: e.detail }),
              },
            )
          }
        >
          NIC.1.AutoConfig will be set to Enable Once on:
          <NodesetView nodes={nodes} />
        </ConfirmDialog>
      }
    />
  );
}
