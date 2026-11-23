import { deleteV1UsersUsernamesMutation } from "@/client/@tanstack/react-query.gen";
import { ConfirmAction } from "@/components/actions/action-card";

export default function UsersDeleteAction({ users }: { users: string }) {
  return (
    <ConfirmAction
      title="Delete"
      description="Permanently remove the selected users from Grendel."
      label="Delete"
      destructive
      mutation={deleteV1UsersUsernamesMutation()}
      variables={{ path: { usernames: users } }}
    >
      These users will be permanently removed from Grendel: {users}
    </ConfirmAction>
  );
}
