export type AvailableStockItem = {
  warehouseId: string;
  warehouseCode: string;
  locationId?: string;
  locationCode?: string;
  sku: string;
  batchId?: string;
  batchNo?: string;
  batchQcStatus?: BatchQCStatus;
  batchStatus?: BatchStatus;
  batchExpiryDate?: string;
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

export type BatchQCStatus = "hold" | "pass" | "fail" | "quarantine" | "retest_required";

export type BatchStatus = "active" | "inactive" | "blocked";

export type BatchQCTransition = {
  id: string;
  batchId: string;
  batchNo: string;
  sku: string;
  fromQcStatus: BatchQCStatus;
  toQcStatus: BatchQCStatus;
  actorId: string;
  reason: string;
  businessRef: string;
  auditLogId: string;
  createdAt: string;
};

export type BatchQCTransitionInput = {
  qcStatus: BatchQCStatus;
  reason: string;
  businessRef?: string;
};

export type BatchQCTransitionResult = {
  transition: BatchQCTransition;
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

export type StockCountStatus = "open" | "submitted" | "variance_review";

export type StockCountLine = {
  id: string;
  itemId?: string;
  sku: string;
  batchId?: string;
  batchNo?: string;
  locationId?: string;
  locationCode?: string;
  expectedQty: string;
  countedQty: string;
  deltaQty: string;
  baseUomCode: string;
  counted: boolean;
  note?: string;
};

export type StockCountSession = {
  id: string;
  countNo: string;
  orgId: string;
  warehouseId: string;
  warehouseCode?: string;
  scope: string;
  status: StockCountStatus;
  createdBy: string;
  submittedBy?: string;
  lines: StockCountLine[];
  auditLogId?: string;
  createdAt: string;
  updatedAt: string;
  submittedAt?: string;
};

export type CreateStockCountInput = {
  countNo?: string;
  warehouseId: string;
  warehouseCode?: string;
  scope: string;
  lines: Array<{
    id?: string;
    itemId?: string;
    sku: string;
    batchId?: string;
    batchNo?: string;
    locationId?: string;
    locationCode?: string;
    expectedQty: string;
    baseUomCode: string;
  }>;
};

export type SubmitStockCountInput = {
  lines: Array<{
    id: string;
    countedQty: string;
    note?: string;
  }>;
};

export type StockAdjustmentStatus = "draft" | "submitted" | "approved" | "rejected" | "posted" | "cancelled";

export type StockAdjustmentLine = {
  id: string;
  itemId?: string;
  sku: string;
  batchId?: string;
  batchNo?: string;
  locationId?: string;
  locationCode?: string;
  expectedQty: string;
  countedQty: string;
  deltaQty: string;
  baseUomCode: string;
  reason?: string;
};

export type StockAdjustment = {
  id: string;
  adjustmentNo: string;
  orgId: string;
  warehouseId: string;
  warehouseCode?: string;
  sourceType?: string;
  sourceId?: string;
  reason: string;
  status: StockAdjustmentStatus;
  requestedBy: string;
  submittedBy?: string;
  approvedBy?: string;
  rejectedBy?: string;
  postedBy?: string;
  lines: StockAdjustmentLine[];
  auditLogId?: string;
  createdAt: string;
  updatedAt: string;
  submittedAt?: string;
  approvedAt?: string;
  rejectedAt?: string;
  postedAt?: string;
};

export type StockAdjustmentAction = "submit" | "approve" | "reject" | "post";
