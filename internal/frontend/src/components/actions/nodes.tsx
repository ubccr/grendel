import { Link } from "@tanstack/react-router";
import { ActionSection, LinkAction } from "./action-card";
import NodesAutoConfigAction from "./nodes/auto-config";
import NodesConsoleAction from "./nodes/console";
import NodesBmcPowerAction from "./nodes/bmc-power";
import NodesDeleteAction from "./nodes/delete";
import NodesExportAction from "./nodes/export";
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
      <LinkAction
        title="Boot Check"
        description="Check that each node's kernel and initrds are on the server and its templates render, before it boots."
        link={<Link to="/nodes/bootcheck/$nodes" params={{ nodes }} />}
      />
      <NodesExportAction nodes={nodes} />
      <NodesDeleteAction nodes={nodes} />

      <ActionSection title="BMC" />
      <NodesOsPowerAction nodes={nodes} />
      <NodesBmcPowerAction nodes={nodes} />
      <NodesIndicatorAction nodes={nodes} />
      <NodesSelAction nodes={nodes} />
      <NodesJobsAction nodes={nodes} />
      <LinkAction
        title="Add interfaces"
        description="Read each node's network interfaces from its BMC, pick one, and add it with an address."
        link={<Link to="/add/node" search={{ tab: "boot", bootStep: "ports", nodes }} />}
      />
      <LinkAction
        title="BIOS Settings"
        description="Compare BIOS settings across the nodes, and stage the same changes on all of them."
        link={<Link to="/nodes/bios/$nodes" params={{ nodes }} />}
      />
      <LinkAction
        title="Secure Boot"
        description="Turn UEFI secure boot on or off, and see which nodes have each key."
        link={<Link to="/nodes/secureboot/$nodes" params={{ nodes }} />}
      />
      <LinkAction
        title="Asset Tag & Location"
        description="Set the rack location from the node names, and asset tags from a template or a CSV of serial numbers."
        link={<Link to="/nodes/identity/$nodes" params={{ nodes }} />}
      />
      <NodesConsoleAction nodes={nodes} />

      <ActionSection title="Dell" />
      <NodesImportConfigAction nodes={nodes} />
      <NodesAutoConfigAction nodes={nodes} />
      <LinkAction
        title="Firmware Updates"
        description="Check the Dell catalog for firmware updates, then install them."
        link={<Link to="/nodes/dell-firmware/$nodes" params={{ nodes }} />}
      />
    </div>
  );
}
