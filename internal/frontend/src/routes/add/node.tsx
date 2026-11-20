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
import { createFileRoute } from "@tanstack/react-router";
import { LoaderCircle } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import z from "zod";
import { themeToMonaco } from "../../hooks/theme-provider";

export const Route = createFileRoute("/add/node")({
  staticData: { crumb: [{ label: "Nodes", to: "/nodes" }, "Add node"] },
  component: RouteComponent,
  validateSearch: z.object({
    tab: z.enum(["form", "json", "switches", "boot"]).default("form").catch("form"),
    step: z.enum(importSteps).default("switches").catch("switches"),
    // the switches are shared by both imports, so the interfaces step scans the ones the BMC import did
    switches: z.string().default("").catch(""),
    bootStep: z.enum(bootSteps).default("nodes").catch("nodes"),
    nodes: z.string().default("").catch(""),
  }),
  beforeLoad: AuthRedirect,
});

function RouteComponent() {
  const { tab, step, switches, bootStep, nodes } = Route.useSearch();
  const navigate = Route.useNavigate();
  // each import is drawn once its tab is first opened, so the form doesn't wait on lookups it doesn't use
  const [opened, setOpened] = useState({ switches: tab === "switches", boot: tab === "boot" });
  if ((tab === "switches" || tab === "boot") && !opened[tab]) setOpened({ ...opened, [tab]: true });
  const onSearch = (patch: object, replace?: boolean) =>
    navigate({ search: (s) => ({ ...s, ...patch }), replace });

  return (
    <SwitchImport search={{ step, switches }} onSearch={onSearch}>
      {(bmc) => (
        <BootImport search={{ bootStep, nodes, switches }} onSearch={onSearch}>
          {(boot) => (
            // the imports' step bars stick to the bottom of the window, which a card clipping its overflow would stop
            <Card className="overflow-visible">
              <CardContent>
                <Tabs
                  className="w-full"
                  value={tab}
                  onValueChange={(v) => navigate({ search: (s) => ({ ...s, tab: v }) })}
                >
                  <div className="grid justify-items-center gap-4 pt-2">
                    <TabsList>
                      <TabsTrigger value="form">Form</TabsTrigger>
                      <TabsTrigger value="json">JSON</TabsTrigger>
                      <TabsTrigger value="switches">Switch</TabsTrigger>
                      <TabsTrigger value="boot">Add Interfaces</TabsTrigger>
                    </TabsList>
                    {tab === "switches" && bmc.stepper}
                    {tab === "boot" && boot.stepper}
                  </div>
                  <TabsContent value="form">
                    <NodeForm />
                  </TabsContent>
                  <TabsContent value="json">
                    <NodeImportJSON />
                  </TabsContent>
                  {/* kept mounted so what is typed into an import survives a look at the other tabs */}
                  <TabsContent value="switches" keepMounted className="grid gap-6 pt-4">
                    {opened.switches && bmc.body}
                  </TabsContent>
                  <TabsContent value="boot" keepMounted className="grid gap-6 pt-4">
                    {opened.boot && boot.body}
                  </TabsContent>
                </Tabs>
              </CardContent>
            </Card>
          )}
        </BootImport>
      )}
    </SwitchImport>
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
