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

  it("filters prototype receipts by purchase order reference", async () => {
    const receipts = await getGoodsReceipts({ referenceDocId: "po-260429-0003" });

    expect(receipts.map((receipt) => receipt.referenceDocId)).toEqual(["po-260429-0003", "po-260429-0003"]);
  });

  it("creates a draft and hydrates line data from the selected batch", async () => {
    const receipt = await createGoodsReceipt({
      id: "grn-ui-test",
      receiptNo: "GRN-260427-UI-TEST",
      warehouseId: "wh-hcm-fg",
      locationId: "loc-hcm-fg-recv-01",
      referenceDocType: "purchase_order",
      referenceDocId: "PO-260427-UI-TEST",
      supplierId: "sup-rm-bioactive",
      deliveryNoteNo: "dn-260427-ui-test",
      lines: [
        {
          purchaseOrderLineId: "po-line-ui-test-001",
          batchId: "batch-cream-2603b",
          lotNo: "lot-2603b",
          expiryDate: "2028-03-01",
          quantity: "6",
          uomCode: "EA",
          baseUomCode: "EA",
          packagingStatus: "intact"
        }
      ]
    });

    expect(receipt.deliveryNoteNo).toBe("DN-260427-UI-TEST");
    expect(receipt.status).toBe("draft");
    expect(receipt.lines[0]).toMatchObject({
      purchaseOrderLineId: "po-line-ui-test-001",
      sku: "CREAM-50G",
      batchNo: "LOT-2603B",
      lotNo: "LOT-2603B",
      expiryDate: "2028-03-01",
      packagingStatus: "intact",
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
      supplierId: "sup-rm-bioactive",
      deliveryNoteNo: "DN-260427-UI-MISS",
      lines: [
        {
          purchaseOrderLineId: "po-line-ui-miss-001",
          itemId: "item-cream-50g",
          sku: "CREAM-50G",
          batchNo: "LOT-2603B",
          lotNo: "LOT-2603B",
          expiryDate: "2028-03-01",
          quantity: "6",
          uomCode: "EA",
          baseUomCode: "EA",
          packagingStatus: "intact"
        }
      ]
    });
    await submitGoodsReceipt(receipt.id);
    await markGoodsReceiptInspectReady(receipt.id);

    await expect(postGoodsReceipt(receipt.id)).rejects.toThrow("Batch and QC data");
  });

  it("rejects duplicate posting", async () => {
    await postGoodsReceipt("grn-hcm-260427-inspect");

    await expect(postGoodsReceipt("grn-hcm-260427-inspect")).rejects.toThrow("already posted");
  });

  it("does not use prototype fallback when the API returns a validation error", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(() =>
        Promise.resolve(
          new Response(
            JSON.stringify({
              success: false,
              error: {
                code: "VALIDATION_ERROR",
                message: "delivery note is required",
                request_id: "req-receiving-validation"
              }
            }),
            { status: 400, headers: { "Content-Type": "application/json" } }
          )
        )
      )
    );

    await expect(
      createGoodsReceipt({
        warehouseId: "wh-hcm-fg",
        locationId: "loc-hcm-fg-recv-01",
        referenceDocType: "purchase_order",
        referenceDocId: "PO-260427-UI-TEST",
        supplierId: "sup-rm-bioactive",
        deliveryNoteNo: "",
        lines: [
          {
            purchaseOrderLineId: "po-line-ui-test-001",
            batchId: "batch-cream-2603b",
            lotNo: "LOT-2603B",
            expiryDate: "2028-03-01",
            quantity: "6",
            uomCode: "EA",
            baseUomCode: "EA",
            packagingStatus: "intact"
          }
        ]
      })
    ).rejects.toThrow("delivery note is required");
  });
});
