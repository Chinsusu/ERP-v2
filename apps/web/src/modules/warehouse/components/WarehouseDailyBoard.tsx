"use client";

import { useMemo, useState } from "react";
import { DataTable, ScanInput, StatusChip, type DataTableColumn, type ToastMessage } from "@/shared/design-system/components";
import { useWarehouseDailyBoard } from "../hooks/useWarehouseDailyBoard";
import { ShiftClosingPanel } from "./ShiftClosingPanel";
import {
  defaultWarehouseDailyBoardDate,
  statusOptions,
  warehouseOptions,
  warehouseTaskTone
} from "../services/warehouseDailyBoardService";
import type { WarehouseDailyBoardQuery, WarehouseDailyTask, WarehouseDailyTaskStatus } from "../types";

const columns: DataTableColumn<WarehouseDailyTask>[] = [
  {
    key: "reference",
    header: "Reference",
    render: (row) => (
      <a className="erp-warehouse-task-link" href={row.href}>
        {row.reference}
      </a>
    ),
    width: "170px"
  },
  {
    key: "task",
    header: "Task",
    render: (row) => row.title,
    width: "260px"
  },
  {
    key: "warehouse",
    header: "Warehouse",
    render: (row) => row.warehouseCode,
    width: "110px"
  },
  {
    key: "owner",
    header: "Owner",
    render: (row) => row.owner,
    width: "120px"
  },
  {
    key: "due",
    header: "Due",
    render: (row) => formatDueTime(row.dueAt),
    width: "110px"
  },
  {
    key: "priority",
    header: "Priority",
    render: (row) => <StatusChip tone={priorityTone(row.priority)}>{row.priority}</StatusChip>,
    width: "100px"
  },
  {
    key: "status",
    header: "Status",
    render: (row) => <StatusChip tone={warehouseTaskTone(row.status)}>{statusLabel(row.status)}</StatusChip>,
    width: "130px"
  }
];

export default function WarehouseDailyBoard() {
  const [warehouseId, setWarehouseId] = useState("");
  const [date, setDate] = useState(defaultWarehouseDailyBoardDate);
  const [status, setStatus] = useState<"" | WarehouseDailyTaskStatus>("");
  const [scanFeedback, setScanFeedback] = useState<ToastMessage | undefined>();
  const query = useMemo<WarehouseDailyBoardQuery>(
    () => ({
      warehouseId: warehouseId || undefined,
      date,
      status: status || undefined
    }),
    [date, status, warehouseId]
  );
  const { board, loading } = useWarehouseDailyBoard(query);
  const exceptions = board?.tasks.filter((task) => task.priority === "P0" || task.status === "mismatch") ?? [];

  return (
    <section className="erp-module-page erp-warehouse-page">
      <header className="erp-page-header">
        <div>
          <p className="erp-module-eyebrow">WH</p>
          <h1 className="erp-page-title">Warehouse Daily Board</h1>
          <p className="erp-page-description">Daily warehouse work queue, handover, returns, and variance control</p>
        </div>
        <div className="erp-page-actions">
          <button className="erp-button erp-button--secondary" type="button">
            Export
          </button>
          <button className="erp-button erp-button--primary" type="button">
            Start closing
          </button>
        </div>
      </header>

      <section className="erp-warehouse-toolbar" aria-label="Warehouse daily board filters">
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
          <span>Date</span>
          <input className="erp-input" type="date" value={date} onChange={(event) => setDate(event.target.value)} />
        </label>
        <label className="erp-field">
          <span>Status</span>
          <select
            className="erp-input"
            value={status}
            onChange={(event) => setStatus(event.target.value as "" | WarehouseDailyTaskStatus)}
          >
            {statusOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
      </section>

      <section className="erp-kpi-grid erp-warehouse-kpis">
        <WarehouseKPI label="Waiting" value={board?.summary.waiting ?? 0} tone="normal" />
        <WarehouseKPI label="Picking" value={board?.summary.picking ?? 0} tone="warning" />
        <WarehouseKPI label="Packed" value={board?.summary.packed ?? 0} tone="success" />
        <WarehouseKPI label="Handover" value={board?.summary.handover ?? 0} tone="info" />
        <WarehouseKPI label="Returns" value={board?.summary.returns ?? 0} tone="warning" />
        <WarehouseKPI label="Mismatch" value={board?.summary.reconciliationMismatch ?? 0} tone="danger" />
      </section>

      <section className="erp-warehouse-ops">
        <div className="erp-card erp-card--padded erp-warehouse-scan-card">
          <div className="erp-section-header">
            <h2 className="erp-section-title">Scan station</h2>
            <StatusChip tone={board?.shiftStatus === "open" ? "warning" : "success"}>
              {board?.shiftStatus ?? "open"}
            </StatusChip>
          </div>
          <ScanInput
            label="Warehouse scan"
            placeholder="Order, manifest, return, or variance code"
            feedback={scanFeedback}
            onScan={(value) =>
              setScanFeedback({
                id: value,
                title: `Queued ${value.toUpperCase()}`,
                tone: "info"
              })
            }
          />
        </div>

        <div className="erp-card erp-card--padded erp-warehouse-exception-card">
          <div className="erp-section-header">
            <h2 className="erp-section-title">Exceptions</h2>
            <StatusChip tone={exceptions.length > 0 ? "danger" : "success"}>{exceptions.length} open</StatusChip>
          </div>
          <div className="erp-warehouse-exception-list">
            {exceptions.length > 0 ? (
              exceptions.map((task) => (
                <a className="erp-warehouse-exception" href={task.href} key={task.id}>
                  <strong>{task.reference}</strong>
                  <span>{task.title}</span>
                </a>
              ))
            ) : (
              <span className="erp-warehouse-empty">No P0 exceptions</span>
            )}
          </div>
        </div>
      </section>

      <ShiftClosingPanel query={{ warehouseId: warehouseId || undefined, date }} />

      <section className="erp-card erp-card--padded erp-module-table-card">
        <div className="erp-section-header">
          <h2 className="erp-section-title">Task board</h2>
          <StatusChip tone={(board?.tasks.length ?? 0) === 0 ? "warning" : "info"}>{board?.tasks.length ?? 0} rows</StatusChip>
        </div>
        <DataTable columns={columns} rows={board?.tasks ?? []} getRowKey={(row) => row.id} loading={loading} />
      </section>
    </section>
  );
}

function WarehouseKPI({
  label,
  value,
  tone
}: {
  label: string;
  value: number;
  tone: "normal" | "success" | "warning" | "danger" | "info";
}) {
  return (
    <article className="erp-card erp-card--padded erp-kpi-card">
      <div className="erp-kpi-label">{label}</div>
      <strong className="erp-kpi-value">{value}</strong>
      <StatusChip tone={tone}>{label}</StatusChip>
    </article>
  );
}

function statusLabel(status: WarehouseDailyTaskStatus) {
  switch (status) {
    case "handover":
      return "Handover";
    case "mismatch":
      return "Mismatch";
    case "picking":
      return "Picking";
    case "packed":
      return "Packed";
    case "returns":
      return "Returns";
    case "waiting":
    default:
      return "Waiting";
  }
}

function priorityTone(priority: WarehouseDailyTask["priority"]) {
  if (priority === "P0") {
    return "danger";
  }
  if (priority === "P1") {
    return "warning";
  }

  return "normal";
}

function formatDueTime(value: string) {
  return new Intl.DateTimeFormat("en-GB", {
    hour: "2-digit",
    minute: "2-digit"
  }).format(new Date(value));
}
