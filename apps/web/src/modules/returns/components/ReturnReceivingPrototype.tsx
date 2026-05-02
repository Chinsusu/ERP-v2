"use client";

import { useEffect, useMemo, useRef, useState, type KeyboardEvent } from "react";
import { DataTable, StatusChip, type DataTableColumn, type StatusTone } from "@/shared/design-system/components";
import { t } from "@/shared/i18n";
import { useReturnReceipts } from "../hooks/useReturnReceipts";
import {
  receiveReturn,
  returnDispositionOptions,
  returnDispositionTone,
  returnReceiptStatusTone,
  returnSourceOptions,
  returnWarehouseOptions
} from "../services/returnReceivingService";
import type { ReturnDisposition, ReturnReceipt, ReturnReceiptQuery, ReturnSource } from "../types";
import { ReturnInspectionPanel } from "./ReturnInspectionPanel";
import { SupplierRejectionPanel } from "./SupplierRejectionPanel";

type ScanFeedback = {
  tone: StatusTone;
  result: "PASS" | "CHECK" | "FAIL";
  message: string;
};

const receiptColumns: DataTableColumn<ReturnReceipt>[] = [
  {
    key: "receipt",
    header: returnReceivingCopy("columns.receipt"),
    render: (row) => (
      <span className="erp-returns-receipt-cell">
        <strong>{row.receiptNo}</strong>
        <small>{row.originalOrderNo ?? returnReceivingCopy("empty.unknownOrder")}</small>
      </span>
    ),
    width: "190px"
  },
  {
    key: "tracking",
    header: returnReceivingCopy("columns.tracking"),
    render: (row) => row.trackingNo ?? row.scanCode,
    width: "160px"
  },
  {
    key: "source",
    header: returnReceivingCopy("columns.source"),
    render: (row) => returnSourceLabel(row.source),
    width: "120px"
  },
  {
    key: "status",
    header: returnReceivingCopy("columns.status"),
    render: (row) => <StatusChip tone={returnReceiptStatusTone(row.status)}>{statusLabel(row.status)}</StatusChip>,
    width: "160px"
  },
  {
    key: "disposition",
    header: returnReceivingCopy("columns.disposition"),
    render: (row) => (
      <StatusChip tone={returnDispositionTone(row.disposition)}>{returnDispositionLabel(row.disposition)}</StatusChip>
    ),
    width: "150px"
  },
  {
    key: "target",
    header: returnReceivingCopy("columns.target"),
    render: (row) => returnReceivingTargetLabel(row.targetLocation),
    width: "210px"
  },
  {
    key: "movement",
    header: returnReceivingCopy("columns.movement"),
    render: (row) => (row.stockMovement ? <StatusChip tone="success">{returnMovementLabel(row.stockMovement.movementType)}</StatusChip> : "-"),
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
  const [feedback, setFeedback] = useState<ScanFeedback | null>(null);
  const [localReceipts, setLocalReceipts] = useState<ReturnReceipt[]>([]);
  const [busy, setBusy] = useState(false);
  const query = useMemo<ReturnReceiptQuery>(
    () => ({
      warehouseId
    }),
    [warehouseId]
  );
  const { receipts, loading, error } = useReturnReceipts(query);
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
        tone: receipt.unknownCase ? "warning" : "success",
        result: receipt.unknownCase ? "CHECK" : "PASS",
        message: receipt.unknownCase
          ? returnReceivingCopy("scan.unknown", { receiptNo: receipt.receiptNo })
          : returnReceivingCopy("scan.matched", {
              receiptNo: receipt.receiptNo,
              code: receipt.originalOrderNo ?? receipt.trackingNo ?? receipt.scanCode
            })
      });
      setScanCode("");
      window.setTimeout(() => scanInputRef.current?.focus(), 0);
    } catch (error) {
      setFeedback({
        tone: "danger",
        result: "FAIL",
        message: error instanceof Error ? error.message : returnReceivingCopy("scan.failed")
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
          <h1 className="erp-page-title">{returnReceivingCopy("title")}</h1>
          <p className="erp-page-description">{returnReceivingCopy("description")}</p>
        </div>
        <div className="erp-page-actions">
          <a className="erp-button erp-button--secondary" href="#return-receiving">
            {returnReceivingCopy("actions.receive")}
          </a>
          <a className="erp-button erp-button--secondary" href="#return-inspection">
            {returnReceivingCopy("actions.inspect")}
          </a>
          <a className="erp-button erp-button--secondary" href="#supplier-rejections">
            {returnReceivingCopy("actions.supplierRejects")}
          </a>
          <a className="erp-button erp-button--primary" href="#return-receipts">
            {returnReceivingCopy("actions.receipts")}
          </a>
        </div>
      </header>

      <section className="erp-returns-toolbar" aria-label={returnReceivingCopy("filters.context")}>
        <label className="erp-field">
          <span>{returnReceivingCopy("filters.warehouse")}</span>
          <select className="erp-input" value={warehouseId} onChange={(event) => setWarehouseId(event.target.value)}>
            {returnWarehouseOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {returnWarehouseLabel(option.value, option.label)}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>{returnReceivingCopy("filters.source")}</span>
          <select className="erp-input" value={source} onChange={(event) => setSource(event.target.value as ReturnSource)}>
            {returnSourceOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {returnSourceLabel(option.value)}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>{returnReceivingCopy("filters.packageCondition")}</span>
          <input
            className="erp-input"
            type="text"
            value={packageCondition}
            onChange={(event) => setPackageCondition(event.target.value)}
          />
        </label>
      </section>

      <section className="erp-kpi-grid erp-returns-kpis">
        <ReturnKPI label={returnReceivingCopy("kpi.pendingInspection")} value={totals.pending} tone="warning" />
        <ReturnKPI label={returnReceivingCopy("kpi.unknownCases")} value={totals.unknown} tone={totals.unknown > 0 ? "danger" : "normal"} />
        <ReturnKPI label={returnReceivingCopy("kpi.returnReceiptMoves")} value={totals.movements} tone="success" />
        <ReturnKPI label={returnReceivingCopy("kpi.labPlaceholders")} value={totals.lab} tone={totals.lab > 0 ? "danger" : "normal"} />
      </section>

      <section className="erp-returns-receiving-grid" id="return-receiving">
        <div className="erp-card erp-card--padded erp-returns-receiving-card">
          <div className="erp-section-header">
            <h2 className="erp-section-title">{returnReceivingCopy("sections.returnScan")}</h2>
            <StatusChip tone={busy ? "warning" : "info"}>{warehouse.code}</StatusChip>
          </div>

          <div className="erp-returns-disposition-options" aria-label={returnReceivingCopy("columns.disposition")}>
            {returnDispositionOptions.map((option) => (
              <button
                aria-pressed={disposition === option.value}
                className="erp-returns-disposition-option"
                data-active={disposition === option.value ? "true" : "false"}
                key={option.value}
                type="button"
                onClick={() => setDisposition(option.value)}
              >
                {returnDispositionLabel(option.value)}
              </button>
            ))}
          </div>

          <label className={`erp-returns-scan-primary erp-returns-scan-primary--${feedback?.tone ?? "normal"}`}>
            <span>{returnReceivingCopy("scan.label")}</span>
            <input
              ref={scanInputRef}
              type="text"
              value={scanCode}
              autoComplete="off"
              disabled={busy}
              aria-invalid={feedback?.result === "FAIL" ? "true" : undefined}
              placeholder={returnReceivingCopy("scan.placeholder")}
              onChange={(event) => setScanCode(event.target.value)}
              onKeyDown={handleScanKeyDown}
            />
            <span className="erp-returns-scan-feedback" role="status" aria-live="polite">
              {feedback ? <StatusChip tone={feedback.tone}>{returnScanResultLabel(feedback.result)}</StatusChip> : null}
              <small>
                {feedback
                  ? feedback.message
                  : returnReceivingCopy("scan.defaultHint", {
                      disposition: returnDispositionLabel(disposition),
                      source: returnSourceLabel(source)
                    })}
              </small>
            </span>
          </label>

          <div className="erp-returns-actions">
            <button
              className="erp-button erp-button--primary"
              type="button"
              disabled={scanCode.trim() === "" || busy}
              onClick={() => void handleConfirmReceipt()}
            >
              {returnReceivingCopy("actions.confirmReceipt")}
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
              {returnReceivingCopy("actions.clear")}
            </button>
          </div>
        </div>

        <div className="erp-card erp-card--padded erp-returns-result-card">
          <div className="erp-section-header">
            <h2 className="erp-section-title">{returnReceivingCopy("sections.latestReceipt")}</h2>
            {latestReceipt ? (
              <StatusChip tone={returnDispositionTone(latestReceipt.disposition)}>
                {returnDispositionLabel(latestReceipt.disposition)}
              </StatusChip>
            ) : null}
          </div>

          {latestReceipt ? (
            <div className="erp-returns-result-grid">
              <ReturnFact label={returnReceivingCopy("facts.receipt")} value={latestReceipt.receiptNo} />
              <ReturnFact label={returnReceivingCopy("facts.order")} value={latestReceipt.originalOrderNo ?? returnReceivingCopy("empty.unknown")} />
              <ReturnFact label={returnReceivingCopy("facts.tracking")} value={latestReceipt.trackingNo ?? latestReceipt.scanCode} />
              <ReturnFact label={returnReceivingCopy("facts.target")} value={returnReceivingTargetLabel(latestReceipt.targetLocation)} />
              <ReturnFact label={returnReceivingCopy("facts.sku")} value={latestReceipt.lines[0]?.sku ?? "-"} />
              <ReturnFact label={returnReceivingCopy("facts.movement")} value={returnMovementLabel(latestReceipt.stockMovement?.movementType)} />
            </div>
          ) : (
            <div className="erp-returns-empty-state">{returnReceivingCopy("empty.noReceiptSelected")}</div>
          )}
        </div>
      </section>

      <ReturnInspectionPanel
        receipts={displayedReceipts}
        onReceiptChange={(receipt) => {
          setLocalReceipts((current) => [receipt, ...current.filter((candidate) => candidate.id !== receipt.id)]);
        }}
      />

      <SupplierRejectionPanel />

      <section className="erp-card erp-card--padded erp-module-table-card" id="return-receipts">
        <div className="erp-section-header">
          <h2 className="erp-section-title">{returnReceivingCopy("sections.returnReceipts")}</h2>
          <StatusChip tone={displayedReceipts.length === 0 ? "warning" : "info"}>
            {returnReceivingCopy("rows", { count: displayedReceipts.length })}
          </StatusChip>
        </div>
        <DataTable
          columns={receiptColumns}
          rows={displayedReceipts}
          getRowKey={(row) => row.id}
          loading={loading}
          error={error?.message}
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
  return returnReceivingCopy(`status.${status}`);
}

function formatQuantity(value: number) {
  return new Intl.NumberFormat("vi-VN", { maximumFractionDigits: 0 }).format(value);
}

function returnReceivingCopy(key: string, values?: Record<string, string | number>) {
  return t(`returns.receiving.${key}`, { values });
}

function returnDispositionLabel(disposition: ReturnDisposition) {
  return returnReceivingCopy(`disposition.${disposition}`);
}

function returnSourceLabel(source: ReturnSource) {
  return returnReceivingCopy(`source.${source}`);
}

function returnWarehouseLabel(value: string, fallback: string) {
  return t(`returns.receiving.warehouse.${value}`, { fallback });
}

function returnScanResultLabel(result: ScanFeedback["result"]) {
  return returnReceivingCopy(`result.${result}`);
}

function returnMovementLabel(movementType: string | undefined) {
  return movementType ? t(`returns.receiving.movementType.${movementType}`, { fallback: movementType }) : "-";
}

function returnReceivingTargetLabel(targetLocation: string) {
  return t(`returns.receiving.targetLocation.${targetLocation}`, { fallback: targetLocation });
}
