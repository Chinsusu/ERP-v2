import { describe, expect, it } from "vitest";
import {
  createReturnInspection,
  formatReturnInspectionCondition,
  formatReturnInspectionDisposition,
  formatReturnDisposition,
  getReturnReceipts,
  matchesReturnReceiptCode,
  receiveReturn,
  returnDispositionTone,
  returnInspectionConditionTone,
  returnInspectionDispositionTone,
  returnInspectionStatusTone,
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

  it("matches return receipts by receipt order tracking return or scan code", async () => {
    const [receipt] = await getReturnReceipts({ warehouseId: "wh-hcm" });

    expect(matchesReturnReceiptCode(receipt, "RR-260426-0001")).toBe(true);
    expect(matchesReturnReceiptCode(receipt, "SO-260425-099")).toBe(true);
    expect(matchesReturnReceiptCode(receipt, "GHN260425099")).toBe(true);
    expect(matchesReturnReceiptCode(receipt, "RET-260425-099")).toBe(true);
  });

  it("creates return inspection results with target status and risk", async () => {
    const [receipt] = await getReturnReceipts({ warehouseId: "wh-hcm" });

    expect(
      createReturnInspection({
        receipt,
        condition: "qa_required",
        disposition: "qa_hold",
        note: "needs QA",
        evidenceLabel: "photo-001"
      })
    ).toMatchObject({
      receiptNo: "RR-260426-0001",
      status: "RETURN_QA_HOLD",
      targetLocation: "return-qa-hold",
      riskLevel: "medium",
      evidenceLabel: "photo-001"
    });
  });

  it("maps inspection conditions dispositions and status to labels and tones", () => {
    expect(returnInspectionConditionTone("intact")).toBe("success");
    expect(returnInspectionConditionTone("damaged")).toBe("danger");
    expect(returnInspectionDispositionTone("not_usable")).toBe("danger");
    expect(returnInspectionStatusTone("RETURN_QA_HOLD")).toBe("warning");
    expect(formatReturnInspectionCondition("seal_torn")).toBe("Seal torn");
    expect(formatReturnInspectionDisposition("qa_hold")).toBe("QA hold");
  });
});
