export type CustomerReceivableStatus = "draft" | "open" | "partially_paid" | "paid" | "disputed" | "void";
export type SupplierPayableStatus =
  | "draft"
  | "open"
  | "payment_requested"
  | "payment_approved"
  | "partially_paid"
  | "paid"
  | "disputed"
  | "void";
export type CODRemittanceStatus = "draft" | "matching" | "submitted" | "approved" | "discrepancy" | "closed" | "void";
export type CODLineMatchStatus = "matched" | "short_paid" | "over_paid";
export type CODDiscrepancyType = "short_paid" | "over_paid" | "carrier_fee" | "return_claim" | "other";
export type CODDiscrepancyStatus = "open" | "resolved";
export type CashTransactionStatus = "draft" | "posted" | "void";
export type CashTransactionDirection = "cash_in" | "cash_out";
export type CashAllocationTargetType =
  | "customer_receivable"
  | "supplier_payable"
  | "cod_remittance"
  | "payment_request"
  | "manual_adjustment";

export type FinanceSourceDocumentType =
  | "sales_order"
  | "shipment"
  | "return_order"
  | "purchase_order"
  | "warehouse_receipt"
  | "qc_inspection"
  | "supplier_payable"
  | "subcontract_order"
  | "subcontract_payment_milestone"
  | "cod_remittance"
  | "manual_adjustment";

export type FinanceSourceDocument = {
  type: FinanceSourceDocumentType;
  id?: string;
  no?: string;
};

export type CustomerReceivableSourceDocument = {
  type: Exclude<FinanceSourceDocumentType, "supplier_payable">;
  id?: string;
  no?: string;
};

export type CustomerReceivableLine = {
  id: string;
  description: string;
  sourceDocument: CustomerReceivableSourceDocument;
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
  sourceDocument?: CustomerReceivableSourceDocument;
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
  sourceDocument: CustomerReceivableSourceDocument;
  amount: string;
};

export type CreateCustomerReceivableInput = {
  id?: string;
  receivableNo?: string;
  customerId: string;
  customerCode?: string;
  customerName: string;
  status?: CustomerReceivableStatus;
  sourceDocument: CustomerReceivableSourceDocument;
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

export type SupplierPayableLine = {
  id: string;
  description: string;
  sourceDocument: FinanceSourceDocument;
  amount: string;
};

export type SupplierPayable = {
  id: string;
  orgId?: string;
  payableNo: string;
  supplierId: string;
  supplierCode?: string;
  supplierName: string;
  status: SupplierPayableStatus;
  sourceDocument?: FinanceSourceDocument;
  lines: SupplierPayableLine[];
  totalAmount: string;
  paidAmount: string;
  outstandingAmount: string;
  currencyCode: string;
  dueDate?: string;
  paymentRequestedBy?: string;
  paymentRequestedAt?: string;
  paymentApprovedBy?: string;
  paymentApprovedAt?: string;
  paymentRejectedBy?: string;
  paymentRejectedAt?: string;
  paymentRejectReason?: string;
  disputeReason?: string;
  voidReason?: string;
  auditLogId?: string;
  createdAt: string;
  updatedAt: string;
  version: number;
};

export type SupplierPayableQuery = {
  search?: string;
  status?: SupplierPayableStatus;
  supplierId?: string;
};

export type SupplierPayableActionResult = {
  supplierPayable: SupplierPayable;
  previousStatus: SupplierPayableStatus;
  currentStatus: SupplierPayableStatus;
  auditLogId?: string;
};

export type SupplierInvoiceStatus = "draft" | "matched" | "mismatch" | "void";
export type SupplierInvoiceMatchStatus = "pending" | "matched" | "mismatch";

export type SupplierInvoiceLine = {
  id: string;
  description: string;
  sourceDocument: FinanceSourceDocument;
  amount: string;
};

export type SupplierInvoice = {
  id: string;
  orgId?: string;
  invoiceNo: string;
  supplierId: string;
  supplierCode?: string;
  supplierName: string;
  payableId: string;
  payableNo: string;
  status: SupplierInvoiceStatus;
  matchStatus: SupplierInvoiceMatchStatus;
  sourceDocument: FinanceSourceDocument;
  lines: SupplierInvoiceLine[];
  invoiceAmount: string;
  expectedAmount: string;
  varianceAmount: string;
  currencyCode: string;
  invoiceDate: string;
  voidReason?: string;
  auditLogId?: string;
  createdAt: string;
  updatedAt: string;
  version: number;
};

export type SupplierInvoiceQuery = {
  search?: string;
  status?: SupplierInvoiceStatus;
  supplierId?: string;
  payableId?: string;
};

export type CreateSupplierInvoiceInput = {
  id?: string;
  invoiceNo?: string;
  supplierId?: string;
  supplierCode?: string;
  supplierName?: string;
  payableId: string;
  invoiceDate?: string;
  invoiceAmount: string;
  currencyCode: string;
};

export type SupplierInvoiceActionResult = {
  supplierInvoice: SupplierInvoice;
  previousStatus: SupplierInvoiceStatus;
  currentStatus: SupplierInvoiceStatus;
  auditLogId?: string;
};

export type CashTransactionAllocation = {
  id: string;
  targetType: CashAllocationTargetType;
  targetId: string;
  targetNo: string;
  amount: string;
};

export type CashTransaction = {
  id: string;
  orgId?: string;
  transactionNo: string;
  direction: CashTransactionDirection;
  status: CashTransactionStatus;
  businessDate: string;
  counterpartyId?: string;
  counterpartyName: string;
  paymentMethod: string;
  referenceNo?: string;
  allocations: CashTransactionAllocation[];
  totalAmount: string;
  currencyCode: string;
  memo?: string;
  postedBy?: string;
  postedAt?: string;
  voidReason?: string;
  voidedBy?: string;
  voidedAt?: string;
  auditLogId?: string;
  createdAt: string;
  updatedAt: string;
  version: number;
};

export type CashTransactionQuery = {
  search?: string;
  status?: CashTransactionStatus;
  direction?: CashTransactionDirection;
  counterpartyId?: string;
};

export type CreateCashTransactionInput = {
  id?: string;
  transactionNo?: string;
  direction: CashTransactionDirection;
  businessDate: string;
  counterpartyId?: string;
  counterpartyName: string;
  paymentMethod: string;
  referenceNo?: string;
  allocations: CashTransactionAllocation[];
  totalAmount: string;
  currencyCode: string;
  memo?: string;
};

export type FinanceDashboardReceivableMetrics = {
  openCount: number;
  overdueCount: number;
  disputedCount: number;
  openAmount: string;
  overdueAmount: string;
  outstandingAmount: string;
};

export type FinanceDashboardPayableMetrics = {
  openCount: number;
  dueCount: number;
  paymentRequestedCount: number;
  paymentApprovedCount: number;
  openAmount: string;
  dueAmount: string;
  outstandingAmount: string;
};

export type FinanceDashboardCODMetrics = {
  pendingCount: number;
  discrepancyCount: number;
  pendingAmount: string;
  discrepancyAmount: string;
};

export type FinanceDashboardCashMetrics = {
  transactionCount: number;
  cashInToday: string;
  cashOutToday: string;
  netCashToday: string;
};

export type FinanceDashboard = {
  businessDate: string;
  generatedAt: string;
  currencyCode: string;
  ar: FinanceDashboardReceivableMetrics;
  ap: FinanceDashboardPayableMetrics;
  cod: FinanceDashboardCODMetrics;
  cash: FinanceDashboardCashMetrics;
};

export type FinanceDashboardQuery = {
  businessDate?: string;
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
