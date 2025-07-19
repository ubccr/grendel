import { createFileRoute } from "@tanstack/react-router";

import { useEffect, useState } from "react";

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import NodeForm from "@/components/nodes/form";
import ActionsSheet from "@/components/actions-sheet";
import NodeActions from "@/components/nodes/actions";
import NodeRedfish from "@/components/nodes/redfish";
import { Button } from "@/components/ui/button";
import { LoaderCircle, RefreshCw } from "lucide-react";
import { useGetV1NodesFindSuspense } from "@/openapi/queries/suspense";
import AuthRedirect from "@/auth";
import { useGetV1Bmc, useGetV1BmcMetrics } from "@/openapi/queries";
import { z } from "zod";
import { TestLineChart } from "@/components/nodes/line-chart";
import { QuerySuspense } from "@/components/query-suspense";

export const Route = createFileRoute("/nodes/$node")({
  component: RouteComponent,
  beforeLoad: AuthRedirect,
});

function RouteComponent() {
  return (
    <>
      <div className="p-4">
        <QuerySuspense>
          <Form />
        </QuerySuspense>
      </div>
    </>
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
  const [chartData, setChartData] = useState<ChartDataMap>(new Map());
  const grendel_host = useGetV1NodesFindSuspense({
    query: { nodeset: node },
  });
  const redfish = useGetV1Bmc({ query: { nodeset: node } }, undefined, {
    staleTime: 5 * 60 * 1000,
  });

  const reports = useGetV1BmcMetrics({ query: { nodeset: node } }, undefined, {
    staleTime: 30 * 60 * 1000,
  });

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
                  Date.parse(metric.Timestamp ?? "")
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
    <div className="mx-auto">
      <Tabs defaultValue="node" className="w-full">
        <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
          <div className="hidden sm:block"></div>
          <div className="sm:text-center">
            <TabsList>
              <TabsTrigger value="node">Node</TabsTrigger>
              <TabsTrigger value="redfish">Redfish</TabsTrigger>
              <TabsTrigger value="reports">Reports</TabsTrigger>
            </TabsList>
          </div>
          <div className="text-end">
            <div className="flex gap-2 justify-end">
              <ActionsSheet checked={node} length={1}>
                <NodeActions nodes={node} length={1} />
              </ActionsSheet>
              <Button
                variant="outline"
                size="sm"
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
                <span className="md:not-sr-only sr-only">Refresh</span>
              </Button>
            </div>
          </div>
        </div>
        <TabsContent value="node">
          {grendel_host.data && grendel_host.data.length > 0 ? (
            <NodeForm
              data={grendel_host.data?.[0]}
              reset={grendel_host.isFetched}
            />
          ) : (
            <div className="flex justify-center">
              <span className="text-center text-muted-foreground p-4">
                404 Node not found.
              </span>
            </div>
          )}
        </TabsContent>
        <TabsContent value="redfish">
          <NodeRedfish redfish={redfish} />
        </TabsContent>
        <TabsContent value="reports">
          {reports.isFetching && (
            <div className="p-4">
              <LoaderCircle className="animate-spin mx-auto" />
            </div>
          )}
          <div className="grid sm:grid-cols-3 gap-4">
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
              <span className="col-span-4 text-center text-muted-foreground p-4">
                No reports could be retrieved from the bmc. Please see our docs
                for help configuring custom metric reports.
              </span>
            )}
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}
