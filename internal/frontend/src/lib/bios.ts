import { BmcResponseModelBmcBiosRegistry } from "@/client";
import { parseNodeSet } from "@/lib/nodeset";

export type BiosValue = string | number | boolean | null;
export type BiosValues = Record<string, BiosValue>;

type ArrayElement<T> = T extends (infer U)[] ? U : never;
type Registry = NonNullable<
  ArrayElement<NonNullable<BmcResponseModelBmcBiosRegistry["results"]>>["data"]
>;
export type BiosAttribute = ArrayElement<NonNullable<Registry["attributes"]>>;

export type InputKind = "enum" | "integer" | "string" | "boolean" | "password";

// asValues narrows the generated client's unknown values to what a bmc reports
export function asValues(raw?: { [key: string]: unknown }): BiosValues {
  const out: BiosValues = {};
  for (const [name, value] of Object.entries(raw ?? {})) {
    out[name] =
      typeof value === "string" || typeof value === "number" || typeof value === "boolean"
        ? value
        : null;
  }
  return out;
}

// inputKind picks the editor for an attribute, from the registry when there is one and otherwise from the type of its current value
export function inputKind(attr: BiosAttribute | undefined, current: BiosValue): InputKind {
  switch (attr?.type) {
    case "Enumeration":
      return "enum";
    case "Integer":
      return "integer";
    case "Boolean":
      return "boolean";
    case "Password":
      return "password";
    case "String":
      return "string";
  }
  if (typeof current === "number") return "integer";
  if (typeof current === "boolean") return "boolean";
  return "string";
}

// editable is whether the registry lets an attribute be changed
export function editable(attr: BiosAttribute | undefined) {
  return !(attr?.read_only || attr?.immutable || attr?.gray_out || attr?.type === "Password");
}

// menuSection names the BIOS setup menu an attribute sits in, from its registry MenuPath such as ./ProcSettings/Turbo, falling back to Other
export function menuSection(menuPath?: string) {
  const top = (menuPath ?? "").replace(/^\.\//, "").split("/")[0];
  if (!top) return "Other";
  return top
    .replace(/Ref$/, "")
    .replace(/([a-z0-9])([A-Z])/g, "$1 $2")
    .replace(/([A-Z]+)([A-Z][a-z])/g, "$1 $2");
}

// sameValue compares bios values the way the bmc would, so 4 and 4.0 match
export function sameValue(a: BiosValue | undefined, b: BiosValue | undefined) {
  return a === b || (a != null && b != null && String(a) === String(b));
}

export function showValue(value: BiosValue | undefined) {
  if (value === undefined) return "(absent)";
  if (value === null) return "-";
  return String(value);
}

export type DiffGroup = { value: string; nodes: string[]; nodeset: string };
export type DiffRow = { name: string; groups: DiffGroup[] };

// biosDiff finds the attributes that differ between nodes, with the nodes holding each value folded into a nodeset. It matches `grendel bmc bios diff`: a staged value counts as the node's, a missing attribute counts as its own value, and the most common value comes first
export function biosDiff(nodes: Record<string, BiosValues>, names?: string[]): DiffRow[] {
  const all = new Set<string>(names?.length ? names : []);
  if (!names?.length) {
    for (const values of Object.values(nodes))
      for (const name of Object.keys(values)) all.add(name);
  }

  const rows: DiffRow[] = [];
  for (const name of [...all].sort()) {
    const byValue = new Map<string, string[]>();
    for (const [node, values] of Object.entries(nodes)) {
      const shown = name in values ? showValue(values[name]) : "(absent)";
      byValue.set(shown, [...(byValue.get(shown) ?? []), node]);
    }
    if (byValue.size < 2) continue;

    const groups = [...byValue.entries()]
      .map(([value, holders]) => {
        const sorted = [...holders].sort();
        return { value, nodes: sorted, nodeset: parseNodeSet(sorted.join(",")).toString() };
      })
      .sort((a, b) => b.nodes.length - a.nodes.length || a.value.localeCompare(b.value));
    rows.push({ name, groups });
  }
  return rows;
}

// parseInput turns what was typed into the value sent for an attribute. The server converts strings too, but sending the right type keeps the change list accurate
export function parseInput(kind: InputKind, raw: string): BiosValue {
  if (kind === "integer") {
    const n = Number(raw);
    return raw.trim() !== "" && Number.isFinite(n) ? n : raw;
  }
  if (kind === "boolean") return raw === "true";
  return raw;
}
