import AuthRedirect from "@/auth";
import ImageForm from "@/components/images/form";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { themeToMonaco, useTheme } from "@/hooks/theme-provider";
import { usePostV1Images } from "@/openapi/queries";
import { Editor } from "@monaco-editor/react";
import { useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { LoaderCircle } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import z from "zod";

export const Route = createFileRoute("/add/image")({
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
            <ImageForm />
          </TabsContent>
          <TabsContent value="json">
            <ImageImportJSON />
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  );
}

const defaultJson = {
  boot_images: [],
};

function ImageImportJSON() {
  const { theme } = useTheme();
  const storeImages = usePostV1Images();
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
            storeImages.mutate(
              { body: JSON.parse(text) },
              {
                onSuccess: (e) => {
                  toast.success(e.data?.title, { description: e.data?.detail });
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
          {storeImages.isPending ? (
            <LoaderCircle className="animate-spin" />
          ) : (
            <span>Submit</span>
          )}
        </Button>
      </div>
    </div>
  );
}
