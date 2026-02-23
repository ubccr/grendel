import { postV1BmcPowerOsMutation } from "@/client/@tanstack/react-query.gen";
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
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useMutation } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

export default function NodesOsPowerAction({ nodes }: { nodes: string }) {
  const [powerOption, setPowerOption] = useState("");
  const [bootOption, setBootOption] = useState("");

  const { mutate, isPending } = useMutation(postV1BmcPowerOsMutation());
  return (
    <Card>
      <CardHeader>
        <CardTitle>OS Power</CardTitle>
      </CardHeader>
      <CardContent className="grid grid-cols-1 gap-2">
        <div>
          <Label>Power Option:</Label>
          <Select defaultValue={powerOption} onValueChange={(e) => setPowerOption(e)}>
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="Action" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="ForceRestart">Power Cycle</SelectItem>
              <SelectItem value="On">On</SelectItem>
              <SelectItem value="ForceOff">Off</SelectItem>
              <SelectItem value="GracefulRestart">Graceful Restart</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div>
          <Label>Boot Option:</Label>
          <Select defaultValue={bootOption} onValueChange={(e) => setBootOption(e)}>
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="Boot Option" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="None">None</SelectItem>
              <SelectItem value="Pxe">PXE</SelectItem>
              <SelectItem value="BiosSetup">BIOS</SelectItem>
              <SelectItem value="Utilities">Utilities</SelectItem>
              <SelectItem value="Diags">Diagnostics</SelectItem>
              <SelectItem value="Usb">USB</SelectItem>
              <SelectItem value="Hdd">HDD</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </CardContent>
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
                Power Option: {powerOption}, Boot Option: {bootOption} <br />
                Nodes: {nodes}
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
                        body: {
                          power_option: powerOption,
                          boot_option: bootOption,
                        },
                      },
                      {
                        onSuccess: () => {
                          toast.success("Successfully sent power command node(s)");
                        },
                        onError: () =>
                          toast.error("Failed to send power command to node(s)", {
                            // description: e.message,
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
