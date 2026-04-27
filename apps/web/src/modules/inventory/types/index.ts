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
