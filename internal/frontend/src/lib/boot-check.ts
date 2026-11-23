import type { NodeCheckResponse } from "@/client";
import type { Language } from "@/lib/highlighter";

export type NodeCheck = NonNullable<NodeCheckResponse["nodes"]>[number];

// the arches a node with none set can be checked as, matching model.Arches, which is what the server takes arch=all to mean
export const arches = ["x86_64", "aarch64"] as const;
export const archChoices = [...arches, "all"] as const;

const kindLabels: Record<string, string> = {
  ipxe: "iPXE script",
  kickstart: "Kickstart",
  user_data: "cloud-init user-data",
  meta_data: "cloud-init meta-data",
  butane: "Butane config",
  ignition: "Ignition config",
  proxmox: "Proxmox answer file",
};

export const kindLabel = (kind: string) => kindLabels[kind] ?? kind;

// the step open on the right is named by the kind it renders as, or for a boot file by this and the name it is served under
const FILE = "file:";
export const fileStep = (name: string) => FILE + name;
export const stepFile = (step: string) =>
  step.startsWith(FILE) ? step.slice(FILE.length) : undefined;

// butane and ignition are one template before and after translation, so they are opened from one step
export const isButanePair = (kind: string) => kind === "butane" || kind === "ignition";

// a butane template only ever reaches a node translated, so that is what opening it shows
export const viewKind = (kind: string) => (kind === "butane" ? "ignition" : kind);

export function fileLabel(name: string) {
  const sig = name.endsWith(".sig") ? " signature" : "";
  const base = name.replace(/\.sig$/, "");
  if (base === "kernel") return `Kernel${sig}`;
  const initrd = /^initrd-(\d+)$/.exec(base);
  return initrd ? `Initrd ${initrd[1]}${sig}` : name;
}

// every reason a node would not boot, each naming the file or template it is about
export function problems(n: NodeCheck): string[] {
  const out: string[] = [];
  if (n.error) out.push(n.error);
  for (const f of n.files ?? []) if (f.error) out.push(`${fileLabel(f.name ?? "")}: ${f.error}`);
  for (const t of n.templates ?? [])
    if (t.error) out.push(`${kindLabel(t.kind ?? "")} (${t.template}): ${t.error}`);
  return out;
}

export type ProblemGroup = { problem: string; nodes: string[] };

// the same problem on many nodes is one thing to fix, so nodes are gathered under each, the most widespread first
export function groupProblems(nodes: NodeCheck[]): ProblemGroup[] {
  const groups = new Map<string, Set<string>>();
  for (const n of nodes)
    for (const p of problems(n)) {
      if (!groups.has(p)) groups.set(p, new Set());
      groups.get(p)!.add(n.name ?? "");
    }
  return [...groups.entries()]
    .map(([problem, set]) => ({ problem, nodes: [...set] }))
    .sort((a, b) => b.nodes.length - a.nodes.length);
}

export function formatBytes(bytes: number) {
  const units = ["B", "KiB", "MiB", "GiB"];
  let value = bytes;
  let unit = 0;
  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024;
    unit++;
  }
  return `${unit === 0 || value >= 100 ? Math.round(value) : value.toFixed(1)} ${units[unit]}`;
}

// an Ignition config is sent on one line, which is indented here so it can be read
export function readable(kind: string, content: string) {
  if (kind !== "ignition") return content;
  try {
    return JSON.stringify(JSON.parse(content), null, 2);
  } catch {
    return content;
  }
}

const kindLanguages: Record<string, Language> = {
  ignition: "json",
  butane: "yaml",
  user_data: "yaml",
  meta_data: "yaml",
  kickstart: "shellscript",
  proxmox: "toml",
};

const extensionLanguages: Record<string, Language> = {
  json: "json",
  ign: "json",
  yaml: "yaml",
  yml: "yaml",
  bu: "yaml",
  toml: "toml",
  sh: "shellscript",
  bash: "shellscript",
  ks: "shellscript",
};

// languageOf is what to color rendered output as: what its kind always is, or else the extension the template's name carries under .tmpl, as in idrac-config.json.tmpl. Anything else, an iPXE script among them, is left plain
export function languageOf(kind: string, template?: string): Language | undefined {
  if (kindLanguages[kind]) return kindLanguages[kind];
  if (kind === "ipxe") return undefined;
  const ext = /\.([a-z0-9]+)$/i.exec((template ?? kind).replace(/\.tmpl$/i, ""))?.[1];
  return ext ? extensionLanguages[ext.toLowerCase()] : undefined;
}

// translationFailure recognises the error for a Butane config that rendered but would not translate to Ignition, and finds the line of the rendered Butane it names, if it names one
export function translationFailure(error: string): { line?: number } | undefined {
  if (!error.startsWith("failed to translate butane config")) return undefined;
  const line = /line (\d+)/.exec(error)?.[1];
  return { line: line ? Number(line) : undefined };
}
