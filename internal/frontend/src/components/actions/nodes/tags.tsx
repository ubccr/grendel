import { patchV1NodesTagsActionMutation } from "@/client/@tanstack/react-query.gen";
import ActionCard from "@/components/actions/action-card";
import { TagsInput } from "@/components/tags-input";
import { Button } from "@/components/ui/button";
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
    <ActionCard
      title="Tags"
      description="Add tags to the nodes or remove them."
      footer={
        <Button
          disabled={isPending || !action || tags.length === 0}
          onClick={() =>
            mutate(
              {
                path: { action: action },
                query: { nodeset: nodes },
                body: { tags: tags.join(",") },
              },
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
      <Select onValueChange={setAction}>
        <SelectTrigger>
          <SelectValue placeholder="Add or remove" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="add">Add</SelectItem>
          <SelectItem value="remove">Remove</SelectItem>
        </SelectContent>
      </Select>
      <TagsInput value={tags} onValueChange={setTags} placeholder="Tags" />
    </ActionCard>
  );
}
