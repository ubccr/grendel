import { createFileRoute } from "@tanstack/react-router";
import { Suspense } from "react";
import { Loading } from "@/components/loading";
import { ErrorBoundary } from "react-error-boundary";
import { Error } from "@/components/error";
import { toast } from "sonner";
import ImageForm from "@/components/images/form";
import { useGetV1ImagesFindSuspense } from "@/openapi/queries/suspense";
import AuthRedirect from "@/auth";
import ImageActions from "@/components/images/actions";
import ActionsSheet from "@/components/actions-sheet";

export const Route = createFileRoute("/images/$image")({
  component: RouteComponent,
  beforeLoad: AuthRedirect,
});

function RouteComponent() {
  return (
    <>
      <div className="p-4">
        <Suspense fallback={<Loading />}>
          <ErrorBoundary
            fallback={<Error />}
            onError={(error) =>
              toast.error("Error loading response", {
                description: error.message,
              })
            }
          >
            <Form />
          </ErrorBoundary>
        </Suspense>
      </div>
    </>
  );
}

function Form() {
  const { image } = Route.useParams();
  const image_query = useGetV1ImagesFindSuspense({ query: { names: image } });

  return (
    <div className="mx-auto">
      <div className="text-end p-2">
        <ActionsSheet checked={image} length={1}>
          <ImageActions images={image} />
        </ActionsSheet>
      </div>
      <ImageForm data={image_query.data?.[0]} reset={image_query.isFetched} />
    </div>
  );
}
