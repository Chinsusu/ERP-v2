import { ApiError, apiGetRaw, apiPatch, apiPost } from "../../../shared/api/client";
import { decimalScales, normalizeDecimalInput } from "../../../shared/format/numberFormat";
import type { AuditLogItem } from "@/modules/audit/types";
import type {
  ChangeSubcontractOrderStatusInput,
  CreateSubcontractOrderInput,
  SubcontractDepositStatus,
  SubcontractFactory,
  SubcontractFinalPaymentStatus,
  SubcontractOrder,
  SubcontractOrderMaterialLine,
  SubcontractOrderQuery,
  SubcontractOrderStatus,
  SubcontractOrderSummary,
  SubcontractProduct,
  SubcontractStatusChangeResult,
  UpdateSubcontractOrderInput
} from "../types";

type SubcontractOrderApiMaterialLine = {
  id: string;
  line_no: number;
  item_id: string;
  sku_code: string;
  item_name: string;
  planned_qty: string;
  issued_qty: string;
  uom_code: string;
  base_planned_qty?: string;
  base_issued_qty?: string;
  base_uom_code?: string;
  conversion_factor?: string;
  unit_cost: string;
  currency_code: string;
  line_cost_amount: string;
  lot_trace_required: boolean;
  note?: string;
};

type SubcontractOrderApi = {
  id: string;
  order_no: string;
  factory_id: string;
  factory_code?: string;
  factory_name: string;
  finished_item_id: string;
  finished_sku_code: string;
  finished_item_name: string;
  planned_qty: string;
  received_qty?: string;
  accepted_qty?: string;
  rejected_qty?: string;
  uom_code: string;
  currency_code: string;
  estimated_cost_amount: string;
  deposit_amount?: string;
  spec_summary?: string;
  sample_required: boolean;
  claim_window_days?: number;
  target_start_date?: string;
  expected_receipt_date: string;
  status: SubcontractOrderStatus;
  material_lines: SubcontractOrderApiMaterialLine[];
  audit_log_id?: string;
  created_at: string;
  updated_at: string;
  submitted_at?: string;
  approved_at?: string;
  factory_confirmed_at?: string;
  cancelled_at?: string;
  closed_at?: string;
  cancel_reason?: string;
  version: number;
};

type SubcontractOrderApiListItem = Omit<SubcontractOrderApi, "material_lines"> & {
  material_line_count: number;
};

type SubcontractOrderApiMaterialLineRequest = {
  item_id: string;
  planned_qty: string;
  uom_code: string;
  unit_cost: string;
  currency_code: string;
  lot_trace_required: boolean;
  note?: string;
};

type CreateSubcontractOrderApiRequest = {
  id?: string;
  order_no?: string;
  factory_id: string;
  finished_item_id: string;
  planned_qty: string;
  uom_code: string;
  currency_code: string;
  spec_summary: string;
  sample_required: boolean;
  claim_window_days: number;
  target_start_date?: string;
  expected_receipt_date: string;
  material_lines: SubcontractOrderApiMaterialLineRequest[];
};

type UpdateSubcontractOrderApiRequest = Partial<CreateSubcontractOrderApiRequest> & {
  expected_version: number;
};

type SubcontractOrderActionApiRequest = {
  expected_version?: number;
  reason?: string;
};

type SubcontractOrderActionApiResult = {
  subcontract_order: SubcontractOrderApi;
  previous_status: SubcontractOrderStatus;
  current_status: SubcontractOrderStatus;
  audit_log_id?: string;
};

type SubcontractOrderAction = "submit" | "approve" | "confirm-factory" | "cancel" | "close";

const defaultAccessToken = "local-dev-access-token";
const prototypeNow = "2026-04-29T10:00:00Z";
let subcontractOrderSequence = 2;
let subcontractAuditSequence = 1;

export const subcontractFactoryOptions: SubcontractFactory[] = [
  { id: "sup-out-lotus", code: "SUP-OUT-LOTUS", name: "Lotus Filling Partner" },
  { id: "sup-rm-bioactive", code: "SUP-RM-BIO", name: "BioActive Raw Materials" },
  { id: "sup-pkg-vina", code: "SUP-PKG-VINA", name: "Vina Packaging Solutions" }
];

export const subcontractProductOptions: SubcontractProduct[] = [
  { id: "item-serum-30ml", sku: "SERUM-30ML", name: "Hydrating Serum 30ml" },
  { id: "item-cream-50g", sku: "CREAM-50G", name: "Repair Cream 50g" },
  { id: "item-toner-100ml", sku: "TONER-100ML", name: "Balancing Toner 100ml" }
];

export const subcontractMaterialItemOptions = [
  { id: "item-cream-50g", sku: "CREAM-50G", name: "Repair Cream 50g", defaultUnitCost: "58000.000000" },
  { id: "item-serum-30ml", sku: "SERUM-30ML", name: "Hydrating Serum 30ml", defaultUnitCost: "64000.000000" },
  { id: "item-toner-100ml", sku: "TONER-100ML", name: "Balancing Toner 100ml", defaultUnitCost: "42000.000000" }
];

export const subcontractDepositStatusOptions: { label: string; value: SubcontractDepositStatus }[] = [
  { label: "Not required", value: "not_required" },
  { label: "Pending", value: "pending" },
  { label: "Paid", value: "paid" }
];

export const subcontractOrderStatusOptions: { label: string; value: SubcontractOrderStatus }[] = [
  { label: "Draft", value: "draft" },
  { label: "Submitted", value: "submitted" },
  { label: "Approved", value: "approved" },
  { label: "Factory confirmed", value: "factory_confirmed" },
  { label: "Deposit recorded", value: "deposit_recorded" },
  { label: "Materials issued", value: "materials_issued_to_factory" },
  { label: "Sample submitted", value: "sample_submitted" },
  { label: "Sample approved", value: "sample_approved" },
  { label: "Sample rejected", value: "sample_rejected" },
  { label: "Mass production", value: "mass_production_started" },
  { label: "Finished received", value: "finished_goods_received" },
  { label: "QC in progress", value: "qc_in_progress" },
  { label: "Accepted", value: "accepted" },
  { label: "Factory issue", value: "rejected_with_factory_issue" },
  { label: "Final payment ready", value: "final_payment_ready" },
  { label: "Closed", value: "closed" },
  { label: "Cancelled", value: "cancelled" }
];

export const subcontractOrderActionOptions: { label: string; value: SubcontractOrderAction }[] = [
  { label: "Submit", value: "submit" },
  { label: "Approve", value: "approve" },
  { label: "Confirm factory", value: "confirm-factory" },
  { label: "Cancel", value: "cancel" },
  { label: "Close", value: "close" }
];

export const prototypeSubcontractOrders: SubcontractOrder[] = [
  createSubcontractOrderRecord({
    id: "sco-260429-0001",
    orderNo: "SCO-260429-0001",
    factoryId: "sup-out-lotus",
    factoryCode: "SUP-OUT-LOTUS",
    factoryName: "Lotus Filling Partner",
    productId: "item-serum-30ml",
    sku: "SERUM-30ML",
    productName: "Hydrating Serum 30ml",
    quantity: 5000,
    specVersion: "SPEC-SERUM-2026.04",
    sampleRequired: true,
    expectedDeliveryDate: "2026-05-20",
    depositStatus: "pending",
    depositAmount: 12000000,
    finalPaymentStatus: "hold",
    status: "approved",
    createdBy: "Subcontract Coordinator",
    createdAt: "2026-04-29T08:00:00Z",
    updatedAt: "2026-04-29T08:15:00Z",
    version: 3,
    estimatedCostAmount: "1160000.00",
    materialLines: [
      {
        id: "sco-260429-0001-material-01",
        itemId: "item-cream-50g",
        skuCode: "CREAM-50G",
        itemName: "Repair Cream 50g",
        plannedQty: "20.000000",
        issuedQty: "0.000000",
        uomCode: "EA",
        unitCost: "58000.000000",
        currencyCode: "VND",
        lineCostAmount: "1160000.00",
        lotTraceRequired: true
      }
    ],
    auditLogIds: ["audit-sco-260429-0001-approved"]
  })
];

let prototypeStore = prototypeSubcontractOrders.map(cloneSubcontractOrder);

export async function getSubcontractOrders(query: SubcontractOrderQuery = {}): Promise<SubcontractOrder[]> {
  try {
    const orders = await apiGetRaw<SubcontractOrderApiListItem[]>(`/subcontract-orders${subcontractOrderQueryString(query)}`, {
      accessToken: defaultAccessToken
    });

    return orders.map(fromApiSubcontractOrderListItem);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return filterPrototypeSubcontractOrders(query);
  }
}

export async function getSubcontractOrder(id: string): Promise<SubcontractOrder> {
  try {
    const order = await apiGetRaw<SubcontractOrderApi>(`/subcontract-orders/${encodeURIComponent(id)}`, {
      accessToken: defaultAccessToken
    });

    return fromApiSubcontractOrder(order);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return getPrototypeSubcontractOrder(id);
  }
}

export async function createSubcontractOrder(input: CreateSubcontractOrderInput): Promise<SubcontractOrder> {
  try {
    const order = await apiPost<SubcontractOrderApi, CreateSubcontractOrderApiRequest>(
      "/subcontract-orders",
      toApiCreateInput(input),
      { accessToken: defaultAccessToken }
    );

    return fromApiSubcontractOrder(order);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return createPrototypeSubcontractOrder(input);
  }
}

export async function updateSubcontractOrder(
  id: string,
  input: UpdateSubcontractOrderInput
): Promise<SubcontractOrder> {
  try {
    const order = await apiPatch<SubcontractOrderApi, UpdateSubcontractOrderApiRequest>(
      `/subcontract-orders/${encodeURIComponent(id)}`,
      toApiUpdateInput(input),
      { accessToken: defaultAccessToken }
    );

    return fromApiSubcontractOrder(order);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return updatePrototypeSubcontractOrder(id, input);
  }
}

export async function submitSubcontractOrder(id: string, expectedVersion?: number): Promise<SubcontractStatusChangeResult> {
  return runSubcontractOrderAction(id, "submit", expectedVersion);
}

export async function approveSubcontractOrder(id: string, expectedVersion?: number): Promise<SubcontractStatusChangeResult> {
  return runSubcontractOrderAction(id, "approve", expectedVersion);
}

export async function confirmFactorySubcontractOrder(
  id: string,
  expectedVersion?: number
): Promise<SubcontractStatusChangeResult> {
  return runSubcontractOrderAction(id, "confirm-factory", expectedVersion);
}

export async function cancelSubcontractOrder(
  id: string,
  reason: string,
  expectedVersion?: number
): Promise<SubcontractStatusChangeResult> {
  return runSubcontractOrderAction(id, "cancel", expectedVersion, reason);
}

export async function closeSubcontractOrder(id: string, expectedVersion?: number): Promise<SubcontractStatusChangeResult> {
  return runSubcontractOrderAction(id, "close", expectedVersion);
}

export function changeSubcontractOrderStatus(input: ChangeSubcontractOrderStatusInput): SubcontractStatusChangeResult {
  const nextStatus = normalizeOrderStatus(input.nextStatus);
  const beforeStatus = input.order.status;
  const auditLog = createStatusAuditLog(input.order, beforeStatus, nextStatus, input);
  const order = createSubcontractOrderRecord({
    ...input.order,
    status: nextStatus,
    updatedAt: prototypeNow,
    version: input.order.version + 1,
    auditLogIds: [...input.order.auditLogIds, auditLog.id]
  });

  return { order, auditLog, auditLogId: auditLog.id };
}

export function summarizeSubcontractOrders(orders: SubcontractOrder[]): SubcontractOrderSummary {
  const sortedDeliveryDates = orders
    .map((order) => order.expectedDeliveryDate)
    .filter((date) => date.trim() !== "")
    .sort();

  return {
    total: orders.length,
    draft: orders.filter((order) => order.status === "draft").length,
    confirmed: orders.filter((order) => ["approved", "factory_confirmed"].includes(order.status)).length,
    active: orders.filter((order) => isActiveOrderStatus(order.status)).length,
    accepted: orders.filter((order) => order.status === "accepted").length,
    rejected: orders.filter((order) => order.status === "rejected_with_factory_issue").length,
    closed: orders.filter((order) => order.status === "closed").length,
    nextDeliveryDate: sortedDeliveryDates[0]
  };
}

export function subcontractOrderStatusTone(
  status: SubcontractOrderStatus
): "normal" | "success" | "warning" | "danger" | "info" {
  switch (status) {
    case "accepted":
    case "final_payment_ready":
    case "closed":
      return "success";
    case "cancelled":
    case "rejected_with_factory_issue":
    case "sample_rejected":
      return "danger";
    case "draft":
    case "submitted":
    case "qc_in_progress":
      return "warning";
    case "approved":
    case "factory_confirmed":
    case "deposit_recorded":
    case "materials_issued_to_factory":
    case "sample_submitted":
    case "sample_approved":
    case "mass_production_started":
    case "finished_goods_received":
    default:
      return "info";
  }
}

export function subcontractDepositStatusTone(
  status: SubcontractDepositStatus
): "normal" | "success" | "warning" {
  switch (status) {
    case "paid":
      return "success";
    case "pending":
      return "warning";
    case "not_required":
    default:
      return "normal";
  }
}

export function formatSubcontractOrderStatus(status: SubcontractOrderStatus) {
  return subcontractOrderStatusOptions.find((option) => option.value === status)?.label ?? status;
}

export function formatSubcontractDepositStatus(status: SubcontractDepositStatus) {
  return subcontractDepositStatusOptions.find((option) => option.value === status)?.label ?? status;
}

export function availableSubcontractOrderActions(status: SubcontractOrderStatus): SubcontractOrderAction[] {
  switch (status) {
    case "draft":
      return ["submit", "cancel"];
    case "submitted":
      return ["approve", "cancel"];
    case "approved":
      return ["confirm-factory", "cancel"];
    case "accepted":
    case "final_payment_ready":
      return ["close"];
    default:
      return [];
  }
}

export function resetPrototypeSubcontractOrdersForTest() {
  subcontractOrderSequence = 2;
  subcontractAuditSequence = 1;
  prototypeStore = prototypeSubcontractOrders.map(cloneSubcontractOrder);
}

async function runSubcontractOrderAction(
  id: string,
  action: SubcontractOrderAction,
  expectedVersion?: number,
  reason?: string
): Promise<SubcontractStatusChangeResult> {
  try {
    const result = await apiPost<SubcontractOrderActionApiResult, SubcontractOrderActionApiRequest>(
      `/subcontract-orders/${encodeURIComponent(id)}/${action}`,
      { expected_version: expectedVersion, reason },
      { accessToken: defaultAccessToken }
    );

    return fromApiActionResult(result);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return transitionPrototypeSubcontractOrder(id, action, expectedVersion, reason);
  }
}

function shouldUsePrototypeFallback(reason: unknown) {
  if (reason instanceof ApiError) {
    return false;
  }

  return !(reason instanceof Error && reason.message.startsWith("API request failed:"));
}

function fromApiSubcontractOrderListItem(order: SubcontractOrderApiListItem): SubcontractOrder {
  return {
    id: order.id,
    orderNo: order.order_no,
    factoryId: order.factory_id,
    factoryCode: order.factory_code ?? "",
    factoryName: order.factory_name,
    productId: order.finished_item_id,
    sku: order.finished_sku_code,
    productName: order.finished_item_name,
    quantity: Number(order.planned_qty),
    specVersion: order.spec_summary ?? "",
    sampleRequired: order.sample_required,
    expectedDeliveryDate: order.expected_receipt_date,
    depositStatus: depositStatusForAmount(order.deposit_amount),
    depositAmount: numberFromDecimal(order.deposit_amount),
    finalPaymentStatus: finalPaymentStatusForOrder(order.status),
    status: order.status,
    createdBy: "API",
    createdAt: order.created_at,
    updatedAt: order.updated_at,
    version: order.version ?? 1,
    estimatedCostAmount: order.estimated_cost_amount,
    materialLines: [],
    auditLogIds: order.audit_log_id ? [order.audit_log_id] : []
  };
}

function fromApiSubcontractOrder(order: SubcontractOrderApi): SubcontractOrder {
  return {
    ...fromApiSubcontractOrderListItem({ ...order, material_line_count: order.material_lines.length }),
    materialLines: order.material_lines.map(fromApiSubcontractMaterialLine),
    auditLogIds: order.audit_log_id ? [order.audit_log_id] : []
  };
}

function fromApiSubcontractMaterialLine(line: SubcontractOrderApiMaterialLine): SubcontractOrderMaterialLine {
  return {
    id: line.id,
    itemId: line.item_id,
    skuCode: line.sku_code,
    itemName: line.item_name,
    plannedQty: line.planned_qty,
    issuedQty: line.issued_qty,
    uomCode: line.uom_code,
    unitCost: line.unit_cost,
    currencyCode: line.currency_code,
    lineCostAmount: line.line_cost_amount,
    lotTraceRequired: line.lot_trace_required,
    note: line.note
  };
}

function fromApiActionResult(result: SubcontractOrderActionApiResult): SubcontractStatusChangeResult {
  const order = fromApiSubcontractOrder(result.subcontract_order);
  const auditLog = createStatusAuditLog(order, result.previous_status, result.current_status, {
    order,
    nextStatus: result.current_status,
    note: result.audit_log_id
  });

  return {
    order,
    auditLog: {
      ...auditLog,
      id: result.audit_log_id ?? auditLog.id
    },
    auditLogId: result.audit_log_id
  };
}

function toApiCreateInput(input: CreateSubcontractOrderInput): CreateSubcontractOrderApiRequest {
  return {
    id: input.id,
    order_no: input.orderNo,
    factory_id: input.factoryId,
    finished_item_id: input.productId,
    planned_qty: normalizeDecimalInput(String(input.quantity), decimalScales.quantity),
    uom_code: "EA",
    currency_code: "VND",
    spec_summary: input.specVersion,
    sample_required: input.sampleRequired,
    claim_window_days: 7,
    expected_receipt_date: input.expectedDeliveryDate,
    material_lines: [toApiMaterialLineInput(input)]
  };
}

function toApiUpdateInput(input: UpdateSubcontractOrderInput): UpdateSubcontractOrderApiRequest {
  return {
    ...toApiCreateInput({
      factoryId: input.factoryId ?? subcontractFactoryOptions[0].id,
      productId: input.productId ?? subcontractProductOptions[0].id,
      quantity: input.quantity ?? 1,
      specVersion: input.specVersion ?? "",
      sampleRequired: input.sampleRequired ?? true,
      expectedDeliveryDate: input.expectedDeliveryDate ?? "",
      depositStatus: input.depositStatus ?? "pending",
      materialItemId: input.materialItemId ?? subcontractMaterialItemOptions[0].id,
      materialQty: input.materialQty ?? "1",
      materialUnitCost: input.materialUnitCost ?? subcontractMaterialItemOptions[0].defaultUnitCost
    }),
    expected_version: input.expectedVersion
  };
}

function toApiMaterialLineInput(input: CreateSubcontractOrderInput): SubcontractOrderApiMaterialLineRequest {
  return {
    item_id: input.materialItemId,
    planned_qty: normalizeDecimalInput(input.materialQty, decimalScales.quantity),
    uom_code: "EA",
    unit_cost: normalizeDecimalInput(input.materialUnitCost, decimalScales.unitCost),
    currency_code: "VND",
    lot_trace_required: input.materialLotTraceRequired ?? true
  };
}

function subcontractOrderQueryString(query: SubcontractOrderQuery) {
  const params = new URLSearchParams();
  if (query.search) {
    params.set("search", query.search);
  }
  if (query.status) {
    params.set("status", query.status);
  }
  if (query.factoryId) {
    params.set("factory_id", query.factoryId);
  }
  if (query.productId) {
    params.set("finished_item_id", query.productId);
  }
  if (query.expectedReceiptFrom) {
    params.set("expected_receipt_from", query.expectedReceiptFrom);
  }
  if (query.expectedReceiptTo) {
    params.set("expected_receipt_to", query.expectedReceiptTo);
  }

  const value = params.toString();
  return value ? `?${value}` : "";
}

function getPrototypeSubcontractOrder(id: string) {
  const order = prototypeStore.find((candidate) => candidate.id === id);
  if (!order) {
    throw new Error("Subcontract order not found");
  }

  return cloneSubcontractOrder(order);
}

function filterPrototypeSubcontractOrders(query: SubcontractOrderQuery) {
  return prototypeStore.filter((order) => matchesSubcontractOrderQuery(order, query)).map(cloneSubcontractOrder);
}

function createPrototypeSubcontractOrder(input: CreateSubcontractOrderInput): SubcontractOrder {
  subcontractOrderSequence += 1;
  const factory = resolveFactory(input.factoryId, input.factoryName);
  const product = resolveProduct(input.productId, input.productName);
  const material = resolveMaterialItem(input.materialItemId);
  const quantity = normalizeQuantity(input.quantity);
  const materialQty = normalizeDecimalInput(input.materialQty, decimalScales.quantity);
  const materialUnitCost = normalizeDecimalInput(input.materialUnitCost, decimalScales.unitCost);
  const id = input.id ?? `sco-ui-${subcontractOrderSequence}`;
  const lineAmount = calculateLineCost(materialQty, materialUnitCost);
  const order = createSubcontractOrderRecord({
    id,
    orderNo: input.orderNo ?? `SCO-260429-${String(subcontractOrderSequence).padStart(4, "0")}`,
    factoryId: factory.id,
    factoryCode: factory.code,
    factoryName: factory.name,
    productId: product.id,
    sku: product.sku,
    productName: product.name,
    quantity,
    specVersion: normalizeRequiredText(input.specVersion, "Spec summary is required"),
    sampleRequired: input.sampleRequired,
    expectedDeliveryDate: normalizeRequiredText(input.expectedDeliveryDate, "Expected receipt date is required"),
    depositStatus: input.depositStatus,
    depositAmount: input.depositAmount,
    finalPaymentStatus: finalPaymentStatusForDeposit(input.depositStatus),
    status: "draft",
    createdBy: input.createdBy?.trim() || "Subcontract Coordinator",
    createdAt: prototypeNow,
    updatedAt: prototypeNow,
    version: 1,
    estimatedCostAmount: lineAmount,
    materialLines: [
      {
        id: `${id}-material-01`,
        itemId: material.id,
        skuCode: material.sku,
        itemName: material.name,
        plannedQty: materialQty,
        issuedQty: "0.000000",
        uomCode: "EA",
        unitCost: materialUnitCost,
        currencyCode: "VND",
        lineCostAmount: lineAmount,
        lotTraceRequired: input.materialLotTraceRequired ?? true
      }
    ],
    auditLogIds: [`audit-${id}-created`]
  });
  prototypeStore = [order, ...prototypeStore.filter((candidate) => candidate.id !== id)];

  return cloneSubcontractOrder(order);
}

function updatePrototypeSubcontractOrder(id: string, input: UpdateSubcontractOrderInput): SubcontractOrder {
  const current = getPrototypeSubcontractOrder(id);
  if (current.status !== "draft") {
    throw new Error("Only draft subcontract orders can be updated");
  }
  if (input.expectedVersion && input.expectedVersion !== current.version) {
    throw new Error("Subcontract order version changed");
  }
  const updated = createPrototypeSubcontractOrder({
    id: current.id,
    orderNo: current.orderNo,
    factoryId: input.factoryId ?? current.factoryId,
    productId: input.productId ?? current.productId,
    quantity: input.quantity ?? current.quantity,
    specVersion: input.specVersion ?? current.specVersion,
    sampleRequired: input.sampleRequired ?? current.sampleRequired,
    expectedDeliveryDate: input.expectedDeliveryDate ?? current.expectedDeliveryDate,
    depositStatus: input.depositStatus ?? current.depositStatus,
    depositAmount: input.depositAmount ?? current.depositAmount,
    materialItemId: input.materialItemId ?? current.materialLines[0]?.itemId ?? subcontractMaterialItemOptions[0].id,
    materialQty: input.materialQty ?? current.materialLines[0]?.plannedQty ?? "1",
    materialUnitCost: input.materialUnitCost ?? current.materialLines[0]?.unitCost ?? subcontractMaterialItemOptions[0].defaultUnitCost,
    materialLotTraceRequired: input.materialLotTraceRequired ?? current.materialLines[0]?.lotTraceRequired ?? true
  });
  updated.version = current.version + 1;
  updated.createdAt = current.createdAt;
  updated.updatedAt = prototypeNow;
  updated.auditLogIds = [...current.auditLogIds, `audit-${id}-updated`];
  prototypeStore = [updated, ...prototypeStore.filter((candidate) => candidate.id !== id)];

  return cloneSubcontractOrder(updated);
}

function transitionPrototypeSubcontractOrder(
  id: string,
  action: SubcontractOrderAction,
  expectedVersion?: number,
  reason?: string
): SubcontractStatusChangeResult {
  const current = getPrototypeSubcontractOrder(id);
  if (expectedVersion && expectedVersion !== current.version) {
    throw new Error("Subcontract order version changed");
  }
  const nextStatus = nextPrototypeStatus(current.status, action);
  const result = changeSubcontractOrderStatus({
    order: current,
    nextStatus,
    actorName: "Subcontract Coordinator",
    note: reason
  });
  prototypeStore = [result.order, ...prototypeStore.filter((candidate) => candidate.id !== id)];

  return result;
}

function nextPrototypeStatus(currentStatus: SubcontractOrderStatus, action: SubcontractOrderAction) {
  if (action === "submit" && currentStatus === "draft") {
    return "submitted";
  }
  if (action === "approve" && currentStatus === "submitted") {
    return "approved";
  }
  if (action === "confirm-factory" && currentStatus === "approved") {
    return "factory_confirmed";
  }
  if (action === "close" && ["accepted", "final_payment_ready"].includes(currentStatus)) {
    return "closed";
  }
  if (action === "cancel" && ["draft", "submitted", "approved", "factory_confirmed"].includes(currentStatus)) {
    return "cancelled";
  }

  throw new Error(`Cannot ${action} subcontract order from ${formatSubcontractOrderStatus(currentStatus)}`);
}

function createSubcontractOrderRecord(input: SubcontractOrder): SubcontractOrder {
  return {
    ...input,
    auditLogIds: [...input.auditLogIds],
    materialLines: input.materialLines.map((line) => ({ ...line }))
  };
}

function createStatusAuditLog(
  order: SubcontractOrder,
  beforeStatus: SubcontractOrderStatus,
  afterStatus: SubcontractOrderStatus,
  input: ChangeSubcontractOrderStatusInput
): AuditLogItem {
  const id = `audit-subcontract-status-${String(subcontractAuditSequence++).padStart(4, "0")}`;

  return {
    id,
    actorId: input.actorId ?? "user-subcontract-coordinator",
    action: "subcontract.order.status_changed",
    entityType: "subcontract_order",
    entityId: order.id,
    requestId: `req_${id}`,
    beforeData: {
      status: beforeStatus
    },
    afterData: {
      status: afterStatus
    },
    metadata: {
      order_no: order.orderNo,
      factory: order.factoryName,
      product: order.productName,
      note: input.note?.trim() || "Status changed from subcontract order UI",
      actor_name: input.actorName ?? "Subcontract Coordinator"
    },
    createdAt: prototypeNow
  };
}

function resolveFactory(factoryId: string, factoryName?: string): SubcontractFactory {
  const normalizedFactoryId = normalizeRequiredText(factoryId, "Factory is required");
  const matchedFactory = subcontractFactoryOptions.find((factory) => factory.id === normalizedFactoryId);

  return (
    matchedFactory ?? {
      id: normalizedFactoryId,
      code: normalizedFactoryId.toUpperCase(),
      name: factoryName?.trim() || normalizedFactoryId
    }
  );
}

function resolveProduct(productId: string, productName?: string): SubcontractProduct {
  const normalizedProductId = normalizeRequiredText(productId, "Finished item is required");
  const matchedProduct = subcontractProductOptions.find((product) => product.id === normalizedProductId);

  return (
    matchedProduct ?? {
      id: normalizedProductId,
      sku: normalizedProductId.toUpperCase(),
      name: productName?.trim() || normalizedProductId
    }
  );
}

function resolveMaterialItem(itemId: string) {
  const normalizedItemId = normalizeRequiredText(itemId, "Material item is required");
  const matched = subcontractMaterialItemOptions.find((item) => item.id === normalizedItemId);
  if (!matched) {
    throw new Error("Material item is required");
  }

  return matched;
}

function normalizeQuantity(quantity: number) {
  if (!Number.isFinite(quantity) || quantity <= 0) {
    throw new Error("Order quantity must be greater than zero");
  }

  return Math.round(quantity * 1000) / 1000;
}

function normalizeRequiredText(value: string, message: string) {
  const normalized = value.trim();
  if (normalized === "") {
    throw new Error(message);
  }

  return normalized;
}

function normalizeOrderStatus(status: SubcontractOrderStatus): SubcontractOrderStatus {
  if (subcontractOrderStatusOptions.some((option) => option.value === status)) {
    return status;
  }

  return "draft";
}

function finalPaymentStatusForDeposit(status: SubcontractDepositStatus): SubcontractFinalPaymentStatus {
  return status === "paid" ? "pending" : "hold";
}

function finalPaymentStatusForOrder(status: SubcontractOrderStatus): SubcontractFinalPaymentStatus {
  if (status === "closed") {
    return "released";
  }
  if (status === "final_payment_ready") {
    return "pending";
  }

  return "hold";
}

function depositStatusForAmount(value?: string): SubcontractDepositStatus {
  const amount = Number(value ?? "0");
  return amount > 0 ? "paid" : "pending";
}

function numberFromDecimal(value?: string) {
  const amount = Number(value ?? "0");
  return Number.isFinite(amount) && amount > 0 ? amount : undefined;
}

function isActiveOrderStatus(status: SubcontractOrderStatus) {
  return !["draft", "accepted", "rejected_with_factory_issue", "closed", "cancelled"].includes(status);
}

function matchesSubcontractOrderQuery(order: SubcontractOrder, query: SubcontractOrderQuery) {
  const search = query.search?.trim().toLowerCase();
  if (search) {
    const haystack = [order.orderNo, order.factoryCode, order.factoryName, order.sku, order.productName].join(" ").toLowerCase();
    if (!haystack.includes(search)) {
      return false;
    }
  }
  if (query.factoryId && order.factoryId !== query.factoryId) {
    return false;
  }
  if (query.productId && order.productId !== query.productId) {
    return false;
  }
  if (query.status && order.status !== query.status) {
    return false;
  }
  if (query.expectedReceiptFrom && order.expectedDeliveryDate < query.expectedReceiptFrom) {
    return false;
  }
  if (query.expectedReceiptTo && order.expectedDeliveryDate > query.expectedReceiptTo) {
    return false;
  }

  return true;
}

function cloneSubcontractOrder(order: SubcontractOrder): SubcontractOrder {
  return {
    ...order,
    auditLogIds: [...order.auditLogIds],
    materialLines: order.materialLines.map((line) => ({ ...line }))
  };
}

function calculateLineCost(quantity: string, unitCost: string) {
  const amount = Number(quantity) * Number(unitCost);
  return Number.isFinite(amount) ? amount.toFixed(2) : "0.00";
}
