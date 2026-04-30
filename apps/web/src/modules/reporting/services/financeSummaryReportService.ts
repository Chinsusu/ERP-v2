import { ApiError, apiGetRaw } from "../../../shared/api/client";
import type {
  FinanceSummaryAgingBucket,
  FinanceSummaryCOD,
  FinanceSummaryCash,
  FinanceSummaryDiscrepancyBucket,
  FinanceSummaryPayable,
  FinanceSummaryQuery,
  FinanceSummaryReceivable,
  FinanceSummaryReport,
  ReportMetadata
} from "../types";

type FinanceSummaryReportApi = {
  metadata: {
    generated_at: string;
    timezone: string;
    source_version: string;
    filters: {
      from_date: string;
      to_date: string;
      business_date: string;
    };
  };
  currency_code: string;
  ar: FinanceSummaryReceivableApi;
  ap: FinanceSummaryPayableApi;
  cod: FinanceSummaryCODApi;
  cash: FinanceSummaryCashApi;
};

type FinanceSummaryReceivableApi = {
  open_count: number;
  overdue_count: number;
  disputed_count: number;
  open_amount: string;
  overdue_amount: string;
  outstanding_amount: string;
  aging_buckets: FinanceSummaryAgingBucketApi[];
};

type FinanceSummaryPayableApi = {
  open_count: number;
  due_count: number;
  payment_requested_count: number;
  payment_approved_count: number;
  open_amount: string;
  due_amount: string;
  outstanding_amount: string;
  aging_buckets: FinanceSummaryAgingBucketApi[];
};

type FinanceSummaryCODApi = {
  pending_count: number;
  discrepancy_count: number;
  pending_amount: string;
  discrepancy_amount: string;
  discrepancy_buckets: FinanceSummaryDiscrepancyBucketApi[];
};

type FinanceSummaryCashApi = {
  transaction_count: number;
  cash_in_amount: string;
  cash_out_amount: string;
  net_cash_amount: string;
};

type FinanceSummaryAgingBucketApi = {
  bucket: string;
  count: number;
  amount: string;
};

type FinanceSummaryDiscrepancyBucketApi = {
  type: string;
  status: string;
  count: number;
  amount: string;
};

const defaultAccessToken = "local-dev-access-token";
const zeroMoney = "0.00";

export async function getFinanceSummaryReport(query: FinanceSummaryQuery = {}): Promise<FinanceSummaryReport> {
  try {
    const report = await apiGetRaw<FinanceSummaryReportApi>(
      `/reports/finance-summary${financeSummaryQueryString(query)}`,
      { accessToken: defaultAccessToken }
    );

    return fromApiFinanceSummaryReport(report);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return createPrototypeFinanceSummaryReport(query);
  }
}

export function createPrototypeFinanceSummaryReport(query: FinanceSummaryQuery = {}): FinanceSummaryReport {
  const businessDate = query.businessDate || query.toDate || todayString();
  const fromDate = query.fromDate || businessDate;
  const toDate = query.toDate || businessDate;
  const includesPrototypeCash = includesDate({ fromDate, toDate }, "2026-04-30");

  return {
    metadata: {
      generatedAt: "2026-04-30T10:55:00Z",
      timezone: "Asia/Ho_Chi_Minh",
      sourceVersion: "reporting-v1",
      filters: {
        fromDate,
        toDate,
        businessDate
      }
    },
    currencyCode: "VND",
    ar: {
      openCount: 1,
      overdueCount: businessDate > "2026-05-03" ? 1 : 0,
      disputedCount: 0,
      openAmount: "1250000.00",
      overdueAmount: businessDate > "2026-05-03" ? "1250000.00" : zeroMoney,
      outstandingAmount: "1250000.00",
      agingBuckets: agingBuckets("2026-05-03", businessDate, "1250000.00")
    },
    ap: {
      openCount: 1,
      dueCount: businessDate >= "2026-05-07" ? 1 : 0,
      paymentRequestedCount: 0,
      paymentApprovedCount: 0,
      openAmount: "4250000.00",
      dueAmount: businessDate >= "2026-05-07" ? "4250000.00" : zeroMoney,
      outstandingAmount: "4250000.00",
      agingBuckets: agingBuckets("2026-05-07", businessDate, "4250000.00")
    },
    cod: {
      pendingCount: includesPrototypeCash ? 1 : 0,
      discrepancyCount: includesPrototypeCash ? 1 : 0,
      pendingAmount: includesPrototypeCash ? "2000000.00" : zeroMoney,
      discrepancyAmount: includesPrototypeCash ? "-50000.00" : zeroMoney,
      discrepancyBuckets: includesPrototypeCash
        ? [{ type: "short_paid", status: "open", count: 1, amount: "-50000.00" }]
        : []
    },
    cash: {
      transactionCount: includesPrototypeCash ? 2 : 0,
      cashInAmount: includesPrototypeCash ? "1250000.00" : zeroMoney,
      cashOutAmount: includesPrototypeCash ? "4250000.00" : zeroMoney,
      netCashAmount: includesPrototypeCash ? "-3000000.00" : zeroMoney
    }
  };
}

export function financeSummaryQueryString(query: FinanceSummaryQuery) {
  const params = new URLSearchParams();
  setQueryParam(params, "from_date", query.fromDate);
  setQueryParam(params, "to_date", query.toDate);
  setQueryParam(params, "business_date", query.businessDate);

  const value = params.toString();
  return value ? `?${value}` : "";
}

function fromApiFinanceSummaryReport(report: FinanceSummaryReportApi): FinanceSummaryReport {
  return {
    metadata: fromApiMetadata(report.metadata),
    currencyCode: report.currency_code,
    ar: fromApiReceivable(report.ar),
    ap: fromApiPayable(report.ap),
    cod: fromApiCOD(report.cod),
    cash: fromApiCash(report.cash)
  };
}

function fromApiMetadata(metadata: FinanceSummaryReportApi["metadata"]): ReportMetadata {
  return {
    generatedAt: metadata.generated_at,
    timezone: metadata.timezone,
    sourceVersion: metadata.source_version,
    filters: {
      fromDate: metadata.filters.from_date,
      toDate: metadata.filters.to_date,
      businessDate: metadata.filters.business_date
    }
  };
}

function fromApiReceivable(receivable: FinanceSummaryReceivableApi): FinanceSummaryReceivable {
  return {
    openCount: receivable.open_count,
    overdueCount: receivable.overdue_count,
    disputedCount: receivable.disputed_count,
    openAmount: receivable.open_amount,
    overdueAmount: receivable.overdue_amount,
    outstandingAmount: receivable.outstanding_amount,
    agingBuckets: receivable.aging_buckets.map(fromApiAgingBucket)
  };
}

function fromApiPayable(payable: FinanceSummaryPayableApi): FinanceSummaryPayable {
  return {
    openCount: payable.open_count,
    dueCount: payable.due_count,
    paymentRequestedCount: payable.payment_requested_count,
    paymentApprovedCount: payable.payment_approved_count,
    openAmount: payable.open_amount,
    dueAmount: payable.due_amount,
    outstandingAmount: payable.outstanding_amount,
    agingBuckets: payable.aging_buckets.map(fromApiAgingBucket)
  };
}

function fromApiCOD(cod: FinanceSummaryCODApi): FinanceSummaryCOD {
  return {
    pendingCount: cod.pending_count,
    discrepancyCount: cod.discrepancy_count,
    pendingAmount: cod.pending_amount,
    discrepancyAmount: cod.discrepancy_amount,
    discrepancyBuckets: cod.discrepancy_buckets.map(fromApiDiscrepancyBucket)
  };
}

function fromApiCash(cash: FinanceSummaryCashApi): FinanceSummaryCash {
  return {
    transactionCount: cash.transaction_count,
    cashInAmount: cash.cash_in_amount,
    cashOutAmount: cash.cash_out_amount,
    netCashAmount: cash.net_cash_amount
  };
}

function fromApiAgingBucket(bucket: FinanceSummaryAgingBucketApi): FinanceSummaryAgingBucket {
  return {
    bucket: bucket.bucket,
    count: bucket.count,
    amount: bucket.amount
  };
}

function fromApiDiscrepancyBucket(bucket: FinanceSummaryDiscrepancyBucketApi): FinanceSummaryDiscrepancyBucket {
  return {
    type: bucket.type,
    status: bucket.status,
    count: bucket.count,
    amount: bucket.amount
  };
}

function shouldUsePrototypeFallback(reason: unknown) {
  if (reason instanceof ApiError) {
    return false;
  }

  return !(reason instanceof Error && reason.message.startsWith("API request failed:"));
}

function agingBuckets(dueDate: string, businessDate: string, amount: string): FinanceSummaryAgingBucket[] {
  const rows = new Map<string, FinanceSummaryAgingBucket>();
  for (const bucket of ["current", "1_7", "8_30", "31_plus"]) {
    rows.set(bucket, { bucket, count: 0, amount: zeroMoney });
  }
  const bucket = agingBucket(dueDate, businessDate);
  rows.set(bucket, { bucket, count: 1, amount });

  return Array.from(rows.values());
}

function agingBucket(dueDate: string, businessDate: string) {
  const days = daysBetween(dueDate, businessDate);
  if (days <= 0) {
    return "current";
  }
  if (days <= 7) {
    return "1_7";
  }
  if (days <= 30) {
    return "8_30";
  }

  return "31_plus";
}

function daysBetween(left: string, right: string) {
  const leftDate = new Date(`${left}T00:00:00Z`).getTime();
  const rightDate = new Date(`${right}T00:00:00Z`).getTime();

  return Math.floor((rightDate - leftDate) / 86400000);
}

function includesDate(range: { fromDate: string; toDate: string }, date: string) {
  return date >= range.fromDate && date <= range.toDate;
}

function setQueryParam(params: URLSearchParams, key: string, value: string | undefined) {
  const normalized = value?.trim();
  if (normalized) {
    params.set(key, normalized);
  }
}

function todayString() {
  return new Date().toISOString().slice(0, 10);
}
