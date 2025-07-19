import { createFileRoute } from "@tanstack/react-router";
import ImageForm from "@/components/images/form";
import { useGetV1ImagesFindSuspense } from "@/openapi/queries/suspense";
import AuthRedirect from "@/auth";
import ImageActions from "@/components/images/actions";
import ActionsSheet from "@/components/actions-sheet";
import { QuerySuspense } from "@/components/query-suspense";

export const Route = createFileRoute("/images/$image")({
  component: RouteComponent,
  beforeLoad: AuthRedirect,
});

function RouteComponent() {
  return (
    <>
      <div className="p-4">
        <QuerySuspense>
          <Form />
        </QuerySuspense>
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
