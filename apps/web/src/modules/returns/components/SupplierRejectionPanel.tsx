"use client";

import { useEffect, useMemo, useState, type FormEvent } from "react";
import { DataTable, EmptyState, StatusChip, type DataTableColumn, type StatusTone } from "@/shared/design-system/components";
import { AttachmentPanel, type AttachmentPanelItem } from "@/shared/design-system/pageTemplates";
import { t } from "@/shared/i18n";
import { getAuditLogs, auditActionTone, compactAuditPayload } from "@/modules/audit/services/auditLogService";
import type { AuditLogItem } from "@/modules/audit/types";
import { useSupplierRejections } from "../hooks/useSupplierRejections";
import {
  confirmSupplierRejection,
  createSupplierRejection,
  formatSupplierRejectionDateTime,
  formatSupplierRejectionQuantity,
  submitSupplierRejection,
  supplierRejectionSampleLines,
  supplierRejectionStatusOptions,
  supplierRejectionStatusTone,
  supplierRejectionSupplierOptions,
  supplierRejectionWarehouseOptions
} from "../services/supplierRejectionService";
import type { SupplierRejection, SupplierRejectionQuery, SupplierRejectionStatus } from "../types";

type StatusFilter = "" | SupplierRejectionStatus;

const rejectionColumns: DataTableColumn<SupplierRejection>[] = [
  {
    key: "rejection",
    header: supplierRejectionCopy("columns.rejection"),
    render: (row) => (
      <span className="erp-supplier-rejection-record-cell">
        <strong>{row.rejectionNo}</strong>
        <small>{row.goodsReceiptNo ?? row.goodsReceiptId}</small>
      </span>
    ),
    width: "210px"
  },
  {
    key: "supplier",
    header: supplierRejectionCopy("columns.supplier"),
    render: (row) => row.supplierCode ?? row.supplierName ?? row.supplierId,
    width: "160px"
  },
  {
    key: "status",
    header: supplierRejectionCopy("columns.status"),
    render: (row) => <StatusChip tone={supplierRejectionStatusTone(row.status)}>{supplierRejectionStatusLabel(row.status)}</StatusChip>,
    width: "130px"
  },
  {
    key: "qty",
    header: supplierRejectionCopy("columns.rejectedQty"),
    render: (row) => formatSupplierRejectionQuantity(totalRejectedQuantity(row), row.lines[0]?.baseUOMCode),
    align: "right",
    width: "140px"
  },
  {
    key: "batch",
    header: supplierRejectionCopy("columns.batchReason"),
    render: (row) => (
      <span className="erp-supplier-rejection-record-cell">
        <strong>{row.lines[0]?.lotNo ?? "-"}</strong>
        <small>{row.reason}</small>
      </span>
    )
  },
  {
    key: "attachments",
    header: supplierRejectionCopy("columns.evidence"),
    render: (row) =>
      row.attachments.length > 0 ? (
        <StatusChip tone="info">{supplierRejectionCopy("attachments.fileCount", { count: row.attachments.length })}</StatusChip>
      ) : (
        "-"
      ),
    width: "120px"
  },
  {
    key: "updated",
    header: supplierRejectionCopy("columns.updated"),
    render: (row) => formatSupplierRejectionDateTime(row.updatedAt),
    width: "150px"
  }
];

const auditColumns: DataTableColumn<AuditLogItem>[] = [
  {
    key: "action",
    header: supplierRejectionCopy("audit.columns.action"),
    render: (row) => <StatusChip tone={auditActionTone(row.action)}>{row.action}</StatusChip>
  },
  {
    key: "actor",
    header: supplierRejectionCopy("audit.columns.actor"),
    render: (row) => row.actorId,
    width: "170px"
  },
  {
    key: "payload",
    header: supplierRejectionCopy("audit.columns.change"),
    render: (row) => compactAuditPayload(row.afterData),
    width: "240px"
  },
  {
    key: "time",
    header: supplierRejectionCopy("audit.columns.time"),
    render: (row) => formatSupplierRejectionDateTime(row.createdAt),
    width: "150px"
  }
];

export function SupplierRejectionPanel() {
  const [warehouseId, setWarehouseId] = useState("wh-hcm-fg");
  const [status, setStatus] = useState<StatusFilter>("");
  const [supplierId, setSupplierId] = useState("supplier-local");
  const [selectedSampleLabel, setSelectedSampleLabel] = useState(supplierRejectionSampleLines[0].label);
  const [rejectedQuantity, setRejectedQuantity] = useState(supplierRejectionSampleLines[0].rejectedQuantity);
  const [reason, setReason] = useState(supplierRejectionSampleLines[0].reason);
  const [attachmentFileName, setAttachmentFileName] = useState("damage-photo.jpg");
  const [localRejections, setLocalRejections] = useState<SupplierRejection[]>([]);
  const [selectedRejectionId, setSelectedRejectionId] = useState("");
  const [feedback, setFeedback] = useState<{ tone: StatusTone; message: string } | null>(null);
  const [busyAction, setBusyAction] = useState("");
  const [auditLogs, setAuditLogs] = useState<AuditLogItem[]>([]);
  const [auditLoading, setAuditLoading] = useState(false);
  const [auditError, setAuditError] = useState<string | null>(null);
  const [auditRefreshKey, setAuditRefreshKey] = useState(0);
  const query = useMemo<SupplierRejectionQuery>(
    () => ({
      warehouseId,
      supplierId: supplierId || undefined,
      status: status || undefined
    }),
    [status, supplierId, warehouseId]
  );
  const { rejections, loading, error } = useSupplierRejections(query);
  const visibleRejections = useMemo(
    () => mergeSupplierRejections(localRejections, rejections, query),
    [localRejections, query, rejections]
  );
  const selectedSample =
    supplierRejectionSampleLines.find((line) => line.label === selectedSampleLabel) ?? supplierRejectionSampleLines[0];
  const selectedSupplier =
    supplierRejectionSupplierOptions.find((supplier) => supplier.value === supplierId) ?? supplierRejectionSupplierOptions[0];
  const selectedWarehouse =
    supplierRejectionWarehouseOptions.find((warehouse) => warehouse.value === warehouseId) ?? supplierRejectionWarehouseOptions[0];
  const selectedRejection =
    visibleRejections.find((rejection) => rejection.id === selectedRejectionId) ?? visibleRejections[0] ?? null;
  const totals = summarizeSupplierRejections(visibleRejections);
  const selectedAttachmentItems = useMemo<AttachmentPanelItem[]>(
    () =>
      selectedRejection?.attachments.map((attachment) => ({
        id: attachment.id,
        name: attachment.fileName,
        kind: supplierRejectionAttachmentKindLabel(attachment.contentType),
        uploadedBy: attachment.uploadedBy ?? selectedRejection.updatedBy,
        uploadedAt: attachment.uploadedAt ?? selectedRejection.updatedAt,
        storageKey: attachment.objectKey,
        status: <StatusChip tone="info">{supplierRejectionAttachmentSourceLabel(attachment.source)}</StatusChip>,
        canDownload: true,
        canDelete: selectedRejection.status === "draft",
        deleteLabel: supplierRejectionCopy("attachments.remove"),
        onDownload: () => setFeedback({ tone: "info", message: attachment.objectKey }),
        onDelete: () => removeSelectedAttachment(attachment.id)
      })) ?? [],
    [selectedRejection]
  );

  useEffect(() => {
    setRejectedQuantity(selectedSample.rejectedQuantity);
    setReason(selectedSample.reason);
  }, [selectedSample.label]);

  useEffect(() => {
    if (!selectedRejection) {
      setSelectedRejectionId("");
      setAuditLogs([]);
      return;
    }
    if (selectedRejection.id !== selectedRejectionId) {
      setSelectedRejectionId(selectedRejection.id);
    }
  }, [selectedRejection?.id, selectedRejectionId]);

  useEffect(() => {
    if (!selectedRejection?.id) {
      return;
    }
    let active = true;
    setAuditLoading(true);
    setAuditError(null);

    getAuditLogs({
      entityType: "inventory.supplier_rejection",
      entityId: selectedRejection.id,
      limit: 10
    })
      .then((items) => {
        if (active) {
          setAuditLogs(items);
        }
      })
      .catch((cause) => {
        if (active) {
          setAuditError(cause instanceof Error ? cause.message : supplierRejectionCopy("feedback.auditLoadFailed"));
        }
      })
      .finally(() => {
        if (active) {
          setAuditLoading(false);
        }
      });

    return () => {
      active = false;
    };
  }, [auditRefreshKey, selectedRejection?.id, selectedRejection?.updatedAt]);

  async function handleCreate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (busyAction) {
      return;
    }

    setBusyAction("create");
    setFeedback(null);
    try {
      const line = {
        ...selectedSample,
        rejectedQuantity,
        reason
      };
      const attachment =
        attachmentFileName.trim() === ""
          ? []
          : [
              {
                fileName: attachmentFileName,
                contentType: attachmentFileName.toLowerCase().endsWith(".pdf") ? "application/pdf" : "image/jpeg",
                source: "inbound_qc"
              }
            ];
      const rejection = await createSupplierRejection({
        supplierId: selectedSupplier.value,
        supplierCode: selectedSupplier.code,
        supplierName: selectedSupplier.name,
        purchaseOrderId: selectedSample.purchaseOrderId,
        purchaseOrderNo: selectedSample.purchaseOrderNo,
        goodsReceiptId: selectedSample.goodsReceiptId,
        goodsReceiptNo: selectedSample.goodsReceiptNo,
        inboundQCInspectionId: selectedSample.inboundQCInspectionId,
        warehouseId: selectedWarehouse.value,
        warehouseCode: selectedWarehouse.code,
        reason,
        lines: [line],
        attachments: attachment
      });

      setLocalRejections((current) => upsertSupplierRejection(current, rejection));
      setSelectedRejectionId(rejection.id);
      setAuditRefreshKey((current) => current + 1);
      setFeedback({ tone: "success", message: supplierRejectionCopy("feedback.created", { rejectionNo: rejection.rejectionNo }) });
    } catch (cause) {
      setFeedback({ tone: "danger", message: cause instanceof Error ? cause.message : supplierRejectionCopy("feedback.createFailed") });
    } finally {
      setBusyAction("");
    }
  }

  async function handleAction(action: "submit" | "confirm") {
    if (!selectedRejection || busyAction) {
      return;
    }

    setBusyAction(action);
    setFeedback(null);
    try {
      const result =
        action === "submit"
          ? await submitSupplierRejection(selectedRejection.id)
          : await confirmSupplierRejection(selectedRejection.id);

      setLocalRejections((current) => upsertSupplierRejection(current, result.rejection));
      setSelectedRejectionId(result.rejection.id);
      setAuditRefreshKey((current) => current + 1);
      setFeedback({
        tone: "success",
        message: supplierRejectionCopy("feedback.transitioned", {
          rejectionNo: result.rejection.rejectionNo,
          status: supplierRejectionStatusLabel(result.currentStatus).toLowerCase()
        })
      });
    } catch (cause) {
      setFeedback({ tone: "danger", message: cause instanceof Error ? cause.message : supplierRejectionCopy("feedback.updateFailed") });
    } finally {
      setBusyAction("");
    }
  }

  function removeSelectedAttachment(attachmentId: string) {
    if (!selectedRejection || selectedRejection.status !== "draft") {
      return;
    }

    const updated = {
      ...selectedRejection,
      attachments: selectedRejection.attachments.filter((attachment) => attachment.id !== attachmentId)
    };
    setLocalRejections((current) => upsertSupplierRejection(current, updated));
    setFeedback({ tone: "warning", message: supplierRejectionCopy("feedback.attachmentRemoved") });
  }

  return (
    <section className="erp-supplier-rejection-panel" id="supplier-rejections">
      <div className="erp-card erp-card--padded erp-supplier-rejection-form-card">
        <div className="erp-section-header">
          <h2 className="erp-section-title">{supplierRejectionCopy("title")}</h2>
          <StatusChip tone={feedback?.tone ?? "info"}>{feedback?.message ?? selectedWarehouse.code}</StatusChip>
        </div>

        <form className="erp-supplier-rejection-form" onSubmit={(event) => void handleCreate(event)}>
          <label className="erp-field">
            <span>{supplierRejectionCopy("fields.warehouse")}</span>
            <select className="erp-input" value={warehouseId} onChange={(event) => setWarehouseId(event.target.value)}>
              {supplierRejectionWarehouseOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>
          <label className="erp-field">
            <span>{supplierRejectionCopy("fields.supplier")}</span>
            <select className="erp-input" value={supplierId} onChange={(event) => setSupplierId(event.target.value)}>
              {supplierRejectionSupplierOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>
          <label className="erp-field">
            <span>{supplierRejectionCopy("fields.rejectedLine")}</span>
            <select className="erp-input" value={selectedSampleLabel} onChange={(event) => setSelectedSampleLabel(event.target.value)}>
              {supplierRejectionSampleLines.map((option) => (
                <option key={option.label} value={option.label}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>
          <label className="erp-field">
            <span>{supplierRejectionCopy("fields.rejectedQuantity")}</span>
            <input
              className="erp-input"
              inputMode="decimal"
              type="text"
              value={rejectedQuantity}
              onChange={(event) => setRejectedQuantity(event.target.value)}
            />
          </label>
          <label className="erp-field">
            <span>{supplierRejectionCopy("fields.reason")}</span>
            <input className="erp-input" type="text" value={reason} onChange={(event) => setReason(event.target.value)} />
          </label>
          <label className="erp-field">
            <span>{supplierRejectionCopy("fields.evidenceFile")}</span>
            <input
              className="erp-input"
              type="text"
              value={attachmentFileName}
              onChange={(event) => setAttachmentFileName(event.target.value)}
            />
          </label>
          <footer className="erp-supplier-rejection-actions">
            <button className="erp-button erp-button--primary" disabled={busyAction === "create"} type="submit">
              {supplierRejectionCopy("actions.create")}
            </button>
            <button
              className="erp-button erp-button--secondary"
              disabled={!selectedRejection || selectedRejection.status !== "draft" || busyAction !== ""}
              type="button"
              onClick={() => void handleAction("submit")}
            >
              {supplierRejectionCopy("actions.submit")}
            </button>
            <button
              className="erp-button erp-button--secondary"
              disabled={!selectedRejection || selectedRejection.status !== "submitted" || busyAction !== ""}
              type="button"
              onClick={() => void handleAction("confirm")}
            >
              {supplierRejectionCopy("actions.confirm")}
            </button>
          </footer>
        </form>
      </div>

      <div className="erp-card erp-card--padded erp-supplier-rejection-detail-card">
        <div className="erp-section-header">
          <h2 className="erp-section-title">{supplierRejectionCopy("detail.title")}</h2>
          {selectedRejection ? (
            <StatusChip tone={supplierRejectionStatusTone(selectedRejection.status)}>
              {supplierRejectionStatusLabel(selectedRejection.status)}
            </StatusChip>
          ) : null}
        </div>
        {selectedRejection ? (
          <>
            <div className="erp-supplier-rejection-facts">
              <SupplierRejectFact label={supplierRejectionCopy("facts.supplier")} value={selectedRejection.supplierName ?? selectedRejection.supplierId} />
              <SupplierRejectFact label="PO" value={selectedRejection.purchaseOrderNo ?? selectedRejection.purchaseOrderId ?? "-"} />
              <SupplierRejectFact label={supplierRejectionCopy("facts.goodsReceipt")} value={selectedRejection.goodsReceiptNo ?? selectedRejection.goodsReceiptId} />
              <SupplierRejectFact label={supplierRejectionCopy("facts.qcInspection")} value={selectedRejection.inboundQCInspectionId} />
              <SupplierRejectFact label={supplierRejectionCopy("facts.batch")} value={selectedRejection.lines[0]?.lotNo ?? "-"} />
              <SupplierRejectFact
                label={supplierRejectionCopy("facts.rejected")}
                value={formatSupplierRejectionQuantity(totalRejectedQuantity(selectedRejection), selectedRejection.lines[0]?.baseUOMCode)}
              />
            </div>
            <AttachmentPanel
              title={supplierRejectionCopy("attachments.title")}
              items={selectedAttachmentItems}
              emptyMessage={supplierRejectionCopy("attachments.empty")}
            />
          </>
        ) : (
          <>
            <EmptyState title={supplierRejectionCopy("empty.noSelection")} />
            <AttachmentPanel
              title={supplierRejectionCopy("attachments.title")}
              items={[]}
              emptyMessage={supplierRejectionCopy("attachments.empty")}
            />
          </>
        )}
      </div>

      <section className="erp-card erp-card--padded erp-module-table-card erp-supplier-rejection-table-card">
        <div className="erp-section-header">
          <h2 className="erp-section-title">{supplierRejectionCopy("list.title")}</h2>
          <StatusChip tone={totals.confirmed > 0 ? "success" : "warning"}>
            {supplierRejectionCopy("list.summary", { total: totals.total, confirmed: totals.confirmed })}
          </StatusChip>
        </div>
        <DataTable
          columns={[
            ...rejectionColumns,
            {
              key: "open",
              header: supplierRejectionCopy("columns.action"),
              render: (row) => (
                <button className="erp-button erp-button--secondary" type="button" onClick={() => setSelectedRejectionId(row.id)}>
                  {supplierRejectionCopy("actions.open")}
                </button>
              ),
              width: "96px",
              sticky: true
            }
          ]}
          rows={visibleRejections}
          getRowKey={(row) => row.id}
          loading={loading}
          error={error?.message}
          emptyState={<EmptyState title={supplierRejectionCopy("empty.noRows")} description={supplierRejectionCopy("empty.noRowsDescription")} />}
          toolbar={
            <div className="erp-supplier-rejection-table-toolbar">
              <label className="erp-field">
                <span>{supplierRejectionCopy("fields.status")}</span>
                <select className="erp-input" value={status} onChange={(event) => setStatus(event.target.value as StatusFilter)}>
                  {supplierRejectionStatusOptions.map((option) => (
                    <option key={option.value || "all"} value={option.value}>
                      {option.value ? supplierRejectionStatusLabel(option.value) : supplierRejectionCopy("filters.allStatuses")}
                    </option>
                  ))}
                </select>
              </label>
            </div>
          }
        />
      </section>

      <section className="erp-card erp-card--padded erp-module-table-card erp-supplier-rejection-audit-card">
        <div className="erp-section-header">
          <h2 className="erp-section-title">{supplierRejectionCopy("audit.title")}</h2>
          <StatusChip tone={auditLogs.length > 0 ? "info" : "warning"}>
            {supplierRejectionCopy("audit.events", { count: auditLogs.length })}
          </StatusChip>
        </div>
        <DataTable
          columns={auditColumns}
          rows={auditLogs}
          getRowKey={(row) => row.id}
          loading={auditLoading}
          error={auditError ?? undefined}
          emptyState={<EmptyState title={supplierRejectionCopy("audit.emptyTitle")} description={supplierRejectionCopy("audit.emptyDescription")} />}
        />
      </section>
    </section>
  );
}

function supplierRejectionCopy(key: string, values?: Record<string, string | number>, fallback?: string) {
  return t(`returns.supplierRejection.${key}`, { values, fallback });
}

function supplierRejectionStatusLabel(status: SupplierRejectionStatus) {
  return supplierRejectionCopy(`status.${status}`);
}

function supplierRejectionAttachmentSourceLabel(source: string | undefined) {
  if (!source) {
    return supplierRejectionCopy("attachments.source.rejection");
  }

  return supplierRejectionCopy(`attachments.source.${source}`, undefined, source);
}

function supplierRejectionAttachmentKindLabel(contentType: string | undefined) {
  if (!contentType) {
    return supplierRejectionCopy("attachments.evidence");
  }
  if (contentType.toLowerCase().startsWith("image/")) {
    return supplierRejectionCopy("attachments.kind.image");
  }
  if (contentType.toLowerCase() === "application/pdf") {
    return supplierRejectionCopy("attachments.kind.pdf");
  }

  return supplierRejectionCopy("attachments.evidence");
}

function SupplierRejectFact({ label, value }: { label: string; value: string }) {
  return (
    <div className="erp-returns-fact">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function mergeSupplierRejections(
  localRejections: SupplierRejection[],
  remoteRejections: SupplierRejection[],
  query: SupplierRejectionQuery
) {
  const localMatches = localRejections.filter((rejection) => matchesSupplierRejectionQuery(rejection, query));
  const localIds = new Set(localMatches.map((rejection) => rejection.id));

  return [...localMatches, ...remoteRejections.filter((rejection) => !localIds.has(rejection.id))].sort((left, right) =>
    right.updatedAt.localeCompare(left.updatedAt)
  );
}

function matchesSupplierRejectionQuery(rejection: SupplierRejection, query: SupplierRejectionQuery) {
  if (query.supplierId && rejection.supplierId !== query.supplierId) {
    return false;
  }
  if (query.warehouseId && rejection.warehouseId !== query.warehouseId) {
    return false;
  }
  if (query.status && rejection.status !== query.status) {
    return false;
  }

  return true;
}

function upsertSupplierRejection(rows: SupplierRejection[], updated: SupplierRejection) {
  return [updated, ...rows.filter((row) => row.id !== updated.id)];
}

function summarizeSupplierRejections(rows: SupplierRejection[]) {
  return rows.reduce(
    (acc, row) => ({
      total: acc.total + 1,
      draft: acc.draft + (row.status === "draft" ? 1 : 0),
      submitted: acc.submitted + (row.status === "submitted" ? 1 : 0),
      confirmed: acc.confirmed + (row.status === "confirmed" ? 1 : 0)
    }),
    { total: 0, draft: 0, submitted: 0, confirmed: 0 }
  );
}

function totalRejectedQuantity(rejection: SupplierRejection) {
  const total = rejection.lines.reduce((sum, line) => sum + Number(line.rejectedQuantity || "0"), 0);
  return total.toFixed(6);
}
