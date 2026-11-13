import { describe, expect, it } from "vitest";
import { renderToStaticMarkup } from "react-dom/server";
import { highlightJson } from "@/lib/highlight-json";

const html = (code: string) => highlightJson(code).map((l) => renderToStaticMarkup(<>{l}</>));

describe("highlightJson", () => {
  it("gives one line per line and keeps the text as it was", () => {
    const code = JSON.stringify({ a: 1, b: [true, null], c: "x" }, null, 2);
    const lines = html(code);
    expect(lines).toHaveLength(code.split("\n").length);
    expect(lines.map((l) => l.replace(/<[^>]+>/g, "").replaceAll("&quot;", '"')).join("\n")).toBe(
      code,
    );
  });
  it("colors keys apart from string values", () => {
    const [line] = html(`"name": "cpn-01:x",`);
    expect(line).toContain('<span class="text-sky-800 dark:text-sky-300">&quot;name&quot;</span>:');
    expect(line).toContain('text-emerald-700 dark:text-emerald-400">&quot;cpn-01:x&quot;</span>,');
  });
  it("handles escaped quotes, numbers and literals", () => {
    const [line] = html(`"a\\"b": -1.5e3, false`);
    expect(line).toContain("&quot;a\\&quot;b&quot;</span>:");
    expect(line).toContain('text-amber-700 dark:text-amber-300">-1.5e3</span>');
    expect(line).toContain('text-violet-700 dark:text-violet-300">false</span>');
  });
});
