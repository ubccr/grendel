import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle, SheetTrigger } from "@/components/ui/sheet";
import { Switch } from "@/components/ui/switch";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { TagsInput } from "@/components/tags-input";
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { useQueryClient } from "@tanstack/react-query";
import { useHostDelete, useHostProvision, useHostTag, useHostUnprovision, useHostUntag } from "@/openapi/queries";
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
import ImageSelect from "../image-select";
import { Input } from "../ui/input";
import { Hammer } from "lucide-react";
import ExportJSON from "../export-json";

type Props = {
    checked: string;
    length: number;
};

export default function Actions({ checked, length }: Props) {
    const queryClient = useQueryClient();
    const mutation_delete = useHostDelete();
    const mutation_provision = useHostProvision();
    const mutation_unprovision = useHostUnprovision();
    const mutation_tag = useHostTag();
    const mutation_untag = useHostUntag();

    const [tags, setTags] = useState<string[]>([]);
    const [tagsAction, setTagsAction] = useState("");
    const [provision, setProvision] = useState(false);
    return (
        <Sheet>
            <SheetTrigger asChild>
                <Button variant="outline" size="sm" className="relative">
                    <Hammer />
                    <span className="sr-only sm:not-sr-only">Actions</span>
                    {length > 0 && (
                        <span className="size-4 absolute -right-1 -top-1 flex">
                            <span className="size-full relative inline-flex justify-center rounded-full bg-sky-500 text-xs text-black">
                                {Math.abs(length).toString().length > 2 ? "-" : length}
                            </span>
                        </span>
                    )}
                </Button>
            </SheetTrigger>
            <SheetContent className="max-w-1/2 overflow-y-scroll">
                <SheetHeader>
                    <SheetTitle>Actions:</SheetTitle>
                    {/* TODO: add copy button */}
                    <SheetDescription className="max-h-36 overflow-y-scroll rounded-md border p-2">
                        {length} Selected node(s): <br />
                        {checked}
                    </SheetDescription>
                </SheetHeader>
                <div className="mt-4 grid grid-cols-1 gap-4">
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
                                            WARNING: Selected nodes: ({checked}) will be permanently removed from
                                            Grendel!
                                        </DialogDescription>
                                    </DialogHeader>
                                    <DialogFooter>
                                        <DialogClose asChild>
                                            <Button
                                                variant="destructive"
                                                size="sm"
                                                onClick={() =>
                                                    mutation_delete.mutate(
                                                        { path: { nodeSet: checked } },
                                                        {
                                                            onSuccess: () => {
                                                                toast.success("Successfully deleted node(s)");
                                                                queryClient.invalidateQueries();
                                                            },
                                                            onError: (e) =>
                                                                toast.error("Failed to delete node(s)", {
                                                                    description: e.message,
                                                                }),
                                                        },
                                                    )
                                                }>
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
                                    provision
                                        ? mutation_provision.mutate(
                                              { path: { nodeSet: checked } },
                                              {
                                                  onSuccess: () => {
                                                      toast.success("Successfully set node(s) to provision");
                                                      queryClient.invalidateQueries();
                                                  },
                                                  onError: (e) =>
                                                      toast.error("Failed to set node(s) to provision", {
                                                          description: e.message,
                                                      }),
                                              },
                                          )
                                        : mutation_unprovision.mutate(
                                              { path: { nodeSet: checked } },
                                              {
                                                  onSuccess: () => {
                                                      toast.success("Successfully set node(s) to unprovision");
                                                      queryClient.invalidateQueries();
                                                  },
                                                  onError: (e) =>
                                                      toast.error("Failed to set node(s) to unprovision", {
                                                          description: e.message,
                                                      }),
                                              },
                                          )
                                }>
                                Submit
                            </Button>
                        </CardFooter>
                    </Card>
                    <Card>
                        <CardHeader>
                            <CardTitle>Tags</CardTitle>
                        </CardHeader>
                        <CardContent>
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
                                    tagsAction === "add"
                                        ? mutation_tag.mutate(
                                              { path: { nodeSet: checked }, query: { tags: tags.join(",") } },
                                              {
                                                  onSuccess: () => {
                                                      queryClient.invalidateQueries();
                                                      toast.success("Successfully tagged node(s)");
                                                  },
                                                  onError: (e) =>
                                                      toast.error("Failed to tag node(s)", {
                                                          description: e.message,
                                                      }),
                                              },
                                          )
                                        : mutation_untag.mutate(
                                              { path: { nodeSet: checked }, query: { tags: tags.join(",") } },
                                              {
                                                  onSuccess: () => {
                                                      queryClient.invalidateQueries();
                                                      toast.success("Successfully untagged node(s)");
                                                  },
                                                  onError: (e) =>
                                                      toast.error("Failed to untag node(s)", {
                                                          description: e.message,
                                                      }),
                                              },
                                          )
                                }>
                                Submit
                            </Button>
                        </CardFooter>
                    </Card>
                    <Card>
                        <CardHeader>
                            <CardTitle>Auto Configure</CardTitle>
                        </CardHeader>
                        <CardFooter>
                            <Button variant="outline" size="sm">
                                Auto Configure
                            </Button>
                        </CardFooter>
                    </Card>
                    <Card>
                        <CardHeader>
                            <CardTitle>Boot Image</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <ImageSelect />
                        </CardContent>
                        <CardFooter>
                            <Button variant="outline" size="sm">
                                Submit
                            </Button>
                        </CardFooter>
                    </Card>
                    <Card>
                        <CardHeader>
                            <CardTitle>OS Power</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <Select>
                                <SelectTrigger className="w-[180px]">
                                    <SelectValue placeholder="Command" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="powerCycle">Power Cycle</SelectItem>
                                    <SelectItem value="powerON">Power On</SelectItem>
                                    <SelectItem value="powerOff">Power Off</SelectItem>
                                </SelectContent>
                            </Select>
                            <Select>
                                <SelectTrigger className="w-[180px]">
                                    <SelectValue placeholder="Override" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="none">None</SelectItem>
                                    <SelectItem value="pxe">PXE</SelectItem>
                                    <SelectItem value="bios">BIOS</SelectItem>
                                </SelectContent>
                            </Select>
                        </CardContent>
                        <CardFooter>
                            <Button variant="outline" size="sm">
                                Submit
                            </Button>
                        </CardFooter>
                    </Card>
                    <Card>
                        <CardHeader>
                            <CardTitle>BMC Powercycle</CardTitle>
                        </CardHeader>
                        <CardFooter>
                            <Button variant="outline" size="sm">
                                Submit
                            </Button>
                        </CardFooter>
                    </Card>
                    <Card>
                        <CardHeader>
                            <CardTitle>Import System Config</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <Select>
                                <SelectTrigger className="w-[180px]">
                                    <SelectValue placeholder="Template" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="idrac">idrac-config.tmpl</SelectItem>
                                </SelectContent>
                            </Select>
                        </CardContent>
                        <CardFooter>
                            <Button variant="outline" size="sm">
                                Submit
                            </Button>
                        </CardFooter>
                    </Card>
                    <Card>
                        <CardHeader>
                            <CardTitle>Export JSON</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <Input placeholder="filename" />
                        </CardContent>
                        <CardFooter>
                            <Dialog>
                                <DialogTrigger asChild>
                                    <Button variant="outline" size="sm">
                                        Submit
                                    </Button>
                                </DialogTrigger>
                                <DialogContent className="h-3/5">
                                    <DialogHeader>
                                        <DialogTitle>Export JSON: {checked}</DialogTitle>
                                    </DialogHeader>
                                    <div className="overflow-y-scroll">
                                        <ExportJSON nodes={checked} />
                                    </div>
                                </DialogContent>
                            </Dialog>
                        </CardFooter>
                    </Card>
                    <Card>
                        <CardHeader>
                            <CardTitle>Export CSV</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <Input placeholder="template" />
                            <Input placeholder="filename" />
                        </CardContent>
                        <CardFooter>
                            <Button variant="outline" size="sm">
                                Submit
                            </Button>
                        </CardFooter>
                    </Card>
                </div>
            </SheetContent>
        </Sheet>
    );
}
