"use client";

import { useWarehouseDailyBoard } from "../hooks/useWarehouseDailyBoard";

const statusClassName = {
  normal: "erp-status-chip--normal",
  warning: "erp-status-chip--warning",
  blocked: "erp-status-chip--danger"
};

export default function WarehouseDailyBoard() {
  const { items, loading } = useWarehouseDailyBoard();

  return (
    <section className="erp-module-page">
      <header className="erp-page-header">
        <div>
          <h1 className="erp-page-title">Warehouse Daily Board</h1>
          <p className="erp-page-description">Phase 1 operational skeleton</p>
        </div>
      </header>

      {loading ? (
        <p>Loading...</p>
      ) : (
        <section className="erp-kpi-grid">
          {items.map((item) => (
            <article className="erp-card erp-card--padded erp-kpi-card" key={item.id}>
              <div className="erp-kpi-label">{item.label}</div>
              <strong className="erp-kpi-value">{item.count}</strong>
              <span className={`erp-status-chip ${statusClassName[item.status]}`}>{item.status}</span>
            </article>
          ))}
        </section>
      )}
    </section>
  );
}
