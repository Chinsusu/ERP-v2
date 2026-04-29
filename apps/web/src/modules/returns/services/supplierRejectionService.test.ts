import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  confirmSupplierRejection,
  createSupplierRejection,
  getSupplierRejections,
  resetPrototypeSupplierRejectionsForTest,
  submitSupplierRejection,
  supplierRejectionSampleLines
} from "./supplierRejectionService";

describe("supplierRejectionService", () => {
  beforeEach(() => {
    resetPrototypeSupplierRejectionsForTest();
    vi.stubGlobal(
      "fetch",
      vi.fn(() => Promise.reject(new Error("offline")))
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("filters prototype supplier rejections by status and warehouse", async () => {
    const rows = await getSupplierRejections({ warehouseId: "wh-hcm-fg", status: "confirmed" });

    expect(rows).toHaveLength(1);
    expect(rows[0]).toMatchObject({
      rejectionNo: "SRJ-260429-0001",
      status: "confirmed",
      warehouseId: "wh-hcm-fg"
    });
  });

  it("creates submits and confirms a supplier rejection without stock movement data", async () => {
    const line = supplierRejectionSampleLines[0];
    const created = await createSupplierRejection({
      id: "srj-ui-test",
      rejectionNo: "SRJ-260429-UI-TEST",
      supplierId: "supplier-local",
      supplierCode: "SUP-LOCAL",
      supplierName: "Local Supplier",
      purchaseOrderId: line.purchaseOrderId,
      purchaseOrderNo: line.purchaseOrderNo,
      goodsReceiptId: line.goodsReceiptId,
      goodsReceiptNo: line.goodsReceiptNo,
      inboundQCInspectionId: line.inboundQCInspectionId,
      warehouseId: line.warehouseId,
      warehouseCode: line.warehouseCode,
      reason: "damaged packaging",
      lines: [
        {
          ...line,
          rejectedQuantity: "6"
        }
      ],
      attachments: [
        {
          fileName: "damage-photo.jpg",
          source: "inbound_qc"
        }
      ]
    });
    const submitted = await submitSupplierRejection(created.id);
    const confirmed = await confirmSupplierRejection(created.id);

    expect(created).toMatchObject({
      status: "draft",
      rejectionNo: "SRJ-260429-UI-TEST",
      attachments: [expect.objectContaining({ objectKey: "supplier-rejections/srj-ui-test/damage-photo.jpg" })]
    });
    expect(created.lines[0].rejectedQuantity).toBe("6.000000");
    expect(submitted).toMatchObject({
      previousStatus: "draft",
      currentStatus: "submitted"
    });
    expect(confirmed).toMatchObject({
      previousStatus: "submitted",
      currentStatus: "confirmed",
      rejection: {
        status: "confirmed",
        auditLogId: "audit-srj-ui-test-confirmed"
      }
    });
  });

  it("maps API supplier rejection rows without prototype fallback", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          success: true,
          data: [
            {
              id: "srj-api-1",
              org_id: "org-my-pham",
              rejection_no: "SRJ-API-1",
              supplier_id: "supplier-local",
              supplier_code: "SUP-LOCAL",
              supplier_name: "Local Supplier",
              purchase_order_id: "po-api-1",
              purchase_order_no: "PO-API-1",
              goods_receipt_id: "grn-api-1",
              goods_receipt_no: "GRN-API-1",
              inbound_qc_inspection_id: "iqc-api-1",
              warehouse_id: "wh-hcm-fg",
              warehouse_code: "WH-HCM-FG",
              status: "draft",
              reason: "damaged packaging",
              lines: [
                {
                  id: "srj-api-line-1",
                  purchase_order_line_id: "po-api-line-1",
                  goods_receipt_line_id: "grn-api-line-1",
                  inbound_qc_inspection_id: "iqc-api-1",
                  item_id: "item-api-1",
                  sku: "SERUM-30ML",
                  item_name: "Vitamin C Serum",
                  batch_id: "batch-api-1",
                  batch_no: "LOT-API-1",
                  lot_no: "LOT-API-1",
                  expiry_date: "2027-04-01",
                  rejected_qty: "6.000000",
                  uom_code: "EA",
                  base_uom_code: "EA",
                  reason: "damaged packaging"
                }
              ],
              attachments: [],
              created_at: "2026-04-29T09:00:00Z",
              created_by: "user-warehouse-lead",
              updated_at: "2026-04-29T09:00:00Z",
              updated_by: "user-warehouse-lead"
            }
          ],
          request_id: "req-srj-list"
        }),
        { status: 200 }
      )
    );
    vi.stubGlobal("fetch", fetchMock);

    const rows = await getSupplierRejections({ warehouseId: "wh-hcm-fg", status: "draft" });

    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/supplier-rejections?warehouse_id=wh-hcm-fg&status=draft",
      {
        headers: {
          Authorization: "Bearer local-dev-access-token"
        }
      }
    );
    expect(rows[0]).toMatchObject({
      id: "srj-api-1",
      rejectionNo: "SRJ-API-1",
      lines: [expect.objectContaining({ rejectedQuantity: "6.000000", baseUOMCode: "EA" })]
    });
  });
});
