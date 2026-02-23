import { patchV1NodesImageMutation } from "@/client/@tanstack/react-query.gen";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { useMutation } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

export default function NodesImageAction({ nodes }: { nodes: string }) {
  const [image, setImage] = useState("");
  const { mutate, isPending } = useMutation(patchV1NodesImageMutation());
  return (
    <Card>
      <CardHeader>
        <CardTitle>Boot Image</CardTitle>
      </CardHeader>
      <CardContent className="grid grid-cols-1 gap-2">
        <Input value={image} onChange={(e) => setImage(e.target.value)} />
      </CardContent>
      <CardFooter>
        <Button
          onClick={() =>
            mutate(
              {
                query: { nodeset: nodes },
                body: { image: image },
              },
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
