import {
  getV1BmcStatusQueryKey,
  getV1GrendelEventsQueryKey,
  getV1NodesQueryKey,
  getV1SwitchScanQueryKey,
} from "@/client/@tanstack/react-query.gen";
import { createQueryClient, operation, readsBmc, readsDatabase } from "@/lib/query-client";
import { newestFirst } from "@/lib/events";
import type { MutationMeta } from "@tanstack/react-query";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { toast } from "sonner";

vi.mock("sonner", () => ({
  toast: { error: vi.fn(), success: vi.fn(), warning: vi.fn() },
}));

const nodes = getV1NodesQueryKey();
const events = getV1GrendelEventsQueryKey();
const status = getV1BmcStatusQueryKey({ query: { nodeset: "cpn-01" } });
const scan = getV1SwitchScanQueryKey({ query: { nodeset: "swe-01" } });
const preview = ["node-import-preview", "{}"];

describe("operation", () => {
  it("is the API operation of a generated key and nothing for a key written by hand", () => {
    expect(operation(nodes)).toBe("getV1Nodes");
    expect(operation(preview)).toBe("");
    expect(operation([])).toBe("");
  });
});

describe("what a read is answered from", () => {
  it("tells the database from BMCs and switches", () => {
    expect([nodes, events, status, scan, preview].map(readsDatabase)).toEqual([
      true,
      true,
      false,
      false,
      false,
    ]);
    expect([nodes, status, scan, preview].map(readsBmc)).toEqual([false, true, false, false]);
  });
});

describe("a settled mutation", () => {
  beforeEach(() => vi.clearAllMocks());

  // each read is put in the cache as fresh, then a mutation is run, and the reads it left out of date are the ones marked invalid
  const invalidatedBy = async (
    meta: MutationMeta | undefined,
    mutationFn: () => Promise<unknown> = async () => ({ summary: { total: 1 } }),
  ) => {
    const client = createQueryClient();
    for (const key of [nodes, events, status, scan, preview]) client.setQueryData(key, []);
    await client
      .getMutationCache()
      .build(client, { mutationFn, meta })
      .execute(undefined)
      .catch(() => undefined);
    return [nodes, events, status, scan, preview].map(
      (key) => client.getQueryState(key)?.isInvalidated,
    );
  };

  it("leaves what was read from the database out of date, and nothing read from hardware", async () => {
    expect(await invalidatedBy(undefined)).toEqual([true, true, false, false, false]);
  });

  it("leaves BMC reads out of date too when it changed BMCs, but never a switch", async () => {
    expect(await invalidatedBy({ bmc: true })).toEqual([true, true, true, false, false]);
  });

  it("does not read BMCs again when the request was refused", async () => {
    const refused = () => Promise.reject({ title: "Bad request", detail: "no such node" });
    expect(await invalidatedBy({ bmc: "Power On" }, refused)).toEqual([
      true,
      true,
      false,
      false,
      false,
    ]);
  });

  it("fetches nothing again after a mutation that changes nothing", async () => {
    expect(await invalidatedBy({ changesNothing: true })).toEqual([
      false,
      false,
      false,
      false,
      false,
    ]);
  });

  it("toasts a failure once, unless the caller shows it", async () => {
    const refused = () => Promise.reject({ title: "Bad request", detail: "no such node" });
    await invalidatedBy(undefined, refused);
    expect(toast.error).toHaveBeenCalledExactlyOnceWith("Bad request", {
      description: "no such node",
    });

    vi.clearAllMocks();
    await invalidatedBy({ quietError: true }, refused);
    expect(toast.error).not.toHaveBeenCalled();
  });

  it("toasts the per-node outcome of a BMC mutation under its name", async () => {
    await invalidatedBy({ bmc: "Power On" });
    expect(toast.success).toHaveBeenCalledExactlyOnceWith("Power On: 1 node(s) succeeded", {
      description: undefined,
    });

    vi.clearAllMocks();
    await invalidatedBy({ bmc: true });
    expect(toast.success).not.toHaveBeenCalled();
  });
});

describe("newestFirst", () => {
  it("orders events by when they ended, or started while still running", () => {
    const list = [
      { id: "audit", started_at: "2026-11-11T10:05:00Z", ended_at: "2026-11-11T10:05:00Z" },
      { id: "provision", started_at: "2026-11-11T10:00:00Z", ended_at: "2026-11-11T10:20:00Z" },
      { id: "running", started_at: "2026-11-11T10:10:00Z" },
    ];
    expect(newestFirst(list).map((e) => e.id)).toEqual(["provision", "running", "audit"]);
  });
});
