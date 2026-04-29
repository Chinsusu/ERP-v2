import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  createInboundQCInspection,
  defaultInboundQCChecklist,
  getInboundQCInspections,
  passInboundQCInspection,
  resetPrototypeInboundQCInspectionsForTest,
  startInboundQCInspection
} from "./inboundQCService";

describe("inboundQCService", () => {
  beforeEach(() => {
    resetPrototypeInboundQCInspectionsForTest();
    vi.stubGlobal(
      "fetch",
      vi.fn(() => Promise.reject(new Error("offline")))
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("creates and filters prototype inbound QC inspections", async () => {
    const created = await createInboundQCInspection({
      goodsReceiptId: "grn-hcm-260427-inspect",
      goodsReceiptLineId: "grn-line-inspect-001",
      inspectorId: "user-qa",
      checklist: defaultInboundQCChecklist
    });

    const rows = await getInboundQCInspections({ status: "pending", warehouseId: "wh-hcm-fg" });

    expect(created.inspection).toMatchObject({
      goodsReceiptNo: "GRN-260427-0003",
      sku: "CREAM-50G",
      quantity: "12.000000",
      status: "pending"
    });
    expect(rows).toHaveLength(1);
    expect(rows[0].id).toBe(created.inspection.id);
  });

  it("starts and passes an inspection with completed checklist", async () => {
    const created = await createInboundQCInspection({
      goodsReceiptId: "grn-hcm-260427-inspect",
      goodsReceiptLineId: "grn-line-inspect-001",
      inspectorId: "user-qa",
      checklist: defaultInboundQCChecklist
    });
    const started = await startInboundQCInspection(created.inspection.id);
    const checklist = started.inspection.checklist.map((item) => ({
      ...item,
      status: item.required ? ("pass" as const) : item.status
    }));

    const passed = await passInboundQCInspection(created.inspection.id, {
      passedQuantity: "12",
      checklist,
      note: "COA checked"
    });

    expect(started.currentStatus).toBe("in_progress");
    expect(passed.inspection).toMatchObject({
      status: "completed",
      result: "pass",
      passedQuantity: "12.000000",
      failedQuantity: "0.000000",
      holdQuantity: "0.000000"
    });
  });

  it("maps API responses without prototype fallback", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          success: true,
          data: [
            {
              id: "iqc-api-1",
              org_id: "org-my-pham",
              goods_receipt_id: "grn-api-1",
              goods_receipt_no: "GRN-API-1",
              goods_receipt_line_id: "grn-api-line-1",
              purchase_order_id: "po-api-1",
              purchase_order_line_id: "po-api-line-1",
              item_id: "item-api-1",
              sku: "SERUM-30ML",
              item_name: "Vitamin C Serum",
              batch_id: "batch-api-1",
              batch_no: "LOT-API-1",
              lot_no: "LOT-API-1",
              expiry_date: "2028-03-01",
              warehouse_id: "wh-hcm-fg",
              location_id: "loc-hcm-fg-recv-01",
              quantity: "6.000000",
              uom_code: "EA",
              inspector_id: "user-qa",
              status: "pending",
              passed_qty: "0.000000",
              failed_qty: "0.000000",
              hold_qty: "0.000000",
              checklist: defaultInboundQCChecklist,
              created_at: "2026-04-29T09:00:00Z",
              created_by: "user-qa",
              updated_at: "2026-04-29T09:00:00Z",
              updated_by: "user-qa"
            }
          ],
          request_id: "req-iqc-list"
        }),
        { status: 200 }
      )
    );
    vi.stubGlobal("fetch", fetchMock);

    const rows = await getInboundQCInspections({ warehouseId: "wh-hcm-fg" });

    expect(fetchMock).toHaveBeenCalledWith("http://localhost:8080/api/v1/inbound-qc-inspections?warehouse_id=wh-hcm-fg", {
      headers: {
        Authorization: "Bearer local-dev-access-token"
      }
    });
    expect(rows[0]).toMatchObject({
      id: "iqc-api-1",
      goodsReceiptNo: "GRN-API-1",
      passedQuantity: "0.000000",
      checklist: expect.arrayContaining([expect.objectContaining({ code: "PACKAGING" })])
    });
  });
});
