export type ReturnReceiptStatus = "pending_inspection";

export type ReturnSource = "SHIPPER" | "CARRIER" | "CUSTOMER" | "MARKETPLACE" | "UNKNOWN";

export type ReturnDisposition = "reusable" | "not_reusable" | "needs_inspection";

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
  scanCode: string;
  customerName: string;
  unknownCase: boolean;
  lines: ReturnReceiptLine[];
  stockMovement?: ReturnStockMovement;
  investigationNote?: string;
  auditLogId?: string;
  createdAt: string;
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
