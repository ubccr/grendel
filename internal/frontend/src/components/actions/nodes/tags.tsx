import {
  getV1NodesTagsQueryKey,
  patchV1NodesTagsActionMutation,
} from "@/client/@tanstack/react-query.gen";
import ActionCard from "@/components/actions/action-card";
import { TagCombobox } from "@/components/tag-combobox";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

export default function NodesTagsAction({ nodes }: { nodes: string }) {
  const [tags, setTags] = useState<string[]>([]);
  const [action, setAction] = useState("");
  const { mutate, isPending } = useMutation(patchV1NodesTagsActionMutation());
  const queryClient = useQueryClient();

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
                onSuccess: (data) => {
                  toast.success(data?.title, { description: data?.detail });
                  queryClient.invalidateQueries({ queryKey: getV1NodesTagsQueryKey() });
                },
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
      <Select
        onValueChange={(v: string | null) => setAction(v ?? "")}
        items={{ add: "Add", remove: "Remove" }}
      >
        <SelectTrigger>
          <SelectValue placeholder="Add or remove" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="add">Add</SelectItem>
          <SelectItem value="remove">Remove</SelectItem>
        </SelectContent>
      </Select>
      <TagCombobox value={tags} onValueChange={setTags} placeholder="Tags" />
    </ActionCard>
  );
}
