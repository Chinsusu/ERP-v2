import { apiGet } from "../../../shared/api/client";
import type { components, operations } from "../../../shared/api/generated/schema";
import type { AvailableStockItem, AvailableStockQuery, AvailableStockSummary } from "../types";

type AvailableStockApiItem = components["schemas"]["AvailableStockItem"];
type AvailableStockApiQuery = operations["listAvailableStock"]["parameters"]["query"];

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

function addQuantity(left: string, right: string) {
  return scaledToQuantity(quantityToScaled(left) + quantityToScaled(right));
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
