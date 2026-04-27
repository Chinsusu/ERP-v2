export type AvailableStockItem = {
  warehouseId: string;
  warehouseCode: string;
  locationId?: string;
  locationCode?: string;
  sku: string;
  batchId?: string;
  batchNo?: string;
  baseUomCode: string;
  physicalQty: string;
  reservedQty: string;
  qcHoldQty: string;
  damagedQty: string;
  returnPendingQty: string;
  blockedQty: string;
  holdQty: string;
  availableQty: string;
};

export type AvailableStockQuery = {
  warehouseId?: string;
  locationId?: string;
  sku?: string;
  batchId?: string;
};

export type AvailableStockSummary = {
  baseUomCode?: string;
  physicalQty: string;
  reservedQty: string;
  qcHoldQty: string;
  blockedQty: string;
  availableQty: string;
};
