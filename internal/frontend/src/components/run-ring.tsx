import { cn } from "@/lib/utils";

const STROKE = 3.5;
// the ring is drawn in a 16 unit box and sized by its class, so it scales with the text beside it
const R = (16 - STROKE) / 2;
// lengths along the ring are shares of it, since each arc's pathLength is 1
// MIN_SHARE is the least a result gets, so one failure among hundreds of nodes still shows
const MIN_SHARE = 0.12;
// GAP parts the arcs, so a mixed result reads as its parts
const GAP = 0.03;

// Arc is a stretch of the ring, share long, starting offset along it from the top
function Arc({
  share = 1,
  offset = 0,
  className,
}: {
  share?: number;
  offset?: number;
  className: string;
}) {
  return (
    <circle
      cx={8}
      cy={8}
      r={R}
      fill="none"
      strokeWidth={STROKE}
      pathLength={1}
      strokeDasharray={`${share} 1`}
      strokeDashoffset={-offset}
      className={className}
    />
  );
}

// ringShares turns counts into each one's share of the ring. A count that is not zero gets at least MIN_SHARE, taken from the larger shares in proportion to what they have to spare
export function ringShares(counts: number[]): number[] {
  const total = counts.reduce((a, b) => a + b, 0);
  if (!total) return counts.map(() => 0);
  const shares = counts.map((n) => (n > 0 ? Math.max(MIN_SHARE, n / total) : 0));
  const over = shares.reduce((a, b) => a + b, 0) - 1;
  const spare = shares.reduce((a, s) => a + Math.max(0, s - MIN_SHARE), 0);
  if (over <= 0 || spare <= 0) return shares;
  return shares.map((s) => (s > MIN_SHARE ? s - (over * (s - MIN_SHARE)) / spare : s));
}

// A run that touched no nodes draws an empty ring. An action still under way also has a pulsing blue share for the nodes still going and an amber one for those that stalled
export function RunRing({
  succeeded,
  failed,
  going = 0,
  stalled = 0,
  className,
}: {
  succeeded: number;
  failed: number;
  going?: number;
  stalled?: number;
  className?: string;
}) {
  const total = succeeded + failed + going + stalled;
  const colors = [
    "stroke-(--node-ready)",
    "stroke-destructive",
    "stroke-(--node-blue) motion-safe:animate-pulse",
    "stroke-(--node-amber)",
  ];
  const shares = ringShares([succeeded, failed, going, stalled]);
  // with more than one present, each arc gives up half the gap at either end
  const gap = shares.filter((s) => s > 0).length > 1 ? GAP : 0;
  const arcs = shares.map((share, i) => ({
    share: share - gap,
    offset: shares.slice(0, i).reduce((a, b) => a + b, 0) + gap / 2,
    className: colors[i],
  }));
  const label =
    going + stalled > 0
      ? `${succeeded} of ${total} nodes done${failed ? `, ${failed} failed` : ""}`
      : `${succeeded} of ${total} nodes succeeded`;

  return (
    <svg
      viewBox="0 0 16 16"
      role="img"
      aria-label={label}
      className={cn("size-3.5 shrink-0 -rotate-90", className)}
    >
      {total === 0 && <Arc className="stroke-muted-foreground/30" />}
      {arcs.map((arc) => arc.share > 0 && <Arc key={arc.className} {...arc} />)}
    </svg>
  );
}
