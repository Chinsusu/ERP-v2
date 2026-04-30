"use client";

import { useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";
import {
  DataTable,
  EmptyState,
  QuantityDisplay,
  StatusChip,
  type DataTableColumn,
  type StatusTone
} from "@/shared/design-system/components";
import { formatDateTimeVI, formatDateVI, formatQuantity } from "@/shared/format/numberFormat";
import { useInventorySnapshotReport } from "../hooks/useInventorySnapshotReport";
import { urlDateParam, urlOptionParam, urlParam, useReportUrlState } from "../hooks/useReportUrlState";
import {
  downloadInventorySnapshotCSV,
  inventorySnapshotCSVFilename,
  inventorySnapshotQueryString,
  inventorySnapshotStatusOptions
} from "../services/inventorySnapshotReportService";
import { ReportExportAction, ReportStateBanner } from "./ReportSharedStates";
import type {
  InventorySnapshotQuery,
  InventorySnapshotReport,
  InventorySnapshotRow,
  InventorySnapshotUOMTotal
} from "../types";

const warehouseOptions = [
  { label: "All warehouses", value: "" },
  { label: "HCM", value: "wh-hcm" },
  { label: "HN", value: "wh-hn" }
];

const rowColumns: DataTableColumn<InventorySnapshotRow>[] = [
  {
    key: "warehouse",
    header: "Warehouse",
    render: (row) => row.warehouseCode || row.warehouseId,
    width: "110px"
  },
  {
    key: "location",
    header: "Location",
    render: (row) => row.locationCode || row.locationId || "-",
    width: "110px"
  },
  {
    key: "sku",
    header: "Item / SKU",
    render: (row) => (
      <span className="erp-reporting-item-cell">
        <strong>{row.sku}</strong>
        <small>{row.itemId || "-"}</small>
      </span>
    ),
    width: "210px"
  },
  {
    key: "batch",
    header: "Batch",
    render: (row) => (
      <span className="erp-reporting-item-cell">
        <strong>{row.batchNo || "-"}</strong>
        <small>{row.batchExpiry ? formatDateVI(row.batchExpiry) : "-"}</small>
      </span>
    ),
    width: "150px"
  },
  {
    key: "physical",
    header: "Physical",
    render: (row) => <QuantityDisplay value={row.physicalQty} uomCode={row.baseUomCode} />,
    align: "right",
    width: "130px"
  },
  {
    key: "reserved",
    header: "Reserved",
    render: (row) => <QuantityDisplay value={row.reservedQty} uomCode={row.baseUomCode} />,
    align: "right",
    width: "130px"
  },
  {
    key: "quarantine",
    header: "Quarantine",
    render: (row) => <QuantityDisplay value={row.quarantineQty} uomCode={row.baseUomCode} />,
    align: "right",
    width: "140px"
  },
  {
    key: "blocked",
    header: "Blocked",
    render: (row) => <QuantityDisplay value={row.blockedQty} uomCode={row.baseUomCode} />,
    align: "right",
    width: "130px"
  },
  {
    key: "available",
    header: "Available",
    render: (row) => <QuantityDisplay value={row.availableQty} uomCode={row.baseUomCode} />,
    align: "right",
    width: "140px"
  },
  {
    key: "state",
    header: "State",
    render: (row) => (
      <span className="erp-reporting-chip-stack">
        <StatusChip tone={sourceStateTone(row)}>{sourceStateLabel(row.sourceStockState)}</StatusChip>
        {row.lowStock ? <StatusChip tone="warning">Low</StatusChip> : null}
        {row.expiryWarning ? <StatusChip tone="warning">Expiry</StatusChip> : null}
        {row.expired ? <StatusChip tone="danger">Expired</StatusChip> : null}
      </span>
    ),
    width: "210px",
    sticky: true
  }
];

const totalColumns: DataTableColumn<InventorySnapshotUOMTotal>[] = [
  {
    key: "uom",
    header: "Base UOM",
    render: (row) => row.baseUomCode,
    width: "110px"
  },
  {
    key: "physical",
    header: "Physical",
    render: (row) => <QuantityDisplay value={row.physicalQty} uomCode={row.baseUomCode} />,
    align: "right"
  },
  {
    key: "reserved",
    header: "Reserved",
    render: (row) => <QuantityDisplay value={row.reservedQty} uomCode={row.baseUomCode} />,
    align: "right"
  },
  {
    key: "quarantine",
    header: "Quarantine",
    render: (row) => <QuantityDisplay value={row.quarantineQty} uomCode={row.baseUomCode} />,
    align: "right"
  },
  {
    key: "blocked",
    header: "Blocked",
    render: (row) => <QuantityDisplay value={row.blockedQty} uomCode={row.baseUomCode} />,
    align: "right"
  },
  {
    key: "available",
    header: "Available",
    render: (row) => <QuantityDisplay value={row.availableQty} uomCode={row.baseUomCode} />,
    align: "right"
  }
];

type InventorySnapshotReportPanelProps = {
  controls?: ReactNode;
};

export function InventorySnapshotReportPanel({ controls }: InventorySnapshotReportPanelProps = {}) {
  const { searchParams, replaceReportUrlParams } = useReportUrlState();
  const defaultDate = defaultBusinessDate();
  const [businessDate, setBusinessDate] = useState(() => urlDateParam(searchParams, "business_date", defaultDate));
  const [warehouseId, setWarehouseId] = useState(() => urlParam(searchParams, "warehouse_id"));
  const [status, setStatus] = useState<InventorySnapshotQuery["status"]>(() =>
    urlOptionParam(
      searchParams,
      "status",
      inventorySnapshotStatusOptions.map((option) => option.value),
      ""
    )
  );
  const [itemId, setItemId] = useState(() => urlParam(searchParams, "item_id"));
  const [sku, setSKU] = useState(() => urlParam(searchParams, "sku").toUpperCase());
  const [lowStockThreshold, setLowStockThreshold] = useState(() => urlParam(searchParams, "low_stock_threshold") || "10");
  const [expiryWarningDays, setExpiryWarningDays] = useState(() => urlParam(searchParams, "expiry_warning_days") || "30");
  const [exporting, setExporting] = useState(false);
  const [exportError, setExportError] = useState<Error | null>(null);
  const [exportedFilename, setExportedFilename] = useState("");

  const query = useMemo<InventorySnapshotQuery>(
    () => ({
      businessDate,
      warehouseId: warehouseId || undefined,
      status: status || undefined,
      itemId: itemId || undefined,
      sku: sku || undefined,
      lowStockThreshold: lowStockThreshold || undefined,
      expiryWarningDays: expiryWarningDays || undefined
    }),
    [businessDate, expiryWarningDays, itemId, lowStockThreshold, sku, status, warehouseId]
  );
  const { report, loading, error } = useInventorySnapshotReport(query);
  const data = report ?? emptyInventorySnapshotReport(businessDate);
  const exportFilename = inventorySnapshotCSVFilename(query);
  const exportQueryKey = inventorySnapshotQueryString(query);

  useEffect(() => {
    replaceReportUrlParams("inventory", {
      business_date: businessDate,
      warehouse_id: warehouseId,
      status,
      item_id: itemId,
      sku,
      low_stock_threshold: lowStockThreshold,
      expiry_warning_days: expiryWarningDays
    });
  }, [businessDate, expiryWarningDays, itemId, lowStockThreshold, replaceReportUrlParams, sku, status, warehouseId]);

  useEffect(() => {
    setExportedFilename("");
    setExportError(null);
  }, [exportQueryKey]);

  async function handleExportCSV() {
    setExporting(true);
    setExportError(null);
    try {
      const download = await downloadInventorySnapshotCSV(query);
      saveBlob(download.blob, download.filename);
      setExportedFilename(download.filename);
    } catch (reason) {
      setExportError(reason instanceof Error ? reason : new Error("Inventory CSV could not be exported"));
      setExportedFilename("");
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
          <p className="erp-page-description">Inventory snapshot by warehouse, item, batch, and stock state</p>
        </div>
        <div className="erp-page-actions">
          {controls}
          <ReportExportAction
            disabled={loading}
            exporting={exporting}
            error={exportError}
            filename={exportFilename}
            exportedFilename={exportedFilename}
            reportLabel="inventory snapshot"
            onExport={handleExportCSV}
          />
        </div>
      </header>

      <section className="erp-reporting-toolbar" aria-label="Inventory snapshot filters">
        <label className="erp-field">
          <span>Business date</span>
          <input className="erp-input" type="date" value={businessDate} onChange={(event) => setBusinessDate(event.target.value)} />
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
          <span>State</span>
          <select
            className="erp-input"
            value={status}
            onChange={(event) => setStatus(event.target.value as InventorySnapshotQuery["status"])}
          >
            {inventorySnapshotStatusOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Item ID</span>
          <input
            className="erp-input"
            type="search"
            value={itemId}
            placeholder="item-serum-30ml"
            onChange={(event) => setItemId(event.target.value)}
          />
        </label>
        <label className="erp-field">
          <span>SKU</span>
          <input
            className="erp-input"
            type="search"
            value={sku}
            placeholder="SERUM-30ML"
            onChange={(event) => setSKU(event.target.value.toUpperCase())}
          />
        </label>
        <label className="erp-field">
          <span>Low stock</span>
          <input
            className="erp-input"
            inputMode="decimal"
            type="text"
            value={lowStockThreshold}
            onChange={(event) => setLowStockThreshold(event.target.value)}
          />
        </label>
        <label className="erp-field">
          <span>Expiry days</span>
          <input
            className="erp-input"
            inputMode="numeric"
            type="text"
            value={expiryWarningDays}
            onChange={(event) => setExpiryWarningDays(event.target.value)}
          />
        </label>
      </section>

      <ReportStateBanner
        loading={loading}
        error={error}
        empty={data.summary.rowCount === 0}
        liveLabel={`${data.summary.rowCount} inventory rows`}
        emptyLabel="No inventory rows match current filters"
      />

      <section className="erp-kpi-grid erp-reporting-kpis">
        <InventorySnapshotKPI label="Rows" value={String(data.summary.rowCount)} tone="info" />
        <InventorySnapshotKPI label="Available" value={quantityTotal(data, "availableQty")} tone="success" />
        <InventorySnapshotKPI label="Quarantine" value={quantityTotal(data, "quarantineQty")} tone="warning" />
        <InventorySnapshotKPI label="Blocked" value={quantityTotal(data, "blockedQty")} tone="danger" />
        <InventorySnapshotKPI label="Low stock" value={String(data.summary.lowStockRowCount)} tone="warning" />
        <InventorySnapshotKPI label="Expiry" value={String(data.summary.expiryWarningRows)} tone="warning" />
      </section>

      <section className="erp-card erp-card--padded erp-module-table-card">
        <div className="erp-section-header">
          <div>
            <h2 className="erp-section-title">Totals by base UOM</h2>
            <p className="erp-section-description">{metadataLabel(data)}</p>
          </div>
          <StatusChip tone={error ? "danger" : loading ? "warning" : "info"}>{reportStatusLabel({ error, loading })}</StatusChip>
        </div>
        <DataTable
          columns={totalColumns}
          rows={data.summary.totalsByUom}
          getRowKey={(row) => row.baseUomCode}
          loading={loading}
          error={error?.message}
          emptyState={<EmptyState title="No totals" description="No stock rows match the selected filters." />}
        />
      </section>

      <section className="erp-card erp-card--padded erp-module-table-card">
        <div className="erp-section-header">
          <h2 className="erp-section-title">Inventory rows</h2>
          <StatusChip tone={data.rows.length === 0 ? "warning" : "info"}>{data.rows.length} rows</StatusChip>
        </div>
        <DataTable
          columns={rowColumns}
          rows={data.rows}
          getRowKey={(row) => `${row.warehouseId}:${row.locationId ?? "-"}:${row.sku}:${row.batchId ?? "-"}:${row.baseUomCode}`}
          loading={loading}
          error={error?.message}
          emptyState={<EmptyState title="No inventory rows" description="No stock rows match the selected filters." />}
        />
      </section>
    </section>
  );
}

function InventorySnapshotKPI({ label, value, tone }: { label: string; value: string; tone: StatusTone }) {
  return (
    <article className="erp-card erp-card--padded erp-kpi-card">
      <span className="erp-kpi-label">{label}</span>
      <strong className="erp-kpi-value">{value}</strong>
      <StatusChip tone={tone}>{label}</StatusChip>
    </article>
  );
}

function quantityTotal(report: InventorySnapshotReport, field: keyof InventorySnapshotUOMTotal) {
  if (report.summary.totalsByUom.length !== 1) {
    return `${report.summary.totalsByUom.length} UOM`;
  }

  const total = report.summary.totalsByUom[0];
  const value = total[field];
  if (typeof value !== "string" || field === "baseUomCode") {
    return total.baseUomCode;
  }

  return formatQuantityLabel(value, total.baseUomCode);
}

function formatQuantityLabel(value: string, uomCode: string) {
  return formatQuantity(value, uomCode);
}

function metadataLabel(report: InventorySnapshotReport) {
  const generatedAt = formatDateTimeVI(report.metadata.generatedAt);
  return `${report.metadata.filters.businessDate} / ${report.metadata.timezone} / ${generatedAt}`;
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

function sourceStateLabel(value: string) {
  switch (value) {
    case "available":
      return "Available";
    case "reserved":
      return "Reserved";
    case "quarantine":
      return "Quarantine";
    case "blocked":
      return "Blocked";
    default:
      return value || "Unknown";
  }
}

function sourceStateTone(row: InventorySnapshotRow): StatusTone {
  if (row.expired || row.sourceStockState === "blocked") {
    return "danger";
  }
  if (row.lowStock || row.expiryWarning || row.sourceStockState === "quarantine") {
    return "warning";
  }
  if (row.sourceStockState === "reserved") {
    return "info";
  }

  return "success";
}

function emptyInventorySnapshotReport(businessDate: string): InventorySnapshotReport {
  return {
    metadata: {
      generatedAt: new Date().toISOString(),
      timezone: "Asia/Ho_Chi_Minh",
      sourceVersion: "reporting-v1",
      filters: {
        fromDate: businessDate,
        toDate: businessDate,
        businessDate
      }
    },
    summary: {
      rowCount: 0,
      lowStockRowCount: 0,
      expiryWarningRows: 0,
      expiredRows: 0,
      totalsByUom: []
    },
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
