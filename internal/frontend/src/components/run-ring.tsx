import { cn } from "@/lib/utils";

const STROKE = 3.5;
// the ring is drawn in a 16 unit box and sized by its class, so it scales with the text beside it
const R = (16 - STROKE) / 2;
// lengths along the ring are shares of it, since each arc's pathLength is 1
// MIN_SHARE is the least a result gets, so one failure among hundreds of nodes still shows
const MIN_SHARE = 0.12;
// GAP parts the two arcs, so a partial result reads as two parts
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

// RunRing shows how a bmc action went across its nodes: green for the share that succeeded, red for the share that failed. A run that touched no nodes draws an empty ring
export function RunRing({
  succeeded,
  failed,
  className,
}: {
  succeeded: number;
  failed: number;
  className?: string;
}) {
  const total = succeeded + failed;
  const split = failed > 0 && succeeded > 0;
  const bad = !total
    ? 0
    : split
      ? Math.min(1 - MIN_SHARE, Math.max(MIN_SHARE, failed / total))
      : failed / total;
  // with both present, each arc gives up half the gap at either end
  const gap = split ? GAP : 0;

  return (
    <svg
      viewBox="0 0 16 16"
      role="img"
      aria-label={`${succeeded} of ${total} nodes succeeded`}
      className={cn("size-3.5 shrink-0 -rotate-90", className)}
    >
      {total === 0 && <Arc className="stroke-muted-foreground/30" />}
      {succeeded > 0 && (
        <Arc share={1 - bad - gap} offset={gap / 2} className="stroke-(--node-ready)" />
      )}
      {failed > 0 && (
        <Arc share={bad - gap} offset={1 - bad + gap / 2} className="stroke-destructive" />
      )}
    </svg>
  );
}
