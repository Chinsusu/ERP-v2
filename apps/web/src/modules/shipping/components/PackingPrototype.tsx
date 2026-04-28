"use client";

import { useEffect, useMemo, useRef, useState, type KeyboardEvent } from "react";
import {
  DataTable,
  QuantityDisplay,
  StatusChip,
  type DataTableColumn,
  type StatusTone
} from "@/shared/design-system/components";
import { usePackTasks } from "../hooks/usePackTasks";
import {
  confirmPackTask,
  packTaskExceptionOptions,
  packTaskLineStatusTone,
  packTaskStatusOptions,
  packTaskStatusTone,
  packTaskWarehouseOptions,
  reportPackTaskException,
  startPackTask
} from "../services/packTaskService";
import type {
  ConfirmPackTaskLineInput,
  PackTask,
  PackTaskExceptionCode,
  PackTaskLine,
  PackTaskQuery,
  PackTaskStatus
} from "../types";

function createTaskColumns(
  selectedTaskId: string,
  onSelectTask: (taskId: string) => void
): DataTableColumn<PackTask>[] {
  return [
    {
      key: "task",
      header: "Pack task",
      render: (row) => (
        <span className="erp-packing-task-cell">
          <strong>{row.packTaskNo}</strong>
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
      render: (row) => <StatusChip tone={packTaskStatusTone(row.status)}>{packTaskStatusLabel(row.status)}</StatusChip>,
      width: "140px"
    },
    {
      key: "lines",
      header: "Lines",
      render: (row) => `${packedLineCount(row)}/${row.lines.length}`,
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

const lineColumns: DataTableColumn<PackTaskLine>[] = [
  {
    key: "sku",
    header: "SKU",
    render: (row) => (
      <span className="erp-packing-line-cell">
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
      <span className="erp-packing-line-cell">
        <strong>{row.batchNo}</strong>
        <small>{row.batchId}</small>
      </span>
    ),
    width: "170px"
  },
  {
    key: "qty",
    header: "Qty",
    render: (row) => <QuantityDisplay value={row.qtyToPack} uomCode={row.baseUOMCode} />,
    align: "right",
    width: "120px"
  },
  {
    key: "packed",
    header: "Packed",
    render: (row) => <QuantityDisplay value={row.qtyPacked} uomCode={row.baseUOMCode} />,
    align: "right",
    width: "120px"
  },
  {
    key: "state",
    header: "State",
    render: (row) => <StatusChip tone={packTaskLineStatusTone(row.status)}>{packTaskLineStatusLabel(row.status)}</StatusChip>,
    width: "130px"
  }
];

type ScanFeedback = {
  tone: StatusTone;
  title: string;
  detail?: string;
};

export function PackingPrototype() {
  const scanInputRef = useRef<HTMLInputElement>(null);
  const [warehouseId, setWarehouseId] = useState("wh-hcm-fg");
  const [status, setStatus] = useState<"" | PackTaskStatus>("");
  const [assignedTo, setAssignedTo] = useState("");
  const [selectedTaskId, setSelectedTaskId] = useState("");
  const [scanValue, setScanValue] = useState("");
  const [packingNote, setPackingNote] = useState("");
  const [feedback, setFeedback] = useState<ScanFeedback>({
    tone: "info",
    title: "Ready to pack",
    detail: "Scan SKU, batch, or pack line code"
  });
  const [busy, setBusy] = useState(false);
  const [localTasks, setLocalTasks] = useState<PackTask[]>([]);
  const query = useMemo<PackTaskQuery>(
    () => ({
      warehouseId: warehouseId || undefined,
      status: status || undefined,
      assignedTo: assignedTo || undefined
    }),
    [assignedTo, status, warehouseId]
  );
  const { packTasks, loading, error } = usePackTasks(query);
  const displayedTasks = useMemo(() => {
    const localMatches = localTasks.filter((task) => matchesPackTaskQuery(task, query));
    const localIds = new Set(localMatches.map((task) => task.id));

    return [...localMatches, ...packTasks.filter((task) => !localIds.has(task.id))];
  }, [localTasks, packTasks, query]);
  const selectedTask =
    displayedTasks.find((task) => task.id === selectedTaskId) ?? displayedTasks[0] ?? null;
  const pendingLine = selectedTask?.lines.find((line) => line.status === "pending") ?? null;
  const exceptionOpen = selectedTask ? isExceptionStatus(selectedTask.status) : false;
  const canStart = selectedTask !== null && !busy && selectedTask.status === "created";
  const canConfirm = selectedTask !== null && !busy && selectedTask.status === "in_progress" && selectedTask.lines.every((line) => line.status === "pending");
  const totals = displayedTasks.reduce(
    (acc, task) => ({
      open: acc.open + (task.status === "created" || task.status === "in_progress" ? 1 : 0),
      packed: acc.packed + packedLineCount(task),
      pending: acc.pending + task.lines.filter((line) => line.status === "pending").length,
      exceptions: acc.exceptions + (isExceptionStatus(task.status) ? 1 : 0)
    }),
    { open: 0, packed: 0, pending: 0, exceptions: 0 }
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
    setPackingNote("");
    setFeedback({
      tone: "info",
      title: "Ready to pack",
      detail: "Scan SKU, batch, or pack line code"
    });
  }

  async function handleStartTask() {
    if (!selectedTask || busy) {
      return;
    }
    await runPackAction(async () => {
      const task = await startPackTask(selectedTask.id);
      patchTask(task);
      setFeedback({ tone: "success", title: "Pack task started", detail: task.packTaskNo });
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

    await runPackAction(async () => {
      let task = selectedTask;
      if (task.status === "created") {
        task = await startPackTask(task.id);
        patchTask(task);
      }
      if (task.status !== "in_progress") {
        setFeedback({ tone: "danger", title: "Task cannot accept scans", detail: packTaskStatusLabel(task.status) });
        return;
      }

      const matchedLine = task.lines.find((line) => line.status === "pending" && matchesPackScan(line, code));
      if (!matchedLine) {
        setFeedback({ tone: "danger", title: "Scan does not match pending line", detail: code.toUpperCase() });
        window.setTimeout(() => scanInputRef.current?.select(), 0);
        return;
      }

      setScanValue("");
      setFeedback({
        tone: "success",
        title: `${matchedLine.skuCode} verified`,
        detail: `${matchedLine.batchNo} / ${matchedLine.qtyToPack} ${matchedLine.baseUOMCode}`
      });
      window.setTimeout(() => scanInputRef.current?.focus(), 0);
    });
  }

  async function handleConfirmTask() {
    if (!selectedTask || !canConfirm || busy) {
      return;
    }
    await runPackAction(async () => {
      const lines: ConfirmPackTaskLineInput[] = selectedTask.lines.map((line) => ({
        lineId: line.id,
        packedQty: line.qtyToPack
      }));
      const task = await confirmPackTask(selectedTask.id, lines);
      patchTask(task);
      setFeedback({ tone: "success", title: "Pack confirmed", detail: task.orderNo });
    });
  }

  async function handleException(code: PackTaskExceptionCode) {
    if (!selectedTask || busy || selectedTask.status === "packed" || exceptionOpen) {
      return;
    }
    await runPackAction(async () => {
      const task = await reportPackTaskException(
        selectedTask.id,
        code,
        packingNote || `Reported from packing station: ${code}`,
        pendingLine?.id
      );
      patchTask(task);
      setFeedback({ tone: "danger", title: "Exception reported", detail: packTaskStatusLabel(task.status) });
    });
  }

  async function runPackAction(action: () => Promise<void>) {
    setBusy(true);
    try {
      await action();
    } catch (cause) {
      setFeedback({
        tone: "danger",
        title: "Packing action failed",
        detail: cause instanceof Error ? cause.message : "Unknown error"
      });
    } finally {
      setBusy(false);
    }
  }

  function patchTask(task: PackTask) {
    setLocalTasks((current) => [task, ...current.filter((candidate) => candidate.id !== task.id)]);
    setSelectedTaskId(task.id);
  }

  return (
    <section className="erp-module-page erp-packing-page">
      <header className="erp-page-header">
        <div>
          <p className="erp-module-eyebrow">PACK</p>
          <h1 className="erp-page-title">Packing Station</h1>
          <p className="erp-page-description">Confirm packed SKU, batch, and base UOM quantity before carrier manifest</p>
        </div>
        <div className="erp-page-actions">
          <button className="erp-button erp-button--secondary" type="button" onClick={handleStartTask} disabled={!canStart}>
            Start
          </button>
          <button className="erp-button erp-button--primary" type="button" onClick={handleConfirmTask} disabled={!canConfirm}>
            Confirm pack
          </button>
        </div>
      </header>

      <section className="erp-packing-toolbar" aria-label="Packing filters">
        <label className="erp-field">
          <span>Warehouse</span>
          <select className="erp-input" value={warehouseId} onChange={(event) => setWarehouseId(event.target.value)}>
            {packTaskWarehouseOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Status</span>
          <select className="erp-input" value={status} onChange={(event) => setStatus(event.target.value as "" | PackTaskStatus)}>
            {packTaskStatusOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Assigned user</span>
          <input className="erp-input" value={assignedTo} onChange={(event) => setAssignedTo(event.target.value)} placeholder="user-packer" />
        </label>
      </section>

      <section className="erp-kpi-grid erp-packing-kpis">
        <PackingKPI label="Open tasks" value={totals.open} tone="info" />
        <PackingKPI label="Pending lines" value={totals.pending} tone={totals.pending === 0 ? "success" : "warning"} />
        <PackingKPI label="Packed lines" value={totals.packed} tone="success" />
        <PackingKPI label="Exceptions" value={totals.exceptions} tone={totals.exceptions === 0 ? "normal" : "danger"} />
      </section>

      <section className="erp-packing-workspace">
        <div className="erp-card erp-card--padded erp-packing-station-card">
          <div className="erp-section-header">
            <div>
              <h2 className="erp-section-title">Packing queue</h2>
              <p className="erp-section-description">{selectedTask?.packTaskNo ?? "No pack task selected"}</p>
            </div>
            {selectedTask ? <StatusChip tone={packTaskStatusTone(selectedTask.status)}>{packTaskStatusLabel(selectedTask.status)}</StatusChip> : null}
          </div>

          <div className="erp-packing-meta-grid">
            <PackingFact label="Order" value={selectedTask?.orderNo ?? "-"} />
            <PackingFact label="Pick task" value={selectedTask?.pickTaskNo ?? "-"} />
            <PackingFact label="Warehouse" value={selectedTask?.warehouseCode ?? "-"} />
            <PackingFact label="Next SKU" value={pendingLine?.skuCode ?? "-"} />
            <PackingFact label="Batch" value={pendingLine?.batchNo ?? "-"} />
            <PackingFact
              label="Qty"
              value={pendingLine ? `${pendingLine.qtyToPack} ${pendingLine.baseUOMCode}` : "-"}
              tone={pendingLine ? "info" : "success"}
            />
          </div>

          <label className={`erp-packing-scan-primary erp-packing-scan-primary--${feedback.tone}`}>
            <span>Scan SKU / batch / pack line</span>
            <input
              ref={scanInputRef}
              value={scanValue}
              onChange={(event) => setScanValue(event.target.value)}
              onKeyDown={handleScanKeyDown}
              placeholder={pendingLine ? `${pendingLine.skuCode} / ${pendingLine.batchNo} / ${pendingLine.id}` : "All lines packed"}
              disabled={!selectedTask || busy || !pendingLine || selectedTask.status === "packed" || exceptionOpen}
            />
            <small>
              <strong>{feedback.title}</strong>
              {feedback.detail ? <span>{feedback.detail}</span> : null}
            </small>
          </label>

          <label className="erp-field erp-packing-note">
            <span>Packing note</span>
            <textarea
              className="erp-input"
              value={packingNote}
              onChange={(event) => setPackingNote(event.target.value)}
              placeholder="Seal damage, shortage, wrong item, or package note"
              rows={3}
            />
          </label>

          <div className="erp-packing-exception-actions">
            {packTaskExceptionOptions.map((option) => (
              <button
                className="erp-button erp-button--danger"
                type="button"
                key={option.value}
                onClick={() => handleException(option.value)}
                disabled={!selectedTask || busy || selectedTask.status === "packed" || exceptionOpen}
              >
                {option.label}
              </button>
            ))}
          </div>
        </div>

        <div className="erp-card erp-card--padded erp-packing-queue-card">
          <div className="erp-section-header">
            <h2 className="erp-section-title">Pack tasks</h2>
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
            <h2 className="erp-section-title">Pack lines</h2>
            <p className="erp-section-description">SKU, batch, base UOM quantity, and packed quantity</p>
          </div>
          <StatusChip tone={selectedTask && selectedTask.lines.every((line) => line.status === "packed") ? "success" : "warning"}>
            {selectedTask ? `${packedLineCount(selectedTask)}/${selectedTask.lines.length}` : "0/0"}
          </StatusChip>
        </div>
        <DataTable
          columns={lineColumns}
          rows={selectedTask?.lines ?? []}
          getRowKey={(row) => row.id}
          emptyState={
            <section className="erp-shipping-empty-state">
              <StatusChip tone="warning">No task</StatusChip>
              <strong>Select a pack task</strong>
            </section>
          }
        />
      </section>
    </section>
  );
}

function PackingKPI({ label, value, tone }: { label: string; value: number; tone: StatusTone }) {
  return (
    <article className="erp-card erp-card--padded erp-kpi-card">
      <div className="erp-kpi-label">{label}</div>
      <strong className="erp-kpi-value">{value}</strong>
      <StatusChip tone={tone}>{label}</StatusChip>
    </article>
  );
}

function PackingFact({
  label,
  value,
  tone = "normal"
}: {
  label: string;
  value: string | number;
  tone?: StatusTone;
}) {
  return (
    <div className="erp-packing-fact">
      <span>{label}</span>
      <strong>{value}</strong>
      {tone !== "normal" ? <StatusChip tone={tone}>{label}</StatusChip> : null}
    </div>
  );
}

function matchesPackScan(line: PackTaskLine, code: string) {
  const normalized = normalizeScanCode(code);
  return [line.id, line.skuCode, line.batchNo, line.batchId, line.salesOrderLineId, line.pickTaskLineId].some(
    (candidate) => normalizeScanCode(candidate) === normalized
  );
}

function matchesPackTaskQuery(task: PackTask, query: PackTaskQuery) {
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

function packedLineCount(task: PackTask) {
  return task.lines.filter((line) => line.status === "packed").length;
}

function isExceptionStatus(status: PackTaskStatus) {
  return ["pack_exception", "cancelled"].includes(status);
}

function packTaskStatusLabel(status: PackTaskStatus) {
  switch (status) {
    case "created":
      return "Created";
    case "in_progress":
      return "In progress";
    case "packed":
      return "Packed";
    case "pack_exception":
      return "Pack exception";
    case "cancelled":
      return "Cancelled";
    default:
      return status;
  }
}

function packTaskLineStatusLabel(status: PackTaskLine["status"]) {
  switch (status) {
    case "pending":
      return "Pending";
    case "packed":
      return "Packed";
    case "pack_exception":
      return "Pack exception";
    case "cancelled":
      return "Cancelled";
    default:
      return status;
  }
}
