"use client";

import { useMemo, useState } from "react";
import type { ReactNode } from "react";
import {
  DataTable,
  EmptyState,
  QuantityDisplay,
  StatusChip,
  type DataTableColumn,
  type StatusTone
} from "@/shared/design-system/components";
import { formatDateTimeVI, formatDateVI } from "@/shared/format/numberFormat";
import { useOperationsDailyReport } from "../hooks/useOperationsDailyReport";
import { downloadOperationsDailyCSV, operationsDailyStatusOptions } from "../services/operationsDailyReportService";
import type {
  OperationsDailyAreaSummary,
  OperationsDailyQuery,
  OperationsDailyReport,
  OperationsDailyRow
} from "../types";

const warehouseOptions = [
  { label: "All warehouses", value: "" },
  { label: "HCM", value: "wh-hcm" },
  { label: "HN", value: "wh-hn" }
];

const areaColumns: DataTableColumn<OperationsDailyAreaSummary>[] = [
  {
    key: "area",
    header: "Area",
    render: (row) => <StatusChip tone={areaTone(row.area)}>{areaLabel(row.area)}</StatusChip>,
    width: "160px"
  },
  {
    key: "signals",
    header: "Signals",
    render: (row) => row.signalCount,
    align: "right"
  },
  {
    key: "pending",
    header: "Pending",
    render: (row) => row.pendingCount,
    align: "right"
  },
  {
    key: "progress",
    header: "In progress",
    render: (row) => row.inProgressCount,
    align: "right"
  },
  {
    key: "blocked",
    header: "Blocked",
    render: (row) => row.blockedCount,
    align: "right"
  },
  {
    key: "exceptions",
    header: "Exceptions",
    render: (row) => row.exceptionCount,
    align: "right"
  },
  {
    key: "completed",
    header: "Completed",
    render: (row) => row.completedCount,
    align: "right"
  }
];

const rowColumns: DataTableColumn<OperationsDailyRow>[] = [
  {
    key: "area",
    header: "Area",
    render: (row) => <StatusChip tone={areaTone(row.area)}>{areaLabel(row.area)}</StatusChip>,
    width: "150px"
  },
  {
    key: "ref",
    header: "Reference",
    render: (row) => (
      <span className="erp-reporting-item-cell">
        <OperationsSourceReference row={row} />
        <small>{sourceLabel(row.sourceType)}</small>
      </span>
    ),
    width: "190px"
  },
  {
    key: "title",
    header: "Work item",
    render: (row) => (
      <span className="erp-reporting-item-cell">
        <strong>{row.title}</strong>
        <small>{row.sourceId}</small>
      </span>
    ),
    width: "280px"
  },
  {
    key: "warehouse",
    header: "Warehouse",
    render: (row) => row.warehouseCode || row.warehouseId,
    width: "110px"
  },
  {
    key: "date",
    header: "Date",
    render: (row) => formatDateVI(row.businessDate),
    width: "120px"
  },
  {
    key: "quantity",
    header: "Quantity",
    render: (row) =>
      row.quantity ? <QuantityDisplay value={row.quantity} uomCode={row.uomCode} /> : <span className="erp-muted">-</span>,
    align: "right",
    width: "130px"
  },
  {
    key: "status",
    header: "Status",
    render: (row) => (
      <span className="erp-reporting-chip-stack">
        <StatusChip tone={statusTone(row.status)}>{statusLabel(row.status)}</StatusChip>
        {row.severity !== "normal" ? <StatusChip tone={severityTone(row.severity)}>{severityLabel(row.severity)}</StatusChip> : null}
      </span>
    ),
    width: "180px"
  },
  {
    key: "owner",
    header: "Owner",
    render: (row) => row.owner || "-",
    width: "130px"
  },
  {
    key: "exception",
    header: "Exception",
    render: (row) => row.exceptionCode || "-",
    width: "190px",
    sticky: true
  }
];

function OperationsSourceReference({ row }: { row: OperationsDailyRow }) {
  if (!row.sourceReference.unavailable && row.sourceReference.href) {
    return (
      <a className="erp-reporting-source-link" href={row.sourceReference.href} aria-label={`Open ${row.refNo}`}>
        {row.refNo}
      </a>
    );
  }

  return <strong>{row.refNo}</strong>;
}

type OperationsDailyReportPanelProps = {
  controls?: ReactNode;
};

export function OperationsDailyReportPanel({ controls }: OperationsDailyReportPanelProps = {}) {
  const [fromDate, setFromDate] = useState(defaultBusinessDate());
  const [toDate, setToDate] = useState(defaultBusinessDate());
  const [warehouseId, setWarehouseId] = useState("wh-hcm");
  const [status, setStatus] = useState<OperationsDailyQuery["status"]>("");
  const [exporting, setExporting] = useState(false);
  const [exportError, setExportError] = useState<Error | null>(null);

  const query = useMemo<OperationsDailyQuery>(
    () => ({
      fromDate,
      toDate,
      businessDate: toDate,
      warehouseId: warehouseId || undefined,
      status: status || undefined
    }),
    [fromDate, status, toDate, warehouseId]
  );
  const { report, loading, error } = useOperationsDailyReport(query);
  const data = report ?? emptyOperationsDailyReport(fromDate, toDate);

  async function handleExportCSV() {
    setExporting(true);
    setExportError(null);
    try {
      const download = await downloadOperationsDailyCSV(query);
      saveBlob(download.blob, download.filename);
    } catch (reason) {
      setExportError(reason instanceof Error ? reason : new Error("Operations CSV could not be exported"));
    } finally {
      setExporting(false);
    }
  }

  return (
    <section className="erp-module-page erp-reporting-page">
      <header className="erp-page-header">
        <div>
          <p className="erp-module-eyebrow">RP</p>
          <h1 className="erp-page-title">Reporting</h1>
          <p className="erp-page-description">Daily operations dashboard by warehouse, status, and exception signal</p>
        </div>
        <div className="erp-page-actions">
          {controls}
          {exportError ? <StatusChip tone="danger">Export failed</StatusChip> : null}
          <button
            className="erp-button erp-button--secondary"
            type="button"
            disabled={loading || exporting}
            title={exportError?.message ?? "Export operations daily CSV"}
            onClick={handleExportCSV}
          >
            {exporting ? "Exporting" : "Export CSV"}
          </button>
        </div>
      </header>

      <section className="erp-reporting-toolbar erp-reporting-toolbar--operations" aria-label="Operations daily filters">
        <label className="erp-field">
          <span>From date</span>
          <input className="erp-input" type="date" value={fromDate} onChange={(event) => setFromDate(event.target.value)} />
        </label>
        <label className="erp-field">
          <span>To date</span>
          <input className="erp-input" type="date" value={toDate} onChange={(event) => setToDate(event.target.value)} />
        </label>
        <label className="erp-field">
          <span>Warehouse</span>
          <select className="erp-input" value={warehouseId} onChange={(event) => setWarehouseId(event.target.value)}>
            {warehouseOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Status</span>
          <select
            className="erp-input"
            value={status}
            onChange={(event) => setStatus(event.target.value as OperationsDailyQuery["status"])}
          >
            {operationsDailyStatusOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
      </section>

      <section className="erp-kpi-grid erp-reporting-kpis">
        <OperationsDailyKPI label="Signals" value={String(data.summary.signalCount)} tone="info" />
        <OperationsDailyKPI label="Pending" value={String(data.summary.pendingCount)} tone="warning" />
        <OperationsDailyKPI label="In progress" value={String(data.summary.inProgressCount)} tone="info" />
        <OperationsDailyKPI label="Blocked" value={String(data.summary.blockedCount)} tone="danger" />
        <OperationsDailyKPI label="Exceptions" value={String(data.summary.exceptionCount)} tone="danger" />
        <OperationsDailyKPI label="Completed" value={String(data.summary.completedCount)} tone="success" />
      </section>

      <section className="erp-card erp-card--padded erp-module-table-card">
        <div className="erp-section-header">
          <div>
            <h2 className="erp-section-title">Area summary</h2>
            <p className="erp-section-description">{metadataLabel(data)}</p>
          </div>
          <StatusChip tone={error ? "danger" : loading ? "warning" : "info"}>{reportStatusLabel({ error, loading })}</StatusChip>
        </div>
        <DataTable
          columns={areaColumns}
          rows={data.areas}
          getRowKey={(row) => row.area}
          loading={loading}
          error={error?.message}
          emptyState={<EmptyState title="No area signals" description="No operations match the selected filters." />}
        />
      </section>

      <section className="erp-card erp-card--padded erp-module-table-card">
        <div className="erp-section-header">
          <h2 className="erp-section-title">Operations rows</h2>
          <StatusChip tone={data.rows.length === 0 ? "warning" : "info"}>{data.rows.length} rows</StatusChip>
        </div>
        <DataTable
          columns={rowColumns}
          rows={data.rows}
          getRowKey={(row) => row.id}
          loading={loading}
          error={error?.message}
          emptyState={<EmptyState title="No operations rows" description="No operations match the selected filters." />}
        />
      </section>
    </section>
  );
}

function OperationsDailyKPI({ label, value, tone }: { label: string; value: string; tone: StatusTone }) {
  return (
    <article className="erp-card erp-card--padded erp-kpi-card">
      <span className="erp-kpi-label">{label}</span>
      <strong className="erp-kpi-value">{value}</strong>
      <StatusChip tone={tone}>{label}</StatusChip>
    </article>
  );
}

function metadataLabel(report: OperationsDailyReport) {
  const generatedAt = formatDateTimeVI(report.metadata.generatedAt);
  return `${report.metadata.filters.fromDate} -> ${report.metadata.filters.toDate} / ${report.metadata.timezone} / ${generatedAt}`;
}

function reportStatusLabel({ error, loading }: { error: Error | null; loading: boolean }) {
  if (error) {
    return "API error";
  }
  if (loading) {
    return "Loading";
  }

  return "Live";
}

function areaLabel(value: string) {
  switch (value) {
    case "inbound":
      return "Inbound";
    case "qc":
      return "QC";
    case "outbound":
      return "Outbound";
    case "returns":
      return "Returns";
    case "stock_count":
      return "Stock count";
    case "subcontract":
      return "Subcontract";
    default:
      return value || "Unknown";
  }
}

function areaTone(value: string): StatusTone {
  switch (value) {
    case "qc":
    case "stock_count":
      return "warning";
    case "outbound":
      return "info";
    case "returns":
      return "danger";
    case "inbound":
    case "subcontract":
      return "success";
    default:
      return "normal";
  }
}

function statusLabel(value: string) {
  switch (value) {
    case "pending":
      return "Pending";
    case "in_progress":
      return "In progress";
    case "completed":
      return "Completed";
    case "blocked":
      return "Blocked";
    case "exception":
      return "Exception";
    default:
      return value || "Unknown";
  }
}

function statusTone(value: string): StatusTone {
  switch (value) {
    case "completed":
      return "success";
    case "pending":
    case "in_progress":
      return "warning";
    case "blocked":
    case "exception":
      return "danger";
    default:
      return "normal";
  }
}

function severityLabel(value: string) {
  switch (value) {
    case "warning":
      return "Warning";
    case "danger":
      return "Danger";
    default:
      return "Normal";
  }
}

function severityTone(value: string): StatusTone {
  return value === "danger" ? "danger" : value === "warning" ? "warning" : "normal";
}

function sourceLabel(value: string) {
  return value.replaceAll("_", " ");
}

function emptyOperationsDailyReport(fromDate: string, toDate: string): OperationsDailyReport {
  return {
    metadata: {
      generatedAt: new Date().toISOString(),
      timezone: "Asia/Ho_Chi_Minh",
      sourceVersion: "reporting-v1",
      filters: {
        fromDate,
        toDate,
        businessDate: toDate
      }
    },
    summary: {
      signalCount: 0,
      pendingCount: 0,
      inProgressCount: 0,
      completedCount: 0,
      blockedCount: 0,
      exceptionCount: 0
    },
    areas: [],
    rows: []
  };
}

function defaultBusinessDate() {
  return new Date().toISOString().slice(0, 10);
}

function saveBlob(blob: Blob, filename: string) {
  const url = window.URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = filename;
  anchor.rel = "noreferrer";
  document.body.append(anchor);
  anchor.click();
  anchor.remove();
  window.setTimeout(() => window.URL.revokeObjectURL(url), 0);
}
