import {
  useDeleteV1UsersUsernames,
  useGetV1Roles,
  usePatchV1UsersUsernamesEnable,
  usePatchV1UsersUsernamesRole,
} from "@/openapi/queries";
import { Button } from "../ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "../ui/card";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../ui/dialog";
import { toast } from "sonner";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";

export default function UserActions({ users }: { users: string }) {
  const queryClient = useQueryClient();
  const [userRole, setUserRole] = useState("");
  const [userEnabled, setUserEnabled] = useState("");

  const query_roles = useGetV1Roles();

  const mutation_delete = useDeleteV1UsersUsernames();
  const mutation_role = usePatchV1UsersUsernamesRole();
  const mutation_enabled = usePatchV1UsersUsernamesEnable();

  return (
    <div className="mt-4 grid gap-4 sm:grid-cols-2">
      <Card>
        <CardHeader>
          <CardTitle>Delete</CardTitle>
        </CardHeader>
        <CardFooter>
          <Dialog>
            <DialogTrigger asChild>
              <Button variant="destructive">
                {mutation_delete.isPending ? (
                  <LoaderCircle className="animate-spin" />
                ) : (
                  <span>Delete</span>
                )}
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Are you sure?</DialogTitle>
                <DialogDescription>
                  WARNING: Selected users: ({users}) will be permanently removed
                  from Grendel!
                </DialogDescription>
              </DialogHeader>
              <DialogFooter>
                <DialogClose asChild>
                  <Button
                    variant="destructive"
                    onClick={() =>
                      mutation_delete.mutate(
                        { path: { usernames: users } },
                        {
                          onSuccess: (e) => {
                            toast.success(e.data?.title, {
                              description: e.data?.detail,
                            });
                            queryClient.invalidateQueries();
                          },
                          onError: (e) =>
                            toast.error(e.title, {
                              description: e.detail,
                            }),
                        },
                      )
                    }
                  >
                    Confirm
                  </Button>
                </DialogClose>
                <DialogClose asChild>
                  <Button>Cancel</Button>
                </DialogClose>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </CardFooter>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>Change Role</CardTitle>
        </CardHeader>
        <CardContent className="grid grid-cols-1 gap-2">
          {query_roles.isFetching ? (
            <LoaderCircle className="animate-spin" />
          ) : (
            <Select onValueChange={(e) => setUserRole(e)}>
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Action" />
              </SelectTrigger>
              <SelectContent>
                {query_roles.data?.roles?.map((role, i) => (
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
              mutation_role.mutate(
                {
                  path: { usernames: users },
                  body: { role: userRole },
                },
                {
                  onSuccess: (e) => {
                    toast.success(e.data?.title, {
                      description: e.data?.detail,
                    });
                    queryClient.invalidateQueries();
                  },
                  onError: (e) =>
                    toast.error(e.title, {
                      description: e.detail,
                    }),
                },
              )
            }
          >
            {mutation_role.isPending ? (
              <LoaderCircle className="animate-spin" />
            ) : (
              <span>Submit</span>
            )}
          </Button>
        </CardFooter>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>Change Enabled flag</CardTitle>
        </CardHeader>
        <CardContent className="grid grid-cols-1 gap-2">
          <Select onValueChange={(e) => setUserEnabled(e)}>
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="Enabled" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="true">True</SelectItem>
              <SelectItem value="false">False</SelectItem>
            </SelectContent>
          </Select>
        </CardContent>
        <CardFooter>
          <Button
            onClick={() =>
              mutation_enabled.mutate(
                {
                  path: { usernames: users },
                  body: { enabled: userEnabled === "true" },
                },
                {
                  onSuccess: (e) => {
                    toast.success(e.data?.title, {
                      description: e.data?.detail,
                    });
                    queryClient.invalidateQueries();
                  },
                  onError: (e) =>
                    toast.error(e.title, {
                      description: e.detail,
                    }),
                },
              )
            }
          >
            {mutation_enabled.isPending ? (
              <LoaderCircle className="animate-spin" />
            ) : (
              <span>Submit</span>
            )}
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
}
