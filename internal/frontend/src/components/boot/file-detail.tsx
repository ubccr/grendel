import { Problem } from "@/components/boot/problem";
import { fileLabel, formatBytes, type NodeCheck } from "@/lib/boot-check";

type File = NonNullable<NodeCheck["files"]>[number];

// FileDetail says where on the server a boot file was looked for, and how that went
export function FileDetail({ file }: { file: File }) {
  const label = fileLabel(file.name ?? "");

  if (file.error)
    return (
      <Problem
        title={
          file.path
            ? `${label} could not be read`
            : `The path of the ${label.toLowerCase()} did not render`
        }
        error={file.error}
      />
    );

  return (
    <section className="grid gap-4 rounded-lg border px-4 py-4">
      <h3 className="text-sm font-medium">{label} is readable</h3>
      <dl className="grid grid-cols-[auto_minmax(0,1fr)] gap-x-6 gap-y-1.5 text-sm">
        <dt className="text-muted-foreground">Path</dt>
        <dd className="font-mono text-xs leading-5 break-all">{file.path}</dd>
        <dt className="text-muted-foreground">Size</dt>
        <dd className="tabular-nums">
          {formatBytes(file.size ?? 0)}
          {file.size === 0 && <span className="ml-2 text-amber-600">The file is empty.</span>}
        </dd>
      </dl>
    </section>
  );
}
