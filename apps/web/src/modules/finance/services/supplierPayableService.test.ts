import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  approveSupplierPayablePayment,
  getSupplierPayablePaymentReadiness,
  getSupplierPayables,
  recordSupplierPayablePayment,
  rejectSupplierPayablePayment,
  requestSupplierPayablePayment,
  resetPrototypeSupplierPayablesForTest,
  voidSupplierPayable
} from "./supplierPayableService";

describe("supplierPayableService", () => {
  beforeEach(() => {
    resetPrototypeSupplierPayablesForTest();
    vi.stubGlobal(
      "fetch",
      vi.fn(() => Promise.reject(new Error("offline")))
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("filters prototype supplier payables by status and search", async () => {
    const payables = await getSupplierPayables({ status: "open", search: "HCM" });

    expect(payables).toHaveLength(1);
    expect(payables[0]).toMatchObject({
      id: "ap-supplier-260430-0001",
      status: "open",
      supplierCode: "SUP-HCM-001"
    });
  });

  it("maps API list rows and sends AP query parameters", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          success: true,
          data: [
            {
              id: "ap-api-1",
              payable_no: "AP-API-1",
              supplier_id: "supplier-api",
              supplier_code: "SUP-API",
              supplier_name: "API Supplier",
              status: "open",
              total_amount: "900000.00",
              paid_amount: "0.00",
              outstanding_amount: "900000.00",
              currency_code: "VND",
              due_date: "2026-05-07",
              created_at: "2026-04-30T10:00:00Z",
              updated_at: "2026-04-30T10:05:00Z",
              version: 2
            }
          ],
          request_id: "req-ap-list"
        }),
        { status: 200 }
      )
    );
    vi.stubGlobal("fetch", fetchMock);

    const payables = await getSupplierPayables({ search: "AP-API", status: "open" });

    expect(fetchMock).toHaveBeenCalledWith("/api/v1/supplier-payables?q=AP-API&status=open", {
      headers: {
        Authorization: "Bearer local-dev-access-token"
      }
    });
    expect(payables[0]).toMatchObject({
      id: "ap-api-1",
      payableNo: "AP-API-1",
      lines: [],
      totalAmount: "900000.00",
      outstandingAmount: "900000.00"
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
              details: { permission: "payment:approve" },
              request_id: "req-denied"
            }
          }),
          { status: 403 }
        )
      )
    );

    await expect(getSupplierPayables()).rejects.toMatchObject({
      name: "ApiError",
      status: 403,
      code: "FORBIDDEN",
      requestId: "req-denied"
    });
  });

  it("requests, approves, and records supplier payment through prototype actions", async () => {
    const requested = await requestSupplierPayablePayment("ap-supplier-260430-0001");

    expect(requested.previousStatus).toBe("open");
    expect(requested.currentStatus).toBe("payment_requested");

    const approved = await approveSupplierPayablePayment("ap-supplier-260430-0001");

    expect(approved.previousStatus).toBe("payment_requested");
    expect(approved.currentStatus).toBe("payment_approved");

    const paid = await recordSupplierPayablePayment("ap-supplier-260430-0001", "1250000");

    expect(paid.previousStatus).toBe("payment_approved");
    expect(paid.currentStatus).toBe("partially_paid");
    expect(paid.supplierPayable).toMatchObject({
      paidAmount: "1250000.00",
      outstandingAmount: "3000000.00",
      auditLogId: "audit-ap-record-payment-ap-supplier-260430-0001"
    });
  });

  it("requires a matched supplier invoice before requesting AP payment", () => {
    const payable = {
      id: "ap-1",
      payableNo: "AP-1",
      supplierId: "supplier-1",
      supplierName: "Supplier 1",
      status: "open" as const,
      lines: [],
      totalAmount: "1000000.00",
      paidAmount: "0.00",
      outstandingAmount: "1000000.00",
      currencyCode: "VND",
      createdAt: "2026-05-05T10:00:00Z",
      updatedAt: "2026-05-05T10:00:00Z",
      version: 1
    };

    expect(getSupplierPayablePaymentReadiness(payable, [], false)).toMatchObject({
      canRequestPayment: false,
      tone: "warning"
    });
    expect(
      getSupplierPayablePaymentReadiness(
        payable,
        [
          {
            id: "si-1",
            invoiceNo: "INV-1",
            supplierId: "supplier-1",
            supplierName: "Supplier 1",
            payableId: "ap-1",
            payableNo: "AP-1",
            status: "mismatch",
            matchStatus: "mismatch",
            sourceDocument: { type: "warehouse_receipt", id: "gr-1", no: "GR-1" },
            lines: [],
            invoiceAmount: "990000.00",
            expectedAmount: "1000000.00",
            varianceAmount: "-10000.00",
            currencyCode: "VND",
            invoiceDate: "2026-05-05",
            createdAt: "2026-05-05T10:00:00Z",
            updatedAt: "2026-05-05T10:00:00Z",
            version: 1
          }
        ],
        false
      )
    ).toMatchObject({
      canRequestPayment: false,
      tone: "danger"
    });
    expect(
      getSupplierPayablePaymentReadiness(
        payable,
        [
          {
            id: "si-2",
            invoiceNo: "INV-2",
            supplierId: "supplier-1",
            supplierName: "Supplier 1",
            payableId: "ap-1",
            payableNo: "AP-1",
            status: "matched",
            matchStatus: "matched",
            sourceDocument: { type: "warehouse_receipt", id: "gr-1", no: "GR-1" },
            lines: [],
            invoiceAmount: "1000000.00",
            expectedAmount: "1000000.00",
            varianceAmount: "0.00",
            currencyCode: "VND",
            invoiceDate: "2026-05-05",
            createdAt: "2026-05-05T10:00:00Z",
            updatedAt: "2026-05-05T10:00:00Z",
            version: 1
          }
        ],
        false
      )
    ).toMatchObject({
      canRequestPayment: true,
      tone: "success"
    });
  });

  it("rejects requested supplier payment with a reason", async () => {
    await requestSupplierPayablePayment("ap-supplier-260430-0001");

    const rejected = await rejectSupplierPayablePayment("ap-supplier-260430-0001", "supplier invoice mismatch");

    expect(rejected.previousStatus).toBe("payment_requested");
    expect(rejected.currentStatus).toBe("open");
    expect(rejected.supplierPayable).toMatchObject({
      paymentRejectedBy: "finance-manager",
      paymentRejectReason: "supplier invoice mismatch"
    });
  });

  it("voids unpaid supplier payables with a reason", async () => {
    const voided = await voidSupplierPayable("ap-supplier-260430-0001", "duplicate supplier invoice");

    expect(voided.previousStatus).toBe("open");
    expect(voided.currentStatus).toBe("void");
    expect(voided.supplierPayable).toMatchObject({
      outstandingAmount: "0.00",
      voidReason: "duplicate supplier invoice"
    });
  });
});
