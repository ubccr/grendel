// A name template is text with {fields}, like cpn-{sw.1}-{port:02}. A numeric field takes one arithmetic step and a zero-padded width: {port+2}, {port/2:02}. A text field can swap its first match of one string for another: {name/cpn/bmc}

export type TemplateVars = {
  // the faceplate number of the switch port
  port?: number;
  // the running count over the ports being imported
  n?: number;
  // the switch's name; {sw.1} is the first run of digits in it
  sw?: string;
  // the node's name, for templates that build on it, like an FQDN
  name?: string;
};

export type TemplateError = { message: string; at: number };

type Part =
  | { text: string }
  | {
      field: "port" | "n" | "sw" | "name" | "swn";
      group?: number;
      op?: Op;
      pad?: number;
      swap?: { from: string; to: string };
      at: number;
    };

type Op = { kind: "+" | "-" | "*" | "/"; by: number };

export type Template = {
  parts: Part[];
  error?: TemplateError;
  uses: Set<string>;
};

const fieldRe = /^(port|n|sw|name)(?:\.(\d+))?(?:([+\-*/])(\d+))?(?::(\d+))?$/;
const swapRe = /^(sw|name)\/([^/]+)\/([^/]*)$/;

export function parseTemplate(source: string): Template {
  const parts: Part[] = [];
  const uses = new Set<string>();
  let i = 0;
  while (i < source.length) {
    const open = source.indexOf("{", i);
    const close = source.indexOf("}", i);
    if (close !== -1 && (open === -1 || close < open))
      return { parts, uses, error: { message: "A } has no { before it.", at: close } };
    if (open === -1) {
      parts.push({ text: source.slice(i) });
      break;
    }
    if (open > i) parts.push({ text: source.slice(i, open) });
    const end = source.indexOf("}", open);
    if (end === -1) return { parts, uses, error: { message: "A { is never closed.", at: open } };

    const body = source.slice(open + 1, end).replace(/\s+/g, "");
    const swap = swapRe.exec(body);
    if (swap) {
      uses.add(swap[1]);
      parts.push({
        field: swap[1] as "sw" | "name",
        swap: { from: swap[2], to: swap[3] },
        at: open,
      });
      i = end + 1;
      continue;
    }
    const m = fieldRe.exec(body);
    if (!m)
      return {
        parts,
        uses,
        error: {
          message: `{${body}} isn't a field. Use {port}, {n}, {sw}, {sw.1}, {name} or {name/from/to}.`,
          at: open,
        },
      };
    const [, name, group, opKind, by, pad] = m;
    if (group && name !== "sw")
      return {
        parts,
        uses,
        error: { message: `Only {sw} takes a .number, as in {sw.1}.`, at: open },
      };
    if (opKind && name !== "port" && name !== "n" && !group)
      return {
        parts,
        uses,
        error: { message: `{${name}} is text, so it can't take arithmetic.`, at: open },
      };
    if (opKind === "/" && Number(by) === 0)
      return { parts, uses, error: { message: "Can't divide by 0.", at: open } };

    const field = group ? "swn" : (name as "port" | "n" | "sw" | "name");
    uses.add(group ? `sw.${group}` : name);
    parts.push({
      field,
      group: group ? Number(group) : undefined,
      op: opKind ? { kind: opKind as Op["kind"], by: Number(by) } : undefined,
      pad: pad ? Number(pad) : undefined,
      at: open,
    });
    i = end + 1;
  }
  return { parts, uses };
}

function apply(v: number, op?: Op) {
  if (!op) return v;
  switch (op.kind) {
    case "+":
      return v + op.by;
    case "-":
      return v - op.by;
    case "*":
      return v * op.by;
    // rounds up, so with two ports to a node, ports 1 and 2 are both 1
    case "/":
      return Math.ceil(v / op.by);
  }
}

// switchDigits is the runs of digits in a switch's short name, so swe-a01-r12 gives ["01", "12"]
export function switchDigits(sw: string) {
  return sw.split(".")[0].match(/\d+/g) ?? [];
}

// renderTemplate fills in a parsed template, or explains which field it couldn't fill
export function renderTemplate(t: Template, vars: TemplateVars): { value: string; error?: string } {
  if (t.error) return { value: "", error: t.error.message };
  let out = "";
  for (const p of t.parts) {
    if ("text" in p) {
      out += p.text;
      continue;
    }
    let text: string;
    switch (p.field) {
      case "sw":
        if (vars.sw === undefined) return { value: out, error: "{sw} has no switch here." };
        text = vars.sw.split(".")[0];
        break;
      case "name":
        if (!vars.name) return { value: out, error: "{name} is empty until the node has a name." };
        text = vars.name;
        break;
      case "swn": {
        const digits = vars.sw === undefined ? [] : switchDigits(vars.sw);
        const d = digits[p.group! - 1];
        if (d === undefined)
          return {
            value: out,
            error: `${vars.sw ?? "The switch"} has no number ${p.group} in its name.`,
          };
        // a bare {sw.1} keeps the digits as written, so a01 stays 01; arithmetic works on the number
        text = p.op ? String(apply(Number(d), p.op)) : d;
        break;
      }
      default: {
        const v = vars[p.field];
        if (v === undefined) return { value: out, error: `{${p.field}} has no value here.` };
        text = String(apply(v, p.op));
      }
    }
    // only the first match is swapped, and text without one is left as it is, so a node renamed by hand keeps its own name
    if (p.swap) text = text.replace(p.swap.from, p.swap.to);
    if (p.pad && /^-?\d+$/.test(text)) {
      const neg = text.startsWith("-");
      text = (neg ? "-" : "") + text.replace("-", "").padStart(p.pad, "0");
    }
    out += text;
  }
  return { value: out };
}

const hostnameRe = /^[A-Za-z0-9](?:[A-Za-z0-9._-]*[A-Za-z0-9])?$/;

// nameProblem says why a rendered node name can't be used, since it also becomes a DNS name and a nodeset member
export function nameProblem(name: string) {
  if (!name) return "No name.";
  if (!hostnameRe.test(name))
    return "Use only letters, digits, dots, dashes and underscores, starting and ending with a letter or digit.";
  return undefined;
}
