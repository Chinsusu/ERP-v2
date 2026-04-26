export type CarrierManifestStatus = "draft" | "ready" | "scanning" | "completed" | "exception";

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
  owner?: string;
};
