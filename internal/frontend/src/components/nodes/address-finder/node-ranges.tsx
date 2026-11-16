import { getV1NodesIpsRangesOptions } from "@/client/@tanstack/react-query.gen";
import { TagCombobox } from "@/components/tag-combobox";
import { Input } from "@/components/ui/input";
import { useDebounced } from "@/hooks/use-debounced";
import { cn } from "@/lib/utils";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import { useId, useState } from "react";

function plural(n: number, noun: string) {
  return `${n} ${noun}${n === 1 ? "" : "s"}`;
}

// NodeRanges finds the ranges some nodes' ports sit in, so a range can be picked by where similar nodes are rather than looked up. Nodes are picked by nodeset, by tags, or both, which narrows to the nodes matching both. The server picks them: the nodeset narrows by name as it is typed, so it works before it is a whole nodeset, and tags pick any node with one of them, as tags do elsewhere in the API. Both tabs stay mounted, so what is typed here survives a trip to the range tab
export function NodeRanges({
  bmc,
  current,
  onPick,
}: {
  bmc: boolean;
  current?: string;
  onPick: (range: string) => void;
}) {
  const id = useId();
  const [nodeset, setNodeset] = useState("");
  const [tags, setTags] = useState<string[]>([]);
  const [typed] = useDebounced(nodeset.trim(), 200);
  const picking = tags.length > 0 || !!typed;
  const found = useQuery({
    ...getV1NodesIpsRangesOptions({ query: { nodeset: typed, tags: tags.join(","), bmc } }),
    enabled: picking,
    placeholderData: keepPreviousData,
  });
  const ranges = picking ? (found.data ?? []) : [];

  let hint = "Type a nodeset like cpn-u25-[01-40] or part of a name, pick tags, or both.";
  if (found.isError) hint = found.error.detail ?? "Couldn't find those nodes.";
  else if (picking && found.data)
    hint = ranges.length
      ? `${plural(ranges.length, "range")}.`
      : "No nodes match, or none have addresses.";

  return (
    <div className="grid gap-4">
      <div className="grid gap-1.5">
        <label htmlFor={`${id}-nodeset`} className="text-sm font-medium">
          Nodeset
        </label>
        <Input
          id={`${id}-nodeset`}
          spellCheck={false}
          autoComplete="off"
          placeholder="cpn-u25-[01-40]"
          aria-describedby={`${id}-nodes-hint`}
          className="font-mono"
          value={nodeset}
          onChange={(e) => setNodeset(e.target.value)}
        />
      </div>
      <div className="grid gap-1.5">
        <label htmlFor={`${id}-tags`} className="text-sm font-medium">
          Tags
        </label>
        <TagCombobox
          id={`${id}-tags`}
          label="Tags"
          placeholder="rack=u25"
          value={tags}
          onValueChange={setTags}
        />
      </div>
      <p
        id={`${id}-nodes-hint`}
        aria-live="polite"
        className={cn(
          "-mt-2 text-xs",
          found.isError ? "text-destructive" : "text-muted-foreground",
        )}
      >
        {picking && found.isFetching ? (
          <span className="inline-flex items-center gap-1.5">
            <LoaderCircle className="size-3 animate-spin" />
            Finding nodes
          </span>
        ) : (
          hint
        )}
      </p>

      {ranges.length > 0 && (
        <ul aria-label="Ranges of those nodes" className="-mx-1 grid gap-0.5 px-1">
          {ranges.map((r) => {
            const bmcPorts = r.bmc_ports ?? 0;
            const kind = bmc ? bmcPorts : (r.ports ?? 0) - bmcPorts;
            return (
              <li key={r.range}>
                <button
                  type="button"
                  onClick={() => r.range && onPick(r.range)}
                  aria-current={r.range === current || undefined}
                  className={cn(
                    "grid w-full grid-cols-[1fr_auto] items-baseline gap-x-3 rounded-md px-2.5 py-1.5 text-left transition-colors hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none aria-current:bg-muted/60",
                    kind === 0 && "text-muted-foreground",
                  )}
                >
                  <span className="font-mono text-sm">{r.range}</span>
                  <span className="text-xs text-muted-foreground tabular-nums">
                    /{r.prefix_len}
                  </span>
                </button>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
