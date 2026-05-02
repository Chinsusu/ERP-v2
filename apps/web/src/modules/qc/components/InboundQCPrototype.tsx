"use client";

import { useEffect, useMemo, useState } from "react";
import { DataTable, EmptyState, StatusChip, type DataTableColumn, type StatusTone } from "@/shared/design-system/components";
import { AttachmentPanel, type AttachmentPanelItem } from "@/shared/design-system/pageTemplates";
import { t } from "@/shared/i18n";
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
  formatInboundQCDateTime,
  formatInboundQCQuantity,
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
  const qcAttachmentItems = useMemo<AttachmentPanelItem[]>(
    () =>
      splitEvidenceRefs(evidenceRefs).map((ref, index) => ({
        id: `qc-evidence-${index}-${ref}`,
        name: ref,
        kind: ref.toLowerCase().endsWith(".pdf") ? qcCopy("inbound.attachments.document") : qcCopy("inbound.attachments.evidence"),
        uploadedBy: inspectorId,
        uploadedAt: selectedInspection?.updatedAt ?? new Date().toISOString(),
        storageKey: selectedInspection ? `inbound-qc/${selectedInspection.id}/${ref}` : ref,
        status: <StatusChip tone="info">{qcCopy("inbound.attachments.reference")}</StatusChip>,
        canDownload: true,
        canDelete: selectedInspection?.status !== "completed",
        deleteLabel: qcCopy("inbound.actions.remove"),
        onDownload: () => setFeedback({ tone: "info", message: ref }),
        onDelete: () => removeEvidenceRef(ref)
      })),
    [evidenceRefs, inspectorId, selectedInspection?.id, selectedInspection?.status, selectedInspection?.updatedAt]
  );
  const selectedLineColumns = useMemo<DataTableColumn<InspectableReceivingLine>[]>(
    () => [
      {
        key: "receipt",
        header: qcCopy("inbound.receivingLines.columns.receipt"),
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
        header: qcCopy("inbound.receivingLines.columns.sku"),
        render: (row) => (
          <span className="erp-qc-record-cell">
            <strong>{row.line.sku}</strong>
            <small>{row.line.itemName ?? row.line.itemId}</small>
          </span>
        )
      },
      {
        key: "lot",
        header: qcCopy("inbound.receivingLines.columns.lotExpiry"),
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
        header: qcCopy("inbound.receivingLines.columns.packaging"),
        render: (row) => (
          <StatusChip tone={row.line.packagingStatus === "intact" ? "success" : "warning"}>
            {formatPackagingStatus(row.line.packagingStatus)}
          </StatusChip>
        ),
        width: "150px"
      },
      {
        key: "quantity",
        header: qcCopy("inbound.receivingLines.columns.quantity"),
        render: (row) => formatReceivingQuantity(row.line.quantity, row.line.baseUomCode),
        align: "right",
        width: "130px"
      },
      {
        key: "action",
        header: qcCopy("inbound.receivingLines.columns.action"),
        render: (row) => {
          const inspection = findInspectionForLine(visibleInspections, row);

          return (
            <button className="erp-button erp-button--secondary" type="button" onClick={() => handleSelectLine(row)}>
              {inspection
                ? qcCopy("inbound.actions.open")
                : row.key === selectedLine?.key
                  ? qcCopy("inbound.actions.selected")
                  : qcCopy("inbound.actions.select")}
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
        header: qcCopy("inbound.inspections.columns.inspection"),
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
        header: qcCopy("inbound.inspections.columns.skuLot"),
        render: (row) => (
          <span className="erp-qc-record-cell">
            <strong>{row.sku}</strong>
            <small>{row.lotNo}</small>
          </span>
        )
      },
      {
        key: "status",
        header: qcCopy("inbound.inspections.columns.status"),
        render: (row) => <StatusChip tone={inboundQCStatusTone(row.status)}>{inboundQCStatusLabel(row.status)}</StatusChip>,
        width: "140px"
      },
      {
        key: "result",
        header: qcCopy("inbound.inspections.columns.result"),
        render: (row) => <StatusChip tone={inboundQCResultTone(row.result)}>{inboundQCResultLabel(row.result)}</StatusChip>,
        width: "120px"
      },
      {
        key: "qty",
        header: qcCopy("inbound.inspections.columns.quantity"),
        render: (row) => formatInboundQCQuantity(row.quantity, row.uomCode),
        align: "right",
        width: "130px"
      },
      {
        key: "updated",
        header: qcCopy("inbound.inspections.columns.updated"),
        render: (row) => formatInboundQCDateTime(row.updatedAt),
        width: "150px"
      },
      {
        key: "open",
        header: qcCopy("inbound.inspections.columns.action"),
        render: (row) => (
          <button className="erp-button erp-button--secondary" type="button" onClick={() => setSelectedInspectionId(row.id)}>
            {qcCopy("inbound.actions.open")}
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
      setFeedback({ tone: "warning", message: qcCopy("inbound.feedback.alreadyCovered", { id: existingLineInspection.id }) });
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
      setFeedback({ tone: "success", message: qcCopy("inbound.feedback.created", { id: result.inspection.id }) });
    } catch (cause) {
      setFeedback({
        tone: "danger",
        message: cause instanceof Error ? cause.message : qcCopy("inbound.feedback.createFailed")
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
      setFeedback({ tone: "success", message: qcCopy("inbound.feedback.started", { id: result.inspection.id }) });
    } catch (cause) {
      setFeedback({
        tone: "danger",
        message: cause instanceof Error ? cause.message : qcCopy("inbound.feedback.startFailed")
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
      setFeedback({ tone: "warning", message: qcCopy("inbound.feedback.startBeforeDecision") });
      return;
    }
    if (result !== "pass" && reason.trim() === "") {
      setFeedback({ tone: "warning", message: qcCopy("inbound.feedback.reasonRequired") });
      return;
    }

    const checklist = checklistForDecision(checklistDraft);
    if (result === "pass" && checklist.some((item) => item.required && item.status === "fail")) {
      setFeedback({ tone: "warning", message: qcCopy("inbound.feedback.requiredFailBlocksPass") });
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
        message: `${actionResult.inspection.id} / ${inboundQCResultLabel(actionResult.inspection.result)}`
      });
    } catch (cause) {
      setFeedback({
        tone: "danger",
        message: cause instanceof Error ? cause.message : qcCopy("inbound.feedback.decisionFailed")
      });
    } finally {
      setBusyAction("");
    }
  }

  function updateChecklistItem(id: string, patch: Partial<InboundQCChecklistItem>) {
    setChecklistDraft((current) => current.map((item) => (item.id === id ? { ...item, ...patch } : item)));
  }

  function removeEvidenceRef(ref: string) {
    setEvidenceRefs((current) => splitEvidenceRefs(current).filter((candidate) => candidate !== ref).join(", "));
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
          <h1 className="erp-page-title">{qcCopy("inbound.title")}</h1>
          <p className="erp-page-description">{qcCopy("inbound.description")}</p>
        </div>
        <div className="erp-page-actions">
          <a className="erp-button erp-button--secondary" href="#qc-receiving-lines">
            {qcCopy("inbound.nav.receipts")}
          </a>
          <a className="erp-button erp-button--secondary" href="#qc-decision">
            {qcCopy("inbound.nav.decision")}
          </a>
          <a className="erp-button erp-button--primary" href="#qc-inspections">
            {qcCopy("inbound.nav.inspections")}
          </a>
        </div>
      </header>

      <section className="erp-qc-toolbar" aria-label={qcCopy("inbound.filters.label")}>
        <label className="erp-field">
          <span>{qcCopy("inbound.filters.warehouse")}</span>
          <select className="erp-input" value={warehouseId} onChange={(event) => setWarehouseId(event.target.value)}>
            {receivingWarehouseOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>{qcCopy("inbound.filters.status")}</span>
          <select className="erp-input" value={status} onChange={(event) => setStatus(event.target.value as StatusFilter)}>
            {inboundQCStatusOptions.map((option) => (
              <option key={option.value || "all"} value={option.value}>
                {option.value ? inboundQCStatusLabel(option.value) : qcCopy("inbound.filters.allStatuses")}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>{qcCopy("inbound.filters.inspector")}</span>
          <input className="erp-input" type="text" value={inspectorId} onChange={(event) => setInspectorId(event.target.value)} />
        </label>
      </section>

      <section className="erp-kpi-grid erp-qc-kpis">
        <QCKPI label={qcCopy("inbound.kpi.inspectableLines")} value={inspectableLines.length} tone={inspectableLines.length > 0 ? "info" : "normal"} />
        <QCKPI label={inboundQCStatusLabel("pending")} value={totals.pending} tone={totals.pending > 0 ? "warning" : "normal"} />
        <QCKPI label={inboundQCStatusLabel("in_progress")} value={totals.inProgress} tone={totals.inProgress > 0 ? "info" : "normal"} />
        <QCKPI label={inboundQCStatusLabel("completed")} value={totals.completed} tone="success" />
        <QCKPI label={qcCopy("inbound.kpi.holdFail")} value={totals.holdOrFail} tone={totals.holdOrFail > 0 ? "danger" : "normal"} />
      </section>

      <section className="erp-qc-workspace">
        <div className="erp-card erp-card--padded erp-qc-panel" id="qc-receiving-lines">
          <div className="erp-section-header">
            <h2 className="erp-section-title">{qcCopy("inbound.receivingLines.title")}</h2>
            <StatusChip tone={receiptsLoading ? "warning" : "info"}>
              {receiptsLoading ? qcCopy("inbound.feedback.loading") : qcCopy("inbound.receivingLines.lines", { count: inspectableLines.length })}
            </StatusChip>
          </div>

          <DataTable
            columns={selectedLineColumns}
            rows={inspectableLines}
            getRowKey={(row) => row.key}
            loading={receiptsLoading}
            error={receiptsError?.message}
            emptyState={
              <EmptyState
                title={qcCopy("inbound.receivingLines.emptyTitle")}
                description={qcCopy("inbound.receivingLines.emptyDescription")}
              />
            }
          />

          {selectedLine ? (
            <div className="erp-qc-selected-line" aria-label={qcCopy("inbound.receivingLines.selectedLabel")}>
              <StatusChip tone={existingLineInspection ? inboundQCStatusTone(existingLineInspection.status) : "info"}>
                {existingLineInspection ? inboundQCStatusLabel(existingLineInspection.status) : qcCopy("inbound.receivingLines.readyForQC")}
              </StatusChip>
              <div>
                <strong>{selectedLine.receipt.receiptNo}</strong>
                <span>
                  {selectedLine.line.sku} / {selectedLine.line.lotNo ?? selectedLine.line.batchNo ?? "-"} /{" "}
                  {formatReceivingQuantity(selectedLine.line.quantity, selectedLine.line.baseUomCode)}
                </span>
              </div>
              <button className="erp-button erp-button--primary" type="button" disabled={busyAction !== ""} onClick={handleCreateInspection}>
                {existingLineInspection
                  ? qcCopy("inbound.actions.openInspection")
                  : busyAction === "create"
                    ? qcCopy("inbound.actions.creating")
                    : qcCopy("inbound.actions.createInspection")}
              </button>
            </div>
          ) : null}
        </div>

        <div className="erp-card erp-card--padded erp-qc-panel" id="qc-decision">
          <div className="erp-section-header">
            <h2 className="erp-section-title">{qcCopy("inbound.decision.title")}</h2>
            <StatusChip tone={selectedInspection ? inboundQCStatusTone(selectedInspection.status) : "normal"}>
              {selectedInspection ? inboundQCStatusLabel(selectedInspection.status) : qcCopy("inbound.empty.noInspection")}
            </StatusChip>
          </div>

          {feedback ? <StatusChip tone={feedback.tone}>{feedback.message}</StatusChip> : null}

          {selectedInspection ? (
            <>
              <div className="erp-qc-fact-grid" aria-label={qcCopy("inbound.decision.factsLabel")}>
                <QCFact label={qcCopy("inbound.fact.receipt")} value={selectedInspection.goodsReceiptNo} />
                <QCFact label={qcCopy("inbound.fact.po")} value={selectedInspection.purchaseOrderId ?? "-"} />
                <QCFact label={qcCopy("inbound.fact.sku")} value={selectedInspection.sku} />
                <QCFact label={qcCopy("inbound.fact.lotExpiry")} value={`${selectedInspection.lotNo} / ${selectedInspection.expiryDate}`} />
                <QCFact label={qcCopy("inbound.fact.quantity")} value={formatInboundQCQuantity(selectedInspection.quantity, selectedInspection.uomCode)} />
                <QCFact label={qcCopy("inbound.fact.inspector")} value={selectedInspection.inspectorId} />
                <QCFact label={qcCopy("inbound.fact.started")} value={formatInboundQCDateTime(selectedInspection.startedAt)} />
                <QCFact label={qcCopy("inbound.fact.decided")} value={formatInboundQCDateTime(selectedInspection.decidedAt)} />
              </div>

              <div className="erp-qc-checklist" aria-label={qcCopy("inbound.checklist.label")}>
                {checklistDraft.map((item) => (
                  <div className="erp-qc-checklist-item" key={item.id}>
                    <div className="erp-qc-checklist-label">
                      <strong>{checklistItemLabel(item)}</strong>
                      <small>
                        {item.code} / {item.required ? qcCopy("inbound.checklist.required") : qcCopy("inbound.checklist.optional")}
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
                          {checklistStatusLabel(option.value)}
                        </option>
                      ))}
                    </select>
                    <input
                      className="erp-input"
                      type="text"
                      value={item.note ?? ""}
                      placeholder={qcCopy("inbound.checklist.notePlaceholder")}
                      disabled={selectedInspection.status === "completed"}
                      onChange={(event) => updateChecklistItem(item.id, { note: event.currentTarget.value })}
                    />
                    <StatusChip tone={checklistStatusTone(item.status)}>{checklistStatusLabel(item.status)}</StatusChip>
                  </div>
                ))}
              </div>

              <div className="erp-qc-decision-grid" aria-label={qcCopy("inbound.decision.quantitiesLabel")}>
                <label className="erp-field">
                  <span>{qcCopy("inbound.decision.passQty")}</span>
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
                  <span>{qcCopy("inbound.decision.failQty")}</span>
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
                  <span>{qcCopy("inbound.decision.holdQty")}</span>
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
                  <span>{qcCopy("inbound.decision.reason")}</span>
                  <textarea
                    className="erp-input"
                    value={reason}
                    placeholder={qcCopy("inbound.decision.reasonPlaceholder")}
                    disabled={selectedInspection.status === "completed"}
                    onChange={(event) => setReason(event.currentTarget.value)}
                  />
                </label>
                <label className="erp-field erp-qc-note-field">
                  <span>{qcCopy("inbound.decision.note")}</span>
                  <textarea
                    className="erp-input"
                    value={decisionNote}
                    disabled={selectedInspection.status === "completed"}
                    onChange={(event) => setDecisionNote(event.currentTarget.value)}
                  />
                </label>
                <label className="erp-field erp-qc-note-field">
                  <span>{qcCopy("inbound.decision.evidenceRefs")}</span>
                  <textarea
                    className="erp-input"
                    value={evidenceRefs}
                    placeholder={qcCopy("inbound.decision.evidencePlaceholder")}
                    disabled={selectedInspection.status === "completed"}
                    onChange={(event) => setEvidenceRefs(event.currentTarget.value)}
                  />
                </label>
              </div>

              <AttachmentPanel
                title={qcCopy("inbound.attachments.title")}
                items={qcAttachmentItems}
                emptyMessage={qcCopy("inbound.attachments.empty")}
              />

              <div className="erp-qc-actions">
                <button
                  className="erp-button erp-button--secondary"
                  type="button"
                  disabled={busyAction !== "" || selectedInspection.status !== "pending"}
                  onClick={handleStartInspection}
                >
                  {busyAction === "start" ? qcCopy("inbound.actions.starting") : qcCopy("inbound.actions.start")}
                </button>
                <button
                  className="erp-button erp-button--primary"
                  type="button"
                  disabled={busyAction !== "" || selectedInspection.status !== "in_progress"}
                  onClick={() => void handleDecision("pass")}
                >
                  {inboundQCResultLabel("pass")}
                </button>
                <button
                  className="erp-button erp-button--secondary"
                  type="button"
                  disabled={busyAction !== "" || selectedInspection.status !== "in_progress"}
                  onClick={() => void handleDecision("hold")}
                >
                  {inboundQCResultLabel("hold")}
                </button>
                <button
                  className="erp-button erp-button--secondary"
                  type="button"
                  disabled={busyAction !== "" || selectedInspection.status !== "in_progress"}
                  onClick={() => void handleDecision("partial")}
                >
                  {inboundQCResultLabel("partial")}
                </button>
                <button
                  className="erp-button erp-button--danger"
                  type="button"
                  disabled={busyAction !== "" || selectedInspection.status !== "in_progress"}
                  onClick={() => void handleDecision("fail")}
                >
                  {inboundQCResultLabel("fail")}
                </button>
              </div>
            </>
          ) : (
            <>
              <EmptyState title={qcCopy("inbound.empty.noInspectionSelected")} description={qcCopy("inbound.empty.noInspectionDescription")} />
              <AttachmentPanel
                title={qcCopy("inbound.attachments.title")}
                items={qcAttachmentItems}
                emptyMessage={qcCopy("inbound.attachments.empty")}
              />
            </>
          )}
        </div>
      </section>

      <section className="erp-card erp-card--padded erp-module-table-card" id="qc-inspections">
        <div className="erp-section-header">
          <h2 className="erp-section-title">{qcCopy("inbound.inspections.title")}</h2>
          <StatusChip tone={inspectionsLoading ? "warning" : "info"}>
            {inspectionsLoading ? qcCopy("inbound.feedback.loading") : qcCopy("inbound.inspections.records", { count: visibleInspections.length })}
          </StatusChip>
        </div>
        <DataTable
          columns={inspectionColumns}
          rows={visibleInspections}
          getRowKey={(row) => row.id}
          loading={inspectionsLoading}
          error={inspectionsError?.message}
          emptyState={<EmptyState title={qcCopy("inbound.inspections.emptyTitle")} description={qcCopy("inbound.inspections.emptyDescription")} />}
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
    `${qcCopy("inbound.note.packaging")}: ${formatPackagingStatus(row.line.packagingStatus)}; ${qcCopy("inbound.note.lot")}: ${row.line.lotNo ?? row.line.batchNo ?? "-"}; ${qcCopy("inbound.note.received")}: ${formatReceivingDateTime(row.receipt.inspectReadyAt)}`,
    evidenceRefs
  );
}

function decisionNoteWithEvidence(note: string, evidenceRefs: string) {
  return [note.trim(), evidenceRefs.trim() ? `${qcCopy("inbound.decision.evidenceRefs")}: ${evidenceRefs.trim()}` : ""]
    .filter(Boolean)
    .join("\n");
}

function splitEvidenceRefs(value: string) {
  return value
    .split(/[\n,]+/)
    .map((ref) => ref.trim())
    .filter(Boolean);
}

function cloneChecklist(items: InboundQCChecklistItem[]) {
  return items.map((item) => ({ ...item }));
}

function formatPackagingStatus(status: ReceivingPackagingStatus) {
  switch (status) {
    case "damaged":
      return qcCopy("inbound.packaging.damaged");
    case "missing_label":
      return qcCopy("inbound.packaging.missing_label");
    case "leaking":
      return qcCopy("inbound.packaging.leaking");
    case "intact":
    default:
      return qcCopy("inbound.packaging.intact");
  }
}

function inboundQCStatusLabel(status: InboundQCInspectionStatus) {
  return qcCopy(`inbound.status.${status}`);
}

function inboundQCResultLabel(result?: InboundQCResult) {
  return result ? qcCopy(`inbound.result.${result}`) : "-";
}

function checklistStatusLabel(status: InboundQCChecklistStatus) {
  return qcCopy(`inbound.checklist.status.${status}`);
}

function checklistItemLabel(item: InboundQCChecklistItem) {
  return qcCopy(`inbound.checklist.item.${item.code}`, undefined, item.label);
}

function qcCopy(key: string, values?: Record<string, string | number>, fallback?: string) {
  return t(`qc.${key}`, { values, fallback });
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
