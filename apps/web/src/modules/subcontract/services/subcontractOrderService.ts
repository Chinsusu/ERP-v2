import { ApiError, apiGetRaw, apiPatch, apiPost } from "../../../shared/api/client";
import { decimalScales, normalizeDecimalInput } from "../../../shared/format/numberFormat";
import type { AuditLogItem } from "@/modules/audit/types";
import type {
  ChangeSubcontractOrderStatusInput,
  CreateSubcontractOrderInput,
  DecideSubcontractSampleInput,
  IssueSubcontractMaterialsInput,
  IssueSubcontractMaterialsResult,
  ReceiveSubcontractFinishedGoodsInput,
  ReceiveSubcontractFinishedGoodsResult,
  SubcontractDepositStatus,
  SubcontractFactory,
  SubcontractFinishedGoodsReceipt,
  SubcontractFinalPaymentStatus,
  SubcontractMaterialTransfer,
  SubcontractOrder,
  SubcontractOrderMaterialLine,
  SubcontractOrderQuery,
  SubcontractSampleApproval,
  SubcontractSampleApprovalResult,
  SubcontractSampleApprovalStatus,
  SubcontractOrderStatus,
  SubcontractStockMovement,
  SubcontractOrderSummary,
  SubcontractProduct,
  SubmitSubcontractSampleInput,
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
  sample_reject_reason?: string;
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

type IssueSubcontractMaterialsApiLineRequest = {
  order_material_line_id: string;
  issue_qty: string;
  uom_code: string;
  base_issue_qty?: string;
  base_uom_code?: string;
  conversion_factor?: string;
  batch_id?: string;
  batch_no?: string;
  lot_no?: string;
  source_bin_id?: string;
  note?: string;
};

type IssueSubcontractMaterialsApiEvidenceRequest = {
  id?: string;
  evidence_type: string;
  file_name?: string;
  object_key?: string;
  external_url?: string;
  note?: string;
};

type IssueSubcontractMaterialsApiRequest = {
  expected_version: number;
  source_warehouse_id: string;
  source_warehouse_code: string;
  handover_by: string;
  handover_at?: string;
  received_by: string;
  receiver_contact?: string;
  vehicle_no?: string;
  note?: string;
  lines: IssueSubcontractMaterialsApiLineRequest[];
  evidence?: IssueSubcontractMaterialsApiEvidenceRequest[];
};

type SubcontractMaterialTransferApiLine = {
  id: string;
  line_no: number;
  order_material_line_id: string;
  item_id: string;
  sku_code: string;
  item_name: string;
  issue_qty: string;
  uom_code: string;
  base_issue_qty: string;
  base_uom_code: string;
  conversion_factor: string;
  batch_id?: string;
  batch_no?: string;
  lot_no?: string;
  source_bin_id?: string;
  lot_trace_required: boolean;
  note?: string;
};

type SubcontractMaterialTransferApiEvidence = {
  id: string;
  evidence_type: string;
  file_name?: string;
  object_key?: string;
  external_url?: string;
  note?: string;
};

type SubcontractMaterialTransferApi = {
  id: string;
  transfer_no: string;
  subcontract_order_id: string;
  subcontract_order_no: string;
  source_warehouse_id: string;
  source_warehouse_code?: string;
  factory_id: string;
  factory_name: string;
  status: "sent_to_factory" | "partially_sent";
  lines: SubcontractMaterialTransferApiLine[];
  evidence?: SubcontractMaterialTransferApiEvidence[];
  handover_by: string;
  handover_at: string;
  received_by: string;
  created_at: string;
};

type SubcontractMaterialIssueMovementApi = {
  movement_no: string;
  movement_type: string;
  item_id: string;
  batch_id?: string;
  warehouse_id: string;
  bin_id?: string;
  quantity: string;
  base_uom_code: string;
  source_quantity: string;
  source_uom_code: string;
  stock_status: string;
  source_doc_id: string;
};

type ReceiveSubcontractFinishedGoodsApiLineRequest = {
  id?: string;
  line_no?: number;
  item_id?: string;
  sku_code?: string;
  item_name?: string;
  batch_id?: string;
  batch_no?: string;
  lot_no?: string;
  expiry_date?: string;
  receive_qty: string;
  uom_code: string;
  base_receive_qty?: string;
  base_uom_code?: string;
  conversion_factor?: string;
  packaging_status: string;
  note?: string;
};

type ReceiveSubcontractFinishedGoodsApiEvidenceRequest = {
  id?: string;
  evidence_type: string;
  file_name?: string;
  object_key?: string;
  external_url?: string;
  note?: string;
};

type ReceiveSubcontractFinishedGoodsApiRequest = {
  expected_version: number;
  receipt_id?: string;
  receipt_no?: string;
  warehouse_id: string;
  warehouse_code: string;
  location_id: string;
  location_code: string;
  delivery_note_no: string;
  received_by: string;
  received_at?: string;
  note?: string;
  lines: ReceiveSubcontractFinishedGoodsApiLineRequest[];
  evidence?: ReceiveSubcontractFinishedGoodsApiEvidenceRequest[];
};

type SubcontractFinishedGoodsReceiptApiLine = {
  id: string;
  line_no: number;
  item_id: string;
  sku_code: string;
  item_name: string;
  batch_id?: string;
  batch_no: string;
  lot_no?: string;
  expiry_date?: string;
  receive_qty: string;
  uom_code: string;
  base_receive_qty: string;
  base_uom_code: string;
  conversion_factor: string;
  packaging_status?: string;
  note?: string;
};

type SubcontractFinishedGoodsReceiptApiEvidence = {
  id: string;
  evidence_type: string;
  file_name?: string;
  object_key?: string;
  external_url?: string;
  note?: string;
};

type SubcontractFinishedGoodsReceiptApi = {
  id: string;
  receipt_no: string;
  subcontract_order_id: string;
  subcontract_order_no: string;
  factory_id: string;
  factory_code?: string;
  factory_name: string;
  warehouse_id: string;
  warehouse_code?: string;
  location_id: string;
  location_code?: string;
  delivery_note_no: string;
  status: "qc_hold";
  lines: SubcontractFinishedGoodsReceiptApiLine[];
  evidence?: SubcontractFinishedGoodsReceiptApiEvidence[];
  received_by: string;
  received_at: string;
  note?: string;
  created_at: string;
  updated_at: string;
  version: number;
};

type ReceiveSubcontractFinishedGoodsApiResult = {
  subcontract_order: SubcontractOrderApi;
  receipt: SubcontractFinishedGoodsReceiptApi;
  stock_movements: SubcontractMaterialIssueMovementApi[];
  audit_log_id?: string;
};

type IssueSubcontractMaterialsApiResult = {
  subcontract_order: SubcontractOrderApi;
  transfer: SubcontractMaterialTransferApi;
  stock_movements: SubcontractMaterialIssueMovementApi[];
  audit_log_id?: string;
};

type SubcontractSampleEvidenceApiRequest = {
  id?: string;
  evidence_type: string;
  file_name?: string;
  object_key?: string;
  external_url?: string;
  note?: string;
};

type SubmitSubcontractSampleApiRequest = {
  expected_version: number;
  sample_approval_id?: string;
  sample_code: string;
  formula_version?: string;
  spec_version?: string;
  submitted_by: string;
  submitted_at?: string;
  note?: string;
  evidence: SubcontractSampleEvidenceApiRequest[];
};

type DecideSubcontractSampleApiRequest = {
  expected_version: number;
  sample_approval_id?: string;
  decision_at?: string;
  reason: string;
  storage_status?: string;
};

type SubcontractSampleEvidenceApi = {
  id: string;
  evidence_type: string;
  file_name?: string;
  object_key?: string;
  external_url?: string;
  note?: string;
  created_at: string;
  created_by: string;
};

type SubcontractSampleApprovalApi = {
  id: string;
  subcontract_order_id: string;
  subcontract_order_no: string;
  sample_code: string;
  formula_version?: string;
  spec_version?: string;
  status: SubcontractSampleApprovalStatus;
  evidence: SubcontractSampleEvidenceApi[];
  submitted_by: string;
  submitted_at: string;
  decision_by?: string;
  decision_at?: string;
  decision_reason?: string;
  storage_status?: string;
  note?: string;
  created_at: string;
  updated_at: string;
  version: number;
};

type SubcontractSampleApprovalApiResult = {
  subcontract_order: SubcontractOrderApi;
  sample_approval: SubcontractSampleApprovalApi;
  previous_status: SubcontractOrderStatus;
  current_status: SubcontractOrderStatus;
  audit_log_id?: string;
};

type SubcontractOrderAction = "submit" | "approve" | "confirm-factory" | "start-mass-production" | "cancel" | "close";

const defaultAccessToken = "local-dev-access-token";
const prototypeNow = "2026-04-29T10:00:00Z";
let subcontractOrderSequence = 2;
let subcontractAuditSequence = 1;
let prototypeSampleApprovalSequence = 1;
let prototypeSampleApprovalStore: SubcontractSampleApproval[] = [];
let prototypeFinishedGoodsReceiptSequence = 1;

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
  { label: "Start mass production", value: "start-mass-production" },
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
    receivedQty: "0.000000",
    acceptedQty: "0.000000",
    rejectedQty: "0.000000",
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

export async function startMassProductionSubcontractOrder(
  id: string,
  expectedVersion?: number
): Promise<SubcontractStatusChangeResult> {
  return runSubcontractOrderAction(id, "start-mass-production", expectedVersion);
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

export async function issueSubcontractMaterials(
  input: IssueSubcontractMaterialsInput
): Promise<IssueSubcontractMaterialsResult> {
  try {
    const result = await apiPost<IssueSubcontractMaterialsApiResult, IssueSubcontractMaterialsApiRequest>(
      `/subcontract-orders/${encodeURIComponent(input.order.id)}/issue-materials`,
      toApiIssueMaterialsInput(input),
      { accessToken: defaultAccessToken }
    );

    return fromApiIssueMaterialsResult(result);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return issuePrototypeSubcontractMaterials(input);
  }
}

export async function submitSubcontractSample(
  input: SubmitSubcontractSampleInput
): Promise<SubcontractSampleApprovalResult> {
  try {
    const result = await apiPost<SubcontractSampleApprovalApiResult, SubmitSubcontractSampleApiRequest>(
      `/subcontract-orders/${encodeURIComponent(input.order.id)}/submit-sample`,
      toApiSubmitSampleInput(input),
      { accessToken: defaultAccessToken }
    );

    return fromApiSampleApprovalResult(result);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return submitPrototypeSubcontractSample(input);
  }
}

export async function approveSubcontractSample(
  input: DecideSubcontractSampleInput
): Promise<SubcontractSampleApprovalResult> {
  try {
    const result = await apiPost<SubcontractSampleApprovalApiResult, DecideSubcontractSampleApiRequest>(
      `/subcontract-orders/${encodeURIComponent(input.order.id)}/approve-sample`,
      toApiDecideSampleInput(input),
      { accessToken: defaultAccessToken }
    );

    return fromApiSampleApprovalResult(result);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return decidePrototypeSubcontractSample(input, "approved");
  }
}

export async function rejectSubcontractSample(
  input: DecideSubcontractSampleInput
): Promise<SubcontractSampleApprovalResult> {
  try {
    const result = await apiPost<SubcontractSampleApprovalApiResult, DecideSubcontractSampleApiRequest>(
      `/subcontract-orders/${encodeURIComponent(input.order.id)}/reject-sample`,
      toApiDecideSampleInput(input),
      { accessToken: defaultAccessToken }
    );

    return fromApiSampleApprovalResult(result);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return decidePrototypeSubcontractSample(input, "rejected");
  }
}

export async function receiveSubcontractFinishedGoods(
  input: ReceiveSubcontractFinishedGoodsInput
): Promise<ReceiveSubcontractFinishedGoodsResult> {
  try {
    const result = await apiPost<ReceiveSubcontractFinishedGoodsApiResult, ReceiveSubcontractFinishedGoodsApiRequest>(
      `/subcontract-orders/${encodeURIComponent(input.order.id)}/receive-finished-goods`,
      toApiReceiveFinishedGoodsInput(input),
      { accessToken: defaultAccessToken }
    );

    return fromApiReceiveFinishedGoodsResult(result);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return receivePrototypeSubcontractFinishedGoods(input);
  }
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

export function availableSubcontractOrderActions(
  status: SubcontractOrderStatus,
  sampleRequired = true
): SubcontractOrderAction[] {
  switch (status) {
    case "draft":
      return ["submit", "cancel"];
    case "submitted":
      return ["approve", "cancel"];
    case "approved":
      return ["confirm-factory", "cancel"];
    case "materials_issued_to_factory":
      return sampleRequired ? [] : ["start-mass-production", "cancel"];
    case "sample_approved":
      return ["start-mass-production", "cancel"];
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
  prototypeSampleApprovalSequence = 1;
  prototypeSampleApprovalStore = [];
  prototypeFinishedGoodsReceiptSequence = 1;
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
    receivedQty: order.received_qty,
    acceptedQty: order.accepted_qty,
    rejectedQty: order.rejected_qty,
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
    auditLogIds: order.audit_log_id ? [order.audit_log_id] : [],
    sampleRejectReason: order.sample_reject_reason
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

function fromApiIssueMaterialsResult(result: IssueSubcontractMaterialsApiResult): IssueSubcontractMaterialsResult {
  const order = fromApiSubcontractOrder(result.subcontract_order);
  const stockMovements = result.stock_movements.map(fromApiMaterialIssueMovement);
  const transfer = fromApiSubcontractMaterialTransfer(result.transfer, stockMovements);
  const auditLog = createMaterialIssueAuditLog(order, transfer, result.audit_log_id);

  return {
    order,
    transfer,
    stockMovements,
    auditLog,
    auditLogId: result.audit_log_id
  };
}

function fromApiSampleApprovalResult(result: SubcontractSampleApprovalApiResult): SubcontractSampleApprovalResult {
  const order = fromApiSubcontractOrder(result.subcontract_order);
  const sampleApproval = fromApiSampleApproval(result.sample_approval);
  const auditLog = createSampleApprovalAuditLog(
    order,
    sampleApproval,
    result.previous_status,
    result.current_status,
    result.audit_log_id
  );

  return {
    order,
    sampleApproval,
    auditLog,
    auditLogId: result.audit_log_id
  };
}

function fromApiReceiveFinishedGoodsResult(
  result: ReceiveSubcontractFinishedGoodsApiResult
): ReceiveSubcontractFinishedGoodsResult {
  const order = fromApiSubcontractOrder(result.subcontract_order);
  const receipt = fromApiSubcontractFinishedGoodsReceipt(result.receipt);
  const stockMovements = result.stock_movements.map((movement) => fromApiSubcontractStockMovement(movement));
  const auditLog = createFinishedGoodsReceiptAuditLog(order, receipt, stockMovements, result.audit_log_id);

  return {
    order,
    receipt,
    stockMovements,
    auditLog,
    auditLogId: result.audit_log_id
  };
}

function fromApiSampleApproval(sampleApproval: SubcontractSampleApprovalApi): SubcontractSampleApproval {
  return {
    id: sampleApproval.id,
    orderId: sampleApproval.subcontract_order_id,
    orderNo: sampleApproval.subcontract_order_no,
    sampleCode: sampleApproval.sample_code,
    formulaVersion: sampleApproval.formula_version,
    specVersion: sampleApproval.spec_version,
    status: sampleApproval.status,
    evidence: sampleApproval.evidence.map((evidence) => ({
      id: evidence.id,
      evidenceType: evidence.evidence_type,
      fileName: evidence.file_name,
      objectKey: evidence.object_key,
      externalURL: evidence.external_url,
      note: evidence.note,
      createdAt: evidence.created_at,
      createdBy: evidence.created_by
    })),
    submittedBy: sampleApproval.submitted_by,
    submittedAt: sampleApproval.submitted_at,
    decisionBy: sampleApproval.decision_by,
    decisionAt: sampleApproval.decision_at,
    decisionReason: sampleApproval.decision_reason,
    storageStatus: sampleApproval.storage_status,
    note: sampleApproval.note,
    createdAt: sampleApproval.created_at,
    updatedAt: sampleApproval.updated_at,
    version: sampleApproval.version
  };
}

function fromApiSubcontractFinishedGoodsReceipt(
  receipt: SubcontractFinishedGoodsReceiptApi
): SubcontractFinishedGoodsReceipt {
  return {
    id: receipt.id,
    receiptNo: receipt.receipt_no,
    orderId: receipt.subcontract_order_id,
    orderNo: receipt.subcontract_order_no,
    factoryId: receipt.factory_id,
    factoryCode: receipt.factory_code,
    factoryName: receipt.factory_name,
    warehouseId: receipt.warehouse_id,
    warehouseCode: receipt.warehouse_code ?? receipt.warehouse_id,
    locationId: receipt.location_id,
    locationCode: receipt.location_code ?? receipt.location_id,
    deliveryNoteNo: receipt.delivery_note_no,
    status: receipt.status,
    lines: receipt.lines.map((line) => ({
      id: line.id,
      lineNo: line.line_no,
      itemId: line.item_id,
      skuCode: line.sku_code,
      itemName: line.item_name,
      batchId: line.batch_id,
      batchNo: line.batch_no,
      lotNo: line.lot_no,
      expiryDate: line.expiry_date ?? "",
      receiveQty: line.receive_qty,
      uomCode: line.uom_code,
      baseReceiveQty: line.base_receive_qty,
      baseUOMCode: line.base_uom_code,
      conversionFactor: line.conversion_factor,
      packagingStatus: line.packaging_status,
      note: line.note
    })),
    evidence: (receipt.evidence ?? []).map((evidence) => ({
      id: evidence.id,
      evidenceType: evidence.evidence_type,
      fileName: evidence.file_name,
      objectKey: evidence.object_key,
      externalURL: evidence.external_url,
      note: evidence.note
    })),
    receivedBy: receipt.received_by,
    receivedAt: receipt.received_at,
    note: receipt.note,
    createdAt: receipt.created_at,
    updatedAt: receipt.updated_at,
    version: receipt.version
  };
}

function fromApiSubcontractMaterialTransfer(
  transfer: SubcontractMaterialTransferApi,
  stockMovements: SubcontractStockMovement[]
): SubcontractMaterialTransfer {
  const lines = transfer.lines.map((line) => ({
    id: line.id,
    itemCode: line.sku_code,
    itemName: line.item_name,
    itemType: line.sku_code.startsWith("PK") ? ("packaging" as const) : ("raw_material" as const),
    quantity: Number(line.issue_qty),
    unit: line.uom_code,
    lotControlled: line.lot_trace_required,
    batchNo: line.batch_no || line.lot_no || line.batch_id,
    qcStatus: "passed" as const
  }));

  return {
    id: transfer.id,
    transferNo: transfer.transfer_no,
    orderId: transfer.subcontract_order_id,
    orderNo: transfer.subcontract_order_no,
    sourceWarehouseId: transfer.source_warehouse_id,
    sourceWarehouseCode: transfer.source_warehouse_code ?? transfer.source_warehouse_id,
    factoryId: transfer.factory_id,
    factoryName: transfer.factory_name,
    signedHandover: transfer.status === "sent_to_factory",
    status: transfer.status === "sent_to_factory" ? "SENT" : "READY_TO_SEND",
    attachmentPlaceholders: createIssueAttachmentPlaceholders(transfer.evidence ?? []),
    lines,
    stockMovements,
    createdBy: transfer.handover_by,
    createdAt: transfer.created_at || transfer.handover_at
  };
}

function fromApiMaterialIssueMovement(movement: SubcontractMaterialIssueMovementApi): SubcontractStockMovement {
  return {
    ...fromApiSubcontractStockMovement(movement, "SUBCONTRACT_ISSUE"),
    targetLocation: `stock_in_subcontractor_hold:${movement.stock_status}`
  };
}

function fromApiSubcontractStockMovement(
  movement: SubcontractMaterialIssueMovementApi,
  fallbackMovementType?: SubcontractStockMovement["movementType"]
): SubcontractStockMovement {
  return {
    id: movement.movement_no,
    movementType:
      movement.movement_type === "subcontract_receipt"
        ? "SUBCONTRACT_RECEIPT"
        : fallbackMovementType ?? "SUBCONTRACT_ISSUE",
    itemCode: movement.item_id,
    quantity: Number(movement.source_quantity || movement.quantity),
    unit: movement.source_uom_code || movement.base_uom_code,
    sourceWarehouseId: movement.warehouse_id,
    targetLocation: `${movement.warehouse_id}/${movement.bin_id || "qc_hold"}:${movement.stock_status}`,
    batchNo: movement.batch_id,
    sourceDocId: movement.source_doc_id
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

function toApiIssueMaterialsInput(input: IssueSubcontractMaterialsInput): IssueSubcontractMaterialsApiRequest {
  return {
    expected_version: input.order.version,
    source_warehouse_id: input.sourceWarehouseId,
    source_warehouse_code: input.sourceWarehouseCode,
    handover_by: input.handoverBy,
    handover_at: input.handoverAt,
    received_by: input.receivedBy,
    receiver_contact: input.receiverContact,
    vehicle_no: input.vehicleNo,
    note: input.note,
    lines: input.lines.map((line) => ({
      order_material_line_id: line.orderMaterialLineId,
      issue_qty: normalizeDecimalInput(line.issueQty, decimalScales.quantity),
      uom_code: line.uomCode,
      base_issue_qty: line.baseIssueQty ? normalizeDecimalInput(line.baseIssueQty, decimalScales.quantity) : undefined,
      base_uom_code: line.baseUOMCode,
      conversion_factor: line.conversionFactor
        ? normalizeDecimalInput(line.conversionFactor, decimalScales.quantity)
        : undefined,
      batch_id: line.batchId,
      batch_no: line.batchNo,
      lot_no: line.lotNo,
      source_bin_id: line.sourceBinId,
      note: line.note
    })),
    evidence: input.evidence?.map((evidence) => ({
      id: evidence.id,
      evidence_type: evidence.evidenceType,
      file_name: evidence.fileName,
      object_key: evidence.objectKey,
      external_url: evidence.externalURL,
      note: evidence.note
    }))
  };
}

function toApiSubmitSampleInput(input: SubmitSubcontractSampleInput): SubmitSubcontractSampleApiRequest {
  return {
    expected_version: input.order.version,
    sample_approval_id: input.sampleApprovalId,
    sample_code: input.sampleCode,
    formula_version: input.formulaVersion,
    spec_version: input.specVersion,
    submitted_by: input.submittedBy,
    submitted_at: input.submittedAt,
    note: input.note,
    evidence: input.evidence.map((evidence) => ({
      id: evidence.id,
      evidence_type: evidence.evidenceType,
      file_name: evidence.fileName,
      object_key: evidence.objectKey,
      external_url: evidence.externalURL,
      note: evidence.note
    }))
  };
}

function toApiDecideSampleInput(input: DecideSubcontractSampleInput): DecideSubcontractSampleApiRequest {
  return {
    expected_version: input.order.version,
    sample_approval_id: input.sampleApprovalId,
    decision_at: input.decisionAt,
    reason: input.reason,
    storage_status: input.storageStatus
  };
}

function toApiReceiveFinishedGoodsInput(
  input: ReceiveSubcontractFinishedGoodsInput
): ReceiveSubcontractFinishedGoodsApiRequest {
  return {
    expected_version: input.order.version,
    receipt_id: input.receiptId,
    receipt_no: input.receiptNo,
    warehouse_id: input.warehouseId,
    warehouse_code: input.warehouseCode,
    location_id: input.locationId,
    location_code: input.locationCode,
    delivery_note_no: input.deliveryNoteNo,
    received_by: input.receivedBy,
    received_at: input.receivedAt,
    note: input.note,
    lines: input.lines.map((line) => ({
      id: line.id,
      line_no: line.lineNo,
      item_id: line.itemId,
      sku_code: line.skuCode,
      item_name: line.itemName,
      batch_id: line.batchId,
      batch_no: line.batchNo,
      lot_no: line.lotNo,
      expiry_date: line.expiryDate,
      receive_qty: normalizeDecimalInput(line.receiveQty, decimalScales.quantity),
      uom_code: line.uomCode,
      base_receive_qty: line.baseReceiveQty
        ? normalizeDecimalInput(line.baseReceiveQty, decimalScales.quantity)
        : undefined,
      base_uom_code: line.baseUOMCode,
      conversion_factor: line.conversionFactor
        ? normalizeDecimalInput(line.conversionFactor, decimalScales.quantity)
        : undefined,
      packaging_status: line.packagingStatus,
      note: line.note
    })),
    evidence: input.evidence?.map((evidence) => ({
      id: evidence.id,
      evidence_type: evidence.evidenceType,
      file_name: evidence.fileName,
      object_key: evidence.objectKey,
      external_url: evidence.externalURL,
      note: evidence.note
    }))
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
    receivedQty: "0.000000",
    acceptedQty: "0.000000",
    rejectedQty: "0.000000",
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
  const nextStatus = nextPrototypeStatus(current, action);
  const result = changeSubcontractOrderStatus({
    order: current,
    nextStatus,
    actorName: "Subcontract Coordinator",
    note: reason
  });
  prototypeStore = [result.order, ...prototypeStore.filter((candidate) => candidate.id !== id)];

  return result;
}

function issuePrototypeSubcontractMaterials(input: IssueSubcontractMaterialsInput): IssueSubcontractMaterialsResult {
  const current = getPrototypeSubcontractOrder(input.order.id);
  if (input.order.version && input.order.version !== current.version) {
    throw new Error("Subcontract order version changed");
  }
  if (!["factory_confirmed", "deposit_recorded"].includes(current.status)) {
    throw new Error(`Cannot issue materials from ${formatSubcontractOrderStatus(current.status)}`);
  }
  const transferId = `sub-transfer-${current.id}-${String(subcontractAuditSequence).padStart(4, "0")}`;
  const updatedLines = current.materialLines.map((line) => {
    const issuedLine = input.lines.find((candidate) => candidate.orderMaterialLineId === line.id);
    if (!issuedLine) {
      return line;
    }

    const issueQty = normalizeDecimalInput(issuedLine.issueQty, decimalScales.quantity);
    if (toScaledBigInt(issueQty, decimalScales.quantity) <= BigInt(0)) {
      throw new Error("Issue quantity must be greater than zero");
    }
    const nextIssuedQty = addQuantityStrings(line.issuedQty, issueQty);
    if (toScaledBigInt(nextIssuedQty, decimalScales.quantity) > toScaledBigInt(line.plannedQty, decimalScales.quantity)) {
      throw new Error(`${line.skuCode} issue quantity exceeds remaining subcontract material quantity`);
    }

    return {
      ...line,
      issuedQty: nextIssuedQty
    };
  });
  const allIssued = updatedLines.every((line) => isQuantityGreaterOrEqual(line.issuedQty, line.plannedQty));
  const order = createSubcontractOrderRecord({
    ...current,
    status: allIssued ? "materials_issued_to_factory" : current.status,
    updatedAt: prototypeNow,
    version: current.version + 1,
    materialLines: updatedLines
  });
  const transferLines = input.lines.map((line) => {
    const orderLine = current.materialLines.find((candidate) => candidate.id === line.orderMaterialLineId);
    if (!orderLine) {
      throw new Error("Material line is required");
    }
    if (orderLine.lotTraceRequired && !line.batchNo?.trim() && !line.batchId?.trim() && !line.lotNo?.trim()) {
      throw new Error(`${orderLine.skuCode} requires batch or lot before factory transfer`);
    }

    return {
      id: `${transferId}-${orderLine.id}`,
      itemCode: orderLine.skuCode,
      itemName: orderLine.itemName,
      itemType: orderLine.skuCode.startsWith("PK") ? ("packaging" as const) : ("raw_material" as const),
      quantity: Number(line.issueQty),
      unit: line.uomCode,
      lotControlled: orderLine.lotTraceRequired,
      batchNo: line.batchNo || line.lotNo || line.batchId,
      qcStatus: "passed" as const
    };
  });
  const stockMovements = transferLines.map((line) => ({
    id: `mov-${transferId}-${line.id}`,
    movementType: "SUBCONTRACT_ISSUE" as const,
    itemCode: line.itemCode,
    quantity: line.quantity,
    unit: line.unit,
    sourceWarehouseId: input.sourceWarehouseId,
    targetLocation: `stock_in_subcontractor_hold:${current.factoryCode}`,
    batchNo: line.batchNo,
    sourceDocId: transferId
  }));
  const transfer: SubcontractMaterialTransfer = {
    id: transferId,
    transferNo: `SUBTR-260429-${String(subcontractAuditSequence).padStart(4, "0")}`,
    orderId: current.id,
    orderNo: current.orderNo,
    sourceWarehouseId: input.sourceWarehouseId,
    sourceWarehouseCode: input.sourceWarehouseCode,
    factoryId: current.factoryId,
    factoryName: current.factoryName,
    signedHandover: true,
    status: allIssued ? "SENT" : "READY_TO_SEND",
    attachmentPlaceholders: createIssueAttachmentPlaceholders(
      input.evidence?.map((evidence) => ({
        id: evidence.id ?? evidence.evidenceType,
        evidence_type: evidence.evidenceType,
        file_name: evidence.fileName,
        object_key: evidence.objectKey,
        external_url: evidence.externalURL
      })) ?? []
    ),
    lines: transferLines,
    stockMovements,
    createdBy: input.handoverBy || "Subcontract Coordinator",
    createdAt: input.handoverAt || prototypeNow
  };
  const auditLog = createMaterialIssueAuditLog(order, transfer, undefined, current.status);
  order.auditLogIds = [...order.auditLogIds, auditLog.id];
  prototypeStore = [order, ...prototypeStore.filter((candidate) => candidate.id !== current.id)];

  return {
    order,
    transfer,
    stockMovements,
    auditLog,
    auditLogId: auditLog.id
  };
}

function submitPrototypeSubcontractSample(input: SubmitSubcontractSampleInput): SubcontractSampleApprovalResult {
  const current = getPrototypeSubcontractOrder(input.order.id);
  if (input.order.version && input.order.version !== current.version) {
    throw new Error("Subcontract order version changed");
  }
  if (!["materials_issued_to_factory", "sample_rejected"].includes(current.status)) {
    throw new Error(`Cannot submit sample from ${formatSubcontractOrderStatus(current.status)}`);
  }
  if (input.evidence.length === 0) {
    throw new Error("Sample evidence is required");
  }
  const sampleApproval: SubcontractSampleApproval = {
    id: input.sampleApprovalId || `sample-${current.id}-${String(prototypeSampleApprovalSequence++).padStart(4, "0")}`,
    orderId: current.id,
    orderNo: current.orderNo,
    sampleCode: input.sampleCode.trim() || `${current.orderNo}-SAMPLE-${prototypeSampleApprovalSequence}`,
    formulaVersion: input.formulaVersion,
    specVersion: input.specVersion || current.specVersion,
    status: "submitted",
    evidence: input.evidence.map((evidence, index) => ({
      id: evidence.id || `${current.id}-sample-evidence-${index + 1}`,
      evidenceType: evidence.evidenceType,
      fileName: evidence.fileName,
      objectKey: evidence.objectKey,
      externalURL: evidence.externalURL,
      note: evidence.note,
      createdAt: input.submittedAt || prototypeNow,
      createdBy: input.submittedBy || "factory-user"
    })),
    submittedBy: input.submittedBy || "factory-user",
    submittedAt: input.submittedAt || prototypeNow,
    note: input.note,
    createdAt: input.submittedAt || prototypeNow,
    updatedAt: input.submittedAt || prototypeNow,
    version: 1
  };
  const order = createSubcontractOrderRecord({
    ...current,
    status: "sample_submitted",
    sampleRejectReason: undefined,
    updatedAt: prototypeNow,
    version: current.version + 1
  });
  const auditLog = createSampleApprovalAuditLog(order, sampleApproval, current.status, order.status);
  order.auditLogIds = [...order.auditLogIds, auditLog.id];
  prototypeStore = [order, ...prototypeStore.filter((candidate) => candidate.id !== current.id)];
  prototypeSampleApprovalStore = [
    sampleApproval,
    ...prototypeSampleApprovalStore.filter((candidate) => candidate.id !== sampleApproval.id)
  ];

  return {
    order,
    sampleApproval,
    auditLog,
    auditLogId: auditLog.id
  };
}

function decidePrototypeSubcontractSample(
  input: DecideSubcontractSampleInput,
  status: Exclude<SubcontractSampleApprovalStatus, "submitted">
): SubcontractSampleApprovalResult {
  const current = getPrototypeSubcontractOrder(input.order.id);
  if (input.order.version && input.order.version !== current.version) {
    throw new Error("Subcontract order version changed");
  }
  if (current.status !== "sample_submitted") {
    throw new Error(`Cannot decide sample from ${formatSubcontractOrderStatus(current.status)}`);
  }
  const existingSample = latestPrototypeSampleApproval(current.id, input.sampleApprovalId);
  if (!existingSample) {
    throw new Error("Sample approval record is required");
  }
  if (status === "rejected" && input.reason.trim() === "") {
    throw new Error("Sample rejection reason is required");
  }
  if (status === "approved" && !input.storageStatus?.trim()) {
    throw new Error("Approved sample storage status is required");
  }
  const sampleApproval: SubcontractSampleApproval = {
    ...existingSample,
    status,
    decisionBy: "qa-lead",
    decisionAt: input.decisionAt || prototypeNow,
    decisionReason: input.reason,
    storageStatus: status === "approved" ? input.storageStatus : undefined,
    updatedAt: input.decisionAt || prototypeNow,
    version: existingSample.version + 1
  };
  const nextStatus: SubcontractOrderStatus = status === "approved" ? "sample_approved" : "sample_rejected";
  const order = createSubcontractOrderRecord({
    ...current,
    status: nextStatus,
    updatedAt: prototypeNow,
    version: current.version + 1,
    auditLogIds: [...current.auditLogIds],
    ...(status === "rejected" ? { sampleRejectReason: input.reason } : {})
  });
  const auditLog = createSampleApprovalAuditLog(order, sampleApproval, current.status, order.status);
  order.auditLogIds = [...order.auditLogIds, auditLog.id];
  prototypeStore = [order, ...prototypeStore.filter((candidate) => candidate.id !== current.id)];
  prototypeSampleApprovalStore = [
    sampleApproval,
    ...prototypeSampleApprovalStore.filter((candidate) => candidate.id !== sampleApproval.id)
  ];

  return {
    order,
    sampleApproval,
    auditLog,
    auditLogId: auditLog.id
  };
}

function receivePrototypeSubcontractFinishedGoods(
  input: ReceiveSubcontractFinishedGoodsInput
): ReceiveSubcontractFinishedGoodsResult {
  const current = getPrototypeSubcontractOrder(input.order.id);
  if (input.order.version && input.order.version !== current.version) {
    throw new Error("Subcontract order version changed");
  }
  if (!["mass_production_started", "finished_goods_received"].includes(current.status)) {
    throw new Error(`Cannot receive finished goods from ${formatSubcontractOrderStatus(current.status)}`);
  }
  if (input.lines.length === 0) {
    throw new Error("At least one finished goods receipt line is required");
  }

  const sequence = prototypeFinishedGoodsReceiptSequence++;
  const receivedAt = input.receivedAt || prototypeNow;
  const receiptId = input.receiptId || `sfgr-${current.id}-${String(sequence).padStart(4, "0")}`;
  const receiptNo = input.receiptNo || `SFGR-260429-${String(sequence).padStart(4, "0")}`;
  const deliveryNoteNo = normalizeRequiredText(input.deliveryNoteNo, "Factory delivery note is required");
  const warehouseId = normalizeRequiredText(input.warehouseId, "Warehouse is required");
  const warehouseCode = normalizeRequiredText(input.warehouseCode, "Warehouse code is required");
  const locationId = normalizeRequiredText(input.locationId, "QC hold location is required");
  const locationCode = normalizeRequiredText(input.locationCode, "QC hold location code is required");
  const receivedBy = normalizeRequiredText(input.receivedBy, "Receiver is required");
  const lines = input.lines.map((line, index) => {
    const receiveQty = normalizeDecimalInput(line.receiveQty, decimalScales.quantity);
    if (toScaledBigInt(receiveQty, decimalScales.quantity) <= BigInt(0)) {
      throw new Error("Receive quantity must be greater than zero");
    }

    return {
      id: line.id || `${receiptId}-line-${String(index + 1).padStart(2, "0")}`,
      lineNo: line.lineNo ?? index + 1,
      itemId: line.itemId || current.productId,
      skuCode: line.skuCode || current.sku,
      itemName: line.itemName || current.productName,
      batchId: line.batchId,
      batchNo: normalizeRequiredText(line.batchNo, "Batch / lot is required"),
      lotNo: line.lotNo || line.batchNo,
      expiryDate: normalizeRequiredText(line.expiryDate, "Expiry date is required"),
      receiveQty,
      uomCode: line.uomCode || "EA",
      baseReceiveQty: normalizeDecimalInput(line.baseReceiveQty || line.receiveQty, decimalScales.quantity),
      baseUOMCode: line.baseUOMCode || line.uomCode || "EA",
      conversionFactor: normalizeDecimalInput(line.conversionFactor || "1", decimalScales.quantity),
      packagingStatus: line.packagingStatus,
      note: line.note
    };
  });
  const totalReceivedQty = lines.reduce(
    (sum, line) => addQuantityStrings(sum, line.receiveQty),
    "0.000000"
  );
  const nextReceivedQty = addQuantityStrings(current.receivedQty ?? "0.000000", totalReceivedQty);
  const plannedQty = normalizeDecimalInput(String(current.quantity), decimalScales.quantity);
  if (toScaledBigInt(nextReceivedQty, decimalScales.quantity) > toScaledBigInt(plannedQty, decimalScales.quantity)) {
    throw new Error("Receive quantity exceeds remaining subcontract finished goods quantity");
  }

  const order = createSubcontractOrderRecord({
    ...current,
    status: "finished_goods_received",
    receivedQty: nextReceivedQty,
    updatedAt: receivedAt,
    version: current.version + 1
  });
  const receipt: SubcontractFinishedGoodsReceipt = {
    id: receiptId,
    receiptNo,
    orderId: current.id,
    orderNo: current.orderNo,
    factoryId: current.factoryId,
    factoryCode: current.factoryCode,
    factoryName: current.factoryName,
    warehouseId,
    warehouseCode,
    locationId,
    locationCode,
    deliveryNoteNo,
    status: "qc_hold",
    lines,
    evidence: (input.evidence ?? []).map((evidence, index) => ({
      id: evidence.id || `${receiptId}-evidence-${index + 1}`,
      evidenceType: evidence.evidenceType,
      fileName: evidence.fileName,
      objectKey: evidence.objectKey,
      externalURL: evidence.externalURL,
      note: evidence.note
    })),
    receivedBy,
    receivedAt,
    note: input.note,
    createdAt: receivedAt,
    updatedAt: receivedAt,
    version: 1
  };
  const stockMovements = receipt.lines.map((line) => ({
    id: `mov-${receipt.id}-${line.id}`,
    movementType: "SUBCONTRACT_RECEIPT" as const,
    itemCode: line.skuCode,
    quantity: Number(line.receiveQty),
    unit: line.uomCode,
    sourceWarehouseId: warehouseId,
    targetLocation: `${warehouseCode}/${locationCode}:qc_hold`,
    batchNo: line.batchNo,
    sourceDocId: receipt.id
  }));
  const auditLog = createFinishedGoodsReceiptAuditLog(order, receipt, stockMovements, undefined, current.status);
  order.auditLogIds = [...order.auditLogIds, auditLog.id];
  prototypeStore = [order, ...prototypeStore.filter((candidate) => candidate.id !== current.id)];

  return {
    order,
    receipt,
    stockMovements,
    auditLog,
    auditLogId: auditLog.id
  };
}

function latestPrototypeSampleApproval(orderId: string, sampleApprovalId?: string) {
  const samples = prototypeSampleApprovalStore.filter((sample) =>
    sampleApprovalId ? sample.id === sampleApprovalId : sample.orderId === orderId
  );

  return samples[0];
}

function nextPrototypeStatus(order: SubcontractOrder, action: SubcontractOrderAction) {
  const currentStatus = order.status;
  if (action === "submit" && currentStatus === "draft") {
    return "submitted";
  }
  if (action === "approve" && currentStatus === "submitted") {
    return "approved";
  }
  if (action === "confirm-factory" && currentStatus === "approved") {
    return "factory_confirmed";
  }
  if (action === "start-mass-production" && currentStatus === "sample_approved") {
    return "mass_production_started";
  }
  if (action === "start-mass-production" && currentStatus === "materials_issued_to_factory" && !order.sampleRequired) {
    return "mass_production_started";
  }
  if (action === "close" && ["accepted", "final_payment_ready"].includes(currentStatus)) {
    return "closed";
  }
  if (
    action === "cancel" &&
    ["draft", "submitted", "approved", "factory_confirmed", "materials_issued_to_factory", "sample_approved"].includes(
      currentStatus
    )
  ) {
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

function createMaterialIssueAuditLog(
  order: SubcontractOrder,
  transfer: SubcontractMaterialTransfer,
  auditLogId?: string,
  beforeStatus: SubcontractOrderStatus = "factory_confirmed"
): AuditLogItem {
  const id = auditLogId ?? `audit-subcontract-material-issue-${String(subcontractAuditSequence++).padStart(4, "0")}`;

  return {
    id,
    actorId: "user-subcontract-coordinator",
    action: "subcontract.materials_issued",
    entityType: "subcontract_order",
    entityId: order.id,
    requestId: `req_${id}`,
    beforeData: {
      status: beforeStatus
    },
    afterData: {
      status: order.status,
      transfer_no: transfer.transferNo,
      stock_movement_count: transfer.stockMovements.length
    },
    metadata: {
      order_no: order.orderNo,
      factory: order.factoryName,
      source_warehouse_id: transfer.sourceWarehouseId,
      actor_name: transfer.createdBy
    },
    createdAt: transfer.createdAt || prototypeNow
  };
}

function createSampleApprovalAuditLog(
  order: SubcontractOrder,
  sampleApproval: SubcontractSampleApproval,
  beforeStatus: SubcontractOrderStatus,
  afterStatus: SubcontractOrderStatus,
  auditLogId?: string
): AuditLogItem {
  const id = auditLogId ?? `audit-subcontract-sample-${String(subcontractAuditSequence++).padStart(4, "0")}`;
  const action =
    sampleApproval.status === "approved"
      ? "subcontract.sample_approved"
      : sampleApproval.status === "rejected"
        ? "subcontract.sample_rejected"
        : "subcontract.sample_submitted";

  return {
    id,
    actorId: "user-subcontract-qa",
    action,
    entityType: "subcontract_order",
    entityId: order.id,
    requestId: `req_${id}`,
    beforeData: {
      status: beforeStatus
    },
    afterData: {
      status: afterStatus,
      sample_approval_id: sampleApproval.id,
      sample_status: sampleApproval.status
    },
    metadata: {
      order_no: order.orderNo,
      sample_code: sampleApproval.sampleCode,
      evidence_count: sampleApproval.evidence.length,
      decision_reason: sampleApproval.decisionReason ?? "",
      storage_status: sampleApproval.storageStatus ?? ""
    },
    createdAt: sampleApproval.updatedAt || prototypeNow
  };
}

function createFinishedGoodsReceiptAuditLog(
  order: SubcontractOrder,
  receipt: SubcontractFinishedGoodsReceipt,
  stockMovements: SubcontractStockMovement[],
  auditLogId?: string,
  beforeStatus: SubcontractOrderStatus = "mass_production_started"
): AuditLogItem {
  const id = auditLogId ?? `audit-subcontract-fg-receipt-${String(subcontractAuditSequence++).padStart(4, "0")}`;

  return {
    id,
    actorId: "user-subcontract-warehouse",
    action: "subcontract.finished_goods_received",
    entityType: "subcontract_order",
    entityId: order.id,
    requestId: `req_${id}`,
    beforeData: {
      status: beforeStatus
    },
    afterData: {
      status: order.status,
      receipt_no: receipt.receiptNo,
      receipt_status: receipt.status,
      stock_movement_count: stockMovements.length
    },
    metadata: {
      order_no: order.orderNo,
      factory: order.factoryName,
      delivery_note_no: receipt.deliveryNoteNo,
      warehouse: receipt.warehouseCode,
      location: receipt.locationCode,
      received_by: receipt.receivedBy
    },
    createdAt: receipt.receivedAt || prototypeNow
  };
}

function addQuantityStrings(left: string, right: string) {
  return fromScaledBigInt(
    toScaledBigInt(left, decimalScales.quantity) + toScaledBigInt(right, decimalScales.quantity),
    decimalScales.quantity
  );
}

function isQuantityGreaterOrEqual(left: string, right: string) {
  return toScaledBigInt(left, decimalScales.quantity) >= toScaledBigInt(right, decimalScales.quantity);
}

function toScaledBigInt(value: string, scale: number) {
  const normalized = normalizeDecimalInput(value, scale);
  return BigInt(normalized.replace(".", ""));
}

function fromScaledBigInt(value: bigint, scale: number) {
  const negative = value < BigInt(0);
  const digits = (negative ? -value : value).toString().padStart(scale + 1, "0");
  const integer = digits.slice(0, -scale);
  const fraction = scale > 0 ? `.${digits.slice(-scale)}` : "";

  return `${negative ? "-" : ""}${integer}${fraction}`;
}

function createIssueAttachmentPlaceholders(
  evidence: SubcontractMaterialTransferApiEvidence[]
): SubcontractMaterialTransfer["attachmentPlaceholders"] {
  const attachedTypes = new Set(evidence.map((item) => item.evidence_type.toLowerCase()));

  return [
    { type: "COA", label: "COA", required: false, attached: attachedTypes.has("coa") },
    { type: "MSDS", label: "MSDS", required: false, attached: attachedTypes.has("msds") },
    { type: "LABEL", label: "Label", required: false, attached: attachedTypes.has("label") },
    { type: "VAT_INVOICE", label: "VAT invoice", required: true, attached: attachedTypes.has("vat_invoice") }
  ];
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
