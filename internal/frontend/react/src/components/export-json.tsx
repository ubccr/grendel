import { useHostFind } from "@/openapi/queries";

export default function ExportJSON({ nodes }: { nodes: string }) {
    const { data } = useHostFind({ path: { nodeSet: nodes } });
    return <pre>{JSON.stringify(data, null, 4)}</pre>;
}
