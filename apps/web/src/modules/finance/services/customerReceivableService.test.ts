import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  getCustomerReceivables,
  markCustomerReceivableDisputed,
  recordCustomerReceivableReceipt,
  resetPrototypeCustomerReceivablesForTest,
  voidCustomerReceivable
} from "./customerReceivableService";

describe("customerReceivableService", () => {
  beforeEach(() => {
    resetPrototypeCustomerReceivablesForTest();
    vi.stubGlobal(
      "fetch",
      vi.fn(() => Promise.reject(new Error("offline")))
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("filters prototype receivables by status and search", async () => {
    const receivables = await getCustomerReceivables({ status: "partially_paid", search: "marketplace" });

    expect(receivables).toHaveLength(1);
    expect(receivables[0]).toMatchObject({
      id: "ar-cod-260430-0002",
      status: "partially_paid",
      customerCode: "KH-MKT-022"
    });
  });

  it("maps API list rows and sends finance query parameters", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          success: true,
          data: [
            {
              id: "ar-api-1",
              receivable_no: "AR-API-1",
              customer_id: "customer-api",
              customer_code: "KH-API",
              customer_name: "API Customer",
              status: "open",
              total_amount: "500000.00",
              paid_amount: "0.00",
              outstanding_amount: "500000.00",
              currency_code: "VND",
              due_date: "2026-05-01",
              created_at: "2026-04-30T08:00:00Z",
              updated_at: "2026-04-30T08:05:00Z",
              version: 2
            }
          ],
          request_id: "req-ar-list"
        }),
        { status: 200 }
      )
    );
    vi.stubGlobal("fetch", fetchMock);

    const receivables = await getCustomerReceivables({ search: "AR-API", status: "open" });

    expect(fetchMock).toHaveBeenCalledWith("http://localhost:8080/api/v1/customer-receivables?q=AR-API&status=open", {
      headers: {
        Authorization: "Bearer local-dev-access-token"
      }
    });
    expect(receivables[0]).toMatchObject({
      id: "ar-api-1",
      receivableNo: "AR-API-1",
      lines: [],
      totalAmount: "500000.00",
      outstandingAmount: "500000.00"
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
              details: { permission: "finance:view" },
              request_id: "req-denied"
            }
          }),
          { status: 403 }
        )
      )
    );

    await expect(getCustomerReceivables()).rejects.toMatchObject({
      name: "ApiError",
      status: 403,
      code: "FORBIDDEN",
      requestId: "req-denied"
    });
  });

  it("records receipts and updates paid/outstanding status", async () => {
    const result = await recordCustomerReceivableReceipt("ar-cod-260430-0001", "250000");

    expect(result.previousStatus).toBe("open");
    expect(result.currentStatus).toBe("partially_paid");
    expect(result.customerReceivable).toMatchObject({
      paidAmount: "250000.00",
      outstandingAmount: "1000000.00",
      auditLogId: "audit-record-receipt-ar-cod-260430-0001"
    });
  });

  it("marks disputed and void states through prototype actions", async () => {
    const disputed = await markCustomerReceivableDisputed("ar-cod-260430-0001", "carrier shortage");

    expect(disputed.currentStatus).toBe("disputed");
    expect(disputed.customerReceivable.disputeReason).toBe("carrier shortage");

    const voided = await voidCustomerReceivable("ar-cod-260430-0001", "duplicate AR");

    expect(voided.previousStatus).toBe("disputed");
    expect(voided.currentStatus).toBe("void");
    expect(voided.customerReceivable.outstandingAmount).toBe("0.00");
  });
});
