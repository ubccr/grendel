import { type Block, type NeighborKind, type Slot } from "@/lib/switch-ports";
import { cn } from "@/lib/utils";
import type { CSSProperties } from "react";

export const kindColor: Record<NeighborKind, string> = {
  node: "var(--node-ready)",
  other: "var(--node-amber)",
};

export const kindLabel: Record<NeighborKind, string> = {
  node: "Node in Grendel",
  other: "Not in Grendel",
};

export function Led({
  kind,
  color: given,
  dim,
  className,
}: {
  kind?: NeighborKind;
  color?: string;
  dim?: boolean;
  className?: string;
}) {
  const color = given ?? (kind ? kindColor[kind] : undefined);
  return (
    <span
      aria-hidden
      className={cn(
        // no opacity transition or fade-in: Chrome animates opacity on a layer of its own, which draws the glow larger and off by a fraction of a pixel until the animation ends and the layer goes, so the lights flicker and the bottom row seems to shift
        "inline-block shrink-0 rounded-[1px]",
        !color && "bg-zinc-400/50 dark:bg-zinc-700",
        dim && "opacity-25",
        className,
      )}
      style={
        color
          ? {
              backgroundColor: color,
              boxShadow: `0 0 5px 0 color-mix(in oklab, ${color} 80%, transparent)`,
            }
          : undefined
      }
    />
  );
}

// The bottom row is drawn upside down, as switches mount it
function Jack({ plugged, flipped }: { plugged: boolean; flipped: boolean }) {
  const w = 24;
  const h = 17;
  return (
    <svg
      aria-hidden
      width={w}
      height={h}
      viewBox={`0 0 ${w} ${h}`}
      className={cn("block", flipped && "-scale-y-100")}
    >
      <rect width={w} height={h} rx={1.5} className="fill-zinc-400 dark:fill-zinc-600" />
      <rect
        x={w / 2 - 3}
        y={0}
        width={6}
        height={1.5}
        className="fill-zinc-300 dark:fill-zinc-700"
      />
      <rect
        x={1.5}
        y={2}
        width={w - 3}
        height={h - 3.5}
        rx={0.75}
        className="fill-zinc-900 dark:fill-zinc-950"
      />
      {plugged ? (
        <>
          <rect
            x={2.5}
            y={3}
            width={w - 5}
            height={h - 5.5}
            rx={0.75}
            className="fill-zinc-400 dark:fill-zinc-500"
          />
          {[w / 2 - 5.5, w / 2 + 0.5].map((x) => (
            <rect
              key={x}
              x={x}
              y={5.5}
              width={5}
              height={5}
              rx={1}
              className="fill-zinc-700 dark:fill-zinc-800"
            />
          ))}
        </>
      ) : (
        <rect
          x={2.5}
          y={h - 3}
          width={w - 5}
          height={1}
          className="fill-zinc-700 dark:fill-zinc-800"
        />
      )}
    </svg>
  );
}

export type Light = {
  label: string;
  plugged: boolean;
  color?: string;
  dim?: boolean;
  // a cage that can't be picked is drawn the same but ignores clicks, like an empty port
  inert?: boolean;
  // a ring in this color marks the cage as wanting attention, which outlasts the light being dimmed
  ring?: string;
};

function Cage({
  light,
  bottom,
  selected,
  onSelect,
}: {
  light: Light;
  bottom: boolean;
  selected: boolean;
  onSelect: () => void;
}) {
  const style: CSSProperties | undefined = light.ring
    ? { boxShadow: `inset 0 0 0 1.5px ${light.ring}` }
    : undefined;

  // each light sits on the cage's outer side, above the top row and below the bottom one, so it reads as that cage's and not its neighbor's
  const body = (
    <>
      <Led color={light.color} dim={light.dim} className="h-[3px] w-2.5" />
      <Jack plugged={light.plugged} flipped={bottom} />
    </>
  );
  const shape = cn(
    "flex items-center gap-[3px] rounded-[3px] p-[3px]",
    bottom ? "flex-col-reverse" : "flex-col",
  );

  if (light.inert)
    return (
      <span title={light.label} className={shape} style={style}>
        {body}
      </span>
    );

  return (
    <button
      type="button"
      aria-label={light.label}
      aria-pressed={selected}
      title={light.label}
      onClick={onSelect}
      className={cn(
        shape,
        "outline-none hover:bg-zinc-300 focus-visible:ring-2 focus-visible:ring-ring dark:hover:bg-zinc-700",
        selected &&
          "bg-sky-500/20 ring-2 ring-sky-500 hover:bg-sky-500/25 dark:hover:bg-sky-500/25",
      )}
      style={style}
    >
      {body}
    </button>
  );
}

export function Faceplate<T>({
  blocks,
  light,
  selected,
  onSelect,
  className,
}: {
  blocks: Block<T>[];
  light: (slot: Slot<T>) => Light;
  selected: (slot: Slot<T>) => boolean;
  onSelect: (slot: Slot<T>) => void;
  className?: string;
}) {
  return (
    <div className={cn("max-w-full overflow-x-auto rounded-lg", className)}>
      <div className="flex w-max items-stretch rounded-lg border border-zinc-300 bg-zinc-200 dark:border-zinc-700 dark:bg-zinc-800">
        <Ear />
        <div className="flex items-center gap-5 px-3 py-2">
          {blocks.map((block) => {
            const columns: Slot<T>[][] = [];
            for (let i = 0; i < block.slots.length; i += 2)
              columns.push(block.slots.slice(i, i + 2));
            return (
              <div key={block.slots[0].number} className="flex gap-1">
                {columns.map((pair) => (
                  <div key={pair[0].number} className="grid gap-1">
                    {pair.map((slot, row) => {
                      const number = (
                        <span
                          aria-hidden
                          className="text-center text-[10px] leading-none text-zinc-500 tabular-nums dark:text-zinc-400"
                        >
                          {slot.number}
                        </span>
                      );
                      return (
                        <div key={slot.number} className="grid justify-items-center gap-0.5">
                          {row === 0 && number}
                          <Cage
                            light={light(slot)}
                            bottom={row === 1}
                            selected={selected(slot)}
                            onSelect={() => onSelect(slot)}
                          />
                          {row === 1 && number}
                        </div>
                      );
                    })}
                  </div>
                ))}
              </div>
            );
          })}
        </div>
        <Ear />
      </div>
    </div>
  );
}

function Ear() {
  return (
    <div
      aria-hidden
      className="flex w-5 flex-col items-center justify-between border-x border-zinc-300/80 py-3 first:border-l-0 last:border-r-0 dark:border-zinc-700/60"
    >
      {[0, 1].map((i) => (
        <span
          key={i}
          className="h-2 w-[5px] rounded-full bg-zinc-400/70 shadow-[inset_0_1px_1px_rgba(0,0,0,0.35)] dark:bg-zinc-950"
        />
      ))}
    </div>
  );
}
