import type { StatusTone } from "../../../shared/design-system/components";
import type { FinanceSourceDocument, SupplierInvoice, SupplierPayable } from "../types";
import { supplierInvoiceMatchesPayable } from "./supplierPayableService";

export type SupplierPayableFactoryCloseoutStepKey =
  | "ap-created"
  | "invoice-match"
  | "payment-request"
  | "payment-approval"
  | "payment-recording";

export type SupplierPayableFactoryCloseoutStepStatus = "complete" | "current" | "pending" | "blocked";

export type SupplierPayableFactoryCloseoutStep = {
  key: SupplierPayableFactoryCloseoutStepKey;
  label: string;
  description: string;
  status: SupplierPayableFactoryCloseoutStepStatus;
  tone: StatusTone;
};

export type SupplierPayableFactoryCloseout = {
  payableNo: string;
  supplierName: string;
  factoryOrderNo?: string;
  factoryMilestoneNo?: string;
  productionHref?: string;
  invoiceNo?: string;
  summaryLabel: string;
  summaryMessage: string;
  summaryTone: StatusTone;
  steps: SupplierPayableFactoryCloseoutStep[];
};

export function buildSupplierPayableFactoryCloseout(
  payable: SupplierPayable | null,
  invoices: SupplierInvoice[],
  invoicesLoading: boolean
): SupplierPayableFactoryCloseout | null {
  if (!payable) {
    return null;
  }

  const sourceDocuments = payableSourceDocuments(payable);
  const factoryOrder = sourceDocuments.find((source) => source.type === "subcontract_order");
  const factoryMilestone = sourceDocuments.find((source) => source.type === "subcontract_payment_milestone");

  if (!factoryOrder && !factoryMilestone) {
    return null;
  }

  const matchedInvoice = invoices.find((invoice) => supplierInvoiceMatchesPayable(invoice, payable));
  const blockingInvoice = invoices.find((invoice) => invoice.payableId === payable.id && invoice.status !== "void");
  const paymentStage = supplierPayablePaymentStage(payable.status);
  const invoiceMatched = Boolean(matchedInvoice);
  const invoiceStepStatus = invoiceMatched ? "complete" : "current";
  const paymentRequestStatus = paymentRequestStepStatus(invoiceMatched, paymentStage);
  const paymentApprovalStatus = paymentApprovalStepStatus(invoiceMatched, paymentStage);
  const paymentRecordingStatus = paymentRecordingStepStatus(invoiceMatched, paymentStage);
  const summary = factoryCloseoutSummary(payable, matchedInvoice, blockingInvoice, invoicesLoading);

  return {
    payableNo: payable.payableNo,
    supplierName: payable.supplierName,
    factoryOrderNo: factoryOrder?.no,
    factoryMilestoneNo: factoryMilestone?.no,
    productionHref: factoryOrderHref(factoryOrder),
    invoiceNo: matchedInvoice?.invoiceNo,
    summaryLabel: summary.label,
    summaryMessage: summary.message,
    summaryTone: summary.tone,
    steps: [
      {
        key: "ap-created",
        label: "AP thanh toán cuối",
        description: `${payable.payableNo} đã được tạo từ ${factoryMilestone?.no ?? factoryOrder?.no ?? "lệnh nhà máy"}.`,
        status: "complete",
        tone: "success"
      },
      {
        key: "invoice-match",
        label: "Hóa đơn NCC",
        description: invoiceStepDescription(matchedInvoice, blockingInvoice, invoicesLoading),
        status: invoiceStepStatus,
        tone: stepTone(invoiceStepStatus, blockingInvoice && !matchedInvoice ? "danger" : undefined)
      },
      {
        key: "payment-request",
        label: "Yêu cầu thanh toán",
        description: invoiceMatched
          ? "Finance có thể gửi yêu cầu thanh toán cho AP đã khớp hóa đơn."
          : "Bị chặn đến khi hóa đơn NCC khớp AP.",
        status: paymentRequestStatus,
        tone: stepTone(paymentRequestStatus)
      },
      {
        key: "payment-approval",
        label: "Duyệt thanh toán",
        description: "Người duyệt Finance kiểm tra AP, hóa đơn, nguồn lệnh nhà máy và duyệt thanh toán.",
        status: paymentApprovalStatus,
        tone: stepTone(paymentApprovalStatus)
      },
      {
        key: "payment-recording",
        label: "Ghi nhận chi tiền",
        description: "Thủ quỹ/Finance ghi nhận chi tiền để đóng AP thanh toán cuối.",
        status: paymentRecordingStatus,
        tone: stepTone(paymentRecordingStatus)
      }
    ]
  };
}

function payableSourceDocuments(payable: SupplierPayable) {
  return [payable.sourceDocument, ...payable.lines.map((line) => line.sourceDocument)].filter(
    (source): source is FinanceSourceDocument => Boolean(source)
  );
}

function supplierPayablePaymentStage(status: SupplierPayable["status"]) {
  if (status === "paid") {
    return 4;
  }
  if (status === "payment_approved" || status === "partially_paid") {
    return 3;
  }
  if (status === "payment_requested") {
    return 2;
  }
  if (status === "open") {
    return 1;
  }

  return 0;
}

function paymentRequestStepStatus(invoiceMatched: boolean, paymentStage: number): SupplierPayableFactoryCloseoutStepStatus {
  if (!invoiceMatched) {
    return "blocked";
  }
  if (paymentStage >= 2) {
    return "complete";
  }
  return paymentStage === 1 ? "current" : "pending";
}

function paymentApprovalStepStatus(invoiceMatched: boolean, paymentStage: number): SupplierPayableFactoryCloseoutStepStatus {
  if (!invoiceMatched || paymentStage < 2) {
    return "pending";
  }
  return paymentStage >= 3 ? "complete" : "current";
}

function paymentRecordingStepStatus(invoiceMatched: boolean, paymentStage: number): SupplierPayableFactoryCloseoutStepStatus {
  if (!invoiceMatched || paymentStage < 3) {
    return "pending";
  }
  return paymentStage >= 4 ? "complete" : "current";
}

function factoryCloseoutSummary(
  payable: SupplierPayable,
  matchedInvoice: SupplierInvoice | undefined,
  blockingInvoice: SupplierInvoice | undefined,
  invoicesLoading: boolean
) {
  if (payable.status === "paid") {
    return {
      label: "Đã đóng AP",
      message: `${payable.payableNo} đã ghi nhận thanh toán cuối.`,
      tone: "success" as StatusTone
    };
  }
  if (matchedInvoice) {
    return {
      label: "Sẵn sàng thanh toán",
      message: `${matchedInvoice.invoiceNo} đã khớp với ${payable.payableNo}; Finance có thể tiếp tục yêu cầu, duyệt và ghi nhận thanh toán.`,
      tone: "success" as StatusTone
    };
  }
  if (blockingInvoice) {
    return {
      label: "Hóa đơn chưa khớp",
      message: `${blockingInvoice.invoiceNo} chưa khớp với ${payable.payableNo}; chưa được thanh toán.`,
      tone: "danger" as StatusTone
    };
  }
  if (invoicesLoading) {
    return {
      label: "Đang tải hóa đơn",
      message: `Đang tải hóa đơn NCC liên kết với ${payable.payableNo}.`,
      tone: "warning" as StatusTone
    };
  }

  return {
    label: "Cần hóa đơn NCC",
    message: `${payable.payableNo} cần hóa đơn NCC khớp trước khi yêu cầu thanh toán.`,
    tone: "warning" as StatusTone
  };
}

function invoiceStepDescription(
  matchedInvoice: SupplierInvoice | undefined,
  blockingInvoice: SupplierInvoice | undefined,
  invoicesLoading: boolean
) {
  if (matchedInvoice) {
    return `${matchedInvoice.invoiceNo} đã khớp AP, nhà cung cấp, tiền tệ và số tiền.`;
  }
  if (blockingInvoice) {
    return `${blockingInvoice.invoiceNo} đang ${blockingInvoice.status}; xử lý chênh lệch trước khi thanh toán.`;
  }
  if (invoicesLoading) {
    return "Đang tải hóa đơn NCC để đối chiếu.";
  }

  return "Tạo hóa đơn NCC từ AP và đối chiếu đến trạng thái đã khớp.";
}

function stepTone(status: SupplierPayableFactoryCloseoutStepStatus, override?: StatusTone): StatusTone {
  if (override) {
    return override;
  }
  switch (status) {
    case "complete":
      return "success";
    case "current":
      return "info";
    case "blocked":
      return "warning";
    case "pending":
    default:
      return "normal";
  }
}

function factoryOrderHref(factoryOrder: FinanceSourceDocument | undefined) {
  const orderId = factoryOrder?.id?.trim();
  if (!orderId) {
    return undefined;
  }

  return `/production/factory-orders/${encodeURIComponent(orderId)}#factory-claim-final-payment-closeout`;
}
