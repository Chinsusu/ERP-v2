"use client";

import { useEffect, useMemo, useState, type KeyboardEvent } from "react";
import { StatusChip, type StatusTone } from "@/shared/design-system/components";
import { AttachmentPanel, type AttachmentPanelItem } from "@/shared/design-system/pageTemplates";
import {
  applyDispositionToReceipt,
  applyInspectionToReceipt,
  applyReturnDisposition,
  formatReturnInspectionCondition,
  formatReturnInspectionDisposition,
  formatReturnDisposition,
  inspectReturn,
  matchesReturnReceiptCode,
  returnInspectionConditionOptions,
  returnInspectionConditionTone,
  returnInspectionDispositionOptions,
  returnInspectionDispositionTone,
  returnInspectionStatusTone,
  uploadReturnAttachment
} from "../services/returnReceivingService";
import type {
  ReturnAttachment,
  ReturnDispositionAction,
  ReturnInspectionCondition,
  ReturnInspectionDisposition,
  ReturnInspectionResult,
  ReturnReceipt
} from "../types";

type ReturnInspectionPanelProps = {
  receipts: ReturnReceipt[];
  onReceiptChange?: (receipt: ReturnReceipt) => void;
};

export function ReturnInspectionPanel({ receipts, onReceiptChange }: ReturnInspectionPanelProps) {
  const [lookupCode, setLookupCode] = useState("");
  const [selectedReceiptId, setSelectedReceiptId] = useState("");
  const [condition, setCondition] = useState<ReturnInspectionCondition>("intact");
  const [disposition, setDisposition] = useState<ReturnInspectionDisposition>("reusable");
  const [note, setNote] = useState("");
  const [evidenceLabel, setEvidenceLabel] = useState("");
  const [attachmentFile, setAttachmentFile] = useState<File | null>(null);
  const [inspectionResult, setInspectionResult] = useState<ReturnInspectionResult | null>(null);
  const [dispositionAction, setDispositionAction] = useState<ReturnDispositionAction | null>(null);
  const [attachmentResult, setAttachmentResult] = useState<ReturnAttachment | null>(null);
  const [lookupFeedback, setLookupFeedback] = useState<{ tone: StatusTone; message: string } | null>(null);
  const [busyAction, setBusyAction] = useState<"inspect" | "dispose" | "attach" | null>(null);
  const selectedReceipt = useMemo(
    () => receipts.find((receipt) => receipt.id === selectedReceiptId) ?? receipts[0] ?? null,
    [receipts, selectedReceiptId]
  );
  const attachmentItems = useMemo<AttachmentPanelItem[]>(
    () =>
      attachmentResult
        ? [
            {
              id: attachmentResult.id,
              name: attachmentResult.fileName,
              kind: attachmentResult.mimeType,
              uploadedBy: attachmentResult.uploadedBy,
              uploadedAt: attachmentResult.uploadedAt,
              detail: `${attachmentResult.fileSizeBytes} bytes`,
              storageKey: attachmentResult.storageKey,
              status: <StatusChip tone="success">{attachmentResult.status}</StatusChip>,
              canDownload: true,
              onDownload: () =>
                setLookupFeedback({
                  tone: "info",
                  message: attachmentResult.storageKey
                })
            }
          ]
        : [],
    [attachmentResult]
  );

  useEffect(() => {
    if (!selectedReceiptId && receipts[0]) {
      setSelectedReceiptId(receipts[0].id);
    }
  }, [receipts, selectedReceiptId]);

  function handleLookup() {
    const code = lookupCode.trim();
    if (code === "") {
      return;
    }

    const match = receipts.find((receipt) => matchesReturnReceiptCode(receipt, code));
    if (!match) {
      setLookupFeedback({ tone: "danger", message: "Return receipt was not found" });
      return;
    }

    setSelectedReceiptId(match.id);
    setLookupFeedback({ tone: "success", message: `${match.receiptNo} selected` });
  }

  function handleLookupKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (event.key === "Enter") {
      handleLookup();
    }
  }

  function handleConditionChange(nextCondition: ReturnInspectionCondition) {
    setCondition(nextCondition);
    if (nextCondition === "seal_torn" || nextCondition === "used") {
      setDisposition("needs_inspection");
    }
    if (nextCondition === "damaged" || nextCondition === "missing_accessory") {
      setDisposition("not_reusable");
    }
    if (nextCondition === "intact") {
      setDisposition("reusable");
    }
  }

  async function handleConfirmInspection(nextDisposition = disposition) {
    if (!selectedReceipt || busyAction) {
      return;
    }

    setBusyAction("inspect");
    try {
      const result = await inspectReturn({
        receipt: selectedReceipt,
        condition,
        disposition: nextDisposition,
        note,
        evidenceLabel
      });
      const updatedReceipt = applyInspectionToReceipt(selectedReceipt, result);
      onReceiptChange?.(updatedReceipt);
      setSelectedReceiptId(updatedReceipt.id);
      setInspectionResult(result);
      setDispositionAction(null);
      setAttachmentResult(null);
      setDisposition(nextDisposition);
      setLookupFeedback({
        tone: returnInspectionStatusTone(result.status),
        message: `${result.receiptNo} / ${inspectionStatusLabel(result.status)}`
      });
    } catch (cause) {
      setLookupFeedback({
        tone: "danger",
        message: cause instanceof Error ? cause.message : "Return inspection could not be recorded"
      });
    } finally {
      setBusyAction(null);
    }
  }

  async function handleApplyDisposition() {
    if (!selectedReceipt || busyAction) {
      return;
    }

    const receiptForAction =
      inspectionResult?.receiptId === selectedReceipt.id && selectedReceipt.status === "pending_inspection"
        ? applyInspectionToReceipt(selectedReceipt, inspectionResult)
        : selectedReceipt;

    setBusyAction("dispose");
    try {
      const action = await applyReturnDisposition({
        receipt: receiptForAction,
        disposition,
        note
      });
      const updatedReceipt = applyDispositionToReceipt(receiptForAction, action);
      onReceiptChange?.(updatedReceipt);
      setSelectedReceiptId(updatedReceipt.id);
      setDispositionAction(action);
      setLookupFeedback({
        tone: returnInspectionDispositionTone(action.disposition),
        message: `${action.receiptNo} / ${action.actionCode}`
      });
    } catch (cause) {
      setLookupFeedback({
        tone: "danger",
        message: cause instanceof Error ? cause.message : "Return disposition could not be applied"
      });
    } finally {
      setBusyAction(null);
    }
  }

  async function handleUploadAttachment() {
    if (!selectedReceipt || !inspectionResult || !attachmentFile || busyAction) {
      return;
    }

    setBusyAction("attach");
    try {
      const attachment = await uploadReturnAttachment({
        receipt: selectedReceipt,
        inspectionId: inspectionResult.id,
        file: attachmentFile,
        note
      });
      setAttachmentResult(attachment);
      setEvidenceLabel(attachment.fileName);
      setLookupFeedback({
        tone: "success",
        message: `${attachment.receiptNo} / attachment uploaded`
      });
    } catch (cause) {
      setLookupFeedback({
        tone: "danger",
        message: cause instanceof Error ? cause.message : "Return attachment could not be uploaded"
      });
    } finally {
      setBusyAction(null);
    }
  }

  return (
    <section className="erp-returns-inspection-grid" id="return-inspection">
      <div className="erp-card erp-card--padded erp-returns-inspection-card">
        <div className="erp-section-header">
          <h2 className="erp-section-title">Return inspection</h2>
          <StatusChip tone={selectedReceipt?.unknownCase ? "danger" : "info"}>
            {selectedReceipt?.receiptNo ?? "No receipt"}
          </StatusChip>
        </div>

        <div className="erp-returns-inspection-lookup">
          <label className="erp-field">
            <span>Receipt / order / tracking</span>
            <input
              className="erp-input"
              type="text"
              value={lookupCode}
              placeholder="RR-260426-0001 / SO-260425-099 / GHN260425099"
              onChange={(event) => setLookupCode(event.target.value)}
              onKeyDown={handleLookupKeyDown}
            />
          </label>
          <button className="erp-button erp-button--secondary" type="button" onClick={handleLookup}>
            Find receipt
          </button>
        </div>
        {lookupFeedback ? <StatusChip tone={lookupFeedback.tone}>{lookupFeedback.message}</StatusChip> : null}

        {selectedReceipt ? (
          <div className="erp-returns-order-grid" aria-label="Return order information">
            <ReturnInspectionFact label="Order" value={selectedReceipt.originalOrderNo ?? "Unknown"} />
            <ReturnInspectionFact label="Customer" value={selectedReceipt.customerName} />
            <ReturnInspectionFact label="Channel" value={selectedReceipt.channel ?? "Unknown"} />
            <ReturnInspectionFact label="SKU" value={selectedReceipt.lines[0]?.productName ?? "-"} />
            <ReturnInspectionFact label="Batch" value={selectedReceipt.batchNo ?? "-"} />
            <ReturnInspectionFact label="Delivered" value={selectedReceipt.deliveredAt ?? "-"} />
            <ReturnInspectionFact label="Reason" value={selectedReceipt.returnReason ?? selectedReceipt.investigationNote ?? "-"} />
            <ReturnInspectionFact label="Received condition" value={selectedReceipt.packageCondition} />
          </div>
        ) : (
          <div className="erp-returns-empty-state">No return receipt selected</div>
        )}
      </div>

      <div className="erp-card erp-card--padded erp-returns-inspection-card">
        <div className="erp-section-header">
          <h2 className="erp-section-title">Condition</h2>
          <StatusChip tone={returnInspectionConditionTone(condition)}>{formatReturnInspectionCondition(condition)}</StatusChip>
        </div>

        <div className="erp-returns-condition-options" aria-label="Return condition">
          {returnInspectionConditionOptions.map((option) => (
            <button
              aria-pressed={condition === option.value}
              className="erp-returns-condition-option"
              data-active={condition === option.value ? "true" : "false"}
              key={option.value}
              type="button"
              onClick={() => handleConditionChange(option.value)}
            >
              {option.label}
            </button>
          ))}
        </div>

        <div className="erp-section-header erp-section-header--compact erp-returns-inspection-subheader">
          <h3 className="erp-subsection-title">Disposition</h3>
          <StatusChip tone={returnInspectionDispositionTone(disposition)}>
            {formatReturnInspectionDisposition(disposition)}
          </StatusChip>
        </div>
        <div className="erp-returns-disposition-options" aria-label="Inspection disposition">
          {returnInspectionDispositionOptions.map((option) => (
            <button
              aria-pressed={disposition === option.value}
              className="erp-returns-disposition-option"
              data-active={disposition === option.value ? "true" : "false"}
              key={option.value}
              type="button"
              onClick={() => setDisposition(option.value)}
            >
              {option.label}
            </button>
          ))}
        </div>

        <div className="erp-returns-note-grid">
          <label className="erp-field">
            <span>Evidence</span>
            <input
              className="erp-input"
              type="text"
              value={evidenceLabel}
              placeholder="photo-001"
              onChange={(event) => setEvidenceLabel(event.target.value)}
            />
          </label>
          <label className="erp-field">
            <span>Evidence file</span>
            <input
              className="erp-input"
              type="file"
              accept="image/jpeg,image/png,image/webp,video/mp4,video/quicktime"
              onChange={(event) => setAttachmentFile(event.target.files?.[0] ?? null)}
            />
          </label>
          <label className="erp-field">
            <span>Note</span>
            <textarea
              className="erp-input erp-returns-textarea"
              value={note}
              onChange={(event) => setNote(event.target.value)}
            />
          </label>
        </div>

        <div className="erp-returns-actions">
          <button
            className="erp-button erp-button--primary"
            type="button"
            disabled={!selectedReceipt || busyAction !== null || selectedReceipt.status !== "pending_inspection"}
            onClick={() => handleConfirmInspection()}
          >
            {busyAction === "inspect" ? "Recording..." : "Confirm inspection"}
          </button>
          <button
            className="erp-button erp-button--secondary"
            type="button"
            disabled={!selectedReceipt || busyAction !== null || selectedReceipt.status !== "pending_inspection"}
            onClick={() => handleConfirmInspection("needs_inspection")}
          >
            Escalate QA
          </button>
          <button
            className="erp-button erp-button--secondary"
            type="button"
            disabled={
              !selectedReceipt ||
              busyAction !== null ||
              selectedReceipt.status === "dispositioned" ||
              (selectedReceipt.status === "pending_inspection" && inspectionResult?.receiptId !== selectedReceipt.id)
            }
            onClick={() => void handleApplyDisposition()}
          >
            {busyAction === "dispose" ? "Applying..." : "Apply routing"}
          </button>
          <button
            className="erp-button erp-button--secondary"
            type="button"
            disabled={!selectedReceipt || !inspectionResult || !attachmentFile || busyAction !== null}
            onClick={() => void handleUploadAttachment()}
          >
            {busyAction === "attach" ? "Uploading..." : "Attach evidence"}
          </button>
        </div>
      </div>

      <div className="erp-card erp-card--padded erp-returns-inspection-card erp-returns-inspection-result">
        <div className="erp-section-header">
          <h2 className="erp-section-title">Inspection result</h2>
          {inspectionResult ? (
            <StatusChip tone={returnInspectionStatusTone(inspectionResult.status)}>{inspectionResult.status}</StatusChip>
          ) : null}
        </div>

        {inspectionResult ? (
          <div className="erp-returns-result-grid">
            <ReturnInspectionFact label="Receipt" value={inspectionResult.receiptNo} />
            <ReturnInspectionFact label="Condition" value={formatReturnInspectionCondition(inspectionResult.condition)} />
            <ReturnInspectionFact label="Disposition" value={formatReturnInspectionDisposition(inspectionResult.disposition)} />
            <ReturnInspectionFact label="Target" value={inspectionResult.targetLocation} />
            <ReturnInspectionFact label="Risk" value={inspectionResult.riskLevel} />
            <ReturnInspectionFact label="Evidence" value={inspectionResult.evidenceLabel ?? "-"} />
            <ReturnInspectionFact label="Inspector" value={inspectionResult.inspectorId} />
          </div>
        ) : (
          <div className="erp-returns-empty-state">No inspection result recorded</div>
        )}
        {dispositionAction ? (
          <div className="erp-returns-result-grid erp-returns-disposition-result">
            <ReturnInspectionFact label="Routing" value={formatReturnDisposition(dispositionAction.disposition)} />
            <ReturnInspectionFact label="Target" value={dispositionAction.targetLocation} />
            <ReturnInspectionFact label="Stock status" value={dispositionAction.targetStockStatus} />
            <ReturnInspectionFact label="Action" value={dispositionAction.actionCode} />
          </div>
        ) : null}
        <AttachmentPanel
          title="Inspection attachments"
          items={attachmentItems}
          emptyMessage="Attach inspection evidence after recording the result."
        />
      </div>
    </section>
  );
}

function inspectionStatusLabel(status: ReturnInspectionResult["status"]) {
  if (status === "return_qa_hold") {
    return "QA hold";
  }

  return "Inspection recorded";
}

function ReturnInspectionFact({ label, value }: { label: string; value: string }) {
  return (
    <div className="erp-returns-fact">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}
