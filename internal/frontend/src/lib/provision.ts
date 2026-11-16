// what each provision state is called, in the order a node goes through them. Match model.ProvisionState
export const stateLabels: Record<string, string> = {
  pxe: "PXE boot",
  ipxe: "iPXE",
  config: "fetched config",
  booting: "booting",
  failed: "failed",
  completed: "completed",
};
