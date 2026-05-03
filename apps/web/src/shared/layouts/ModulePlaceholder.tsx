"use client";

import type { AuthenticatedUser } from "../auth/session";
import { DataTable, StatusChip, type DataTableColumn } from "@/shared/design-system/components";
import { getActionLabel } from "@/shared/i18n/action-labels";
import { t } from "@/shared/i18n";
import { getNavigationItemLabel } from "@/shared/i18n/navigation-labels";
import { getStatusLabel } from "@/shared/i18n/status-labels";
import { getVisibleActions, moduleActions, type AppMenuItem } from "@/shared/permissions/menu";

type ModulePlaceholderProps = {
  item: AppMenuItem;
  user: AuthenticatedUser;
};

const kpis = [
  { labelKey: "openQueue", value: "0", status: "normal", chipKey: "ready" },
  { labelKey: "needsReview", value: "0", status: "normal", chipKey: "clear" },
  { labelKey: "blocked", value: "0", status: "normal", chipKey: "clear" }
] as const;

type WorkQueueRow = {
  queueKey: "newReceipts" | "exceptionReview" | "dailyCheckpoint";
  ownerKey: "ops" | "lead";
  status: "Ready" | "Review" | "Blocked";
};

const rows: WorkQueueRow[] = [
  { queueKey: "newReceipts", ownerKey: "ops", status: "Ready" },
  { queueKey: "exceptionReview", ownerKey: "lead", status: "Review" },
  { queueKey: "dailyCheckpoint", ownerKey: "ops", status: "Ready" }
];

const columns: DataTableColumn<WorkQueueRow>[] = [
  { key: "queue", header: t("common.queue"), render: (row) => t(`common.${row.queueKey}`) },
  { key: "owner", header: t("common.owner"), render: (row) => t(`common.${row.ownerKey}`), width: "140px" },
  {
    key: "status",
    header: t("common.status"),
    render: (row) => (
      <StatusChip tone={row.status === "Blocked" ? "danger" : row.status === "Review" ? "warning" : "success"}>
        {getStatusLabel(row.status)}
      </StatusChip>
    ),
    width: "140px"
  }
];

export function ModulePlaceholder({ item, user }: ModulePlaceholderProps) {
  const actions = getVisibleActions(user, moduleActions);
  const title = getNavigationItemLabel(item);

  return (
    <section className="erp-module-page">
      <header className="erp-page-header">
        <div>
          <p className="erp-module-eyebrow">{item.code}</p>
          <h1 className="erp-page-title">{title}</h1>
        </div>
        {actions.length > 0 ? (
          <div className="erp-page-actions">
            {actions.map((action) => (
              <button
                className={`erp-button erp-button--${action.variant}`}
                key={action.label}
                type="button"
              >
                {getActionLabel(action.label)}
              </button>
            ))}
          </div>
        ) : null}
      </header>

      <section className="erp-kpi-grid">
        {kpis.map((kpi) => (
          <article className="erp-card erp-card--padded erp-kpi-card" key={kpi.labelKey}>
            <div className="erp-kpi-label">{t(`common.${kpi.labelKey}`)}</div>
            <strong className="erp-kpi-value">{kpi.value}</strong>
            <span className={`erp-status-chip erp-status-chip--${kpi.status}`}>
              {t(`common.${kpi.chipKey}`)}
            </span>
          </article>
        ))}
      </section>

      <section className="erp-card erp-card--padded erp-module-table-card">
        <div className="erp-section-header">
          <h2 className="erp-section-title">{t("common.workQueue")}</h2>
          <StatusChip>{t("common.mock")}</StatusChip>
        </div>
        <DataTable columns={columns} rows={rows} getRowKey={(row) => row.queueKey} />
      </section>
    </section>
  );
}
