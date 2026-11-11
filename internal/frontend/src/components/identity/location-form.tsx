import { BmcIdentityDefaults } from "@/client";
import { postV1BmcLocationMutation } from "@/client/@tanstack/react-query.gen";
import { ConfirmDialog } from "@/components/actions/action-card";
import PreviewTable, { counts, PreviewRow } from "@/components/identity/preview-table";
import TemplateHelp from "@/components/identity/template-help";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { toastBmcResponse } from "@/lib/bmc";
import { formatLocation } from "@/lib/identity";
import { useMutation } from "@tanstack/react-query";
import { Eye, LoaderCircle, MapPin } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

type Fields = { datacenter: string; room: string; row: string; rack: string; rack_unit: string };

const fieldInfo: Array<{ key: keyof Fields; label: string; hint: string }> = [
  { key: "rack", label: "Rack", hint: "{{ index .Parts 1 }}" },
  { key: "rack_unit", label: "Rack unit", hint: "{{ index .Parts 2 }}" },
  { key: "row", label: "Row", hint: "iDRAC calls this the aisle" },
  { key: "room", label: "Room", hint: "Suite 3" },
  { key: "datacenter", label: "Data center", hint: "Building or site" },
];

// LocationForm renders the location templates for each node, previews the result against what the BMCs hold, and writes it. Editing a field drops the preview, so what gets applied is always what was last previewed
export default function LocationForm({
  nodes,
  example,
  defaults,
}: {
  nodes: string;
  example?: string;
  defaults?: BmcIdentityDefaults;
}) {
  const [fields, setFields] = useState<Fields>({
    datacenter: defaults?.datacenter ?? "",
    room: defaults?.room ?? "",
    row: defaults?.row ?? "",
    rack: defaults?.rack ?? "",
    rack_unit: defaults?.rack_unit ?? "",
  });
  const [preview, setPreview] = useState<{ rows: PreviewRow[]; applied: boolean }>();
  const [running, setRunning] = useState<"preview" | "apply">();
  const { mutate } = useMutation(postV1BmcLocationMutation());

  const empty = Object.values(fields).every((v) => v.trim() === "");

  const run = (dryRun: boolean) => {
    setRunning(dryRun ? "preview" : "apply");
    mutate(
      { query: { nodeset: nodes }, body: { ...fields, dry_run: dryRun } },
      {
        onSuccess: (data) => {
          if (!dryRun) toastBmcResponse("Set location", data);
          setPreview({
            applied: !dryRun,
            rows: (data?.results ?? []).map((r) => ({
              node: r.node ?? "",
              current: formatLocation(r.data?.current),
              next: formatLocation(r.data?.desired),
              changed: !!r.data?.changed,
              success: !!r.success,
              message: r.message ?? "",
            })),
          });
        },
        onError: (e) => toast.error(e.title, { description: e.detail }),
        onSettled: () => setRunning(undefined),
      },
    );
  };

  const c = preview ? counts(preview.rows) : undefined;

  return (
    <div className="grid gap-4">
      <div className="flex items-start justify-between gap-4">
        <p className="text-sm text-muted-foreground">
          Each field is rendered per node. Leave a field empty to keep what the BMC has.
        </p>
        <TemplateHelp example={example} />
      </div>
      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-5">
        {fieldInfo.map((f) => (
          <div key={f.key} className="grid gap-1.5">
            <Label htmlFor={`loc-${f.key}`}>{f.label}</Label>
            <Input
              id={`loc-${f.key}`}
              className="font-mono text-xs"
              placeholder={f.hint}
              value={fields[f.key]}
              onChange={(e) => {
                setFields({ ...fields, [f.key]: e.target.value });
                setPreview(undefined);
              }}
            />
          </div>
        ))}
      </div>

      <div className="flex flex-wrap justify-end gap-2">
        <Button variant="secondary" disabled={empty || !!running} onClick={() => run(true)}>
          {running === "preview" ? <LoaderCircle className="animate-spin" /> : <Eye />}
          Preview
        </Button>
        <ConfirmDialog
          trigger={
            <Button disabled={!preview || preview.applied || !c?.change || !!running}>
              {running === "apply" ? <LoaderCircle className="animate-spin" /> : <MapPin />}
              Apply {c?.change ? `to ${c.change} node(s)` : ""}
            </Button>
          }
          onConfirm={() => run(false)}
        >
          Write the previewed location to the BMC of {c?.change} node(s).
          {c?.failed ? ` The ${c.failed} node(s) that failed the preview are left unchanged.` : ""}
        </ConfirmDialog>
      </div>

      {preview && <PreviewTable rows={preview.rows} applied={preview.applied} />}
    </div>
  );
}
