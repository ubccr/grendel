import { deleteV1BmcJobsMutation } from "@/client/@tanstack/react-query.gen";
import { Button } from "@/components/ui/button";
import { Card, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { ScrollArea } from "@/components/ui/scroll-area";
import { useMutation } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import { toast } from "sonner";

export default function JobsDeleteAction({ list }: { list: { [key: string]: string[] } }) {
  const { mutate, isPending } = useMutation(deleteV1BmcJobsMutation());

  return (
    <Card>
      <CardHeader>
        <CardTitle>Delete Jobs</CardTitle>
      </CardHeader>
      <CardFooter>
        <Dialog>
          <DialogTrigger asChild>
            <Button>
              {isPending ? <LoaderCircle className="animate-spin" /> : <span>Open</span>}
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Are you sure?</DialogTitle>
              <DialogDescription>
                The following nodes will have the listed jobs deleted: <br />
                <ScrollArea className="max-h-[80dvh] overflow-scroll">
                  <pre>{JSON.stringify(list, null, 4)}</pre>
                </ScrollArea>
              </DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <DialogClose asChild>
                <Button
                  variant="destructive"
                  onClick={() =>
                    mutate(
                      {
                        body: { node_job_list: list },
                      },
                      {
                        onSuccess: () => {
                          toast.success("Successfully deleted job(s) on node(s)");
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
  );
}
