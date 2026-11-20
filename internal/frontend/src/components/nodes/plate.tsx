import type { Host, ProvisionStatus } from "@/client";
import { getV1NodesProvisionStatusOptions } from "@/client/@tanstack/react-query.gen";
import { Fact, Unset } from "@/components/detail-header";
import { cn, relativeTime } from "@/lib/utils";
import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";

type Nic = NonNullable<Host["interfaces"]>[number];

// bootNic and bmcNic pick the interfaces the server does: the first one not marked BMC boots, the first one marked BMC is the BMC
export function bootNic(nics?: Nic[]) {
  return nics?.find((n) => !n.bmc);
}
export function bmcNic(nics?: Nic[]) {
  return nics?.find((n) => n.bmc);
}

// match model.ProvisionState
const provisionText: Record<string, string> = {
  pxe: "PXE booting",
  ipxe: "loading image",
  config: "fetched config",
  booting: "booting",
  failed: "failed",
  completed: "completed",
};

// ProvisionLight is whether a node provisions: a green light when it does
export function ProvisionLight({ on, className }: { on?: boolean; className?: string }) {
  return (
    <span
      aria-hidden
      className={cn(
        "size-2 shrink-0 rounded-full",
        on ? "bg-emerald-500" : "bg-muted-foreground/40",
        className,
      )}
    />
  );
}

// LastProvision is how a node's last provision went and when, linked to its provision tab
export function LastProvision({ status, node }: { status: ProvisionStatus; node: string }) {
  const updated = status.updated_at ? relativeTime(status.updated_at) : undefined;
  return (
    <Link
      to="/nodes/$node/provision"
      params={{ node }}
      className={cn(
        "underline-offset-4 hover:underline",
        status.state === "failed" && "text-destructive",
      )}
      title={updated?.title}
    >
      {provisionText[status.state ?? ""] ?? status.state} {updated?.text}
    </Link>
  );
}

// NodeFacts are the facts about a node that every tab of the node view needs at a glance, like the asset label on the front of a server
export function NodeFacts({ host, name }: { host?: Host; name: string }) {
  const { data: statuses } = useQuery(
    getV1NodesProvisionStatusOptions({ query: { nodeset: name } }),
  );
  const status = statuses?.[0];
  const boot = bootNic(host?.interfaces);
  const bmc = bmcNic(host?.interfaces);
  const bmcHost = bmc?.fqdn || bmc?.ip?.split("/")[0];
  const isSwitch = host?.tags?.includes("switch");

  return (
    <>
      {!isSwitch && (
        <Fact label="Provision">
          <span className="inline-flex items-center gap-1.5">
            <ProvisionLight on={host?.provision} />
            {host?.provision ? "On" : "Off"}
          </span>
        </Fact>
      )}
      {!isSwitch && (
        <Fact label="Image">
          {host?.boot_image ? (
            <Link
              to="/images/$image"
              params={{ image: host.boot_image }}
              className="underline-offset-4 hover:underline"
            >
              {host.boot_image}
            </Link>
          ) : (
            <Unset>default</Unset>
          )}
        </Fact>
      )}
      <Fact label={isSwitch ? "Management" : "Boot interface"}>
        {boot?.mac || boot?.ip ? (
          <span className="inline-flex gap-3 font-mono">
            {boot.mac && <span>{boot.mac}</span>}
            {boot.ip && <span>{boot.ip}</span>}
          </span>
        ) : (
          <Unset>none</Unset>
        )}
      </Fact>
      {!isSwitch && (
        <Fact label="BMC">
          {bmcHost ? (
            <a
              href={`https://${bmcHost}`}
              target="_blank"
              rel="noreferrer"
              className="font-mono underline-offset-4 hover:underline"
            >
              {bmcHost}
            </a>
          ) : (
            <Unset>none</Unset>
          )}
        </Fact>
      )}
      {status?.state && (
        <Fact label="Last provision">
          <LastProvision status={status} node={name} />
        </Fact>
      )}
    </>
  );
}
