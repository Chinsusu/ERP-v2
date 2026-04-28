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
