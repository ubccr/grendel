import { useQuery } from "@tanstack/react-query";
import { useMemo, type CSSProperties, type ReactNode } from "react";
import { createBundledHighlighter, createSingletonShorthands } from "shiki/core";
import { createJavaScriptRegexEngine } from "shiki/engine/javascript";

const langs = {
  json: () => import("shiki/langs/json.mjs"),
  yaml: () => import("shiki/langs/yaml.mjs"),
  shellscript: () => import("shiki/langs/shellscript.mjs"),
  toml: () => import("shiki/langs/toml.mjs"),
};

export type Language = keyof typeof langs;

// only what is listed here is bundled into the Grendel binary, and each loads the first time it is used
const { codeToTokens } = createSingletonShorthands(
  createBundledHighlighter({
    langs,
    themes: {
      "github-light": () => import("shiki/themes/github-light.mjs"),
      "github-dark": () => import("shiki/themes/github-dark.mjs"),
    },
    engine: () => createJavaScriptRegexEngine({ forgiving: true }),
  }),
);

// highlight colors code for both themes, as one node per line
async function highlight(code: string, lang: Language): Promise<ReactNode[]> {
  const { tokens } = await codeToTokens(code, {
    lang,
    themes: { light: "github-light", dark: "github-dark" },
    defaultColor: false,
  });
  return tokens.map((line) =>
    line.map((token, i) => (
      <span
        key={i}
        style={token.htmlStyle as CSSProperties}
        className="text-(--shiki-light) dark:text-(--shiki-dark)"
      >
        {token.content}
      </span>
    )),
  );
}

// MAX_HIGHLIGHT is the most code that gets colored, so a very large file shows as plain text rather than stalling the page
const MAX_HIGHLIGHT = 512 * 1024;

// useHighlighted is code as one node per line: plain at first, and for code with no language or too much of it, then colored once that is ready. key names what the code was made from, so the code itself, which can be long, is not hashed on every render
export function useHighlighted(
  code: string,
  lang: Language | undefined,
  key: readonly unknown[],
): ReactNode[] {
  const text = useMemo(() => code.replace(/\n$/, ""), [code]);
  const highlighted = useQuery({
    queryKey: ["highlighted", lang, ...key],
    queryFn: () => highlight(text, lang!),
    enabled: !!lang && !!text && text.length <= MAX_HIGHLIGHT,
    staleTime: Infinity,
    gcTime: 0,
  });
  return useMemo(() => highlighted.data ?? text.split("\n"), [highlighted.data, text]);
}
