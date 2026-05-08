import type { StatusTone } from "../../../shared/design-system/components";
import type { CashTransaction, FinanceSourceDocument, SupplierInvoice, SupplierPayable } from "../types";
import { supplierInvoiceMatchesPayable } from "./supplierPayableService";

export type SupplierPayableFactoryCloseoutStepKey =
  | "ap-created"
  | "invoice-match"
  | "payment-request"
  | "payment-approval"
  | "payment-recording"
  | "payment-voucher";

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
  voucherNo?: string;
  voucherReferenceNo?: string;
  voucherHref?: string;
  createVoucherHref?: string;
  summaryLabel: string;
  summaryMessage: string;
  summaryTone: StatusTone;
  steps: SupplierPayableFactoryCloseoutStep[];
};

export function buildSupplierPayableFactoryCloseout(
  payable: SupplierPayable | null,
  invoices: SupplierInvoice[],
  invoicesLoading: boolean,
  cashTransactions: CashTransaction[] = [],
  cashTransactionsLoading = false
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
  const paymentVoucher = findSupplierPayableCashOutVoucher(payable, cashTransactions);
  const invoiceStepStatus = invoiceMatched ? "complete" : "current";
  const paymentRequestStatus = paymentRequestStepStatus(invoiceMatched, paymentStage);
  const paymentApprovalStatus = paymentApprovalStepStatus(invoiceMatched, paymentStage);
  const paymentRecordingStatus = paymentRecordingStepStatus(invoiceMatched, paymentStage);
  const paymentVoucherStatus = paymentVoucherStepStatus(invoiceMatched, paymentStage, paymentVoucher, cashTransactionsLoading);
  const summary = factoryCloseoutSummary(
    payable,
    matchedInvoice,
    blockingInvoice,
    invoicesLoading,
    paymentVoucher,
    paymentVoucherStatus
  );

  return {
    payableNo: payable.payableNo,
    supplierName: payable.supplierName,
    factoryOrderNo: factoryOrder?.no,
    factoryMilestoneNo: factoryMilestone?.no,
    productionHref: factoryOrderHref(factoryOrder),
    invoiceNo: matchedInvoice?.invoiceNo,
    voucherNo: paymentVoucher?.transactionNo,
    voucherReferenceNo: paymentVoucher?.referenceNo,
    voucherHref: paymentVoucherHref(paymentVoucher),
    createVoucherHref:
      !paymentVoucher && !cashTransactionsLoading && paymentVoucherStatus === "current"
        ? buildSupplierPayableVoucherHref(payable, matchedInvoice, factoryOrder, factoryMilestone)
        : undefined,
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
      },
      {
        key: "payment-voucher",
        label: "Chứng từ chi",
        description: paymentVoucherStepDescription(paymentVoucher, paymentVoucherStatus, cashTransactionsLoading),
        status: paymentVoucherStatus,
        tone: stepTone(paymentVoucherStatus)
      }
    ]
  };
}

export function findSupplierPayableCashOutVoucher(
  payable: SupplierPayable,
  transactions: CashTransaction[]
): CashTransaction | undefined {
  const payableID = payable.id.trim().toLowerCase();
  const payableNo = payable.payableNo.trim().toLowerCase();

  return transactions.find((transaction) => {
    if (transaction.direction !== "cash_out" || transaction.status !== "posted") {
      return false;
    }

    return transaction.allocations.some((allocation) => {
      if (allocation.targetType !== "supplier_payable") {
        return false;
      }
      return allocation.targetId.trim().toLowerCase() === payableID || allocation.targetNo.trim().toLowerCase() === payableNo;
    });
  });
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

function paymentVoucherStepStatus(
  invoiceMatched: boolean,
  paymentStage: number,
  voucher: CashTransaction | undefined,
  cashTransactionsLoading: boolean
): SupplierPayableFactoryCloseoutStepStatus {
  if (!invoiceMatched || paymentStage < 4) {
    return "pending";
  }
  if (voucher) {
    return "complete";
  }
  if (cashTransactionsLoading) {
    return "current";
  }

  return "current";
}

function factoryCloseoutSummary(
  payable: SupplierPayable,
  matchedInvoice: SupplierInvoice | undefined,
  blockingInvoice: SupplierInvoice | undefined,
  invoicesLoading: boolean,
  paymentVoucher: CashTransaction | undefined,
  paymentVoucherStatus: SupplierPayableFactoryCloseoutStepStatus
) {
  if (payable.status === "paid" && paymentVoucher) {
    return {
      label: "Đã có chứng từ chi",
      message: `${paymentVoucher.transactionNo} đã ghi nhận chứng từ chi cho ${payable.payableNo}.`,
      tone: "success" as StatusTone
    };
  }
  if (payable.status === "paid" && paymentVoucherStatus === "current") {
    return {
      label: "Cần chứng từ chi",
      message: `${payable.payableNo} đã ghi nhận thanh toán AP; Finance cần post chứng từ chi để đủ evidence cash/bank.`,
      tone: "warning" as StatusTone
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

function paymentVoucherStepDescription(
  voucher: CashTransaction | undefined,
  status: SupplierPayableFactoryCloseoutStepStatus,
  cashTransactionsLoading: boolean
) {
  if (voucher) {
    return `${voucher.transactionNo} / ${voucher.referenceNo ?? "chưa có số tham chiếu"} đã allocate vào AP.`;
  }
  if (cashTransactionsLoading) {
    return "Đang kiểm tra chứng từ chi cash_out đã allocate vào AP.";
  }
  if (status === "current") {
    return "Tạo hoặc mở chứng từ chi cash_out allocate vào AP để hoàn tất evidence thanh toán.";
  }

  return "Chờ AP được ghi nhận thanh toán trước khi chứng từ chi được xem là bắt buộc.";
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

function paymentVoucherHref(voucher: CashTransaction | undefined) {
  if (!voucher) {
    return undefined;
  }
  return `/finance?cash_q=${encodeURIComponent(voucher.transactionNo)}#cash-transactions`;
}

function buildSupplierPayableVoucherHref(
  payable: SupplierPayable,
  invoice: SupplierInvoice | undefined,
  factoryOrder: FinanceSourceDocument | undefined,
  factoryMilestone: FinanceSourceDocument | undefined
) {
  const params = new URLSearchParams({
    cash_q: payable.payableNo,
    cash_direction: "cash_out",
    cash_counterparty_id: payable.supplierId,
    cash_counterparty_name: payable.supplierName,
    cash_payment_method: "bank_transfer",
    cash_reference_no: `PAY-${payable.payableNo}`,
    cash_amount: voucherAmount(payable),
    cash_target_type: "supplier_payable",
    cash_target_id: payable.id,
    cash_target_no: payable.payableNo,
    cash_memo: voucherMemo(payable, invoice, factoryOrder, factoryMilestone)
  });

  return `/finance?${params.toString()}#cash-transactions`;
}

function voucherAmount(payable: SupplierPayable) {
  if (Number(payable.paidAmount) > 0) {
    return payable.paidAmount;
  }

  return payable.outstandingAmount;
}

function voucherMemo(
  payable: SupplierPayable,
  invoice: SupplierInvoice | undefined,
  factoryOrder: FinanceSourceDocument | undefined,
  factoryMilestone: FinanceSourceDocument | undefined
) {
  return [
    `Thanh toan AP ${payable.payableNo}`,
    factoryOrder?.no ? `cho lenh ${factoryOrder.no}` : "",
    invoice?.invoiceNo ? `hoa don ${invoice.invoiceNo}` : "",
    factoryMilestone?.no ? `moc ${factoryMilestone.no}` : ""
  ]
    .filter(Boolean)
    .join("; ");
}
