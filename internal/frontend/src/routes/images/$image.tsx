import { BootImage, getV1ImagesFind } from "@/client";
import { getV1NodesOptions } from "@/client/@tanstack/react-query.gen";
import ActionsSheet from "@/components/actions-sheet";
import ImagesDeleteAction from "@/components/actions/images/delete";
import ImagesExportAction from "@/components/actions/images/export";
import { DetailHeader, Fact, Unset } from "@/components/detail-header";
import ImageForm from "@/components/images/form";
import { ErrorPanel } from "@/components/error-panel";
import { Card, CardContent } from "@/components/ui/card";
import AuthRedirect from "@/lib/auth";
import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import { ImageOff } from "lucide-react";

export const Route = createFileRoute("/images/$image")({
  component: RouteComponent,
  beforeLoad: AuthRedirect,
  loader: ({ params: { image } }) => getV1ImagesFind({ query: { names: image } }),
});

function plural(n: number, one: string) {
  return `${n} ${one}${n === 1 ? "" : "s"}`;
}

// ImageFacts are what an image boots and how widely it is used, at a glance
function ImageFacts({ image }: { image: BootImage }) {
  const { data: nodes } = useQuery({ ...getV1NodesOptions(), staleTime: 60_000 });
  const assigned = nodes?.filter((n) => n.boot_image === image.name).length;
  const initrds = image.initrd?.length ?? 0;
  const templates = Object.keys(image.provision_templates ?? {}).length;

  return (
    <>
      <Fact label="Version">{image.version || <Unset>none</Unset>}</Fact>
      <Fact label="Initrds">{initrds ? initrds : <Unset>none</Unset>}</Fact>
      <Fact label="Templates">{templates ? templates : <Unset>defaults</Unset>}</Fact>
      <Fact label="Signatures">{image.verify ? "Checked" : <Unset>not checked</Unset>}</Fact>
      {assigned !== undefined && (
        <Fact label="Assigned to">
          <span title="Nodes with this image set. Nodes using the default image aren't counted.">
            {assigned ? (
              <Link
                to="/nodes"
                search={{ boot_image: image.name }}
                className="underline-offset-4 hover:underline"
              >
                {plural(assigned, "node")}
              </Link>
            ) : (
              <Unset>no nodes</Unset>
            )}
          </span>
        </Fact>
      )}
    </>
  );
}

function RouteComponent() {
  const { data } = Route.useLoaderData();
  const imageName = Route.useParams().image;
  const image = data?.length === 1 ? data[0] : undefined;

  if (!image)
    return (
      <Card>
        <CardContent className="p-6">
          <ErrorPanel
            title="Image not found"
            message={`No image is named ${imageName}.`}
            icon={ImageOff}
          />
        </CardContent>
      </Card>
    );

  return (
    <Card>
      <CardContent className="p-0">
        <DetailHeader
          title={image.name}
          facts={<ImageFacts image={image} />}
          actions={
            <ActionsSheet checked={imageName} length={1}>
              <div className="grid gap-4 sm:grid-cols-2">
                <ImagesDeleteAction images={imageName} />
                <ImagesExportAction images={imageName} />
              </div>
            </ActionsSheet>
          }
        />
        <div className="px-6 py-5">
          <ImageForm data={image} />
        </div>
      </CardContent>
    </Card>
  );
}
