import { decimalScales, normalizeDecimalInput } from "../../../shared/format/numberFormat";
import type {
  IssueSubcontractMaterialsInput,
  SubcontractOrder,
  SubcontractOrderMaterialLine,
  SubcontractOrderStatus
} from "../types";

export type FactoryMaterialHandoverStatus = "blocked" | "ready" | "complete";
export type FactoryMaterialHandoverLineStatus = "pending" | "ready" | "complete";

export type FactoryMaterialHandoverLine = {
  id: string;
  itemId: string;
  skuCode: string;
  itemName: string;
  plannedQty: string;
  issuedQty: string;
  remainingQty: string;
  uomCode: string;
  lotTraceRequired: boolean;
  status: FactoryMaterialHandoverLineStatus;
  note?: string;
};

export type FactoryMaterialHandover = {
  status: FactoryMaterialHandoverStatus;
  canIssue: boolean;
  blockedReason?: string;
  totalLines: number;
  completeLines: number;
  pendingLines: number;
  lines: FactoryMaterialHandoverLine[];
};

export type FactoryMaterialHandoverLineDraft = {
  issueQty?: string;
  batchNo?: string;
  sourceBinId?: string;
  note?: string;
};

export type BuildFactoryMaterialHandoverIssueInputArgs = {
  order: SubcontractOrder;
  sourceWarehouseId: string;
  sourceWarehouseCode: string;
  handoverBy: string;
  receivedBy: string;
  receiverContact?: string;
  vehicleNo?: string;
  note?: string;
  evidenceFileName?: string;
  lineDrafts: Record<string, FactoryMaterialHandoverLineDraft>;
};

export function buildSubcontractFactoryMaterialHandover(order: SubcontractOrder): FactoryMaterialHandover {
  const completeLines = order.materialLines.filter((line) => isMaterialLineComplete(line)).length;
  const totalLines = order.materialLines.length;
  const pendingLines = totalLines - completeLines;
  const complete = totalLines > 0 && completeLines === totalLines;
  const blockedReason = complete ? undefined : materialHandoverBlockedReason(order);
  const canIssue = !complete && !blockedReason;
  const status: FactoryMaterialHandoverStatus = complete ? "complete" : canIssue ? "ready" : "blocked";

  return {
    status,
    canIssue,
    blockedReason,
    totalLines,
    completeLines,
    pendingLines,
    lines: order.materialLines.map((line) => toHandoverLine(line, canIssue))
  };
}

export function buildFactoryMaterialHandoverIssueInput({
  order,
  sourceWarehouseId,
  sourceWarehouseCode,
  handoverBy,
  receivedBy,
  receiverContact,
  vehicleNo,
  note,
  evidenceFileName,
  lineDrafts
}: BuildFactoryMaterialHandoverIssueInputArgs): IssueSubcontractMaterialsInput {
  const handover = buildSubcontractFactoryMaterialHandover(order);
  const lines = handover.lines
    .filter((line) => isPositiveQuantity(line.remainingQty))
    .map((line) => {
      const draft = lineDrafts[line.id] ?? {};

      return {
        orderMaterialLineId: line.id,
        issueQty: normalizeDecimalInput(draft.issueQty || line.remainingQty, decimalScales.quantity),
        uomCode: line.uomCode,
        batchNo: trimOptional(draft.batchNo),
        sourceBinId: trimOptional(draft.sourceBinId),
        note: trimOptional(draft.note)
      };
    });

  if (lines.length === 0) {
    throw new Error("Không còn dòng vật tư cần bàn giao.");
  }

  const normalizedEvidenceFileName = trimOptional(evidenceFileName);
  const evidence = normalizedEvidenceFileName
    ? [
        {
          evidenceType: "handover" as const,
          fileName: normalizedEvidenceFileName,
          note: "Biên bản bàn giao vật tư cho nhà máy"
        }
      ]
    : undefined;

  return {
    order,
    sourceWarehouseId: sourceWarehouseId.trim(),
    sourceWarehouseCode: sourceWarehouseCode.trim(),
    handoverBy: handoverBy.trim(),
    receivedBy: receivedBy.trim(),
    receiverContact: trimOptional(receiverContact),
    vehicleNo: trimOptional(vehicleNo),
    note: trimOptional(note),
    lines,
    evidence
  };
}

function toHandoverLine(line: SubcontractOrderMaterialLine, canIssue: boolean): FactoryMaterialHandoverLine {
  const remainingQty = remainingMaterialQty(line);
  const complete = !isPositiveQuantity(remainingQty);

  return {
    id: line.id,
    itemId: line.itemId,
    skuCode: line.skuCode,
    itemName: line.itemName,
    plannedQty: normalizeDecimalInput(line.plannedQty, decimalScales.quantity),
    issuedQty: normalizeDecimalInput(line.issuedQty, decimalScales.quantity),
    remainingQty,
    uomCode: line.uomCode,
    lotTraceRequired: line.lotTraceRequired,
    status: complete ? "complete" : canIssue ? "ready" : "pending",
    note: line.note
  };
}

function materialHandoverBlockedReason(order: SubcontractOrder) {
  if (order.materialLines.length === 0) {
    return "Lệnh chưa có dòng vật tư để bàn giao cho nhà máy.";
  }
  if (!isAtLeast(order.status, "factory_confirmed")) {
    return "Chờ nhà máy xác nhận trước khi bàn giao vật tư.";
  }
  if (!depositSatisfied(order)) {
    return "Chờ ghi nhận đặt cọc trước khi bàn giao vật tư cho nhà máy.";
  }
  if (!["factory_confirmed", "deposit_recorded"].includes(order.status)) {
    return "Lệnh đã qua bước bàn giao vật tư; không tạo thêm bàn giao từ màn này.";
  }

  return undefined;
}

function depositSatisfied(order: SubcontractOrder) {
  return order.depositStatus === "not_required" || order.depositStatus === "paid" || isAtLeast(order.status, "deposit_recorded");
}

function isMaterialLineComplete(line: SubcontractOrderMaterialLine) {
  return toScaledBigInt(line.issuedQty) >= toScaledBigInt(line.plannedQty);
}

function remainingMaterialQty(line: SubcontractOrderMaterialLine) {
  const remaining = toScaledBigInt(line.plannedQty) - toScaledBigInt(line.issuedQty);
  return formatScaledBigInt(remaining > BigInt(0) ? remaining : BigInt(0));
}

function isPositiveQuantity(value: string) {
  return toScaledBigInt(value) > BigInt(0);
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

function toScaledBigInt(value: string | number | null | undefined) {
  const normalized = normalizeDecimalInput(value, decimalScales.quantity);
  const [integerPart, fractionPart = ""] = normalized.split(".");
  return BigInt(`${integerPart}${fractionPart.padEnd(decimalScales.quantity, "0")}`);
}

function formatScaledBigInt(value: bigint) {
  const negative = value < BigInt(0);
  const unsigned = String(negative ? -value : value).padStart(decimalScales.quantity + 1, "0");
  const integerPart = unsigned.slice(0, -decimalScales.quantity);
  const fractionPart = unsigned.slice(-decimalScales.quantity);

  return `${negative ? "-" : ""}${integerPart}.${fractionPart}`;
}

function trimOptional(value?: string) {
  const trimmed = value?.trim();
  return trimmed ? trimmed : undefined;
}
