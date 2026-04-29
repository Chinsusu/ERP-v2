"use client";

import { useEffect, useMemo, useState, type FormEvent } from "react";
import { DataTable, EmptyState, StatusChip, type DataTableColumn, type StatusTone } from "@/shared/design-system/components";
import { getAuditLogs, auditActionTone, compactAuditPayload } from "@/modules/audit/services/auditLogService";
import type { AuditLogItem } from "@/modules/audit/types";
import { useSupplierRejections } from "../hooks/useSupplierRejections";
import {
  confirmSupplierRejection,
  createSupplierRejection,
  formatSupplierRejectionDateTime,
  formatSupplierRejectionQuantity,
  formatSupplierRejectionStatus,
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
    header: "Reject record",
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
    header: "Supplier",
    render: (row) => row.supplierCode ?? row.supplierName ?? row.supplierId,
    width: "160px"
  },
  {
    key: "status",
    header: "Status",
    render: (row) => <StatusChip tone={supplierRejectionStatusTone(row.status)}>{formatSupplierRejectionStatus(row.status)}</StatusChip>,
    width: "130px"
  },
  {
    key: "qty",
    header: "Rejected qty",
    render: (row) => formatSupplierRejectionQuantity(totalRejectedQuantity(row), row.lines[0]?.baseUOMCode),
    align: "right",
    width: "140px"
  },
  {
    key: "batch",
    header: "Batch / reason",
    render: (row) => (
      <span className="erp-supplier-rejection-record-cell">
        <strong>{row.lines[0]?.lotNo ?? "-"}</strong>
        <small>{row.reason}</small>
      </span>
    )
  },
  {
    key: "attachments",
    header: "Evidence",
    render: (row) => (row.attachments.length > 0 ? <StatusChip tone="info">{row.attachments.length} files</StatusChip> : "-"),
    width: "120px"
  },
  {
    key: "updated",
    header: "Updated",
    render: (row) => formatSupplierRejectionDateTime(row.updatedAt),
    width: "150px"
  }
];

const auditColumns: DataTableColumn<AuditLogItem>[] = [
  {
    key: "action",
    header: "Action",
    render: (row) => <StatusChip tone={auditActionTone(row.action)}>{row.action}</StatusChip>
  },
  {
    key: "actor",
    header: "Actor",
    render: (row) => row.actorId,
    width: "170px"
  },
  {
    key: "payload",
    header: "Change",
    render: (row) => compactAuditPayload(row.afterData),
    width: "240px"
  },
  {
    key: "time",
    header: "Time",
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
          setAuditError(cause instanceof Error ? cause.message : "Audit trail could not be loaded");
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
      setFeedback({ tone: "success", message: `${rejection.rejectionNo} created` });
    } catch (cause) {
      setFeedback({ tone: "danger", message: cause instanceof Error ? cause.message : "Supplier rejection could not be created" });
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
        message: `${result.rejection.rejectionNo} ${formatSupplierRejectionStatus(result.currentStatus).toLowerCase()}`
      });
    } catch (cause) {
      setFeedback({ tone: "danger", message: cause instanceof Error ? cause.message : "Supplier rejection could not be updated" });
    } finally {
      setBusyAction("");
    }
  }

  return (
    <section className="erp-supplier-rejection-panel" id="supplier-rejections">
      <div className="erp-card erp-card--padded erp-supplier-rejection-form-card">
        <div className="erp-section-header">
          <h2 className="erp-section-title">Supplier rejection</h2>
          <StatusChip tone={feedback?.tone ?? "info"}>{feedback?.message ?? selectedWarehouse.code}</StatusChip>
        </div>

        <form className="erp-supplier-rejection-form" onSubmit={(event) => void handleCreate(event)}>
          <label className="erp-field">
            <span>Warehouse</span>
            <select className="erp-input" value={warehouseId} onChange={(event) => setWarehouseId(event.target.value)}>
              {supplierRejectionWarehouseOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>
          <label className="erp-field">
            <span>Supplier</span>
            <select className="erp-input" value={supplierId} onChange={(event) => setSupplierId(event.target.value)}>
              {supplierRejectionSupplierOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>
          <label className="erp-field">
            <span>Rejected inbound line</span>
            <select className="erp-input" value={selectedSampleLabel} onChange={(event) => setSelectedSampleLabel(event.target.value)}>
              {supplierRejectionSampleLines.map((option) => (
                <option key={option.label} value={option.label}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>
          <label className="erp-field">
            <span>Rejected quantity</span>
            <input
              className="erp-input"
              inputMode="decimal"
              type="text"
              value={rejectedQuantity}
              onChange={(event) => setRejectedQuantity(event.target.value)}
            />
          </label>
          <label className="erp-field">
            <span>Reason</span>
            <input className="erp-input" type="text" value={reason} onChange={(event) => setReason(event.target.value)} />
          </label>
          <label className="erp-field">
            <span>Evidence file</span>
            <input
              className="erp-input"
              type="text"
              value={attachmentFileName}
              onChange={(event) => setAttachmentFileName(event.target.value)}
            />
          </label>
          <footer className="erp-supplier-rejection-actions">
            <button className="erp-button erp-button--primary" disabled={busyAction === "create"} type="submit">
              Create reject
            </button>
            <button
              className="erp-button erp-button--secondary"
              disabled={!selectedRejection || selectedRejection.status !== "draft" || busyAction !== ""}
              type="button"
              onClick={() => void handleAction("submit")}
            >
              Submit
            </button>
            <button
              className="erp-button erp-button--secondary"
              disabled={!selectedRejection || selectedRejection.status !== "submitted" || busyAction !== ""}
              type="button"
              onClick={() => void handleAction("confirm")}
            >
              Confirm
            </button>
          </footer>
        </form>
      </div>

      <div className="erp-card erp-card--padded erp-supplier-rejection-detail-card">
        <div className="erp-section-header">
          <h2 className="erp-section-title">Reject detail</h2>
          {selectedRejection ? (
            <StatusChip tone={supplierRejectionStatusTone(selectedRejection.status)}>
              {formatSupplierRejectionStatus(selectedRejection.status)}
            </StatusChip>
          ) : null}
        </div>
        {selectedRejection ? (
          <>
            <div className="erp-supplier-rejection-facts">
              <SupplierRejectFact label="Supplier" value={selectedRejection.supplierName ?? selectedRejection.supplierId} />
              <SupplierRejectFact label="PO" value={selectedRejection.purchaseOrderNo ?? selectedRejection.purchaseOrderId ?? "-"} />
              <SupplierRejectFact label="Goods receipt" value={selectedRejection.goodsReceiptNo ?? selectedRejection.goodsReceiptId} />
              <SupplierRejectFact label="QC inspection" value={selectedRejection.inboundQCInspectionId} />
              <SupplierRejectFact label="Batch" value={selectedRejection.lines[0]?.lotNo ?? "-"} />
              <SupplierRejectFact
                label="Rejected"
                value={formatSupplierRejectionQuantity(totalRejectedQuantity(selectedRejection), selectedRejection.lines[0]?.baseUOMCode)}
              />
            </div>
            <div className="erp-supplier-rejection-evidence">
              <strong>Evidence</strong>
              {selectedRejection.attachments.length > 0 ? (
                selectedRejection.attachments.map((attachment) => (
                  <span key={attachment.id}>
                    {attachment.fileName}
                    <small>{attachment.objectKey}</small>
                  </span>
                ))
              ) : (
                <small>No attachment metadata</small>
              )}
            </div>
          </>
        ) : (
          <EmptyState title="No supplier rejection selected" />
        )}
      </div>

      <section className="erp-card erp-card--padded erp-module-table-card erp-supplier-rejection-table-card">
        <div className="erp-section-header">
          <h2 className="erp-section-title">Rejected inbound goods</h2>
          <StatusChip tone={totals.confirmed > 0 ? "success" : "warning"}>
            {totals.total} rows / {totals.confirmed} confirmed
          </StatusChip>
        </div>
        <DataTable
          columns={[
            ...rejectionColumns,
            {
              key: "open",
              header: "Action",
              render: (row) => (
                <button className="erp-button erp-button--secondary" type="button" onClick={() => setSelectedRejectionId(row.id)}>
                  Open
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
          emptyState={<EmptyState title="No rejected inbound goods" description="Create a supplier rejection from a failed QC line." />}
          toolbar={
            <div className="erp-supplier-rejection-table-toolbar">
              <label className="erp-field">
                <span>Status</span>
                <select className="erp-input" value={status} onChange={(event) => setStatus(event.target.value as StatusFilter)}>
                  {supplierRejectionStatusOptions.map((option) => (
                    <option key={option.value || "all"} value={option.value}>
                      {option.label}
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
          <h2 className="erp-section-title">Audit trail</h2>
          <StatusChip tone={auditLogs.length > 0 ? "info" : "warning"}>{auditLogs.length} events</StatusChip>
        </div>
        <DataTable
          columns={auditColumns}
          rows={auditLogs}
          getRowKey={(row) => row.id}
          loading={auditLoading}
          error={auditError ?? undefined}
          emptyState={<EmptyState title="No audit events loaded" description="Create, submit, or confirm a supplier rejection to populate audit." />}
        />
      </section>
    </section>
  );
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
