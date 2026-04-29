export type InboundQCInspectionStatus = "pending" | "in_progress" | "completed" | "cancelled";

export type InboundQCResult = "pass" | "fail" | "hold" | "partial";

export type InboundQCChecklistStatus = "pending" | "pass" | "fail" | "not_applicable";

export type InboundQCChecklistItem = {
  id: string;
  code: string;
  label: string;
  required: boolean;
  status: InboundQCChecklistStatus;
  note?: string;
};

export type InboundQCInspection = {
  id: string;
  orgId: string;
  goodsReceiptId: string;
  goodsReceiptNo: string;
  goodsReceiptLineId: string;
  purchaseOrderId?: string;
  purchaseOrderLineId?: string;
  itemId: string;
  sku: string;
  itemName?: string;
  batchId: string;
  batchNo: string;
  lotNo: string;
  expiryDate: string;
  warehouseId: string;
  locationId: string;
  quantity: string;
  uomCode: string;
  inspectorId: string;
  status: InboundQCInspectionStatus;
  result?: InboundQCResult;
  passedQuantity: string;
  failedQuantity: string;
  holdQuantity: string;
  checklist: InboundQCChecklistItem[];
  reason?: string;
  note?: string;
  auditLogId?: string;
  createdAt: string;
  createdBy: string;
  updatedAt: string;
  updatedBy: string;
  startedAt?: string;
  startedBy?: string;
  decidedAt?: string;
  decidedBy?: string;
};

export type InboundQCInspectionQuery = {
  status?: InboundQCInspectionStatus;
  goodsReceiptId?: string;
  goodsReceiptLineId?: string;
  warehouseId?: string;
};

export type CreateInboundQCInspectionInput = {
  id?: string;
  orgId?: string;
  goodsReceiptId: string;
  goodsReceiptLineId: string;
  inspectorId?: string;
  checklist?: InboundQCChecklistItem[];
  note?: string;
};

export type InboundQCDecisionInput = {
  passedQuantity?: string;
  failedQuantity?: string;
  holdQuantity?: string;
  checklist?: InboundQCChecklistItem[];
  reason?: string;
  note?: string;
};

export type InboundQCActionResult = {
  inspection: InboundQCInspection;
  previousStatus?: InboundQCInspectionStatus;
  currentStatus: InboundQCInspectionStatus;
  previousResult?: InboundQCResult;
  currentResult?: InboundQCResult;
  auditLogId?: string;
};
