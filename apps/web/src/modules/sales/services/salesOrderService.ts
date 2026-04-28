import { apiGet, apiPatch, apiPost } from "../../../shared/api/client";
import type { components, operations } from "../../../shared/api/generated/schema";
import {
  decimalScales,
  formatDateVI,
  formatMoney,
  formatQuantity,
  normalizeDecimalInput
} from "../../../shared/format/numberFormat";
import type {
  CreateSalesOrderInput,
  SalesOrder,
  SalesOrderActionResult,
  SalesOrderLine,
  SalesOrderLineInput,
  SalesOrderQuery,
  SalesOrderStatus,
  UpdateSalesOrderInput
} from "../types";

type SalesOrderApi = components["schemas"]["SalesOrder"];
type SalesOrderListItemApi = components["schemas"]["SalesOrderListItem"];
type SalesOrderApiQuery = operations["listSalesOrders"]["parameters"]["query"];
type CreateSalesOrderApiRequest = components["schemas"]["CreateSalesOrderRequest"];
type UpdateSalesOrderApiRequest = components["schemas"]["UpdateSalesOrderRequest"];
type SalesOrderActionApiResult = components["schemas"]["SalesOrderActionResult"];

type CustomerOption = {
  label: string;
  value: string;
  code: string;
  channel: string;
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
const prototypeNow = "2026-04-28T12:00:00Z";
let orderSequence = 20;

export const salesCustomerOptions: CustomerOption[] = [
  { label: "Minh Anh Distributor", value: "cus-dl-minh-anh", code: "CUS-DL-MINHANH", channel: "B2B" },
  { label: "Linh Chi Dealer", value: "cus-dealer-linh-chi", code: "CUS-DL-LINHCHI", channel: "DEALER" },
  { label: "Shopee Marketplace", value: "cus-mp-shopee", code: "CUS-MP-SHOPEE", channel: "MP" },
  { label: "HCM Internal Store", value: "cus-internal-hcm-store", code: "CUS-INT-HCMSTORE", channel: "INT" }
];

export const salesWarehouseOptions: WarehouseOption[] = [
  { label: "Finished Goods HCM", value: "wh-hcm-fg", code: "WH-HCM-FG" },
  { label: "Raw Material HCM", value: "wh-hcm-rm", code: "WH-HCM-RM" }
];

export const salesItemOptions: ItemOption[] = [
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

export const salesChannelOptions = ["B2B", "DEALER", "MP", "INT"] as const;

export const salesStatusOptions: { label: string; value: "" | SalesOrderStatus }[] = [
  { label: "All statuses", value: "" },
  { label: "Draft", value: "draft" },
  { label: "Confirmed", value: "confirmed" },
  { label: "Reserved", value: "reserved" },
  { label: "Picking", value: "picking" },
  { label: "Packed", value: "packed" },
  { label: "Waiting handover", value: "waiting_handover" },
  { label: "Cancelled", value: "cancelled" }
];

let prototypeSalesOrders = createPrototypeSalesOrders();

export async function getSalesOrders(query: SalesOrderQuery = {}): Promise<SalesOrder[]> {
  try {
    const orders = await apiGet("/sales-orders", {
      accessToken: defaultAccessToken,
      query: toApiQuery(query)
    });

    return orders.map(fromApiSalesOrderListItem);
  } catch {
    return filterPrototypeSalesOrders(query);
  }
}

export async function getSalesOrder(id: string): Promise<SalesOrder> {
  try {
    const order = await apiGet(`/sales-orders/${encodeURIComponent(id)}` as "/sales-orders/{sales_order_id}", {
      accessToken: defaultAccessToken
    });

    return fromApiSalesOrder(order);
  } catch {
    return getPrototypeSalesOrder(id);
  }
}

export async function createSalesOrder(input: CreateSalesOrderInput): Promise<SalesOrder> {
  try {
    const order = await apiPost<SalesOrderApi, CreateSalesOrderApiRequest>(
      "/sales-orders",
      toApiCreateInput(input),
      { accessToken: defaultAccessToken }
    );

    return fromApiSalesOrder(order);
  } catch {
    return createPrototypeSalesOrder(input);
  }
}

export async function updateSalesOrder(id: string, input: UpdateSalesOrderInput): Promise<SalesOrder> {
  try {
    const order = await apiPatch<SalesOrderApi, UpdateSalesOrderApiRequest>(
      `/sales-orders/${encodeURIComponent(id)}`,
      toApiUpdateInput(input),
      { accessToken: defaultAccessToken }
    );

    return fromApiSalesOrder(order);
  } catch {
    return updatePrototypeSalesOrder(id, input);
  }
}

export async function confirmSalesOrder(id: string, expectedVersion?: number): Promise<SalesOrderActionResult> {
  try {
    const result = await apiPost<SalesOrderActionApiResult, components["schemas"]["SalesOrderActionRequest"]>(
      `/sales-orders/${encodeURIComponent(id)}/confirm`,
      { expected_version: expectedVersion },
      { accessToken: defaultAccessToken }
    );

    return fromApiActionResult(result);
  } catch {
    return transitionPrototypeSalesOrder(id, "confirmed", expectedVersion);
  }
}

export async function cancelSalesOrder(
  id: string,
  reason: string,
  expectedVersion?: number
): Promise<SalesOrderActionResult> {
  try {
    const result = await apiPost<SalesOrderActionApiResult, components["schemas"]["CancelSalesOrderRequest"]>(
      `/sales-orders/${encodeURIComponent(id)}/cancel`,
      { reason, expected_version: expectedVersion },
      { accessToken: defaultAccessToken }
    );

    return fromApiActionResult(result);
  } catch {
    return transitionPrototypeSalesOrder(id, "cancelled", expectedVersion, reason);
  }
}

export function resetPrototypeSalesOrdersForTest() {
  orderSequence = 20;
  prototypeSalesOrders = createPrototypeSalesOrders();
}

export function salesOrderStatusTone(status: SalesOrderStatus): "success" | "warning" | "danger" | "info" | "normal" {
  switch (status) {
    case "confirmed":
    case "reserved":
      return "info";
    case "picked":
    case "packed":
    case "handed_over":
    case "closed":
      return "success";
    case "reservation_failed":
    case "pick_exception":
    case "pack_exception":
    case "handover_exception":
    case "cancelled":
      return "danger";
    case "draft":
    case "picking":
    case "packing":
    case "waiting_handover":
    default:
      return "warning";
  }
}

export function formatSalesOrderStatus(status: SalesOrderStatus) {
  return status
    .split("_")
    .map((part) => `${part.slice(0, 1).toUpperCase()}${part.slice(1)}`)
    .join(" ");
}

export function formatSalesMoney(value: string, currencyCode = "VND") {
  return formatMoney(value, currencyCode);
}

export function formatSalesQuantity(value: string, uomCode?: string) {
  return formatQuantity(value, uomCode);
}

export function formatSalesDate(value: string) {
  return formatDateVI(value);
}

function fromApiSalesOrderListItem(order: SalesOrderListItemApi): SalesOrder {
  return {
    id: order.id,
    orderNo: order.order_no,
    customerId: order.customer_id,
    customerCode: order.customer_code,
    customerName: order.customer_name,
    channel: order.channel,
    warehouseId: order.warehouse_id,
    warehouseCode: order.warehouse_code,
    orderDate: order.order_date,
    status: order.status,
    currencyCode: order.currency_code,
    subtotalAmount: order.total_amount,
    discountAmount: "0.00",
    taxAmount: "0.00",
    shippingFeeAmount: "0.00",
    netAmount: order.total_amount,
    totalAmount: order.total_amount,
    lineCount: order.line_count,
    lines: [],
    createdAt: order.created_at,
    updatedAt: order.updated_at,
    version: order.version ?? 1
  };
}

function fromApiSalesOrder(order: SalesOrderApi): SalesOrder {
  return {
    id: order.id,
    orderNo: order.order_no,
    customerId: order.customer_id,
    customerCode: order.customer_code,
    customerName: order.customer_name,
    channel: order.channel,
    warehouseId: order.warehouse_id,
    warehouseCode: order.warehouse_code,
    orderDate: order.order_date,
    status: order.status,
    currencyCode: order.currency_code,
    subtotalAmount: order.subtotal_amount,
    discountAmount: order.discount_amount,
    taxAmount: order.tax_amount,
    shippingFeeAmount: order.shipping_fee_amount,
    netAmount: order.net_amount,
    totalAmount: order.total_amount,
    note: order.note,
    lineCount: order.lines.length,
    lines: order.lines.map(fromApiSalesOrderLine),
    auditLogId: order.audit_log_id,
    createdAt: order.created_at,
    updatedAt: order.updated_at,
    confirmedAt: order.confirmed_at,
    cancelledAt: order.cancelled_at,
    cancelReason: order.cancel_reason,
    version: order.version
  };
}

function fromApiSalesOrderLine(line: components["schemas"]["SalesOrderLine"]): SalesOrderLine {
  return {
    id: line.id,
    lineNo: line.line_no,
    itemId: line.item_id,
    skuCode: line.sku_code,
    itemName: line.item_name,
    orderedQty: line.ordered_qty,
    uomCode: line.uom_code,
    baseOrderedQty: line.base_ordered_qty,
    baseUomCode: line.base_uom_code,
    conversionFactor: line.conversion_factor,
    unitPrice: line.unit_price,
    currencyCode: line.currency_code,
    lineDiscountAmount: line.line_discount_amount,
    lineAmount: line.line_amount,
    reservedQty: line.reserved_qty,
    shippedQty: line.shipped_qty,
    batchId: line.batch_id,
    batchNo: line.batch_no
  };
}

function fromApiActionResult(result: SalesOrderActionApiResult): SalesOrderActionResult {
  return {
    salesOrder: fromApiSalesOrder(result.sales_order),
    previousStatus: result.previous_status,
    currentStatus: result.current_status,
    auditLogId: result.audit_log_id
  };
}

function toApiQuery(query: SalesOrderQuery): SalesOrderApiQuery {
  return {
    q: query.search,
    status: query.status,
    customer_id: query.customerId,
    channel: query.channel,
    warehouse_id: query.warehouseId
  };
}

function toApiCreateInput(input: CreateSalesOrderInput): CreateSalesOrderApiRequest {
  return {
    customer_id: input.customerId,
    channel: input.channel,
    warehouse_id: input.warehouseId,
    order_date: input.orderDate,
    currency_code: input.currencyCode,
    note: input.note,
    lines: input.lines.map(toApiLineInput)
  };
}

function toApiUpdateInput(input: UpdateSalesOrderInput): UpdateSalesOrderApiRequest {
  return {
    customer_id: input.customerId,
    channel: input.channel,
    warehouse_id: input.warehouseId,
    order_date: input.orderDate,
    note: input.note,
    expected_version: input.expectedVersion,
    lines: input.lines?.map((line) => ({ ...toApiLineInput(line), id: line.id, line_no: line.lineNo }))
  };
}

function toApiLineInput(line: SalesOrderLineInput): components["schemas"]["CreateSalesOrderLineRequest"] {
  return {
    item_id: line.itemId,
    ordered_qty: normalizeDecimalInput(line.orderedQty, decimalScales.quantity),
    uom_code: line.uomCode,
    unit_price: normalizeDecimalInput(line.unitPrice, decimalScales.unitPrice),
    currency_code: line.currencyCode ?? "VND",
    line_discount_amount: normalizeDecimalInput(line.lineDiscountAmount ?? "0", decimalScales.money)
  };
}

function getPrototypeSalesOrder(id: string) {
  const order = prototypeSalesOrders.find((candidate) => candidate.id === id);
  if (!order) {
    throw new Error("Sales order not found");
  }

  return cloneSalesOrder(order);
}

function filterPrototypeSalesOrders(query: SalesOrderQuery) {
  const search = query.search?.trim().toLowerCase();

  return prototypeSalesOrders
    .filter((order) => {
      if (search) {
        const haystack = [order.orderNo, order.customerCode, order.customerName, order.channel]
          .join(" ")
          .toLowerCase();
        if (!haystack.includes(search)) {
          return false;
        }
      }
      if (query.status && order.status !== query.status) {
        return false;
      }
      if (query.channel && order.channel !== query.channel) {
        return false;
      }
      if (query.customerId && order.customerId !== query.customerId) {
        return false;
      }
      if (query.warehouseId && order.warehouseId !== query.warehouseId) {
        return false;
      }

      return true;
    })
    .map(cloneSalesOrder);
}

function createPrototypeSalesOrder(input: CreateSalesOrderInput): SalesOrder {
  if (input.lines.length === 0) {
    throw new Error("At least one line item is required");
  }
  orderSequence += 1;
  const customer = findCustomer(input.customerId);
  const warehouse = input.warehouseId ? findWarehouse(input.warehouseId) : undefined;
  const id = input.id ?? `so-ui-${orderSequence}`;
  const orderNo = input.orderNo ?? `SO-260428-${String(orderSequence).padStart(4, "0")}`;
  const lines = input.lines.map((line, index) => hydrateLineInput(line, id, index));
  const amounts = calculateAmounts(lines);
  const order: SalesOrder = {
    id,
    orderNo,
    customerId: customer.value,
    customerCode: customer.code,
    customerName: customer.label,
    channel: input.channel,
    warehouseId: warehouse?.value,
    warehouseCode: warehouse?.code,
    orderDate: input.orderDate,
    status: "draft",
    currencyCode: input.currencyCode,
    subtotalAmount: amounts.subtotalAmount,
    discountAmount: amounts.discountAmount,
    taxAmount: "0.00",
    shippingFeeAmount: "0.00",
    netAmount: amounts.netAmount,
    totalAmount: amounts.netAmount,
    note: input.note,
    lines,
    auditLogId: `audit-${id}`,
    createdAt: prototypeNow,
    updatedAt: prototypeNow,
    version: 1
  };
  prototypeSalesOrders = [order, ...prototypeSalesOrders.filter((candidate) => candidate.id !== id)];

  return cloneSalesOrder(order);
}

function updatePrototypeSalesOrder(id: string, input: UpdateSalesOrderInput): SalesOrder {
  const current = getPrototypeSalesOrder(id);
  if (current.status !== "draft") {
    throw new Error("Only draft sales orders can be updated");
  }
  if (input.expectedVersion && input.expectedVersion !== current.version) {
    throw new Error("Sales order version changed");
  }
  const customer = input.customerId ? findCustomer(input.customerId) : undefined;
  const warehouse = input.warehouseId ? findWarehouse(input.warehouseId) : undefined;
  const lines = input.lines ? input.lines.map((line, index) => hydrateLineInput(line, id, index)) : current.lines;
  const amounts = calculateAmounts(lines);
  const updated: SalesOrder = {
    ...current,
    customerId: customer?.value ?? current.customerId,
    customerCode: customer?.code ?? current.customerCode,
    customerName: customer?.label ?? current.customerName,
    channel: input.channel ?? current.channel,
    warehouseId: warehouse?.value ?? current.warehouseId,
    warehouseCode: warehouse?.code ?? current.warehouseCode,
    orderDate: input.orderDate ?? current.orderDate,
    note: input.note ?? current.note,
    lines,
    subtotalAmount: amounts.subtotalAmount,
    discountAmount: amounts.discountAmount,
    netAmount: amounts.netAmount,
    totalAmount: amounts.netAmount,
    updatedAt: prototypeNow,
    version: current.version + 1
  };
  prototypeSalesOrders = [updated, ...prototypeSalesOrders.filter((candidate) => candidate.id !== id)];

  return cloneSalesOrder(updated);
}

function transitionPrototypeSalesOrder(
  id: string,
  status: Extract<SalesOrderStatus, "confirmed" | "cancelled">,
  expectedVersion?: number,
  reason?: string
): SalesOrderActionResult {
  const current = getPrototypeSalesOrder(id);
  if (expectedVersion && expectedVersion !== current.version) {
    throw new Error("Sales order version changed");
  }
  if (status === "confirmed" && current.status !== "draft") {
    throw new Error("Only draft sales orders can be confirmed");
  }
  if (status === "cancelled" && !["draft", "confirmed"].includes(current.status)) {
    throw new Error("Only draft or confirmed sales orders can be cancelled");
  }
  const updated: SalesOrder = {
    ...current,
    status,
    updatedAt: prototypeNow,
    confirmedAt: status === "confirmed" ? prototypeNow : current.confirmedAt,
    cancelledAt: status === "cancelled" ? prototypeNow : current.cancelledAt,
    cancelReason: status === "cancelled" ? reason : current.cancelReason,
    version: current.version + 1
  };
  prototypeSalesOrders = [updated, ...prototypeSalesOrders.filter((candidate) => candidate.id !== id)];

  return {
    salesOrder: cloneSalesOrder(updated),
    previousStatus: current.status,
    currentStatus: status,
    auditLogId: `audit-${id}-${status}`
  };
}

function hydrateLineInput(input: SalesOrderLineInput, orderId: string, index: number): SalesOrderLine {
  const item = findItem(input.itemId);
  const orderedQty = normalizeDecimalInput(input.orderedQty, decimalScales.quantity);
  const unitPrice = normalizeDecimalInput(input.unitPrice, decimalScales.unitPrice);
  const lineDiscountAmount = normalizeDecimalInput(input.lineDiscountAmount ?? "0", decimalScales.money);
  const lineAmount = calculateLineAmount(orderedQty, unitPrice, lineDiscountAmount);

  return {
    id: input.id ?? `${orderId}-line-${String(index + 1).padStart(2, "0")}`,
    lineNo: input.lineNo ?? index + 1,
    itemId: item.value,
    skuCode: item.skuCode,
    itemName: item.itemName,
    orderedQty,
    uomCode: item.baseUomCode,
    baseOrderedQty: orderedQty,
    baseUomCode: item.baseUomCode,
    conversionFactor: "1.000000",
    unitPrice,
    currencyCode: input.currencyCode ?? "VND",
    lineDiscountAmount,
    lineAmount,
    reservedQty: "0.000000",
    shippedQty: "0.000000"
  };
}

function calculateAmounts(lines: SalesOrderLine[]) {
  const subtotal = lines.reduce(
    (sum, line) => addMoney(sum, addMoney(line.lineAmount, line.lineDiscountAmount)),
    "0.00"
  );
  const discount = lines.reduce((sum, line) => addMoney(sum, line.lineDiscountAmount), "0.00");
  const net = lines.reduce((sum, line) => addMoney(sum, line.lineAmount), "0.00");

  return {
    subtotalAmount: subtotal,
    discountAmount: discount,
    netAmount: net
  };
}

function calculateLineAmount(orderedQty: string, unitPrice: string, discount: string) {
  const quantity = toScaledBigInt(orderedQty, decimalScales.quantity);
  const price = toScaledBigInt(unitPrice, decimalScales.unitPrice);
  const gross = roundScaled(quantity * price, decimalScales.quantity + decimalScales.unitPrice, decimalScales.money);
  const amount = gross - toScaledBigInt(discount, decimalScales.money);
  if (amount < BigInt(0)) {
    throw new Error("Line discount cannot exceed gross amount");
  }

  return fromScaledBigInt(amount, decimalScales.money);
}

function addMoney(left: string, right: string) {
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

function findCustomer(id: string) {
  const customer = salesCustomerOptions.find((candidate) => candidate.value === id);
  if (!customer) {
    throw new Error("Customer is required");
  }

  return customer;
}

function findWarehouse(id: string) {
  const warehouse = salesWarehouseOptions.find((candidate) => candidate.value === id);
  if (!warehouse) {
    throw new Error("Warehouse is required");
  }

  return warehouse;
}

function findItem(id: string) {
  const item = salesItemOptions.find((candidate) => candidate.value === id);
  if (!item) {
    throw new Error("Item is required");
  }

  return item;
}

function cloneSalesOrder(order: SalesOrder): SalesOrder {
  return {
    ...order,
    lines: order.lines.map((line) => ({ ...line }))
  };
}

function createPrototypeSalesOrders(): SalesOrder[] {
  const draft = createOrderSeed({
    id: "so-260428-0001",
    orderNo: "SO-260428-0001",
    customerId: "cus-dl-minh-anh",
    channel: "B2B",
    status: "draft",
    lines: [
      { itemId: "item-serum-30ml", orderedQty: "12", uomCode: "EA", unitPrice: "125000" },
      {
        itemId: "item-cream-50g",
        orderedQty: "6",
        uomCode: "EA",
        unitPrice: "95000",
        lineDiscountAmount: "50000"
      }
    ]
  });
  const confirmed = createOrderSeed({
    id: "so-260428-0002",
    orderNo: "SO-260428-0002",
    customerId: "cus-mp-shopee",
    channel: "MP",
    status: "confirmed",
    lines: [{ itemId: "item-toner-100ml", orderedQty: "24", uomCode: "EA", unitPrice: "82000" }],
    confirmedAt: "2026-04-28T10:20:00Z"
  });
  const reserved = createOrderSeed({
    id: "so-260428-0003",
    orderNo: "SO-260428-0003",
    customerId: "cus-dealer-linh-chi",
    channel: "DEALER",
    status: "reserved",
    lines: [{ itemId: "item-cream-50g", orderedQty: "8", uomCode: "EA", unitPrice: "95000" }]
  });

  return [draft, confirmed, reserved];
}

function createOrderSeed(input: {
  id: string;
  orderNo: string;
  customerId: string;
  channel: string;
  status: SalesOrderStatus;
  lines: SalesOrderLineInput[];
  confirmedAt?: string;
}) {
  const customer = findCustomer(input.customerId);
  const warehouse = salesWarehouseOptions[0];
  const lines = input.lines.map((line, index) => hydrateLineInput(line, input.id, index));
  const amounts = calculateAmounts(lines);

  return {
    id: input.id,
    orderNo: input.orderNo,
    customerId: customer.value,
    customerCode: customer.code,
    customerName: customer.label,
    channel: input.channel,
    warehouseId: warehouse.value,
    warehouseCode: warehouse.code,
    orderDate: "2026-04-28",
    status: input.status,
    currencyCode: "VND",
    subtotalAmount: amounts.subtotalAmount,
    discountAmount: amounts.discountAmount,
    taxAmount: "0.00",
    shippingFeeAmount: "0.00",
    netAmount: amounts.netAmount,
    totalAmount: amounts.netAmount,
    lines,
    auditLogId: `audit-${input.id}`,
    createdAt: "2026-04-28T09:00:00Z",
    updatedAt: input.confirmedAt ?? "2026-04-28T09:00:00Z",
    confirmedAt: input.confirmedAt,
    version: input.status === "draft" ? 1 : 2
  } satisfies SalesOrder;
}
