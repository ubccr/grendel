import { postV1BmcPowerBmcMutation } from "@/client/@tanstack/react-query.gen";
import NodesetView from "@/components/nodeset-view";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
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

export default function NodesBmcPowerAction({ nodes }: { nodes: string }) {
  const { mutate, isPending } = useMutation(postV1BmcPowerBmcMutation());
  return (
    <Card>
      <CardHeader>
        <CardTitle>BMC PowerCycle</CardTitle>
      </CardHeader>
      <CardContent className="grid grid-cols-1 gap-2"></CardContent>
      <CardFooter>
        <Dialog>
          <DialogTrigger asChild>
            <Button>
              {isPending ? <LoaderCircle className="animate-spin" /> : <span>Submit</span>}
            </Button>
          </DialogTrigger>
          <DialogContent className="max-h-dvh overflow-scroll">
            <DialogHeader>
              <DialogTitle>Are you sure?</DialogTitle>
              <DialogDescription>
                bmc(s) will be unavailable while they reboot
                <br />
                <NodesetView nodes={nodes} />
              </DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <DialogClose asChild>
                <Button
                  variant="destructive"
                  onClick={() =>
                    mutate(
                      {
                        query: { nodeset: nodes },
                      },
                      {
                        onSuccess: () => {
                          toast.success("Successfully rebooted bmc(s)");
                        },
                        onError: (data) =>
                          toast.error(data.title, {
                            description: data.detail,
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
