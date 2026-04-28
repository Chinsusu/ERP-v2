import { ApiError, apiGet, apiPost, apiPostForm } from "../../../shared/api/client";
import type { components, operations } from "../../../shared/api/generated/schema";
import type {
  ApplyReturnDispositionInput,
  InspectReturnInput,
  ReceiveReturnInput,
  ReturnAttachment,
  ReturnDispositionAction,
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
  ReturnStockMovement,
  UploadReturnAttachmentInput
} from "../types";

type ReturnReceiptApi = components["schemas"]["ReturnReceipt"];
type ReturnReceiptApiQuery = operations["listReturnReceipts"]["parameters"]["query"];
type ScanReturnApiRequest = components["schemas"]["ScanReturnRequest"];
type ReturnInspectionApi = components["schemas"]["ReturnInspection"];
type InspectReturnApiRequest = components["schemas"]["InspectReturnRequest"];
type ReturnDispositionActionApi = components["schemas"]["ReturnDispositionAction"];
type ApplyReturnDispositionApiRequest = components["schemas"]["ApplyReturnDispositionRequest"];
type ReturnAttachmentApi = components["schemas"]["ReturnAttachment"];

type ExpectedReturnRecord = {
  orderNo: string;
  orderStatus: "handed_over" | "delivered" | "waiting_handover";
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

const defaultAccessToken = "local-dev-access-token";

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
  { label: "Missing accessory", value: "missing_accessory" }
];

export const returnInspectionDispositionOptions: { label: string; value: ReturnInspectionDisposition }[] = [
  { label: "Reusable", value: "reusable" },
  { label: "Not reusable", value: "not_reusable" },
  { label: "Needs QA", value: "needs_inspection" }
];

export let prototypeReturnReceipts: ReturnReceipt[] = createPrototypeReturnReceipts();

const expectedReturnRecords: ExpectedReturnRecord[] = [
  {
    orderNo: "SO-260426-001",
    orderStatus: "handed_over",
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
    orderStatus: "delivered",
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
    orderStatus: "handed_over",
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
  },
  {
    orderNo: "SO-260426-009",
    orderStatus: "waiting_handover",
    trackingNo: "GHN260426009",
    returnCode: "RET-260426-009",
    shipmentId: "ship-hcm-260426-009",
    customerName: "Vu Nhi",
    sku: "SERUM-30ML",
    productName: "Hydrating Serum 30ml",
    quantity: 1,
    channel: "Website",
    batchNo: "SER-260426-WAIT",
    deliveredAt: "",
    returnReason: "Premature return scan",
    warehouseId: "wh-hcm",
    warehouseCode: "HCM",
    source: "CARRIER"
  }
];

export async function getReturnReceipts(query: ReturnReceiptQuery = {}): Promise<ReturnReceipt[]> {
  try {
    const receipts = await apiGet("/returns/receipts", {
      accessToken: defaultAccessToken,
      query: toApiQuery(query)
    });

    return receipts.map(fromApiReturnReceipt).sort(sortReceipts);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return filterPrototypeReturnReceipts(query);
  }
}

export async function receiveReturn(input: ReceiveReturnInput): Promise<ReturnReceipt> {
  try {
    const receipt = await apiPost<ReturnReceiptApi, ScanReturnApiRequest>(
      "/returns/scan",
      toApiScanInput(input),
      { accessToken: defaultAccessToken }
    );

    return fromApiReturnReceipt(receipt);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return receivePrototypeReturn(input);
  }
}

export async function inspectReturn(input: InspectReturnInput): Promise<ReturnInspectionResult> {
  try {
    const inspection = await apiPost<ReturnInspectionApi, InspectReturnApiRequest>(
      `/returns/${input.receipt.id}/inspect`,
      toApiInspectInput(input),
      { accessToken: defaultAccessToken }
    );

    return fromApiReturnInspection(inspection);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    const result = createReturnInspection(input);
    prototypeReturnReceipts = [
      applyInspectionToReceipt(input.receipt, result),
      ...prototypeReturnReceipts.filter((candidate) => candidate.id !== input.receipt.id)
    ];

    return result;
  }
}

export async function applyReturnDisposition(input: ApplyReturnDispositionInput): Promise<ReturnDispositionAction> {
  try {
    const action = await apiPost<ReturnDispositionActionApi, ApplyReturnDispositionApiRequest>(
      `/returns/${input.receipt.id}/disposition`,
      toApiDispositionInput(input),
      { accessToken: defaultAccessToken }
    );

    return fromApiReturnDispositionAction(action);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    const action = createReturnDispositionAction(input);
    prototypeReturnReceipts = [
      applyDispositionToReceipt(input.receipt, action),
      ...prototypeReturnReceipts.filter((candidate) => candidate.id !== input.receipt.id)
    ];

    return action;
  }
}

export async function uploadReturnAttachment(input: UploadReturnAttachmentInput): Promise<ReturnAttachment> {
  try {
    const form = new FormData();
    form.set("inspection_id", input.inspectionId);
    form.set("file", input.file);
    if (input.note?.trim()) {
      form.set("note", input.note.trim());
    }

    const attachment = await apiPostForm<ReturnAttachmentApi>(
      `/returns/${input.receipt.id}/attachments`,
      form,
      { accessToken: defaultAccessToken }
    );

    return fromApiReturnAttachment(attachment);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return createReturnAttachment(input);
  }
}

export function resetPrototypeReturnReceiptsForTest() {
  prototypeReturnReceipts = createPrototypeReturnReceipts();
}

function receivePrototypeReturn(input: ReceiveReturnInput): ReturnReceipt {
  const scanCode = normalizeReturnScanCode(input.code);
  if (scanCode === "") {
    throw new Error("Return scan code is required");
  }

  const disposition = normalizeDisposition(input.disposition);
  const expected = expectedReturnRecords.find((candidate) => matchesExpectedReturn(candidate, scanCode));
  if (expected && !isReturnReceivableOrderStatus(expected.orderStatus)) {
    throw new Error("Order status is not eligible for return receiving");
  }
  const warehouse = returnWarehouseOptions.find((option) => option.value === (expected?.warehouseId ?? input.warehouseId));
  const warehouseId = expected?.warehouseId ?? input.warehouseId;
  const warehouseCode = expected?.warehouseCode ?? warehouse?.code ?? input.warehouseCode;
  const packageCondition = input.packageCondition.trim() || "pending inspection";
  const receiptNo = expected ? `RR-${expected.orderNo.replace("SO-", "")}` : `RR-UNKNOWN-${scanCode}`;
  const duplicateIdentifiers = [
    receiptNo,
    scanCode,
    expected?.orderNo,
    expected?.trackingNo,
    expected?.returnCode,
    expected?.shipmentId
  ];
  if (
    prototypeReturnReceipts.some((receipt) =>
      duplicateIdentifiers.some((identifier) => matchesReturnReceiptCode(receipt, identifier ?? ""))
    )
  ) {
    throw new Error("Return receipt already exists for this scan");
  }

  const id = receiptNo.toLowerCase();
  const lines = createReceiptLines(expected, packageCondition);
  const stockMovement = disposition === "reusable" ? createReturnReceiptMovement(id, warehouseId, lines[0]) : undefined;

  const receipt = createReturnReceipt({
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
  prototypeReturnReceipts = [receipt, ...prototypeReturnReceipts.filter((candidate) => candidate.id !== receipt.id)];

  return createReturnReceipt(receipt);
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
  const status: ReturnInspectionStatus = disposition === "needs_inspection" ? "return_qa_hold" : "inspection_recorded";

  return {
    id: `inspect-${input.receipt.id}-${condition}-${disposition}`,
    receiptId: input.receipt.id,
    receiptNo: input.receipt.receiptNo,
    condition,
    disposition,
    status,
    targetLocation: inspectionTargetLocation(disposition),
    riskLevel: inspectionRiskLevel(condition, disposition),
    inspectorId: "user-return-inspector",
    note: input.note?.trim() || undefined,
    evidenceLabel: input.evidenceLabel?.trim() || undefined,
    inspectedAt: "2026-04-26T11:00:00Z"
  };
}

export function createReturnDispositionAction(input: ApplyReturnDispositionInput): ReturnDispositionAction {
  const disposition = normalizeDisposition(input.disposition);

  return {
    id: `dispose-${input.receipt.id}-${disposition}`,
    receiptId: input.receipt.id,
    receiptNo: input.receipt.receiptNo,
    disposition,
    targetLocation: dispositionTargetLocation(disposition),
    targetStockStatus: dispositionTargetStockStatus(disposition),
    actionCode: dispositionActionCode(disposition),
    actorId: "user-return-inspector",
    note: input.note?.trim() || undefined,
    auditLogId: "audit-return-disposition-prototype",
    decidedAt: "2026-04-26T11:30:00Z"
  };
}

export function createReturnAttachment(input: UploadReturnAttachmentInput): ReturnAttachment {
  const fileName = input.file.name.trim();
  const fileExt = fileName.includes(".") ? fileName.split(".").pop()?.toLowerCase() : undefined;

  return {
    id: `attach-${input.receipt.id}-${input.inspectionId}-${fileName.toLowerCase().replace(/[^a-z0-9._-]+/g, "-")}`,
    receiptId: input.receipt.id,
    receiptNo: input.receipt.receiptNo,
    inspectionId: input.inspectionId,
    fileName,
    fileExt,
    mimeType: input.file.type || "application/octet-stream",
    fileSizeBytes: input.file.size,
    storageBucket: "erp-return-attachments-dev",
    storageKey: `returns/${input.receipt.id}/inspections/${input.inspectionId}/${fileName}`,
    status: "active",
    uploadedBy: "user-return-inspector",
    note: input.note?.trim() || undefined,
    auditLogId: "audit-return-attachment-prototype",
    uploadedAt: "2026-04-26T11:45:00Z"
  };
}

export function applyInspectionToReceipt(receipt: ReturnReceipt, inspection: ReturnInspectionResult): ReturnReceipt {
  return createReturnReceipt({
    ...receipt,
    status: "inspected",
    packageCondition: inspection.condition,
    disposition: inspection.disposition,
    targetLocation: inspection.targetLocation,
    lines: receipt.lines.map((line) => ({ ...line, condition: inspection.condition })),
    stockMovement: undefined
  });
}

export function applyDispositionToReceipt(receipt: ReturnReceipt, action: ReturnDispositionAction): ReturnReceipt {
  return createReturnReceipt({
    ...receipt,
    status: "dispositioned",
    disposition: action.disposition,
    targetLocation: action.targetLocation,
    stockMovement: undefined
  });
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
    case "missing_accessory":
      return "danger";
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
    case "reusable":
      return "success";
    case "not_reusable":
      return "danger";
    case "needs_inspection":
    default:
      return "warning";
  }
}

export function returnInspectionStatusTone(status: ReturnInspectionStatus): "success" | "warning" {
  return status === "return_qa_hold" ? "warning" : "success";
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

function shouldUsePrototypeFallback(reason: unknown) {
  if (reason instanceof ApiError) {
    return false;
  }

  return !(reason instanceof Error && reason.message.startsWith("API request failed:"));
}

function fromApiReturnReceipt(receipt: ReturnReceiptApi): ReturnReceipt {
  return createReturnReceipt({
    id: receipt.id,
    receiptNo: receipt.receipt_no,
    warehouseId: receipt.warehouse_id,
    warehouseCode: receipt.warehouse_code,
    source: receipt.source,
    receivedBy: receipt.received_by,
    receivedAt: receipt.received_at,
    packageCondition: receipt.package_condition,
    status: receipt.status,
    disposition: receipt.disposition,
    targetLocation: receipt.target_location,
    originalOrderNo: receipt.original_order_no,
    trackingNo: receipt.tracking_no,
    returnCode: receipt.return_code,
    scanCode: receipt.scan_code,
    customerName: receipt.customer_name,
    unknownCase: receipt.unknown_case,
    lines: receipt.lines.map((line) => ({
      id: line.id,
      sku: line.sku,
      productName: line.product_name,
      quantity: line.quantity,
      condition: line.condition
    })),
    stockMovement: receipt.stock_movement
      ? {
          id: receipt.stock_movement.id,
          movementType: receipt.stock_movement.movement_type,
          sku: receipt.stock_movement.sku,
          warehouseId: receipt.stock_movement.warehouse_id,
          quantity: receipt.stock_movement.quantity,
          targetStockStatus: receipt.stock_movement.target_stock_status,
          sourceDocId: receipt.stock_movement.source_doc_id
        }
      : undefined,
    investigationNote: receipt.investigation_note,
    auditLogId: receipt.audit_log_id,
    createdAt: receipt.created_at
  });
}

function fromApiReturnInspection(inspection: ReturnInspectionApi): ReturnInspectionResult {
  return {
    id: inspection.id,
    receiptId: inspection.receipt_id,
    receiptNo: inspection.receipt_no,
    condition: inspection.condition,
    disposition: inspection.disposition,
    status: inspection.status,
    targetLocation: inspection.target_location,
    riskLevel: inspection.risk_level,
    inspectorId: inspection.inspector_id,
    note: inspection.note,
    evidenceLabel: inspection.evidence_label,
    inspectedAt: inspection.inspected_at
  };
}

function fromApiReturnDispositionAction(action: ReturnDispositionActionApi): ReturnDispositionAction {
  return {
    id: action.id,
    receiptId: action.receipt_id,
    receiptNo: action.receipt_no,
    disposition: action.disposition,
    targetLocation: action.target_location,
    targetStockStatus: action.target_stock_status,
    actionCode: action.action_code,
    actorId: action.actor_id,
    note: action.note,
    auditLogId: action.audit_log_id,
    decidedAt: action.decided_at
  };
}

function fromApiReturnAttachment(attachment: ReturnAttachmentApi): ReturnAttachment {
  return {
    id: attachment.id,
    receiptId: attachment.receipt_id,
    receiptNo: attachment.receipt_no,
    inspectionId: attachment.inspection_id,
    fileName: attachment.file_name,
    fileExt: attachment.file_ext,
    mimeType: attachment.mime_type,
    fileSizeBytes: attachment.file_size_bytes,
    storageBucket: attachment.storage_bucket,
    storageKey: attachment.storage_key,
    status: attachment.status,
    uploadedBy: attachment.uploaded_by,
    note: attachment.note,
    auditLogId: attachment.audit_log_id,
    uploadedAt: attachment.uploaded_at
  };
}

function toApiQuery(query: ReturnReceiptQuery): ReturnReceiptApiQuery {
  return {
    warehouse_id: query.warehouseId,
    status: query.status
  };
}

function toApiScanInput(input: ReceiveReturnInput): ScanReturnApiRequest {
  return {
    warehouse_id: input.warehouseId,
    warehouse_code: input.warehouseCode,
    source: input.source,
    code: input.code,
    package_condition: input.packageCondition,
    disposition: input.disposition,
    investigation_note: input.investigationNote
  };
}

function toApiInspectInput(input: InspectReturnInput): InspectReturnApiRequest {
  return {
    condition: input.condition,
    disposition: input.disposition,
    note: input.note,
    evidence_label: input.evidenceLabel
  };
}

function toApiDispositionInput(input: ApplyReturnDispositionInput): ApplyReturnDispositionApiRequest {
  return {
    disposition: input.disposition,
    note: input.note
  };
}

function filterPrototypeReturnReceipts(query: ReturnReceiptQuery): ReturnReceipt[] {
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
    .map(createReturnReceipt)
    .sort(sortReceipts);
}

function createPrototypeReturnReceipts(): ReturnReceipt[] {
  return [
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

  return "missing_accessory";
}

function normalizeInspectionDisposition(disposition: ReturnInspectionDisposition): ReturnInspectionDisposition {
  if (returnInspectionDispositionOptions.some((option) => option.value === disposition)) {
    return disposition;
  }

  return "needs_inspection";
}

function inspectionTargetLocation(disposition: ReturnInspectionDisposition) {
  switch (disposition) {
    case "reusable":
      return "return-area-qc-release";
    case "not_reusable":
      return "lab-damaged-placeholder";
    case "needs_inspection":
    default:
      return "return-qa-hold";
  }
}

function inspectionRiskLevel(condition: ReturnInspectionCondition, disposition: ReturnInspectionDisposition) {
  if (condition === "damaged" || condition === "missing_accessory" || disposition === "not_reusable") {
    return "high";
  }
  if (condition === "seal_torn" || disposition === "needs_inspection" || condition === "used") {
    return "medium";
  }

  return "low";
}

function dispositionTargetLocation(disposition: ReturnDisposition) {
  switch (disposition) {
    case "reusable":
      return "return-putaway-ready";
    case "not_reusable":
      return "lab-damaged-placeholder";
    case "needs_inspection":
    default:
      return "return-quarantine-hold";
  }
}

function dispositionTargetStockStatus(disposition: ReturnDisposition): ReturnDispositionAction["targetStockStatus"] {
  switch (disposition) {
    case "reusable":
      return "return_pending";
    case "not_reusable":
      return "damaged";
    case "needs_inspection":
    default:
      return "qc_hold";
  }
}

function dispositionActionCode(disposition: ReturnDisposition): ReturnDispositionAction["actionCode"] {
  switch (disposition) {
    case "reusable":
      return "route_to_putaway";
    case "not_reusable":
      return "route_to_lab_or_damaged";
    case "needs_inspection":
    default:
      return "route_to_quarantine_hold";
  }
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

function isReturnReceivableOrderStatus(status: ExpectedReturnRecord["orderStatus"]) {
  return status === "handed_over" || status === "delivered";
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
