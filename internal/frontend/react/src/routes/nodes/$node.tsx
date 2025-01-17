import { createFileRoute } from "@tanstack/react-router";

import { Suspense } from "react";
import { Loading } from "@/components/loading";
import { ErrorBoundary } from "react-error-boundary";
import { Error } from "@/components/error";

import { useHostFindSuspense } from "@/openapi/queries/suspense";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { toast } from "sonner";
import NodeForm from "@/components/nodes/form";
import Actions from "@/components/data-table/actions";

export const Route = createFileRoute("/nodes/$node")({
    component: RouteComponent,
});

function RouteComponent() {
    return (
        <>
            <div className="p-4">
                <Suspense fallback={<Loading />}>
                    <ErrorBoundary
                        fallback={<Error />}
                        onError={(error) =>
                            toast.error("Error loading response", {
                                description: error.message,
                            })
                        }>
                        <Form />
                    </ErrorBoundary>
                </Suspense>
            </div>
        </>
    );
}

function Form() {
    const { node } = Route.useParams();
    const { data } = useHostFindSuspense({ path: { nodeSet: node } });

    return (
        <div className="mx-auto">
            <Tabs defaultValue="node" className="w-full">
                <div className="grid grid-cols-3">
                    <div></div>
                    <div className="text-center">
                        <TabsList>
                            <TabsTrigger value="node">Node</TabsTrigger>
                            <TabsTrigger value="redfish">Redfish</TabsTrigger>
                        </TabsList>
                    </div>
                    <div className="text-end">
                        <Actions checked={node} length={1} />
                    </div>
                </div>
                <TabsContent value="node">
                    <NodeForm data={data?.[0]} />
                </TabsContent>
                <TabsContent value="redfish">Refish data goes here.</TabsContent>
            </Tabs>
        </div>
    );
}
