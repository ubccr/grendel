import { postV1NodesMutation } from "@/client/@tanstack/react-query.gen";
import { BootImport, bootSteps } from "@/components/import/boot-import";
import { importSteps, SwitchImport } from "@/components/import/switch-import";
import NodeForm from "@/components/nodes/form";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useTheme } from "@/hooks/theme-provider";
import AuthRedirect from "@/lib/auth";
import { Editor } from "@monaco-editor/react";
import { useMutation } from "@tanstack/react-query";
import { createFileRoute, stripSearchParams } from "@tanstack/react-router";
import { LoaderCircle } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import z from "zod";
import { themeToMonaco } from "../../hooks/theme-provider";

// a step or choice still at its default is left out of the url, which then only carries where an import has got to
const searchDefaults = {
  tab: "form",
  step: "switches",
  switches: "",
  bootStep: "nodes",
  nodes: "",
  bootSwitches: "",
} as const;

export const Route = createFileRoute("/add/node")({
  staticData: { crumb: [{ label: "Nodes", to: "/nodes" }, "Add node"] },
  component: RouteComponent,
  validateSearch: z.object({
    tab: z.enum(["form", "json", "switches", "boot"]).default("form").catch("form"),
    step: z.enum(importSteps).default("switches").catch("switches"),
    switches: z.string().default("").catch(""),
    bootStep: z.enum(bootSteps).default("nodes").catch("nodes"),
    nodes: z.string().default("").catch(""),
    // the interfaces import's own switches, apart from the BMC import's, so picking switches in one tab never changes what the other has scanned
    bootSwitches: z.string().default("").catch(""),
  }),
  search: { middlewares: [stripSearchParams(searchDefaults)] },
  beforeLoad: AuthRedirect,
});

function RouteComponent() {
  const { tab } = Route.useSearch();
  const navigate = Route.useNavigate();
  // each import is drawn once its tab is first opened, so the form doesn't wait on lookups it doesn't use
  const [opened, setOpened] = useState({ switches: tab === "switches", boot: tab === "boot" });
  if ((tab === "switches" || tab === "boot") && !opened[tab]) setOpened({ ...opened, [tab]: true });

  return (
    // the imports' step bars stick to the bottom of the window, which a card clipping its overflow would stop
    <Card className="overflow-visible">
      <CardContent>
        <Tabs
          className="w-full"
          value={tab}
          onValueChange={(v) => navigate({ search: (s) => ({ ...s, tab: v }) })}
        >
          <div className="grid justify-items-center pt-2">
            <TabsList>
              <TabsTrigger value="form">Form</TabsTrigger>
              <TabsTrigger value="json">JSON</TabsTrigger>
              <TabsTrigger value="switches">Switch</TabsTrigger>
              <TabsTrigger value="boot">Add Interfaces</TabsTrigger>
            </TabsList>
          </div>
          <TabsContent value="form">
            <NodeForm />
          </TabsContent>
          <TabsContent value="json">
            <NodeImportJSON />
          </TabsContent>
          {/* kept mounted so what is typed into an import survives a look at the other tabs */}
          <TabsContent value="switches" keepMounted className="grid gap-6 pt-4">
            {opened.switches && <SwitchImport />}
          </TabsContent>
          <TabsContent value="boot" keepMounted className="grid gap-6 pt-4">
            {opened.boot && <BootImport />}
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
  const [text, setText] = useState(() => JSON.stringify(defaultJson, null, 4));

  const submit = () => {
    let body;
    try {
      body = JSON.parse(text);
    } catch (e) {
      toast.error("This isn't valid JSON", { description: (e as Error).message });
      return;
    }
    mutate(
      { body },
      { onSuccess: (data) => toast.success(data?.title, { description: data?.detail }) },
    );
  };

  return (
    <div>
      <Editor
        height="80vh"
        language="json"
        value={text}
        onChange={(e) => setText(e ?? "")}
        theme={themeToMonaco(theme)}
      />
      <div className="mt-2 flex justify-end">
        <Button onClick={submit}>
          {isPending ? <LoaderCircle className="animate-spin" /> : <span>Submit</span>}
        </Button>
      </div>
    </div>
  );
}
