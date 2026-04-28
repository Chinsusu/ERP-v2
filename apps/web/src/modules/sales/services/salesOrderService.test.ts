import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  cancelSalesOrder,
  confirmSalesOrder,
  createSalesOrder,
  getSalesOrders,
  resetPrototypeSalesOrdersForTest,
  updateSalesOrder
} from "./salesOrderService";

describe("salesOrderService", () => {
  beforeEach(() => {
    resetPrototypeSalesOrdersForTest();
    vi.stubGlobal(
      "fetch",
      vi.fn(() => Promise.reject(new Error("offline")))
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("filters prototype sales orders by status and search", async () => {
    const orders = await getSalesOrders({ status: "confirmed", search: "shopee" });

    expect(orders).toHaveLength(1);
    expect(orders[0]).toMatchObject({
      id: "so-260428-0002",
      status: "confirmed",
      customerCode: "CUS-MP-SHOPEE"
    });
  });

  it("maps API list metadata without loading details", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          success: true,
          data: [
            {
              id: "so-api-1",
              order_no: "SO-API-1",
              customer_id: "cus-api",
              customer_code: "CUS-API",
              customer_name: "API Customer",
              channel: "B2B",
              warehouse_id: "wh-hcm-fg",
              warehouse_code: "WH-HCM-FG",
              order_date: "2026-04-28",
              status: "confirmed",
              currency_code: "VND",
              total_amount: "2020000.00",
              line_count: 2,
              reserved_line_count: 1,
              created_at: "2026-04-28T09:00:00Z",
              updated_at: "2026-04-28T09:15:00Z",
              version: 3
            }
          ],
          pagination: { page: 1, page_size: 20, total: 1 },
          request_id: "req-sales-list"
        }),
        { status: 200 }
      )
    );
    vi.stubGlobal("fetch", fetchMock);

    const orders = await getSalesOrders({ search: "SO-API", warehouseId: "wh-hcm-fg" });

    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/sales-orders?q=SO-API&warehouse_id=wh-hcm-fg",
      {
        headers: {
          Authorization: "Bearer local-dev-access-token"
        }
      }
    );
    expect(orders[0]).toMatchObject({
      id: "so-api-1",
      lineCount: 2,
      lines: [],
      totalAmount: "2020000.00",
      version: 3
    });
  });

  it("creates a draft with normalized decimal lines and VND totals", async () => {
    const order = await createSalesOrder({
      customerId: "cus-dl-minh-anh",
      channel: "B2B",
      warehouseId: "wh-hcm-fg",
      orderDate: "2026-04-28",
      currencyCode: "VND",
      lines: [
        {
          itemId: "item-serum-30ml",
          orderedQty: "2",
          uomCode: "EA",
          unitPrice: "125000.5",
          lineDiscountAmount: "1000"
        }
      ]
    });

    expect(order.status).toBe("draft");
    expect(order.lines[0]).toMatchObject({
      skuCode: "SERUM-30ML",
      orderedQty: "2.000000",
      unitPrice: "125000.5000",
      lineAmount: "249001.00"
    });
    expect(order).toMatchObject({
      subtotalAmount: "250001.00",
      discountAmount: "1000.00",
      totalAmount: "249001.00"
    });
  });

  it("replaces draft lines and checks optimistic version", async () => {
    const created = await createSalesOrder({
      customerId: "cus-dl-minh-anh",
      channel: "B2B",
      warehouseId: "wh-hcm-fg",
      orderDate: "2026-04-28",
      currencyCode: "VND",
      lines: [{ itemId: "item-serum-30ml", orderedQty: "1", uomCode: "EA", unitPrice: "125000" }]
    });

    const updated = await updateSalesOrder(created.id, {
      expectedVersion: 1,
      lines: [{ itemId: "item-cream-50g", orderedQty: "3", uomCode: "EA", unitPrice: "95000" }]
    });

    expect(updated.version).toBe(2);
    expect(updated.lines[0].skuCode).toBe("CREAM-50G");
    expect(updated.totalAmount).toBe("285000.00");
    await expect(updateSalesOrder(created.id, { expectedVersion: 1 })).rejects.toThrow("version changed");
  });

  it("confirms and cancels through the prototype state flow", async () => {
    const confirmed = await confirmSalesOrder("so-260428-0001", 1);

    expect(confirmed.previousStatus).toBe("draft");
    expect(confirmed.currentStatus).toBe("confirmed");
    expect(confirmed.salesOrder.version).toBe(2);

    const cancelled = await cancelSalesOrder("so-260428-0001", "customer changed order", 2);

    expect(cancelled.previousStatus).toBe("confirmed");
    expect(cancelled.currentStatus).toBe("cancelled");
    expect(cancelled.salesOrder.cancelReason).toBe("customer changed order");
  });
});
