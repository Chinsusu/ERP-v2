export type SalesOrderStatus =
  | "draft"
  | "confirmed"
  | "reserved"
  | "picking"
  | "picked"
  | "packing"
  | "packed"
  | "waiting_handover"
  | "handed_over"
  | "delivered"
  | "returned"
  | "closed"
  | "cancelled"
  | "reservation_failed"
  | "pick_exception"
  | "pack_exception"
  | "handover_exception";

export type SalesOrderLine = {
  id: string;
  lineNo: number;
  itemId: string;
  skuCode: string;
  itemName: string;
  orderedQty: string;
  uomCode: string;
  baseOrderedQty: string;
  baseUomCode: string;
  conversionFactor?: string;
  unitPrice: string;
  currencyCode: string;
  lineDiscountAmount: string;
  lineAmount: string;
  reservedQty: string;
  shippedQty: string;
  batchId?: string;
  batchNo?: string;
};

export type SalesOrder = {
  id: string;
  orderNo: string;
  customerId: string;
  customerCode?: string;
  customerName: string;
  channel: string;
  warehouseId?: string;
  warehouseCode?: string;
  orderDate: string;
  status: SalesOrderStatus;
  currencyCode: string;
  subtotalAmount: string;
  discountAmount: string;
  taxAmount: string;
  shippingFeeAmount: string;
  netAmount: string;
  totalAmount: string;
  note?: string;
  lineCount?: number;
  lines: SalesOrderLine[];
  auditLogId?: string;
  createdAt: string;
  updatedAt: string;
  confirmedAt?: string;
  cancelledAt?: string;
  cancelReason?: string;
  version: number;
};

export type SalesOrderQuery = {
  search?: string;
  status?: SalesOrderStatus;
  channel?: string;
  customerId?: string;
  warehouseId?: string;
};

export type SalesOrderLineInput = {
  id?: string;
  lineNo?: number;
  itemId: string;
  orderedQty: string;
  uomCode: string;
  unitPrice: string;
  currencyCode?: string;
  lineDiscountAmount?: string;
};

export type CreateSalesOrderInput = {
  id?: string;
  orderNo?: string;
  customerId: string;
  channel: string;
  warehouseId?: string;
  orderDate: string;
  currencyCode: string;
  note?: string;
  lines: SalesOrderLineInput[];
};

export type UpdateSalesOrderInput = {
  customerId?: string;
  channel?: string;
  warehouseId?: string;
  orderDate?: string;
  note?: string;
  expectedVersion?: number;
  lines?: SalesOrderLineInput[];
};

export type SalesOrderActionResult = {
  salesOrder: SalesOrder;
  previousStatus: SalesOrderStatus;
  currentStatus: SalesOrderStatus;
  auditLogId?: string;
};
