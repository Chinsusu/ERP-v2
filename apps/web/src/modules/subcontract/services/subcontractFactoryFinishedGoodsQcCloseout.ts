import type {
  AcceptSubcontractFinishedGoodsInput,
  CreateSubcontractFactoryClaimInput,
  PartialAcceptSubcontractFinishedGoodsInput,
  SubcontractFactoryClaimSeverity,
  SubcontractFinishedGoodsReceipt,
  SubcontractOrder
} from "../types";

export type FactoryFinishedGoodsQcCloseoutStatus =
  | "blocked"
  | "ready_for_qc"
  | "partially_closed"
  | "passed"
  | "failed";

export type FactoryFinishedGoodsQcCloseoutGate = {
  status: FactoryFinishedGoodsQcCloseoutStatus;
  canCloseout: boolean;
  blockedReason?: string;
  receiptNo?: string;
  receivedQty: string;
  acceptedQty: string;
  rejectedQty: string;
  remainingQcQty: string;
  uomCode: string;
};

type QcClaimArgs = {
  order: SubcontractOrder;
  latestReceipt?: SubcontractFinishedGoodsReceipt;
  ownerId: string;
  openedBy: string;
  openedAt?: string;
  reasonCode: string;
  reason: string;
  severity: SubcontractFactoryClaimSeverity;
  evidenceFileName?: string;
  evidenceNote?: string;
};

export type BuildFactoryFinishedGoodsQcAcceptInputArgs = {
  order: SubcontractOrder;
  latestReceipt?: SubcontractFinishedGoodsReceipt;
  acceptedBy: string;
  acceptedAt?: string;
  note?: string;
};

export type BuildFactoryFinishedGoodsQcPartialAcceptInputArgs = QcClaimArgs & {
  acceptedQty: string;
  rejectedQty: string;
  acceptedBy: string;
  acceptedAt?: string;
  note?: string;
};

export type BuildFactoryFinishedGoodsQcRejectInputArgs = QcClaimArgs;

export function buildSubcontractFactoryFinishedGoodsQcCloseout(
  order: SubcontractOrder,
  latestReceipt?: SubcontractFinishedGoodsReceipt
): FactoryFinishedGoodsQcCloseoutGate {
  const receivedQty = quantityString(order.receivedQty ?? latestReceiptReceivedQty(latestReceipt));
  const acceptedQty = quantityString(order.acceptedQty ?? "0");
  const rejectedQty = quantityString(order.rejectedQty ?? "0");
  const remainingQcQty = quantityString(numeric(receivedQty) - numeric(acceptedQty) - numeric(rejectedQty));
  const uomCode = latestReceipt?.lines[0]?.uomCode ?? order.uomCode ?? "PCS";

  if (numeric(receivedQty) <= 0) {
    return {
      status: "blocked",
      canCloseout: false,
      blockedReason: "Chưa có phiếu nhận thành phẩm vào QC hold.",
      receivedQty,
      acceptedQty,
      rejectedQty,
      remainingQcQty: "0.000000",
      uomCode
    };
  }

  if (["accepted", "final_payment_ready", "closed"].includes(order.status)) {
    return {
      status: "passed",
      canCloseout: false,
      receiptNo: latestReceipt?.receiptNo,
      receivedQty,
      acceptedQty,
      rejectedQty,
      remainingQcQty: positiveQuantityString(remainingQcQty),
      uomCode
    };
  }

  if (order.status === "rejected_with_factory_issue") {
    return {
      status: "failed",
      canCloseout: false,
      receiptNo: latestReceipt?.receiptNo,
      receivedQty,
      acceptedQty,
      rejectedQty,
      remainingQcQty: positiveQuantityString(remainingQcQty),
      uomCode
    };
  }

  if (!["finished_goods_received", "qc_in_progress"].includes(order.status)) {
    return {
      status: "blocked",
      canCloseout: false,
      blockedReason: "Chỉ QC sau khi thành phẩm đã vào QC hold.",
      receiptNo: latestReceipt?.receiptNo,
      receivedQty,
      acceptedQty,
      rejectedQty,
      remainingQcQty: positiveQuantityString(remainingQcQty),
      uomCode
    };
  }

  if (numeric(remainingQcQty) <= 0) {
    return {
      status: numeric(rejectedQty) > 0 ? "partially_closed" : "passed",
      canCloseout: false,
      receiptNo: latestReceipt?.receiptNo,
      receivedQty,
      acceptedQty,
      rejectedQty,
      remainingQcQty: "0.000000",
      uomCode
    };
  }

  return {
    status: numeric(acceptedQty) > 0 || numeric(rejectedQty) > 0 ? "partially_closed" : "ready_for_qc",
    canCloseout: true,
    receiptNo: latestReceipt?.receiptNo,
    receivedQty,
    acceptedQty,
    rejectedQty,
    remainingQcQty,
    uomCode
  };
}

export function buildFactoryFinishedGoodsQcAcceptInput({
  order,
  acceptedBy,
  acceptedAt,
  note
}: BuildFactoryFinishedGoodsQcAcceptInputArgs): AcceptSubcontractFinishedGoodsInput {
  return {
    order,
    acceptedBy: requiredText(acceptedBy, "Người QC là bắt buộc"),
    acceptedAt,
    note: optionalText(note)
  };
}

export function buildFactoryFinishedGoodsQcPartialAcceptInput({
  order,
  latestReceipt,
  acceptedQty,
  rejectedQty,
  acceptedBy,
  acceptedAt,
  openedBy,
  openedAt,
  ownerId,
  reasonCode,
  reason,
  severity,
  evidenceFileName,
  evidenceNote,
  note
}: BuildFactoryFinishedGoodsQcPartialAcceptInputArgs): PartialAcceptSubcontractFinishedGoodsInput {
  const line = latestReceipt?.lines[0];
  const uomCode = line?.uomCode ?? order.uomCode ?? "PCS";
  const baseUOMCode = line?.baseUOMCode ?? uomCode;

  return {
    order,
    acceptedQty: requiredPositiveQuantity(acceptedQty, "Số lượng đạt QC là bắt buộc"),
    rejectedQty: requiredPositiveQuantity(rejectedQty, "Số lượng lỗi là bắt buộc"),
    uomCode,
    baseAcceptedQty: requiredPositiveQuantity(acceptedQty, "Số lượng đạt QC là bắt buộc"),
    baseRejectedQty: requiredPositiveQuantity(rejectedQty, "Số lượng lỗi là bắt buộc"),
    baseUOMCode,
    claimId: `sfc-${order.id}-${Date.now()}`,
    receiptId: latestReceipt?.id,
    receiptNo: latestReceipt?.receiptNo,
    reasonCode: requiredText(reasonCode, "Mã lý do claim là bắt buộc"),
    reason: requiredText(reason, "Lý do claim là bắt buộc"),
    severity,
    evidence: buildClaimEvidence(order, evidenceFileName, evidenceNote),
    ownerId: requiredText(ownerId, "Owner claim là bắt buộc"),
    acceptedBy: requiredText(acceptedBy, "Người QC là bắt buộc"),
    acceptedAt,
    openedBy: requiredText(openedBy, "Người mở claim là bắt buộc"),
    openedAt,
    note: optionalText(note)
  };
}

export function buildFactoryFinishedGoodsQcRejectInput({
  order,
  latestReceipt,
  openedBy,
  openedAt,
  ownerId,
  reasonCode,
  reason,
  severity,
  evidenceFileName,
  evidenceNote
}: BuildFactoryFinishedGoodsQcRejectInputArgs): CreateSubcontractFactoryClaimInput {
  const gate = buildSubcontractFactoryFinishedGoodsQcCloseout(order, latestReceipt);
  const line = latestReceipt?.lines[0];
  const uomCode = line?.uomCode ?? order.uomCode ?? "PCS";
  const baseUOMCode = line?.baseUOMCode ?? uomCode;

  return {
    order,
    claimId: `sfc-${order.id}-${Date.now()}`,
    receiptId: latestReceipt?.id,
    receiptNo: latestReceipt?.receiptNo,
    reasonCode: requiredText(reasonCode, "Mã lý do claim là bắt buộc"),
    reason: requiredText(reason, "Lý do claim là bắt buộc"),
    severity,
    affectedQty: gate.remainingQcQty,
    uomCode,
    baseAffectedQty: gate.remainingQcQty,
    baseUOMCode,
    ownerId: requiredText(ownerId, "Owner claim là bắt buộc"),
    openedBy: requiredText(openedBy, "Người mở claim là bắt buộc"),
    openedAt,
    evidence: buildClaimEvidence(order, evidenceFileName, evidenceNote)
  };
}

function buildClaimEvidence(order: SubcontractOrder, fileName?: string, note?: string) {
  const normalizedFileName = optionalText(fileName) || "qc-fail-evidence.jpg";

  return [
    {
      evidenceType: "qc_photo" as const,
      fileName: normalizedFileName,
      objectKey: `subcontract-finished-goods/${order.id}/qc/${normalizedFileName}`,
      note: optionalText(note)
    }
  ];
}

function latestReceiptReceivedQty(receipt?: SubcontractFinishedGoodsReceipt) {
  if (!receipt) {
    return "0";
  }

  return receipt.lines.reduce((sum, line) => sum + numeric(line.receiveQty), 0).toString();
}

function requiredText(value: string | undefined, message: string) {
  const trimmed = value?.trim();
  if (!trimmed) {
    throw new Error(message);
  }

  return trimmed;
}

function optionalText(value: string | undefined) {
  return value?.trim() || undefined;
}

function requiredPositiveQuantity(value: string, message: string) {
  const trimmed = value.trim();
  if (numeric(trimmed) <= 0) {
    throw new Error(message);
  }

  return trimmed;
}

function positiveQuantityString(value: string) {
  return numeric(value) > 0 ? quantityString(value) : "0.000000";
}

function quantityString(value: string | number) {
  return numeric(value).toFixed(6);
}

function numeric(value: string | number | undefined) {
  if (value === undefined) {
    return 0;
  }

  return Number.parseFloat(String(value).replace(",", ".")) || 0;
}
