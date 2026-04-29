export type ReturnReceiptStatus = "pending_inspection" | "inspected" | "dispositioned";

export type ReturnSource = "SHIPPER" | "CARRIER" | "CUSTOMER" | "MARKETPLACE" | "UNKNOWN";

export type ReturnDisposition = "reusable" | "not_reusable" | "needs_inspection";

export type ReturnInspectionCondition =
  | "intact"
  | "dented_box"
  | "seal_torn"
  | "used"
  | "damaged"
  | "missing_accessory";

export type ReturnInspectionDisposition = ReturnDisposition;

export type ReturnInspectionStatus = "inspection_recorded" | "return_qa_hold";

export type ReturnReceiptLine = {
  id: string;
  sku: string;
  productName: string;
  quantity: number;
  condition: string;
};

export type ReturnStockMovement = {
  id: string;
  movementType: "RETURN_RECEIPT";
  sku: string;
  warehouseId: string;
  quantity: number;
  targetStockStatus: "return_pending";
  sourceDocId: string;
};

export type ReturnReceipt = {
  id: string;
  receiptNo: string;
  warehouseId: string;
  warehouseCode: string;
  source: ReturnSource;
  receivedBy: string;
  receivedAt: string;
  packageCondition: string;
  status: ReturnReceiptStatus;
  disposition: ReturnDisposition;
  targetLocation: string;
  originalOrderNo?: string;
  trackingNo?: string;
  returnCode?: string;
  channel?: string;
  batchNo?: string;
  deliveredAt?: string;
  returnReason?: string;
  scanCode: string;
  customerName: string;
  unknownCase: boolean;
  lines: ReturnReceiptLine[];
  stockMovement?: ReturnStockMovement;
  investigationNote?: string;
  auditLogId?: string;
  createdAt: string;
};

export type ReturnInspectionResult = {
  id: string;
  receiptId: string;
  receiptNo: string;
  condition: ReturnInspectionCondition;
  disposition: ReturnInspectionDisposition;
  status: ReturnInspectionStatus;
  targetLocation: string;
  riskLevel: "low" | "medium" | "high";
  inspectorId: string;
  note?: string;
  evidenceLabel?: string;
  inspectedAt: string;
};

export type ReturnDispositionAction = {
  id: string;
  receiptId: string;
  receiptNo: string;
  disposition: ReturnDisposition;
  targetLocation: string;
  targetStockStatus: "return_pending" | "damaged" | "qc_hold";
  actionCode: "route_to_putaway" | "route_to_lab_or_damaged" | "route_to_quarantine_hold";
  actorId: string;
  note?: string;
  auditLogId?: string;
  decidedAt: string;
};

export type ReturnAttachment = {
  id: string;
  receiptId: string;
  receiptNo: string;
  inspectionId: string;
  fileName: string;
  fileExt?: string;
  mimeType: string;
  fileSizeBytes: number;
  storageBucket: string;
  storageKey: string;
  status: "active" | "deleted" | "quarantined";
  uploadedBy: string;
  note?: string;
  auditLogId?: string;
  uploadedAt: string;
};

export type InspectReturnInput = {
  receipt: ReturnReceipt;
  condition: ReturnInspectionCondition;
  disposition: ReturnInspectionDisposition;
  note?: string;
  evidenceLabel?: string;
};

export type ApplyReturnDispositionInput = {
  receipt: ReturnReceipt;
  disposition: ReturnDisposition;
  note?: string;
};

export type UploadReturnAttachmentInput = {
  receipt: ReturnReceipt;
  inspectionId: string;
  file: File;
  note?: string;
};

export type ReturnReceiptQuery = {
  warehouseId?: string;
  status?: ReturnReceiptStatus;
};

export type ReceiveReturnInput = {
  warehouseId: string;
  warehouseCode: string;
  source: ReturnSource;
  code: string;
  packageCondition: string;
  disposition: ReturnDisposition;
  investigationNote?: string;
};

export type SupplierRejectionStatus = "draft" | "submitted" | "confirmed" | "cancelled";

export type SupplierRejectionLine = {
  id: string;
  purchaseOrderLineId?: string;
  goodsReceiptLineId: string;
  inboundQCInspectionId: string;
  itemId: string;
  sku: string;
  itemName?: string;
  batchId: string;
  batchNo: string;
  lotNo: string;
  expiryDate: string;
  rejectedQuantity: string;
  uomCode: string;
  baseUOMCode: string;
  reason: string;
};

export type SupplierRejectionAttachment = {
  id: string;
  lineId?: string;
  fileName: string;
  objectKey: string;
  contentType?: string;
  uploadedAt?: string;
  uploadedBy?: string;
  source?: string;
};

export type SupplierRejection = {
  id: string;
  orgId: string;
  rejectionNo: string;
  supplierId: string;
  supplierCode?: string;
  supplierName?: string;
  purchaseOrderId?: string;
  purchaseOrderNo?: string;
  goodsReceiptId: string;
  goodsReceiptNo?: string;
  inboundQCInspectionId: string;
  warehouseId: string;
  warehouseCode?: string;
  status: SupplierRejectionStatus;
  reason: string;
  lines: SupplierRejectionLine[];
  attachments: SupplierRejectionAttachment[];
  auditLogId?: string;
  createdAt: string;
  createdBy: string;
  updatedAt: string;
  updatedBy: string;
  submittedAt?: string;
  submittedBy?: string;
  confirmedAt?: string;
  confirmedBy?: string;
};

export type SupplierRejectionQuery = {
  supplierId?: string;
  warehouseId?: string;
  status?: SupplierRejectionStatus;
};

export type CreateSupplierRejectionLineInput = {
  id?: string;
  purchaseOrderLineId?: string;
  goodsReceiptLineId: string;
  inboundQCInspectionId: string;
  itemId: string;
  sku: string;
  itemName?: string;
  batchId: string;
  batchNo: string;
  lotNo: string;
  expiryDate: string;
  rejectedQuantity: string;
  uomCode: string;
  baseUOMCode: string;
  reason: string;
};

export type CreateSupplierRejectionAttachmentInput = {
  id?: string;
  lineId?: string;
  fileName: string;
  objectKey?: string;
  contentType?: string;
  source?: string;
};

export type CreateSupplierRejectionInput = {
  id?: string;
  orgId?: string;
  rejectionNo?: string;
  supplierId: string;
  supplierCode?: string;
  supplierName?: string;
  purchaseOrderId?: string;
  purchaseOrderNo?: string;
  goodsReceiptId: string;
  goodsReceiptNo?: string;
  inboundQCInspectionId: string;
  warehouseId: string;
  warehouseCode?: string;
  reason: string;
  lines: CreateSupplierRejectionLineInput[];
  attachments?: CreateSupplierRejectionAttachmentInput[];
};

export type SupplierRejectionActionResult = {
  rejection: SupplierRejection;
  previousStatus?: SupplierRejectionStatus;
  currentStatus: SupplierRejectionStatus;
  auditLogId?: string;
};
