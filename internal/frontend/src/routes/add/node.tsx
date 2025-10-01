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
import { themeToMonaco } from "../../hooks/theme-provider";
import { Card, CardContent } from "@/components/ui/card";
import z from "zod";

export const Route = createFileRoute("/add/node")({
  component: RouteComponent,
  validateSearch: z.object({
    tab: z.string().optional().catch("form"),
  }),
  beforeLoad: AuthRedirect,
});

function RouteComponent() {
  const search = Route.useSearch();
  const navigate = Route.useNavigate();

  return (
    <Card>
      <CardContent>
        <Tabs
          className="w-full"
          defaultValue={search.tab ?? "form"}
          onValueChange={(v) => navigate({ search: { tab: v } })}
        >
          <div className="pt-2 text-center">
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
      </CardContent>
    </Card>
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
        theme={themeToMonaco(theme)}
      />
      <div className="mt-2 flex justify-end">
        <Button
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
              },
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
