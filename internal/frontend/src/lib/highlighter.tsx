import type { CSSProperties, ReactNode } from "react";
import { createBundledHighlighter, createSingletonShorthands } from "shiki/core";
import { createJavaScriptRegexEngine } from "shiki/engine/javascript";

// only what is listed here is bundled into the Grendel binary, and each loads the first time it is used
const { codeToTokens } = createSingletonShorthands(
  createBundledHighlighter({
    langs: { json: () => import("shiki/langs/json.mjs") },
    themes: {
      "github-light": () => import("shiki/themes/github-light.mjs"),
      "github-dark": () => import("shiki/themes/github-dark.mjs"),
    },
    engine: () => createJavaScriptRegexEngine({ forgiving: true }),
  }),
);

// highlightJson colors JSON for both themes, as one node per line
export async function highlightJson(code: string): Promise<ReactNode[]> {
  const { tokens } = await codeToTokens(code, {
    lang: "json",
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
