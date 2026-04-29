import type { AuditLogItem } from "@/modules/audit/types";

export type SubcontractOrderStatus =
  | "draft"
  | "submitted"
  | "approved"
  | "factory_confirmed"
  | "deposit_recorded"
  | "materials_issued_to_factory"
  | "sample_submitted"
  | "sample_approved"
  | "sample_rejected"
  | "mass_production_started"
  | "finished_goods_received"
  | "qc_in_progress"
  | "accepted"
  | "rejected_with_factory_issue"
  | "final_payment_ready"
  | "closed"
  | "cancelled";

export type SubcontractDepositStatus = "not_required" | "pending" | "paid";

export type SubcontractFinalPaymentStatus = "pending" | "hold" | "released";

export type SubcontractTransferStatus = "DRAFT" | "READY_TO_SEND" | "SENT";

export type SubcontractMaterialItemType = "raw_material" | "packaging";

export type SubcontractMaterialQcStatus = "passed" | "pending" | "failed";

export type SubcontractTransferAttachmentType = "COA" | "MSDS" | "LABEL" | "VAT_INVOICE";

export type SubcontractFinishedGoodsReceiptStatus = "qc_hold";

export type SubcontractFinishedGoodsPackagingStatus = "intact" | "damaged" | "mixed";

export type SubcontractFactoryClaimStatus = "open" | "acknowledged" | "resolved" | "closed" | "cancelled";

export type SubcontractFactoryClaimSeverity = "P1" | "P2" | "P3";

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
  receivedQty?: string;
  acceptedQty?: string;
  rejectedQty?: string;
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
  version: number;
  estimatedCostAmount: string;
  materialLines: SubcontractOrderMaterialLine[];
  auditLogIds: string[];
  sampleRejectReason?: string;
};

export type SubcontractOrderMaterialLine = {
  id: string;
  itemId: string;
  skuCode: string;
  itemName: string;
  plannedQty: string;
  issuedQty: string;
  uomCode: string;
  unitCost: string;
  currencyCode: string;
  lineCostAmount: string;
  lotTraceRequired: boolean;
  note?: string;
};

export type CreateSubcontractOrderInput = {
  id?: string;
  orderNo?: string;
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
  materialItemId: string;
  materialQty: string;
  materialUnitCost: string;
  materialLotTraceRequired?: boolean;
  createdBy?: string;
};

export type UpdateSubcontractOrderInput = Partial<CreateSubcontractOrderInput> & {
  expectedVersion: number;
};

export type SubcontractOrderQuery = {
  search?: string;
  factoryId?: string;
  productId?: string;
  status?: SubcontractOrderStatus;
  expectedReceiptFrom?: string;
  expectedReceiptTo?: string;
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
  auditLogId?: string;
};

export type IssueSubcontractMaterialLineInput = {
  orderMaterialLineId: string;
  issueQty: string;
  uomCode: string;
  baseIssueQty?: string;
  baseUOMCode?: string;
  conversionFactor?: string;
  batchId?: string;
  batchNo?: string;
  lotNo?: string;
  sourceBinId?: string;
  note?: string;
};

export type IssueSubcontractMaterialsEvidenceInput = {
  id?: string;
  evidenceType: "handover" | "coa" | "msds" | "label" | "vat_invoice";
  fileName?: string;
  objectKey?: string;
  externalURL?: string;
  note?: string;
};

export type IssueSubcontractMaterialsInput = {
  order: SubcontractOrder;
  sourceWarehouseId: string;
  sourceWarehouseCode: string;
  handoverBy: string;
  handoverAt?: string;
  receivedBy: string;
  receiverContact?: string;
  vehicleNo?: string;
  note?: string;
  lines: IssueSubcontractMaterialLineInput[];
  evidence?: IssueSubcontractMaterialsEvidenceInput[];
};

export type IssueSubcontractMaterialsResult = {
  order: SubcontractOrder;
  transfer: SubcontractMaterialTransfer;
  stockMovements: SubcontractStockMovement[];
  auditLog: AuditLogItem;
  auditLogId?: string;
};

export type SubcontractSampleApprovalStatus = "submitted" | "approved" | "rejected";

export type SubcontractSampleEvidenceInput = {
  id?: string;
  evidenceType: "photo" | "coa" | "spec_sheet" | "label" | "decision_note";
  fileName?: string;
  objectKey?: string;
  externalURL?: string;
  note?: string;
};

export type SubmitSubcontractSampleInput = {
  order: SubcontractOrder;
  sampleApprovalId?: string;
  sampleCode: string;
  formulaVersion?: string;
  specVersion?: string;
  submittedBy: string;
  submittedAt?: string;
  note?: string;
  evidence: SubcontractSampleEvidenceInput[];
};

export type DecideSubcontractSampleInput = {
  order: SubcontractOrder;
  sampleApprovalId?: string;
  reason: string;
  storageStatus?: string;
  decisionAt?: string;
};

export type SubcontractSampleEvidence = {
  id: string;
  evidenceType: string;
  fileName?: string;
  objectKey?: string;
  externalURL?: string;
  note?: string;
  createdAt: string;
  createdBy: string;
};

export type SubcontractSampleApproval = {
  id: string;
  orderId: string;
  orderNo: string;
  sampleCode: string;
  formulaVersion?: string;
  specVersion?: string;
  status: SubcontractSampleApprovalStatus;
  evidence: SubcontractSampleEvidence[];
  submittedBy: string;
  submittedAt: string;
  decisionBy?: string;
  decisionAt?: string;
  decisionReason?: string;
  storageStatus?: string;
  note?: string;
  createdAt: string;
  updatedAt: string;
  version: number;
};

export type SubcontractSampleApprovalResult = {
  order: SubcontractOrder;
  sampleApproval: SubcontractSampleApproval;
  auditLog: AuditLogItem;
  auditLogId?: string;
};

export type ReceiveSubcontractFinishedGoodsLineInput = {
  id?: string;
  lineNo?: number;
  itemId?: string;
  skuCode?: string;
  itemName?: string;
  batchId?: string;
  batchNo: string;
  lotNo?: string;
  expiryDate: string;
  receiveQty: string;
  uomCode: string;
  baseReceiveQty?: string;
  baseUOMCode?: string;
  conversionFactor?: string;
  packagingStatus: SubcontractFinishedGoodsPackagingStatus;
  note?: string;
};

export type ReceiveSubcontractFinishedGoodsEvidenceInput = {
  id?: string;
  evidenceType: "delivery_note" | "packing_list" | "coa" | "photo";
  fileName?: string;
  objectKey?: string;
  externalURL?: string;
  note?: string;
};

export type ReceiveSubcontractFinishedGoodsInput = {
  order: SubcontractOrder;
  receiptId?: string;
  receiptNo?: string;
  warehouseId: string;
  warehouseCode: string;
  locationId: string;
  locationCode: string;
  deliveryNoteNo: string;
  receivedBy: string;
  receivedAt?: string;
  note?: string;
  lines: ReceiveSubcontractFinishedGoodsLineInput[];
  evidence?: ReceiveSubcontractFinishedGoodsEvidenceInput[];
};

export type SubcontractFinishedGoodsReceiptLine = {
  id: string;
  lineNo: number;
  itemId: string;
  skuCode: string;
  itemName: string;
  batchId?: string;
  batchNo: string;
  lotNo?: string;
  expiryDate: string;
  receiveQty: string;
  uomCode: string;
  baseReceiveQty: string;
  baseUOMCode: string;
  conversionFactor: string;
  packagingStatus?: string;
  note?: string;
};

export type SubcontractFinishedGoodsReceiptEvidence = {
  id: string;
  evidenceType: string;
  fileName?: string;
  objectKey?: string;
  externalURL?: string;
  note?: string;
};

export type SubcontractFinishedGoodsReceipt = {
  id: string;
  receiptNo: string;
  orderId: string;
  orderNo: string;
  factoryId: string;
  factoryCode?: string;
  factoryName: string;
  warehouseId: string;
  warehouseCode: string;
  locationId: string;
  locationCode: string;
  deliveryNoteNo: string;
  status: SubcontractFinishedGoodsReceiptStatus;
  lines: SubcontractFinishedGoodsReceiptLine[];
  evidence: SubcontractFinishedGoodsReceiptEvidence[];
  receivedBy: string;
  receivedAt: string;
  note?: string;
  createdAt: string;
  updatedAt: string;
  version: number;
};

export type ReceiveSubcontractFinishedGoodsResult = {
  order: SubcontractOrder;
  receipt: SubcontractFinishedGoodsReceipt;
  stockMovements: SubcontractStockMovement[];
  auditLog: AuditLogItem;
  auditLogId?: string;
};

export type SubcontractFactoryClaimEvidenceInput = {
  id?: string;
  evidenceType: "qc_photo" | "inspection_note" | "delivery_note" | "coa";
  fileName?: string;
  objectKey?: string;
  externalURL?: string;
  note?: string;
};

export type CreateSubcontractFactoryClaimInput = {
  order: SubcontractOrder;
  claimId?: string;
  claimNo?: string;
  receiptId?: string;
  receiptNo?: string;
  reasonCode: string;
  reason: string;
  severity: SubcontractFactoryClaimSeverity;
  affectedQty: string;
  uomCode: string;
  baseAffectedQty?: string;
  baseUOMCode?: string;
  ownerId: string;
  openedBy: string;
  openedAt?: string;
  evidence: SubcontractFactoryClaimEvidenceInput[];
};

export type SubcontractFactoryClaimEvidence = {
  id: string;
  evidenceType: string;
  fileName?: string;
  objectKey?: string;
  externalURL?: string;
  note?: string;
  createdAt: string;
  createdBy: string;
};

export type SubcontractFactoryClaim = {
  id: string;
  claimNo: string;
  orderId: string;
  orderNo: string;
  factoryId: string;
  factoryCode: string;
  factoryName: string;
  receiptId?: string;
  receiptNo?: string;
  reasonCode: string;
  reason: string;
  severity: SubcontractFactoryClaimSeverity;
  status: SubcontractFactoryClaimStatus;
  affectedQty: string;
  uomCode: string;
  baseAffectedQty: string;
  baseUOMCode: string;
  evidence: SubcontractFactoryClaimEvidence[];
  ownerId: string;
  openedBy: string;
  openedAt: string;
  dueAt: string;
  resolutionNote?: string;
  blocksFinalPayment: boolean;
  createdAt: string;
  updatedAt: string;
  version: number;
};

export type SubcontractFactoryClaimResult = {
  order: SubcontractOrder;
  claim: SubcontractFactoryClaim;
  auditLog: AuditLogItem;
  auditLogId?: string;
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
  movementType: "SUBCONTRACT_ISSUE" | "SUBCONTRACT_RECEIPT";
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
