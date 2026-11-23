import { postV1BmcSecurebootKeysDeleteMutation } from "@/client/@tanstack/react-query.gen";
import { ConfirmDialog } from "@/components/actions/action-card";
import { DataTable, type DataTableActions } from "@/components/data-table/data-table";
import type { DataTableColumnDef } from "@/components/data-table/features";
import { DataTableColumnHeader, type FilterOption } from "@/components/data-table/header";
import NodesetView from "@/components/nodeset-view";
import { Button } from "@/components/ui/button";
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from "@/components/ui/empty";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { fold } from "@/lib/nodeset";
import {
  expiry,
  groupFingerprint,
  hashName,
  normalizeHex,
  signatureTypes,
  type Coverage,
} from "@/lib/secureboot";
import { cn } from "@/lib/utils";
import { useMutation } from "@tanstack/react-query";
import { Check, Copy, LoaderCircle, Trash2 } from "lucide-react";
import { useMemo, useState } from "react";
import { toast } from "sonner";

type Props = {
  db: string;
  rows: Coverage[];
  // reached is every node whose BMC answered, so a key can say which of them lack it
  reached: string[];
  // reading is true while the BMCs are being asked again
  reading?: boolean;
};

type Row = Coverage & { kindLabel: string; missing: string[]; reached: number };

const expiryTone = {
  expired: "text-red-600",
  soon: "text-yellow-600",
  ok: "",
  unknown: "text-muted-foreground",
};

const kindOptions: FilterOption[] = ["Certificate", ...signatureTypes.map(hashName)].map((v) => ({
  value: v,
  label: v,
}));

const base: DataTableColumnDef<Row, any>[] = [
  {
    accessorKey: "subject",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Subject" />,
    cell: ({ row: { original: k } }) =>
      k.kind === "cert" && (
        <span className="block max-w-xs truncate font-medium" title={k.subject}>
          {k.subject || "Unnamed certificate"}
        </span>
      ),
  },
  {
    id: "kind",
    accessorFn: (k) => k.kindLabel,
    filterFn: "equals",
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title="Type" options={kindOptions} />
    ),
  },
  {
    accessorKey: "issuer",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Issuer" />,
    cell: ({ row: { original: k } }) =>
      k.kind === "cert" && (
        <span className="block max-w-xs truncate" title={k.issuer}>
          {k.issuer && k.issuer !== k.subject ? k.issuer : "Self signed"}
        </span>
      ),
  },
  {
    id: "expires",
    accessorFn: (k) => k.notAfter,
    sortFn: "text",
    enableColumnFilter: false,
    header: ({ column }) => <DataTableColumnHeader column={column} title="Expires" />,
    cell: ({ row: { original: k } }) => {
      if (!k.notAfter) return null;
      const e = expiry(k.notAfter);
      return (
        <span className={cn("tabular-nums", expiryTone[e])}>
          {new Date(k.notAfter).toLocaleDateString()}
          {e === "expired" && ", expired"}
        </span>
      );
    },
  },
  {
    accessorKey: "fingerprint",
    // a fingerprint is often pasted with the colons or spaces it was printed with
    filterFn: (row, id, value: string) =>
      row.getValue<string>(id).includes(normalizeHex(String(value))),
    header: ({ column }) => <DataTableColumnHeader column={column} title="Fingerprint" />,
    cell: ({ row }) => <Fingerprint value={row.original.fingerprint} />,
  },
];

const remove: DataTableColumnDef<Row, any> = {
  id: "remove",
  enableHiding: false,
  enableSorting: false,
  enableColumnFilter: false,
  header: () => <span className="sr-only">Remove</span>,
  cell: ({ row }) => <RemoveKey k={row.original} />,
};

const columns: Record<"one" | "many", DataTableColumnDef<Row, any>[]> = {
  one: [
    ...base,
    {
      accessorKey: "id",
      sortFn: "alphanumeric",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Id on the BMC" />,
      cell: ({ row }) => <span className="font-mono text-xs">{row.original.id}</span>,
    },
    remove,
  ],
  many: [
    ...base,
    {
      id: "nodes",
      accessorFn: (k) => k.nodes.length,
      enableColumnFilter: false,
      header: ({ column }) => <DataTableColumnHeader column={column} title="Nodes" />,
      cell: ({ row: { original: k } }) => (
        <div className="max-w-xs">
          <div className="tabular-nums">
            {k.nodes.length} of {k.reached}
          </div>
          {k.missing.length > 0 && (
            <div className="truncate text-xs text-muted-foreground" title={fold(k.missing)}>
              Missing on {fold(k.missing)}
            </div>
          )}
        </div>
      ),
    },
    remove,
  ],
};

const rowId = (k: Row) => `${k.db}/${k.fingerprint || k.nodes[0] + "/" + k.id}`;

export default function KeyList({ db, rows, reached, reading }: Props) {
  const data = useMemo<Row[]>(
    () =>
      rows.map((k) => ({
        ...k,
        kindLabel: k.kind === "cert" ? "Certificate" : hashName(k.type),
        missing: reached.filter((n) => !k.nodes.includes(n)),
        reached: reached.length,
      })),
    [rows, reached],
  );

  if (rows.length === 0)
    return (
      <Empty className="border border-dashed">
        <EmptyHeader>
          <EmptyTitle>No keys in {db}</EmptyTitle>
          <EmptyDescription>{emptyHint(db)}</EmptyDescription>
        </EmptyHeader>
      </Empty>
    );

  const certs = rows.some((k) => k.kind === "cert");
  return (
    <DataTable
      // a table per database, so each opens on the columns its own keys fill and its own filters
      key={db}
      columns={columns[reached.length > 1 ? "many" : "one"]}
      data={data}
      getRowId={rowId}
      progress={reading}
      selectable
      Actions={BulkRemove}
      syncFilters={false}
      initialVisibility={{ subject: certs, issuer: certs, expires: certs }}
    />
  );
}

function emptyHint(db: string) {
  switch (db) {
    case "PK":
      return "Without a platform key the node is in setup mode, and secure boot enforces nothing.";
    case "KEK":
      return "With no key exchange key, db and dbx can only change through the firmware's own setup.";
    case "db":
      return "With secure boot on, a node with an empty db boots nothing.";
    default:
      return "Nothing is revoked.";
  }
}

// Fingerprint shows the start of a fingerprint and copies the whole thing, which is what the cli and openssl print
function Fingerprint({ value }: { value: string }) {
  const [copied, setCopied] = useState(false);
  if (!value)
    return (
      <span className="text-xs text-muted-foreground">
        None reported, so this key is removed by its id
      </span>
    );

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      toast.error("Could not copy", { description: "The browser only allows copying over https." });
    }
  };

  return (
    <Tooltip>
      <TooltipTrigger
        render={
          <button
            type="button"
            onClick={copy}
            className="inline-flex max-w-full items-center gap-1.5 rounded-sm font-mono text-xs text-muted-foreground hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
          />
        }
      >
        <span className="truncate">{groupFingerprint(value.slice(0, 24))}…</span>
        {copied ? <Check className="size-3 text-green-600" /> : <Copy className="size-3" />}
        <span className="sr-only">Copy fingerprint</span>
      </TooltipTrigger>
      <TooltipContent className="max-w-sm font-mono break-all">
        {groupFingerprint(value)}
      </TooltipContent>
    </Tooltip>
  );
}

function RemoveKey({ k }: { k: Coverage }) {
  const { mutate, isPending } = useMutation({
    ...postV1BmcSecurebootKeysDeleteMutation(),
    meta: { bmc: `Remove from ${k.db}` },
  });
  const nodes = fold(k.nodes);
  const body = k.fingerprint
    ? { database: k.db, fingerprint: k.fingerprint }
    : { database: k.db, id: k.id };

  return (
    <ConfirmDialog
      trigger={
        <Button variant="ghost" size="sm" disabled={isPending}>
          {isPending ? <LoaderCircle className="animate-spin" /> : <Trash2 />}
          Remove
        </Button>
      }
      onConfirm={() => mutate({ query: { nodeset: nodes }, body })}
    >
      {removeWarning(k.db)} It will be removed from {k.db} on:
      <NodesetView nodes={nodes} />
    </ConfirmDialog>
  );
}

// BulkRemove takes the ticked keys out in as few requests as it can: every key with a fingerprint goes in one, to all the nodes holding any of them, and a key known only by its id goes to the one node that gave it
const BulkRemove: DataTableActions<Row> = ({ table }) => {
  const { mutateAsync, isPending } = useMutation({
    ...postV1BmcSecurebootKeysDeleteMutation(),
    meta: { bmc: "Remove keys" },
  });
  const keys = table.getSelectedRowModel().flatRows.map((r) => r.original);
  if (keys.length === 0) return <></>;

  const db = keys[0].db;
  const printed = keys.filter((k) => k.fingerprint);
  const byNode = new Map<string, string[]>();
  for (const k of keys)
    if (!k.fingerprint) byNode.set(k.nodes[0], [...(byNode.get(k.nodes[0]) ?? []), k.id]);
  const nodes = fold([...new Set(keys.flatMap((k) => k.nodes))]);

  const remove = async () => {
    if (printed.length > 0)
      await mutateAsync({
        query: { nodeset: fold([...new Set(printed.flatMap((k) => k.nodes))]) },
        body: { database: db, fingerprints: printed.map((k) => k.fingerprint) },
      });
    for (const [node, ids] of byNode)
      await mutateAsync({ query: { nodeset: node }, body: { database: db, ids } });
  };

  return (
    <ConfirmDialog
      trigger={
        <Button variant="destructive" disabled={isPending}>
          {isPending ? <LoaderCircle className="animate-spin" /> : <Trash2 />}
          Remove {keys.length} {keys.length === 1 ? "key" : "keys"}
        </Button>
      }
      // a failure is already toasted, and the keys removed before it are gone from the table once it is read again
      onConfirm={() => void remove().catch(() => {})}
    >
      {removeWarning(db, keys.length > 1)}{" "}
      {keys.length === 1 ? "This key" : `These ${keys.length} keys`} will be removed from {db} on:
      <NodesetView nodes={nodes} />
    </ConfirmDialog>
  );
};

function removeWarning(db: string, many = false) {
  switch (db) {
    case "PK":
      return "Removing the platform key puts the node in setup mode, where secure boot enforces nothing.";
    case "KEK":
      return many
        ? "Without these keys, updates they signed to db and dbx are no longer accepted."
        : "Without this key, updates it signed to db and dbx are no longer accepted.";
    case "db":
      return many
        ? "With secure boot on, anything signed only by these keys stops booting."
        : "With secure boot on, anything signed only by this key stops booting.";
    default:
      return many
        ? "Whatever these entries revoked will be allowed to boot again."
        : "Whatever this entry revoked will be allowed to boot again.";
  }
}
