import { Link } from "@tanstack/react-router";

import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ChevronDown, ChevronUp, ExternalLink, Plus, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Host } from "@/openapi/requests";

import { useForm } from "@tanstack/react-form";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { TagsInput } from "@/components/tags-input";
import { Info, LoaderCircle } from "lucide-react";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { useStoreHosts } from "@/openapi/queries";
import { toast } from "sonner";
import { useQueryClient } from "@tanstack/react-query";

type Props = {
    data?: Host;
};

export default function NodeForm({ data }: Props) {
    const storeHosts = useStoreHosts();
    const queryClient = useQueryClient();

    const form = useForm({
        defaultValues: data,
        onSubmit: async ({ value }) => {
            if (value != undefined) {
                await storeHosts.mutateAsync(
                    { body: [value] },
                    {
                        onSuccess: () => {
                            toast.success("Successfully saved Node");
                            queryClient.invalidateQueries();
                        },
                        onError: (e) => {
                            toast.error("Error saving Node", {
                                description: e.message,
                            });
                        },
                    },
                );
            }
        },
    });

    return (
        <form
            onSubmit={(e) => {
                e.preventDefault();
                e.stopPropagation();
                form.handleSubmit();
            }}>
            <Card>
                <CardHeader>
                    <CardTitle className="flex justify-between">
                        <span>Node:</span>
                        <Button type="submit" variant={"outline"} size={"sm"}>
                            {!storeHosts.isPending && <span>Submit</span>}
                            {storeHosts.isPending && (
                                <>
                                    <LoaderCircle className="animate-spin" />
                                    <span className="sr-only">Loading</span>
                                </>
                            )}
                        </Button>
                    </CardTitle>
                </CardHeader>
                <CardContent className="grid grid-cols-2 gap-6">
                    <form.Field
                        name="name"
                        children={(field) => (
                            <div>
                                <Label>Name:</Label>
                                <Input
                                    value={field.state.value}
                                    onBlur={field.handleBlur}
                                    onChange={(e) => field.handleChange(e.target.value)}
                                />
                            </div>
                        )}
                    />
                    <form.Field
                        name="node_group"
                        children={(field) => (
                            <div>
                                <Label>Group:</Label>
                                <Select value={field.state.value} onValueChange={(e) => field.handleChange(e)}>
                                    <SelectTrigger>
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="hpc">HPC</SelectItem>
                                        <SelectItem value="cloud">Cloud</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>
                        )}
                    />
                    <form.Field
                        name="provision"
                        children={(field) => (
                            <div className="flex items-center space-x-2">
                                <Label>Provision:</Label>
                                <Switch
                                    checked={field.state.value}
                                    onBlur={field.handleBlur}
                                    onCheckedChange={(e) => field.handleChange(e)}
                                />
                            </div>
                        )}
                    />
                    <form.Field
                        name="boot_image"
                        children={(field) => (
                            <div>
                                <Label>Image:</Label>
                                <Input
                                    value={field.state.value}
                                    onBlur={field.handleBlur}
                                    onChange={(e) => field.handleChange(e.target.value)}
                                />
                            </div>
                        )}
                    />
                    <form.Field
                        name="tags"
                        children={(field) => (
                            <div className="col-span-2">
                                <div className="mb-1 flex gap-2">
                                    <Label>Tags:</Label>
                                    <Popover>
                                        <PopoverTrigger>
                                            <Info className="size-3" />
                                        </PopoverTrigger>
                                        <PopoverContent>
                                            <span className="text-md">
                                                Both keys and key value pairs are accepted. Key value pairs should be
                                                separated by ":"
                                            </span>
                                            <br />
                                            <span className="text-sm font-light">
                                                Example key only: "dell" <br />
                                            </span>
                                            <span className="text-sm font-light">
                                                Example key value pair: "brand:dell"
                                            </span>
                                        </PopoverContent>
                                    </Popover>
                                </div>
                                <TagsInput
                                    className="px-3 py-2"
                                    value={field.state.value ?? []}
                                    onValueChange={(e) => field.handleChange(e)}
                                />
                            </div>
                        )}
                    />
                    <div className="col-span-2">
                        <form.Field name="interfaces" mode="array">
                            {(field) => (
                                <>
                                    <div className="mb-2">
                                        <Button
                                            type="button"
                                            size="sm"
                                            variant="outline"
                                            onClick={() => field.pushValue({})}>
                                            <Plus />
                                            <span>Add Interface</span>
                                        </Button>
                                    </div>
                                    <div className="grid grid-cols-1 gap-4">
                                        {field.state.value?.map((iface, i) => (
                                            <Card key={i}>
                                                <CardHeader>
                                                    <CardTitle className="flex justify-between">
                                                        <span>Interface {i + 1}:</span>
                                                        <div className="flex gap-2">
                                                            <Button
                                                                type="button"
                                                                size="sm"
                                                                variant="outline"
                                                                disabled={!iface.fqdn}
                                                                asChild>
                                                                <Link target="_blank" to={"https://" + iface.fqdn}>
                                                                    <ExternalLink />
                                                                    <span className="sr-only">Go to FQDN</span>
                                                                </Link>
                                                            </Button>
                                                            <Button
                                                                type="button"
                                                                size="sm"
                                                                variant="outline"
                                                                onClick={() => field.moveValue(i, i - 1)}
                                                                disabled={i == 0}>
                                                                <ChevronUp />
                                                                <span className="sr-only">Move Interface Up</span>
                                                            </Button>
                                                            <Button
                                                                type="button"
                                                                size="sm"
                                                                variant="outline"
                                                                onClick={() => field.moveValue(i, i + 1)}
                                                                disabled={
                                                                    field.state.value &&
                                                                    i == field.state.value.length - 1
                                                                }>
                                                                <ChevronDown />
                                                                <span className="sr-only">Move Interface Down</span>
                                                            </Button>
                                                            <Button
                                                                type="button"
                                                                size="sm"
                                                                variant="destructive"
                                                                onClick={() => field.removeValue(i)}>
                                                                <X />
                                                                <span className="sr-only">
                                                                    Delete Interface {i + 1}
                                                                </span>
                                                            </Button>
                                                        </div>
                                                    </CardTitle>
                                                </CardHeader>
                                                <CardContent>
                                                    <div className="grid grid-cols-2 gap-2">
                                                        <form.Field name={`interfaces[${i}].fqdn`}>
                                                            {(subField) => (
                                                                <div>
                                                                    <Label>FQDN:</Label>
                                                                    <Input
                                                                        value={subField.state.value}
                                                                        onChange={(e) =>
                                                                            subField.handleChange(e.target.value)
                                                                        }
                                                                    />
                                                                </div>
                                                            )}
                                                        </form.Field>
                                                        <form.Field name={`interfaces[${i}].ip`}>
                                                            {(subField) => (
                                                                <div>
                                                                    <Label>IP:</Label>
                                                                    <Input
                                                                        value={subField.state.value}
                                                                        onChange={(e) =>
                                                                            subField.handleChange(e.target.value)
                                                                        }
                                                                    />
                                                                </div>
                                                            )}
                                                        </form.Field>
                                                        <form.Field name={`interfaces[${i}].name`}>
                                                            {(subField) => (
                                                                <div>
                                                                    <Label>Name:</Label>
                                                                    <Input
                                                                        value={subField.state.value}
                                                                        onChange={(e) =>
                                                                            subField.handleChange(e.target.value)
                                                                        }
                                                                    />
                                                                </div>
                                                            )}
                                                        </form.Field>
                                                        <form.Field name={`interfaces[${i}].mac`}>
                                                            {(subField) => (
                                                                <div>
                                                                    <Label>MAC:</Label>
                                                                    <Input
                                                                        value={subField.state.value}
                                                                        onChange={(e) =>
                                                                            subField.handleChange(e.target.value)
                                                                        }
                                                                    />
                                                                </div>
                                                            )}
                                                        </form.Field>
                                                        <form.Field name={`interfaces[${i}].vlan`}>
                                                            {(subField) => (
                                                                <div>
                                                                    <Label>VLAN:</Label>
                                                                    <Input
                                                                        value={subField.state.value}
                                                                        onChange={(e) =>
                                                                            subField.handleChange(e.target.value)
                                                                        }
                                                                    />
                                                                </div>
                                                            )}
                                                        </form.Field>
                                                        <form.Field name={`interfaces[${i}].mtu`}>
                                                            {(subField) => (
                                                                <div>
                                                                    <Label>MTU:</Label>
                                                                    <Input
                                                                        value={subField.state.value}
                                                                        onChange={(e) =>
                                                                            subField.handleChange(e.target.value)
                                                                        }
                                                                    />
                                                                </div>
                                                            )}
                                                        </form.Field>
                                                        <form.Field name={`interfaces[${i}].bmc`}>
                                                            {(subField) => (
                                                                <div className="flex items-center space-x-2">
                                                                    <Label>BMC:</Label>
                                                                    <Switch
                                                                        checked={subField.state.value}
                                                                        onBlur={subField.handleBlur}
                                                                        onCheckedChange={(e) =>
                                                                            subField.handleChange(e)
                                                                        }
                                                                    />
                                                                </div>
                                                            )}
                                                        </form.Field>
                                                    </div>
                                                </CardContent>
                                            </Card>
                                        ))}
                                    </div>
                                </>
                            )}
                        </form.Field>
                    </div>
                </CardContent>
            </Card>
        </form>
    );
}