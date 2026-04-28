"use client";

import { useEffect, useMemo, useRef, useState, type KeyboardEvent } from "react";
import {
  DataTable,
  QuantityDisplay,
  StatusChip,
  type DataTableColumn,
  type StatusTone
} from "@/shared/design-system/components";
import { usePickTasks } from "../hooks/usePickTasks";
import {
  completePickTask,
  confirmPickTaskLine,
  pickTaskExceptionOptions,
  pickTaskLineStatusTone,
  pickTaskStatusOptions,
  pickTaskStatusTone,
  pickTaskWarehouseOptions,
  reportPickTaskException,
  startPickTask
} from "../services/pickTaskService";
import type { PickTask, PickTaskExceptionCode, PickTaskLine, PickTaskQuery, PickTaskStatus } from "../types";

function createTaskColumns(
  selectedTaskId: string,
  onSelectTask: (taskId: string) => void
): DataTableColumn<PickTask>[] {
  return [
    {
      key: "task",
      header: "Pick task",
      render: (row) => (
        <span className="erp-picking-task-cell">
          <strong>{row.pickTaskNo}</strong>
          <small>{row.orderNo}</small>
        </span>
      ),
      width: "220px"
    },
    {
      key: "warehouse",
      header: "Warehouse",
      render: (row) => row.warehouseCode,
      width: "120px"
    },
    {
      key: "status",
      header: "Status",
      render: (row) => <StatusChip tone={pickTaskStatusTone(row.status)}>{pickTaskStatusLabel(row.status)}</StatusChip>,
      width: "140px"
    },
    {
      key: "lines",
      header: "Lines",
      render: (row) => `${pickedLineCount(row)}/${row.lines.length}`,
      align: "right",
      width: "90px"
    },
    {
      key: "select",
      header: "",
      render: (row) => (
        <button
          className={`erp-button ${row.id === selectedTaskId ? "erp-button--primary" : "erp-button--secondary"}`}
          type="button"
          onClick={() => onSelectTask(row.id)}
        >
          Select
        </button>
      ),
      align: "right",
      width: "110px"
    }
  ];
}

const lineColumns: DataTableColumn<PickTaskLine>[] = [
  {
    key: "sku",
    header: "SKU",
    render: (row) => (
      <span className="erp-picking-line-cell">
        <strong>{row.skuCode}</strong>
        <small>{row.itemId}</small>
      </span>
    ),
    width: "180px"
  },
  {
    key: "batch",
    header: "Batch",
    render: (row) => (
      <span className="erp-picking-line-cell">
        <strong>{row.batchNo}</strong>
        <small>{row.batchId}</small>
      </span>
    ),
    width: "170px"
  },
  {
    key: "location",
    header: "Location",
    render: (row) => (
      <span className="erp-picking-line-cell">
        <strong>{row.binCode}</strong>
        <small>{row.binId}</small>
      </span>
    ),
    width: "170px"
  },
  {
    key: "qty",
    header: "Qty",
    render: (row) => <QuantityDisplay value={row.qtyToPick} uomCode={row.baseUOMCode} />,
    align: "right",
    width: "120px"
  },
  {
    key: "picked",
    header: "Picked",
    render: (row) => <QuantityDisplay value={row.qtyPicked} uomCode={row.baseUOMCode} />,
    align: "right",
    width: "120px"
  },
  {
    key: "state",
    header: "State",
    render: (row) => <StatusChip tone={pickTaskLineStatusTone(row.status)}>{pickTaskLineStatusLabel(row.status)}</StatusChip>,
    width: "130px"
  }
];

type ScanFeedback = {
  tone: StatusTone;
  title: string;
  detail?: string;
};

export function PickingPrototype() {
  const scanInputRef = useRef<HTMLInputElement>(null);
  const [warehouseId, setWarehouseId] = useState("wh-hcm-fg");
  const [status, setStatus] = useState<"" | PickTaskStatus>("");
  const [assignedTo, setAssignedTo] = useState("");
  const [selectedTaskId, setSelectedTaskId] = useState("");
  const [scanValue, setScanValue] = useState("");
  const [feedback, setFeedback] = useState<ScanFeedback>({
    tone: "info",
    title: "Ready to scan",
    detail: "Scan SKU, batch, bin, or pick line code"
  });
  const [busy, setBusy] = useState(false);
  const [localTasks, setLocalTasks] = useState<PickTask[]>([]);
  const query = useMemo<PickTaskQuery>(
    () => ({
      warehouseId: warehouseId || undefined,
      status: status || undefined,
      assignedTo: assignedTo || undefined
    }),
    [assignedTo, status, warehouseId]
  );
  const { pickTasks, loading, error } = usePickTasks(query);
  const displayedTasks = useMemo(() => {
    const localMatches = localTasks.filter((task) => matchesPickTaskQuery(task, query));
    const localIds = new Set(localMatches.map((task) => task.id));

    return [...localMatches, ...pickTasks.filter((task) => !localIds.has(task.id))];
  }, [localTasks, pickTasks, query]);
  const selectedTask =
    displayedTasks.find((task) => task.id === selectedTaskId) ?? displayedTasks[0] ?? null;
  const pendingLine = selectedTask?.lines.find((line) => line.status === "pending") ?? null;
  const exceptionOpen = selectedTask ? isExceptionStatus(selectedTask.status) : false;
  const canStart = selectedTask !== null && !busy && (selectedTask.status === "created" || selectedTask.status === "assigned");
  const canComplete = selectedTask !== null && selectedTask.status === "in_progress" && selectedTask.lines.every((line) => line.status === "picked");
  const totals = displayedTasks.reduce(
    (acc, task) => ({
      open: acc.open + (task.status === "created" || task.status === "assigned" || task.status === "in_progress" ? 1 : 0),
      picked: acc.picked + pickedLineCount(task),
      pending: acc.pending + task.lines.filter((line) => line.status === "pending").length,
      exceptions: acc.exceptions + (isExceptionStatus(task.status) ? 1 : 0)
    }),
    { open: 0, picked: 0, pending: 0, exceptions: 0 }
  );
  const taskColumns = useMemo(() => createTaskColumns(selectedTask?.id ?? "", handleSelectTask), [selectedTask?.id]);

  useEffect(() => {
    if (!selectedTaskId && displayedTasks.length > 0) {
      setSelectedTaskId(displayedTasks[0].id);
    }
  }, [displayedTasks, selectedTaskId]);

  useEffect(() => {
    window.setTimeout(() => scanInputRef.current?.focus(), 0);
  }, [selectedTask?.id, pendingLine?.id]);

  function handleSelectTask(taskId: string) {
    setSelectedTaskId(taskId);
    setScanValue("");
    setFeedback({
      tone: "info",
      title: "Ready to scan",
      detail: "Scan SKU, batch, bin, or pick line code"
    });
  }

  async function handleStartTask() {
    if (!selectedTask || busy) {
      return;
    }
    await runPickAction(async () => {
      const task = await startPickTask(selectedTask.id);
      patchTask(task);
      setFeedback({ tone: "success", title: "Pick task started", detail: task.pickTaskNo });
    });
  }

  async function handleScanKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (event.key !== "Enter" || busy) {
      return;
    }
    const code = scanValue.trim();
    if (!selectedTask || code === "") {
      return;
    }

    await runPickAction(async () => {
      let task = selectedTask;
      if (task.status === "created" || task.status === "assigned") {
        task = await startPickTask(task.id);
        patchTask(task);
      }
      if (task.status !== "in_progress") {
        setFeedback({ tone: "danger", title: "Task cannot accept scans", detail: pickTaskStatusLabel(task.status) });
        return;
      }

      const matchedLine = task.lines.find((line) => line.status === "pending" && matchesPickScan(line, code));
      if (!matchedLine) {
        setFeedback({ tone: "danger", title: "Scan does not match pending line", detail: code.toUpperCase() });
        window.setTimeout(() => scanInputRef.current?.select(), 0);
        return;
      }

      const updated = await confirmPickTaskLine(task.id, matchedLine.id, matchedLine.qtyToPick);
      patchTask(updated);
      setScanValue("");
      setFeedback({
        tone: updated.lines.every((line) => line.status === "picked") ? "success" : "info",
        title: `${matchedLine.skuCode} picked`,
        detail: `${matchedLine.batchNo} / ${matchedLine.binCode}`
      });
      window.setTimeout(() => scanInputRef.current?.focus(), 0);
    });
  }

  async function handleCompleteTask() {
    if (!selectedTask || !canComplete || busy) {
      return;
    }
    await runPickAction(async () => {
      const task = await completePickTask(selectedTask.id);
      patchTask(task);
      setFeedback({ tone: "success", title: "Pick task completed", detail: task.pickTaskNo });
    });
  }

  async function handleException(code: PickTaskExceptionCode) {
    if (!selectedTask || busy || selectedTask.status === "completed" || exceptionOpen) {
      return;
    }
    await runPickAction(async () => {
      const task = await reportPickTaskException(selectedTask.id, code, `Reported from picking station: ${code}`);
      patchTask(task);
      setFeedback({ tone: "danger", title: "Exception reported", detail: pickTaskStatusLabel(task.status) });
    });
  }

  async function runPickAction(action: () => Promise<void>) {
    setBusy(true);
    try {
      await action();
    } catch (cause) {
      setFeedback({
        tone: "danger",
        title: "Picking action failed",
        detail: cause instanceof Error ? cause.message : "Unknown error"
      });
    } finally {
      setBusy(false);
    }
  }

  function patchTask(task: PickTask) {
    setLocalTasks((current) => [task, ...current.filter((candidate) => candidate.id !== task.id)]);
    setSelectedTaskId(task.id);
  }

  return (
    <section className="erp-module-page erp-picking-page">
      <header className="erp-page-header">
        <div>
          <p className="erp-module-eyebrow">PICK</p>
          <h1 className="erp-page-title">Picking Station</h1>
          <p className="erp-page-description">Scan-first picking by SKU, batch, location, and base UOM quantity</p>
        </div>
        <div className="erp-page-actions">
          <button className="erp-button erp-button--secondary" type="button" onClick={handleStartTask} disabled={!canStart}>
            Start
          </button>
          <button className="erp-button erp-button--primary" type="button" onClick={handleCompleteTask} disabled={!canComplete || busy}>
            Complete
          </button>
        </div>
      </header>

      <section className="erp-picking-toolbar" aria-label="Picking filters">
        <label className="erp-field">
          <span>Warehouse</span>
          <select className="erp-input" value={warehouseId} onChange={(event) => setWarehouseId(event.target.value)}>
            {pickTaskWarehouseOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Status</span>
          <select className="erp-input" value={status} onChange={(event) => setStatus(event.target.value as "" | PickTaskStatus)}>
            {pickTaskStatusOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Assigned user</span>
          <input className="erp-input" value={assignedTo} onChange={(event) => setAssignedTo(event.target.value)} placeholder="user-picker" />
        </label>
      </section>

      <section className="erp-kpi-grid erp-picking-kpis">
        <PickingKPI label="Open tasks" value={totals.open} tone="info" />
        <PickingKPI label="Pending lines" value={totals.pending} tone={totals.pending === 0 ? "success" : "warning"} />
        <PickingKPI label="Picked lines" value={totals.picked} tone="success" />
        <PickingKPI label="Exceptions" value={totals.exceptions} tone={totals.exceptions === 0 ? "normal" : "danger"} />
      </section>

      <section className="erp-picking-workspace">
        <div className="erp-card erp-card--padded erp-picking-station-card">
          <div className="erp-section-header">
            <div>
              <h2 className="erp-section-title">Scan queue</h2>
              <p className="erp-section-description">{selectedTask?.pickTaskNo ?? "No pick task selected"}</p>
            </div>
            {selectedTask ? <StatusChip tone={pickTaskStatusTone(selectedTask.status)}>{pickTaskStatusLabel(selectedTask.status)}</StatusChip> : null}
          </div>

          <div className="erp-picking-meta-grid">
            <PickingFact label="Order" value={selectedTask?.orderNo ?? "-"} />
            <PickingFact label="Warehouse" value={selectedTask?.warehouseCode ?? "-"} />
            <PickingFact label="Next SKU" value={pendingLine?.skuCode ?? "-"} />
            <PickingFact label="Batch" value={pendingLine?.batchNo ?? "-"} />
            <PickingFact label="Location" value={pendingLine?.binCode ?? "-"} />
            <PickingFact
              label="Qty"
              value={pendingLine ? `${pendingLine.qtyToPick} ${pendingLine.baseUOMCode}` : "-"}
              tone={pendingLine ? "info" : "success"}
            />
          </div>

          <label className={`erp-picking-scan-primary erp-picking-scan-primary--${feedback.tone}`}>
            <span>Scan SKU / batch / location</span>
            <input
              ref={scanInputRef}
              value={scanValue}
              onChange={(event) => setScanValue(event.target.value)}
              onKeyDown={handleScanKeyDown}
              placeholder={pendingLine ? `${pendingLine.skuCode} / ${pendingLine.batchNo} / ${pendingLine.binCode}` : "All lines picked"}
              disabled={!selectedTask || busy || !pendingLine || selectedTask.status === "completed" || exceptionOpen}
            />
            <small>
              <strong>{feedback.title}</strong>
              {feedback.detail ? <span>{feedback.detail}</span> : null}
            </small>
          </label>

          <div className="erp-picking-exception-actions">
            {pickTaskExceptionOptions.map((option) => (
              <button
                className="erp-button erp-button--danger"
                type="button"
                key={option.value}
                onClick={() => handleException(option.value)}
                disabled={!selectedTask || busy || selectedTask.status === "completed" || exceptionOpen}
              >
                {option.label}
              </button>
            ))}
          </div>
        </div>

        <div className="erp-card erp-card--padded erp-picking-queue-card">
          <div className="erp-section-header">
            <h2 className="erp-section-title">Pick tasks</h2>
            <StatusChip tone={displayedTasks.length === 0 ? "warning" : "info"}>{displayedTasks.length} rows</StatusChip>
          </div>
          <DataTable
            columns={taskColumns}
            rows={displayedTasks}
            getRowKey={(row) => row.id}
            loading={loading}
            error={error?.message}
          />
        </div>
      </section>

      <section className="erp-card erp-card--padded erp-module-table-card">
        <div className="erp-section-header">
          <div>
            <h2 className="erp-section-title">Pick lines</h2>
            <p className="erp-section-description">SKU, batch, location, base UOM quantity, and picked quantity</p>
          </div>
          <StatusChip tone={selectedTask && selectedTask.lines.every((line) => line.status === "picked") ? "success" : "warning"}>
            {selectedTask ? `${pickedLineCount(selectedTask)}/${selectedTask.lines.length}` : "0/0"}
          </StatusChip>
        </div>
        <DataTable
          columns={lineColumns}
          rows={selectedTask?.lines ?? []}
          getRowKey={(row) => row.id}
          emptyState={
            <section className="erp-shipping-empty-state">
              <StatusChip tone="warning">No task</StatusChip>
              <strong>Select a pick task</strong>
            </section>
          }
        />
      </section>
    </section>
  );
}

function PickingKPI({ label, value, tone }: { label: string; value: number; tone: StatusTone }) {
  return (
    <article className="erp-card erp-card--padded erp-kpi-card">
      <div className="erp-kpi-label">{label}</div>
      <strong className="erp-kpi-value">{value}</strong>
      <StatusChip tone={tone}>{label}</StatusChip>
    </article>
  );
}

function PickingFact({
  label,
  value,
  tone = "normal"
}: {
  label: string;
  value: string | number;
  tone?: StatusTone;
}) {
  return (
    <div className="erp-picking-fact">
      <span>{label}</span>
      <strong>{value}</strong>
      {tone !== "normal" ? <StatusChip tone={tone}>{label}</StatusChip> : null}
    </div>
  );
}

function matchesPickScan(line: PickTaskLine, code: string) {
  const normalized = normalizeScanCode(code);
  return [line.id, line.skuCode, line.batchNo, line.batchId, line.binCode, line.binId].some(
    (candidate) => normalizeScanCode(candidate) === normalized
  );
}

function matchesPickTaskQuery(task: PickTask, query: PickTaskQuery) {
  if (query.warehouseId && task.warehouseId !== query.warehouseId) {
    return false;
  }
  if (query.status && task.status !== query.status) {
    return false;
  }
  if (query.assignedTo && task.assignedTo !== query.assignedTo) {
    return false;
  }

  return true;
}

function normalizeScanCode(value: string) {
  return value.trim().toUpperCase();
}

function pickedLineCount(task: PickTask) {
  return task.lines.filter((line) => line.status === "picked").length;
}

function isExceptionStatus(status: PickTaskStatus) {
  return ["missing_stock", "wrong_sku", "wrong_batch", "wrong_location", "cancelled"].includes(status);
}

function pickTaskStatusLabel(status: PickTaskStatus) {
  switch (status) {
    case "created":
      return "Created";
    case "assigned":
      return "Assigned";
    case "in_progress":
      return "In progress";
    case "completed":
      return "Completed";
    case "missing_stock":
      return "Missing stock";
    case "wrong_sku":
      return "Wrong SKU";
    case "wrong_batch":
      return "Wrong batch";
    case "wrong_location":
      return "Wrong location";
    case "cancelled":
      return "Cancelled";
    default:
      return status;
  }
}

function pickTaskLineStatusLabel(status: PickTaskLine["status"]) {
  switch (status) {
    case "pending":
      return "Pending";
    case "picked":
      return "Picked";
    case "missing_stock":
      return "Missing stock";
    case "wrong_sku":
      return "Wrong SKU";
    case "wrong_batch":
      return "Wrong batch";
    case "wrong_location":
      return "Wrong location";
    case "cancelled":
      return "Cancelled";
    default:
      return status;
  }
}
