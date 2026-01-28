import { patchV1NodesTagsActionMutation } from "@/client/@tanstack/react-query.gen";
import { TagsInput } from "@/components/tags-input";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useMutation } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

export default function NodesTagsAction({ nodes }: { nodes: string }) {
  const [tags, setTags] = useState<string[]>([]);
  const [action, setAction] = useState("");

  const { mutate, isPending } = useMutation(patchV1NodesTagsActionMutation());
  return (
    <Card>
      <CardHeader>
        <CardTitle>Tags</CardTitle>
      </CardHeader>
      <CardContent className="grid grid-cols-1 gap-2">
        <TagsInput value={tags} onValueChange={setTags} placeholder="Tags" />
        <Select onValueChange={(e) => setAction(e)}>
          <SelectTrigger className="w-[180px]">
            <SelectValue placeholder="Action" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="add">Add</SelectItem>
            <SelectItem value="remove">Remove</SelectItem>
          </SelectContent>
        </Select>
      </CardContent>
      <CardFooter>
        <Button
          onClick={() =>
            mutate(
              {
                path: { action: action },
                query: { nodeset: nodes },
                body: { tags: tags.join(",") },
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
