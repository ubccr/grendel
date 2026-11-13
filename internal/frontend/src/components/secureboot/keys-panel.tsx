import TrustChain from "@/components/secureboot/chain";
import EnrollDialog from "@/components/secureboot/enroll-dialog";
import KeyList from "@/components/secureboot/key-list";
import ResetDialog from "@/components/secureboot/reset-dialog";
import { coverage, type SecureBootResult, type Writable } from "@/lib/secureboot";
import { useState } from "react";

const headings: Record<Writable, string> = {
  PK: "Platform key",
  KEK: "Key exchange keys",
  db: "Allowed to boot",
  dbx: "Revoked",
};

type Props = {
  nodes: string;
  results: SecureBootResult[];
  onChanged: () => void;
};

// KeysPanel is the chain of databases and the keys of the one picked, for one node or many. With many nodes a key is one row however many nodes hold it
export default function KeysPanel({ nodes, results, onChanged }: Props) {
  const [db, setDb] = useState<Writable>("db");
  const rows = coverage(results);
  const reached = results.filter((r) => r.success).map((r) => r.node ?? "");

  const counts: Record<string, number> = {};
  for (const r of rows) counts[r.db] = (counts[r.db] ?? 0) + 1;
  const shown = rows.filter((r) => r.db === db);

  return (
    <section aria-labelledby="sb-keys" className="grid gap-4">
      <TrustChain
        counts={counts}
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
            {db} holds {shown.length} {shown.length === 1 ? "key" : "keys"}
            {reached.length > 1 && ` across the ${reached.length} nodes that answered`}
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          <ResetDialog nodes={nodes} onDone={onChanged} />
          <EnrollDialog nodes={nodes} db={db} onDone={onChanged} />
        </div>
      </div>

      <KeyList db={db} rows={shown} reached={reached} onChanged={onChanged} />
    </section>
  );
}
