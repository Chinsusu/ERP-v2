"use client";

import { useWarehouseDailyBoard } from "../hooks/useWarehouseDailyBoard";

const statusColor = {
  normal: "#1d6f5f",
  warning: "#b7791f",
  blocked: "#b42318"
};

export default function WarehouseDailyBoard() {
  const { items, loading } = useWarehouseDailyBoard();

  return (
    <main style={{ padding: 24 }}>
      <header style={{ marginBottom: 20 }}>
        <h1 style={{ fontSize: 24, margin: 0 }}>Warehouse Daily Board</h1>
        <p style={{ color: "#607089", marginTop: 8 }}>Phase 1 operational skeleton</p>
      </header>

      {loading ? (
        <p>Loading...</p>
      ) : (
        <section
          style={{
            display: "grid",
            gap: 12,
            gridTemplateColumns: "repeat(auto-fit, minmax(180px, 1fr))"
          }}
        >
          {items.map((item) => (
            <article
              key={item.id}
              style={{
                background: "#fff",
                border: "1px solid #d7dee8",
                borderRadius: 8,
                padding: 16
              }}
            >
              <div style={{ color: "#607089", fontSize: 13 }}>{item.label}</div>
              <strong style={{ display: "block", fontSize: 28, marginTop: 8 }}>{item.count}</strong>
              <span style={{ color: statusColor[item.status], fontSize: 13 }}>{item.status}</span>
            </article>
          ))}
        </section>
      )}
    </main>
  );
}
