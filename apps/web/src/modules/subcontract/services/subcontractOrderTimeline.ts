import type { StatusTone } from "../../../shared/design-system/components";
import type { SubcontractOrder, SubcontractOrderStatus } from "../types";

export type SubcontractOrderTimelineStatus = "complete" | "current" | "pending" | "blocked";

export type SubcontractOrderTimelineItem = {
  id: string;
  label: string;
  description: string;
  status: SubcontractOrderTimelineStatus;
  tone: StatusTone;
  occurredAt?: string;
  action?: {
    label: string;
    href: string;
    disabled?: boolean;
  };
};

type TimelineStep = {
  id: string;
  label: string;
  description: (order: SubcontractOrder) => string;
  completeAt: number;
  currentAt: number[];
  action?: (order: SubcontractOrder, status: SubcontractOrderTimelineStatus) => SubcontractOrderTimelineItem["action"];
};

export function buildSubcontractOrderTimeline(order: SubcontractOrder): SubcontractOrderTimelineItem[] {
  const index = statusProgressIndex(order.status);
  const terminal = terminalStatus(order.status);

  return timelineSteps.map((step) => {
    const status = timelineStepStatus(step, index, terminal);

    return {
      id: step.id,
      label: step.label,
      description: step.description(order),
      status,
      tone: toneForTimelineStatus(status),
      action: step.action?.(order, status)
    };
  });
}

export function productionFactoryOrderHref(order: Pick<SubcontractOrder, "id">) {
  return `/production/factory-orders/${encodeURIComponent(order.id)}`;
}

export function productionFactoryOrderSourcePlanHref(order: Pick<SubcontractOrder, "sourceProductionPlanId">) {
  return order.sourceProductionPlanId ? `/production/plans/${encodeURIComponent(order.sourceProductionPlanId)}` : undefined;
}

export function subcontractOperationsHref(order: Pick<SubcontractOrder, "sourceProductionPlanId" | "sourceProductionPlanNo" | "orderNo">, hash = "subcontract-orders") {
  const params = new URLSearchParams();
  if (order.sourceProductionPlanId) {
    params.set("source_production_plan_id", order.sourceProductionPlanId);
  }
  params.set("search", order.sourceProductionPlanNo || order.orderNo);

  return `/subcontract?${params.toString()}#${hash}`;
}

const timelineSteps: TimelineStep[] = [
  {
    id: "created",
    label: "Tạo lệnh",
    completeAt: 0,
    currentAt: [],
    description: (order) => `${order.orderNo} được tạo cho ${order.factoryName}.`
  },
  {
    id: "submitted",
    label: "Gửi duyệt",
    completeAt: 1,
    currentAt: [0],
    description: () => "Lệnh nhà máy chờ người phụ trách gửi duyệt hoặc quản lý duyệt."
  },
  {
    id: "approved",
    label: "Duyệt lệnh",
    completeAt: 2,
    currentAt: [1],
    description: () => "Sau khi duyệt, công ty có thể gửi lệnh cho nhà máy ngoài xác nhận."
  },
  {
    id: "factory-confirmed",
    label: "Nhà máy xác nhận",
    completeAt: 3,
    currentAt: [2],
    description: (order) => `${order.factoryName} xác nhận số lượng, quy cách, mẫu mã và lịch giao.`,
    action: (order, status) => ({
      label: "Mở xử lý lệnh",
      href: subcontractOperationsHref(order),
      disabled: status === "pending" || status === "blocked"
    })
  },
  {
    id: "deposit",
    label: "Ghi nhận cọc",
    completeAt: 4,
    currentAt: [3],
    description: (order) =>
      order.depositStatus === "not_required" ? "Lệnh này không yêu cầu đặt cọc." : "Ghi nhận cọc trước khi xuất vật tư hoặc chạy sản xuất."
  },
  {
    id: "materials-issued",
    label: "Xuất vật tư",
    completeAt: 5,
    currentAt: [4],
    description: () => "Xuất nguyên liệu/bao bì cho nhà máy và lưu bằng chứng bàn giao.",
    action: (order, status) => ({
      label: "Mở xuất vật tư",
      href: subcontractOperationsHref(order, "subcontract-transfer"),
      disabled: status === "pending" || status === "blocked"
    })
  },
  {
    id: "sample",
    label: "Duyệt mẫu",
    completeAt: 6,
    currentAt: [5],
    description: (order) =>
      order.sampleRequired ? "Gửi mẫu, duyệt mẫu hoặc ghi lý do từ chối trước khi sản xuất hàng loạt." : "Không yêu cầu duyệt mẫu cho lệnh này."
  },
  {
    id: "mass-production",
    label: "Sản xuất hàng loạt",
    completeAt: 7,
    currentAt: [6],
    description: () => "Nhà máy chạy sản xuất hàng loạt sau khi đủ vật tư và mẫu đạt."
  },
  {
    id: "finished-goods-received",
    label: "Nhận thành phẩm",
    completeAt: 8,
    currentAt: [7],
    description: (order) => `Nhận thành phẩm từ ${order.factoryName} về khu QC hold.`,
    action: (order, status) => ({
      label: "Mở nhận thành phẩm",
      href: subcontractOperationsHref(order, "subcontract-inbound"),
      disabled: status === "pending" || status === "blocked"
    })
  },
  {
    id: "qc",
    label: "QC thành phẩm",
    completeAt: 9,
    currentAt: [8],
    description: () => "Chỉ thành phẩm đạt QC mới được nhập vào tồn khả dụng; lỗi mở claim nhà máy."
  },
  {
    id: "final-payment",
    label: "Thanh toán cuối",
    completeAt: 10,
    currentAt: [9],
    description: () => "Chỉ mở thanh toán cuối khi QC đạt hoặc claim đã có ngoại lệ được duyệt.",
    action: (order, status) => ({
      label: "Mở thanh toán",
      href: subcontractOperationsHref(order, "subcontract-payment"),
      disabled: status === "pending" || status === "blocked"
    })
  },
  {
    id: "closed",
    label: "Đóng lệnh",
    completeAt: 11,
    currentAt: [10],
    description: () => "Đóng lệnh khi nhận hàng, QC, claim và thanh toán cuối đã xong."
  }
];

function timelineStepStatus(
  step: Pick<TimelineStep, "completeAt" | "currentAt">,
  currentIndex: number,
  terminal?: SubcontractOrderStatus
): SubcontractOrderTimelineStatus {
  if (terminal) {
    return currentIndex >= step.completeAt ? "complete" : "blocked";
  }
  if (currentIndex >= step.completeAt) {
    return "complete";
  }
  if (step.currentAt.includes(currentIndex)) {
    return "current";
  }

  return "pending";
}

function statusProgressIndex(status: SubcontractOrderStatus) {
  switch (status) {
    case "draft":
      return 0;
    case "submitted":
      return 1;
    case "approved":
      return 2;
    case "factory_confirmed":
      return 3;
    case "deposit_recorded":
      return 4;
    case "materials_issued_to_factory":
      return 5;
    case "sample_submitted":
    case "sample_rejected":
      return 5;
    case "sample_approved":
      return 6;
    case "mass_production_started":
      return 7;
    case "finished_goods_received":
      return 8;
    case "qc_in_progress":
      return 8;
    case "accepted":
      return 9;
    case "final_payment_ready":
      return 10;
    case "closed":
      return 11;
    case "rejected_with_factory_issue":
    case "cancelled":
    default:
      return 0;
  }
}

function terminalStatus(status: SubcontractOrderStatus) {
  return ["cancelled", "rejected_with_factory_issue"].includes(status) ? status : undefined;
}

function toneForTimelineStatus(status: SubcontractOrderTimelineStatus): StatusTone {
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
