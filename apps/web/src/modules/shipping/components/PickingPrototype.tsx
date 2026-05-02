"use client";

import { useEffect, useMemo, useRef, useState, type KeyboardEvent } from "react";
import {
  DataTable,
  QuantityDisplay,
  StatusChip,
  type DataTableColumn,
  type StatusTone
} from "@/shared/design-system/components";
import { t } from "@/shared/i18n";
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
      header: pickingCopy("columns.task"),
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
      header: pickingCopy("columns.warehouse"),
      render: (row) => row.warehouseCode,
      width: "120px"
    },
    {
      key: "status",
      header: pickingCopy("columns.status"),
      render: (row) => <StatusChip tone={pickTaskStatusTone(row.status)}>{pickTaskStatusLabel(row.status)}</StatusChip>,
      width: "140px"
    },
    {
      key: "lines",
      header: pickingCopy("columns.lines"),
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
          {pickingCopy("actions.select")}
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
    header: pickingCopy("columns.batch"),
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
    header: pickingCopy("columns.location"),
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
    header: pickingCopy("columns.qty"),
    render: (row) => <QuantityDisplay value={row.qtyToPick} uomCode={row.baseUOMCode} />,
    align: "right",
    width: "120px"
  },
  {
    key: "picked",
    header: pickingCopy("columns.picked"),
    render: (row) => <QuantityDisplay value={row.qtyPicked} uomCode={row.baseUOMCode} />,
    align: "right",
    width: "120px"
  },
  {
    key: "state",
    header: pickingCopy("columns.state"),
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
    title: pickingCopy("scan.readyTitle"),
    detail: pickingCopy("scan.readyDetail")
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
      title: pickingCopy("scan.readyTitle"),
      detail: pickingCopy("scan.readyDetail")
    });
  }

  async function handleStartTask() {
    if (!selectedTask || busy) {
      return;
    }
    await runPickAction(async () => {
      const task = await startPickTask(selectedTask.id);
      patchTask(task);
      setFeedback({ tone: "success", title: pickingCopy("scan.started"), detail: task.pickTaskNo });
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
        setFeedback({ tone: "danger", title: pickingCopy("scan.cannotAccept"), detail: pickTaskStatusLabel(task.status) });
        return;
      }

      const matchedLine = task.lines.find((line) => line.status === "pending" && matchesPickScan(line, code));
      if (!matchedLine) {
        setFeedback({ tone: "danger", title: pickingCopy("scan.notMatched"), detail: code.toUpperCase() });
        window.setTimeout(() => scanInputRef.current?.select(), 0);
        return;
      }

      const updated = await confirmPickTaskLine(task.id, matchedLine.id, matchedLine.qtyToPick);
      patchTask(updated);
      setScanValue("");
      setFeedback({
        tone: updated.lines.every((line) => line.status === "picked") ? "success" : "info",
        title: pickingCopy("scan.picked", { sku: matchedLine.skuCode }),
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
      setFeedback({ tone: "success", title: pickingCopy("scan.completed"), detail: task.pickTaskNo });
    });
  }

  async function handleException(code: PickTaskExceptionCode) {
    if (!selectedTask || busy || selectedTask.status === "completed" || exceptionOpen) {
      return;
    }
    await runPickAction(async () => {
      const task = await reportPickTaskException(selectedTask.id, code, pickingCopy("exceptionReason", { code }));
      patchTask(task);
      setFeedback({ tone: "danger", title: pickingCopy("scan.exceptionReported"), detail: pickTaskStatusLabel(task.status) });
    });
  }

  async function runPickAction(action: () => Promise<void>) {
    setBusy(true);
    try {
      await action();
    } catch (cause) {
      setFeedback({
        tone: "danger",
        title: pickingCopy("scan.actionFailed"),
        detail: cause instanceof Error ? cause.message : pickingCopy("scan.unknownError")
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
          <h1 className="erp-page-title">{pickingCopy("title")}</h1>
          <p className="erp-page-description">{pickingCopy("description")}</p>
        </div>
        <div className="erp-page-actions">
          <button className="erp-button erp-button--secondary" type="button" onClick={handleStartTask} disabled={!canStart}>
            {pickingCopy("actions.start")}
          </button>
          <button className="erp-button erp-button--primary" type="button" onClick={handleCompleteTask} disabled={!canComplete || busy}>
            {pickingCopy("actions.complete")}
          </button>
        </div>
      </header>

      <section className="erp-picking-toolbar" aria-label={pickingCopy("filters.label")}>
        <label className="erp-field">
          <span>{pickingCopy("filters.warehouse")}</span>
          <select className="erp-input" value={warehouseId} onChange={(event) => setWarehouseId(event.target.value)}>
            {pickTaskWarehouseOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {pickingWarehouseLabel(option.value, option.label)}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>{pickingCopy("filters.status")}</span>
          <select className="erp-input" value={status} onChange={(event) => setStatus(event.target.value as "" | PickTaskStatus)}>
            {pickTaskStatusOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.value ? pickTaskStatusLabel(option.value) : pickingCopy("filters.allStatuses")}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>{pickingCopy("filters.assignedUser")}</span>
          <input
            className="erp-input"
            value={assignedTo}
            onChange={(event) => setAssignedTo(event.target.value)}
            placeholder={pickingCopy("filters.assignedPlaceholder")}
          />
        </label>
      </section>

      <section className="erp-kpi-grid erp-picking-kpis">
        <PickingKPI label={pickingCopy("kpi.openTasks")} value={totals.open} tone="info" />
        <PickingKPI label={pickingCopy("kpi.pendingLines")} value={totals.pending} tone={totals.pending === 0 ? "success" : "warning"} />
        <PickingKPI label={pickingCopy("kpi.pickedLines")} value={totals.picked} tone="success" />
        <PickingKPI label={pickingCopy("kpi.exceptions")} value={totals.exceptions} tone={totals.exceptions === 0 ? "normal" : "danger"} />
      </section>

      <section className="erp-picking-workspace">
        <div className="erp-card erp-card--padded erp-picking-station-card">
          <div className="erp-section-header">
            <div>
              <h2 className="erp-section-title">{pickingCopy("sections.scanQueue")}</h2>
              <p className="erp-section-description">{selectedTask?.pickTaskNo ?? pickingCopy("empty.noTaskSelected")}</p>
            </div>
            {selectedTask ? <StatusChip tone={pickTaskStatusTone(selectedTask.status)}>{pickTaskStatusLabel(selectedTask.status)}</StatusChip> : null}
          </div>

          <div className="erp-picking-meta-grid">
            <PickingFact label={pickingCopy("facts.order")} value={selectedTask?.orderNo ?? "-"} />
            <PickingFact label={pickingCopy("facts.warehouse")} value={selectedTask?.warehouseCode ?? "-"} />
            <PickingFact label={pickingCopy("facts.nextSku")} value={pendingLine?.skuCode ?? "-"} />
            <PickingFact label={pickingCopy("facts.batch")} value={pendingLine?.batchNo ?? "-"} />
            <PickingFact label={pickingCopy("facts.location")} value={pendingLine?.binCode ?? "-"} />
            <PickingFact
              label={pickingCopy("facts.qty")}
              value={pendingLine ? `${pendingLine.qtyToPick} ${pendingLine.baseUOMCode}` : "-"}
              tone={pendingLine ? "info" : "success"}
            />
          </div>

          <label className={`erp-picking-scan-primary erp-picking-scan-primary--${feedback.tone}`}>
            <span>{pickingCopy("scan.label")}</span>
            <input
              ref={scanInputRef}
              value={scanValue}
              onChange={(event) => setScanValue(event.target.value)}
              onKeyDown={handleScanKeyDown}
              placeholder={
                pendingLine
                  ? `${pendingLine.skuCode} / ${pendingLine.batchNo} / ${pendingLine.binCode}`
                  : pickingCopy("scan.placeholderDone")
              }
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
                {pickingExceptionLabel(option.value)}
              </button>
            ))}
          </div>
        </div>

        <div className="erp-card erp-card--padded erp-picking-queue-card">
          <div className="erp-section-header">
            <h2 className="erp-section-title">{pickingCopy("sections.pickTasks")}</h2>
            <StatusChip tone={displayedTasks.length === 0 ? "warning" : "info"}>
              {pickingCopy("rows", { count: displayedTasks.length })}
            </StatusChip>
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
            <h2 className="erp-section-title">{pickingCopy("sections.pickLines")}</h2>
            <p className="erp-section-description">{pickingCopy("sections.pickLinesDescription")}</p>
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
              <StatusChip tone="warning">{pickingCopy("empty.noTaskSelected")}</StatusChip>
              <strong>{pickingCopy("empty.selectTask")}</strong>
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

function pickingCopy(key: string, values?: Record<string, string | number>) {
  return t(`shipping.picking.${key}`, { values });
}

function pickingWarehouseLabel(value: string, fallback: string) {
  if (value === "") {
    return pickingCopy("warehouse.all");
  }

  return t(`shipping.picking.warehouse.${value}`, { fallback });
}

function pickingExceptionLabel(value: PickTaskExceptionCode) {
  return pickingCopy(`exceptions.${value}`);
}

function pickTaskStatusLabel(status: PickTaskStatus) {
  return pickingCopy(`status.${status}`);
}

function pickTaskLineStatusLabel(status: PickTaskLine["status"]) {
  return pickingCopy(`status.${status}`);
}
