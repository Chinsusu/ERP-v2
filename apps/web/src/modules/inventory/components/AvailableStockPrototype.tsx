"use client";

import { useMemo, useState } from "react";
import { DataTable, StatusChip, type DataTableColumn } from "@/shared/design-system/components";
import { useAvailableStock } from "../hooks/useAvailableStock";
import { availabilityTone } from "../services/stockAvailabilityService";
import type { AvailableStockItem, AvailableStockQuery } from "../types";

const warehouseOptions = [
  { label: "All warehouses", value: "" },
  { label: "HCM", value: "wh-hcm" },
  { label: "HN", value: "wh-hn" }
];

const columns: DataTableColumn<AvailableStockItem>[] = [
  {
    key: "warehouse",
    header: "Warehouse",
    render: (row) => row.warehouseCode,
    width: "120px"
  },
  {
    key: "sku",
    header: "SKU",
    render: (row) => row.sku,
    width: "180px"
  },
  {
    key: "batch",
    header: "Batch",
    render: (row) => row.batchNo ?? "-",
    width: "150px"
  },
  {
    key: "physical",
    header: "Physical",
    render: (row) => formatQuantity(row.physicalStock),
    align: "right",
    width: "120px"
  },
  {
    key: "reserved",
    header: "Reserved",
    render: (row) => formatQuantity(row.reservedStock),
    align: "right",
    width: "120px"
  },
  {
    key: "hold",
    header: "Hold",
    render: (row) => formatQuantity(row.holdStock),
    align: "right",
    width: "120px"
  },
  {
    key: "available",
    header: "Available",
    render: (row) => formatQuantity(row.availableStock),
    align: "right",
    width: "120px"
  },
  {
    key: "state",
    header: "State",
    render: (row) => <StatusChip tone={availabilityTone(row)}>{statusLabel(row)}</StatusChip>,
    width: "130px"
  }
];

export function AvailableStockPrototype() {
  const [warehouseId, setWarehouseId] = useState("");
  const [sku, setSKU] = useState("");
  const query = useMemo<AvailableStockQuery>(
    () => ({
      warehouseId: warehouseId || undefined,
      sku: sku || undefined
    }),
    [warehouseId, sku]
  );
  const { items, loading, summary } = useAvailableStock(query);

  return (
    <section className="erp-module-page erp-stock-page">
      <header className="erp-page-header">
        <div>
          <p className="erp-module-eyebrow">IV</p>
          <h1 className="erp-page-title">Inventory</h1>
          <p className="erp-page-description">Available stock by warehouse, SKU, and batch</p>
        </div>
      </header>

      <section className="erp-stock-toolbar" aria-label="Inventory filters">
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
          <span>SKU</span>
          <input
            className="erp-input"
            type="search"
            value={sku}
            placeholder="SERUM-30ML"
            onChange={(event) => setSKU(event.target.value.toUpperCase())}
          />
        </label>
      </section>

      <section className="erp-kpi-grid erp-stock-kpis">
        <StockKPI label="Physical" value={summary.physicalStock} tone="normal" />
        <StockKPI label="Reserved" value={summary.reservedStock} tone="warning" />
        <StockKPI label="Hold" value={summary.holdStock} tone="danger" />
        <StockKPI label="Available" value={summary.availableStock} tone="success" />
      </section>

      <section className="erp-card erp-card--padded erp-module-table-card">
        <div className="erp-section-header">
          <h2 className="erp-section-title">Available stock</h2>
          <StatusChip tone={items.length === 0 ? "warning" : "info"}>{items.length} rows</StatusChip>
        </div>
        <DataTable columns={columns} rows={items} getRowKey={(row) => `${row.warehouseId}:${row.sku}:${row.batchId}`} loading={loading} />
      </section>
    </section>
  );
}

function StockKPI({ label, value, tone }: { label: string; value: number; tone: "normal" | "success" | "warning" | "danger" }) {
  return (
    <article className="erp-card erp-card--padded erp-kpi-card">
      <div className="erp-kpi-label">{label}</div>
      <strong className="erp-kpi-value">{formatQuantity(value)}</strong>
      <StatusChip tone={tone}>{label}</StatusChip>
    </article>
  );
}

function formatQuantity(value: number) {
  return new Intl.NumberFormat("en-US", { maximumFractionDigits: 0 }).format(value);
}

function statusLabel(item: AvailableStockItem) {
  if (item.availableStock <= 0) {
    return "Blocked";
  }
  if (item.holdStock > 0) {
    return "Hold";
  }

  return "Available";
}
