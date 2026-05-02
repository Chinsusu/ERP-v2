import { apiGetRaw, apiPost } from "../../../shared/api/client";
import { shouldUsePrototypeFallback } from "../../../shared/api/prototypeFallback";
import type { components, operations } from "../../../shared/api/generated/schema";
import { decimalScales, formatDateTimeVI, formatQuantity, normalizeDecimalInput } from "../../../shared/format/numberFormat";
import type {
  CreateInboundQCInspectionInput,
  InboundQCActionResult,
  InboundQCChecklistItem,
  InboundQCChecklistStatus,
  InboundQCDecisionInput,
  InboundQCInspection,
  InboundQCInspectionQuery,
  InboundQCInspectionStatus,
  InboundQCResult
} from "../types";

type InboundQCChecklistItemApi = components["schemas"]["InboundQCChecklistItem"];
type InboundQCInspectionApi = components["schemas"]["InboundQCInspection"];
type CreateInboundQCInspectionApiRequest = components["schemas"]["CreateInboundQCInspectionRequest"];
type InboundQCDecisionApiRequest = components["schemas"]["InboundQCDecisionRequest"];
type InboundQCActionResultApi = components["schemas"]["InboundQCActionResult"];
type InboundQCInspectionListApiQuery = operations["listInboundQCInspections"]["parameters"]["query"];

type DecisionAction = "pass" | "fail" | "hold" | "partial";

type PrototypeInspectableLine = {
  goodsReceiptId: string;
  goodsReceiptNo: string;
  goodsReceiptLineId: string;
  purchaseOrderId: string;
  purchaseOrderLineId: string;
  itemId: string;
  sku: string;
  itemName: string;
  batchId: string;
  batchNo: string;
  lotNo: string;
  expiryDate: string;
  warehouseId: string;
  locationId: string;
  quantity: string;
  uomCode: string;
};

const defaultAccessToken = "local-dev-access-token";
const prototypeNow = "2026-04-27T11:00:00Z";
const zeroQuantity = "0.000000";

export const inboundQCStatusOptions: { label: string; value: "" | InboundQCInspectionStatus }[] = [
  { label: "All statuses", value: "" },
  { label: "Pending", value: "pending" },
  { label: "In progress", value: "in_progress" },
  { label: "Completed", value: "completed" },
  { label: "Cancelled", value: "cancelled" }
];

export const inboundQCChecklistStatusOptions: { label: string; value: InboundQCChecklistStatus }[] = [
  { label: "Pending", value: "pending" },
  { label: "Pass", value: "pass" },
  { label: "Fail", value: "fail" },
  { label: "N/A", value: "not_applicable" }
];

export const defaultInboundQCChecklist: InboundQCChecklistItem[] = [
  {
    id: "check-packaging",
    code: "PACKAGING",
    label: "Packaging condition",
    required: true,
    status: "pending"
  },
  {
    id: "check-lot-expiry",
    code: "LOT_EXPIRY",
    label: "Lot, expiry, and supplier label",
    required: true,
    status: "pending"
  },
  {
    id: "check-coa-msds",
    code: "COA_MSDS",
    label: "COA / MSDS attached when required",
    required: true,
    status: "pending"
  },
  {
    id: "check-photo-evidence",
    code: "PHOTO_EVIDENCE",
    label: "QC photos captured",
    required: false,
    status: "pending"
  },
  {
    id: "check-sample",
    code: "SAMPLE",
    label: "Sample retained when required",
    required: false,
    status: "pending"
  }
];

const prototypeInspectableLines: PrototypeInspectableLine[] = [
  {
    goodsReceiptId: "grn-hcm-260427-inspect",
    goodsReceiptNo: "GRN-260427-0003",
    goodsReceiptLineId: "grn-line-inspect-001",
    purchaseOrderId: "po-260429-0003",
    purchaseOrderLineId: "po-260429-0003-line-01",
    itemId: "item-cream-50g",
    sku: "CREAM-50G",
    itemName: "Moisturizing Cream",
    batchId: "batch-cream-2603b",
    batchNo: "LOT-2603B",
    lotNo: "LOT-2603B",
    expiryDate: "2028-03-01",
    warehouseId: "wh-hcm-fg",
    locationId: "loc-hcm-fg-recv-01",
    quantity: "12.000000",
    uomCode: "EA"
  }
];

let prototypeInspections = createPrototypeInspections();

export async function getInboundQCInspections(
  query: InboundQCInspectionQuery = {}
): Promise<InboundQCInspection[]> {
  try {
    const inspections = await apiGetRaw<InboundQCInspectionApi[]>(
      `/inbound-qc-inspections${inboundQCQueryString(query)}`,
      { accessToken: defaultAccessToken }
    );

    return inspections.map(fromApiInspection);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return filterPrototypeInspections(query);
  }
}

export async function createInboundQCInspection(
  input: CreateInboundQCInspectionInput
): Promise<InboundQCActionResult> {
  try {
    const result = await apiPost<InboundQCActionResultApi, CreateInboundQCInspectionApiRequest>(
      "/inbound-qc-inspections",
      toApiCreateInput(input),
      { accessToken: defaultAccessToken }
    );

    return fromApiActionResult(result);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return createPrototypeInspection(input);
  }
}

export function startInboundQCInspection(id: string): Promise<InboundQCActionResult> {
  return actionInboundQCInspection(id, "start", {});
}

export function passInboundQCInspection(id: string, input: InboundQCDecisionInput): Promise<InboundQCActionResult> {
  return actionInboundQCInspection(id, "pass", input);
}

export function failInboundQCInspection(id: string, input: InboundQCDecisionInput): Promise<InboundQCActionResult> {
  return actionInboundQCInspection(id, "fail", input);
}

export function holdInboundQCInspection(id: string, input: InboundQCDecisionInput): Promise<InboundQCActionResult> {
  return actionInboundQCInspection(id, "hold", input);
}

export function partialInboundQCInspection(id: string, input: InboundQCDecisionInput): Promise<InboundQCActionResult> {
  return actionInboundQCInspection(id, "partial", input);
}

export function resetPrototypeInboundQCInspectionsForTest() {
  prototypeInspections = createPrototypeInspections();
}

export function inboundQCStatusTone(status: InboundQCInspectionStatus) {
  switch (status) {
    case "completed":
      return "success" as const;
    case "in_progress":
      return "info" as const;
    case "cancelled":
      return "normal" as const;
    case "pending":
    default:
      return "warning" as const;
  }
}

export function inboundQCResultTone(result?: InboundQCResult) {
  switch (result) {
    case "pass":
      return "success" as const;
    case "fail":
      return "danger" as const;
    case "hold":
    case "partial":
      return "warning" as const;
    default:
      return "normal" as const;
  }
}

export function checklistStatusTone(status: InboundQCChecklistStatus) {
  switch (status) {
    case "pass":
      return "success" as const;
    case "fail":
      return "danger" as const;
    case "not_applicable":
      return "normal" as const;
    case "pending":
    default:
      return "warning" as const;
  }
}

export function formatInboundQCStatus(status: InboundQCInspectionStatus) {
  switch (status) {
    case "in_progress":
      return "In progress";
    case "completed":
      return "Completed";
    case "cancelled":
      return "Cancelled";
    case "pending":
    default:
      return "Pending";
  }
}

export function formatInboundQCResult(result?: InboundQCResult) {
  switch (result) {
    case "pass":
      return "Pass";
    case "fail":
      return "Fail";
    case "hold":
      return "Hold";
    case "partial":
      return "Partial";
    default:
      return "-";
  }
}

export function formatChecklistStatus(status: InboundQCChecklistStatus) {
  switch (status) {
    case "pass":
      return "Pass";
    case "fail":
      return "Fail";
    case "not_applicable":
      return "N/A";
    case "pending":
    default:
      return "Pending";
  }
}

export function formatInboundQCQuantity(value: string, uomCode?: string) {
  return formatQuantity(value, uomCode);
}

export function formatInboundQCDateTime(value?: string) {
  return value ? formatDateTimeVI(value) : "-";
}

function actionInboundQCInspection(
  id: string,
  action: "start" | DecisionAction,
  input: InboundQCDecisionInput
): Promise<InboundQCActionResult> {
  return apiPost<InboundQCActionResultApi, InboundQCDecisionApiRequest>(
    `/inbound-qc-inspections/${encodeURIComponent(id)}/${action}`,
    action === "start" ? {} : toApiDecisionInput(input),
    { accessToken: defaultAccessToken }
  )
    .then(fromApiActionResult)
    .catch((cause) => {
      if (!shouldUsePrototypeFallback(cause)) {
        throw cause;
      }

      if (action === "start") {
        return startPrototypeInspection(id);
      }

      return decidePrototypeInspection(id, action, input);
    });
}

function fromApiInspection(inspection: InboundQCInspectionApi): InboundQCInspection {
  return {
    id: inspection.id,
    orgId: inspection.org_id,
    goodsReceiptId: inspection.goods_receipt_id,
    goodsReceiptNo: inspection.goods_receipt_no,
    goodsReceiptLineId: inspection.goods_receipt_line_id,
    purchaseOrderId: inspection.purchase_order_id,
    purchaseOrderLineId: inspection.purchase_order_line_id,
    itemId: inspection.item_id,
    sku: inspection.sku,
    itemName: inspection.item_name,
    batchId: inspection.batch_id,
    batchNo: inspection.batch_no,
    lotNo: inspection.lot_no,
    expiryDate: inspection.expiry_date,
    warehouseId: inspection.warehouse_id,
    locationId: inspection.location_id,
    quantity: inspection.quantity,
    uomCode: inspection.uom_code,
    inspectorId: inspection.inspector_id,
    status: inspection.status,
    result: inspection.result || undefined,
    passedQuantity: inspection.passed_qty,
    failedQuantity: inspection.failed_qty,
    holdQuantity: inspection.hold_qty,
    checklist: inspection.checklist.map(fromApiChecklistItem),
    reason: inspection.reason,
    note: inspection.note,
    auditLogId: inspection.audit_log_id,
    createdAt: inspection.created_at,
    createdBy: inspection.created_by,
    updatedAt: inspection.updated_at,
    updatedBy: inspection.updated_by,
    startedAt: inspection.started_at,
    startedBy: inspection.started_by,
    decidedAt: inspection.decided_at,
    decidedBy: inspection.decided_by
  };
}

function fromApiChecklistItem(item: InboundQCChecklistItemApi): InboundQCChecklistItem {
  return {
    id: item.id,
    code: item.code,
    label: item.label,
    required: item.required,
    status: item.status,
    note: item.note
  };
}

function fromApiActionResult(result: InboundQCActionResultApi): InboundQCActionResult {
  return {
    inspection: fromApiInspection(result.inspection),
    previousStatus: result.previous_status,
    currentStatus: result.current_status,
    previousResult: result.previous_result,
    currentResult: result.current_result,
    auditLogId: result.audit_log_id
  };
}

function toApiCreateInput(input: CreateInboundQCInspectionInput): CreateInboundQCInspectionApiRequest {
  return {
    id: input.id,
    org_id: input.orgId,
    goods_receipt_id: input.goodsReceiptId,
    goods_receipt_line_id: input.goodsReceiptLineId,
    inspector_id: input.inspectorId,
    checklist: input.checklist?.map(toApiChecklistItem),
    note: input.note
  };
}

function toApiDecisionInput(input: InboundQCDecisionInput): InboundQCDecisionApiRequest {
  return {
    passed_qty: input.passedQuantity,
    failed_qty: input.failedQuantity,
    hold_qty: input.holdQuantity,
    checklist: input.checklist?.map(toApiChecklistItem),
    reason: input.reason,
    note: input.note
  };
}

function toApiChecklistItem(item: InboundQCChecklistItem): InboundQCChecklistItemApi {
  return {
    id: item.id,
    code: item.code,
    label: item.label,
    required: item.required,
    status: item.status,
    note: item.note
  };
}

function inboundQCQueryString(query: InboundQCInspectionQuery) {
  const params = new URLSearchParams();
  Object.entries(toApiInboundQCQuery(query) ?? {}).forEach(([key, value]) => {
    if (value) {
      params.set(key, value);
    }
  });

  const value = params.toString();
  return value ? `?${value}` : "";
}

function toApiInboundQCQuery(query: InboundQCInspectionQuery): InboundQCInspectionListApiQuery {
  return {
    status: query.status,
    goods_receipt_id: query.goodsReceiptId,
    goods_receipt_line_id: query.goodsReceiptLineId,
    warehouse_id: query.warehouseId
  };
}


function createPrototypeInspections(): InboundQCInspection[] {
  return [];
}

function filterPrototypeInspections(query: InboundQCInspectionQuery) {
  return prototypeInspections.filter((inspection) => matchesInspectionQuery(inspection, query)).map(cloneInspection).sort(sortInspections);
}

function createPrototypeInspection(input: CreateInboundQCInspectionInput): InboundQCActionResult {
  const line = prototypeInspectableLines.find(
    (candidate) =>
      candidate.goodsReceiptId === input.goodsReceiptId && candidate.goodsReceiptLineId === input.goodsReceiptLineId
  );
  if (!line) {
    throw new Error("Goods receipt line is not inspection-ready");
  }
  if (
    prototypeInspections.some(
      (inspection) =>
        inspection.goodsReceiptId === input.goodsReceiptId &&
        inspection.goodsReceiptLineId === input.goodsReceiptLineId &&
        inspection.status !== "cancelled"
    )
  ) {
    throw new Error("Inbound QC inspection already exists for this goods receipt line");
  }

  const inspection: InboundQCInspection = {
    id: input.id?.trim() || `iqc-${line.goodsReceiptId}-${line.goodsReceiptLineId}`,
    orgId: input.orgId?.trim() || "org-my-pham",
    goodsReceiptId: line.goodsReceiptId,
    goodsReceiptNo: line.goodsReceiptNo,
    goodsReceiptLineId: line.goodsReceiptLineId,
    purchaseOrderId: line.purchaseOrderId,
    purchaseOrderLineId: line.purchaseOrderLineId,
    itemId: line.itemId,
    sku: line.sku,
    itemName: line.itemName,
    batchId: line.batchId,
    batchNo: line.batchNo,
    lotNo: line.lotNo,
    expiryDate: line.expiryDate,
    warehouseId: line.warehouseId,
    locationId: line.locationId,
    quantity: normalizeQuantity(line.quantity),
    uomCode: line.uomCode,
    inspectorId: input.inspectorId?.trim() || "user-qa",
    status: "pending",
    passedQuantity: zeroQuantity,
    failedQuantity: zeroQuantity,
    holdQuantity: zeroQuantity,
    checklist: cloneChecklist(input.checklist?.length ? input.checklist : defaultInboundQCChecklist),
    note: input.note?.trim(),
    auditLogId: "audit-iqc-prototype-created",
    createdAt: prototypeNow,
    createdBy: input.inspectorId?.trim() || "user-qa",
    updatedAt: prototypeNow,
    updatedBy: input.inspectorId?.trim() || "user-qa"
  };

  prototypeInspections = [inspection, ...prototypeInspections];

  return {
    inspection: cloneInspection(inspection),
    currentStatus: inspection.status,
    auditLogId: inspection.auditLogId
  };
}

function startPrototypeInspection(id: string): InboundQCActionResult {
  const current = findPrototypeInspection(id);
  if (current.status !== "pending") {
    throw new Error("Inbound QC inspection status transition is not allowed");
  }

  const updated: InboundQCInspection = {
    ...current,
    status: "in_progress",
    startedAt: "2026-04-27T11:10:00Z",
    startedBy: current.inspectorId,
    updatedAt: "2026-04-27T11:10:00Z",
    updatedBy: current.inspectorId,
    auditLogId: "audit-iqc-prototype-started"
  };

  savePrototypeInspection(updated);

  return {
    inspection: cloneInspection(updated),
    previousStatus: current.status,
    currentStatus: updated.status,
    auditLogId: updated.auditLogId
  };
}

function decidePrototypeInspection(
  id: string,
  result: DecisionAction,
  input: InboundQCDecisionInput
): InboundQCActionResult {
  const current = findPrototypeInspection(id);
  if (current.status !== "in_progress") {
    throw new Error("Inbound QC inspection status transition is not allowed");
  }

  const quantities = normalizeDecisionQuantities(current.quantity, result, input);
  const checklist = cloneChecklist(input.checklist ?? current.checklist);
  if (checklist.some((item) => item.required && item.status === "pending")) {
    throw new Error("Required checklist items must be completed");
  }
  if ((result === "fail" || result === "hold" || result === "partial") && !input.reason?.trim()) {
    throw new Error("Reason is required for this QC decision");
  }

  const updated: InboundQCInspection = {
    ...current,
    status: "completed",
    result,
    passedQuantity: quantities.passedQuantity,
    failedQuantity: quantities.failedQuantity,
    holdQuantity: quantities.holdQuantity,
    checklist,
    reason: input.reason?.trim(),
    note: input.note?.trim(),
    decidedAt: "2026-04-27T11:20:00Z",
    decidedBy: current.inspectorId,
    updatedAt: "2026-04-27T11:20:00Z",
    updatedBy: current.inspectorId,
    auditLogId: `audit-iqc-prototype-${result}`
  };

  savePrototypeInspection(updated);

  return {
    inspection: cloneInspection(updated),
    previousStatus: current.status,
    currentStatus: updated.status,
    previousResult: current.result,
    currentResult: updated.result,
    auditLogId: updated.auditLogId
  };
}

function normalizeDecisionQuantities(
  totalQuantity: string,
  result: DecisionAction,
  input: InboundQCDecisionInput
) {
  const total = normalizeQuantity(totalQuantity);
  if (result === "pass") {
    return { passedQuantity: normalizeQuantity(input.passedQuantity || total), failedQuantity: zeroQuantity, holdQuantity: zeroQuantity };
  }
  if (result === "fail") {
    return { passedQuantity: zeroQuantity, failedQuantity: normalizeQuantity(input.failedQuantity || total), holdQuantity: zeroQuantity };
  }
  if (result === "hold") {
    return { passedQuantity: zeroQuantity, failedQuantity: zeroQuantity, holdQuantity: normalizeQuantity(input.holdQuantity || total) };
  }

  const passedQuantity = normalizeQuantity(input.passedQuantity || zeroQuantity);
  const failedQuantity = normalizeQuantity(input.failedQuantity || zeroQuantity);
  const holdQuantity = normalizeQuantity(input.holdQuantity || zeroQuantity);
  const bucketTotal = quantityUnits(passedQuantity) + quantityUnits(failedQuantity) + quantityUnits(holdQuantity);
  const nonZeroBuckets = [passedQuantity, failedQuantity, holdQuantity].filter(
    (value) => quantityUnits(value) > BigInt(0)
  ).length;
  if (bucketTotal !== quantityUnits(total) || nonZeroBuckets < 2) {
    throw new Error("Partial QC quantities must sum to the inspection quantity");
  }

  return { passedQuantity, failedQuantity, holdQuantity };
}

function normalizeQuantity(value: string) {
  return normalizeDecimalInput(value, decimalScales.quantity);
}

function quantityUnits(value: string) {
  return BigInt(normalizeQuantity(value).replace(".", ""));
}

function findPrototypeInspection(id: string) {
  const inspection = prototypeInspections.find((candidate) => candidate.id === id);
  if (!inspection) {
    throw new Error("Inbound QC inspection not found");
  }

  return cloneInspection(inspection);
}

function savePrototypeInspection(next: InboundQCInspection) {
  prototypeInspections = [next, ...prototypeInspections.filter((candidate) => candidate.id !== next.id)];
}

function matchesInspectionQuery(inspection: InboundQCInspection, query: InboundQCInspectionQuery) {
  if (query.status && inspection.status !== query.status) {
    return false;
  }
  if (query.goodsReceiptId && inspection.goodsReceiptId !== query.goodsReceiptId) {
    return false;
  }
  if (query.goodsReceiptLineId && inspection.goodsReceiptLineId !== query.goodsReceiptLineId) {
    return false;
  }
  if (query.warehouseId && inspection.warehouseId !== query.warehouseId) {
    return false;
  }

  return true;
}

function cloneInspection(inspection: InboundQCInspection): InboundQCInspection {
  return {
    ...inspection,
    checklist: cloneChecklist(inspection.checklist)
  };
}

function cloneChecklist(items: InboundQCChecklistItem[]): InboundQCChecklistItem[] {
  return items.map((item) => ({ ...item }));
}

function sortInspections(left: InboundQCInspection, right: InboundQCInspection) {
  return right.updatedAt.localeCompare(left.updatedAt);
}
