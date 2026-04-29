import { ApiError, apiGetRaw, apiPatch, apiPost } from "../../../shared/api/client";
import {
  decimalScales,
  formatDateVI,
  formatMoney,
  formatQuantity,
  normalizeDecimalInput
} from "../../../shared/format/numberFormat";
import type {
  CreatePurchaseOrderInput,
  PurchaseOrder,
  PurchaseOrderActionResult,
  PurchaseOrderLine,
  PurchaseOrderLineInput,
  PurchaseOrderQuery,
  PurchaseOrderStatus,
  UpdatePurchaseOrderInput
} from "../types";

type PurchaseOrderLineApi = {
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
  status: PurchaseOrderStatus;
  currency_code: string;
  subtotal_amount: string;
  total_amount: string;
  note?: string;
  lines: PurchaseOrderLineApi[];
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

type PurchaseOrderListItemApi = {
  id: string;
  po_no: string;
  supplier_id: string;
  supplier_code?: string;
  supplier_name: string;
  warehouse_id: string;
  warehouse_code?: string;
  expected_date: string;
  status: PurchaseOrderStatus;
  currency_code: string;
  total_amount: string;
  line_count: number;
  received_line_count: number;
  created_at: string;
  updated_at: string;
  version: number;
};

type PurchaseOrderLineApiRequest = {
  id?: string;
  line_no?: number;
  item_id: string;
  ordered_qty: string;
  uom_code: string;
  unit_price: string;
  currency_code?: string;
  expected_date?: string;
  note?: string;
};

type CreatePurchaseOrderApiRequest = {
  id?: string;
  po_no?: string;
  supplier_id: string;
  warehouse_id: string;
  expected_date: string;
  currency_code: string;
  note?: string;
  lines: PurchaseOrderLineApiRequest[];
};

type UpdatePurchaseOrderApiRequest = {
  supplier_id?: string;
  warehouse_id?: string;
  expected_date?: string;
  note?: string;
  expected_version?: number;
  lines?: PurchaseOrderLineApiRequest[];
};

type PurchaseOrderActionApiResult = {
  purchase_order: PurchaseOrderApi;
  previous_status: PurchaseOrderStatus;
  current_status: PurchaseOrderStatus;
  audit_log_id?: string;
};

type SupplierOption = {
  label: string;
  value: string;
  code: string;
};

type WarehouseOption = {
  label: string;
  value: string;
  code: string;
};

type ItemOption = {
  label: string;
  value: string;
  skuCode: string;
  itemName: string;
  baseUomCode: string;
  defaultUnitPrice: string;
};

const defaultAccessToken = "local-dev-access-token";
const prototypeNow = "2026-04-29T10:00:00Z";
let orderSequence = 20;

export const purchaseSupplierOptions: SupplierOption[] = [
  { label: "BioActive Raw Materials", value: "sup-rm-bioactive", code: "SUP-RM-BIO" },
  { label: "Vina Packaging Solutions", value: "sup-pkg-vina", code: "SUP-PKG-VINA" },
  { label: "Lotus Filling Partner", value: "sup-out-lotus", code: "SUP-OUT-LOTUS" }
];

export const purchaseWarehouseOptions: WarehouseOption[] = [
  { label: "Raw Material HCM", value: "wh-hcm-rm", code: "WH-HCM-RM" },
  { label: "Finished Goods HCM", value: "wh-hcm-fg", code: "WH-HCM-FG" }
];

export const purchaseItemOptions: ItemOption[] = [
  {
    label: "SERUM-30ML / Hydrating Serum 30ml",
    value: "item-serum-30ml",
    skuCode: "SERUM-30ML",
    itemName: "Hydrating Serum 30ml",
    baseUomCode: "EA",
    defaultUnitPrice: "125000.0000"
  },
  {
    label: "CREAM-50G / Repair Cream 50g",
    value: "item-cream-50g",
    skuCode: "CREAM-50G",
    itemName: "Repair Cream 50g",
    baseUomCode: "EA",
    defaultUnitPrice: "95000.0000"
  },
  {
    label: "TONER-100ML / Balancing Toner 100ml",
    value: "item-toner-100ml",
    skuCode: "TONER-100ML",
    itemName: "Balancing Toner 100ml",
    baseUomCode: "EA",
    defaultUnitPrice: "82000.0000"
  }
];

export const purchaseStatusOptions: { label: string; value: "" | PurchaseOrderStatus }[] = [
  { label: "All statuses", value: "" },
  { label: "Draft", value: "draft" },
  { label: "Submitted", value: "submitted" },
  { label: "Approved", value: "approved" },
  { label: "Partially received", value: "partially_received" },
  { label: "Received", value: "received" },
  { label: "Closed", value: "closed" },
  { label: "Cancelled", value: "cancelled" },
  { label: "Rejected", value: "rejected" }
];

let prototypePurchaseOrders = createPrototypePurchaseOrders();

export async function getPurchaseOrders(query: PurchaseOrderQuery = {}): Promise<PurchaseOrder[]> {
  try {
    const orders = await apiGetRaw<PurchaseOrderListItemApi[]>(`/purchase-orders${purchaseOrderQueryString(query)}`, {
      accessToken: defaultAccessToken
    });

    return orders.map(fromApiPurchaseOrderListItem);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return filterPrototypePurchaseOrders(query);
  }
}

export async function getPurchaseOrder(id: string): Promise<PurchaseOrder> {
  try {
    const order = await apiGetRaw<PurchaseOrderApi>(`/purchase-orders/${encodeURIComponent(id)}`, {
      accessToken: defaultAccessToken
    });

    return fromApiPurchaseOrder(order);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return getPrototypePurchaseOrder(id);
  }
}

export async function createPurchaseOrder(input: CreatePurchaseOrderInput): Promise<PurchaseOrder> {
  try {
    const order = await apiPost<PurchaseOrderApi, CreatePurchaseOrderApiRequest>(
      "/purchase-orders",
      toApiCreateInput(input),
      { accessToken: defaultAccessToken }
    );

    return fromApiPurchaseOrder(order);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return createPrototypePurchaseOrder(input);
  }
}

export async function updatePurchaseOrder(id: string, input: UpdatePurchaseOrderInput): Promise<PurchaseOrder> {
  try {
    const order = await apiPatch<PurchaseOrderApi, UpdatePurchaseOrderApiRequest>(
      `/purchase-orders/${encodeURIComponent(id)}`,
      toApiUpdateInput(input),
      { accessToken: defaultAccessToken }
    );

    return fromApiPurchaseOrder(order);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return updatePrototypePurchaseOrder(id, input);
  }
}

export async function submitPurchaseOrder(id: string, expectedVersion?: number): Promise<PurchaseOrderActionResult> {
  return runPurchaseOrderAction(id, "submit", expectedVersion);
}

export async function approvePurchaseOrder(id: string, expectedVersion?: number): Promise<PurchaseOrderActionResult> {
  return runPurchaseOrderAction(id, "approve", expectedVersion);
}

export async function cancelPurchaseOrder(
  id: string,
  reason: string,
  expectedVersion?: number
): Promise<PurchaseOrderActionResult> {
  return runPurchaseOrderAction(id, "cancel", expectedVersion, reason);
}

export async function closePurchaseOrder(id: string, expectedVersion?: number): Promise<PurchaseOrderActionResult> {
  return runPurchaseOrderAction(id, "close", expectedVersion);
}

export function resetPrototypePurchaseOrdersForTest() {
  orderSequence = 20;
  prototypePurchaseOrders = createPrototypePurchaseOrders();
}

export function purchaseOrderStatusTone(status: PurchaseOrderStatus): "success" | "warning" | "danger" | "info" | "normal" {
  switch (status) {
    case "approved":
    case "partially_received":
      return "info";
    case "received":
    case "closed":
      return "success";
    case "cancelled":
    case "rejected":
      return "danger";
    case "draft":
    case "submitted":
    default:
      return "warning";
  }
}

export function formatPurchaseOrderStatus(status: PurchaseOrderStatus) {
  return status
    .split("_")
    .map((part) => `${part.slice(0, 1).toUpperCase()}${part.slice(1)}`)
    .join(" ");
}

export function formatPurchaseMoney(value: string, currencyCode = "VND") {
  return formatMoney(value, currencyCode);
}

export function formatPurchaseQuantity(value: string, uomCode?: string) {
  return formatQuantity(value, uomCode);
}

export function formatPurchaseDate(value: string) {
  return formatDateVI(value);
}

async function runPurchaseOrderAction(
  id: string,
  action: "submit" | "approve" | "cancel" | "close",
  expectedVersion?: number,
  reason?: string
): Promise<PurchaseOrderActionResult> {
  try {
    const result = await apiPost<PurchaseOrderActionApiResult, { expected_version?: number; reason?: string }>(
      `/purchase-orders/${encodeURIComponent(id)}/${action}`,
      { expected_version: expectedVersion, reason },
      { accessToken: defaultAccessToken }
    );

    return fromApiActionResult(result);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return transitionPrototypePurchaseOrder(id, action, expectedVersion, reason);
  }
}

function shouldUsePrototypeFallback(reason: unknown) {
  if (reason instanceof ApiError) {
    return false;
  }

  return !(reason instanceof Error && reason.message.startsWith("API request failed:"));
}

function fromApiPurchaseOrderListItem(order: PurchaseOrderListItemApi): PurchaseOrder {
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
    subtotalAmount: order.total_amount,
    totalAmount: order.total_amount,
    lineCount: order.line_count,
    receivedLineCount: order.received_line_count,
    lines: [],
    createdAt: order.created_at,
    updatedAt: order.updated_at,
    version: order.version ?? 1
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
    lines: order.lines.map(fromApiPurchaseOrderLine),
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

function fromApiPurchaseOrderLine(line: PurchaseOrderLineApi): PurchaseOrderLine {
  return {
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
  };
}

function fromApiActionResult(result: PurchaseOrderActionApiResult): PurchaseOrderActionResult {
  return {
    purchaseOrder: fromApiPurchaseOrder(result.purchase_order),
    previousStatus: result.previous_status,
    currentStatus: result.current_status,
    auditLogId: result.audit_log_id
  };
}

function toApiCreateInput(input: CreatePurchaseOrderInput): CreatePurchaseOrderApiRequest {
  return {
    id: input.id,
    po_no: input.poNo,
    supplier_id: input.supplierId,
    warehouse_id: input.warehouseId,
    expected_date: input.expectedDate,
    currency_code: input.currencyCode,
    note: input.note,
    lines: input.lines.map(toApiLineInput)
  };
}

function toApiUpdateInput(input: UpdatePurchaseOrderInput): UpdatePurchaseOrderApiRequest {
  return {
    supplier_id: input.supplierId,
    warehouse_id: input.warehouseId,
    expected_date: input.expectedDate,
    note: input.note,
    expected_version: input.expectedVersion,
    lines: input.lines?.map((line) => ({ ...toApiLineInput(line), id: line.id, line_no: line.lineNo }))
  };
}

function toApiLineInput(line: PurchaseOrderLineInput): PurchaseOrderLineApiRequest {
  return {
    item_id: line.itemId,
    ordered_qty: normalizeDecimalInput(line.orderedQty, decimalScales.quantity),
    uom_code: line.uomCode,
    unit_price: normalizeDecimalInput(line.unitPrice, decimalScales.unitPrice),
    currency_code: line.currencyCode ?? "VND",
    expected_date: line.expectedDate,
    note: line.note
  };
}

function purchaseOrderQueryString(query: PurchaseOrderQuery) {
  const params = new URLSearchParams();
  if (query.search) {
    params.set("search", query.search);
  }
  if (query.status) {
    params.set("status", query.status);
  }
  if (query.supplierId) {
    params.set("supplier_id", query.supplierId);
  }
  if (query.warehouseId) {
    params.set("warehouse_id", query.warehouseId);
  }
  if (query.expectedFrom) {
    params.set("expected_from", query.expectedFrom);
  }
  if (query.expectedTo) {
    params.set("expected_to", query.expectedTo);
  }

  const value = params.toString();
  return value ? `?${value}` : "";
}

function getPrototypePurchaseOrder(id: string) {
  const order = prototypePurchaseOrders.find((candidate) => candidate.id === id);
  if (!order) {
    throw new Error("Purchase order not found");
  }

  return clonePurchaseOrder(order);
}

function filterPrototypePurchaseOrders(query: PurchaseOrderQuery) {
  return prototypePurchaseOrders.filter((order) => matchesPurchaseOrderQuery(order, query)).map(clonePurchaseOrder);
}

function createPrototypePurchaseOrder(input: CreatePurchaseOrderInput): PurchaseOrder {
  if (input.lines.length === 0) {
    throw new Error("At least one line item is required");
  }
  orderSequence += 1;
  const supplier = findSupplier(input.supplierId);
  const warehouse = findWarehouse(input.warehouseId);
  const id = input.id ?? `po-ui-${orderSequence}`;
  const poNo = input.poNo ?? `PO-260429-${String(orderSequence).padStart(4, "0")}`;
  const lines = input.lines.map((line, index) => hydrateLineInput(line, id, input.expectedDate, index));
  const totalAmount = calculateTotalAmount(lines);
  const order: PurchaseOrder = {
    id,
    poNo,
    supplierId: supplier.value,
    supplierCode: supplier.code,
    supplierName: supplier.label,
    warehouseId: warehouse.value,
    warehouseCode: warehouse.code,
    expectedDate: input.expectedDate,
    status: "draft",
    currencyCode: input.currencyCode,
    subtotalAmount: totalAmount,
    totalAmount,
    note: input.note,
    lines,
    lineCount: lines.length,
    receivedLineCount: 0,
    auditLogId: `audit-${id}`,
    createdAt: prototypeNow,
    updatedAt: prototypeNow,
    version: 1
  };
  prototypePurchaseOrders = [order, ...prototypePurchaseOrders.filter((candidate) => candidate.id !== id)];

  return clonePurchaseOrder(order);
}

function updatePrototypePurchaseOrder(id: string, input: UpdatePurchaseOrderInput): PurchaseOrder {
  const current = getPrototypePurchaseOrder(id);
  if (current.status !== "draft") {
    throw new Error("Only draft purchase orders can be updated");
  }
  if (input.expectedVersion && input.expectedVersion !== current.version) {
    throw new Error("Purchase order version changed");
  }
  const supplier = input.supplierId ? findSupplier(input.supplierId) : undefined;
  const warehouse = input.warehouseId ? findWarehouse(input.warehouseId) : undefined;
  const expectedDate = input.expectedDate ?? current.expectedDate;
  const lines = input.lines ? input.lines.map((line, index) => hydrateLineInput(line, id, expectedDate, index)) : current.lines;
  const totalAmount = calculateTotalAmount(lines);
  const updated: PurchaseOrder = {
    ...current,
    supplierId: supplier?.value ?? current.supplierId,
    supplierCode: supplier?.code ?? current.supplierCode,
    supplierName: supplier?.label ?? current.supplierName,
    warehouseId: warehouse?.value ?? current.warehouseId,
    warehouseCode: warehouse?.code ?? current.warehouseCode,
    expectedDate,
    note: input.note ?? current.note,
    lines,
    lineCount: lines.length,
    subtotalAmount: totalAmount,
    totalAmount,
    updatedAt: prototypeNow,
    version: current.version + 1
  };
  prototypePurchaseOrders = [updated, ...prototypePurchaseOrders.filter((candidate) => candidate.id !== id)];

  return clonePurchaseOrder(updated);
}

function transitionPrototypePurchaseOrder(
  id: string,
  action: "submit" | "approve" | "cancel" | "close",
  expectedVersion?: number,
  reason?: string
): PurchaseOrderActionResult {
  const current = getPrototypePurchaseOrder(id);
  if (expectedVersion && expectedVersion !== current.version) {
    throw new Error("Purchase order version changed");
  }
  const nextStatus = nextPrototypeStatus(current.status, action);
  const updated: PurchaseOrder = {
    ...current,
    status: nextStatus,
    updatedAt: prototypeNow,
    submittedAt: nextStatus === "submitted" ? prototypeNow : current.submittedAt,
    approvedAt: nextStatus === "approved" ? prototypeNow : current.approvedAt,
    closedAt: nextStatus === "closed" ? prototypeNow : current.closedAt,
    cancelledAt: nextStatus === "cancelled" ? prototypeNow : current.cancelledAt,
    cancelReason: nextStatus === "cancelled" ? reason : current.cancelReason,
    version: current.version + 1
  };
  prototypePurchaseOrders = [updated, ...prototypePurchaseOrders.filter((candidate) => candidate.id !== id)];

  return {
    purchaseOrder: clonePurchaseOrder(updated),
    previousStatus: current.status,
    currentStatus: updated.status,
    auditLogId: `audit-${id}-${action}`
  };
}

function nextPrototypeStatus(currentStatus: PurchaseOrderStatus, action: "submit" | "approve" | "cancel" | "close") {
  if (action === "submit" && currentStatus === "draft") {
    return "submitted";
  }
  if (action === "approve" && currentStatus === "submitted") {
    return "approved";
  }
  if (action === "close" && ["approved", "partially_received", "received"].includes(currentStatus)) {
    return "closed";
  }
  if (action === "cancel" && ["draft", "submitted", "approved"].includes(currentStatus)) {
    return "cancelled";
  }

  throw new Error(`Cannot ${action} purchase order from ${formatPurchaseOrderStatus(currentStatus)}`);
}

function hydrateLineInput(
  input: PurchaseOrderLineInput,
  orderId: string,
  expectedDate: string,
  index: number
): PurchaseOrderLine {
  const item = findItem(input.itemId);
  const orderedQty = normalizeDecimalInput(input.orderedQty, decimalScales.quantity);
  const unitPrice = normalizeDecimalInput(input.unitPrice, decimalScales.unitPrice);
  const lineAmount = calculateLineAmount(orderedQty, unitPrice);

  return {
    id: input.id ?? `${orderId}-line-${String(index + 1).padStart(2, "0")}`,
    lineNo: input.lineNo ?? index + 1,
    itemId: item.value,
    skuCode: item.skuCode,
    itemName: item.itemName,
    orderedQty,
    receivedQty: "0.000000",
    uomCode: input.uomCode,
    baseOrderedQty: orderedQty,
    baseReceivedQty: "0.000000",
    baseUomCode: item.baseUomCode,
    conversionFactor: "1.000000",
    unitPrice,
    currencyCode: input.currencyCode ?? "VND",
    lineAmount,
    expectedDate: input.expectedDate ?? expectedDate,
    note: input.note
  };
}

function calculateTotalAmount(lines: PurchaseOrderLine[]) {
  return lines.reduce((sum, line) => addMoneyStrings(sum, line.lineAmount), "0.00");
}

function calculateLineAmount(orderedQty: string, unitPrice: string) {
  const quantity = toScaledBigInt(orderedQty, decimalScales.quantity);
  const price = toScaledBigInt(unitPrice, decimalScales.unitPrice);
  const gross = roundScaled(quantity * price, decimalScales.quantity + decimalScales.unitPrice, decimalScales.money);

  return fromScaledBigInt(gross, decimalScales.money);
}

function addMoneyStrings(left: string, right: string) {
  return fromScaledBigInt(
    toScaledBigInt(left, decimalScales.money) + toScaledBigInt(right, decimalScales.money),
    decimalScales.money
  );
}

function roundScaled(value: bigint, fromScale: number, toScale: number) {
  if (fromScale <= toScale) {
    return value * BigInt(10) ** BigInt(toScale - fromScale);
  }
  const factor = BigInt(10) ** BigInt(fromScale - toScale);
  const quotient = value / factor;
  const remainder = value % factor;

  return remainder * BigInt(2) >= factor ? quotient + BigInt(1) : quotient;
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

function findSupplier(id: string) {
  const supplier = purchaseSupplierOptions.find((candidate) => candidate.value === id);
  if (!supplier) {
    throw new Error("Supplier is required");
  }

  return supplier;
}

function findWarehouse(id: string) {
  const warehouse = purchaseWarehouseOptions.find((candidate) => candidate.value === id);
  if (!warehouse) {
    throw new Error("Warehouse is required");
  }

  return warehouse;
}

function findItem(id: string) {
  const item = purchaseItemOptions.find((candidate) => candidate.value === id);
  if (!item) {
    throw new Error("Item is required");
  }

  return item;
}

function clonePurchaseOrder(order: PurchaseOrder): PurchaseOrder {
  return {
    ...order,
    lines: order.lines.map((line) => ({ ...line }))
  };
}

function createPrototypePurchaseOrders(): PurchaseOrder[] {
  const draft = createOrderSeed({
    id: "po-260429-0001",
    poNo: "PO-260429-0001",
    supplierId: "sup-rm-bioactive",
    warehouseId: "wh-hcm-rm",
    expectedDate: "2026-05-02",
    status: "draft",
    lines: [
      { itemId: "item-serum-30ml", orderedQty: "12", uomCode: "EA", unitPrice: "125000" },
      { itemId: "item-cream-50g", orderedQty: "6", uomCode: "EA", unitPrice: "95000" }
    ]
  });
  const submitted = createOrderSeed({
    id: "po-260429-0002",
    poNo: "PO-260429-0002",
    supplierId: "sup-pkg-vina",
    warehouseId: "wh-hcm-fg",
    expectedDate: "2026-05-03",
    status: "submitted",
    lines: [{ itemId: "item-toner-100ml", orderedQty: "40", uomCode: "EA", unitPrice: "82000" }],
    submittedAt: "2026-04-29T09:20:00Z"
  });
  const approved = createOrderSeed({
    id: "po-260429-0003",
    poNo: "PO-260429-0003",
    supplierId: "sup-rm-bioactive",
    warehouseId: "wh-hcm-rm",
    expectedDate: "2026-05-04",
    status: "approved",
    lines: [{ itemId: "item-cream-50g", orderedQty: "20", uomCode: "EA", unitPrice: "95000" }],
    submittedAt: "2026-04-29T08:30:00Z",
    approvedAt: "2026-04-29T09:40:00Z"
  });

  return [draft, submitted, approved];
}

function createOrderSeed(input: {
  id: string;
  poNo: string;
  supplierId: string;
  warehouseId: string;
  expectedDate: string;
  status: PurchaseOrderStatus;
  lines: PurchaseOrderLineInput[];
  submittedAt?: string;
  approvedAt?: string;
}) {
  const supplier = findSupplier(input.supplierId);
  const warehouse = findWarehouse(input.warehouseId);
  const lines = input.lines.map((line, index) => hydrateLineInput(line, input.id, input.expectedDate, index));
  const totalAmount = calculateTotalAmount(lines);

  return {
    id: input.id,
    poNo: input.poNo,
    supplierId: supplier.value,
    supplierCode: supplier.code,
    supplierName: supplier.label,
    warehouseId: warehouse.value,
    warehouseCode: warehouse.code,
    expectedDate: input.expectedDate,
    status: input.status,
    currencyCode: "VND",
    subtotalAmount: totalAmount,
    totalAmount,
    lines,
    lineCount: lines.length,
    receivedLineCount: 0,
    auditLogId: `audit-${input.id}`,
    createdAt: "2026-04-29T08:00:00Z",
    updatedAt: input.approvedAt ?? input.submittedAt ?? "2026-04-29T08:00:00Z",
    submittedAt: input.submittedAt,
    approvedAt: input.approvedAt,
    version: input.status === "draft" ? 1 : input.status === "submitted" ? 2 : 3
  } satisfies PurchaseOrder;
}

function matchesPurchaseOrderQuery(order: PurchaseOrder, query: PurchaseOrderQuery) {
  const search = query.search?.trim().toLowerCase();
  if (search) {
    const haystack = [order.poNo, order.supplierCode, order.supplierName, order.warehouseCode].join(" ").toLowerCase();
    if (!haystack.includes(search)) {
      return false;
    }
  }
  if (query.status && order.status !== query.status) {
    return false;
  }
  if (query.supplierId && order.supplierId !== query.supplierId) {
    return false;
  }
  if (query.warehouseId && order.warehouseId !== query.warehouseId) {
    return false;
  }
  if (query.expectedFrom && order.expectedDate < query.expectedFrom) {
    return false;
  }
  if (query.expectedTo && order.expectedDate > query.expectedTo) {
    return false;
  }

  return true;
}
