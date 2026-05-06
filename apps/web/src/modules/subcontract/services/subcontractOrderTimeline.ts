import type { StatusTone } from "../../../shared/design-system/components";
import type { SubcontractOrder, SubcontractOrderStatus } from "../types";

export type SubcontractOrderTimelineStatus = "complete" | "current" | "pending" | "blocked";
export type FactoryDispatchTimelineStatus = "draft" | "ready" | "sent" | "confirmed" | "revision_requested" | "rejected" | "cancelled";

export type SubcontractOrderTimelineContext = {
  dispatchStatus?: FactoryDispatchTimelineStatus;
};

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

export function buildSubcontractOrderTimeline(order: SubcontractOrder, context: SubcontractOrderTimelineContext = {}): SubcontractOrderTimelineItem[] {
  const index = statusProgressIndex(order.status, context.dispatchStatus);
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
    label: "T\u1ea1o l\u1ec7nh",
    completeAt: 0,
    currentAt: [],
    description: (order) => `${order.orderNo} \u0111\u01b0\u1ee3c t\u1ea1o cho ${order.factoryName}.`
  },
  {
    id: "submitted",
    label: "G\u1eedi duy\u1ec7t",
    completeAt: 1,
    currentAt: [0],
    description: () => "L\u1ec7nh nh\u00e0 m\u00e1y ch\u1edd ng\u01b0\u1eddi ph\u1ee5 tr\u00e1ch g\u1eedi duy\u1ec7t ho\u1eb7c qu\u1ea3n l\u00fd duy\u1ec7t."
  },
  {
    id: "approved",
    label: "Duy\u1ec7t l\u1ec7nh",
    completeAt: 2,
    currentAt: [1],
    description: () => "Sau khi duy\u1ec7t, c\u00f4ng ty c\u00f3 th\u1ec3 chu\u1ea9n b\u1ecb b\u1ed9 g\u1eedi nh\u00e0 m\u00e1y."
  },
  {
    id: "factory-dispatch",
    label: "G\u1eedi nh\u00e0 m\u00e1y",
    completeAt: 3,
    currentAt: [2],
    description: () => "T\u1ea1o b\u1ed9 g\u1eedi nh\u00e0 m\u00e1y, ghi nh\u1eadn \u0111\u00e3 g\u1eedi th\u1ee7 c\u00f4ng v\u00e0 l\u01b0u b\u1eb1ng ch\u1ee9ng g\u1eedi."
  },
  {
    id: "factory-confirmed",
    label: "Nh\u00e0 m\u00e1y x\u00e1c nh\u1eadn",
    completeAt: 4,
    currentAt: [3],
    description: (order) => `${order.factoryName} x\u00e1c nh\u1eadn s\u1ed1 l\u01b0\u1ee3ng, quy c\u00e1ch, m\u1eabu m\u00e3 v\u00e0 l\u1ecbch giao.`,
    action: (order, status) => ({
      label: "M\u1edf x\u1eed l\u00fd l\u1ec7nh",
      href: subcontractOperationsHref(order),
      disabled: status === "pending" || status === "blocked"
    })
  },
  {
    id: "deposit",
    label: "Ghi nh\u1eadn c\u1ecdc",
    completeAt: 5,
    currentAt: [4],
    description: (order) =>
      order.depositStatus === "not_required" ? "L\u1ec7nh n\u00e0y kh\u00f4ng y\u00eau c\u1ea7u \u0111\u1eb7t c\u1ecdc." : "Ghi nh\u1eadn c\u1ecdc tr\u01b0\u1edbc khi xu\u1ea5t v\u1eadt t\u01b0 ho\u1eb7c ch\u1ea1y s\u1ea3n xu\u1ea5t."
  },
  {
    id: "materials-issued",
    label: "Xu\u1ea5t v\u1eadt t\u01b0",
    completeAt: 6,
    currentAt: [5],
    description: () => "Xu\u1ea5t nguy\u00ean li\u1ec7u/bao b\u00ec cho nh\u00e0 m\u00e1y v\u00e0 l\u01b0u b\u1eb1ng ch\u1ee9ng b\u00e0n giao.",
    action: (order, status) => ({
      label: "M\u1edf xu\u1ea5t v\u1eadt t\u01b0",
      href: `${productionFactoryOrderHref(order)}#factory-material-handover`,
      disabled: status === "pending" || status === "blocked"
    })
  },
  {
    id: "sample",
    label: "Duy\u1ec7t m\u1eabu",
    completeAt: 7,
    currentAt: [6],
    description: (order) =>
      order.sampleRequired ? "G\u1eedi m\u1eabu, duy\u1ec7t m\u1eabu ho\u1eb7c ghi l\u00fd do t\u1eeb ch\u1ed1i tr\u01b0\u1edbc khi s\u1ea3n xu\u1ea5t h\u00e0ng lo\u1ea1t." : "Kh\u00f4ng y\u00eau c\u1ea7u duy\u1ec7t m\u1eabu cho l\u1ec7nh n\u00e0y.",
    action: (order, status) => ({
      label: "M\u1edf duy\u1ec7t m\u1eabu",
      href: `${productionFactoryOrderHref(order)}#factory-sample-approval`,
      disabled: status === "pending" || status === "blocked"
    })
  },
  {
    id: "mass-production",
    label: "S\u1ea3n xu\u1ea5t h\u00e0ng lo\u1ea1t",
    completeAt: 8,
    currentAt: [7],
    description: () => "Nh\u00e0 m\u00e1y ch\u1ea1y s\u1ea3n xu\u1ea5t h\u00e0ng lo\u1ea1t sau khi \u0111\u1ee7 v\u1eadt t\u01b0 v\u00e0 m\u1eabu \u0111\u1ea1t.",
    action: (order, status) => ({
      label: "M\u1edf s\u1ea3n xu\u1ea5t h\u00e0ng lo\u1ea1t",
      href: `${productionFactoryOrderHref(order)}#factory-mass-production`,
      disabled: status === "pending" || status === "blocked"
    })
  },
  {
    id: "finished-goods-received",
    label: "Nh\u1eadn th\u00e0nh ph\u1ea9m",
    completeAt: 9,
    currentAt: [8],
    description: (order) => `Nh\u1eadn th\u00e0nh ph\u1ea9m t\u1eeb ${order.factoryName} v\u1ec1 khu QC hold.`,
    action: (order, status) => ({
      label: "M\u1edf nh\u1eadn th\u00e0nh ph\u1ea9m",
      href: `${productionFactoryOrderHref(order)}#factory-finished-goods-receipt`,
      disabled: status === "pending" || status === "blocked"
    })
  },
  {
    id: "qc",
    label: "QC th\u00e0nh ph\u1ea9m",
    completeAt: 10,
    currentAt: [9],
    description: () => "Ch\u1ec9 th\u00e0nh ph\u1ea9m \u0111\u1ea1t QC m\u1edbi \u0111\u01b0\u1ee3c nh\u1eadp v\u00e0o t\u1ed3n kh\u1ea3 d\u1ee5ng; l\u1ed7i m\u1edf claim nh\u00e0 m\u00e1y."
  },
  {
    id: "final-payment",
    label: "Thanh to\u00e1n cu\u1ed1i",
    completeAt: 11,
    currentAt: [10],
    description: () => "Ch\u1ec9 m\u1edf thanh to\u00e1n cu\u1ed1i khi QC \u0111\u1ea1t ho\u1eb7c claim \u0111\u00e3 c\u00f3 ngo\u1ea1i l\u1ec7 \u0111\u01b0\u1ee3c duy\u1ec7t.",
    action: (order, status) => ({
      label: "M\u1edf thanh to\u00e1n",
      href: subcontractOperationsHref(order, "subcontract-payment"),
      disabled: status === "pending" || status === "blocked"
    })
  },
  {
    id: "closed",
    label: "\u0110\u00f3ng l\u1ec7nh",
    completeAt: 12,
    currentAt: [11],
    description: () => "\u0110\u00f3ng l\u1ec7nh khi nh\u1eadn h\u00e0ng, QC, claim v\u00e0 thanh to\u00e1n cu\u1ed1i \u0111\u00e3 xong."
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

function statusProgressIndex(status: SubcontractOrderStatus, dispatchStatus?: FactoryDispatchTimelineStatus) {
  switch (status) {
    case "draft":
      return 0;
    case "submitted":
      return 1;
    case "approved":
      return approvedProgressIndex(dispatchStatus);
    case "factory_confirmed":
      return 4;
    case "deposit_recorded":
      return 5;
    case "materials_issued_to_factory":
      return 6;
    case "sample_submitted":
    case "sample_rejected":
      return 6;
    case "sample_approved":
      return 7;
    case "mass_production_started":
      return 8;
    case "finished_goods_received":
      return 9;
    case "qc_in_progress":
      return 9;
    case "accepted":
      return 10;
    case "final_payment_ready":
      return 11;
    case "closed":
      return 12;
    case "rejected_with_factory_issue":
    case "cancelled":
    default:
      return 0;
  }
}

function approvedProgressIndex(dispatchStatus?: FactoryDispatchTimelineStatus) {
  if (dispatchStatus === "sent" || dispatchStatus === "confirmed") {
    return 3;
  }

  return 2;
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
