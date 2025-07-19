import AuthRedirect from "@/auth";
import ImageForm from "@/components/images/form";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useTheme } from "@/hooks/theme-provider";
import { usePostV1Images } from "@/openapi/queries";
import { Editor } from "@monaco-editor/react";
import { useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { LoaderCircle } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

export const Route = createFileRoute("/add/image")({
  component: RouteComponent,
  beforeLoad: AuthRedirect,
});

function RouteComponent() {
  return (
    <div className="p-4">
      <Tabs defaultValue="form" className="w-full">
        <div className="text-center">
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
    </div>
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
        theme={theme == "dark" ? "vs-dark" : "light"}
      />
      <div className="flex justify-end mt-2">
        <Button
          variant="outline"
          size="sm"
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
              }
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
