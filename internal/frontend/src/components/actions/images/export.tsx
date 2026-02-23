import { getV1ImagesFindOptions } from "@/client/@tanstack/react-query.gen";
import { Button } from "@/components/ui/button";
import { Card, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { useQuery } from "@tanstack/react-query";
import { Copy, LoaderCircle } from "lucide-react";
import { toast } from "sonner";

export default function ImagesExportAction({ images }: { images: string }) {
  const { data, isLoading, refetch } = useQuery(
    getV1ImagesFindOptions({ query: { names: images } }),
  );

  return (
    <Card>
      <CardHeader>
        <CardTitle>Export JSON</CardTitle>
      </CardHeader>
      <CardFooter>
        <Dialog>
          <DialogTrigger asChild>
            <Button onClick={() => refetch()}>Submit</Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Export JSON: {images}</DialogTitle>
            </DialogHeader>
            <div className="max-h-[calc(70dvh)] overflow-scroll">
              <div className="text-muted-foreground">
                {isLoading ? (
                  <LoaderCircle className="mx-auto animate-spin" />
                ) : (
                  <pre>{JSON.stringify(data, null, 4)}</pre>
                )}
              </div>
            </div>
            <DialogFooter>
              <Button
                type="button"
                onClick={() => {
                  navigator.clipboard.writeText(JSON.stringify(data, null, 4));
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
  );
}
