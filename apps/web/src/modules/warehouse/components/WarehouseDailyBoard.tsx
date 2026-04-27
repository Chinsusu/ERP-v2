"use client";

import { useMemo, useState } from "react";
import {
  DataTable,
  EmptyState,
  ScanInput,
  StatusChip,
  type DataTableColumn,
  type ToastMessage
} from "@/shared/design-system/components";
import { useWarehouseDailyBoard } from "../hooks/useWarehouseDailyBoard";
import { ShiftClosingPanel } from "./ShiftClosingPanel";
import {
  defaultWarehouseDailyBoardShiftCode,
  defaultWarehouseDailyBoardDate,
  warehouseShiftOptions,
  warehouseOptions,
  warehouseTaskTone
} from "../services/warehouseDailyBoardService";
import type { WarehouseDailyBoardQuery, WarehouseDailyShiftCode, WarehouseDailyTask, WarehouseDailyTaskStatus } from "../types";

type QueueFilter = "" | WarehouseDailyTaskStatus | "overdue";

const queueOptions: { label: string; value: QueueFilter }[] = [
  { label: "All active queues", value: "" },
  { label: "New orders", value: "waiting" },
  { label: "Picking", value: "picking" },
  { label: "Packed", value: "packed" },
  { label: "Handover", value: "handover" },
  { label: "Returns", value: "returns" },
  { label: "Stock variance", value: "mismatch" },
  { label: "P0 exceptions", value: "overdue" }
];

const columns: DataTableColumn<WarehouseDailyTask>[] = [
  {
    key: "type",
    header: "Type",
    render: (row) => (
      <span className="erp-warehouse-task-type">
        <strong>{taskTypeLabel(row.status)}</strong>
        <small>{row.title}</small>
      </span>
    ),
    width: "250px"
  },
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
    key: "status",
    header: "Status",
    render: (row) => <StatusChip tone={warehouseTaskTone(row.status)}>{statusLabel(row.status)}</StatusChip>,
    width: "130px"
  },
  {
    key: "source",
    header: "Source",
    render: (row) => (
      <span className="erp-warehouse-source-cell">
        <strong>{sourceLabel(row.source)}</strong>
        <small>{row.sourceField}</small>
      </span>
    ),
    width: "170px"
  },
  {
    key: "owner",
    header: "Owner",
    render: (row) => row.owner,
    width: "120px"
  },
  {
    key: "sla",
    header: "SLA",
    render: (row) => <StatusChip tone={priorityTone(row.priority)}>{formatSla(row)}</StatusChip>,
    width: "130px"
  },
  {
    key: "action",
    header: "Action",
    render: (row) => (
      <a className="erp-button erp-button--secondary erp-button--compact" href={row.href}>
        {taskActionLabel(row.status)}
      </a>
    ),
    width: "150px"
  }
];

export default function WarehouseDailyBoard() {
  const [warehouseId, setWarehouseId] = useState("");
  const [date, setDate] = useState(defaultWarehouseDailyBoardDate);
  const [shiftCode, setShiftCode] = useState<WarehouseDailyShiftCode>(defaultWarehouseDailyBoardShiftCode);
  const [queueFilter, setQueueFilter] = useState<QueueFilter>("");
  const [scanFeedback, setScanFeedback] = useState<ToastMessage | undefined>();
  const query = useMemo<WarehouseDailyBoardQuery>(
    () => ({
      warehouseId: warehouseId || undefined,
      date,
      shiftCode
    }),
    [date, shiftCode, warehouseId]
  );
  const { board, loading } = useWarehouseDailyBoard(query);
  const allTasks = board?.tasks ?? [];
  const visibleTasks = useMemo(() => filterTasksByQueue(allTasks, queueFilter), [allTasks, queueFilter]);
  const exceptions = allTasks.filter((task) => task.priority === "P0" || task.status === "mismatch");
  const selectedWarehouseLabel =
    warehouseOptions.find((option) => option.value === warehouseId)?.label ?? board?.warehouseCode ?? "All warehouses";
  const activeQueueLabel = queueLabel(queueFilter);
  const queueCards: {
    key: QueueFilter;
    label: string;
    value: number;
    tone: "normal" | "success" | "warning" | "danger" | "info";
    helper: string;
  }[] = [
    { key: "waiting", label: "New orders", value: board?.summary.waiting ?? 0, tone: "normal", helper: "Receive" },
    { key: "picking", label: "Picking", value: board?.summary.picking ?? 0, tone: "warning", helper: "Pick" },
    { key: "packed", label: "Packed", value: board?.summary.packed ?? 0, tone: "success", helper: "Pack" },
    { key: "handover", label: "Handover", value: board?.summary.handover ?? 0, tone: "info", helper: "Scan" },
    { key: "returns", label: "Returns", value: board?.summary.returns ?? 0, tone: "warning", helper: "Inspect" },
    {
      key: "mismatch",
      label: "Stock variance",
      value: board?.summary.reconciliationMismatch ?? 0,
      tone: "danger",
      helper: "Reconcile"
    },
    { key: "overdue", label: "P0 exceptions", value: board?.summary.overdue ?? 0, tone: "danger", helper: "Resolve" }
  ];

  return (
    <section className="erp-module-page erp-warehouse-page">
      <header className="erp-page-header">
        <div>
          <p className="erp-module-eyebrow">WH</p>
          <h1 className="erp-page-title">Warehouse Daily Board</h1>
          <p className="erp-page-description">Daily warehouse work queue, handover, returns, and variance control</p>
        </div>
        <div className="erp-page-actions">
          <a className="erp-button erp-button--secondary" href="#scan-station">
            Scan
          </a>
          <a className="erp-button erp-button--primary" href="#shift-closing">
            Shift closing
          </a>
        </div>
      </header>

      <section className="erp-warehouse-context" aria-label="Warehouse shift context">
        <WarehouseContext label="Date" value={formatBoardDate(date)} />
        <WarehouseContext label="Shift" value={shiftLabel(shiftCode)} />
        <WarehouseContext label="Warehouse" value={board?.warehouseCode ?? selectedWarehouseLabel} />
        <WarehouseContext label="Owner" value={board?.owner ?? "Warehouse Lead"} />
        <WarehouseContext label="Shift status" value={statusLabel(board?.shiftStatus ?? "open")} />
      </section>

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
          <span>Shift</span>
          <select
            className="erp-input"
            value={shiftCode}
            onChange={(event) => setShiftCode(event.target.value as WarehouseDailyShiftCode)}
          >
            {warehouseShiftOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Queue</span>
          <select
            className="erp-input"
            value={queueFilter}
            onChange={(event) => setQueueFilter(event.target.value as QueueFilter)}
          >
            {queueOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
      </section>

      <section className="erp-kpi-grid erp-warehouse-kpis">
        {queueCards.map((card) => (
          <WarehouseKPI
            active={queueFilter === card.key}
            helper={card.helper}
            key={card.key}
            label={card.label}
            tone={card.tone}
            value={card.value}
            onSelect={() => setQueueFilter(card.key)}
          />
        ))}
      </section>

      <section className="erp-warehouse-source-strip" aria-label="Counter source fields">
        {(board?.sourceFields ?? []).map((source) => (
          <span className="erp-warehouse-source-item" key={source.counter}>
            <strong>{source.label}</strong>
            <small>{source.fields.join(" / ")}</small>
          </span>
        ))}
      </section>

      <section className="erp-warehouse-ops">
        <div className="erp-card erp-card--padded erp-warehouse-scan-card" id="scan-station">
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
                  <span>
                    {task.title} / {formatSla(task)}
                  </span>
                </a>
              ))
            ) : (
              <span className="erp-warehouse-empty">No P0 exceptions</span>
            )}
          </div>
        </div>
      </section>

      <div id="shift-closing">
        <ShiftClosingPanel query={{ warehouseId: warehouseId || undefined, date, shiftCode }} />
      </div>

      <section className="erp-card erp-card--padded erp-module-table-card">
        <div className="erp-section-header">
          <div>
            <h2 className="erp-section-title">Task board</h2>
            <p className="erp-section-description">{activeQueueLabel}</p>
          </div>
          <StatusChip tone={visibleTasks.length === 0 ? "warning" : "info"}>{visibleTasks.length} rows</StatusChip>
        </div>
        <DataTable
          columns={columns}
          rows={visibleTasks}
          getRowKey={(row) => row.id}
          loading={loading}
          emptyState={<EmptyState title="No tasks in this queue" description="Change the warehouse, date, or queue filter." />}
        />
      </section>
    </section>
  );
}

function WarehouseContext({ label, value }: { label: string; value: string }) {
  return (
    <div className="erp-warehouse-context-item">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function WarehouseKPI({
  active,
  helper,
  label,
  onSelect,
  value,
  tone
}: {
  active: boolean;
  helper: string;
  label: string;
  onSelect: () => void;
  value: number;
  tone: "normal" | "success" | "warning" | "danger" | "info";
}) {
  return (
    <button
      aria-pressed={active}
      className={`erp-card erp-card--padded erp-kpi-card erp-warehouse-kpi-button${active ? " is-active" : ""}`}
      type="button"
      onClick={onSelect}
    >
      <div className="erp-kpi-label">{label}</div>
      <strong className="erp-kpi-value">{value}</strong>
      <span className="erp-warehouse-kpi-footer">
        <StatusChip tone={tone}>{helper}</StatusChip>
        <span className="erp-warehouse-kpi-action">Open queue</span>
      </span>
    </button>
  );
}

function filterTasksByQueue(tasks: WarehouseDailyTask[], queueFilter: QueueFilter) {
  if (queueFilter === "overdue") {
    return tasks.filter((task) => task.priority === "P0");
  }
  if (queueFilter) {
    return tasks.filter((task) => task.status === queueFilter);
  }

  return tasks;
}

function queueLabel(queueFilter: QueueFilter) {
  return queueOptions.find((option) => option.value === queueFilter)?.label ?? "All active queues";
}

function sourceLabel(source: WarehouseDailyTask["source"]) {
  switch (source) {
    case "receiving":
      return "Receiving";
    case "shipping":
      return "Shipping";
    case "returns":
      return "Returns";
    case "stock_movement":
      return "Stock movement";
    case "reconciliation":
      return "Reconciliation";
    case "order_queue":
    default:
      return "Order queue";
  }
}

function shiftLabel(shiftCode: WarehouseDailyShiftCode) {
  return warehouseShiftOptions.find((option) => option.value === shiftCode)?.label ?? shiftCode;
}

function taskTypeLabel(status: WarehouseDailyTaskStatus) {
  switch (status) {
    case "handover":
      return "Handover";
    case "mismatch":
      return "Variance";
    case "picking":
      return "Picking";
    case "packed":
      return "Packed";
    case "returns":
      return "Return";
    case "waiting":
    default:
      return "New order";
  }
}

function statusLabel(status: WarehouseDailyTaskStatus | "open" | "closing" | "closed") {
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
    case "closing":
      return "Closing";
    case "closed":
      return "Closed";
    case "open":
      return "Open";
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

function taskActionLabel(status: WarehouseDailyTaskStatus) {
  switch (status) {
    case "handover":
      return "Scan";
    case "mismatch":
      return "Reconcile";
    case "picking":
      return "Continue";
    case "packed":
      return "Review";
    case "returns":
      return "Inspect";
    case "waiting":
    default:
      return "Start";
  }
}

function formatSla(row: WarehouseDailyTask) {
  const due = formatDueTime(row.dueAt);
  if (row.priority === "P0") {
    return `P0 / ${due}`;
  }

  return `${row.priority} / ${due}`;
}

function formatDueTime(value: string) {
  return new Intl.DateTimeFormat("en-GB", {
    hour: "2-digit",
    minute: "2-digit"
  }).format(new Date(value));
}

function formatBoardDate(value: string) {
  return new Intl.DateTimeFormat("en-GB", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric"
  }).format(new Date(`${value}T00:00:00Z`));
}
