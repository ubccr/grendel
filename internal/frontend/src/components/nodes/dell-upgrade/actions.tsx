import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

import FirmwareForm from "./firmware-form";

export default function DellFirmwareUpgradeActions({
  nodes,
}: {
  nodes: string;
}) {
  return (
    <div className="mt-4 grid gap-4 sm:grid-cols-1">
      <Card>
        <CardHeader>
          <CardTitle>Upgrade Firmware</CardTitle>
          <CardDescription>
            Submit a request to the BMC to download the specified catalog and
            compare available firmware for updates.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <FirmwareForm nodes={nodes} />
        </CardContent>
      </Card>
    </div>
  );
}
