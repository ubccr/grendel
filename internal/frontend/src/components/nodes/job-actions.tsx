import { Card, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { useQueryClient } from "@tanstack/react-query";
import { Button } from "../ui/button";
import { toast } from "sonner";
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
import { LoaderCircle } from "lucide-react";
import { useDeleteV1BmcJobsJids, useGetV1BmcJobsKey } from "@/openapi/queries";

export default function JobActions({
  nodes,
  jids,
}: {
  nodes: string;
  jids: string;
}) {
  const queryClient = useQueryClient();
  const mutation_delete = useDeleteV1BmcJobsJids();

  return (
    <div className="mt-4 grid grid-cols-2 gap-4">
      <Card>
        <CardHeader>
          <CardTitle>Delete</CardTitle>
        </CardHeader>
        <CardFooter>
          <Dialog>
            <DialogTrigger asChild>
              <Button size="sm" variant="destructive" disabled={jids === ""}>
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
                <DialogDescription className="break-all">
                  WARNING: Selected jobs: ({jids}) will be removed from node: (
                  {nodes})!
                </DialogDescription>
              </DialogHeader>
              <DialogFooter>
                <DialogClose asChild>
                  <Button
                    variant="destructive"
                    size="sm"
                    onClick={() =>
                      mutation_delete.mutate(
                        { path: { jids: jids }, query: { nodeset: nodes } },
                        {
                          onSuccess: () => {
                            toast.success("Successfully deleted job(s)");
                            queryClient.invalidateQueries({
                              queryKey: [useGetV1BmcJobsKey],
                            });
                          },
                          onError: () =>
                            toast.error("Failed to delete job(s)", {
                              // description: e.message,
                            }),
                        }
                      )
                    }
                    disabled={jids === ""}
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
