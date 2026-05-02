"use client";

import { useEffect, useMemo, useRef, useState, type KeyboardEvent } from "react";
import { DataTable, StatusChip, type DataTableColumn } from "@/shared/design-system/components";
import { t } from "@/shared/i18n";
import { useCarrierManifests } from "../hooks/useCarrierManifests";
import {
  addShipmentToManifest,
  cancelCarrierManifest,
  carrierManifestScanSeverityTone,
  carrierManifestStatusOptions,
  carrierManifestStatusTone,
  carrierOptions,
  confirmCarrierManifestHandover,
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
  CarrierManifestScanResultCode,
  CarrierManifestStatus
} from "../types";

function createManifestColumns(
  selectedManifestId: string,
  onSelectManifest: (manifestId: string) => void
): DataTableColumn<CarrierManifest>[] {
  return [
    {
      key: "manifest",
      header: handoverCopy("columns.manifest"),
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
      header: handoverCopy("columns.carrier"),
      render: (row) => row.carrierName,
      width: "150px"
    },
    {
      key: "warehouse",
      header: handoverCopy("columns.warehouse"),
      render: (row) => row.warehouseCode,
      width: "100px"
    },
    {
      key: "status",
      header: handoverCopy("columns.status"),
      render: (row) => <StatusChip tone={carrierManifestStatusTone(row.status)}>{statusLabel(row.status)}</StatusChip>,
      width: "130px"
    },
    {
      key: "orders",
      header: handoverCopy("columns.orders"),
      render: (row) => row.summary.expectedCount,
      align: "right",
      width: "90px"
    },
    {
      key: "scanned",
      header: handoverCopy("columns.scanned"),
      render: (row) => row.summary.scannedCount,
      align: "right",
      width: "100px"
    },
    {
      key: "missing",
      header: handoverCopy("columns.missing"),
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
          {handoverCopy("actions.select")}
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
    { key: "order", header: handoverCopy("columns.order"), render: (row) => row.orderNo, width: "150px" },
    { key: "tracking", header: handoverCopy("columns.tracking"), render: (row) => row.trackingNo, width: "160px" },
    { key: "package", header: handoverCopy("columns.package"), render: (row) => row.packageCode, width: "120px" },
    { key: "zone", header: handoverCopy("columns.zone"), render: (row) => row.handoverZoneCode || row.stagingZone, width: "130px" },
    { key: "bin", header: handoverCopy("columns.bin"), render: (row) => row.handoverBinCode || "-", width: "90px" },
    {
      key: "scan",
      header: handoverCopy("columns.scanState"),
      render: (row) => (
        <StatusChip tone={row.scanned ? "success" : "warning"}>
          {handoverCopy(row.scanned ? "scanState.scanned" : "scanState.missing")}
        </StatusChip>
      ),
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
          {handoverCopy("actions.remove")}
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
    { key: "order", header: handoverCopy("columns.order"), render: (row) => row.orderNo, width: "150px" },
    { key: "tracking", header: handoverCopy("columns.tracking"), render: (row) => row.trackingNo, width: "160px" },
    { key: "zone", header: handoverCopy("columns.packingZone"), render: (row) => row.handoverZoneCode || row.stagingZone || "handover", width: "140px" },
    {
      key: "state",
      header: handoverCopy("columns.state"),
      render: () => <StatusChip tone="danger">{handoverCopy("scanState.missing")}</StatusChip>,
      width: "110px"
    },
    {
      key: "action",
      header: handoverCopy("columns.action"),
      render: (row) => (
        <div className="erp-shipping-missing-actions">
          <button className="erp-button erp-button--secondary" type="button" onClick={() => onMissingAction("find", row.orderNo)}>
            {handoverCopy("actions.findOrder")}
          </button>
          <button className="erp-button erp-button--danger" type="button" onClick={() => onMissingAction("report", row.orderNo)}>
            {handoverCopy("actions.report")}
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
  const scanIssues = recentScans.filter((scan) => scan.severity !== "success");
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
    const params = new URLSearchParams(window.location.search);
    const nextWarehouseId = optionValueFromParam(params.get("warehouse_id"), manifestWarehouseOptions);
    const nextCarrierCode = optionValueFromParam(params.get("carrier_code"), carrierOptions);
    const nextStatus = optionValueFromParam(params.get("status"), carrierManifestStatusOptions);
    const nextDate = params.get("date");

    if (nextWarehouseId !== null) {
      setWarehouseId(nextWarehouseId);
    }
    if (nextDate) {
      setDate(nextDate);
    }
    if (nextCarrierCode !== null) {
      setCarrierCode(nextCarrierCode);
    }
    if (nextStatus !== null) {
      setStatus(nextStatus as "" | CarrierManifestStatus);
    }
  }, []);

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
      setFeedback(handoverCopy("feedback.created", { manifestId: created.id }));
    });
  }

  async function handleAddShipment() {
    if (!selectedManifest || shipmentId.trim() === "") {
      return;
    }
    await runManifestAction(async () => {
      const updated = await addShipmentToManifest(selectedManifest.id, shipmentId.trim());
      patchManifest(updated);
      setFeedback(handoverCopy("feedback.added", { shipmentId: shipmentId.trim(), manifestId: updated.id }));
    });
  }

  async function handleMarkReady() {
    if (!selectedManifest || !canMarkReady) {
      return;
    }

    await runManifestAction(async () => {
      const updated = await markCarrierManifestReady(selectedManifest.id);
      patchManifest(updated);
      setFeedback(handoverCopy("feedback.ready", { manifestId: updated.id }));
    });
  }

  async function handleCancelManifest() {
    if (!selectedManifest || !canCancel) {
      return;
    }

    await runManifestAction(async () => {
      const updated = await cancelCarrierManifest(selectedManifest.id, "carrier pickup moved");
      patchManifest(updated);
      setFeedback(handoverCopy("feedback.cancelled", { manifestId: updated.id }));
    });
  }

  async function handleRemoveShipment(line: CarrierManifestLine) {
    if (!selectedManifest || !canRemoveLine) {
      return;
    }

    await runManifestAction(async () => {
      const updated = await removeShipmentFromManifest(selectedManifest.id, line.shipmentId);
      patchManifest(updated);
      setFeedback(handoverCopy("feedback.removed", { orderNo: line.orderNo, manifestId: updated.id }));
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
          stationId: `${selectedManifest.warehouseCode.toLowerCase()}-${selectedManifest.handoverZoneCode || selectedManifest.stagingZone}`,
          deviceId: `${selectedManifest.warehouseCode.toLowerCase()}-handover-ui`,
          source: "handover_scan_ui"
        },
        displayedManifests
      );
      setLocalManifests((current) => [result.manifest, ...current.filter((manifest) => manifest.id !== result.manifest.id)]);
      setSelectedManifestId(result.manifest.id);
      setScanResult(result);
      setRecentScans((current) => [result, ...current].slice(0, 6));
      setFeedback(`${scanResultLabel(result.resultCode)}: ${scanResultMessage(result)}`);
      return result;
    } catch (cause) {
      setFeedback(cause instanceof Error ? cause.message : handoverCopy("scan.failed"));
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

  async function handleConfirmHandover() {
    if (!selectedManifest || !canConfirmHandover) {
      return;
    }

    await runManifestAction(async () => {
      const handedOver = await confirmCarrierManifestHandover(selectedManifest.id);
      patchManifest(handedOver);
      setFeedback(handoverCopy("feedback.confirmed", { manifestId: handedOver.id }));
    });
  }

  function handleMissingAction(action: "find" | "report", orderNo: string) {
    setFeedback(handoverCopy(action === "find" ? "feedback.find" : "feedback.report", { orderNo }));
  }

  function handleReportManifestMissing() {
    setFeedback(handoverCopy("feedback.reportMissing", { manifestId: selectedManifest?.id ?? handoverCopy("columns.manifest") }));
  }

  async function runManifestAction(action: () => Promise<void>) {
    if (actionBusy) {
      return;
    }

    setActionBusy(true);
    try {
      await action();
    } catch (cause) {
      setFeedback(cause instanceof Error ? cause.message : handoverCopy("feedback.actionFailed"));
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
          <h1 className="erp-page-title">{handoverCopy("title")}</h1>
          <p className="erp-page-description">{handoverCopy("description")}</p>
        </div>
        <div className="erp-page-actions">
          <button className="erp-button erp-button--secondary" type="button" onClick={() => setStatus("exception")}>
            {handoverCopy("actions.exceptions")}
          </button>
          <button className="erp-button erp-button--primary" type="button" onClick={handleCreateManifest} disabled={actionBusy}>
            {handoverCopy("actions.createManifest")}
          </button>
        </div>
      </header>

      <section className="erp-shipping-toolbar" aria-label={handoverCopy("filters.label")}>
        <label className="erp-field">
          <span>{handoverCopy("filters.warehouse")}</span>
          <select className="erp-input" value={warehouseId} onChange={(event) => setWarehouseId(event.target.value)}>
            {manifestWarehouseOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {manifestWarehouseLabel(option.value, option.label)}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>{handoverCopy("filters.date")}</span>
          <input className="erp-input" type="date" value={date} onChange={(event) => setDate(event.target.value)} />
        </label>
        <label className="erp-field">
          <span>{handoverCopy("filters.carrier")}</span>
          <select className="erp-input" value={carrierCode} onChange={(event) => setCarrierCode(event.target.value)}>
            {carrierOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {carrierOptionLabel(option.value, option.label)}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>{handoverCopy("filters.status")}</span>
          <select className="erp-input" value={status} onChange={(event) => setStatus(event.target.value as "" | CarrierManifestStatus)}>
            {carrierManifestStatusOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.value ? statusLabel(option.value) : handoverCopy("filters.allStatuses")}
              </option>
            ))}
          </select>
        </label>
      </section>

      <section className="erp-kpi-grid erp-shipping-kpis">
        <ShippingKPI label={handoverCopy("kpi.expected")} value={totals.expected} tone="info" />
        <ShippingKPI label={handoverCopy("kpi.scanned")} value={totals.scanned} tone="success" />
        <ShippingKPI label={handoverCopy("kpi.missing")} value={totals.missing} tone={totals.missing === 0 ? "success" : "danger"} />
        <ShippingKPI label={handoverCopy("kpi.manifests")} value={displayedManifests.length} tone="normal" />
      </section>

      <section className="erp-card erp-card--padded erp-module-table-card" id="carrier-manifest-list">
        <div className="erp-section-header">
          <h2 className="erp-section-title">{handoverCopy("sections.manifestBatches")}</h2>
          <StatusChip tone={displayedManifests.length === 0 ? "warning" : "info"}>
            {handoverCopy("rows", { count: displayedManifests.length })}
          </StatusChip>
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
              <h2 className="erp-section-title">{handoverCopy("sections.carrierHandover")}</h2>
              <p className="erp-section-description">{selectedManifest?.id ?? handoverCopy("empty.noManifestSelected")}</p>
            </div>
            {selectedManifest ? (
              <StatusChip tone={carrierManifestStatusTone(selectedManifest.status)}>{statusLabel(selectedManifest.status)}</StatusChip>
            ) : null}
          </div>

          <div className="erp-shipping-handover-meta">
            <label className="erp-field">
              <span>{handoverCopy("facts.manifest")}</span>
              <select className="erp-input" value={selectedManifest?.id ?? ""} onChange={(event) => setSelectedManifestId(event.target.value)}>
                {displayedManifests.map((manifest) => (
                  <option key={manifest.id} value={manifest.id}>
                    {manifest.id}
                  </option>
                ))}
              </select>
            </label>
            <HandoverMetric label={handoverCopy("facts.carrier")} value={selectedManifest?.carrierName ?? "-"} />
            <HandoverMetric label={handoverCopy("facts.zone")} value={selectedManifest ? manifestLocationLabel(selectedManifest) : "-"} />
            <HandoverMetric label={handoverCopy("facts.bin")} value={selectedManifest?.handoverBinCode ?? "-"} />
            <HandoverMetric label={handoverCopy("facts.owner")} value={selectedManifest?.owner ?? "-"} />
            <HandoverMetric label={handoverCopy("facts.expected")} value={selectedManifest?.summary.expectedCount ?? 0} />
            <HandoverMetric label={handoverCopy("facts.scanned")} value={selectedManifest?.summary.scannedCount ?? 0} />
            <HandoverMetric label={handoverCopy("facts.missing")} value={selectedManifest?.summary.missingCount ?? 0} tone={(selectedManifest?.summary.missingCount ?? 0) === 0 ? "success" : "danger"} />
          </div>

          <label className={`erp-shipping-scan-primary erp-shipping-scan-primary--${scanResult?.severity ?? "normal"}`}>
            <span>{handoverCopy("scan.label")}</span>
            <input
              ref={scanInputRef}
              value={scanCode}
              onChange={(event) => setScanCode(event.target.value)}
              onKeyDown={handleScanKeyDown}
              placeholder={handoverCopy("scan.placeholder")}
              disabled={!selectedManifest || actionBusy}
            />
            {scanResult ? (
              <small>
                <strong>{scanResultLabel(scanResult.resultCode)}</strong>
                <span>{scanResultMessage(scanResult)}</span>
                {scanResult.expectedManifestId ? (
                  <span>{handoverCopy("scan.relatedManifest", { manifestId: scanResult.expectedManifestId })}</span>
                ) : null}
              </small>
            ) : (
              <small>
                {scanBusy
                  ? handoverCopy("scan.checkingCode")
                  : selectedManifest
                    ? manifestLocationLabel(selectedManifest)
                    : handoverCopy("scan.noHandoverZone")}
              </small>
            )}
          </label>

          <div className="erp-shipping-scan-actions">
            <button
              className="erp-button erp-button--secondary"
              type="button"
              onClick={() =>
                setFeedback(
                  handoverCopy("feedback.printed", { manifestId: selectedManifest?.id ?? handoverCopy("columns.manifest") })
                )
              }
            >
              {handoverCopy("actions.printManifest")}
            </button>
            <button className="erp-button erp-button--secondary" type="button" onClick={handleMarkReady} disabled={!canMarkReady}>
              {handoverCopy("actions.readyToScan")}
            </button>
            <button className="erp-button erp-button--danger" type="button" onClick={handleReportManifestMissing} disabled={missingLines.length === 0}>
              {handoverCopy("actions.reportMissing")}
            </button>
            <button className="erp-button erp-button--danger" type="button" onClick={handleCancelManifest} disabled={!canCancel}>
              {handoverCopy("actions.cancelManifest")}
            </button>
            <button className="erp-button erp-button--primary" type="button" onClick={handleConfirmHandover} disabled={!canConfirmHandover}>
              {handoverCopy("actions.confirmHandover")}
            </button>
          </div>

          {feedback ? <small className="erp-shipping-feedback">{feedback}</small> : null}
        </div>

        <div className="erp-card erp-card--padded erp-shipping-exception-card">
          <div className="erp-section-header">
            <div>
              <h2 className="erp-section-title">{handoverCopy("sections.missingQueue")}</h2>
              <p className="erp-section-description">
                {selectedManifest ? manifestLocationLabel(selectedManifest) : handoverCopy("scan.noHandoverZone")}
              </p>
            </div>
            <StatusChip tone={missingLines.length === 0 ? "success" : "danger"}>
              {handoverCopy("open", { count: missingLines.length })}
            </StatusChip>
          </div>
          <DataTable
            columns={missingColumns}
            rows={missingLines}
            getRowKey={(row) => row.id}
            emptyState={
              <section className="erp-shipping-empty-state">
                <StatusChip tone="success">{handoverCopy("empty.ready")}</StatusChip>
                <strong>{handoverCopy("empty.allExpectedScanned")}</strong>
              </section>
            }
          />
        </div>
      </section>

      <section className="erp-shipping-detail-grid">
        <div className="erp-card erp-card--padded">
          <div className="erp-section-header">
            <h2 className="erp-section-title">{handoverCopy("sections.recentScans")}</h2>
            <StatusChip tone={scanIssues.length === 0 ? "success" : "danger"}>
              {handoverCopy("issues", { count: scanIssues.length })}
            </StatusChip>
          </div>
          {scanIssues.length > 0 ? (
            <section className="erp-shipping-scan-panel">
              <div className="erp-shipping-scan-meta">
                <StatusChip tone="danger">{handoverCopy("scan.issues")}</StatusChip>
                {scanResult ? (
                  <StatusChip tone={carrierManifestScanSeverityTone(scanResult.severity)}>
                    {scanResultLabel(scanResult.resultCode)}
                  </StatusChip>
                ) : null}
              </div>
              <ol className="erp-shipping-scan-list" aria-label={handoverCopy("scan.issuesLabel")}>
                {scanIssues.map((scan) => (
                  <li key={`issue-${scan.scanEvent.id}`}>
                    <StatusChip tone={carrierManifestScanSeverityTone(scan.severity)}>{scanResultLabel(scan.resultCode)}</StatusChip>
                    <strong>{scan.scanEvent.code}</strong>
                    <span>{scan.expectedManifestId ?? scanResultMessage(scan)}</span>
                  </li>
                ))}
              </ol>
            </section>
          ) : null}
          {recentScans.length > 0 ? (
            <ol className="erp-shipping-scan-list" aria-label={handoverCopy("scan.recentLabel")}>
              {recentScans.map((scan) => (
                <li key={scan.scanEvent.id}>
                  <StatusChip tone={carrierManifestScanSeverityTone(scan.severity)}>{scanResultLabel(scan.resultCode)}</StatusChip>
                  <strong>{scan.scanEvent.code}</strong>
                  <span>{scan.line?.orderNo ?? scan.expectedManifestId ?? scanResultMessage(scan)}</span>
                </li>
              ))}
            </ol>
          ) : (
            <section className="erp-shipping-empty-state">
              <StatusChip tone="normal">{handoverCopy("empty.idle")}</StatusChip>
              <strong>{handoverCopy("empty.noScansYet")}</strong>
            </section>
          )}
        </div>

        <div className="erp-card erp-card--padded">
          <div className="erp-section-header">
            <h2 className="erp-section-title">{handoverCopy("sections.shipmentLines")}</h2>
            <StatusChip tone={(selectedManifest?.summary.missingCount ?? 0) === 0 ? "success" : "danger"}>
              {handoverCopy("missingCount", { count: selectedManifest?.summary.missingCount ?? 0 })}
            </StatusChip>
          </div>
          <DataTable columns={lineColumns} rows={selectedManifest?.lines ?? []} getRowKey={(row) => row.id} />
          <div className="erp-shipping-add-box">
            <label className="erp-field">
              <span>{handoverCopy("facts.shipmentId")}</span>
              <input className="erp-input" value={shipmentId} onChange={(event) => setShipmentId(event.target.value)} />
            </label>
            <button className="erp-button erp-button--primary" type="button" onClick={handleAddShipment} disabled={actionBusy || !selectedManifest}>
              {handoverCopy("actions.addShipment")}
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

function optionValueFromParam<TValue extends string>(
  value: string | null,
  options: readonly { value: TValue; label: string }[]
): TValue | null {
  if (value === null) {
    return null;
  }

  return options.some((option) => option.value === value) ? (value as TValue) : null;
}

function handoverCopy(key: string, values?: Record<string, string | number>) {
  return t(`shipping.handover.${key}`, { values });
}

function statusLabel(status: CarrierManifestStatus) {
  return handoverCopy(`status.${status}`);
}

function scanResultLabel(resultCode: CarrierManifestScanResultCode) {
  return handoverCopy(`scanResult.${resultCode}`);
}

function scanResultMessage(result: CarrierManifestScanResult) {
  return t(`shipping.handover.scanMessage.${result.resultCode}`, { fallback: result.message });
}

function manifestWarehouseLabel(value: string, fallback: string) {
  if (value === "") {
    return handoverCopy("warehouse.all");
  }

  return t(`shipping.handover.warehouse.${value}`, { fallback });
}

function carrierOptionLabel(value: string, fallback: string) {
  return value === "" ? handoverCopy("filters.allCarriers") : fallback;
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
