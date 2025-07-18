import { createFileRoute } from "@tanstack/react-router";
import NodeForm from "@/components/nodes/form";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Editor } from "@monaco-editor/react";
import { useTheme } from "@/hooks/theme-provider";
import { Button } from "@/components/ui/button";
import { useState } from "react";
import { toast } from "sonner";
import { useQueryClient } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import { usePostV1Nodes } from "@/openapi/queries";
import AuthRedirect from "@/auth";

export const Route = createFileRoute("/add/node")({
  component: RouteComponent,
  beforeLoad: AuthRedirect,
});

function RouteComponent() {
  return (
    <div className="p-4 mx-auto">
      <Tabs defaultValue="form" className="w-full">
        <div className="text-center">
          <TabsList>
            <TabsTrigger value="form">Form</TabsTrigger>
            <TabsTrigger value="json">JSON</TabsTrigger>
          </TabsList>
        </div>
        <TabsContent value="form">
          <NodeForm />
        </TabsContent>
        <TabsContent value="json">
          <NodeImportJSON />
        </TabsContent>
      </Tabs>
    </div>
  );
}

const defaultJson = {
  node_list: [],
};

function NodeImportJSON() {
  const { theme } = useTheme();
  const storeHosts = usePostV1Nodes();
  const [text, setText] = useState("");
  const queryClient = useQueryClient();

  return (
    <div>
      <Editor
        height="80vh"
        language="json"
        value={text}
        defaultValue={JSON.stringify(defaultJson, null, 4)}
        onChange={(e) => setText(e ?? "")}
        theme={theme == "dark" ? "vs-dark" : "light"}
      />
      <div className="flex justify-end mt-2">
        <Button
          variant="outline"
          size="sm"
          onClick={() =>
            storeHosts.mutate(
              { body: JSON.parse(text) },
              {
                onSuccess: (e) => {
                  toast.success(e.data?.title, {
                    description: e.data?.detail,
                  });
                  queryClient.invalidateQueries();
                },
                onError: (e) => {
                  toast.error(e.title, {
                    description: e.detail,
                  });
                },
              }
            )
          }
        >
          {storeHosts.isPending ? (
            <LoaderCircle className="animate-spin" />
          ) : (
            <span>Submit</span>
          )}
        </Button>
      </div>
    </div>
  );
}
