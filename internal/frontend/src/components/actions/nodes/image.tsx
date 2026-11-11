import { patchV1NodesImageMutation } from "@/client/@tanstack/react-query.gen";
import ActionCard from "@/components/actions/action-card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useMutation } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

export default function NodesImageAction({ nodes }: { nodes: string }) {
  const [image, setImage] = useState("");
  const { mutate, isPending } = useMutation(patchV1NodesImageMutation());

  return (
    <ActionCard
      title="Boot Image"
      description="Set the image the nodes boot when they provision."
      footer={
        <Button
          disabled={isPending}
          onClick={() =>
            mutate(
              { query: { nodeset: nodes }, body: { image: image } },
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
      <Input placeholder="Image name" value={image} onChange={(e) => setImage(e.target.value)} />
    </ActionCard>
  );
}
