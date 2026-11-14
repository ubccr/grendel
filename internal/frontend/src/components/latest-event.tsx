import type { Event } from "@/client";
import { getV1GrendelEventsOptions } from "@/client/@tanstack/react-query.gen";
import { RunRing } from "@/components/run-ring";
import { useHoverPin } from "@/hooks/use-hover-pin";
import { useStoredValue } from "@/hooks/use-stored-value";
import { useUser } from "@/hooks/user-provider";
import { parseRun, severityDot } from "@/lib/events";
import { islandSpring } from "@/lib/motion";
import { cn, relativeTime, severityOf } from "@/lib/utils";
import { useQuery } from "@tanstack/react-query";
import { Link, useMatchRoute } from "@tanstack/react-router";
import { AnimatePresence, motion, useIsPresent } from "motion/react";
import { useEffect, useId, useReducer } from "react";

// how often the latest event is fetched; a change made in this browser also refetches it at once, through the query client's mutation cache in main.tsx
const POLL_MS = 30_000;
// SEEN_KEY stores when the events were last looked at, so every tab shares it
const SEEN_KEY = "grendel-events-seen";
// RECENT is how many events the opened island shows, the latest in its own row and the rest under it
const RECENT = 5;

// LatestEvent is an island at the top right of the page, across from the breadcrumbs, holding the newest event, who made it, and how many changes other people made since the events page was last open. A new event slides in, so a change made elsewhere is noticed without a toast taking over the page. Hovering or pressing it opens it downward over the page to show the events before it, from the same query the events page reads, with a link to the events page at the bottom
export function LatestEvent() {
  const { user } = useUser();
  const { data } = useQuery({
    ...getV1GrendelEventsOptions(),
    enabled: !!user,
    refetchInterval: POLL_MS,
  });
  const onEvents = !!useMatchRoute()({ to: "/events" });
  const [storedSeen, writeSeen] = useStoredValue(SEEN_KEY);
  const seen = Number(storedSeen) || undefined;
  const { open, containerProps, triggerProps } = useHoverPin<HTMLDivElement, HTMLButtonElement>();
  const listId = useId();
  // the relative time is redrawn each minute, since a poll that finds nothing new leaves the query's data, and so this, untouched
  const [, tick] = useReducer((n: number) => n + 1, 0);
  useEffect(() => {
    const id = setInterval(tick, 60_000);
    return () => clearInterval(id);
  }, []);

  const latest = data?.[0];
  const latestMs = latest?.Time ? Date.parse(latest.Time) : undefined;

  // the events count as seen while the events page is open, and the first time this browser sees any, so a new browser does not start with every event unseen
  useEffect(() => {
    if (latestMs === undefined || (!onEvents && seen !== undefined) || latestMs === seen) return;
    writeSeen(String(latestMs));
  }, [latestMs, onEvents, seen, writeSeen]);

  if (!user || !latest) return null;

  // unseen are other people's changes since the events were last seen; your own you already know about
  const isUnseen = (e: Event) =>
    seen !== undefined && e.User !== user.username && Date.parse(e.Time ?? "") > seen;
  const unseen = data.filter(isUnseen).length;
  const { message } = parseRun(latest.Message ?? "");
  // the dot pulses only for a warning or error someone else made that has not been seen
  const severity = severityOf(latest.Severity);
  const alarm = isUnseen(latest) && (severity === "warning" || severity === "error");

  return (
    // the island holds one row's place in the header, and opened it grows down over the page rather than pushing it; on a phone it grows leftward from the right edge, past the space it holds
    <div className="ml-auto flex h-10 max-w-[min(32rem,50%)] min-w-0 justify-end">
      <motion.div
        layout
        transition={islandSpring}
        style={{ borderRadius: 14 }}
        {...containerProps}
        className={cn(
          "relative z-30 flex h-fit min-w-0 flex-col overflow-hidden bg-sidebar text-sidebar-foreground shadow-float ring-1 ring-sidebar-border",
          // opened, it widens to fit the longest event up to 32rem, and on a phone, where closed it is just a dot, to the screen
          open
            ? "max-w-[min(32rem,calc(100vw-1.5rem))] shrink-0 max-sm:w-[calc(100vw-1.5rem)]"
            : "max-w-full",
        )}
      >
        <motion.div layout="position" className="min-w-0">
          {/* pressing the row pins the island open, or closes it when it already is */}
          <button
            type="button"
            {...triggerProps}
            aria-controls={listId}
            aria-label={`Latest event: ${latest.Severity} ${message} by ${latest.User}${unseen ? `, ${unseen} unseen` : ""}. Show recent events`}
            className="flex h-10 w-full min-w-0 cursor-pointer items-center gap-2 px-3 text-left text-sm transition-colors hover:bg-sidebar-accent focus-visible:bg-sidebar-accent focus-visible:outline-none"
          >
            {/* the event slides in under its key, so a new one replaces the last rather than the text just changing */}
            <AnimatePresence mode="popLayout" initial={false}>
              <motion.span
                key={latest.Time}
                initial={{ opacity: 0, y: -10 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: 10 }}
                className="flex min-w-0 flex-1 items-center gap-2"
              >
                <EventLine event={latest} ping={alarm} hideText={!open} />
              </motion.span>
            </AnimatePresence>
            {unseen > 0 && (
              <span className="shrink-0 rounded-full bg-primary px-1.5 text-xs leading-5 font-medium text-primary-foreground tabular-nums">
                {unseen}
              </span>
            )}
          </button>
        </motion.div>
        <AnimatePresence initial={false}>
          {open && (
            // the rows grow from nothing and shrink back to it, so the island opens and closes as one motion; animating their height rather than the island's layout means closing animates too, since removing them would not re-render the island
            <motion.div
              layout="position"
              initial={{ height: 0, opacity: 0 }}
              animate={{ height: "auto", opacity: 1 }}
              exit={{ height: 0, opacity: 0 }}
              transition={islandSpring}
            >
              <EarlierEvents
                id={listId}
                events={data.slice(1, RECENT)}
                isUnseen={isUnseen}
                total={data.length}
              />
            </motion.div>
          )}
        </AnimatePresence>
      </motion.div>
    </div>
  );
}

// EarlierEvents are the events before the latest, under it in the opened island, each on one line laid out the same as the latest's
function EarlierEvents({
  id,
  events,
  isUnseen,
  total,
}: {
  id: string;
  events: Event[];
  isUnseen: (e: Event) => boolean;
  total: number;
}) {
  // on its way out these rows stop counting towards the island's width, so the island narrows back to the latest event's row in the same motion as it shrinks
  const present = useIsPresent();
  return (
    <div
      id={id}
      className={cn(
        "grid border-t border-sidebar-border text-sm",
        !present && "contain-inline-size",
      )}
    >
      {events.length > 0 && (
        <ol aria-label="Earlier events" className="grid divide-y divide-sidebar-border">
          {events.map((e, i) => (
            <li key={e.Time ?? i} className="flex h-10 min-w-0 items-center gap-2 px-3">
              <EventLine event={e} fresh={isUnseen(e)} />
            </li>
          ))}
        </ol>
      )}
      <Link
        to="/events"
        className="border-t border-sidebar-border px-3 py-2 text-xs text-muted-foreground transition-colors first:border-t-0 hover:bg-sidebar-accent hover:text-foreground focus-visible:bg-sidebar-accent focus-visible:outline-none"
      >
        See all {total} events
      </Link>
    </div>
  );
}

// EventLine is one event on one line: its severity, who made it, what they did, how it went and how long ago, the same for the latest event and the ones under it. A fresh event, one other people made since the events page was last open, is set in bold. hideText leaves just the dot, for the closed island on a phone
function EventLine({
  event,
  ping,
  fresh,
  hideText,
}: {
  event: Event;
  ping?: boolean;
  fresh?: boolean;
  hideText?: boolean;
}) {
  const run = parseRun(event.Message ?? "");
  const time = relativeTime(event.Time ?? "", { narrow: true });
  return (
    <>
      {/* a bmc action shows how its nodes went as a ring in place of the severity dot, which it says more than; any other event keeps the dot, with its severity spelled out for screen readers */}
      <span className="relative inline-flex size-3.5 shrink-0 items-center justify-center">
        {ping && (
          <span
            aria-hidden
            className={cn(
              "absolute inset-0 rounded-full opacity-75 motion-safe:animate-ping",
              severityDot(event.Severity),
            )}
          />
        )}
        {run.total > 0 ? (
          <RunRing succeeded={run.succeeded} failed={run.failed} className="relative" />
        ) : (
          <>
            <span
              aria-hidden
              className={cn("relative size-2 rounded-full", severityDot(event.Severity))}
            />
            <span className="sr-only">{event.Severity}</span>
          </>
        )}
      </span>
      <span
        className={cn("min-w-0 flex-1 items-center gap-2", hideText ? "hidden sm:flex" : "flex")}
      >
        <span className="shrink-0 text-muted-foreground">{event.User}</span>
        <span className={cn("truncate", fresh && "font-medium")} title={event.Message}>
          {run.message}
        </span>
        {/* the count only shows when something failed; a clean run is said by its full green ring */}
        {run.failed > 0 && (
          <span className="shrink-0 text-muted-foreground tabular-nums">
            {run.succeeded}/{run.total}
          </span>
        )}
        <span className="ml-auto shrink-0 text-muted-foreground tabular-nums" title={time.title}>
          {time.text}
        </span>
      </span>
    </>
  );
}
