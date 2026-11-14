export { cn } from "cn";

// severityVariant picks the Badge variant for an event's severity
export function severityVariant(severity?: string) {
  switch (severity?.toLowerCase()) {
    case "error":
      return "destructive";
    case "warning":
      return "warning";
    case "success":
      return "success";
    default:
      return "secondary";
  }
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

// relativeTime describes a timestamp from now, such as "5 minutes ago", with the absolute local time for a tooltip. A value that does not parse as a date comes back as is
export function relativeTime(value: string, now = Date.now()) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return { text: value, title: value };

  const seconds = Math.round((date.getTime() - now) / 1000);
  const [unit, size] =
    relativeUnits.find(([, size]) => Math.abs(seconds) >= size) ??
    relativeUnits[relativeUnits.length - 1];
  return {
    text: relativeFormat.format(Math.round(seconds / size), unit),
    title: date.toLocaleString(),
  };
}
