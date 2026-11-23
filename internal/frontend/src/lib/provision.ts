// the states of a provision in the order a node goes through them, ending in the two it can finish in. label names a state in a badge or a list, short heads its step on the timeline, and doing says what a node in it is busy with. Match model.ProvisionState
export const provisionStates = [
  { state: "pxe", label: "PXE boot", short: "PXE", doing: "Network booting" },
  { state: "ipxe", label: "iPXE", short: "iPXE", doing: "Loading the boot image" },
  { state: "config", label: "fetched config", short: "Config", doing: "Configuring the OS" },
  { state: "booting", label: "booting", short: "Booting", doing: "Starting services" },
  { state: "failed", label: "failed", short: "Failed", doing: undefined },
  { state: "completed", label: "completed", short: "Completed", doing: undefined },
] as const;

export type ProvisionState = (typeof provisionStates)[number]["state"];

const byState = new Map<string, (typeof provisionStates)[number]>(
  provisionStates.map((s) => [s.state, s]),
);

// provisionState looks a state up by the name the API gives it, which is a plain string and may be one this table does not know
export const provisionState = (state?: string) => byState.get(state ?? "");

export const stateLabel = (state?: string) => provisionState(state)?.label ?? state ?? "";

export function isFinished(state?: string) {
  return state === "completed" || state === "failed";
}

// the states a provision is still going in, in order
export const runningStates = provisionStates.filter((s) => !isFinished(s.state));
