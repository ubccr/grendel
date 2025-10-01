import { createFileRoute } from "@tanstack/react-router";
import ImageForm from "@/components/images/form";
import AuthRedirect from "@/auth";
import ImageActions from "@/components/images/actions";
import ActionsSheet from "@/components/actions-sheet";
import { useEffect } from "react";
import { toast } from "sonner";
import { useGetV1ImagesFind } from "@/openapi/queries";
import { Card, CardContent } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";

export const Route = createFileRoute("/images/$image")({
  component: RouteComponent,
  beforeLoad: AuthRedirect,
});

function RouteComponent() {
  return (
    <div>
      <Form />
    </div>
  );
}

function Form() {
  const { image } = Route.useParams();
  const image_query = useGetV1ImagesFind({ query: { names: image } });

  useEffect(() => {
    if (image_query.error) {
      toast.error(image_query.error.title, {
        description: image_query.error.detail,
      });
    }
  }, [image_query.error]);

  return (
    <Card>
      <CardContent>
        {image_query.isFetching && <Progress className="h-1" />}
        {image_query.data && image_query.data.length === 1 ? (
          <div>
            <div className="grid grid-cols-2 gap-3 pt-2 sm:grid-cols-3">
              <div className="hidden sm:block"></div>
              <div></div>
              <div className="flex justify-end gap-2">
                <ActionsSheet checked={image} length={1}>
                  <ImageActions images={image} />
                </ActionsSheet>
              </div>
            </div>
            <div className="pt-2">
              <ImageForm
                data={image_query.data?.[0]}
                reset={image_query.isFetched}
              />
            </div>
          </div>
        ) : (
          <div className="flex justify-center">
            <span className="text-muted-foreground p-4 text-center">
              404 Image not found.
            </span>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
