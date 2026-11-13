import { postV1BmcConfigureImportMutation } from "@/client/@tanstack/react-query.gen";
import ActionCard, { ConfirmDialog } from "@/components/actions/action-card";
import NodesetView from "@/components/nodeset-view";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { toastBmcResponse } from "@/lib/bmc";
import { useMutation } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

const shutdownTypes = { NoReboot: "No Reboot", Graceful: "Graceful", Forced: "Forced" };

export default function NodesImportConfigAction({ nodes }: { nodes: string }) {
  const [shutdownType, setShutdownType] = useState("");
  const [file, setFile] = useState("");
  const { mutate, isPending } = useMutation(postV1BmcConfigureImportMutation());

  return (
    <ActionCard
      title="Import System Config"
      description="Apply a server configuration profile template that Grendel renders for each node."
      footer={
        <ConfirmDialog
          trigger={
            <Button disabled={isPending || !shutdownType || !file}>
              {isPending && <LoaderCircle className="animate-spin" />}
              Import
            </Button>
          }
          onConfirm={() =>
            mutate(
              {
                query: { nodeset: nodes },
                body: { file: file, shutdown_type: shutdownType },
              },
              {
                onSuccess: (data) => toastBmcResponse("Import config", data),
                onError: (e) => toast.error(e.title, { description: e.detail }),
              },
            )
          }
        >
          Import {file} with a {shutdownTypes[shutdownType as keyof typeof shutdownTypes]} shutdown
          on:
          <NodesetView nodes={nodes} />
        </ConfirmDialog>
      }
    >
      <div className="grid grid-cols-2 gap-2">
        <div className="grid gap-1.5">
          <Label htmlFor="import-file">File</Label>
          <Input
            id="import-file"
            placeholder="idrac-config.json"
            value={file}
            onChange={(e) => setFile(e.target.value)}
          />
        </div>
        <div className="grid gap-1.5">
          <Label htmlFor="shutdown-type">Shutdown</Label>
          <Select
            value={shutdownType}
            onValueChange={(v) => setShutdownType(v ?? "")}
            items={shutdownTypes}
          >
            <SelectTrigger id="shutdown-type">
              <SelectValue placeholder="Type" />
            </SelectTrigger>
            <SelectContent>
              {Object.entries(shutdownTypes).map(([value, label]) => (
                <SelectItem key={value} value={value}>
                  {label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>
    </ActionCard>
  );
}
