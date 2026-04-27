import { afterEach, describe, expect, it, vi } from "vitest";
import {
  availabilityTone,
  formatQuantity,
  getAvailableStock,
  prototypeAvailableStock,
  summarizeAvailableStock
} from "./stockAvailabilityService";

describe("stockAvailabilityService", () => {
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
});
