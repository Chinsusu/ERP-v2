import type {
  CarrierManifest,
  CarrierManifestLine,
  CarrierManifestQuery,
  CarrierManifestScanResult,
  CarrierManifestScanResultCode,
  CarrierManifestScanSeverity,
  CarrierManifestStatus,
  CarrierManifestSummary,
  CreateCarrierManifestInput,
  VerifyCarrierManifestScanInput
} from "../types";

export const defaultCarrierManifestDate = "2026-04-26";

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
  { label: "Exception", value: "exception" }
];

export const prototypeCarrierManifests: CarrierManifest[] = [
  createManifest({
    id: "manifest-hcm-ghn-morning",
    carrierCode: "GHN",
    carrierName: "GHN Express",
    warehouseId: "wh-hcm",
    warehouseCode: "HCM",
    date: "2026-04-26",
    handoverBatch: "morning",
    stagingZone: "handover-a",
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
        scanned: true
      },
      {
        id: "line-ship-hcm-002",
        shipmentId: "ship-hcm-260426-002",
        orderNo: "SO-260426-002",
        trackingNo: "GHN260426002",
        packageCode: "TOTE-A01",
        stagingZone: "handover-a",
        scanned: true
      },
      {
        id: "line-ship-hcm-003",
        shipmentId: "ship-hcm-260426-003",
        orderNo: "SO-260426-003",
        trackingNo: "GHN260426003",
        packageCode: "TOTE-A02",
        stagingZone: "handover-a",
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
        scanned: true
      }
    ]
  })
];

const prototypeShipmentStates = [
  {
    shipmentId: "ship-hcm-260426-099",
    orderNo: "SO-260426-099",
    trackingNo: "GHN260426099",
    packageCode: "PACKING-LANE-01",
    stagingZone: "packing",
    packed: false
  }
];

let prototypeScanSequence = 0;

export async function getCarrierManifests(query: CarrierManifestQuery = {}): Promise<CarrierManifest[]> {
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
    .sort(sortManifests);
}

export async function createCarrierManifest(input: CreateCarrierManifestInput): Promise<CarrierManifest> {
  const manifest = createManifest({
    id: `manifest-${input.warehouseCode.toLowerCase()}-${input.carrierCode.toLowerCase()}-${input.date.replaceAll("-", "")}`,
    carrierCode: input.carrierCode.toUpperCase(),
    carrierName: input.carrierName,
    warehouseId: input.warehouseId,
    warehouseCode: input.warehouseCode,
    date: input.date,
    handoverBatch: input.handoverBatch || "day",
    stagingZone: input.stagingZone || "handover",
    status: "draft",
    owner: input.owner || "Warehouse Lead",
    auditLogId: "audit-manifest-created-prototype",
    createdAt: "2026-04-26T10:00:00Z",
    lines: []
  });

  return manifest;
}

export async function addShipmentToManifest(manifestId: string, shipmentId: string): Promise<CarrierManifest> {
  const manifest = prototypeCarrierManifests.find((candidate) => candidate.id === manifestId);
  if (!manifest) {
    throw new Error("Carrier manifest not found");
  }
  if (manifest.lines.some((line) => line.shipmentId === shipmentId)) {
    throw new Error("Shipment already exists in carrier manifest");
  }

  return createManifest({
    ...manifest,
    status: manifest.status === "draft" ? "ready" : manifest.status,
    auditLogId: "audit-manifest-shipment-added-prototype",
    lines: [
      ...manifest.lines,
      {
        id: `line-${shipmentId}`,
        shipmentId,
        orderNo: shipmentId.toUpperCase(),
        trackingNo: shipmentId.toUpperCase(),
        packageCode: "TOTE-A03",
        stagingZone: manifest.stagingZone,
        scanned: false
      }
    ]
  });
}

export async function verifyCarrierManifestScan(
  input: VerifyCarrierManifestScanInput,
  manifests: CarrierManifest[] = prototypeCarrierManifests
): Promise<CarrierManifestScanResult> {
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
        stationId: input.stationId
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
        stationId: input.stationId
      });
    }

    const updatedManifest = createManifest({
      ...manifest,
      status: manifest.status === "ready" ? "scanning" : manifest.status,
      auditLogId: "audit-manifest-scan-prototype",
      lines: manifest.lines.map((candidate, index) => (index === lineIndex ? { ...candidate, scanned: true } : candidate))
    });

    return createScanResult({
      code,
      manifest: updatedManifest,
      line: updatedManifest.lines[lineIndex],
      resultCode: "MATCHED",
      severity: "success",
      message: "Scan matched manifest line",
      stationId: input.stationId,
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
      stationId: input.stationId
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
      stationId: input.stationId
    });
  }

  return createScanResult({
    code,
    manifest,
    resultCode: "NOT_FOUND",
    severity: "danger",
    message: "Scan code was not found",
    stationId: input.stationId
  });
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
      return "success";
    case "exception":
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

function createManifest(input: Omit<CarrierManifest, "summary">): CarrierManifest {
  return {
    ...input,
    summary: summarizeCarrierManifestLines(input.lines)
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
