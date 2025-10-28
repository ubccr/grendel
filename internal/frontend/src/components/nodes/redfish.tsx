import { HeartPulse, Power } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "../ui/card";
import { UseQueryResult } from "@tanstack/react-query";
import { GetV1BmcResponse } from "@/openapi/requests";

export default function NodeRedfish({
  redfish,
}: {
  redfish: UseQueryResult<GetV1BmcResponse | undefined, unknown>;
}) {
  return (
    <div>
      {redfish.data && redfish.data.length > 0 ? (
        <div className="grid gap-4 sm:grid-cols-3">
          <Card>
            <CardHeader>
              <CardTitle>Hostname:</CardTitle>
            </CardHeader>
            <CardContent>
              <span className="text-muted-foreground">
                {redfish.data?.[0]?.host_name}
              </span>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>BIOS Version:</CardTitle>
            </CardHeader>
            <CardContent>
              <span className="text-muted-foreground">
                {redfish.data?.[0]?.bios_version}
              </span>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Serial Number:</CardTitle>
            </CardHeader>
            <CardContent>
              <span className="text-muted-foreground">
                {redfish.data?.[0]?.serial_number}
              </span>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Manufacturer:</CardTitle>
            </CardHeader>
            <CardContent>
              <span className="text-muted-foreground">
                {redfish.data?.[0]?.manufacturer}
              </span>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Model:</CardTitle>
            </CardHeader>
            <CardContent>
              <span className="text-muted-foreground">
                {redfish.data?.[0]?.model}
              </span>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Power Status:</CardTitle>
            </CardHeader>
            <CardContent className="flex gap-2">
              <Power
                className={
                  redfish.data?.[0]?.power_status === "On"
                    ? "text-green-600"
                    : "text-red-600"
                }
              />
              <span className="text-muted-foreground">
                {redfish.data?.[0]?.power_status}
              </span>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Health:</CardTitle>
            </CardHeader>
            <CardContent className="flex gap-2">
              <HeartPulse
                className={healthColor(redfish.data?.[0]?.health ?? "")}
              />
              <span className="text-muted-foreground">
                {redfish.data?.[0]?.health}
              </span>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Total Memory:</CardTitle>
            </CardHeader>
            <CardContent>
              <span className="text-muted-foreground">
                {redfish.data?.[0]?.total_memory} GB
              </span>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Processor Count:</CardTitle>
            </CardHeader>
            <CardContent>
              <span className="text-muted-foreground">
                {redfish.data?.[0]?.processor_count} Cores
              </span>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Boot Next:</CardTitle>
            </CardHeader>
            <CardContent>
              <span className="text-muted-foreground">
                {redfish.data?.[0]?.boot_next}
              </span>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Boot Order:</CardTitle>
            </CardHeader>
            <CardContent>
              <span className="text-muted-foreground">
                {redfish.data?.[0]?.boot_order?.join(",")}
              </span>
            </CardContent>
          </Card>
        </div>
      ) : (
        <div className="flex justify-center">
          <span className="text-muted-foreground p-4 text-center">
            No redfish data could be retrieved from the node. Check the server
            logs for more details.
          </span>
        </div>
      )}
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
