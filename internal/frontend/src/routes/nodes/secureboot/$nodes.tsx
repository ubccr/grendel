import { getV1BmcSecureboot } from "@/client";
import KeysPanel from "@/components/secureboot/keys-panel";
import { look, type Tone } from "@/components/secureboot/state";
import SecureBootToggle from "@/components/secureboot/toggle";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import AuthRedirect from "@/lib/auth";
import { fold } from "@/lib/nodeset";
import type { SecureBootResult } from "@/lib/secureboot";
import { cn } from "@/lib/utils";
import { createFileRoute, Link, useRouter } from "@tanstack/react-router";
import { ServerOff, type LucideIcon } from "lucide-react";

export const Route = createFileRoute("/nodes/secureboot/$nodes")({
  component: RouteComponent,
  beforeLoad: AuthRedirect,
  // one request reads the state and the keys over a single login to each BMC
  loader: ({ params: { nodes } }) => getV1BmcSecureboot({ query: { nodeset: nodes, keys: true } }),
});

type Group = { title: string; icon: LucideIcon; tone: Tone; nodes: string[]; reason?: string };

const unreachable: Tone = { text: "text-red-600", border: "border-red-400/60" };

// groups sorts the nodes by what secure boot does on them, so the ones that differ from the rest are listed apart rather than found by scanning a table. Nodes that could not be read are grouped by the reason, so each reason is shown
function groups(results: SecureBootResult[]) {
  const out = new Map<string, Group>();
  const add = (key: string, node: string, group: Omit<Group, "nodes">) => {
    const g = out.get(key) ?? { ...group, nodes: [] };
    g.nodes.push(node);
    out.set(key, g);
  };
  for (const r of results) {
    const node = r.node ?? "";
    if (!r.success || !r.data) {
      const reason = r.message || "The server gave no reason.";
      add(`failed/${reason}`, node, {
        title: "Could not be read",
        icon: ServerOff,
        tone: unreachable,
        reason,
      });
      continue;
    }
    const l = look(r.data);
    add(l.title, node, { title: l.title, icon: l.icon, tone: l.tone });
  }
  return [...out.entries()]
    .sort(([, a], [, b]) => b.nodes.length - a.nodes.length)
    .map(([key, g]) => ({ key, ...g }));
}

function RouteComponent() {
  const router = useRouter();
  const { nodes } = Route.useParams();
  const { data } = Route.useLoaderData();
  const results = data?.results ?? [];
  const refresh = () => router.invalidate();
  const byState = groups(results);
  const total = results.length;
  // failures are split by reason, but count as one state
  const states = new Set(byState.map((g) => g.title)).size;

  return (
    <div className="grid gap-4">
      <Card>
        <CardHeader className="flex flex-row flex-wrap items-start justify-between gap-4 space-y-0">
          <div className="space-y-1.5">
            <CardTitle>Secure boot</CardTitle>
            <CardDescription>
              {states === 1 && byState.length === 1
                ? `All ${total} nodes: ${byState[0].title.charAt(0).toLowerCase()}${byState[0].title.slice(1)}.`
                : `The ${total} nodes are in ${states} different states.`}
            </CardDescription>
          </div>
          <div className="flex flex-wrap gap-2">
            <SecureBootToggle nodes={nodes} enable={false} variant="secondary" onDone={refresh} />
            <SecureBootToggle nodes={nodes} enable onDone={refresh} />
          </div>
        </CardHeader>
        <CardContent>
          <ul className="grid gap-2 sm:grid-cols-2">
            {byState.map((g) => {
              const Icon = g.icon;
              const folded = fold(g.nodes);
              return (
                <li
                  key={g.key}
                  className="flex min-w-0 items-start gap-3 rounded-lg border px-4 py-3"
                >
                  <Icon className={cn("mt-0.5 size-5 shrink-0", g.tone.text)} />
                  <div className="min-w-0">
                    <div className="text-sm font-medium">
                      {g.title}
                      <span className="ml-2 text-muted-foreground tabular-nums">
                        {g.nodes.length}
                      </span>
                    </div>
                    <div className="truncate text-xs text-muted-foreground" title={folded}>
                      {g.nodes.length === 1 ? (
                        <Link
                          to="/nodes/$node/secureboot"
                          params={{ node: g.nodes[0] }}
                          className="hover:underline"
                        >
                          {folded}
                        </Link>
                      ) : (
                        folded
                      )}
                    </div>
                    {g.reason && (
                      <div className="mt-1 text-xs break-words text-red-600">{g.reason}</div>
                    )}
                  </div>
                </li>
              );
            })}
          </ul>
        </CardContent>
      </Card>

      <Card>
        <CardContent className="pt-6">
          <KeysPanel nodes={nodes} results={results} onChanged={refresh} />
        </CardContent>
      </Card>
    </div>
  );
}
