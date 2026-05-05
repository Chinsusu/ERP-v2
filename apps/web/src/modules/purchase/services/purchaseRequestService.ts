import { apiGetRaw, apiPost } from "../../../shared/api/client";
import { shouldUsePrototypeFallback } from "../../../shared/api/prototypeFallback";
import { formatPurchaseDate, formatPurchaseQuantity, createPurchaseOrder } from "./purchaseOrderService";
import { getProductionPlans } from "../../production-planning/services/productionPlanService";
import type { ProductionPlan, PurchaseRequestDraft } from "../../production-planning/types";
import type {
  ConvertPurchaseRequestToPurchaseOrderInput,
  ConvertPurchaseRequestToPurchaseOrderResult,
  CreatePurchaseOrderInput,
  PurchaseOrder,
  PurchaseRequest,
  PurchaseRequestActionResult,
  PurchaseRequestLine,
  PurchaseRequestQuery,
  PurchaseRequestStatus
} from "../types";

type PurchaseRequestApi = {
  id?: string;
  request_no?: string;
  source_production_plan_id?: string;
  source_production_plan_no?: string;
  status?: PurchaseRequestStatus;
  lines: PurchaseRequestLineApi[];
  created_at?: string;
  created_by?: string;
  submitted_at?: string;
  submitted_by?: string;
  approved_at?: string;
  approved_by?: string;
  converted_at?: string;
  converted_by?: string;
  converted_purchase_order_id?: string;
  converted_purchase_order_no?: string;
  cancelled_at?: string;
  cancelled_by?: string;
  rejected_at?: string;
  rejected_by?: string;
  reject_reason?: string;
};

type PurchaseRequestLineApi = {
  id: string;
  line_no: number;
  source_production_plan_line_id: string;
  item_id?: string;
  sku: string;
  item_name: string;
  requested_qty: string;
  uom_code: string;
  note?: string;
};

type PurchaseRequestActionApiResult = {
  purchase_request: PurchaseRequestApi;
  previous_status: PurchaseRequestStatus;
  current_status: PurchaseRequestStatus;
  audit_log_id?: string;
};

type ConvertPurchaseRequestApiRequest = {
  supplier_id: string;
  warehouse_id: string;
  expected_date: string;
  currency_code?: string;
  unit_price?: string;
};

type ConvertPurchaseRequestApiResult = {
  purchase_request: PurchaseRequestApi;
  purchase_order: PurchaseOrderApi;
  audit_log_id?: string;
};

type PurchaseOrderApi = {
  id: string;
  po_no: string;
  supplier_id: string;
  supplier_code?: string;
  supplier_name: string;
  warehouse_id: string;
  warehouse_code?: string;
  expected_date: string;
  status: PurchaseOrder["status"];
  currency_code: string;
  subtotal_amount: string;
  total_amount: string;
  note?: string;
  lines: Array<{
    id: string;
    line_no: number;
    item_id: string;
    sku_code: string;
    item_name: string;
    ordered_qty: string;
    received_qty: string;
    uom_code: string;
    base_ordered_qty: string;
    base_received_qty: string;
    base_uom_code: string;
    conversion_factor: string;
    unit_price: string;
    currency_code: string;
    line_amount: string;
    expected_date: string;
    note?: string;
  }>;
  audit_log_id?: string;
  created_at: string;
  updated_at: string;
  submitted_at?: string;
  approved_at?: string;
  closed_at?: string;
  cancelled_at?: string;
  rejected_at?: string;
  cancel_reason?: string;
  reject_reason?: string;
  version: number;
};

const defaultAccessToken = "local-dev-access-token";
const localStatusOverrides = new Map<string, Partial<PurchaseRequest>>();

export async function getPurchaseRequests(query: PurchaseRequestQuery = {}): Promise<PurchaseRequest[]> {
  try {
    const requests = await apiGetRaw<PurchaseRequestApi[]>(`/purchase-requests${purchaseRequestQueryString(query)}`, {
      accessToken: defaultAccessToken
    });

    return requests.map(fromApiPurchaseRequest);
  } catch (reason) {
    if (!shouldUsePrototypeFallback(reason)) {
      throw reason;
    }

    const plans = await getProductionPlans();
    return filterPurchaseRequests(plans.flatMap(purchaseRequestsFromProductionPlan), query);
  }
}

export async function getPurchaseRequest(id: string): Promise<PurchaseRequest> {
  try {
    const request = await apiGetRaw<PurchaseRequestApi>(`/purchase-requests/${encodeURIComponent(id)}`, {
      accessToken: defaultAccessToken
    });

    return fromApiPurchaseRequest(request);
  } catch (reason) {
    if (!shouldUsePrototypeFallback(reason)) {
      throw reason;
    }

    const requests = await getPurchaseRequests();
    const request = requests.find((candidate) => candidate.id === id || candidate.requestNo === id);
    if (!request) {
      throw new Error("Purchase request not found");
    }

    return request;
  }
}

export async function submitPurchaseRequest(id: string): Promise<PurchaseRequestActionResult> {
  return runPurchaseRequestAction(id, "submit");
}

export async function approvePurchaseRequest(id: string): Promise<PurchaseRequestActionResult> {
  return runPurchaseRequestAction(id, "approve");
}

export async function convertPurchaseRequestToPurchaseOrder(
  id: string,
  input: ConvertPurchaseRequestToPurchaseOrderInput
): Promise<ConvertPurchaseRequestToPurchaseOrderResult> {
  try {
    const result = await apiPost<ConvertPurchaseRequestApiResult, ConvertPurchaseRequestApiRequest>(
      `/purchase-requests/${encodeURIComponent(id)}/convert-to-po`,
      toApiConvertInput(input),
      { accessToken: defaultAccessToken }
    );

    return {
      purchaseRequest: fromApiPurchaseRequest(result.purchase_request),
      purchaseOrder: fromApiPurchaseOrder(result.purchase_order),
      auditLogId: result.audit_log_id
    };
  } catch (reason) {
    if (!shouldUsePrototypeFallback(reason)) {
      throw reason;
    }

    const request = await getPurchaseRequest(id);
    if (request.status !== "approved") {
      throw new Error("Purchase request must be approved before PO conversion");
    }
    const purchaseOrder = await createPurchaseOrder(buildPurchaseOrderFromPurchaseRequest(request, input));
    const converted = applyLocalPurchaseRequestStatus(request, "converted_to_po", {
      convertedAt: new Date().toISOString(),
      convertedBy: "local-user",
      convertedPurchaseOrderId: purchaseOrder.id,
      convertedPurchaseOrderNo: purchaseOrder.poNo
    });

    return { purchaseRequest: converted, purchaseOrder };
  }
}

export function purchaseRequestStatusLabel(status: PurchaseRequestStatus) {
  switch (status) {
    case "draft":
      return "Nháp";
    case "submitted":
      return "Chờ duyệt";
    case "approved":
      return "Đã duyệt";
    case "converted_to_po":
      return "Đã chuyển PO";
    case "cancelled":
      return "Đã hủy";
    case "rejected":
      return "Từ chối";
    default:
      return status;
  }
}

export function purchaseRequestStatusTone(status: PurchaseRequestStatus) {
  switch (status) {
    case "approved":
    case "converted_to_po":
      return "success" as const;
    case "submitted":
      return "info" as const;
    case "cancelled":
    case "rejected":
      return "danger" as const;
    case "draft":
    default:
      return "warning" as const;
  }
}

export function formatPurchaseRequestQuantity(value: string, uomCode: string) {
  return formatPurchaseQuantity(value, uomCode);
}

export function formatPurchaseRequestDate(value?: string) {
  return value ? formatPurchaseDate(value) : "-";
}

export function purchaseRequestDetailHref(request: Pick<PurchaseRequest, "id">) {
  return `/purchase/requests/${encodeURIComponent(request.id)}`;
}

function runPurchaseRequestAction(
  id: string,
  action: "submit" | "approve"
): Promise<PurchaseRequestActionResult> {
  return apiPost<PurchaseRequestActionApiResult, Record<string, never>>(
    `/purchase-requests/${encodeURIComponent(id)}/${action}`,
    {},
    { accessToken: defaultAccessToken }
  )
    .then((result) => ({
      purchaseRequest: fromApiPurchaseRequest(result.purchase_request),
      previousStatus: result.previous_status,
      currentStatus: result.current_status,
      auditLogId: result.audit_log_id
    }))
    .catch(async (reason) => {
      if (!shouldUsePrototypeFallback(reason)) {
        throw reason;
      }
      const request = await getPurchaseRequest(id);
      const nextStatus = action === "submit" ? "submitted" : "approved";
      if ((action === "submit" && request.status !== "draft") || (action === "approve" && request.status !== "submitted")) {
        throw new Error("Purchase request status transition is invalid");
      }
      const transitionPatch: Partial<PurchaseRequest> =
        nextStatus === "submitted"
          ? { submittedAt: new Date().toISOString(), submittedBy: "local-user" }
          : { approvedAt: new Date().toISOString(), approvedBy: "local-user" };
      const next = applyLocalPurchaseRequestStatus(request, nextStatus, transitionPatch);

      return {
        purchaseRequest: next,
        previousStatus: request.status,
        currentStatus: next.status
      };
    });
}

function fromApiPurchaseRequest(request: PurchaseRequestApi): PurchaseRequest {
  return withLocalOverride({
    id: request.id ?? "",
    requestNo: request.request_no ?? "",
    sourceProductionPlanId: request.source_production_plan_id ?? "",
    sourceProductionPlanNo: request.source_production_plan_no ?? "",
    status: request.status ?? "draft",
    lines: request.lines.map(fromApiPurchaseRequestLine),
    createdAt: request.created_at,
    createdBy: request.created_by,
    submittedAt: request.submitted_at,
    submittedBy: request.submitted_by,
    approvedAt: request.approved_at,
    approvedBy: request.approved_by,
    convertedAt: request.converted_at,
    convertedBy: request.converted_by,
    convertedPurchaseOrderId: request.converted_purchase_order_id,
    convertedPurchaseOrderNo: request.converted_purchase_order_no,
    cancelledAt: request.cancelled_at,
    cancelledBy: request.cancelled_by,
    rejectedAt: request.rejected_at,
    rejectedBy: request.rejected_by,
    rejectReason: request.reject_reason
  });
}

function fromApiPurchaseRequestLine(line: PurchaseRequestLineApi): PurchaseRequestLine {
  return {
    id: line.id,
    lineNo: line.line_no,
    sourceProductionPlanLineId: line.source_production_plan_line_id,
    itemId: line.item_id,
    sku: line.sku,
    itemName: line.item_name,
    requestedQty: line.requested_qty,
    uomCode: line.uom_code,
    note: line.note
  };
}

function fromApiPurchaseOrder(order: PurchaseOrderApi): PurchaseOrder {
  return {
    id: order.id,
    poNo: order.po_no,
    supplierId: order.supplier_id,
    supplierCode: order.supplier_code,
    supplierName: order.supplier_name,
    warehouseId: order.warehouse_id,
    warehouseCode: order.warehouse_code,
    expectedDate: order.expected_date,
    status: order.status,
    currencyCode: order.currency_code,
    subtotalAmount: order.subtotal_amount,
    totalAmount: order.total_amount,
    note: order.note,
    lineCount: order.lines.length,
    receivedLineCount: order.lines.filter((line) => line.received_qty !== "0.000000").length,
    lines: order.lines.map((line) => ({
      id: line.id,
      lineNo: line.line_no,
      itemId: line.item_id,
      skuCode: line.sku_code,
      itemName: line.item_name,
      orderedQty: line.ordered_qty,
      receivedQty: line.received_qty,
      uomCode: line.uom_code,
      baseOrderedQty: line.base_ordered_qty,
      baseReceivedQty: line.base_received_qty,
      baseUomCode: line.base_uom_code,
      conversionFactor: line.conversion_factor,
      unitPrice: line.unit_price,
      currencyCode: line.currency_code,
      lineAmount: line.line_amount,
      expectedDate: line.expected_date,
      note: line.note
    })),
    auditLogId: order.audit_log_id,
    createdAt: order.created_at,
    updatedAt: order.updated_at,
    submittedAt: order.submitted_at,
    approvedAt: order.approved_at,
    closedAt: order.closed_at,
    cancelledAt: order.cancelled_at,
    rejectedAt: order.rejected_at,
    cancelReason: order.cancel_reason,
    rejectReason: order.reject_reason,
    version: order.version
  };
}

function purchaseRequestsFromProductionPlan(plan: ProductionPlan): PurchaseRequest[] {
  if (plan.purchaseRequestDraft.lines.length === 0 || !plan.purchaseRequestDraft.id || !plan.purchaseRequestDraft.requestNo) {
    return [];
  }
  return [fromProductionPlanPurchaseRequestDraft(plan, plan.purchaseRequestDraft)];
}

function fromProductionPlanPurchaseRequestDraft(plan: ProductionPlan, draft: PurchaseRequestDraft): PurchaseRequest {
  return withLocalOverride({
    id: draft.id ?? "",
    requestNo: draft.requestNo ?? "",
    sourceProductionPlanId: draft.sourceProductionPlanId ?? plan.id,
    sourceProductionPlanNo: draft.sourceProductionPlanNo ?? plan.planNo,
    status: draft.status ?? "draft",
    lines: draft.lines.map((line) => ({ ...line })),
    createdAt: draft.createdAt ?? plan.createdAt,
    createdBy: draft.createdBy ?? plan.createdBy,
    submittedAt: draft.submittedAt,
    submittedBy: draft.submittedBy,
    approvedAt: draft.approvedAt,
    approvedBy: draft.approvedBy,
    convertedAt: draft.convertedAt,
    convertedBy: draft.convertedBy,
    convertedPurchaseOrderId: draft.convertedPurchaseOrderId,
    convertedPurchaseOrderNo: draft.convertedPurchaseOrderNo,
    cancelledAt: draft.cancelledAt,
    cancelledBy: draft.cancelledBy,
    rejectedAt: draft.rejectedAt,
    rejectedBy: draft.rejectedBy,
    rejectReason: draft.rejectReason
  });
}

function buildPurchaseOrderFromPurchaseRequest(
  request: PurchaseRequest,
  input: ConvertPurchaseRequestToPurchaseOrderInput
): CreatePurchaseOrderInput {
  return {
    id: purchaseOrderIdForPurchaseRequest(request),
    poNo: purchaseOrderNoForPurchaseRequest(request),
    supplierId: input.supplierId,
    warehouseId: input.warehouseId,
    expectedDate: input.expectedDate,
    currencyCode: input.currencyCode ?? "VND",
    note: `Tạo từ đề nghị mua ${request.requestNo} / kế hoạch sản xuất ${request.sourceProductionPlanNo}`,
    lines: request.lines.map((line) => {
      if (!line.itemId) {
        throw new Error(`Purchase request line ${line.sku} has no item id`);
      }

      return {
        id: `${purchaseOrderIdForPurchaseRequest(request)}-line-${String(line.lineNo).padStart(2, "0")}`,
        lineNo: line.lineNo,
        itemId: line.itemId,
        orderedQty: line.requestedQty,
        uomCode: line.uomCode,
        unitPrice: input.unitPrice ?? "0",
        currencyCode: input.currencyCode ?? "VND",
        expectedDate: input.expectedDate,
        note: `Từ ${request.requestNo} dòng ${line.lineNo}: ${line.sku}`
      };
    })
  };
}

function filterPurchaseRequests(requests: PurchaseRequest[], query: PurchaseRequestQuery) {
  const search = query.search?.trim().toLowerCase() ?? "";
  return requests.filter((request) => {
    if (query.status && request.status !== query.status) {
      return false;
    }
    if (query.sourceProductionPlanId && request.sourceProductionPlanId !== query.sourceProductionPlanId) {
      return false;
    }
    if (!search) {
      return true;
    }

    return [request.requestNo, request.sourceProductionPlanNo, ...request.lines.flatMap((line) => [line.sku, line.itemName])]
      .join(" ")
      .toLowerCase()
      .includes(search);
  });
}

function applyLocalPurchaseRequestStatus(
  request: PurchaseRequest,
  status: PurchaseRequestStatus,
  patch: Partial<PurchaseRequest>
) {
  const next = { ...request, ...patch, status };
  localStatusOverrides.set(request.id, { ...localStatusOverrides.get(request.id), ...patch, status });

  return next;
}

function withLocalOverride(request: PurchaseRequest) {
  return { ...request, ...localStatusOverrides.get(request.id), lines: request.lines.map((line) => ({ ...line })) };
}

function purchaseRequestQueryString(query: PurchaseRequestQuery) {
  const params = new URLSearchParams();
  if (query.search) {
    params.set("q", query.search);
  }
  if (query.status) {
    params.set("status", query.status);
  }
  if (query.sourceProductionPlanId) {
    params.set("source_production_plan_id", query.sourceProductionPlanId);
  }
  const value = params.toString();
  return value ? `?${value}` : "";
}

function toApiConvertInput(input: ConvertPurchaseRequestToPurchaseOrderInput): ConvertPurchaseRequestApiRequest {
  return {
    supplier_id: input.supplierId,
    warehouse_id: input.warehouseId,
    expected_date: input.expectedDate,
    currency_code: input.currencyCode,
    unit_price: input.unitPrice
  };
}

function purchaseOrderIdForPurchaseRequest(request: PurchaseRequest) {
  return `po-${purchaseRequestSuffix(request)}`;
}

function purchaseOrderNoForPurchaseRequest(request: PurchaseRequest) {
  return `PO-${purchaseRequestSuffix(request).toUpperCase()}`;
}

function purchaseRequestSuffix(request: PurchaseRequest) {
  return (request.id || request.requestNo).toLowerCase().replace(/^pr-draft-/, "").replace(/^pr-/, "");
}
