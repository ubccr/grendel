import type { Event } from "@/client";
import { RunRing } from "@/components/run-ring";
import { useNow } from "@/hooks/use-now";
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
import { eventsOptions, followEvents } from "@/lib/queries";
import { cn, relativeTime, severityOf } from "@/lib/utils";
import { Popover } from "@base-ui/react/popover";
import { useQuery } from "@tanstack/react-query";
import { Link, useMatchRoute } from "@tanstack/react-router";
import { AnimatePresence, motion } from "motion/react";
import { useEffect, useState } from "react";

// SEEN_KEY stores when the events were last looked at, so every tab shares it
const SEEN_KEY = "grendel-events-seen";
// RECENT is how many events the island lists when opened
const RECENT = 5;
// the actor of an event nobody asked for, such as nodes network booting. Match model.EventActorSystem
const SYSTEM = "system";

const noEvents: Event[] = [];

const when = (e: Event) => Date.parse(eventTime(e)) || 0;

// An event still running, such as nodes provisioning, holds the island's row while it runs
export function LatestEvent() {
  const { user } = useUser();
  const { data } = useQuery({
    ...eventsOptions(),
    enabled: !!user,
    refetchInterval: followEvents,
  });
  const onEvents = !!useMatchRoute()({ to: "/events" });
  const [storedSeen, writeSeen] = useStoredValue(SEEN_KEY);
  const seen = Number(storedSeen) || undefined;
  const [open, setOpen] = useState(false);
  // the relative time is redrawn each minute, since a poll that finds nothing new leaves the query's data, and so this, untouched
  useNow(60_000);

  // newest first by when each counts as having happened, the order the events page lists them in
  const events = data ?? noEvents;
  const latest = events[0];
  const latestMs = latest ? when(latest) : 0;

  // the events count as seen while the events page is open, and the first time this browser sees any, so a new browser does not start with every event unseen
  useEffect(() => {
    if (latestMs === 0 || (!onEvents && seen !== undefined) || latestMs === seen) return;
    writeSeen(String(latestMs));
  }, [latestMs, onEvents, seen, writeSeen]);

  if (!user || !latest) return null;

  const live = events.find(isRunning);
  const shown = live ?? latest;

  // unseen are other people's changes since the events were last seen; your own you already know about
  const isUnseen = (e: Event) => seen !== undefined && e.actor !== user.username && when(e) > seen;
  const unseen = events.filter(isUnseen).length;
  // the dot pulses only for a warning or error someone else made that has not been seen
  const isAlarm = (e: Event) => {
    const severity = severityOf(e.severity);
    return isUnseen(e) && (severity === "warning" || severity === "error");
  };
  // an alarm behind the running event in the row colours the count instead
  const hidden = shown.id !== latest.id && isAlarm(latest) ? latest : undefined;

  return (
    <Popover.Root open={open} onOpenChange={setOpen}>
      {/* the island holds one row's place in the header, and narrows or widens with the event in it */}
      <div className="ml-auto flex h-10 max-w-[min(32rem,50%)] min-w-0 justify-end">
        <motion.div
          layout
          transition={islandSpring}
          style={{ borderRadius: 14 }}
          className={cn(island, "relative z-10 flex h-10 max-w-full min-w-0 overflow-hidden")}
        >
          <motion.div layout="position" className="min-w-0">
            <Popover.Trigger
              openOnHover
              delay={150}
              closeDelay={200}
              aria-label={`${isRunning(shown) ? "Running" : "Latest event"}: ${shown.severity} ${eventText(shown)} by ${shown.actor}${unseen ? `, ${unseen} unseen` : ""}. Show recent events`}
              className="flex h-10 w-full min-w-0 cursor-pointer items-center gap-2 px-3 text-left text-sm transition-colors hover:bg-sidebar-accent focus-visible:bg-sidebar-accent focus-visible:outline-none data-popup-open:bg-sidebar-accent"
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
                  <EventLine event={shown} ping={isAlarm(shown)} hideText />
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
            </Popover.Trigger>
          </motion.div>
        </motion.div>
      </div>
      {/* the list opens under the island's right edge on the same surface, so the two read as one piece opened out */}
      <Popover.Portal>
        <Popover.Positioner align="end" sideOffset={6} className="isolate z-50">
          <Popover.Popup
            className={cn(
              island,
              "w-[min(32rem,calc(100vw-1.5rem))] origin-(--transform-origin) overflow-hidden rounded-[14px] outline-hidden duration-150 data-open:animate-in data-open:fade-in-0 data-open:slide-in-from-top-2 data-closed:animate-out data-closed:fade-out-0 data-closed:slide-out-to-top-2",
            )}
          >
            <RecentEvents
              shown={shown}
              parts={live ? eventParts(live) : []}
              events={events.filter((e) => e.id !== shown.id).slice(0, RECENT - 1)}
              isUnseen={isUnseen}
              onLeave={() => setOpen(false)}
            />
          </Popover.Popup>
        </Popover.Positioner>
      </Popover.Portal>
    </Popover.Root>
  );
}

// island is the surface the header's islands share with the sidebar
const island = "bg-sidebar text-sidebar-foreground shadow-float ring-1 ring-sidebar-border";

// On a phone the island's row is only a dot, so its event leads the list there
function RecentEvents({
  shown,
  parts,
  events,
  isUnseen,
  onLeave,
}: {
  shown: Event;
  parts: EventPart[];
  events: Event[];
  isUnseen: (e: Event) => boolean;
  // onLeave closes the island when a link in it is followed
  onLeave: () => void;
}) {
  return (
    <div className="grid divide-y divide-sidebar-border text-sm">
      <div className="flex h-10 min-w-0 items-center gap-2 px-3 sm:hidden">
        <EventLine event={shown} />
      </div>
      {parts.length > 0 && <Stages parts={parts} onLeave={onLeave} />}
      {events.length > 0 && (
        <ol aria-label="Other recent events" className="grid divide-y divide-sidebar-border">
          {events.map((e) => (
            <li key={e.id} className="flex h-10 min-w-0 items-center gap-2 px-3">
              <EventLine event={e} fresh={isUnseen(e)} />
            </li>
          ))}
        </ol>
      )}
      <Link
        to="/events"
        onClick={onLeave}
        className="px-3 py-2 text-xs text-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-foreground focus-visible:bg-sidebar-accent focus-visible:outline-none"
      >
        See all events
      </Link>
    </div>
  );
}

// parts are ordered furthest stage first, so the bar fills from the left as the nodes finish. A count whose stage the nodes list can filter on is a link to that list
function Stages({ parts, onLeave }: { parts: EventPart[]; onLeave: () => void }) {
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
                  onClick={onLeave}
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

// A fresh event, one other people made since the events page was last open, is set in bold. hideText leaves just the dot, for the island's row on a phone
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
