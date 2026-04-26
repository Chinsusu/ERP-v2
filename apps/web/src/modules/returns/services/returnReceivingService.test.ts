import { describe, expect, it } from "vitest";
import {
  formatReturnDisposition,
  getReturnReceipts,
  receiveReturn,
  returnDispositionTone,
  returnReceiptStatusTone
} from "./returnReceivingService";

describe("returnReceivingService", () => {
  it("filters return receipts by warehouse and pending inspection status", async () => {
    await expect(
      getReturnReceipts({
        warehouseId: "wh-hcm",
        status: "pending_inspection"
      })
    ).resolves.toMatchObject([
      {
        receiptNo: "RR-260426-0001",
        status: "pending_inspection"
      }
    ]);
  });

  it("creates a reusable return receipt movement for a known tracking code", async () => {
    await expect(
      receiveReturn({
        warehouseId: "wh-hcm",
        warehouseCode: "HCM",
        source: "CARRIER",
        code: "GHN260426001",
        packageCondition: "sealed",
        disposition: "reusable"
      })
    ).resolves.toMatchObject({
      originalOrderNo: "SO-260426-001",
      unknownCase: false,
      status: "pending_inspection",
      targetLocation: "return-area-pending-inspection",
      stockMovement: {
        movementType: "RETURN_RECEIPT",
        targetStockStatus: "return_pending"
      }
    });
  });

  it("creates an unknown return case when no order or tracking code matches", async () => {
    await expect(
      receiveReturn({
        warehouseId: "wh-hcm",
        warehouseCode: "HCM",
        source: "SHIPPER",
        code: "UNKNOWN-RETURN",
        packageCondition: "damaged box",
        disposition: "needs_inspection"
      })
    ).resolves.toMatchObject({
      unknownCase: true,
      targetLocation: "return-inspection-queue",
      stockMovement: undefined
    });
  });

  it("routes not reusable returns to the lab damaged placeholder", async () => {
    await expect(
      receiveReturn({
        warehouseId: "wh-hcm",
        warehouseCode: "HCM",
        source: "SHIPPER",
        code: "SO-260426-004",
        packageCondition: "leaking",
        disposition: "not_reusable"
      })
    ).resolves.toMatchObject({
      targetLocation: "lab-damaged-placeholder",
      stockMovement: undefined
    });
  });

  it("maps returns status and disposition to UI labels and tones", () => {
    expect(returnReceiptStatusTone("pending_inspection")).toBe("warning");
    expect(returnDispositionTone("reusable")).toBe("success");
    expect(returnDispositionTone("not_reusable")).toBe("danger");
    expect(formatReturnDisposition("needs_inspection")).toBe("Needs inspection");
  });
});
