"use client";

import { useEffect, useMemo, useRef, useState, type KeyboardEvent } from "react";
import { DataTable, StatusChip, type DataTableColumn, type StatusTone } from "@/shared/design-system/components";
import { useReturnReceipts } from "../hooks/useReturnReceipts";
import {
  formatReturnDisposition,
  receiveReturn,
  returnDispositionOptions,
  returnDispositionTone,
  returnReceiptStatusTone,
  returnSourceOptions,
  returnWarehouseOptions
} from "../services/returnReceivingService";
import type { ReturnDisposition, ReturnReceipt, ReturnReceiptQuery, ReturnSource } from "../types";

const receiptColumns: DataTableColumn<ReturnReceipt>[] = [
  {
    key: "receipt",
    header: "Receipt",
    render: (row) => (
      <span className="erp-returns-receipt-cell">
        <strong>{row.receiptNo}</strong>
        <small>{row.originalOrderNo ?? "Unknown order"}</small>
      </span>
    ),
    width: "190px"
  },
  {
    key: "tracking",
    header: "Tracking",
    render: (row) => row.trackingNo ?? row.scanCode,
    width: "160px"
  },
  {
    key: "source",
    header: "Source",
    render: (row) => row.source,
    width: "120px"
  },
  {
    key: "status",
    header: "Status",
    render: (row) => <StatusChip tone={returnReceiptStatusTone(row.status)}>{statusLabel(row.status)}</StatusChip>,
    width: "160px"
  },
  {
    key: "disposition",
    header: "Disposition",
    render: (row) => (
      <StatusChip tone={returnDispositionTone(row.disposition)}>{formatReturnDisposition(row.disposition)}</StatusChip>
    ),
    width: "150px"
  },
  {
    key: "target",
    header: "Target",
    render: (row) => row.targetLocation,
    width: "210px"
  },
  {
    key: "movement",
    header: "Movement",
    render: (row) => (row.stockMovement ? <StatusChip tone="success">{row.stockMovement.movementType}</StatusChip> : "-"),
    width: "150px"
  }
];

export function ReturnReceivingPrototype() {
  const scanInputRef = useRef<HTMLInputElement>(null);
  const [warehouseId, setWarehouseId] = useState("wh-hcm");
  const [source, setSource] = useState<ReturnSource>("CARRIER");
  const [packageCondition, setPackageCondition] = useState("sealed");
  const [disposition, setDisposition] = useState<ReturnDisposition>("needs_inspection");
  const [scanCode, setScanCode] = useState("");
  const [feedback, setFeedback] = useState<{ tone: StatusTone; message: string } | null>(null);
  const [localReceipts, setLocalReceipts] = useState<ReturnReceipt[]>([]);
  const [busy, setBusy] = useState(false);
  const query = useMemo<ReturnReceiptQuery>(
    () => ({
      warehouseId,
      status: "pending_inspection"
    }),
    [warehouseId]
  );
  const { receipts, loading } = useReturnReceipts(query);
  const displayedReceipts = useMemo(() => {
    const localMatches = localReceipts.filter((receipt) => matchesReceiptQuery(receipt, query));
    const localIds = new Set(localMatches.map((receipt) => receipt.id));

    return [...localMatches, ...receipts.filter((receipt) => !localIds.has(receipt.id))];
  }, [localReceipts, query, receipts]);
  const latestReceipt = displayedReceipts[0] ?? null;
  const warehouse = returnWarehouseOptions.find((option) => option.value === warehouseId) ?? returnWarehouseOptions[0];
  const totals = displayedReceipts.reduce(
    (acc, receipt) => ({
      pending: acc.pending + (receipt.status === "pending_inspection" ? 1 : 0),
      unknown: acc.unknown + (receipt.unknownCase ? 1 : 0),
      movements: acc.movements + (receipt.stockMovement ? 1 : 0),
      lab: acc.lab + (receipt.targetLocation === "lab-damaged-placeholder" ? 1 : 0)
    }),
    { pending: 0, unknown: 0, movements: 0, lab: 0 }
  );

  useEffect(() => {
    scanInputRef.current?.focus();
  }, [warehouseId, disposition]);

  async function handleConfirmReceipt() {
    const code = scanCode.trim();
    if (code === "" || busy) {
      return;
    }

    setBusy(true);
    try {
      const receipt = await receiveReturn({
        warehouseId: warehouse.value,
        warehouseCode: warehouse.code,
        source,
        code,
        packageCondition,
        disposition
      });
      setLocalReceipts((current) => [receipt, ...current.filter((candidate) => candidate.id !== receipt.id)]);
      setFeedback({
        tone: receipt.unknownCase ? "warning" : receipt.stockMovement ? "success" : "info",
        message: receipt.stockMovement
          ? `${receipt.receiptNo} / ${receipt.stockMovement.movementType}`
          : `${receipt.receiptNo} / ${receipt.targetLocation}`
      });
      setScanCode("");
      window.setTimeout(() => scanInputRef.current?.focus(), 0);
    } catch (error) {
      setFeedback({
        tone: "danger",
        message: error instanceof Error ? error.message : "Return receipt could not be created"
      });
      window.setTimeout(() => scanInputRef.current?.select(), 0);
    } finally {
      setBusy(false);
    }
  }

  function handleScanKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (event.key === "Enter") {
      void handleConfirmReceipt();
    }
  }

  return (
    <section className="erp-module-page erp-returns-page">
      <header className="erp-page-header">
        <div>
          <p className="erp-module-eyebrow">RT</p>
          <h1 className="erp-page-title">Returns</h1>
          <p className="erp-page-description">Return receiving, pending inspection, and disposition routing</p>
        </div>
        <div className="erp-page-actions">
          <a className="erp-button erp-button--secondary" href="#return-receiving">
            Receive
          </a>
          <a className="erp-button erp-button--primary" href="#return-receipts">
            Receipts
          </a>
        </div>
      </header>

      <section className="erp-returns-toolbar" aria-label="Return receiving context">
        <label className="erp-field">
          <span>Warehouse</span>
          <select className="erp-input" value={warehouseId} onChange={(event) => setWarehouseId(event.target.value)}>
            {returnWarehouseOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Source</span>
          <select className="erp-input" value={source} onChange={(event) => setSource(event.target.value as ReturnSource)}>
            {returnSourceOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Package condition</span>
          <input
            className="erp-input"
            type="text"
            value={packageCondition}
            onChange={(event) => setPackageCondition(event.target.value)}
          />
        </label>
      </section>

      <section className="erp-kpi-grid erp-returns-kpis">
        <ReturnKPI label="Pending inspection" value={totals.pending} tone="warning" />
        <ReturnKPI label="Unknown cases" value={totals.unknown} tone={totals.unknown > 0 ? "danger" : "normal"} />
        <ReturnKPI label="Return receipt moves" value={totals.movements} tone="success" />
        <ReturnKPI label="Lab placeholders" value={totals.lab} tone={totals.lab > 0 ? "danger" : "normal"} />
      </section>

      <section className="erp-returns-receiving-grid" id="return-receiving">
        <div className="erp-card erp-card--padded erp-returns-receiving-card">
          <div className="erp-section-header">
            <h2 className="erp-section-title">Return receiving</h2>
            <StatusChip tone={busy ? "warning" : "info"}>{warehouse.code}</StatusChip>
          </div>

          <div className="erp-returns-disposition-options" aria-label="Return disposition">
            {returnDispositionOptions.map((option) => (
              <button
                aria-pressed={disposition === option.value}
                className="erp-returns-disposition-option"
                data-active={disposition === option.value ? "true" : "false"}
                key={option.value}
                type="button"
                onClick={() => setDisposition(option.value)}
              >
                {option.label}
              </button>
            ))}
          </div>

          <label className={`erp-returns-scan-primary erp-returns-scan-primary--${feedback?.tone ?? "normal"}`}>
            <span>Order / tracking / return code</span>
            <input
              ref={scanInputRef}
              type="text"
              value={scanCode}
              placeholder="SO-260426-001 / GHN260426001 / RET-260426-001"
              onChange={(event) => setScanCode(event.target.value)}
              onKeyDown={handleScanKeyDown}
            />
            <small>{feedback ? feedback.message : `${formatReturnDisposition(disposition)} / ${source}`}</small>
          </label>

          <div className="erp-returns-actions">
            <button
              className="erp-button erp-button--primary"
              type="button"
              disabled={scanCode.trim() === "" || busy}
              onClick={() => void handleConfirmReceipt()}
            >
              Confirm receipt
            </button>
            <button
              className="erp-button erp-button--secondary"
              type="button"
              onClick={() => {
                setScanCode("");
                setFeedback(null);
                scanInputRef.current?.focus();
              }}
            >
              Clear
            </button>
          </div>
        </div>

        <div className="erp-card erp-card--padded erp-returns-result-card">
          <div className="erp-section-header">
            <h2 className="erp-section-title">Latest receipt</h2>
            {latestReceipt ? (
              <StatusChip tone={returnDispositionTone(latestReceipt.disposition)}>
                {formatReturnDisposition(latestReceipt.disposition)}
              </StatusChip>
            ) : null}
          </div>

          {latestReceipt ? (
            <div className="erp-returns-result-grid">
              <ReturnFact label="Receipt" value={latestReceipt.receiptNo} />
              <ReturnFact label="Order" value={latestReceipt.originalOrderNo ?? "Unknown"} />
              <ReturnFact label="Tracking" value={latestReceipt.trackingNo ?? latestReceipt.scanCode} />
              <ReturnFact label="Target" value={latestReceipt.targetLocation} />
              <ReturnFact label="SKU" value={latestReceipt.lines[0]?.sku ?? "-"} />
              <ReturnFact label="Movement" value={latestReceipt.stockMovement?.movementType ?? "-"} />
            </div>
          ) : (
            <div className="erp-returns-empty-state">No return receipt selected</div>
          )}
        </div>
      </section>

      <section className="erp-card erp-card--padded erp-module-table-card" id="return-receipts">
        <div className="erp-section-header">
          <h2 className="erp-section-title">Return receipts</h2>
          <StatusChip tone={displayedReceipts.length === 0 ? "warning" : "info"}>{displayedReceipts.length} rows</StatusChip>
        </div>
        <DataTable
          columns={receiptColumns}
          rows={displayedReceipts}
          getRowKey={(row) => row.id}
          loading={loading}
        />
      </section>
    </section>
  );
}

function ReturnKPI({
  label,
  value,
  tone
}: {
  label: string;
  value: number;
  tone: "normal" | "success" | "warning" | "danger";
}) {
  return (
    <article className="erp-card erp-card--padded erp-kpi-card">
      <div className="erp-kpi-label">{label}</div>
      <strong className="erp-kpi-value">{formatQuantity(value)}</strong>
      <StatusChip tone={tone}>{label}</StatusChip>
    </article>
  );
}

function ReturnFact({ label, value }: { label: string; value: string }) {
  return (
    <div className="erp-returns-fact">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function matchesReceiptQuery(receipt: ReturnReceipt, query: ReturnReceiptQuery) {
  if (query.warehouseId && receipt.warehouseId !== query.warehouseId) {
    return false;
  }
  if (query.status && receipt.status !== query.status) {
    return false;
  }

  return true;
}

function statusLabel(status: string) {
  if (status === "pending_inspection") {
    return "Pending inspection";
  }

  return status;
}

function formatQuantity(value: number) {
  return new Intl.NumberFormat("en-US", { maximumFractionDigits: 0 }).format(value);
}
