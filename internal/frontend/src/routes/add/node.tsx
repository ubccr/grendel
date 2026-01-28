import { postV1NodesMutation } from "@/client/@tanstack/react-query.gen";
import NodeForm from "@/components/nodes/form";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useTheme } from "@/hooks/theme-provider";
import AuthRedirect from "@/lib/auth";
import { Editor } from "@monaco-editor/react";
import { useMutation } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { LoaderCircle } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import z from "zod";
import { themeToMonaco } from "../../hooks/theme-provider";

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
  const { mutate, isPending } = useMutation(postV1NodesMutation());
  const [text, setText] = useState("");

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
            mutate(
              { body: JSON.parse(text) },
              {
                onSuccess: (data) => {
                  toast.success(data?.title, {
                    description: data?.detail,
                  });
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
          {isPending ? <LoaderCircle className="animate-spin" /> : <span>Submit</span>}
        </Button>
      </div>
    </div>
  );
}
