import type { StatusTone } from "@/shared/design-system/components";
import type { SubcontractFactoryDispatchStatus, SubcontractOrder, SubcontractOrderStatus } from "../types";
import { subcontractOperationsHref } from "./subcontractOrderTimeline";

export type FactoryExecutionWorkStatus = "complete" | "current" | "pending" | "blocked";

export type FactoryExecutionTrackerContext = {
  dispatchStatus?: SubcontractFactoryDispatchStatus;
};

export type FactoryExecutionWorkItem = {
  id: string;
  title: string;
  description: string;
  status: FactoryExecutionWorkStatus;
  tone: StatusTone;
  metric: string;
  action?: {
    label: string;
    href: string;
    disabled: boolean;
  };
};

export type FactoryExecutionTracker = {
  currentGate: FactoryExecutionWorkItem;
  items: FactoryExecutionWorkItem[];
};

export function buildSubcontractFactoryExecutionTracker(
  order: SubcontractOrder,
  context: FactoryExecutionTrackerContext = {}
): FactoryExecutionTracker {
  if (order.status === "cancelled") {
    const cancelled = workItem({
      id: "cancelled",
      title: "Lệnh đã hủy",
      description: "Lệnh nhà máy đã bị hủy; không tiếp tục gửi, bàn giao vật tư, sản xuất, nhận hàng hoặc thanh toán.",
      status: "blocked",
      metric: "Đã hủy"
    });

    return {
      currentGate: cancelled,
      items: [cancelled]
    };
  }

  const items = [
    buildDispatchItem(order, context.dispatchStatus),
    buildFactoryConfirmationItem(order, context.dispatchStatus),
    buildDepositItem(order),
    buildMaterialHandoverItem(order),
    buildSampleGateItem(order),
    buildMassProductionItem(order),
    buildFinishedGoodsReceiptItem(order),
    buildQcCloseoutItem(order),
    buildFinalPaymentItem(order)
  ];

  return {
    currentGate:
      items.find((item) => item.status === "blocked") ??
      items.find((item) => item.status === "current") ??
      items.find((item) => item.status === "pending") ??
      items[items.length - 1],
    items
  };
}

function buildDispatchItem(order: SubcontractOrder, dispatchStatus?: SubcontractFactoryDispatchStatus): FactoryExecutionWorkItem {
  const status =
    isAtLeast(order.status, "factory_confirmed") || dispatchStatus === "confirmed"
      ? "complete"
      : dispatchStatus === "rejected" || dispatchStatus === "cancelled"
        ? "blocked"
        : order.status === "approved" || dispatchStatus === "draft" || dispatchStatus === "ready" || dispatchStatus === "sent"
          ? "current"
          : "pending";

  return workItem({
    id: "factory-dispatch",
    title: "Gửi lệnh cho nhà máy",
    description: "Tạo bộ gửi, ghi nhận sẵn sàng, đã gửi thủ công và phản hồi từ nhà máy.",
    status,
    metric: dispatchStatus ? formatDispatchStatus(dispatchStatus) : "Chưa có bộ gửi",
    action: {
      label: "Mở gửi nhà máy",
      href: productionFactoryOrderSectionHref(order, "factory-dispatch"),
      disabled: status === "pending"
    }
  });
}

function buildFactoryConfirmationItem(order: SubcontractOrder, dispatchStatus?: SubcontractFactoryDispatchStatus): FactoryExecutionWorkItem {
  const status =
    isAtLeast(order.status, "factory_confirmed")
      ? "complete"
      : dispatchStatus === "rejected"
        ? "blocked"
        : dispatchStatus === "sent"
          ? "current"
          : "pending";

  return workItem({
    id: "factory-confirmation",
    title: "Nhà máy xác nhận",
    description: "Nhà máy xác nhận số lượng, quy cách, mẫu mã và lịch giao trước khi theo dõi thực thi.",
    status,
    metric: isAtLeast(order.status, "factory_confirmed") ? "Đã xác nhận" : "Chờ phản hồi",
    action: {
      label: "Mở phản hồi",
      href: productionFactoryOrderSectionHref(order, "factory-dispatch"),
      disabled: status === "pending"
    }
  });
}

function buildDepositItem(order: SubcontractOrder): FactoryExecutionWorkItem {
  const depositComplete = depositSatisfied(order);
  const status = depositComplete ? "complete" : isAtLeast(order.status, "factory_confirmed") ? "current" : "pending";

  return workItem({
    id: "deposit",
    title: "Đặt cọc / điều kiện thanh toán",
    description: "Ghi nhận đặt cọc nếu nhà máy yêu cầu trước khi bàn giao vật tư hoặc chạy sản xuất.",
    status,
    metric: order.depositStatus === "not_required" ? "Không yêu cầu cọc" : order.depositStatus === "paid" ? "Đã ghi nhận cọc" : "Chờ ghi nhận",
    action: {
      label: "Mở thanh toán",
      href: subcontractOperationsHref(order, "subcontract-payment"),
      disabled: status === "pending"
    }
  });
}

function buildMaterialHandoverItem(order: SubcontractOrder): FactoryExecutionWorkItem {
  const summary = materialIssueSummary(order);
  const canIssueMaterials = depositSatisfied(order) && isAtLeast(order.status, "factory_confirmed");
  const status = summary.complete
    ? "complete"
    : canIssueMaterials
      ? "current"
      : "pending";

  return workItem({
    id: "material-handover",
    title: "Bàn giao vật tư cho nhà máy",
    description: "Xuất nguyên liệu/bao bì theo lệnh, kèm bằng chứng bàn giao và lô nếu có kiểm soát.",
    status,
    metric: `${summary.completeLines}/${summary.totalLines} dòng đủ`,
    action: {
      label: "Mở xuất vật tư",
      href: productionFactoryOrderSectionHref(order, "factory-material-handover"),
      disabled: status === "pending"
    }
  });
}

function buildSampleGateItem(order: SubcontractOrder): FactoryExecutionWorkItem {
  if (!order.sampleRequired) {
    return workItem({
      id: "sample-gate",
      title: "Duyệt mẫu",
      description: "Lệnh này không yêu cầu duyệt mẫu trước khi chạy hàng loạt.",
      status: "complete",
      metric: "Không yêu cầu mẫu"
    });
  }

  const status =
    order.status === "sample_rejected"
      ? "blocked"
      : isAtLeast(order.status, "sample_approved")
        ? "complete"
        : isAtLeast(order.status, "materials_issued_to_factory") || order.status === "sample_submitted"
          ? "current"
          : "pending";

  return workItem({
    id: "sample-gate",
    title: "Duyệt mẫu",
    description: "Gửi mẫu, duyệt mẫu hoặc ghi nhận lý do từ chối trước khi sản xuất hàng loạt.",
    status,
    metric: formatSampleMetric(order.status),
    action: {
      label: "Mở duyệt mẫu",
      href: productionFactoryOrderSectionHref(order, "factory-sample-approval"),
      disabled: status === "pending"
    }
  });
}

function buildMassProductionItem(order: SubcontractOrder): FactoryExecutionWorkItem {
  const status =
    order.status === "sample_rejected"
      ? "blocked"
      : isAtLeast(order.status, "mass_production_started")
        ? "complete"
        : canStartMassProduction(order)
          ? "current"
          : "pending";

  return workItem({
    id: "mass-production",
    title: "Chạy sản xuất hàng loạt",
    description: "Nhà máy bắt đầu sản xuất hàng loạt sau khi đủ vật tư và mẫu đạt hoặc không cần mẫu.",
    status,
    metric: isAtLeast(order.status, "mass_production_started") ? "Đã bắt đầu" : "Chưa bắt đầu",
    action: {
      label: "Mở sản xuất hàng loạt",
      href: productionFactoryOrderSectionHref(order, "factory-mass-production"),
      disabled: status === "pending" || status === "blocked"
    }
  });
}

function buildFinishedGoodsReceiptItem(order: SubcontractOrder): FactoryExecutionWorkItem {
  const receivedQty = numeric(order.receivedQty);
  const status = receivedQty > 0 || isAtLeast(order.status, "finished_goods_received") ? "complete" : isAtLeast(order.status, "mass_production_started") ? "current" : "pending";

  return workItem({
    id: "finished-goods-receipt",
    title: "Nhận thành phẩm về QC hold",
    description: "Nhận thành phẩm từ nhà máy về khu QC hold; chưa nhập tồn khả dụng trước khi QC đạt.",
    status,
    metric: `${formatQty(order.receivedQty)} / ${formatQty(String(order.quantity))} ${order.uomCode ?? "PCS"}`,
    action: {
      label: "Mở nhận hàng",
      href: subcontractOperationsHref(order, "subcontract-inbound"),
      disabled: status === "pending"
    }
  });
}

function buildQcCloseoutItem(order: SubcontractOrder): FactoryExecutionWorkItem {
  const status =
    order.status === "rejected_with_factory_issue"
      ? "blocked"
      : ["accepted", "final_payment_ready", "closed"].includes(order.status)
        ? "complete"
        : ["finished_goods_received", "qc_in_progress"].includes(order.status)
          ? "current"
          : "pending";

  return workItem({
    id: "qc-closeout",
    title: "QC thành phẩm / claim nhà máy",
    description: "QC đạt mới vào tồn khả dụng; QC lỗi phải mở claim nhà máy trước khi đóng thanh toán.",
    status,
    metric: `Đạt ${formatQty(order.acceptedQty)} / lỗi ${formatQty(order.rejectedQty)} ${order.uomCode ?? "PCS"}`,
    action: {
      label: status === "blocked" ? "Mở claim" : "Mở QC",
      href: subcontractOperationsHref(order, status === "blocked" ? "subcontract-claim" : "subcontract-inbound"),
      disabled: status === "pending"
    }
  });
}

function buildFinalPaymentItem(order: SubcontractOrder): FactoryExecutionWorkItem {
  const status =
    ["final_payment_ready", "closed"].includes(order.status)
      ? "complete"
      : order.status === "rejected_with_factory_issue"
        ? "blocked"
        : order.status === "accepted"
          ? "current"
          : "pending";

  return workItem({
    id: "final-payment",
    title: "Sẵn sàng thanh toán cuối",
    description: "Chỉ mở thanh toán cuối khi thành phẩm đạt QC hoặc claim đã có ngoại lệ được duyệt.",
    status,
    metric: formatFinalPaymentMetric(order),
    action: {
      label: "Mở thanh toán",
      href: subcontractOperationsHref(order, "subcontract-payment"),
      disabled: status === "pending" || status === "blocked"
    }
  });
}

function workItem(input: Omit<FactoryExecutionWorkItem, "tone">): FactoryExecutionWorkItem {
  return {
    ...input,
    tone: toneForStatus(input.status)
  };
}

function materialIssueSummary(order: SubcontractOrder) {
  const totalLines = order.materialLines.length;
  const completeLines = order.materialLines.filter((line) => numeric(line.issuedQty) + 0.000001 >= numeric(line.plannedQty)).length;

  return {
    totalLines,
    completeLines,
    complete: totalLines > 0 && completeLines === totalLines
  };
}

function depositSatisfied(order: SubcontractOrder) {
  return order.depositStatus === "not_required" || order.depositStatus === "paid" || isAtLeast(order.status, "deposit_recorded");
}

function canStartMassProduction(order: SubcontractOrder) {
  const materialComplete = materialIssueSummary(order).complete || isAtLeast(order.status, "materials_issued_to_factory");
  const sampleComplete = !order.sampleRequired || isAtLeast(order.status, "sample_approved");

  return materialComplete && sampleComplete;
}

function isAtLeast(status: SubcontractOrderStatus, target: SubcontractOrderStatus) {
  return statusRank(status) >= statusRank(target);
}

function statusRank(status: SubcontractOrderStatus) {
  const ranks: Record<SubcontractOrderStatus, number> = {
    draft: 0,
    submitted: 1,
    approved: 2,
    factory_confirmed: 3,
    deposit_recorded: 4,
    materials_issued_to_factory: 5,
    sample_submitted: 6,
    sample_rejected: 6,
    sample_approved: 7,
    mass_production_started: 8,
    finished_goods_received: 9,
    qc_in_progress: 10,
    accepted: 11,
    rejected_with_factory_issue: 11,
    final_payment_ready: 12,
    closed: 13,
    cancelled: -1
  };

  return ranks[status];
}

function toneForStatus(status: FactoryExecutionWorkStatus): StatusTone {
  switch (status) {
    case "complete":
      return "success";
    case "current":
      return "info";
    case "blocked":
      return "danger";
    case "pending":
    default:
      return "normal";
  }
}

function productionFactoryOrderSectionHref(order: Pick<SubcontractOrder, "id">, sectionId: string) {
  return `/production/factory-orders/${encodeURIComponent(order.id)}#${sectionId}`;
}

function numeric(value?: string) {
  if (!value) {
    return 0;
  }

  return Number.parseFloat(value.replace(",", ".")) || 0;
}

function formatQty(value?: string) {
  return numeric(value).toLocaleString("vi-VN", { maximumFractionDigits: 6 });
}

function formatDispatchStatus(status: SubcontractFactoryDispatchStatus) {
  switch (status) {
    case "draft":
      return "Nháp";
    case "ready":
      return "Sẵn sàng gửi";
    case "sent":
      return "Đã gửi";
    case "confirmed":
      return "Nhà máy xác nhận";
    case "revision_requested":
      return "Cần chỉnh";
    case "rejected":
      return "Bị từ chối";
    case "cancelled":
      return "Đã hủy";
    default:
      return status;
  }
}

function formatSampleMetric(status: SubcontractOrderStatus) {
  if (status === "sample_rejected") {
    return "Mẫu bị từ chối";
  }
  if (isAtLeast(status, "sample_approved")) {
    return "Mẫu đã duyệt";
  }
  if (status === "sample_submitted") {
    return "Đã gửi mẫu";
  }

  return "Cần duyệt mẫu";
}

function formatFinalPaymentMetric(order: SubcontractOrder) {
  if (order.status === "closed") {
    return "Đã đóng lệnh";
  }
  if (order.status === "final_payment_ready") {
    return "Sẵn sàng thanh toán";
  }
  if (order.status === "rejected_with_factory_issue") {
    return "Đang giữ vì claim";
  }

  return order.finalPaymentStatus === "hold" ? "Đang giữ" : "Chờ xử lý";
}
