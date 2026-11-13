import { BmcResponseModelRedfishSystem, getV1BmcStatus } from "@/client";
import { getV1BmcJobsOptions, getV1BmcSelOptions } from "@/client/@tanstack/react-query.gen";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { ErrorPanel } from "@/components/error-panel";
import { Skeleton } from "@/components/ui/skeleton";
import { logSeverityColor } from "@/lib/bmc";
import { formatLocation } from "@/lib/identity";
import { cn, relativeTime } from "@/lib/utils";
import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import type { ReactNode } from "react";
import {
  Activity,
  ArrowRight,
  BatteryFull,
  CircuitBoard,
  Cpu,
  Fan,
  HardDrive,
  HeartPulse,
  KeyRound,
  ListChecks,
  MemoryStick,
  Plug,
  Power,
  ScrollText,
  ShieldAlert,
  Thermometer,
  Zap,
  type LucideIcon,
} from "lucide-react";

export const Route = createFileRoute("/nodes/$node/redfish")({
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
  const dell = system?.oem_dell ?? undefined;

  return (
    <div className="grid gap-4">
      <div className="grid gap-4 lg:grid-cols-2">
        <Overview system={system} dell={dell} />
        <Boot system={system} />
      </div>

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <Stat
          icon={Cpu}
          label="Processors"
          value={
            [
              system?.processor_cores ? `${system.processor_cores} cores` : "",
              system?.processor_count ? `${system.processor_count} threads` : "",
            ]
              .filter(Boolean)
              .join(" - ") || "-"
          }
          detail={[cpuModel(system?.processor_model, system?.processor_sockets)]}
        />
        <Stat
          icon={MemoryStick}
          label="Memory"
          value={system?.total_memory ? `${system.total_memory} GB` : "-"}
          detail={[
            dell?.MaxDIMMSlots
              ? `${dell.PopulatedDIMMSlots ?? 0} of ${dell.MaxDIMMSlots} DIMMs`
              : "",
          ]}
        />
        <Stat
          icon={CircuitBoard}
          label="BIOS"
          value={system?.bios_version || "-"}
          detail={[dell?.BIOSReleaseDate ? `released ${dell.BIOSReleaseDate}` : ""]}
        />
        <Stat
          icon={Activity}
          label="Boot progress"
          value={system?.boot_progress ? bootState(system.boot_progress) : "-"}
          detail={[
            system?.boot_progress_time
              ? relativeTime(system.boot_progress_time)
              : system?.boot_progress
                ? ""
                : "not reported by this BMC",
          ]}
        />
      </div>

      {dell && <Subsystems dell={dell} />}

      <div className="grid gap-4 lg:grid-cols-2">
        <RecentEvents node={node} />
        <JobQueue node={node} />
      </div>
    </div>
  );
}

// tone maps a Redfish health or state to a colour family
function tone(value?: string) {
  switch (value) {
    case "OK":
    case "On":
      return { dot: "bg-green-500", text: "text-green-600", ring: "border-green-400/60" };
    case "Warning":
      return { dot: "bg-yellow-500", text: "text-yellow-600", ring: "border-yellow-400/60" };
    case "Critical":
    case "Off":
      return { dot: "bg-red-500", text: "text-red-600", ring: "border-red-400/60" };
    default:
      return { dot: "bg-muted-foreground/40", text: "text-muted-foreground", ring: "" };
  }
}

function Overview({ system, dell }: { system?: System; dell?: Dell }) {
  const inventoried = dell?.LastSystemInventoryTime
    ? relativeTime(dell.LastSystemInventoryTime)
    : undefined;
  // a fact is label, value, hover title, and what to show when it is more than the value
  const facts: Array<[string, string | undefined, string?, ReactNode?]> = [
    ["Serial", system?.serial_number],
    ["Hostname", system?.host_name],
    ["Asset tag", system?.asset_tag],
    ["Location", formatLocation(system?.location) || undefined],
    ["Express service code", dell?.ExpressServiceCode],
    [
      "PCIe slots",
      dell?.MaxPCIeSlots ? `${dell.PopulatedPCIeSlots ?? 0} of ${dell.MaxPCIeSlots}` : undefined,
    ],
    ["Height", dell?.ChassisSystemHeightUnit ? `${dell.ChassisSystemHeightUnit}U` : undefined],
    ["Inventoried", inventoried?.text, inventoried?.title],
    [
      "Locate LED",
      system?.location_indicator,
      undefined,
      <LedState key="led" state={system?.location_indicator} />,
    ],
  ];

  return (
    <Card>
      <CardContent className="grid gap-4 pt-6">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div className="min-w-0">
            <div className="text-sm text-muted-foreground">
              {[system?.manufacturer, dell?.SystemGeneration].filter(Boolean).join(" · ")}
            </div>
            <h2 className="truncate text-2xl font-semibold tracking-tight">
              {system?.model || "Unknown model"}
            </h2>
          </div>
          <div className="flex flex-wrap gap-2">
            <Pill icon={Power} label="Power" value={system?.power_status} />
            <Pill icon={HeartPulse} label="Health" value={system?.health} />
          </div>
        </div>
        <dl className="grid grid-cols-1 gap-x-6 gap-y-2 text-sm sm:grid-cols-2">
          {facts
            .filter(([, value]) => value)
            .map(([label, value, title, display]) => (
              <div key={label} className="min-w-0">
                <dt className="text-xs text-muted-foreground">{label}</dt>
                <dd className="truncate font-mono" title={title ?? value}>
                  {display ?? value}
                </dd>
              </div>
            ))}
        </dl>
      </CardContent>
    </Card>
  );
}

// LedState blinks a dot while the locate light is on, the way the light itself does in the rack
function LedState({ state }: { state?: string }) {
  const on = state === "On";
  return (
    <span className="inline-flex items-center gap-2">
      <span className="relative ml-1 inline-flex size-2">
        {on && (
          <span className="absolute inline-flex size-full rounded-full bg-sky-400 opacity-75 motion-safe:animate-ping" />
        )}
        <span
          className={cn(
            "relative inline-flex size-2 rounded-full",
            on ? "bg-sky-500" : "bg-muted-foreground/40",
          )}
        />
      </span>
      {state}
    </span>
  );
}

function Pill({ icon: Icon, label, value }: { icon: LucideIcon; label: string; value?: string }) {
  const t = tone(value);
  return (
    <div className={cn("flex items-center gap-2 rounded-lg border px-3 py-2", t.ring)}>
      <Icon className={cn("size-4", t.text)} />
      <div className="leading-tight">
        <div className="text-xs text-muted-foreground">{label}</div>
        <div className={cn("text-sm font-medium", t.text)}>{value || "Unknown"}</div>
      </div>
    </div>
  );
}

function Stat({
  icon: Icon,
  label,
  value,
  detail,
}: {
  icon: LucideIcon;
  label: string;
  value: string;
  detail: Array<string | { text: string; title: string }>;
}) {
  const lines = detail
    .filter(Boolean)
    .map((line) => (typeof line === "string" ? { text: line, title: line } : line));
  return (
    <Card>
      <CardContent className="flex items-start gap-3 pt-6">
        <div className="rounded-md bg-muted p-2">
          <Icon className="size-5 text-muted-foreground" />
        </div>
        <div className="min-w-0">
          <div className="text-xs text-muted-foreground">{label}</div>
          <div className="truncate text-xl font-semibold">{value}</div>
          {lines.map((line) => (
            <div
              key={line.text}
              className="truncate text-xs text-muted-foreground"
              title={line.title}
            >
              {line.text}
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

const subsystems: Array<[string, keyof Dell, LucideIcon]> = [
  ["Processors", "CPURollupStatus", Cpu],
  ["Memory", "SysMemPrimaryStatus", MemoryStick],
  ["Storage", "StorageRollupStatus", HardDrive],
  ["Fans", "FanRollupStatus", Fan],
  ["Temperature", "TempRollupStatus", Thermometer],
  ["Power supplies", "PSRollupStatus", Plug],
  ["Voltage", "VoltRollupStatus", Zap],
  ["Current", "CurrentRollupStatus", Zap],
  ["Battery", "BatteryRollupStatus", BatteryFull],
  ["Intrusion", "IntrusionRollupStatus", ShieldAlert],
  ["Event log", "SELRollupStatus", ScrollText],
  ["Licensing", "LicensingRollupStatus", KeyRound],
];

function Subsystems({ dell }: { dell: Dell }) {
  const present = subsystems.filter(([, key]) => typeof dell[key] === "string" && dell[key]);
  if (present.length === 0) return null;

  const unhealthy = present.filter(([, key]) => dell[key] !== "OK").length;

  return (
    <Card>
      <CardHeader>
        <CardTitle>Subsystem health</CardTitle>
        <CardDescription>
          {unhealthy === 0
            ? "Every subsystem the iDRAC reports is OK"
            : `${unhealthy} subsystem(s) need attention`}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <ul className="grid gap-x-8 gap-y-2 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {present.map(([label, key, Icon]) => {
            const value = dell[key] as string;
            const t = tone(value);
            return (
              <li key={key} className="flex items-center justify-between gap-2 text-sm">
                <span className="flex items-center gap-2">
                  <Icon className="size-4 text-muted-foreground" />
                  {label}
                </span>
                <span className={cn("flex items-center gap-1.5 font-medium", t.text)}>
                  <span className={cn("size-2 rounded-full", t.dot)} />
                  {value}
                </span>
              </li>
            );
          })}
        </ul>
      </CardContent>
    </Card>
  );
}

function Boot({ system }: { system?: System }) {
  const order = system?.boot_order ?? [];
  const override =
    system?.boot_override_enabled &&
    system.boot_override_enabled !== "Disabled" &&
    system.boot_override_target &&
    system.boot_override_target !== "None";
  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between gap-2 space-y-0">
        <div className="space-y-1.5">
          <CardTitle>Boot</CardTitle>
          <CardDescription>
            {override ? (
              <>
                Override:{" "}
                <span className="font-mono text-foreground">{system?.boot_override_target}</span> (
                {system?.boot_override_enabled?.toLowerCase()})
              </>
            ) : (
              "No boot override is set"
            )}
            {system?.boot_next && (
              <>
                <br />
                UEFI BootNext: <span className="font-mono text-foreground">{system.boot_next}</span>
              </>
            )}
          </CardDescription>
        </div>
        {system?.boot_mode && <Badge variant="outline">{system.boot_mode}</Badge>}
      </CardHeader>
      <CardContent>
        {order.length === 0 ? (
          <span className="text-sm text-muted-foreground">The BMC reports no boot order</span>
        ) : (
          <ol className="grid gap-1.5">
            {order.map((device, i) => (
              <li key={device} className="flex items-center gap-3 text-sm">
                <span className="flex size-6 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-medium">
                  {i + 1}
                </span>
                <span className="truncate font-mono">{device}</span>
              </li>
            ))}
          </ol>
        )}
      </CardContent>
    </Card>
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

function RowsSkeleton() {
  return (
    <div className="grid gap-2">
      {[0, 1, 2].map((i) => (
        <Skeleton key={i} className="h-5 w-full" />
      ))}
    </div>
  );
}

function RecentEvents({ node }: { node: string }) {
  const { data, isPending, isError } = useQuery({
    ...getV1BmcSelOptions({ query: { nodeset: node, limit: "5" } }),
    staleTime: 60_000,
  });
  const result = data?.results?.[0];
  const entries = result?.data?.entries ?? [];

  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between space-y-0">
        <div className="space-y-1.5">
          <CardTitle className="flex items-center gap-2">
            <ScrollText className="size-4" /> Recent events
          </CardTitle>
          <CardDescription>The latest entries in the system event log</CardDescription>
        </div>
        <PanelLink to="/nodes/sel/$nodes" node={node} label="Full log" />
      </CardHeader>
      <CardContent>
        {isPending ? (
          <RowsSkeleton />
        ) : isError || (result && !result.success) ? (
          <span className="text-sm text-muted-foreground">
            {result?.message || "Could not read the event log"}
          </span>
        ) : entries.length === 0 ? (
          <span className="text-sm text-muted-foreground">The event log is empty</span>
        ) : (
          <ul className="grid grid-cols-[auto_1fr] gap-x-3 gap-y-3">
            {entries.map((e) => (
              <li key={e.id} className="col-span-2 grid grid-cols-subgrid items-start text-sm">
                <Badge
                  variant="secondary"
                  className={cn("justify-center rounded-sm", logSeverityColor(e.severity))}
                >
                  {e.severity || "-"}
                </Badge>
                <div className="min-w-0">
                  <div className="line-clamp-2">{e.message}</div>
                  <div className="text-xs text-muted-foreground">
                    {e.created ? new Date(e.created).toLocaleString() : ""}
                  </div>
                </div>
              </li>
            ))}
          </ul>
        )}
      </CardContent>
    </Card>
  );
}

function JobQueue({ node }: { node: string }) {
  const { data, isPending, isError } = useQuery({
    ...getV1BmcJobsOptions({ query: { nodeset: node } }),
    staleTime: 60_000,
  });
  const result = data?.results?.[0];
  const jobs = result?.data?.jobs ?? [];

  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between space-y-0">
        <div className="space-y-1.5">
          <CardTitle className="flex items-center gap-2">
            <ListChecks className="size-4" /> Job queue
          </CardTitle>
          <CardDescription>
            {isPending || isError || !result?.success
              ? "Jobs queued on the BMC"
              : `${jobs.length} job(s) queued on the BMC`}
          </CardDescription>
        </div>
        <PanelLink to="/nodes/jobs/$nodes" node={node} label="All jobs" />
      </CardHeader>
      <CardContent>
        {isPending ? (
          <RowsSkeleton />
        ) : isError || (result && !result.success) ? (
          <span className="text-sm text-muted-foreground">
            {result?.message || "Could not read the job queue"}
          </span>
        ) : jobs.length === 0 ? (
          <span className="text-sm text-muted-foreground">The job queue is empty</span>
        ) : (
          <ul className="grid gap-3">
            {jobs.slice(0, 5).map((job) => (
              <li key={job.id} className="grid gap-1 text-sm">
                <div className="flex items-center justify-between gap-2">
                  <span className="truncate">{job.name || job.id}</span>
                  <span className="shrink-0 text-xs text-muted-foreground">{job.state}</span>
                </div>
                <div className="h-1.5 overflow-hidden rounded-full bg-muted">
                  <div
                    className="h-full rounded-full bg-primary"
                    style={{ width: `${Math.min(100, job.percent_complete ?? 0)}%` }}
                  />
                </div>
              </li>
            ))}
            {jobs.length > 5 && (
              <li className="text-xs text-muted-foreground">and {jobs.length - 5} more</li>
            )}
          </ul>
        )}
      </CardContent>
    </Card>
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

// bootState turns a Redfish boot progress value such as OSRunning or PCIResourceConfigStarted into words, keeping acronyms
function bootState(state: string) {
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
