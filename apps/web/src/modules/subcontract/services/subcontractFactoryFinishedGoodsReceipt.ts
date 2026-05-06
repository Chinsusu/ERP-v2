import type {
  ReceiveSubcontractFinishedGoodsInput,
  SubcontractFinishedGoodsPackagingStatus,
  SubcontractOrder,
  SubcontractOrderStatus
} from "../types";

export type FactoryFinishedGoodsReceiptStatus = "blocked" | "ready_to_receive" | "partial" | "complete";

export type FactoryFinishedGoodsReceiptGate = {
  status: FactoryFinishedGoodsReceiptStatus;
  blockedReason?: string;
  canReceive: boolean;
  plannedQty: string;
  receivedQty: string;
  remainingQty: string;
  uomCode: string;
};

export type FactoryFinishedGoodsReceiptDraft = {
  receiveQty: string;
  batchNo: string;
  lotNo: string;
  expiryDate: string;
  packagingStatus: SubcontractFinishedGoodsPackagingStatus;
  note: string;
};

export type BuildFactoryFinishedGoodsReceiptInputArgs = {
  order: SubcontractOrder;
  warehouseId: string;
  warehouseCode: string;
  locationId: string;
  locationCode: string;
  deliveryNoteNo: string;
  receivedBy: string;
  evidenceFileName?: string;
  note?: string;
  receivedAt?: string;
  draft: FactoryFinishedGoodsReceiptDraft;
};

export function buildSubcontractFactoryFinishedGoodsReceipt(order: SubcontractOrder): FactoryFinishedGoodsReceiptGate {
  const plannedQty = normalizeQuantity(String(order.quantity));
  const receivedQty = normalizeQuantity(order.receivedQty ?? "0");
  const remainingQty = normalizeQuantity(String(Math.max(numeric(plannedQty) - numeric(receivedQty), 0)));
  const uomCode = order.uomCode ?? "PCS";

  if (order.status === "cancelled" || order.status === "rejected_with_factory_issue") {
    return {
      status: "blocked",
      blockedReason: "Lệnh đã bị chặn hoặc hủy; không thể nhận thêm thành phẩm.",
      canReceive: false,
      plannedQty,
      receivedQty,
      remainingQty,
      uomCode
    };
  }

  if (isAtLeast(order.status, "qc_in_progress")) {
    return {
      status: numeric(remainingQty) <= 0 ? "complete" : "blocked",
      blockedReason: numeric(remainingQty) <= 0 ? undefined : "Lệnh đã chuyển sang QC/closeout; không nhận thêm tại bước này.",
      canReceive: false,
      plannedQty,
      receivedQty,
      remainingQty,
      uomCode
    };
  }

  if (!isAtLeast(order.status, "mass_production_started")) {
    return {
      status: "blocked",
      blockedReason: "Chỉ nhận thành phẩm sau khi nhà máy đã bắt đầu sản xuất hàng loạt.",
      canReceive: false,
      plannedQty,
      receivedQty,
      remainingQty,
      uomCode
    };
  }

  if (numeric(remainingQty) <= 0) {
    return {
      status: "complete",
      canReceive: false,
      plannedQty,
      receivedQty,
      remainingQty,
      uomCode
    };
  }

  return {
    status: numeric(receivedQty) > 0 ? "partial" : "ready_to_receive",
    canReceive: true,
    plannedQty,
    receivedQty,
    remainingQty,
    uomCode
  };
}

export function buildFactoryFinishedGoodsReceiptInput({
  order,
  warehouseId,
  warehouseCode,
  locationId,
  locationCode,
  deliveryNoteNo,
  receivedBy,
  evidenceFileName,
  note,
  receivedAt,
  draft
}: BuildFactoryFinishedGoodsReceiptInputArgs): ReceiveSubcontractFinishedGoodsInput {
  const normalizedEvidenceFileName = evidenceFileName?.trim();
  const batchNo = draft.batchNo.trim();
  const lotNo = draft.lotNo.trim() || batchNo;
  const receiveQty = draft.receiveQty.trim();
  const uomCode = order.uomCode ?? "PCS";

  return {
    order,
    receiptId: `sfgr-${order.id}-${Date.now()}`,
    warehouseId: warehouseId.trim(),
    warehouseCode: warehouseCode.trim(),
    locationId: locationId.trim(),
    locationCode: locationCode.trim(),
    deliveryNoteNo: deliveryNoteNo.trim(),
    receivedBy: receivedBy.trim(),
    receivedAt,
    note: note?.trim() || undefined,
    lines: [
      {
        lineNo: 1,
        itemId: order.productId,
        skuCode: order.sku,
        itemName: order.productName,
        batchId: batchNo,
        batchNo,
        lotNo,
        expiryDate: draft.expiryDate.trim(),
        receiveQty,
        uomCode,
        baseReceiveQty: receiveQty,
        baseUOMCode: uomCode,
        conversionFactor: "1",
        packagingStatus: draft.packagingStatus,
        note: draft.note.trim() || undefined
      }
    ],
    evidence: normalizedEvidenceFileName
      ? [
          {
            evidenceType: "delivery_note",
            fileName: normalizedEvidenceFileName,
            objectKey: `subcontract-finished-goods/${order.id}/${normalizedEvidenceFileName}`
          }
        ]
      : undefined
  };
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

function normalizeQuantity(value: string) {
  return numeric(value).toFixed(6);
}

function numeric(value?: string) {
  if (!value) {
    return 0;
  }

  return Number.parseFloat(value.replace(",", ".")) || 0;
}
