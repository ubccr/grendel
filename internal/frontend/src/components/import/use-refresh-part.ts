import { fold } from "@/lib/nodeset";
import { useQueryClient, type QueryKey } from "@tanstack/react-query";
import { useState } from "react";
import { toast } from "sonner";

// useRefreshPart asks again for some entries of a cached list, such as switches to rescan or BMCs that didn't answer, and puts the answers in place of the old entries without fetching the rest
export function useRefreshPart<D, E>({
  queryKey,
  fetch,
  entries,
  withEntries,
  name,
}: {
  queryKey: QueryKey;
  fetch: (nodeset: string) => Promise<E[]>;
  entries: (data: D) => E[];
  withEntries: (data: D, entries: E[]) => D;
  name: (entry: E) => string | undefined;
}) {
  const queryClient = useQueryClient();
  const [busy, setBusy] = useState<ReadonlySet<string>>(new Set());

  const refresh = async (names: string[]) => {
    setBusy((b) => new Set([...b, ...names]));
    try {
      const fresh = new Map((await fetch(fold(names))).map((e) => [name(e), e]));
      queryClient.setQueryData<D>(
        queryKey,
        (old) =>
          old &&
          withEntries(
            old,
            entries(old).map((e) => fresh.get(name(e)) ?? e),
          ),
      );
    } catch (e) {
      toast.error(`Couldn't ask ${fold(names)} again`, {
        description: (e as { detail?: string }).detail,
      });
    } finally {
      setBusy((b) => new Set([...b].filter((n) => !names.includes(n))));
    }
  };

  return { refresh, busy };
}
