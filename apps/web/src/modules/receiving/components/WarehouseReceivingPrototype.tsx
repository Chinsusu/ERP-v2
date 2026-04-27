"use client";

import { useMemo, useState, type FormEvent } from "react";
import {
  ConfirmDialog,
  DataTable,
  EmptyState,
  StatusChip,
  type DataTableColumn,
  type StatusTone
} from "@/shared/design-system/components";
import { useGoodsReceipts } from "../hooks/useGoodsReceipts";
import {
  createGoodsReceipt,
  formatGoodsReceiptStatus,
  formatQCStatus,
  formatReceivingDateTime,
  formatReceivingQuantity,
  goodsReceiptStatusTone,
  markGoodsReceiptInspectReady,
  postGoodsReceipt,
  qcStatusTone,
  receivingBatchOptions,
  receivingLocationOptions,
  receivingWarehouseOptions,
  submitGoodsReceipt
} from "../services/warehouseReceivingService";
import type { GoodsReceipt, GoodsReceiptLine, GoodsReceiptQuery, GoodsReceiptStatus, GoodsReceiptStockMovement } from "../types";

type StatusFilter = "" | GoodsReceiptStatus;

const statusOptions: { label: string; value: StatusFilter }[] = [
  { label: "All statuses", value: "" },
  { label: "Draft", value: "draft" },
  { label: "Submitted", value: "submitted" },
  { label: "Inspect ready", value: "inspect_ready" },
  { label: "Posted", value: "posted" }
];

const receiptColumns: DataTableColumn<GoodsReceipt>[] = [
  {
    key: "receipt",
    header: "Receipt",
    render: (row) => (
      <span className="erp-receiving-receipt-cell">
        <strong>{row.receiptNo}</strong>
        <small>{row.referenceDocId}</small>
      </span>
    ),
    width: "190px"
  },
  {
    key: "warehouse",
    header: "Warehouse",
    render: (row) => row.warehouseCode,
    width: "150px"
  },
  {
    key: "location",
    header: "Location",
    render: (row) => row.locationCode,
    width: "140px"
  },
  {
    key: "status",
    header: "Status",
    render: (row) => <StatusChip tone={goodsReceiptStatusTone(row.status)}>{formatGoodsReceiptStatus(row.status)}</StatusChip>,
    width: "150px"
  },
  {
    key: "sku",
    header: "SKU",
    render: (row) => row.lines[0]?.sku ?? "-",
    width: "160px"
  },
  {
    key: "quantity",
    header: "Quantity",
    render: (row) => formatReceivingQuantity(row.lines[0]?.quantity ?? "0", row.lines[0]?.baseUomCode),
    align: "right",
    width: "140px"
  },
  {
    key: "movement",
    header: "Movement",
    render: (row) => (row.stockMovements?.length ? <StatusChip tone="success">{row.stockMovements.length} posted</StatusChip> : "-"),
    width: "130px"
  }
];

const lineColumns: DataTableColumn<GoodsReceiptLine>[] = [
  {
    key: "sku",
    header: "SKU",
    render: (row) => (
      <span className="erp-receiving-receipt-cell">
        <strong>{row.sku}</strong>
        <small>{row.itemName ?? row.itemId}</small>
      </span>
    )
  },
  {
    key: "batch",
    header: "Batch",
    render: (row) => row.batchNo ?? row.batchId ?? "-",
    width: "160px"
  },
  {
    key: "qc",
    header: "QC",
    render: (row) => <StatusChip tone={qcStatusTone(row.qcStatus)}>{formatQCStatus(row.qcStatus)}</StatusChip>,
    width: "130px"
  },
  {
    key: "quantity",
    header: "Quantity",
    render: (row) => formatReceivingQuantity(row.quantity, row.baseUomCode),
    align: "right",
    width: "130px"
  }
];

const movementColumns: DataTableColumn<GoodsReceiptStockMovement>[] = [
  {
    key: "movement",
    header: "Movement",
    render: (row) => row.movementNo
  },
  {
    key: "status",
    header: "Stock",
    render: (row) => <StatusChip tone={row.stockStatus === "available" ? "success" : "warning"}>{row.stockStatus}</StatusChip>,
    width: "120px"
  },
  {
    key: "quantity",
    header: "Quantity",
    render: (row) => formatReceivingQuantity(row.quantity, row.baseUomCode),
    align: "right",
    width: "130px"
  }
];

export function WarehouseReceivingPrototype() {
  const [warehouseId, setWarehouseId] = useState("wh-hcm-fg");
  const [status, setStatus] = useState<StatusFilter>("");
  const [locationId, setLocationId] = useState("loc-hcm-fg-recv-01");
  const [referenceDocId, setReferenceDocId] = useState("PO-260427-UI");
  const [batchId, setBatchId] = useState("batch-cream-2603b");
  const [quantity, setQuantity] = useState("12");
  const [localReceipts, setLocalReceipts] = useState<GoodsReceipt[]>([]);
  const [selectedReceiptId, setSelectedReceiptId] = useState("grn-hcm-260427-inspect");
  const [feedback, setFeedback] = useState<{ tone: StatusTone; message: string } | null>(null);
  const [busyAction, setBusyAction] = useState("");
  const [confirmPostId, setConfirmPostId] = useState<string | null>(null);
  const query = useMemo<GoodsReceiptQuery>(
    () => ({
      warehouseId: warehouseId || undefined,
      status: status || undefined
    }),
    [status, warehouseId]
  );
  const { receipts, loading, error } = useGoodsReceipts(query);
  const visibleReceipts = useMemo(() => mergeReceipts(localReceipts, receipts, query), [localReceipts, query, receipts]);
  const selectedReceipt = visibleReceipts.find((receipt) => receipt.id === selectedReceiptId) ?? visibleReceipts[0] ?? null;
  const selectedBatch = receivingBatchOptions.find((batch) => batch.value === batchId) ?? receivingBatchOptions[0];
  const locationOptions = receivingLocationOptions.filter((location) => location.warehouseId === warehouseId);
  const totals = summarizeReceipts(visibleReceipts);

  function handleWarehouseChange(nextWarehouseId: string) {
    setWarehouseId(nextWarehouseId);
    const nextLocation = receivingLocationOptions.find((location) => location.warehouseId === nextWarehouseId);
    setLocationId(nextLocation?.value ?? "");
  }

  async function handleCreateDraft(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (busyAction) {
      return;
    }

    setBusyAction("create");
    setFeedback(null);
    try {
      const receipt = await createGoodsReceipt({
        warehouseId,
        locationId,
        referenceDocType: "purchase_order",
        referenceDocId,
        lines: [
          {
            id: "line-ui-001",
            batchId,
            quantity,
            baseUomCode: selectedBatch.baseUomCode
          }
        ]
      });
      upsertLocalReceipt(receipt);
      setSelectedReceiptId(receipt.id);
      setFeedback({ tone: "success", message: `${receipt.receiptNo} created` });
    } catch (reason) {
      setFeedback({ tone: "danger", message: reason instanceof Error ? reason.message : "Goods receipt could not be created" });
    } finally {
      setBusyAction("");
    }
  }

  async function runAction(receipt: GoodsReceipt, action: "submit" | "inspect" | "post") {
    if (busyAction) {
      return;
    }

    setBusyAction(`${action}:${receipt.id}`);
    setFeedback(null);
    try {
      const updated =
        action === "submit"
          ? await submitGoodsReceipt(receipt.id)
          : action === "inspect"
            ? await markGoodsReceiptInspectReady(receipt.id)
            : await postGoodsReceipt(receipt.id);
      upsertLocalReceipt(updated);
      setSelectedReceiptId(updated.id);
      setFeedback({ tone: "success", message: `${updated.receiptNo} / ${formatGoodsReceiptStatus(updated.status)}` });
    } catch (reason) {
      setFeedback({ tone: "danger", message: reason instanceof Error ? reason.message : "Goods receipt action failed" });
    } finally {
      setBusyAction("");
      setConfirmPostId(null);
    }
  }

  function upsertLocalReceipt(receipt: GoodsReceipt) {
    setLocalReceipts((current) => [receipt, ...current.filter((candidate) => candidate.id !== receipt.id)]);
  }

  return (
    <section className="erp-module-page erp-receiving-page">
      <header className="erp-page-header">
        <div>
          <p className="erp-module-eyebrow">RC</p>
          <h1 className="erp-page-title">Warehouse Receiving</h1>
          <p className="erp-page-description">Goods receipt draft, inspection handoff, posting, and movement evidence</p>
        </div>
        <div className="erp-page-actions">
          <a className="erp-button erp-button--secondary" href="#receiving-draft">
            Draft
          </a>
          <a className="erp-button erp-button--secondary" href="#receiving-detail">
            Detail
          </a>
          <a className="erp-button erp-button--primary" href="#receiving-list">
            Receipts
          </a>
        </div>
      </header>

      <section className="erp-receiving-toolbar" aria-label="Goods receipt filters">
        <label className="erp-field">
          <span>Warehouse</span>
          <select className="erp-input" value={warehouseId} onChange={(event) => handleWarehouseChange(event.target.value)}>
            {receivingWarehouseOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Status</span>
          <select className="erp-input" value={status} onChange={(event) => setStatus(event.target.value as StatusFilter)}>
            {statusOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
      </section>

      <section className="erp-kpi-grid erp-receiving-kpis">
        <ReceivingKPI label="Draft" tone="warning" value={totals.draft} />
        <ReceivingKPI label="Submitted" tone="info" value={totals.submitted} />
        <ReceivingKPI label="Inspect ready" tone="warning" value={totals.inspectReady} />
        <ReceivingKPI label="Posted" tone="success" value={totals.posted} />
      </section>

      <section className="erp-receiving-workspace">
        <form className="erp-card erp-card--padded erp-receiving-draft" id="receiving-draft" onSubmit={handleCreateDraft}>
          <div className="erp-section-header">
            <h2 className="erp-section-title">New receipt draft</h2>
            <StatusChip tone={feedback?.tone ?? "info"}>{feedback?.message ?? "Ready"}</StatusChip>
          </div>

          <div className="erp-receiving-form-grid">
            <label className="erp-field">
              <span>Location</span>
              <select className="erp-input" value={locationId} onChange={(event) => setLocationId(event.target.value)}>
                {locationOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </label>
            <label className="erp-field">
              <span>Reference</span>
              <input
                className="erp-input"
                type="text"
                value={referenceDocId}
                onChange={(event) => setReferenceDocId(event.target.value.toUpperCase())}
                required
              />
            </label>
            <label className="erp-field">
              <span>Batch / SKU</span>
              <select className="erp-input" value={batchId} onChange={(event) => setBatchId(event.target.value)}>
                {receivingBatchOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </label>
            <label className="erp-field">
              <span>Quantity</span>
              <input
                className="erp-input"
                min="0.000001"
                step="0.000001"
                type="number"
                value={quantity}
                onChange={(event) => setQuantity(event.target.value)}
                required
              />
            </label>
          </div>

          <div className="erp-receiving-selected-line">
            <StatusChip tone={qcStatusTone(selectedBatch.qcStatus)}>{formatQCStatus(selectedBatch.qcStatus)}</StatusChip>
            <span>{selectedBatch.itemName}</span>
            <strong>{formatReceivingQuantity(quantity || "0", selectedBatch.baseUomCode)}</strong>
          </div>

          <div className="erp-receiving-actions">
            <button className="erp-button erp-button--primary" type="submit" disabled={busyAction === "create"}>
              {busyAction === "create" ? "Creating" : "Create draft"}
            </button>
          </div>
        </form>

        <section className="erp-card erp-card--padded erp-receiving-detail" id="receiving-detail">
          <div className="erp-section-header">
            <h2 className="erp-section-title">Receipt detail</h2>
            {selectedReceipt ? (
              <StatusChip tone={goodsReceiptStatusTone(selectedReceipt.status)}>
                {formatGoodsReceiptStatus(selectedReceipt.status)}
              </StatusChip>
            ) : null}
          </div>

          {selectedReceipt ? (
            <>
              <div className="erp-receiving-detail-grid">
                <ReceivingFact label="Receipt" value={selectedReceipt.receiptNo} />
                <ReceivingFact label="Reference" value={selectedReceipt.referenceDocId} />
                <ReceivingFact label="Warehouse" value={selectedReceipt.warehouseCode} />
                <ReceivingFact label="Location" value={selectedReceipt.locationCode} />
                <ReceivingFact label="Updated" value={formatReceivingDateTime(selectedReceipt.updatedAt)} />
                <ReceivingFact label="Audit" value={selectedReceipt.auditLogId ?? "-"} />
              </div>

              <div className="erp-receiving-actions">
                <button
                  className="erp-button erp-button--secondary"
                  type="button"
                  disabled={selectedReceipt.status !== "draft" || Boolean(busyAction)}
                  onClick={() => void runAction(selectedReceipt, "submit")}
                >
                  Submit
                </button>
                <button
                  className="erp-button erp-button--secondary"
                  type="button"
                  disabled={selectedReceipt.status !== "submitted" || Boolean(busyAction)}
                  onClick={() => void runAction(selectedReceipt, "inspect")}
                >
                  Inspect ready
                </button>
                <button
                  className="erp-button erp-button--primary"
                  type="button"
                  disabled={selectedReceipt.status !== "inspect_ready" || Boolean(busyAction)}
                  onClick={() => setConfirmPostId(selectedReceipt.id)}
                >
                  Post
                </button>
              </div>

              <div className="erp-receiving-subsection">
                <h3 className="erp-section-title">Lines</h3>
                <DataTable columns={lineColumns} rows={selectedReceipt.lines} getRowKey={(row) => row.id} />
              </div>

              <div className="erp-receiving-subsection">
                <h3 className="erp-section-title">Stock movements</h3>
                <DataTable
                  columns={movementColumns}
                  rows={selectedReceipt.stockMovements ?? []}
                  getRowKey={(row) => row.movementNo}
                  emptyState={<EmptyState title="No stock movement posted" description="Post the receipt after inspection." />}
                />
              </div>
            </>
          ) : (
            <EmptyState title="No goods receipt selected" description="Create or select a receipt." />
          )}
        </section>
      </section>

      <section className="erp-card erp-card--padded erp-module-table-card" id="receiving-list">
        <div className="erp-section-header">
          <h2 className="erp-section-title">Goods receipts</h2>
          <StatusChip tone={visibleReceipts.length === 0 ? "warning" : "info"}>{visibleReceipts.length} rows</StatusChip>
        </div>
        <DataTable
          columns={[
            ...receiptColumns,
            {
              key: "action",
              header: "Action",
              render: (row) => (
                <button
                  className="erp-button erp-button--secondary erp-button--compact"
                  type="button"
                  onClick={() => setSelectedReceiptId(row.id)}
                >
                  View
                </button>
              ),
              width: "110px"
            }
          ]}
          rows={visibleReceipts}
          getRowKey={(row) => row.id}
          loading={loading}
          error={error?.message}
          emptyState={<EmptyState title="No goods receipts" description="Change warehouse or status filter." />}
        />
      </section>

      <ConfirmDialog
        open={Boolean(confirmPostId && selectedReceipt)}
        title="Post goods receipt"
        description={`${selectedReceipt?.receiptNo ?? "Receipt"} will create stock movement records.`}
        confirmLabel="Post"
        onCancel={() => setConfirmPostId(null)}
        onConfirm={() => {
          if (selectedReceipt) {
            void runAction(selectedReceipt, "post");
          }
        }}
      />
    </section>
  );
}

function ReceivingKPI({
  label,
  value,
  tone
}: {
  label: string;
  value: number;
  tone: "normal" | "success" | "warning" | "danger" | "info";
}) {
  return (
    <article className="erp-card erp-card--padded erp-kpi-card">
      <div className="erp-kpi-label">{label}</div>
      <strong className="erp-kpi-value">{value}</strong>
      <StatusChip tone={tone}>{label}</StatusChip>
    </article>
  );
}

function ReceivingFact({ label, value }: { label: string; value: string }) {
  return (
    <div className="erp-receiving-fact">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function mergeReceipts(localReceipts: GoodsReceipt[], remoteReceipts: GoodsReceipt[], query: GoodsReceiptQuery) {
  const localMatches = localReceipts.filter((receipt) => matchesReceiptQuery(receipt, query));
  const localIds = new Set(localMatches.map((receipt) => receipt.id));

  return [...localMatches, ...remoteReceipts.filter((receipt) => !localIds.has(receipt.id))];
}

function matchesReceiptQuery(receipt: GoodsReceipt, query: GoodsReceiptQuery) {
  if (query.warehouseId && receipt.warehouseId !== query.warehouseId) {
    return false;
  }
  if (query.status && receipt.status !== query.status) {
    return false;
  }

  return true;
}

function summarizeReceipts(receipts: GoodsReceipt[]) {
  return receipts.reduce(
    (summary, receipt) => ({
      draft: summary.draft + (receipt.status === "draft" ? 1 : 0),
      submitted: summary.submitted + (receipt.status === "submitted" ? 1 : 0),
      inspectReady: summary.inspectReady + (receipt.status === "inspect_ready" ? 1 : 0),
      posted: summary.posted + (receipt.status === "posted" ? 1 : 0)
    }),
    { draft: 0, submitted: 0, inspectReady: 0, posted: 0 }
  );
}
