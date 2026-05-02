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
      header: packingCopy("columns.task"),
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
      header: packingCopy("columns.warehouse"),
      render: (row) => row.warehouseCode,
      width: "120px"
    },
    {
      key: "status",
      header: packingCopy("columns.status"),
      render: (row) => <StatusChip tone={packTaskStatusTone(row.status)}>{packTaskStatusLabel(row.status)}</StatusChip>,
      width: "140px"
    },
    {
      key: "lines",
      header: packingCopy("columns.lines"),
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
          {packingCopy("actions.select")}
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
    header: packingCopy("columns.batch"),
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
    header: packingCopy("columns.qty"),
    render: (row) => <QuantityDisplay value={row.qtyToPack} uomCode={row.baseUOMCode} />,
    align: "right",
    width: "120px"
  },
  {
    key: "packed",
    header: packingCopy("columns.packed"),
    render: (row) => <QuantityDisplay value={row.qtyPacked} uomCode={row.baseUOMCode} />,
    align: "right",
    width: "120px"
  },
  {
    key: "state",
    header: packingCopy("columns.state"),
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
    title: packingCopy("scan.readyTitle"),
    detail: packingCopy("scan.readyDetail")
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
      title: packingCopy("scan.readyTitle"),
      detail: packingCopy("scan.readyDetail")
    });
  }

  async function handleStartTask() {
    if (!selectedTask || busy) {
      return;
    }
    await runPackAction(async () => {
      const task = await startPackTask(selectedTask.id);
      patchTask(task);
      setFeedback({ tone: "success", title: packingCopy("scan.started"), detail: task.packTaskNo });
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
        setFeedback({ tone: "danger", title: packingCopy("scan.cannotAccept"), detail: packTaskStatusLabel(task.status) });
        return;
      }

      const matchedLine = task.lines.find((line) => line.status === "pending" && matchesPackScan(line, code));
      if (!matchedLine) {
        setFeedback({ tone: "danger", title: packingCopy("scan.notMatched"), detail: code.toUpperCase() });
        window.setTimeout(() => scanInputRef.current?.select(), 0);
        return;
      }

      setScanValue("");
      setFeedback({
        tone: "success",
        title: packingCopy("scan.verified", { sku: matchedLine.skuCode }),
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
      setFeedback({ tone: "success", title: packingCopy("scan.confirmed"), detail: task.orderNo });
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
        packingNote || packingCopy("exceptionReason", { code }),
        pendingLine?.id
      );
      patchTask(task);
      setFeedback({ tone: "danger", title: packingCopy("scan.exceptionReported"), detail: packTaskStatusLabel(task.status) });
    });
  }

  async function runPackAction(action: () => Promise<void>) {
    setBusy(true);
    try {
      await action();
    } catch (cause) {
      setFeedback({
        tone: "danger",
        title: packingCopy("scan.actionFailed"),
        detail: cause instanceof Error ? cause.message : packingCopy("scan.unknownError")
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
          <h1 className="erp-page-title">{packingCopy("title")}</h1>
          <p className="erp-page-description">{packingCopy("description")}</p>
        </div>
        <div className="erp-page-actions">
          <button className="erp-button erp-button--secondary" type="button" onClick={handleStartTask} disabled={!canStart}>
            {packingCopy("actions.start")}
          </button>
          <button className="erp-button erp-button--primary" type="button" onClick={handleConfirmTask} disabled={!canConfirm}>
            {packingCopy("actions.confirmPack")}
          </button>
        </div>
      </header>

      <section className="erp-packing-toolbar" aria-label={packingCopy("filters.label")}>
        <label className="erp-field">
          <span>{packingCopy("filters.warehouse")}</span>
          <select className="erp-input" value={warehouseId} onChange={(event) => setWarehouseId(event.target.value)}>
            {packTaskWarehouseOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {packingWarehouseLabel(option.value, option.label)}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>{packingCopy("filters.status")}</span>
          <select className="erp-input" value={status} onChange={(event) => setStatus(event.target.value as "" | PackTaskStatus)}>
            {packTaskStatusOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.value ? packTaskStatusLabel(option.value) : packingCopy("filters.allStatuses")}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>{packingCopy("filters.assignedUser")}</span>
          <input
            className="erp-input"
            value={assignedTo}
            onChange={(event) => setAssignedTo(event.target.value)}
            placeholder={packingCopy("filters.assignedPlaceholder")}
          />
        </label>
      </section>

      <section className="erp-kpi-grid erp-packing-kpis">
        <PackingKPI label={packingCopy("kpi.openTasks")} value={totals.open} tone="info" />
        <PackingKPI label={packingCopy("kpi.pendingLines")} value={totals.pending} tone={totals.pending === 0 ? "success" : "warning"} />
        <PackingKPI label={packingCopy("kpi.packedLines")} value={totals.packed} tone="success" />
        <PackingKPI label={packingCopy("kpi.exceptions")} value={totals.exceptions} tone={totals.exceptions === 0 ? "normal" : "danger"} />
      </section>

      <section className="erp-packing-workspace">
        <div className="erp-card erp-card--padded erp-packing-station-card">
          <div className="erp-section-header">
            <div>
              <h2 className="erp-section-title">{packingCopy("sections.packingQueue")}</h2>
              <p className="erp-section-description">{selectedTask?.packTaskNo ?? packingCopy("empty.noTaskSelected")}</p>
            </div>
            {selectedTask ? <StatusChip tone={packTaskStatusTone(selectedTask.status)}>{packTaskStatusLabel(selectedTask.status)}</StatusChip> : null}
          </div>

          <div className="erp-packing-meta-grid">
            <PackingFact label={packingCopy("facts.order")} value={selectedTask?.orderNo ?? "-"} />
            <PackingFact label={packingCopy("facts.pickTask")} value={selectedTask?.pickTaskNo ?? "-"} />
            <PackingFact label={packingCopy("facts.warehouse")} value={selectedTask?.warehouseCode ?? "-"} />
            <PackingFact label={packingCopy("facts.nextSku")} value={pendingLine?.skuCode ?? "-"} />
            <PackingFact label={packingCopy("facts.batch")} value={pendingLine?.batchNo ?? "-"} />
            <PackingFact
              label={packingCopy("facts.qty")}
              value={pendingLine ? `${pendingLine.qtyToPack} ${pendingLine.baseUOMCode}` : "-"}
              tone={pendingLine ? "info" : "success"}
            />
          </div>

          <label className={`erp-packing-scan-primary erp-packing-scan-primary--${feedback.tone}`}>
            <span>{packingCopy("scan.label")}</span>
            <input
              ref={scanInputRef}
              value={scanValue}
              onChange={(event) => setScanValue(event.target.value)}
              onKeyDown={handleScanKeyDown}
              placeholder={
                pendingLine
                  ? `${pendingLine.skuCode} / ${pendingLine.batchNo} / ${pendingLine.id}`
                  : packingCopy("scan.placeholderDone")
              }
              disabled={!selectedTask || busy || !pendingLine || selectedTask.status === "packed" || exceptionOpen}
            />
            <small>
              <strong>{feedback.title}</strong>
              {feedback.detail ? <span>{feedback.detail}</span> : null}
            </small>
          </label>

          <label className="erp-field erp-packing-note">
            <span>{packingCopy("note.label")}</span>
            <textarea
              className="erp-input"
              value={packingNote}
              onChange={(event) => setPackingNote(event.target.value)}
              placeholder={packingCopy("note.placeholder")}
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
                {packingExceptionLabel(option.value)}
              </button>
            ))}
          </div>
        </div>

        <div className="erp-card erp-card--padded erp-packing-queue-card">
          <div className="erp-section-header">
            <h2 className="erp-section-title">{packingCopy("sections.packTasks")}</h2>
            <StatusChip tone={displayedTasks.length === 0 ? "warning" : "info"}>
              {packingCopy("rows", { count: displayedTasks.length })}
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
            <h2 className="erp-section-title">{packingCopy("sections.packLines")}</h2>
            <p className="erp-section-description">{packingCopy("sections.packLinesDescription")}</p>
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
              <StatusChip tone="warning">{packingCopy("empty.noTaskSelected")}</StatusChip>
              <strong>{packingCopy("empty.selectTask")}</strong>
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

function packingCopy(key: string, values?: Record<string, string | number>) {
  return t(`shipping.packing.${key}`, { values });
}

function packingWarehouseLabel(value: string, fallback: string) {
  if (value === "") {
    return packingCopy("warehouse.all");
  }

  return t(`shipping.packing.warehouse.${value}`, { fallback });
}

function packingExceptionLabel(value: PackTaskExceptionCode) {
  return packingCopy(`exceptions.${value}`);
}

function packTaskStatusLabel(status: PackTaskStatus) {
  return packingCopy(`status.${status}`);
}

function packTaskLineStatusLabel(status: PackTaskLine["status"]) {
  return packingCopy(`status.${status}`);
}
