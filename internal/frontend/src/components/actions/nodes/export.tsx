import { getV1NodesFindOptions } from "@/client/@tanstack/react-query.gen";
import ActionCard from "@/components/actions/action-card";
import { ExportJson } from "@/components/actions/export-json";
import { Button } from "@/components/ui/button";

export default function NodesExportAction({ nodes }: { nodes: string }) {
  return (
    <ActionCard
      title="Export JSON"
      description="Show the nodes as JSON, ready to copy or import elsewhere."
      footer={
        <ExportJson
          noun="node"
          options={getV1NodesFindOptions({ query: { nodeset: nodes } })}
          trigger={<Button>Export</Button>}
        />
      }
    />
  );
}
