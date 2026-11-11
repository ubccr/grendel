import { deleteV1UsersUsernamesMutation } from "@/client/@tanstack/react-query.gen";
import ActionCard, { ConfirmDialog } from "@/components/actions/action-card";
import { Button } from "@/components/ui/button";
import { useMutation } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import { toast } from "sonner";

export default function UsersDeleteAction({ users }: { users: string }) {
  const { mutate, isPending } = useMutation(deleteV1UsersUsernamesMutation());

  return (
    <ActionCard
      title="Delete"
      description="Permanently remove the selected users from Grendel."
      footer={
        <ConfirmDialog
          trigger={
            <Button variant="destructive" disabled={isPending}>
              {isPending && <LoaderCircle className="animate-spin" />}
              Delete
            </Button>
          }
          onConfirm={() =>
            mutate(
              { path: { usernames: users } },
              {
                onSuccess: (data) => toast.success(data?.title, { description: data?.detail }),
                onError: (e) => toast.error(e.title, { description: e.detail }),
              },
            )
          }
        >
          These users will be permanently removed from Grendel: {users}
        </ConfirmDialog>
      }
    />
  );
}
