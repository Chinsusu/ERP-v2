import { apiGetRaw, apiPost } from "../../../shared/api/client";
import { shouldUsePrototypeFallback } from "../../../shared/api/prototypeFallback";
import type {
  CreateStockTransferInput,
  CreateWarehouseIssueInput,
  StockTransfer,
  StockTransferStatus,
  WarehouseDocumentAction,
  WarehouseIssue,
  WarehouseIssueStatus
} from "../types";

type StockTransferApiLine = {
  id: string;
  item_id?: string;
  sku: string;
  batch_id?: string;
  batch_no?: string;
  source_location_id?: string;
  source_location_code?: string;
  destination_location_id?: string;
  destination_location_code?: string;
  quantity: string;
  base_uom_code: string;
  note?: string;
};

type StockTransferApiItem = {
  id: string;
  transfer_no: string;
  org_id: string;
  source_warehouse_id: string;
  source_warehouse_code?: string;
  destination_warehouse_id: string;
  destination_warehouse_code?: string;
  reason_code: string;
  status: StockTransferStatus;
  requested_by: string;
  submitted_by?: string;
  approved_by?: string;
  posted_by?: string;
  lines: StockTransferApiLine[];
  audit_log_id?: string;
  created_at: string;
  updated_at: string;
  submitted_at?: string;
  approved_at?: string;
  posted_at?: string;
};

type CreateStockTransferApiRequest = {
  transfer_no?: string;
  source_warehouse_id: string;
  source_warehouse_code?: string;
  destination_warehouse_id: string;
  destination_warehouse_code?: string;
  reason_code: string;
  lines: Array<{
    id?: string;
    item_id?: string;
    sku: string;
    batch_id?: string;
    batch_no?: string;
    source_location_id?: string;
    source_location_code?: string;
    destination_location_id?: string;
    destination_location_code?: string;
    quantity: string;
    base_uom_code: string;
    note?: string;
  }>;
};

type WarehouseIssueApiLine = {
  id: string;
  item_id?: string;
  sku: string;
  item_name?: string;
  category?: string;
  batch_id?: string;
  batch_no?: string;
  location_id?: string;
  location_code?: string;
  quantity: string;
  base_uom_code: string;
  specification?: string;
  source_document_type?: string;
  source_document_id?: string;
  note?: string;
};

type WarehouseIssueApiItem = {
  id: string;
  issue_no: string;
  org_id: string;
  warehouse_id: string;
  warehouse_code?: string;
  destination_type: string;
  destination_name: string;
  reason_code: string;
  status: WarehouseIssueStatus;
  requested_by: string;
  submitted_by?: string;
  approved_by?: string;
  posted_by?: string;
  lines: WarehouseIssueApiLine[];
  audit_log_id?: string;
  created_at: string;
  updated_at: string;
  submitted_at?: string;
  approved_at?: string;
  posted_at?: string;
};

type CreateWarehouseIssueApiRequest = {
  issue_no?: string;
  warehouse_id: string;
  warehouse_code?: string;
  destination_type: string;
  destination_name: string;
  reason_code: string;
  lines: Array<{
    id?: string;
    item_id?: string;
    sku: string;
    item_name?: string;
    category?: string;
    batch_id?: string;
    batch_no?: string;
    location_id?: string;
    location_code?: string;
    quantity: string;
    base_uom_code: string;
    specification?: string;
    source_document_type?: string;
    source_document_id?: string;
    note?: string;
  }>;
};

const defaultAccessToken = "local-dev-access-token";
const nowSeed = "2026-05-05T08:00:00Z";

const initialPrototypeStockTransfers: StockTransfer[] = [];
const initialPrototypeWarehouseIssues: WarehouseIssue[] = [];

let prototypeStockTransfers = initialPrototypeStockTransfers.map(cloneStockTransfer);
let prototypeWarehouseIssues = initialPrototypeWarehouseIssues.map(cloneWarehouseIssue);

export async function getStockTransfers(): Promise<StockTransfer[]> {
  try {
    const items = await apiGetRaw<StockTransferApiItem[]>("/stock-transfers", {
      accessToken: defaultAccessToken
    });

    return items.map(fromApiStockTransfer);
  } catch (reason) {
    if (!shouldUsePrototypeFallback(reason)) {
      throw reason;
    }

    return prototypeStockTransfers.map(cloneStockTransfer);
  }
}

export async function createStockTransfer(input: CreateStockTransferInput): Promise<StockTransfer> {
  try {
    const item = await apiPost<StockTransferApiItem, CreateStockTransferApiRequest>(
      "/stock-transfers",
      toApiCreateStockTransfer(input),
      { accessToken: defaultAccessToken }
    );

    return fromApiStockTransfer(item);
  } catch (reason) {
    if (!shouldUsePrototypeFallback(reason)) {
      throw reason;
    }

    return createPrototypeStockTransfer(input);
  }
}

export async function transitionStockTransfer(id: string, action: WarehouseDocumentAction): Promise<StockTransfer> {
  try {
    const item = await apiPost<StockTransferApiItem, Record<string, never>>(
      `/stock-transfers/${encodeURIComponent(id)}/${action}`,
      {},
      { accessToken: defaultAccessToken }
    );

    return fromApiStockTransfer(item);
  } catch (reason) {
    if (!shouldUsePrototypeFallback(reason)) {
      throw reason;
    }

    return transitionPrototypeStockTransfer(id, action);
  }
}

export async function getWarehouseIssues(): Promise<WarehouseIssue[]> {
  try {
    const items = await apiGetRaw<WarehouseIssueApiItem[]>("/warehouse-issues", {
      accessToken: defaultAccessToken
    });

    return items.map(fromApiWarehouseIssue);
  } catch (reason) {
    if (!shouldUsePrototypeFallback(reason)) {
      throw reason;
    }

    return prototypeWarehouseIssues.map(cloneWarehouseIssue);
  }
}

export async function createWarehouseIssue(input: CreateWarehouseIssueInput): Promise<WarehouseIssue> {
  try {
    const item = await apiPost<WarehouseIssueApiItem, CreateWarehouseIssueApiRequest>(
      "/warehouse-issues",
      toApiCreateWarehouseIssue(input),
      { accessToken: defaultAccessToken }
    );

    return fromApiWarehouseIssue(item);
  } catch (reason) {
    if (!shouldUsePrototypeFallback(reason)) {
      throw reason;
    }

    return createPrototypeWarehouseIssue(input);
  }
}

export async function transitionWarehouseIssue(id: string, action: WarehouseDocumentAction): Promise<WarehouseIssue> {
  try {
    const item = await apiPost<WarehouseIssueApiItem, Record<string, never>>(
      `/warehouse-issues/${encodeURIComponent(id)}/${action}`,
      {},
      { accessToken: defaultAccessToken }
    );

    return fromApiWarehouseIssue(item);
  } catch (reason) {
    if (!shouldUsePrototypeFallback(reason)) {
      throw reason;
    }

    return transitionPrototypeWarehouseIssue(id, action);
  }
}

export function resetPrototypeWarehouseDocumentsForTest() {
  prototypeStockTransfers = initialPrototypeStockTransfers.map(cloneStockTransfer);
  prototypeWarehouseIssues = initialPrototypeWarehouseIssues.map(cloneWarehouseIssue);
}

function fromApiStockTransfer(item: StockTransferApiItem): StockTransfer {
  return {
    id: item.id,
    transferNo: item.transfer_no,
    orgId: item.org_id,
    sourceWarehouseId: item.source_warehouse_id,
    sourceWarehouseCode: item.source_warehouse_code,
    destinationWarehouseId: item.destination_warehouse_id,
    destinationWarehouseCode: item.destination_warehouse_code,
    reasonCode: item.reason_code,
    status: item.status,
    requestedBy: item.requested_by,
    submittedBy: item.submitted_by,
    approvedBy: item.approved_by,
    postedBy: item.posted_by,
    lines: item.lines.map((line) => ({
      id: line.id,
      itemId: line.item_id,
      sku: line.sku,
      batchId: line.batch_id,
      batchNo: line.batch_no,
      sourceLocationId: line.source_location_id,
      sourceLocationCode: line.source_location_code,
      destinationLocationId: line.destination_location_id,
      destinationLocationCode: line.destination_location_code,
      quantity: line.quantity,
      baseUomCode: line.base_uom_code,
      note: line.note
    })),
    auditLogId: item.audit_log_id,
    createdAt: item.created_at,
    updatedAt: item.updated_at,
    submittedAt: item.submitted_at,
    approvedAt: item.approved_at,
    postedAt: item.posted_at
  };
}

function fromApiWarehouseIssue(item: WarehouseIssueApiItem): WarehouseIssue {
  return {
    id: item.id,
    issueNo: item.issue_no,
    orgId: item.org_id,
    warehouseId: item.warehouse_id,
    warehouseCode: item.warehouse_code,
    destinationType: item.destination_type,
    destinationName: item.destination_name,
    reasonCode: item.reason_code,
    status: item.status,
    requestedBy: item.requested_by,
    submittedBy: item.submitted_by,
    approvedBy: item.approved_by,
    postedBy: item.posted_by,
    lines: item.lines.map((line) => ({
      id: line.id,
      itemId: line.item_id,
      sku: line.sku,
      itemName: line.item_name,
      category: line.category,
      batchId: line.batch_id,
      batchNo: line.batch_no,
      locationId: line.location_id,
      locationCode: line.location_code,
      quantity: line.quantity,
      baseUomCode: line.base_uom_code,
      specification: line.specification,
      sourceDocumentType: line.source_document_type,
      sourceDocumentId: line.source_document_id,
      note: line.note
    })),
    auditLogId: item.audit_log_id,
    createdAt: item.created_at,
    updatedAt: item.updated_at,
    submittedAt: item.submitted_at,
    approvedAt: item.approved_at,
    postedAt: item.posted_at
  };
}

function toApiCreateStockTransfer(input: CreateStockTransferInput): CreateStockTransferApiRequest {
  return {
    transfer_no: input.transferNo,
    source_warehouse_id: input.sourceWarehouseId,
    source_warehouse_code: input.sourceWarehouseCode,
    destination_warehouse_id: input.destinationWarehouseId,
    destination_warehouse_code: input.destinationWarehouseCode,
    reason_code: input.reasonCode,
    lines: input.lines.map((line) => ({
      id: line.id,
      item_id: line.itemId,
      sku: line.sku,
      batch_id: line.batchId,
      batch_no: line.batchNo,
      source_location_id: line.sourceLocationId,
      source_location_code: line.sourceLocationCode,
      destination_location_id: line.destinationLocationId,
      destination_location_code: line.destinationLocationCode,
      quantity: line.quantity,
      base_uom_code: line.baseUomCode,
      note: line.note
    }))
  };
}

function toApiCreateWarehouseIssue(input: CreateWarehouseIssueInput): CreateWarehouseIssueApiRequest {
  return {
    issue_no: input.issueNo,
    warehouse_id: input.warehouseId,
    warehouse_code: input.warehouseCode,
    destination_type: input.destinationType,
    destination_name: input.destinationName,
    reason_code: input.reasonCode,
    lines: input.lines.map((line) => ({
      id: line.id,
      item_id: line.itemId,
      sku: line.sku,
      item_name: line.itemName,
      category: line.category,
      batch_id: line.batchId,
      batch_no: line.batchNo,
      location_id: line.locationId,
      location_code: line.locationCode,
      quantity: line.quantity,
      base_uom_code: line.baseUomCode,
      specification: line.specification,
      source_document_type: line.sourceDocumentType,
      source_document_id: line.sourceDocumentId,
      note: line.note
    }))
  };
}

function createPrototypeStockTransfer(input: CreateStockTransferInput): StockTransfer {
  const now = new Date().toISOString();
  const created: StockTransfer = {
    id: `stock-transfer-${Date.now()}`,
    transferNo: input.transferNo || `ST-${prototypeStockTransfers.length + 1}`,
    orgId: "org-my-pham",
    sourceWarehouseId: input.sourceWarehouseId,
    sourceWarehouseCode: input.sourceWarehouseCode,
    destinationWarehouseId: input.destinationWarehouseId,
    destinationWarehouseCode: input.destinationWarehouseCode,
    reasonCode: input.reasonCode,
    status: "draft",
    requestedBy: "local-dev",
    lines: input.lines.map((line, index) => ({
      id: line.id || `stock-transfer-line-${index + 1}`,
      itemId: line.itemId,
      sku: line.sku,
      batchId: line.batchId,
      batchNo: line.batchNo,
      sourceLocationId: line.sourceLocationId,
      sourceLocationCode: line.sourceLocationCode,
      destinationLocationId: line.destinationLocationId,
      destinationLocationCode: line.destinationLocationCode,
      quantity: line.quantity,
      baseUomCode: line.baseUomCode,
      note: line.note
    })),
    auditLogId: `audit-stock-transfer-create-${Date.now()}`,
    createdAt: now,
    updatedAt: now
  };
  prototypeStockTransfers = [cloneStockTransfer(created), ...prototypeStockTransfers];

  return cloneStockTransfer(created);
}

function createPrototypeWarehouseIssue(input: CreateWarehouseIssueInput): WarehouseIssue {
  const now = new Date().toISOString();
  const created: WarehouseIssue = {
    id: `warehouse-issue-${Date.now()}`,
    issueNo: input.issueNo || `WI-${prototypeWarehouseIssues.length + 1}`,
    orgId: "org-my-pham",
    warehouseId: input.warehouseId,
    warehouseCode: input.warehouseCode,
    destinationType: input.destinationType,
    destinationName: input.destinationName,
    reasonCode: input.reasonCode,
    status: "draft",
    requestedBy: "local-dev",
    lines: input.lines.map((line, index) => ({
      id: line.id || `warehouse-issue-line-${index + 1}`,
      itemId: line.itemId,
      sku: line.sku,
      itemName: line.itemName,
      category: line.category,
      batchId: line.batchId,
      batchNo: line.batchNo,
      locationId: line.locationId,
      locationCode: line.locationCode,
      quantity: line.quantity,
      baseUomCode: line.baseUomCode,
      specification: line.specification,
      sourceDocumentType: line.sourceDocumentType,
      sourceDocumentId: line.sourceDocumentId,
      note: line.note
    })),
    auditLogId: `audit-warehouse-issue-create-${Date.now()}`,
    createdAt: now,
    updatedAt: now
  };
  prototypeWarehouseIssues = [cloneWarehouseIssue(created), ...prototypeWarehouseIssues];

  return cloneWarehouseIssue(created);
}

function transitionPrototypeStockTransfer(id: string, action: WarehouseDocumentAction): StockTransfer {
  const index = prototypeStockTransfers.findIndex((transfer) => transfer.id === id);
  if (index < 0) {
    throw new Error("Stock transfer not found");
  }
  const updated = nextTransferDocument(prototypeStockTransfers[index], action);
  prototypeStockTransfers = prototypeStockTransfers.map((transfer) => (transfer.id === id ? updated : transfer));

  return cloneStockTransfer(updated);
}

function transitionPrototypeWarehouseIssue(id: string, action: WarehouseDocumentAction): WarehouseIssue {
  const index = prototypeWarehouseIssues.findIndex((issue) => issue.id === id);
  if (index < 0) {
    throw new Error("Warehouse issue not found");
  }
  const updated = nextIssueDocument(prototypeWarehouseIssues[index], action);
  prototypeWarehouseIssues = prototypeWarehouseIssues.map((issue) => (issue.id === id ? updated : issue));

  return cloneWarehouseIssue(updated);
}

function nextTransferDocument(document: StockTransfer, action: WarehouseDocumentAction): StockTransfer {
  const now = new Date().toISOString();
  const nextStatus = nextDocumentStatus(document.status, action);

  return {
    ...document,
    status: nextStatus,
    auditLogId: `audit-stock-transfer-${action}-${Date.now()}`,
    updatedAt: now,
    submittedBy: action === "submit" ? "local-dev" : document.submittedBy,
    approvedBy: action === "approve" ? "local-dev" : document.approvedBy,
    postedBy: action === "post" ? "local-dev" : document.postedBy,
    submittedAt: action === "submit" ? now : document.submittedAt,
    approvedAt: action === "approve" ? now : document.approvedAt,
    postedAt: action === "post" ? now : document.postedAt
  };
}

function nextIssueDocument(document: WarehouseIssue, action: WarehouseDocumentAction): WarehouseIssue {
  const now = new Date().toISOString();
  const nextStatus = nextDocumentStatus(document.status, action);

  return {
    ...document,
    status: nextStatus,
    auditLogId: `audit-warehouse-issue-${action}-${Date.now()}`,
    updatedAt: now,
    submittedBy: action === "submit" ? "local-dev" : document.submittedBy,
    approvedBy: action === "approve" ? "local-dev" : document.approvedBy,
    postedBy: action === "post" ? "local-dev" : document.postedBy,
    submittedAt: action === "submit" ? now : document.submittedAt,
    approvedAt: action === "approve" ? now : document.approvedAt,
    postedAt: action === "post" ? now : document.postedAt
  };
}

function nextDocumentStatus<TStatus extends StockTransferStatus | WarehouseIssueStatus>(
  status: TStatus,
  action: WarehouseDocumentAction
): TStatus {
  if (action === "submit" && status === "draft") {
    return "submitted" as TStatus;
  }
  if (action === "approve" && status === "submitted") {
    return "approved" as TStatus;
  }
  if (action === "post" && status === "approved") {
    return "posted" as TStatus;
  }

  throw new Error("Warehouse document action is not allowed");
}

function cloneStockTransfer(transfer: StockTransfer): StockTransfer {
  return {
    ...transfer,
    createdAt: transfer.createdAt || nowSeed,
    updatedAt: transfer.updatedAt || nowSeed,
    lines: transfer.lines.map((line) => ({ ...line }))
  };
}

function cloneWarehouseIssue(issue: WarehouseIssue): WarehouseIssue {
  return {
    ...issue,
    createdAt: issue.createdAt || nowSeed,
    updatedAt: issue.updatedAt || nowSeed,
    lines: issue.lines.map((line) => ({ ...line }))
  };
}
