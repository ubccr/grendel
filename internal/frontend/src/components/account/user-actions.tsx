import {
  useDeleteV1UsersUsernames,
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

  const mutation_delete = useDeleteV1UsersUsernames();
  const mutation_role = usePatchV1UsersUsernamesRole();

  return (
    <div className="mt-4 grid sm:grid-cols-2 gap-4">
      <Card>
        <CardHeader>
          <CardTitle>Delete</CardTitle>
        </CardHeader>
        <CardFooter>
          <Dialog>
            <DialogTrigger asChild>
              <Button size="sm" variant="destructive">
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
                    size="sm"
                    onClick={() =>
                      mutation_delete.mutate(
                        { path: { usernames: users } },
                        {
                          onSuccess: () => {
                            toast.success("Successfully deleted user(s)");
                            queryClient.invalidateQueries();
                          },
                          onError: () =>
                            toast.error("Failed to delete user(s)", {
                              // description: e.message,
                            }),
                        }
                      )
                    }
                  >
                    Confirm
                  </Button>
                </DialogClose>
                <DialogClose asChild>
                  <Button variant="outline" size="sm">
                    Cancel
                  </Button>
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
          <Select onValueChange={(e) => setUserRole(e)}>
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="Action" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="disabled">Disabled</SelectItem>
              <SelectItem value="user">User</SelectItem>
              <SelectItem value="admin">Admin</SelectItem>
            </SelectContent>
          </Select>
        </CardContent>
        <CardFooter>
          <Button
            variant="outline"
            size="sm"
            onClick={() =>
              mutation_role.mutate(
                {
                  path: { usernames: users },
                  body: { role: userRole },
                },
                {
                  onSuccess: () => {
                    queryClient.invalidateQueries();
                    toast.success("Successfully changed user(s) role");
                  },
                  onError: () =>
                    toast.error("Failed to change user(s) role", {
                      // description: e.message,
                    }),
                }
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
    </div>
  );
}
