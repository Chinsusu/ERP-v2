import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  getStockAdjustments,
  resetPrototypeStockAdjustmentsForTest,
  summarizeStockAdjustmentDelta,
  transitionStockAdjustment
} from "./stockAdjustmentService";

describe("stockAdjustmentService", () => {
  beforeEach(() => {
    resetPrototypeStockAdjustmentsForTest();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("falls back to submitted prototype adjustments when the API is not reachable", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("offline")));

    const rows = await getStockAdjustments();

    expect(rows[0]).toMatchObject({
      adjustmentNo: "ADJ-260426-0001",
      status: "submitted",
      requestedBy: "user-warehouse"
    });
    expect(summarizeStockAdjustmentDelta(rows[0])).toEqual({
      deltaQty: "-2.000000",
      baseUomCode: "PCS"
    });
  });

  it("approves and posts prototype adjustments when the API is not reachable", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("offline")));

    const [pending] = await getStockAdjustments();
    const approved = await transitionStockAdjustment(pending.id, "approve");
    const posted = await transitionStockAdjustment(pending.id, "post");

    expect(approved).toMatchObject({
      status: "approved",
      approvedBy: "local-dev"
    });
    expect(posted).toMatchObject({
      status: "posted",
      postedBy: "local-dev"
    });
    expect(posted.auditLogId).toContain("audit-stock-adjustment-post");
  });
});
