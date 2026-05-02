import { apiDelete, apiGetRaw, apiPost } from "../../../shared/api/client";
import { shouldUsePrototypeFallback } from "../../../shared/api/prototypeFallback";
import type {
  CarrierManifest,
  CarrierManifestLine,
  CarrierManifestQuery,
  CarrierManifestScanEvent,
  CarrierManifestScanResult,
  CarrierManifestScanResultCode,
  CarrierManifestScanSeverity,
  CarrierManifestStatus,
  CarrierManifestSummary,
  CreateCarrierManifestInput,
  VerifyCarrierManifestScanInput
} from "../types";

export const defaultCarrierManifestDate = "2026-04-26";

const defaultAccessToken = "local-dev-access-token";

type CarrierManifestSummaryApi = {
  expected_count: number;
  scanned_count: number;
  missing_count: number;
};

type CarrierManifestLineApi = {
  id: string;
  shipment_id: string;
  order_no: string;
  tracking_no: string;
  package_code: string;
  staging_zone: string;
  handover_zone_code?: string;
  handover_bin_code?: string;
  scanned: boolean;
};

type CarrierManifestApi = {
  id: string;
  carrier_code: string;
  carrier_name: string;
  warehouse_id: string;
  warehouse_code: string;
  date: string;
  handover_batch: string;
  staging_zone: string;
  handover_zone_code?: string;
  handover_bin_code?: string;
  status: CarrierManifestStatus;
  owner: string;
  audit_log_id?: string;
  summary: CarrierManifestSummaryApi;
  lines: CarrierManifestLineApi[];
  missing_lines?: CarrierManifestLineApi[];
  created_at?: string;
};

type CreateCarrierManifestApiRequest = {
  carrier_code: string;
  carrier_name: string;
  warehouse_id: string;
  warehouse_code: string;
  date: string;
  handover_batch?: string;
  staging_zone?: string;
  handover_zone_code?: string;
  handover_bin_code?: string;
  owner?: string;
};

type CarrierManifestScanEventApi = {
  id: string;
  manifest_id: string;
  expected_manifest_id?: string;
  code: string;
  result_code: CarrierManifestScanResultCode;
  severity: CarrierManifestScanSeverity;
  message: string;
  shipment_id?: string;
  order_no?: string;
  tracking_no?: string;
  actor_id: string;
  station_id: string;
  device_id?: string;
  source: string;
  warehouse_id: string;
  carrier_code: string;
  created_at: string;
};

type CarrierManifestScanResultApi = {
  result_code: CarrierManifestScanResultCode;
  severity: CarrierManifestScanSeverity;
  message: string;
  expected_manifest_id?: string;
  line?: CarrierManifestLineApi;
  scan_event: CarrierManifestScanEventApi;
  manifest: CarrierManifestApi;
  audit_log_id?: string;
};

export const carrierOptions = [
  { label: "All carriers", value: "" },
  { label: "GHN", value: "GHN" },
  { label: "Viettel Post", value: "VTP" },
  { label: "Ninja Van", value: "NJV" }
] as const;

export const manifestWarehouseOptions = [
  { label: "All warehouses", value: "" },
  { label: "HCM", value: "wh-hcm", code: "HCM" },
  { label: "HN", value: "wh-hn", code: "HN" }
] as const;

export const carrierManifestStatusOptions: { label: string; value: "" | CarrierManifestStatus }[] = [
  { label: "All statuses", value: "" },
  { label: "Draft", value: "draft" },
  { label: "Ready", value: "ready" },
  { label: "Scanning", value: "scanning" },
  { label: "Completed", value: "completed" },
  { label: "Handed over", value: "handed_over" },
  { label: "Exception", value: "exception" },
  { label: "Cancelled", value: "cancelled" }
];

export let prototypeCarrierManifests: CarrierManifest[] = createPrototypeCarrierManifests();

const prototypeShipmentStates = [
  {
    shipmentId: "ship-hcm-260426-004",
    orderNo: "SO-260426-004",
    trackingNo: "GHN260426004",
    carrierCode: "GHN",
    packageCode: "TOTE-A03",
    stagingZone: "handover-a",
    handoverZoneCode: "handover-a",
    handoverBinCode: "A03",
    packed: true
  },
  {
    shipmentId: "ship-hcm-260426-099",
    orderNo: "SO-260426-099",
    trackingNo: "GHN260426099",
    carrierCode: "GHN",
    packageCode: "PACKING-LANE-01",
    stagingZone: "packing",
    packed: false
  },
  {
    shipmentId: "ship-hcm-vtp-260426-002",
    orderNo: "SO-260426-012",
    trackingNo: "VTP260426012",
    carrierCode: "VTP",
    packageCode: "TOTE-B02",
    stagingZone: "handover-b",
    handoverZoneCode: "handover-b",
    handoverBinCode: "B02",
    packed: true
  }
];

let prototypeScanSequence = 0;

export async function getCarrierManifests(query: CarrierManifestQuery = {}): Promise<CarrierManifest[]> {
  try {
    const manifests = await apiGetRaw<CarrierManifestApi[]>(`/shipping/manifests${queryString(toApiQuery(query))}`, {
      accessToken: defaultAccessToken
    });

    return manifests.map(fromApiCarrierManifest).sort(sortManifests);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return filterPrototypeCarrierManifests(query);
  }
}

export async function createCarrierManifest(input: CreateCarrierManifestInput): Promise<CarrierManifest> {
  try {
    return fromApiCarrierManifest(
      await apiPost<CarrierManifestApi, CreateCarrierManifestApiRequest>(
        "/shipping/manifests",
        toApiCreateInput(input),
        { accessToken: defaultAccessToken }
      )
    );
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return createPrototypeCarrierManifest(input);
  }
}

export async function addShipmentToManifest(manifestId: string, shipmentId: string): Promise<CarrierManifest> {
  try {
    return fromApiCarrierManifest(
      await apiPost<CarrierManifestApi, { shipment_id: string }>(
        `/shipping/manifests/${encodeURIComponent(manifestId)}/shipments`,
        { shipment_id: shipmentId },
        { accessToken: defaultAccessToken }
      )
    );
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return addPrototypeShipmentToManifest(manifestId, shipmentId);
  }
}

export async function removeShipmentFromManifest(manifestId: string, shipmentId: string): Promise<CarrierManifest> {
  try {
    return fromApiCarrierManifest(
      await apiDelete<CarrierManifestApi>(
        `/shipping/manifests/${encodeURIComponent(manifestId)}/shipments/${encodeURIComponent(shipmentId)}`,
        { accessToken: defaultAccessToken }
      )
    );
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return removePrototypeShipmentFromManifest(manifestId, shipmentId);
  }
}

export async function markCarrierManifestReady(manifestId: string): Promise<CarrierManifest> {
  try {
    return fromApiCarrierManifest(
      await apiPost<CarrierManifestApi, Record<string, never>>(
        `/shipping/manifests/${encodeURIComponent(manifestId)}/ready`,
        {},
        { accessToken: defaultAccessToken }
      )
    );
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return markPrototypeManifestReady(manifestId);
  }
}

export async function cancelCarrierManifest(manifestId: string, reason = ""): Promise<CarrierManifest> {
  try {
    return fromApiCarrierManifest(
      await apiPost<CarrierManifestApi, { reason: string }>(
        `/shipping/manifests/${encodeURIComponent(manifestId)}/cancel`,
        { reason },
        { accessToken: defaultAccessToken }
      )
    );
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return cancelPrototypeManifest(manifestId);
  }
}

export async function confirmCarrierManifestHandover(manifestId: string): Promise<CarrierManifest> {
  try {
    return fromApiCarrierManifest(
      await apiPost<CarrierManifestApi, Record<string, never>>(
        `/shipping/manifests/${encodeURIComponent(manifestId)}/confirm-handover`,
        {},
        { accessToken: defaultAccessToken }
      )
    );
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return confirmPrototypeManifestHandover(manifestId);
  }
}

export async function verifyCarrierManifestScan(
  input: VerifyCarrierManifestScanInput,
  manifests: CarrierManifest[] = prototypeCarrierManifests
): Promise<CarrierManifestScanResult> {
  try {
    return fromApiCarrierManifestScanResult(
      await apiPost<
        CarrierManifestScanResultApi,
        { code: string; station_id?: string; device_id?: string; source?: string }
      >(
        `/shipping/manifests/${encodeURIComponent(input.manifestId)}/scan`,
        { code: input.code, station_id: input.stationId, device_id: input.deviceId, source: input.source },
        { accessToken: defaultAccessToken }
      )
    );
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return verifyPrototypeCarrierManifestScan(input, manifests);
  }
}

export function summarizeCarrierManifestLines(lines: CarrierManifestLine[]): CarrierManifestSummary {
  const expectedCount = lines.length;
  const scannedCount = lines.filter((line) => line.scanned).length;

  return {
    expectedCount,
    scannedCount,
    missingCount: Math.max(expectedCount - scannedCount, 0)
  };
}

export function carrierManifestStatusTone(
  status: CarrierManifestStatus
): "normal" | "success" | "warning" | "danger" | "info" {
  switch (status) {
    case "completed":
    case "handed_over":
      return "success";
    case "exception":
    case "cancelled":
      return "danger";
    case "scanning":
      return "warning";
    case "ready":
      return "info";
    case "draft":
    default:
      return "normal";
  }
}

export function carrierManifestScanSeverityTone(severity: CarrierManifestScanSeverity) {
  return severity;
}

export function resetPrototypeCarrierManifestsForTest() {
  prototypeCarrierManifests = createPrototypeCarrierManifests();
  prototypeScanSequence = 0;
}


function fromApiCarrierManifest(manifest: CarrierManifestApi): CarrierManifest {
  return createManifest({
    id: manifest.id,
    carrierCode: manifest.carrier_code,
    carrierName: manifest.carrier_name,
    warehouseId: manifest.warehouse_id,
    warehouseCode: manifest.warehouse_code,
    date: manifest.date,
    handoverBatch: manifest.handover_batch,
    stagingZone: manifest.staging_zone,
    handoverZoneCode: manifest.handover_zone_code || manifest.staging_zone,
    handoverBinCode: manifest.handover_bin_code,
    status: manifest.status,
    owner: manifest.owner,
    auditLogId: manifest.audit_log_id,
    lines: manifest.lines.map(fromApiCarrierManifestLine),
    createdAt: manifest.created_at ?? ""
  });
}

function fromApiCarrierManifestLine(line: CarrierManifestLineApi): CarrierManifestLine {
  return {
    id: line.id,
    shipmentId: line.shipment_id,
    orderNo: line.order_no,
    trackingNo: line.tracking_no,
    packageCode: line.package_code,
    stagingZone: line.staging_zone,
    handoverZoneCode: line.handover_zone_code || line.staging_zone,
    handoverBinCode: line.handover_bin_code,
    scanned: line.scanned
  };
}

function fromApiCarrierManifestScanResult(result: CarrierManifestScanResultApi): CarrierManifestScanResult {
  return {
    resultCode: result.result_code,
    severity: result.severity,
    message: result.message,
    expectedManifestId: result.expected_manifest_id,
    line: result.line ? fromApiCarrierManifestLine(result.line) : undefined,
    scanEvent: fromApiCarrierManifestScanEvent(result.scan_event),
    manifest: fromApiCarrierManifest(result.manifest),
    auditLogId: result.audit_log_id
  };
}

function fromApiCarrierManifestScanEvent(event: CarrierManifestScanEventApi): CarrierManifestScanEvent {
  return {
    id: event.id,
    manifestId: event.manifest_id,
    expectedManifestId: event.expected_manifest_id,
    code: event.code,
    resultCode: event.result_code,
    severity: event.severity,
    message: event.message,
    shipmentId: event.shipment_id,
    orderNo: event.order_no,
    trackingNo: event.tracking_no,
    actorId: event.actor_id,
    stationId: event.station_id,
    deviceId: event.device_id,
    source: event.source,
    warehouseId: event.warehouse_id,
    carrierCode: event.carrier_code,
    createdAt: event.created_at
  };
}

function toApiQuery(query: CarrierManifestQuery) {
  return {
    warehouse_id: query.warehouseId,
    date: query.date,
    carrier_code: query.carrierCode,
    status: query.status
  };
}

function toApiCreateInput(input: CreateCarrierManifestInput): CreateCarrierManifestApiRequest {
  return {
    carrier_code: input.carrierCode,
    carrier_name: input.carrierName,
    warehouse_id: input.warehouseId,
    warehouse_code: input.warehouseCode,
    date: input.date,
    handover_batch: input.handoverBatch,
    staging_zone: input.stagingZone,
    handover_zone_code: input.handoverZoneCode,
    handover_bin_code: input.handoverBinCode,
    owner: input.owner
  };
}

function queryString(query: Record<string, string | undefined>) {
  const params = new URLSearchParams();
  Object.entries(query).forEach(([key, value]) => {
    if (value) {
      params.set(key, value);
    }
  });

  const value = params.toString();
  return value ? `?${value}` : "";
}

function filterPrototypeCarrierManifests(query: CarrierManifestQuery) {
  return prototypeCarrierManifests
    .filter((manifest) => {
      if (query.warehouseId && manifest.warehouseId !== query.warehouseId) {
        return false;
      }
      if (query.date && manifest.date !== query.date) {
        return false;
      }
      if (query.carrierCode && manifest.carrierCode !== query.carrierCode) {
        return false;
      }
      if (query.status && manifest.status !== query.status) {
        return false;
      }

      return true;
    })
    .sort(sortManifests)
    .map(cloneManifest);
}

function createPrototypeCarrierManifest(input: CreateCarrierManifestInput) {
  const manifest = createManifest({
    id: `manifest-${input.warehouseCode.toLowerCase()}-${input.carrierCode.toLowerCase()}-${input.date.replaceAll("-", "")}`,
    carrierCode: input.carrierCode.toUpperCase(),
    carrierName: input.carrierName,
    warehouseId: input.warehouseId,
    warehouseCode: input.warehouseCode,
    date: input.date,
    handoverBatch: input.handoverBatch || "day",
    stagingZone: input.stagingZone || input.handoverZoneCode || "handover",
    handoverZoneCode: input.handoverZoneCode || input.stagingZone || "handover",
    handoverBinCode: input.handoverBinCode,
    status: "draft",
    owner: input.owner || "Warehouse Lead",
    auditLogId: "audit-manifest-created-prototype",
    createdAt: "2026-04-26T10:00:00Z",
    lines: []
  });
  prototypeCarrierManifests = [
    manifest,
    ...prototypeCarrierManifests.filter((candidate) => candidate.id !== manifest.id)
  ];

  return cloneManifest(manifest);
}

function addPrototypeShipmentToManifest(manifestId: string, shipmentId: string): CarrierManifest {
  const manifest = findPrototypeManifest(manifestId);
  if (manifest.lines.some((line) => line.shipmentId === shipmentId)) {
    throw new Error("Shipment already exists in carrier manifest");
  }
  if (manifest.status !== "draft" && manifest.status !== "ready") {
    throw new Error("Carrier manifest status transition is invalid");
  }
  const shipment = prototypeShipmentStates.find((candidate) => candidate.shipmentId === shipmentId);
  if (shipment && !shipment.packed) {
    throw new Error("Shipment must be packed before adding to carrier manifest");
  }
  if (shipment?.carrierCode && shipment.carrierCode !== manifest.carrierCode) {
    throw new Error("Shipment carrier does not match carrier manifest");
  }

  const updated = createManifest({
    ...manifest,
    auditLogId: "audit-manifest-shipment-added-prototype",
    lines: [
      ...manifest.lines,
      {
        id: `line-${shipmentId}`,
        shipmentId,
        orderNo: shipment?.orderNo ?? shipmentId.toUpperCase(),
        trackingNo: shipment?.trackingNo ?? shipmentId.toUpperCase(),
        packageCode: shipment?.packageCode ?? "TOTE-A03",
        stagingZone: shipment?.stagingZone ?? manifest.stagingZone,
        handoverZoneCode: shipment?.handoverZoneCode ?? manifest.handoverZoneCode ?? manifest.stagingZone,
        handoverBinCode: shipment?.handoverBinCode ?? manifest.handoverBinCode,
        scanned: false
      }
    ]
  });

  replacePrototypeManifest(updated);
  return cloneManifest(updated);
}

function removePrototypeShipmentFromManifest(manifestId: string, shipmentId: string): CarrierManifest {
  const manifest = findPrototypeManifest(manifestId);
  if (manifest.status !== "draft" && manifest.status !== "ready") {
    throw new Error("Carrier manifest status transition is invalid");
  }
  if (!manifest.lines.some((line) => line.shipmentId === shipmentId)) {
    throw new Error("Shipment was not found in carrier manifest");
  }

  const updated = createManifest({
    ...manifest,
    status: manifest.lines.length === 1 ? "draft" : manifest.status,
    auditLogId: "audit-manifest-shipment-removed-prototype",
    lines: manifest.lines.filter((line) => line.shipmentId !== shipmentId)
  });

  replacePrototypeManifest(updated);
  return cloneManifest(updated);
}

function markPrototypeManifestReady(manifestId: string): CarrierManifest {
  const manifest = findPrototypeManifest(manifestId);
  if (manifest.status === "ready" || manifest.status === "scanning") {
    return cloneManifest(manifest);
  }
  if (manifest.status !== "draft" || manifest.lines.length === 0) {
    throw new Error("Carrier manifest status transition is invalid");
  }

  const updated = createManifest({
    ...manifest,
    status: "ready",
    auditLogId: "audit-manifest-ready-prototype"
  });
  replacePrototypeManifest(updated);

  return cloneManifest(updated);
}

function cancelPrototypeManifest(manifestId: string): CarrierManifest {
  const manifest = findPrototypeManifest(manifestId);
  if (manifest.status === "completed" || manifest.status === "handed_over") {
    throw new Error("Carrier manifest is already completed");
  }
  if (manifest.status === "cancelled") {
    return cloneManifest(manifest);
  }

  const updated = createManifest({
    ...manifest,
    status: "cancelled",
    auditLogId: "audit-manifest-cancelled-prototype"
  });
  replacePrototypeManifest(updated);

  return cloneManifest(updated);
}

function confirmPrototypeManifestHandover(manifestId: string): CarrierManifest {
  const manifest = findPrototypeManifest(manifestId);
  if (manifest.status === "handed_over") {
    return cloneManifest(manifest);
  }
  if (manifest.status !== "ready" && manifest.status !== "scanning") {
    throw new Error("Carrier manifest status transition is invalid");
  }
  if (manifest.summary.expectedCount === 0 || manifest.summary.missingCount > 0) {
    throw new Error("Carrier manifest has missing orders");
  }

  const updated = createManifest({
    ...manifest,
    status: "handed_over",
    auditLogId: "audit-manifest-handed-over-prototype"
  });
  replacePrototypeManifest(updated);

  return cloneManifest(updated);
}

function verifyPrototypeCarrierManifestScan(
  input: VerifyCarrierManifestScanInput,
  manifests: CarrierManifest[]
): CarrierManifestScanResult {
  const manifest = manifests.find((candidate) => candidate.id === input.manifestId);
  if (!manifest) {
    throw new Error("Carrier manifest not found");
  }

  const code = normalizeScanCode(input.code);
  if (code === "") {
    throw new Error("Scan code is required");
  }

  const lineIndex = manifest.lines.findIndex((line) => matchesScanCode(line, code));
  if (lineIndex >= 0) {
    const line = manifest.lines[lineIndex];
    if (manifest.status !== "ready" && manifest.status !== "scanning") {
      return createScanResult({
        code,
        manifest,
        line,
        resultCode: "INVALID_STATE",
        severity: "danger",
        message: "Manifest cannot accept scans in its current state",
        stationId: input.stationId,
        deviceId: input.deviceId,
        source: input.source
      });
    }
    if (line.scanned) {
      return createScanResult({
        code,
        manifest,
        line,
        resultCode: "DUPLICATE_SCAN",
        severity: "warning",
        message: "Shipment was already scanned for this manifest",
        stationId: input.stationId,
        deviceId: input.deviceId,
        source: input.source
      });
    }

    const updatedManifest = createManifest({
      ...manifest,
      status: manifest.status === "ready" ? "scanning" : manifest.status,
      auditLogId: "audit-manifest-scan-prototype",
      lines: manifest.lines.map((candidate, index) => (index === lineIndex ? { ...candidate, scanned: true } : candidate))
    });
    if (manifests === prototypeCarrierManifests) {
      replacePrototypeManifest(updatedManifest);
    }

    return createScanResult({
      code,
      manifest: updatedManifest,
      line: updatedManifest.lines[lineIndex],
      resultCode: "MATCHED",
      severity: "success",
      message: "Scan matched manifest line",
      stationId: input.stationId,
      deviceId: input.deviceId,
      source: input.source,
      auditLogId: "audit-manifest-scan-prototype"
    });
  }

  const expectedManifest = manifests.find((candidate) => candidate.lines.some((line) => matchesScanCode(line, code)));
  if (expectedManifest) {
    const expectedLine = expectedManifest.lines.find((line) => matchesScanCode(line, code));

    return createScanResult({
      code,
      manifest,
      line: expectedLine,
      resultCode: "MANIFEST_MISMATCH",
      severity: "danger",
      message: "Scan code belongs to a different manifest",
      expectedManifestId: expectedManifest.id,
      stationId: input.stationId,
      deviceId: input.deviceId,
      source: input.source
    });
  }

  const shipment = prototypeShipmentStates.find((candidate) =>
    [candidate.shipmentId, candidate.orderNo, candidate.trackingNo, candidate.packageCode].some(
      (value) => normalizeScanCode(value) === code
    )
  );
  if (shipment && !shipment.packed) {
    return createScanResult({
      code,
      manifest,
      line: {
        id: `line-${shipment.shipmentId}`,
        shipmentId: shipment.shipmentId,
        orderNo: shipment.orderNo,
        trackingNo: shipment.trackingNo,
        packageCode: shipment.packageCode,
        stagingZone: shipment.stagingZone,
        scanned: false
      },
      resultCode: "INVALID_STATE",
      severity: "danger",
      message: "Shipment is not packed and cannot be handed over",
      stationId: input.stationId,
      deviceId: input.deviceId,
      source: input.source
    });
  }
  if (shipment?.carrierCode && shipment.carrierCode !== manifest.carrierCode) {
    return createScanResult({
      code,
      manifest,
      line: {
        id: `line-${shipment.shipmentId}`,
        shipmentId: shipment.shipmentId,
        orderNo: shipment.orderNo,
        trackingNo: shipment.trackingNo,
        packageCode: shipment.packageCode,
        stagingZone: shipment.stagingZone,
        handoverZoneCode: shipment.handoverZoneCode,
        handoverBinCode: shipment.handoverBinCode,
        scanned: false
      },
      resultCode: "MANIFEST_MISMATCH",
      severity: "danger",
      message: "Shipment carrier does not match carrier manifest",
      stationId: input.stationId,
      deviceId: input.deviceId,
      source: input.source
    });
  }
  if (shipment) {
    return createScanResult({
      code,
      manifest,
      line: {
        id: `line-${shipment.shipmentId}`,
        shipmentId: shipment.shipmentId,
        orderNo: shipment.orderNo,
        trackingNo: shipment.trackingNo,
        packageCode: shipment.packageCode,
        stagingZone: shipment.stagingZone,
        handoverZoneCode: shipment.handoverZoneCode,
        handoverBinCode: shipment.handoverBinCode,
        scanned: false
      },
      resultCode: "MANIFEST_MISMATCH",
      severity: "danger",
      message: "Packed shipment is not expected on this manifest",
      stationId: input.stationId,
      deviceId: input.deviceId,
      source: input.source
    });
  }

  return createScanResult({
    code,
    manifest,
    resultCode: "NOT_FOUND",
    severity: "danger",
    message: "Scan code was not found",
    stationId: input.stationId,
    deviceId: input.deviceId,
    source: input.source
  });
}

function createManifest(input: Omit<CarrierManifest, "summary" | "missingLines">): CarrierManifest {
  return {
    ...input,
    summary: summarizeCarrierManifestLines(input.lines),
    missingLines: input.lines.filter((line) => !line.scanned)
  };
}

function createScanResult({
  code,
  manifest,
  line,
  resultCode,
  severity,
  message,
  expectedManifestId,
  stationId,
  deviceId,
  source,
  auditLogId
}: {
  code: string;
  manifest: CarrierManifest;
  line?: CarrierManifestLine;
  resultCode: CarrierManifestScanResultCode;
  severity: CarrierManifestScanSeverity;
  message: string;
  expectedManifestId?: string;
  stationId?: string;
  deviceId?: string;
  source?: string;
  auditLogId?: string;
}): CarrierManifestScanResult {
  const scanEvent = {
    id: `scan-${manifest.id}-${code}-${resultCode}-${++prototypeScanSequence}`.toLowerCase(),
    manifestId: manifest.id,
    expectedManifestId,
    code,
    resultCode,
    severity,
    message,
    shipmentId: line?.shipmentId,
    orderNo: line?.orderNo,
    trackingNo: line?.trackingNo,
    actorId: "user-handover-operator",
    stationId: stationId || "shipping-handover",
    deviceId,
    source: source || "shipping_handover",
    warehouseId: manifest.warehouseId,
    carrierCode: manifest.carrierCode,
    createdAt: "2026-04-26T10:15:00Z"
  };

  return {
    resultCode,
    severity,
    message,
    expectedManifestId,
    line,
    scanEvent,
    manifest,
    auditLogId
  };
}

function createPrototypeCarrierManifests(): CarrierManifest[] {
  return [
    createManifest({
      id: "manifest-hcm-ghn-morning",
      carrierCode: "GHN",
      carrierName: "GHN Express",
      warehouseId: "wh-hcm",
      warehouseCode: "HCM",
      date: "2026-04-26",
      handoverBatch: "morning",
      stagingZone: "handover-a",
      handoverZoneCode: "handover-a",
      handoverBinCode: "A01",
      status: "scanning",
      owner: "Handover Operator",
      createdAt: "2026-04-26T07:45:00Z",
      lines: [
        {
          id: "line-ship-hcm-001",
          shipmentId: "ship-hcm-260426-001",
          orderNo: "SO-260426-001",
          trackingNo: "GHN260426001",
          packageCode: "TOTE-A01",
          stagingZone: "handover-a",
          handoverZoneCode: "handover-a",
          handoverBinCode: "A01",
          scanned: true
        },
        {
          id: "line-ship-hcm-002",
          shipmentId: "ship-hcm-260426-002",
          orderNo: "SO-260426-002",
          trackingNo: "GHN260426002",
          packageCode: "TOTE-A01",
          stagingZone: "handover-a",
          handoverZoneCode: "handover-a",
          handoverBinCode: "A01",
          scanned: true
        },
        {
          id: "line-ship-hcm-003",
          shipmentId: "ship-hcm-260426-003",
          orderNo: "SO-260426-003",
          trackingNo: "GHN260426003",
          packageCode: "TOTE-A02",
          stagingZone: "handover-a",
          handoverZoneCode: "handover-a",
          handoverBinCode: "A02",
          scanned: false
        }
      ]
    }),
    createManifest({
      id: "manifest-hcm-vtp-noon",
      carrierCode: "VTP",
      carrierName: "Viettel Post",
      warehouseId: "wh-hcm",
      warehouseCode: "HCM",
      date: "2026-04-26",
      handoverBatch: "noon",
      stagingZone: "handover-b",
      handoverZoneCode: "handover-b",
      handoverBinCode: "B01",
      status: "ready",
      owner: "Warehouse Lead",
      createdAt: "2026-04-26T09:00:00Z",
      lines: [
        {
          id: "line-ship-hcm-vtp-001",
          shipmentId: "ship-hcm-vtp-260426-001",
          orderNo: "SO-260426-011",
          trackingNo: "VTP260426011",
          packageCode: "TOTE-B01",
          stagingZone: "handover-b",
          handoverZoneCode: "handover-b",
          handoverBinCode: "B01",
          scanned: false
        }
      ]
    }),
    createManifest({
      id: "manifest-hn-ghn-day",
      carrierCode: "GHN",
      carrierName: "GHN Express",
      warehouseId: "wh-hn",
      warehouseCode: "HN",
      date: "2026-04-26",
      handoverBatch: "day",
      stagingZone: "hn-handover",
      handoverZoneCode: "hn-handover",
      handoverBinCode: "HN-01",
      status: "completed",
      owner: "HN Lead",
      createdAt: "2026-04-26T08:20:00Z",
      lines: [
        {
          id: "line-ship-hn-001",
          shipmentId: "ship-hn-260426-001",
          orderNo: "SO-260426-HN-011",
          trackingNo: "GHNHN260426001",
          packageCode: "HN-TOTE-01",
          stagingZone: "hn-handover",
          handoverZoneCode: "hn-handover",
          handoverBinCode: "HN-01",
          scanned: true
        }
      ]
    })
  ];
}

function findPrototypeManifest(id: string) {
  const manifest = prototypeCarrierManifests.find((candidate) => candidate.id === id);
  if (!manifest) {
    throw new Error("Carrier manifest not found");
  }

  return manifest;
}

function replacePrototypeManifest(manifest: CarrierManifest) {
  prototypeCarrierManifests = [
    manifest,
    ...prototypeCarrierManifests.filter((candidate) => candidate.id !== manifest.id)
  ];
}

function cloneManifest(manifest: CarrierManifest): CarrierManifest {
  return {
    ...manifest,
    lines: manifest.lines.map((line) => ({ ...line })),
    missingLines: manifest.missingLines.map((line) => ({ ...line })),
    summary: { ...manifest.summary }
  };
}

function matchesScanCode(line: CarrierManifestLine, code: string) {
  return [line.orderNo, line.trackingNo, line.shipmentId, line.packageCode].some((candidate) => normalizeScanCode(candidate) === code);
}

function normalizeScanCode(code: string) {
  return code.trim().toUpperCase();
}

function sortManifests(left: CarrierManifest, right: CarrierManifest) {
  if (left.date !== right.date) {
    return right.date.localeCompare(left.date);
  }
  if (left.warehouseCode !== right.warehouseCode) {
    return left.warehouseCode.localeCompare(right.warehouseCode);
  }
  if (left.carrierCode !== right.carrierCode) {
    return left.carrierCode.localeCompare(right.carrierCode);
  }

  return left.handoverBatch.localeCompare(right.handoverBatch);
}
