import { apiGet, apiGetRaw, apiPost } from "../../../shared/api/client";
import { shouldUsePrototypeFallback } from "../../../shared/api/prototypeFallback";
import type { components, operations } from "../../../shared/api/generated/schema";
import { decimalScales, normalizeDecimalInput } from "../../../shared/format/numberFormat";
import type {
  FinanceSourceDocument,
  SupplierInvoice,
  SupplierPayable,
  SupplierPayableActionResult,
  SupplierPayableLine,
  SupplierPayableQuery,
  SupplierPayableStatus
} from "../types";

type SupplierPayableApi = components["schemas"]["SupplierPayable"];
type SupplierPayableListItemApi = components["schemas"]["SupplierPayableListItem"];
type SupplierPayableLineApi = components["schemas"]["SupplierPayableLine"];
type SupplierPayableActionApiRequest = components["schemas"]["SupplierPayableActionRequest"];
type SupplierPayableRejectPaymentApiRequest = components["schemas"]["SupplierPayableRejectPaymentRequest"];
type SupplierPayableActionApiResult = components["schemas"]["SupplierPayableActionResult"];
type SupplierPayableListApiQuery = operations["listSupplierPayables"]["parameters"]["query"];

const defaultAccessToken = "local-dev-access-token";
const prototypeNow = "2026-04-30T10:00:00Z";

let prototypeSupplierPayables = createPrototypeSupplierPayables();

export const supplierPayableStatusOptions: { label: string; value: "" | SupplierPayableStatus }[] = [
  { label: "All statuses", value: "" },
  { label: "Draft", value: "draft" },
  { label: "Open", value: "open" },
  { label: "Payment requested", value: "payment_requested" },
  { label: "Payment approved", value: "payment_approved" },
  { label: "Partially paid", value: "partially_paid" },
  { label: "Paid", value: "paid" },
  { label: "Disputed", value: "disputed" },
  { label: "Void", value: "void" }
];

export async function getSupplierPayables(query: SupplierPayableQuery = {}): Promise<SupplierPayable[]> {
  try {
    const payables = await apiGet("/supplier-payables", {
      accessToken: defaultAccessToken,
      query: toApiSupplierPayableQuery(query)
    });

    return payables.map(fromApiSupplierPayableListItem);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return filterPrototypeSupplierPayables(query);
  }
}

export async function getSupplierPayable(id: string): Promise<SupplierPayable> {
  try {
    const payable = await apiGetRaw<SupplierPayableApi>(`/supplier-payables/${encodeURIComponent(id)}`, {
      accessToken: defaultAccessToken
    });

    return fromApiSupplierPayable(payable);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return getPrototypeSupplierPayable(id);
  }
}

export async function approveSupplierPayablePayment(id: string): Promise<SupplierPayableActionResult> {
  return runSupplierPayableAction(id, "approve-payment", {});
}

export async function requestSupplierPayablePayment(id: string): Promise<SupplierPayableActionResult> {
  return runSupplierPayableAction(id, "request-payment", {});
}

export async function rejectSupplierPayablePayment(
  id: string,
  reason: string
): Promise<SupplierPayableActionResult> {
  return runSupplierPayableAction(id, "reject-payment", { reason: reason.trim() });
}

export async function recordSupplierPayablePayment(
  id: string,
  amount: string
): Promise<SupplierPayableActionResult> {
  return runSupplierPayableAction(id, "record-payment", {
    amount: normalizeDecimalInput(amount, decimalScales.money)
  });
}

export async function voidSupplierPayable(id: string, reason: string): Promise<SupplierPayableActionResult> {
  return runSupplierPayableAction(id, "void", { reason });
}

export function resetPrototypeSupplierPayablesForTest() {
  prototypeSupplierPayables = createPrototypeSupplierPayables();
}

export function supplierPayableStatusTone(
  status: SupplierPayableStatus
): "success" | "warning" | "danger" | "info" | "normal" {
  switch (status) {
    case "paid":
      return "success";
    case "payment_requested":
    case "payment_approved":
    case "partially_paid":
      return "info";
    case "disputed":
      return "warning";
    case "void":
      return "danger";
    case "draft":
      return "normal";
    case "open":
    default:
      return "warning";
  }
}

export function formatSupplierPayableStatus(status: SupplierPayableStatus) {
  return status
    .split("_")
    .map((part) => `${part.slice(0, 1).toUpperCase()}${part.slice(1)}`)
    .join(" ");
}

export function canApproveSupplierPayablePayment(payable: SupplierPayable | null) {
  return Boolean(payable && payable.status === "payment_requested");
}

export function canRequestSupplierPayablePayment(payable: SupplierPayable | null) {
  return Boolean(payable && payable.status === "open");
}

export type SupplierPayablePaymentReadiness = {
  canRequestPayment: boolean;
  label: string;
  message: string;
  tone: "success" | "warning" | "danger" | "info" | "normal";
};

export function getSupplierPayablePaymentReadiness(
  payable: SupplierPayable | null,
  invoices: SupplierInvoice[],
  loading: boolean
): SupplierPayablePaymentReadiness {
  if (!payable) {
    return {
      canRequestPayment: false,
      label: "Chưa chọn AP",
      message: "Chọn một AP để kiểm tra điều kiện thanh toán.",
      tone: "normal"
    };
  }
  if (loading) {
    return {
      canRequestPayment: false,
      label: "Đang kiểm tra hóa đơn",
      message: "Hệ thống đang tải hóa đơn NCC liên kết với AP này.",
      tone: "warning"
    };
  }
  const matchedInvoice = invoices.find((invoice) => supplierInvoiceMatchesPayable(invoice, payable));
  if (matchedInvoice) {
    return {
      canRequestPayment: canRequestSupplierPayablePayment(payable),
      label: "Sẵn sàng thanh toán",
      message: `${matchedInvoice.invoiceNo} đã khớp với ${payable.payableNo}.`,
      tone: "success"
    };
  }
  const blockingInvoice = invoices.find((invoice) => invoice.payableId === payable.id && invoice.status !== "void");
  if (blockingInvoice) {
    return {
      canRequestPayment: false,
      label: "Hóa đơn chưa khớp",
      message: `${blockingInvoice.invoiceNo} đang ${formatSupplierInvoiceGateStatus(blockingInvoice)}; xử lý lệch hoặc tranh chấp trước khi thanh toán.`,
      tone: "danger"
    };
  }

  return {
    canRequestPayment: false,
    label: "Cần hóa đơn NCC",
    message: "Tạo và đối chiếu hóa đơn NCC đến trạng thái đã khớp trước khi yêu cầu thanh toán.",
    tone: "warning"
  };
}

export function supplierInvoiceMatchesPayable(invoice: SupplierInvoice, payable: SupplierPayable) {
  return (
    invoice.payableId === payable.id &&
    invoice.supplierId === payable.supplierId &&
    invoice.currencyCode === payable.currencyCode &&
    invoice.expectedAmount === payable.totalAmount &&
    invoice.status === "matched" &&
    invoice.matchStatus === "matched" &&
    Number(invoice.varianceAmount) === 0
  );
}

function formatSupplierInvoiceGateStatus(invoice: SupplierInvoice) {
  if (invoice.status === "mismatch") {
    return "lệch đối chiếu";
  }
  if (invoice.status === "draft") {
    return "nháp";
  }

  return "chưa sẵn sàng";
}

export function canRejectSupplierPayablePayment(payable: SupplierPayable | null) {
  return Boolean(payable && payable.status === "payment_requested");
}

export function canRecordSupplierPayablePayment(payable: SupplierPayable | null) {
  return Boolean(
    payable &&
      Number(payable.outstandingAmount) > 0 &&
      ["payment_approved", "partially_paid"].includes(payable.status)
  );
}

export function canVoidSupplierPayable(payable: SupplierPayable | null) {
  return Boolean(payable && !["paid", "void"].includes(payable.status) && Number(payable.paidAmount) === 0);
}

async function runSupplierPayableAction(
  id: string,
  action: "request-payment" | "approve-payment" | "reject-payment" | "record-payment" | "void",
  body: SupplierPayableActionApiRequest | SupplierPayableRejectPaymentApiRequest
): Promise<SupplierPayableActionResult> {
  try {
    const result = await apiPost<
      SupplierPayableActionApiResult,
      SupplierPayableActionApiRequest | SupplierPayableRejectPaymentApiRequest
    >(
      `/supplier-payables/${encodeURIComponent(id)}/${action}`,
      body,
      { accessToken: defaultAccessToken }
    );

    return fromApiActionResult(result);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return transitionPrototypeSupplierPayable(id, action, body);
  }
}


function toApiSupplierPayableQuery(query: SupplierPayableQuery): SupplierPayableListApiQuery {
  return {
    q: query.search,
    status: query.status,
    supplier_id: query.supplierId
  };
}

function fromApiSupplierPayableListItem(item: SupplierPayableListItemApi): SupplierPayable {
  return {
    id: item.id,
    payableNo: item.payable_no,
    supplierId: item.supplier_id,
    supplierCode: item.supplier_code,
    supplierName: item.supplier_name,
    status: item.status,
    lines: [],
    totalAmount: item.total_amount,
    paidAmount: item.paid_amount,
    outstandingAmount: item.outstanding_amount,
    currencyCode: item.currency_code,
    dueDate: item.due_date,
    createdAt: item.created_at,
    updatedAt: item.updated_at,
    version: item.version
  };
}

function fromApiSupplierPayable(payable: SupplierPayableApi): SupplierPayable {
  return {
    id: payable.id,
    orgId: payable.org_id,
    payableNo: payable.payable_no,
    supplierId: payable.supplier_id,
    supplierCode: payable.supplier_code,
    supplierName: payable.supplier_name,
    status: payable.status,
    sourceDocument: fromApiSourceDocument(payable.source_document),
    lines: payable.lines.map(fromApiSupplierPayableLine),
    totalAmount: payable.total_amount,
    paidAmount: payable.paid_amount,
    outstandingAmount: payable.outstanding_amount,
    currencyCode: payable.currency_code,
    dueDate: payable.due_date,
    paymentRequestedBy: payable.payment_requested_by,
    paymentRequestedAt: payable.payment_requested_at,
    paymentApprovedBy: payable.payment_approved_by,
    paymentApprovedAt: payable.payment_approved_at,
    paymentRejectedBy: payable.payment_rejected_by,
    paymentRejectedAt: payable.payment_rejected_at,
    paymentRejectReason: payable.payment_reject_reason,
    disputeReason: payable.dispute_reason,
    voidReason: payable.void_reason,
    auditLogId: payable.audit_log_id,
    createdAt: payable.created_at,
    updatedAt: payable.updated_at,
    version: payable.version
  };
}

function fromApiSupplierPayableLine(line: SupplierPayableLineApi): SupplierPayableLine {
  return {
    id: line.id,
    description: line.description,
    sourceDocument: fromApiSourceDocument(line.source_document),
    amount: line.amount
  };
}

function fromApiSourceDocument(source: components["schemas"]["FinanceSourceDocument"]): FinanceSourceDocument {
  return {
    type: source.type,
    id: source.id,
    no: source.no
  };
}

function fromApiActionResult(result: SupplierPayableActionApiResult): SupplierPayableActionResult {
  return {
    supplierPayable: fromApiSupplierPayable(result.supplier_payable),
    previousStatus: result.previous_status,
    currentStatus: result.current_status,
    auditLogId: result.audit_log_id
  };
}

function filterPrototypeSupplierPayables(query: SupplierPayableQuery) {
  return prototypeSupplierPayables
    .filter((payable) => matchesSupplierPayableQuery(payable, query))
    .map(cloneSupplierPayable);
}

function getPrototypeSupplierPayable(id: string) {
  const payable = prototypeSupplierPayables.find((candidate) => candidate.id === id);
  if (!payable) {
    throw new Error("Supplier payable not found");
  }

  return cloneSupplierPayable(payable);
}

function transitionPrototypeSupplierPayable(
  id: string,
  action: "request-payment" | "approve-payment" | "reject-payment" | "record-payment" | "void",
  body: SupplierPayableActionApiRequest | SupplierPayableRejectPaymentApiRequest
): SupplierPayableActionResult {
  const current = getPrototypeSupplierPayable(id);
  const previousStatus = current.status;
  let next = cloneSupplierPayable(current);

  if (action === "request-payment") {
    if (!canRequestSupplierPayablePayment(next)) {
      throw new Error("Only open AP can request payment");
    }
    next = {
      ...next,
      status: "payment_requested",
      paymentRequestedBy: "finance-user",
      paymentRequestedAt: prototypeNow,
      paymentRejectedBy: undefined,
      paymentRejectedAt: undefined,
      paymentRejectReason: undefined
    };
  } else if (action === "approve-payment") {
    if (!canApproveSupplierPayablePayment(next)) {
      throw new Error("Only requested AP can be approved");
    }
    next = {
      ...next,
      status: "payment_approved",
      paymentApprovedBy: "finance-manager",
      paymentApprovedAt: prototypeNow
    };
  } else if (action === "reject-payment") {
    const reason = "reason" in body ? body.reason?.trim() : "";
    if (!canRejectSupplierPayablePayment(next) || !reason) {
      throw new Error("Only requested AP can be rejected with a reason");
    }
    next = {
      ...next,
      status: "open",
      paymentApprovedBy: undefined,
      paymentApprovedAt: undefined,
      paymentRejectedBy: "finance-manager",
      paymentRejectedAt: prototypeNow,
      paymentRejectReason: reason
    };
  } else if (action === "record-payment") {
    const amount = normalizeDecimalInput("amount" in body ? body.amount : undefined, decimalScales.money);
    const amountValue = Number(amount);
    if (!canRecordSupplierPayablePayment(next) || amountValue <= 0 || amountValue > Number(next.outstandingAmount)) {
      throw new Error("Payment amount is invalid for this AP");
    }

    const paidAmount = Number(next.paidAmount) + amountValue;
    const outstandingAmount = Math.max(0, Number(next.totalAmount) - paidAmount);
    next = {
      ...next,
      paidAmount: normalizeDecimalInput(paidAmount, decimalScales.money),
      outstandingAmount: normalizeDecimalInput(outstandingAmount, decimalScales.money),
      status: outstandingAmount === 0 ? "paid" : "partially_paid"
    };
  } else {
    if (!canVoidSupplierPayable(next)) {
      throw new Error("Paid, void, or partially paid AP cannot be voided");
    }
    next = {
      ...next,
      status: "void",
      outstandingAmount: "0.00",
      voidReason: body.reason?.trim() || "Voided by finance"
    };
  }

  next = {
    ...next,
    updatedAt: prototypeNow,
    version: next.version + 1,
    auditLogId: `audit-ap-${action}-${next.id}`
  };
  prototypeSupplierPayables = [next, ...prototypeSupplierPayables.filter((payable) => payable.id !== next.id)];

  return {
    supplierPayable: cloneSupplierPayable(next),
    previousStatus,
    currentStatus: next.status,
    auditLogId: next.auditLogId
  };
}

function matchesSupplierPayableQuery(payable: SupplierPayable, query: SupplierPayableQuery) {
  const search = query.search?.trim().toLowerCase();
  return (
    (!query.status || payable.status === query.status) &&
    (!query.supplierId || payable.supplierId === query.supplierId) &&
    (!search ||
      [payable.payableNo, payable.supplierCode, payable.supplierName, payable.sourceDocument?.no]
        .filter(Boolean)
        .some((value) => String(value).toLowerCase().includes(search)))
  );
}

function createPrototypeSupplierPayables(): SupplierPayable[] {
  return [
    createSupplierPayableSeed({
      id: "ap-supplier-260430-0001",
      payableNo: "AP-SUP-260430-0001",
      supplierId: "supplier-hcm-001",
      supplierCode: "SUP-HCM-001",
      supplierName: "Nguyen Lieu HCM",
      status: "open",
      totalAmount: "4250000.00",
      paidAmount: "0.00",
      sourceDocument: { type: "qc_inspection", id: "qc-inbound-260430-0001", no: "QC-260430-0001" },
      dueDate: "2026-05-07",
      lines: [
        {
          id: "ap-supplier-260430-0001-line-1",
          description: "Accepted raw material after inbound QC",
          sourceDocument: { type: "warehouse_receipt", id: "gr-260430-0001", no: "GR-260430-0001" },
          amount: "3000000.00"
        },
        {
          id: "ap-supplier-260430-0001-line-2",
          description: "Accepted packaging after inbound QC",
          sourceDocument: { type: "purchase_order", id: "po-260430-0001", no: "PO-260430-0001" },
          amount: "1250000.00"
        }
      ]
    }),
    createSupplierPayableSeed({
      id: "ap-subcontract-260430-0002",
      payableNo: "AP-SUB-260430-0002",
      supplierId: "factory-bd-002",
      supplierCode: "FACT-BD-002",
      supplierName: "Binh Duong Gia Cong",
      status: "payment_approved",
      totalAmount: "6800000.00",
      paidAmount: "0.00",
      sourceDocument: { type: "subcontract_payment_milestone", id: "sco-pay-260430-0002", no: "SCO-PAY-260430-0002" },
      dueDate: "2026-05-03",
      paymentApprovedBy: "finance-manager",
      paymentApprovedAt: prototypeNow,
      lines: [
        {
          id: "ap-subcontract-260430-0002-line-1",
          description: "Final subcontract manufacturing milestone",
          sourceDocument: { type: "subcontract_order", id: "sco-260430-0002", no: "SCO-260430-0002" },
          amount: "6800000.00"
        }
      ]
    })
  ];
}

function createSupplierPayableSeed(input: {
  id: string;
  payableNo: string;
  supplierId: string;
  supplierCode: string;
  supplierName: string;
  status: SupplierPayableStatus;
  totalAmount: string;
  paidAmount: string;
  sourceDocument: FinanceSourceDocument;
  dueDate: string;
  paymentApprovedBy?: string;
  paymentApprovedAt?: string;
  paymentRejectedBy?: string;
  paymentRejectedAt?: string;
  paymentRejectReason?: string;
  lines: SupplierPayableLine[];
}): SupplierPayable {
  return {
    id: input.id,
    orgId: "org-my-pham",
    payableNo: input.payableNo,
    supplierId: input.supplierId,
    supplierCode: input.supplierCode,
    supplierName: input.supplierName,
    status: input.status,
    sourceDocument: input.sourceDocument,
    lines: input.lines,
    totalAmount: normalizeDecimalInput(input.totalAmount, decimalScales.money),
    paidAmount: normalizeDecimalInput(input.paidAmount, decimalScales.money),
    outstandingAmount: normalizeDecimalInput(Number(input.totalAmount) - Number(input.paidAmount), decimalScales.money),
    currencyCode: "VND",
    dueDate: input.dueDate,
    paymentApprovedBy: input.paymentApprovedBy,
    paymentApprovedAt: input.paymentApprovedAt,
    paymentRejectedBy: input.paymentRejectedBy,
    paymentRejectedAt: input.paymentRejectedAt,
    paymentRejectReason: input.paymentRejectReason,
    createdAt: prototypeNow,
    updatedAt: prototypeNow,
    version: 1
  };
}

function cloneSupplierPayable(payable: SupplierPayable): SupplierPayable {
  return {
    ...payable,
    sourceDocument: payable.sourceDocument ? { ...payable.sourceDocument } : undefined,
    lines: payable.lines.map((line) => ({
      ...line,
      sourceDocument: { ...line.sourceDocument }
    }))
  };
}
