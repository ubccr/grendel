// import { createFileRoute } from '@tanstack/react-router'

// export const Route = createFileRoute('/nodes/$node/reports')({
//   component: RouteComponent,
// })

// function RouteComponent() {
//   return <div>Hello "/nodes/$node/reports"!</div>
// }

// type ChartDataMap = Map<string, ChartData>;

// type ChartData = {
//   title: string;
//   description: string;
//   xAxisKey: string;
//   yAxisKey: string;
//   data: {
//     Time: string;
//     Value: number;
//   }[];
// };

// const [chartData, setChartData] = useState<ChartDataMap>(new Map());
// const grendel_host = useGetV1NodesFind({
//   query: { nodeset: node },
// });
// const redfish = useGetV1Bmc({ query: { nodeset: node } }, undefined, {
//   staleTime: 5 * 60 * 1000,
// });

// const reports = useGetV1BmcMetrics({ query: { nodeset: node } }, undefined, {
//   staleTime: 30 * 60 * 1000,
// });

// useEffect(() => {
//   if (grendel_host.error) {
//     toast.error(grendel_host.error.title, {
//       description: grendel_host.error.detail,
//     });
//   }
// }, [grendel_host.error]);

// TODO: move logic into backend

// useEffect(() => {
//   if (!reports.isSuccess || !reports.data) return;
//   if (reports.data.length != 1) {
//     return;
//   }

//   const oemSchema = z.object({
//     Dell: z.object({
//       ContextID: z.string(),
//       FQDD: z.string(),
//       Label: z.string(),
//       Source: z.string(),
//     }),
//   });

//   console.log(reports);

//   const charts: ChartDataMap = new Map();
//   reports.data?.[0].reports?.forEach((report) => {
//     if (!report?.Id?.startsWith("Grendel")) return;

//     report?.MetricValues?.forEach((metric) => {
//       try {
//         const oem = oemSchema.parse(metric.Oem);
//         const key = `${report.Id}:${metric.MetricID}:${oem.Dell.FQDD}`;
//         const prev = charts.get(key);
//         charts.set(key, {
//           title: oem.Dell.ContextID ?? "",
//           description: metric.MetricID ?? "",
//           xAxisKey: "Time",
//           yAxisKey: "Value",
//           data: [
//             ...(prev?.data ?? []),
//             {
//               Time: new Date(
//                 Date.parse(metric.Timestamp ?? ""),
//               ).toLocaleTimeString("en-US", {
//                 hour: "numeric",
//                 minute: "numeric",
//               }),
//               Value: Number.parseInt(metric.MetricValue ?? ""),
//             },
//           ],
//         });
//       } catch {
//         return;
//       }
//     });
//   });

//   setChartData(charts);
// }, [reports.isSuccess, reports.data]);

// {/*<TabsContent value="reports">
//   <div className="grid gap-4 sm:grid-cols-3">
//     {Array.from(chartData).map(([, chart], i) => (
//       <TestLineChart
//         key={i}
//         data={chart.data}
//         XAxisKey={chart.xAxisKey}
//         YAxisKey={chart.yAxisKey}
//         title={chart.title}
//         description={chart.description}
//       />
//     ))}
//     {chartData.size === 0 && (
//       <span className="text-muted-foreground col-span-4 p-4 text-center">
//         No reports could be retrieved from the bmc. Please see our
//         docs for help configuring custom metric reports.
//       </span>
//     )}
//   </div>
// </TabsContent>*/}
