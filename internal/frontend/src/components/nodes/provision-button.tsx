import { patchV1NodesProvisionMutation } from "@/client/@tanstack/react-query.gen";
import { useMutation } from "@tanstack/react-query";
import { LoaderCircle, Zap, ZapOff } from "lucide-react";
import { Button } from "../ui/button";

type Props = {
  name?: string;
  provision?: boolean;
};

export default function ProvisionIcon({ provision, name }: Props) {
  const { mutate, isPending, variables } = useMutation(patchV1NodesProvisionMutation());
  // what was asked for stands in while the change is stored and the nodes are read again
  const on = isPending ? variables.body.provision : provision;

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
          onClick={() => mutate({ query: { nodeset: name }, body: { provision: !on } })}
        >
          {on ? <Zap className="text-green-600" /> : <ZapOff className="text-red-600" />}
        </Button>
      )}
    </>
  );
}
