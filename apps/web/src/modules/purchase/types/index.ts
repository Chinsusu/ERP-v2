export type PurchaseOrderStatus =
  | "draft"
  | "submitted"
  | "approved"
  | "partially_received"
  | "received"
  | "closed"
  | "cancelled"
  | "rejected";

export type PurchaseOrderLine = {
  id: string;
  lineNo: number;
  itemId: string;
  skuCode: string;
  itemName: string;
  orderedQty: string;
  receivedQty: string;
  uomCode: string;
  baseOrderedQty: string;
  baseReceivedQty: string;
  baseUomCode: string;
  conversionFactor: string;
  unitPrice: string;
  currencyCode: string;
  lineAmount: string;
  expectedDate: string;
  note?: string;
};

export type PurchaseOrder = {
  id: string;
  poNo: string;
  supplierId: string;
  supplierCode?: string;
  supplierName: string;
  warehouseId: string;
  warehouseCode?: string;
  expectedDate: string;
  status: PurchaseOrderStatus;
  currencyCode: string;
  subtotalAmount: string;
  totalAmount: string;
  note?: string;
  lineCount?: number;
  receivedLineCount?: number;
  lines: PurchaseOrderLine[];
  auditLogId?: string;
  createdAt: string;
  updatedAt: string;
  submittedAt?: string;
  approvedAt?: string;
  partiallyReceivedAt?: string;
  receivedAt?: string;
  closedAt?: string;
  cancelledAt?: string;
  rejectedAt?: string;
  cancelReason?: string;
  rejectReason?: string;
  version: number;
};

export type PurchaseOrderQuery = {
  search?: string;
  status?: PurchaseOrderStatus;
  supplierId?: string;
  warehouseId?: string;
  expectedFrom?: string;
  expectedTo?: string;
};

export type PurchaseOrderLineInput = {
  id?: string;
  lineNo?: number;
  itemId: string;
  orderedQty: string;
  uomCode: string;
  unitPrice: string;
  currencyCode?: string;
  expectedDate?: string;
  note?: string;
};

export type CreatePurchaseOrderInput = {
  id?: string;
  poNo?: string;
  supplierId: string;
  warehouseId: string;
  expectedDate: string;
  currencyCode: string;
  note?: string;
  lines: PurchaseOrderLineInput[];
};

export type UpdatePurchaseOrderInput = {
  supplierId?: string;
  warehouseId?: string;
  expectedDate?: string;
  note?: string;
  expectedVersion?: number;
  lines?: PurchaseOrderLineInput[];
};

export type PurchaseOrderActionResult = {
  purchaseOrder: PurchaseOrder;
  previousStatus: PurchaseOrderStatus;
  currentStatus: PurchaseOrderStatus;
  auditLogId?: string;
};

export type PurchaseRequestStatus =
  | "draft"
  | "submitted"
  | "approved"
  | "converted_to_po"
  | "cancelled"
  | "rejected";

export type PurchaseRequestLine = {
  id: string;
  lineNo: number;
  sourceProductionPlanLineId: string;
  itemId?: string;
  sku: string;
  itemName: string;
  requestedQty: string;
  uomCode: string;
  note?: string;
};

export type PurchaseRequest = {
  id: string;
  requestNo: string;
  sourceProductionPlanId: string;
  sourceProductionPlanNo: string;
  status: PurchaseRequestStatus;
  lines: PurchaseRequestLine[];
  createdAt?: string;
  createdBy?: string;
  submittedAt?: string;
  submittedBy?: string;
  approvedAt?: string;
  approvedBy?: string;
  convertedAt?: string;
  convertedBy?: string;
  convertedPurchaseOrderId?: string;
  convertedPurchaseOrderNo?: string;
  cancelledAt?: string;
  cancelledBy?: string;
  rejectedAt?: string;
  rejectedBy?: string;
  rejectReason?: string;
};

export type PurchaseRequestQuery = {
  search?: string;
  status?: PurchaseRequestStatus | "";
  sourceProductionPlanId?: string;
};

export type PurchaseRequestActionResult = {
  purchaseRequest: PurchaseRequest;
  previousStatus: PurchaseRequestStatus;
  currentStatus: PurchaseRequestStatus;
  auditLogId?: string;
};

export type ConvertPurchaseRequestToPurchaseOrderInput = {
  supplierId: string;
  warehouseId: string;
  expectedDate: string;
  currencyCode?: string;
  unitPrice?: string;
};

export type ConvertPurchaseRequestToPurchaseOrderResult = {
  purchaseRequest: PurchaseRequest;
  purchaseOrder: PurchaseOrder;
  auditLogId?: string;
};
