import type { NextAddressResponse } from "@/client";
import { getV1NodesIpsNextOptions } from "@/client/@tanstack/react-query.gen";
import { Button } from "@/components/ui/button";
import { DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { RadioGroup } from "@/components/ui/radio-group";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useDebounced } from "@/hooks/use-debounced";
import { useStoredValue } from "@/hooks/use-stored-value";
import { lastOctet, parseRange, rangeKey, rangeOf, stepRange, stripPrefix } from "@/lib/addresses";
import { cn } from "@/lib/utils";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { ChevronLeft, ChevronRight, LoaderCircle } from "lucide-react";
import { useId, useRef, useState } from "react";
import { flushSync } from "react-dom";
import type { Port } from ".";
import { NodeRanges } from "./node-ranges";
import { Option, type Choice } from "./option";
import { RangeMap } from "./range-map";

type Suggestion = NonNullable<NextAddressResponse["lowest"]>[number];

// pick is the address a suggestion list offers, its first one nothing answered at, and the ones that answered before it, which were skipped
function pick(list: Suggestion[] = []) {
  const at = list.findIndex((s) => !s.answered);
  return {
    ip: list[at]?.ip,
    skipped: list.slice(0, at < 0 ? undefined : at).map((s) => s.ip ?? ""),
  };
}

// subnetNote says which prefix length a picked address will carry, and the subnet that makes it part of when that is wider than the range searched
function subnetNote(data: NextAddressResponse) {
  if (data.subnet === data.range) return `Addresses get /${data.prefix_len}.`;
  return `Addresses get /${data.prefix_len}, as part of ${data.subnet}.`;
}

// Finder offers two addresses, the lowest free and the first after the highest used, each pinged to check nothing unknown to Grendel is there
export function Finder({
  value,
  port,
  label,
  onPick,
}: {
  value: string;
  port: Port;
  label: string;
  onPick: (ip: string) => void;
}) {
  const id = useId();
  const rangeRef = useRef<HTMLInputElement>(null);
  const [remembered, remember] = useStoredValue(rangeKey(port.bmc));
  const [tab, setTab] = useState<"range" | "nodes">("range");
  const [text, setText] = useState(() => rangeOf(value) ?? remembered ?? "");
  const [choice, setChoice] = useState<Choice>("lowest");
  // what is typed is searched once typing pauses, while a step searches at once
  const [typed, setTyped] = useDebounced(text, 300);
  const range = parseRange(typed);
  const badRange = typed.trim() !== "" && !range;

  const shown = parseRange(text);
  const previousRange = shown && stepRange(shown, -1);
  const nextRange = shown && stepRange(shown, 1);
  const go = (to?: string) => {
    if (!to) return;
    setText(to);
    setTyped(to);
  };

  const others = port.others.map(stripPrefix).filter(Boolean);
  const search = useQuery({
    ...getV1NodesIpsNextOptions({
      query: { cidr: range ?? "", count: 1, exclude: others.join(",") },
    }),
    enabled: !!range,
    staleTime: 0,
    // the last range stays in view, dimmed, while the next one is searched, so the dialog doesn't empty between them
    placeholderData: keepPreviousData,
  });
  const data = search.data;
  const stale = search.isPlaceholderData;

  const lowest = pick(data?.lowest);
  const after = pick(data?.after_highest);
  const chosen = choice === "after" ? after.ip : lowest.ip;
  const full = chosen && data ? `${chosen}/${data.prefix_len}` : undefined;
  // a range wider than a /24 can't name its addresses by last octet alone
  const wide = Number(data?.range?.split("/")[1]) < 24;
  const highest = data?.highest && (wide ? data.highest : lastOctet(data.highest));

  const use = () => {
    // what is typed is searched once typing pauses, so until then the address on show belongs to the range before it
    if (!full || !range || stale || shown !== range) return;
    remember(range);
    onPick(full);
  };

  return (
    <div className="grid gap-4">
      <DialogTitle>Find a free address for {label}</DialogTitle>

      <Tabs value={tab} onValueChange={(v) => setTab(v)}>
        <TabsList variant="line" className="-mx-1">
          <TabsTrigger value="range">Range</TabsTrigger>
          <TabsTrigger value="nodes">From nodes</TabsTrigger>
        </TabsList>
        <TabsContent value="nodes" keepMounted className="pt-2">
          <NodeRanges
            bmc={port.bmc}
            current={range}
            onPick={(r) => {
              go(r);
              // the range tab is shown before its input is focused, as a hidden panel can't take focus
              flushSync(() => setTab("range"));
              rangeRef.current?.focus();
            }}
          />
        </TabsContent>
        <TabsContent value="range" keepMounted className="grid gap-4 pt-2">
          <div className="grid gap-1.5">
            <label htmlFor={`${id}-range`} className="text-sm font-medium">
              Range
            </label>
            <div className="flex gap-1.5">
              <Button
                type="button"
                variant="outline"
                size="icon"
                disabled={!previousRange}
                onClick={() => go(previousRange)}
                aria-label={previousRange ? `Previous range, ${previousRange}` : "Previous range"}
              >
                <ChevronLeft />
              </Button>
              <Input
                ref={rangeRef}
                id={`${id}-range`}
                autoFocus
                spellCheck={false}
                autoComplete="off"
                placeholder="10.17.25.0/24"
                aria-invalid={badRange}
                aria-describedby={`${id}-range-hint`}
                className="font-mono"
                value={text}
                onChange={(e) => setText(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === "Enter") {
                    e.preventDefault();
                    use();
                  } else if (e.key === "ArrowUp" || e.key === "ArrowDown") {
                    e.preventDefault();
                    go(e.key === "ArrowUp" ? nextRange : previousRange);
                  }
                }}
              />
              <Button
                type="button"
                variant="outline"
                size="icon"
                disabled={!nextRange}
                onClick={() => go(nextRange)}
                aria-label={nextRange ? `Next range, ${nextRange}` : "Next range"}
              >
                <ChevronRight />
              </Button>
            </div>
            <p
              id={`${id}-range-hint`}
              aria-live="polite"
              className={cn("text-xs", badRange ? "text-destructive" : "text-muted-foreground")}
            >
              {badRange ? (
                "Type a range up to a /16, like 10.17.25.0/24, or an address for its /24."
              ) : search.isFetching ? (
                <span className="inline-flex items-center gap-1.5">
                  <LoaderCircle className="size-3 animate-spin" />
                  Looking through {range}
                </span>
              ) : data ? (
                subnetNote(data)
              ) : (
                "A rack's /24, or any range up to a /16. An address stands for its /24."
              )}
            </p>
          </div>

          {search.isError && (
            <p role="alert" className="text-sm text-destructive">
              {search.error.detail ?? "Couldn't search the range."}
            </p>
          )}

          {data && !search.isError && (
            <div
              className={cn(
                "grid gap-4 transition-opacity motion-reduce:transition-none",
                stale && "opacity-60",
              )}
            >
              <RangeMap data={data} others={others} chosen={chosen} />

              <RadioGroup
                value={choice}
                onValueChange={(v) => setChoice(v as Choice)}
                aria-label="Suggested address"
                className="gap-2"
              >
                <Option
                  value="lowest"
                  label="Lowest free"
                  {...lowest}
                  prefix={data.prefix_len}
                  pinged={data.pinged}
                  pingError={data.ping_error}
                  wide={wide}
                />
                <Option
                  value="after"
                  label={highest ? `After the highest used, ${highest}` : "After the highest used"}
                  {...after}
                  prefix={data.prefix_len}
                  pinged={data.pinged}
                  pingError={data.ping_error}
                  wide={wide}
                  empty="Nothing is free after it."
                />
              </RadioGroup>

              <Button
                type="button"
                disabled={!full || stale}
                onClick={use}
                className="justify-self-end"
              >
                {full ? (
                  <>
                    Use <span className="font-mono">{full}</span>
                  </>
                ) : (
                  "Use address"
                )}
              </Button>
            </div>
          )}
        </TabsContent>
      </Tabs>
    </div>
  );
}
