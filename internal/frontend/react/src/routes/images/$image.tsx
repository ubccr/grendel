import { useImageFindSuspense } from "@/openapi/queries/suspense";
import { useForm } from "@tanstack/react-form";
import { createFileRoute } from "@tanstack/react-router";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { Suspense } from "react";
import { Loading } from "@/components/loading";
import { ErrorBoundary } from "react-error-boundary";
import { Error } from "@/components/error";
import { toast } from "sonner";

export const Route = createFileRoute("/images/$image")({
    component: RouteComponent,
});

function Form() {
    const { image } = Route.useParams();
    const { data } = useImageFindSuspense({ path: { name: image } });

    const form = useForm({
        defaultValues: {
            name: data?.[0].name ?? "",
            kernel: data?.[0].kernel ?? "",
            initrd: data?.[0].initrd ?? [],
            liveimg: data?.[0].liveimg ?? "",
            cmdline: data?.[0].cmdline ?? "",
            verify: data?.[0].verify ?? false,
            // provision_template : data?.[0]. ?? "",
            // provision_templates : data?.[0].name ?? "",
            // user_data : data?.[0].user_data ?? "",
            // butane : data?.[0].butane ?? "",
        },
    });

    return (
        <>
            <div className="mb-4 flex justify-center">
                <div className="border-muted-foreground rounded-lg border px-4 py-2 text-3xl">{image}</div>
            </div>
            <div className="mx-auto">
                <Card>
                    <CardHeader>
                        <CardTitle>Image:</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <form className="grid grid-cols-2 gap-6">
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
                                name="kernel"
                                children={(field) => (
                                    <div>
                                        <Label>Kernel:</Label>
                                        <Input
                                            value={field.state.value}
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
                                            checked={field.state.value}
                                            onBlur={field.handleBlur}
                                            onCheckedChange={(e) => field.handleChange(e)}
                                        />
                                    </div>
                                )}
                            />
                            <form.Field
                                name="liveimg"
                                children={(field) => (
                                    <div>
                                        <Label>liveimg:</Label>
                                        <Input
                                            value={field.state.value}
                                            onBlur={field.handleBlur}
                                            onChange={(e) => field.handleChange(e.target.value)}
                                        />
                                    </div>
                                )}
                            />
                            <form.Field
                                name="cmdline"
                                children={(field) => (
                                    <div className="col-span-2">
                                        <Label>cmdline:</Label>
                                        <Input
                                            value={field.state.value}
                                            onBlur={field.handleBlur}
                                            onChange={(e) => field.handleChange(e.target.value)}
                                        />
                                    </div>
                                )}
                            />
                            <div className="col-span-2">
                                <form.Field name="initrd" mode="array">
                                    {(field) => (
                                        <div className="grid grid-cols-1 gap-4">
                                            {field.state.value.map((rd, i) => (
                                                <Card key={i}>
                                                    <CardHeader>
                                                        <CardTitle>Initrd {i + 1}:</CardTitle>
                                                    </CardHeader>
                                                    <CardContent>
                                                        <div className="grid grid-cols-2 gap-2">
                                                            <div>
                                                                <Input value={rd} />
                                                            </div>
                                                        </div>
                                                    </CardContent>
                                                </Card>
                                            ))}
                                        </div>
                                    )}
                                </form.Field>
                            </div>
                        </form>
                    </CardContent>
                </Card>
            </div>
        </>
    );
}

function RouteComponent() {
    return (
        <>
            <div className="p-4">
                <Suspense fallback={<Loading />}>
                    <ErrorBoundary
                        fallback={<Error />}
                        onError={(error) =>
                            toast.error("Error loading response", {
                                description: error.message,
                            })
                        }>
                        <Form />
                    </ErrorBoundary>
                </Suspense>
            </div>
        </>
    );
}

// const image = [
//     {
//         id: "2kDAhPVkrfIDtgTbPKamgt5zBjX",
//         name: "frosty",
//         kernel: "/var/lib/grendel/images/frosty/noble/frosty-uki-24.11.1-1-ubuntu-noble-x86-64.efi",
//         initrd: [],
//         liveimg: "",
//         cmdline:
//             "console=tty0 console=ttyS0,115200n8 root=LABEL=ROOT selinux=0 systemd.hostname={{$.nic.FQDN}} ifname=bootnet:{{$.nic.MAC}} ip={{$.nic.AddrString}}::{{$.nic.Gateway}}:{{$.nic.NetmaskString}}::bootnet:none:{{ or $.nic.MTU 9000 }} ignition.firstboot=1 ignition.platform.id=metal ignition.config.url={{ $.endpoints.IgnitionURL }} ",
//         verify: false,
//         provision_template: "",
//         provision_templates: {
//             "idrac-config.json": "idrac-config.json.tmpl",
//         },
//         user_data: "",
//         butane: "frosty-ignition.tmpl",
//     },
// ];
