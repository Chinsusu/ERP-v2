import { apiGetRaw, apiPost } from "../../../shared/api/client";
import { shouldUsePrototypeFallback } from "../../../shared/api/prototypeFallback";
import { decimalScales, normalizeDecimalInput, normalizeUOMCode } from "../../../shared/format/numberFormat";
import { formatFormulaQuantity } from "../../masterdata/services/formulaMasterDataService";
import type {
  ProductionPlan,
  ProductionPlanInput,
  ProductionPlanLine,
  ProductionPlanQuery,
  ProductionPlanStatus,
  ProductionPlanSummary,
  PurchaseRequestDraft,
  PurchaseRequestDraftLine
} from "../types";

type ProductionPlanApi = {
  id: string;
  org_id: string;
  plan_no: string;
  output_item_id: string;
  output_sku: string;
  output_item_name: string;
  output_item_type: ProductionPlan["outputItemType"];
  planned_qty: string;
  uom_code: string;
  formula_id: string;
  formula_code: string;
  formula_version: string;
  formula_batch_qty: string;
  formula_batch_uom_code: string;
  planned_start_date?: string;
  planned_end_date?: string;
  status: ProductionPlanStatus;
  lines: ProductionPlanLineApi[];
  purchase_request_draft: PurchaseRequestDraftApi;
  audit_log_id?: string;
  created_at: string;
  created_by: string;
  updated_at: string;
  updated_by: string;
  version: number;
};

type ProductionPlanLineApi = {
  id: string;
  formula_line_id: string;
  line_no: number;
  component_item_id?: string;
  component_sku: string;
  component_name: string;
  component_type: ProductionPlanLine["componentType"];
  formula_qty: string;
  formula_uom_code: string;
  required_qty: string;
  required_uom_code: string;
  required_stock_base_qty: string;
  stock_base_uom_code: string;
  available_qty: string;
  shortage_qty: string;
  purchase_draft_qty: string;
  purchase_draft_uom_code: string;
  is_stock_managed: boolean;
  needs_purchase: boolean;
  note?: string;
};

type PurchaseRequestDraftApi = {
  id?: string;
  request_no?: string;
  source_production_plan_id?: string;
  source_production_plan_no?: string;
  status?: PurchaseRequestDraft["status"];
  lines: PurchaseRequestDraftLineApi[];
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

type PurchaseRequestDraftLineApi = {
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

type CreateProductionPlanApiRequest = {
  output_item_id: string;
  formula_id?: string;
  planned_qty: string;
  uom_code: string;
  planned_start_date?: string;
  planned_end_date?: string;
};

const defaultAccessToken = "local-dev-access-token";
const quantityScale = decimalScales.quantity;

let localPlans: ProductionPlan[] = [];

export async function getProductionPlans(query: ProductionPlanQuery = {}): Promise<ProductionPlan[]> {
  try {
    const plans = await apiGetRaw<ProductionPlanApi[]>(`/production-plans${productionPlanQueryString(query)}`, {
      accessToken: defaultAccessToken
    });

    return plans.map(fromApiProductionPlan);
  } catch (reason) {
    if (!shouldUsePrototypeFallback(reason)) {
      throw reason;
    }

    return filterProductionPlans(localPlans, query).map(cloneProductionPlan);
  }
}

export async function getProductionPlan(id: string): Promise<ProductionPlan> {
  try {
    const plan = await apiGetRaw<ProductionPlanApi>(`/production-plans/${encodeURIComponent(id)}`, {
      accessToken: defaultAccessToken
    });

    return fromApiProductionPlan(plan);
  } catch (reason) {
    if (!shouldUsePrototypeFallback(reason)) {
      throw reason;
    }

    const localPlan = localPlans.find((plan) => plan.id === id);
    if (!localPlan) {
      throw new Error("Production plan was not found");
    }

    return cloneProductionPlan(localPlan);
  }
}

export async function createProductionPlan(input: ProductionPlanInput): Promise<ProductionPlan> {
  const normalized = normalizeProductionPlanInput(input);
  validateProductionPlanInput(normalized);

  return createNormalizedProductionPlan(normalized);
}

export async function createProductionPlans(inputs: ProductionPlanInput[]): Promise<ProductionPlan[]> {
  if (inputs.length === 0) {
    throw new Error("At least one production plan line is required");
  }

  const normalized = inputs.map(normalizeProductionPlanInput);
  normalized.forEach(validateProductionPlanInput);

  const plans: ProductionPlan[] = [];
  for (const input of normalized) {
    plans.push(await createNormalizedProductionPlan(input));
  }

  return plans;
}

async function createNormalizedProductionPlan(normalized: ProductionPlanInput): Promise<ProductionPlan> {
  try {
    const plan = await apiPost<ProductionPlanApi, CreateProductionPlanApiRequest>(
      "/production-plans",
      toApiProductionPlanRequest(normalized),
      { accessToken: defaultAccessToken }
    );

    return fromApiProductionPlan(plan);
  } catch (reason) {
    if (!shouldUsePrototypeFallback(reason)) {
      throw reason;
    }

    const now = new Date().toISOString();
    const localID = `local-production-plan-${Date.now()}-${localPlans.length + 1}`;
    const plan: ProductionPlan = {
      id: localID,
      orgId: "org-my-pham",
      planNo: `PP-LOCAL-${Date.now()}`,
      outputItemId: normalized.outputItemId,
      outputSku: normalized.outputItemId.toUpperCase(),
      outputItemName: normalized.outputItemId,
      outputItemType: "finished_good",
      plannedQty: normalized.plannedQty,
      uomCode: normalized.uomCode,
      formulaId: normalized.formulaId ?? "",
      formulaCode: normalized.formulaId ?? "",
      formulaVersion: "local",
      formulaBatchQty: "1.000000",
      formulaBatchUomCode: normalized.uomCode,
      plannedStartDate: normalized.plannedStartDate,
      plannedEndDate: normalized.plannedEndDate,
      status: "draft",
      lines: [],
      purchaseRequestDraft: { lines: [] },
      createdAt: now,
      createdBy: "local-user",
      updatedAt: now,
      updatedBy: "local-user",
      version: 1
    };
    localPlans = [plan, ...localPlans];

    return cloneProductionPlan(plan);
  }
}

export function normalizeProductionPlanInput(input: ProductionPlanInput): ProductionPlanInput {
  return {
    outputItemId: input.outputItemId.trim(),
    formulaId: input.formulaId?.trim() || undefined,
    plannedQty: normalizeDecimalInput(input.plannedQty, quantityScale),
    uomCode: normalizeUOMCode(input.uomCode),
    plannedStartDate: input.plannedStartDate?.trim() || undefined,
    plannedEndDate: input.plannedEndDate?.trim() || undefined
  };
}

export function summarizeProductionPlans(plans: ProductionPlan[]): ProductionPlanSummary {
  return {
    total: plans.length,
    draft: plans.filter((plan) => plan.status === "draft").length,
    shortageLines: plans.reduce((total, plan) => total + plan.lines.filter((line) => line.needsPurchase).length, 0),
    purchaseDraftLines: plans.reduce((total, plan) => total + plan.purchaseRequestDraft.lines.length, 0)
  };
}

export function formatProductionPlanQuantity(value: string, uomCode: string) {
  return formatFormulaQuantity(value, uomCode);
}

export function productionPlanStatusDisplay(status: ProductionPlanStatus) {
  switch (status) {
    case "draft":
      return "Nháp";
    case "purchase_request_draft_created":
      return "Đã tạo đề nghị mua";
    case "cancelled":
      return "Đã hủy";
    default:
      return status;
  }
}

export function productionPlanStatusTone(status: ProductionPlanStatus) {
  switch (status) {
    case "draft":
      return "info" as const;
    case "purchase_request_draft_created":
      return "warning" as const;
    case "cancelled":
      return "danger" as const;
    default:
      return "normal" as const;
  }
}

export function resetPrototypeProductionPlansForTest() {
  localPlans = [];
}

function validateProductionPlanInput(input: ProductionPlanInput) {
  if (!input.outputItemId || !input.plannedQty || !input.uomCode) {
    throw new Error("Production plan required fields are missing");
  }
  if (input.plannedQty === "0.000000") {
    throw new Error("Production plan quantity must be greater than zero");
  }
}

function toApiProductionPlanRequest(input: ProductionPlanInput): CreateProductionPlanApiRequest {
  return {
    output_item_id: input.outputItemId,
    formula_id: input.formulaId,
    planned_qty: input.plannedQty,
    uom_code: input.uomCode,
    planned_start_date: input.plannedStartDate,
    planned_end_date: input.plannedEndDate
  };
}

function fromApiProductionPlan(plan: ProductionPlanApi): ProductionPlan {
  return {
    id: plan.id,
    orgId: plan.org_id,
    planNo: plan.plan_no,
    outputItemId: plan.output_item_id,
    outputSku: plan.output_sku,
    outputItemName: plan.output_item_name,
    outputItemType: plan.output_item_type,
    plannedQty: plan.planned_qty,
    uomCode: plan.uom_code,
    formulaId: plan.formula_id,
    formulaCode: plan.formula_code,
    formulaVersion: plan.formula_version,
    formulaBatchQty: plan.formula_batch_qty,
    formulaBatchUomCode: plan.formula_batch_uom_code,
    plannedStartDate: plan.planned_start_date,
    plannedEndDate: plan.planned_end_date,
    status: plan.status,
    lines: plan.lines.map(fromApiProductionPlanLine),
    purchaseRequestDraft: fromApiPurchaseRequestDraft(plan.purchase_request_draft),
    auditLogId: plan.audit_log_id,
    createdAt: plan.created_at,
    createdBy: plan.created_by,
    updatedAt: plan.updated_at,
    updatedBy: plan.updated_by,
    version: plan.version
  };
}

function fromApiProductionPlanLine(line: ProductionPlanLineApi): ProductionPlanLine {
  return {
    id: line.id,
    formulaLineId: line.formula_line_id,
    lineNo: line.line_no,
    componentItemId: line.component_item_id,
    componentSku: line.component_sku,
    componentName: line.component_name,
    componentType: line.component_type,
    formulaQty: line.formula_qty,
    formulaUomCode: line.formula_uom_code,
    requiredQty: line.required_qty,
    requiredUomCode: line.required_uom_code,
    requiredStockBaseQty: line.required_stock_base_qty,
    stockBaseUomCode: line.stock_base_uom_code,
    availableQty: line.available_qty,
    shortageQty: line.shortage_qty,
    purchaseDraftQty: line.purchase_draft_qty,
    purchaseDraftUomCode: line.purchase_draft_uom_code,
    isStockManaged: line.is_stock_managed,
    needsPurchase: line.needs_purchase,
    note: line.note
  };
}

function fromApiPurchaseRequestDraft(draft: PurchaseRequestDraftApi): PurchaseRequestDraft {
  return {
    id: draft.id,
    requestNo: draft.request_no,
    sourceProductionPlanId: draft.source_production_plan_id,
    sourceProductionPlanNo: draft.source_production_plan_no,
    status: draft.status,
    lines: draft.lines.map(fromApiPurchaseRequestDraftLine),
    createdAt: draft.created_at,
    createdBy: draft.created_by,
    submittedAt: draft.submitted_at,
    submittedBy: draft.submitted_by,
    approvedAt: draft.approved_at,
    approvedBy: draft.approved_by,
    convertedAt: draft.converted_at,
    convertedBy: draft.converted_by,
    convertedPurchaseOrderId: draft.converted_purchase_order_id,
    convertedPurchaseOrderNo: draft.converted_purchase_order_no,
    cancelledAt: draft.cancelled_at,
    cancelledBy: draft.cancelled_by,
    rejectedAt: draft.rejected_at,
    rejectedBy: draft.rejected_by,
    rejectReason: draft.reject_reason
  };
}

function fromApiPurchaseRequestDraftLine(line: PurchaseRequestDraftLineApi): PurchaseRequestDraftLine {
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

function productionPlanQueryString(query: ProductionPlanQuery) {
  const params = new URLSearchParams();
  if (query.search) {
    params.set("q", query.search);
  }
  if (query.status) {
    params.set("status", query.status);
  }
  if (query.outputItemId) {
    params.set("output_item_id", query.outputItemId);
  }
  const value = params.toString();
  return value ? `?${value}` : "";
}

function filterProductionPlans(plans: ProductionPlan[], query: ProductionPlanQuery) {
  const search = query.search?.trim().toLowerCase() ?? "";
  return plans.filter((plan) => {
    if (query.status && plan.status !== query.status) {
      return false;
    }
    if (query.outputItemId && plan.outputItemId !== query.outputItemId) {
      return false;
    }
    if (!search) {
      return true;
    }

    return [plan.planNo, plan.outputSku, plan.outputItemName, plan.formulaCode].join(" ").toLowerCase().includes(search);
  });
}

function cloneProductionPlan(plan: ProductionPlan): ProductionPlan {
  return {
    ...plan,
    lines: plan.lines.map((line) => ({ ...line })),
    purchaseRequestDraft: {
      ...plan.purchaseRequestDraft,
      lines: plan.purchaseRequestDraft.lines.map((line) => ({ ...line }))
    }
  };
}
