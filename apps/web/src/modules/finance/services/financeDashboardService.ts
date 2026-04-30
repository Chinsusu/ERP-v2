import { ApiError, apiGet } from "../../../shared/api/client";
import type { components, operations } from "../../../shared/api/generated/schema";
import type { FinanceDashboard, FinanceDashboardQuery } from "../types";

type FinanceDashboardApi = components["schemas"]["FinanceDashboard"];
type FinanceDashboardApiQuery = operations["getFinanceDashboard"]["parameters"]["query"];

const defaultAccessToken = "local-dev-access-token";

export async function getFinanceDashboard(query: FinanceDashboardQuery = {}): Promise<FinanceDashboard> {
  try {
    const dashboard = await apiGet("/finance/dashboard", {
      accessToken: defaultAccessToken,
      query: toApiFinanceDashboardQuery(query)
    });

    return fromApiFinanceDashboard(dashboard);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return createPrototypeFinanceDashboard(query.businessDate ?? "2026-04-30");
  }
}

export function resetPrototypeFinanceDashboardForTest() {
  return createPrototypeFinanceDashboard("2026-04-30");
}

function shouldUsePrototypeFallback(reason: unknown) {
  if (reason instanceof ApiError) {
    return false;
  }

  return !(reason instanceof Error && reason.message.startsWith("API request failed:"));
}

function toApiFinanceDashboardQuery(query: FinanceDashboardQuery): FinanceDashboardApiQuery {
  return {
    business_date: query.businessDate
  };
}

function fromApiFinanceDashboard(dashboard: FinanceDashboardApi): FinanceDashboard {
  return {
    businessDate: dashboard.business_date,
    generatedAt: dashboard.generated_at,
    currencyCode: dashboard.currency_code,
    ar: {
      openCount: dashboard.ar.open_count,
      overdueCount: dashboard.ar.overdue_count,
      disputedCount: dashboard.ar.disputed_count,
      openAmount: dashboard.ar.open_amount,
      overdueAmount: dashboard.ar.overdue_amount,
      outstandingAmount: dashboard.ar.outstanding_amount
    },
    ap: {
      openCount: dashboard.ap.open_count,
      dueCount: dashboard.ap.due_count,
      paymentRequestedCount: dashboard.ap.payment_requested_count,
      paymentApprovedCount: dashboard.ap.payment_approved_count,
      openAmount: dashboard.ap.open_amount,
      dueAmount: dashboard.ap.due_amount,
      outstandingAmount: dashboard.ap.outstanding_amount
    },
    cod: {
      pendingCount: dashboard.cod.pending_count,
      discrepancyCount: dashboard.cod.discrepancy_count,
      pendingAmount: dashboard.cod.pending_amount,
      discrepancyAmount: dashboard.cod.discrepancy_amount
    },
    cash: {
      transactionCount: dashboard.cash.transaction_count,
      cashInToday: dashboard.cash.cash_in_today,
      cashOutToday: dashboard.cash.cash_out_today,
      netCashToday: dashboard.cash.net_cash_today
    }
  };
}

function createPrototypeFinanceDashboard(businessDate: string): FinanceDashboard {
  return {
    businessDate,
    generatedAt: "2026-04-30T10:30:00Z",
    currencyCode: "VND",
    ar: {
      openCount: 1,
      overdueCount: businessDate > "2026-05-03" ? 1 : 0,
      disputedCount: 0,
      openAmount: "1250000.00",
      overdueAmount: businessDate > "2026-05-03" ? "1250000.00" : "0.00",
      outstandingAmount: "1250000.00"
    },
    ap: {
      openCount: 1,
      dueCount: businessDate >= "2026-05-07" ? 1 : 0,
      paymentRequestedCount: 0,
      paymentApprovedCount: 0,
      openAmount: "4250000.00",
      dueAmount: businessDate >= "2026-05-07" ? "4250000.00" : "0.00",
      outstandingAmount: "4250000.00"
    },
    cod: {
      pendingCount: 1,
      discrepancyCount: 1,
      pendingAmount: "2000000.00",
      discrepancyAmount: "-50000.00"
    },
    cash: {
      transactionCount: businessDate === "2026-04-30" ? 2 : 0,
      cashInToday: businessDate === "2026-04-30" ? "1250000.00" : "0.00",
      cashOutToday: businessDate === "2026-04-30" ? "4250000.00" : "0.00",
      netCashToday: businessDate === "2026-04-30" ? "-3000000.00" : "0.00"
    }
  };
}
