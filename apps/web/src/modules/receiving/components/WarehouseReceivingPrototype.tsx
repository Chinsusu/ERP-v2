"use client";

import { useEffect, useMemo, useState, type FormEvent } from "react";
import {
  ConfirmDialog,
  DataTable,
  EmptyState,
  StatusChip,
  type DataTableColumn,
  type StatusTone
} from "@/shared/design-system/components";
import { AttachmentPanel, type AttachmentPanelItem } from "@/shared/design-system/pageTemplates";
import { t } from "@/shared/i18n";
import {
  formatPurchaseQuantity,
  getPurchaseOrder,
  getPurchaseOrders,
  purchaseOrderStatusTone,
  purchaseSupplierOptions
} from "../../purchase/services/purchaseOrderService";
import type { PurchaseOrder, PurchaseOrderLine } from "../../purchase/types";
import { useGoodsReceipts } from "../hooks/useGoodsReceipts";
import {
  createGoodsReceipt,
  formatReceivingDateTime,
  formatReceivingQuantity,
  goodsReceiptStatusTone,
  markGoodsReceiptInspectReady,
  postGoodsReceipt,
  qcStatusTone,
  receivingBatchOptions,
  receivingLocationOptions,
  receivingPackagingStatusOptions,
  receivingWarehouseOptions,
  submitGoodsReceipt
} from "../services/warehouseReceivingService";
import type {
  BatchQCStatus,
  GoodsReceipt,
  GoodsReceiptLine,
  GoodsReceiptQuery,
  GoodsReceiptStatus,
  GoodsReceiptStockMovement,
  ReceivingPackagingStatus
} from "../types";

type StatusFilter = "" | GoodsReceiptStatus;

const statusOptions: { label: string; value: StatusFilter }[] = [
  { label: receivingCopy("filters.allStatuses"), value: "" },
  { label: goodsReceiptStatusLabel("draft"), value: "draft" },
  { label: goodsReceiptStatusLabel("submitted"), value: "submitted" },
  { label: goodsReceiptStatusLabel("inspect_ready"), value: "inspect_ready" },
  { label: goodsReceiptStatusLabel("posted"), value: "posted" }
];

const defaultBatch = receivingBatchOptions[1] ?? receivingBatchOptions[0];

const receiptColumns: DataTableColumn<GoodsReceipt>[] = [
  {
    key: "receipt",
    header: receivingCopy("columns.receipt"),
    render: (row) => (
      <span className="erp-receiving-receipt-cell">
        <strong>{row.receiptNo}</strong>
        <small>{row.deliveryNoteNo ?? row.referenceDocId}</small>
      </span>
    ),
    width: "190px"
  },
  {
    key: "reference",
    header: receivingCopy("columns.po"),
    render: (row) => row.referenceDocId,
    width: "170px"
  },
  {
    key: "warehouse",
    header: receivingCopy("columns.warehouse"),
    render: (row) => row.warehouseCode,
    width: "150px"
  },
  {
    key: "status",
    header: receivingCopy("columns.status"),
    render: (row) => <StatusChip tone={goodsReceiptStatusTone(row.status)}>{goodsReceiptStatusLabel(row.status)}</StatusChip>,
    width: "150px"
  },
  {
    key: "sku",
    header: receivingCopy("columns.sku"),
    render: (row) => row.lines[0]?.sku ?? "-",
    width: "160px"
  },
  {
    key: "quantity",
    header: receivingCopy("columns.quantity"),
    render: (row) => formatReceivingQuantity(row.lines[0]?.quantity ?? "0", row.lines[0]?.baseUomCode),
    align: "right",
    width: "140px"
  },
  {
    key: "movement",
    header: receivingCopy("columns.movement"),
    render: (row) =>
      row.stockMovements?.length ? (
        <StatusChip tone="success">{receivingCopy("movement.postedCount", { count: row.stockMovements.length })}</StatusChip>
      ) : (
        "-"
      ),
    width: "130px"
  }
];

const lineColumns: DataTableColumn<GoodsReceiptLine>[] = [
  {
    key: "sku",
    header: receivingCopy("line.columns.sku"),
    render: (row) => (
      <span className="erp-receiving-receipt-cell">
        <strong>{row.sku}</strong>
        <small>{row.itemName ?? row.itemId}</small>
      </span>
    )
  },
  {
    key: "poLine",
    header: receivingCopy("line.columns.poLine"),
    render: (row) => row.purchaseOrderLineId ?? "-",
    width: "180px"
  },
  {
    key: "lot",
    header: receivingCopy("line.columns.lotExpiry"),
    render: (row) => (
      <span className="erp-receiving-receipt-cell">
        <strong>{row.lotNo ?? row.batchNo ?? row.batchId ?? "-"}</strong>
        <small>{row.expiryDate ?? "-"}</small>
      </span>
    ),
    width: "150px"
  },
  {
    key: "packaging",
    header: receivingCopy("line.columns.packaging"),
    render: (row) => <StatusChip tone={row.packagingStatus === "intact" ? "success" : "warning"}>{formatPackagingStatus(row.packagingStatus)}</StatusChip>,
    width: "150px"
  },
  {
    key: "qc",
    header: receivingCopy("line.columns.qc"),
    render: (row) => <StatusChip tone={qcStatusTone(row.qcStatus)}>{qcStatusLabel(row.qcStatus)}</StatusChip>,
    width: "130px"
  },
  {
    key: "quantity",
    header: receivingCopy("line.columns.quantity"),
    render: (row) => formatReceivingQuantity(row.quantity, row.baseUomCode),
    align: "right",
    width: "130px"
  }
];

const movementColumns: DataTableColumn<GoodsReceiptStockMovement>[] = [
  {
    key: "movement",
    header: receivingCopy("movement.columns.movement"),
    render: (row) => row.movementNo
  },
  {
    key: "status",
    header: receivingCopy("movement.columns.stock"),
    render: (row) => <StatusChip tone={row.stockStatus === "available" ? "success" : "warning"}>{stockStatusLabel(row.stockStatus)}</StatusChip>,
    width: "120px"
  },
  {
    key: "quantity",
    header: receivingCopy("movement.columns.quantity"),
    render: (row) => formatReceivingQuantity(row.quantity, row.baseUomCode),
    align: "right",
    width: "130px"
  }
];

export function WarehouseReceivingPrototype() {
  const [warehouseId, setWarehouseId] = useState("wh-hcm-fg");
  const [status, setStatus] = useState<StatusFilter>("");
  const [locationId, setLocationId] = useState("loc-hcm-fg-recv-01");
  const [supplierId, setSupplierId] = useState("sup-rm-bioactive");
  const [selectedPurchaseOrderId, setSelectedPurchaseOrderId] = useState("");
  const [purchaseOrderDetail, setPurchaseOrderDetail] = useState<PurchaseOrder | null>(null);
  const [purchaseOrders, setPurchaseOrders] = useState<PurchaseOrder[]>([]);
  const [purchaseOrdersLoading, setPurchaseOrdersLoading] = useState(false);
  const [purchaseOrderError, setPurchaseOrderError] = useState<string | null>(null);
  const [referenceDocId, setReferenceDocId] = useState("");
  const [purchaseOrderLineId, setPurchaseOrderLineId] = useState("");
  const [deliveryNoteNo, setDeliveryNoteNo] = useState("DN-260429-UI");
  const [batchId, setBatchId] = useState(defaultBatch.value);
  const [lotNo, setLotNo] = useState(defaultBatch.lotNo);
  const [expiryDate, setExpiryDate] = useState(defaultBatch.expiryDate);
  const [quantity, setQuantity] = useState("12");
  const [uomCode, setUomCode] = useState(defaultBatch.baseUomCode);
  const [baseUomCode, setBaseUomCode] = useState(defaultBatch.baseUomCode);
  const [packagingStatus, setPackagingStatus] = useState<ReceivingPackagingStatus>("intact");
  const [receivingNote, setReceivingNote] = useState("");
  const [attachmentRef, setAttachmentRef] = useState("");
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
  const selectedBatch = receivingBatchOptions.find((batch) => batch.value === batchId) ?? defaultBatch;
  const selectedPurchaseOrderLine = purchaseOrderDetail?.lines.find((line) => line.id === purchaseOrderLineId);
  const availableBatchOptions = useMemo(() => {
    const matches = selectedPurchaseOrderLine
      ? receivingBatchOptions.filter((batch) => batch.itemId === selectedPurchaseOrderLine.itemId)
      : receivingBatchOptions;

    return matches.length > 0 ? matches : receivingBatchOptions;
  }, [selectedPurchaseOrderLine]);
  const locationOptions = receivingLocationOptions.filter((location) => location.warehouseId === warehouseId);
  const totals = summarizeReceipts(visibleReceipts);
  const receivingAttachmentItems = useMemo<AttachmentPanelItem[]>(() => {
    const items: AttachmentPanelItem[] = [];
    if (selectedReceipt?.deliveryNoteNo) {
      items.push({
        id: `${selectedReceipt.id}:delivery-note`,
        name: selectedReceipt.deliveryNoteNo,
        kind: receivingCopy("attachments.deliveryNote"),
        uploadedBy: selectedReceipt.supplierId ?? receivingCopy("attachments.supplier"),
        uploadedAt: selectedReceipt.updatedAt,
        storageKey: `receiving/${selectedReceipt.id}/${selectedReceipt.deliveryNoteNo}`,
        status: <StatusChip tone="info">{receivingCopy("attachments.document")}</StatusChip>,
        canDownload: true,
        onDownload: () => setFeedback({ tone: "info", message: selectedReceipt.deliveryNoteNo ?? "" })
      });
    }
    if (attachmentRef.trim()) {
      items.push({
        id: "draft-attachment-ref",
        name: attachmentRef.trim(),
        kind: receivingCopy("attachments.receivingEvidence"),
        uploadedBy: receivingCopy("attachments.warehouse"),
        uploadedAt: selectedReceipt?.updatedAt ?? new Date().toISOString(),
        detail: receivingCopy("attachments.draftReference"),
        status: <StatusChip tone="warning">{goodsReceiptStatusLabel("draft")}</StatusChip>,
        canDelete: true,
        deleteLabel: receivingCopy("actions.remove"),
        onDelete: () => setAttachmentRef("")
      });
    }

    return items;
  }, [attachmentRef, selectedReceipt]);

  useEffect(() => {
    let active = true;
    setPurchaseOrdersLoading(true);
    setPurchaseOrderError(null);

    Promise.all([
      getPurchaseOrders({ warehouseId, status: "approved" }),
      getPurchaseOrders({ warehouseId, status: "partially_received" })
    ])
      .then(([approved, partial]) => {
        if (active) {
          setPurchaseOrders(uniquePurchaseOrders([...approved, ...partial]));
        }
      })
      .catch((reason) => {
        if (active) {
          setPurchaseOrderError(reason instanceof Error ? reason.message : receivingCopy("feedback.purchaseOrdersLoadFailed"));
        }
      })
      .finally(() => {
        if (active) {
          setPurchaseOrdersLoading(false);
        }
      });

    return () => {
      active = false;
    };
  }, [warehouseId]);

  function handleWarehouseChange(nextWarehouseId: string) {
    setWarehouseId(nextWarehouseId);
    const nextLocation = receivingLocationOptions.find((location) => location.warehouseId === nextWarehouseId);
    setLocationId(nextLocation?.value ?? "");
    setSelectedPurchaseOrderId("");
    setPurchaseOrderDetail(null);
    setReferenceDocId("");
    setPurchaseOrderLineId("");
  }

  async function handlePurchaseOrderChange(nextPurchaseOrderId: string) {
    setSelectedPurchaseOrderId(nextPurchaseOrderId);
    setPurchaseOrderError(null);
    if (!nextPurchaseOrderId) {
      setPurchaseOrderDetail(null);
      return;
    }

    setPurchaseOrdersLoading(true);
    try {
      const order = await getPurchaseOrder(nextPurchaseOrderId);
      setPurchaseOrderDetail(order);
      applyPurchaseOrder(order, order.lines[0]?.id);
    } catch (reason) {
      setPurchaseOrderDetail(null);
      setPurchaseOrderError(reason instanceof Error ? reason.message : receivingCopy("feedback.purchaseOrderLoadFailed"));
    } finally {
      setPurchaseOrdersLoading(false);
    }
  }

  function applyPurchaseOrder(order: PurchaseOrder, lineId?: string) {
    setReferenceDocId(order.id);
    setSupplierId(order.supplierId);
    setWarehouseId(order.warehouseId);
    const nextLocation = receivingLocationOptions.find((location) => location.warehouseId === order.warehouseId);
    setLocationId(nextLocation?.value ?? "");
    const nextLine = lineId ? order.lines.find((line) => line.id === lineId) : order.lines[0];
    if (nextLine) {
      applyPurchaseOrderLine(nextLine);
    }
  }

  function applyPurchaseOrderLine(line: PurchaseOrderLine) {
    setPurchaseOrderLineId(line.id);
    setQuantity(line.orderedQty);
    setUomCode(line.uomCode);
    setBaseUomCode(line.baseUomCode);
    const nextBatch = receivingBatchOptions.find((batch) => batch.itemId === line.itemId) ?? defaultBatch;
    applyBatch(nextBatch.value, line);
  }

  function applyBatch(nextBatchId: string, line: PurchaseOrderLine | undefined = selectedPurchaseOrderLine) {
    const nextBatch = receivingBatchOptions.find((batch) => batch.value === nextBatchId) ?? defaultBatch;
    setBatchId(nextBatch.value);
    setLotNo(nextBatch.lotNo);
    setExpiryDate(nextBatch.expiryDate);
    setUomCode(line?.uomCode ?? nextBatch.baseUomCode);
    setBaseUomCode(line?.baseUomCode ?? nextBatch.baseUomCode);
  }

  function handlePurchaseOrderLineChange(nextLineId: string) {
    const nextLine = purchaseOrderDetail?.lines.find((line) => line.id === nextLineId);
    if (nextLine) {
      applyPurchaseOrderLine(nextLine);
    } else {
      setPurchaseOrderLineId(nextLineId);
    }
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
        supplierId,
        deliveryNoteNo,
        referenceDocType: "purchase_order",
        referenceDocId,
        lines: [
          {
            id: "line-ui-001",
            purchaseOrderLineId,
            itemId: selectedPurchaseOrderLine?.itemId ?? selectedBatch.itemId,
            sku: selectedPurchaseOrderLine?.skuCode ?? selectedBatch.sku,
            itemName: selectedPurchaseOrderLine?.itemName ?? selectedBatch.itemName,
            batchId,
            batchNo: selectedBatch.batchNo,
            lotNo,
            expiryDate,
            quantity,
            uomCode,
            baseUomCode,
            packagingStatus,
            qcStatus: selectedBatch.qcStatus
          }
        ]
      });
      upsertLocalReceipt(receipt);
      setSelectedReceiptId(receipt.id);
      setFeedback({ tone: "success", message: receivingCopy("feedback.created", { receiptNo: receipt.receiptNo }) });
    } catch (reason) {
      setFeedback({ tone: "danger", message: reason instanceof Error ? reason.message : receivingCopy("feedback.createFailed") });
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
      setFeedback({ tone: "success", message: `${updated.receiptNo} / ${goodsReceiptStatusLabel(updated.status)}` });
    } catch (reason) {
      setFeedback({ tone: "danger", message: reason instanceof Error ? reason.message : receivingCopy("feedback.actionFailed") });
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
          <h1 className="erp-page-title">{receivingCopy("title")}</h1>
          <p className="erp-page-description">{receivingCopy("description")}</p>
        </div>
        <div className="erp-page-actions">
          <a className="erp-button erp-button--secondary" href="#receiving-draft">
            {goodsReceiptStatusLabel("draft")}
          </a>
          <a className="erp-button erp-button--secondary" href="#receiving-detail">
            {receivingCopy("detail.title")}
          </a>
          <a className="erp-button erp-button--primary" href="#receiving-list">
            {receivingCopy("list.title")}
          </a>
        </div>
      </header>

      <section className="erp-receiving-toolbar" aria-label={receivingCopy("filters.label")}>
        <label className="erp-field">
          <span>{receivingCopy("filters.warehouse")}</span>
          <select className="erp-input" value={warehouseId} onChange={(event) => handleWarehouseChange(event.target.value)}>
            {receivingWarehouseOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>{receivingCopy("filters.status")}</span>
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
        <ReceivingKPI label={goodsReceiptStatusLabel("draft")} tone="warning" value={totals.draft} />
        <ReceivingKPI label={goodsReceiptStatusLabel("submitted")} tone="info" value={totals.submitted} />
        <ReceivingKPI label={goodsReceiptStatusLabel("inspect_ready")} tone="warning" value={totals.inspectReady} />
        <ReceivingKPI label={goodsReceiptStatusLabel("posted")} tone="success" value={totals.posted} />
      </section>

      <section className="erp-receiving-workspace">
        <form className="erp-card erp-card--padded erp-receiving-draft" id="receiving-draft" onSubmit={handleCreateDraft}>
          <div className="erp-section-header">
            <h2 className="erp-section-title">{receivingCopy("draft.title")}</h2>
            <StatusChip tone={feedback?.tone ?? (purchaseOrderError ? "danger" : "info")}>
              {feedback?.message ?? purchaseOrderError ?? receivingCopy("feedback.ready")}
            </StatusChip>
          </div>

          <div className="erp-receiving-form-section">
            <div className="erp-receiving-section-label">
              <strong>{receivingCopy("po.context")}</strong>
              <StatusChip tone={purchaseOrdersLoading ? "info" : purchaseOrders.length > 0 ? "success" : "warning"}>
                {purchaseOrdersLoading ? receivingCopy("feedback.loading") : receivingCopy("po.openCount", { count: purchaseOrders.length })}
              </StatusChip>
            </div>
            <div className="erp-receiving-form-grid">
              <label className="erp-field">
                <span>{receivingCopy("po.openPO")}</span>
                <select
                  className="erp-input"
                  value={selectedPurchaseOrderId}
                  onChange={(event) => void handlePurchaseOrderChange(event.target.value)}
                >
                  <option value="">{receivingCopy("po.manualReference")}</option>
                  {purchaseOrders.map((order) => (
                    <option key={order.id} value={order.id}>
                      {order.poNo} / {order.supplierCode ?? order.supplierName}
                    </option>
                  ))}
                </select>
              </label>
              <label className="erp-field">
                <span>{receivingCopy("po.reference")}</span>
                <input
                  className="erp-input"
                  type="text"
                  value={referenceDocId}
                  onChange={(event) => setReferenceDocId(event.target.value)}
                  required
                />
              </label>
              <label className="erp-field">
                <span>{receivingCopy("fields.supplier")}</span>
                <select className="erp-input" value={supplierId} onChange={(event) => setSupplierId(event.target.value)} required>
                  {purchaseSupplierOptions.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              </label>
              <label className="erp-field">
                <span>{receivingCopy("po.line")}</span>
                {purchaseOrderDetail?.lines.length ? (
                  <select
                    className="erp-input"
                    value={purchaseOrderLineId}
                    onChange={(event) => handlePurchaseOrderLineChange(event.target.value)}
                    required
                  >
                    {purchaseOrderDetail.lines.map((line) => (
                      <option key={line.id} value={line.id}>
                        {line.lineNo} / {line.skuCode} / {formatPurchaseQuantity(line.receivedQty, line.baseUomCode)} {receivingCopy("po.of")}{" "}
                        {formatPurchaseQuantity(line.orderedQty, line.baseUomCode)}
                      </option>
                    ))}
                  </select>
                ) : (
                  <input
                    className="erp-input"
                    type="text"
                    value={purchaseOrderLineId}
                    onChange={(event) => setPurchaseOrderLineId(event.target.value)}
                    required
                  />
                )}
              </label>
            </div>
            {purchaseOrderDetail ? (
              <div className="erp-receiving-po-summary">
                <StatusChip tone={purchaseOrderStatusTone(purchaseOrderDetail.status)}>
                  {purchaseStatusLabel(purchaseOrderDetail.status)}
                </StatusChip>
                <span>{purchaseOrderDetail.poNo}</span>
                <strong>{purchaseOrderDetail.supplierName}</strong>
              </div>
            ) : null}
          </div>

          <div className="erp-receiving-form-section">
            <div className="erp-receiving-section-label">
              <strong>{receivingCopy("delivery.title")}</strong>
              <StatusChip tone={packagingStatus === "intact" ? "success" : "warning"}>{formatPackagingStatus(packagingStatus)}</StatusChip>
            </div>
            <div className="erp-receiving-form-grid">
              <label className="erp-field">
                <span>{receivingCopy("delivery.note")}</span>
                <input
                  className="erp-input"
                  type="text"
                  value={deliveryNoteNo}
                  onChange={(event) => setDeliveryNoteNo(event.target.value.toUpperCase())}
                  required
                />
              </label>
              <label className="erp-field">
                <span>{receivingCopy("fields.location")}</span>
                <select className="erp-input" value={locationId} onChange={(event) => setLocationId(event.target.value)} required>
                  {locationOptions.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              </label>
              <label className="erp-field">
                <span>{receivingCopy("delivery.packaging")}</span>
                <select
                  className="erp-input"
                  value={packagingStatus}
                  onChange={(event) => setPackagingStatus(event.target.value as ReceivingPackagingStatus)}
                  required
                >
                  {receivingPackagingStatusOptions.map((option) => (
                    <option key={option.value} value={option.value}>
                      {packagingStatusLabel(option.value)}
                    </option>
                  ))}
                </select>
              </label>
              <label className="erp-field">
                <span>{receivingCopy("attachments.reference")}</span>
                <input
                  className="erp-input"
                  type="text"
                  value={attachmentRef}
                  onChange={(event) => setAttachmentRef(event.target.value)}
                  placeholder={receivingCopy("attachments.placeholder")}
                />
              </label>
            </div>
          </div>

          <div className="erp-receiving-form-section">
            <div className="erp-receiving-section-label">
              <strong>{receivingCopy("line.check")}</strong>
              <StatusChip tone={qcStatusTone(selectedBatch.qcStatus)}>{qcStatusLabel(selectedBatch.qcStatus)}</StatusChip>
            </div>
            <div className="erp-receiving-form-grid">
              <label className="erp-field">
                <span>{receivingCopy("line.batchSku")}</span>
                <select className="erp-input" value={batchId} onChange={(event) => applyBatch(event.target.value)} required>
                  {availableBatchOptions.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              </label>
              <label className="erp-field">
                <span>{receivingCopy("line.quantity")}</span>
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
              <label className="erp-field">
                <span>{receivingCopy("line.uom")}</span>
                <input className="erp-input" type="text" value={uomCode} onChange={(event) => setUomCode(event.target.value.toUpperCase())} required />
              </label>
              <label className="erp-field">
                <span>{receivingCopy("line.baseUom")}</span>
                <input
                  className="erp-input"
                  type="text"
                  value={baseUomCode}
                  onChange={(event) => setBaseUomCode(event.target.value.toUpperCase())}
                  required
                />
              </label>
              <label className="erp-field">
                <span>{receivingCopy("line.lot")}</span>
                <input className="erp-input" type="text" value={lotNo} onChange={(event) => setLotNo(event.target.value.toUpperCase())} required />
              </label>
              <label className="erp-field">
                <span>{receivingCopy("line.expiry")}</span>
                <input className="erp-input" type="date" value={expiryDate} onChange={(event) => setExpiryDate(event.target.value)} required />
              </label>
            </div>
            <label className="erp-field erp-receiving-note-field">
              <span>{receivingCopy("fields.note")}</span>
              <textarea className="erp-input" value={receivingNote} onChange={(event) => setReceivingNote(event.target.value)} rows={3} />
            </label>
          </div>

          <div className="erp-receiving-selected-line">
            <StatusChip tone={qcStatusTone(selectedBatch.qcStatus)}>{qcStatusLabel(selectedBatch.qcStatus)}</StatusChip>
            <span>
              {selectedPurchaseOrderLine?.skuCode ?? selectedBatch.sku} / {lotNo} / {expiryDate}
            </span>
            <strong>{formatReceivingQuantity(quantity || "0", baseUomCode)}</strong>
          </div>

          <div className="erp-receiving-actions">
            <button className="erp-button erp-button--primary" type="submit" disabled={busyAction === "create"}>
              {busyAction === "create" ? receivingCopy("actions.creating") : receivingCopy("actions.createDraft")}
            </button>
          </div>
        </form>

        <section className="erp-card erp-card--padded erp-receiving-detail" id="receiving-detail">
          <div className="erp-section-header">
            <h2 className="erp-section-title">{receivingCopy("detail.title")}</h2>
            {selectedReceipt ? (
              <StatusChip tone={goodsReceiptStatusTone(selectedReceipt.status)}>
                {goodsReceiptStatusLabel(selectedReceipt.status)}
              </StatusChip>
            ) : null}
          </div>

          {selectedReceipt ? (
            <>
              <div className="erp-receiving-detail-grid">
                <ReceivingFact label={receivingCopy("detail.receipt")} value={selectedReceipt.receiptNo} />
                <ReceivingFact label={receivingCopy("detail.deliveryNote")} value={selectedReceipt.deliveryNoteNo ?? "-"} />
                <ReceivingFact label={receivingCopy("detail.po")} value={selectedReceipt.referenceDocId} />
                <ReceivingFact label={receivingCopy("detail.supplier")} value={selectedReceipt.supplierId ?? "-"} />
                <ReceivingFact label={receivingCopy("detail.warehouse")} value={selectedReceipt.warehouseCode} />
                <ReceivingFact label={receivingCopy("detail.location")} value={selectedReceipt.locationCode} />
                <ReceivingFact label={receivingCopy("detail.updated")} value={formatReceivingDateTime(selectedReceipt.updatedAt)} />
                <ReceivingFact label={receivingCopy("detail.audit")} value={selectedReceipt.auditLogId ?? "-"} />
              </div>

              <div className="erp-receiving-actions">
                <button
                  className="erp-button erp-button--secondary"
                  type="button"
                  disabled={selectedReceipt.status !== "draft" || Boolean(busyAction)}
                  onClick={() => void runAction(selectedReceipt, "submit")}
                >
                  {receivingCopy("actions.submit")}
                </button>
                <button
                  className="erp-button erp-button--secondary"
                  type="button"
                  disabled={selectedReceipt.status !== "submitted" || Boolean(busyAction)}
                  onClick={() => void runAction(selectedReceipt, "inspect")}
                >
                  {receivingCopy("actions.inspectReady")}
                </button>
                <button
                  className="erp-button erp-button--primary"
                  type="button"
                  disabled={selectedReceipt.status !== "inspect_ready" || Boolean(busyAction)}
                  onClick={() => setConfirmPostId(selectedReceipt.id)}
                >
                  {receivingCopy("actions.post")}
                </button>
              </div>

              <AttachmentPanel
                title={receivingCopy("attachments.title")}
                items={receivingAttachmentItems}
                emptyMessage={receivingCopy("attachments.empty")}
              />

              <div className="erp-receiving-subsection">
                <h3 className="erp-section-title">{receivingCopy("line.lines")}</h3>
                <DataTable columns={lineColumns} rows={selectedReceipt.lines} getRowKey={(row) => row.id} />
              </div>

              <div className="erp-receiving-subsection">
                <h3 className="erp-section-title">{receivingCopy("movement.title")}</h3>
                <DataTable
                  columns={movementColumns}
                  rows={selectedReceipt.stockMovements ?? []}
                  getRowKey={(row) => row.movementNo}
                  emptyState={
                    <EmptyState
                      title={receivingCopy("movement.emptyTitle")}
                      description={receivingCopy("movement.emptyDescription")}
                    />
                  }
                />
              </div>
            </>
          ) : (
            <EmptyState title={receivingCopy("empty.noSelection")} description={receivingCopy("empty.noSelectionDescription")} />
          )}
        </section>
      </section>

      <section className="erp-card erp-card--padded erp-module-table-card" id="receiving-list">
        <div className="erp-section-header">
          <h2 className="erp-section-title">{receivingCopy("list.title")}</h2>
          <StatusChip tone={visibleReceipts.length === 0 ? "warning" : "info"}>
            {receivingCopy("list.rows", { count: visibleReceipts.length })}
          </StatusChip>
        </div>
        <DataTable
          columns={[
            ...receiptColumns,
            {
              key: "action",
              header: receivingCopy("columns.action"),
              render: (row) => (
                <button
                  className="erp-button erp-button--secondary erp-button--compact"
                  type="button"
                  onClick={() => setSelectedReceiptId(row.id)}
                >
                  {receivingCopy("actions.view")}
                </button>
              ),
              width: "110px"
            }
          ]}
          rows={visibleReceipts}
          getRowKey={(row) => row.id}
          loading={loading}
          error={error?.message}
          emptyState={<EmptyState title={receivingCopy("empty.noReceipts")} description={receivingCopy("empty.changeFilters")} />}
        />
      </section>

      <ConfirmDialog
        open={Boolean(confirmPostId && selectedReceipt)}
        title={receivingCopy("confirmPost.title")}
        description={receivingCopy("confirmPost.description", { receiptNo: selectedReceipt?.receiptNo ?? receivingCopy("detail.receipt") })}
        confirmLabel={receivingCopy("actions.post")}
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

function formatPackagingStatus(status: ReceivingPackagingStatus) {
  return packagingStatusLabel(status);
}

function goodsReceiptStatusLabel(status: GoodsReceiptStatus) {
  return receivingCopy(`status.${status}`);
}

function purchaseStatusLabel(status: PurchaseOrder["status"]) {
  return t(`purchase.status.${status}`);
}

function packagingStatusLabel(status: ReceivingPackagingStatus) {
  return receivingCopy(`packaging.${status}`);
}

function qcStatusLabel(status?: BatchQCStatus) {
  return status ? receivingCopy(`qcStatus.${status}`) : "-";
}

function stockStatusLabel(status: GoodsReceiptStockMovement["stockStatus"]) {
  return receivingCopy(`stockStatus.${status}`);
}

function receivingCopy(key: string, values?: Record<string, string | number>) {
  return t(`purchase.receiving.${key}`, { values });
}

function uniquePurchaseOrders(orders: PurchaseOrder[]) {
  const seen = new Set<string>();
  return orders.filter((order) => {
    if (seen.has(order.id)) {
      return false;
    }
    seen.add(order.id);
    return true;
  });
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
