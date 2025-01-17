import { createFileRoute } from "@tanstack/react-router";
import NodeForm from "@/components/nodes/form";

export const Route = createFileRoute("/add/node")({
    component: RouteComponent,
});

function RouteComponent() {
    return (
        <div className="p-4">
            <NodeForm />
        </div>
    );
}
