import { patchV1NodesProvisionMutation } from "@/client/@tanstack/react-query.gen";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Switch } from "@/components/ui/switch";
import { useMutation } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

export default function NodesProvisionAction({ nodes }: { nodes: string }) {
  const [provision, setProvision] = useState(false);

  const { mutate, isPending } = useMutation(patchV1NodesProvisionMutation());
  return (
    <Card>
      <CardHeader>
        <CardTitle>Provision</CardTitle>
      </CardHeader>
      <CardContent>
        <Switch onCheckedChange={(e) => setProvision(e)} />
      </CardContent>
      <CardFooter>
        <Button
          onClick={() =>
            mutate(
              { query: { nodeset: nodes }, body: { provision: provision } },
              {
                onSuccess: (data) => {
                  toast.success(data?.title, {
                    description: data?.detail,
                  });
                },
                onError: (e) =>
                  toast.error(e.title, {
                    description: e.detail,
                  }),
              },
            )
          }
        >
          {isPending ? <LoaderCircle className="animate-spin" /> : <span>Submit</span>}
        </Button>
      </CardFooter>
    </Card>
  );
}
