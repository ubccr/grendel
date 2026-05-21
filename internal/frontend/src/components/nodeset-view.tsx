import { parseNodeSet } from "@/lib/nodeset";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "./ui/tabs";
import { ScrollArea } from "./ui/scroll-area";
import { useState } from "react";
import { Button } from "./ui/button";
import { CopyCheckIcon, CopyIcon } from "lucide-react";
import clsx from "clsx";

export default function NodesetView({ nodes, size = "md" }: { nodes: string; size?: "md" | "sm" }) {
  const nodeset = parseNodeSet(nodes);
  const [tab, setTab] = useState("nodeset");
  const [clicked, setClicked] = useState(false);

  return (
    <Tabs value={tab} onValueChange={setTab} className="mt-2">
      <div className="flex justify-between">
        <TabsList>
          <TabsTrigger value="nodeset">Nodeset</TabsTrigger>
          <TabsTrigger value="nodes">Nodes</TabsTrigger>
        </TabsList>
        <div>
          <Button
            size="icon"
            variant="outline"
            onClick={() => {
              setClicked(true);
              navigator.clipboard.writeText(tab === "nodeset" ? nodeset.toString() : nodes);
              setTimeout(() => {
                setClicked(false);
              }, 3000);
            }}
          >
            {clicked ? <CopyCheckIcon className="text-green-600" /> : <CopyIcon />}
          </Button>
        </div>
      </div>
      <ScrollArea
        className={clsx(
          "m-2 overflow-scroll text-sm text-muted-foreground",
          {
            "max-h-28": size === "sm",
          },
          {
            "max-h-56": size === "md",
          },
        )}
      >
        <TabsContent value="nodeset">{nodeset.toString()}</TabsContent>
        <TabsContent value="nodes">{nodes}</TabsContent>
      </ScrollArea>
    </Tabs>
  );
}
