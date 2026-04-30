export type CustomerReceivableStatus = "draft" | "open" | "partially_paid" | "paid" | "disputed" | "void";

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
