import { createFileRoute } from "@tanstack/react-router";

import { useEffect, useState } from "react";

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import NodeForm from "@/components/nodes/form";
import NodeRedfish from "@/components/nodes/redfish";
import { LoaderCircle, RefreshCw } from "lucide-react";
import AuthRedirect from "@/auth";
import {
  useGetV1Bmc,
  useGetV1BmcMetrics,
  useGetV1NodesFind,
} from "@/openapi/queries";
import { z } from "zod";
import { TestLineChart } from "@/components/nodes/line-chart";
import LldpTable from "@/components/nodes/lldp-table";
import { useNavigate } from "@tanstack/react-router";
import { toast } from "sonner";
import { Card, CardContent } from "@/components/ui/card";
import ActionsSheet from "@/components/actions-sheet";
import { Button } from "@/components/ui/button";
import NodeActions from "@/components/nodes/actions";
import { Progress } from "@/components/ui/progress";

export const Route = createFileRoute("/nodes/$node")({
  component: RouteComponent,
  validateSearch: z.object({
    tab: z.string().optional().catch("node"),
  }),
  beforeLoad: AuthRedirect,
});

function RouteComponent() {
  return (
    <div>
      <Form />
    </div>
  );
}

type ChartDataMap = Map<string, ChartData>;

type ChartData = {
  title: string;
  description: string;
  xAxisKey: string;
  yAxisKey: string;
  data: {
    Time: string;
    Value: number;
  }[];
};

function Form() {
  const { node } = Route.useParams();
  const search = Route.useSearch();
  const navigate = useNavigate({ from: Route.fullPath });

  const [chartData, setChartData] = useState<ChartDataMap>(new Map());
  const grendel_host = useGetV1NodesFind({
    query: { nodeset: node },
  });
  const redfish = useGetV1Bmc({ query: { nodeset: node } }, undefined, {
    staleTime: 5 * 60 * 1000,
  });

  const reports = useGetV1BmcMetrics({ query: { nodeset: node } }, undefined, {
    staleTime: 30 * 60 * 1000,
  });

  useEffect(() => {
    if (grendel_host.error) {
      toast.error(grendel_host.error.title, {
        description: grendel_host.error.detail,
      });
    }
  }, [grendel_host.error]);

  // TODO: move logic into backend

  useEffect(() => {
    if (!reports.isSuccess || !reports.data) return;
    if (reports.data.length != 1) {
      return;
    }

    const oemSchema = z.object({
      Dell: z.object({
        ContextID: z.string(),
        FQDD: z.string(),
        Label: z.string(),
        Source: z.string(),
      }),
    });

    console.log(reports);

    const charts: ChartDataMap = new Map();
    reports.data?.[0].reports?.forEach((report) => {
      if (!report?.Id?.startsWith("Grendel")) return;

      report?.MetricValues?.forEach((metric) => {
        try {
          const oem = oemSchema.parse(metric.Oem);
          const key = `${report.Id}:${metric.MetricID}:${oem.Dell.FQDD}`;
          const prev = charts.get(key);
          charts.set(key, {
            title: oem.Dell.ContextID ?? "",
            description: metric.MetricID ?? "",
            xAxisKey: "Time",
            yAxisKey: "Value",
            data: [
              ...(prev?.data ?? []),
              {
                Time: new Date(
                  Date.parse(metric.Timestamp ?? ""),
                ).toLocaleTimeString("en-US", {
                  hour: "numeric",
                  minute: "numeric",
                }),
                Value: Number.parseInt(metric.MetricValue ?? ""),
              },
            ],
          });
        } catch {
          return;
        }
      });
    });

    setChartData(charts);
  }, [reports.isSuccess, reports.data]);

  return (
    <Card>
      <CardContent>
        {(grendel_host.isFetching ||
          redfish.isFetching ||
          reports.isFetching) && <Progress className="h-1" />}
        {grendel_host.data && grendel_host.data.length > 0 ? (
          <div>
            <Tabs
              defaultValue={search.tab ?? "node"}
              onValueChange={(v) => navigate({ search: { tab: v } })}
            >
              <div className="grid grid-cols-2 gap-3 pt-2 sm:grid-cols-3">
                <div className="hidden sm:block"></div>
                <div className="sm:text-center">
                  <TabsList>
                    <TabsTrigger value="node">Node</TabsTrigger>
                    {grendel_host.data?.[0].tags?.includes("switch") ? (
                      <TabsTrigger value="lldp">LLDP</TabsTrigger>
                    ) : (
                      <>
                        <TabsTrigger value="redfish">Redfish</TabsTrigger>
                        <TabsTrigger value="reports">Reports</TabsTrigger>
                      </>
                    )}
                  </TabsList>
                </div>
                <div>
                  <div className="flex justify-end gap-2">
                    <ActionsSheet checked={node} length={1}>
                      <NodeActions nodes={node} length={1} />
                    </ActionsSheet>
                    <Button
                      variant="secondary"
                      type="button"
                      onClick={() => {
                        grendel_host.refetch();
                        redfish.refetch();
                        reports.refetch();
                      }}
                    >
                      <RefreshCw
                        className={
                          grendel_host.isFetching || redfish.isFetching
                            ? "animate-spin"
                            : ""
                        }
                      />
                      <span className="sr-only md:not-sr-only">Refresh</span>
                    </Button>
                  </div>
                </div>
              </div>
              <TabsContent value="node">
                <NodeForm
                  data={grendel_host.data?.[0]}
                  reset={grendel_host.isFetched}
                />
              </TabsContent>
              <TabsContent value="redfish">
                <NodeRedfish redfish={redfish} />
              </TabsContent>
              <TabsContent value="reports">
                {reports.isFetching && (
                  <div className="p-4">
                    <LoaderCircle className="mx-auto animate-spin" />
                  </div>
                )}
                <div className="grid gap-4 sm:grid-cols-3">
                  {Array.from(chartData).map(([, chart], i) => (
                    <TestLineChart
                      key={i}
                      data={chart.data}
                      XAxisKey={chart.xAxisKey}
                      YAxisKey={chart.yAxisKey}
                      title={chart.title}
                      description={chart.description}
                    />
                  ))}
                  {chartData.size === 0 && (
                    <span className="text-muted-foreground col-span-4 p-4 text-center">
                      No reports could be retrieved from the bmc. Please see our
                      docs for help configuring custom metric reports.
                    </span>
                  )}
                </div>
              </TabsContent>
              <TabsContent value="lldp">
                <LldpTable node={node} />
              </TabsContent>
            </Tabs>
          </div>
        ) : (
          <div className="flex justify-center">
            <span className="text-muted-foreground p-4 text-center">
              404 Node not found.
            </span>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
