import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  createSupplierInvoice,
  getSupplierInvoices,
  resetPrototypeSupplierInvoicesForTest,
  voidSupplierInvoice
} from "./supplierInvoiceService";

describe("supplierInvoiceService", () => {
  beforeEach(() => {
    resetPrototypeSupplierInvoicesForTest();
    vi.stubGlobal(
      "fetch",
      vi.fn(() => Promise.reject(new Error("offline")))
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("filters prototype supplier invoices by payable and source document", async () => {
    const invoices = await getSupplierInvoices({ payableId: "ap-supplier-260430-0001", search: "PO-260430" });

    expect(invoices).toHaveLength(1);
    expect(invoices[0]).toMatchObject({
      id: "si-supplier-260430-0001",
      status: "matched",
      payableNo: "AP-SUP-260430-0001"
    });
  });

  it("maps API invoice rows and sends query parameters", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          success: true,
          data: [
            {
              id: "si-api-1",
              invoice_no: "INV-API-1",
              supplier_id: "supplier-api",
              supplier_code: "SUP-API",
              supplier_name: "API Supplier",
              payable_id: "ap-api-1",
              payable_no: "AP-API-1",
              status: "matched",
              match_status: "matched",
              source_document: { type: "warehouse_receipt", id: "gr-api-1", no: "GR-API-1" },
              invoice_amount: "900000.00",
              expected_amount: "900000.00",
              variance_amount: "0.00",
              currency_code: "VND",
              invoice_date: "2026-05-05",
              created_at: "2026-05-05T10:00:00Z",
              updated_at: "2026-05-05T10:00:00Z",
              version: 1
            }
          ],
          request_id: "req-si-list"
        }),
        { status: 200 }
      )
    );
    vi.stubGlobal("fetch", fetchMock);

    const invoices = await getSupplierInvoices({ payableId: "ap-api-1", search: "INV-API", status: "matched" });

    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/supplier-invoices?q=INV-API&status=matched&payable_id=ap-api-1",
      {
        headers: {
          Authorization: "Bearer local-dev-access-token"
        }
      }
    );
    expect(invoices[0]).toMatchObject({
      id: "si-api-1",
      invoiceNo: "INV-API-1",
      matchStatus: "matched",
      lines: []
    });
  });

  it("creates prototype mismatch invoices with variance", async () => {
    const invoice = await createSupplierInvoice({
      invoiceNo: "INV-MISMATCH",
      payableId: "ap-supplier-260430-0001",
      invoiceDate: "2026-05-05",
      invoiceAmount: "4200000",
      currencyCode: "VND"
    });

    expect(invoice).toMatchObject({
      invoiceNo: "INV-MISMATCH",
      status: "mismatch",
      varianceAmount: "-50000.00"
    });
  });

  it("voids prototype supplier invoices", async () => {
    const result = await voidSupplierInvoice("si-supplier-260430-0001", "duplicate invoice");

    expect(result.previousStatus).toBe("matched");
    expect(result.currentStatus).toBe("void");
    expect(result.supplierInvoice).toMatchObject({
      voidReason: "duplicate invoice",
      auditLogId: "audit-si-void-si-supplier-260430-0001"
    });
  });
});
