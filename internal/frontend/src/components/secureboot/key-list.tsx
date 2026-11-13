import { postV1BmcSecurebootKeysDeleteMutation } from "@/client/@tanstack/react-query.gen";
import { ConfirmDialog } from "@/components/actions/action-card";
import NodesetView from "@/components/nodeset-view";
import { Button } from "@/components/ui/button";
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from "@/components/ui/empty";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { toastBmcResponse } from "@/lib/bmc";
import { fold } from "@/lib/nodeset";
import { expiry, groupFingerprint, hashName, type Coverage } from "@/lib/secureboot";
import { useMutation } from "@tanstack/react-query";
import { Check, Copy, LoaderCircle, Trash2 } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

type Props = {
  db: string;
  rows: Coverage[];
  // reached is every node whose BMC answered, so a key can say which of them lack it
  reached: string[];
  onChanged: () => void;
};

const expiryTone = {
  expired: "text-red-600",
  soon: "text-yellow-600",
  ok: "",
  unknown: "text-muted-foreground",
};

// KeyList is the keys of one database, with how many nodes hold each when there is more than one node
export default function KeyList({ db, rows, reached, onChanged }: Props) {
  const many = reached.length > 1;

  if (rows.length === 0)
    return (
      <Empty className="border border-dashed">
        <EmptyHeader>
          <EmptyTitle>No keys in {db}</EmptyTitle>
          <EmptyDescription>{emptyHint(db)}</EmptyDescription>
        </EmptyHeader>
      </Empty>
    );

  return (
    <ul className="divide-y rounded-lg border">
      {rows.map((k) => {
        const missing = reached.filter((n) => !k.nodes.includes(n));
        const e = expiry(k.notAfter);
        return (
          <li
            key={`${k.db}/${k.fingerprint || k.nodes[0] + "/" + k.id}`}
            className="grid gap-x-6 gap-y-2 px-4 py-3 md:grid-cols-[minmax(0,2fr)_minmax(0,1fr)_auto] md:items-center"
          >
            <div className="min-w-0">
              <div className="truncate font-medium" title={k.subject}>
                {k.kind === "cert"
                  ? k.subject || "Unnamed certificate"
                  : `${hashName(k.type)} hash`}
              </div>
              <div className="truncate text-xs text-muted-foreground">
                {k.kind === "hash" ? (
                  k.id && `Id ${k.id}`
                ) : (
                  <>
                    {k.issuer && k.issuer !== k.subject ? `Issued by ${k.issuer}` : "Self signed"}
                    {k.notAfter && (
                      <span className={expiryTone[e]}>, {expiryText(k.notAfter, e)}</span>
                    )}
                  </>
                )}
              </div>
              <Fingerprint value={k.fingerprint} />
            </div>

            <div className="text-sm">
              {many ? (
                <>
                  <div>
                    On {k.nodes.length} of {reached.length} nodes
                  </div>
                  <div className="truncate text-xs text-muted-foreground" title={fold(missing)}>
                    {missing.length === 0
                      ? "Every node that answered"
                      : `Missing on ${fold(missing)}`}
                  </div>
                </>
              ) : (
                k.id && <span className="text-xs text-muted-foreground">Id {k.id} on the BMC</span>
              )}
            </div>

            <RemoveKey k={k} onDone={onChanged} />
          </li>
        );
      })}
    </ul>
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

function expiryText(notAfter: string, e: ReturnType<typeof expiry>) {
  const day = new Date(notAfter).toLocaleDateString();
  if (e === "expired") return `expired ${day}`;
  return `expires ${day}`;
}

// Fingerprint shows the start of a fingerprint and copies the whole thing, which is what the cli and openssl print
function Fingerprint({ value }: { value: string }) {
  const [copied, setCopied] = useState(false);
  if (!value)
    return (
      <div className="mt-1 text-xs text-muted-foreground">
        The BMC reports no fingerprint, so this key can only be removed by its id
      </div>
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
            className="mt-1 inline-flex max-w-full items-center gap-1.5 rounded-sm font-mono text-xs text-muted-foreground hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
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

function RemoveKey({ k, onDone }: { k: Coverage; onDone: () => void }) {
  const { mutate, isPending } = useMutation(postV1BmcSecurebootKeysDeleteMutation());
  const nodes = fold(k.nodes);
  const body = k.fingerprint
    ? { database: k.db, fingerprint: k.fingerprint }
    : { database: k.db, id: k.id };

  return (
    <ConfirmDialog
      trigger={
        <Button
          variant="ghost"
          size="sm"
          disabled={isPending}
          className="justify-self-start md:justify-self-end"
        >
          {isPending ? <LoaderCircle className="animate-spin" /> : <Trash2 />}
          Remove
        </Button>
      }
      onConfirm={() =>
        mutate(
          { query: { nodeset: nodes }, body },
          {
            onSuccess: (data) => {
              toastBmcResponse(`Remove from ${k.db}`, data);
              onDone();
            },
            onError: (e) => toast.error(e.title, { description: e.detail }),
          },
        )
      }
    >
      {removeWarning(k.db)} It will be removed from {k.db} on:
      <NodesetView nodes={nodes} />
    </ConfirmDialog>
  );
}

function removeWarning(db: string) {
  switch (db) {
    case "PK":
      return "Removing the platform key puts the node in setup mode, where secure boot enforces nothing.";
    case "KEK":
      return "Without this key, updates it signed to db and dbx are no longer accepted.";
    case "db":
      return "With secure boot on, anything signed only by this key stops booting.";
    default:
      return "Whatever this entry revoked will be allowed to boot again.";
  }
}
