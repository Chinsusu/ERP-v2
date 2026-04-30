"use client";

import { useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";
import {
  DataTable,
  EmptyState,
  StatusChip,
  type DataTableColumn,
  type StatusTone
} from "@/shared/design-system/components";
import { formatDateTimeVI, formatMoney } from "@/shared/format/numberFormat";
import { useFinanceSummaryReport } from "../hooks/useFinanceSummaryReport";
import { urlDateParam, useReportUrlState } from "../hooks/useReportUrlState";
import { downloadFinanceSummaryCSV } from "../services/financeSummaryReportService";
import { ReportSourceReferenceLink, ReportStateBanner } from "./ReportSharedStates";
import type {
  FinanceSummaryAgingBucket,
  FinanceSummaryDiscrepancyBucket,
  FinanceSummaryQuery,
  FinanceSummaryReport
} from "../types";

const agingColumns: DataTableColumn<FinanceSummaryAgingBucket>[] = [
  {
    key: "bucket",
    header: "Bucket",
    render: (row) => (
      <ReportSourceReferenceLink reference={row.sourceReference}>
        <StatusChip tone={agingTone(row.bucket)}>{agingLabel(row.bucket)}</StatusChip>
      </ReportSourceReferenceLink>
    ),
    width: "150px"
  },
  {
    key: "count",
    header: "Count",
    render: (row) => row.count,
    align: "right"
  },
  {
    key: "amount",
    header: "Amount",
    render: (row) => formatMoney(row.amount),
    align: "right",
    width: "170px"
  }
];

const discrepancyColumns: DataTableColumn<FinanceSummaryDiscrepancyBucket>[] = [
  {
    key: "type",
    header: "Type",
    render: (row) => (
      <ReportSourceReferenceLink reference={row.sourceReference}>
        {discrepancyTypeLabel(row.type)}
      </ReportSourceReferenceLink>
    ),
    width: "180px"
  },
  {
    key: "status",
    header: "Status",
    render: (row) => <StatusChip tone={row.status === "resolved" ? "success" : "warning"}>{statusLabel(row.status)}</StatusChip>,
    width: "140px"
  },
  {
    key: "count",
    header: "Count",
    render: (row) => row.count,
    align: "right"
  },
  {
    key: "amount",
    header: "Amount",
    render: (row) => formatMoney(row.amount),
    align: "right",
    width: "170px"
  }
];

type FinanceSummaryReportPanelProps = {
  controls?: ReactNode;
};

export function FinanceSummaryReportPanel({ controls }: FinanceSummaryReportPanelProps = {}) {
  const { searchParams, replaceReportUrlParams } = useReportUrlState();
  const defaultDate = defaultBusinessDate();
  const initialToDate = urlDateParam(searchParams, "to_date", defaultDate);
  const [fromDate, setFromDate] = useState(() => urlDateParam(searchParams, "from_date", initialToDate));
  const [toDate, setToDate] = useState(() => initialToDate);
  const [businessDate, setBusinessDate] = useState(() => urlDateParam(searchParams, "business_date", initialToDate));
  const [exporting, setExporting] = useState(false);
  const [exportError, setExportError] = useState<Error | null>(null);

  const query = useMemo<FinanceSummaryQuery>(
    () => ({
      fromDate,
      toDate,
      businessDate
    }),
    [businessDate, fromDate, toDate]
  );
  const { report, loading, error } = useFinanceSummaryReport(query);
  const data = report ?? emptyFinanceSummaryReport(fromDate, toDate, businessDate);

  useEffect(() => {
    replaceReportUrlParams("finance", {
      from_date: fromDate,
      to_date: toDate,
      business_date: businessDate
    });
  }, [businessDate, fromDate, replaceReportUrlParams, toDate]);

  async function handleExportCSV() {
    setExporting(true);
    setExportError(null);
    try {
      const download = await downloadFinanceSummaryCSV(query);
      saveBlob(download.blob, download.filename);
    } catch (reason) {
      setExportError(reason instanceof Error ? reason : new Error("Finance CSV could not be exported"));
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
          <p className="erp-page-description">Finance summary by AR/AP aging, COD discrepancy, and cash movement</p>
        </div>
        <div className="erp-page-actions">
          {controls}
          {exportError ? <StatusChip tone="danger">Export failed</StatusChip> : null}
          <button
            className="erp-button erp-button--secondary"
            type="button"
            disabled={loading || exporting}
            title={exportError?.message ?? "Export finance summary CSV"}
            onClick={handleExportCSV}
          >
            {exporting ? "Exporting" : "Export CSV"}
          </button>
        </div>
      </header>

      <section className="erp-reporting-toolbar erp-reporting-toolbar--finance" aria-label="Finance summary filters">
        <label className="erp-field">
          <span>From date</span>
          <input className="erp-input" type="date" value={fromDate} onChange={(event) => setFromDate(event.target.value)} />
        </label>
        <label className="erp-field">
          <span>To date</span>
          <input
            className="erp-input"
            type="date"
            value={toDate}
            onChange={(event) => {
              setToDate(event.target.value);
              setBusinessDate(event.target.value);
            }}
          />
        </label>
        <label className="erp-field">
          <span>As of</span>
          <input
            className="erp-input"
            type="date"
            value={businessDate}
            onChange={(event) => setBusinessDate(event.target.value)}
          />
        </label>
      </section>

      <ReportStateBanner
        loading={loading}
        error={error}
        empty={isFinanceSummaryEmpty(data)}
        liveLabel="Finance summary loaded"
        emptyLabel="No finance records match current filters"
      />

      <section className="erp-kpi-grid erp-reporting-kpis">
        <FinanceSummaryKPI label="AR open" value={formatMoney(data.ar.openAmount, data.currencyCode)} tone="info" />
        <FinanceSummaryKPI label="AR overdue" value={formatMoney(data.ar.overdueAmount, data.currencyCode)} tone="danger" />
        <FinanceSummaryKPI label="AP due" value={formatMoney(data.ap.dueAmount, data.currencyCode)} tone="warning" />
        <FinanceSummaryKPI label="COD pending" value={formatMoney(data.cod.pendingAmount, data.currencyCode)} tone="info" />
        <FinanceSummaryKPI
          label="COD discrepancy"
          value={formatMoney(data.cod.discrepancyAmount, data.currencyCode)}
          tone={data.cod.discrepancyCount > 0 ? "danger" : "success"}
        />
        <FinanceSummaryKPI
          label="Net cash"
          value={formatMoney(data.cash.netCashAmount, data.currencyCode)}
          tone={data.cash.netCashAmount.startsWith("-") ? "danger" : "success"}
        />
      </section>

      <section className="erp-reporting-bucket-grid">
        <section className="erp-card erp-card--padded erp-module-table-card">
          <div className="erp-section-header">
            <div>
              <h2 className="erp-section-title">AR aging</h2>
              <p className="erp-section-description">{receivableLabel(data)}</p>
            </div>
            <StatusChip tone={error ? "danger" : loading ? "warning" : "info"}>{reportStatusLabel({ error, loading })}</StatusChip>
          </div>
          <DataTable
            columns={agingColumns}
            rows={data.ar.agingBuckets}
            getRowKey={(row) => row.bucket}
            loading={loading}
            error={error?.message}
            emptyState={<EmptyState title="No AR aging" description="No receivable aging buckets are available." />}
          />
        </section>

        <section className="erp-card erp-card--padded erp-module-table-card">
          <div className="erp-section-header">
            <div>
              <h2 className="erp-section-title">AP aging</h2>
              <p className="erp-section-description">{payableLabel(data)}</p>
            </div>
            <StatusChip tone={data.ap.dueCount > 0 ? "warning" : "info"}>{data.ap.dueCount} due</StatusChip>
          </div>
          <DataTable
            columns={agingColumns}
            rows={data.ap.agingBuckets}
            getRowKey={(row) => row.bucket}
            loading={loading}
            error={error?.message}
            emptyState={<EmptyState title="No AP aging" description="No payable aging buckets are available." />}
          />
        </section>
      </section>

      <section className="erp-card erp-card--padded erp-module-table-card">
        <div className="erp-section-header">
          <div>
            <h2 className="erp-section-title">COD discrepancy buckets</h2>
            <p className="erp-section-description">{metadataLabel(data)}</p>
          </div>
          <StatusChip tone={data.cod.discrepancyCount > 0 ? "danger" : "success"}>{data.cod.discrepancyCount} discrepancies</StatusChip>
        </div>
        <DataTable
          columns={discrepancyColumns}
          rows={data.cod.discrepancyBuckets}
          getRowKey={(row) => `${row.type}:${row.status}`}
          loading={loading}
          error={error?.message}
          emptyState={<EmptyState title="No COD discrepancies" description="No discrepancy buckets match the selected date range." />}
        />
      </section>

      <section className="erp-card erp-card--padded">
        <div className="erp-section-header">
          <div>
            <h2 className="erp-section-title">Cash movement</h2>
            <p className="erp-section-description">{data.cash.transactionCount} posted transactions</p>
          </div>
          <StatusChip tone={data.cash.netCashAmount.startsWith("-") ? "danger" : "success"}>{formatMoney(data.cash.netCashAmount)}</StatusChip>
        </div>
        <div className="erp-reporting-cash-grid">
          <FinanceSummaryFact label="Cash in" value={formatMoney(data.cash.cashInAmount, data.currencyCode)} tone="success" />
          <FinanceSummaryFact label="Cash out" value={formatMoney(data.cash.cashOutAmount, data.currencyCode)} tone="warning" />
          <FinanceSummaryFact label="Net cash" value={formatMoney(data.cash.netCashAmount, data.currencyCode)} tone="info" />
        </div>
      </section>
    </section>
  );
}

function FinanceSummaryKPI({ label, value, tone }: { label: string; value: string; tone: StatusTone }) {
  return (
    <article className="erp-card erp-card--padded erp-kpi-card">
      <span className="erp-kpi-label">{label}</span>
      <strong className="erp-kpi-value erp-kpi-value--small">{value}</strong>
      <StatusChip tone={tone}>{label}</StatusChip>
    </article>
  );
}

function FinanceSummaryFact({ label, value, tone }: { label: string; value: string; tone: StatusTone }) {
  return (
    <div className="erp-reporting-fact">
      <span>{label}</span>
      <strong>{value}</strong>
      <StatusChip tone={tone}>{label}</StatusChip>
    </div>
  );
}

function metadataLabel(report: FinanceSummaryReport) {
  const generatedAt = formatDateTimeVI(report.metadata.generatedAt);
  return `${report.metadata.filters.fromDate} -> ${report.metadata.filters.toDate} / ${report.metadata.timezone} / ${generatedAt}`;
}

function receivableLabel(report: FinanceSummaryReport) {
  return `${report.ar.openCount} open / ${report.ar.overdueCount} overdue / ${formatMoney(report.ar.outstandingAmount, report.currencyCode)}`;
}

function payableLabel(report: FinanceSummaryReport) {
  return `${report.ap.openCount} open / ${report.ap.dueCount} due / ${formatMoney(report.ap.outstandingAmount, report.currencyCode)}`;
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

function agingLabel(value: string) {
  switch (value) {
    case "current":
      return "Current";
    case "1_7":
      return "1-7 days";
    case "8_30":
      return "8-30 days";
    case "31_plus":
      return "31+ days";
    default:
      return value || "Unknown";
  }
}

function agingTone(value: string): StatusTone {
  switch (value) {
    case "current":
      return "success";
    case "1_7":
      return "warning";
    case "8_30":
    case "31_plus":
      return "danger";
    default:
      return "normal";
  }
}

function discrepancyTypeLabel(value: string) {
  return value.replaceAll("_", " ");
}

function statusLabel(value: string) {
  return value.replaceAll("_", " ");
}

function isFinanceSummaryEmpty(report: FinanceSummaryReport) {
  return (
    report.ar.openCount === 0 &&
    report.ap.openCount === 0 &&
    report.cod.pendingCount === 0 &&
    report.cash.transactionCount === 0
  );
}

function emptyFinanceSummaryReport(fromDate: string, toDate: string, businessDate: string): FinanceSummaryReport {
  return {
    metadata: {
      generatedAt: new Date().toISOString(),
      timezone: "Asia/Ho_Chi_Minh",
      sourceVersion: "reporting-v1",
      filters: {
        fromDate,
        toDate,
        businessDate
      }
    },
    currencyCode: "VND",
    ar: {
      openCount: 0,
      overdueCount: 0,
      disputedCount: 0,
      openAmount: "0.00",
      overdueAmount: "0.00",
      outstandingAmount: "0.00",
      agingBuckets: [],
      sourceReferences: []
    },
    ap: {
      openCount: 0,
      dueCount: 0,
      paymentRequestedCount: 0,
      paymentApprovedCount: 0,
      openAmount: "0.00",
      dueAmount: "0.00",
      outstandingAmount: "0.00",
      agingBuckets: [],
      sourceReferences: []
    },
    cod: {
      pendingCount: 0,
      discrepancyCount: 0,
      pendingAmount: "0.00",
      discrepancyAmount: "0.00",
      discrepancyBuckets: [],
      sourceReferences: []
    },
    cash: {
      transactionCount: 0,
      cashInAmount: "0.00",
      cashOutAmount: "0.00",
      netCashAmount: "0.00",
      sourceReferences: []
    }
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
