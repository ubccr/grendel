import { getV1ImagesFind } from "@/client";
import ActionsSheet from "@/components/actions-sheet";
import ImagesDeleteAction from "@/components/actions/images/delete";
import ImagesExportAction from "@/components/actions/images/export";
import ImageForm from "@/components/images/form";
import { Card, CardContent } from "@/components/ui/card";
import AuthRedirect from "@/lib/auth";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/images/$image")({
  component: Form,
  beforeLoad: AuthRedirect,
  loader: ({ params: { image } }) => getV1ImagesFind({ query: { names: image } }),
});

function Form() {
  const image = Route.useLoaderData();
  const imageName = Route.useParams().image;

  return (
    <Card>
      <CardContent>
        {image.data && image.data.length === 1 ? (
          <div>
            <div className="grid grid-cols-2 gap-3 pt-2 sm:grid-cols-3">
              <div className="hidden sm:block"></div>
              <div></div>
              <div className="flex justify-end gap-2">
                <ActionsSheet checked={imageName} length={1}>
                  <div className="mt-4 grid gap-4 sm:grid-cols-2">
                    <ImagesDeleteAction images={imageName} />
                    <ImagesExportAction images={imageName} />
                  </div>
                </ActionsSheet>
              </div>
            </div>
            <div className="pt-2">
              <ImageForm
                data={image.data?.[0]}
                // reset={image.isFetched}
              />
            </div>
          </div>
        ) : (
          <div className="flex justify-center">
            <span className="p-4 text-center text-muted-foreground">404 Image not found.</span>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
