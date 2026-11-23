import { postV1BmcPowerMutation } from "@/client/@tanstack/react-query.gen";
import { ConfirmAction } from "@/components/actions/action-card";
import NodesetView from "@/components/nodeset-view";

export default function NodesBmcPowerAction({ nodes }: { nodes: string }) {
  return (
    <ConfirmAction
      title="BMC Reboot"
      description="Restart the BMCs. The hosts keep running, but the BMCs are unreachable for a few minutes."
      label="Reboot"
      mutation={{ ...postV1BmcPowerMutation(), meta: { bmc: "BMC reboot" } }}
      variables={{ query: { nodeset: nodes }, body: { target: "bmc" } }}
    >
      These BMCs will be unavailable while they reboot:
      <NodesetView nodes={nodes} />
    </ConfirmAction>
  );
}
