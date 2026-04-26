"use client";

import type { MockUser } from "@/shared/auth/mockSession";
import { DataTable, StatusChip, type DataTableColumn } from "@/shared/design-system/components";
import { getVisibleActions, moduleActions, type AppMenuItem } from "@/shared/permissions/menu";

type ModulePlaceholderProps = {
  item: AppMenuItem;
  user: MockUser;
};

const kpis = [
  { label: "Open queue", value: "0", status: "normal", chip: "Ready" },
  { label: "Needs review", value: "0", status: "normal", chip: "Clear" },
  { label: "Blocked", value: "0", status: "normal", chip: "Clear" }
] as const;

type WorkQueueRow = {
  queue: string;
  owner: string;
  status: "Ready" | "Review" | "Blocked";
};

const rows: WorkQueueRow[] = [
  { queue: "New receipts", owner: "Ops", status: "Ready" },
  { queue: "Exception review", owner: "Lead", status: "Review" },
  { queue: "Daily checkpoint", owner: "Ops", status: "Ready" }
];

const columns: DataTableColumn<WorkQueueRow>[] = [
  { key: "queue", header: "Queue", render: (row) => row.queue },
  { key: "owner", header: "Owner", render: (row) => row.owner, width: "140px" },
  {
    key: "status",
    header: "Status",
    render: (row) => (
      <StatusChip tone={row.status === "Blocked" ? "danger" : row.status === "Review" ? "warning" : "success"}>
        {row.status}
      </StatusChip>
    ),
    width: "140px"
  }
];

export function ModulePlaceholder({ item, user }: ModulePlaceholderProps) {
  const actions = getVisibleActions(user, moduleActions);

  return (
    <section className="erp-module-page">
      <header className="erp-page-header">
        <div>
          <p className="erp-module-eyebrow">{item.code}</p>
          <h1 className="erp-page-title">{item.label}</h1>
        </div>
        {actions.length > 0 ? (
          <div className="erp-page-actions">
            {actions.map((action) => (
              <button
                className={`erp-button erp-button--${action.variant}`}
                key={action.label}
                type="button"
              >
                {action.label}
              </button>
            ))}
          </div>
        ) : null}
      </header>

      <section className="erp-kpi-grid">
        {kpis.map((kpi) => (
          <article className="erp-card erp-card--padded erp-kpi-card" key={kpi.label}>
            <div className="erp-kpi-label">{kpi.label}</div>
            <strong className="erp-kpi-value">{kpi.value}</strong>
            <span className={`erp-status-chip erp-status-chip--${kpi.status}`}>{kpi.chip}</span>
          </article>
        ))}
      </section>

      <section className="erp-card erp-card--padded erp-module-table-card">
        <div className="erp-section-header">
          <h2 className="erp-section-title">Work queue</h2>
          <StatusChip>Mock</StatusChip>
        </div>
        <DataTable columns={columns} rows={rows} getRowKey={(row) => row.queue} />
      </section>
    </section>
  );
}
