import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  availabilityTone,
  createStockCount,
  formatQuantity,
  getAvailableStock,
  getBatchQCTransitions,
  getStockCounts,
  prototypeBatchQCTransitions,
  prototypeAvailableStock,
  resetPrototypeStockCountsForTest,
  submitStockCount,
  summarizeAvailableStock
} from "./stockAvailabilityService";

describe("stockAvailabilityService", () => {
  beforeEach(() => {
    resetPrototypeStockCountsForTest();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("summarizes physical, reserved, QC hold, blocked, and available stock", () => {
    expect(summarizeAvailableStock(prototypeAvailableStock)).toEqual({
      baseUomCode: "PCS",
      physicalQty: "264.000000",
      reservedQty: "42.000000",
      qcHoldQty: "8.000000",
      blockedQty: "7.000000",
      availableQty: "207.000000"
    });
  });

  it("marks rows with QC hold or blocked stock as warning", () => {
    expect(availabilityTone(prototypeAvailableStock[0])).toBe("warning");
    expect(
      availabilityTone({
        ...prototypeAvailableStock[0],
        qcHoldQty: "0.000000",
        blockedQty: "0.000000",
        reservedQty: "1.000000"
      })
    ).toBe("success");
    expect(availabilityTone({ ...prototypeAvailableStock[0], availableQty: "0.000000" })).toBe("danger");
  });

  it("formats decimal quantities with vi-VN separators without using number arithmetic", () => {
    expect(formatQuantity("1250.500000", "PCS")).toBe("1.250,5 PCS");
  });

  it("falls back to prototype rows when the API is not reachable", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("offline")));

    await expect(getAvailableStock({ warehouseId: "wh-hn", sku: "toner-100ml" })).resolves.toEqual([
      prototypeAvailableStock[2]
    ]);
  });

  it("falls back to prototype batch QC transition history", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("offline")));

    await expect(getBatchQCTransitions("batch-cream-2603b")).resolves.toEqual(prototypeBatchQCTransitions);
  });

  it("opens and submits a prototype stock count with variance when the API is not reachable", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("offline")));

    const created = await createStockCount({
      warehouseId: "wh-hcm",
      warehouseCode: "HCM",
      scope: "cycle-count",
      lines: [
        {
          id: "count-line-test-1",
          sku: "SERUM-30ML",
          batchId: "batch-serum-2604a",
          batchNo: "LOT-2604A",
          locationId: "bin-hcm-a01",
          locationCode: "A-01",
          expectedQty: "10.000000",
          baseUomCode: "PCS"
        }
      ]
    });

    const submitted = await submitStockCount(created.id, {
      lines: [{ id: created.lines[0].id, countedQty: "11.500000", note: "cycle count" }]
    });

    expect(created.status).toBe("open");
    expect(submitted.status).toBe("variance_review");
    expect(submitted.lines[0]).toMatchObject({
      countedQty: "11.500000",
      deltaQty: "1.500000",
      counted: true
    });
    await expect(getStockCounts()).resolves.toEqual(expect.arrayContaining([submitted]));
  });
});
