import {
  getV1RolesOptions,
  patchV1UsersUsernamesRoleMutation,
} from "@/client/@tanstack/react-query.gen";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
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
  const [userRole, setUserRole] = useState("");

  const { data, isFetching } = useQuery(getV1RolesOptions());
  const { mutate, isPending } = useMutation(patchV1UsersUsernamesRoleMutation());
  return (
    <Card>
      <CardHeader>
        <CardTitle>Change Role</CardTitle>
      </CardHeader>
      <CardContent className="grid grid-cols-1 gap-2">
        {isFetching ? (
          <LoaderCircle className="animate-spin" />
        ) : (
          <Select onValueChange={(e) => setUserRole(e)}>
            <SelectTrigger className="w-45">
              <SelectValue placeholder="Action" />
            </SelectTrigger>
            <SelectContent>
              {data?.roles?.map((role, i) => (
                <SelectItem key={i} value={role.name ?? ""}>
                  {role.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}
      </CardContent>
      <CardFooter>
        <Button
          onClick={() =>
            mutate(
              {
                path: { usernames: users },
                body: { role: userRole },
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
