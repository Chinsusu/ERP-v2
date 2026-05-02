import { apiGetRaw, apiPost } from "../../../shared/api/client";
import { shouldUsePrototypeFallback } from "../../../shared/api/prototypeFallback";
import type { StockAdjustment, StockAdjustmentAction, StockAdjustmentStatus } from "../types";

type StockAdjustmentApiLine = {
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
  reason?: string;
};

type StockAdjustmentApiItem = {
  id: string;
  adjustment_no: string;
  org_id: string;
  warehouse_id: string;
  warehouse_code?: string;
  source_type?: string;
  source_id?: string;
  reason: string;
  status: StockAdjustmentStatus;
  requested_by: string;
  submitted_by?: string;
  approved_by?: string;
  rejected_by?: string;
  posted_by?: string;
  lines: StockAdjustmentApiLine[];
  audit_log_id?: string;
  created_at: string;
  updated_at: string;
  submitted_at?: string;
  approved_at?: string;
  rejected_at?: string;
  posted_at?: string;
};

const defaultAccessToken = "local-dev-access-token";
const quantityScale = 6;
const zeroQuantity = "0.000000";

const initialPrototypeStockAdjustments: StockAdjustment[] = [
  {
    id: "adj-hcm-260426-0001",
    adjustmentNo: "ADJ-260426-0001",
    orgId: "org-my-pham",
    warehouseId: "wh-hcm",
    warehouseCode: "HCM",
    sourceType: "stock_count",
    sourceId: "count-hcm-260426-0001",
    reason: "Cycle count variance approval",
    status: "submitted",
    requestedBy: "user-warehouse",
    submittedBy: "user-warehouse",
    lines: [
      {
        id: "adj-line-hcm-0001",
        sku: "SERUM-30ML",
        batchId: "batch-serum-2604a",
        batchNo: "LOT-2604A",
        locationId: "bin-hcm-a01",
        locationCode: "A-01",
        expectedQty: "128.000000",
        countedQty: "126.000000",
        deltaQty: "-2.000000",
        baseUomCode: "PCS",
        reason: "two damaged units found during count"
      }
    ],
    auditLogId: "audit-stock-adjustment-260427-0001",
    createdAt: "2026-04-26T09:00:00Z",
    updatedAt: "2026-04-26T09:05:00Z",
    submittedAt: "2026-04-26T09:05:00Z"
  }
];

let prototypeStockAdjustments = initialPrototypeStockAdjustments.map(cloneStockAdjustment);

export async function getStockAdjustments(): Promise<StockAdjustment[]> {
  try {
    const items = await apiGetRaw<StockAdjustmentApiItem[]>("/stock-adjustments", {
      accessToken: defaultAccessToken
    });

    return items.map(fromApiStockAdjustment);
  } catch (reason) {
    if (!shouldUsePrototypeFallback(reason)) {
      throw reason;
    }

    return prototypeStockAdjustments.map(cloneStockAdjustment);
  }
}

export async function transitionStockAdjustment(
  id: string,
  action: StockAdjustmentAction
): Promise<StockAdjustment> {
  try {
    const item = await apiPost<StockAdjustmentApiItem, Record<string, never>>(
      `/stock-adjustments/${encodeURIComponent(id)}/${action}`,
      {},
      { accessToken: defaultAccessToken }
    );

    return fromApiStockAdjustment(item);
  } catch (error) {
    if (!shouldUsePrototypeFallback(error)) {
      throw error;
    }

    return transitionPrototypeStockAdjustment(id, action);
  }
}

export function resetPrototypeStockAdjustmentsForTest() {
  prototypeStockAdjustments = initialPrototypeStockAdjustments.map(cloneStockAdjustment);
}

function fromApiStockAdjustment(item: StockAdjustmentApiItem): StockAdjustment {
  return {
    id: item.id,
    adjustmentNo: item.adjustment_no,
    orgId: item.org_id,
    warehouseId: item.warehouse_id,
    warehouseCode: item.warehouse_code,
    sourceType: item.source_type,
    sourceId: item.source_id,
    reason: item.reason,
    status: item.status,
    requestedBy: item.requested_by,
    submittedBy: item.submitted_by,
    approvedBy: item.approved_by,
    rejectedBy: item.rejected_by,
    postedBy: item.posted_by,
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
      reason: line.reason
    })),
    auditLogId: item.audit_log_id,
    createdAt: item.created_at,
    updatedAt: item.updated_at,
    submittedAt: item.submitted_at,
    approvedAt: item.approved_at,
    rejectedAt: item.rejected_at,
    postedAt: item.posted_at
  };
}

function transitionPrototypeStockAdjustment(id: string, action: StockAdjustmentAction): StockAdjustment {
  const index = prototypeStockAdjustments.findIndex((adjustment) => adjustment.id === id);
  if (index < 0) {
    throw new Error("Stock adjustment not found");
  }

  const current = prototypeStockAdjustments[index];
  const now = new Date().toISOString();
  const next = nextStatus(current.status, action);
  const updated: StockAdjustment = {
    ...current,
    status: next,
    auditLogId: `audit-stock-adjustment-${action}-${Date.now()}`,
    updatedAt: now,
    submittedBy: action === "submit" ? "local-dev" : current.submittedBy,
    approvedBy: action === "approve" ? "local-dev" : current.approvedBy,
    rejectedBy: action === "reject" ? "local-dev" : current.rejectedBy,
    postedBy: action === "post" ? "local-dev" : current.postedBy,
    submittedAt: action === "submit" ? now : current.submittedAt,
    approvedAt: action === "approve" ? now : current.approvedAt,
    rejectedAt: action === "reject" ? now : current.rejectedAt,
    postedAt: action === "post" ? now : current.postedAt
  };

  prototypeStockAdjustments = prototypeStockAdjustments.map((adjustment) =>
    adjustment.id === id ? cloneStockAdjustment(updated) : adjustment
  );

  return cloneStockAdjustment(updated);
}

function nextStatus(status: StockAdjustmentStatus, action: StockAdjustmentAction): StockAdjustmentStatus {
  if (action === "submit" && status === "draft") {
    return "submitted";
  }
  if (action === "approve" && status === "submitted") {
    return "approved";
  }
  if (action === "reject" && status === "submitted") {
    return "rejected";
  }
  if (action === "post" && status === "approved") {
    return "posted";
  }

  throw new Error("Stock adjustment action is not allowed");
}

function cloneStockAdjustment(adjustment: StockAdjustment): StockAdjustment {
  return {
    ...adjustment,
    lines: adjustment.lines.map((line) => ({ ...line }))
  };
}

export function summarizeStockAdjustmentDelta(adjustment: StockAdjustment) {
  const firstLine = adjustment.lines[0];
  if (!firstLine) {
    return { deltaQty: zeroQuantity, baseUomCode: undefined };
  }
  const sameUOM = adjustment.lines.every((line) => line.baseUomCode === firstLine.baseUomCode);
  if (!sameUOM) {
    return { deltaQty: zeroQuantity, baseUomCode: undefined };
  }

  return {
    deltaQty: adjustment.lines.reduce((total, line) => addQuantity(total, line.deltaQty), zeroQuantity),
    baseUomCode: firstLine.baseUomCode
  };
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
