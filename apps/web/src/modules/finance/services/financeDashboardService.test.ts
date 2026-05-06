import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { buildFinanceSummaryReportHref, getFinanceDashboard } from "./financeDashboardService";

describe("financeDashboardService", () => {
  beforeEach(() => {
    vi.stubGlobal(
      "fetch",
      vi.fn(() => Promise.reject(new Error("offline")))
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("returns prototype dashboard metrics when API is unavailable", async () => {
    const dashboard = await getFinanceDashboard({ businessDate: "2026-04-30" });

    expect(dashboard).toMatchObject({
      businessDate: "2026-04-30",
      ar: { openCount: 1, openAmount: "1250000.00" },
      ap: { openCount: 1, openAmount: "4250000.00" },
      cod: { pendingCount: 1, discrepancyAmount: "-50000.00" },
      cash: { transactionCount: 2, netCashToday: "-3000000.00" }
    });
  });

  it("builds finance summary report links with dashboard business date", () => {
    expect(buildFinanceSummaryReportHref({ businessDate: "2026-04-30" })).toBe(
      "/reports?report=finance&from_date=2026-04-30&to_date=2026-04-30&business_date=2026-04-30"
    );
    expect(buildFinanceSummaryReportHref()).toBe("/reports?report=finance");
  });

  it("maps API dashboard metrics and sends business date query", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          success: true,
          data: {
            business_date: "2026-05-08",
            generated_at: "2026-05-08T03:00:00Z",
            currency_code: "VND",
            ar: {
              open_count: 2,
              overdue_count: 1,
              disputed_count: 1,
              open_amount: "2000000.00",
              overdue_amount: "750000.00",
              outstanding_amount: "2500000.00"
            },
            ap: {
              open_count: 3,
              due_count: 1,
              payment_requested_count: 1,
              payment_approved_count: 1,
              open_amount: "5000000.00",
              due_amount: "1500000.00",
              outstanding_amount: "6000000.00"
            },
            cod: {
              pending_count: 1,
              discrepancy_count: 1,
              pending_amount: "2000000.00",
              discrepancy_amount: "-50000.00"
            },
            cash: {
              transaction_count: 2,
              cash_in_today: "1250000.00",
              cash_out_today: "4250000.00",
              net_cash_today: "-3000000.00"
            }
          },
          request_id: "req-dashboard"
        }),
        { status: 200 }
      )
    );
    vi.stubGlobal("fetch", fetchMock);

    const dashboard = await getFinanceDashboard({ businessDate: "2026-05-08" });

    expect(fetchMock).toHaveBeenCalledWith("/api/v1/finance/dashboard?business_date=2026-05-08", {
      headers: {
        Authorization: "Bearer local-dev-access-token"
      }
    });
    expect(dashboard).toMatchObject({
      businessDate: "2026-05-08",
      ar: { overdueCount: 1, overdueAmount: "750000.00" },
      cash: { netCashToday: "-3000000.00" }
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

    await expect(getFinanceDashboard()).rejects.toMatchObject({
      name: "ApiError",
      status: 403,
      code: "FORBIDDEN",
      requestId: "req-denied"
    });
  });
});
