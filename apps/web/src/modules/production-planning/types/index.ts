import type { FormulaComponentType, ProductType } from "../../masterdata/types";

export type ProductionPlanStatus = "draft" | "purchase_request_draft_created" | "cancelled";
export type PurchaseRequestDraftStatus =
  | "draft"
  | "submitted"
  | "approved"
  | "converted_to_po"
  | "cancelled"
  | "rejected";

export type ProductionPlanIssueStatus =
  | "shortage"
  | "ready_to_issue"
  | "issue_draft"
  | "issue_submitted"
  | "issue_approved"
  | "partially_issued"
  | "issued"
  | "waived"
  | "blocked";

export type ProductionPlanWarehouseIssueRef = {
  id: string;
  issueNo: string;
  lineId: string;
  status: string;
  quantity: string;
};

export type ProductionPlanLine = {
  id: string;
  formulaLineId: string;
  lineNo: number;
  componentItemId?: string;
  componentSku: string;
  componentName: string;
  componentType: FormulaComponentType;
  formulaQty: string;
  formulaUomCode: string;
  requiredQty: string;
  requiredUomCode: string;
  requiredStockBaseQty: string;
  stockBaseUomCode: string;
  availableQty: string;
  shortageQty: string;
  purchaseDraftQty: string;
  purchaseDraftUomCode: string;
  issuedQty: string;
  remainingIssueQty: string;
  issueStatus: ProductionPlanIssueStatus;
  warehouseIssues: ProductionPlanWarehouseIssueRef[];
  isStockManaged: boolean;
  needsPurchase: boolean;
  note?: string;
};

export type PurchaseRequestDraftLine = {
  id: string;
  lineNo: number;
  sourceProductionPlanLineId: string;
  itemId?: string;
  sku: string;
  itemName: string;
  requestedQty: string;
  uomCode: string;
  note?: string;
};

export type PurchaseRequestDraft = {
  id?: string;
  requestNo?: string;
  sourceProductionPlanId?: string;
  sourceProductionPlanNo?: string;
  status?: PurchaseRequestDraftStatus;
  lines: PurchaseRequestDraftLine[];
  createdAt?: string;
  createdBy?: string;
  submittedAt?: string;
  submittedBy?: string;
  approvedAt?: string;
  approvedBy?: string;
  convertedAt?: string;
  convertedBy?: string;
  convertedPurchaseOrderId?: string;
  convertedPurchaseOrderNo?: string;
  cancelledAt?: string;
  cancelledBy?: string;
  rejectedAt?: string;
  rejectedBy?: string;
  rejectReason?: string;
};

export type ProductionPlan = {
  id: string;
  orgId: string;
  planNo: string;
  outputItemId: string;
  outputSku: string;
  outputItemName: string;
  outputItemType: ProductType;
  plannedQty: string;
  uomCode: string;
  formulaId: string;
  formulaCode: string;
  formulaVersion: string;
  formulaBatchQty: string;
  formulaBatchUomCode: string;
  plannedStartDate?: string;
  plannedEndDate?: string;
  status: ProductionPlanStatus;
  lines: ProductionPlanLine[];
  purchaseRequestDraft: PurchaseRequestDraft;
  auditLogId?: string;
  createdAt: string;
  createdBy: string;
  updatedAt: string;
  updatedBy: string;
  version: number;
};

export type ProductionPlanInput = {
  outputItemId: string;
  formulaId?: string;
  plannedQty: string;
  uomCode: string;
  plannedStartDate?: string;
  plannedEndDate?: string;
};

export type ProductionPlanDraftLine = ProductionPlanInput & {
  rowId: string;
};

export type ProductionPlanQuery = {
  search?: string;
  status?: ProductionPlanStatus | "";
  outputItemId?: string;
};

export type ProductionPlanSummary = {
  total: number;
  draft: number;
  shortageLines: number;
  purchaseDraftLines: number;
};
