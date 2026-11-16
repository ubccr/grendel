import { describe, expect, it } from "vitest";
import { ringShares } from "../run-ring";

const sum = (shares: number[]) => shares.reduce((a, b) => a + b, 0);

describe("ringShares", () => {
  it("gives nothing to anyone when there are no nodes", () => {
    expect(ringShares([0, 0, 0])).toEqual([0, 0, 0]);
  });

  it("splits the ring by count", () => {
    expect(ringShares([3, 1])).toEqual([0.75, 0.25]);
    expect(ringShares([4, 0])).toEqual([1, 0]);
  });

  it("keeps a small count visible, taking its share from the larger ones", () => {
    const shares = ringShares([300, 1, 99, 0]);
    expect(shares[1]).toBeCloseTo(0.12);
    expect(shares[3]).toBe(0);
    expect(shares[0]).toBeGreaterThan(shares[2]);
    expect(sum(shares)).toBeCloseTo(1);
  });
});
