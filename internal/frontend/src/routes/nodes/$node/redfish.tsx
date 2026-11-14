import { BmcResponseModelRedfishSystem, getV1BmcStatus } from "@/client";
import {
  getV1BmcJobsOptions,
  getV1BmcSelOptions,
  postV1BmcIndicatorMutation,
} from "@/client/@tanstack/react-query.gen";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ErrorPanel } from "@/components/error-panel";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Progress, ProgressLabel, ProgressValue } from "@/components/ui/progress";
import { Skeleton } from "@/components/ui/skeleton";
import { logSeverityVariant, toastBmcResponse } from "@/lib/bmc";
import { formatLocation } from "@/lib/identity";
import { cn, relativeTime } from "@/lib/utils";
import { useMutation, useQuery } from "@tanstack/react-query";
import { createFileRoute, Link, useRouter } from "@tanstack/react-router";
import type { ReactNode } from "react";
import { ArrowRight } from "lucide-react";
import { toast } from "sonner";

export const Route = createFileRoute("/nodes/$node/redfish")({
  staticData: { crumb: "Redfish" },
  component: RouteComponent,
  loader: ({ params: { node } }) => getV1BmcStatus({ query: { nodeset: node } }),
  staleTime: 60_000,
});

type System = NonNullable<NonNullable<BmcResponseModelRedfishSystem["results"]>[number]["data"]>;
type Dell = NonNullable<System["oem_dell"]>;

function RouteComponent() {
  const { node } = Route.useParams();
  const { data: res } = Route.useLoaderData();
  const result = res?.results?.[0];

  if (result && !result.success)
    return <ErrorPanel title="Could not read the BMC" message={result.message} />;

  const system = result?.data;
  // a BMC that is not a Dell iDRAC sends oem_dell as null, which the generated type leaves out
  const dell = system?.oem_dell ?? undefined;

  return (
    <div className="grid grid-cols-1 gap-8">
      <FrontPanel node={node} system={system} dell={dell} />

      <div className="grid grid-cols-1 gap-8 lg:grid-cols-[minmax(0,3fr)_minmax(0,2fr)]">
        <SpecSheet system={system} dell={dell} />
        <Boot system={system} />
      </div>

      <div className="grid grid-cols-1 gap-8 border-t pt-6 lg:grid-cols-2">
        <RecentEvents node={node} />
        <JobQueue node={node} />
      </div>
    </div>
  );
}

// a Tone is what an LED shows: changing is a server powering on or off, locate is the blue locate light
type Tone = "ok" | "changing" | "warning" | "critical" | "off" | "unknown" | "locate";

// toneOf reads a Redfish health or power state as the light an LED would show; power off is an unlit LED, not a fault
function toneOf(value?: string): Tone {
  switch (value) {
    case "OK":
    case "On":
      return "ok";
    case "PoweringOn":
    case "PoweringOff":
      return "changing";
    case "Warning":
    case "Paused":
      return "warning";
    case "Critical":
      return "critical";
    case "Off":
      return "off";
    default:
      return "unknown";
  }
}

const ledClass: Record<Tone, string> = {
  ok: "bg-(--node-ready) shadow-[0_0_6px_var(--node-ready)]",
  changing: "bg-(--node-amber) shadow-[0_0_6px_var(--node-amber)] motion-safe:animate-pulse",
  warning: "bg-(--node-amber) shadow-[0_0_6px_var(--node-amber)]",
  critical: "bg-destructive shadow-[0_0_6px_var(--color-destructive)]",
  off: "ring-1 ring-inset ring-muted-foreground/50",
  unknown: "bg-muted-foreground/25",
  locate: "bg-(--node-blue) shadow-[0_0_6px_var(--node-blue)]",
};

// toneText is the colour of the words beside a light; warning matches the warning badge rather than the LED, whose amber is too light to read as text
const toneText: Record<Tone, string> = {
  ok: "text-foreground",
  changing: "text-foreground",
  warning: "text-amber-700 dark:text-amber-400",
  critical: "text-destructive",
  off: "text-muted-foreground",
  unknown: "text-muted-foreground",
  locate: "text-foreground",
};

// Led is one indicator light; the locate light pulses the way it blinks in the rack
function Led({ tone }: { tone: Tone }) {
  return (
    <span aria-hidden className="relative inline-flex size-2 shrink-0">
      {tone === "locate" && (
        <span className="absolute inset-0 rounded-full bg-(--node-blue) opacity-75 motion-safe:animate-ping" />
      )}
      <span className={cn("relative size-full rounded-full", ledClass[tone])} />
    </span>
  );
}

// subsystems are the iDRAC rollups shown on the front panel, in the order a person checks them
const subsystems = [
  ["Processors", "CPURollupStatus"],
  ["Memory", "SysMemPrimaryStatus"],
  ["Storage", "StorageRollupStatus"],
  ["Fans", "FanRollupStatus"],
  ["Temperature", "TempRollupStatus"],
  ["Power supplies", "PSRollupStatus"],
  ["Voltage", "VoltRollupStatus"],
  ["Current", "CurrentRollupStatus"],
  ["Battery", "BatteryRollupStatus"],
  ["Intrusion", "IntrusionRollupStatus"],
  ["Event log", "SELRollupStatus"],
  ["Licensing", "LicensingRollupStatus"],
] as const satisfies ReadonlyArray<readonly [string, keyof Dell]>;

// FrontPanel draws the server's bezel as the BMC sees it: power and health on the left ear, the locate light on the right, and the subsystem health across the front. On a phone the model takes the top row and the ears sit side by side under it
function FrontPanel({ node, system, dell }: { node: string; system?: System; dell?: Dell }) {
  const present = dell
    ? subsystems.flatMap(([label, key]) => {
        const value = dell[key];
        return value ? [{ label, value, tone: toneOf(value) }] : [];
      })
    : [];

  return (
    <section
      aria-label="Front panel"
      className="grid grid-cols-2 overflow-hidden rounded-lg border bg-muted/40 sm:min-h-19 sm:grid-cols-[auto_minmax(0,1fr)_auto]"
    >
      <div className="col-span-2 row-start-1 flex min-w-0 flex-wrap items-center justify-between gap-x-8 gap-y-2 px-5 py-3 sm:col-span-1 sm:col-start-2">
        <div className="min-w-0">
          <div className="truncate text-sm text-muted-foreground">
            {[system?.manufacturer, dell?.SystemGeneration].filter(Boolean).join(", ") ||
              "Unknown maker"}
          </div>
          <h2 className="truncate text-2xl font-semibold tracking-tight">
            {system?.model || "Unknown model"}
          </h2>
        </div>
        {present.length > 0 && <SubsystemHealth present={present} />}
      </div>

      <Ear side="left">
        <EarLight label="Power" value={system?.power_status} />
        <EarLight label="Health" value={system?.health} />
      </Ear>
      <Ear side="right">
        <LocateLight node={node} state={system?.location_indicator} />
      </Ear>
    </section>
  );
}

type Subsystem = { label: string; value: string; tone: Tone };

// SubsystemHealth says in one line that every subsystem is OK, or names the ones that are not; the full list opens from it
function SubsystemHealth({ present }: { present: Subsystem[] }) {
  const faults = present.filter((s) => s.tone !== "ok");
  const healthy = present.length - faults.length;
  const summary =
    faults.length === 0
      ? `All ${present.length} subsystems OK`
      : healthy > 0
        ? `${healthy} more OK`
        : "All subsystems";

  return (
    <div className="flex flex-wrap items-center gap-x-4 gap-y-1.5 text-sm">
      {faults.map((s) => (
        <span key={s.label} className="flex items-center gap-1.5">
          <Led tone={s.tone} />
          {s.label}
          <span className={cn("font-medium", toneText[s.tone])}>{s.value}</span>
        </span>
      ))}
      <Popover>
        <PopoverTrigger className="flex items-center gap-1.5 rounded-sm text-muted-foreground underline decoration-dotted underline-offset-4 hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none">
          {faults.length === 0 && <Led tone="ok" />}
          {summary}
        </PopoverTrigger>
        <PopoverContent align="end" className="w-60">
          <ul className="grid gap-1.5 text-sm" aria-label="Subsystem health">
            {present.map((s) => (
              <li key={s.label} className="flex items-center justify-between gap-3">
                <span className="flex items-center gap-2">
                  <Led tone={s.tone} />
                  {s.label}
                </span>
                <span className={cn("text-xs", toneText[s.tone])}>{s.value}</span>
              </li>
            ))}
          </ul>
        </PopoverContent>
      </Popover>
    </div>
  );
}

// Ear is one rack ear of the front panel, with the screw holes that hold it to the rack at its outer edge, clear of its lights
function Ear({ side, children }: { side: "left" | "right"; children: ReactNode }) {
  const left = side === "left";
  return (
    <div
      className={cn(
        "relative flex items-center border-t bg-muted/60 py-3 sm:border-t-0",
        left ? "border-r pr-4 pl-7" : "pr-7 pl-4 sm:border-l",
      )}
    >
      {["top-2", "bottom-2"].map((y) => (
        <span
          key={y}
          aria-hidden
          className={cn(
            "absolute size-1.5 rounded-full bg-background ring-1 ring-border",
            y,
            left ? "left-2" : "right-2",
          )}
        />
      ))}
      <div className="grid grid-cols-[auto_auto_auto] items-center gap-x-2 gap-y-2.5 text-xs">
        {children}
      </div>
    </div>
  );
}

// EarLight is one light on an ear with its name and state on one line, so two fit in a 1U panel
function EarLight({ label, value }: { label: string; value?: string }) {
  const tone = toneOf(value);
  return (
    <div className="col-span-3 grid grid-cols-subgrid items-center">
      <Led tone={tone} />
      <span className="text-muted-foreground">{label}</span>
      <span className={cn("font-medium", toneText[tone])}>
        {value ? readable(value) : "Unknown"}
      </span>
    </div>
  );
}

// LocateLight is the locate light on the right ear, and a switch for it: clicking it turns the light on the node on or off, then rereads the BMC
function LocateLight({ node, state }: { node: string; state?: string }) {
  const router = useRouter();
  const { mutate, isPending, variables } = useMutation({
    ...postV1BmcIndicatorMutation(),
    // an onSuccess given here is awaited, so the mutation stays pending through the reread and its variables stand in for the light until the BMC reports it. A BMC that refused still answers 200, and returning early lets the light fall back to what the BMC last reported
    onSuccess: async (data, { body }) => {
      if (!data?.results?.[0]?.success) return toastBmcResponse(`Locate LED ${body.state}`, data);
      await router.invalidate();
    },
    onError: (e) => toast.error(e.title, { description: e.detail }),
  });
  const on = isPending ? variables.body.state === "on" : state === "On";
  const next = on ? "off" : "on";

  return (
    // aria-disabled rather than disabled while pending, so the button keeps keyboard focus
    <button
      type="button"
      onClick={() => !isPending && mutate({ query: { nodeset: node }, body: { state: next } })}
      aria-disabled={isPending}
      aria-pressed={on}
      title={`Turn the locate LED ${next}`}
      className="col-span-3 -mx-2 grid cursor-pointer grid-cols-subgrid items-center rounded-md px-2 py-1 text-left transition-colors hover:bg-background/70 focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none aria-disabled:cursor-wait aria-disabled:opacity-70"
    >
      <Led tone={on ? "locate" : "unknown"} />
      <span className="text-muted-foreground">Locate</span>
      <span className="font-medium">{isPending ? (on ? "On" : "Off") : state || "Unknown"}</span>
    </button>
  );
}

// Section heads one part of the tab below the front panel, with an optional link or badge beside its title
function Section({
  title,
  aside,
  children,
}: {
  title: string;
  aside?: ReactNode;
  children: ReactNode;
}) {
  return (
    <section className="grid grid-cols-1 content-start gap-3">
      <div className="flex min-h-8 items-center justify-between gap-2">
        <h2 className="font-medium">{title}</h2>
        {aside}
      </div>
      {children}
    </section>
  );
}

// Note is the line a section shows in place of what it could not show
function Note({ children }: { children: ReactNode }) {
  return <p className="text-sm text-muted-foreground">{children}</p>;
}

// joined lists the parts that are there, so a missing detail leaves no stray comma
const joined = (...parts: Array<string | 0 | false | undefined>) =>
  parts.filter(Boolean).join(", ");

type Spec = { label: string; value?: string; title?: string; wide?: boolean };

// SpecSheet lists what the BMC knows about the hardware, laid out like the service tag pulled from the front of the server
function SpecSheet({ system, dell }: { system?: System; dell?: Dell }) {
  const inventoried = dell?.LastSystemInventoryTime
    ? relativeTime(dell.LastSystemInventoryTime)
    : undefined;
  const cores = system?.processor_cores;
  // processor_count is Redfish's LogicalProcessorCount, the threads
  const threads = system?.processor_count;
  const dimms = dell?.MaxDIMMSlots;
  const slots = dell?.MaxPCIeSlots;
  const specs: Spec[] = [
    {
      label: "Processor",
      value: cpuModel(system?.processor_model, system?.processor_sockets),
      wide: true,
    },
    { label: "Cores", value: joined(cores && `${cores} cores`, threads && `${threads} threads`) },
    {
      label: "Memory",
      value: system?.total_memory
        ? joined(
            `${system.total_memory} GB`,
            dimms && `${dell.PopulatedDIMMSlots ?? 0} of ${dimms} DIMMs`,
          )
        : "",
    },
    {
      label: "BIOS",
      value: system?.bios_version && joined(system.bios_version, dell?.BIOSReleaseDate),
    },
    { label: "PCIe slots", value: slots ? `${dell.PopulatedPCIeSlots ?? 0} of ${slots} used` : "" },
    {
      label: "Height",
      value: dell?.ChassisSystemHeightUnit ? `${dell.ChassisSystemHeightUnit}U` : "",
    },
    { label: "Serial", value: system?.serial_number },
    { label: "Express service code", value: dell?.ExpressServiceCode },
    { label: "Asset tag", value: system?.asset_tag },
    { label: "Hostname", value: system?.host_name },
    { label: "Location", value: formatLocation(system?.location) },
    { label: "Inventoried", value: inventoried?.text, title: inventoried?.title },
  ];
  const shown = specs.filter((s) => s.value);

  return (
    <Section title="Hardware">
      {shown.length === 0 ? (
        <Note>The BMC reports no hardware details</Note>
      ) : (
        <dl className="grid gap-x-8 text-sm sm:grid-cols-2">
          {shown.map((s) => (
            <div
              key={s.label}
              className={cn(
                "flex min-w-0 items-baseline justify-between gap-4 border-b border-dashed py-2",
                s.wide && "sm:col-span-2",
              )}
            >
              <dt className="shrink-0 text-muted-foreground">{s.label}</dt>
              <dd
                className="truncate text-right font-mono text-[0.8125rem]"
                title={s.title ?? s.value}
              >
                {s.value}
              </dd>
            </div>
          ))}
        </dl>
      )}
    </Section>
  );
}

function Boot({ system }: { system?: System }) {
  const order = system?.boot_order ?? [];
  const override =
    system?.boot_override_enabled &&
    system.boot_override_enabled !== "Disabled" &&
    system.boot_override_target &&
    system.boot_override_target !== "None";
  const progressed = system?.boot_progress_time
    ? relativeTime(system.boot_progress_time)
    : undefined;

  return (
    <Section
      title="Boot"
      aside={system?.boot_mode && <Badge variant="outline">{system.boot_mode}</Badge>}
    >
      {system?.boot_progress ? (
        <p className="text-sm">
          {readable(system.boot_progress)}
          {progressed && (
            <span className="text-muted-foreground" title={progressed.title}>
              {" "}
              {progressed.text}
            </span>
          )}
        </p>
      ) : (
        <Note>This BMC does not report boot progress</Note>
      )}

      {(override || system?.boot_next) && (
        <div className="grid gap-1 rounded-md border border-(--node-amber)/50 bg-(--node-amber)/10 px-3 py-2 text-sm">
          {override && (
            <div>
              Next boot goes to{" "}
              <span className="font-mono font-medium">{system?.boot_override_target}</span>{" "}
              <span className="text-muted-foreground">
                ({system?.boot_override_enabled?.toLowerCase()})
              </span>
            </div>
          )}
          {system?.boot_next && (
            <div>
              UEFI BootNext is <span className="font-mono font-medium">{system.boot_next}</span>
            </div>
          )}
        </div>
      )}

      {order.length === 0 ? (
        <Note>The BMC reports no boot order</Note>
      ) : (
        <ol className="grid text-sm">
          {order.map((device, i) => (
            <li
              key={device}
              className="flex items-center gap-3 border-l-2 py-1.5 pl-3 first:border-foreground"
            >
              <span className="w-4 shrink-0 text-right text-xs text-muted-foreground tabular-nums">
                {i + 1}
              </span>
              <span className={cn("truncate font-mono", i > 0 && "text-muted-foreground")}>
                {device}
              </span>
            </li>
          ))}
        </ol>
      )}
    </Section>
  );
}

type BmcQuery = {
  isPending: boolean;
  isError: boolean;
  data?: { results?: Array<{ success?: boolean; message?: string }> };
};

// BmcSection is a Section filled from one bmc query of one node: a skeleton while it loads, the BMC's message when it fails, and the empty line when it found nothing
function BmcSection({
  title,
  aside,
  query,
  count,
  failed,
  empty,
  children,
}: {
  title: string;
  aside: ReactNode;
  query: BmcQuery;
  count: number;
  failed: string;
  empty: string;
  children: ReactNode;
}) {
  const result = query.data?.results?.[0];
  return (
    <Section title={title} aside={aside}>
      {query.isPending ? (
        <div className="grid gap-2">
          {[0, 1, 2].map((i) => (
            <Skeleton key={i} className="h-5 w-full" />
          ))}
        </div>
      ) : query.isError || (result && !result.success) ? (
        <Note>{result?.message || failed}</Note>
      ) : count === 0 ? (
        <Note>{empty}</Note>
      ) : (
        children
      )}
    </Section>
  );
}

function RecentEvents({ node }: { node: string }) {
  const query = useQuery({
    ...getV1BmcSelOptions({ query: { nodeset: node, limit: "5" } }),
    staleTime: 60_000,
  });
  const entries = query.data?.results?.[0]?.data?.entries ?? [];

  return (
    <BmcSection
      title="Recent events"
      aside={<PanelLink to="/nodes/sel/$nodes" node={node} label="Full log" />}
      query={query}
      count={entries.length}
      failed="Could not read the event log"
      empty="The event log is empty"
    >
      <ul className="grid grid-cols-[auto_1fr] gap-x-3 divide-y">
        {entries.map((e) => {
          const created = e.created ? relativeTime(e.created) : undefined;
          return (
            <li
              key={e.id}
              className="col-span-2 grid grid-cols-subgrid items-start py-2.5 text-sm first:pt-0"
            >
              <Badge variant={logSeverityVariant(e.severity)}>{e.severity || "-"}</Badge>
              <div className="min-w-0">
                <div className="line-clamp-2">{e.message}</div>
                {created && (
                  <div className="text-xs text-muted-foreground" title={created.title}>
                    {created.text}
                  </div>
                )}
              </div>
            </li>
          );
        })}
      </ul>
    </BmcSection>
  );
}

function JobQueue({ node }: { node: string }) {
  const query = useQuery({
    ...getV1BmcJobsOptions({ query: { nodeset: node } }),
    staleTime: 60_000,
  });
  const jobs = query.data?.results?.[0]?.data?.jobs ?? [];

  return (
    <BmcSection
      title="Job queue"
      aside={<PanelLink to="/nodes/jobs/$nodes" node={node} label="All jobs" />}
      query={query}
      count={jobs.length}
      failed="Could not read the job queue"
      empty="No jobs are queued on the BMC"
    >
      <ul className="grid grid-cols-1 gap-3">
        {jobs.slice(0, 5).map((job) => (
          <li key={job.id}>
            <Progress value={Math.min(100, job.percent_complete ?? 0)} className="gap-x-2 gap-y-1">
              <ProgressLabel className="min-w-0 flex-1 truncate font-normal">
                {job.name || job.id}
              </ProgressLabel>
              <ProgressValue className="text-xs">
                {(percent) => joined(job.state, percent ?? undefined)}
              </ProgressValue>
            </Progress>
          </li>
        ))}
        {jobs.length > 5 && (
          <li className="text-xs text-muted-foreground">and {jobs.length - 5} more</li>
        )}
      </ul>
    </BmcSection>
  );
}

function PanelLink({
  to,
  node,
  label,
}: {
  to: "/nodes/sel/$nodes" | "/nodes/jobs/$nodes";
  node: string;
  label: string;
}) {
  return (
    <Button
      variant="ghost"
      size="sm"
      render={<Link to={to} params={{ nodes: node }} />}
      nativeButton={false}
    >
      {label}
      <ArrowRight />
    </Button>
  );
}

// cpuModel shortens a processor model string, which bmcs report with trademark marks and a redundant CPU, and prefixes the socket count
function cpuModel(model?: string, sockets?: number) {
  const name = (model ?? "")
    .replace(/\((R|TM)\)/gi, "")
    .replace(/\bCPU\b/g, "")
    .replace(/\s+/g, " ")
    .trim();
  if (!name) return sockets ? `${sockets} socket(s)` : "";
  return sockets && sockets > 1 ? `${sockets} × ${name}` : name;
}

// readable turns a Redfish value such as OSRunning, PoweringOn or PCIResourceConfigStarted into words, keeping acronyms
function readable(state: string) {
  const words = state
    .replace(/([a-z\d])([A-Z])/g, "$1 $2")
    .replace(/([A-Z]+)([A-Z][a-z])/g, "$1 $2")
    .split(" ");
  return words
    .map((w, i) => {
      if (w.length > 1 && w === w.toUpperCase()) return w;
      return i === 0 ? w[0].toUpperCase() + w.slice(1).toLowerCase() : w.toLowerCase();
    })
    .join(" ");
}
