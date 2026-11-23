import { patchV1UsersUsernamesEnableMutation } from "@/client/@tanstack/react-query.gen";
import ActionCard from "@/components/actions/action-card";
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

const states = { true: "Enabled", false: "Disabled" };

export default function UsersEnabledAction({ users }: { users: string }) {
  const [enabled, setEnabled] = useState<keyof typeof states>();
  const { mutate, isPending } = useMutation(patchV1UsersUsernamesEnableMutation());

  return (
    <ActionCard
      title="Account access"
      description="A disabled account keeps its role and password, but can't sign in."
      footer={
        <Button
          disabled={isPending || enabled === undefined}
          onClick={() =>
            mutate(
              { path: { usernames: users }, body: { enabled: enabled === "true" } },
              { onSuccess: (data) => toast.success(data?.title, { description: data?.detail }) },
            )
          }
        >
          {isPending && <LoaderCircle className="animate-spin" />}
          Apply
        </Button>
      }
    >
      <Select
        value={enabled ?? null}
        onValueChange={(v) => setEnabled(v ?? undefined)}
        items={states}
      >
        <SelectTrigger>
          <SelectValue placeholder="Enable or disable" />
        </SelectTrigger>
        <SelectContent>
          {Object.entries(states).map(([value, label]) => (
            <SelectItem key={value} value={value}>
              {label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </ActionCard>
  );
}
