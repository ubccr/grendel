import { getV1NodesFindOptions } from "@/client/@tanstack/react-query.gen";
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
import { Copy } from "lucide-react";
import { toast } from "sonner";

export default function NodesExportAction({ nodes }: { nodes: string }) {
  const { data, refetch } = useQuery(getV1NodesFindOptions({ query: { nodeset: nodes } }));
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
          <DialogContent className="max-w-1/2">
            <DialogHeader>
              <DialogTitle>Export JSON:</DialogTitle>
            </DialogHeader>
            <div className="max-h-[80dvh]">
              <pre className="max-h-full overflow-scroll text-muted-foreground">
                {JSON.stringify(data, null, 4)}
              </pre>
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
