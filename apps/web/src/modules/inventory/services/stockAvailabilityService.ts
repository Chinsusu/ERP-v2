import { apiGet } from "../../../shared/api/client";
import type { AvailableStockItem, AvailableStockQuery, AvailableStockSummary } from "../types";

type AvailableStockApiItem = {
  warehouse_id: string;
  warehouse_code: string;
  sku: string;
  batch_id?: string;
  batch_no?: string;
  physical_stock: number;
  reserved_stock: number;
  hold_stock: number;
  available_stock: number;
};

const defaultAccessToken = "local-dev-access-token";

export const prototypeAvailableStock: AvailableStockItem[] = [
  {
    warehouseId: "wh-hcm",
    warehouseCode: "HCM",
    sku: "SERUM-30ML",
    batchId: "batch-serum-2604a",
    batchNo: "LOT-2604A",
    physicalStock: 128,
    reservedStock: 10,
    holdStock: 8,
    availableStock: 110
  },
  {
    warehouseId: "wh-hcm",
    warehouseCode: "HCM",
    sku: "CREAM-50G",
    batchId: "batch-cream-2603b",
    batchNo: "LOT-2603B",
    physicalStock: 46,
    reservedStock: 12,
    holdStock: 2,
    availableStock: 32
  },
  {
    warehouseId: "wh-hn",
    warehouseCode: "HN",
    sku: "TONER-100ML",
    batchId: "batch-toner-2604c",
    batchNo: "LOT-2604C",
    physicalStock: 90,
    reservedStock: 20,
    holdStock: 5,
    availableStock: 65
  }
];

export async function getAvailableStock(query: AvailableStockQuery = {}): Promise<AvailableStockItem[]> {
  try {
    const items = await apiGet<AvailableStockApiItem[]>(`/inventory/available-stock${queryString(query)}`, {
      accessToken: defaultAccessToken
    });

    return items.map(fromApiItem);
  } catch {
    return filterPrototypeStock(query);
  }
}

export function summarizeAvailableStock(items: AvailableStockItem[]): AvailableStockSummary {
  return items.reduce(
    (summary, item) => ({
      physicalStock: summary.physicalStock + item.physicalStock,
      reservedStock: summary.reservedStock + item.reservedStock,
      holdStock: summary.holdStock + item.holdStock,
      availableStock: summary.availableStock + item.availableStock
    }),
    {
      physicalStock: 0,
      reservedStock: 0,
      holdStock: 0,
      availableStock: 0
    }
  );
}

export function availabilityTone(item: AvailableStockItem): "success" | "warning" | "danger" {
  if (item.availableStock <= 0) {
    return "danger";
  }
  if (item.holdStock > 0 || item.reservedStock > item.availableStock) {
    return "warning";
  }

  return "success";
}

function fromApiItem(item: AvailableStockApiItem): AvailableStockItem {
  return {
    warehouseId: item.warehouse_id,
    warehouseCode: item.warehouse_code,
    sku: item.sku,
    batchId: item.batch_id,
    batchNo: item.batch_no,
    physicalStock: item.physical_stock,
    reservedStock: item.reserved_stock,
    holdStock: item.hold_stock,
    availableStock: item.available_stock
  };
}

function filterPrototypeStock(query: AvailableStockQuery): AvailableStockItem[] {
  const normalizedSKU = query.sku?.trim().toUpperCase();
  return prototypeAvailableStock.filter((item) => {
    if (query.warehouseId && item.warehouseId !== query.warehouseId) {
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

function queryString(query: AvailableStockQuery) {
  const params = new URLSearchParams();
  if (query.warehouseId) {
    params.set("warehouse_id", query.warehouseId);
  }
  if (query.sku) {
    params.set("sku", query.sku);
  }
  if (query.batchId) {
    params.set("batch_id", query.batchId);
  }

  const value = params.toString();
  return value ? `?${value}` : "";
}
