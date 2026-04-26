import type {
  CreateSubcontractMaterialTransferInput,
  SubcontractMaterialTransfer,
  SubcontractMaterialTransferLine,
  SubcontractMaterialTransferSummary,
  SubcontractStockMovement,
  SubcontractTransferAttachmentPlaceholder,
  SubcontractTransferAttachmentType,
  SubcontractTransferStatus
} from "../types";

export const subcontractTransferWarehouseOptions = [
  { label: "HCM Main Warehouse", value: "wh-hcm", code: "HCM" },
  { label: "HN Main Warehouse", value: "wh-hn", code: "HN" }
] as const;

export const prototypeTransferLines: SubcontractMaterialTransferLine[] = [
  {
    id: "line-cream-base",
    itemCode: "RM-CREAM-BASE",
    itemName: "Cream base bulk",
    itemType: "raw_material",
    quantity: 250,
    unit: "kg",
    lotControlled: true,
    batchNo: "RM-260426-A",
    qcStatus: "passed"
  },
  {
    id: "line-jar-50ml",
    itemCode: "PKG-JAR-50ML",
    itemName: "50ml jar packaging",
    itemType: "packaging",
    quantity: 5000,
    unit: "pcs",
    lotControlled: false,
    qcStatus: "passed"
  }
];

let transferSequence = 1;

export function createSubcontractMaterialTransfer(
  input: CreateSubcontractMaterialTransferInput
): SubcontractMaterialTransfer {
  const sourceWarehouseId = normalizeRequiredText(input.sourceWarehouseId, "Source warehouse is required");
  const sourceWarehouseCode = normalizeRequiredText(input.sourceWarehouseCode, "Source warehouse code is required");
  const factoryId = normalizeRequiredText(input.order.factoryId, "Factory is required");
  const factoryName = normalizeRequiredText(input.order.factoryName, "Factory name is required");
  const lines = normalizeTransferLines(input.lines);
  const sequence = transferSequence++;
  const id = `sub-transfer-260426-${String(sequence).padStart(4, "0")}`;
  const attachmentPlaceholders = createAttachmentPlaceholders(lines);
  const stockMovements = lines.map((line) =>
    createSubcontractIssueMovement({
      transferId: id,
      sourceWarehouseId,
      factoryCode: input.order.factoryCode,
      line
    })
  );

  return {
    id,
    transferNo: `SUBTR-260426-${String(sequence).padStart(4, "0")}`,
    orderId: input.order.id,
    orderNo: input.order.orderNo,
    sourceWarehouseId,
    sourceWarehouseCode,
    factoryId,
    factoryName,
    signedHandover: input.signedHandover,
    status: input.signedHandover ? "SENT" : "READY_TO_SEND",
    attachmentPlaceholders,
    lines,
    stockMovements,
    createdBy: input.createdBy?.trim() || "Subcontract Coordinator",
    createdAt: "2026-04-26T13:00:00Z"
  };
}

export function summarizeSubcontractMaterialTransfers(
  transfers: SubcontractMaterialTransfer[]
): SubcontractMaterialTransferSummary {
  return {
    total: transfers.length,
    signed: transfers.filter((transfer) => transfer.signedHandover).length,
    movementCount: transfers.reduce((sum, transfer) => sum + transfer.stockMovements.length, 0),
    attachmentPlaceholderCount: transfers.reduce(
      (sum, transfer) => sum + transfer.attachmentPlaceholders.length,
      0
    )
  };
}

export function subcontractTransferStatusTone(
  status: SubcontractTransferStatus
): "normal" | "success" | "warning" | "info" {
  switch (status) {
    case "SENT":
      return "success";
    case "READY_TO_SEND":
      return "info";
    case "DRAFT":
    default:
      return "warning";
  }
}

export function formatSubcontractTransferStatus(status: SubcontractTransferStatus) {
  switch (status) {
    case "READY_TO_SEND":
      return "Ready to send";
    case "SENT":
      return "Sent";
    case "DRAFT":
    default:
      return "Draft";
  }
}

export function formatSubcontractAttachmentType(type: SubcontractTransferAttachmentType) {
  switch (type) {
    case "COA":
      return "COA";
    case "MSDS":
      return "MSDS";
    case "LABEL":
      return "Label";
    case "VAT_INVOICE":
    default:
      return "VAT invoice";
  }
}

function normalizeTransferLines(lines: SubcontractMaterialTransferLine[]) {
  if (lines.length === 0) {
    throw new Error("At least one material or packaging line is required");
  }

  return lines.map((line) => {
    const itemCode = normalizeRequiredText(line.itemCode, "Item code is required");
    const itemName = normalizeRequiredText(line.itemName, "Item name is required");
    if (!Number.isFinite(line.quantity) || line.quantity <= 0) {
      throw new Error(`${itemCode} quantity must be greater than zero`);
    }
    if (line.lotControlled && !line.batchNo?.trim()) {
      throw new Error(`${itemCode} requires batch or lot before factory transfer`);
    }
    if (line.qcStatus !== "passed") {
      throw new Error(`${itemCode} must pass QC before factory transfer`);
    }

    return {
      ...line,
      itemCode,
      itemName,
      quantity: Math.round(line.quantity * 1000) / 1000,
      batchNo: line.batchNo?.trim() || undefined
    };
  });
}

function createAttachmentPlaceholders(
  lines: SubcontractMaterialTransferLine[]
): SubcontractTransferAttachmentPlaceholder[] {
  const hasRawMaterial = lines.some((line) => line.itemType === "raw_material");
  const hasPackaging = lines.some((line) => line.itemType === "packaging");

  return [
    { type: "COA", label: "COA", required: hasRawMaterial, attached: false },
    { type: "MSDS", label: "MSDS", required: hasRawMaterial, attached: false },
    { type: "LABEL", label: "Label", required: hasPackaging, attached: false },
    { type: "VAT_INVOICE", label: "VAT invoice", required: true, attached: false }
  ];
}

function createSubcontractIssueMovement({
  transferId,
  sourceWarehouseId,
  factoryCode,
  line
}: {
  transferId: string;
  sourceWarehouseId: string;
  factoryCode: string;
  line: SubcontractMaterialTransferLine;
}): SubcontractStockMovement {
  return {
    id: `mov-${transferId}-${line.id}`,
    movementType: "SUBCONTRACT_ISSUE",
    itemCode: line.itemCode,
    quantity: line.quantity,
    unit: line.unit,
    sourceWarehouseId,
    targetLocation: `stock_in_subcontractor_hold:${factoryCode}`,
    batchNo: line.batchNo,
    sourceDocId: transferId
  };
}

function normalizeRequiredText(value: string, message: string) {
  const normalized = value.trim();
  if (normalized === "") {
    throw new Error(message);
  }

  return normalized;
}
