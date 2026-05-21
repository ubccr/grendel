import { ParseNodeSetError } from "./errors";
import { formatPattern } from "./format";
import { NodeSetIterator } from "./iterator";
import { RangeSetND } from "./range-set-nd";

const rangeSetRegexp = /(\[[^[\]]+\]|[0-9]+)/g;

interface PatternEntry {
  format: string;
  rangeSet: RangeSetND | null;
}

export class NodeSet implements Iterable<string> {
  private readonly patterns = new Map<string, RangeSetND | null>();

  constructor(nodestr = "") {
    if (nodestr === "") return;
    this.add(nodestr);
  }

  add(nodestr: string): void {
    const cleaned = nodestr.replace(/ /g, "");
    if (cleaned === "") {
      throw new ParseNodeSetError("empty nodeset");
    }

    const ranges: string[] = [];
    for (const m of cleaned.matchAll(rangeSetRegexp)) {
      ranges.push(m[1]);
    }
    const patterns = cleaned.replace(rangeSetRegexp, "%s");

    if (patterns.indexOf("[") !== -1) {
      throw new ParseNodeSetError(`unbalanced '[' found while parsing ${cleaned}`);
    }
    if (patterns.indexOf("]") !== -1) {
      throw new ParseNodeSetError(`unbalanced ']' found while parsing ${cleaned}`);
    }

    let ridx = 0;
    for (const pattern of patterns.split(",")) {
      const rangeSetCount = countOccurrences(pattern, "%s");
      if (rangeSetCount === 0) {
        this.patterns.set(pattern, null);
        continue;
      }

      const rangeSets: string[] = [];
      for (let i = ridx; i < ridx + rangeSetCount; i++) {
        rangeSets.push(trimBrackets(ranges[i]));
      }

      const rs = new RangeSetND([rangeSets]);
      const existing = this.patterns.get(pattern);
      if (existing == null) {
        this.patterns.set(pattern, rs);
      } else {
        existing.update(rs);
      }

      ridx += rangeSetCount;
    }
  }

  len(): number {
    let size = 0;
    for (const rs of this.patterns.values()) {
      if (rs === null) size++;
      else size += rs.len();
    }
    return size;
  }

  toString(): string {
    return this.toStringList().join(",");
  }

  toStringList(): string[] {
    const items: PatternEntry[] = [];
    for (const [format, rangeSet] of this.patterns) {
      items.push({ format, rangeSet });
    }

    sortStable(items, (a, b) => {
      const alen = a.rangeSet === null ? 1 : a.rangeSet.len();
      const blen = b.rangeSet === null ? 1 : b.rangeSet.len();
      return blen - alen;
    });

    const list: string[] = [];
    for (const entry of items) {
      if (entry.rangeSet === null) {
        list.push(entry.format);
        continue;
      }
      for (const params of entry.rangeSet.formatList()) {
        list.push(formatPattern(entry.format, params));
      }
    }
    return list;
  }

  iterator(): NodeSetIterator {
    const items: PatternEntry[] = [];
    for (const [format, rangeSet] of this.patterns) {
      items.push({ format, rangeSet });
    }

    sortStable(items, (a, b) => {
      const alen = a.rangeSet === null ? 1 : a.rangeSet.len();
      const blen = b.rangeSet === null ? 1 : b.rangeSet.len();
      return blen - alen;
    });

    const nodes: string[] = [];
    for (const entry of items) {
      if (entry.rangeSet === null) {
        nodes.push(entry.format);
        continue;
      }
      const it = entry.rangeSet.iterator();
      while (it.next()) {
        nodes.push(formatPattern(entry.format, it.formatList()));
      }
    }

    return new NodeSetIterator(nodes);
  }

  toJSON(): string[] {
    return this.toStringList();
  }

  static fromJSON(value: unknown): NodeSet {
    if (!Array.isArray(value) || !value.every((v) => typeof v === "string")) {
      throw new ParseNodeSetError("expected a JSON array of strings");
    }
    return new NodeSet(value.join(","));
  }

  [Symbol.iterator](): Iterator<string> {
    return this.iterator()[Symbol.iterator]();
  }
}

function trimBrackets(s: string): string {
  let start = 0;
  let end = s.length;
  while (start < end && (s[start] === "[" || s[start] === "]")) start++;
  while (end > start && (s[end - 1] === "[" || s[end - 1] === "]")) end--;
  return s.slice(start, end);
}

function countOccurrences(haystack: string, needle: string): number {
  if (needle === "") return 0;
  let count = 0;
  let pos = 0;
  while ((pos = haystack.indexOf(needle, pos)) !== -1) {
    count++;
    pos += needle.length;
  }
  return count;
}

function sortStable<T>(arr: T[], cmp: (a: T, b: T) => number): void {
  const indexed = arr.map((value, index) => ({ value, index }));
  indexed.sort((a, b) => {
    const c = cmp(a.value, b.value);
    return c !== 0 ? c : a.index - b.index;
  });
  for (let i = 0; i < arr.length; i++) arr[i] = indexed[i].value;
}
