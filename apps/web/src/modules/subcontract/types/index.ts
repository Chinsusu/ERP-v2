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

export type SubcontractTransferStatus = "DRAFT" | "READY_TO_SEND" | "SENT";

export type SubcontractMaterialItemType = "raw_material" | "packaging";

export type SubcontractMaterialQcStatus = "passed" | "pending" | "failed";

export type SubcontractTransferAttachmentType = "COA" | "MSDS" | "LABEL" | "VAT_INVOICE";

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

export type SubcontractMaterialTransferLine = {
  id: string;
  itemCode: string;
  itemName: string;
  itemType: SubcontractMaterialItemType;
  quantity: number;
  unit: string;
  lotControlled: boolean;
  batchNo?: string;
  qcStatus: SubcontractMaterialQcStatus;
};

export type SubcontractTransferAttachmentPlaceholder = {
  type: SubcontractTransferAttachmentType;
  label: string;
  required: boolean;
  attached: boolean;
};

export type SubcontractStockMovement = {
  id: string;
  movementType: "SUBCONTRACT_ISSUE";
  itemCode: string;
  quantity: number;
  unit: string;
  sourceWarehouseId: string;
  targetLocation: string;
  batchNo?: string;
  sourceDocId: string;
};

export type SubcontractMaterialTransfer = {
  id: string;
  transferNo: string;
  orderId: string;
  orderNo: string;
  sourceWarehouseId: string;
  sourceWarehouseCode: string;
  factoryId: string;
  factoryName: string;
  signedHandover: boolean;
  status: SubcontractTransferStatus;
  attachmentPlaceholders: SubcontractTransferAttachmentPlaceholder[];
  lines: SubcontractMaterialTransferLine[];
  stockMovements: SubcontractStockMovement[];
  createdBy: string;
  createdAt: string;
};

export type CreateSubcontractMaterialTransferInput = {
  order: SubcontractOrder;
  sourceWarehouseId: string;
  sourceWarehouseCode: string;
  signedHandover: boolean;
  lines: SubcontractMaterialTransferLine[];
  createdBy?: string;
};

export type SubcontractMaterialTransferSummary = {
  total: number;
  signed: number;
  movementCount: number;
  attachmentPlaceholderCount: number;
};
