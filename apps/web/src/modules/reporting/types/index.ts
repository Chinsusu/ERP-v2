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

export type ReportSourceReference = {
  entityType: string;
  id: string;
  label: string;
  href?: string;
  unavailable: boolean;
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
  sourceReferences: ReportSourceReference[];
};

export type OperationsDailyQuery = {
  fromDate?: string;
  toDate?: string;
  businessDate?: string;
  warehouseId?: string;
  status?: OperationsDailyStatusFilter;
};

export type OperationsDailyStatusFilter = "" | "pending" | "in_progress" | "completed" | "blocked" | "exception";

export type OperationsDailyReport = {
  metadata: ReportMetadata;
  summary: OperationsDailySummary;
  areas: OperationsDailyAreaSummary[];
  rows: OperationsDailyRow[];
};

export type OperationsDailySummary = {
  signalCount: number;
  pendingCount: number;
  inProgressCount: number;
  completedCount: number;
  blockedCount: number;
  exceptionCount: number;
};

export type OperationsDailyAreaSummary = OperationsDailySummary & {
  area: string;
};

export type OperationsDailyRow = {
  id: string;
  area: string;
  sourceType: string;
  sourceId: string;
  sourceReference: ReportSourceReference;
  refNo: string;
  title: string;
  warehouseId: string;
  warehouseCode?: string;
  businessDate: string;
  status: string;
  severity: string;
  quantity?: string;
  uomCode?: string;
  exceptionCode?: string;
  owner?: string;
};

export type FinanceSummaryQuery = {
  fromDate?: string;
  toDate?: string;
  businessDate?: string;
};

export type FinanceSummaryReport = {
  metadata: ReportMetadata;
  currencyCode: string;
  ar: FinanceSummaryReceivable;
  ap: FinanceSummaryPayable;
  cod: FinanceSummaryCOD;
  cash: FinanceSummaryCash;
};

export type FinanceSummaryReceivable = {
  openCount: number;
  overdueCount: number;
  disputedCount: number;
  openAmount: string;
  overdueAmount: string;
  outstandingAmount: string;
  agingBuckets: FinanceSummaryAgingBucket[];
};

export type FinanceSummaryPayable = {
  openCount: number;
  dueCount: number;
  paymentRequestedCount: number;
  paymentApprovedCount: number;
  openAmount: string;
  dueAmount: string;
  outstandingAmount: string;
  agingBuckets: FinanceSummaryAgingBucket[];
};

export type FinanceSummaryCOD = {
  pendingCount: number;
  discrepancyCount: number;
  pendingAmount: string;
  discrepancyAmount: string;
  discrepancyBuckets: FinanceSummaryDiscrepancyBucket[];
};

export type FinanceSummaryCash = {
  transactionCount: number;
  cashInAmount: string;
  cashOutAmount: string;
  netCashAmount: string;
};

export type FinanceSummaryAgingBucket = {
  bucket: string;
  count: number;
  amount: string;
};

export type FinanceSummaryDiscrepancyBucket = {
  type: string;
  status: string;
  count: number;
  amount: string;
};
