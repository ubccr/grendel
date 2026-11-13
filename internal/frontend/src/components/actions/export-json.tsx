import { Button } from "@/components/ui/button";
import {
  CodeBlock,
  CodeBlockActions,
  CodeBlockContent,
  CodeBlockCopyButton,
  CodeBlockHeader,
  CodeBlockLine,
  CodeBlockTitle,
} from "@/components/ui/code-block";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Skeleton } from "@/components/ui/skeleton";
import { Switch } from "@/components/ui/switch";
import { highlightJson } from "@/lib/highlight-json";
import { cn } from "@/lib/utils";
import { type QueryKey, useQuery, type UseQueryOptions } from "@tanstack/react-query";
import { Download, FileJson } from "lucide-react";
import { type ReactNode, useId, useMemo, useState } from "react";

type Named = { name?: string };

type Props<T extends Named> = {
  trigger: ReactNode;
  // noun is the CLI noun, so the dialog can name the import command: node or image
  noun: "node" | "image";
  options: UseQueryOptions<T[] | undefined, Error, T[] | undefined, QueryKey>;
};

// MAX_HIGHLIGHT is the most JSON that gets colored, so a very large export shows as plain text rather than stalling the page
const MAX_HIGHLIGHT = 512 * 1024;

// withoutIds drops the database ids from nodes, their interfaces and bonds, and images. The store upserts on id, so an import that carries them overwrites whatever records have those ids, in another Grendel a different node entirely. Without them the uids are generated fresh
export function withoutIds<T>(records: T[]): T[] {
  return records.map((record) => {
    const { id: _id, uid: _uid, ...rest } = record as Record<string, unknown>;
    for (const key of ["interfaces", "bonds"]) {
      const nics = rest[key];
      if (Array.isArray(nics))
        rest[key] = nics.map((nic: Record<string, unknown>) => {
          const { id: _nicId, ...n } = nic;
          return n;
        });
    }
    return rest as T;
  });
}

function formatSize(bytes: number) {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
}

// ExportJson shows records as the JSON the CLI's import command reads, to copy or save as a file. It only fetches while open, so the records are current each time
export function ExportJson<T extends Named>({ trigger, noun, options }: Props<T>) {
  const [open, setOpen] = useState(false);
  const [ids, setIds] = useState(false);
  const idsSwitch = useId();
  // the client throws the server's error body rather than an Error, so its detail says what went wrong
  const { data, isPending, isError, error } = useQuery({ ...options, enabled: open });
  const reason = (error as { detail?: string } | null)?.detail;

  const code = useMemo(
    () => (data ? JSON.stringify(ids ? data : withoutIds(data), null, 2) : ""),
    [data, ids],
  );
  const lines = useMemo(
    () => (code.length > MAX_HIGHLIGHT ? code.split("\n") : highlightJson(code)),
    [code],
  );
  const size = useMemo(() => new Blob([code]).size, [code]);
  const count = data?.length ?? 0;
  const plural = count === 1 ? noun : `${noun}s`;
  const filename = count === 1 && data?.[0]?.name ? `${data[0].name}.json` : `${noun}s.json`;

  const download = () => {
    const url = URL.createObjectURL(new Blob([code], { type: "application/json" }));
    const a = document.createElement("a");
    a.href = url;
    a.download = filename;
    a.click();
    setTimeout(() => URL.revokeObjectURL(url));
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{trigger}</DialogTrigger>
      <DialogContent className="flex max-h-[calc(100dvh-2rem)] max-w-[calc(100%-2rem)] flex-col gap-4 overflow-hidden sm:max-w-4xl">
        <DialogHeader className="gap-1.5 pr-6">
          <DialogTitle>{data ? `Export ${count} ${plural}` : `Export ${noun}s`}</DialogTitle>
          <DialogDescription>
            Copy the JSON, or save it and run{" "}
            <code className="rounded bg-muted px-1.5 py-0.5 font-mono text-[0.8125rem] text-foreground">
              grendel {noun} import {filename}
            </code>{" "}
            to add {count === 1 ? "it" : "them"} to another Grendel.
          </DialogDescription>
        </DialogHeader>
        <div
          className={cn(
            "flex items-start gap-3 rounded-md border px-3 py-2.5 text-sm transition-colors",
            ids && "border-amber-500/40 bg-amber-500/10",
          )}
        >
          <Switch
            id={idsSwitch}
            className="mt-0.5 data-[state=checked]:bg-amber-600"
            checked={ids}
            onCheckedChange={setIds}
          />
          <label htmlFor={idsSwitch} className="grid gap-0.5">
            <span className="font-medium">Include IDs</span>
            <span className="text-muted-foreground">
              {ids
                ? `Warning: importing this JSON will replace any existing ${noun === "node" ? "nodes or interfaces" : "images"} with matching IDs`
                : `Unique IDs are removed so this JSON is safe to import as a new ${noun}`}
            </span>
          </label>
        </div>
        {isError ? (
          <p className="rounded-md border border-destructive/40 bg-destructive/5 px-4 py-3 text-sm text-destructive">
            Couldn't load the {noun}s: {reason || "the request failed"}. Close this and try again.
          </p>
        ) : isPending || !data ? (
          <div className="grid gap-2.5 rounded-lg border p-4" aria-label="Loading">
            {[40, 64, 52, 72, 48, 60, 36].map((w, i) => (
              <Skeleton key={i} className="h-3.5" style={{ width: `${w}%` }} />
            ))}
          </div>
        ) : (
          <CodeBlock className="min-h-0 flex-1">
            <CodeBlockHeader>
              <CodeBlockTitle className="font-mono">
                <FileJson />
                {filename}
              </CodeBlockTitle>
              <span className="hidden text-xs whitespace-nowrap tabular-nums sm:inline">
                {lines.length.toLocaleString()} lines, {formatSize(size)}
              </span>
              <CodeBlockActions>
                <CodeBlockCopyButton value={code} className="size-8" />
                <Button type="button" size="sm" variant="ghost" onClick={download}>
                  <Download />
                  Download
                </Button>
              </CodeBlockActions>
            </CodeBlockHeader>
            <CodeBlockContent lineNumbers aria-label={filename}>
              {lines.map((line, i) => (
                <CodeBlockLine key={i}>{line}</CodeBlockLine>
              ))}
            </CodeBlockContent>
          </CodeBlock>
        )}
      </DialogContent>
    </Dialog>
  );
}
