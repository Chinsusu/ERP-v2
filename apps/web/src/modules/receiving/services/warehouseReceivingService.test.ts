import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  createGoodsReceipt,
  getGoodsReceipts,
  markGoodsReceiptInspectReady,
  postGoodsReceipt,
  resetPrototypeGoodsReceiptsForTest,
  submitGoodsReceipt
} from "./warehouseReceivingService";

describe("warehouseReceivingService", () => {
  beforeEach(() => {
    resetPrototypeGoodsReceiptsForTest();
    vi.stubGlobal(
      "fetch",
      vi.fn(() => Promise.reject(new Error("offline")))
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("filters prototype receipts by status", async () => {
    const receipts = await getGoodsReceipts({ status: "inspect_ready" });

    expect(receipts).toHaveLength(1);
    expect(receipts[0].id).toBe("grn-hcm-260427-inspect");
  });

  it("creates a draft and hydrates line data from the selected batch", async () => {
    const receipt = await createGoodsReceipt({
      id: "grn-ui-test",
      receiptNo: "GRN-260427-UI-TEST",
      warehouseId: "wh-hcm-fg",
      locationId: "loc-hcm-fg-recv-01",
      referenceDocType: "purchase_order",
      referenceDocId: "PO-260427-UI-TEST",
      lines: [{ batchId: "batch-cream-2603b", quantity: "6", baseUomCode: "EA" }]
    });

    expect(receipt.status).toBe("draft");
    expect(receipt.lines[0]).toMatchObject({
      sku: "CREAM-50G",
      batchNo: "LOT-2603B",
      qcStatus: "pass",
      quantity: "6.000000"
    });
  });

  it("moves a receipt through submit, inspect-ready, and posted states", async () => {
    const submitted = await submitGoodsReceipt("grn-hcm-260427-draft");
    const inspectReady = await markGoodsReceiptInspectReady(submitted.id);
    const posted = await postGoodsReceipt(inspectReady.id);

    expect(submitted.status).toBe("submitted");
    expect(inspectReady.status).toBe("inspect_ready");
    expect(posted.status).toBe("posted");
    expect(posted.stockMovements).toHaveLength(1);
    expect(posted.stockMovements?.[0]).toMatchObject({
      movementType: "purchase_receipt",
      stockStatus: "qc_hold",
      sourceDocId: "grn-hcm-260427-draft"
    });
  });

  it("rejects posting when batch or QC data is missing", async () => {
    const receipt = await createGoodsReceipt({
      id: "grn-ui-missing-batch",
      receiptNo: "GRN-260427-UI-MISS",
      warehouseId: "wh-hcm-fg",
      locationId: "loc-hcm-fg-recv-01",
      referenceDocType: "purchase_order",
      referenceDocId: "PO-260427-UI-MISS",
      lines: [{ itemId: "item-cream-50g", sku: "CREAM-50G", quantity: "6", baseUomCode: "EA" }]
    });
    await submitGoodsReceipt(receipt.id);
    await markGoodsReceiptInspectReady(receipt.id);

    await expect(postGoodsReceipt(receipt.id)).rejects.toThrow("Batch and QC data");
  });

  it("rejects duplicate posting", async () => {
    await postGoodsReceipt("grn-hcm-260427-inspect");

    await expect(postGoodsReceipt("grn-hcm-260427-inspect")).rejects.toThrow("already posted");
  });
});
