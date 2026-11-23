import { getV1BmcBiosRegistry } from "@/client";
import { BiosAttribute } from "@/lib/bios";
import { useQuery } from "@tanstack/react-query";

type Identity = { model?: string; bios_version?: string; attribute_registry?: string };

// The registry is keyed by what it depends on rather than the node, so every node of one model and BIOS version shares one fetch, as in the server's cache
export function useBiosRegistry(node: string | undefined, identity: Identity | undefined) {
  const query = useQuery({
    queryKey: [
      "bios-registry",
      identity?.model,
      identity?.bios_version,
      identity?.attribute_registry,
    ],
    queryFn: async () => {
      const res = await getV1BmcBiosRegistry({ query: { nodeset: node }, throwOnError: true });
      const result = res.data?.results?.[0];
      if (!result?.success) throw new Error(result?.message || "failed to read the registry");
      return result.data;
    },
    enabled: !!node && !!identity?.attribute_registry,
    staleTime: Infinity,
    retry: false,
  });

  const registry: BiosAttribute[] | undefined = query.data?.available
    ? (query.data.attributes ?? [])
    : undefined;
  return { registry, isLoading: query.isLoading, error: query.error };
}
