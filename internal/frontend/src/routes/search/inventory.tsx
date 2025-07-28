import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { createFileRoute, Link } from "@tanstack/react-router";
import { useForm } from "@tanstack/react-form";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useGetV1NodesFind } from "@/openapi/queries";
import { useEffect, useState } from "react";
import { Host } from "@/openapi/requests";
import AuthRedirect from "@/auth";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { Button } from "@/components/ui/button";
import { Info } from "lucide-react";

export const Route = createFileRoute("/search/inventory")({
  component: RouteComponent,
  beforeLoad: AuthRedirect,
});

type HostInventory = {
  node: Host;
  search: string;
};

function RouteComponent() {
  const [query, setQuery] = useState<Array<HostInventory>>(
    Array<HostInventory>
  );
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

  const grendel_node = useGetV1NodesFind(
    {
      query: {
        tags: `grendel:serial=${searchParam},grendel:asset=${searchParam}`,
      },
    },
    undefined,
    {
      refetchOnWindowFocus: false,
      refetchOnMount: false,
    }
  );

  useEffect(() => {
    if (grendel_node.data)
      setQuery((n) => [
        { node: grendel_node.data?.[0] ?? {}, search: searchParam },
        ...n,
      ]);
  }, [grendel_node.data]);

  useEffect(() => {
    if (grendel_node.isError && searchParam !== "")
      setQuery((n) => [
        {
          node: {},
          search: searchParam,
        },
        ...n,
      ]);
  }, [grendel_node.isError, searchParam]);

  return (
    <div className="p-4">
      <form
        onSubmit={(e) => {
          e.preventDefault();
          e.stopPropagation();
          form.handleSubmit();
        }}
      >
        <Card>
          <CardHeader>
            <CardTitle className="flex justify-center">Inventory</CardTitle>
          </CardHeader>
          <CardContent className="grid grid-cols-1 gap-2">
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
                          <Button variant="outline">
                            <Info />
                          </Button>
                        </TooltipTrigger>
                        <TooltipContent>
                          Nodes must be tagged with "grendel:serial=1234" or
                          "grendel:asset=1234" to return here.
                          <br /> This form can be used with a barcode scanner:
                          <br /> Set the scanner to include an enter keystroke
                          or Carriage Return as a suffix, focus the textbox and
                          scan.
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
                  {query.map((query, i) => (
                    <TableRow
                      key={i}
                      className={`${i == 0 && !!query.node.name && "dark:bg-green-800 hover:dark:bg-green-700 bg-green-50 hover:bg-green-100"} ${i == 0 && !query.node.name && "dark:bg-red-800 hover:dark:bg-red-700 bg-red-50 hover:bg-red-100"}`}
                    >
                      <TableCell>{query.search}</TableCell>
                      <TableCell>
                        <Link
                          className="hover:font-medium"
                          to="/nodes/$node"
                          params={{ node: query.node.name ?? "" }}
                        >
                          {query.node.name}
                        </Link>
                      </TableCell>
                      <TableCell>
                        {query.node.tags
                          ?.filter((tag) =>
                            tag.includes("grendel:serial=")
                          )?.[0]
                          ?.replace("grendel:serial=", "")}
                      </TableCell>
                      <TableCell>
                        {query.node.tags
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
    </div>
  );
}
