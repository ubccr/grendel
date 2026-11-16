import type { Event } from "@/client";
import { getV1GrendelEventsOptions } from "@/client/@tanstack/react-query.gen";
import { RunRing } from "@/components/run-ring";
import { useHoverPin } from "@/hooks/use-hover-pin";
import { useStoredValue } from "@/hooks/use-stored-value";
import { useUser } from "@/hooks/user-provider";
import {
  eventParts,
  eventTally,
  eventText,
  eventTime,
  isRunning,
  severityDot,
  type EventPart,
} from "@/lib/events";
import { islandSpring } from "@/lib/motion";
import { cn, relativeTime, severityOf } from "@/lib/utils";
import { useQuery } from "@tanstack/react-query";
import { Link, useMatchRoute } from "@tanstack/react-router";
import { AnimatePresence, motion, useIsPresent } from "motion/react";
import { useEffect, useId, useReducer, useState } from "react";

// how often the events are fetched, while one is running and while none is; a change made in this browser also refetches them at once, through the query client's mutation cache in main.tsx
const LIVE_POLL_MS = 5_000;
const POLL_MS = 30_000;
// SEEN_KEY stores when the events were last looked at, so every tab shares it
const SEEN_KEY = "grendel-events-seen";
// RECENT is how many events the opened island shows, the one in its row and the rest under it
const RECENT = 5;
// how long a new event holds the island's row while another is running, before the running one takes it back
const EVENT_HOLD_MS = 6_000;
// how long an event that was running keeps the row once it ends, so how it went is seen
const ENDED_HOLD_MS = 10_000;
// the actor of an event nobody asked for, such as nodes network booting. Match model.EventActorSystem
const SYSTEM = "system";

// useHeld is true for ms after key changes to a new value, and not for the first value it gets
function useHeld(key: string | undefined, ms: number) {
  const [last, setLast] = useState(key);
  const [held, setHeld] = useState<string>();
  if (key !== last) {
    setLast(key);
    if (last !== undefined) setHeld(key);
  }
  useEffect(() => {
    if (held === undefined) return;
    const id = setTimeout(() => setHeld(undefined), ms);
    return () => clearTimeout(id);
  }, [held, ms]);
  return held !== undefined;
}

// LatestEvent is an island at the top right of the page, across from the breadcrumbs, holding the newest event, who made it, and how many changes other people made since the events page was last open. A new event slides in, so a change made elsewhere is noticed without a toast taking over the page. Hovering or pressing it opens it downward over the page to show the events before it, with a link to the events page at the bottom. An event still running, such as nodes provisioning, holds the row while it runs, with a ring filling as its nodes finish; a new event takes the row for a few seconds and gives it back. Opened, the running event stays in the row with its nodes by stage under it
export function LatestEvent() {
  const { user } = useUser();
  const { data } = useQuery({
    ...getV1GrendelEventsOptions(),
    enabled: !!user,
    refetchInterval: (q) => (q.state.data?.some(isRunning) ? LIVE_POLL_MS : POLL_MS),
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

  const events = data ?? [];
  const latest = events[0];
  // the newest running event is the one followed, and the last one followed is remembered for when it ends
  const live = events.find(isRunning);
  const [followed, setFollowed] = useState(live?.id);
  if (live?.id !== undefined && live.id !== followed) setFollowed(live.id);
  const justEnded = useHeld(live ? "live" : "idle", ENDED_HOLD_MS) && !live;
  const tracked = live ?? (justEnded ? events.find((e) => e.id === followed) : undefined);
  const eventHeld = useHeld(latest?.id, EVENT_HOLD_MS);

  // an event counts from when it ended, so one that ran for a while is unseen when it finishes
  const latestMs = events.reduce((ms, e) => Math.max(ms, Date.parse(eventTime(e)) || 0), 0);

  // the events count as seen while the events page is open, and the first time this browser sees any, so a new browser does not start with every event unseen
  useEffect(() => {
    if (latestMs === 0 || (!onEvents && seen !== undefined) || latestMs === seen) return;
    writeSeen(String(latestMs));
  }, [latestMs, onEvents, seen, writeSeen]);

  if (!user || !latest) return null;

  // the followed event has the row unless a newer event is holding it; opened it always does, so the row does not change under the pointer
  const shown = tracked && (open || !eventHeld || latest.id === tracked.id) ? tracked : latest;

  // unseen are other people's changes since the events were last seen; your own you already know about
  const isUnseen = (e: Event) =>
    seen !== undefined && e.actor !== user.username && Date.parse(eventTime(e)) > seen;
  const unseen = events.filter(isUnseen).length;
  // the dot pulses only for a warning or error someone else made that has not been seen
  const isAlarm = (e: Event) => {
    const severity = severityOf(e.severity);
    return isUnseen(e) && (severity === "warning" || severity === "error");
  };
  // an alarm behind the event in the row colours the count instead
  const hidden = shown.id !== latest.id && isAlarm(latest) ? latest : undefined;
  const parts = tracked && shown.id === tracked.id ? eventParts(tracked) : [];

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
            aria-label={`${isRunning(shown) ? "Running" : "Latest event"}: ${shown.severity} ${eventText(shown)} by ${shown.actor}${unseen ? `, ${unseen} unseen` : ""}. Show recent events`}
            className="flex h-10 w-full min-w-0 cursor-pointer items-center gap-2 px-3 text-left text-sm transition-colors hover:bg-sidebar-accent focus-visible:bg-sidebar-accent focus-visible:outline-none"
          >
            {/* the event slides in under its key, so a new one replaces the last rather than the text just changing */}
            <AnimatePresence mode="popLayout" initial={false}>
              <motion.span
                key={shown.id}
                initial={{ opacity: 0, y: -10 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: 10 }}
                className="flex min-w-0 flex-1 items-center gap-2"
              >
                <EventLine event={shown} ping={isAlarm(shown)} hideText={!open} />
              </motion.span>
            </AnimatePresence>
            {unseen > 0 && (
              <span
                className={cn(
                  "shrink-0 rounded-full px-1.5 text-xs leading-5 font-medium tabular-nums",
                  hidden
                    ? cn("text-white", severityDot(hidden.severity))
                    : "bg-primary text-primary-foreground",
                )}
              >
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
              <OtherEvents
                id={listId}
                parts={parts}
                events={events.filter((e) => e.id !== shown.id).slice(0, RECENT - 1)}
                isUnseen={isUnseen}
                total={events.length}
              />
            </motion.div>
          )}
        </AnimatePresence>
      </motion.div>
    </div>
  );
}

// OtherEvents are the latest events besides the one in the island's row, under it in the opened island, each on one line laid out the same. Over them are the nodes of the row's event by stage, when that event is one being followed
function OtherEvents({
  id,
  parts,
  events,
  isUnseen,
  total,
}: {
  id: string;
  parts: EventPart[];
  events: Event[];
  isUnseen: (e: Event) => boolean;
  total: number;
}) {
  // on its way out these rows stop counting towards the island's width, so the island narrows back to the row in the same motion as it shrinks
  const present = useIsPresent();
  return (
    <div
      id={id}
      className={cn(
        "grid border-t border-sidebar-border text-sm",
        !present && "contain-inline-size",
      )}
    >
      {parts.length > 0 && <Stages parts={parts} />}
      {events.length > 0 && (
        <ol
          aria-label="Other recent events"
          className="grid divide-y divide-sidebar-border border-t border-sidebar-border first:border-t-0"
        >
          {events.map((e) => (
            <li key={e.id} className="flex h-10 min-w-0 items-center gap-2 px-3">
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

// Stages are an event's nodes by how far they got: a bar with a stretch for each stage, the furthest along first, so it fills from the left as the nodes finish, with the count at each stage under it. A count whose stage the nodes list can filter on is a link to that list
function Stages({ parts }: { parts: EventPart[] }) {
  const item = "flex items-center gap-1.5 whitespace-nowrap";
  return (
    <div className="grid gap-2 px-3 py-2.5">
      <div aria-hidden className="flex h-1.5 gap-0.5 overflow-hidden rounded-full">
        {parts.map((p) => (
          <span
            key={p.key}
            style={{ flexGrow: p.count }}
            className={cn("min-w-1.5 basis-0 transition-[flex-grow] duration-500", p.color)}
          />
        ))}
      </div>
      <ul aria-label="Nodes by stage" className="flex flex-wrap gap-x-3 gap-y-1 text-xs">
        {parts.map((p) => {
          const count = (
            <>
              <span aria-hidden className={cn("size-1.5 rounded-full", p.color)} />
              <span className="text-muted-foreground group-hover:text-sidebar-foreground group-focus-visible:text-sidebar-foreground">
                <span className="text-sidebar-foreground tabular-nums">{p.count}</span> {p.label}
              </span>
            </>
          );
          return (
            <li key={p.key}>
              {p.state ? (
                <Link
                  to="/nodes"
                  search={{ provision_status: p.state }}
                  title={`List the nodes whose last provision is ${p.label}`}
                  className={cn(
                    item,
                    "group underline-offset-4 hover:underline focus-visible:underline focus-visible:outline-none",
                  )}
                >
                  {count}
                </Link>
              ) : (
                <span className={item}>{count}</span>
              )}
            </li>
          );
        })}
      </ul>
    </div>
  );
}

// EventLine is one event on one line: how it went, who made it, what they did and how long ago, the same for the event in the island's row and the ones under it. A fresh event, one other people made since the events page was last open, is set in bold. hideText leaves just the dot, for the closed island on a phone
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
  const tally = eventTally(event);
  const running = isRunning(event);
  const time = relativeTime(eventTime(event), { narrow: true });
  const text = eventText(event);
  return (
    <>
      {/* an event with nodes shows how they went as a ring in place of the severity dot, which it says more than; any other event keeps the dot, with its severity spelled out for screen readers */}
      <span className="relative inline-flex size-3.5 shrink-0 items-center justify-center">
        {ping && (
          <span
            aria-hidden
            className={cn(
              "absolute inset-0 rounded-full opacity-75 motion-safe:animate-ping",
              severityDot(event.severity),
            )}
          />
        )}
        {tally.total > 0 ? (
          <RunRing {...tally} className="relative" />
        ) : (
          <>
            <span
              aria-hidden
              className={cn("relative size-2 rounded-full", severityDot(event.severity))}
            />
            <span className="sr-only">{event.severity}</span>
          </>
        )}
      </span>
      <span
        className={cn("min-w-0 flex-1 items-center gap-2", hideText ? "hidden sm:flex" : "flex")}
      >
        {/* what nodes did on their own has nobody to name */}
        {event.actor !== SYSTEM && (
          <span className="shrink-0 text-muted-foreground">{event.actor}</span>
        )}
        <span className={cn("truncate", fresh && "font-medium")} title={text}>
          {text}
        </span>
        {tally.failed > 0 && running && (
          <span className="shrink-0 text-destructive">{tally.failed} failed</span>
        )}
        {/* the count shows while an event runs, and after only when something failed; a clean run is said by its full green ring */}
        {(running || tally.failed > 0) && tally.total > 0 && (
          <span className="shrink-0 text-muted-foreground tabular-nums">
            {tally.succeeded}/{tally.total}
          </span>
        )}
        <span
          className="ml-auto shrink-0 text-muted-foreground tabular-nums"
          title={running ? `Started ${time.title}` : time.title}
        >
          {time.text}
        </span>
      </span>
    </>
  );
}
