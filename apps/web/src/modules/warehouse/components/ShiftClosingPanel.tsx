"use client";

import { useMemo, useState } from "react";
import { DataTable, StatusChip, type DataTableColumn, type ToastMessage } from "@/shared/design-system/components";
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
    header: "Batch",
    render: (row) => row.batchNo,
    width: "130px"
  },
  {
    key: "bin",
    header: "Bin",
    render: (row) => row.binCode,
    width: "90px"
  },
  {
    key: "system",
    header: "System",
    render: (row) => row.systemQuantity,
    align: "right",
    width: "90px"
  },
  {
    key: "counted",
    header: "Counted",
    render: (row) => row.countedQuantity,
    align: "right",
    width: "90px"
  },
  {
    key: "variance",
    header: "Variance",
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
    header: "Owner",
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
        title: "Shift closed",
        description: closed.auditLogId ? `Audit log ${closed.auditLogId}` : undefined,
        tone: "success"
      });
    } catch (error) {
      setFeedback({
        id: "close-error",
        title: error instanceof Error ? error.message : "Shift close failed",
        tone: "danger"
      });
    }
  }

  if (loading) {
    return (
      <section className="erp-card erp-card--padded erp-shift-closing-panel">
        <div className="erp-section-header">
          <h2 className="erp-section-title">End-of-day closing</h2>
          <StatusChip tone="info">Loading</StatusChip>
        </div>
      </section>
    );
  }

  if (!activeReconciliation) {
    return (
      <section className="erp-card erp-card--padded erp-shift-closing-panel">
        <div className="erp-section-header">
          <h2 className="erp-section-title">End-of-day closing</h2>
          <StatusChip tone="warning">No session</StatusChip>
        </div>
      </section>
    );
  }

  return (
    <section className="erp-card erp-card--padded erp-shift-closing-panel">
      <div className="erp-section-header">
        <div>
          <h2 className="erp-section-title">End-of-day closing</h2>
          <p className="erp-section-description">
            {activeReconciliation.warehouseCode} / {activeReconciliation.date} / {activeReconciliation.shiftCode}
          </p>
        </div>
        <StatusChip tone={reconciliationStatusTone(activeReconciliation.status)}>
          {statusLabel(activeReconciliation.status)}
        </StatusChip>
      </div>

      <section className="erp-shift-closing-summary" aria-label="End-of-day reconciliation summary">
        <ClosingMetric label="System" value={activeReconciliation.summary.systemQuantity} tone="normal" />
        <ClosingMetric label="Counted" value={activeReconciliation.summary.countedQuantity} tone="info" />
        <ClosingMetric
          label="Variance"
          value={activeReconciliation.summary.varianceQuantity}
          tone={activeReconciliation.summary.varianceQuantity === 0 ? "success" : "danger"}
        />
        <ClosingMetric
          label="Checklist"
          value={`${activeReconciliation.summary.checklistCompleted}/${activeReconciliation.summary.checklistTotal}`}
          tone={openBlockingItems.length === 0 ? "success" : "warning"}
        />
      </section>

      <section className="erp-shift-operations" aria-label="Shift closing operating counters">
        <OperationMetric label="Orders" value={activeReconciliation.operations.orderCount} />
        <OperationMetric label="Handover" value={activeReconciliation.operations.handoverOrderCount} />
        <OperationMetric label="Returns" value={activeReconciliation.operations.returnOrderCount} />
        <OperationMetric label="Movements" value={activeReconciliation.operations.stockMovementCount} />
        <OperationMetric label="Stock count" value={activeReconciliation.operations.stockCountSessionCount} />
        <OperationMetric
          label="Issues"
          value={activeReconciliation.operations.pendingIssueCount}
          tone={activeReconciliation.operations.pendingIssueCount > 0 ? "danger" : "success"}
        />
      </section>

      <section className="erp-shift-closing-grid">
        <div>
          <div className="erp-section-header erp-section-header--compact">
            <h3 className="erp-subsection-title">Checklist</h3>
            <StatusChip tone={openBlockingItems.length === 0 ? "success" : "warning"}>
              {openBlockingItems.length} blockers
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
                  <strong>{item.label}</strong>
                  {item.note ? <small>{item.note}</small> : null}
                </span>
                <StatusChip tone={item.complete ? "success" : isOperationalClosingBlocker(item) ? "danger" : "warning"}>
                  {item.complete ? "Done" : isOperationalClosingBlocker(item) ? "Resolve" : "Exception"}
                </StatusChip>
              </li>
            ))}
          </ol>
        </div>

        <div className="erp-shift-close-box">
          <label className="erp-field">
            <span>Exception note</span>
            <textarea
              className="erp-input erp-textarea"
              value={exceptionNote}
              onChange={(event) => setExceptionNote(event.target.value)}
              placeholder={exceptionBlockingItems.length > 0 ? "Required for variance exception" : "No exception needed"}
            />
          </label>
          <button className="erp-button erp-button--primary erp-button--full" type="button" disabled={closeDisabled} onClick={handleClose}>
            Close shift
          </button>
          {operationalBlockingItems.length > 0 ? (
            <small className="erp-shift-close-feedback erp-shift-close-feedback--danger">
              Resolve operational blockers before closing.
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
          <h3 className="erp-subsection-title">System vs counted quantity</h3>
          <StatusChip tone={activeReconciliation.summary.varianceCount === 0 ? "success" : "danger"}>
            {activeReconciliation.summary.varianceCount} variance lines
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
      <StatusChip tone={tone}>Ops</StatusChip>
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
      return "Closed";
    case "in_review":
      return "In Review";
    case "open":
    default:
      return "Open";
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
