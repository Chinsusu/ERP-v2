"use client";

import { useEffect, useMemo, useRef, useState, type KeyboardEvent } from "react";
import { DataTable, StatusChip, type DataTableColumn } from "@/shared/design-system/components";
import { useCarrierManifests } from "../hooks/useCarrierManifests";
import {
  addShipmentToManifest,
  cancelCarrierManifest,
  carrierManifestScanSeverityTone,
  carrierManifestStatusOptions,
  carrierManifestStatusTone,
  carrierOptions,
  createCarrierManifest,
  defaultCarrierManifestDate,
  markCarrierManifestReady,
  manifestWarehouseOptions,
  removeShipmentFromManifest,
  verifyCarrierManifestScan
} from "../services/carrierManifestService";
import type {
  CarrierManifest,
  CarrierManifestLine,
  CarrierManifestQuery,
  CarrierManifestScanResult,
  CarrierManifestStatus
} from "../types";

function createManifestColumns(
  selectedManifestId: string,
  onSelectManifest: (manifestId: string) => void
): DataTableColumn<CarrierManifest>[] {
  return [
    {
      key: "manifest",
      header: "Manifest",
      render: (row) => (
        <span className="erp-shipping-manifest-cell">
          <strong>{row.id}</strong>
          <small>{row.handoverBatch} / {manifestLocationLabel(row)}</small>
        </span>
      ),
      width: "280px"
    },
    {
      key: "carrier",
      header: "Carrier",
      render: (row) => row.carrierName,
      width: "150px"
    },
    {
      key: "warehouse",
      header: "Warehouse",
      render: (row) => row.warehouseCode,
      width: "100px"
    },
    {
      key: "status",
      header: "Status",
      render: (row) => <StatusChip tone={carrierManifestStatusTone(row.status)}>{statusLabel(row.status)}</StatusChip>,
      width: "130px"
    },
    {
      key: "orders",
      header: "Orders",
      render: (row) => row.summary.expectedCount,
      align: "right",
      width: "90px"
    },
    {
      key: "scanned",
      header: "Scanned",
      render: (row) => row.summary.scannedCount,
      align: "right",
      width: "100px"
    },
    {
      key: "missing",
      header: "Missing",
      render: (row) => (
        <StatusChip tone={row.summary.missingCount === 0 ? "success" : "danger"}>{row.summary.missingCount}</StatusChip>
      ),
      align: "right",
      width: "100px"
    },
    {
      key: "select",
      header: "",
      render: (row) => (
        <button
          className={`erp-button ${row.id === selectedManifestId ? "erp-button--primary" : "erp-button--secondary"}`}
          type="button"
          onClick={() => onSelectManifest(row.id)}
        >
          Select
        </button>
      ),
      align: "right",
      width: "110px"
    }
  ];
}

function createLineColumns(
  canRemoveLine: boolean,
  onRemoveLine: (line: CarrierManifestLine) => void
): DataTableColumn<CarrierManifestLine>[] {
  return [
    { key: "order", header: "Order", render: (row) => row.orderNo, width: "150px" },
    { key: "tracking", header: "Tracking", render: (row) => row.trackingNo, width: "160px" },
    { key: "package", header: "Package", render: (row) => row.packageCode, width: "120px" },
    { key: "zone", header: "Zone", render: (row) => row.handoverZoneCode || row.stagingZone, width: "130px" },
    { key: "bin", header: "Bin", render: (row) => row.handoverBinCode || "-", width: "90px" },
    {
      key: "scan",
      header: "Scan state",
      render: (row) => <StatusChip tone={row.scanned ? "success" : "warning"}>{row.scanned ? "Scanned" : "Missing"}</StatusChip>,
      width: "120px"
    },
    {
      key: "action",
      header: "",
      render: (row) => (
        <button
          className="erp-button erp-button--secondary"
          type="button"
          onClick={() => onRemoveLine(row)}
          disabled={!canRemoveLine}
        >
          Remove
        </button>
      ),
      align: "right",
      width: "110px"
    }
  ];
}

function createMissingLineColumns(
  onMissingAction: (action: "find" | "report", orderNo: string) => void
): DataTableColumn<CarrierManifestLine>[] {
  return [
    { key: "order", header: "Order", render: (row) => row.orderNo, width: "150px" },
    { key: "tracking", header: "Tracking", render: (row) => row.trackingNo, width: "160px" },
    { key: "zone", header: "Packing zone", render: (row) => row.handoverZoneCode || row.stagingZone || "handover", width: "140px" },
    {
      key: "state",
      header: "State",
      render: () => <StatusChip tone="danger">Missing</StatusChip>,
      width: "110px"
    },
    {
      key: "action",
      header: "Action",
      render: (row) => (
        <div className="erp-shipping-missing-actions">
          <button className="erp-button erp-button--secondary" type="button" onClick={() => onMissingAction("find", row.orderNo)}>
            Find order
          </button>
          <button className="erp-button erp-button--danger" type="button" onClick={() => onMissingAction("report", row.orderNo)}>
            Report
          </button>
        </div>
      ),
      width: "220px"
    }
  ];
}

export function CarrierManifestPrototype() {
  const scanInputRef = useRef<HTMLInputElement>(null);
  const [warehouseId, setWarehouseId] = useState("");
  const [date, setDate] = useState(defaultCarrierManifestDate);
  const [carrierCode, setCarrierCode] = useState("");
  const [status, setStatus] = useState<"" | CarrierManifestStatus>("");
  const [selectedManifestId, setSelectedManifestId] = useState("manifest-hcm-ghn-morning");
  const [shipmentId, setShipmentId] = useState("ship-hcm-260426-004");
  const [scanCode, setScanCode] = useState("");
  const [feedback, setFeedback] = useState("");
  const [scanResult, setScanResult] = useState<CarrierManifestScanResult | null>(null);
  const [recentScans, setRecentScans] = useState<CarrierManifestScanResult[]>([]);
  const [scanBusy, setScanBusy] = useState(false);
  const [actionBusy, setActionBusy] = useState(false);
  const [localManifests, setLocalManifests] = useState<CarrierManifest[]>([]);
  const query = useMemo<CarrierManifestQuery>(
    () => ({
      warehouseId: warehouseId || undefined,
      date,
      carrierCode: carrierCode || undefined,
      status: status || undefined
    }),
    [carrierCode, date, status, warehouseId]
  );
  const { manifests, loading, error } = useCarrierManifests(query);
  const displayedManifests = useMemo(() => {
    const localMatches = localManifests.filter((manifest) => matchesManifestQuery(manifest, query));
    const localIds = new Set(localMatches.map((manifest) => manifest.id));

    return [...localMatches, ...manifests.filter((manifest) => !localIds.has(manifest.id))];
  }, [localManifests, manifests, query]);
  const selectedManifest =
    displayedManifests.find((manifest) => manifest.id === selectedManifestId) ?? displayedManifests[0] ?? null;
  const missingLines = selectedManifest?.missingLines ?? [];
  const canConfirmHandover =
    selectedManifest !== null &&
    !actionBusy &&
    selectedManifest.summary.expectedCount > 0 &&
    selectedManifest.summary.missingCount === 0 &&
    selectedManifest.status !== "completed" &&
    selectedManifest.status !== "handed_over" &&
    selectedManifest.status !== "cancelled";
  const canMarkReady =
    selectedManifest !== null &&
    !actionBusy &&
    selectedManifest.status === "draft" &&
    selectedManifest.summary.expectedCount > 0;
  const canCancel =
    selectedManifest !== null &&
    !actionBusy &&
    selectedManifest.status !== "completed" &&
    selectedManifest.status !== "handed_over" &&
    selectedManifest.status !== "cancelled";
  const canRemoveLine =
    selectedManifest !== null && !actionBusy && (selectedManifest.status === "draft" || selectedManifest.status === "ready");
  const totals = displayedManifests.reduce(
    (acc, manifest) => ({
      expected: acc.expected + manifest.summary.expectedCount,
      scanned: acc.scanned + manifest.summary.scannedCount,
      missing: acc.missing + manifest.summary.missingCount
    }),
    { expected: 0, scanned: 0, missing: 0 }
  );

  useEffect(() => {
    if (selectedManifest) {
      window.setTimeout(() => scanInputRef.current?.focus(), 0);
    }
  }, [selectedManifest, selectedManifestId]);
  const manifestColumns = useMemo(
    () => createManifestColumns(selectedManifest?.id ?? "", handleSelectManifest),
    [selectedManifest?.id]
  );
  const lineColumns = useMemo(
    () => createLineColumns(canRemoveLine, handleRemoveShipment),
    [canRemoveLine, selectedManifest?.id]
  );
  const missingColumns = createMissingLineColumns(handleMissingAction);

  function handleSelectManifest(manifestId: string) {
    setSelectedManifestId(manifestId);
    setScanCode("");
    setScanResult(null);
  }

  async function handleCreateManifest() {
    await runManifestAction(async () => {
      const warehouseOption = manifestWarehouseOptions.find((option) => option.value === (warehouseId || "wh-hcm"));
      const carrierOption = carrierOptions.find((option) => option.value === (carrierCode || "GHN"));
      const handoverZoneCode = carrierOption?.value === "VTP" ? "handover-b" : "handover-a";
      const created = await createCarrierManifest({
        carrierCode: carrierOption?.value || "GHN",
        carrierName: carrierOption?.label || "GHN",
        warehouseId: warehouseOption?.value || "wh-hcm",
        warehouseCode: warehouseOption && "code" in warehouseOption ? warehouseOption.code : "HCM",
        date,
        stagingZone: handoverZoneCode,
        handoverZoneCode,
        handoverBinCode: "A01"
      });
      patchManifest(created);
      setSelectedManifestId(created.id);
      setFeedback(`Created ${created.id}`);
    });
  }

  async function handleAddShipment() {
    if (!selectedManifest || shipmentId.trim() === "") {
      return;
    }
    await runManifestAction(async () => {
      const updated = await addShipmentToManifest(selectedManifest.id, shipmentId.trim());
      patchManifest(updated);
      setFeedback(`Added ${shipmentId.trim()} to ${updated.id}`);
    });
  }

  async function handleMarkReady() {
    if (!selectedManifest || !canMarkReady) {
      return;
    }

    await runManifestAction(async () => {
      const updated = await markCarrierManifestReady(selectedManifest.id);
      patchManifest(updated);
      setFeedback(`${updated.id} is ready to scan`);
    });
  }

  async function handleCancelManifest() {
    if (!selectedManifest || !canCancel) {
      return;
    }

    await runManifestAction(async () => {
      const updated = await cancelCarrierManifest(selectedManifest.id, "carrier pickup moved");
      patchManifest(updated);
      setFeedback(`Cancelled ${updated.id}`);
    });
  }

  async function handleRemoveShipment(line: CarrierManifestLine) {
    if (!selectedManifest || !canRemoveLine) {
      return;
    }

    await runManifestAction(async () => {
      const updated = await removeShipmentFromManifest(selectedManifest.id, line.shipmentId);
      patchManifest(updated);
      setFeedback(`Removed ${line.orderNo} from ${updated.id}`);
    });
  }

  async function handleVerifyScan(code: string) {
    if (!selectedManifest || scanBusy) {
      return null;
    }

    setScanBusy(true);
    try {
      const result = await verifyCarrierManifestScan(
        {
          manifestId: selectedManifest.id,
          code,
          stationId: `${selectedManifest.warehouseCode.toLowerCase()}-${selectedManifest.handoverZoneCode || selectedManifest.stagingZone}`
        },
        displayedManifests
      );
      setLocalManifests((current) => [result.manifest, ...current.filter((manifest) => manifest.id !== result.manifest.id)]);
      setSelectedManifestId(result.manifest.id);
      setScanResult(result);
      setRecentScans((current) => [result, ...current].slice(0, 6));
      setFeedback(`${result.resultCode}: ${result.message}`);
      return result;
    } catch (cause) {
      setFeedback(cause instanceof Error ? cause.message : "Scan could not be verified");
      return null;
    } finally {
      setScanBusy(false);
    }
  }

  async function handleScanKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (event.key !== "Enter") {
      return;
    }

    const code = scanCode.trim();
    if (code === "") {
      return;
    }

    const result = await handleVerifyScan(code);
    if (!result) {
      return;
    }
    if (result.resultCode === "MATCHED") {
      setScanCode("");
      scanInputRef.current?.focus();
      return;
    }

    setScanCode(result.scanEvent.code);
    window.setTimeout(() => scanInputRef.current?.select(), 0);
  }

  function handleConfirmHandover() {
    if (!selectedManifest || !canConfirmHandover) {
      return;
    }

    const handedOver: CarrierManifest = {
      ...selectedManifest,
      status: "handed_over"
    };
    patchManifest(handedOver);
    setFeedback(`Confirmed handover for ${handedOver.id}`);
  }

  function handleMissingAction(action: "find" | "report", orderNo: string) {
    setFeedback(action === "find" ? `Locate ${orderNo} in packing zone` : `Reported missing ${orderNo}`);
  }

  function handleReportManifestMissing() {
    setFeedback(`Reported missing lines for ${selectedManifest?.id ?? "manifest"}`);
  }

  async function runManifestAction(action: () => Promise<void>) {
    if (actionBusy) {
      return;
    }

    setActionBusy(true);
    try {
      await action();
    } catch (cause) {
      setFeedback(cause instanceof Error ? cause.message : "Manifest action failed");
    } finally {
      setActionBusy(false);
    }
  }

  function patchManifest(manifest: CarrierManifest) {
    setLocalManifests((current) => [manifest, ...current.filter((candidate) => candidate.id !== manifest.id)]);
    setSelectedManifestId(manifest.id);
  }

  return (
    <section className="erp-module-page erp-shipping-page">
      <header className="erp-page-header">
        <div>
          <p className="erp-module-eyebrow">SHIP</p>
          <h1 className="erp-page-title">Carrier Manifest</h1>
          <p className="erp-page-description">Carrier handover batches by warehouse, date, expected count, scanned count, and missing count</p>
        </div>
        <div className="erp-page-actions">
          <button className="erp-button erp-button--secondary" type="button" onClick={() => setStatus("exception")}>
            Exceptions
          </button>
          <button className="erp-button erp-button--primary" type="button" onClick={handleCreateManifest} disabled={actionBusy}>
            Create manifest
          </button>
        </div>
      </header>

      <section className="erp-shipping-toolbar" aria-label="Carrier manifest filters">
        <label className="erp-field">
          <span>Warehouse</span>
          <select className="erp-input" value={warehouseId} onChange={(event) => setWarehouseId(event.target.value)}>
            {manifestWarehouseOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Date</span>
          <input className="erp-input" type="date" value={date} onChange={(event) => setDate(event.target.value)} />
        </label>
        <label className="erp-field">
          <span>Carrier</span>
          <select className="erp-input" value={carrierCode} onChange={(event) => setCarrierCode(event.target.value)}>
            {carrierOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Status</span>
          <select className="erp-input" value={status} onChange={(event) => setStatus(event.target.value as "" | CarrierManifestStatus)}>
            {carrierManifestStatusOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
      </section>

      <section className="erp-kpi-grid erp-shipping-kpis">
        <ShippingKPI label="Expected" value={totals.expected} tone="info" />
        <ShippingKPI label="Scanned" value={totals.scanned} tone="success" />
        <ShippingKPI label="Missing" value={totals.missing} tone={totals.missing === 0 ? "success" : "danger"} />
        <ShippingKPI label="Manifests" value={displayedManifests.length} tone="normal" />
      </section>

      <section className="erp-card erp-card--padded erp-module-table-card">
        <div className="erp-section-header">
          <h2 className="erp-section-title">Manifest batches</h2>
          <StatusChip tone={displayedManifests.length === 0 ? "warning" : "info"}>{displayedManifests.length} rows</StatusChip>
        </div>
        <DataTable
          columns={manifestColumns}
          rows={displayedManifests}
          getRowKey={(row) => row.id}
          loading={loading}
          error={error?.message}
        />
      </section>

      <section className="erp-shipping-handover-grid">
        <div className="erp-card erp-card--padded erp-shipping-handover-card">
          <div className="erp-section-header">
            <div>
              <h2 className="erp-section-title">Carrier handover</h2>
              <p className="erp-section-description">{selectedManifest?.id ?? "No manifest selected"}</p>
            </div>
            {selectedManifest ? (
              <StatusChip tone={carrierManifestStatusTone(selectedManifest.status)}>{statusLabel(selectedManifest.status)}</StatusChip>
            ) : null}
          </div>

          <div className="erp-shipping-handover-meta">
            <label className="erp-field">
              <span>Manifest</span>
              <select className="erp-input" value={selectedManifest?.id ?? ""} onChange={(event) => setSelectedManifestId(event.target.value)}>
                {displayedManifests.map((manifest) => (
                  <option key={manifest.id} value={manifest.id}>
                    {manifest.id}
                  </option>
                ))}
              </select>
            </label>
            <HandoverMetric label="Carrier" value={selectedManifest?.carrierName ?? "-"} />
            <HandoverMetric label="Zone" value={selectedManifest ? manifestLocationLabel(selectedManifest) : "-"} />
            <HandoverMetric label="Bin" value={selectedManifest?.handoverBinCode ?? "-"} />
            <HandoverMetric label="Owner" value={selectedManifest?.owner ?? "-"} />
            <HandoverMetric label="Expected" value={selectedManifest?.summary.expectedCount ?? 0} />
            <HandoverMetric label="Scanned" value={selectedManifest?.summary.scannedCount ?? 0} />
            <HandoverMetric label="Missing" value={selectedManifest?.summary.missingCount ?? 0} tone={(selectedManifest?.summary.missingCount ?? 0) === 0 ? "success" : "danger"} />
          </div>

          <label className={`erp-shipping-scan-primary erp-shipping-scan-primary--${scanResult?.severity ?? "normal"}`}>
            <span>Scan order or tracking</span>
            <input
              ref={scanInputRef}
              value={scanCode}
              onChange={(event) => setScanCode(event.target.value)}
              onKeyDown={handleScanKeyDown}
              placeholder="SO-260426-003 / GHN260426003"
              disabled={!selectedManifest || actionBusy}
            />
            {scanResult ? (
              <small>
                <strong>{scanResult.resultCode}</strong>
                <span>{scanResult.message}</span>
                {scanResult.expectedManifestId ? <span>Related {scanResult.expectedManifestId}</span> : null}
              </small>
            ) : (
              <small>{scanBusy ? "Checking code" : selectedManifest ? manifestLocationLabel(selectedManifest) : "No handover zone"}</small>
            )}
          </label>

          <div className="erp-shipping-scan-actions">
            <button className="erp-button erp-button--secondary" type="button" onClick={() => setFeedback(`Printed ${selectedManifest?.id ?? "manifest"}`)}>
              Print manifest
            </button>
            <button className="erp-button erp-button--secondary" type="button" onClick={handleMarkReady} disabled={!canMarkReady}>
              Ready to scan
            </button>
            <button className="erp-button erp-button--danger" type="button" onClick={handleReportManifestMissing} disabled={missingLines.length === 0}>
              Report missing
            </button>
            <button className="erp-button erp-button--danger" type="button" onClick={handleCancelManifest} disabled={!canCancel}>
              Cancel manifest
            </button>
            <button className="erp-button erp-button--primary" type="button" onClick={handleConfirmHandover} disabled={!canConfirmHandover}>
              Confirm handover
            </button>
          </div>

          {feedback ? <small className="erp-shipping-feedback">{feedback}</small> : null}
        </div>

        <div className="erp-card erp-card--padded erp-shipping-exception-card">
          <div className="erp-section-header">
            <div>
              <h2 className="erp-section-title">Missing / exception queue</h2>
              <p className="erp-section-description">{selectedManifest ? manifestLocationLabel(selectedManifest) : "No zone"}</p>
            </div>
            <StatusChip tone={missingLines.length === 0 ? "success" : "danger"}>{missingLines.length} open</StatusChip>
          </div>
          <DataTable
            columns={missingColumns}
            rows={missingLines}
            getRowKey={(row) => row.id}
            emptyState={
              <section className="erp-shipping-empty-state">
                <StatusChip tone="success">Ready</StatusChip>
                <strong>All expected lines scanned</strong>
              </section>
            }
          />
        </div>
      </section>

      <section className="erp-shipping-detail-grid">
        <div className="erp-card erp-card--padded">
          <div className="erp-section-header">
            <h2 className="erp-section-title">Recent scans</h2>
            {scanResult ? <StatusChip tone={carrierManifestScanSeverityTone(scanResult.severity)}>{scanResult.resultCode}</StatusChip> : null}
          </div>
          {recentScans.length > 0 ? (
            <ol className="erp-shipping-scan-list" aria-label="Recent manifest scans">
              {recentScans.map((scan) => (
                <li key={scan.scanEvent.id}>
                  <StatusChip tone={carrierManifestScanSeverityTone(scan.severity)}>{scan.resultCode}</StatusChip>
                  <strong>{scan.scanEvent.code}</strong>
                  <span>{scan.line?.orderNo ?? scan.expectedManifestId ?? scan.message}</span>
                </li>
              ))}
            </ol>
          ) : (
            <section className="erp-shipping-empty-state">
              <StatusChip tone="normal">Idle</StatusChip>
              <strong>No scans yet</strong>
            </section>
          )}
        </div>

        <div className="erp-card erp-card--padded">
          <div className="erp-section-header">
            <h2 className="erp-section-title">Shipment lines</h2>
            <StatusChip tone={(selectedManifest?.summary.missingCount ?? 0) === 0 ? "success" : "danger"}>
              {selectedManifest?.summary.missingCount ?? 0} missing
            </StatusChip>
          </div>
          <DataTable columns={lineColumns} rows={selectedManifest?.lines ?? []} getRowKey={(row) => row.id} />
          <div className="erp-shipping-add-box">
            <label className="erp-field">
              <span>Shipment ID</span>
              <input className="erp-input" value={shipmentId} onChange={(event) => setShipmentId(event.target.value)} />
            </label>
            <button className="erp-button erp-button--primary" type="button" onClick={handleAddShipment} disabled={actionBusy || !selectedManifest}>
              Add shipment
            </button>
          </div>
        </div>
      </section>
    </section>
  );
}

function ShippingKPI({
  label,
  value,
  tone
}: {
  label: string;
  value: number;
  tone: "normal" | "success" | "warning" | "danger" | "info";
}) {
  return (
    <article className="erp-card erp-card--padded erp-kpi-card">
      <div className="erp-kpi-label">{label}</div>
      <strong className="erp-kpi-value">{value}</strong>
      <StatusChip tone={tone}>{label}</StatusChip>
    </article>
  );
}

function HandoverMetric({
  label,
  value,
  tone = "normal"
}: {
  label: string;
  value: string | number;
  tone?: "normal" | "success" | "warning" | "danger" | "info";
}) {
  return (
    <div className="erp-shipping-handover-metric">
      <span>{label}</span>
      <strong>{value}</strong>
      {tone !== "normal" ? <StatusChip tone={tone}>{label}</StatusChip> : null}
    </div>
  );
}

function statusLabel(status: CarrierManifestStatus) {
  switch (status) {
    case "cancelled":
      return "Cancelled";
    case "completed":
      return "Completed";
    case "handed_over":
      return "Handed over";
    case "exception":
      return "Exception";
    case "scanning":
      return "Scanning";
    case "ready":
      return "Ready";
    case "draft":
    default:
      return "Draft";
  }
}

function manifestLocationLabel(manifest: CarrierManifest) {
  return [manifest.handoverZoneCode || manifest.stagingZone, manifest.handoverBinCode].filter(Boolean).join(" / ");
}

function matchesManifestQuery(manifest: CarrierManifest, query: CarrierManifestQuery) {
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
}
