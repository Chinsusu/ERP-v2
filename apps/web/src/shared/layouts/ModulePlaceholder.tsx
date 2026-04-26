import type { MockUser } from "@/shared/auth/mockSession";
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

const rows = ["New receipts", "Exception review", "Daily checkpoint"];

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
          <span className="erp-status-chip erp-status-chip--normal">Mock</span>
        </div>
        <div className="erp-module-table" role="table" aria-label={`${item.label} work queue`}>
          <div className="erp-module-table-row erp-module-table-row--head" role="row">
            <span role="columnheader">Queue</span>
            <span role="columnheader">Owner</span>
            <span role="columnheader">Status</span>
          </div>
          {rows.map((row) => (
            <div className="erp-module-table-row" role="row" key={row}>
              <span role="cell">{row}</span>
              <span role="cell">Ops</span>
              <span role="cell">Ready</span>
            </div>
          ))}
        </div>
      </section>
    </section>
  );
}
