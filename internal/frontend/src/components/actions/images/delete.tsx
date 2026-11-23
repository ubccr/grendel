import { deleteV1ImagesMutation } from "@/client/@tanstack/react-query.gen";
import { ConfirmAction } from "@/components/actions/action-card";

export default function ImagesDeleteAction({ images }: { images: string }) {
  return (
    <ConfirmAction
      title="Delete"
      description="Permanently remove the selected images from Grendel."
      label="Delete"
      destructive
      mutation={deleteV1ImagesMutation()}
      variables={{ query: { names: images } }}
    >
      These images will be permanently removed from Grendel: {images}
    </ConfirmAction>
  );
}
