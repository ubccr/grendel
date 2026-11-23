import { postV1BmcConfigureAutoMutation } from "@/client/@tanstack/react-query.gen";
import { ConfirmAction } from "@/components/actions/action-card";
import NodesetView from "@/components/nodeset-view";

export default function NodesAutoConfigAction({ nodes }: { nodes: string }) {
  return (
    <ConfirmAction
      title="BMC Auto Configure"
      description="Set the iDRAC to auto configure once, from DHCP, on the node's next boot."
      label="Enable"
      mutation={{ ...postV1BmcConfigureAutoMutation(), meta: { bmc: "Auto configure" } }}
      variables={{ query: { nodeset: nodes } }}
    >
      NIC.1.AutoConfig will be set to Enable Once on:
      <NodesetView nodes={nodes} />
    </ConfirmAction>
  );
}
