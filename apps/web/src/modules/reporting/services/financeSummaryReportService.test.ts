import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  createPrototypeFinanceSummaryReport,
  downloadFinanceSummaryCSV,
  financeSummaryQueryString,
  getFinanceSummaryReport
} from "./financeSummaryReportService";

describe("financeSummaryReportService", () => {
  beforeEach(() => {
    vi.stubGlobal(
      "fetch",
      vi.fn(() => Promise.reject(new Error("offline")))
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("returns prototype finance summary when API is unavailable", async () => {
    const report = await getFinanceSummaryReport({
      fromDate: "2026-04-30",
      toDate: "2026-05-08",
      businessDate: "2026-05-08"
    });

    expect(report.metadata.filters).toMatchObject({
      fromDate: "2026-04-30",
      toDate: "2026-05-08",
      businessDate: "2026-05-08"
    });
    expect(report.currencyCode).toBe("VND");
    expect(report.ar).toMatchObject({
      openCount: 1,
      overdueCount: 1,
      overdueAmount: "1250000.00"
    });
    expect(report.ap).toMatchObject({
      openCount: 1,
      dueCount: 1,
      dueAmount: "4250000.00"
    });
    expect(report.cod.discrepancyBuckets[0]).toMatchObject({
      type: "short_paid",
      status: "open",
      amount: "-50000.00"
    });
    expect(report.cash.netCashAmount).toBe("-3000000.00");
  });

  it("maps API report and sends finance summary filters", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          success: true,
          data: {
            metadata: {
              generated_at: "2026-04-30T07:33:13Z",
              timezone: "Asia/Ho_Chi_Minh",
              source_version: "reporting-v1",
              filters: {
                from_date: "2026-04-30",
                to_date: "2026-05-08",
                business_date: "2026-05-08"
              }
            },
            currency_code: "VND",
            ar: {
              open_count: 1,
              overdue_count: 1,
              disputed_count: 0,
              open_amount: "1250000.00",
              overdue_amount: "1250000.00",
              outstanding_amount: "1250000.00",
              aging_buckets: [{ bucket: "1_7", count: 1, amount: "1250000.00" }]
            },
            ap: {
              open_count: 1,
              due_count: 1,
              payment_requested_count: 0,
              payment_approved_count: 0,
              open_amount: "4250000.00",
              due_amount: "4250000.00",
              outstanding_amount: "4250000.00",
              aging_buckets: [{ bucket: "1_7", count: 1, amount: "4250000.00" }]
            },
            cod: {
              pending_count: 1,
              discrepancy_count: 1,
              pending_amount: "2000000.00",
              discrepancy_amount: "-50000.00",
              discrepancy_buckets: []
            },
            cash: {
              transaction_count: 2,
              cash_in_amount: "1250000.00",
              cash_out_amount: "4250000.00",
              net_cash_amount: "-3000000.00"
            }
          },
          request_id: "req-report-finance"
        }),
        { status: 200 }
      )
    );
    vi.stubGlobal("fetch", fetchMock);

    const report = await getFinanceSummaryReport({
      fromDate: "2026-04-30",
      toDate: "2026-05-08",
      businessDate: "2026-05-08"
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/reports/finance-summary?from_date=2026-04-30&to_date=2026-05-08&business_date=2026-05-08",
      {
        headers: {
          Authorization: "Bearer local-dev-access-token"
        }
      }
    );
    expect(report.ar.agingBuckets[0]).toMatchObject({ bucket: "1_7", count: 1 });
    expect(report.cash.transactionCount).toBe(2);
  });

  it("downloads finance CSV with current filters", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response("section,metric,count,amount,currency_code\nar,open,1,1250000.00,VND\n", {
        status: 200,
        headers: {
          "Content-Disposition": `attachment; filename="finance-summary-2026-04-30-to-2026-05-08.csv"`,
          "Content-Type": "text/csv; charset=utf-8"
        }
      })
    );
    vi.stubGlobal("fetch", fetchMock);

    const download = await downloadFinanceSummaryCSV({
      fromDate: "2026-04-30",
      toDate: "2026-05-08",
      businessDate: "2026-05-08"
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/reports/finance-summary/export.csv?from_date=2026-04-30&to_date=2026-05-08&business_date=2026-05-08",
      {
        headers: {
          Authorization: "Bearer local-dev-access-token"
        }
      }
    );
    await expect(download.blob.text()).resolves.toContain("1250000.00");
    expect(download.filename).toBe("finance-summary-2026-04-30-to-2026-05-08.csv");
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
              details: { permission: "reports:finance:view" },
              request_id: "req-denied"
            }
          }),
          { status: 403 }
        )
      )
    );

    await expect(getFinanceSummaryReport()).rejects.toMatchObject({
      name: "ApiError",
      status: 403,
      code: "FORBIDDEN",
      requestId: "req-denied"
    });
  });

  it("builds stable query strings without blank filters", () => {
    expect(
      financeSummaryQueryString({
        fromDate: "",
        toDate: "2026-05-08",
        businessDate: "2026-05-08"
      })
    ).toBe("?to_date=2026-05-08&business_date=2026-05-08");
  });

  it("keeps out-of-range COD and cash at zero in prototype reports", () => {
    const report = createPrototypeFinanceSummaryReport({
      fromDate: "2026-05-01",
      toDate: "2026-05-08",
      businessDate: "2026-05-08"
    });

    expect(report.cod.pendingCount).toBe(0);
    expect(report.cash.transactionCount).toBe(0);
    expect(report.cash.netCashAmount).toBe("0.00");
  });
});
