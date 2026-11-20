import { ErrorPanel } from "@/components/error-panel";
import { KeyRound, Tag, Unplug } from "lucide-react";

// SwitchProblem explains why a switch couldn't be read and what to do about it, from the stable code the API sends with the error: error_code on a scan, type on an HTTP error
export function SwitchProblem({
  what,
  code,
  title,
  detail,
  address,
}: {
  // what is what was being read, like "LLDP neighbors"
  what: string;
  code?: string;
  title?: string;
  detail?: string;
  address?: string;
}) {
  switch (code) {
    case "unknown_nos":
      return (
        <ErrorPanel
          icon={Tag}
          title="Grendel doesn't know what this switch runs"
          message={`Tag the node with arista, dellos10 or sonic so Grendel knows how to ask it for ${what}.`}
        />
      );
    case "no_credentials":
      return (
        <ErrorPanel
          icon={KeyRound}
          title="No switch login set"
          message="Set switch.username and switch.password in grendel.toml, then restart grendeld. Arista and SONiC switches need them for gNMI."
        />
      );
    case "no_interfaces":
      return (
        <ErrorPanel
          icon={Unplug}
          title="This switch has no address"
          message="Add the switch's management interface. Grendel connects to its first interface."
        />
      );
    case "unreachable":
      return (
        <ErrorPanel
          icon={Unplug}
          title="The switch didn't answer"
          message={`Grendel couldn't read ${what} from ${address ?? "the switch"}. Check that its type tag is right, that gNMI (Arista, SONiC) or SNMP (Dell OS10) is enabled on it, and that the [switch] settings in grendel.toml match.${detail ? ` The error was: ${detail}` : ""}`}
        />
      );
  }
  return <ErrorPanel title={`Couldn't read ${what}`} message={detail || title} />;
}
