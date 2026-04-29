import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  approvePurchaseOrder,
  cancelPurchaseOrder,
  closePurchaseOrder,
  createPurchaseOrder,
  getPurchaseOrders,
  resetPrototypePurchaseOrdersForTest,
  submitPurchaseOrder,
  updatePurchaseOrder
} from "./purchaseOrderService";

describe("purchaseOrderService", () => {
  beforeEach(() => {
    resetPrototypePurchaseOrdersForTest();
    vi.stubGlobal(
      "fetch",
      vi.fn(() => Promise.reject(new Error("offline")))
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("filters prototype purchase orders by status and search", async () => {
    const orders = await getPurchaseOrders({ status: "submitted", search: "vina" });

    expect(orders).toHaveLength(1);
    expect(orders[0]).toMatchObject({
      id: "po-260429-0002",
      status: "submitted",
      supplierCode: "SUP-PKG-VINA"
    });
  });

  it("maps API list metadata without loading detail lines", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          success: true,
          data: [
            {
              id: "po-api-1",
              po_no: "PO-API-1",
              supplier_id: "sup-api",
              supplier_code: "SUP-API",
              supplier_name: "API Supplier",
              warehouse_id: "wh-hcm-rm",
              warehouse_code: "WH-HCM-RM",
              expected_date: "2026-05-02",
              status: "approved",
              currency_code: "VND",
              total_amount: "2020000.00",
              line_count: 2,
              received_line_count: 1,
              created_at: "2026-04-29T09:00:00Z",
              updated_at: "2026-04-29T09:15:00Z",
              version: 3
            }
          ],
          request_id: "req-purchase-list"
        }),
        { status: 200 }
      )
    );
    vi.stubGlobal("fetch", fetchMock);

    const orders = await getPurchaseOrders({ search: "PO-API", warehouseId: "wh-hcm-rm" });

    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/purchase-orders?search=PO-API&warehouse_id=wh-hcm-rm",
      {
        headers: {
          Authorization: "Bearer local-dev-access-token"
        }
      }
    );
    expect(orders[0]).toMatchObject({
      id: "po-api-1",
      lineCount: 2,
      receivedLineCount: 1,
      lines: [],
      totalAmount: "2020000.00",
      version: 3
    });
  });

  it("does not hide API permission errors behind prototype fallback", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            error: {
              code: "FORBIDDEN",
              message: "Permission denied",
              details: { permission: "purchase:view" },
              request_id: "req-denied"
            }
          }),
          { status: 403 }
        )
      )
    );

    await expect(getPurchaseOrders()).rejects.toMatchObject({
      name: "ApiError",
      status: 403,
      code: "FORBIDDEN",
      requestId: "req-denied"
    });
  });

  it("creates a draft with normalized decimal lines and VND totals", async () => {
    const order = await createPurchaseOrder({
      supplierId: "sup-rm-bioactive",
      warehouseId: "wh-hcm-rm",
      expectedDate: "2026-05-02",
      currencyCode: "VND",
      lines: [
        {
          itemId: "item-serum-30ml",
          orderedQty: "2",
          uomCode: "EA",
          unitPrice: "125000.5"
        }
      ]
    });

    expect(order.status).toBe("draft");
    expect(order.lines[0]).toMatchObject({
      skuCode: "SERUM-30ML",
      orderedQty: "2.000000",
      baseOrderedQty: "2.000000",
      baseUomCode: "EA",
      conversionFactor: "1.000000",
      unitPrice: "125000.5000",
      lineAmount: "250001.00"
    });
    expect(order.totalAmount).toBe("250001.00");
  });

  it("validates that a purchase order has at least one line", async () => {
    await expect(
      createPurchaseOrder({
        supplierId: "sup-rm-bioactive",
        warehouseId: "wh-hcm-rm",
        expectedDate: "2026-05-02",
        currencyCode: "VND",
        lines: []
      })
    ).rejects.toThrow("At least one line item is required");
  });

  it("replaces draft lines and checks optimistic version", async () => {
    const created = await createPurchaseOrder({
      supplierId: "sup-rm-bioactive",
      warehouseId: "wh-hcm-rm",
      expectedDate: "2026-05-02",
      currencyCode: "VND",
      lines: [{ itemId: "item-serum-30ml", orderedQty: "1", uomCode: "EA", unitPrice: "125000" }]
    });

    const updated = await updatePurchaseOrder(created.id, {
      expectedVersion: 1,
      lines: [{ itemId: "item-cream-50g", orderedQty: "3", uomCode: "EA", unitPrice: "95000" }]
    });

    expect(updated.version).toBe(2);
    expect(updated.lines[0].skuCode).toBe("CREAM-50G");
    expect(updated.totalAmount).toBe("285000.00");
    await expect(updatePurchaseOrder(created.id, { expectedVersion: 1 })).rejects.toThrow("version changed");
  });

  it("submits approves closes and cancels through the prototype state flow", async () => {
    const submitted = await submitPurchaseOrder("po-260429-0001", 1);

    expect(submitted.previousStatus).toBe("draft");
    expect(submitted.currentStatus).toBe("submitted");
    expect(submitted.purchaseOrder.version).toBe(2);

    const approved = await approvePurchaseOrder("po-260429-0001", 2);

    expect(approved.previousStatus).toBe("submitted");
    expect(approved.currentStatus).toBe("approved");
    expect(approved.purchaseOrder.version).toBe(3);

    const closed = await closePurchaseOrder("po-260429-0001", 3);

    expect(closed.previousStatus).toBe("approved");
    expect(closed.currentStatus).toBe("closed");
    expect(closed.purchaseOrder.version).toBe(4);

    const cancelled = await cancelPurchaseOrder("po-260429-0002", "supplier delay", 2);

    expect(cancelled.previousStatus).toBe("submitted");
    expect(cancelled.currentStatus).toBe("cancelled");
    expect(cancelled.purchaseOrder.cancelReason).toBe("supplier delay");
  });
});
