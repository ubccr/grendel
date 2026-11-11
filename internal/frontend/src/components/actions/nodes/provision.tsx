import { patchV1NodesProvisionMutation } from "@/client/@tanstack/react-query.gen";
import ActionCard from "@/components/actions/action-card";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { useMutation } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

export default function NodesProvisionAction({ nodes }: { nodes: string }) {
  const [provision, setProvision] = useState(false);
  const { mutate, isPending } = useMutation(patchV1NodesProvisionMutation());

  return (
    <ActionCard
      title="Provision"
      description="While on, the node installs its boot image the next time it PXE boots."
      footer={
        <Button
          disabled={isPending}
          onClick={() =>
            mutate(
              { query: { nodeset: nodes }, body: { provision: provision } },
              {
                onSuccess: (data) => toast.success(data?.title, { description: data?.detail }),
                onError: (e) => toast.error(e.title, { description: e.detail }),
              },
            )
          }
        >
          {isPending && <LoaderCircle className="animate-spin" />}
          Apply
        </Button>
      }
    >
      <div className="flex items-center gap-2">
        <Switch id="provision" checked={provision} onCheckedChange={setProvision} />
        <Label htmlFor="provision">{provision ? "On" : "Off"}</Label>
      </div>
    </ActionCard>
  );
}
