import { deleteV1ImagesMutation } from "@/client/@tanstack/react-query.gen";
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
import { useMutation } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import { toast } from "sonner";

export default function ImagesDeleteAction({ images }: { images: string }) {
  const { mutate, isPending } = useMutation(deleteV1ImagesMutation());
  return (
    <Card>
      <CardHeader>
        <CardTitle>Delete</CardTitle>
      </CardHeader>
      <CardFooter>
        <Dialog>
          <DialogTrigger asChild>
            <Button>
              {isPending ? <LoaderCircle className="animate-spin" /> : <span>Submit</span>}
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Are you sure?</DialogTitle>
              <DialogDescription>
                WARNING: Selected images will be permanently removed from Grendel!
              </DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <DialogClose asChild>
                <Button
                  variant="destructive"
                  onClick={() =>
                    mutate(
                      { query: { names: images } },
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
                  Delete
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
