import { afterEach, describe, expect, it, vi } from "vitest";
import {
  availabilityTone,
  getAvailableStock,
  prototypeAvailableStock,
  summarizeAvailableStock
} from "./stockAvailabilityService";

describe("stockAvailabilityService", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("summarizes physical, reserved, hold, and available stock", () => {
    expect(summarizeAvailableStock(prototypeAvailableStock)).toEqual({
      physicalStock: 264,
      reservedStock: 42,
      holdStock: 15,
      availableStock: 207
    });
  });

  it("marks rows with hold stock as warning", () => {
    expect(availabilityTone(prototypeAvailableStock[0])).toBe("warning");
    expect(availabilityTone({ ...prototypeAvailableStock[0], holdStock: 0, reservedStock: 1 })).toBe("success");
    expect(availabilityTone({ ...prototypeAvailableStock[0], availableStock: 0 })).toBe("danger");
  });

  it("falls back to prototype rows when the API is not reachable", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("offline")));

    await expect(getAvailableStock({ warehouseId: "wh-hn", sku: "toner-100ml" })).resolves.toEqual([
      prototypeAvailableStock[2]
    ]);
  });
});
