import { MismatchedDimensionsError } from "./errors";
import { RangeSetNDIterator } from "./iterator";
import { RangeSet } from "./range-set";

export class RangeSetND {
  private readonly rangesData: RangeSet[][];
  private dirty: boolean;

  constructor(args: readonly (readonly string[])[]) {
    this.dirty = true;
    this.rangesData = args.map((rgvec) => rgvec.map((rg) => new RangeSet(rg)));
  }

  update(other: RangeSetND): void {
    if (this.dim() !== other.dim()) {
      throw new MismatchedDimensionsError(
        `mismatched dimensions ${this.dim()} != ${other.dim()}`,
      );
    }
    this.rangesData.push(...other.rangesData);
    this.dirty = true;
  }

  dim(): number {
    if (this.rangesData.length === 0) return 0;
    return this.rangesData[0].length;
  }

  len(): number {
    return this.iterator().len();
  }

  fold(): void {
    if (!this.dirty) return;
    if (this.rangesData.length === 0) {
      this.dirty = false;
      return;
    }

    const dim = this.rangesData[0].length;
    let vardim = 0;
    let dimdiff = 0;
    if (dim > 1) {
      for (let i = 0; i < dim; i++) {
        const slist = new Set<string>();
        for (const rs of this.rangesData) {
          slist.add(rs[i].toString());
        }
        if (slist.size !== 1) {
          dimdiff += 1;
          if (dimdiff > 1) break;
          vardim = i;
        }
      }
    }

    if (dim === 1 || dimdiff === 1) {
      for (let i = 1; i < this.rangesData.length; i++) {
        this.rangesData[0][vardim].inPlaceUnion(this.rangesData[i][vardim]);
      }
      this.rangesData.length = 1;
    } else {
      this.foldMultivariate();
    }

    this.dirty = false;
  }

  private foldMultivariate(): void {
    this.foldMultivariateExpand();
    this.sort();
    this.foldMultivariateMerge();
    this.sort();
  }

  private foldMultivariateExpand(): void {
    let index1 = 0;
    while (index1 + 1 < this.rangesData.length) {
      let item1 = this.rangesData[index1];
      let index2 = index1 + 1;
      index1++;
      while (index2 < this.rangesData.length) {
        const item2 = this.rangesData[index2];
        index2++;
        let newItem: RangeSet[] | null = null;
        let disjoint = false;
        const suppl: RangeSet[][] = [];

        for (let pos = 0; pos < item1.length; pos++) {
          const rg1 = item1[pos];
          const rg2 = item2[pos];
          const rg1Intersect = rg1.intersection(rg2);
          if (rg1Intersect.empty()) {
            disjoint = true;
            break;
          }
          if (newItem === null) {
            newItem = new Array<RangeSet>(item1.length);
          }
          if (rg1.equal(rg2)) {
            newItem[pos] = rg1;
          } else {
            newItem[pos] = rg1Intersect;
            const rg1Diff = rg1.difference(rg2);
            if (!rg1Diff.empty()) {
              suppl.push([...item1.slice(0, pos), rg1Diff, ...item1.slice(pos + 1)]);
            }
            const rg2Diff = rg2.difference(rg1);
            if (!rg2Diff.empty()) {
              suppl.push([...item2.slice(0, pos), rg2Diff, ...item2.slice(pos + 1)]);
            }
          }
        }

        if (!disjoint && newItem !== null) {
          item1 = newItem;
          this.rangesData[index1 - 1] = newItem;
          index2--;
          this.rangesData.splice(index2, 1);
          this.rangesData.push(...suppl);
        }
      }
    }
  }

  private foldMultivariateMerge(): void {
    let chg = true;
    while (chg) {
      chg = false;
      let index1 = 0;
      while (index1 + 1 < this.rangesData.length) {
        let item1 = this.rangesData[index1];
        let index2 = index1 + 1;
        index1++;
        while (index2 < this.rangesData.length) {
          const item2 = this.rangesData[index2];
          index2++;
          const newItem = new Array<RangeSet>(item1.length);
          let nbDiff = 0;

          for (let pos = 0; pos < item1.length; pos++) {
            const rg1 = item1[pos];
            const rg2 = item2[pos];
            const rg1Intersect = rg1.intersection(rg2);

            if (rg1.equal(rg2)) {
              newItem[pos] = rg1.clone();
            } else if (rg1Intersect.empty()) {
              nbDiff++;
              if (nbDiff > 1) break;
              newItem[pos] = rg1.union(rg2);
            } else if (rg1.greater(rg2) || rg1.less(rg2)) {
              nbDiff++;
              if (nbDiff > 1) break;
              newItem[pos] = rg1.greater(rg2) ? rg1.clone() : rg2.clone();
            } else {
              nbDiff = 2;
              break;
            }
          }

          if (nbDiff <= 1) {
            chg = true;
            item1 = newItem;
            this.rangesData[index1 - 1] = newItem;
            index2--;
            this.rangesData.splice(index2, 1);
          }
        }
      }
    }
  }

  dump(): string[] {
    return this.rangesData.map((rgvec) => rgvec.map((rs) => rs.toString()).join(","));
  }

  formatList(): string[][] {
    this.fold();
    return this.rangesData.map((rgvec) =>
      rgvec.map((rs) => (rs.len() > 1 ? `[${rs.toString()}]` : rs.toString())),
    );
  }

  toString(): string {
    this.fold();
    let buffer = "";
    for (const rgvec of this.rangesData) {
      for (let j = 0; j < rgvec.length; j++) {
        buffer += rgvec[j].toString();
        if (j !== rgvec.length - 1) buffer += "; ";
      }
      buffer += "\n";
    }
    return buffer;
  }

  ranges(): readonly RangeSet[][] {
    return this.rangesData;
  }

  sort(): void {
    this.rangesData.sort((a, b) => {
      let isize = a[0].len();
      for (let k = 1; k < a.length; k++) isize *= a[k].len();
      let jsize = b[0].len();
      for (let k = 1; k < b.length; k++) jsize *= b[k].len();

      if (isize === jsize && a.length === b.length) {
        if (a[0].len() === b[0].len()) {
          return b[b.length - 1].len() - a[a.length - 1].len();
        }
        return b[0].len() - a[0].len();
      }
      if (isize === jsize) {
        return b.length - a.length;
      }
      return jsize - isize;
    });
  }

  iterator(): RangeSetNDIterator {
    const it = new RangeSetNDIterator();
    const dim = this.dim();
    for (const rgvec of this.rangesData) {
      const slices = Array.from({ length: dim }, (_, i) => rgvec[i].items());
      it.product([], ...slices);
    }
    return it;
  }
}
