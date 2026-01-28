import { postV1BmcConfigureImportMutation } from "@/client/@tanstack/react-query.gen";
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
import { Input } from "@/components/ui/input";
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

export default function NodesImportConfigAction({ nodes }: { nodes: string }) {
  const [shutdownType, setShutdownType] = useState("");
  const [file, setFile] = useState("");

  const { mutate, isPending } = useMutation(postV1BmcConfigureImportMutation());
  return (
    <Card>
      <CardHeader>
        <CardTitle>Import System Config</CardTitle>
      </CardHeader>
      <CardContent className="grid grid-cols-1 gap-2">
        <div>
          <Label>Shutdown Type:</Label>
          <Select defaultValue={shutdownType} onValueChange={(e) => setShutdownType(e)}>
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="Action" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="NoReboot">No Reboot</SelectItem>
              <SelectItem value="Graceful">Graceful</SelectItem>
              <SelectItem value="Forced">Forced</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div>
          <Label>File:</Label>
          <Input
            placeholder="filename.json.tmpl"
            defaultValue={file}
            onChange={(e) => setFile(e.target.value)}
          />
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
                Shutdown Type: {shutdownType}, Filename: {file} <br />
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
                          file: file,
                          shutdown_type: shutdownType,
                        },
                      },
                      {
                        onSuccess: () => {
                          toast.success("Successfully submitted bmc configure job");
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
