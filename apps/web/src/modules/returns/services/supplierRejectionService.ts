import { ApiError, apiGetRaw, apiPost } from "../../../shared/api/client";
import { decimalScales, formatDateTimeVI, formatQuantity, normalizeDecimalInput } from "../../../shared/format/numberFormat";
import type {
  CreateSupplierRejectionAttachmentInput,
  CreateSupplierRejectionInput,
  CreateSupplierRejectionLineInput,
  SupplierRejection,
  SupplierRejectionActionResult,
  SupplierRejectionAttachment,
  SupplierRejectionLine,
  SupplierRejectionQuery,
  SupplierRejectionStatus
} from "../types";

type SupplierRejectionLineApi = {
  id: string;
  purchase_order_line_id?: string;
  goods_receipt_line_id: string;
  inbound_qc_inspection_id: string;
  item_id: string;
  sku: string;
  item_name?: string;
  batch_id: string;
  batch_no: string;
  lot_no: string;
  expiry_date: string;
  rejected_qty: string;
  uom_code: string;
  base_uom_code: string;
  reason: string;
};

type SupplierRejectionAttachmentApi = {
  id: string;
  line_id?: string;
  file_name: string;
  object_key: string;
  content_type?: string;
  uploaded_at?: string;
  uploaded_by?: string;
  source?: string;
};

type SupplierRejectionApi = {
  id: string;
  org_id: string;
  rejection_no: string;
  supplier_id: string;
  supplier_code?: string;
  supplier_name?: string;
  purchase_order_id?: string;
  purchase_order_no?: string;
  goods_receipt_id: string;
  goods_receipt_no?: string;
  inbound_qc_inspection_id: string;
  warehouse_id: string;
  warehouse_code?: string;
  status: SupplierRejectionStatus;
  reason: string;
  lines: SupplierRejectionLineApi[];
  attachments: SupplierRejectionAttachmentApi[];
  audit_log_id?: string;
  created_at: string;
  created_by: string;
  updated_at: string;
  updated_by: string;
  submitted_at?: string;
  submitted_by?: string;
  confirmed_at?: string;
  confirmed_by?: string;
};

type CreateSupplierRejectionLineApiRequest = {
  id?: string;
  purchase_order_line_id?: string;
  goods_receipt_line_id: string;
  inbound_qc_inspection_id: string;
  item_id: string;
  sku: string;
  item_name?: string;
  batch_id: string;
  batch_no: string;
  lot_no: string;
  expiry_date: string;
  rejected_qty: string;
  uom_code: string;
  base_uom_code: string;
  reason: string;
};

type CreateSupplierRejectionAttachmentApiRequest = {
  id?: string;
  line_id?: string;
  file_name: string;
  object_key: string;
  content_type?: string;
  source?: string;
};

type CreateSupplierRejectionApiRequest = {
  id: string;
  org_id?: string;
  rejection_no: string;
  supplier_id: string;
  supplier_code?: string;
  supplier_name?: string;
  purchase_order_id?: string;
  purchase_order_no?: string;
  goods_receipt_id: string;
  goods_receipt_no?: string;
  inbound_qc_inspection_id: string;
  warehouse_id: string;
  warehouse_code?: string;
  reason: string;
  lines: CreateSupplierRejectionLineApiRequest[];
  attachments?: CreateSupplierRejectionAttachmentApiRequest[];
};

type SupplierRejectionActionResultApi = {
  rejection: SupplierRejectionApi;
  previous_status?: SupplierRejectionStatus;
  current_status: SupplierRejectionStatus;
  audit_log_id?: string;
};

type SupplierOption = {
  label: string;
  value: string;
  code: string;
  name: string;
};

type SupplierRejectionSampleLine = CreateSupplierRejectionLineInput & {
  label: string;
  purchaseOrderId: string;
  purchaseOrderNo: string;
  goodsReceiptId: string;
  goodsReceiptNo: string;
  warehouseId: string;
  warehouseCode: string;
};

const defaultAccessToken = "local-dev-access-token";
const prototypeNow = "2026-04-29T10:30:00Z";

let supplierRejectionSequence = 1;

export const supplierRejectionStatusOptions: { label: string; value: "" | SupplierRejectionStatus }[] = [
  { label: "All statuses", value: "" },
  { label: "Draft", value: "draft" },
  { label: "Submitted", value: "submitted" },
  { label: "Confirmed", value: "confirmed" },
  { label: "Cancelled", value: "cancelled" }
];

export const supplierRejectionSupplierOptions: SupplierOption[] = [
  { label: "SUP-LOCAL / Local Supplier", value: "supplier-local", code: "SUP-LOCAL", name: "Local Supplier" },
  { label: "SUP-BIOACTIVE / BioActive Raw Materials", value: "sup-rm-bioactive", code: "SUP-BIOACTIVE", name: "BioActive Raw Materials" },
  { label: "SUP-PACK-HCM / HCM Packaging Partner", value: "sup-pack-hcm", code: "SUP-PACK-HCM", name: "HCM Packaging Partner" }
];

export const supplierRejectionWarehouseOptions = [
  { label: "Finished Goods HCM", value: "wh-hcm-fg", code: "WH-HCM-FG" },
  { label: "Raw Material HCM", value: "wh-hcm-rm", code: "WH-HCM-RM" },
  { label: "QC Hold HCM", value: "wh-hcm-qh", code: "WH-HCM-QH" }
] as const;

export const supplierRejectionSampleLines: SupplierRejectionSampleLine[] = [
  {
    label: "GRN-260427-0003 / SERUM-30ML / LOT-2604A",
    purchaseOrderId: "po-260427-0003",
    purchaseOrderNo: "PO-260427-0003",
    goodsReceiptId: "grn-hcm-260427-inspect",
    goodsReceiptNo: "GRN-260427-0003",
    warehouseId: "wh-hcm-fg",
    warehouseCode: "WH-HCM-FG",
    purchaseOrderLineId: "po-line-260427-0003-001",
    goodsReceiptLineId: "grn-line-draft-001",
    inboundQCInspectionId: "iqc-fail-001",
    itemId: "item-serum-30ml",
    sku: "SERUM-30ML",
    itemName: "Vitamin C Serum",
    batchId: "batch-serum-2604a",
    batchNo: "LOT-2604A",
    lotNo: "LOT-2604A",
    expiryDate: "2027-04-01",
    rejectedQuantity: "6.000000",
    uomCode: "EA",
    baseUOMCode: "EA",
    reason: "Damaged packaging"
  },
  {
    label: "GRN-260427-0003 / TONER-100ML / LOT-2604C",
    purchaseOrderId: "po-260427-0004",
    purchaseOrderNo: "PO-260427-0004",
    goodsReceiptId: "grn-hcm-260427-fail",
    goodsReceiptNo: "GRN-260427-0004",
    warehouseId: "wh-hcm-fg",
    warehouseCode: "WH-HCM-FG",
    purchaseOrderLineId: "po-line-260427-0004-001",
    goodsReceiptLineId: "grn-line-fail-001",
    inboundQCInspectionId: "iqc-fail-002",
    itemId: "item-toner-100ml",
    sku: "TONER-100ML",
    itemName: "Hydrating Toner",
    batchId: "batch-toner-2604c",
    batchNo: "LOT-2604C",
    lotNo: "LOT-2604C",
    expiryDate: "2027-10-10",
    rejectedQuantity: "4.000000",
    uomCode: "EA",
    baseUOMCode: "EA",
    reason: "Supplier label mismatch"
  }
];

let prototypeSupplierRejections = createPrototypeSupplierRejections();

export async function getSupplierRejections(query: SupplierRejectionQuery = {}): Promise<SupplierRejection[]> {
  try {
    const rows = await apiGetRaw<SupplierRejectionApi[]>(`/supplier-rejections${supplierRejectionQueryString(query)}`, {
      accessToken: defaultAccessToken
    });

    return rows.map(fromApiSupplierRejection);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return filterPrototypeSupplierRejections(query);
  }
}

export async function createSupplierRejection(input: CreateSupplierRejectionInput): Promise<SupplierRejection> {
  try {
    const rejection = await apiPost<SupplierRejectionApi, CreateSupplierRejectionApiRequest>(
      "/supplier-rejections",
      toApiCreateInput(input),
      { accessToken: defaultAccessToken }
    );

    return fromApiSupplierRejection(rejection);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return createPrototypeSupplierRejection(input);
  }
}

export async function submitSupplierRejection(id: string): Promise<SupplierRejectionActionResult> {
  return transitionSupplierRejection(id, "submit");
}

export async function confirmSupplierRejection(id: string): Promise<SupplierRejectionActionResult> {
  return transitionSupplierRejection(id, "confirm");
}

export function resetPrototypeSupplierRejectionsForTest() {
  supplierRejectionSequence = 1;
  prototypeSupplierRejections = createPrototypeSupplierRejections();
}

export function supplierRejectionStatusTone(status: SupplierRejectionStatus) {
  switch (status) {
    case "confirmed":
      return "success" as const;
    case "submitted":
      return "info" as const;
    case "cancelled":
      return "normal" as const;
    case "draft":
    default:
      return "warning" as const;
  }
}

export function formatSupplierRejectionStatus(status: SupplierRejectionStatus) {
  switch (status) {
    case "submitted":
      return "Submitted";
    case "confirmed":
      return "Confirmed";
    case "cancelled":
      return "Cancelled";
    case "draft":
    default:
      return "Draft";
  }
}

export function formatSupplierRejectionQuantity(value: string, uomCode?: string) {
  return formatQuantity(value, uomCode);
}

export function formatSupplierRejectionDateTime(value?: string) {
  return value ? formatDateTimeVI(value) : "-";
}

function transitionSupplierRejection(id: string, action: "submit" | "confirm"): Promise<SupplierRejectionActionResult> {
  return apiPost<SupplierRejectionActionResultApi, Record<string, never>>(
    `/supplier-rejections/${encodeURIComponent(id)}/${action}`,
    {},
    { accessToken: defaultAccessToken }
  )
    .then(fromApiActionResult)
    .catch((cause) => {
      if (!shouldUsePrototypeFallback(cause)) {
        throw cause;
      }

      return transitionPrototypeSupplierRejection(id, action);
    });
}

function fromApiSupplierRejection(rejection: SupplierRejectionApi): SupplierRejection {
  return {
    id: rejection.id,
    orgId: rejection.org_id,
    rejectionNo: rejection.rejection_no,
    supplierId: rejection.supplier_id,
    supplierCode: rejection.supplier_code,
    supplierName: rejection.supplier_name,
    purchaseOrderId: rejection.purchase_order_id,
    purchaseOrderNo: rejection.purchase_order_no,
    goodsReceiptId: rejection.goods_receipt_id,
    goodsReceiptNo: rejection.goods_receipt_no,
    inboundQCInspectionId: rejection.inbound_qc_inspection_id,
    warehouseId: rejection.warehouse_id,
    warehouseCode: rejection.warehouse_code,
    status: rejection.status,
    reason: rejection.reason,
    lines: rejection.lines.map(fromApiSupplierRejectionLine),
    attachments: rejection.attachments.map(fromApiSupplierRejectionAttachment),
    auditLogId: rejection.audit_log_id,
    createdAt: rejection.created_at,
    createdBy: rejection.created_by,
    updatedAt: rejection.updated_at,
    updatedBy: rejection.updated_by,
    submittedAt: rejection.submitted_at,
    submittedBy: rejection.submitted_by,
    confirmedAt: rejection.confirmed_at,
    confirmedBy: rejection.confirmed_by
  };
}

function fromApiSupplierRejectionLine(line: SupplierRejectionLineApi): SupplierRejectionLine {
  return {
    id: line.id,
    purchaseOrderLineId: line.purchase_order_line_id,
    goodsReceiptLineId: line.goods_receipt_line_id,
    inboundQCInspectionId: line.inbound_qc_inspection_id,
    itemId: line.item_id,
    sku: line.sku,
    itemName: line.item_name,
    batchId: line.batch_id,
    batchNo: line.batch_no,
    lotNo: line.lot_no,
    expiryDate: line.expiry_date,
    rejectedQuantity: line.rejected_qty,
    uomCode: line.uom_code,
    baseUOMCode: line.base_uom_code,
    reason: line.reason
  };
}

function fromApiSupplierRejectionAttachment(attachment: SupplierRejectionAttachmentApi): SupplierRejectionAttachment {
  return {
    id: attachment.id,
    lineId: attachment.line_id,
    fileName: attachment.file_name,
    objectKey: attachment.object_key,
    contentType: attachment.content_type,
    uploadedAt: attachment.uploaded_at,
    uploadedBy: attachment.uploaded_by,
    source: attachment.source
  };
}

function fromApiActionResult(result: SupplierRejectionActionResultApi): SupplierRejectionActionResult {
  return {
    rejection: fromApiSupplierRejection(result.rejection),
    previousStatus: result.previous_status,
    currentStatus: result.current_status,
    auditLogId: result.audit_log_id
  };
}

function toApiCreateInput(input: CreateSupplierRejectionInput): CreateSupplierRejectionApiRequest {
  const id = input.id?.trim() || nextSupplierRejectionID();
  const firstLine = input.lines[0];

  return {
    id,
    org_id: input.orgId,
    rejection_no: input.rejectionNo?.trim().toUpperCase() || nextSupplierRejectionNo(),
    supplier_id: input.supplierId,
    supplier_code: input.supplierCode,
    supplier_name: input.supplierName,
    purchase_order_id: input.purchaseOrderId,
    purchase_order_no: input.purchaseOrderNo,
    goods_receipt_id: input.goodsReceiptId,
    goods_receipt_no: input.goodsReceiptNo,
    inbound_qc_inspection_id: input.inboundQCInspectionId,
    warehouse_id: input.warehouseId,
    warehouse_code: input.warehouseCode,
    reason: input.reason,
    lines: input.lines.map((line, index) => toApiCreateLineInput(line, index)),
    attachments: input.attachments?.map((attachment, index) => toApiCreateAttachmentInput(attachment, index, id, firstLine?.id))
  };
}

function toApiCreateLineInput(
  input: CreateSupplierRejectionLineInput,
  index: number
): CreateSupplierRejectionLineApiRequest {
  return {
    id: input.id?.trim() || `line-${String(index + 1).padStart(3, "0")}`,
    purchase_order_line_id: input.purchaseOrderLineId,
    goods_receipt_line_id: input.goodsReceiptLineId,
    inbound_qc_inspection_id: input.inboundQCInspectionId,
    item_id: input.itemId,
    sku: input.sku,
    item_name: input.itemName,
    batch_id: input.batchId,
    batch_no: input.batchNo,
    lot_no: input.lotNo,
    expiry_date: input.expiryDate,
    rejected_qty: normalizeDecimalInput(input.rejectedQuantity, decimalScales.quantity),
    uom_code: input.uomCode,
    base_uom_code: input.baseUOMCode,
    reason: input.reason
  };
}

function toApiCreateAttachmentInput(
  input: CreateSupplierRejectionAttachmentInput,
  index: number,
  rejectionID: string,
  fallbackLineID?: string
): CreateSupplierRejectionAttachmentApiRequest {
  const fileName = input.fileName.trim();
  return {
    id: input.id?.trim() || `att-${String(index + 1).padStart(3, "0")}`,
    line_id: input.lineId?.trim() || fallbackLineID,
    file_name: fileName,
    object_key: input.objectKey?.trim() || `supplier-rejections/${rejectionID}/${fileName}`,
    content_type: input.contentType,
    source: input.source
  };
}

function supplierRejectionQueryString(query: SupplierRejectionQuery) {
  const params = new URLSearchParams();
  if (query.supplierId) {
    params.set("supplier_id", query.supplierId);
  }
  if (query.warehouseId) {
    params.set("warehouse_id", query.warehouseId);
  }
  if (query.status) {
    params.set("status", query.status);
  }

  const value = params.toString();
  return value ? `?${value}` : "";
}

function shouldUsePrototypeFallback(reason: unknown) {
  if (reason instanceof ApiError) {
    return false;
  }

  return !(reason instanceof Error && reason.message.startsWith("API request failed:"));
}

function filterPrototypeSupplierRejections(query: SupplierRejectionQuery) {
  return prototypeSupplierRejections
    .filter((rejection) => {
      if (query.supplierId && rejection.supplierId !== query.supplierId) {
        return false;
      }
      if (query.warehouseId && rejection.warehouseId !== query.warehouseId) {
        return false;
      }
      if (query.status && rejection.status !== query.status) {
        return false;
      }

      return true;
    })
    .map(cloneSupplierRejection)
    .sort((left, right) => right.updatedAt.localeCompare(left.updatedAt));
}

function createPrototypeSupplierRejection(input: CreateSupplierRejectionInput): SupplierRejection {
  const payload = toApiCreateInput(input);
  if (payload.lines.length === 0) {
    throw new Error("At least one rejected line is required");
  }
  if (
    payload.supplier_id.trim() === "" ||
    payload.goods_receipt_id.trim() === "" ||
    payload.inbound_qc_inspection_id.trim() === "" ||
    payload.warehouse_id.trim() === "" ||
    payload.reason.trim() === ""
  ) {
    throw new Error("Supplier, goods receipt, QC inspection, warehouse, and reason are required");
  }

  const rejection: SupplierRejection = {
    id: payload.id,
    orgId: payload.org_id?.trim() || "org-my-pham",
    rejectionNo: payload.rejection_no,
    supplierId: payload.supplier_id,
    supplierCode: payload.supplier_code,
    supplierName: payload.supplier_name,
    purchaseOrderId: payload.purchase_order_id,
    purchaseOrderNo: payload.purchase_order_no,
    goodsReceiptId: payload.goods_receipt_id,
    goodsReceiptNo: payload.goods_receipt_no,
    inboundQCInspectionId: payload.inbound_qc_inspection_id,
    warehouseId: payload.warehouse_id,
    warehouseCode: payload.warehouse_code,
    status: "draft",
    reason: payload.reason.trim(),
    lines: payload.lines.map(fromCreateLineApiInput),
    attachments: (payload.attachments ?? []).map(fromCreateAttachmentApiInput),
    auditLogId: `audit-${payload.id}-created`,
    createdAt: prototypeNow,
    createdBy: "user-warehouse-lead",
    updatedAt: prototypeNow,
    updatedBy: "user-warehouse-lead"
  };

  prototypeSupplierRejections = [rejection, ...prototypeSupplierRejections.filter((candidate) => candidate.id !== rejection.id)];

  return cloneSupplierRejection(rejection);
}

function transitionPrototypeSupplierRejection(
  id: string,
  action: "submit" | "confirm"
): SupplierRejectionActionResult {
  const current = prototypeSupplierRejections.find((candidate) => candidate.id === id);
  if (!current) {
    throw new Error("Supplier rejection not found");
  }
  if (action === "submit" && current.status !== "draft") {
    throw new Error("Supplier rejection must be draft before submit");
  }
  if (action === "confirm" && current.status !== "submitted") {
    throw new Error("Supplier rejection must be submitted before confirm");
  }

  const updated: SupplierRejection = {
    ...cloneSupplierRejection(current),
    status: action === "submit" ? "submitted" : "confirmed",
    updatedAt: action === "submit" ? "2026-04-29T10:40:00Z" : "2026-04-29T10:50:00Z",
    updatedBy: "user-warehouse-lead",
    auditLogId: `audit-${current.id}-${action === "submit" ? "submitted" : "confirmed"}`
  };
  if (action === "submit") {
    updated.submittedAt = updated.updatedAt;
    updated.submittedBy = updated.updatedBy;
  } else {
    updated.confirmedAt = updated.updatedAt;
    updated.confirmedBy = updated.updatedBy;
  }
  prototypeSupplierRejections = prototypeSupplierRejections.map((candidate) =>
    candidate.id === id ? updated : candidate
  );

  return {
    rejection: cloneSupplierRejection(updated),
    previousStatus: current.status,
    currentStatus: updated.status,
    auditLogId: updated.auditLogId
  };
}

function fromCreateLineApiInput(input: CreateSupplierRejectionLineApiRequest): SupplierRejectionLine {
  return {
    id: input.id || "line-001",
    purchaseOrderLineId: input.purchase_order_line_id,
    goodsReceiptLineId: input.goods_receipt_line_id,
    inboundQCInspectionId: input.inbound_qc_inspection_id,
    itemId: input.item_id,
    sku: input.sku.trim().toUpperCase(),
    itemName: input.item_name,
    batchId: input.batch_id,
    batchNo: input.batch_no.trim().toUpperCase(),
    lotNo: input.lot_no.trim().toUpperCase(),
    expiryDate: input.expiry_date,
    rejectedQuantity: input.rejected_qty,
    uomCode: input.uom_code.trim().toUpperCase(),
    baseUOMCode: input.base_uom_code.trim().toUpperCase(),
    reason: input.reason.trim()
  };
}

function fromCreateAttachmentApiInput(input: CreateSupplierRejectionAttachmentApiRequest): SupplierRejectionAttachment {
  return {
    id: input.id || "att-001",
    lineId: input.line_id,
    fileName: input.file_name,
    objectKey: input.object_key,
    contentType: input.content_type,
    uploadedAt: prototypeNow,
    uploadedBy: "user-warehouse-lead",
    source: input.source
  };
}

function createPrototypeSupplierRejections(): SupplierRejection[] {
  const firstLine = supplierRejectionSampleLines[0];
  const supplier = supplierRejectionSupplierOptions[0];

  return [
    {
      id: "srj-prototype-confirmed",
      orgId: "org-my-pham",
      rejectionNo: "SRJ-260429-0001",
      supplierId: supplier.value,
      supplierCode: supplier.code,
      supplierName: supplier.name,
      purchaseOrderId: firstLine.purchaseOrderId,
      purchaseOrderNo: firstLine.purchaseOrderNo,
      goodsReceiptId: firstLine.goodsReceiptId,
      goodsReceiptNo: firstLine.goodsReceiptNo,
      inboundQCInspectionId: firstLine.inboundQCInspectionId,
      warehouseId: firstLine.warehouseId,
      warehouseCode: firstLine.warehouseCode,
      status: "confirmed",
      reason: "Damaged packaging",
      lines: [{ ...firstLine, id: "srj-prototype-line-001" }],
      attachments: [
        {
          id: "srj-prototype-att-001",
          lineId: "srj-prototype-line-001",
          fileName: "damage-photo.jpg",
          objectKey: "supplier-rejections/srj-prototype-confirmed/damage-photo.jpg",
          contentType: "image/jpeg",
          uploadedAt: "2026-04-29T09:05:00Z",
          uploadedBy: "user-warehouse-lead",
          source: "inbound_qc"
        }
      ],
      auditLogId: "audit-srj-prototype-confirmed",
      createdAt: "2026-04-29T09:00:00Z",
      createdBy: "user-warehouse-lead",
      updatedAt: "2026-04-29T09:20:00Z",
      updatedBy: "user-warehouse-lead",
      submittedAt: "2026-04-29T09:10:00Z",
      submittedBy: "user-warehouse-lead",
      confirmedAt: "2026-04-29T09:20:00Z",
      confirmedBy: "user-warehouse-lead"
    }
  ];
}

function cloneSupplierRejection(rejection: SupplierRejection): SupplierRejection {
  return {
    ...rejection,
    lines: rejection.lines.map((line) => ({ ...line })),
    attachments: rejection.attachments.map((attachment) => ({ ...attachment }))
  };
}

function nextSupplierRejectionID() {
  return `srj-ui-${Date.now()}-${supplierRejectionSequence++}`;
}

function nextSupplierRejectionNo() {
  const datePart = new Date().toISOString().slice(2, 10).replaceAll("-", "");
  return `SRJ-${datePart}-${String(Date.now()).slice(-8)}`;
}
