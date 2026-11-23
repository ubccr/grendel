import { LinkAction } from "@/components/actions/action-card";
import { expand } from "@/lib/nodeset";
import { Link } from "@tanstack/react-router";

// NodesConsoleAction links to the serial console, which attaches to one node at a time
export default function NodesConsoleAction({ nodes }: { nodes: string }) {
  let single: string | undefined;
  try {
    const list = expand(nodes);
    if (list.length === 1) single = list[0];
  } catch {
    // an unparsable nodeset just hides the card
  }
  if (!single) return null;

  return (
    <LinkAction
      title="Serial Console"
      description="Watch the node boot or log in on its serial console, over IPMI SOL."
      link={<Link to="/nodes/$node/console" params={{ node: single }} />}
    />
  );
}
