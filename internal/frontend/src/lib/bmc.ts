import { severityColor } from "@/lib/utils";
import { toast } from "sonner";

type BmcSummary = {
  total?: number;
  succeeded?: number;
  failed?: number;
  failed_nodes?: string;
};

type BmcResult = {
  node?: string;
  success?: boolean;
  message?: string;
};

export type BmcResponse = {
  summary?: BmcSummary;
  results?: Array<BmcResult>;
};

// A bmc route answers 200 for a completed run even when nodes in it failed, so the outcome is read from the summary rather than the status code
export function toastBmcResponse(action: string, res: BmcResponse | undefined) {
  const summary = res?.summary ?? {};
  const failed = summary.failed ?? 0;
  const total = summary.total ?? 0;

  if (failed === 0) {
    const note = res?.results?.length === 1 ? res.results[0].message : undefined;
    toast.success(`${action}: ${total} node(s) succeeded`, { description: note || undefined });
    return;
  }

  const first = res?.results?.find((r) => !r.success);
  const description = `Failed: ${summary.failed_nodes}${first ? `. ${first.node}: ${first.message}` : ""}`;

  if ((summary.succeeded ?? 0) > 0) {
    toast.warning(`${action}: ${failed} of ${total} node(s) failed`, { description });
  } else {
    toast.error(`${action}: every node failed`, { description });
  }
}

// logSeverityColor colours a Redfish severity the way severityColor colours an event's
export function logSeverityColor(severity?: string) {
  switch (severity) {
    case "Critical":
      return severityColor("error");
    case "Warning":
      return severityColor("warning");
    case "OK":
      return severityColor("success");
    default:
      return "";
  }
}
