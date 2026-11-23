import { ErrorPanel } from "@/components/error-panel";
import TrustChain from "@/components/secureboot/chain";
import EnrollDialog from "@/components/secureboot/enroll-dialog";
import KeyList from "@/components/secureboot/key-list";
import ResetDialog from "@/components/secureboot/reset-dialog";
import { Skeleton } from "@/components/ui/skeleton";
import { fold } from "@/lib/nodeset";
import { secureBootKeysOptions } from "@/lib/queries";
import { coverage, writable, type Writable } from "@/lib/secureboot";
import { useQuery } from "@tanstack/react-query";
import { KeyRound, TriangleAlert } from "lucide-react";
import { useState } from "react";

const headings: Record<Writable, string> = {
  PK: "Platform key",
  KEK: "Key exchange keys",
  db: "Allowed to boot",
  dbx: "Revoked",
};

type Props = {
  nodes: string;
};

// With many nodes a key is one row however many nodes hold it
export default function KeysPanel({ nodes }: Props) {
  const [db, setDb] = useState<Writable>("db");
  const trusted = useQuery(secureBootKeysOptions(nodes, "trusted"));
  // dbx waits its turn behind the databases shown first, unless it is the one asked for
  const revoked = useQuery({
    ...secureBootKeysOptions(nodes, "revoked"),
    enabled: db === "dbx" || !trusted.isPending,
  });
  const read = (d: Writable) => (d === "dbx" ? revoked : trusted);

  const rows = coverage([...(trusted.data?.results ?? []), ...(revoked.data?.results ?? [])]);
  const counts: Record<string, number> = {};
  for (const r of rows) counts[r.db] = (counts[r.db] ?? 0) + 1;

  const unread: Partial<Record<Writable, "reading" | "failed">> = {};
  for (const d of writable) {
    const q = read(d);
    if (q.isPending) unread[d] = "reading";
    else if (q.isError || !q.data.results?.some((r) => r.success)) unread[d] = "failed";
  }

  const current = read(db);
  const results = current.data?.results ?? [];
  const reached = results.filter((r) => r.success).map((r) => r.node ?? "");
  const failed = results.filter((r) => !r.success);
  const shown = rows.filter((r) => r.db === db);
  const problem = (current.error as { detail?: string; title?: string } | null) ?? undefined;

  return (
    <section aria-labelledby="sb-keys" className="grid gap-4">
      <TrustChain
        counts={counts}
        unread={unread}
        selected={db}
        onSelect={setDb}
        unit={reached.length > 1 ? "distinct key" : "key"}
      />

      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h3 id="sb-keys" className="text-base font-semibold">
            {headings[db]}
          </h3>
          <p className="text-sm text-muted-foreground">
            {unread[db] === "reading"
              ? `Reading ${db}…`
              : unread[db] === "failed"
                ? `${db} could not be read`
                : `${db} holds ${shown.length} ${shown.length === 1 ? "key" : "keys"}${
                    reached.length > 1 ? ` across the ${reached.length} nodes that answered` : ""
                  }`}
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          <ResetDialog nodes={nodes} />
          <EnrollDialog nodes={nodes} db={db} />
        </div>
      </div>

      {unread[db] === "reading" ? (
        <div className="grid gap-2" aria-busy>
          <Skeleton className="h-16 w-full" />
          <Skeleton className="h-16 w-full" />
          {db === "dbx" && (
            <p className="text-xs text-muted-foreground">
              A BMC hands over each revoked hash in a request of its own, so a long list takes a
              while.
            </p>
          )}
        </div>
      ) : unread[db] === "failed" ? (
        <ErrorPanel
          icon={KeyRound}
          title={`Could not read ${db}`}
          message={problem?.detail ?? problem?.title ?? failed[0]?.message}
        />
      ) : (
        <>
          {failed.length > 0 && (
            <p className="flex items-start gap-1.5 text-xs text-amber-700 dark:text-amber-300">
              <TriangleAlert className="mt-px size-3.5 shrink-0" />
              <span>
                {db} could not be read on {fold(failed.map((r) => r.node ?? ""))}:{" "}
                {failed[0].message || "the server gave no reason"}
              </span>
            </p>
          )}
          <KeyList db={db} rows={shown} reached={reached} />
        </>
      )}
    </section>
  );
}
