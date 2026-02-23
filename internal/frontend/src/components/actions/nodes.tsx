import JobsDeleteAction from "./jobs/delete";
import NodesBmcPowerAction from "./nodes/bmc-power";
import NodesDeleteAction from "./nodes/delete";
import NodesExportAction from "./nodes/export";
import NodesFirmwareAction from "./nodes/firmware";
import NodesImageAction from "./nodes/image";
import NodesImportConfigAction from "./nodes/import-config";
import NodesJobsAction from "./nodes/jobs";
import NodesOsPowerAction from "./nodes/os-power";
import NodesProvisionAction from "./nodes/provision";
import NodesTagsAction from "./nodes/tags";

export default function NodesAction({ nodes }: { nodes: string }) {
  const list = new Map<string, string[]>();
  nodes.split(",").map((node) => list.set(node, ["JID_CLEARALL"]));

  return (
    <div className="mt-4 grid gap-4 sm:grid-cols-2">
      <NodesDeleteAction nodes={nodes} />
      <NodesExportAction nodes={nodes} />
      <NodesProvisionAction nodes={nodes} />
      <NodesImageAction nodes={nodes} />
      <NodesTagsAction nodes={nodes} />
      <NodesOsPowerAction nodes={nodes} />
      <NodesImportConfigAction nodes={nodes} />
      <NodesBmcPowerAction nodes={nodes} />
      <NodesJobsAction nodes={nodes} />
      <JobsDeleteAction list={Object.fromEntries(list)} />
      <NodesFirmwareAction nodes={nodes} />
    </div>
  );
}
