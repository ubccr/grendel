export { cn } from "cn";

export type Severity = "success" | "info" | "warning" | "error";

// severityOf reads the severities grendel writes (an event's Success, Info, Warning or Error, a node's success or error under it) and the ones Redfish reports (OK, Warning, Critical) as one of four, so every badge, dot and rail colours them the same way. Anything else has none
export function severityOf(value?: string): Severity | undefined {
  switch (value?.toLowerCase()) {
    case "success":
    case "ok":
      return "success";
    case "info":
      return "info";
    case "warning":
      return "warning";
    case "error":
    case "critical":
      return "error";
  }
}

const severityVariants = {
  success: "success",
  info: "secondary",
  warning: "warning",
  error: "destructive",
} as const;

// severityVariant picks the Badge variant for a severity
export function severityVariant(value?: string) {
  const severity = severityOf(value);
  return severity ? severityVariants[severity] : "secondary";
}

const relativeUnits: Array<[Intl.RelativeTimeFormatUnit, number]> = [
  ["year", 365 * 24 * 60 * 60],
  ["month", 30 * 24 * 60 * 60],
  ["week", 7 * 24 * 60 * 60],
  ["day", 24 * 60 * 60],
  ["hour", 60 * 60],
  ["minute", 60],
  ["second", 1],
];

const relativeFormat = new Intl.RelativeTimeFormat(undefined, { numeric: "auto" });
const narrowFormat = new Intl.RelativeTimeFormat(undefined, { numeric: "auto", style: "narrow" });

// relativeTime describes a time from now, such as "5 minutes ago", or "5m ago" when narrow, with the absolute local time for a tooltip. It takes a timestamp string, a Date, or milliseconds since the epoch; a string that does not parse as a date comes back as is
export function relativeTime(
  value: string | number | Date,
  { now = Date.now(), narrow = false }: { now?: number; narrow?: boolean } = {},
) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return { text: String(value), title: String(value) };

  const seconds = Math.round((date.getTime() - now) / 1000);
  const [unit, size] =
    relativeUnits.find(([, size]) => Math.abs(seconds) >= size) ??
    relativeUnits[relativeUnits.length - 1];
  return {
    text: (narrow ? narrowFormat : relativeFormat).format(Math.round(seconds / size), unit),
    title: date.toLocaleString(),
  };
}
