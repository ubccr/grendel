import { Switch } from "@/components/ui/switch";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { TagsInput } from "@/components/tags-input";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useQueryClient } from "@tanstack/react-query";
import { Button } from "../ui/button";
import { toast } from "sonner";
import { useState } from "react";
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
import { Copy, Info, LoaderCircle } from "lucide-react";
import { Label } from "../ui/label";
import RedfishJobList from "./job-table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "../ui/tooltip";
import {
  useDeleteV1BmcJobsJids,
  useDeleteV1Nodes,
  useGetV1NodesFind,
  usePatchV1NodesImage,
  usePatchV1NodesProvision,
  usePatchV1NodesTagsAction,
  usePostV1BmcConfigureImport,
  usePostV1BmcPowerBmc,
  usePostV1BmcPowerOs,
} from "@/openapi/queries";
import { Input } from "../ui/input";

export default function NodeActions({
  nodes,
  length,
}: {
  nodes: string;
  length: number;
}) {
  const queryClient = useQueryClient();
  const mutation_delete = useDeleteV1Nodes();
  const mutation_provision = usePatchV1NodesProvision();
  const mutation_tag = usePatchV1NodesTagsAction();
  const mutation_image = usePatchV1NodesImage();
  const mutation_power_bmc = usePostV1BmcPowerBmc();
  const mutation_configure_import = usePostV1BmcConfigureImport();
  const mutation_job_clear = useDeleteV1BmcJobsJids();
  const hosts_query = useGetV1NodesFind(
    { query: { nodeset: nodes } },
    undefined,
    {
      enabled: false,
    }
  );

  const mutation_power = usePostV1BmcPowerOs();

  const [tags, setTags] = useState<string[]>([]);
  const [bootImage, setBootImage] = useState<string>("");
  const [tagsAction, setTagsAction] = useState("");
  const [osPowerAction, setOSPowerAction] = useState("");
  const [importSystemConfigShutdownType, setImportSystemConfigShutdownType] =
    useState("");
  const [importSystemConfigFile, setImportSystemConfigFile] = useState("");
  const [osBootAction, setOSBootAction] = useState("None");
  const [provision, setProvision] = useState(false);

  return (
    <div className="mt-4 grid sm:grid-cols-2 gap-4">
      <Card>
        <CardHeader>
          <CardTitle>Delete</CardTitle>
        </CardHeader>
        <CardFooter>
          <Dialog>
            <DialogTrigger asChild>
              <Button size="sm" variant="destructive">
                Delete
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Are you sure?</DialogTitle>
                <DialogDescription>
                  WARNING: Selected nodes: ({nodes}) will be permanently removed
                  from Grendel!
                </DialogDescription>
              </DialogHeader>
              <DialogFooter>
                <DialogClose asChild>
                  <Button
                    variant="destructive"
                    size="sm"
                    onClick={() =>
                      mutation_delete.mutate(
                        { query: { nodeset: nodes } },
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
                        }
                      )
                    }
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
      <Card>
        <CardHeader>
          <CardTitle>Export JSON</CardTitle>
        </CardHeader>
        <CardFooter>
          <Dialog>
            <DialogTrigger asChild>
              <Button
                variant="outline"
                size="sm"
                onClick={() => hosts_query.refetch()}
              >
                Submit
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Export JSON: {nodes}</DialogTitle>
              </DialogHeader>
              <div className="max-h-[calc(70dvh)] overflow-scroll">
                <div className="text-muted-foreground">
                  {hosts_query.isLoading ? (
                    <LoaderCircle className="animate-spin mx-auto" />
                  ) : (
                    <pre>{JSON.stringify(hosts_query.data, null, 4)}</pre>
                  )}
                </div>
              </div>
              <DialogFooter>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    navigator.clipboard.writeText(
                      JSON.stringify(hosts_query.data, null, 4)
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
      <Card>
        <CardHeader>
          <CardTitle>Provision</CardTitle>
        </CardHeader>
        <CardContent>
          <Switch onCheckedChange={(e) => setProvision(e)} />
        </CardContent>
        <CardFooter>
          <Button
            variant="outline"
            size="sm"
            onClick={() =>
              mutation_provision.mutate(
                { query: { nodeset: nodes }, body: { provision: provision } },
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
                }
              )
            }
          >
            Submit
          </Button>
        </CardFooter>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>Boot Image</CardTitle>
        </CardHeader>
        <CardContent className="grid grid-cols-1 gap-2">
          <Input
            value={bootImage}
            onChange={(e) => setBootImage(e.target.value)}
          />
        </CardContent>
        <CardFooter>
          <Button
            variant="outline"
            size="sm"
            onClick={() =>
              mutation_image.mutate(
                {
                  query: { nodeset: nodes },
                  body: { image: bootImage },
                },
                {
                  onSuccess: (e) => {
                    queryClient.invalidateQueries();
                    toast.success(e.data?.title, {
                      description: e.data?.detail,
                    });
                  },
                  onError: (e) =>
                    toast.error(e.title, {
                      description: e.detail,
                    }),
                }
              )
            }
          >
            Submit
          </Button>
        </CardFooter>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>Tags</CardTitle>
        </CardHeader>
        <CardContent className="grid grid-cols-1 gap-2">
          <TagsInput value={tags} onValueChange={setTags} placeholder="Tags" />
          <Select onValueChange={(e) => setTagsAction(e)}>
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="Action" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="add">Add</SelectItem>
              <SelectItem value="remove">Remove</SelectItem>
            </SelectContent>
          </Select>
        </CardContent>
        <CardFooter>
          <Button
            variant="outline"
            size="sm"
            onClick={() =>
              mutation_tag.mutate(
                {
                  path: { action: tagsAction },
                  query: { nodeset: nodes },
                  body: { tags: tags.join(",") },
                },
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
                }
              )
            }
          >
            Submit
          </Button>
        </CardFooter>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>OS Power</CardTitle>
        </CardHeader>
        <CardContent className="grid grid-cols-1 gap-2">
          <div>
            <Label>Power Option:</Label>
            <Select
              defaultValue={osPowerAction}
              onValueChange={(e) => setOSPowerAction(e)}
            >
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Action" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="ForceRestart">Power Cycle</SelectItem>
                <SelectItem value="On">On</SelectItem>
                <SelectItem value="ForceOff">Off</SelectItem>
                <SelectItem value="GracefulRestart">
                  Graceful Restart
                </SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div>
            <Label>Boot Option:</Label>
            <Select
              defaultValue={osBootAction}
              onValueChange={(e) => setOSBootAction(e)}
            >
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
              <Button variant="outline" size="sm">
                {mutation_power.isPending ? (
                  <LoaderCircle className="animate-spin" />
                ) : (
                  <span>Submit</span>
                )}
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Are you sure?</DialogTitle>
                <DialogDescription>
                  Power Option: {osPowerAction}, Boot Option: {osBootAction}{" "}
                  <br />
                  Nodes: {nodes}
                </DialogDescription>
              </DialogHeader>
              <DialogFooter>
                <DialogClose asChild>
                  <Button
                    variant="destructive"
                    size="sm"
                    onClick={() =>
                      mutation_power.mutate(
                        {
                          query: { nodeset: nodes },
                          body: {
                            power_option: osPowerAction,
                            boot_option: osBootAction,
                          },
                        },
                        {
                          onSuccess: () => {
                            toast.success(
                              "Successfully sent power command node(s)"
                            );
                          },
                          onError: () =>
                            toast.error(
                              "Failed to send power command to node(s)",
                              {
                                // description: e.message,
                              }
                            ),
                        }
                      )
                    }
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
      <Card>
        <CardHeader>
          <CardTitle>Import System Config</CardTitle>
        </CardHeader>
        <CardContent className="grid grid-cols-1 gap-2">
          <div>
            <Label>Shutdown Type:</Label>
            <Select
              defaultValue={importSystemConfigShutdownType}
              onValueChange={(e) => setImportSystemConfigShutdownType(e)}
            >
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
              defaultValue={importSystemConfigFile}
              onChange={(e) => setImportSystemConfigFile(e.target.value)}
            />
          </div>
        </CardContent>
        <CardFooter>
          <Dialog>
            <DialogTrigger asChild>
              <Button variant="outline" size="sm">
                {mutation_configure_import.isPending ? (
                  <LoaderCircle className="animate-spin" />
                ) : (
                  <span>Submit</span>
                )}
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Are you sure?</DialogTitle>
                <DialogDescription>
                  Shutdown Type: {importSystemConfigShutdownType}, Filename:{" "}
                  {importSystemConfigFile} <br />
                  Nodes: {nodes}
                </DialogDescription>
              </DialogHeader>
              <DialogFooter>
                <DialogClose asChild>
                  <Button
                    variant="destructive"
                    size="sm"
                    onClick={() =>
                      mutation_configure_import.mutate(
                        {
                          query: { nodeset: nodes },
                          body: {
                            file: importSystemConfigFile,
                            shutdown_type: importSystemConfigShutdownType,
                          },
                        },
                        {
                          onSuccess: () => {
                            toast.success(
                              "Successfully submitted bmc configure job"
                            );
                          },
                          onError: (e) =>
                            toast.error(e.title, {
                              description: e.detail,
                            }),
                        }
                      )
                    }
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
      <Card>
        <CardHeader>
          <CardTitle>BMC PowerCycle</CardTitle>
        </CardHeader>
        <CardContent className="grid grid-cols-1 gap-2"></CardContent>
        <CardFooter>
          <Dialog>
            <DialogTrigger asChild>
              <Button variant="outline" size="sm">
                {mutation_power_bmc.isPending ? (
                  <LoaderCircle className="animate-spin" />
                ) : (
                  <span>Submit</span>
                )}
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Are you sure?</DialogTitle>
                <DialogDescription>
                  bmc(s) will be unavailable while they reboot
                  <br />
                  Nodes: {nodes}
                </DialogDescription>
              </DialogHeader>
              <DialogFooter>
                <DialogClose asChild>
                  <Button
                    variant="destructive"
                    size="sm"
                    onClick={() =>
                      mutation_power_bmc.mutate(
                        {
                          query: { nodeset: nodes },
                        },
                        {
                          onSuccess: () => {
                            toast.success("Successfully rebooted bmc(s)");
                          },
                          onError: (e) =>
                            toast.error(e.title, {
                              description: e.detail,
                            }),
                        }
                      )
                    }
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
      <Card>
        <CardHeader>
          <CardTitle>View Jobs</CardTitle>
        </CardHeader>
        <CardFooter>
          <Dialog>
            <DialogTrigger asChild>
              <Button variant="outline" size="sm" disabled={length !== 1}>
                Submit
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-7xl">
              <DialogHeader>
                <DialogTitle>Job List: {nodes}</DialogTitle>
              </DialogHeader>
              <div className="max-h-[calc(90dvh)] overflow-scroll">
                <RedfishJobList nodes={nodes} />
              </div>
            </DialogContent>
          </Dialog>
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button variant="outline" size="sm">
                  <Info />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                Currently only querying one node at a time is supported
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </CardFooter>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>Clear Jobs</CardTitle>
        </CardHeader>
        <CardFooter>
          <Dialog>
            <DialogTrigger asChild>
              <Button size="sm" variant="outline">
                {mutation_job_clear.isPending ? (
                  <LoaderCircle className="animate-spin" />
                ) : (
                  <span>Submit</span>
                )}
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Are you sure?</DialogTitle>
                <DialogDescription>
                  Selected nodes: ({nodes}) will have all BMC jobs cleared
                </DialogDescription>
              </DialogHeader>
              <DialogFooter>
                <DialogClose asChild>
                  <Button
                    variant="destructive"
                    size="sm"
                    onClick={() =>
                      mutation_job_clear.mutate(
                        {
                          query: { nodeset: nodes },
                          path: { jids: "JID_CLEARALL" },
                        },
                        {
                          onSuccess: () => {
                            toast.success(
                              "Successfully cleared all jobs on node(s)"
                            );
                            queryClient.invalidateQueries();
                          },
                          onError: (e) =>
                            toast.error(e.title, {
                              description: e.detail,
                            }),
                        }
                      )
                    }
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
