import { ApiError, apiGet, apiGetBlob } from "../../../shared/api/client";
import type { ApiGetQuery, ApiGetResponse } from "../../../shared/api/client";
import type {
  FinanceSummaryAgingBucket,
  FinanceSummaryCOD,
  FinanceSummaryCash,
  FinanceSummaryDiscrepancyBucket,
  FinanceSummaryPayable,
  FinanceSummaryQuery,
  FinanceSummaryReceivable,
  FinanceSummaryReport,
  ReportMetadata,
  ReportSourceReference
} from "../types";

type FinanceSummaryReportApi = ApiGetResponse<"/reports/finance-summary">;
type FinanceSummaryQueryApi = NonNullable<ApiGetQuery<"/reports/finance-summary">>;
type FinanceSummaryReceivableApi = FinanceSummaryReportApi["ar"];
type FinanceSummaryPayableApi = FinanceSummaryReportApi["ap"];
type FinanceSummaryCODApi = FinanceSummaryReportApi["cod"];
type FinanceSummaryCashApi = FinanceSummaryReportApi["cash"];
type FinanceSummaryAgingBucketApi = FinanceSummaryReceivableApi["aging_buckets"][number];
type FinanceSummaryDiscrepancyBucketApi = FinanceSummaryCODApi["discrepancy_buckets"][number];

const defaultAccessToken = "local-dev-access-token";
const zeroMoney = "0.00";

export async function getFinanceSummaryReport(query: FinanceSummaryQuery = {}): Promise<FinanceSummaryReport> {
  try {
    const report = await apiGet("/reports/finance-summary", {
      accessToken: defaultAccessToken,
      query: financeSummaryApiQuery(query)
    });

    return fromApiFinanceSummaryReport(report);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return createPrototypeFinanceSummaryReport(query);
  }
}

export async function downloadFinanceSummaryCSV(
  query: FinanceSummaryQuery = {}
): Promise<{ blob: Blob; filename: string }> {
  const result = await apiGetBlob(`/reports/finance-summary/export.csv${financeSummaryQueryString(query)}`, {
    accessToken: defaultAccessToken
  });

  return {
    blob: result.blob,
    filename: result.filename ?? financeSummaryCSVFilename(query)
  };
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
      agingBuckets: agingBuckets("customer_receivable", "2026-05-03", businessDate, "1250000.00", { fromDate, toDate, businessDate }),
      sourceReferences: sectionReferences(["customer_receivable"], { fromDate, toDate, businessDate })
    },
    ap: {
      openCount: 1,
      dueCount: businessDate >= "2026-05-07" ? 1 : 0,
      paymentRequestedCount: 0,
      paymentApprovedCount: 0,
      openAmount: "4250000.00",
      dueAmount: businessDate >= "2026-05-07" ? "4250000.00" : zeroMoney,
      outstandingAmount: "4250000.00",
      agingBuckets: agingBuckets("supplier_payable", "2026-05-07", businessDate, "4250000.00", { fromDate, toDate, businessDate }),
      sourceReferences: sectionReferences(["supplier_payable", "payment_approval"], { fromDate, toDate, businessDate })
    },
    cod: {
      pendingCount: includesPrototypeCash ? 1 : 0,
      discrepancyCount: includesPrototypeCash ? 1 : 0,
      pendingAmount: includesPrototypeCash ? "2000000.00" : zeroMoney,
      discrepancyAmount: includesPrototypeCash ? "-50000.00" : zeroMoney,
      discrepancyBuckets: includesPrototypeCash
        ? [
            {
              type: "short_paid",
              status: "open",
              count: 1,
              amount: "-50000.00",
              sourceReference: financeSourceReference("cod_discrepancy", "short_paid:open", "short_paid:open", {
                fromDate,
                toDate,
                businessDate,
                type: "short_paid",
                status: "open"
              })
            }
          ]
        : [],
      sourceReferences: sectionReferences(["cod_remittance", "cod_discrepancy"], { fromDate, toDate, businessDate })
    },
    cash: {
      transactionCount: includesPrototypeCash ? 2 : 0,
      cashInAmount: includesPrototypeCash ? "1250000.00" : zeroMoney,
      cashOutAmount: includesPrototypeCash ? "4250000.00" : zeroMoney,
      netCashAmount: includesPrototypeCash ? "-3000000.00" : zeroMoney,
      sourceReferences: sectionReferences(["cash_transaction"], { fromDate, toDate, businessDate })
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

function financeSummaryApiQuery(query: FinanceSummaryQuery): FinanceSummaryQueryApi {
  return {
    ...(optionalString(query.fromDate) ? { from_date: optionalString(query.fromDate) } : {}),
    ...(optionalString(query.toDate) ? { to_date: optionalString(query.toDate) } : {}),
    ...(optionalString(query.businessDate) ? { business_date: optionalString(query.businessDate) } : {})
  };
}

function financeSummaryCSVFilename(query: FinanceSummaryQuery) {
  const toDate = query.toDate || query.businessDate || todayString();
  const fromDate = query.fromDate || toDate;
  return `finance-summary-${fromDate}-to-${toDate}.csv`;
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
    agingBuckets: receivable.aging_buckets.map(fromApiAgingBucket),
    sourceReferences: fromApiSourceReferences(receivable.source_references)
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
    agingBuckets: payable.aging_buckets.map(fromApiAgingBucket),
    sourceReferences: fromApiSourceReferences(payable.source_references)
  };
}

function fromApiCOD(cod: FinanceSummaryCODApi): FinanceSummaryCOD {
  return {
    pendingCount: cod.pending_count,
    discrepancyCount: cod.discrepancy_count,
    pendingAmount: cod.pending_amount,
    discrepancyAmount: cod.discrepancy_amount,
    discrepancyBuckets: cod.discrepancy_buckets.map(fromApiDiscrepancyBucket),
    sourceReferences: fromApiSourceReferences(cod.source_references)
  };
}

function fromApiCash(cash: FinanceSummaryCashApi): FinanceSummaryCash {
  return {
    transactionCount: cash.transaction_count,
    cashInAmount: cash.cash_in_amount,
    cashOutAmount: cash.cash_out_amount,
    netCashAmount: cash.net_cash_amount,
    sourceReferences: fromApiSourceReferences(cash.source_references)
  };
}

function fromApiAgingBucket(bucket: FinanceSummaryAgingBucketApi): FinanceSummaryAgingBucket {
  return {
    bucket: bucket.bucket,
    count: bucket.count,
    amount: bucket.amount,
    sourceReference: fromApiSourceReference(bucket.source_reference)
  };
}

function fromApiDiscrepancyBucket(bucket: FinanceSummaryDiscrepancyBucketApi): FinanceSummaryDiscrepancyBucket {
  return {
    type: bucket.type,
    status: bucket.status,
    count: bucket.count,
    amount: bucket.amount,
    sourceReference: fromApiSourceReference(bucket.source_reference)
  };
}

function fromApiSourceReferences(references: Array<FinanceSummaryAgingBucketApi["source_reference"]> | undefined) {
  return references?.map(fromApiSourceReference) ?? [];
}

function fromApiSourceReference(reference: FinanceSummaryAgingBucketApi["source_reference"]): ReportSourceReference {
  return {
    entityType: reference.entity_type,
    id: reference.id,
    label: reference.label,
    href: reference.href,
    unavailable: reference.unavailable
  };
}

function shouldUsePrototypeFallback(reason: unknown) {
  if (reason instanceof ApiError) {
    return false;
  }

  return !(reason instanceof Error && reason.message.startsWith("API request failed:"));
}

function agingBuckets(
  entityType: string,
  dueDate: string,
  businessDate: string,
  amount: string,
  filters: { fromDate: string; toDate: string; businessDate: string }
): FinanceSummaryAgingBucket[] {
  const rows = new Map<string, FinanceSummaryAgingBucket>();
  for (const bucket of ["current", "1_7", "8_30", "31_plus"]) {
    rows.set(bucket, {
      bucket,
      count: 0,
      amount: zeroMoney,
      sourceReference: financeSourceReference(entityType, bucket, bucket, { ...filters, bucket })
    });
  }
  const bucket = agingBucket(dueDate, businessDate);
  rows.set(bucket, {
    bucket,
    count: 1,
    amount,
    sourceReference: financeSourceReference(entityType, bucket, bucket, { ...filters, bucket })
  });

  return Array.from(rows.values());
}

function sectionReferences(
  entityTypes: string[],
  filters: { fromDate: string; toDate: string; businessDate: string }
): ReportSourceReference[] {
  return entityTypes.map((entityType) => financeSourceReference(entityType, entityType, entityType, filters));
}

function financeSourceReference(
  entityType: string,
  id: string,
  label: string,
  filters: { fromDate: string; toDate: string; businessDate: string; bucket?: string; type?: string; status?: string }
): ReportSourceReference {
  const params = new URLSearchParams();
  params.set("source_type", entityType);
  params.set("from_date", filters.fromDate);
  params.set("to_date", filters.toDate);
  params.set("business_date", filters.businessDate);
  if (filters.bucket) {
    params.set("bucket", filters.bucket);
  }
  if (filters.type) {
    params.set("type", filters.type);
  }
  if (filters.status) {
    params.set("status", filters.status);
  }

  return {
    entityType,
    id,
    label,
    href: `/finance?${params.toString()}`,
    unavailable: false
  };
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
  const normalized = optionalString(value);
  if (normalized) {
    params.set(key, normalized);
  }
}

function optionalString(value: string | undefined) {
  const normalized = value?.trim();
  return normalized || undefined;
}

function todayString() {
  return new Date().toISOString().slice(0, 10);
}
