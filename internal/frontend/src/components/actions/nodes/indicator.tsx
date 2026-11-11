import { postV1BmcIndicatorMutation } from "@/client/@tanstack/react-query.gen";
import ActionCard from "@/components/actions/action-card";
import { Button } from "@/components/ui/button";
import { toastBmcResponse } from "@/lib/bmc";
import { useMutation } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

// the indicator only lights an LED, so unlike the other bmc actions it runs without a confirmation
export default function NodesIndicatorAction({ nodes }: { nodes: string }) {
  const [pending, setPending] = useState<"on" | "off">();
  const { mutate, isPending } = useMutation(postV1BmcIndicatorMutation());

  const set = (state: "on" | "off") => {
    setPending(state);
    mutate(
      { query: { nodeset: nodes }, body: { state } },
      {
        onSuccess: (data) => toastBmcResponse(`Locate LED ${state}`, data),
        onError: (e) => toast.error(e.title, { description: e.detail }),
        onSettled: () => setPending(undefined),
      },
    );
  };

  return (
    <ActionCard
      title="Locate LED"
      description="Blink the identify light on the nodes so they can be found in the rack."
      footer={
        <>
          <Button variant="secondary" disabled={isPending} onClick={() => set("off")}>
            {pending === "off" && <LoaderCircle className="animate-spin" />}
            Off
          </Button>
          <Button disabled={isPending} onClick={() => set("on")}>
            {pending === "on" && <LoaderCircle className="animate-spin" />}
            On
          </Button>
        </>
      }
    />
  );
}
