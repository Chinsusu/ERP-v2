"use client";

import { useMemo, useState } from "react";
import { DataTable, EmptyState, StatusChip, type DataTableColumn, type StatusTone } from "@/shared/design-system/components";
import { useFinanceDashboard } from "../hooks/useFinanceDashboard";
import { formatFinanceDate, formatFinanceMoney } from "../services/customerReceivableService";
import { buildFinanceSummaryReportHref } from "../services/financeDashboardService";
import type { FinanceDashboard } from "../types";

type DashboardActionRow = {
  id: string;
  area: string;
  metric: string;
  amount: string;
  tone: StatusTone;
  sectionHref: string;
  reportHref: string;
};

const actionColumns: DataTableColumn<DashboardActionRow>[] = [
  {
    key: "area",
    header: "Area",
    render: (row) => (
      <span className="erp-finance-receivable-cell">
        <strong>{row.area}</strong>
        <small>{row.metric}</small>
      </span>
    )
  },
  {
    key: "amount",
    header: "Amount",
    render: (row) => formatFinanceMoney(row.amount),
    align: "right",
    width: "160px"
  },
  {
    key: "status",
    header: "Status",
    render: (row) => <StatusChip tone={row.tone}>{row.metric}</StatusChip>,
    width: "180px"
  },
  {
    key: "action",
    header: "Action",
    render: (row) => (
      <span className="erp-finance-action-links">
        <a className="erp-button erp-button--secondary" href={row.sectionHref}>
          Open
        </a>
        <a className="erp-button erp-button--secondary" href={row.reportHref}>
          Report
        </a>
      </span>
    ),
    width: "180px",
    sticky: true
  }
];

export function FinanceDashboardPanel() {
  const [businessDate, setBusinessDate] = useState(defaultBusinessDate());
  const query = useMemo(() => ({ businessDate }), [businessDate]);
  const { dashboard, loading, error } = useFinanceDashboard(query);
  const metrics = dashboard ?? emptyFinanceDashboard(businessDate);
  const financeReportHref = buildFinanceSummaryReportHref(query);
  const actionRows = useMemo(() => createActionRows(metrics, financeReportHref), [financeReportHref, metrics]);

  return (
    <section className="erp-finance-section" id="finance-dashboard">
      <div className="erp-section-header">
        <div>
          <h2 className="erp-section-title">Finance dashboard</h2>
          <p className="erp-section-description">AR, AP, COD, and cash position for the selected business date</p>
        </div>
        <div className="erp-finance-action-links">
          <StatusChip tone={loading ? "warning" : "info"}>{formatFinanceDate(metrics.generatedAt)}</StatusChip>
          <a className="erp-button erp-button--secondary" href={financeReportHref}>
            Finance report
          </a>
        </div>
      </div>

      <section className="erp-finance-toolbar" aria-label="Finance dashboard filters">
        <label className="erp-field">
          <span>Business date</span>
          <input className="erp-input" type="date" value={businessDate} onChange={(event) => setBusinessDate(event.target.value)} />
        </label>
        <StatusChip tone={error ? "danger" : "success"}>{error ? "API fallback" : metrics.currencyCode}</StatusChip>
      </section>

      <section className="erp-kpi-grid erp-finance-kpis">
        <FinanceDashboardKPI
          label="AR open"
          value={metrics.ar.openCount}
          amount={metrics.ar.openAmount}
          tone={metrics.ar.overdueCount > 0 ? "warning" : "success"}
        />
        <FinanceDashboardKPI
          label="AP due"
          value={metrics.ap.dueCount}
          amount={metrics.ap.dueAmount}
          tone={metrics.ap.dueCount > 0 ? "warning" : "success"}
        />
        <FinanceDashboardKPI
          label="COD pending"
          value={metrics.cod.pendingCount}
          amount={metrics.cod.pendingAmount}
          tone={metrics.cod.discrepancyCount > 0 ? "danger" : "info"}
        />
        <FinanceDashboardKPI
          label="Cash net"
          value={metrics.cash.transactionCount}
          amount={metrics.cash.netCashToday}
          tone={Number(metrics.cash.netCashToday) >= 0 ? "success" : "danger"}
        />
      </section>

      <section className="erp-card erp-card--padded">
        <div className="erp-section-header">
          <h2 className="erp-section-title">Action queue</h2>
          <StatusChip tone="info">{actionRows.length} queues</StatusChip>
        </div>
        <DataTable
          columns={actionColumns}
          rows={actionRows}
          getRowKey={(row) => row.id}
          loading={loading}
          error={error?.message}
          emptyState={<EmptyState title="No finance queue" description="Dashboard metrics did not return actionable rows." />}
        />
      </section>
    </section>
  );
}

function FinanceDashboardKPI({
  label,
  value,
  amount,
  tone
}: {
  label: string;
  value: number;
  amount: string;
  tone: StatusTone;
}) {
  return (
    <div className="erp-card erp-card--padded erp-kpi-card">
      <span className="erp-kpi-label">{label}</span>
      <strong className="erp-kpi-value">{formatFinanceMoney(amount)}</strong>
      <StatusChip tone={tone}>{value}</StatusChip>
    </div>
  );
}

function createActionRows(metrics: FinanceDashboard, reportHref: string): DashboardActionRow[] {
  return [
    {
      id: "ar-overdue",
      area: "Customer receivables",
      metric: `${metrics.ar.overdueCount} overdue`,
      amount: metrics.ar.overdueAmount,
      tone: metrics.ar.overdueCount > 0 ? "warning" : "success",
      sectionHref: "#customer-receivables",
      reportHref
    },
    {
      id: "ap-due",
      area: "Supplier payables",
      metric: `${metrics.ap.dueCount} due`,
      amount: metrics.ap.dueAmount,
      tone: metrics.ap.dueCount > 0 ? "warning" : "success",
      sectionHref: "#supplier-payables",
      reportHref
    },
    {
      id: "cod-discrepancy",
      area: "COD reconciliation",
      metric: `${metrics.cod.discrepancyCount} discrepancy`,
      amount: metrics.cod.discrepancyAmount,
      tone: metrics.cod.discrepancyCount > 0 ? "danger" : "success",
      sectionHref: "#cod-reconciliation",
      reportHref
    },
    {
      id: "cash-today",
      area: "Cash transactions",
      metric: `${metrics.cash.transactionCount} posted`,
      amount: metrics.cash.netCashToday,
      tone: Number(metrics.cash.netCashToday) >= 0 ? "success" : "danger",
      sectionHref: "#cash-transactions",
      reportHref
    }
  ];
}

function emptyFinanceDashboard(businessDate: string): FinanceDashboard {
  return {
    businessDate,
    generatedAt: new Date().toISOString(),
    currencyCode: "VND",
    ar: {
      openCount: 0,
      overdueCount: 0,
      disputedCount: 0,
      openAmount: "0.00",
      overdueAmount: "0.00",
      outstandingAmount: "0.00"
    },
    ap: {
      openCount: 0,
      dueCount: 0,
      paymentRequestedCount: 0,
      paymentApprovedCount: 0,
      openAmount: "0.00",
      dueAmount: "0.00",
      outstandingAmount: "0.00"
    },
    cod: {
      pendingCount: 0,
      discrepancyCount: 0,
      pendingAmount: "0.00",
      discrepancyAmount: "0.00"
    },
    cash: {
      transactionCount: 0,
      cashInToday: "0.00",
      cashOutToday: "0.00",
      netCashToday: "0.00"
    }
  };
}

function defaultBusinessDate() {
  return new Date().toISOString().slice(0, 10);
}
