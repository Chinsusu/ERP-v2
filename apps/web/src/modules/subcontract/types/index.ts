import type { AuditLogItem } from "@/modules/audit/types";

export type SubcontractOrderStatus =
  | "DRAFT"
  | "CONFIRMED"
  | "MATERIAL_TRANSFERRED"
  | "SAMPLE_APPROVED"
  | "IN_PRODUCTION"
  | "DELIVERED"
  | "QC_REVIEW"
  | "ACCEPTED"
  | "REJECTED"
  | "CLOSED";

export type SubcontractDepositStatus = "not_required" | "pending" | "paid";

export type SubcontractFinalPaymentStatus = "pending" | "hold" | "released";

export type SubcontractFactory = {
  id: string;
  code: string;
  name: string;
};

export type SubcontractProduct = {
  id: string;
  sku: string;
  name: string;
};

export type SubcontractOrder = {
  id: string;
  orderNo: string;
  factoryId: string;
  factoryCode: string;
  factoryName: string;
  productId: string;
  sku: string;
  productName: string;
  quantity: number;
  specVersion: string;
  sampleRequired: boolean;
  expectedDeliveryDate: string;
  depositStatus: SubcontractDepositStatus;
  depositAmount?: number;
  finalPaymentStatus: SubcontractFinalPaymentStatus;
  status: SubcontractOrderStatus;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
  auditLogIds: string[];
};

export type CreateSubcontractOrderInput = {
  factoryId: string;
  factoryName?: string;
  productId: string;
  productName?: string;
  quantity: number;
  specVersion: string;
  sampleRequired: boolean;
  expectedDeliveryDate: string;
  depositStatus: SubcontractDepositStatus;
  depositAmount?: number;
  createdBy?: string;
};

export type SubcontractOrderQuery = {
  factoryId?: string;
  productId?: string;
  status?: SubcontractOrderStatus;
};

export type ChangeSubcontractOrderStatusInput = {
  order: SubcontractOrder;
  nextStatus: SubcontractOrderStatus;
  actorId?: string;
  actorName?: string;
  note?: string;
};

export type SubcontractStatusChangeResult = {
  order: SubcontractOrder;
  auditLog: AuditLogItem;
};

export type SubcontractOrderSummary = {
  total: number;
  draft: number;
  confirmed: number;
  active: number;
  accepted: number;
  rejected: number;
  closed: number;
  nextDeliveryDate?: string;
};
