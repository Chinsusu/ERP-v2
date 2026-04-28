import { ApiError, apiGet, apiGetRaw, apiPost } from "../../../shared/api/client";
import type { components, operations } from "../../../shared/api/generated/schema";
import type {
  AvailableStockItem,
  AvailableStockQuery,
  AvailableStockSummary,
  BatchQCTransition,
  BatchQCTransitionInput,
  BatchQCTransitionResult,
  CreateStockCountInput,
  StockCountSession,
  StockCountStatus,
  SubmitStockCountInput
} from "../types";

type AvailableStockApiItem = components["schemas"]["AvailableStockItem"];
type AvailableStockApiQuery = operations["listAvailableStock"]["parameters"]["query"];
type BatchQCTransitionApiItem = components["schemas"]["BatchQCTransition"];
type BatchQCTransitionResultApi = components["schemas"]["BatchQCTransitionResult"];
type StockCountApiLine = {
  id: string;
  item_id?: string;
  sku: string;
  batch_id?: string;
  batch_no?: string;
  location_id?: string;
  location_code?: string;
  expected_qty: string;
  counted_qty: string;
  delta_qty: string;
  base_uom_code: string;
  counted: boolean;
  note?: string;
};
type StockCountApiSession = {
  id: string;
  count_no: string;
  org_id: string;
  warehouse_id: string;
  warehouse_code?: string;
  scope: string;
  status: StockCountStatus;
  created_by: string;
  submitted_by?: string;
  lines: StockCountApiLine[];
  audit_log_id?: string;
  created_at: string;
  updated_at: string;
  submitted_at?: string;
};
type CreateStockCountApiRequest = {
  count_no?: string;
  warehouse_id: string;
  warehouse_code?: string;
  scope: string;
  lines: Array<{
    id?: string;
    item_id?: string;
    sku: string;
    batch_id?: string;
    batch_no?: string;
    location_id?: string;
    location_code?: string;
    expected_qty: string;
    base_uom_code: string;
  }>;
};
type SubmitStockCountApiRequest = {
  lines: Array<{
    id: string;
    counted_qty: string;
    note?: string;
  }>;
};

const defaultAccessToken = "local-dev-access-token";
const quantityScale = 6;
const zeroQuantity = "0.000000";

export const prototypeAvailableStock: AvailableStockItem[] = [
  {
    warehouseId: "wh-hcm",
    warehouseCode: "HCM",
    locationId: "bin-hcm-a01",
    locationCode: "A-01",
    sku: "SERUM-30ML",
    batchId: "batch-serum-2604a",
    batchNo: "LOT-2604A",
    batchQcStatus: "hold",
    batchStatus: "active",
    batchExpiryDate: "2027-04-01",
    baseUomCode: "PCS",
    physicalQty: "128.000000",
    reservedQty: "10.000000",
    qcHoldQty: "8.000000",
    damagedQty: "0.000000",
    returnPendingQty: "0.000000",
    blockedQty: "0.000000",
    holdQty: "8.000000",
    availableQty: "110.000000"
  },
  {
    warehouseId: "wh-hcm",
    warehouseCode: "HCM",
    locationId: "bin-hcm-a01",
    locationCode: "A-01",
    sku: "CREAM-50G",
    batchId: "batch-cream-2603b",
    batchNo: "LOT-2603B",
    batchQcStatus: "pass",
    batchStatus: "active",
    batchExpiryDate: "2028-03-01",
    baseUomCode: "PCS",
    physicalQty: "46.000000",
    reservedQty: "12.000000",
    qcHoldQty: "0.000000",
    damagedQty: "2.000000",
    returnPendingQty: "0.000000",
    blockedQty: "2.000000",
    holdQty: "2.000000",
    availableQty: "32.000000"
  },
  {
    warehouseId: "wh-hn",
    warehouseCode: "HN",
    locationId: "bin-hn-r01",
    locationCode: "R-01",
    sku: "TONER-100ML",
    batchId: "batch-toner-2604c",
    batchNo: "LOT-2604C",
    batchQcStatus: "fail",
    batchStatus: "blocked",
    batchExpiryDate: "2027-10-10",
    baseUomCode: "PCS",
    physicalQty: "90.000000",
    reservedQty: "20.000000",
    qcHoldQty: "0.000000",
    damagedQty: "0.000000",
    returnPendingQty: "5.000000",
    blockedQty: "5.000000",
    holdQty: "5.000000",
    availableQty: "65.000000"
  }
];

export const prototypeBatchQCTransitions: BatchQCTransition[] = [
  {
    id: "audit-batch-qc-260426-0004",
    batchId: "batch-cream-2603b",
    batchNo: "LOT-2603B",
    sku: "CREAM-50G",
    fromQcStatus: "hold",
    toQcStatus: "pass",
    actorId: "user-qa",
    reason: "incoming inspection passed",
    businessRef: "QC-260426-0004",
    auditLogId: "audit-batch-qc-260426-0004",
    createdAt: "2026-04-26T07:40:00Z"
  }
];

const initialPrototypeStockCounts: StockCountSession[] = [
  {
    id: "count-hcm-260427-0001",
    countNo: "CNT-260427-0001",
    orgId: "org-my-pham",
    warehouseId: "wh-hcm",
    warehouseCode: "HCM",
    scope: "cycle-count",
    status: "open",
    createdBy: "user-warehouse",
    lines: [
      {
        id: "count-line-hcm-0001",
        sku: "SERUM-30ML",
        batchId: "batch-serum-2604a",
        batchNo: "LOT-2604A",
        locationId: "bin-hcm-a01",
        locationCode: "A-01",
        expectedQty: "128.000000",
        countedQty: zeroQuantity,
        deltaQty: zeroQuantity,
        baseUomCode: "PCS",
        counted: false
      }
    ],
    createdAt: "2026-04-27T08:15:00Z",
    updatedAt: "2026-04-27T08:15:00Z"
  }
];

let prototypeStockCounts = initialPrototypeStockCounts.map(cloneStockCount);

export async function getAvailableStock(query: AvailableStockQuery = {}): Promise<AvailableStockItem[]> {
  try {
    const items = await apiGet("/inventory/available-stock", {
      accessToken: defaultAccessToken,
      query: toApiQuery(query)
    });

    return items.map(fromApiItem);
  } catch {
    return filterPrototypeStock(query);
  }
}

export async function getBatchQCTransitions(batchId: string): Promise<BatchQCTransition[]> {
  try {
    const items = await apiGetRaw<BatchQCTransitionApiItem[]>(
      `/inventory/batches/${encodeURIComponent(batchId)}/qc-transitions`,
      { accessToken: defaultAccessToken }
    );

    return items.map(fromApiTransition);
  } catch {
    return prototypeBatchQCTransitions.filter((transition) => transition.batchId === batchId);
  }
}

export async function createBatchQCTransition(
  batchId: string,
  input: BatchQCTransitionInput
): Promise<BatchQCTransitionResult> {
  const result = await apiPost<BatchQCTransitionResultApi, components["schemas"]["CreateBatchQCTransitionRequest"]>(
    `/inventory/batches/${encodeURIComponent(batchId)}/qc-transitions`,
    {
      qc_status: input.qcStatus,
      reason: input.reason,
      business_ref: input.businessRef
    },
    { accessToken: defaultAccessToken }
  );

  return {
    transition: fromApiTransition(result.transition)
  };
}

export async function getStockCounts(): Promise<StockCountSession[]> {
  try {
    const items = await apiGetRaw<StockCountApiSession[]>("/stock-counts", {
      accessToken: defaultAccessToken
    });

    return items.map(fromApiStockCount);
  } catch {
    return prototypeStockCounts.map(cloneStockCount);
  }
}

export async function createStockCount(input: CreateStockCountInput): Promise<StockCountSession> {
  try {
    const result = await apiPost<StockCountApiSession, CreateStockCountApiRequest>(
      "/stock-counts",
      toApiCreateStockCount(input),
      { accessToken: defaultAccessToken }
    );

    return fromApiStockCount(result);
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }

    return createPrototypeStockCount(input);
  }
}

export async function submitStockCount(id: string, input: SubmitStockCountInput): Promise<StockCountSession> {
  try {
    const result = await apiPost<StockCountApiSession, SubmitStockCountApiRequest>(
      `/stock-counts/${encodeURIComponent(id)}/submit`,
      toApiSubmitStockCount(input),
      { accessToken: defaultAccessToken }
    );

    return fromApiStockCount(result);
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }

    return submitPrototypeStockCount(id, input);
  }
}

export function resetPrototypeStockCountsForTest() {
  prototypeStockCounts = initialPrototypeStockCounts.map(cloneStockCount);
}

export function summarizeAvailableStock(items: AvailableStockItem[]): AvailableStockSummary {
  return items.reduce<AvailableStockSummary>(
    (summary, item) => ({
      baseUomCode:
        summary.baseUomCode === undefined || summary.baseUomCode === item.baseUomCode
          ? (summary.baseUomCode ?? item.baseUomCode)
          : undefined,
      physicalQty: addQuantity(summary.physicalQty, item.physicalQty),
      reservedQty: addQuantity(summary.reservedQty, item.reservedQty),
      qcHoldQty: addQuantity(summary.qcHoldQty, item.qcHoldQty),
      blockedQty: addQuantity(summary.blockedQty, item.blockedQty),
      availableQty: addQuantity(summary.availableQty, item.availableQty)
    }),
    {
      baseUomCode: undefined,
      physicalQty: zeroQuantity,
      reservedQty: zeroQuantity,
      qcHoldQty: zeroQuantity,
      blockedQty: zeroQuantity,
      availableQty: zeroQuantity
    }
  );
}

export function availabilityTone(item: AvailableStockItem): "success" | "warning" | "danger" {
  if (compareQuantity(item.availableQty, zeroQuantity) <= 0) {
    return "danger";
  }
  if (compareQuantity(item.qcHoldQty, zeroQuantity) > 0 || compareQuantity(item.blockedQty, zeroQuantity) > 0) {
    return "warning";
  }
  if (compareQuantity(item.reservedQty, item.availableQty) > 0) {
    return "warning";
  }

  return "success";
}

export function formatQuantity(value: string, uomCode?: string) {
  const normalized = normalizeQuantity(value);
  const negative = normalized.startsWith("-");
  const unsigned = negative ? normalized.slice(1) : normalized;
  const [integerPart, fractionalPart = ""] = unsigned.split(".");
  const groupedInteger = integerPart.replace(/\B(?=(\d{3})+(?!\d))/g, ".");
  const trimmedFraction = fractionalPart.replace(/0+$/, "");
  const formatted = `${negative ? "-" : ""}${groupedInteger}${trimmedFraction ? `,${trimmedFraction}` : ""}`;

  return uomCode ? `${formatted} ${uomCode}` : formatted;
}

function fromApiItem(item: AvailableStockApiItem): AvailableStockItem {
  return {
    warehouseId: item.warehouse_id,
    warehouseCode: item.warehouse_code,
    locationId: item.location_id,
    locationCode: item.location_code,
    sku: item.sku,
    batchId: item.batch_id,
    batchNo: item.batch_no,
    batchQcStatus: item.batch_qc_status,
    batchStatus: item.batch_status,
    batchExpiryDate: item.batch_expiry_date,
    baseUomCode: item.base_uom_code,
    physicalQty: item.physical_qty,
    reservedQty: item.reserved_qty,
    qcHoldQty: item.qc_hold_qty,
    damagedQty: item.damaged_qty,
    returnPendingQty: item.return_pending_qty,
    blockedQty: item.blocked_qty,
    holdQty: item.hold_qty,
    availableQty: item.available_qty
  };
}

function fromApiTransition(item: BatchQCTransitionApiItem): BatchQCTransition {
  return {
    id: item.id,
    batchId: item.batch_id,
    batchNo: item.batch_no,
    sku: item.sku,
    fromQcStatus: item.from_qc_status,
    toQcStatus: item.to_qc_status,
    actorId: item.actor_id,
    reason: item.reason,
    businessRef: item.business_ref,
    auditLogId: item.audit_log_id,
    createdAt: item.created_at
  };
}

function fromApiStockCount(item: StockCountApiSession): StockCountSession {
  return {
    id: item.id,
    countNo: item.count_no,
    orgId: item.org_id,
    warehouseId: item.warehouse_id,
    warehouseCode: item.warehouse_code,
    scope: item.scope,
    status: item.status,
    createdBy: item.created_by,
    submittedBy: item.submitted_by,
    lines: item.lines.map((line) => ({
      id: line.id,
      itemId: line.item_id,
      sku: line.sku,
      batchId: line.batch_id,
      batchNo: line.batch_no,
      locationId: line.location_id,
      locationCode: line.location_code,
      expectedQty: line.expected_qty,
      countedQty: line.counted_qty,
      deltaQty: line.delta_qty,
      baseUomCode: line.base_uom_code,
      counted: line.counted,
      note: line.note
    })),
    auditLogId: item.audit_log_id,
    createdAt: item.created_at,
    updatedAt: item.updated_at,
    submittedAt: item.submitted_at
  };
}

function toApiCreateStockCount(input: CreateStockCountInput): CreateStockCountApiRequest {
  return {
    count_no: input.countNo,
    warehouse_id: input.warehouseId,
    warehouse_code: input.warehouseCode,
    scope: input.scope,
    lines: input.lines.map((line) => ({
      id: line.id,
      item_id: line.itemId,
      sku: line.sku,
      batch_id: line.batchId,
      batch_no: line.batchNo,
      location_id: line.locationId,
      location_code: line.locationCode,
      expected_qty: line.expectedQty,
      base_uom_code: line.baseUomCode
    }))
  };
}

function toApiSubmitStockCount(input: SubmitStockCountInput): SubmitStockCountApiRequest {
  return {
    lines: input.lines.map((line) => ({
      id: line.id,
      counted_qty: line.countedQty,
      note: line.note
    }))
  };
}

function toApiQuery(query: AvailableStockQuery): AvailableStockApiQuery {
  return {
    warehouse_id: query.warehouseId,
    location_id: query.locationId,
    sku: query.sku,
    batch_id: query.batchId
  };
}

function filterPrototypeStock(query: AvailableStockQuery): AvailableStockItem[] {
  const normalizedSKU = query.sku?.trim().toUpperCase();
  return prototypeAvailableStock.filter((item) => {
    if (query.warehouseId && item.warehouseId !== query.warehouseId) {
      return false;
    }
    if (query.locationId && item.locationId !== query.locationId) {
      return false;
    }
    if (normalizedSKU && item.sku !== normalizedSKU) {
      return false;
    }
    if (query.batchId && item.batchId !== query.batchId) {
      return false;
    }

    return true;
  });
}

function createPrototypeStockCount(input: CreateStockCountInput): StockCountSession {
  const now = new Date().toISOString();
  const id = `count-local-${Date.now()}`;
  const count: StockCountSession = {
    id,
    countNo: input.countNo || `CNT-LOCAL-${String(prototypeStockCounts.length + 1).padStart(4, "0")}`,
    orgId: "org-my-pham",
    warehouseId: input.warehouseId,
    warehouseCode: input.warehouseCode,
    scope: input.scope,
    status: "open",
    createdBy: "local-dev",
    lines: input.lines.map((line, index) => ({
      id: line.id || `count-line-local-${index + 1}`,
      itemId: line.itemId,
      sku: line.sku,
      batchId: line.batchId,
      batchNo: line.batchNo,
      locationId: line.locationId,
      locationCode: line.locationCode,
      expectedQty: normalizeQuantity(line.expectedQty),
      countedQty: zeroQuantity,
      deltaQty: zeroQuantity,
      baseUomCode: line.baseUomCode,
      counted: false
    })),
    auditLogId: `audit-${id}`,
    createdAt: now,
    updatedAt: now
  };
  prototypeStockCounts = [cloneStockCount(count), ...prototypeStockCounts];

  return cloneStockCount(count);
}

function submitPrototypeStockCount(id: string, input: SubmitStockCountInput): StockCountSession {
  const index = prototypeStockCounts.findIndex((count) => count.id === id);
  if (index < 0) {
    throw new Error("Stock count not found");
  }
  const current = prototypeStockCounts[index];
  if (current.status !== "open") {
    throw new Error("Stock count is not open");
  }

  const countedByLine = new Map(input.lines.map((line) => [line.id, line]));
  const now = new Date().toISOString();
  const updatedLines = current.lines.map((line) => {
    const counted = countedByLine.get(line.id);
    if (!counted) {
      throw new Error("All stock count lines must be counted");
    }
    const countedQty = normalizeQuantity(counted.countedQty);

    return {
      ...line,
      countedQty,
      deltaQty: subtractQuantity(countedQty, line.expectedQty),
      counted: true,
      note: counted.note
    };
  });
  const status: StockCountStatus = updatedLines.some((line) => compareQuantity(line.deltaQty, zeroQuantity) !== 0)
    ? "variance_review"
    : "submitted";
  const updated: StockCountSession = {
    ...current,
    status,
    submittedBy: "local-dev",
    lines: updatedLines,
    updatedAt: now,
    submittedAt: now
  };
  prototypeStockCounts = prototypeStockCounts.map((count) => (count.id === id ? cloneStockCount(updated) : count));

  return cloneStockCount(updated);
}

function cloneStockCount(count: StockCountSession): StockCountSession {
  return {
    ...count,
    lines: count.lines.map((line) => ({ ...line }))
  };
}

function addQuantity(left: string, right: string) {
  return scaledToQuantity(quantityToScaled(left) + quantityToScaled(right));
}

function subtractQuantity(left: string, right: string) {
  return scaledToQuantity(quantityToScaled(left) - quantityToScaled(right));
}

function compareQuantity(left: string, right: string) {
  const leftValue = quantityToScaled(left);
  const rightValue = quantityToScaled(right);
  if (leftValue === rightValue) {
    return 0;
  }

  return leftValue > rightValue ? 1 : -1;
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
