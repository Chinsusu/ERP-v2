import type {
  DecideSubcontractSampleInput,
  SubcontractOrder,
  SubcontractOrderStatus,
  SubcontractSampleApproval,
  SubmitSubcontractSampleInput
} from "../types";

export type FactorySampleStatus = "blocked" | "pending" | "ready_to_submit" | "submitted" | "approved" | "rejected" | "not_required";

export type FactoryMassProductionStatus = "blocked" | "pending" | "ready_to_start" | "started";

export type FactorySampleMassProductionGate = {
  sampleStatus: FactorySampleStatus;
  sampleBlockedReason?: string;
  canSubmitSample: boolean;
  canApproveSample: boolean;
  canRejectSample: boolean;
  latestSampleCode?: string;
  massStatus: FactoryMassProductionStatus;
  massBlockedReason?: string;
  canStartMassProduction: boolean;
};

export type BuildFactorySampleSubmissionInputArgs = {
  order: SubcontractOrder;
  sampleCode: string;
  formulaVersion?: string;
  evidenceFileName?: string;
  note?: string;
  submittedAt?: string;
};

export type BuildFactorySampleDecisionInputArgs = {
  order: SubcontractOrder;
  sampleApproval?: SubcontractSampleApproval;
  decision: "approve" | "reject";
  reason: string;
  storageStatus?: string;
  decisionAt?: string;
};

export function buildSubcontractFactorySampleMassProduction(
  order: SubcontractOrder,
  latestSample?: SubcontractSampleApproval
): FactorySampleMassProductionGate {
  const sampleStatus = resolveSampleStatus(order, latestSample);
  const canSubmitSample = sampleStatus === "ready_to_submit" || sampleStatus === "rejected";
  const canApproveSample = sampleStatus === "submitted";
  const canRejectSample = sampleStatus === "submitted";
  const massStatus = resolveMassProductionStatus(order, sampleStatus);

  return {
    sampleStatus,
    sampleBlockedReason: sampleBlockedReason(order, sampleStatus),
    canSubmitSample,
    canApproveSample,
    canRejectSample,
    latestSampleCode: latestSample?.sampleCode,
    massStatus,
    massBlockedReason: massProductionBlockedReason(order, massStatus, sampleStatus),
    canStartMassProduction: massStatus === "ready_to_start"
  };
}

export function buildFactorySampleSubmissionInput({
  order,
  sampleCode,
  formulaVersion,
  evidenceFileName,
  note,
  submittedAt
}: BuildFactorySampleSubmissionInputArgs): SubmitSubcontractSampleInput {
  const normalizedEvidenceFileName = evidenceFileName?.trim();

  return {
    order,
    sampleApprovalId: `sample-${order.id}-${Date.now()}`,
    sampleCode: sampleCode.trim(),
    formulaVersion: formulaVersion?.trim() || undefined,
    specVersion: order.specVersion,
    submittedBy: "factory-user",
    submittedAt,
    note: note?.trim() || undefined,
    evidence: normalizedEvidenceFileName
      ? [
          {
            evidenceType: "photo",
            fileName: normalizedEvidenceFileName,
            objectKey: `subcontract-samples/${order.id}/${normalizedEvidenceFileName}`
          }
        ]
      : []
  };
}

export function buildFactorySampleDecisionInput({
  order,
  sampleApproval,
  decision,
  reason,
  storageStatus,
  decisionAt
}: BuildFactorySampleDecisionInputArgs): DecideSubcontractSampleInput {
  return {
    order,
    sampleApprovalId: sampleApproval?.id,
    reason: reason.trim(),
    storageStatus: decision === "approve" ? storageStatus?.trim() || undefined : undefined,
    decisionAt
  };
}

function resolveSampleStatus(order: SubcontractOrder, latestSample?: SubcontractSampleApproval): FactorySampleStatus {
  if (!order.sampleRequired) {
    return "not_required";
  }
  if (order.status === "sample_rejected" || latestSample?.status === "rejected") {
    return "rejected";
  }
  if (isAtLeast(order.status, "sample_approved") || latestSample?.status === "approved") {
    return "approved";
  }
  if (order.status === "sample_submitted" || latestSample?.status === "submitted") {
    return "submitted";
  }
  if (isAtLeast(order.status, "materials_issued_to_factory")) {
    return "ready_to_submit";
  }

  return "pending";
}

function resolveMassProductionStatus(order: SubcontractOrder, sampleStatus: FactorySampleStatus): FactoryMassProductionStatus {
  if (isAtLeast(order.status, "mass_production_started")) {
    return "started";
  }
  if (sampleStatus === "rejected") {
    return "blocked";
  }
  if (materialIssueComplete(order) && (sampleStatus === "approved" || sampleStatus === "not_required")) {
    return "ready_to_start";
  }

  return "pending";
}

function sampleBlockedReason(order: SubcontractOrder, sampleStatus: FactorySampleStatus) {
  if (sampleStatus === "pending" && !materialIssueComplete(order)) {
    return "Cần bàn giao đủ vật tư cho nhà máy trước khi gửi mẫu.";
  }
  if (sampleStatus === "rejected") {
    return order.sampleRejectReason || "Mẫu bị từ chối; cần gửi lại mẫu đạt trước khi sản xuất hàng loạt.";
  }

  return undefined;
}

function massProductionBlockedReason(
  order: SubcontractOrder,
  massStatus: FactoryMassProductionStatus,
  sampleStatus: FactorySampleStatus
) {
  if (massStatus === "blocked") {
    return order.sampleRejectReason || "Mẫu chưa đạt; không được bắt đầu sản xuất hàng loạt.";
  }
  if (massStatus === "pending" && !materialIssueComplete(order)) {
    return "Cần bàn giao đủ vật tư cho nhà máy trước khi sản xuất hàng loạt.";
  }
  if (massStatus === "pending" && sampleStatus !== "approved" && sampleStatus !== "not_required") {
    return "Cần mẫu đạt hoặc lệnh không yêu cầu mẫu trước khi sản xuất hàng loạt.";
  }

  return undefined;
}

function materialIssueComplete(order: SubcontractOrder) {
  if (isAtLeast(order.status, "materials_issued_to_factory")) {
    return true;
  }

  return (
    order.materialLines.length > 0 &&
    order.materialLines.every((line) => numeric(line.issuedQty) + 0.000001 >= numeric(line.plannedQty))
  );
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

function numeric(value?: string) {
  if (!value) {
    return 0;
  }

  return Number.parseFloat(value.replace(",", ".")) || 0;
}
