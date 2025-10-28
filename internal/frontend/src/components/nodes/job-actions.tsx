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
import { useDeleteV1BmcJobs, useGetV1BmcJobsKey } from "@/openapi/queries";

export default function JobActions({
  checked,
}: {
  checked: Map<string, string[]>;
}) {
  const queryClient = useQueryClient();
  const mutation_delete = useDeleteV1BmcJobs();

  return (
    <div className="mt-4 grid grid-cols-2 gap-4">
      <Card>
        <CardHeader>
          <CardTitle>Delete</CardTitle>
        </CardHeader>
        <CardFooter>
          <Dialog>
            <DialogTrigger asChild>
              <Button variant="destructive" disabled={checked.size < 1}>
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
                  WARNING: Selected jobs will be deleted!
                </DialogDescription>
              </DialogHeader>
              <DialogFooter>
                <DialogClose asChild>
                  <Button
                    variant="destructive"
                    onClick={() => {
                      const b: any = {};
                      checked.forEach((v, k) => (b[k] = v));
                      mutation_delete.mutate(
                        { body: { node_job_list: b } },
                        {
                          onSuccess: () => {
                            toast.success("Successfully deleted job(s)");
                            queryClient.invalidateQueries({
                              queryKey: [useGetV1BmcJobsKey],
                            });
                          },
                          onError: (e) =>
                            toast.error(e.title, {
                              description: e.detail,
                            }),
                        },
                      );
                    }}
                    disabled={checked.size < 1}
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
    </div>
  );
}
