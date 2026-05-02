"use client";

import { useEffect, useMemo, useState, type FormEvent } from "react";
import {
  DataTable,
  EmptyState,
  StatusChip,
  type DataTableColumn,
  type StatusTone
} from "@/shared/design-system/components";
import { t } from "@/shared/i18n";
import { useSalesOrders } from "../hooks/useSalesOrders";
import {
  cancelSalesOrder,
  confirmSalesOrder,
  createSalesOrder,
  formatSalesDate,
  formatSalesMoney,
  formatSalesQuantity,
  salesChannelOptions,
  salesCustomerOptions,
  salesItemOptions,
  salesOrderStatusTone,
  salesStatusOptions,
  salesWarehouseOptions,
  updateSalesOrder
} from "../services/salesOrderService";
import type { CreateSalesOrderInput, SalesOrder, SalesOrderLine, SalesOrderLineInput, SalesOrderQuery, SalesOrderStatus } from "../types";

type StatusFilter = "" | SalesOrderStatus;

const orderColumns = (
  onSelect: (order: SalesOrder) => void
): DataTableColumn<SalesOrder>[] => [
  {
    key: "order",
    header: salesCopy("columns.order"),
    render: (row) => (
      <span className="erp-sales-order-cell">
        <strong>{row.orderNo}</strong>
        <small>{row.customerName}</small>
      </span>
    ),
    width: "220px"
  },
  {
    key: "channel",
    header: salesCopy("columns.channel"),
    render: (row) => row.channel,
    width: "100px"
  },
  {
    key: "warehouse",
    header: salesCopy("columns.warehouse"),
    render: (row) => row.warehouseCode ?? "-",
    width: "130px"
  },
  {
    key: "status",
    header: salesCopy("columns.status"),
    render: (row) => <StatusChip tone={salesOrderStatusTone(row.status)}>{salesStatusLabel(row.status)}</StatusChip>,
    width: "150px"
  },
  {
    key: "date",
    header: salesCopy("columns.date"),
    render: (row) => formatSalesDate(row.orderDate),
    width: "120px"
  },
  {
    key: "lines",
    header: salesCopy("columns.lines"),
    render: (row) => row.lineCount ?? row.lines.length,
    align: "right",
    width: "80px"
  },
  {
    key: "total",
    header: salesCopy("columns.total"),
    render: (row) => formatSalesMoney(row.totalAmount, row.currencyCode),
    align: "right",
    width: "140px"
  },
  {
    key: "action",
    header: salesCopy("columns.action"),
    render: (row) => (
      <button className="erp-button erp-button--secondary" type="button" onClick={() => onSelect(row)}>
        {salesCopy("actions.open")}
      </button>
    ),
    width: "96px",
    sticky: true
  }
];

const lineColumns: DataTableColumn<SalesOrderLine>[] = [
  {
    key: "sku",
    header: "SKU",
    render: (row) => (
      <span className="erp-sales-order-cell">
        <strong>{row.skuCode}</strong>
        <small>{row.itemName}</small>
      </span>
    )
  },
  {
    key: "qty",
    header: salesCopy("columns.qty"),
    render: (row) => formatSalesQuantity(row.orderedQty, row.uomCode),
    align: "right",
    width: "140px"
  },
  {
    key: "base",
    header: salesCopy("columns.base"),
    render: (row) => formatSalesQuantity(row.baseOrderedQty, row.baseUomCode),
    align: "right",
    width: "140px"
  },
  {
    key: "price",
    header: salesCopy("columns.unitPrice"),
    render: (row) => formatSalesMoney(row.unitPrice, row.currencyCode),
    align: "right",
    width: "140px"
  },
  {
    key: "amount",
    header: salesCopy("columns.amount"),
    render: (row) => formatSalesMoney(row.lineAmount, row.currencyCode),
    align: "right",
    width: "140px"
  }
];

export function SalesOrderPrototype() {
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState<StatusFilter>("");
  const [filterChannel, setFilterChannel] = useState("");
  const [filterWarehouseId, setFilterWarehouseId] = useState("wh-hcm-fg");
  const [customerId, setCustomerId] = useState("cus-dl-minh-anh");
  const [orderChannel, setOrderChannel] = useState("B2B");
  const [orderWarehouseId, setOrderWarehouseId] = useState("wh-hcm-fg");
  const [orderDate, setOrderDate] = useState("2026-04-28");
  const [note, setNote] = useState("");
  const [draftLines, setDraftLines] = useState<SalesOrderLineInput[]>([]);
  const [lineItemId, setLineItemId] = useState("item-serum-30ml");
  const [lineQty, setLineQty] = useState("1");
  const [lineUnitPrice, setLineUnitPrice] = useState("125000");
  const [lineDiscount, setLineDiscount] = useState("0");
  const [localOrders, setLocalOrders] = useState<SalesOrder[]>([]);
  const [selectedOrderId, setSelectedOrderId] = useState("so-260428-0001");
  const [feedback, setFeedback] = useState<{ tone: StatusTone; message: string } | null>(null);
  const [busyAction, setBusyAction] = useState("");
  const query = useMemo<SalesOrderQuery>(
    () => ({
      search: search || undefined,
      status: status || undefined,
      channel: filterChannel || undefined,
      warehouseId: filterWarehouseId || undefined
    }),
    [filterChannel, filterWarehouseId, search, status]
  );
  const { orders, loading, error } = useSalesOrders(query);
  const visibleOrders = useMemo(() => mergeOrders(localOrders, orders, query), [localOrders, orders, query]);
  const selectedOrder = visibleOrders.find((order) => order.id === selectedOrderId) ?? visibleOrders[0] ?? null;
  const selectedCustomer = salesCustomerOptions.find((customer) => customer.value === customerId) ?? salesCustomerOptions[0];
  const selectedLineItem = salesItemOptions.find((item) => item.value === lineItemId) ?? salesItemOptions[0];
  const totals = summarizeOrders(visibleOrders);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const nextStatus = salesStatusFromParam(params.get("status"));
    const nextWarehouseId = salesWarehouseFromParam(params.get("warehouse_id"));

    if (nextStatus !== null) {
      setStatus(nextStatus);
    }
    if (nextWarehouseId !== null) {
      setFilterWarehouseId(nextWarehouseId);
    }
  }, []);

  function handleCustomerChange(nextCustomerId: string) {
    const customer = salesCustomerOptions.find((candidate) => candidate.value === nextCustomerId) ?? salesCustomerOptions[0];
    setCustomerId(customer.value);
    setOrderChannel(customer.channel);
  }

  function handleLineItemChange(nextItemId: string) {
    const item = salesItemOptions.find((candidate) => candidate.value === nextItemId) ?? salesItemOptions[0];
    setLineItemId(item.value);
    setLineUnitPrice(item.defaultUnitPrice);
  }

  function handleAddLine(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const nextLine: SalesOrderLineInput = {
      lineNo: draftLines.length + 1,
      itemId: selectedLineItem.value,
      orderedQty: lineQty,
      uomCode: selectedLineItem.baseUomCode,
      unitPrice: lineUnitPrice,
      currencyCode: "VND",
      lineDiscountAmount: lineDiscount
    };
    setDraftLines((current) => [...current, nextLine]);
    setLineQty("1");
    setLineDiscount("0");
    setFeedback({ tone: "info", message: salesCopy("feedback.lineAdded", { sku: selectedLineItem.skuCode }) });
  }

  async function handleCreateOrder(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (busyAction || draftLines.length === 0) {
      setFeedback({ tone: "danger", message: salesCopy("feedback.addLineRequired") });
      return;
    }

    setBusyAction("create");
    setFeedback(null);
    try {
      const input: CreateSalesOrderInput = {
        customerId: selectedCustomer.value,
        channel: orderChannel || selectedCustomer.channel,
        warehouseId: orderWarehouseId,
        orderDate,
        currencyCode: "VND",
        note,
        lines: draftLines
      };
      const order = await createSalesOrder(input);
      upsertLocalOrder(order);
      setSelectedOrderId(order.id);
      setDraftLines([]);
      setFeedback({ tone: "success", message: salesCopy("feedback.created", { orderNo: order.orderNo }) });
    } catch (reason) {
      setFeedback({ tone: "danger", message: salesErrorMessage(reason, "feedback.createFailed") });
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
      const order = await updateSalesOrder(selectedOrder.id, {
        expectedVersion: selectedOrder.version,
        lines: draftLines
      });
      upsertLocalOrder(order);
      setSelectedOrderId(order.id);
      setDraftLines([]);
      setFeedback({ tone: "success", message: salesCopy("feedback.linesUpdated", { orderNo: order.orderNo }) });
    } catch (reason) {
      setFeedback({ tone: "danger", message: salesErrorMessage(reason, "feedback.updateFailed") });
    } finally {
      setBusyAction("");
    }
  }

  async function runAction(action: "confirm" | "cancel") {
    if (!selectedOrder || busyAction) {
      return;
    }
    const reason = action === "cancel" ? salesCopy("feedback.cancelReason") : undefined;
    setBusyAction(`${action}:${selectedOrder.id}`);
    setFeedback(null);
    try {
      const result =
        action === "confirm"
          ? await confirmSalesOrder(selectedOrder.id, selectedOrder.version)
          : await cancelSalesOrder(selectedOrder.id, reason ?? "", selectedOrder.version);
      upsertLocalOrder(result.salesOrder);
      setSelectedOrderId(result.salesOrder.id);
      setFeedback({
        tone: action === "confirm" ? "success" : "warning",
        message: salesCopy("feedback.actionDone", {
          orderNo: result.salesOrder.orderNo,
          status: salesStatusLabel(result.currentStatus)
        })
      });
    } catch (reason) {
      setFeedback({ tone: "danger", message: salesErrorMessage(reason, "feedback.actionFailed") });
    } finally {
      setBusyAction("");
    }
  }

  function upsertLocalOrder(order: SalesOrder) {
    setLocalOrders((current) => [order, ...current.filter((candidate) => candidate.id !== order.id)]);
  }

  return (
    <section className="erp-module-page erp-sales-page">
      <header className="erp-page-header">
        <div>
          <p className="erp-module-eyebrow">SO</p>
          <h1 className="erp-page-title">{salesCopy("orders")}</h1>
          <p className="erp-page-description">{salesCopy("description")}</p>
        </div>
        <div className="erp-page-actions">
          <a className="erp-button erp-button--secondary" href="#sales-create">
            {salesCopy("actions.create")}
          </a>
          <a className="erp-button erp-button--primary" href="#sales-list">
            {salesCopy("actions.orders")}
          </a>
        </div>
      </header>

      <section className="erp-sales-toolbar" aria-label={salesCopy("filters.label")}>
        <label className="erp-field">
          <span>{salesCopy("filters.search")}</span>
          <input className="erp-input" type="search" value={search} onChange={(event) => setSearch(event.target.value)} />
        </label>
        <label className="erp-field">
          <span>{salesCopy("filters.status")}</span>
          <select className="erp-input" value={status} onChange={(event) => setStatus(event.target.value as StatusFilter)}>
            {salesStatusOptions.map((option) => (
              <option key={option.value || "all"} value={option.value}>
                {option.value ? salesStatusLabel(option.value) : salesCopy("filters.allStatuses")}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>{salesCopy("filters.channel")}</span>
          <select className="erp-input" value={filterChannel} onChange={(event) => setFilterChannel(event.target.value)}>
            <option value="">{salesCopy("filters.allChannels")}</option>
            {salesChannelOptions.map((option) => (
              <option key={option} value={option}>
                {option}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>{salesCopy("filters.warehouse")}</span>
          <select className="erp-input" value={filterWarehouseId} onChange={(event) => setFilterWarehouseId(event.target.value)}>
            {salesWarehouseOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {salesWarehouseLabel(option.value, option.label)}
              </option>
            ))}
          </select>
        </label>
      </section>

      <section className="erp-kpi-grid erp-sales-kpis" aria-label={salesCopy("orders")}>
        <KPI label={salesCopy("kpi.orders")} value={String(totals.count)} />
        <KPI label={salesCopy("kpi.draft")} value={String(totals.draft)} tone={totals.draft > 0 ? "warning" : "normal"} />
        <KPI label={salesCopy("kpi.confirmed")} value={String(totals.confirmed)} tone="info" />
        <KPI label={salesCopy("kpi.total")} value={formatSalesMoney(totals.totalAmount)} tone="success" />
      </section>

      {feedback ? (
        <p className={`erp-sales-feedback erp-sales-feedback--${feedback.tone}`} role="status">
          {feedback.message}
        </p>
      ) : null}

      <section className="erp-sales-workspace">
        <section className="erp-card erp-card--padded erp-sales-create" id="sales-create">
          <header className="erp-section-header">
            <h2 className="erp-section-title">{salesCopy("sections.createOrder")}</h2>
            <StatusChip tone="info">{salesCopy("lineCount", { count: draftLines.length })}</StatusChip>
          </header>
          <form className="erp-sales-form-grid" onSubmit={handleCreateOrder}>
            <label className="erp-field">
              <span>{salesCopy("fields.customer")}</span>
              <select className="erp-input" value={customerId} onChange={(event) => handleCustomerChange(event.target.value)}>
                {salesCustomerOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </label>
            <label className="erp-field">
              <span>{salesCopy("fields.channel")}</span>
              <select className="erp-input" value={orderChannel} onChange={(event) => setOrderChannel(event.target.value)}>
                {salesChannelOptions.map((option) => (
                  <option key={option} value={option}>
                    {option}
                  </option>
                ))}
              </select>
            </label>
            <label className="erp-field">
              <span>{salesCopy("fields.warehouse")}</span>
              <select className="erp-input" value={orderWarehouseId} onChange={(event) => setOrderWarehouseId(event.target.value)}>
                {salesWarehouseOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {salesWarehouseLabel(option.value, option.label)}
                  </option>
                ))}
              </select>
            </label>
            <label className="erp-field">
              <span>{salesCopy("fields.orderDate")}</span>
              <input className="erp-input" type="date" value={orderDate} onChange={(event) => setOrderDate(event.target.value)} />
            </label>
            <label className="erp-field erp-sales-note-field">
              <span>{salesCopy("fields.note")}</span>
              <input className="erp-input" value={note} onChange={(event) => setNote(event.target.value)} />
            </label>
            <button className="erp-button erp-button--primary" type="submit" disabled={busyAction === "create"}>
              {salesCopy("actions.createOrder")}
            </button>
          </form>

          <form className="erp-sales-line-editor" onSubmit={handleAddLine}>
            <label className="erp-field">
              <span>{salesCopy("fields.lineItem")}</span>
              <select className="erp-input" value={lineItemId} onChange={(event) => handleLineItemChange(event.target.value)}>
                {salesItemOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </label>
            <label className="erp-field">
              <span>{salesCopy("fields.qty")}</span>
              <input className="erp-input" inputMode="decimal" value={lineQty} onChange={(event) => setLineQty(event.target.value)} />
            </label>
            <label className="erp-field">
              <span>{salesCopy("fields.unitPrice")}</span>
              <input
                className="erp-input"
                inputMode="decimal"
                value={lineUnitPrice}
                onChange={(event) => setLineUnitPrice(event.target.value)}
              />
            </label>
            <label className="erp-field">
              <span>{salesCopy("fields.discount")}</span>
              <input
                className="erp-input"
                inputMode="decimal"
                value={lineDiscount}
                onChange={(event) => setLineDiscount(event.target.value)}
              />
            </label>
            <button className="erp-button erp-button--secondary" type="submit">
              {salesCopy("actions.addLine")}
            </button>
          </form>

          <DraftLineList lines={draftLines} onRemove={(index) => setDraftLines((current) => current.filter((_, i) => i !== index))} />
          <div className="erp-sales-actions">
            <button
              className="erp-button erp-button--secondary"
              type="button"
              disabled={!selectedOrder || selectedOrder.status !== "draft" || draftLines.length === 0}
              onClick={handleReplaceDraftLines}
            >
              {salesCopy("actions.replaceDraftLines")}
            </button>
          </div>
        </section>

        <section className="erp-card erp-card--padded erp-sales-detail" id="sales-detail">
          <header className="erp-section-header">
            <h2 className="erp-section-title">{salesCopy("sections.detail")}</h2>
            {selectedOrder ? <StatusChip tone={salesOrderStatusTone(selectedOrder.status)}>{salesStatusLabel(selectedOrder.status)}</StatusChip> : null}
          </header>
          {selectedOrder ? (
            <>
              <div className="erp-sales-detail-grid">
                <Fact label={salesCopy("fields.order")} value={selectedOrder.orderNo} />
                <Fact label={salesCopy("fields.customer")} value={selectedOrder.customerName} />
                <Fact label={salesCopy("fields.warehouse")} value={selectedOrder.warehouseCode ?? "-"} />
                <Fact label={salesCopy("fields.date")} value={formatSalesDate(selectedOrder.orderDate)} />
                <Fact label={salesCopy("fields.total")} value={formatSalesMoney(selectedOrder.totalAmount, selectedOrder.currencyCode)} />
                <Fact label={salesCopy("fields.version")} value={String(selectedOrder.version)} />
              </div>
              <div className="erp-sales-actions">
                <button
                  className="erp-button erp-button--primary"
                  type="button"
                  disabled={selectedOrder.status !== "draft" || Boolean(busyAction)}
                  onClick={() => runAction("confirm")}
                >
                  {salesCopy("actions.confirm")}
                </button>
                <button
                  className="erp-button erp-button--danger"
                  type="button"
                  disabled={!["draft", "confirmed"].includes(selectedOrder.status) || Boolean(busyAction)}
                  onClick={() => runAction("cancel")}
                >
                  {salesCopy("actions.cancel")}
                </button>
              </div>
              <div className="erp-sales-subsection">
                <h3 className="erp-section-title">{salesCopy("sections.lineItems")}</h3>
                <DataTable columns={lineColumns} rows={selectedOrder.lines} getRowKey={(row) => row.id} />
              </div>
            </>
          ) : (
            <EmptyState title={salesCopy("empty.noSelected")} />
          )}
        </section>
      </section>

      <section id="sales-list">
        <DataTable
          columns={orderColumns((order) => setSelectedOrderId(order.id))}
          rows={visibleOrders}
          getRowKey={(row) => row.id}
          loading={loading}
          error={error?.message}
          emptyState={<EmptyState title={salesCopy("empty.noOrders")} />}
        />
      </section>
    </section>
  );
}

function DraftLineList({ lines, onRemove }: { lines: SalesOrderLineInput[]; onRemove: (index: number) => void }) {
  if (lines.length === 0) {
    return <p className="erp-sales-empty-line">{salesCopy("empty.noDraftLines")}</p>;
  }

  return (
    <ol className="erp-sales-draft-lines" aria-label={salesCopy("draftLines")}>
      {lines.map((line, index) => {
        const item = salesItemOptions.find((candidate) => candidate.value === line.itemId) ?? salesItemOptions[0];

        return (
          <li key={`${line.itemId}-${index}`}>
            <span>
              <strong>{item.skuCode}</strong>
              <small>{line.orderedQty} {line.uomCode} / {formatSalesMoney(line.unitPrice)}</small>
            </span>
            <button className="erp-button erp-button--secondary" type="button" onClick={() => onRemove(index)}>
              {salesCopy("actions.remove")}
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

function salesCopy(key: string, values?: Record<string, string | number>) {
  return t(`sales.${key}`, { values });
}

function salesStatusLabel(status: SalesOrderStatus) {
  return salesCopy(`status.${status}`, { status });
}

function salesWarehouseLabel(value: string, fallback: string) {
  return t(`sales.warehouse.${value}`, { fallback });
}

function salesErrorMessage(reason: unknown, fallbackKey: string) {
  if (!(reason instanceof Error)) {
    return salesCopy(fallbackKey);
  }

  switch (reason.message) {
    case "At least one line item is required":
      return salesCopy("feedback.addLineRequired");
    default:
      return salesCopy(fallbackKey);
  }
}

function salesStatusFromParam(value: string | null): StatusFilter | null {
  if (value === null) {
    return null;
  }
  if (salesStatusOptions.some((option) => option.value === value)) {
    return value as StatusFilter;
  }

  return null;
}

function salesWarehouseFromParam(value: string | null) {
  if (value === null) {
    return null;
  }
  if (value === "wh-hcm") {
    return "wh-hcm-fg";
  }

  return salesWarehouseOptions.some((option) => option.value === value) ? value : null;
}

function Fact({ label, value }: { label: string; value: string }) {
  return (
    <div className="erp-sales-fact">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function mergeOrders(localOrders: SalesOrder[], fetchedOrders: SalesOrder[], query: SalesOrderQuery) {
  const localMatches = localOrders.filter((order) => matchesOrderQuery(order, query));
  const localIds = new Set(localMatches.map((order) => order.id));

  return [...localMatches, ...fetchedOrders.filter((order) => !localIds.has(order.id))];
}

function matchesOrderQuery(order: SalesOrder, query: SalesOrderQuery) {
  const search = query.search?.trim().toLowerCase();
  if (search) {
    const haystack = [order.orderNo, order.customerCode, order.customerName, order.channel].join(" ").toLowerCase();
    if (!haystack.includes(search)) {
      return false;
    }
  }
  if (query.status && order.status !== query.status) {
    return false;
  }
  if (query.channel && order.channel !== query.channel) {
    return false;
  }
  if (query.customerId && order.customerId !== query.customerId) {
    return false;
  }
  if (query.warehouseId && order.warehouseId !== query.warehouseId) {
    return false;
  }

  return true;
}

function summarizeOrders(orders: SalesOrder[]) {
  return orders.reduce(
    (summary, order) => ({
      count: summary.count + 1,
      draft: summary.draft + (order.status === "draft" ? 1 : 0),
      confirmed: summary.confirmed + (["confirmed", "reserved"].includes(order.status) ? 1 : 0),
      totalAmount: addMoneyStrings(summary.totalAmount, order.totalAmount)
    }),
    { count: 0, draft: 0, confirmed: 0, totalAmount: "0.00" }
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
