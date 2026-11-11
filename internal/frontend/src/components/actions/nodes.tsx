import { ActionSection } from "./action-card";
import NodesAutoConfigAction from "./nodes/auto-config";
import NodesBiosAction from "./nodes/bios";
import NodesIdentityAction from "./nodes/identity";
import NodesBmcPowerAction from "./nodes/bmc-power";
import NodesDeleteAction from "./nodes/delete";
import NodesExportAction from "./nodes/export";
import NodesFirmwareAction from "./nodes/firmware";
import NodesImageAction from "./nodes/image";
import NodesImportConfigAction from "./nodes/import-config";
import NodesIndicatorAction from "./nodes/indicator";
import NodesJobsAction from "./nodes/jobs";
import NodesOsPowerAction from "./nodes/os-power";
import NodesProvisionAction from "./nodes/provision";
import NodesSelAction from "./nodes/sel";
import NodesTagsAction from "./nodes/tags";

export default function NodesAction({ nodes }: { nodes: string }) {
  return (
    <div className="grid gap-4 sm:grid-cols-2">
      <ActionSection title="Grendel" />
      <NodesProvisionAction nodes={nodes} />
      <NodesImageAction nodes={nodes} />
      <NodesTagsAction nodes={nodes} />
      <NodesExportAction nodes={nodes} />
      <NodesDeleteAction nodes={nodes} />

      <ActionSection title="BMC" />
      <NodesOsPowerAction nodes={nodes} />
      <NodesBmcPowerAction nodes={nodes} />
      <NodesIndicatorAction nodes={nodes} />
      <NodesSelAction nodes={nodes} />
      <NodesJobsAction nodes={nodes} />
      <NodesBiosAction nodes={nodes} />
      <NodesIdentityAction nodes={nodes} />

      <ActionSection title="Dell" />
      <NodesImportConfigAction nodes={nodes} />
      <NodesAutoConfigAction nodes={nodes} />
      <NodesFirmwareAction nodes={nodes} />
    </div>
  );
}
