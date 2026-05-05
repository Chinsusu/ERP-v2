import { formatProductionPlanQuantity } from "./productionPlanService";
import type { ProductionPlan, ProductionPlanIssueStatus } from "../types";

type WorkTaskTone = "normal" | "success" | "warning" | "danger" | "info";

export type ProductionPlanWorkTask = {
  id: string;
  step: number;
  title: string;
  statusLabel: string;
  statusTone: WorkTaskTone;
  detail: string;
  action?: {
    label: string;
    href?: string;
    disabled?: boolean;
  };
};

export function buildProductionPlanWorklist(plan: ProductionPlan): ProductionPlanWorkTask[] {
  const lineCount = plan.lines.length;
  const shortageLineCount = plan.lines.filter(
    (line) => line.issueStatus === "shortage" || line.needsPurchase || Number(line.shortageQty) > 0
  ).length;
  const purchaseLineCount = plan.purchaseRequestDraft.lines.length;
  const quantityLabel = formatProductionPlanQuantity(plan.plannedQty, plan.uomCode);
  const materialDemand = buildMaterialDemandTaskState(lineCount, shortageLineCount);
  const purchaseRequest = buildPurchaseRequestTaskState(plan);
  const approval = buildPurchaseApprovalTaskState(plan);
  const purchaseOrder = buildPurchaseOrderTaskState(plan);
  const receiving = buildReceivingTaskState(plan);
  const materialIssue = buildMaterialIssueTaskState(plan);

  return [
    {
      id: "production-plan",
      step: 1,
      title: "Kế hoạch sản xuất",
      statusLabel: "Đã tạo",
      statusTone: "success",
      detail: `${plan.planNo} - ${plan.outputSku} - ${quantityLabel}; công thức ${plan.formulaCode} - ${plan.formulaVersion}.`
    },
    {
      id: "material-demand",
      step: 2,
      title: "Nhu cầu vật tư",
      statusLabel: materialDemand.statusLabel,
      statusTone: materialDemand.statusTone,
      detail: materialDemand.detail
    },
    {
      id: "purchase-request",
      step: 3,
      title: "Đề nghị mua",
      statusLabel: purchaseRequest.statusLabel,
      statusTone: purchaseRequest.statusTone,
      detail:
        purchaseLineCount > 0
          ? `${plan.purchaseRequestDraft.requestNo} có ${purchaseLineCount} dòng vật tư thiếu từ ${plan.planNo}.`
          : "Kế hoạch không phát sinh đề nghị mua từ MRP.",
      action: purchaseRequest.action
    },
    {
      id: "purchase-request-approval",
      step: 4,
      title: "Duyệt đề nghị mua",
      statusLabel: approval.statusLabel,
      statusTone: approval.statusTone,
      detail:
        purchaseLineCount > 0
          ? "Đề nghị mua phải được gửi duyệt và duyệt trước khi chuyển thành PO."
          : "Không có đề nghị mua cần duyệt.",
      action: approval.action
    },
    {
      id: "purchase-order",
      step: 5,
      title: "PO vật tư",
      statusLabel: purchaseOrder.statusLabel,
      statusTone: purchaseOrder.statusTone,
      detail: purchaseLineCount > 0 ? purchaseOrder.detail : "Kế hoạch không cần PO vật tư.",
      action: purchaseOrder.action
    },
    {
      id: "receiving-qc",
      step: 6,
      title: "Nhập kho/QC vật tư",
      statusLabel: receiving.statusLabel,
      statusTone: receiving.statusTone,
      detail: purchaseLineCount > 0 ? receiving.detail : "Không có vật tư mua thêm cần nhập kho hoặc QC.",
      action: receiving.action
    },
    {
      id: "warehouse-issue",
      step: 7,
      title: "Phiếu xuất kho vật tư",
      statusLabel: materialIssue.statusLabel,
      statusTone: materialIssue.statusTone,
      detail: materialIssue.detail,
      action: materialIssue.action
    },
    {
      id: "subcontract-order",
      step: 8,
      title: "Lệnh gia công",
      statusLabel: materialIssue.readyForSubcontract ? "Sẵn sàng tạo lệnh" : "Chờ xuất vật tư",
      statusTone: materialIssue.readyForSubcontract ? "success" : "warning",
      detail: materialIssue.readyForSubcontract
        ? "Mở module Gia công để tạo hoặc theo dõi lệnh sản xuất từ kế hoạch này."
        : "Tạo lệnh gia công sau khi vật tư đã được xuất kho hoặc có waiver.",
      action: {
        label: materialIssue.readyForSubcontract ? "Mở gia công" : "Chờ bước 7",
        href: materialIssue.readyForSubcontract ? subcontractHref(plan) : undefined,
        disabled: !materialIssue.readyForSubcontract
      }
    }
  ];
}

function buildMaterialDemandTaskState(
  lineCount: number,
  shortageLineCount: number
): Pick<ProductionPlanWorkTask, "statusLabel" | "statusTone" | "detail"> {
  if (lineCount === 0) {
    return { statusLabel: "Chưa có dòng vật tư", statusTone: "warning", detail: "Kế hoạch chưa có nhu cầu vật tư." };
  }
  if (shortageLineCount > 0) {
    return {
      statusLabel: `Thiếu ${shortageLineCount} dòng vật tư`,
      statusTone: "warning",
      detail: `${lineCount - shortageLineCount}/${lineCount} dòng vật tư đủ tồn khả dụng; ${shortageLineCount} dòng cần mua thêm.`
    };
  }

  return {
    statusLabel: "Đủ vật tư",
    statusTone: "success",
    detail: `${lineCount} dòng vật tư đã đủ tồn khả dụng.`
  };
}

function buildPurchaseRequestTaskState(plan: ProductionPlan): Pick<ProductionPlanWorkTask, "statusLabel" | "statusTone" | "action"> {
  const draft = plan.purchaseRequestDraft;
  if (draft.lines.length === 0) {
    return { statusLabel: "Không cần đề nghị mua", statusTone: "success" };
  }

  return {
    statusLabel: purchaseRequestStatusLabel(draft.status),
    statusTone: purchaseRequestStatusTone(draft.status),
    action: {
      label: "Mở đề nghị",
      href: purchaseRequestHref(plan),
      disabled: !draft.id
    }
  };
}

function buildPurchaseApprovalTaskState(plan: ProductionPlan): Pick<ProductionPlanWorkTask, "statusLabel" | "statusTone" | "action"> {
  const draft = plan.purchaseRequestDraft;
  if (draft.lines.length === 0) {
    return { statusLabel: "Không cần duyệt", statusTone: "success" };
  }
  if (draft.status === "draft") {
    return {
      statusLabel: "Chờ gửi duyệt",
      statusTone: "warning",
      action: { label: "Mở đề nghị", href: purchaseRequestHref(plan), disabled: !draft.id }
    };
  }
  if (draft.status === "submitted") {
    return {
      statusLabel: "Chờ duyệt",
      statusTone: "info",
      action: { label: "Mở đề nghị", href: purchaseRequestHref(plan), disabled: !draft.id }
    };
  }
  if (draft.status === "rejected") {
    return { statusLabel: "Đã từ chối", statusTone: "danger" };
  }
  if (draft.status === "cancelled") {
    return { statusLabel: "Đã hủy", statusTone: "danger" };
  }

  return { statusLabel: "Đã duyệt", statusTone: "success" };
}

function buildPurchaseOrderTaskState(plan: ProductionPlan): Pick<ProductionPlanWorkTask, "statusLabel" | "statusTone" | "detail" | "action"> {
  const draft = plan.purchaseRequestDraft;
  if (draft.lines.length === 0) {
    return { statusLabel: "Không cần PO", statusTone: "success", detail: "Kế hoạch không phát sinh vật tư thiếu." };
  }
  if (draft.status === "converted_to_po") {
    return {
      statusLabel: "Đã tạo PO",
      statusTone: "success",
      detail: `${draft.convertedPurchaseOrderNo ?? "PO"} được tạo từ ${draft.requestNo}.`,
      action: draft.convertedPurchaseOrderId
        ? { label: "Mở PO", href: `/purchase/orders/${encodeURIComponent(draft.convertedPurchaseOrderId)}`, disabled: false }
        : undefined
    };
  }
  if (draft.status === "approved") {
    return {
      statusLabel: "Cần tạo PO",
      statusTone: "warning",
      detail: `${draft.requestNo} đã duyệt; mở đề nghị mua để tạo PO.`,
      action: { label: "Mở đề nghị", href: purchaseRequestHref(plan), disabled: !draft.id }
    };
  }

  return {
    statusLabel: "Chờ duyệt đề nghị",
    statusTone: "info",
    detail: "PO chỉ được tạo sau khi đề nghị mua đã duyệt.",
    action: { label: "Mở đề nghị", href: purchaseRequestHref(plan), disabled: !draft.id }
  };
}

function buildReceivingTaskState(plan: ProductionPlan): Pick<ProductionPlanWorkTask, "statusLabel" | "statusTone" | "detail" | "action"> {
  const draft = plan.purchaseRequestDraft;
  if (draft.lines.length === 0) {
    return { statusLabel: "Không chờ nhập mua", statusTone: "success", detail: "Không có vật tư mua thêm." };
  }
  if (draft.status === "converted_to_po") {
    return {
      statusLabel: "Chờ nhập kho/QC",
      statusTone: "info",
      detail: "Theo dõi lịch giao, phiếu nhập và QC vật tư theo PO đã tạo.",
      action: { label: "Mở nhập kho", href: "/receiving", disabled: false }
    };
  }

  return {
    statusLabel: "Chờ PO",
    statusTone: "warning",
    detail: "Chưa thể nhập kho vật tư khi đề nghị mua chưa chuyển thành PO."
  };
}

function buildMaterialIssueTaskState(
  plan: ProductionPlan
): Pick<ProductionPlanWorkTask, "statusLabel" | "statusTone" | "detail" | "action"> & { readyForSubcontract: boolean } {
  const stockLines = plan.lines.filter((line) => line.isStockManaged);
  const shortageCount = stockLines.filter((line) => line.issueStatus === "shortage").length;
  const readyCount = stockLines.filter((line) => line.issueStatus === "ready_to_issue" || line.issueStatus === "partially_issued").length;
  const pendingCount = stockLines.filter((line) => pendingIssueStatuses.includes(line.issueStatus)).length;
  const issuedCount = stockLines.filter((line) => line.issueStatus === "issued" || line.issueStatus === "waived").length;

  if (stockLines.length === 0) {
    return {
      statusLabel: "Không cần xuất kho",
      statusTone: "success",
      detail: "Kế hoạch không có dòng vật tư quản lý tồn.",
      readyForSubcontract: true
    };
  }
  if (issuedCount === stockLines.length) {
    return {
      statusLabel: "Đã xuất đủ",
      statusTone: "success",
      detail: `${issuedCount}/${stockLines.length} dòng vật tư đã có bằng chứng xuất kho.`,
      action: { label: "Mở phiếu xuất", href: "/inventory#warehouse-issues", disabled: false },
      readyForSubcontract: true
    };
  }
  if (pendingCount > 0) {
    return {
      statusLabel: "Có phiếu xuất đang xử lý",
      statusTone: "info",
      detail: `${pendingCount} dòng đã có phiếu xuất chưa post; chỉ mở gia công sau khi post.`,
      action: { label: "Mở phiếu xuất", href: "/inventory#warehouse-issues", disabled: false },
      readyForSubcontract: false
    };
  }
  if (readyCount > 0 && shortageCount === 0) {
    return {
      statusLabel: "Sẵn sàng xuất kho",
      statusTone: "warning",
      detail: `${readyCount} dòng đủ tồn; tạo phiếu xuất ở bảng nhu cầu vật tư.`,
      action: { label: "Tạo ở bảng vật tư", disabled: true },
      readyForSubcontract: false
    };
  }

  return {
    statusLabel: "Chờ đủ vật tư",
    statusTone: "warning",
    detail: `Còn ${shortageCount} dòng thiếu hoặc chưa đủ điều kiện xuất kho.`,
    action: { label: "Chờ bước 6", disabled: true },
    readyForSubcontract: false
  };
}

const pendingIssueStatuses: ProductionPlanIssueStatus[] = ["issue_draft", "issue_submitted", "issue_approved"];

function purchaseRequestStatusLabel(status = "draft") {
  switch (status) {
    case "submitted":
      return "Đã gửi duyệt";
    case "approved":
      return "Đã duyệt";
    case "converted_to_po":
      return "Đã chuyển PO";
    case "cancelled":
      return "Đã hủy";
    case "rejected":
      return "Đã từ chối";
    case "draft":
    default:
      return "Đề nghị nháp";
  }
}

function purchaseRequestStatusTone(status = "draft"): WorkTaskTone {
  switch (status) {
    case "approved":
    case "converted_to_po":
      return "success";
    case "submitted":
      return "info";
    case "cancelled":
    case "rejected":
      return "danger";
    case "draft":
    default:
      return "warning";
  }
}

function purchaseRequestHref(plan: ProductionPlan) {
  const requestID = plan.purchaseRequestDraft.id;
  return requestID ? `/purchase/requests/${encodeURIComponent(requestID)}` : `/purchase?search=${encodeURIComponent(plan.planNo)}#purchase-list`;
}

function subcontractHref(plan: ProductionPlan) {
  return `/subcontract?source_production_plan_id=${encodeURIComponent(plan.id)}&search=${encodeURIComponent(plan.planNo)}#subcontract-orders`;
}
