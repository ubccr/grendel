import { deleteV1RolesNamesMutation } from "@/client/@tanstack/react-query.gen";
import ActionCard, { ConfirmDialog } from "@/components/actions/action-card";
import { Button } from "@/components/ui/button";
import { useMutation } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import { toast } from "sonner";

export default function RolesDeleteAction({ roles }: { roles: string }) {
  const { mutate, isPending } = useMutation(deleteV1RolesNamesMutation());

  return (
    <ActionCard
      title="Delete"
      description="Permanently remove the selected roles from Grendel."
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
              { path: { names: roles } },
              {
                onSuccess: (data) => toast.success(data?.title, { description: data?.detail }),
                onError: (e) => toast.error(e.title, { description: e.detail }),
              },
            )
          }
        >
          These roles will be permanently removed from Grendel: {roles}
        </ConfirmDialog>
      }
    />
  );
}
