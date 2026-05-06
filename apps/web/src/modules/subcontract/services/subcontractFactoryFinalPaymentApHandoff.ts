import type { SubcontractPaymentMilestoneResult } from "../types";

export type FactoryFinalPaymentApHandoffStatus = "ready" | "pending";

export type FactoryFinalPaymentApHandoff = {
  status: FactoryFinalPaymentApHandoffStatus;
  payableId?: string;
  payableNo?: string;
  financeHref?: string;
  message: string;
};

export function buildFactoryFinalPaymentApHandoff(
  result?: Pick<SubcontractPaymentMilestoneResult, "supplierPayable">
): FactoryFinalPaymentApHandoff {
  const payable = result?.supplierPayable;
  if (!payable || !payable.payableNo.trim()) {
    return {
      status: "pending",
      message: "Thanh toán cuối đã sẵn sàng nhưng chưa có AP handoff để mở Finance."
    };
  }

  return {
    status: "ready",
    payableId: payable.payableId,
    payableNo: payable.payableNo,
    financeHref: buildFactoryFinalPaymentApHandoffHref(payable.payableNo),
    message: `${payable.payableNo} đã được tạo để Finance đối chiếu hóa đơn và thanh toán.`
  };
}

export function buildFactoryFinalPaymentApHandoffFromSource(sourceNo: string): FactoryFinalPaymentApHandoff {
  const trimmedSourceNo = sourceNo.trim();

  return {
    status: "pending",
    financeHref: trimmedSourceNo ? buildFactoryFinalPaymentApHandoffHref(trimmedSourceNo) : undefined,
    message: trimmedSourceNo
      ? `Mở Finance theo lệnh ${trimmedSourceNo} để kiểm tra AP đã tạo.`
      : "Thanh toán cuối đã sẵn sàng; mở Finance để kiểm tra AP đã tạo."
  };
}

export function buildFactoryFinalPaymentApHandoffHref(payableNoOrId: string) {
  const params = new URLSearchParams({ ap_q: payableNoOrId.trim() });

  return `/finance?${params.toString()}#supplier-payables`;
}
