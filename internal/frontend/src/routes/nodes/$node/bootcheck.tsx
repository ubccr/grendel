import { ArchToggle } from "@/components/boot/arch-toggle";
import { FileDetail } from "@/components/boot/file-detail";
import { Manifest } from "@/components/boot/manifest";
import { Rendered } from "@/components/boot/rendered";
import { ErrorPanel } from "@/components/error-panel";
import { Skeleton } from "@/components/ui/skeleton";
import { arches, problems, stepFile } from "@/lib/boot-check";
import { bootCheckOptions, nodeOptions } from "@/lib/queries";
import { cn } from "@/lib/utils";
import { useQuery, useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { FileX2, HardDriveDownload } from "lucide-react";
import { z } from "zod";

export const Route = createFileRoute("/nodes/$node/bootcheck")({
  staticData: { crumb: "Boot check" },
  component: RouteComponent,
  validateSearch: z.object({
    // open is the step shown on the right: the kind a template renders as, such as ipxe, or a boot file
    open: z.string().optional(),
    arch: z.enum(arches).optional(),
  }),
});

const s = (n: number) => (n === 1 ? "" : "s");

function RouteComponent() {
  const { node } = Route.useParams();
  const search = Route.useSearch();
  const navigate = Route.useNavigate();
  const { data: hosts } = useSuspenseQuery(nodeOptions(node));
  const host = hosts[0];

  // a node's own arch always wins on the server, so the choice is only offered, and only sent, without one
  const arch = host?.arch ? undefined : (search.arch ?? "x86_64");
  const open = search.open ?? "ipxe";
  const setOpen = (step: string) =>
    void navigate({ search: (p) => ({ ...p, open: step }), replace: true });

  const { data, error, isPending, isFetching } = useQuery(bootCheckOptions(node, arch));
  const check = data?.nodes?.[0];

  if (isPending)
    return (
      <div className="grid gap-6 lg:grid-cols-[minmax(0,22rem)_minmax(0,1fr)]">
        <Skeleton className="h-64 w-full" />
        <Skeleton className="h-96 w-full" />
      </div>
    );
  if (error || !check)
    return (
      <ErrorPanel
        title="Could not check the boot"
        message={error?.detail || error?.title}
        icon={FileX2}
      />
    );

  const fileName = stepFile(open);
  const file = check.files?.find((f) => f.name === fileName);

  const failed = problems(check).length;
  const files = check.files?.length ?? 0;
  const templates = check.templates?.length ?? 0;

  return (
    <div className={cn("my-2 grid gap-6 transition-opacity", isFetching && "opacity-70")}>
      <header className="flex flex-wrap items-start justify-between gap-4">
        <div className="grid gap-1">
          <h2
            className={cn("text-xl font-semibold tracking-tight", failed > 0 && "text-destructive")}
          >
            {failed > 0
              ? `${failed} problem${s(failed)} would stop this boot`
              : `Ready to boot ${check.image}`}
          </h2>
          <p className="text-sm text-muted-foreground">
            {check.error
              ? check.error
              : `Rendered image ${check.image} with ${files} file${s(files)} and ${templates} template${s(templates)}`}
          </p>
        </div>
        {arch && (
          <ArchToggle
            options={arches}
            value={arch}
            onChange={(a) => void navigate({ search: (p) => ({ ...p, arch: a }), replace: true })}
          />
        )}
      </header>

      {check.error ? (
        <ErrorPanel
          title="Nothing to send this node"
          message={check.error}
          icon={HardDriveDownload}
        />
      ) : (
        <div className="grid gap-6 lg:grid-cols-[minmax(0,22rem)_minmax(0,1fr)]">
          <Manifest check={check} open={open} onOpen={setOpen} />
          <div className="min-w-0">
            {fileName !== undefined ? (
              file ? (
                <FileDetail file={file} />
              ) : (
                <ErrorPanel
                  title="No such boot file"
                  message={`The image has no ${fileName}.`}
                  icon={FileX2}
                />
              )
            ) : (
              <Rendered
                key={`${open}/${arch}`}
                node={node}
                kind={open}
                arch={arch}
                onKind={setOpen}
              />
            )}
          </div>
        </div>
      )}
    </div>
  );
}
