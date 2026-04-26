export type AvailableStockItem = {
  warehouseId: string;
  warehouseCode: string;
  sku: string;
  batchId?: string;
  batchNo?: string;
  physicalStock: number;
  reservedStock: number;
  holdStock: number;
  availableStock: number;
};

export type AvailableStockQuery = {
  warehouseId?: string;
  sku?: string;
  batchId?: string;
};

export type AvailableStockSummary = {
  physicalStock: number;
  reservedStock: number;
  holdStock: number;
  availableStock: number;
};
