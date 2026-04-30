import { ApiError, apiGet, apiGetBlob } from "../../../shared/api/client";
import type { ApiGetQuery, ApiGetResponse } from "../../../shared/api/client";
import type {
  InventorySnapshotQuery,
  InventorySnapshotReport,
  InventorySnapshotRow,
  InventorySnapshotStatusFilter,
  InventorySnapshotSummary,
  InventorySnapshotUOMTotal,
  ReportMetadata
} from "../types";

type InventorySnapshotReportApi = ApiGetResponse<"/reports/inventory-snapshot">;
type InventorySnapshotQueryApi = NonNullable<ApiGetQuery<"/reports/inventory-snapshot">>;
type InventorySnapshotUOMTotalApi = InventorySnapshotReportApi["summary"]["totals_by_uom"][number];
type InventorySnapshotRowApi = InventorySnapshotReportApi["rows"][number];

const defaultAccessToken = "local-dev-access-token";
const quantityScale = 6;
const zeroQuantity = "0.000000";

const prototypeRows: InventorySnapshotRow[] = [
  {
    warehouseId: "wh-hcm",
    warehouseCode: "HCM",
    locationId: "bin-hcm-a01",
    locationCode: "A-01",
    itemId: "item-serum-30ml",
    sku: "SERUM-30ML",
    batchId: "batch-serum-2604a",
    batchNo: "LOT-2604A",
    batchExpiry: "2026-05-20",
    baseUomCode: "PCS",
    physicalQty: "128.000000",
    reservedQty: "10.000000",
    quarantineQty: "8.000000",
    blockedQty: "0.000000",
    availableQty: "110.000000",
    lowStock: false,
    expiryWarning: true,
    expired: false,
    batchQcStatus: "pass",
    batchStatus: "active",
    sourceStockState: "quarantine"
  },
  {
    warehouseId: "wh-hcm",
    warehouseCode: "HCM",
    locationId: "bin-hcm-a01",
    locationCode: "A-01",
    itemId: "item-cream-50g",
    sku: "CREAM-50G",
    batchId: "batch-cream-2603b",
    batchNo: "LOT-2603B",
    baseUomCode: "PCS",
    physicalQty: "46.000000",
    reservedQty: "12.000000",
    quarantineQty: "0.000000",
    blockedQty: "2.000000",
    availableQty: "32.000000",
    lowStock: false,
    expiryWarning: false,
    expired: false,
    batchQcStatus: "pass",
    batchStatus: "active",
    sourceStockState: "blocked"
  },
  {
    warehouseId: "wh-hn",
    warehouseCode: "HN",
    locationId: "bin-hn-r01",
    locationCode: "R-01",
    itemId: "item-toner-100ml",
    sku: "TONER-100ML",
    batchId: "batch-toner-2604c",
    batchNo: "LOT-2604C",
    baseUomCode: "PCS",
    physicalQty: "90.000000",
    reservedQty: "20.000000",
    quarantineQty: "0.000000",
    blockedQty: "5.000000",
    availableQty: "65.000000",
    lowStock: false,
    expiryWarning: false,
    expired: false,
    batchQcStatus: "pass",
    batchStatus: "active",
    sourceStockState: "blocked"
  }
];

export async function getInventorySnapshotReport(
  query: InventorySnapshotQuery = {}
): Promise<InventorySnapshotReport> {
  try {
    const report = await apiGet("/reports/inventory-snapshot", {
      accessToken: defaultAccessToken,
      query: inventorySnapshotApiQuery(query)
    });

    return fromApiInventorySnapshotReport(report);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return createPrototypeInventorySnapshotReport(query);
  }
}

export async function downloadInventorySnapshotCSV(
  query: InventorySnapshotQuery = {}
): Promise<{ blob: Blob; filename: string }> {
  const result = await apiGetBlob(`/reports/inventory-snapshot/export.csv${inventorySnapshotQueryString(query)}`, {
    accessToken: defaultAccessToken
  });

  return {
    blob: result.blob,
    filename: result.filename ?? inventorySnapshotCSVFilename(query)
  };
}

export function createPrototypeInventorySnapshotReport(
  query: InventorySnapshotQuery = {}
): InventorySnapshotReport {
  const businessDate = query.businessDate || todayString();
  const rows = prototypeRows.filter((row) => matchesPrototypeFilter(row, query));

  return {
    metadata: {
      generatedAt: "2026-04-30T10:30:00Z",
      timezone: "Asia/Ho_Chi_Minh",
      sourceVersion: "reporting-v1",
      filters: {
        fromDate: query.fromDate || businessDate,
        toDate: query.toDate || businessDate,
        businessDate,
        warehouseId: query.warehouseId,
        status: query.status,
        itemId: query.itemId,
        category: query.category
      }
    },
    summary: summarizeInventorySnapshotRows(rows),
    rows
  };
}

export function inventorySnapshotQueryString(query: InventorySnapshotQuery) {
  const params = new URLSearchParams();
  setQueryParam(params, "from_date", query.fromDate);
  setQueryParam(params, "to_date", query.toDate);
  setQueryParam(params, "business_date", query.businessDate);
  setQueryParam(params, "warehouse_id", query.warehouseId);
  setQueryParam(params, "status", query.status);
  setQueryParam(params, "item_id", query.itemId);
  setQueryParam(params, "category", query.category);
  setQueryParam(params, "location_id", query.locationId);
  setQueryParam(params, "sku", query.sku);
  setQueryParam(params, "batch_id", query.batchId);
  setQueryParam(params, "low_stock_threshold", query.lowStockThreshold);
  setQueryParam(params, "expiry_warning_days", query.expiryWarningDays);

  const value = params.toString();
  return value ? `?${value}` : "";
}

function inventorySnapshotApiQuery(query: InventorySnapshotQuery): InventorySnapshotQueryApi {
  const expiryWarningDays = optionalNumber(query.expiryWarningDays);

  return {
    ...(optionalString(query.fromDate) ? { from_date: optionalString(query.fromDate) } : {}),
    ...(optionalString(query.toDate) ? { to_date: optionalString(query.toDate) } : {}),
    ...(optionalString(query.businessDate) ? { business_date: optionalString(query.businessDate) } : {}),
    ...(optionalString(query.warehouseId) ? { warehouse_id: optionalString(query.warehouseId) } : {}),
    ...(optionalString(query.status) ? { status: optionalString(query.status) } : {}),
    ...(optionalString(query.itemId) ? { item_id: optionalString(query.itemId) } : {}),
    ...(optionalString(query.category) ? { category: optionalString(query.category) } : {}),
    ...(optionalString(query.locationId) ? { location_id: optionalString(query.locationId) } : {}),
    ...(optionalString(query.sku) ? { sku: optionalString(query.sku) } : {}),
    ...(optionalString(query.batchId) ? { batch_id: optionalString(query.batchId) } : {}),
    ...(optionalString(query.lowStockThreshold) ? { low_stock_threshold: optionalString(query.lowStockThreshold) } : {}),
    ...(expiryWarningDays !== undefined ? { expiry_warning_days: expiryWarningDays } : {})
  };
}

function inventorySnapshotCSVFilename(query: InventorySnapshotQuery) {
  return `inventory-snapshot-${query.businessDate || todayString()}.csv`;
}

function fromApiInventorySnapshotReport(report: InventorySnapshotReportApi): InventorySnapshotReport {
  return {
    metadata: fromApiMetadata(report.metadata),
    summary: {
      rowCount: report.summary.row_count,
      lowStockRowCount: report.summary.low_stock_row_count,
      expiryWarningRows: report.summary.expiry_warning_rows,
      expiredRows: report.summary.expired_rows,
      totalsByUom: report.summary.totals_by_uom.map(fromApiTotal)
    },
    rows: report.rows.map(fromApiRow)
  };
}

function fromApiMetadata(metadata: InventorySnapshotReportApi["metadata"]): ReportMetadata {
  return {
    generatedAt: metadata.generated_at,
    timezone: metadata.timezone,
    sourceVersion: metadata.source_version,
    filters: {
      fromDate: metadata.filters.from_date,
      toDate: metadata.filters.to_date,
      businessDate: metadata.filters.business_date,
      warehouseId: metadata.filters.warehouse_id,
      status: metadata.filters.status,
      itemId: metadata.filters.item_id,
      category: metadata.filters.category
    }
  };
}

function fromApiTotal(total: InventorySnapshotUOMTotalApi): InventorySnapshotUOMTotal {
  return {
    baseUomCode: total.base_uom_code,
    physicalQty: total.physical_qty,
    reservedQty: total.reserved_qty,
    quarantineQty: total.quarantine_qty,
    blockedQty: total.blocked_qty,
    availableQty: total.available_qty
  };
}

function fromApiRow(row: InventorySnapshotRowApi): InventorySnapshotRow {
  return {
    warehouseId: row.warehouse_id,
    warehouseCode: row.warehouse_code,
    locationId: row.location_id,
    locationCode: row.location_code,
    itemId: row.item_id,
    sku: row.sku,
    batchId: row.batch_id,
    batchNo: row.batch_no,
    batchExpiry: row.batch_expiry,
    baseUomCode: row.base_uom_code,
    physicalQty: row.physical_qty,
    reservedQty: row.reserved_qty,
    quarantineQty: row.quarantine_qty,
    blockedQty: row.blocked_qty,
    availableQty: row.available_qty,
    lowStock: row.low_stock,
    expiryWarning: row.expiry_warning,
    expired: row.expired,
    batchQcStatus: row.batch_qc_status,
    batchStatus: row.batch_status,
    sourceStockState: row.source_stock_state
  };
}

function summarizeInventorySnapshotRows(rows: InventorySnapshotRow[]): InventorySnapshotSummary {
  const totalsByUom = new Map<string, InventorySnapshotUOMTotal>();
  for (const row of rows) {
    const current = totalsByUom.get(row.baseUomCode) ?? {
      baseUomCode: row.baseUomCode,
      physicalQty: zeroQuantity,
      reservedQty: zeroQuantity,
      quarantineQty: zeroQuantity,
      blockedQty: zeroQuantity,
      availableQty: zeroQuantity
    };
    totalsByUom.set(row.baseUomCode, {
      ...current,
      physicalQty: addQuantity(current.physicalQty, row.physicalQty),
      reservedQty: addQuantity(current.reservedQty, row.reservedQty),
      quarantineQty: addQuantity(current.quarantineQty, row.quarantineQty),
      blockedQty: addQuantity(current.blockedQty, row.blockedQty),
      availableQty: addQuantity(current.availableQty, row.availableQty)
    });
  }

  return {
    rowCount: rows.length,
    lowStockRowCount: rows.filter((row) => row.lowStock).length,
    expiryWarningRows: rows.filter((row) => row.expiryWarning).length,
    expiredRows: rows.filter((row) => row.expired).length,
    totalsByUom: Array.from(totalsByUom.values()).sort((left, right) =>
      left.baseUomCode.localeCompare(right.baseUomCode)
    )
  };
}

function shouldUsePrototypeFallback(reason: unknown) {
  if (reason instanceof ApiError) {
    return false;
  }

  return !(reason instanceof Error && reason.message.startsWith("API request failed:"));
}

function matchesPrototypeFilter(row: InventorySnapshotRow, query: InventorySnapshotQuery) {
  if (query.warehouseId && row.warehouseId !== query.warehouseId) {
    return false;
  }
  if (query.locationId && row.locationId !== query.locationId) {
    return false;
  }
  if (query.itemId && row.itemId !== query.itemId) {
    return false;
  }
  if (query.sku && row.sku !== query.sku.trim().toUpperCase()) {
    return false;
  }
  if (query.batchId && row.batchId !== query.batchId) {
    return false;
  }
  if (query.status && row.sourceStockState !== query.status) {
    return false;
  }

  return true;
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

function optionalNumber(value: string | undefined) {
  const normalized = optionalString(value);
  if (!normalized) {
    return undefined;
  }

  const parsed = Number(normalized);
  return parsed;
}

function todayString() {
  return new Date().toISOString().slice(0, 10);
}

function addQuantity(left: string, right: string) {
  return scaledToQuantity(quantityToScaled(left) + quantityToScaled(right));
}

function quantityToScaled(value: string) {
  const normalized = normalizeQuantity(value);
  const negative = normalized.startsWith("-");
  const unsigned = negative ? normalized.slice(1) : normalized;
  const [integerPart, fractionalPart = ""] = unsigned.split(".");
  const scaled = `${integerPart}${fractionalPart.padEnd(quantityScale, "0").slice(0, quantityScale)}`;
  const parsed = BigInt(scaled);

  return negative ? -parsed : parsed;
}

function scaledToQuantity(value: bigint) {
  const negative = value < BigInt(0);
  const unsigned = negative ? -value : value;
  const digits = unsigned.toString().padStart(quantityScale + 1, "0");
  const integerPart = digits.slice(0, -quantityScale);
  const fractionalPart = digits.slice(-quantityScale);

  return `${negative ? "-" : ""}${integerPart}.${fractionalPart}`;
}

function normalizeQuantity(value: string) {
  const raw = value.trim();
  if (!/^-?\d+(\.\d{1,6})?$/.test(raw)) {
    return zeroQuantity;
  }

  const [integerPart, fractionalPart = ""] = raw.split(".");
  return `${integerPart}.${fractionalPart.padEnd(quantityScale, "0").slice(0, quantityScale)}`;
}

export const inventorySnapshotStatusOptions: Array<{ label: string; value: InventorySnapshotStatusFilter }> = [
  { label: "All states", value: "" },
  { label: "Available", value: "available" },
  { label: "Reserved", value: "reserved" },
  { label: "Quarantine", value: "quarantine" },
  { label: "Blocked", value: "blocked" }
];
