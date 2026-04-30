export type CustomerReceivableStatus = "draft" | "open" | "partially_paid" | "paid" | "disputed" | "void";
export type CODRemittanceStatus = "draft" | "matching" | "submitted" | "approved" | "discrepancy" | "closed" | "void";
export type CODLineMatchStatus = "matched" | "short_paid" | "over_paid";
export type CODDiscrepancyType = "short_paid" | "over_paid" | "carrier_fee" | "return_claim" | "other";
export type CODDiscrepancyStatus = "open" | "resolved";

export type FinanceSourceDocumentType =
  | "sales_order"
  | "shipment"
  | "return_order"
  | "purchase_order"
  | "warehouse_receipt"
  | "qc_inspection"
  | "subcontract_order"
  | "subcontract_payment_milestone"
  | "cod_remittance"
  | "manual_adjustment";

export type FinanceSourceDocument = {
  type: FinanceSourceDocumentType;
  id?: string;
  no?: string;
};

export type CustomerReceivableLine = {
  id: string;
  description: string;
  sourceDocument: FinanceSourceDocument;
  amount: string;
};

export type CustomerReceivable = {
  id: string;
  orgId?: string;
  receivableNo: string;
  customerId: string;
  customerCode?: string;
  customerName: string;
  status: CustomerReceivableStatus;
  sourceDocument?: FinanceSourceDocument;
  lines: CustomerReceivableLine[];
  totalAmount: string;
  paidAmount: string;
  outstandingAmount: string;
  currencyCode: string;
  dueDate?: string;
  disputeReason?: string;
  voidReason?: string;
  auditLogId?: string;
  createdAt: string;
  updatedAt: string;
  version: number;
};

export type CustomerReceivableQuery = {
  search?: string;
  status?: CustomerReceivableStatus;
  customerId?: string;
};

export type CreateCustomerReceivableLineInput = {
  id: string;
  description: string;
  sourceDocument: FinanceSourceDocument;
  amount: string;
};

export type CreateCustomerReceivableInput = {
  id?: string;
  receivableNo?: string;
  customerId: string;
  customerCode?: string;
  customerName: string;
  status?: CustomerReceivableStatus;
  sourceDocument: FinanceSourceDocument;
  lines: CreateCustomerReceivableLineInput[];
  totalAmount: string;
  currencyCode: string;
  dueDate?: string;
};

export type CustomerReceivableActionResult = {
  customerReceivable: CustomerReceivable;
  previousStatus: CustomerReceivableStatus;
  currentStatus: CustomerReceivableStatus;
  auditLogId?: string;
};

export type CODRemittanceLine = {
  id: string;
  receivableId: string;
  receivableNo: string;
  shipmentId?: string;
  trackingNo: string;
  customerName?: string;
  expectedAmount: string;
  remittedAmount: string;
  discrepancyAmount: string;
  matchStatus: CODLineMatchStatus;
};

export type CODDiscrepancy = {
  id: string;
  lineId: string;
  receivableId: string;
  type: CODDiscrepancyType;
  status: CODDiscrepancyStatus;
  amount: string;
  reason: string;
  ownerId: string;
  recordedBy: string;
  recordedAt: string;
  resolvedBy?: string;
  resolvedAt?: string;
  resolution?: string;
};

export type CODRemittance = {
  id: string;
  orgId?: string;
  remittanceNo: string;
  carrierId: string;
  carrierCode?: string;
  carrierName: string;
  status: CODRemittanceStatus;
  businessDate: string;
  expectedAmount: string;
  remittedAmount: string;
  discrepancyAmount: string;
  currencyCode: string;
  lines: CODRemittanceLine[];
  discrepancies: CODDiscrepancy[];
  lineCount: number;
  discrepancyCount: number;
  auditLogId?: string;
  submittedBy?: string;
  submittedAt?: string;
  approvedBy?: string;
  approvedAt?: string;
  closedBy?: string;
  closedAt?: string;
  createdAt: string;
  updatedAt: string;
  version: number;
};

export type CODRemittanceQuery = {
  search?: string;
  status?: CODRemittanceStatus;
  carrierId?: string;
};

export type CODRemittanceDiscrepancyInput = {
  id?: string;
  lineId: string;
  type?: CODDiscrepancyType;
  status?: CODDiscrepancyStatus;
  reason: string;
  ownerId: string;
};

export type CODRemittanceActionResult = {
  codRemittance: CODRemittance;
  previousStatus: CODRemittanceStatus;
  currentStatus: CODRemittanceStatus;
  auditLogId?: string;
};
