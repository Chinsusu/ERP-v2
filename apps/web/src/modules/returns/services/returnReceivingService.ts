import type {
  InspectReturnInput,
  ReceiveReturnInput,
  ReturnInspectionCondition,
  ReturnInspectionDisposition,
  ReturnInspectionResult,
  ReturnInspectionStatus,
  ReturnDisposition,
  ReturnReceipt,
  ReturnReceiptLine,
  ReturnReceiptQuery,
  ReturnReceiptStatus,
  ReturnSource,
  ReturnStockMovement
} from "../types";

type ExpectedReturnRecord = {
  orderNo: string;
  trackingNo: string;
  returnCode: string;
  shipmentId: string;
  customerName: string;
  sku: string;
  productName: string;
  quantity: number;
  channel: string;
  batchNo: string;
  deliveredAt: string;
  returnReason: string;
  warehouseId: string;
  warehouseCode: string;
  source: ReturnSource;
};

export const returnWarehouseOptions = [
  { label: "HCM", value: "wh-hcm", code: "HCM" },
  { label: "HN", value: "wh-hn", code: "HN" }
] as const;

export const returnSourceOptions: { label: string; value: ReturnSource }[] = [
  { label: "Carrier", value: "CARRIER" },
  { label: "Shipper", value: "SHIPPER" },
  { label: "Customer", value: "CUSTOMER" },
  { label: "Marketplace", value: "MARKETPLACE" },
  { label: "Unknown", value: "UNKNOWN" }
];

export const returnDispositionOptions: { label: string; value: ReturnDisposition }[] = [
  { label: "Reusable", value: "reusable" },
  { label: "Not reusable", value: "not_reusable" },
  { label: "Needs inspection", value: "needs_inspection" }
];

export const returnInspectionConditionOptions: { label: string; value: ReturnInspectionCondition }[] = [
  { label: "Intact", value: "intact" },
  { label: "Dented box", value: "dented_box" },
  { label: "Seal torn", value: "seal_torn" },
  { label: "Used", value: "used" },
  { label: "Damaged", value: "damaged" },
  { label: "QA required", value: "qa_required" }
];

export const returnInspectionDispositionOptions: { label: string; value: ReturnInspectionDisposition }[] = [
  { label: "Usable", value: "usable" },
  { label: "Not usable", value: "not_usable" },
  { label: "QA hold", value: "qa_hold" }
];

export const prototypeReturnReceipts: ReturnReceipt[] = [
  createReturnReceipt({
    id: "rr-260426-0001",
    receiptNo: "RR-260426-0001",
    warehouseId: "wh-hcm",
    warehouseCode: "HCM",
    source: "CARRIER",
    receivedBy: "Return Inspector",
    receivedAt: "2026-04-26T08:30:00Z",
    packageCondition: "sealed bag",
    status: "pending_inspection",
    disposition: "needs_inspection",
    targetLocation: "return-inspection-queue",
    originalOrderNo: "SO-260425-099",
    trackingNo: "GHN260425099",
    returnCode: "RET-260425-099",
    channel: "Website",
    batchNo: "CREAM-260425-A",
    deliveredAt: "2026-04-24",
    returnReason: "Customer reported wrong shade",
    scanCode: "GHN260425099",
    customerName: "Tran Binh",
    unknownCase: false,
    lines: [
      {
        id: "line-cream-50ml",
        sku: "CREAM-50ML",
        productName: "Repair Cream 50ml",
        quantity: 1,
        condition: "sealed bag"
      }
    ],
    investigationNote: "Customer reported wrong shade",
    createdAt: "2026-04-26T08:30:00Z"
  })
];

const expectedReturnRecords: ExpectedReturnRecord[] = [
  {
    orderNo: "SO-260426-001",
    trackingNo: "GHN260426001",
    returnCode: "RET-260426-001",
    shipmentId: "ship-hcm-260426-001",
    customerName: "Nguyen An",
    sku: "SERUM-30ML",
    productName: "Hydrating Serum 30ml",
    quantity: 1,
    channel: "Website",
    batchNo: "SER-260426-A",
    deliveredAt: "2026-04-24",
    returnReason: "Customer refused delivery",
    warehouseId: "wh-hcm",
    warehouseCode: "HCM",
    source: "CARRIER"
  },
  {
    orderNo: "SO-260426-004",
    trackingNo: "GHN260426004",
    returnCode: "RET-260426-004",
    shipmentId: "ship-hcm-260426-004",
    customerName: "Le Chi",
    sku: "TONER-100ML",
    productName: "Balancing Toner 100ml",
    quantity: 2,
    channel: "Marketplace",
    batchNo: "TON-260426-B",
    deliveredAt: "2026-04-25",
    returnReason: "Damaged packaging",
    warehouseId: "wh-hcm",
    warehouseCode: "HCM",
    source: "SHIPPER"
  },
  {
    orderNo: "SO-260426-HN-011",
    trackingNo: "GHNHN260426001",
    returnCode: "RET-HN-260426-011",
    shipmentId: "ship-hn-260426-001",
    customerName: "Pham Ha",
    sku: "MASK-SET-05",
    productName: "Sheet Mask Set",
    quantity: 1,
    channel: "TikTok Shop",
    batchNo: "MASK-260426-HN",
    deliveredAt: "2026-04-25",
    returnReason: "Wrong item claim",
    warehouseId: "wh-hn",
    warehouseCode: "HN",
    source: "MARKETPLACE"
  }
];

let receiptSequence = 1;

export async function getReturnReceipts(query: ReturnReceiptQuery = {}): Promise<ReturnReceipt[]> {
  return prototypeReturnReceipts
    .filter((receipt) => {
      if (query.warehouseId && receipt.warehouseId !== query.warehouseId) {
        return false;
      }
      if (query.status && receipt.status !== query.status) {
        return false;
      }

      return true;
    })
    .sort(sortReceipts);
}

export async function receiveReturn(input: ReceiveReturnInput): Promise<ReturnReceipt> {
  const scanCode = normalizeReturnScanCode(input.code);
  if (scanCode === "") {
    throw new Error("Return scan code is required");
  }

  const disposition = normalizeDisposition(input.disposition);
  const expected = expectedReturnRecords.find((candidate) => matchesExpectedReturn(candidate, scanCode));
  const warehouse = returnWarehouseOptions.find((option) => option.value === (expected?.warehouseId ?? input.warehouseId));
  const warehouseId = expected?.warehouseId ?? input.warehouseId;
  const warehouseCode = expected?.warehouseCode ?? warehouse?.code ?? input.warehouseCode;
  const packageCondition = input.packageCondition.trim() || "pending inspection";
  const receiptNo = expected ? `RR-${expected.orderNo.replace("SO-", "")}` : `RR-UNKNOWN-${scanCode}`;
  const id = `${receiptNo.toLowerCase()}-${receiptSequence++}`;
  const lines = createReceiptLines(expected, packageCondition);
  const stockMovement = disposition === "reusable" ? createReturnReceiptMovement(id, warehouseId, lines[0]) : undefined;

  return createReturnReceipt({
    id,
    receiptNo,
    warehouseId,
    warehouseCode,
    source: input.source || expected?.source || "UNKNOWN",
    receivedBy: "Return Inspector",
    receivedAt: "2026-04-26T10:30:00Z",
    packageCondition,
    status: "pending_inspection",
    disposition,
    targetLocation: targetLocationForDisposition(disposition),
    originalOrderNo: expected?.orderNo,
    trackingNo: expected?.trackingNo ?? scanCode,
    returnCode: expected?.returnCode,
    channel: expected?.channel,
    batchNo: expected?.batchNo,
    deliveredAt: expected?.deliveredAt,
    returnReason: expected?.returnReason,
    scanCode,
    customerName: expected?.customerName ?? "Unknown customer",
    unknownCase: !expected,
    lines,
    stockMovement,
    investigationNote: expected ? input.investigationNote : input.investigationNote || "Unknown return case created from receiving scan",
    auditLogId: "audit-return-receipt-prototype",
    createdAt: "2026-04-26T10:30:00Z"
  });
}

export function returnReceiptStatusTone(status: ReturnReceiptStatus): "warning" | "normal" {
  return status === "pending_inspection" ? "warning" : "normal";
}

export function returnDispositionTone(
  disposition: ReturnDisposition
): "success" | "warning" | "danger" | "info" {
  switch (disposition) {
    case "reusable":
      return "success";
    case "not_reusable":
      return "danger";
    case "needs_inspection":
    default:
      return "warning";
  }
}

export function createReturnInspection(input: InspectReturnInput): ReturnInspectionResult {
  const disposition = normalizeInspectionDisposition(input.disposition);
  const condition = normalizeInspectionCondition(input.condition);
  const status: ReturnInspectionStatus = disposition === "qa_hold" ? "RETURN_QA_HOLD" : "INSPECTION_RECORDED";

  return {
    id: `inspect-${input.receipt.id}-${condition}-${disposition}`,
    receiptId: input.receipt.id,
    receiptNo: input.receipt.receiptNo,
    condition,
    disposition,
    status,
    targetLocation: inspectionTargetLocation(disposition),
    riskLevel: inspectionRiskLevel(condition, disposition),
    inspector: "Return Inspector",
    note: input.note?.trim() || undefined,
    evidenceLabel: input.evidenceLabel?.trim() || undefined,
    inspectedAt: "2026-04-26T11:00:00Z"
  };
}

export function matchesReturnReceiptCode(receipt: ReturnReceipt, code: string) {
  const scanCode = normalizeReturnScanCode(code);
  if (scanCode === "") {
    return false;
  }

  return [
    receipt.id,
    receipt.receiptNo,
    receipt.originalOrderNo,
    receipt.trackingNo,
    receipt.returnCode,
    receipt.scanCode
  ].some((value) => normalizeReturnScanCode(value ?? "") === scanCode);
}

export function returnInspectionConditionTone(
  condition: ReturnInspectionCondition
): "normal" | "success" | "warning" | "danger" | "info" {
  switch (condition) {
    case "intact":
      return "success";
    case "damaged":
      return "danger";
    case "qa_required":
      return "info";
    case "dented_box":
    case "seal_torn":
    case "used":
    default:
      return "warning";
  }
}

export function returnInspectionDispositionTone(
  disposition: ReturnInspectionDisposition
): "success" | "warning" | "danger" {
  switch (disposition) {
    case "usable":
      return "success";
    case "not_usable":
      return "danger";
    case "qa_hold":
    default:
      return "warning";
  }
}

export function returnInspectionStatusTone(status: ReturnInspectionStatus): "success" | "warning" {
  return status === "RETURN_QA_HOLD" ? "warning" : "success";
}

export function formatReturnInspectionCondition(condition: ReturnInspectionCondition) {
  return returnInspectionConditionOptions.find((option) => option.value === condition)?.label ?? condition;
}

export function formatReturnInspectionDisposition(disposition: ReturnInspectionDisposition) {
  return returnInspectionDispositionOptions.find((option) => option.value === disposition)?.label ?? disposition;
}

export function formatReturnDisposition(disposition: ReturnDisposition) {
  return returnDispositionOptions.find((option) => option.value === disposition)?.label ?? disposition;
}

function createReturnReceipt(input: ReturnReceipt): ReturnReceipt {
  return {
    ...input,
    lines: input.lines.map((line) => ({ ...line })),
    stockMovement: input.stockMovement ? { ...input.stockMovement } : undefined
  };
}

function createReceiptLines(expected: ExpectedReturnRecord | undefined, packageCondition: string): ReturnReceiptLine[] {
  if (!expected) {
    return [
      {
        id: "line-unknown-return",
        sku: "UNKNOWN-SKU",
        productName: "Unknown return item",
        quantity: 1,
        condition: packageCondition
      }
    ];
  }

  return [
    {
      id: `line-${expected.sku.toLowerCase()}`,
      sku: expected.sku,
      productName: expected.productName,
      quantity: expected.quantity,
      condition: packageCondition
    }
  ];
}

function createReturnReceiptMovement(
  receiptId: string,
  warehouseId: string,
  line: ReturnReceiptLine
): ReturnStockMovement {
  return {
    id: `mov-${receiptId}`,
    movementType: "RETURN_RECEIPT",
    sku: line.sku,
    warehouseId,
    quantity: line.quantity,
    targetStockStatus: "return_pending",
    sourceDocId: receiptId
  };
}

function targetLocationForDisposition(disposition: ReturnDisposition) {
  switch (disposition) {
    case "reusable":
      return "return-area-pending-inspection";
    case "not_reusable":
      return "lab-damaged-placeholder";
    case "needs_inspection":
    default:
      return "return-inspection-queue";
  }
}

function normalizeInspectionCondition(condition: ReturnInspectionCondition): ReturnInspectionCondition {
  if (returnInspectionConditionOptions.some((option) => option.value === condition)) {
    return condition;
  }

  return "qa_required";
}

function normalizeInspectionDisposition(disposition: ReturnInspectionDisposition): ReturnInspectionDisposition {
  if (returnInspectionDispositionOptions.some((option) => option.value === disposition)) {
    return disposition;
  }

  return "qa_hold";
}

function inspectionTargetLocation(disposition: ReturnInspectionDisposition) {
  switch (disposition) {
    case "usable":
      return "return-area-qc-release";
    case "not_usable":
      return "lab-damaged-placeholder";
    case "qa_hold":
    default:
      return "return-qa-hold";
  }
}

function inspectionRiskLevel(condition: ReturnInspectionCondition, disposition: ReturnInspectionDisposition) {
  if (condition === "damaged" || disposition === "not_usable") {
    return "high";
  }
  if (condition === "qa_required" || disposition === "qa_hold" || condition === "used") {
    return "medium";
  }

  return "low";
}

function matchesExpectedReturn(candidate: ExpectedReturnRecord, scanCode: string) {
  return [candidate.orderNo, candidate.trackingNo, candidate.returnCode, candidate.shipmentId].some(
    (value) => normalizeReturnScanCode(value) === scanCode
  );
}

function normalizeDisposition(disposition: ReturnDisposition): ReturnDisposition {
  if (returnDispositionOptions.some((option) => option.value === disposition)) {
    return disposition;
  }

  return "needs_inspection";
}

function normalizeReturnScanCode(code: string) {
  return code.trim().toUpperCase();
}

function sortReceipts(left: ReturnReceipt, right: ReturnReceipt) {
  if (left.createdAt !== right.createdAt) {
    return right.createdAt.localeCompare(left.createdAt);
  }

  return right.receiptNo.localeCompare(left.receiptNo);
}
