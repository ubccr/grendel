import { getV1NodesFind, Host } from "@/client";
import ActionsSheet from "@/components/actions-sheet";
import NodeActions from "@/components/actions/nodes";
import { AnimateIcon } from "@/components/animate-ui/icons/icon";
import { SquareArrowOutUpRight } from "@/components/animate-ui/icons/square-arrow-out-up-right";
import ProvisionIcon from "@/components/nodes/provision-button";
import TagsList from "@/components/tags";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import AuthRedirect from "@/lib/auth";
import { createFileRoute, Link } from "@tanstack/react-router";
import { Settings2 } from "lucide-react";
import { useState } from "react";

export const Route = createFileRoute("/racks/$rack")({
  component: RackTable,
  beforeLoad: AuthRedirect,
  loader: ({ params: { rack } }) => getV1NodesFind({ query: { tags: `rack=${rack}` } }),
});

type rackArr = {
  u: string;
  hosts: Host[];
};

function RackTable() {
  const { data } = Route.useLoaderData();
  const [view, setView] = useState(["provision", "tags", "bmc"]);
  const [checked, setChecked] = useState<string[]>([]);
  const fields = ["provision", "tags", "firmware", "boot image", "bmc"];

  const arr: Array<rackArr> = [];
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

  return (
    <Card>
      <CardContent className="p-2">
        <Table>
          <TableHeader className="*:text-center">
            <TableRow className="*:my-auto *:text-center">
              <TableHead className="w-16">u</TableHead>
              <TableHead className="grid grid-cols-3 *:my-auto *:text-center">
                <div>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button type="button" variant="secondary">
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
                  <ActionsSheet checked={checked.join(",")} length={checked.length}>
                    <NodeActions nodes={checked.join(",")} />
                  </ActionsSheet>
                </div>
              </TableHead>
              <TableHead className="w-16">
                <Checkbox
                  onCheckedChange={(e) => (e ? setChecked(checkAll(arr)) : setChecked([]))}
                />
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {arr.map((u, i) => (
              <TableRow key={i} className="*:text-center">
                <TableCell>{u.u}</TableCell>
                <TableCell className={`grid-cols- grid${u.hosts.length} gap-4`}>
                  {u.hosts.map((host, i) => (
                    <div key={i} className="grid grid-cols-1 gap-4 sm:grid-cols-3">
                      <div className="flex justify-center gap-6 sm:justify-start">
                        {!!view.find((col) => col === "provision") && (
                          <ProvisionIcon provision={host.provision} name={host.name} />
                        )}
                        {!!view.find((col) => col === "firmware") && (
                          <span className="my-auto">{host.firmware}</span>
                        )}
                      </div>
                      <Link
                        to={`/nodes/$node/node`}
                        params={{ node: host.name ?? "unknown" }}
                        className="my-auto"
                      >
                        {host.name}
                      </Link>
                      <div className="flex justify-center gap-6 sm:justify-end">
                        {!!view.find((col) => col === "tags") && (
                          <TagsList tags={host.tags ?? []} />
                        )}
                        {!!view.find((col) => col === "boot image") && (
                          <span className="my-auto">{host.boot_image}</span>
                        )}
                        {!!view.find((col) => col === "bmc") && (
                          <TooltipProvider>
                            <Tooltip>
                              <TooltipTrigger>
                                <Button variant="secondary" type="button" size="icon" asChild>
                                  <Link
                                    to={
                                      "https://" +
                                      host.interfaces?.filter((v) => v?.bmc == true)?.[0]?.fqdn
                                    }
                                    target="_blank"
                                  >
                                    <AnimateIcon animateOnHover loop>
                                      <SquareArrowOutUpRight animation="default-loop" />
                                    </AnimateIcon>
                                  </Link>
                                </Button>
                              </TooltipTrigger>
                              <TooltipContent>
                                <span>
                                  https://
                                  {host.interfaces?.filter((v) => v?.bmc == true)?.[0]?.fqdn}
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
                  <div className="flex justify-center gap-1">
                    {u.hosts.map((host, i) => (
                      <Checkbox
                        key={i}
                        checked={!!checked.find((v) => v === host.name)}
                        onCheckedChange={(e) =>
                          e
                            ? setChecked([host.name ?? "unknown", ...checked])
                            : setChecked(checked.filter((val) => val != host.name))
                        }
                      />
                    ))}
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}

function checkAll(arr: rackArr[]) {
  const allHosts: Array<string> = [];
  arr.forEach((racku) => racku.hosts.forEach((host) => host.name && allHosts.push(host.name)));

  return allHosts;
}
