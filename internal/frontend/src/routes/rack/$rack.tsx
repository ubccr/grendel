import { createFileRoute } from "@tanstack/react-router";

import { Suspense, useState } from "react";
import { Loading } from "@/components/loading";
import { ErrorBoundary } from "react-error-boundary";
import { toast } from "sonner";
import { Error } from "@/components/error";

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Link } from "@tanstack/react-router";
import { Checkbox } from "@/components/ui/checkbox";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

import { Button } from "@/components/ui/button";

import ProvisionIcon from "@/components/nodes/provision-button";
import { ExternalLink, Settings2 } from "lucide-react";
import TagsList from "@/components/tags";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import ActionsSheet from "@/components/actions-sheet";
import NodeActions from "@/components/nodes/actions";
import { Host } from "@/openapi/requests";
import { useGetV1NodesFindSuspense } from "@/openapi/queries/suspense";
import AuthRedirect from "@/auth";

export const Route = createFileRoute("/rack/$rack")({
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
            <RackTable />
          </ErrorBoundary>
        </Suspense>
      </div>
    </>
  );
}

type rackArr = {
  u: string;
  hosts: Host[];
};

function RackTable() {
  const { rack } = Route.useParams();
  const { data, isSuccess } = useGetV1NodesFindSuspense({
    query: { tags: rack },
  });
  const [view, setView] = useState(["provision", "tags", "bmc"]);
  const [checked, setChecked] = useState<string[]>([]);
  const fields = ["provision", "tags", "firmware", "boot image", "bmc"];

  const arr: Array<rackArr> = [];
  if (isSuccess) {
    for (let x = 42; x >= 3; x--) {
      let str = x.toString();
      if (x < 10) str = "0" + str;

      const found = data?.filter((host) => {
        if (!host.name) return;
        const parts = host.name.split("-");
        if (parts.length < 3) return false;
        if (parts[2] === str) return true;
      });

      arr.push({ u: str, hosts: found ?? [] });
    }
  }
  return (
    <div>
      <Table>
        <TableHeader className="*:text-center">
          <TableRow className="*:my-auto *:text-center">
            <TableHead className="w-16">u</TableHead>
            <TableHead className="*:my-auto *:text-center grid grid-cols-3">
              <div>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="outline" size="sm">
                      <Settings2 />
                      <span className="sr-only sm:not-sr-only">View</span>
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    {fields.map((col, i) => (
                      <DropdownMenuCheckboxItem
                        key={i}
                        className="capitalize"
                        checked={!!view.find((view) => view == col)}
                        onCheckedChange={(value) =>
                          value
                            ? setView([...view, col])
                            : setView(view.filter((view) => view != col))
                        }
                      >
                        {col}
                      </DropdownMenuCheckboxItem>
                    ))}
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
              <span>Node</span>
              <div>
                <ActionsSheet
                  checked={checked.join(",")}
                  length={checked.length}
                >
                  <NodeActions
                    nodes={checked.join(",")}
                    length={checked.length}
                  />
                </ActionsSheet>
              </div>
            </TableHead>
            <TableHead className=" w-16 ">
              <Checkbox />
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {arr.map((u, i) => (
            <TableRow key={i} className="*:text-center">
              <TableCell>{u.u}</TableCell>
              <TableCell className={`grid grid-cols-${u.hosts.length} gap-4`}>
                {u.hosts.map((host, i) => (
                  <div
                    key={i}
                    className="grid grid-cols-1 gap-4 sm:grid-cols-3"
                  >
                    <div className="flex justify-center gap-6 sm:justify-start">
                      {!!view.find((col) => col === "provision") && (
                        <ProvisionIcon
                          provision={host.provision}
                          name={host.name}
                        />
                      )}
                      {!!view.find((col) => col === "firmware") && (
                        <span className="my-auto">{host.firmware}</span>
                      )}
                    </div>
                    <Link
                      to={`/nodes/$node`}
                      params={{ node: host.name ?? "unknown" }}
                      className="my-auto"
                    >
                      {host.name}
                    </Link>
                    <div className="flex justify-center gap-6 sm:justify-end">
                      {!!view.find((col) => col === "tags") && (
                        <TagsList tags={host.tags} />
                      )}
                      {!!view.find((col) => col === "boot image") && (
                        <span className="my-auto">{host.boot_image}</span>
                      )}
                      {!!view.find((col) => col === "bmc") && (
                        <TooltipProvider>
                          <Tooltip>
                            <TooltipTrigger>
                              <Button
                                variant="outline"
                                type="button"
                                size="sm"
                                asChild
                              >
                                <Link
                                  to={
                                    "https://" +
                                    host.interfaces?.filter(
                                      (v) => v?.bmc == true
                                    )?.[0]?.fqdn
                                  }
                                  target="_blank"
                                >
                                  <ExternalLink />
                                </Link>
                              </Button>
                            </TooltipTrigger>
                            <TooltipContent>
                              <span>
                                https://
                                {
                                  host.interfaces?.filter(
                                    (v) => v?.bmc == true
                                  )?.[0]?.fqdn
                                }
                              </span>
                            </TooltipContent>
                          </Tooltip>
                        </TooltipProvider>
                      )}
                    </div>
                  </div>
                ))}
              </TableCell>
              <TableCell>
                <div className="flex gap-1">
                  {u.hosts.map((host, i) => (
                    <Checkbox
                      key={i}
                      onCheckedChange={(e) =>
                        e
                          ? setChecked([host.name ?? "unknown", ...checked])
                          : setChecked(
                              checked.filter((val) => val != host.name)
                            )
                      }
                    />
                  ))}
                </div>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
