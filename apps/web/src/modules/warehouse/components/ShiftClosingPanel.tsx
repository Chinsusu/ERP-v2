"use client";

import { useMemo, useState } from "react";
import { DataTable, StatusChip, type DataTableColumn, type ToastMessage } from "@/shared/design-system/components";
import { t } from "@/shared/i18n";
import { useEndOfDayReconciliation } from "../hooks/useEndOfDayReconciliation";
import {
  closeEndOfDayReconciliation,
  isOperationalClosingBlocker,
  reconciliationStatusTone
} from "../services/warehouseDailyBoardService";
import type {
  EndOfDayReconciliation,
  EndOfDayReconciliationQuery,
  ReconciliationLine
} from "../types";

const reconciliationColumns: DataTableColumn<ReconciliationLine>[] = [
  {
    key: "sku",
    header: "SKU",
    render: (row) => row.sku,
    width: "140px"
  },
  {
    key: "batch",
    header: closingCopy("columns.batch"),
    render: (row) => row.batchNo,
    width: "130px"
  },
  {
    key: "bin",
    header: closingCopy("columns.bin"),
    render: (row) => row.binCode,
    width: "90px"
  },
  {
    key: "system",
    header: closingCopy("columns.system"),
    render: (row) => row.systemQuantity,
    align: "right",
    width: "90px"
  },
  {
    key: "counted",
    header: closingCopy("columns.counted"),
    render: (row) => row.countedQuantity,
    align: "right",
    width: "90px"
  },
  {
    key: "variance",
    header: closingCopy("columns.variance"),
    render: (row) => (
      <StatusChip tone={row.varianceQuantity === 0 ? "success" : "danger"}>
        {formatVariance(row.varianceQuantity)}
      </StatusChip>
    ),
    align: "right",
    width: "110px"
  },
  {
    key: "owner",
    header: closingCopy("columns.owner"),
    render: (row) => row.owner,
    width: "140px"
  }
];

export type ShiftClosingPanelProps = {
  query: EndOfDayReconciliationQuery;
};

export function ShiftClosingPanel({ query }: ShiftClosingPanelProps) {
  const [exceptionNote, setExceptionNote] = useState("");
  const [closeResult, setCloseResult] = useState<EndOfDayReconciliation | null>(null);
  const [feedback, setFeedback] = useState<ToastMessage | null>(null);
  const { reconciliations, loading } = useEndOfDayReconciliation(query);
  const activeReconciliation = useMemo(() => {
    const first = reconciliations[0] ?? null;
    if (!closeResult || closeResult.id !== first?.id) {
      return first;
    }

    return closeResult;
  }, [closeResult, reconciliations]);
  const openBlockingItems =
    activeReconciliation?.checklist.filter((item) => item.blocking && !item.complete) ?? [];
  const operationalBlockingItems = openBlockingItems.filter(isOperationalClosingBlocker);
  const exceptionBlockingItems = openBlockingItems.filter((item) => !isOperationalClosingBlocker(item));
  const closeDisabled =
    !activeReconciliation ||
    activeReconciliation.status === "closed" ||
    operationalBlockingItems.length > 0 ||
    (exceptionBlockingItems.length > 0 && exceptionNote.trim() === "");

  async function handleClose() {
    if (!activeReconciliation || closeDisabled) {
      return;
    }

    try {
      const closed = await closeEndOfDayReconciliation(activeReconciliation.id, exceptionNote);
      setCloseResult(closed);
      setFeedback({
        id: closed.auditLogId ?? closed.id,
        title: closingCopy("shiftClosingPanel.shiftClosed"),
        description: closed.auditLogId ? closingCopy("shiftClosingPanel.auditLog", { id: closed.auditLogId }) : undefined,
        tone: "success"
      });
    } catch (error) {
      setFeedback({
        id: "close-error",
        title: error instanceof Error ? error.message : closingCopy("shiftClosingPanel.closeFailed"),
        tone: "danger"
      });
    }
  }

  if (loading) {
    return (
      <section className="erp-card erp-card--padded erp-shift-closing-panel">
        <div className="erp-section-header">
          <h2 className="erp-section-title">{closingCopy("shiftClosingPanel.title")}</h2>
          <StatusChip tone="info">{closingCopy("status.loading")}</StatusChip>
        </div>
      </section>
    );
  }

  if (!activeReconciliation) {
    return (
      <section className="erp-card erp-card--padded erp-shift-closing-panel">
        <div className="erp-section-header">
          <h2 className="erp-section-title">{closingCopy("shiftClosingPanel.title")}</h2>
          <StatusChip tone="warning">{closingCopy("shiftClosingPanel.noSession")}</StatusChip>
        </div>
      </section>
    );
  }

  return (
    <section className="erp-card erp-card--padded erp-shift-closing-panel">
      <div className="erp-section-header">
        <div>
          <h2 className="erp-section-title">{closingCopy("shiftClosingPanel.title")}</h2>
          <p className="erp-section-description">
            {activeReconciliation.warehouseCode} / {activeReconciliation.date} / {activeReconciliation.shiftCode}
          </p>
        </div>
        <StatusChip tone={reconciliationStatusTone(activeReconciliation.status)}>
          {statusLabel(activeReconciliation.status)}
        </StatusChip>
      </div>

      <section className="erp-shift-closing-summary" aria-label={closingCopy("shiftClosingPanel.summaryLabel")}>
        <ClosingMetric label={closingCopy("shiftClosingPanel.system")} value={activeReconciliation.summary.systemQuantity} tone="normal" />
        <ClosingMetric label={closingCopy("shiftClosingPanel.counted")} value={activeReconciliation.summary.countedQuantity} tone="info" />
        <ClosingMetric
          label={closingCopy("shiftClosingPanel.variance")}
          value={activeReconciliation.summary.varianceQuantity}
          tone={activeReconciliation.summary.varianceQuantity === 0 ? "success" : "danger"}
        />
        <ClosingMetric
          label={closingCopy("shiftClosingPanel.checklist")}
          value={`${activeReconciliation.summary.checklistCompleted}/${activeReconciliation.summary.checklistTotal}`}
          tone={openBlockingItems.length === 0 ? "success" : "warning"}
        />
      </section>

      <section className="erp-shift-operations" aria-label={closingCopy("shiftClosingPanel.operationsLabel")}>
        <OperationMetric label={closingCopy("shiftClosingPanel.orders")} value={activeReconciliation.operations.orderCount} />
        <OperationMetric label={closingCopy("shiftClosingPanel.handover")} value={activeReconciliation.operations.handoverOrderCount} />
        <OperationMetric label={closingCopy("shiftClosingPanel.returns")} value={activeReconciliation.operations.returnOrderCount} />
        <OperationMetric label={closingCopy("shiftClosingPanel.movements")} value={activeReconciliation.operations.stockMovementCount} />
        <OperationMetric label={closingCopy("shiftClosingPanel.stockCount")} value={activeReconciliation.operations.stockCountSessionCount} />
        <OperationMetric
          label={closingCopy("shiftClosingPanel.issues")}
          value={activeReconciliation.operations.pendingIssueCount}
          tone={activeReconciliation.operations.pendingIssueCount > 0 ? "danger" : "success"}
        />
      </section>

      <section className="erp-shift-closing-grid">
        <div>
          <div className="erp-section-header erp-section-header--compact">
            <h3 className="erp-subsection-title">{closingCopy("shiftClosingPanel.checklist")}</h3>
            <StatusChip tone={openBlockingItems.length === 0 ? "success" : "warning"}>
              {closingCopy("shiftClosingPanel.blockers", { count: openBlockingItems.length })}
            </StatusChip>
          </div>
          <ol className="erp-shift-checklist">
            {activeReconciliation.checklist.map((item) => (
              <li
                className={`erp-shift-checklist-item${item.blocking && !item.complete ? " is-blocking" : ""}`}
                key={item.key}
              >
                <span className={item.complete ? "erp-shift-checkmark is-complete" : "erp-shift-checkmark"} aria-hidden="true" />
                <span>
                  <strong>{checklistLabel(item.key, item.label)}</strong>
                  {item.note ? <small>{checklistNote(item.key, item.note)}</small> : null}
                </span>
                <StatusChip tone={item.complete ? "success" : isOperationalClosingBlocker(item) ? "danger" : "warning"}>
                  {item.complete
                    ? closingCopy("status.done")
                    : isOperationalClosingBlocker(item)
                      ? closingCopy("status.resolve")
                      : closingCopy("status.exception")}
                </StatusChip>
              </li>
            ))}
          </ol>
        </div>

        <div className="erp-shift-close-box">
          <label className="erp-field">
            <span>{closingCopy("shiftClosingPanel.exceptionNote")}</span>
            <textarea
              className="erp-input erp-textarea"
              value={exceptionNote}
              onChange={(event) => setExceptionNote(event.target.value)}
              placeholder={
                exceptionBlockingItems.length > 0
                  ? closingCopy("shiftClosingPanel.requiredVarianceException")
                  : closingCopy("shiftClosingPanel.noExceptionNeeded")
              }
            />
          </label>
          <button className="erp-button erp-button--primary erp-button--full" type="button" disabled={closeDisabled} onClick={handleClose}>
            {closingCopy("shiftClosingPanel.closeShift")}
          </button>
          {operationalBlockingItems.length > 0 ? (
            <small className="erp-shift-close-feedback erp-shift-close-feedback--danger">
              {closingCopy("shiftClosingPanel.resolveOperationalBlockers")}
            </small>
          ) : null}
          {feedback ? (
            <small className={`erp-shift-close-feedback erp-shift-close-feedback--${feedback.tone ?? "normal"}`}>
              {feedback.title}
            </small>
          ) : null}
        </div>
      </section>

      <div className="erp-shift-lines">
        <div className="erp-section-header erp-section-header--compact">
          <h3 className="erp-subsection-title">{closingCopy("shiftClosingPanel.systemVsCounted")}</h3>
          <StatusChip tone={activeReconciliation.summary.varianceCount === 0 ? "success" : "danger"}>
            {closingCopy("shiftClosingPanel.varianceLines", { count: activeReconciliation.summary.varianceCount })}
          </StatusChip>
        </div>
        <DataTable
          columns={reconciliationColumns}
          rows={activeReconciliation.lines}
          getRowKey={(row) => row.id}
        />
      </div>
    </section>
  );
}

function OperationMetric({
  label,
  value,
  tone = "info"
}: {
  label: string;
  value: number;
  tone?: "success" | "danger" | "info";
}) {
  return (
    <article className="erp-shift-operation">
      <span>{label}</span>
      <strong>{formatCount(value)}</strong>
      <StatusChip tone={tone}>{closingCopy("shiftClosingPanel.ops")}</StatusChip>
    </article>
  );
}

function ClosingMetric({
  label,
  value,
  tone
}: {
  label: string;
  value: number | string;
  tone: "normal" | "success" | "warning" | "danger" | "info";
}) {
  return (
    <article className="erp-shift-metric">
      <span>{label}</span>
      <strong>{typeof value === "number" ? formatCount(value) : value}</strong>
      <StatusChip tone={tone}>{label}</StatusChip>
    </article>
  );
}

function statusLabel(status: EndOfDayReconciliation["status"]) {
  switch (status) {
    case "closed":
      return closingCopy("status.closed");
    case "in_review":
      return closingCopy("status.inReview");
    case "open":
    default:
      return closingCopy("status.open");
  }
}

function formatVariance(value: number) {
  if (value > 0) {
    return `+${value}`;
  }

  return String(value);
}

function formatCount(value: number) {
  return value.toLocaleString("vi-VN");
}

function closingCopy(key: string, values?: Record<string, string | number>) {
  return t(`warehouse.${key}`, { values });
}

function checklistLabel(key: string, fallback: string) {
  return t(`warehouse.shiftClosingPanel.checklistItems.${key}`, { fallback });
}

function checklistNote(key: string, fallback: string) {
  return t(`warehouse.shiftClosingPanel.checklistNotes.${key}`, { fallback });
}
