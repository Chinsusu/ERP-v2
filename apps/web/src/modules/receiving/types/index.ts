export type GoodsReceiptStatus = "draft" | "submitted" | "inspect_ready" | "posted";

export type BatchQCStatus = "hold" | "pass" | "fail" | "quarantine" | "retest_required";

export type GoodsReceiptLine = {
  id: string;
  itemId: string;
  sku: string;
  itemName?: string;
  batchId?: string;
  batchNo?: string;
  warehouseId: string;
  locationId: string;
  quantity: string;
  baseUomCode: string;
  qcStatus?: BatchQCStatus;
};

export type GoodsReceiptStockMovement = {
  movementNo: string;
  movementType: "purchase_receipt";
  itemId: string;
  batchId: string;
  warehouseId: string;
  locationId: string;
  quantity: string;
  baseUomCode: string;
  stockStatus: "available" | "qc_hold";
  sourceDocId: string;
  sourceDocLineId: string;
};

export type GoodsReceipt = {
  id: string;
  orgId: string;
  receiptNo: string;
  warehouseId: string;
  warehouseCode: string;
  locationId: string;
  locationCode: string;
  referenceDocType: string;
  referenceDocId: string;
  supplierId?: string;
  status: GoodsReceiptStatus;
  lines: GoodsReceiptLine[];
  stockMovements?: GoodsReceiptStockMovement[];
  createdBy: string;
  submittedBy?: string;
  inspectReadyBy?: string;
  postedBy?: string;
  auditLogId?: string;
  createdAt: string;
  updatedAt: string;
  submittedAt?: string;
  inspectReadyAt?: string;
  postedAt?: string;
};

export type GoodsReceiptQuery = {
  warehouseId?: string;
  status?: GoodsReceiptStatus;
};

export type CreateGoodsReceiptLineInput = {
  id?: string;
  itemId?: string;
  sku?: string;
  itemName?: string;
  batchId?: string;
  batchNo?: string;
  quantity: string;
  baseUomCode: string;
  qcStatus?: BatchQCStatus;
};

export type CreateGoodsReceiptInput = {
  id?: string;
  orgId?: string;
  receiptNo?: string;
  warehouseId: string;
  locationId: string;
  referenceDocType: string;
  referenceDocId: string;
  supplierId?: string;
  lines: CreateGoodsReceiptLineInput[];
};
