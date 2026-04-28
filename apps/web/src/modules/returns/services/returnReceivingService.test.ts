import { beforeEach, describe, expect, it } from "vitest";
import {
  applyDispositionToReceipt,
  applyInspectionToReceipt,
  applyReturnDisposition,
  createReturnAttachment,
  createReturnDispositionAction,
  createReturnInspection,
  formatReturnInspectionCondition,
  formatReturnInspectionDisposition,
  formatReturnDisposition,
  getReturnReceipts,
  inspectReturn,
  matchesReturnReceiptCode,
  receiveReturn,
  resetPrototypeReturnReceiptsForTest,
  returnDispositionTone,
  returnInspectionConditionTone,
  returnInspectionDispositionTone,
  returnInspectionStatusTone,
  returnReceiptStatusTone,
  uploadReturnAttachment
} from "./returnReceivingService";

describe("returnReceivingService", () => {
  beforeEach(() => {
    resetPrototypeReturnReceiptsForTest();
  });

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

  it("rejects duplicate scans across tracking and order number", async () => {
    const input = {
      warehouseId: "wh-hcm",
      warehouseCode: "HCM",
      source: "CARRIER" as const,
      code: "GHN260426001",
      packageCondition: "sealed",
      disposition: "needs_inspection" as const
    };

    await expect(receiveReturn(input)).resolves.toMatchObject({ originalOrderNo: "SO-260426-001" });
    await expect(receiveReturn({ ...input, code: "SO-260426-001" })).rejects.toThrow("already exists");
  });

  it("rejects returns before handover or delivery", async () => {
    await expect(
      receiveReturn({
        warehouseId: "wh-hcm",
        warehouseCode: "HCM",
        source: "CARRIER",
        code: "GHN260426009",
        packageCondition: "sealed",
        disposition: "needs_inspection"
      })
    ).rejects.toThrow("not eligible");
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
        condition: "missing_accessory",
        disposition: "needs_inspection",
        note: "needs QA",
        evidenceLabel: "photo-001"
      })
    ).toMatchObject({
      receiptNo: "RR-260426-0001",
      status: "return_qa_hold",
      targetLocation: "return-qa-hold",
      riskLevel: "high",
      evidenceLabel: "photo-001"
    });
  });

  it("updates receipt state after inspection and disposition actions", async () => {
    const [receipt] = await getReturnReceipts({ warehouseId: "wh-hcm" });
    const inspection = await inspectReturn({
      receipt,
      condition: "intact",
      disposition: "reusable",
      note: "usable",
      evidenceLabel: "photo-001"
    });
    const inspectedReceipt = applyInspectionToReceipt(receipt, inspection);

    expect(inspectedReceipt).toMatchObject({
      status: "inspected",
      disposition: "reusable",
      targetLocation: "return-area-qc-release",
      stockMovement: undefined
    });

    const action = await applyReturnDisposition({
      receipt: inspectedReceipt,
      disposition: "reusable",
      note: "ready"
    });
    const dispositionedReceipt = applyDispositionToReceipt(inspectedReceipt, action);

    expect(action).toMatchObject({
      actionCode: "route_to_putaway",
      targetLocation: "return-putaway-ready",
      targetStockStatus: "return_pending"
    });
    expect(dispositionedReceipt).toMatchObject({
      status: "dispositioned",
      targetLocation: "return-putaway-ready",
      stockMovement: undefined
    });
  });

  it("creates prototype disposition actions for QA hold", async () => {
    const [receipt] = await getReturnReceipts({ warehouseId: "wh-hcm" });

    expect(
      createReturnDispositionAction({
        receipt,
        disposition: "needs_inspection",
        note: "QA review"
      })
    ).toMatchObject({
      actionCode: "route_to_quarantine_hold",
      targetLocation: "return-quarantine-hold",
      targetStockStatus: "qc_hold"
    });
  });

  it("creates return attachment metadata after inspection", async () => {
    const [receipt] = await getReturnReceipts({ warehouseId: "wh-hcm" });
    const inspection = await inspectReturn({
      receipt,
      condition: "intact",
      disposition: "reusable",
      note: "usable",
      evidenceLabel: "photo-001"
    });
    const inspectedReceipt = applyInspectionToReceipt(receipt, inspection);
    const file = new File(["fake image bytes"], "return-photo.png", { type: "image/png" });

    const attachment = await uploadReturnAttachment({
      receipt: inspectedReceipt,
      inspectionId: inspection.id,
      file,
      note: "front photo"
    });

    expect(attachment).toMatchObject({
      receiptNo: "RR-260426-0001",
      inspectionId: inspection.id,
      fileName: "return-photo.png",
      mimeType: "image/png",
      storageBucket: "erp-return-attachments-dev",
      status: "active"
    });
    expect(
      createReturnAttachment({
        receipt: inspectedReceipt,
        inspectionId: inspection.id,
        file,
        note: "front photo"
      })
    ).toMatchObject({
      auditLogId: "audit-return-attachment-prototype",
      fileSizeBytes: file.size
    });
  });

  it("maps inspection conditions dispositions and status to labels and tones", () => {
    expect(returnInspectionConditionTone("intact")).toBe("success");
    expect(returnInspectionConditionTone("damaged")).toBe("danger");
    expect(returnInspectionDispositionTone("not_reusable")).toBe("danger");
    expect(returnInspectionStatusTone("return_qa_hold")).toBe("warning");
    expect(formatReturnInspectionCondition("seal_torn")).toBe("Seal torn");
    expect(formatReturnInspectionDisposition("needs_inspection")).toBe("Needs QA");
  });
});
