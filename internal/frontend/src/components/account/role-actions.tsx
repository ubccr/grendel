import { useDeleteV1RolesNames } from "@/openapi/queries";
import { Button } from "../ui/button";
import { Card, CardFooter, CardHeader, CardTitle } from "../ui/card";
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
import { useQueryClient } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";

export default function RoleActions({ roles }: { roles: string }) {
  const queryClient = useQueryClient();

  const mutation_delete = useDeleteV1RolesNames();

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
                  WARNING: Selected roles: ({roles}) will be permanently removed
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
                        { path: { names: roles } },
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
    </div>
  );
}
