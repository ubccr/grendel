import { toastBmcResponse, type BmcResponse } from "@/lib/bmc";
import { MutationCache, QueryClient, type QueryKey } from "@tanstack/react-query";
import { toast } from "sonner";

declare module "@tanstack/react-query" {
  interface Register {
    mutationMeta: {
      // quietError keeps the failure toast away from a mutation whose caller shows the error itself
      quietError?: boolean;
      // bmc marks a mutation that changes BMCs, so what was read from BMCs is fetched again once it succeeds. Given the action's name, its per-node outcome is toasted under that name; given true, the caller reports the outcome
      bmc?: string | true;
      // changesNothing marks a mutation that leaves what the queries read as it was, such as signing in or out, so nothing is fetched again after it
      changesNothing?: boolean;
    };
  }
}

// how long a read stays fresh. A page opened again within this shows what it has without asking the server
const FRESH_MS = 10_000;

// operation is the API operation a generated query key was made for, such as getV1Nodes, or "" for a key written by hand
export function operation(queryKey: QueryKey): string {
  const head = queryKey[0];
  if (typeof head !== "object" || head === null || !("_id" in head)) return "";
  return typeof head._id === "string" ? head._id : "";
}

// readsBmc is true for the reads that log into BMCs, which take seconds for each one asked
export const readsBmc = (queryKey: QueryKey) => operation(queryKey).startsWith("getV1Bmc");

// readsDatabase is true for the reads answered from Grendel's own database. Reads of BMCs and switches are not, and neither is a query keyed by hand
export function readsDatabase(queryKey: QueryKey) {
  const op = operation(queryKey);
  return op.startsWith("getV1") && !op.startsWith("getV1Bmc") && !op.startsWith("getV1Switch");
}

export function errorToast(error: unknown) {
  const e = error as { title?: string; detail?: string; message?: string };
  toast.error(e.title ?? e.message ?? "The request failed", { description: e.detail });
}

export function createQueryClient(): QueryClient {
  const queryClient: QueryClient = new QueryClient({
    mutationCache: new MutationCache({
      onError: (error, _variables, _result, mutation) => {
        if (!mutation.meta?.quietError) errorToast(error);
      },
      onSuccess: (data, _variables, _result, mutation) => {
        const bmc = mutation.meta?.bmc;
        if (typeof bmc === "string") toastBmcResponse(bmc, data as BmcResponse);
      },
      // any change may show in anything read from Grendel's database, so all of that is fetched again, and the mutation stays pending until it has been so no control shows the old value in between. BMCs are read again only after a mutation that changed them, without being waited on, and switches never are
      onSettled: (_data, error, _variables, _result, mutation) => {
        if (mutation.meta?.changesNothing) return;
        if (error === null && mutation.meta?.bmc !== undefined)
          void queryClient.invalidateQueries({ predicate: (query) => readsBmc(query.queryKey) });
        return queryClient.invalidateQueries({
          predicate: (query) => readsDatabase(query.queryKey),
        });
      },
    }),
    defaultOptions: {
      queries: {
        retry: false,
        staleTime: FRESH_MS,
        refetchOnWindowFocus: false,
        refetchOnReconnect: false,
      },
    },
  });
  return queryClient;
}
