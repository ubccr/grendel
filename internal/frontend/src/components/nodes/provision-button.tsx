import { LoaderCircle, Zap, ZapOff } from "lucide-react";
import { Button } from "../ui/button";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useEffect, useState } from "react";
import { usePatchV1NodesProvision } from "@/openapi/queries";

type Props = {
  name?: string;
  provision?: boolean;
};

export default function ProvisionIcon({ provision, name }: Props) {
  const mutate_provision = usePatchV1NodesProvision();
  const queryClient = useQueryClient();

  const [ping, setPing] = useState(false);

  useEffect(() => {
    if (ping) setTimeout(() => setPing(false), 2000);
  }, [ping]);

  return (
    <>
      {name == undefined || provision == undefined ? (
        <LoaderCircle />
      ) : (
        <Button
          size="sm"
          variant="outline"
          type="button"
          className={ping ? "animate-pulse" : ""}
          onClick={() => {
            setPing(true);
            mutate_provision.mutate(
              { query: { nodeset: name }, body: { provision: !provision } },
              {
                onSuccess: () => {
                  queryClient.invalidateQueries();
                },
                onError: (e) =>
                  toast.error(e.title, {
                    description: e.detail,
                  }),
              }
            );
          }}
        >
          {provision && <Zap className="text-green-600" />}
          {!provision && <ZapOff className="text-red-600" />}
        </Button>
      )}
    </>
  );
}
