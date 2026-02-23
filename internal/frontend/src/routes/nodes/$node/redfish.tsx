import { getV1Bmc } from "@/client";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { createFileRoute } from "@tanstack/react-router";
import { HeartPulse, Power } from "lucide-react";

export const Route = createFileRoute("/nodes/$node/redfish")({
  component: RouteComponent,
  loader: ({ params: { node } }) => getV1Bmc({ query: { nodeset: node } }),
  staleTime: 60_000,
});

function RouteComponent() {
  const { data } = Route.useLoaderData();

  return (
    <div className="grid gap-4 sm:grid-cols-3">
      <Card>
        <CardHeader>
          <CardTitle>Hostname:</CardTitle>
        </CardHeader>
        <CardContent>
          <span className="text-muted-foreground">{data?.[0].host_name}</span>
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>BIOS Version:</CardTitle>
        </CardHeader>
        <CardContent>
          <span className="text-muted-foreground">{data?.[0].bios_version}</span>
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>Serial Number:</CardTitle>
        </CardHeader>
        <CardContent>
          <span className="text-muted-foreground">{data?.[0].serial_number}</span>
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>Manufacturer:</CardTitle>
        </CardHeader>
        <CardContent>
          <span className="text-muted-foreground">{data?.[0].manufacturer}</span>
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>Model:</CardTitle>
        </CardHeader>
        <CardContent>
          <span className="text-muted-foreground">{data?.[0].model}</span>
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>Power Status:</CardTitle>
        </CardHeader>
        <CardContent className="flex gap-2">
          <Power className={data?.[0].power_status === "On" ? "text-green-600" : "text-red-600"} />
          <span className="text-muted-foreground">{data?.[0].power_status}</span>
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>Health:</CardTitle>
        </CardHeader>
        <CardContent className="flex gap-2">
          <HeartPulse className={healthColor(data?.[0].health ?? "")} />
          <span className="text-muted-foreground">{data?.[0].health}</span>
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>Total Memory:</CardTitle>
        </CardHeader>
        <CardContent>
          <span className="text-muted-foreground">{data?.[0].total_memory} GB</span>
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>Processor Count:</CardTitle>
        </CardHeader>
        <CardContent>
          <span className="text-muted-foreground">{data?.[0].processor_count} Cores</span>
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>Boot Next:</CardTitle>
        </CardHeader>
        <CardContent>
          <span className="text-muted-foreground">{data?.[0].boot_next}</span>
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>Boot Order:</CardTitle>
        </CardHeader>
        <CardContent>
          <span className="text-muted-foreground">{data?.[0].boot_order?.join(",")}</span>
        </CardContent>
      </Card>
    </div>
  );
}

function healthColor(health: string) {
  switch (health) {
    case "Critical":
      return "text-red-600";
    case "Warning":
      return "text-yellow-500";
    case "OK":
      return "text-green-600";
    default:
      return "text-gray-600";
  }
}
