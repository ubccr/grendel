import { postV1BmcAssetTagMutation } from "@/client/@tanstack/react-query.gen";
import { ConfirmDialog } from "@/components/actions/action-card";
import PreviewTable, { counts, PreviewRow } from "@/components/identity/preview-table";
import TemplateHelp from "@/components/identity/template-help";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { toastBmcResponse } from "@/lib/bmc";
import { assetCsv, guessColumns, parseCsv } from "@/lib/identity";
import { useMutation } from "@tanstack/react-query";
import { Eye, FileSpreadsheet, LoaderCircle, Tag, TriangleAlert } from "lucide-react";
import { useMemo, useState } from "react";
import { toast } from "sonner";

type Csv = { file: string; rows: string[][]; header: boolean; serial: number; tag: number };

// AssetTagForm sets asset tags from a template, or from a CSV matched to each node by the serial its BMC reports
export default function AssetTagForm({
  nodes,
  example,
  defaultTemplate,
}: {
  nodes: string;
  example?: string;
  defaultTemplate?: string;
}) {
  const [mode, setMode] = useState<"template" | "csv">("csv");
  const [template, setTemplate] = useState(defaultTemplate ?? "");
  const [csv, setCsv] = useState<Csv>();
  const [preview, setPreview] = useState<{ rows: PreviewRow[]; applied: boolean }>();
  const [running, setRunning] = useState<"preview" | "apply">();
  const { mutate } = useMutation(postV1BmcAssetTagMutation());

  const parsed = useMemo(
    () => (csv ? assetCsv(csv.rows, csv.header, csv.serial, csv.tag) : undefined),
    [csv],
  );
  const ready = mode === "template" ? template.trim() !== "" : !!parsed && !parsed.error;

  const changeCsv = (next: Csv | undefined) => {
    setCsv(next);
    setPreview(undefined);
  };

  const readFile = async (file?: File) => {
    if (!file) return changeCsv(undefined);
    const rows = parseCsv(await file.text());
    const g = guessColumns(rows);
    changeCsv({ file: file.name, rows, header: g.header, serial: g.serialIndex, tag: g.tagIndex });
  };

  const run = (dryRun: boolean) => {
    const body =
      mode === "template"
        ? { template, dry_run: dryRun }
        : { by_serial: parsed?.bySerial, dry_run: dryRun };
    setRunning(dryRun ? "preview" : "apply");
    mutate(
      { query: { nodeset: nodes }, body },
      {
        onSuccess: (data) => {
          if (!dryRun) toastBmcResponse("Set asset tag", data);
          setPreview({
            applied: !dryRun,
            rows: (data?.results ?? []).map((r) => ({
              node: r.node ?? "",
              serial: r.data?.serial,
              current: r.data?.current ?? "",
              next: r.data?.desired ?? "",
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
  // a csv row whose serial no reached node reported. Nodes that failed before reading their serial cannot claim a row, so this is what was not applied rather than proof the serial is absent
  const unused =
    preview && parsed && mode === "csv"
      ? Object.keys(parsed.bySerial).filter(
          (s) => !preview.rows.some((r) => r.success && r.serial === s),
        )
      : [];
  const columns =
    csv?.rows[0]?.map((h, i) => ({
      i,
      label: csv.header ? h || `Column ${i + 1}` : `Column ${i + 1}`,
    })) ?? [];

  return (
    <div className="grid gap-4">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <ToggleGroup
          variant="outline"
          value={[mode]}
          onValueChange={(v) => {
            if (!v[0]) return;
            setMode(v[0] as "template" | "csv");
            setPreview(undefined);
          }}
        >
          <ToggleGroupItem value="csv" className="gap-1.5 px-3">
            <FileSpreadsheet className="size-4" />
            CSV by serial
          </ToggleGroupItem>
          <ToggleGroupItem value="template" className="gap-1.5 px-3">
            <Tag className="size-4" />
            Template
          </ToggleGroupItem>
        </ToggleGroup>
        {mode === "template" && <TemplateHelp example={example} />}
      </div>

      {mode === "template" ? (
        <div className="grid max-w-xl gap-1.5">
          <Label htmlFor="asset-template">Asset tag</Label>
          <Input
            id="asset-template"
            className="font-mono text-xs"
            placeholder="UB-{{ .Name }}"
            value={template}
            onChange={(e) => {
              setTemplate(e.target.value);
              setPreview(undefined);
            }}
          />
        </div>
      ) : (
        <div className="grid gap-3">
          <p className="text-sm text-muted-foreground">
            Each node is matched by the serial (service tag) its BMC reports. A selected node with
            no row in the file fails and is left unchanged.
          </p>
          <div className="grid max-w-xl gap-1.5">
            <Label htmlFor="asset-csv">CSV file</Label>
            <Input
              id="asset-csv"
              type="file"
              accept=".csv,text/csv,text/plain"
              onChange={(e) => readFile(e.target.files?.[0])}
            />
          </div>
          {csv && (
            <div className="grid gap-3 rounded-lg border bg-muted/30 p-3">
              <div className="flex flex-wrap items-end gap-3">
                <ColumnSelect
                  label="Serial column"
                  value={csv.serial}
                  columns={columns}
                  onChange={(serial) => changeCsv({ ...csv, serial })}
                />
                <ColumnSelect
                  label="Asset tag column"
                  value={csv.tag}
                  columns={columns}
                  onChange={(tag) => changeCsv({ ...csv, tag })}
                />
                <div className="flex h-9 items-center gap-2">
                  <Checkbox
                    id="csv-header"
                    checked={csv.header}
                    onCheckedChange={(v) => changeCsv({ ...csv, header: v === true })}
                  />
                  <Label htmlFor="csv-header" className="font-normal">
                    First row is a header
                  </Label>
                </div>
              </div>
              {parsed?.error ? (
                <p className="text-sm text-destructive">{parsed.error}</p>
              ) : (
                <p className="text-sm text-muted-foreground">
                  {Object.keys(parsed?.bySerial ?? {}).length} serial(s) read from{" "}
                  <span className="font-mono text-foreground">{csv.file}</span>.
                </p>
              )}
            </div>
          )}
        </div>
      )}

      <div className="flex flex-wrap justify-end gap-2">
        <Button variant="secondary" disabled={!ready || !!running} onClick={() => run(true)}>
          {running === "preview" ? <LoaderCircle className="animate-spin" /> : <Eye />}
          Preview
        </Button>
        <ConfirmDialog
          trigger={
            <Button disabled={!preview || preview.applied || !c?.change || !!running}>
              {running === "apply" ? <LoaderCircle className="animate-spin" /> : <Tag />}
              Apply {c?.change ? `to ${c.change} node(s)` : ""}
            </Button>
          }
          onConfirm={() => run(false)}
        >
          Write the previewed asset tags to the BMC of {c?.change} node(s).
          {c?.failed ? ` The ${c.failed} node(s) that failed the preview are left unchanged.` : ""}
        </ConfirmDialog>
      </div>

      {preview && (
        <PreviewTable rows={preview.rows} applied={preview.applied} showSerial={mode === "csv"} />
      )}
      {unused.length > 0 && (
        <div className="flex items-start gap-2 rounded-md border border-amber-500/40 bg-amber-500/10 px-3 py-2 text-sm text-amber-700 dark:text-amber-300">
          <TriangleAlert className="mt-0.5 size-4 shrink-0" />
          <span>
            {unused.length} CSV row(s) went to no node:{" "}
            <span className="font-mono">{unused.join(", ")}</span>
          </span>
        </div>
      )}
    </div>
  );
}

function ColumnSelect({
  label,
  value,
  columns,
  onChange,
}: {
  label: string;
  value: number;
  columns: Array<{ i: number; label: string }>;
  onChange: (i: number) => void;
}) {
  return (
    <div className="grid gap-1.5">
      <Label className="text-xs">{label}</Label>
      <Select
        value={value >= 0 ? String(value) : undefined}
        items={columns.map((c) => ({ value: String(c.i), label: c.label }))}
        onValueChange={(v) => onChange(Number(v))}
      >
        <SelectTrigger className="w-48">
          <SelectValue placeholder="Pick a column" />
        </SelectTrigger>
        <SelectContent>
          {columns.map((c) => (
            <SelectItem key={c.i} value={String(c.i)}>
              {c.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
}
