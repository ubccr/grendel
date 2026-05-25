export function padInt(value: number, width: number): string {
  const s = String(value);
  if (width <= s.length) return s;
  return "0".repeat(width - s.length) + s;
}

export function formatPattern(template: string, params: readonly string[]): string {
  let result = "";
  let paramIdx = 0;
  let i = 0;
  while (i < template.length) {
    if (template[i] === "%" && template[i + 1] === "s") {
      if (paramIdx >= params.length) {
        throw new Error(`formatPattern: not enough params for template "${template}"`);
      }
      result += params[paramIdx++];
      i += 2;
    } else {
      result += template[i];
      i++;
    }
  }
  if (paramIdx !== params.length) {
    throw new Error(
      `formatPattern: ${params.length - paramIdx} unused params for template "${template}"`,
    );
  }
  return result;
}
