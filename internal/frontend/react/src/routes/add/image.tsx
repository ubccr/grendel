import { useTheme } from "@/components/theme-provider";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Editor } from "@monaco-editor/react";
import { useForm } from "@tanstack/react-form";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/add/image")({
    component: RouteComponent,
});

function RouteComponent() {
    const { theme } = useTheme();

    const form = useForm({
        defaultValues: {
            commandline: "",
        },
    });
    return (
        <div className="p-4">
            <Card>
                <CardHeader>
                    <CardTitle>Add Image:</CardTitle>
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
                            name="arch"
                            children={(field) => (
                                <div>
                                    <Label>Architecture:</Label>
                                    <Input
                                        value={field.state.value}
                                        onBlur={field.handleBlur}
                                        onChange={(e) => field.handleChange(e.target.value)}
                                    />
                                </div>
                            )}
                        />
                        <form.Field
                            name="version"
                            children={(field) => (
                                <div>
                                    <Label>Version:</Label>
                                    <Input
                                        value={field.state.value}
                                        onBlur={field.handleBlur}
                                        onChange={(e) => field.handleChange(e.target.value)}
                                    />
                                </div>
                            )}
                        />
                        <form.Field
                            name="initrd"
                            children={(field) => (
                                <div>
                                    <Label>Initrd:</Label>
                                    <Input
                                        type="file"
                                        multiple
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
                                        type="file"
                                        value={field.state.value}
                                        onBlur={field.handleBlur}
                                        onChange={(e) => field.handleChange(e.target.value)}
                                    />
                                </div>
                            )}
                        />
                        <form.Field
                            name="commandline"
                            children={(field) => (
                                <div>
                                    <Label>Command Line Template:</Label>
                                    <Editor
                                        value={field.state.value}
                                        onChange={(e) => field.handleChange(e)}
                                        height="60px"
                                        language="text"
                                        theme={theme == "dark" ? "vs-dark" : "light"}
                                    />
                                </div>
                            )}
                        />
                    </form>
                </CardContent>
            </Card>
        </div>
    );
}
