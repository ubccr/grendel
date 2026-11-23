import {
  getV1RolesOptions,
  patchV1UsersUsernamesRoleMutation,
} from "@/client/@tanstack/react-query.gen";
import ActionCard from "@/components/actions/action-card";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useMutation, useQuery } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

export default function UsersRoleAction({ users }: { users: string }) {
  const [role, setRole] = useState("");
  const { data, isPending: loading } = useQuery(getV1RolesOptions());
  const { mutate, isPending } = useMutation(patchV1UsersUsernamesRoleMutation());
  const names = (data?.roles ?? []).flatMap((r) => (r.name ? [r.name] : []));

  return (
    <ActionCard
      title="Role"
      description="Give the selected users another role, and with it that role's permissions."
      footer={
        <Button
          disabled={isPending || !role}
          onClick={() =>
            mutate(
              { path: { usernames: users }, body: { role } },
              { onSuccess: (data) => toast.success(data?.title, { description: data?.detail }) },
            )
          }
        >
          {isPending && <LoaderCircle className="animate-spin" />}
          Apply
        </Button>
      }
    >
      <Select value={role || null} onValueChange={(v) => setRole(v ?? "")} disabled={loading}>
        <SelectTrigger>
          <SelectValue placeholder={loading ? "Loading roles" : "Role"} />
        </SelectTrigger>
        <SelectContent>
          {names.map((name) => (
            <SelectItem key={name} value={name}>
              {name}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </ActionCard>
  );
}
