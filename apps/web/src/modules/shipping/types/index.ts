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
