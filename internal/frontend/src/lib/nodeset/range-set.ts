import { InvalidRangeSetError, ParseRangeSetError } from "./errors";
import { padInt } from "./format";

export interface Slice {
  start: number;
  stop: number;
  step: number;
  pad: number;
}

export class RangeSetItem {
  constructor(
    readonly value: number,
    readonly padding: number,
  ) {}

  toString(): string {
    return padInt(this.value, this.padding);
  }
}

export class RangeSet {
  private readonly bits = new Set<number>();
  padding = 0;

  constructor(pattern = "") {
    if (pattern.length === 0) return;
    for (const subrange of pattern.split(",")) {
      this.addString(subrange);
    }
  }

  addString(subrange: string): void {
    if (subrange === "") {
      throw new ParseRangeSetError("empty range");
    }

    let baserange = subrange;
    let step = 1;
    if (subrange.indexOf("/") >= 0) {
      const parts = splitN(subrange, "/", 2);
      baserange = parts[0];
      if (parts.length !== 2 || parts[1] === "") {
        throw new ParseRangeSetError(`cannot parse step ${subrange}`);
      }
      const parsedStep = parseStrictInt(parts[1]);
      if (parsedStep === null) {
        throw new ParseRangeSetError(`cannot convert step to integer ${subrange}`);
      }
      step = parsedStep;
    }

    let parts: string[] = [baserange];
    if (baserange.indexOf("-") < 0) {
      if (step !== 1) {
        throw new ParseRangeSetError(`invalid step usage ${subrange}`);
      }
    } else {
      parts = splitN(baserange, "-", 2);
      if (parts.length !== 2 || parts[1] === "") {
        throw new ParseRangeSetError(`cannot parse end value ${subrange}`);
      }
    }

    const start = parseStrictInt(parts[0]);
    if (start === null) {
      throw new ParseRangeSetError(`cannot convert starting range to integer ${parts[0]}`);
    }

    let pad = 0;
    if (start !== 0) {
      const begins = trimLeft(parts[0], "0");
      if (parts[0].length - begins.length > 0) {
        pad = parts[0].length;
      }
    } else {
      if (parts[0].length > 1) {
        pad = parts[0].length;
      }
    }

    let stop: number;
    if (parts.length === 2) {
      const parsedStop = parseStrictInt(parts[1]);
      if (parsedStop === null) {
        throw new ParseRangeSetError(`cannot convert ending range to integer ${parts[1]}`);
      }
      stop = parsedStop;
    } else {
      stop = start;
    }

    if (start > stop || step < 1) {
      throw new ParseRangeSetError(`invalid value in range ${subrange}`);
    }

    this.addSlice({ start, stop: stop + 1, step, pad });
  }

  addSlice(slice: Slice): void {
    if (slice.start > slice.stop) {
      throw new InvalidRangeSetError("invalid range start > stop");
    }
    if (slice.step <= 0) {
      throw new InvalidRangeSetError("invalid range step <= 0");
    }
    if (slice.pad < 0) {
      throw new InvalidRangeSetError("invalid range padding < 0");
    }

    if (slice.pad > 0 && this.padding === 0) {
      this.padding = slice.pad;
    }

    for (let i = slice.start; i < slice.stop; i += slice.step) {
      this.bits.add(i);
    }
  }

  clone(): RangeSet {
    const out = new RangeSet();
    out.padding = this.padding;
    for (const v of this.bits) out.bits.add(v);
    return out;
  }

  intersection(other: RangeSet): RangeSet {
    const out = new RangeSet();
    out.padding = this.padding;
    for (const v of this.bits) {
      if (other.bits.has(v)) out.bits.add(v);
    }
    return out;
  }

  inPlaceIntersection(other: RangeSet): void {
    for (const v of this.bits) {
      if (!other.bits.has(v)) this.bits.delete(v);
    }
    if (this.padding < other.padding) this.padding = other.padding;
  }

  union(other: RangeSet): RangeSet {
    const out = this.clone();
    for (const v of other.bits) out.bits.add(v);
    return out;
  }

  inPlaceUnion(other: RangeSet): void {
    for (const v of other.bits) this.bits.add(v);
    if (this.padding < other.padding) this.padding = other.padding;
  }

  difference(other: RangeSet): RangeSet {
    const out = new RangeSet();
    out.padding = this.padding;
    for (const v of this.bits) {
      if (!other.bits.has(v)) out.bits.add(v);
    }
    return out;
  }

  inPlaceDifference(other: RangeSet): void {
    for (const v of other.bits) this.bits.delete(v);
    if (this.padding < other.padding) this.padding = other.padding;
  }

  symmetricDifference(other: RangeSet): RangeSet {
    const out = new RangeSet();
    out.padding = this.padding;
    for (const v of this.bits) {
      if (!other.bits.has(v)) out.bits.add(v);
    }
    for (const v of other.bits) {
      if (!this.bits.has(v)) out.bits.add(v);
    }
    return out;
  }

  inPlaceSymmetricDifference(other: RangeSet): void {
    const toAdd: number[] = [];
    for (const v of other.bits) {
      if (this.bits.has(v)) {
        this.bits.delete(v);
      } else {
        toAdd.push(v);
      }
    }
    for (const v of toAdd) this.bits.add(v);
    if (this.padding < other.padding) this.padding = other.padding;
  }

  superset(other: RangeSet): boolean {
    for (const v of other.bits) {
      if (!this.bits.has(v)) return false;
    }
    return true;
  }

  subset(other: RangeSet): boolean {
    return other.superset(this);
  }

  greater(other: RangeSet): boolean {
    return this.len() > other.len() && this.superset(other);
  }

  less(other: RangeSet): boolean {
    return this.len() < other.len() && this.subset(other);
  }

  equal(other: RangeSet): boolean {
    if (this.bits.size !== other.bits.size) return false;
    for (const v of this.bits) {
      if (!other.bits.has(v)) return false;
    }
    return true;
  }

  empty(): boolean {
    return this.bits.size === 0;
  }

  len(): number {
    return this.bits.size;
  }

  sortedValues(): number[] {
    return [...this.bits].sort((a, b) => a - b);
  }

  slices(): Slice[] {
    const result: Slice[] = [];
    const values = this.sortedValues();
    if (values.length === 0) return result;

    let k = values[0];
    let j = values[0];
    for (let idx = 1; idx < values.length; idx++) {
      const i = values[idx];
      if (i - j > 1) {
        result.push({ start: k, stop: j + 1, step: 1, pad: this.padding });
        k = i;
      }
      j = i;
    }
    result.push({ start: k, stop: j + 1, step: 1, pad: this.padding });
    return result;
  }

  toString(): string {
    const parts: string[] = [];
    for (const sli of this.slices()) {
      if (sli.start + 1 === sli.stop) {
        parts.push(padInt(sli.start, this.padding));
      } else {
        parts.push(`${padInt(sli.start, this.padding)}-${padInt(sli.stop - 1, this.padding)}`);
      }
    }
    return parts.join(",");
  }

  strings(): string[] {
    const out: string[] = [];
    for (const sli of this.slices()) {
      for (let i = sli.start; i < sli.stop; i += sli.step) {
        out.push(padInt(i, this.padding));
      }
    }
    return out;
  }

  ints(): number[] {
    const out: number[] = [];
    for (const sli of this.slices()) {
      for (let i = sli.start; i < sli.stop; i += sli.step) {
        out.push(i);
      }
    }
    return out;
  }

  items(): RangeSetItem[] {
    const out: RangeSetItem[] = [];
    for (const sli of this.slices()) {
      for (let i = sli.start; i < sli.stop; i += sli.step) {
        out.push(new RangeSetItem(i, this.padding));
      }
    }
    return out;
  }
}

function splitN(s: string, sep: string, n: number): string[] {
  if (n <= 0) return s.split(sep);
  const parts: string[] = [];
  let rest = s;
  while (parts.length < n - 1) {
    const idx = rest.indexOf(sep);
    if (idx < 0) break;
    parts.push(rest.slice(0, idx));
    rest = rest.slice(idx + sep.length);
  }
  parts.push(rest);
  return parts;
}

function trimLeft(s: string, cutset: string): string {
  let i = 0;
  while (i < s.length && cutset.indexOf(s[i]) >= 0) i++;
  return s.slice(i);
}

function parseStrictInt(s: string): number | null {
  if (s === "" || !/^-?\d+$/.test(s)) return null;
  const n = Number(s);
  if (!Number.isSafeInteger(n)) return null;
  return n;
}
