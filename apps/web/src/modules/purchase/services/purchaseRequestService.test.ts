import { afterEach, describe, expect, it, vi } from "vitest";
import {
  approvePurchaseRequest,
  convertPurchaseRequestToPurchaseOrder,
  getPurchaseRequest,
  getPurchaseRequests,
  purchaseRequestStatusLabel,
  submitPurchaseRequest
} from "./purchaseRequestService";

describe("purchaseRequestService", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("loads purchase requests from the backend with lifecycle fields", async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse([
        {
          id: "pr-draft-001",
          request_no: "PR-DRAFT-260505-0001",
          source_production_plan_id: "pp-260505-0001",
          source_production_plan_no: "PP-260505-0001",
          status: "approved",
          lines: [purchaseRequestLineApi()],
          created_at: "2026-05-05T03:00:00Z",
          submitted_at: "2026-05-05T04:00:00Z",
          approved_at: "2026-05-05T05:00:00Z"
        }
      ])
    );
    vi.stubGlobal("fetch", fetchMock);

    const requests = await getPurchaseRequests({ status: "approved", search: "PP-260505" });

    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining("/purchase-requests?q=PP-260505&status=approved"),
      expect.objectContaining({ headers: expect.any(Object) })
    );
    expect(requests[0]).toMatchObject({
      id: "pr-draft-001",
      requestNo: "PR-DRAFT-260505-0001",
      sourceProductionPlanNo: "PP-260505-0001",
      status: "approved",
      approvedAt: "2026-05-05T05:00:00Z"
    });
    expect(requests[0].lines[0]).toMatchObject({
      sku: "ACI_BHA",
      requestedQty: "0.099900",
      uomCode: "KG"
    });
  });

  it("submits and approves a purchase request", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        jsonResponse({
          purchase_request: purchaseRequestApi({ status: "submitted" }),
          previous_status: "draft",
          current_status: "submitted",
          audit_log_id: "audit-submit"
        })
      )
      .mockResolvedValueOnce(
        jsonResponse({
          purchase_request: purchaseRequestApi({ status: "approved" }),
          previous_status: "submitted",
          current_status: "approved",
          audit_log_id: "audit-approve"
        })
      );
    vi.stubGlobal("fetch", fetchMock);

    const submitted = await submitPurchaseRequest("pr-draft-001");
    const approved = await approvePurchaseRequest("pr-draft-001");

    expect(submitted.currentStatus).toBe("submitted");
    expect(approved.currentStatus).toBe("approved");
    expect(purchaseRequestStatusLabel(approved.currentStatus)).toBe("Đã duyệt");
  });

  it("converts an approved purchase request to a purchase order", async () => {
    const fetchMock = vi.fn(async (_url, init) => {
      expect(String((init as RequestInit).body)).toContain('"supplier_id":"sup-rm-bioactive"');

      return jsonResponse({
        purchase_request: purchaseRequestApi({
          status: "converted_to_po",
          converted_purchase_order_id: "po-260505-0001",
          converted_purchase_order_no: "PO-260505-0001"
        }),
        purchase_order: purchaseOrderApi(),
        audit_log_id: "audit-convert"
      });
    });
    vi.stubGlobal("fetch", fetchMock);

    const result = await convertPurchaseRequestToPurchaseOrder("pr-draft-001", {
      supplierId: "sup-rm-bioactive",
      warehouseId: "wh-hcm-rm",
      expectedDate: "2026-05-08"
    });

    expect(result.purchaseRequest.status).toBe("converted_to_po");
    expect(result.purchaseOrder.poNo).toBe("PO-260505-0001");
  });

  it("loads one purchase request by id", async () => {
    const fetchMock = vi.fn(async () => jsonResponse(purchaseRequestApi({ status: "draft" })));
    vi.stubGlobal("fetch", fetchMock);

    const request = await getPurchaseRequest("pr-draft-001");

    expect(fetchMock).toHaveBeenCalledWith(expect.stringContaining("/purchase-requests/pr-draft-001"), expect.any(Object));
    expect(request.requestNo).toBe("PR-DRAFT-260505-0001");
  });
});

function jsonResponse(data: unknown) {
  return new Response(JSON.stringify({ success: true, request_id: "req-test", data }), {
    status: 200,
    headers: { "Content-Type": "application/json" }
  });
}

function purchaseRequestApi(overrides: Record<string, unknown> = {}) {
  return {
    id: "pr-draft-001",
    request_no: "PR-DRAFT-260505-0001",
    source_production_plan_id: "pp-260505-0001",
    source_production_plan_no: "PP-260505-0001",
    status: "draft",
    lines: [purchaseRequestLineApi()],
    created_at: "2026-05-05T03:00:00Z",
    ...overrides
  };
}

function purchaseRequestLineApi() {
  return {
    id: "pr-line-001",
    line_no: 1,
    source_production_plan_line_id: "pp-line-001",
    item_id: "item-aci-bha",
    sku: "ACI_BHA",
    item_name: "ACID SALICYLIC",
    requested_qty: "0.099900",
    uom_code: "KG"
  };
}

function purchaseOrderApi() {
  return {
    id: "po-260505-0001",
    po_no: "PO-260505-0001",
    supplier_id: "sup-rm-bioactive",
    supplier_name: "BioActive Raw Materials",
    warehouse_id: "wh-hcm-rm",
    warehouse_code: "WH-HCM-RM",
    expected_date: "2026-05-08",
    status: "draft",
    currency_code: "VND",
    subtotal_amount: "0",
    total_amount: "0",
    lines: [
      {
        id: "po-line-001",
        line_no: 1,
        item_id: "item-aci-bha",
        sku_code: "ACI_BHA",
        item_name: "ACID SALICYLIC",
        ordered_qty: "0.099900",
        received_qty: "0.000000",
        uom_code: "KG",
        base_ordered_qty: "0.099900",
        base_received_qty: "0.000000",
        base_uom_code: "KG",
        conversion_factor: "1",
        unit_price: "0",
        currency_code: "VND",
        line_amount: "0",
        expected_date: "2026-05-08"
      }
    ],
    created_at: "2026-05-05T06:00:00Z",
    updated_at: "2026-05-05T06:00:00Z",
    version: 1
  };
}
