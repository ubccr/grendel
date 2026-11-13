export { cn } from "cn";

export function severityColor(severity?: string) {
  switch (severity?.toLowerCase()) {
    case "error":
      return "text-red-600 border-red-400";
    case "warning":
      return "text-yellow-600 border-yellow-400";
    case "success":
      return "text-green-600 border-green-400";
    default:
      return "";
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
