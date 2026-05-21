import { RangeSetItem } from "./range-set";

export class NodeSetIterator implements Iterable<string> {
  private current = -1;

  constructor(private readonly nodes: readonly string[]) {}

  next(): boolean {
    this.current++;
    return this.current < this.nodes.length;
  }

  len(): number {
    return this.nodes.length;
  }

  value(): string {
    return this.nodes[this.current];
  }

  stringSlice(): string[] {
    const out: string[] = [];
    while (this.next()) {
      out.push(this.value());
    }
    return out;
  }

  [Symbol.iterator](): Iterator<string> {
    let i = 0;
    const nodes = this.nodes;
    return {
      next(): IteratorResult<string> {
        if (i < nodes.length) {
          return { value: nodes[i++], done: false };
        }
        return { value: undefined, done: true };
      },
    };
  }
}

export class RangeSetNDIterator implements Iterable<RangeSetItem[]> {
  private readonly vects: RangeSetItem[][] = [];
  private readonly seen = new Set<string>();
  private current = -1;

  next(): boolean {
    this.current++;
    return this.current < this.vects.length;
  }

  len(): number {
    return this.vects.length;
  }

  intValue(): number[] {
    return this.vects[this.current].map((v) => v.value);
  }

  formatList(): string[] {
    return this.vects[this.current].map((v) => v.toString());
  }

  values(): readonly RangeSetItem[][] {
    return this.vects;
  }

  sort(): void {
    this.vects.sort((a, b) => {
      for (let x = 0; x < a.length; x++) {
        if (a[x].value !== b[x].value) {
          return a[x].value - b[x].value;
        }
      }
      return 0;
    });
  }

  product(result: RangeSetItem[], ...params: RangeSetItem[][]): void {
    if (params.length === 0) {
      const key = result.map((i) => i.value).join(",");
      if (!this.seen.has(key)) {
        this.seen.add(key);
        this.vects.push(result);
      }
      return;
    }

    const [p, ...rest] = params;
    for (let i = 0; i < p.length; i++) {
      this.product([...result, p[i]], ...rest);
    }
  }

  [Symbol.iterator](): Iterator<RangeSetItem[]> {
    let i = 0;
    const vects = this.vects;
    return {
      next(): IteratorResult<RangeSetItem[]> {
        if (i < vects.length) {
          return { value: vects[i++], done: false };
        }
        return { value: undefined, done: true };
      },
    };
  }
}
