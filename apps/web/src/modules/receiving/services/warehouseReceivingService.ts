import { apiGet, apiPost } from "../../../shared/api/client";
import type { components, operations } from "../../../shared/api/generated/schema";
import { decimalScales, formatDateTimeVI, formatQuantity, normalizeDecimalInput } from "../../../shared/format/numberFormat";
import type {
  BatchQCStatus,
  CreateGoodsReceiptInput,
  CreateGoodsReceiptLineInput,
  GoodsReceipt,
  GoodsReceiptLine,
  GoodsReceiptQuery,
  GoodsReceiptStatus,
  GoodsReceiptStockMovement
} from "../types";

type GoodsReceiptApi = components["schemas"]["GoodsReceipt"];
type GoodsReceiptApiQuery = operations["listGoodsReceipts"]["parameters"]["query"];
type CreateGoodsReceiptApiRequest = components["schemas"]["CreateGoodsReceiptRequest"];

type WarehouseOption = {
  label: string;
  value: string;
  code: string;
};

type LocationOption = {
  label: string;
  value: string;
  code: string;
  warehouseId: string;
};

type BatchOption = {
  label: string;
  value: string;
  batchNo: string;
  itemId: string;
  sku: string;
  itemName: string;
  qcStatus: BatchQCStatus;
  baseUomCode: string;
};

const defaultAccessToken = "local-dev-access-token";
const prototypeNow = "2026-04-27T10:00:00Z";
let receiptSequence = 10;

export const receivingWarehouseOptions: WarehouseOption[] = [
  { label: "Finished Goods HCM", value: "wh-hcm-fg", code: "WH-HCM-FG" },
  { label: "Raw Material HCM", value: "wh-hcm-rm", code: "WH-HCM-RM" },
  { label: "QC Hold HCM", value: "wh-hcm-qh", code: "WH-HCM-QH" }
];

export const receivingLocationOptions: LocationOption[] = [
  { label: "FG-RECV-01", value: "loc-hcm-fg-recv-01", code: "FG-RECV-01", warehouseId: "wh-hcm-fg" },
  { label: "RM-RECV-01", value: "loc-hcm-rm-recv-01", code: "RM-RECV-01", warehouseId: "wh-hcm-rm" },
  { label: "QH-HOLD-01", value: "loc-hcm-qh-hold-01", code: "QH-HOLD-01", warehouseId: "wh-hcm-qh" }
];

export const receivingBatchOptions: BatchOption[] = [
  {
    label: "LOT-2604A / SERUM-30ML",
    value: "batch-serum-2604a",
    batchNo: "LOT-2604A",
    itemId: "item-serum-30ml",
    sku: "SERUM-30ML",
    itemName: "Vitamin C Serum",
    qcStatus: "hold",
    baseUomCode: "EA"
  },
  {
    label: "LOT-2603B / CREAM-50G",
    value: "batch-cream-2603b",
    batchNo: "LOT-2603B",
    itemId: "item-cream-50g",
    sku: "CREAM-50G",
    itemName: "Moisturizing Cream",
    qcStatus: "pass",
    baseUomCode: "EA"
  },
  {
    label: "LOT-2604C / TONER-100ML",
    value: "batch-toner-2604c",
    batchNo: "LOT-2604C",
    itemId: "item-toner-100ml",
    sku: "TONER-100ML",
    itemName: "Hydrating Toner",
    qcStatus: "fail",
    baseUomCode: "EA"
  }
];

let prototypeGoodsReceipts = createPrototypeGoodsReceipts();

export async function getGoodsReceipts(query: GoodsReceiptQuery = {}): Promise<GoodsReceipt[]> {
  try {
    const receipts = await apiGet("/goods-receipts", {
      accessToken: defaultAccessToken,
      query: toApiQuery(query)
    });

    return receipts.map(fromApiGoodsReceipt);
  } catch {
    return filterPrototypeReceipts(query);
  }
}

export async function createGoodsReceipt(input: CreateGoodsReceiptInput): Promise<GoodsReceipt> {
  try {
    const receipt = await apiPost<GoodsReceiptApi, CreateGoodsReceiptApiRequest>(
      "/goods-receipts",
      toApiCreateInput(input),
      { accessToken: defaultAccessToken }
    );

    return fromApiGoodsReceipt(receipt);
  } catch {
    return createPrototypeGoodsReceipt(input);
  }
}

export async function submitGoodsReceipt(id: string): Promise<GoodsReceipt> {
  return transitionGoodsReceipt(id, "submitted", `/goods-receipts/${encodeURIComponent(id)}/submit`);
}

export async function markGoodsReceiptInspectReady(id: string): Promise<GoodsReceipt> {
  return transitionGoodsReceipt(id, "inspect_ready", `/goods-receipts/${encodeURIComponent(id)}/inspect-ready`);
}

export async function postGoodsReceipt(id: string): Promise<GoodsReceipt> {
  return transitionGoodsReceipt(id, "posted", `/goods-receipts/${encodeURIComponent(id)}/post`);
}

export function resetPrototypeGoodsReceiptsForTest() {
  receiptSequence = 10;
  prototypeGoodsReceipts = createPrototypeGoodsReceipts();
}

export function goodsReceiptStatusTone(status: GoodsReceiptStatus): "success" | "warning" | "danger" | "info" {
  switch (status) {
    case "posted":
      return "success";
    case "inspect_ready":
      return "warning";
    case "submitted":
      return "info";
    case "draft":
    default:
      return "warning";
  }
}

export function qcStatusTone(status?: BatchQCStatus): "success" | "warning" | "danger" | "info" | "normal" {
  switch (status) {
    case "pass":
      return "success";
    case "fail":
      return "danger";
    case "quarantine":
    case "retest_required":
      return "danger";
    case "hold":
      return "warning";
    default:
      return "normal";
  }
}

export function formatGoodsReceiptStatus(status: GoodsReceiptStatus) {
  switch (status) {
    case "inspect_ready":
      return "Inspect ready";
    case "submitted":
      return "Submitted";
    case "posted":
      return "Posted";
    case "draft":
    default:
      return "Draft";
  }
}

export function formatQCStatus(status?: BatchQCStatus) {
  switch (status) {
    case "pass":
      return "Pass";
    case "fail":
      return "Fail";
    case "quarantine":
      return "Quarantine";
    case "retest_required":
      return "Retest";
    case "hold":
      return "Hold";
    default:
      return "-";
  }
}

export function formatReceivingQuantity(value: string, uomCode?: string) {
  return formatQuantity(value, uomCode);
}

export function formatReceivingDateTime(value?: string) {
  return value ? formatDateTimeVI(value) : "-";
}

async function transitionGoodsReceipt(id: string, status: GoodsReceiptStatus, path: string): Promise<GoodsReceipt> {
  try {
    const receipt = await apiPost<GoodsReceiptApi, Record<string, never>>(path, {}, { accessToken: defaultAccessToken });

    return fromApiGoodsReceipt(receipt);
  } catch {
    return transitionPrototypeReceipt(id, status);
  }
}

function fromApiGoodsReceipt(receipt: GoodsReceiptApi): GoodsReceipt {
  return {
    id: receipt.id,
    orgId: receipt.org_id,
    receiptNo: receipt.receipt_no,
    warehouseId: receipt.warehouse_id,
    warehouseCode: receipt.warehouse_code,
    locationId: receipt.location_id,
    locationCode: receipt.location_code,
    referenceDocType: receipt.reference_doc_type,
    referenceDocId: receipt.reference_doc_id,
    supplierId: receipt.supplier_id,
    status: receipt.status,
    lines: receipt.lines.map(fromApiGoodsReceiptLine),
    stockMovements: receipt.stock_movements?.map(fromApiGoodsReceiptStockMovement),
    createdBy: receipt.created_by,
    submittedBy: receipt.submitted_by,
    inspectReadyBy: receipt.inspect_ready_by,
    postedBy: receipt.posted_by,
    auditLogId: receipt.audit_log_id,
    createdAt: receipt.created_at,
    updatedAt: receipt.updated_at,
    submittedAt: receipt.submitted_at,
    inspectReadyAt: receipt.inspect_ready_at,
    postedAt: receipt.posted_at
  };
}

function fromApiGoodsReceiptLine(line: components["schemas"]["GoodsReceiptLine"]): GoodsReceiptLine {
  return {
    id: line.id,
    itemId: line.item_id,
    sku: line.sku,
    itemName: line.item_name,
    batchId: line.batch_id,
    batchNo: line.batch_no,
    warehouseId: line.warehouse_id,
    locationId: line.location_id,
    quantity: line.quantity,
    baseUomCode: line.base_uom_code,
    qcStatus: line.qc_status
  };
}

function fromApiGoodsReceiptStockMovement(
  movement: components["schemas"]["GoodsReceiptStockMovement"]
): GoodsReceiptStockMovement {
  return {
    movementNo: movement.movement_no,
    movementType: movement.movement_type,
    itemId: movement.item_id,
    batchId: movement.batch_id,
    warehouseId: movement.warehouse_id,
    locationId: movement.location_id,
    quantity: movement.quantity,
    baseUomCode: movement.base_uom_code,
    stockStatus: movement.stock_status,
    sourceDocId: movement.source_doc_id,
    sourceDocLineId: movement.source_doc_line_id
  };
}

function toApiQuery(query: GoodsReceiptQuery): GoodsReceiptApiQuery {
  return {
    warehouse_id: query.warehouseId,
    status: query.status
  };
}

function toApiCreateInput(input: CreateGoodsReceiptInput): CreateGoodsReceiptApiRequest {
  return {
    id: input.id,
    org_id: input.orgId,
    receipt_no: input.receiptNo,
    warehouse_id: input.warehouseId,
    location_id: input.locationId,
    reference_doc_type: input.referenceDocType,
    reference_doc_id: input.referenceDocId,
    supplier_id: input.supplierId,
    lines: input.lines.map((line) => ({
      id: line.id,
      item_id: line.itemId,
      sku: line.sku,
      item_name: line.itemName,
      batch_id: line.batchId,
      batch_no: line.batchNo,
      quantity: line.quantity,
      base_uom_code: line.baseUomCode,
      qc_status: line.qcStatus
    }))
  };
}

function filterPrototypeReceipts(query: GoodsReceiptQuery): GoodsReceipt[] {
  return prototypeGoodsReceipts
    .filter((receipt) => {
      if (query.warehouseId && receipt.warehouseId !== query.warehouseId) {
        return false;
      }
      if (query.status && receipt.status !== query.status) {
        return false;
      }

      return true;
    })
    .map(cloneReceipt)
    .sort((left, right) => right.updatedAt.localeCompare(left.updatedAt));
}

function createPrototypeGoodsReceipt(input: CreateGoodsReceiptInput): GoodsReceipt {
  const warehouse = receivingWarehouseOptions.find((option) => option.value === input.warehouseId);
  const location = receivingLocationOptions.find(
    (option) => option.value === input.locationId && option.warehouseId === input.warehouseId
  );
  if (!warehouse || !location || input.referenceDocType.trim() === "" || input.referenceDocId.trim() === "") {
    throw new Error("Warehouse, location, and reference document are required");
  }
  if (input.lines.length === 0) {
    throw new Error("At least one receiving line is required");
  }

  const id = input.id?.trim() || `grn-ui-${receiptSequence++}`;
  const receiptNo = input.receiptNo?.trim().toUpperCase() || `GRN-260427-UI-${receiptSequence}`;
  const lines = input.lines.map((line, index) => createPrototypeLine(line, index, warehouse.value, location.value));
  const receipt: GoodsReceipt = {
    id,
    orgId: input.orgId?.trim() || "org-my-pham",
    receiptNo,
    warehouseId: warehouse.value,
    warehouseCode: warehouse.code,
    locationId: location.value,
    locationCode: location.code,
    referenceDocType: input.referenceDocType.trim(),
    referenceDocId: input.referenceDocId.trim(),
    supplierId: input.supplierId?.trim() || undefined,
    status: "draft",
    lines,
    createdBy: "user-warehouse-lead",
    auditLogId: "audit-receiving-prototype",
    createdAt: prototypeNow,
    updatedAt: prototypeNow
  };
  prototypeGoodsReceipts = [receipt, ...prototypeGoodsReceipts.filter((candidate) => candidate.id !== receipt.id)];

  return cloneReceipt(receipt);
}

function createPrototypeLine(
  input: CreateGoodsReceiptLineInput,
  index: number,
  warehouseId: string,
  locationId: string
): GoodsReceiptLine {
  const batch = input.batchId ? receivingBatchOptions.find((option) => option.value === input.batchId) : undefined;
  const quantity = normalizeDecimalInput(input.quantity, decimalScales.quantity);
  if (Number(quantity) <= 0) {
    throw new Error("Quantity must be positive");
  }
  if (!batch && (!input.itemId || !input.sku)) {
    throw new Error("Item/SKU or batch is required");
  }

  return {
    id: input.id?.trim() || `line-${index + 1}`,
    itemId: input.itemId?.trim() || batch?.itemId || "",
    sku: input.sku?.trim().toUpperCase() || batch?.sku || "",
    itemName: input.itemName?.trim() || batch?.itemName,
    batchId: input.batchId?.trim() || undefined,
    batchNo: input.batchNo?.trim().toUpperCase() || batch?.batchNo,
    warehouseId,
    locationId,
    quantity,
    baseUomCode: (input.baseUomCode || batch?.baseUomCode || "EA").trim().toUpperCase(),
    qcStatus: input.qcStatus || batch?.qcStatus
  };
}

function transitionPrototypeReceipt(id: string, nextStatus: GoodsReceiptStatus): GoodsReceipt {
  const current = prototypeGoodsReceipts.find((receipt) => receipt.id === id);
  if (!current) {
    throw new Error("Goods receipt not found");
  }
  if (current.status === "posted") {
    throw new Error("Goods receipt is already posted");
  }

  const updated = cloneReceipt(current);
  updated.updatedAt = "2026-04-27T11:00:00Z";
  if (nextStatus === "submitted") {
    if (updated.status !== "draft") {
      throw new Error("Goods receipt must be draft before submit");
    }
    updated.status = "submitted";
    updated.submittedBy = "user-warehouse-lead";
    updated.submittedAt = updated.updatedAt;
  } else if (nextStatus === "inspect_ready") {
    if (updated.status !== "submitted") {
      throw new Error("Goods receipt must be submitted before inspection");
    }
    updated.status = "inspect_ready";
    updated.inspectReadyBy = "user-qa";
    updated.inspectReadyAt = updated.updatedAt;
  } else if (nextStatus === "posted") {
    if (updated.status !== "inspect_ready") {
      throw new Error("Goods receipt must be inspection-ready before posting");
    }
    if (updated.lines.some((line) => !line.batchId || !line.qcStatus)) {
      throw new Error("Batch and QC data are required before posting");
    }
    updated.status = "posted";
    updated.postedBy = "user-warehouse-lead";
    updated.postedAt = updated.updatedAt;
    updated.stockMovements = updated.lines.map((line, index) => ({
      movementNo: `${updated.receiptNo}-MV-${String(index + 1).padStart(3, "0")}`,
      movementType: "purchase_receipt",
      itemId: line.itemId,
      batchId: line.batchId ?? "",
      warehouseId: line.warehouseId,
      locationId: line.locationId,
      quantity: line.quantity,
      baseUomCode: line.baseUomCode,
      stockStatus: line.qcStatus === "pass" ? "available" : "qc_hold",
      sourceDocId: updated.id,
      sourceDocLineId: line.id
    }));
  }

  prototypeGoodsReceipts = prototypeGoodsReceipts.map((receipt) => (receipt.id === id ? updated : receipt));

  return cloneReceipt(updated);
}

function createPrototypeGoodsReceipts(): GoodsReceipt[] {
  return [
    {
      id: "grn-hcm-260427-draft",
      orgId: "org-my-pham",
      receiptNo: "GRN-260427-0001",
      warehouseId: "wh-hcm-fg",
      warehouseCode: "WH-HCM-FG",
      locationId: "loc-hcm-fg-recv-01",
      locationCode: "FG-RECV-01",
      referenceDocType: "purchase_order",
      referenceDocId: "PO-260427-0001",
      supplierId: "supplier-local",
      status: "draft",
      lines: [
        createPrototypeLine(
          { id: "grn-line-draft-001", batchId: "batch-serum-2604a", quantity: "24", baseUomCode: "EA" },
          0,
          "wh-hcm-fg",
          "loc-hcm-fg-recv-01"
        )
      ],
      createdBy: "user-warehouse-lead",
      createdAt: "2026-04-27T09:00:00Z",
      updatedAt: "2026-04-27T09:00:00Z"
    },
    {
      id: "grn-hcm-260427-inspect",
      orgId: "org-my-pham",
      receiptNo: "GRN-260427-0003",
      warehouseId: "wh-hcm-fg",
      warehouseCode: "WH-HCM-FG",
      locationId: "loc-hcm-fg-recv-01",
      locationCode: "FG-RECV-01",
      referenceDocType: "purchase_order",
      referenceDocId: "PO-260427-0003",
      supplierId: "supplier-local",
      status: "inspect_ready",
      lines: [
        createPrototypeLine(
          { id: "grn-line-inspect-001", batchId: "batch-cream-2603b", quantity: "12", baseUomCode: "EA" },
          0,
          "wh-hcm-fg",
          "loc-hcm-fg-recv-01"
        )
      ],
      createdBy: "user-warehouse-lead",
      submittedBy: "user-warehouse-lead",
      inspectReadyBy: "user-qa",
      createdAt: "2026-04-27T09:30:00Z",
      updatedAt: "2026-04-27T10:30:00Z",
      submittedAt: "2026-04-27T10:00:00Z",
      inspectReadyAt: "2026-04-27T10:30:00Z"
    }
  ];
}

function cloneReceipt(receipt: GoodsReceipt): GoodsReceipt {
  return {
    ...receipt,
    lines: receipt.lines.map((line) => ({ ...line })),
    stockMovements: receipt.stockMovements?.map((movement) => ({ ...movement }))
  };
}
