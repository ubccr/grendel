import { patchV1NodesProvisionMutation } from "@/client/@tanstack/react-query.gen";
import { useMutation } from "@tanstack/react-query";
import { LoaderCircle, Zap, ZapOff } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { Button } from "../ui/button";

type Props = {
  name?: string;
  provision?: boolean;
};

export default function ProvisionIcon({ provision, name }: Props) {
  const { mutate, isPending } = useMutation(patchV1NodesProvisionMutation());
  const [localProvision, setLocalProvision] = useState(provision);

  return (
    <>
      {name == undefined || provision == undefined ? (
        <LoaderCircle />
      ) : (
        <Button
          size="icon"
          variant="secondary"
          type="button"
          className={isPending ? "animate-pulse" : ""}
          onClick={() => {
            mutate(
              { query: { nodeset: name }, body: { provision: !localProvision } },
              {
                onSuccess: () => {
                  setLocalProvision(!localProvision);
                },
                onError: (e) =>
                  toast.error(e.title, {
                    description: e.detail,
                  }),
              },
            );
          }}
        >
          {localProvision ? (
            <Zap className="text-green-600" />
          ) : (
            <ZapOff className="text-red-600" />
          )}
        </Button>
      )}
    </>
  );
}
