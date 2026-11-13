import { postV1BmcSecurebootKeysMutation } from "@/client/@tanstack/react-query.gen";
import NodesetView from "@/components/nodeset-view";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { toastBmcResponse } from "@/lib/bmc";
import {
  fingerprint,
  groupFingerprint,
  hashBytes,
  hashName,
  normalizeHex,
  readCertificateFile,
  signatureTypes,
  type SignatureType,
  splitPem,
  writable,
  type Writable,
} from "@/lib/secureboot";
import { useMutation, useQuery } from "@tanstack/react-query";
import { FileKey, LoaderCircle, Plus } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

type Props = {
  nodes: string;
  db: Writable;
  onDone: () => void;
};

// guid takes the forms the server does: plain, in braces, or as a urn
const guid = /^(urn:uuid:)?\{?[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\}?$/i;

const dbHelp: Record<Writable, string> = {
  PK: "The platform key. Most firmware takes a new one only in setup mode, or signed by the current one.",
  KEK: "Keys allowed to change db and dbx. Firmware in user or deployed mode may want the change signed by the platform key.",
  db: "Certificates and hashes allowed to boot. Add the certificate of the key you sign your kernels or shim with.",
  dbx: "Certificates and hashes that never boot, even when db allows them.",
};

// EnrollDialog adds certificates or a hash to one database on every node. Certificates are read in the browser and shown with their fingerprint before anything is sent, so they can be checked against openssl's
export default function EnrollDialog({ nodes, db: initial, onDone }: Props) {
  const [open, setOpen] = useState(false);
  const [db, setDb] = useState<Writable>(initial);
  const [kind, setKind] = useState<"cert" | "hash">(initial === "dbx" ? "hash" : "cert");
  const [pasted, setPasted] = useState("");
  const [file, setFile] = useState<{ name: string; pems: string[] }>();
  const [hash, setHash] = useState("");
  const [hashType, setHashType] = useState<SignatureType>("EFI_CERT_SHA256_GUID");
  const [owner, setOwner] = useState("");
  const { mutate, isPending } = useMutation(postV1BmcSecurebootKeysMutation());

  const pems = file ? file.pems : splitPem(pasted);
  // React Query hashes the key by value, so the fingerprints are worked out again only when the certificates change, and a slow result never lands on newer ones
  const { data: prints } = useQuery({
    queryKey: ["certificate-fingerprints", pems],
    queryFn: () => Promise.all(pems.map(fingerprint)),
    enabled: pems.length > 0,
    staleTime: Infinity,
  });

  const hex = normalizeHex(hash);
  const wantHex = hashBytes[hashType] * 2;
  const hashError =
    hex === ""
      ? undefined
      : !/^[0-9a-f]*$/.test(hex)
        ? "Only hex digits."
        : hex.length !== wantHex
          ? `A ${hashName(hashType)} hash is ${wantHex} hex digits, this is ${hex.length}.`
          : undefined;
  const ownerError = owner && !guid.test(owner.trim()) ? "Not a GUID." : undefined;
  const ready =
    !ownerError && (kind === "cert" ? pems.length > 0 : hex.length === wantHex && !hashError);

  const readFile = async (f?: File) => {
    if (!f) return setFile(undefined);
    const bytes = new Uint8Array(await f.arrayBuffer());
    const found = readCertificateFile(bytes);
    if (found.length === 0)
      toast.error(`${f.name} holds no certificate`, {
        description: "Choose a PEM file, or a single DER certificate.",
      });
    setFile({ name: f.name, pems: found });
  };

  const submit = () => {
    const base = { database: db, owner: owner.trim() || undefined };
    // every certificate goes in one request, so a bundle is one run, and each node says what it enrolled and what it already had
    const body =
      kind === "cert"
        ? { ...base, certificates: pems }
        : { ...base, signature: hex, signature_type: hashType };
    mutate(
      { query: { nodeset: nodes }, body },
      {
        onSuccess: (data) => {
          toastBmcResponse(`Enroll in ${db}`, data);
          setPasted("");
          setFile(undefined);
          setHash("");
          setOwner("");
          setOpen(false);
          onDone();
        },
        onError: (e) => toast.error(e.title, { description: e.detail }),
      },
    );
  };

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        // the database picked follows the place the dialog was opened from, each time it opens
        if (next) {
          setDb(initial);
          setKind(initial === "dbx" ? "hash" : "cert");
        }
        setOpen(next);
      }}
    >
      <DialogTrigger render={<Button />}>
        <Plus />
        Enroll key
      </DialogTrigger>
      <DialogContent className="max-h-dvh overflow-y-auto sm:max-w-xl">
        <DialogHeader>
          <DialogTitle>Enroll a key</DialogTitle>
          <DialogDescription>
            Nodes that already have the key are left alone. It takes effect at the next boot.
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-4">
          <div className="grid gap-1.5">
            <Label htmlFor="enroll-db">Database</Label>
            <Select value={db} onValueChange={(v) => setDb(v as Writable)}>
              <SelectTrigger id="enroll-db" className="w-40">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {writable.map((d) => (
                  <SelectItem key={d} value={d}>
                    {d}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <p className="text-xs text-muted-foreground">{dbHelp[db]}</p>
          </div>

          <Tabs value={kind} onValueChange={(v) => setKind(v as "cert" | "hash")}>
            <TabsList>
              <TabsTrigger value="cert">Certificate</TabsTrigger>
              <TabsTrigger value="hash">Hash</TabsTrigger>
            </TabsList>

            <TabsContent value="cert" className="grid gap-3 pt-2">
              <div className="grid gap-1.5">
                <Label htmlFor="enroll-file">Certificate file</Label>
                <Input
                  id="enroll-file"
                  type="file"
                  accept=".pem,.crt,.cer,.der,application/x-pem-file,application/pkix-cert"
                  onChange={(e) => readFile(e.target.files?.[0])}
                />
              </div>
              {!file && (
                <div className="grid gap-1.5">
                  <Label htmlFor="enroll-pem">Or paste PEM</Label>
                  <Textarea
                    id="enroll-pem"
                    rows={5}
                    className="font-mono text-xs"
                    placeholder="-----BEGIN CERTIFICATE-----"
                    value={pasted}
                    onChange={(e) => setPasted(e.target.value)}
                  />
                </div>
              )}
              {pems.length > 0 && (
                <ul className="grid gap-2 rounded-lg border bg-muted/30 p-3 text-sm">
                  {pems.map((_, i) => (
                    <li key={i} className="flex items-start gap-2">
                      <FileKey className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
                      <div className="min-w-0">
                        <div>
                          Certificate {i + 1}
                          {file && pems.length === 1 ? ` from ${file.name}` : ""}
                        </div>
                        <div className="font-mono text-xs break-all text-muted-foreground">
                          {prints?.[i]
                            ? `SHA-256 ${groupFingerprint(prints[i])}`
                            : "The fingerprint shows over https only"}
                        </div>
                      </div>
                    </li>
                  ))}
                </ul>
              )}
            </TabsContent>

            <TabsContent value="hash" className="grid gap-3 pt-2">
              <div className="grid gap-1.5">
                <Label htmlFor="enroll-hash-type">Type</Label>
                <Select
                  value={hashType}
                  onValueChange={(v) => setHashType(v as SignatureType)}
                  items={signatureTypes.map((t) => ({ value: t, label: hashName(t) }))}
                >
                  <SelectTrigger id="enroll-hash-type" className="w-56">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {signatureTypes.map((t) => (
                      <SelectItem key={t} value={t}>
                        {hashName(t)}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="grid gap-1.5">
                <Label htmlFor="enroll-hash">Hash</Label>
                <Input
                  id="enroll-hash"
                  className="font-mono text-xs"
                  placeholder={`${wantHex} hex digits`}
                  aria-invalid={!!hashError}
                  value={hash}
                  onChange={(e) => setHash(e.target.value)}
                />
                <p
                  className={
                    hashError ? "text-xs text-destructive" : "text-xs text-muted-foreground"
                  }
                >
                  {hashError ?? "Usually the hash of a boot loader or kernel to revoke in dbx."}
                </p>
              </div>
            </TabsContent>
          </Tabs>

          <div className="grid gap-1.5">
            <Label htmlFor="enroll-owner">Owner GUID (optional)</Label>
            <Input
              id="enroll-owner"
              className="font-mono text-xs"
              placeholder="Left to the BMC"
              aria-invalid={!!ownerError}
              value={owner}
              onChange={(e) => setOwner(e.target.value)}
            />
            {ownerError && <p className="text-xs text-destructive">{ownerError}</p>}
          </div>

          <div className="text-sm text-muted-foreground">
            Enrolls on:
            <NodesetView nodes={nodes} size="sm" />
          </div>
        </div>

        <DialogFooter>
          <DialogClose render={<Button variant="secondary" />}>Cancel</DialogClose>
          <Button disabled={!ready || isPending} onClick={submit}>
            {isPending && <LoaderCircle className="animate-spin" />}
            {kind === "cert" && pems.length > 1 ? `Enroll ${pems.length} keys` : "Enroll key"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
