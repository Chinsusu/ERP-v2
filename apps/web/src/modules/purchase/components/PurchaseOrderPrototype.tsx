"use client";

import { useEffect, useMemo, useRef, useState, type FormEvent } from "react";
import {
  DataTable,
  EmptyState,
  StatusChip,
  type DataTableColumn,
  type StatusTone
} from "@/shared/design-system/components";
import { AttachmentPanel, type AttachmentPanelItem } from "@/shared/design-system/pageTemplates";
import { t } from "@/shared/i18n";
import { usePurchaseOrders } from "../hooks/usePurchaseOrders";
import {
  approvePurchaseOrder,
  cancelPurchaseOrder,
  closePurchaseOrder,
  createPurchaseOrder,
  formatPurchaseDate,
  formatPurchaseMoney,
  formatPurchaseQuantity,
  getPurchaseOrder,
  purchaseItemOptions,
  purchaseOrderStatusTone,
  purchaseStatusOptions,
  purchaseSupplierOptions,
  purchaseWarehouseOptions,
  submitPurchaseOrder,
  updatePurchaseOrder
} from "../services/purchaseOrderService";
import type {
  CreatePurchaseOrderInput,
  PurchaseOrder,
  PurchaseOrderLine,
  PurchaseOrderLineInput,
  PurchaseOrderQuery,
  PurchaseOrderStatus
} from "../types";

type StatusFilter = "" | PurchaseOrderStatus;

const orderColumns = (onSelect: (order: PurchaseOrder) => void): DataTableColumn<PurchaseOrder>[] => [
  {
    key: "order",
    header: purchaseCopy("order.columns.po"),
    render: (row) => (
      <span className="erp-purchase-order-cell">
        <strong>{row.poNo}</strong>
        <small>{row.supplierName}</small>
      </span>
    ),
    width: "220px"
  },
  {
    key: "warehouse",
    header: purchaseCopy("order.columns.warehouse"),
    render: (row) => row.warehouseCode ?? "-",
    width: "130px"
  },
  {
    key: "status",
    header: purchaseCopy("order.columns.status"),
    render: (row) => (
      <StatusChip tone={purchaseOrderStatusTone(row.status)}>{purchaseOrderStatusLabel(row.status)}</StatusChip>
    ),
    width: "160px"
  },
  {
    key: "expected",
    header: purchaseCopy("order.columns.expected"),
    render: (row) => formatPurchaseDate(row.expectedDate),
    width: "120px"
  },
  {
    key: "lines",
    header: purchaseCopy("order.columns.lines"),
    render: (row) => row.lineCount ?? row.lines.length,
    align: "right",
    width: "80px"
  },
  {
    key: "received",
    header: purchaseCopy("order.columns.received"),
    render: (row) => row.receivedLineCount ?? 0,
    align: "right",
    width: "100px"
  },
  {
    key: "total",
    header: purchaseCopy("order.columns.total"),
    render: (row) => formatPurchaseMoney(row.totalAmount, row.currencyCode),
    align: "right",
    width: "140px"
  },
  {
    key: "action",
    header: purchaseCopy("order.columns.action"),
    render: (row) => (
      <button className="erp-button erp-button--secondary" type="button" onClick={() => onSelect(row)}>
        {purchaseCopy("actions.open")}
      </button>
    ),
    width: "96px",
    sticky: true
  }
];

const lineColumns: DataTableColumn<PurchaseOrderLine>[] = [
  {
    key: "sku",
    header: purchaseCopy("line.columns.sku"),
    render: (row) => (
      <span className="erp-purchase-order-cell">
        <strong>{row.skuCode}</strong>
        <small>{row.itemName}</small>
      </span>
    )
  },
  {
    key: "qty",
    header: purchaseCopy("line.columns.ordered"),
    render: (row) => formatPurchaseQuantity(row.orderedQty, row.uomCode),
    align: "right",
    width: "140px"
  },
  {
    key: "received",
    header: purchaseCopy("line.columns.received"),
    render: (row) => formatPurchaseQuantity(row.receivedQty, row.uomCode),
    align: "right",
    width: "140px"
  },
  {
    key: "base",
    header: purchaseCopy("line.columns.base"),
    render: (row) => formatPurchaseQuantity(row.baseOrderedQty, row.baseUomCode),
    align: "right",
    width: "140px"
  },
  {
    key: "price",
    header: purchaseCopy("line.columns.unitPrice"),
    render: (row) => formatPurchaseMoney(row.unitPrice, row.currencyCode),
    align: "right",
    width: "140px"
  },
  {
    key: "amount",
    header: purchaseCopy("line.columns.amount"),
    render: (row) => formatPurchaseMoney(row.lineAmount, row.currencyCode),
    align: "right",
    width: "140px"
  }
];

export function PurchaseOrderPrototype() {
  const mountedRef = useRef(true);
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState<StatusFilter>("");
  const [filterSupplierId, setFilterSupplierId] = useState("");
  const [filterWarehouseId, setFilterWarehouseId] = useState("wh-hcm-rm");
  const [supplierId, setSupplierId] = useState("sup-rm-bioactive");
  const [warehouseId, setWarehouseId] = useState("wh-hcm-rm");
  const [expectedDate, setExpectedDate] = useState("2026-05-02");
  const [note, setNote] = useState("");
  const [draftLines, setDraftLines] = useState<PurchaseOrderLineInput[]>([]);
  const [lineItemId, setLineItemId] = useState("item-serum-30ml");
  const [lineQty, setLineQty] = useState("10");
  const [lineUnitPrice, setLineUnitPrice] = useState("125000");
  const [lineNote, setLineNote] = useState("");
  const [poAttachmentName, setPoAttachmentName] = useState("supplier-quote.pdf");
  const [purchaseAttachmentRefs, setPurchaseAttachmentRefs] = useState<Record<string, string[]>>({});
  const [localOrders, setLocalOrders] = useState<PurchaseOrder[]>([]);
  const [selectedOrderId, setSelectedOrderId] = useState("po-260429-0001");
  const [autoOpenPurchaseSearch, setAutoOpenPurchaseSearch] = useState("");
  const [feedback, setFeedback] = useState<{ tone: StatusTone; message: string } | null>(null);
  const [busyAction, setBusyAction] = useState("");
  const query = useMemo<PurchaseOrderQuery>(
    () => ({
      search: search || undefined,
      status: status || undefined,
      supplierId: filterSupplierId || undefined,
      warehouseId: filterWarehouseId || undefined
    }),
    [filterSupplierId, filterWarehouseId, search, status]
  );
  const { orders, loading, error } = usePurchaseOrders(query);
  const visibleOrders = useMemo(() => mergeOrders(localOrders, orders, query), [localOrders, orders, query]);
  const selectedOrder = visibleOrders.find((order) => order.id === selectedOrderId) ?? visibleOrders[0] ?? null;
  const selectedSupplier = purchaseSupplierOptions.find((supplier) => supplier.value === supplierId) ?? purchaseSupplierOptions[0];
  const selectedLineItem = purchaseItemOptions.find((item) => item.value === lineItemId) ?? purchaseItemOptions[0];
  const totals = summarizeOrders(visibleOrders);
  const purchaseAttachmentItems = useMemo<AttachmentPanelItem[]>(
    () =>
      selectedOrder
        ? (purchaseAttachmentRefs[selectedOrder.id] ?? []).map((name) => ({
            id: `${selectedOrder.id}:${name}`,
            name,
            kind: name.toLowerCase().endsWith(".pdf")
              ? purchaseCopy("attachments.supplierDocument")
              : purchaseCopy("attachments.poEvidence"),
            uploadedBy: selectedOrder.supplierCode ?? selectedOrder.supplierName,
            uploadedAt: selectedOrder.updatedAt,
            storageKey: `purchase-orders/${selectedOrder.id}/${name}`,
            status: (
              <StatusChip tone={selectedOrder.status === "draft" ? "warning" : "info"}>
                {purchaseOrderStatusLabel(selectedOrder.status)}
              </StatusChip>
            ),
            canDownload: true,
            canDelete: selectedOrder.status === "draft",
            deleteLabel: purchaseCopy("actions.remove"),
            onDownload: () => setFeedback({ tone: "info", message: `purchase-orders/${selectedOrder.id}/${name}` }),
            onDelete: () => removePurchaseAttachment(selectedOrder.id, name)
          }))
        : [],
    [purchaseAttachmentRefs, selectedOrder]
  );

  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
    };
  }, []);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const nextSearch = params.get("search");
    const nextStatus = purchaseStatusFromParam(params.get("status"));
    const nextWarehouseId = purchaseWarehouseFromParam(params.get("warehouse_id"));

    if (nextSearch !== null) {
      const normalizedSearch = nextSearch.trim();
      setSearch(normalizedSearch);
      setAutoOpenPurchaseSearch(normalizedSearch);
      if (!params.has("warehouse_id")) {
        setFilterWarehouseId("");
      }
    }
    if (nextStatus !== null) {
      setStatus(nextStatus);
    }
    if (nextWarehouseId !== null) {
      setFilterWarehouseId(nextWarehouseId);
    }
  }, []);

  function handleLineItemChange(nextItemId: string) {
    const item = purchaseItemOptions.find((candidate) => candidate.value === nextItemId) ?? purchaseItemOptions[0];
    setLineItemId(item.value);
    setLineUnitPrice(item.defaultUnitPrice);
  }

  function handleAddLine(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const nextLine: PurchaseOrderLineInput = {
      lineNo: draftLines.length + 1,
      itemId: selectedLineItem.value,
      orderedQty: lineQty,
      uomCode: selectedLineItem.baseUomCode,
      unitPrice: lineUnitPrice,
      currencyCode: "VND",
      expectedDate,
      note: lineNote
    };
    setDraftLines((current) => [...current, nextLine]);
    setLineQty("10");
    setLineNote("");
    setFeedback({ tone: "info", message: purchaseCopy("feedback.lineAdded", { sku: selectedLineItem.skuCode }) });
  }

  async function handleCreateOrder(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (busyAction || draftLines.length === 0) {
      setFeedback({ tone: "danger", message: purchaseCopy("feedback.addLineRequired") });
      return;
    }

    setBusyAction("create");
    setFeedback(null);
    try {
      const input: CreatePurchaseOrderInput = {
        supplierId: selectedSupplier.value,
        warehouseId,
        expectedDate,
        currencyCode: "VND",
        note,
        lines: draftLines
      };
      const order = await createPurchaseOrder(input);
      upsertLocalOrder(order);
      setSelectedOrderId(order.id);
      setDraftLines([]);
      setFeedback({ tone: "success", message: purchaseCopy("feedback.created", { poNo: order.poNo }) });
    } catch (reason) {
      setFeedback({ tone: "danger", message: reason instanceof Error ? reason.message : purchaseCopy("feedback.createFailed") });
    } finally {
      setBusyAction("");
    }
  }

  async function handleReplaceDraftLines() {
    if (!selectedOrder || selectedOrder.status !== "draft" || draftLines.length === 0 || busyAction) {
      return;
    }

    setBusyAction(`update:${selectedOrder.id}`);
    setFeedback(null);
    try {
      const order = await updatePurchaseOrder(selectedOrder.id, {
        expectedVersion: selectedOrder.version,
        supplierId: selectedOrder.supplierId,
        warehouseId: selectedOrder.warehouseId,
        expectedDate: selectedOrder.expectedDate,
        note: selectedOrder.note,
        lines: draftLines
      });
      upsertLocalOrder(order);
      setSelectedOrderId(order.id);
      setDraftLines([]);
      setFeedback({ tone: "success", message: purchaseCopy("feedback.linesUpdated", { poNo: order.poNo }) });
    } catch (reason) {
      setFeedback({ tone: "danger", message: reason instanceof Error ? reason.message : purchaseCopy("feedback.updateFailed") });
    } finally {
      setBusyAction("");
    }
  }

  async function handleSelectOrder(order: PurchaseOrder) {
    setSelectedOrderId(order.id);
    if (order.lines.length > 0 || busyAction) {
      return;
    }
    setBusyAction(`load:${order.id}`);
    try {
      const detail = await getPurchaseOrder(order.id);
      upsertLocalOrder(detail);
    } catch (reason) {
      setFeedback({ tone: "danger", message: reason instanceof Error ? reason.message : purchaseCopy("feedback.detailFailed") });
    } finally {
      setBusyAction("");
    }
  }

  useEffect(() => {
    if (!autoOpenPurchaseSearch || loading || busyAction || visibleOrders.length === 0) {
      return;
    }

    const order = visibleOrders[0];
    setAutoOpenPurchaseSearch("");

    setSelectedOrderId(order.id);
    if (order.lines.length > 0) {
      return;
    }

    setBusyAction(`load:${order.id}`);
    getPurchaseOrder(order.id)
      .then((detail) => {
        if (mountedRef.current) {
          setLocalOrders((current) => [detail, ...current.filter((candidate) => candidate.id !== detail.id)]);
        }
      })
      .catch((reason) => {
        if (mountedRef.current) {
          setFeedback({ tone: "danger", message: reason instanceof Error ? reason.message : purchaseCopy("feedback.detailFailed") });
        }
      })
      .finally(() => {
        if (mountedRef.current) {
          setBusyAction("");
        }
      });
  }, [autoOpenPurchaseSearch, busyAction, loading, visibleOrders]);

  async function runAction(action: "submit" | "approve" | "cancel" | "close") {
    if (!selectedOrder || busyAction) {
      return;
    }
    const reason = action === "cancel" ? "Cancelled from purchase order board" : undefined;
    setBusyAction(`${action}:${selectedOrder.id}`);
    setFeedback(null);
    try {
      const result =
        action === "submit"
          ? await submitPurchaseOrder(selectedOrder.id, selectedOrder.version)
          : action === "approve"
            ? await approvePurchaseOrder(selectedOrder.id, selectedOrder.version)
            : action === "cancel"
              ? await cancelPurchaseOrder(selectedOrder.id, reason ?? "", selectedOrder.version)
              : await closePurchaseOrder(selectedOrder.id, selectedOrder.version);
      upsertLocalOrder(result.purchaseOrder);
      setSelectedOrderId(result.purchaseOrder.id);
      setFeedback({
        tone: action === "cancel" ? "warning" : "success",
        message: `${result.purchaseOrder.poNo} / ${purchaseOrderStatusLabel(result.currentStatus)}`
      });
    } catch (reason) {
      setFeedback({ tone: "danger", message: reason instanceof Error ? reason.message : purchaseCopy("feedback.actionFailed") });
    } finally {
      setBusyAction("");
    }
  }

  function upsertLocalOrder(order: PurchaseOrder) {
    setLocalOrders((current) => [order, ...current.filter((candidate) => candidate.id !== order.id)]);
  }

  function handleAddPurchaseAttachment(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedOrder || poAttachmentName.trim() === "") {
      return;
    }

    const fileName = poAttachmentName.trim();
    setPurchaseAttachmentRefs((current) => ({
      ...current,
      [selectedOrder.id]: Array.from(new Set([...(current[selectedOrder.id] ?? []), fileName]))
    }));
    setPoAttachmentName("");
    setFeedback({ tone: "info", message: purchaseCopy("feedback.attachmentLinked", { fileName, poNo: selectedOrder.poNo }) });
  }

  function removePurchaseAttachment(orderId: string, fileName: string) {
    setPurchaseAttachmentRefs((current) => ({
      ...current,
      [orderId]: (current[orderId] ?? []).filter((name) => name !== fileName)
    }));
  }

  return (
    <section className="erp-module-page erp-purchase-page">
      <header className="erp-page-header">
        <div>
          <p className="erp-module-eyebrow">PU</p>
          <h1 className="erp-page-title">{purchaseCopy("order.title")}</h1>
          <p className="erp-page-description">{purchaseCopy("order.description")}</p>
        </div>
        <div className="erp-page-actions">
          <a className="erp-button erp-button--secondary" href="#purchase-create">
            {purchaseCopy("actions.create")}
          </a>
          <a className="erp-button erp-button--primary" href="#purchase-list">
            {purchaseCopy("actions.orders")}
          </a>
        </div>
      </header>

      <section className="erp-purchase-toolbar" aria-label={purchaseCopy("filters.label")}>
        <label className="erp-field">
          <span>{purchaseCopy("filters.search")}</span>
          <input className="erp-input" type="search" value={search} onChange={(event) => setSearch(event.target.value)} />
        </label>
        <label className="erp-field">
          <span>{purchaseCopy("filters.status")}</span>
          <select className="erp-input" value={status} onChange={(event) => setStatus(event.target.value as StatusFilter)}>
            {purchaseStatusOptions.map((option) => (
              <option key={option.value || "all"} value={option.value}>
                {option.value ? purchaseOrderStatusLabel(option.value) : purchaseCopy("filters.allStatuses")}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>{purchaseCopy("filters.supplier")}</span>
          <select className="erp-input" value={filterSupplierId} onChange={(event) => setFilterSupplierId(event.target.value)}>
            <option value="">{purchaseCopy("filters.allSuppliers")}</option>
            {purchaseSupplierOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>{purchaseCopy("filters.warehouse")}</span>
          <select className="erp-input" value={filterWarehouseId} onChange={(event) => setFilterWarehouseId(event.target.value)}>
            <option value="">{purchaseCopy("filters.allWarehouses")}</option>
            {purchaseWarehouseOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
      </section>

      <section className="erp-kpi-grid erp-purchase-kpis" aria-label={purchaseCopy("summary.label")}>
        <KPI label={purchaseCopy("summary.pos")} value={String(totals.count)} />
        <KPI label={purchaseCopy("summary.draft")} value={String(totals.draft)} tone={totals.draft > 0 ? "warning" : "normal"} />
        <KPI label={purchaseCopy("summary.approved")} value={String(totals.approved)} tone="info" />
        <KPI label={purchaseCopy("summary.total")} value={formatPurchaseMoney(totals.totalAmount)} tone="success" />
      </section>

      {feedback ? (
        <p className={`erp-purchase-feedback erp-purchase-feedback--${feedback.tone}`} role="status">
          {feedback.message}
        </p>
      ) : null}

      <section className="erp-purchase-workspace">
        <section className="erp-card erp-card--padded erp-purchase-create" id="purchase-create">
          <header className="erp-section-header">
            <h2 className="erp-section-title">{purchaseCopy("create.title")}</h2>
            <StatusChip tone="info">{purchaseCopy("line.count", { count: draftLines.length })}</StatusChip>
          </header>
          <form className="erp-purchase-form-grid" onSubmit={handleCreateOrder}>
            <label className="erp-field">
              <span>{purchaseCopy("fields.supplier")}</span>
              <select className="erp-input" value={supplierId} onChange={(event) => setSupplierId(event.target.value)}>
                {purchaseSupplierOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </label>
            <label className="erp-field">
              <span>{purchaseCopy("fields.warehouse")}</span>
              <select className="erp-input" value={warehouseId} onChange={(event) => setWarehouseId(event.target.value)}>
                {purchaseWarehouseOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </label>
            <label className="erp-field">
              <span>{purchaseCopy("fields.expectedDate")}</span>
              <input className="erp-input" type="date" value={expectedDate} onChange={(event) => setExpectedDate(event.target.value)} />
            </label>
            <label className="erp-field erp-purchase-note-field">
              <span>{purchaseCopy("fields.note")}</span>
              <input className="erp-input" value={note} onChange={(event) => setNote(event.target.value)} />
            </label>
            <button className="erp-button erp-button--primary" type="submit" disabled={busyAction === "create"}>
              {purchaseCopy("actions.createPO")}
            </button>
          </form>

          <form className="erp-purchase-line-editor" onSubmit={handleAddLine}>
            <label className="erp-field">
              <span>{purchaseCopy("line.item")}</span>
              <select className="erp-input" value={lineItemId} onChange={(event) => handleLineItemChange(event.target.value)}>
                {purchaseItemOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </label>
            <label className="erp-field">
              <span>{purchaseCopy("line.qty")}</span>
              <input className="erp-input" inputMode="decimal" value={lineQty} onChange={(event) => setLineQty(event.target.value)} />
            </label>
            <label className="erp-field">
              <span>{purchaseCopy("line.unitPrice")}</span>
              <input
                className="erp-input"
                inputMode="decimal"
                value={lineUnitPrice}
                onChange={(event) => setLineUnitPrice(event.target.value)}
              />
            </label>
            <label className="erp-field">
              <span>{purchaseCopy("line.note")}</span>
              <input className="erp-input" value={lineNote} onChange={(event) => setLineNote(event.target.value)} />
            </label>
            <button className="erp-button erp-button--secondary" type="submit">
              {purchaseCopy("actions.addLine")}
            </button>
          </form>

          <DraftLineList lines={draftLines} onRemove={(index) => setDraftLines((current) => current.filter((_, i) => i !== index))} />
          <div className="erp-purchase-actions">
            <button
              className="erp-button erp-button--secondary"
              type="button"
              disabled={!selectedOrder || selectedOrder.status !== "draft" || draftLines.length === 0}
              onClick={handleReplaceDraftLines}
            >
              {purchaseCopy("actions.replaceDraftLines")}
            </button>
          </div>
        </section>

        <section className="erp-card erp-card--padded erp-purchase-detail" id="purchase-detail">
          <header className="erp-section-header">
            <h2 className="erp-section-title">{purchaseCopy("detail.title")}</h2>
            {selectedOrder ? (
              <StatusChip tone={purchaseOrderStatusTone(selectedOrder.status)}>
                {purchaseOrderStatusLabel(selectedOrder.status)}
              </StatusChip>
            ) : null}
          </header>
          {selectedOrder ? (
            <>
              <div className="erp-purchase-detail-grid">
                <Fact label={purchaseCopy("detail.po")} value={selectedOrder.poNo} />
                <Fact label={purchaseCopy("detail.supplier")} value={selectedOrder.supplierName} />
                <Fact label={purchaseCopy("detail.warehouse")} value={selectedOrder.warehouseCode ?? "-"} />
                <Fact label={purchaseCopy("detail.expected")} value={formatPurchaseDate(selectedOrder.expectedDate)} />
                <Fact label={purchaseCopy("detail.total")} value={formatPurchaseMoney(selectedOrder.totalAmount, selectedOrder.currencyCode)} />
                <Fact label={purchaseCopy("detail.version")} value={String(selectedOrder.version)} />
              </div>
              <div className="erp-purchase-actions">
                <button
                  className="erp-button erp-button--secondary"
                  type="button"
                  disabled={selectedOrder.status !== "draft" || Boolean(busyAction)}
                  onClick={() => runAction("submit")}
                >
                  {purchaseCopy("actions.submit")}
                </button>
                <button
                  className="erp-button erp-button--primary"
                  type="button"
                  disabled={selectedOrder.status !== "submitted" || Boolean(busyAction)}
                  onClick={() => runAction("approve")}
                >
                  {purchaseCopy("actions.approve")}
                </button>
                <button
                  className="erp-button erp-button--secondary"
                  type="button"
                  disabled={!["approved", "partially_received", "received"].includes(selectedOrder.status) || Boolean(busyAction)}
                  onClick={() => runAction("close")}
                >
                  {purchaseCopy("actions.close")}
                </button>
                <button
                  className="erp-button erp-button--danger"
                  type="button"
                  disabled={!["draft", "submitted", "approved"].includes(selectedOrder.status) || Boolean(busyAction)}
                  onClick={() => runAction("cancel")}
                >
                  {purchaseCopy("actions.cancel")}
                </button>
              </div>
              <div className="erp-purchase-subsection">
                <h3 className="erp-section-title">{purchaseCopy("line.items")}</h3>
                <DataTable columns={lineColumns} rows={selectedOrder.lines} getRowKey={(row) => row.id} />
              </div>
              <AttachmentPanel
                title={purchaseCopy("attachments.title")}
                items={purchaseAttachmentItems}
                emptyMessage={purchaseCopy("attachments.empty")}
                uploadAction={
                  <form className="erp-purchase-attachment-form" onSubmit={handleAddPurchaseAttachment}>
                    <input
                      aria-label={purchaseCopy("attachments.file")}
                      className="erp-input"
                      value={poAttachmentName}
                      onChange={(event) => setPoAttachmentName(event.currentTarget.value)}
                    />
                    <button
                      className="erp-button erp-button--secondary erp-button--compact"
                      type="submit"
                      disabled={!selectedOrder || selectedOrder.status === "cancelled"}
                    >
                      {purchaseCopy("attachments.upload")}
                    </button>
                  </form>
                }
              />
            </>
          ) : (
            <>
              <EmptyState title={purchaseCopy("empty.noSelection")} />
              <AttachmentPanel title={purchaseCopy("attachments.title")} items={[]} emptyMessage={purchaseCopy("attachments.empty")} />
            </>
          )}
        </section>
      </section>

      <section id="purchase-list">
        <DataTable
          columns={orderColumns((order) => void handleSelectOrder(order))}
          rows={visibleOrders}
          getRowKey={(row) => row.id}
          loading={loading}
          error={error?.message}
          emptyState={<EmptyState title={purchaseCopy("empty.noMatches")} />}
        />
      </section>
    </section>
  );
}

function DraftLineList({ lines, onRemove }: { lines: PurchaseOrderLineInput[]; onRemove: (index: number) => void }) {
  if (lines.length === 0) {
    return <p className="erp-purchase-empty-line">{purchaseCopy("line.noDraftLines")}</p>;
  }

  return (
    <ol className="erp-purchase-draft-lines" aria-label={purchaseCopy("line.draftLines")}>
      {lines.map((line, index) => {
        const item = purchaseItemOptions.find((candidate) => candidate.value === line.itemId) ?? purchaseItemOptions[0];

        return (
          <li key={`${line.itemId}-${index}`}>
            <span>
              <strong>{item.skuCode}</strong>
              <small>
                {line.orderedQty} {line.uomCode} / {formatPurchaseMoney(line.unitPrice)}
              </small>
            </span>
            <button className="erp-button erp-button--secondary" type="button" onClick={() => onRemove(index)}>
              {purchaseCopy("actions.remove")}
            </button>
          </li>
        );
      })}
    </ol>
  );
}

function KPI({ label, value, tone = "normal" }: { label: string; value: string; tone?: StatusTone }) {
  return (
    <article className="erp-card erp-card--padded erp-kpi-card">
      <span className="erp-kpi-label">{label}</span>
      <strong className="erp-kpi-value">{value}</strong>
      <StatusChip tone={tone}>{label}</StatusChip>
    </article>
  );
}

function purchaseStatusFromParam(value: string | null): StatusFilter | null {
  if (value === null) {
    return null;
  }
  if (purchaseStatusOptions.some((option) => option.value === value)) {
    return value as StatusFilter;
  }

  return null;
}

function purchaseWarehouseFromParam(value: string | null) {
  if (value === null) {
    return null;
  }

  return purchaseWarehouseOptions.some((option) => option.value === value) ? value : null;
}

function Fact({ label, value }: { label: string; value: string }) {
  return (
    <div className="erp-purchase-fact">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function purchaseOrderStatusLabel(status: PurchaseOrderStatus) {
  return purchaseCopy(`status.${status}`);
}

function purchaseCopy(key: string, values?: Record<string, string | number>) {
  return t(`purchase.${key}`, { values });
}

function mergeOrders(localOrders: PurchaseOrder[], fetchedOrders: PurchaseOrder[], query: PurchaseOrderQuery) {
  const localMatches = localOrders.filter((order) => matchesOrderQuery(order, query));
  const localIds = new Set(localMatches.map((order) => order.id));

  return [...localMatches, ...fetchedOrders.filter((order) => !localIds.has(order.id))];
}

function matchesOrderQuery(order: PurchaseOrder, query: PurchaseOrderQuery) {
  const search = query.search?.trim().toLowerCase();
  if (search) {
    const haystack = [order.poNo, order.supplierCode, order.supplierName, order.warehouseCode, order.note]
      .join(" ")
      .toLowerCase();
    if (!haystack.includes(search)) {
      return false;
    }
  }
  if (query.status && order.status !== query.status) {
    return false;
  }
  if (query.supplierId && order.supplierId !== query.supplierId) {
    return false;
  }
  if (query.warehouseId && order.warehouseId !== query.warehouseId) {
    return false;
  }

  return true;
}

function summarizeOrders(orders: PurchaseOrder[]) {
  return orders.reduce(
    (summary, order) => ({
      count: summary.count + 1,
      draft: summary.draft + (order.status === "draft" ? 1 : 0),
      approved: summary.approved + (["approved", "partially_received", "received"].includes(order.status) ? 1 : 0),
      totalAmount: addMoneyStrings(summary.totalAmount, order.totalAmount)
    }),
    { count: 0, draft: 0, approved: 0, totalAmount: "0.00" }
  );
}

function addMoneyStrings(left: string, right: string) {
  const [leftInt, leftFrac = ""] = left.split(".");
  const [rightInt, rightFrac = ""] = right.split(".");
  const leftValue = BigInt(`${leftInt}${leftFrac.padEnd(2, "0")}`);
  const rightValue = BigInt(`${rightInt}${rightFrac.padEnd(2, "0")}`);
  const digits = String(leftValue + rightValue).padStart(3, "0");

  return `${digits.slice(0, -2)}.${digits.slice(-2)}`;
}
