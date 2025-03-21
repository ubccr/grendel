import { BootImage } from "@/openapi/requests";
import { Card, CardContent, CardHeader, CardTitle } from "../ui/card";
import { useQueryClient } from "@tanstack/react-query";
import { useForm } from "@tanstack/react-form";
import { toast } from "sonner";
import { Label } from "../ui/label";
import { Input } from "../ui/input";
import { Switch } from "../ui/switch";
import { Button } from "../ui/button";
import { ChevronDown, ChevronUp, LoaderCircle, Plus, X } from "lucide-react";
import { usePostV1Images } from "@/openapi/queries";
import { useEffect } from "react";

export default function ImageForm({
  data,
  reset,
}: {
  data?: BootImage;
  reset?: boolean;
}) {
  const storeImages = usePostV1Images();
  const queryClient = useQueryClient();

  const form = useForm({
    defaultValues: data,
    onSubmit: async ({ value }) => {
      if (value != undefined) {
        await storeImages.mutateAsync(
          { body: { boot_images: [value] } },
          {
            onSuccess: () => {
              toast.success("Successfully saved Image");
              queryClient.invalidateQueries();
            },
            onError: () => {
              toast.error("Error saving Image", {
                // description: e.message,
              });
            },
          }
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
      <Card>
        <CardHeader>
          <CardTitle className="flex justify-between">
            <span>Image:</span>
            <Button type="submit" variant={"outline"} size={"sm"}>
              {!storeImages.isPending && <span>Submit</span>}
              {storeImages.isPending && (
                <>
                  <LoaderCircle className="animate-spin" />
                  <span className="sr-only">Loading</span>
                </>
              )}
            </Button>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid sm:grid-cols-2 gap-6">
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
              name="kernel"
              children={(field) => (
                <div>
                  <Label>Kernel:</Label>
                  <Input
                    value={field.state.value ?? ""}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                  />
                </div>
              )}
            />
            <form.Field
              name="verify"
              children={(field) => (
                <div className="flex items-center space-x-2">
                  <Label>Verify:</Label>
                  <Switch
                    checked={field.state.value ?? false}
                    onBlur={field.handleBlur}
                    onCheckedChange={(e) => field.handleChange(e)}
                  />
                </div>
              )}
            />
            <form.Field
              name="butane"
              children={(field) => (
                <div>
                  <Label>Butane:</Label>
                  <Input
                    value={field.state.value ?? ""}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                  />
                </div>
              )}
            />
            <form.Field
              name="cmdline"
              children={(field) => (
                <div className="col-span-1 md:col-span-2">
                  <Label>Command Line:</Label>
                  <Input
                    value={field.state.value ?? ""}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                  />
                </div>
              )}
            />
            <form.Field
              name="user_data"
              children={(field) => (
                <div className="col-span-1 md:col-span-2">
                  <Label>User Data:</Label>
                  <Input
                    value={field.state.value ?? ""}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                  />
                </div>
              )}
            />
            <div>
              <form.Field name="provision_templates" mode="array">
                {(field) => (
                  <>
                    <div className="mb-2">
                      <Button
                        type="button"
                        size="sm"
                        variant="outline"
                        onClick={() =>
                          field.setValue({ ...field.state.value, "": "" })
                        }
                      >
                        <Plus />
                        <span>Add Template</span>
                      </Button>
                    </div>
                    <div className="grid grid-cols-1 gap-4">
                      {Object.keys(field.state.value ?? { "": "" }).map(
                        (key, i) => (
                          <Card key={i}>
                            <CardHeader>
                              <CardTitle className="flex justify-between">
                                <span>Template {i + 1}:</span>
                                <div className="flex gap-2">
                                  <Button
                                    type="button"
                                    size="sm"
                                    variant="destructive"
                                    onClick={() => {
                                      delete field.state.value?.[key];
                                      field.setValue(field.state.value);
                                    }}
                                  >
                                    <X />
                                    <span className="sr-only">
                                      Delete Template {i + 1}
                                    </span>
                                  </Button>
                                </div>
                              </CardTitle>
                            </CardHeader>
                            <CardContent>
                              <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
                                <div>
                                  <Input
                                    value={key}
                                    placeholder="name"
                                    onChange={(e) => {
                                      const preChange = field.state.value ?? {};
                                      const value = preChange[key];
                                      delete preChange[key];
                                      field.setValue({
                                        ...preChange,
                                        [e.target.value]: value,
                                      });
                                    }}
                                  />
                                </div>
                                <div>
                                  <Input
                                    placeholder="path"
                                    value={field.state.value?.[key] ?? ""}
                                    onChange={(e) => {
                                      const preChange = field.state.value ?? {};
                                      preChange[key] = e.target.value;
                                      field.setValue(preChange);
                                    }}
                                  />
                                </div>
                              </div>
                            </CardContent>
                          </Card>
                        )
                      )}
                    </div>
                  </>
                )}
              </form.Field>
            </div>
            <div>
              <form.Field name="initrd" mode="array">
                {(field) => (
                  <>
                    <div className="mb-2">
                      <Button
                        type="button"
                        size="sm"
                        variant="outline"
                        onClick={() => field.pushValue("")}
                      >
                        <Plus />
                        <span>Add Initrd</span>
                      </Button>
                    </div>
                    <div className="grid grid-cols-1 gap-4">
                      {field.state.value?.map((rd, i) => (
                        <Card key={i}>
                          <CardHeader>
                            <CardTitle className="flex justify-between">
                              <span>Initrd {i + 1}:</span>
                              <div className="flex gap-2">
                                <Button
                                  type="button"
                                  size="sm"
                                  variant="outline"
                                  onClick={() => field.moveValue(i, i - 1)}
                                  disabled={i == 0}
                                >
                                  <ChevronUp />
                                  <span className="sr-only">
                                    Move Initrd Up
                                  </span>
                                </Button>
                                <Button
                                  type="button"
                                  size="sm"
                                  variant="outline"
                                  onClick={() => field.moveValue(i, i + 1)}
                                  disabled={
                                    !!field.state.value &&
                                    i == field.state.value.length - 1
                                  }
                                >
                                  <ChevronDown />
                                  <span className="sr-only">
                                    Move Initrd Down
                                  </span>
                                </Button>
                                <Button
                                  type="button"
                                  size="sm"
                                  variant="destructive"
                                  onClick={() => field.removeValue(i)}
                                >
                                  <X />
                                  <span className="sr-only">
                                    Delete Initrd {i + 1}
                                  </span>
                                </Button>
                              </div>
                            </CardTitle>
                          </CardHeader>
                          <CardContent>
                            <div className="grid grid-cols-1 gap-2">
                              <div>
                                <Input
                                  placeholder="path"
                                  value={rd}
                                  onChange={(e) =>
                                    field.replaceValue(i, e.target.value)
                                  }
                                />
                              </div>
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
        </CardContent>
      </Card>
    </form>
  );
}
