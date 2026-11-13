import type { ReactNode } from "react";

// token matches one JSON value worth coloring: a string, then a colon if that string is a key, or a number, or a literal
const token =
  /("(?:\\.|[^"\\])*")(\s*:)?|-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?|\b(?:true|false|null)\b/g;

const colors = {
  key: "text-sky-800 dark:text-sky-300",
  string: "text-emerald-700 dark:text-emerald-400",
  number: "text-amber-700 dark:text-amber-300",
  literal: "text-violet-700 dark:text-violet-300",
};

function colorOf(m: RegExpExecArray): string {
  if (m[1]) return m[2] ? colors.key : colors.string;
  return /^[-\d]/.test(m[0]) ? colors.number : colors.literal;
}

// highlightJson colors JSON one line at a time, so it suits JSON.stringify output, which never breaks a value across lines
export function highlightJson(code: string): ReactNode[] {
  return code.split("\n").map((line) => {
    const parts: ReactNode[] = [];
    let at = 0;
    for (const m of line.matchAll(token)) {
      if (m.index > at) parts.push(line.slice(at, m.index));
      const text = m[1] ?? m[0];
      parts.push(
        <span key={m.index} className={colorOf(m)}>
          {text}
        </span>,
      );
      if (m[2]) parts.push(m[2]);
      at = m.index + m[0].length;
    }
    if (at < line.length) parts.push(line.slice(at));
    return parts;
  });
}
