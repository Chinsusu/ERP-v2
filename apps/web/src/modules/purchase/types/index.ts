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
