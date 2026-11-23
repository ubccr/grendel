import { deleteV1RolesNamesMutation } from "@/client/@tanstack/react-query.gen";
import { ConfirmAction } from "@/components/actions/action-card";

export default function RolesDeleteAction({ roles }: { roles: string }) {
  return (
    <ConfirmAction
      title="Delete"
      description="Permanently remove the selected roles from Grendel."
      label="Delete"
      destructive
      mutation={deleteV1RolesNamesMutation()}
      variables={{ path: { names: roles } }}
    >
      These roles will be permanently removed from Grendel: {roles}
    </ConfirmAction>
  );
}
