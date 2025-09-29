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
import { Copy, LoaderCircle } from "lucide-react";
import { useDeleteV1Images, useGetV1ImagesFind } from "@/openapi/queries";

export default function ImageActions({ images }: { images: string }) {
  const queryClient = useQueryClient();
  const mutation_delete = useDeleteV1Images();
  const images_query = useGetV1ImagesFind(
    { query: { names: images } },
    undefined,
    {
      enabled: false,
    },
  );

  return (
    <div className="mt-4 grid gap-4 sm:grid-cols-2">
      <Card>
        <CardHeader>
          <CardTitle>Delete</CardTitle>
        </CardHeader>
        <CardFooter>
          <Dialog>
            <DialogTrigger asChild>
              <Button variant="destructive">Delete</Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Are you sure?</DialogTitle>
                <DialogDescription>
                  WARNING: Selected images: ({images}) will be permanently
                  removed from Grendel!
                </DialogDescription>
              </DialogHeader>
              <DialogFooter>
                <DialogClose asChild>
                  <Button
                    variant="destructive"
                    onClick={() =>
                      mutation_delete.mutate(
                        { query: { name: images } },
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
          <CardTitle>Export JSON</CardTitle>
        </CardHeader>
        <CardFooter>
          <Dialog>
            <DialogTrigger asChild>
              <Button onClick={() => images_query.refetch()}>Submit</Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Export JSON: {images}</DialogTitle>
              </DialogHeader>
              <div className="max-h-[calc(70dvh)] overflow-scroll">
                <div className="text-muted-foreground">
                  {images_query.isLoading ? (
                    <LoaderCircle className="mx-auto animate-spin" />
                  ) : (
                    <pre>{JSON.stringify(images_query.data, null, 4)}</pre>
                  )}
                </div>
              </div>
              <DialogFooter>
                <Button
                  type="button"
                  onClick={() => {
                    navigator.clipboard.writeText(
                      JSON.stringify(images_query.data, null, 4),
                    );
                    toast.success("Successfully copied JSON to clipboard");
                  }}
                >
                  <Copy />
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </CardFooter>
      </Card>
    </div>
  );
}
