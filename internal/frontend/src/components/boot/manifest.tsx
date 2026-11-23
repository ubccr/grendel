import {
  fileLabel,
  fileStep,
  formatBytes,
  isButanePair,
  kindLabel,
  viewKind,
  type NodeCheck,
} from "@/lib/boot-check";
import { cn } from "@/lib/utils";
import { Check, ChevronRight, X } from "lucide-react";
import type { ReactNode } from "react";

type Props = {
  check: NodeCheck;
  // open is the step shown on the right: the kind a template renders as, or a boot file as fileStep names it
  open: string;
  onOpen: (step: string) => void;
};

// ok is undefined for a step that was not checked, since the image does not name it
function Mark({ ok }: { ok?: boolean }) {
  const base = "relative z-10 mt-0.5 flex size-5 shrink-0 items-center justify-center rounded-full";
  if (ok === undefined)
    return (
      <span className={cn(base, "border-2 border-border bg-background")}>
        <span className="sr-only">Not checked</span>
      </span>
    );
  return (
    <span className={cn(base, "text-white", ok ? "bg-emerald-600" : "bg-destructive")}>
      {ok ? <Check className="size-3" strokeWidth={3} /> : <X className="size-3" strokeWidth={3} />}
      <span className="sr-only">{ok ? "Ready" : "Failed"}</span>
    </span>
  );
}

function Step({
  ok,
  label,
  detail,
  aside,
  selected,
  onSelect,
}: {
  ok?: boolean;
  label: string;
  detail?: string;
  aside?: ReactNode;
  selected: boolean;
  onSelect: () => void;
}) {
  return (
    // the rail joins each mark to the next, since this is the order the node asks for them in
    <li className="relative after:absolute after:top-7 after:-bottom-3.5 after:left-[1.125rem] after:w-px after:bg-border last:after:hidden">
      <button
        type="button"
        onClick={onSelect}
        aria-current={selected ? "true" : undefined}
        className={cn(
          "flex w-full gap-3 rounded-lg px-2 py-2 text-left transition-colors outline-none hover:bg-muted/60 focus-visible:ring-2 focus-visible:ring-ring",
          selected && "bg-muted hover:bg-muted",
        )}
      >
        <Mark ok={ok} />
        <span className="grid min-w-0 flex-1 gap-0.5">
          <span className="flex items-baseline justify-between gap-3">
            <span className={cn("text-sm font-medium", ok === false && "text-destructive")}>
              {label}
            </span>
            {aside && (
              <span className="shrink-0 text-xs text-muted-foreground tabular-nums">{aside}</span>
            )}
          </span>
          {detail && (
            <span className="truncate font-mono text-xs text-muted-foreground" title={detail}>
              {detail}
            </span>
          )}
        </span>
        <ChevronRight
          className={cn(
            "mt-0.5 size-4 shrink-0 text-muted-foreground transition-opacity",
            !selected && "opacity-40",
          )}
        />
      </button>
    </li>
  );
}

// Manifest lists what the node is sent as it boots: the iPXE script, the kernel and initrds that script loads, then the configs the booted installer asks for. Opening a step shows its rendered output, or where its file is, or why it failed
export function Manifest({ check, open, onOpen }: Props) {
  const templates = check.templates ?? [];
  const ipxe = templates.filter((t) => t.kind === "ipxe");
  const configs = templates.filter((t) => t.kind !== "ipxe");
  const onRequest = check.on_request ?? [];

  // a butane template is one step whichever side of its translation is open
  const isOpen = (kind: string) => open === kind || (isButanePair(open) && isButanePair(kind));

  const template = (t: (typeof templates)[number]) => {
    const kind = viewKind(t.kind ?? "");
    return (
      <Step
        key={t.kind}
        ok={!t.error}
        label={kindLabel(t.kind ?? "")}
        detail={t.template}
        selected={isOpen(kind)}
        onSelect={() => onOpen(kind)}
      />
    );
  };

  return (
    <div className="grid content-start gap-4">
      <ol className="grid gap-1" aria-label="What the node is sent, in order">
        {ipxe.map(template)}
        {(check.files ?? []).map((f) => {
          const step = fileStep(f.name ?? "");
          return (
            <Step
              key={step}
              ok={!f.error}
              label={fileLabel(f.name ?? "")}
              detail={f.path}
              aside={!f.error && f.size !== undefined ? formatBytes(f.size) : undefined}
              selected={open === step}
              onSelect={() => onOpen(step)}
            />
          );
        })}
        {configs.map(template)}
      </ol>

      {onRequest.length > 0 && (
        <div className="grid gap-2 border-t pt-4">
          <h3 className="px-2 text-xs text-muted-foreground">
            Served on request, though the image doesn't name them
          </h3>
          <ul className="grid gap-1">
            {onRequest.map((t) => (
              <Step
                key={t.kind}
                label={kindLabel(t.kind ?? "")}
                detail={t.template}
                selected={isOpen(t.kind ?? "")}
                onSelect={() => onOpen(t.kind ?? "")}
              />
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}
