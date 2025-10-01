import { Link } from "@tanstack/react-router";

import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ExternalLink, Plus, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Host } from "@/openapi/requests";

import { useForm } from "@tanstack/react-form";
import { TagsInput } from "@/components/tags-input";
import { Info, LoaderCircle } from "lucide-react";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { toast } from "sonner";
import { useQueryClient } from "@tanstack/react-query";
import { usePostV1Nodes } from "@/openapi/queries";
import { useEffect } from "react";

export default function NodeForm({
  data,
  reset,
}: {
  data?: Host;
  reset?: boolean;
}) {
  const storeHosts = usePostV1Nodes();
  const queryClient = useQueryClient();

  const form = useForm({
    defaultValues: data,
    onSubmit: async ({ value }) => {
      if (value != undefined) {
        await storeHosts.mutateAsync(
          { body: { node_list: [value] } },
          {
            onSuccess: (e) => {
              toast.success(e.data?.title, { description: e.data?.detail });
              queryClient.invalidateQueries();
            },
            onError: (e) => {
              console.log(e);

              toast.error(e.title, {
                description: e.detail,
              });
            },
          },
        );
      }
    },
  });

  useEffect(() => {
    form.reset();
  }, [reset]);

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        e.stopPropagation();
        form.handleSubmit();
      }}
    >
      <div className="flex justify-between">
      <span>Node:</span>
      <Button type="submit">
        {!storeHosts.isPending && <span>Submit</span>}
        {storeHosts.isPending && (
          <>
            <LoaderCircle className="animate-spin" />
            <span className="sr-only">Loading</span>
          </>
        )}
      </Button>
      </div>
      <div className="grid grid-cols-1 gap-6">
        <form.Field
          name="name"
          children={(field) => (
            <div>
              <Label>Name:</Label>
              <Input
                value={field.state.value ?? ""}
                onBlur={field.handleBlur}
                onChange={(e) => field.handleChange(e.target.value)}
              />
            </div>
          )}
        />
        <form.Field
          name="provision"
          children={(field) => (
            <div className="flex items-center space-x-2">
              <Label>Provision:</Label>
              <Switch
                checked={field.state.value ?? false}
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
                value={field.state.value ?? ""}
                onBlur={field.handleBlur}
                onChange={(e) => field.handleChange(e.target.value)}
              />
            </div>
          )}
        />
        <form.Field
          name="tags"
          children={(field) => (
            <div>
              <div className="mb-1 flex gap-2">
                <Label>Tags:</Label>
                <Popover>
                  <PopoverTrigger>
                    <Info className="size-3" />
                  </PopoverTrigger>
                  <PopoverContent>
                    <span className="text-md">
                      Both keys and key value pairs are accepted. Key value
                      pairs should be separated by "=".
                    </span>
                    <br />
                    <span className="text-sm font-light">
                      Example key only: "dell"
                    </span>
                    <br />
                    <span className="text-sm font-light">
                      Example key value pair: "brand=dell"
                    </span>
                    <br />
                    <span className="text-sm font-light">
                      Example key value pair with namespace:
                      "grendel:brand=dell"
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
        <div>
          <form.Field name="interfaces" mode="array">
            {(field) => (
              <>
                <div className="mb-2">
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={() => field.pushValue({})}
                  >
                    <Plus />
                    <span>Add Interface</span>
                  </Button>
                </div>
                <div className="grid grid-cols-1 gap-4">
                  {field.state.value?.map((iface, i) => (
                    <Card key={i}>
                      <CardHeader>
                        <CardTitle className="grid gap-3 sm:grid-cols-2">
                          <span>Interface {i + 1}:</span>
                          <div className="flex justify-center gap-2 sm:justify-end">
                            <Button
                              type="button"
                              size="icon"
                              variant="secondary"
                              disabled={!iface?.fqdn}
                              asChild
                            >
                              <Link
                                target="_blank"
                                to={"https://" + iface?.fqdn}
                              >
                                <ExternalLink />
                                <span className="sr-only">Go to FQDN</span>
                              </Link>
                            </Button>
                            {/* <Button
                                type="button"
                                variant="secondary"
                                onClick={() => field.moveValue(i, i - 1)}
                                disabled={i == 0}
                              >
                                <ChevronUp />
                                <span className="sr-only">
                                  Move Interface Up
                                </span>
                              </Button>
                              <Button
                                type="button"
                                variant="secondary"
                                onClick={() => field.moveValue(i, i + 1)}
                                disabled={
                                  !!field.state.value &&
                                  i == field.state.value.length - 1
                                }
                              >
                                <ChevronDown />
                                <span className="sr-only">
                                  Move Interface Down
                                </span>
                              </Button> */}
                            <Button
                              type="button"
                              size="icon"
                              variant="destructive"
                              onClick={() => field.removeValue(i)}
                            >
                              <X />
                              <span className="sr-only">
                                Delete Interface {i + 1}
                              </span>
                            </Button>
                          </div>
                        </CardTitle>
                      </CardHeader>
                      <CardContent>
                        <div className="grid gap-2 sm:grid-cols-2">
                          <form.Field name={`interfaces[${i}].fqdn`}>
                            {(subField) => (
                              <div>
                                <Label>FQDN:</Label>
                                <Input
                                  value={subField.state.value ?? ""}
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
                                  value={subField.state.value as string}
                                  onChange={(e) =>
                                    subField.handleChange(e.target.value)
                                  }
                                />
                              </div>
                            )}
                          </form.Field>
                          <form.Field name={`interfaces[${i}].ifname`}>
                            {(subField) => (
                              <div>
                                <Label>Name:</Label>
                                <Input
                                  value={subField.state.value ?? ""}
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
                                  value={subField.state.value ?? ""}
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
                                  value={subField.state.value ?? ""}
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
                                  type="number"
                                  value={subField.state.value ?? ""}
                                  onChange={(e) =>
                                    subField.handleChange(+e.target.value)
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
                                  checked={subField.state.value ?? false}
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
      </div>
    </form>
  );
}
