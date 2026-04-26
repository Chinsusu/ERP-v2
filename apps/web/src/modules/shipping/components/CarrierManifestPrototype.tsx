"use client";

import { useMemo, useState } from "react";
import { DataTable, StatusChip, type DataTableColumn } from "@/shared/design-system/components";
import { useCarrierManifests } from "../hooks/useCarrierManifests";
import {
  addShipmentToManifest,
  carrierManifestStatusOptions,
  carrierManifestStatusTone,
  carrierOptions,
  createCarrierManifest,
  defaultCarrierManifestDate,
  manifestWarehouseOptions
} from "../services/carrierManifestService";
import type { CarrierManifest, CarrierManifestLine, CarrierManifestQuery, CarrierManifestStatus } from "../types";

const manifestColumns: DataTableColumn<CarrierManifest>[] = [
  {
    key: "manifest",
    header: "Manifest",
    render: (row) => (
      <span className="erp-shipping-manifest-cell">
        <strong>{row.id}</strong>
        <small>{row.handoverBatch} / {row.stagingZone}</small>
      </span>
    ),
    width: "260px"
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
    width: "120px"
  },
  {
    key: "expected",
    header: "Expected",
    render: (row) => row.summary.expectedCount,
    align: "right",
    width: "100px"
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
  }
];

const lineColumns: DataTableColumn<CarrierManifestLine>[] = [
  { key: "order", header: "Order", render: (row) => row.orderNo, width: "150px" },
  { key: "tracking", header: "Tracking", render: (row) => row.trackingNo, width: "160px" },
  { key: "package", header: "Package", render: (row) => row.packageCode, width: "120px" },
  { key: "zone", header: "Zone", render: (row) => row.stagingZone, width: "130px" },
  {
    key: "scan",
    header: "Scan state",
    render: (row) => <StatusChip tone={row.scanned ? "success" : "warning"}>{row.scanned ? "Scanned" : "Missing"}</StatusChip>,
    width: "120px"
  }
];

export function CarrierManifestPrototype() {
  const [warehouseId, setWarehouseId] = useState("");
  const [date, setDate] = useState(defaultCarrierManifestDate);
  const [carrierCode, setCarrierCode] = useState("");
  const [status, setStatus] = useState<"" | CarrierManifestStatus>("");
  const [selectedManifestId, setSelectedManifestId] = useState("manifest-hcm-ghn-morning");
  const [shipmentId, setShipmentId] = useState("ship-hcm-260426-004");
  const [feedback, setFeedback] = useState("");
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
  const { manifests, loading } = useCarrierManifests(query);
  const displayedManifests = useMemo(() => {
    const localMatches = localManifests.filter((manifest) => matchesManifestQuery(manifest, query));
    const localIds = new Set(localMatches.map((manifest) => manifest.id));

    return [...localMatches, ...manifests.filter((manifest) => !localIds.has(manifest.id))];
  }, [localManifests, manifests, query]);
  const selectedManifest =
    displayedManifests.find((manifest) => manifest.id === selectedManifestId) ?? displayedManifests[0] ?? null;
  const totals = displayedManifests.reduce(
    (acc, manifest) => ({
      expected: acc.expected + manifest.summary.expectedCount,
      scanned: acc.scanned + manifest.summary.scannedCount,
      missing: acc.missing + manifest.summary.missingCount
    }),
    { expected: 0, scanned: 0, missing: 0 }
  );

  async function handleCreateManifest() {
    const warehouseOption = manifestWarehouseOptions.find((option) => option.value === (warehouseId || "wh-hcm"));
    const carrierOption = carrierOptions.find((option) => option.value === (carrierCode || "GHN"));
    const created = await createCarrierManifest({
      carrierCode: carrierOption?.value || "GHN",
      carrierName: carrierOption?.label || "GHN",
      warehouseId: warehouseOption?.value || "wh-hcm",
      warehouseCode: warehouseOption && "code" in warehouseOption ? warehouseOption.code : "HCM",
      date
    });
    setLocalManifests((current) => [created, ...current.filter((manifest) => manifest.id !== created.id)]);
    setSelectedManifestId(created.id);
    setFeedback(`Created ${created.id}`);
  }

  async function handleAddShipment() {
    if (!selectedManifest || shipmentId.trim() === "") {
      return;
    }
    const updated = await addShipmentToManifest(selectedManifest.id, shipmentId.trim());
    setLocalManifests((current) => [updated, ...current.filter((manifest) => manifest.id !== updated.id)]);
    setFeedback(`Added ${shipmentId.trim()} to ${updated.id}`);
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
          <button className="erp-button erp-button--primary" type="button" onClick={handleCreateManifest}>
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
        <ShippingKPI label="Manifests" value={manifests.length} tone="normal" />
      </section>

      <section className="erp-card erp-card--padded erp-module-table-card">
        <div className="erp-section-header">
          <h2 className="erp-section-title">Manifest batches</h2>
          <StatusChip tone={displayedManifests.length === 0 ? "warning" : "info"}>{displayedManifests.length} rows</StatusChip>
        </div>
        <DataTable columns={manifestColumns} rows={displayedManifests} getRowKey={(row) => row.id} loading={loading} />
      </section>

      <section className="erp-shipping-detail-grid">
        <div className="erp-card erp-card--padded">
          <div className="erp-section-header">
            <div>
              <h2 className="erp-section-title">Selected manifest</h2>
              <p className="erp-section-description">{selectedManifest?.id ?? "No manifest selected"}</p>
            </div>
            {selectedManifest ? (
              <StatusChip tone={carrierManifestStatusTone(selectedManifest.status)}>{statusLabel(selectedManifest.status)}</StatusChip>
            ) : null}
          </div>
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
          <div className="erp-shipping-add-box">
            <label className="erp-field">
              <span>Shipment ID</span>
              <input className="erp-input" value={shipmentId} onChange={(event) => setShipmentId(event.target.value)} />
            </label>
            <button className="erp-button erp-button--primary" type="button" onClick={handleAddShipment}>
              Add shipment
            </button>
          </div>
          {feedback ? <small className="erp-shipping-feedback">{feedback}</small> : null}
        </div>

        <div className="erp-card erp-card--padded">
          <div className="erp-section-header">
            <h2 className="erp-section-title">Shipment lines</h2>
            <StatusChip tone={(selectedManifest?.summary.missingCount ?? 0) === 0 ? "success" : "danger"}>
              {selectedManifest?.summary.missingCount ?? 0} missing
            </StatusChip>
          </div>
          <DataTable columns={lineColumns} rows={selectedManifest?.lines ?? []} getRowKey={(row) => row.id} />
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

function statusLabel(status: CarrierManifestStatus) {
  switch (status) {
    case "completed":
      return "Completed";
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
