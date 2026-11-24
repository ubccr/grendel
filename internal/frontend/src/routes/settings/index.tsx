import { getV1SettingsOptions } from "@/client/@tanstack/react-query.gen";
import { ErrorPanel } from "@/components/error-panel";
import { DetailHeader, Fact, Unset } from "@/components/detail-header";
import SettingsEditor from "@/components/settings/editor";
import { Card, CardContent } from "@/components/ui/card";
import { useUser } from "@/hooks/user-provider";
import AuthRedirect from "@/lib/auth";
import { fileName } from "@/lib/settings";
import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { ShieldOff } from "lucide-react";

export const Route = createFileRoute("/settings/")({
  staticData: { crumb: "Settings" },
  component: RouteComponent,
  beforeLoad: AuthRedirect,
  // a failure is shown in the page, so the settings are waited on but not thrown
  loader: ({ context: { queryClient } }) => queryClient.prefetchQuery(getV1SettingsOptions()),
});

function RouteComponent() {
  const { data, error, isPending } = useQuery(getV1SettingsOptions());
  const { user } = useUser();

  if (isPending) return null;
  if (!data)
    return (
      <Card>
        <CardContent className="p-6">
          <ErrorPanel
            title="Could not load the settings"
            message={error?.detail || error?.title}
            icon={ShieldOff}
          />
        </CardContent>
      </Card>
    );

  const all = (data.sections ?? []).flatMap((s) => s.options ?? []);
  const from = (...sources: string[]) => all.filter((o) => sources.includes(o.source ?? "")).length;
  const canEdit = user?.role === "admin";

  return (
    // the card clips its content by default, which would stop the section list and the save bar from staying in view while the page scrolls
    <Card className="overflow-visible">
      <CardContent className="p-0">
        <DetailHeader
          title="Settings"
          facts={
            <>
              <Fact label="Config file">
                {data.config_file ? (
                  <span className="font-mono" title={data.config_file}>
                    {fileName(data.config_file)}
                  </span>
                ) : (
                  <Unset>none</Unset>
                )}
              </Fact>
              <Fact label="Stored in the database">{from("database", "generated")}</Fact>
              <Fact label="Set outside it">{from("file", "env", "flag")}</Fact>
              <Fact label="At their default">{from("default", "inherited")}</Fact>
            </>
          }
        />
        <SettingsEditor listing={data} canEdit={canEdit} />
      </CardContent>
    </Card>
  );
}
