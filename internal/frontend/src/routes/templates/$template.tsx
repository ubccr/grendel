import { createFileRoute } from "@tanstack/react-router";
import Editor from "@monaco-editor/react";
import { themeToMonaco, useTheme } from "@/hooks/theme-provider";

export const Route = createFileRoute("/templates/$template")({
  component: RouteComponent,
});

function RouteComponent() {
  const { theme } = useTheme();

  return (
    <div className="flex justify-center">
      <Editor
        height="90vh"
        language="yaml"
        defaultValue={""}
        theme={themeToMonaco(theme)}
      />
    </div>
  );
}
