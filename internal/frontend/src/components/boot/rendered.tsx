import { Problem } from "@/components/boot/problem";
import { ErrorPanel } from "@/components/error-panel";
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
import { Skeleton } from "@/components/ui/skeleton";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import {
  isButanePair,
  kindLabel,
  languageOf,
  readable,
  translationFailure,
} from "@/lib/boot-check";
import { useHighlighted } from "@/lib/highlighter";
import { renderOptions } from "@/lib/queries";
import { cn } from "@/lib/utils";
import { useQuery } from "@tanstack/react-query";
import { FileX2 } from "lucide-react";
import { useMemo } from "react";

type Props = {
  node: string;
  kind: string;
  arch?: string;
  // onKind opens the other side of a butane template's translation
  onKind: (kind: string) => void;
};

// Translation switches a butane template between what the node is sent and what that was translated from
function Translation({ kind, onKind }: Pick<Props, "kind" | "onKind">) {
  return (
    <ToggleGroup
      aria-label="Before or after translation"
      variant="outline"
      size="sm"
      spacing={0}
      value={[kind]}
      onValueChange={(v) => v[0] && onKind(v[0])}
    >
      <ToggleGroupItem value="ignition" className="px-2.5 text-xs">
        Ignition
      </ToggleGroupItem>
      <ToggleGroupItem value="butane" className="px-2.5 text-xs">
        Butane
      </ToggleGroupItem>
    </ToggleGroup>
  );
}

// Rendered shows what the provision server would send node for kind, read from the templates on disk each time it is opened
export function Rendered({ node, kind, arch, onKind }: Props) {
  const { data, error, isPending, isFetching, dataUpdatedAt } = useQuery(
    renderOptions(node, kind, arch),
  );
  const result = data?.nodes?.[0];
  const content = useMemo(() => readable(kind, result?.content ?? ""), [kind, result?.content]);
  const lines = useHighlighted(content, languageOf(kind, result?.template), [
    node,
    kind,
    arch,
    dataUpdatedAt,
  ]);

  if (isPending) return <Skeleton className="h-96 w-full" />;
  if (error || !result)
    return (
      <ErrorPanel
        title="Could not render the template"
        message={error?.detail || error?.title}
        icon={FileX2}
      />
    );

  if (result.error) {
    const untranslated = translationFailure(result.error);
    return (
      <Problem
        title={
          untranslated
            ? `${kindLabel("butane")} did not translate to Ignition`
            : `${kindLabel(kind)} did not render`
        }
        error={result.error}
      >
        {untranslated && (
          <div>
            <Button variant="outline" size="sm" onClick={() => onKind("butane")}>
              Show the Butane{untranslated.line ? `, at line ${untranslated.line}` : ""}
            </Button>
          </div>
        )}
      </Problem>
    );
  }

  return (
    <CodeBlock className={cn("transition-opacity", isFetching && "opacity-60")}>
      <CodeBlockHeader className="flex-wrap border-b pb-2.5">
        <CodeBlockTitle>
          {kindLabel(kind)}
          {result.template && (
            <span className="ml-2 font-mono font-normal text-muted-foreground">
              {result.template}
            </span>
          )}
        </CodeBlockTitle>
        <CodeBlockActions className="gap-3">
          {isButanePair(kind) && <Translation kind={kind} onKind={onKind} />}
          <CodeBlockCopyButton value={content} />
        </CodeBlockActions>
      </CodeBlockHeader>
      <CodeBlockContent lineNumbers className="max-h-[70vh] text-xs">
        {lines.map((line, i) => (
          <CodeBlockLine key={i}>{line}</CodeBlockLine>
        ))}
      </CodeBlockContent>
    </CodeBlock>
  );
}
