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

export const Route = createFileRoute("/search/inventory")({
  component: RouteComponent,
});

function RouteComponent() {
  const [nodes, setNodes] = useState<Array<Host>>(Array<Host>);
  const [params, setParams] = useState({ serial: "", asset_tag: "" });

  const form = useForm({
    defaultValues: {
      serial: "",
      asset_tag: "",
    },
    onSubmit: async ({ value }) => {
      setParams({ serial: value.serial, asset_tag: value.asset_tag });
      form.reset();
    },
  });

  const grendel_node = useGetV1NodesFind(
    {
      query: {
        tags: `grendel:serial=${params.serial},grendel:asset=${params.asset_tag}`,
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
      setNodes((n) => [...(grendel_node.data ?? []), ...n]);
  }, [grendel_node.data]);

  useEffect(() => {
    if (
      grendel_node.isError &&
      (params.serial !== "" || params.asset_tag !== "")
    )
      setNodes((n) => [
        {
          name: "",
          tags: [
            `grendel:serial=${params.serial}`,
            `grendel:asset=${params.asset_tag}`,
          ],
        },
        ...n,
      ]);
  }, [grendel_node.isError, params]);

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
          <CardContent className="grid grid-cols-2 gap-2">
            <form.Field
              name="serial"
              children={(field) => (
                <div>
                  <Label>Serial:</Label>
                  <Input
                    value={field.state.value ?? ""}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                  />
                </div>
              )}
            />
            <form.Field
              name="asset_tag"
              children={(field) => (
                <div>
                  <Label>Asset Tag:</Label>
                  <Input
                    value={field.state.value ?? ""}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                  />
                </div>
              )}
            />
            <input type="submit" className="invisible" />
            <div className="col-span-2">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Node</TableHead>
                    <TableHead>Serial Number</TableHead>
                    <TableHead>Asset Tag</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {nodes.map((node, i) => (
                    <TableRow
                      key={i}
                      className={`${i == 0 && node.name !== "" && "dark:bg-green-800 hover:dark:bg-green-700 bg-green-50 hover:bg-green-100"} ${i == 0 && node.name == "" && "dark:bg-red-800 hover:dark:bg-red-700 bg-red-50 hover:bg-red-100"}`}
                    >
                      <TableCell>
                        <Link
                          className="hover:font-medium"
                          to="/nodes/$node"
                          params={{ node: node.name ?? "" }}
                        >
                          {node.name}
                        </Link>
                      </TableCell>
                      <TableCell>
                        {node.tags
                          ?.filter((tag) =>
                            tag.includes("grendel:serial=")
                          )?.[0]
                          ?.replace("grendel:serial=", "")}
                      </TableCell>
                      <TableCell>
                        {node.tags
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
