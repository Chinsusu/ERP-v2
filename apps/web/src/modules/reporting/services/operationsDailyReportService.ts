import { apiGet, apiGetBlob } from "../../../shared/api/client";
import { shouldUsePrototypeFallback } from "../../../shared/api/prototypeFallback";
import type { ApiGetQuery, ApiGetResponse } from "../../../shared/api/client";
import type {
  OperationsDailyAreaSummary,
  OperationsDailyQuery,
  OperationsDailyReport,
  OperationsDailyRow,
  OperationsDailyStatusFilter,
  OperationsDailySummary,
  ReportMetadata,
  ReportSourceReference
} from "../types";

type OperationsDailyReportApi = ApiGetResponse<"/reports/operations-daily">;
type OperationsDailyQueryApi = NonNullable<ApiGetQuery<"/reports/operations-daily">>;
type OperationsDailySummaryApi = OperationsDailyReportApi["summary"];
type OperationsDailyAreaSummaryApi = OperationsDailyReportApi["areas"][number];
type OperationsDailyRowApi = OperationsDailyReportApi["rows"][number];

const defaultAccessToken = "local-dev-access-token";

const prototypeRows: OperationsDailyRow[] = withSourceReferences([
  {
    id: "ops-inbound-hcm-260430-0001",
    area: "inbound",
    sourceType: "goods_receipt",
    sourceId: "gr-260430-0001",
    refNo: "GR-260430-0001",
    title: "Supplier delivery awaiting receiving check",
    warehouseId: "wh-hcm",
    warehouseCode: "HCM",
    businessDate: "2026-04-30",
    status: "pending",
    severity: "warning",
    quantity: "12.000000",
    uomCode: "PCS",
    owner: "warehouse"
  },
  {
    id: "ops-qc-hcm-260430-0001",
    area: "qc",
    sourceType: "inbound_qc",
    sourceId: "iqc-260430-fail",
    refNo: "IQC-260430-FAIL",
    title: "Inbound QC failed for damaged packaging",
    warehouseId: "wh-hcm",
    warehouseCode: "HCM",
    businessDate: "2026-04-30",
    status: "exception",
    severity: "danger",
    exceptionCode: "QC_FAIL",
    owner: "qa"
  },
  {
    id: "ops-outbound-hcm-260430-0001",
    area: "outbound",
    sourceType: "pick_task",
    sourceId: "pick-260430-0001",
    refNo: "PICK-260430-0001",
    title: "Pick wave in progress for ecommerce orders",
    warehouseId: "wh-hcm",
    warehouseCode: "HCM",
    businessDate: "2026-04-30",
    status: "in_progress",
    severity: "normal",
    quantity: "24.000000",
    uomCode: "PCS",
    owner: "warehouse"
  },
  {
    id: "ops-outbound-hcm-260430-0002",
    area: "outbound",
    sourceType: "carrier_manifest",
    sourceId: "manifest-260430-ghn",
    refNo: "MAN-260430-GHN",
    title: "Carrier handover missing scan",
    warehouseId: "wh-hcm",
    warehouseCode: "HCM",
    businessDate: "2026-04-30",
    status: "blocked",
    severity: "danger",
    exceptionCode: "MISSING_HANDOVER_SCAN",
    owner: "shipping"
  },
  {
    id: "ops-returns-hcm-260430-0001",
    area: "returns",
    sourceType: "return_receipt",
    sourceId: "ret-260430-0001",
    refNo: "RET-260430-0001",
    title: "Return receipt awaiting inspection",
    warehouseId: "wh-hcm",
    warehouseCode: "HCM",
    businessDate: "2026-04-30",
    status: "pending",
    severity: "warning",
    quantity: "3.000000",
    uomCode: "PCS",
    owner: "returns"
  },
  {
    id: "ops-stock-hcm-260430-0001",
    area: "stock_count",
    sourceType: "stock_count",
    sourceId: "count-260430-0001",
    refNo: "CNT-260430-0001",
    title: "Cycle count variance needs review",
    warehouseId: "wh-hcm",
    warehouseCode: "HCM",
    businessDate: "2026-04-30",
    status: "blocked",
    severity: "danger",
    exceptionCode: "VARIANCE_REVIEW",
    owner: "warehouse_lead"
  },
  {
    id: "ops-subcontract-hcm-260430-0001",
    area: "subcontract",
    sourceType: "subcontract_order",
    sourceId: "sco-260430-0001",
    refNo: "SCO-260430-0001",
    title: "Material issue in progress for subcontract factory",
    warehouseId: "wh-hcm",
    warehouseCode: "HCM",
    businessDate: "2026-04-30",
    status: "in_progress",
    severity: "normal",
    quantity: "80.000000",
    uomCode: "PCS",
    owner: "production"
  },
  {
    id: "ops-outbound-hn-260430-0001",
    area: "outbound",
    sourceType: "carrier_manifest",
    sourceId: "manifest-260430-hn",
    refNo: "MAN-260430-HN",
    title: "Hanoi carrier handover completed",
    warehouseId: "wh-hn",
    warehouseCode: "HN",
    businessDate: "2026-04-30",
    status: "completed",
    severity: "normal",
    owner: "shipping"
  }
]);

export async function getOperationsDailyReport(query: OperationsDailyQuery = {}): Promise<OperationsDailyReport> {
  try {
    const report = await apiGet("/reports/operations-daily", {
      accessToken: defaultAccessToken,
      query: operationsDailyApiQuery(query)
    });

    return fromApiOperationsDailyReport(report);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return createPrototypeOperationsDailyReport(query);
  }
}

export async function downloadOperationsDailyCSV(
  query: OperationsDailyQuery = {}
): Promise<{ blob: Blob; filename: string }> {
  const result = await apiGetBlob(`/reports/operations-daily/export.csv${operationsDailyQueryString(query)}`, {
    accessToken: defaultAccessToken
  });

  return {
    blob: result.blob,
    filename: result.filename ?? operationsDailyCSVFilename(query)
  };
}

export function createPrototypeOperationsDailyReport(query: OperationsDailyQuery = {}): OperationsDailyReport {
  const businessDate = query.businessDate || todayString();
  const fromDate = query.fromDate || businessDate;
  const toDate = query.toDate || businessDate;
  const rows = prototypeRows.filter((row) => matchesPrototypeFilter(row, { ...query, fromDate, toDate, businessDate }));

  return {
    metadata: {
      generatedAt: "2026-04-30T10:45:00Z",
      timezone: "Asia/Ho_Chi_Minh",
      sourceVersion: "reporting-v1",
      filters: {
        fromDate,
        toDate,
        businessDate,
        warehouseId: query.warehouseId,
        status: query.status
      }
    },
    summary: summarizeOperationsDailyRows(rows),
    areas: summarizeOperationsDailyAreas(rows),
    rows: rows.slice().sort(compareOperationsRows)
  };
}

export function operationsDailyQueryString(query: OperationsDailyQuery) {
  const params = new URLSearchParams();
  setQueryParam(params, "from_date", query.fromDate);
  setQueryParam(params, "to_date", query.toDate);
  setQueryParam(params, "business_date", query.businessDate);
  setQueryParam(params, "warehouse_id", query.warehouseId);
  setQueryParam(params, "status", query.status);

  const value = params.toString();
  return value ? `?${value}` : "";
}

function operationsDailyApiQuery(query: OperationsDailyQuery): OperationsDailyQueryApi {
  return {
    ...(optionalString(query.fromDate) ? { from_date: optionalString(query.fromDate) } : {}),
    ...(optionalString(query.toDate) ? { to_date: optionalString(query.toDate) } : {}),
    ...(optionalString(query.businessDate) ? { business_date: optionalString(query.businessDate) } : {}),
    ...(optionalString(query.warehouseId) ? { warehouse_id: optionalString(query.warehouseId) } : {}),
    ...(optionalString(query.status) ? { status: optionalString(query.status) } : {})
  };
}

export function operationsDailyCSVFilename(query: OperationsDailyQuery) {
  const toDate = query.toDate || query.businessDate || todayString();
  const fromDate = query.fromDate || toDate;
  return `operations-daily-${fromDate}-to-${toDate}.csv`;
}

function fromApiOperationsDailyReport(report: OperationsDailyReportApi): OperationsDailyReport {
  return {
    metadata: fromApiMetadata(report.metadata),
    summary: fromApiSummary(report.summary),
    areas: report.areas.map(fromApiAreaSummary),
    rows: report.rows.map(fromApiRow)
  };
}

function fromApiMetadata(metadata: OperationsDailyReportApi["metadata"]): ReportMetadata {
  return {
    generatedAt: metadata.generated_at,
    timezone: metadata.timezone,
    sourceVersion: metadata.source_version,
    filters: {
      fromDate: metadata.filters.from_date,
      toDate: metadata.filters.to_date,
      businessDate: metadata.filters.business_date,
      warehouseId: metadata.filters.warehouse_id,
      status: metadata.filters.status
    }
  };
}

function fromApiSummary(summary: OperationsDailySummaryApi): OperationsDailySummary {
  return {
    signalCount: summary.signal_count,
    pendingCount: summary.pending_count,
    inProgressCount: summary.in_progress_count,
    completedCount: summary.completed_count,
    blockedCount: summary.blocked_count,
    exceptionCount: summary.exception_count
  };
}

function fromApiAreaSummary(summary: OperationsDailyAreaSummaryApi): OperationsDailyAreaSummary {
  return {
    ...fromApiSummary(summary),
    area: summary.area
  };
}

function fromApiRow(row: OperationsDailyRowApi): OperationsDailyRow {
  return {
    id: row.id,
    area: row.area,
    sourceType: row.source_type,
    sourceId: row.source_id,
    sourceReference: fromApiSourceReference(row.source_reference, row),
    refNo: row.ref_no,
    title: row.title,
    warehouseId: row.warehouse_id,
    warehouseCode: row.warehouse_code,
    businessDate: row.business_date,
    status: row.status,
    severity: row.severity,
    quantity: row.quantity,
    uomCode: row.uom_code,
    exceptionCode: row.exception_code,
    owner: row.owner
  };
}

function fromApiSourceReference(
  reference: OperationsDailyRowApi["source_reference"] | undefined,
  row: Pick<OperationsDailyRowApi, "source_type" | "source_id" | "ref_no">
): ReportSourceReference {
  if (reference) {
    return {
      entityType: reference.entity_type,
      id: reference.id,
      label: reference.label,
      href: reference.href,
      unavailable: reference.unavailable
    };
  }

  return buildOperationsSourceReference(row.source_type, row.source_id, row.ref_no);
}

function withSourceReferences(rows: Array<Omit<OperationsDailyRow, "sourceReference">>): OperationsDailyRow[] {
  return rows.map((row) => ({
    ...row,
    sourceReference: buildOperationsSourceReference(row.sourceType, row.sourceId, row.refNo)
  }));
}

function buildOperationsSourceReference(sourceType: string, sourceId: string, label: string): ReportSourceReference {
  const module = operationsSourceModule(sourceType);
  if (!module) {
    return {
      entityType: sourceType,
      id: sourceId,
      label,
      unavailable: true
    };
  }

  const params = new URLSearchParams();
  params.set("source_type", sourceType);
  params.set("source_id", sourceId);

  return {
    entityType: sourceType,
    id: sourceId,
    label,
    href: `/${module}?${params.toString()}`,
    unavailable: false
  };
}

function operationsSourceModule(sourceType: string) {
  switch (sourceType) {
    case "goods_receipt":
      return "receiving";
    case "inbound_qc":
      return "qc";
    case "pick_task":
    case "carrier_manifest":
      return "shipping";
    case "return_receipt":
      return "returns";
    case "stock_count":
      return "inventory";
    case "subcontract_order":
      return "subcontract";
    default:
      return "";
  }
}

function summarizeOperationsDailyRows(rows: OperationsDailyRow[]): OperationsDailySummary {
  const summary = emptySummary();
  for (const row of rows) {
    addRowToSummary(summary, row);
  }

  return summary;
}

function summarizeOperationsDailyAreas(rows: OperationsDailyRow[]): OperationsDailyAreaSummary[] {
  const areas = new Map<string, OperationsDailyAreaSummary>();
  for (const row of rows) {
    const summary = areas.get(row.area) ?? { ...emptySummary(), area: row.area };
    addRowToSummary(summary, row);
    areas.set(row.area, summary);
  }

  return Array.from(areas.values()).sort((left, right) => areaOrder(left.area) - areaOrder(right.area));
}

function emptySummary(): OperationsDailySummary {
  return {
    signalCount: 0,
    pendingCount: 0,
    inProgressCount: 0,
    completedCount: 0,
    blockedCount: 0,
    exceptionCount: 0
  };
}

function addRowToSummary(summary: OperationsDailySummary, row: OperationsDailyRow) {
  summary.signalCount += 1;
  switch (row.status) {
    case "pending":
      summary.pendingCount += 1;
      break;
    case "in_progress":
      summary.inProgressCount += 1;
      break;
    case "completed":
      summary.completedCount += 1;
      break;
    case "blocked":
      summary.blockedCount += 1;
      break;
    case "exception":
      summary.exceptionCount += 1;
      break;
  }
}


function matchesPrototypeFilter(row: OperationsDailyRow, query: OperationsDailyQuery) {
  if (query.warehouseId && row.warehouseId !== query.warehouseId) {
    return false;
  }
  if (query.status && row.status !== query.status) {
    return false;
  }
  if (query.fromDate && row.businessDate < query.fromDate) {
    return false;
  }
  if (query.toDate && row.businessDate > query.toDate) {
    return false;
  }

  return true;
}

function compareOperationsRows(left: OperationsDailyRow, right: OperationsDailyRow) {
  if (left.businessDate !== right.businessDate) {
    return left.businessDate.localeCompare(right.businessDate);
  }
  if (areaOrder(left.area) !== areaOrder(right.area)) {
    return areaOrder(left.area) - areaOrder(right.area);
  }
  if (statusOrder(left.status) !== statusOrder(right.status)) {
    return statusOrder(left.status) - statusOrder(right.status);
  }

  return left.refNo.localeCompare(right.refNo);
}

function areaOrder(area: string) {
  switch (area) {
    case "inbound":
      return 1;
    case "qc":
      return 2;
    case "outbound":
      return 3;
    case "returns":
      return 4;
    case "stock_count":
      return 5;
    case "subcontract":
      return 6;
    default:
      return 99;
  }
}

function statusOrder(status: string) {
  switch (status) {
    case "blocked":
      return 1;
    case "exception":
      return 2;
    case "pending":
      return 3;
    case "in_progress":
      return 4;
    case "completed":
      return 5;
    default:
      return 99;
  }
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

export const operationsDailyStatusOptions: Array<{ label: string; value: OperationsDailyStatusFilter }> = [
  { label: "All statuses", value: "" },
  { label: "Pending", value: "pending" },
  { label: "In progress", value: "in_progress" },
  { label: "Completed", value: "completed" },
  { label: "Blocked", value: "blocked" },
  { label: "Exception", value: "exception" }
];
