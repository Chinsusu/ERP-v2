export type ReportFilters = {
  fromDate: string;
  toDate: string;
  businessDate: string;
  warehouseId?: string;
  status?: string;
  itemId?: string;
  category?: string;
};

export type ReportMetadata = {
  generatedAt: string;
  timezone: string;
  sourceVersion: string;
  filters: ReportFilters;
};

export type InventorySnapshotQuery = {
  fromDate?: string;
  toDate?: string;
  businessDate?: string;
  warehouseId?: string;
  status?: InventorySnapshotStatusFilter;
  itemId?: string;
  category?: string;
  locationId?: string;
  sku?: string;
  batchId?: string;
  lowStockThreshold?: string;
  expiryWarningDays?: string;
};

export type InventorySnapshotStatusFilter = "" | "available" | "reserved" | "quarantine" | "blocked";

export type InventorySnapshotReport = {
  metadata: ReportMetadata;
  summary: InventorySnapshotSummary;
  rows: InventorySnapshotRow[];
};

export type InventorySnapshotSummary = {
  rowCount: number;
  lowStockRowCount: number;
  expiryWarningRows: number;
  expiredRows: number;
  totalsByUom: InventorySnapshotUOMTotal[];
};

export type InventorySnapshotUOMTotal = {
  baseUomCode: string;
  physicalQty: string;
  reservedQty: string;
  quarantineQty: string;
  blockedQty: string;
  availableQty: string;
};

export type InventorySnapshotRow = {
  warehouseId: string;
  warehouseCode?: string;
  locationId?: string;
  locationCode?: string;
  itemId?: string;
  sku: string;
  batchId?: string;
  batchNo?: string;
  batchExpiry?: string;
  baseUomCode: string;
  physicalQty: string;
  reservedQty: string;
  quarantineQty: string;
  blockedQty: string;
  availableQty: string;
  lowStock: boolean;
  expiryWarning: boolean;
  expired: boolean;
  batchQcStatus?: string;
  batchStatus?: string;
  sourceStockState: string;
};
