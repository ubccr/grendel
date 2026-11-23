import { postV1NodesImport, type NodeImportRequest, type NodeImportResponse } from "@/client";
import { postV1NodesImportMutation } from "@/client/@tanstack/react-query.gen";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import type { ImportRow } from "@/lib/import-rows";
import { cn } from "@/lib/utils";
import { useMutation, useQuery } from "@tanstack/react-query";
import { ArrowRight, Eye, LoaderCircle, Upload } from "lucide-react";
import { useStoredValue } from "@/hooks/use-stored-value";
import { useId, useMemo, useState, type ReactNode } from "react";
import { toast } from "sonner";
import { rangeKey } from "@/lib/addresses";
import { cellInput } from "@/components/form-layout";
import { ResetButton } from "./port-grid";
import { RangePicker } from "./range-picker";
import { StepBar } from "./step-bar";

type Result = NonNullable<NodeImportResponse["rows"]>[number];
type Action = "create" | "add" | "replace" | "unchanged" | "conflict";

const s = (n: number) => (n === 1 ? "" : "s");

// an address typed over the one a row would get is kept by the row's port, so it survives a preview that reorders or renumbers the rows
const editKey = (r: ImportRow) => `${r.node}|${r.mac}|${r.bmc}`;

const actionText: Record<Action, { label: string; count: (n: number) => string; color?: string }> =
  {
    conflict: { label: "Conflict", count: (n) => `conflict${s(n)}`, color: "var(--destructive)" },
    create: { label: "New node", count: (n) => `new node${s(n)}`, color: "var(--node-ready)" },
    add: {
      label: "Add port",
      count: (n) => `port${s(n)} added to existing nodes`,
      color: "var(--node-ready)",
    },
    replace: {
      label: "Replace port",
      count: (n) => `port${s(n)} changed`,
      color: "var(--node-amber)",
    },
    unchanged: { label: "No change", count: () => "unchanged" },
  };
const order: Action[] = ["conflict", "create", "add", "replace", "unchanged"];

export function Review({
  rows,
  bmc,
  needsAddresses,
  back,
  onCommitted,
}: {
  rows: ImportRow[];
  bmc: boolean;
  // needsAddresses is whether any row is a port Grendel doesn't have yet. A port it already has keeps its address, so an import of only those needs no range
  needsAddresses: boolean;
  back?: ReactNode;
  onCommitted: (res: NodeImportResponse, nodes: string[]) => void;
}) {
  const id = useId();
  // starts as the range last used for this kind of port, and stays empty once cleared
  const [remembered] = useStoredValue(rangeKey(bmc));
  const [range, setRange] = useState(remembered ?? "");
  const [from, setFrom] = useState<"lowest" | "after_highest">("lowest");
  const [ping, setPing] = useState(true);
  const { mutate, isPending } = useMutation(postV1NodesImportMutation());
  const [edits, setEdits] = useState<Record<string, string>>({});
  // the edits the preview on screen was made with; any others haven't been checked by the server yet
  const [checked, setChecked] = useState("{}");
  const editsKey = useMemo(() => JSON.stringify(edits), [edits]);

  const request = useMemo<NodeImportRequest>(
    () => ({
      ping,
      from,
      ranges: bmc ? { bmc: range.trim() } : { boot: range.trim() },
      rows,
    }),
    [rows, range, from, ping, bmc],
  );
  // a dry run is keyed on the request and only fetched on asking, so any change to the rows or the range leaves no preview until the next one, and what is committed is always what was last shown
  // keyed on the request as text, so its thousands of rows are serialized once per change rather than hashed on every render
  const requestKey = useMemo(() => JSON.stringify(request), [request]);
  const dry = useQuery({
    queryKey: ["node-import-preview", requestKey],
    queryFn: async () =>
      (
        await postV1NodesImport({
          body: {
            ...request,
            dry_run: true,
            rows: rows.map((r) => ({ ...r, ip: edits[editKey(r)]?.trim() || r.ip })),
          },
          throwOnError: true,
        })
      ).data,
    enabled: false,
    // dropped as soon as the request changes, so returning to an earlier request never shows a preview from before other changes were stored
    gcTime: 0,
  });
  const preview = dry.data;
  // an address changed since the preview is one the server hasn't checked for conflicts or pinged
  const edited = !!preview && editsKey !== checked;

  const runPreview = async () => {
    const sent = editsKey;
    const { error } = await dry.refetch();
    if (error) {
      const err = error as { title?: string; detail?: string };
      toast.error(err.title ?? "Preview failed", { description: err.detail });
    } else setChecked(sent);
  };

  const setAddress = (key: string, ip: string | undefined) =>
    setEdits((e) => {
      const { [key]: _, ...rest } = e;
      return ip === undefined ? rest : { ...rest, [key]: ip };
    });

  const commit = () => {
    // the previewed addresses are sent back, so the server stores exactly what was shown, and a range filling up in between is a conflict rather than a different address
    const shown = preview?.rows ?? [];
    const body = {
      ...request,
      dry_run: false,
      ping: false,
      rows: rows.map((r, i) => ({ ...r, ip: shown[i]?.ip || r.ip })),
    };
    mutate(
      { body },
      { onSuccess: (res) => onCommitted(res, [...new Set(rows.map((r) => r.node ?? ""))]) },
    );
  };

  const summary = preview?.summary;
  const changes = summary ? (summary.create ?? 0) + (summary.add ?? 0) + (summary.replace ?? 0) : 0;
  const blocked = (summary?.conflict ?? 0) > 0;
  const dryRunning = dry.isFetching;
  // a new port takes its address from the range, so there is nothing to preview until one is given
  const needsRange = !range.trim() && needsAddresses;

  return (
    <div className="grid gap-6">
      <div className="grid gap-5 rounded-lg border bg-muted/30 p-4 md:grid-cols-[minmax(0,22rem)_1fr]">
        <RangePicker bmc={bmc} value={range} onChange={setRange} />
        <div className="grid content-start gap-4">
          <div className="grid gap-2">
            <span className="text-sm font-medium">Start from</span>
            <RadioGroup
              value={from}
              onValueChange={(v) => setFrom(v as typeof from)}
              className="flex flex-wrap gap-x-5 gap-y-2"
            >
              <label className="flex items-center gap-2 text-sm">
                <RadioGroupItem value="lowest" />
                The lowest free address
              </label>
              <label className="flex items-center gap-2 text-sm">
                <RadioGroupItem value="after_highest" />
                After the highest one in use
              </label>
            </RadioGroup>
          </div>
          <div className="flex items-center gap-2">
            <Checkbox
              id={`${id}-ping`}
              checked={ping}
              onCheckedChange={(v) => setPing(v === true)}
            />
            <Label htmlFor={`${id}-ping`} className="font-normal">
              Ping each address first and skip any that answer
            </Label>
          </div>
        </div>
      </div>

      {preview && (
        <div className="grid gap-4">
          {preview.ping_error && (
            <p className="text-sm text-amber-700 dark:text-amber-300">
              Addresses weren't pinged: {preview.ping_error}
            </p>
          )}
          <Summary summary={summary} />
          <PreviewTable
            rows={preview.rows ?? []}
            keys={rows.map(editKey)}
            edits={edits}
            onEdit={setAddress}
            onSubmit={runPreview}
          />
        </div>
      )}

      <StepBar>
        {back}
        <p
          className={cn("mr-auto text-sm", blocked ? "text-destructive" : "text-muted-foreground")}
        >
          {needsRange && !preview
            ? `Choose a ${bmc ? "BMC" : "boot"} address range to hand out addresses from.`
            : !preview
              ? `${rows.length} ${bmc ? "BMC" : "boot"} port${rows.length === 1 ? "" : "s"} to import. Preview to hand out addresses and check for conflicts.`
              : edited
                ? "Preview again to check the addresses you changed."
                : blocked
                  ? "Resolve the conflicts before importing. Nothing is stored while any row conflicts."
                  : changes === 0
                    ? "Every port is already in Grendel as shown."
                    : "Stored in one step: every row or none."}
        </p>
        <Button
          variant={preview && !edited ? "outline" : "default"}
          disabled={rows.length === 0 || isPending || dryRunning || needsRange}
          onClick={runPreview}
        >
          {dryRunning ? <LoaderCircle className="animate-spin" /> : <Eye />}
          {preview ? "Preview again" : "Preview"}
        </Button>
        {preview && (
          <Button
            variant={edited ? "outline" : "default"}
            disabled={edited || blocked || changes === 0 || isPending || dryRunning}
            onClick={commit}
          >
            {isPending ? <LoaderCircle className="animate-spin" /> : <Upload />}
            Import {changes} port{changes === 1 ? "" : "s"}
          </Button>
        )}
      </StepBar>
    </div>
  );
}

function Summary({ summary }: { summary: NodeImportResponse["summary"] }) {
  return (
    <ul className="flex flex-wrap gap-2">
      {order
        .filter((a) => summary?.[a])
        .map((a) => (
          <li
            key={a}
            className="inline-flex items-center gap-2 rounded-full border px-3 py-1 text-xs"
          >
            <span
              aria-hidden
              className={cn(
                "size-2 rounded-full",
                !actionText[a].color && "bg-muted-foreground/40",
              )}
              style={actionText[a].color ? { backgroundColor: actionText[a].color } : undefined}
            />
            <span className="font-medium tabular-nums">{summary?.[a]}</span>
            {actionText[a].count(summary?.[a] ?? 0)}
          </li>
        ))}
    </ul>
  );
}

function PreviewTable({
  rows,
  keys,
  edits,
  onEdit,
  onSubmit,
}: {
  rows: Result[];
  // keys names each row's edit, in the order the rows were sent, which is the order the server answers in
  keys: string[];
  edits: Record<string, string>;
  onEdit: (key: string, ip: string | undefined) => void;
  onSubmit: () => void;
}) {
  const sorted = rows
    .map((r, i) => ({ r, key: keys[i] ?? "" }))
    .sort((a, b) => order.indexOf(a.r.action as Action) - order.indexOf(b.r.action as Action));
  return (
    <div className="max-h-[min(60vh,40rem)] overflow-auto rounded-lg border">
      <table className="w-full min-w-[48rem] text-sm">
        <thead className="sticky top-0 bg-muted text-left text-xs text-muted-foreground shadow-[0_1px_0_var(--border)]">
          <tr>
            <th className="px-3 py-2 font-medium">Node</th>
            <th className="px-2 py-2 font-medium">Change</th>
            <th className="px-2 py-2 font-medium">MAC</th>
            <th className="px-2 py-2 font-medium">Address</th>
            <th className="px-3 py-2 font-medium">FQDN</th>
          </tr>
        </thead>
        <tbody>
          {sorted.map(({ r, key }) => {
            const a = actionText[r.action as Action] ?? actionText.unchanged;
            const typed = edits[key];
            return (
              <tr
                key={`${r.node}-${r.mac}-${r.bmc}`}
                className={cn(
                  "border-b last:border-b-0",
                  r.action === "unchanged" && "text-muted-foreground",
                )}
                style={a.color ? { boxShadow: `inset 3px 0 0 ${a.color}` } : undefined}
              >
                <td className="px-3 py-1.5 font-mono text-xs">{r.node}</td>
                <td className="px-2 py-1.5">
                  <span className="text-xs font-medium">
                    {r.action === "replace" && !r.replaces ? "Update port" : a.label}
                  </span>
                  {r.reason && (
                    <p
                      className={cn(
                        "text-xs",
                        r.action === "conflict" ? "text-destructive" : "text-muted-foreground",
                      )}
                    >
                      {r.reason}
                    </p>
                  )}
                </td>
                <td className="px-2 py-1.5 font-mono text-xs whitespace-nowrap">
                  {r.replaces && (
                    <>
                      <span className="text-muted-foreground line-through">{r.replaces}</span>
                      <ArrowRight className="mx-1 inline size-3 text-muted-foreground" />
                    </>
                  )}
                  {r.mac}
                </td>
                <td className="px-2 py-1 text-xs whitespace-nowrap">
                  {/* sized in the input's own font to hold the longest address, 255.255.255.255/32, beside the reset button */}
                  <span className="relative inline-block w-[calc(19ch+2.5rem)] align-middle font-mono text-xs">
                    <Input
                      aria-label={`Address for ${r.node}`}
                      value={typed ?? r.ip ?? ""}
                      placeholder="From the range"
                      spellCheck={false}
                      className={cn(
                        cellInput,
                        "h-7 pr-8 text-xs md:text-xs",
                        typed !== undefined && "bg-sky-500/10",
                      )}
                      onChange={(e) => onEdit(key, e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === "Enter") onSubmit();
                      }}
                    />
                    {typed !== undefined && (
                      <ResetButton
                        title="Take the address from the range or the port again"
                        label={`Reset the address for ${r.node}`}
                        onClick={() => onEdit(key, undefined)}
                      />
                    )}
                  </span>
                  {r.assigned && typed === undefined && (
                    <span className="ml-1.5 font-sans text-[11px] text-muted-foreground">new</span>
                  )}
                  {r.answered && (
                    <span
                      className="ml-1.5 font-sans text-[11px] text-amber-700 dark:text-amber-300"
                      title="Something already answers a ping at this address"
                    >
                      answered a ping
                    </span>
                  )}
                </td>
                <td
                  className="max-w-64 truncate px-3 py-1.5 font-mono text-xs text-muted-foreground"
                  title={r.fqdn}
                >
                  {r.fqdn}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
