import type { AuditLogItem } from "@/modules/audit/types";
import type {
  ChangeSubcontractOrderStatusInput,
  CreateSubcontractOrderInput,
  SubcontractDepositStatus,
  SubcontractFactory,
  SubcontractFinalPaymentStatus,
  SubcontractOrder,
  SubcontractOrderQuery,
  SubcontractOrderStatus,
  SubcontractOrderSummary,
  SubcontractProduct,
  SubcontractStatusChangeResult
} from "../types";

export const subcontractFactoryOptions: SubcontractFactory[] = [
  { id: "factory-lotus", code: "LOTUS", name: "Lotus GMP Factory" },
  { id: "factory-saigon-lab", code: "SGLAB", name: "Saigon Cosmetic Lab" },
  { id: "factory-north-pack", code: "NPK", name: "North Pack Partner" }
];

export const subcontractProductOptions: SubcontractProduct[] = [
  { id: "product-repair-cream-50ml", sku: "CREAM-50ML", name: "Repair Cream 50ml" },
  { id: "product-hydrating-serum-30ml", sku: "SERUM-30ML", name: "Hydrating Serum 30ml" },
  { id: "product-balancing-toner-100ml", sku: "TONER-100ML", name: "Balancing Toner 100ml" }
];

export const subcontractDepositStatusOptions: { label: string; value: SubcontractDepositStatus }[] = [
  { label: "Not required", value: "not_required" },
  { label: "Pending", value: "pending" },
  { label: "Paid", value: "paid" }
];

export const subcontractOrderStatusOptions: { label: string; value: SubcontractOrderStatus }[] = [
  { label: "Draft", value: "DRAFT" },
  { label: "Confirmed", value: "CONFIRMED" },
  { label: "Material transferred", value: "MATERIAL_TRANSFERRED" },
  { label: "Sample approved", value: "SAMPLE_APPROVED" },
  { label: "In production", value: "IN_PRODUCTION" },
  { label: "Delivered", value: "DELIVERED" },
  { label: "QC review", value: "QC_REVIEW" },
  { label: "Accepted", value: "ACCEPTED" },
  { label: "Rejected", value: "REJECTED" },
  { label: "Closed", value: "CLOSED" }
];

export const prototypeSubcontractOrders: SubcontractOrder[] = [
  createSubcontractOrderRecord({
    id: "sub-order-260426-0001",
    orderNo: "SUB-260426-0001",
    factoryId: "factory-lotus",
    factoryCode: "LOTUS",
    factoryName: "Lotus GMP Factory",
    productId: "product-repair-cream-50ml",
    sku: "CREAM-50ML",
    productName: "Repair Cream 50ml",
    quantity: 5000,
    specVersion: "SPEC-CREAM-50ML-v3",
    sampleRequired: true,
    expectedDeliveryDate: "2026-05-12",
    depositStatus: "pending",
    depositAmount: 12000000,
    finalPaymentStatus: "hold",
    status: "CONFIRMED",
    createdBy: "Subcontract Coordinator",
    createdAt: "2026-04-26T08:00:00Z",
    updatedAt: "2026-04-26T08:15:00Z",
    auditLogIds: ["audit-sub-order-260426-0001-confirmed"]
  })
];

let subcontractOrderSequence = 2;
let subcontractAuditSequence = 1;

export async function getSubcontractOrders(query: SubcontractOrderQuery = {}): Promise<SubcontractOrder[]> {
  return prototypeSubcontractOrders
    .filter((order) => {
      if (query.factoryId && order.factoryId !== query.factoryId) {
        return false;
      }
      if (query.productId && order.productId !== query.productId) {
        return false;
      }
      if (query.status && order.status !== query.status) {
        return false;
      }

      return true;
    })
    .sort(sortSubcontractOrders);
}

export function createSubcontractOrder(input: CreateSubcontractOrderInput): SubcontractOrder {
  const factory = resolveFactory(input.factoryId, input.factoryName);
  const product = resolveProduct(input.productId, input.productName);
  const quantity = normalizeQuantity(input.quantity);
  const expectedDeliveryDate = normalizeRequiredText(input.expectedDeliveryDate, "Expected delivery date is required");
  const specVersion = normalizeRequiredText(input.specVersion, "Spec version is required");
  const depositStatus = normalizeDepositStatus(input.depositStatus);
  const sequence = subcontractOrderSequence++;
  const id = `sub-order-260426-${String(sequence).padStart(4, "0")}`;

  return createSubcontractOrderRecord({
    id,
    orderNo: `SUB-260426-${String(sequence).padStart(4, "0")}`,
    factoryId: factory.id,
    factoryCode: factory.code,
    factoryName: factory.name,
    productId: product.id,
    sku: product.sku,
    productName: product.name,
    quantity,
    specVersion,
    sampleRequired: input.sampleRequired,
    expectedDeliveryDate,
    depositStatus,
    depositAmount: depositStatus === "not_required" ? undefined : normalizeOptionalAmount(input.depositAmount),
    finalPaymentStatus: finalPaymentStatusForDeposit(depositStatus),
    status: "DRAFT",
    createdBy: input.createdBy?.trim() || "Subcontract Coordinator",
    createdAt: "2026-04-26T12:00:00Z",
    updatedAt: "2026-04-26T12:00:00Z",
    auditLogIds: []
  });
}

export function changeSubcontractOrderStatus(
  input: ChangeSubcontractOrderStatusInput
): SubcontractStatusChangeResult {
  const nextStatus = normalizeOrderStatus(input.nextStatus);
  const beforeStatus = input.order.status;
  const auditLog = createStatusAuditLog(input.order, beforeStatus, nextStatus, input);
  const order = createSubcontractOrderRecord({
    ...input.order,
    status: nextStatus,
    updatedAt: "2026-04-26T12:15:00Z",
    auditLogIds: [...input.order.auditLogIds, auditLog.id]
  });

  return { order, auditLog };
}

export function summarizeSubcontractOrders(orders: SubcontractOrder[]): SubcontractOrderSummary {
  const sortedDeliveryDates = orders
    .map((order) => order.expectedDeliveryDate)
    .filter((date) => date.trim() !== "")
    .sort();

  return {
    total: orders.length,
    draft: orders.filter((order) => order.status === "DRAFT").length,
    confirmed: orders.filter((order) => order.status === "CONFIRMED").length,
    active: orders.filter((order) => isActiveOrderStatus(order.status)).length,
    accepted: orders.filter((order) => order.status === "ACCEPTED").length,
    rejected: orders.filter((order) => order.status === "REJECTED").length,
    closed: orders.filter((order) => order.status === "CLOSED").length,
    nextDeliveryDate: sortedDeliveryDates[0]
  };
}

export function subcontractOrderStatusTone(
  status: SubcontractOrderStatus
): "normal" | "success" | "warning" | "danger" | "info" {
  switch (status) {
    case "ACCEPTED":
    case "CLOSED":
      return "success";
    case "REJECTED":
      return "danger";
    case "DRAFT":
    case "QC_REVIEW":
      return "warning";
    case "CONFIRMED":
    case "MATERIAL_TRANSFERRED":
    case "SAMPLE_APPROVED":
    case "IN_PRODUCTION":
    case "DELIVERED":
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

function createSubcontractOrderRecord(input: SubcontractOrder): SubcontractOrder {
  return {
    ...input,
    auditLogIds: [...input.auditLogIds]
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
      note: input.note?.trim() || "Status changed from subcontract order skeleton",
      actor_name: input.actorName ?? "Subcontract Coordinator"
    },
    createdAt: "2026-04-26T12:15:00Z"
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
  const normalizedProductId = normalizeRequiredText(productId, "Product is required");
  const matchedProduct = subcontractProductOptions.find((product) => product.id === normalizedProductId);

  return (
    matchedProduct ?? {
      id: normalizedProductId,
      sku: normalizedProductId.toUpperCase(),
      name: productName?.trim() || normalizedProductId
    }
  );
}

function normalizeQuantity(quantity: number) {
  if (!Number.isFinite(quantity) || quantity <= 0) {
    throw new Error("Order quantity must be greater than zero");
  }

  return Math.round(quantity * 1000) / 1000;
}

function normalizeOptionalAmount(amount: number | undefined) {
  if (amount === undefined || amount <= 0 || !Number.isFinite(amount)) {
    return undefined;
  }

  return Math.round(amount);
}

function normalizeRequiredText(value: string, message: string) {
  const normalized = value.trim();
  if (normalized === "") {
    throw new Error(message);
  }

  return normalized;
}

function normalizeDepositStatus(status: SubcontractDepositStatus): SubcontractDepositStatus {
  if (subcontractDepositStatusOptions.some((option) => option.value === status)) {
    return status;
  }

  return "pending";
}

function normalizeOrderStatus(status: SubcontractOrderStatus): SubcontractOrderStatus {
  if (subcontractOrderStatusOptions.some((option) => option.value === status)) {
    return status;
  }

  return "DRAFT";
}

function finalPaymentStatusForDeposit(status: SubcontractDepositStatus): SubcontractFinalPaymentStatus {
  return status === "paid" ? "pending" : "hold";
}

function isActiveOrderStatus(status: SubcontractOrderStatus) {
  return !["DRAFT", "ACCEPTED", "REJECTED", "CLOSED"].includes(status);
}

function sortSubcontractOrders(left: SubcontractOrder, right: SubcontractOrder) {
  if (left.expectedDeliveryDate !== right.expectedDeliveryDate) {
    return left.expectedDeliveryDate.localeCompare(right.expectedDeliveryDate);
  }

  return right.orderNo.localeCompare(left.orderNo);
}
