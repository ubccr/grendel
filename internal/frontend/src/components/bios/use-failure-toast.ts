import { useEffect, useRef } from "react";
import { toast } from "sonner";

type Result = { node?: string; success?: boolean; message?: string };

// useFailureToast reports the nodes a bios read could not reach, once per run, so the page itself only shows what was read
export function useFailureToast(runID: string | undefined, results: Result[]) {
  const shown = useRef<string>(undefined);

  useEffect(() => {
    if (!runID || shown.current === runID) return;
    shown.current = runID;

    const failed = results.filter((r) => !r.success);
    if (failed.length === 0) return;

    const first = failed[0];
    const description = `${first.node}: ${first.message}${failed.length > 1 ? ` (and ${failed.length - 1} more)` : ""}`;
    if (failed.length === results.length) toast.error("Could not read the BIOS", { description });
    else toast.warning(`${failed.length} node(s) could not be read`, { description });
  }, [runID, results]);
}
