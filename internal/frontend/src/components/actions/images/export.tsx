import { getV1ImagesFindOptions } from "@/client/@tanstack/react-query.gen";
import ActionCard from "@/components/actions/action-card";
import { ExportJson } from "@/components/actions/export-json";
import { Button } from "@/components/ui/button";

export default function ImagesExportAction({ images }: { images: string }) {
  return (
    <ActionCard
      title="Export JSON"
      description="Show the images as JSON, ready to copy or import elsewhere."
      footer={
        <ExportJson
          noun="image"
          options={getV1ImagesFindOptions({ query: { names: images } })}
          trigger={<Button>Export</Button>}
        />
      }
    />
  );
}
