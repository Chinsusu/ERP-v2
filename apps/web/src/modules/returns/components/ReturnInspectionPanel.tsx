"use client";

import { useEffect, useMemo, useState, type KeyboardEvent } from "react";
import { StatusChip, type StatusTone } from "@/shared/design-system/components";
import { AttachmentPanel, type AttachmentPanelItem } from "@/shared/design-system/pageTemplates";
import { t } from "@/shared/i18n";
import {
  applyDispositionToReceipt,
  applyInspectionToReceipt,
  applyReturnDisposition,
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
              kind: returnAttachmentKindLabel(attachmentResult.mimeType),
              uploadedBy: attachmentResult.uploadedBy,
              uploadedAt: attachmentResult.uploadedAt,
              detail: returnInspectionCopy("fileSize", { bytes: attachmentResult.fileSizeBytes }),
              storageKey: attachmentResult.storageKey,
              status: <StatusChip tone="success">{attachmentStatusLabel(attachmentResult.status)}</StatusChip>,
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
      setLookupFeedback({ tone: "danger", message: returnInspectionCopy("feedback.notFound") });
      return;
    }

    setSelectedReceiptId(match.id);
    setLookupFeedback({ tone: "success", message: returnInspectionCopy("feedback.selected", { receiptNo: match.receiptNo }) });
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
        message: cause instanceof Error ? cause.message : returnInspectionCopy("feedback.couldNotRecord")
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
        message: `${action.receiptNo} / ${returnActionCodeLabel(action.actionCode)}`
      });
    } catch (cause) {
      setLookupFeedback({
        tone: "danger",
        message: cause instanceof Error ? cause.message : returnInspectionCopy("feedback.couldNotApplyDisposition")
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
        message: returnInspectionCopy("feedback.attachmentUploaded", { receiptNo: attachment.receiptNo })
      });
    } catch (cause) {
      setLookupFeedback({
        tone: "danger",
        message: cause instanceof Error ? cause.message : returnInspectionCopy("feedback.couldNotUploadAttachment")
      });
    } finally {
      setBusyAction(null);
    }
  }

  return (
    <section className="erp-returns-inspection-grid" id="return-inspection">
      <div className="erp-card erp-card--padded erp-returns-inspection-card">
        <div className="erp-section-header">
          <h2 className="erp-section-title">{returnInspectionCopy("title")}</h2>
          <StatusChip tone={selectedReceipt?.unknownCase ? "danger" : "info"}>
            {selectedReceipt?.receiptNo ?? returnInspectionCopy("empty.noReceipt")}
          </StatusChip>
        </div>

        <div className="erp-returns-inspection-lookup">
          <label className="erp-field">
            <span>{returnInspectionCopy("lookupLabel")}</span>
            <input
              className="erp-input"
              type="text"
              value={lookupCode}
              placeholder={returnInspectionCopy("lookupPlaceholder")}
              onChange={(event) => setLookupCode(event.target.value)}
              onKeyDown={handleLookupKeyDown}
            />
          </label>
          <button className="erp-button erp-button--secondary" type="button" onClick={handleLookup}>
            {returnInspectionCopy("actions.findReceipt")}
          </button>
        </div>
        {lookupFeedback ? <StatusChip tone={lookupFeedback.tone}>{lookupFeedback.message}</StatusChip> : null}

        {selectedReceipt ? (
          <div className="erp-returns-order-grid" aria-label={returnInspectionCopy("orderInfoLabel")}>
            <ReturnInspectionFact label={returnInspectionCopy("facts.order")} value={selectedReceipt.originalOrderNo ?? returnInspectionCopy("empty.unknown")} />
            <ReturnInspectionFact label={returnInspectionCopy("facts.customer")} value={selectedReceipt.customerName} />
            <ReturnInspectionFact label={returnInspectionCopy("facts.channel")} value={selectedReceipt.channel ?? returnInspectionCopy("empty.unknown")} />
            <ReturnInspectionFact label={returnInspectionCopy("facts.sku")} value={selectedReceipt.lines[0]?.productName ?? "-"} />
            <ReturnInspectionFact label={returnInspectionCopy("facts.batch")} value={selectedReceipt.batchNo ?? "-"} />
            <ReturnInspectionFact label={returnInspectionCopy("facts.delivered")} value={selectedReceipt.deliveredAt ?? "-"} />
            <ReturnInspectionFact label={returnInspectionCopy("facts.reason")} value={selectedReceipt.returnReason ?? selectedReceipt.investigationNote ?? "-"} />
            <ReturnInspectionFact label={returnInspectionCopy("facts.receivedCondition")} value={selectedReceipt.packageCondition} />
          </div>
        ) : (
          <div className="erp-returns-empty-state">{returnInspectionCopy("empty.noReceiptSelected")}</div>
        )}
      </div>

      <div className="erp-card erp-card--padded erp-returns-inspection-card">
        <div className="erp-section-header">
          <h2 className="erp-section-title">{returnInspectionCopy("sections.condition")}</h2>
          <StatusChip tone={returnInspectionConditionTone(condition)}>{returnConditionLabel(condition)}</StatusChip>
        </div>

        <div className="erp-returns-condition-options" aria-label={returnInspectionCopy("sections.condition")}>
          {returnInspectionConditionOptions.map((option) => (
            <button
              aria-pressed={condition === option.value}
              className="erp-returns-condition-option"
              data-active={condition === option.value ? "true" : "false"}
              key={option.value}
              type="button"
              onClick={() => handleConditionChange(option.value)}
            >
              {returnConditionLabel(option.value)}
            </button>
          ))}
        </div>

        <div className="erp-section-header erp-section-header--compact erp-returns-inspection-subheader">
          <h3 className="erp-subsection-title">{returnInspectionCopy("sections.disposition")}</h3>
          <StatusChip tone={returnInspectionDispositionTone(disposition)}>
            {returnInspectionDispositionLabel(disposition)}
          </StatusChip>
        </div>
        <div className="erp-returns-disposition-options" aria-label={returnInspectionCopy("sections.disposition")}>
          {returnInspectionDispositionOptions.map((option) => (
            <button
              aria-pressed={disposition === option.value}
              className="erp-returns-disposition-option"
              data-active={disposition === option.value ? "true" : "false"}
              key={option.value}
              type="button"
              onClick={() => setDisposition(option.value)}
            >
              {returnInspectionDispositionLabel(option.value)}
            </button>
          ))}
        </div>

        <div className="erp-returns-note-grid">
          <label className="erp-field">
            <span>{returnInspectionCopy("fields.evidence")}</span>
            <input
              className="erp-input"
              type="text"
              value={evidenceLabel}
              placeholder="photo-001"
              onChange={(event) => setEvidenceLabel(event.target.value)}
            />
          </label>
          <label className="erp-field">
            <span>{returnInspectionCopy("fields.evidenceFile")}</span>
            <input
              className="erp-input"
              type="file"
              accept="image/jpeg,image/png,image/webp,video/mp4,video/quicktime"
              onChange={(event) => setAttachmentFile(event.target.files?.[0] ?? null)}
            />
          </label>
          <label className="erp-field">
            <span>{returnInspectionCopy("fields.note")}</span>
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
            {busyAction === "inspect" ? returnInspectionCopy("actions.recording") : returnInspectionCopy("actions.confirmInspection")}
          </button>
          <button
            className="erp-button erp-button--secondary"
            type="button"
            disabled={!selectedReceipt || busyAction !== null || selectedReceipt.status !== "pending_inspection"}
            onClick={() => handleConfirmInspection("needs_inspection")}
          >
            {returnInspectionCopy("actions.escalateQA")}
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
            {busyAction === "dispose" ? returnInspectionCopy("actions.applying") : returnInspectionCopy("actions.applyRouting")}
          </button>
          <button
            className="erp-button erp-button--secondary"
            type="button"
            disabled={!selectedReceipt || !inspectionResult || !attachmentFile || busyAction !== null}
            onClick={() => void handleUploadAttachment()}
          >
            {busyAction === "attach" ? returnInspectionCopy("actions.uploading") : returnInspectionCopy("actions.attachEvidence")}
          </button>
        </div>
      </div>

      <div className="erp-card erp-card--padded erp-returns-inspection-card erp-returns-inspection-result">
        <div className="erp-section-header">
          <h2 className="erp-section-title">{returnInspectionCopy("sections.result")}</h2>
          {inspectionResult ? (
            <StatusChip tone={returnInspectionStatusTone(inspectionResult.status)}>{inspectionStatusLabel(inspectionResult.status)}</StatusChip>
          ) : null}
        </div>

        {inspectionResult ? (
          <div className="erp-returns-result-grid">
            <ReturnInspectionFact label={returnInspectionCopy("facts.receipt")} value={inspectionResult.receiptNo} />
            <ReturnInspectionFact label={returnInspectionCopy("facts.condition")} value={returnConditionLabel(inspectionResult.condition)} />
            <ReturnInspectionFact label={returnInspectionCopy("facts.disposition")} value={returnInspectionDispositionLabel(inspectionResult.disposition)} />
            <ReturnInspectionFact label={returnInspectionCopy("facts.target")} value={returnInspectionTargetLabel(inspectionResult.targetLocation)} />
            <ReturnInspectionFact label={returnInspectionCopy("facts.risk")} value={returnRiskLabel(inspectionResult.riskLevel)} />
            <ReturnInspectionFact label={returnInspectionCopy("facts.evidence")} value={inspectionResult.evidenceLabel ?? "-"} />
            <ReturnInspectionFact label={returnInspectionCopy("facts.inspector")} value={inspectionResult.inspectorId} />
          </div>
        ) : (
          <div className="erp-returns-empty-state">{returnInspectionCopy("empty.noResult")}</div>
        )}
        {dispositionAction ? (
          <div className="erp-returns-result-grid erp-returns-disposition-result">
            <ReturnInspectionFact label={returnInspectionCopy("facts.routing")} value={returnInspectionDispositionLabel(dispositionAction.disposition)} />
            <ReturnInspectionFact label={returnInspectionCopy("facts.target")} value={returnInspectionTargetLabel(dispositionAction.targetLocation)} />
            <ReturnInspectionFact label={returnInspectionCopy("facts.stockStatus")} value={returnStockStatusLabel(dispositionAction.targetStockStatus)} />
            <ReturnInspectionFact label={returnInspectionCopy("facts.action")} value={returnActionCodeLabel(dispositionAction.actionCode)} />
          </div>
        ) : null}
        <AttachmentPanel
          title={returnInspectionCopy("sections.attachments")}
          items={attachmentItems}
          emptyMessage={returnInspectionCopy("empty.attachmentEmpty")}
        />
      </div>
    </section>
  );
}

function inspectionStatusLabel(status: ReturnInspectionResult["status"]) {
  return returnInspectionCopy(`status.${status}`);
}

function returnInspectionCopy(key: string, values?: Record<string, string | number>) {
  return t(`returns.inspection.${key}`, { values });
}

function returnConditionLabel(condition: ReturnInspectionCondition) {
  return returnInspectionCopy(`condition.${condition}`);
}

function returnInspectionDispositionLabel(disposition: ReturnInspectionDisposition) {
  return returnInspectionCopy(`disposition.${disposition}`);
}

function returnRiskLabel(riskLevel: ReturnInspectionResult["riskLevel"]) {
  return returnInspectionCopy(`risk.${riskLevel}`);
}

function returnActionCodeLabel(actionCode: ReturnDispositionAction["actionCode"]) {
  return returnInspectionCopy(`actionCode.${actionCode}`);
}

function returnStockStatusLabel(stockStatus: ReturnDispositionAction["targetStockStatus"]) {
  return returnInspectionCopy(`stockStatus.${stockStatus}`);
}

function returnInspectionTargetLabel(targetLocation: string) {
  return t(`returns.inspection.targetLocation.${targetLocation}`, { fallback: targetLocation });
}

function attachmentStatusLabel(status: ReturnAttachment["status"]) {
  return returnInspectionCopy(`attachmentStatus.${status}`);
}

function returnAttachmentKindLabel(mimeType: string) {
  if (mimeType.toLowerCase().startsWith("image/")) {
    return returnInspectionCopy("attachmentKind.image");
  }
  if (mimeType.toLowerCase() === "application/pdf") {
    return returnInspectionCopy("attachmentKind.pdf");
  }

  return returnInspectionCopy("attachmentKind.evidence");
}

function ReturnInspectionFact({ label, value }: { label: string; value: string }) {
  return (
    <div className="erp-returns-fact">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}
