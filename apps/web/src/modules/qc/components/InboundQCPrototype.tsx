"use client";

import { useEffect, useMemo, useState } from "react";
import { DataTable, EmptyState, StatusChip, type DataTableColumn, type StatusTone } from "@/shared/design-system/components";
import { useGoodsReceipts } from "../../receiving/hooks/useGoodsReceipts";
import {
  formatReceivingDateTime,
  formatReceivingQuantity,
  receivingWarehouseOptions
} from "../../receiving/services/warehouseReceivingService";
import type { GoodsReceipt, GoodsReceiptLine, GoodsReceiptQuery, ReceivingPackagingStatus } from "../../receiving/types";
import { useInboundQCInspections } from "../hooks/useInboundQCInspections";
import {
  checklistStatusTone,
  createInboundQCInspection,
  defaultInboundQCChecklist,
  failInboundQCInspection,
  formatChecklistStatus,
  formatInboundQCDateTime,
  formatInboundQCQuantity,
  formatInboundQCResult,
  formatInboundQCStatus,
  holdInboundQCInspection,
  inboundQCChecklistStatusOptions,
  inboundQCResultTone,
  inboundQCStatusOptions,
  inboundQCStatusTone,
  partialInboundQCInspection,
  passInboundQCInspection,
  startInboundQCInspection
} from "../services/inboundQCService";
import type {
  InboundQCChecklistItem,
  InboundQCChecklistStatus,
  InboundQCInspection,
  InboundQCInspectionQuery,
  InboundQCInspectionStatus,
  InboundQCResult
} from "../types";

type StatusFilter = "" | InboundQCInspectionStatus;

type InspectableReceivingLine = {
  key: string;
  receipt: GoodsReceipt;
  line: GoodsReceiptLine;
};

const zeroQuantity = "0.000000";

export function InboundQCPrototype() {
  const [warehouseId, setWarehouseId] = useState("wh-hcm-fg");
  const [status, setStatus] = useState<StatusFilter>("");
  const [inspectorId, setInspectorId] = useState("user-qa");
  const [selectedLineKey, setSelectedLineKey] = useState("");
  const [selectedInspectionId, setSelectedInspectionId] = useState("");
  const [localInspections, setLocalInspections] = useState<InboundQCInspection[]>([]);
  const [checklistDraft, setChecklistDraft] = useState<InboundQCChecklistItem[]>(cloneChecklist(defaultInboundQCChecklist));
  const [passedQuantity, setPassedQuantity] = useState("");
  const [failedQuantity, setFailedQuantity] = useState("0");
  const [holdQuantity, setHoldQuantity] = useState("0");
  const [reason, setReason] = useState("");
  const [decisionNote, setDecisionNote] = useState("");
  const [evidenceRefs, setEvidenceRefs] = useState("");
  const [feedback, setFeedback] = useState<{ tone: StatusTone; message: string } | null>(null);
  const [busyAction, setBusyAction] = useState("");
  const receivingQuery = useMemo<GoodsReceiptQuery>(
    () => ({
      warehouseId,
      status: "inspect_ready"
    }),
    [warehouseId]
  );
  const inspectionQuery = useMemo<InboundQCInspectionQuery>(
    () => ({
      warehouseId,
      status: status || undefined
    }),
    [status, warehouseId]
  );
  const { receipts, loading: receiptsLoading, error: receiptsError } = useGoodsReceipts(receivingQuery);
  const { inspections, loading: inspectionsLoading, error: inspectionsError } = useInboundQCInspections(inspectionQuery);
  const inspectableLines = useMemo(() => flattenInspectableLines(receipts), [receipts]);
  const visibleInspections = useMemo(
    () => mergeInspections(localInspections, inspections, inspectionQuery),
    [inspectionQuery, inspections, localInspections]
  );
  const selectedLine = inspectableLines.find((item) => item.key === selectedLineKey) ?? inspectableLines[0] ?? null;
  const existingLineInspection = selectedLine ? findInspectionForLine(visibleInspections, selectedLine) : null;
  const selectedInspection =
    visibleInspections.find((inspection) => inspection.id === selectedInspectionId) ??
    existingLineInspection ??
    visibleInspections[0] ??
    null;
  const totals = summarizeInspections(visibleInspections);
  const selectedLineColumns = useMemo<DataTableColumn<InspectableReceivingLine>[]>(
    () => [
      {
        key: "receipt",
        header: "Receipt",
        render: (row) => (
          <span className="erp-qc-record-cell">
            <strong>{row.receipt.receiptNo}</strong>
            <small>{row.receipt.deliveryNoteNo ?? row.receipt.referenceDocId}</small>
          </span>
        ),
        width: "190px"
      },
      {
        key: "sku",
        header: "SKU",
        render: (row) => (
          <span className="erp-qc-record-cell">
            <strong>{row.line.sku}</strong>
            <small>{row.line.itemName ?? row.line.itemId}</small>
          </span>
        )
      },
      {
        key: "lot",
        header: "Lot / Expiry",
        render: (row) => (
          <span className="erp-qc-record-cell">
            <strong>{row.line.lotNo ?? row.line.batchNo ?? "-"}</strong>
            <small>{row.line.expiryDate ?? "-"}</small>
          </span>
        ),
        width: "150px"
      },
      {
        key: "packaging",
        header: "Packaging",
        render: (row) => (
          <StatusChip tone={row.line.packagingStatus === "intact" ? "success" : "warning"}>
            {formatPackagingStatus(row.line.packagingStatus)}
          </StatusChip>
        ),
        width: "150px"
      },
      {
        key: "quantity",
        header: "Quantity",
        render: (row) => formatReceivingQuantity(row.line.quantity, row.line.baseUomCode),
        align: "right",
        width: "130px"
      },
      {
        key: "action",
        header: "Action",
        render: (row) => {
          const inspection = findInspectionForLine(visibleInspections, row);

          return (
            <button className="erp-button erp-button--secondary" type="button" onClick={() => handleSelectLine(row)}>
              {inspection ? "Open" : row.key === selectedLine?.key ? "Selected" : "Select"}
            </button>
          );
        },
        width: "100px",
        sticky: true
      }
    ],
    [selectedLine?.key, visibleInspections]
  );
  const inspectionColumns = useMemo<DataTableColumn<InboundQCInspection>[]>(
    () => [
      {
        key: "inspection",
        header: "Inspection",
        render: (row) => (
          <span className="erp-qc-record-cell">
            <strong>{row.id}</strong>
            <small>{row.goodsReceiptNo}</small>
          </span>
        ),
        width: "240px"
      },
      {
        key: "sku",
        header: "SKU / Lot",
        render: (row) => (
          <span className="erp-qc-record-cell">
            <strong>{row.sku}</strong>
            <small>{row.lotNo}</small>
          </span>
        )
      },
      {
        key: "status",
        header: "Status",
        render: (row) => <StatusChip tone={inboundQCStatusTone(row.status)}>{formatInboundQCStatus(row.status)}</StatusChip>,
        width: "140px"
      },
      {
        key: "result",
        header: "Result",
        render: (row) => <StatusChip tone={inboundQCResultTone(row.result)}>{formatInboundQCResult(row.result)}</StatusChip>,
        width: "120px"
      },
      {
        key: "qty",
        header: "Quantity",
        render: (row) => formatInboundQCQuantity(row.quantity, row.uomCode),
        align: "right",
        width: "130px"
      },
      {
        key: "updated",
        header: "Updated",
        render: (row) => formatInboundQCDateTime(row.updatedAt),
        width: "150px"
      },
      {
        key: "open",
        header: "Action",
        render: (row) => (
          <button className="erp-button erp-button--secondary" type="button" onClick={() => setSelectedInspectionId(row.id)}>
            Open
          </button>
        ),
        width: "96px",
        sticky: true
      }
    ],
    []
  );

  useEffect(() => {
    if (inspectableLines.length === 0) {
      setSelectedLineKey("");
      return;
    }
    if (!inspectableLines.some((item) => item.key === selectedLineKey)) {
      setSelectedLineKey(inspectableLines[0].key);
    }
  }, [inspectableLines, selectedLineKey]);

  useEffect(() => {
    if (!selectedInspection) {
      setChecklistDraft(cloneChecklist(defaultInboundQCChecklist));
      setPassedQuantity("");
      setFailedQuantity("0");
      setHoldQuantity("0");
      setReason("");
      setDecisionNote("");
      return;
    }

    setChecklistDraft(cloneChecklist(selectedInspection.checklist));
    setPassedQuantity(selectedInspection.passedQuantity !== zeroQuantity ? selectedInspection.passedQuantity : selectedInspection.quantity);
    setFailedQuantity(selectedInspection.failedQuantity !== zeroQuantity ? selectedInspection.failedQuantity : "0");
    setHoldQuantity(selectedInspection.holdQuantity !== zeroQuantity ? selectedInspection.holdQuantity : "0");
    setReason(selectedInspection.reason ?? "");
    setDecisionNote(selectedInspection.note ?? "");
  }, [selectedInspection?.id, selectedInspection?.updatedAt]);

  function handleSelectLine(row: InspectableReceivingLine) {
    setSelectedLineKey(row.key);
    const inspection = findInspectionForLine(visibleInspections, row);
    setSelectedInspectionId(inspection?.id ?? "");
    setFeedback(null);
  }

  async function handleCreateInspection() {
    if (!selectedLine || busyAction) {
      return;
    }
    if (existingLineInspection) {
      setSelectedInspectionId(existingLineInspection.id);
      setFeedback({ tone: "warning", message: `${existingLineInspection.id} already covers this receiving line` });
      return;
    }

    setBusyAction("create");
    try {
      const result = await createInboundQCInspection({
        goodsReceiptId: selectedLine.receipt.id,
        goodsReceiptLineId: selectedLine.line.id,
        inspectorId,
        checklist: defaultInboundQCChecklist,
        note: createInspectionNote(selectedLine, evidenceRefs)
      });
      upsertInspection(result.inspection);
      setFeedback({ tone: "success", message: `${result.inspection.id} created` });
    } catch (cause) {
      setFeedback({
        tone: "danger",
        message: cause instanceof Error ? cause.message : "Inbound QC inspection could not be created"
      });
    } finally {
      setBusyAction("");
    }
  }

  async function handleStartInspection() {
    if (!selectedInspection || busyAction) {
      return;
    }

    setBusyAction("start");
    try {
      const result = await startInboundQCInspection(selectedInspection.id);
      upsertInspection(result.inspection);
      setFeedback({ tone: "success", message: `${result.inspection.id} started` });
    } catch (cause) {
      setFeedback({
        tone: "danger",
        message: cause instanceof Error ? cause.message : "Inbound QC inspection could not be started"
      });
    } finally {
      setBusyAction("");
    }
  }

  async function handleDecision(result: InboundQCResult) {
    if (!selectedInspection || busyAction) {
      return;
    }
    if (selectedInspection.status !== "in_progress") {
      setFeedback({ tone: "warning", message: "Start the inspection before recording a QC result" });
      return;
    }
    if (result !== "pass" && reason.trim() === "") {
      setFeedback({ tone: "warning", message: "Reason is required for fail, hold, and partial results" });
      return;
    }

    const checklist = checklistForDecision(checklistDraft);
    if (result === "pass" && checklist.some((item) => item.required && item.status === "fail")) {
      setFeedback({ tone: "warning", message: "A failed required checklist item cannot be passed" });
      return;
    }

    const decisionInput = {
      passedQuantity: result === "pass" ? selectedInspection.quantity : result === "partial" ? passedQuantity : "0",
      failedQuantity: result === "fail" ? selectedInspection.quantity : result === "partial" ? failedQuantity : "0",
      holdQuantity: result === "hold" ? selectedInspection.quantity : result === "partial" ? holdQuantity : "0",
      checklist,
      reason: result === "pass" ? undefined : reason,
      note: decisionNoteWithEvidence(decisionNote, evidenceRefs)
    };

    setBusyAction(result);
    try {
      const actionResult = await runDecisionAction(selectedInspection.id, result, decisionInput);
      upsertInspection(actionResult.inspection);
      setFeedback({
        tone: inboundQCResultTone(actionResult.inspection.result),
        message: `${actionResult.inspection.id} / ${formatInboundQCResult(actionResult.inspection.result)}`
      });
    } catch (cause) {
      setFeedback({
        tone: "danger",
        message: cause instanceof Error ? cause.message : "Inbound QC decision could not be recorded"
      });
    } finally {
      setBusyAction("");
    }
  }

  function updateChecklistItem(id: string, patch: Partial<InboundQCChecklistItem>) {
    setChecklistDraft((current) => current.map((item) => (item.id === id ? { ...item, ...patch } : item)));
  }

  function upsertInspection(inspection: InboundQCInspection) {
    setLocalInspections((current) => [inspection, ...current.filter((candidate) => candidate.id !== inspection.id)]);
    setSelectedInspectionId(inspection.id);
  }

  return (
    <section className="erp-module-page erp-qc-page">
      <header className="erp-page-header">
        <div>
          <p className="erp-module-eyebrow">QC</p>
          <h1 className="erp-page-title">Inbound QC</h1>
          <p className="erp-page-description">Receiving inspection for quantity, packaging, lot, expiry, COA/MSDS, and evidence</p>
        </div>
        <div className="erp-page-actions">
          <a className="erp-button erp-button--secondary" href="#qc-receiving-lines">
            Receipts
          </a>
          <a className="erp-button erp-button--secondary" href="#qc-decision">
            Decision
          </a>
          <a className="erp-button erp-button--primary" href="#qc-inspections">
            Inspections
          </a>
        </div>
      </header>

      <section className="erp-qc-toolbar" aria-label="Inbound QC filters">
        <label className="erp-field">
          <span>Warehouse</span>
          <select className="erp-input" value={warehouseId} onChange={(event) => setWarehouseId(event.target.value)}>
            {receivingWarehouseOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Inspection status</span>
          <select className="erp-input" value={status} onChange={(event) => setStatus(event.target.value as StatusFilter)}>
            {inboundQCStatusOptions.map((option) => (
              <option key={option.value || "all"} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Inspector</span>
          <input className="erp-input" type="text" value={inspectorId} onChange={(event) => setInspectorId(event.target.value)} />
        </label>
      </section>

      <section className="erp-kpi-grid erp-qc-kpis">
        <QCKPI label="Inspectable lines" value={inspectableLines.length} tone={inspectableLines.length > 0 ? "info" : "normal"} />
        <QCKPI label="Pending" value={totals.pending} tone={totals.pending > 0 ? "warning" : "normal"} />
        <QCKPI label="In progress" value={totals.inProgress} tone={totals.inProgress > 0 ? "info" : "normal"} />
        <QCKPI label="Completed" value={totals.completed} tone="success" />
        <QCKPI label="Hold / Fail" value={totals.holdOrFail} tone={totals.holdOrFail > 0 ? "danger" : "normal"} />
      </section>

      <section className="erp-qc-workspace">
        <div className="erp-card erp-card--padded erp-qc-panel" id="qc-receiving-lines">
          <div className="erp-section-header">
            <h2 className="erp-section-title">Inspectable receiving lines</h2>
            <StatusChip tone={receiptsLoading ? "warning" : "info"}>{receiptsLoading ? "Loading" : `${inspectableLines.length} lines`}</StatusChip>
          </div>

          <DataTable
            columns={selectedLineColumns}
            rows={inspectableLines}
            getRowKey={(row) => row.key}
            loading={receiptsLoading}
            error={receiptsError?.message}
            emptyState={
              <EmptyState
                title="No inspection-ready receiving lines"
                description="Submit a goods receipt and mark it inspection-ready before QC can start."
              />
            }
          />

          {selectedLine ? (
            <div className="erp-qc-selected-line" aria-label="Selected receiving line">
              <StatusChip tone={existingLineInspection ? inboundQCStatusTone(existingLineInspection.status) : "info"}>
                {existingLineInspection ? formatInboundQCStatus(existingLineInspection.status) : "Ready for QC"}
              </StatusChip>
              <div>
                <strong>{selectedLine.receipt.receiptNo}</strong>
                <span>
                  {selectedLine.line.sku} / {selectedLine.line.lotNo ?? selectedLine.line.batchNo ?? "-"} /{" "}
                  {formatReceivingQuantity(selectedLine.line.quantity, selectedLine.line.baseUomCode)}
                </span>
              </div>
              <button className="erp-button erp-button--primary" type="button" disabled={busyAction !== ""} onClick={handleCreateInspection}>
                {existingLineInspection ? "Open inspection" : busyAction === "create" ? "Creating" : "Create inspection"}
              </button>
            </div>
          ) : null}
        </div>

        <div className="erp-card erp-card--padded erp-qc-panel" id="qc-decision">
          <div className="erp-section-header">
            <h2 className="erp-section-title">Inspection decision</h2>
            <StatusChip tone={selectedInspection ? inboundQCStatusTone(selectedInspection.status) : "normal"}>
              {selectedInspection ? formatInboundQCStatus(selectedInspection.status) : "No inspection"}
            </StatusChip>
          </div>

          {feedback ? <StatusChip tone={feedback.tone}>{feedback.message}</StatusChip> : null}

          {selectedInspection ? (
            <>
              <div className="erp-qc-fact-grid" aria-label="Inbound QC inspection facts">
                <QCFact label="Receipt" value={selectedInspection.goodsReceiptNo} />
                <QCFact label="PO" value={selectedInspection.purchaseOrderId ?? "-"} />
                <QCFact label="SKU" value={selectedInspection.sku} />
                <QCFact label="Lot / expiry" value={`${selectedInspection.lotNo} / ${selectedInspection.expiryDate}`} />
                <QCFact label="Quantity" value={formatInboundQCQuantity(selectedInspection.quantity, selectedInspection.uomCode)} />
                <QCFact label="Inspector" value={selectedInspection.inspectorId} />
                <QCFact label="Started" value={formatInboundQCDateTime(selectedInspection.startedAt)} />
                <QCFact label="Decided" value={formatInboundQCDateTime(selectedInspection.decidedAt)} />
              </div>

              <div className="erp-qc-checklist" aria-label="QC checklist">
                {checklistDraft.map((item) => (
                  <div className="erp-qc-checklist-item" key={item.id}>
                    <div className="erp-qc-checklist-label">
                      <strong>{item.label}</strong>
                      <small>
                        {item.code} / {item.required ? "Required" : "Optional"}
                      </small>
                    </div>
                    <select
                      className="erp-input"
                      value={item.status}
                      disabled={selectedInspection.status === "completed"}
                      onChange={(event) =>
                        updateChecklistItem(item.id, { status: event.currentTarget.value as InboundQCChecklistStatus })
                      }
                    >
                      {inboundQCChecklistStatusOptions.map((option) => (
                        <option key={option.value} value={option.value}>
                          {option.label}
                        </option>
                      ))}
                    </select>
                    <input
                      className="erp-input"
                      type="text"
                      value={item.note ?? ""}
                      placeholder="Checklist note"
                      disabled={selectedInspection.status === "completed"}
                      onChange={(event) => updateChecklistItem(item.id, { note: event.currentTarget.value })}
                    />
                    <StatusChip tone={checklistStatusTone(item.status)}>{formatChecklistStatus(item.status)}</StatusChip>
                  </div>
                ))}
              </div>

              <div className="erp-qc-decision-grid" aria-label="QC decision quantities">
                <label className="erp-field">
                  <span>Pass qty</span>
                  <input
                    className="erp-input"
                    inputMode="decimal"
                    type="text"
                    value={passedQuantity}
                    disabled={selectedInspection.status === "completed"}
                    onChange={(event) => setPassedQuantity(event.currentTarget.value)}
                  />
                </label>
                <label className="erp-field">
                  <span>Fail qty</span>
                  <input
                    className="erp-input"
                    inputMode="decimal"
                    type="text"
                    value={failedQuantity}
                    disabled={selectedInspection.status === "completed"}
                    onChange={(event) => setFailedQuantity(event.currentTarget.value)}
                  />
                </label>
                <label className="erp-field">
                  <span>Hold qty</span>
                  <input
                    className="erp-input"
                    inputMode="decimal"
                    type="text"
                    value={holdQuantity}
                    disabled={selectedInspection.status === "completed"}
                    onChange={(event) => setHoldQuantity(event.currentTarget.value)}
                  />
                </label>
              </div>

              <div className="erp-qc-note-grid">
                <label className="erp-field erp-qc-note-field">
                  <span>Reason</span>
                  <textarea
                    className="erp-input"
                    value={reason}
                    placeholder="Required for fail, hold, or partial"
                    disabled={selectedInspection.status === "completed"}
                    onChange={(event) => setReason(event.currentTarget.value)}
                  />
                </label>
                <label className="erp-field erp-qc-note-field">
                  <span>QC note</span>
                  <textarea
                    className="erp-input"
                    value={decisionNote}
                    disabled={selectedInspection.status === "completed"}
                    onChange={(event) => setDecisionNote(event.currentTarget.value)}
                  />
                </label>
                <label className="erp-field erp-qc-note-field">
                  <span>COA / MSDS / photo refs</span>
                  <textarea
                    className="erp-input"
                    value={evidenceRefs}
                    placeholder="COA-260429-001, MSDS-SERUM, IMG-..."
                    disabled={selectedInspection.status === "completed"}
                    onChange={(event) => setEvidenceRefs(event.currentTarget.value)}
                  />
                </label>
              </div>

              <div className="erp-qc-actions">
                <button
                  className="erp-button erp-button--secondary"
                  type="button"
                  disabled={busyAction !== "" || selectedInspection.status !== "pending"}
                  onClick={handleStartInspection}
                >
                  {busyAction === "start" ? "Starting" : "Start"}
                </button>
                <button
                  className="erp-button erp-button--primary"
                  type="button"
                  disabled={busyAction !== "" || selectedInspection.status !== "in_progress"}
                  onClick={() => void handleDecision("pass")}
                >
                  Pass
                </button>
                <button
                  className="erp-button erp-button--secondary"
                  type="button"
                  disabled={busyAction !== "" || selectedInspection.status !== "in_progress"}
                  onClick={() => void handleDecision("hold")}
                >
                  Hold
                </button>
                <button
                  className="erp-button erp-button--secondary"
                  type="button"
                  disabled={busyAction !== "" || selectedInspection.status !== "in_progress"}
                  onClick={() => void handleDecision("partial")}
                >
                  Partial
                </button>
                <button
                  className="erp-button erp-button--danger"
                  type="button"
                  disabled={busyAction !== "" || selectedInspection.status !== "in_progress"}
                  onClick={() => void handleDecision("fail")}
                >
                  Fail
                </button>
              </div>
            </>
          ) : (
            <EmptyState title="No inbound QC inspection selected" description="Create one from an inspection-ready receiving line." />
          )}
        </div>
      </section>

      <section className="erp-card erp-card--padded erp-module-table-card" id="qc-inspections">
        <div className="erp-section-header">
          <h2 className="erp-section-title">Inbound QC inspections</h2>
          <StatusChip tone={inspectionsLoading ? "warning" : "info"}>
            {inspectionsLoading ? "Loading" : `${visibleInspections.length} records`}
          </StatusChip>
        </div>
        <DataTable
          columns={inspectionColumns}
          rows={visibleInspections}
          getRowKey={(row) => row.id}
          loading={inspectionsLoading}
          error={inspectionsError?.message}
          emptyState={<EmptyState title="No inbound QC inspections" description="Create an inspection from a receiving line." />}
        />
      </section>
    </section>
  );
}

function flattenInspectableLines(receipts: GoodsReceipt[]): InspectableReceivingLine[] {
  return receipts.flatMap((receipt) =>
    receipt.lines.map((line) => ({
      key: receivingLineKey(receipt.id, line.id),
      receipt,
      line
    }))
  );
}

function receivingLineKey(receiptId: string, lineId: string) {
  return `${receiptId}:${lineId}`;
}

function findInspectionForLine(inspections: InboundQCInspection[], row: InspectableReceivingLine) {
  return inspections.find(
    (inspection) =>
      inspection.goodsReceiptId === row.receipt.id &&
      inspection.goodsReceiptLineId === row.line.id &&
      inspection.status !== "cancelled"
  );
}

function mergeInspections(
  localInspections: InboundQCInspection[],
  remoteInspections: InboundQCInspection[],
  query: InboundQCInspectionQuery
) {
  const localMatches = localInspections.filter((inspection) => matchesInspectionQuery(inspection, query));
  const localIds = new Set(localMatches.map((inspection) => inspection.id));

  return [...localMatches, ...remoteInspections.filter((inspection) => !localIds.has(inspection.id))].sort((left, right) =>
    right.updatedAt.localeCompare(left.updatedAt)
  );
}

function matchesInspectionQuery(inspection: InboundQCInspection, query: InboundQCInspectionQuery) {
  if (query.status && inspection.status !== query.status) {
    return false;
  }
  if (query.warehouseId && inspection.warehouseId !== query.warehouseId) {
    return false;
  }

  return true;
}

function summarizeInspections(inspections: InboundQCInspection[]) {
  return inspections.reduce(
    (acc, inspection) => ({
      pending: acc.pending + (inspection.status === "pending" ? 1 : 0),
      inProgress: acc.inProgress + (inspection.status === "in_progress" ? 1 : 0),
      completed: acc.completed + (inspection.status === "completed" ? 1 : 0),
      holdOrFail: acc.holdOrFail + (inspection.result === "hold" || inspection.result === "fail" ? 1 : 0)
    }),
    { pending: 0, inProgress: 0, completed: 0, holdOrFail: 0 }
  );
}

function checklistForDecision(items: InboundQCChecklistItem[]) {
  return items.map((item) => ({
    ...item,
    status: item.required && item.status === "pending" ? ("pass" as const) : item.status
  }));
}

function createInspectionNote(row: InspectableReceivingLine, evidenceRefs: string) {
  return decisionNoteWithEvidence(
    `Packaging: ${formatPackagingStatus(row.line.packagingStatus)}; lot: ${row.line.lotNo ?? row.line.batchNo ?? "-"}; received: ${formatReceivingDateTime(row.receipt.inspectReadyAt)}`,
    evidenceRefs
  );
}

function decisionNoteWithEvidence(note: string, evidenceRefs: string) {
  return [note.trim(), evidenceRefs.trim() ? `Evidence refs: ${evidenceRefs.trim()}` : ""].filter(Boolean).join("\n");
}

function cloneChecklist(items: InboundQCChecklistItem[]) {
  return items.map((item) => ({ ...item }));
}

function formatPackagingStatus(status: ReceivingPackagingStatus) {
  switch (status) {
    case "damaged":
      return "Damaged";
    case "missing_label":
      return "Missing label";
    case "leaking":
      return "Leaking";
    case "intact":
    default:
      return "Intact";
  }
}

async function runDecisionAction(id: string, result: InboundQCResult, input: Parameters<typeof passInboundQCInspection>[1]) {
  switch (result) {
    case "pass":
      return passInboundQCInspection(id, input);
    case "fail":
      return failInboundQCInspection(id, input);
    case "hold":
      return holdInboundQCInspection(id, input);
    case "partial":
    default:
      return partialInboundQCInspection(id, input);
  }
}

function QCKPI({ label, value, tone }: { label: string; value: number; tone: StatusTone }) {
  return (
    <div className="erp-card erp-card--padded erp-kpi-card">
      <span className="erp-kpi-label">{label}</span>
      <strong className="erp-kpi-value">{value}</strong>
      <StatusChip tone={tone}>{label}</StatusChip>
    </div>
  );
}

function QCFact({ label, value }: { label: string; value: string }) {
  return (
    <div className="erp-qc-fact">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}
