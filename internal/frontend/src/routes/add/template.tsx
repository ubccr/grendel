import AuthRedirect from "@/auth";
import { useTheme } from "@/hooks/theme-provider";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Editor } from "@monaco-editor/react";
import { useForm } from "@tanstack/react-form";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/add/template")({
  component: RouteComponent,
  beforeLoad: AuthRedirect,
});

function RouteComponent() {
  const { theme } = useTheme();

  const form = useForm({
    defaultValues: {
      data: "",
      name: "",
      type: "",
      version: "",
      language: "",
    },
  });

  return (
    <div className="p-4">
      <Card>
        <CardHeader>
          <CardTitle>Add Template:</CardTitle>
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
              name="type"
              children={(field) => (
                <div>
                  <Label>Type:</Label>
                  <Select
                    value={field.state.value}
                    onValueChange={(e) => field.handleChange(e)}
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="Command Line">Command Line</SelectItem>
                      <SelectItem value="Butane">Butane</SelectItem>
                      <SelectItem value="Cloud Init">Cloud Init</SelectItem>
                      <SelectItem value="other">other</SelectItem>
                    </SelectContent>
                  </Select>
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
              name="language"
              children={(field) => (
                <div>
                  <Label>Language:</Label>
                  <Select
                    value={field.state.value}
                    onValueChange={(e) => field.handleChange(e)}
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="yaml">yaml</SelectItem>
                      <SelectItem value="json">json</SelectItem>
                      <SelectItem value="text">text</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              )}
            />
            <form.Field
              name="data"
              children={(field) => (
                <div className="col-span-2">
                  <Label>Template:</Label>
                  <Editor
                    value={field.state.value}
                    onChange={(e) => field.handleChange(e ?? "")}
                    height="50vh"
                    language={form.getFieldValue("language")}
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
