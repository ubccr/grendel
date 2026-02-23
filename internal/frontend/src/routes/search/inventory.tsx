import type { Host } from "@/client";
import { getV1NodesFindOptions } from "@/client/@tanstack/react-query.gen";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
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
import { useForm } from "@tanstack/react-form";
import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import { Info } from "lucide-react";
import { useEffect, useState } from "react";

export const Route = createFileRoute("/search/inventory")({
  component: RouteComponent,
  beforeLoad: AuthRedirect,
});

type HostInventory = {
  node: Host;
  search: string;
};

function RouteComponent() {
  const [nodes, setNodes] = useState<Array<HostInventory>>(Array<HostInventory>);
  const [searchParam, setSearchParam] = useState("");

  const form = useForm({
    defaultValues: {
      search: "",
    },
    onSubmit: async ({ value }) => {
      setSearchParam(value.search);
      form.reset();
    },
  });

  const { data, isError } = useQuery({
    ...getV1NodesFindOptions({
      query: {
        tags: `grendel:serial=${searchParam},grendel:asset=${searchParam}`,
      },
    }),
    refetchOnWindowFocus: false,
    refetchOnMount: false,
  });

  useEffect(() => {
    if (!data) return;

    setNodes((n) => [{ node: data?.[0] ?? {}, search: searchParam }, ...n]);
  }, [data]);

  useEffect(() => {
    if (!isError || searchParam === "") return;

    setNodes((n) => [
      {
        node: {},
        search: searchParam,
      },
      ...n,
    ]);
  }, [isError, searchParam]);

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        e.stopPropagation();
        form.handleSubmit();
      }}
    >
      <Card>
        <CardContent className="grid grid-cols-1 gap-2 pt-2">
          <form.Field
            name="search"
            children={(field) => (
              <div>
                <Label>Serial or Asset:</Label>
                <div className="flex gap-2">
                  <Input
                    value={field.state.value ?? ""}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                  />
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Button variant="secondary">
                          <Info />
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent>
                        Nodes must be tagged with "grendel:serial=1234" or "grendel:asset=1234" to
                        return here.
                        <br /> This form can be used with a barcode scanner:
                        <br /> Set the scanner to include an enter keystroke or Carriage Return as a
                        suffix, focus the textbox and scan.
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                </div>
              </div>
            )}
          />
          <input type="submit" className="invisible" />
          <div className="col-span-2">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Search</TableHead>
                  <TableHead>Node</TableHead>
                  <TableHead>Serial Number</TableHead>
                  <TableHead>Asset Tag</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {nodes.map((val, i) => (
                  <TableRow
                    key={i}
                    className={`${i == 0 && !!val.node.name && "bg-green-50 hover:bg-green-100 dark:bg-green-800 hover:dark:bg-green-700"} ${i == 0 && !val.node.name && "bg-red-50 hover:bg-red-100 dark:bg-red-800 hover:dark:bg-red-700"}`}
                  >
                    <TableCell>{val.search}</TableCell>
                    <TableCell>
                      <Link
                        className="hover:font-medium"
                        to="/nodes/$node/node"
                        params={{ node: val.node.name ?? "" }}
                      >
                        {val.node.name}
                      </Link>
                    </TableCell>
                    <TableCell>
                      {val.node.tags
                        ?.filter((tag) => tag.includes("grendel:serial="))?.[0]
                        ?.replace("grendel:serial=", "")}
                    </TableCell>
                    <TableCell>
                      {val.node.tags
                        ?.filter((tag) => tag.includes("grendel:asset="))?.[0]
                        ?.replace("grendel:asset=", "")}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>
    </form>
  );
}
