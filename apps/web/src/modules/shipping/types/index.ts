export type CarrierManifestStatus =
  | "draft"
  | "ready"
  | "scanning"
  | "completed"
  | "handed_over"
  | "exception"
  | "cancelled";

export type CarrierManifestSummary = {
  expectedCount: number;
  scannedCount: number;
  missingCount: number;
};

export type CarrierManifestLine = {
  id: string;
  shipmentId: string;
  orderNo: string;
  trackingNo: string;
  packageCode: string;
  stagingZone: string;
  handoverZoneCode?: string;
  handoverBinCode?: string;
  scanned: boolean;
};

export type CarrierManifest = {
  id: string;
  carrierCode: string;
  carrierName: string;
  warehouseId: string;
  warehouseCode: string;
  date: string;
  handoverBatch: string;
  stagingZone: string;
  handoverZoneCode?: string;
  handoverBinCode?: string;
  status: CarrierManifestStatus;
  owner: string;
  auditLogId?: string;
  summary: CarrierManifestSummary;
  lines: CarrierManifestLine[];
  createdAt: string;
};

export type CarrierManifestQuery = {
  warehouseId?: string;
  date?: string;
  carrierCode?: string;
  status?: CarrierManifestStatus;
};

export type CarrierManifestScanResultCode =
  | "MATCHED"
  | "NOT_FOUND"
  | "INVALID_STATE"
  | "MANIFEST_MISMATCH"
  | "DUPLICATE_SCAN";

export type CarrierManifestScanSeverity = "success" | "warning" | "danger";

export type CarrierManifestScanEvent = {
  id: string;
  manifestId: string;
  expectedManifestId?: string;
  code: string;
  resultCode: CarrierManifestScanResultCode;
  severity: CarrierManifestScanSeverity;
  message: string;
  shipmentId?: string;
  orderNo?: string;
  trackingNo?: string;
  actorId: string;
  stationId: string;
  warehouseId: string;
  carrierCode: string;
  createdAt: string;
};

export type CarrierManifestScanResult = {
  resultCode: CarrierManifestScanResultCode;
  severity: CarrierManifestScanSeverity;
  message: string;
  expectedManifestId?: string;
  line?: CarrierManifestLine;
  scanEvent: CarrierManifestScanEvent;
  manifest: CarrierManifest;
  auditLogId?: string;
};

export type VerifyCarrierManifestScanInput = {
  manifestId: string;
  code: string;
  stationId?: string;
};

export type CreateCarrierManifestInput = {
  carrierCode: string;
  carrierName: string;
  warehouseId: string;
  warehouseCode: string;
  date: string;
  handoverBatch?: string;
  stagingZone?: string;
  handoverZoneCode?: string;
  handoverBinCode?: string;
  owner?: string;
};

export type PickTaskStatus =
  | "created"
  | "assigned"
  | "in_progress"
  | "completed"
  | "missing_stock"
  | "wrong_sku"
  | "wrong_batch"
  | "wrong_location"
  | "cancelled";

export type PickTaskLineStatus =
  | "pending"
  | "picked"
  | "missing_stock"
  | "wrong_sku"
  | "wrong_batch"
  | "wrong_location"
  | "cancelled";

export type PickTaskLine = {
  id: string;
  lineNo: number;
  salesOrderLineId: string;
  stockReservationId: string;
  itemId: string;
  skuCode: string;
  batchId: string;
  batchNo: string;
  warehouseId: string;
  binId: string;
  binCode: string;
  qtyToPick: string;
  qtyPicked: string;
  baseUOMCode: string;
  status: PickTaskLineStatus;
  pickedAt?: string;
  pickedBy?: string;
  createdAt: string;
  updatedAt: string;
};

export type PickTask = {
  id: string;
  orgId: string;
  pickTaskNo: string;
  salesOrderId: string;
  orderNo: string;
  warehouseId: string;
  warehouseCode: string;
  status: PickTaskStatus;
  assignedTo?: string;
  assignedAt?: string;
  startedAt?: string;
  startedBy?: string;
  completedAt?: string;
  completedBy?: string;
  auditLogId?: string;
  lines: PickTaskLine[];
  createdAt: string;
  updatedAt: string;
};

export type PickTaskQuery = {
  warehouseId?: string;
  status?: PickTaskStatus;
  assignedTo?: string;
};

export type PickTaskExceptionCode = "missing_stock" | "wrong_sku" | "wrong_batch" | "wrong_location" | "cancelled";

export type PackTaskStatus = "created" | "in_progress" | "packed" | "pack_exception" | "cancelled";

export type PackTaskLineStatus = "pending" | "packed" | "pack_exception" | "cancelled";

export type PackTaskLine = {
  id: string;
  lineNo: number;
  pickTaskLineId: string;
  salesOrderLineId: string;
  itemId: string;
  skuCode: string;
  batchId: string;
  batchNo: string;
  warehouseId: string;
  qtyToPack: string;
  qtyPacked: string;
  baseUOMCode: string;
  status: PackTaskLineStatus;
  packedAt?: string;
  packedBy?: string;
  createdAt: string;
  updatedAt: string;
};

export type PackTask = {
  id: string;
  orgId: string;
  packTaskNo: string;
  salesOrderId: string;
  salesOrderStatus?: string;
  orderNo: string;
  pickTaskId: string;
  pickTaskNo: string;
  warehouseId: string;
  warehouseCode: string;
  status: PackTaskStatus;
  assignedTo?: string;
  assignedAt?: string;
  startedAt?: string;
  startedBy?: string;
  packedAt?: string;
  packedBy?: string;
  auditLogId?: string;
  lines: PackTaskLine[];
  createdAt: string;
  updatedAt: string;
};

export type PackTaskQuery = {
  warehouseId?: string;
  status?: PackTaskStatus;
  assignedTo?: string;
};

export type ConfirmPackTaskLineInput = {
  lineId: string;
  packedQty: string;
};

export type PackTaskExceptionCode = "pack_exception";
