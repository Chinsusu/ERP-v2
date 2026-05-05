export type GoodsReceiptStatus = "draft" | "submitted" | "inspect_ready" | "posted";

export type BatchQCStatus = "hold" | "pass" | "fail" | "quarantine" | "retest_required";

export type ReceivingPackagingStatus = "intact" | "damaged" | "missing_label" | "leaking";

export type GoodsReceiptLine = {
  id: string;
  purchaseOrderLineId?: string;
  itemId: string;
  sku: string;
  itemName?: string;
  batchId?: string;
  batchNo?: string;
  lotNo?: string;
  expiryDate?: string;
  warehouseId: string;
  locationId: string;
  quantity: string;
  uomCode: string;
  baseUomCode: string;
  packagingStatus: ReceivingPackagingStatus;
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
  deliveryNoteNo?: string;
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
  referenceDocId?: string;
};

export type CreateGoodsReceiptLineInput = {
  id?: string;
  purchaseOrderLineId: string;
  itemId?: string;
  sku?: string;
  itemName?: string;
  batchId?: string;
  batchNo?: string;
  lotNo: string;
  expiryDate: string;
  quantity: string;
  uomCode: string;
  baseUomCode: string;
  packagingStatus: ReceivingPackagingStatus;
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
  supplierId: string;
  deliveryNoteNo: string;
  lines: CreateGoodsReceiptLineInput[];
};
