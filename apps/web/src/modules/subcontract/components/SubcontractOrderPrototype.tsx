"use client";

import { useMemo, useState } from "react";
import type { AuditLogItem } from "@/modules/audit/types";
import { DataTable, StatusChip, type DataTableColumn, type StatusTone } from "@/shared/design-system/components";
import {
  changeSubcontractOrderStatus,
  createSubcontractOrder,
  formatSubcontractDepositStatus,
  formatSubcontractOrderStatus,
  prototypeSubcontractOrders,
  subcontractDepositStatusOptions,
  subcontractDepositStatusTone,
  subcontractFactoryOptions,
  subcontractOrderStatusOptions,
  subcontractOrderStatusTone,
  subcontractProductOptions,
  summarizeSubcontractOrders
} from "../services/subcontractOrderService";
import type { SubcontractDepositStatus, SubcontractOrder, SubcontractOrderStatus } from "../types";

const orderColumns: DataTableColumn<SubcontractOrder>[] = [
  {
    key: "order",
    header: "Order",
    render: (row) => (
      <span className="erp-subcontract-order-cell">
        <strong>{row.orderNo}</strong>
        <small>{row.factoryName}</small>
      </span>
    ),
    width: "190px"
  },
  {
    key: "product",
    header: "Product",
    render: (row) => (
      <span className="erp-subcontract-order-cell">
        <strong>{row.productName}</strong>
        <small>{row.sku}</small>
      </span>
    ),
    width: "210px"
  },
  {
    key: "quantity",
    header: "Qty",
    render: (row) => formatQuantity(row.quantity),
    align: "right",
    width: "110px"
  },
  {
    key: "spec",
    header: "Spec",
    render: (row) => row.specVersion,
    width: "180px"
  },
  {
    key: "sample",
    header: "Sample",
    render: (row) => (row.sampleRequired ? "Required" : "Not required"),
    width: "130px"
  },
  {
    key: "expected",
    header: "Expected",
    render: (row) => row.expectedDeliveryDate,
    width: "130px"
  },
  {
    key: "deposit",
    header: "Deposit",
    render: (row) => (
      <StatusChip tone={subcontractDepositStatusTone(row.depositStatus)}>
        {formatSubcontractDepositStatus(row.depositStatus)}
      </StatusChip>
    ),
    width: "140px"
  },
  {
    key: "status",
    header: "Status",
    render: (row) => (
      <StatusChip tone={subcontractOrderStatusTone(row.status)}>{formatSubcontractOrderStatus(row.status)}</StatusChip>
    ),
    width: "170px"
  }
];

export function SubcontractOrderPrototype() {
  const [factoryId, setFactoryId] = useState(subcontractFactoryOptions[0].id);
  const [productId, setProductId] = useState(subcontractProductOptions[0].id);
  const [quantity, setQuantity] = useState("2500");
  const [specVersion, setSpecVersion] = useState("SPEC-CREAM-50ML-v4");
  const [sampleRequired, setSampleRequired] = useState(true);
  const [expectedDeliveryDate, setExpectedDeliveryDate] = useState("2026-05-20");
  const [depositStatus, setDepositStatus] = useState<SubcontractDepositStatus>("pending");
  const [depositAmount, setDepositAmount] = useState("5000000");
  const [orders, setOrders] = useState<SubcontractOrder[]>(prototypeSubcontractOrders);
  const [selectedOrderId, setSelectedOrderId] = useState(prototypeSubcontractOrders[0]?.id ?? "");
  const [feedback, setFeedback] = useState<{ tone: StatusTone; message: string } | null>(null);
  const [lastAudit, setLastAudit] = useState<AuditLogItem | null>(null);
  const summary = useMemo(() => summarizeSubcontractOrders(orders), [orders]);
  const selectedOrder = orders.find((order) => order.id === selectedOrderId) ?? orders[0] ?? null;

  function handleCreateOrder() {
    try {
      const order = createSubcontractOrder({
        factoryId,
        productId,
        quantity: Number(quantity),
        specVersion,
        sampleRequired,
        expectedDeliveryDate,
        depositStatus,
        depositAmount: depositAmount.trim() === "" ? undefined : Number(depositAmount)
      });

      setOrders((current) => [order, ...current]);
      setSelectedOrderId(order.id);
      setLastAudit(null);
      setFeedback({
        tone: "success",
        message: `${order.orderNo} created as ${formatSubcontractOrderStatus(order.status)}`
      });
    } catch (error) {
      setFeedback({
        tone: "danger",
        message: error instanceof Error ? error.message : "Subcontract order could not be created"
      });
    }
  }

  function handleStatusChange(nextStatus: SubcontractOrderStatus) {
    if (!selectedOrder) {
      return;
    }

    const result = changeSubcontractOrderStatus({
      order: selectedOrder,
      nextStatus,
      actorName: "Subcontract Coordinator"
    });

    setOrders((current) => current.map((order) => (order.id === result.order.id ? result.order : order)));
    setSelectedOrderId(result.order.id);
    setLastAudit(result.auditLog);
    setFeedback({
      tone: subcontractOrderStatusTone(result.order.status),
      message: `${result.order.orderNo} moved to ${formatSubcontractOrderStatus(result.order.status)}`
    });
  }

  return (
    <section className="erp-module-page erp-subcontract-page">
      <header className="erp-page-header">
        <div>
          <p className="erp-module-eyebrow">SUB</p>
          <h1 className="erp-page-title">Subcontract Orders</h1>
          <p className="erp-page-description">External factory order creation and status audit skeleton</p>
        </div>
        <div className="erp-page-actions">
          <a className="erp-button erp-button--secondary" href="#subcontract-order-form">
            New order
          </a>
          <a className="erp-button erp-button--secondary" href="#subcontract-workflow">
            Workflow
          </a>
          <a className="erp-button erp-button--primary" href="#subcontract-orders">
            Orders
          </a>
        </div>
      </header>

      <section className="erp-kpi-grid erp-subcontract-kpis">
        <SubcontractKPI label="Total orders" value={summary.total} tone="info" />
        <SubcontractKPI label="Active" value={summary.active} tone="warning" />
        <SubcontractKPI label="Accepted" value={summary.accepted} tone="success" />
        <article className="erp-card erp-card--padded erp-kpi-card">
          <div className="erp-kpi-label">Next delivery</div>
          <strong className="erp-kpi-value erp-kpi-value--small">{summary.nextDeliveryDate ?? "-"}</strong>
          <StatusChip tone="normal">Expected</StatusChip>
        </article>
      </section>

      <section className="erp-subcontract-workspace">
        <div className="erp-card erp-card--padded erp-subcontract-card" id="subcontract-order-form">
          <div className="erp-section-header">
            <h2 className="erp-section-title">External factory order</h2>
            <StatusChip tone="warning">Draft</StatusChip>
          </div>

          <div className="erp-subcontract-form-grid">
            <label className="erp-field">
              <span>Factory</span>
              <select className="erp-input" value={factoryId} onChange={(event) => setFactoryId(event.target.value)}>
                {subcontractFactoryOptions.map((factory) => (
                  <option key={factory.id} value={factory.id}>
                    {factory.name}
                  </option>
                ))}
              </select>
            </label>
            <label className="erp-field">
              <span>Product</span>
              <select className="erp-input" value={productId} onChange={(event) => setProductId(event.target.value)}>
                {subcontractProductOptions.map((product) => (
                  <option key={product.id} value={product.id}>
                    {product.name}
                  </option>
                ))}
              </select>
            </label>
            <label className="erp-field">
              <span>Quantity</span>
              <input
                className="erp-input"
                min="1"
                type="number"
                value={quantity}
                onChange={(event) => setQuantity(event.target.value)}
              />
            </label>
            <label className="erp-field">
              <span>Spec</span>
              <input
                className="erp-input"
                type="text"
                value={specVersion}
                onChange={(event) => setSpecVersion(event.target.value)}
              />
            </label>
            <label className="erp-field">
              <span>Expected delivery date</span>
              <input
                className="erp-input"
                type="date"
                value={expectedDeliveryDate}
                onChange={(event) => setExpectedDeliveryDate(event.target.value)}
              />
            </label>
            <label className="erp-field">
              <span>Deposit status</span>
              <select
                className="erp-input"
                value={depositStatus}
                onChange={(event) => setDepositStatus(event.target.value as SubcontractDepositStatus)}
              >
                {subcontractDepositStatusOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </label>
            <label className="erp-field">
              <span>Deposit amount</span>
              <input
                className="erp-input"
                min="0"
                type="number"
                value={depositAmount}
                onChange={(event) => setDepositAmount(event.target.value)}
              />
            </label>
            <label className="erp-subcontract-sample-toggle">
              <input
                checked={sampleRequired}
                type="checkbox"
                onChange={(event) => setSampleRequired(event.target.checked)}
              />
              <span>Sample required</span>
            </label>
          </div>

          <div className="erp-subcontract-actions">
            <button className="erp-button erp-button--primary" type="button" onClick={handleCreateOrder}>
              Create order
            </button>
            {feedback ? <StatusChip tone={feedback.tone}>{feedback.message}</StatusChip> : null}
          </div>
        </div>

        <div className="erp-card erp-card--padded erp-subcontract-card" id="subcontract-workflow">
          <div className="erp-section-header">
            <h2 className="erp-section-title">Status workflow</h2>
            {selectedOrder ? (
              <StatusChip tone={subcontractOrderStatusTone(selectedOrder.status)}>
                {formatSubcontractOrderStatus(selectedOrder.status)}
              </StatusChip>
            ) : null}
          </div>

          {selectedOrder ? (
            <>
              <div className="erp-subcontract-fact-grid">
                <SubcontractFact label="Order" value={selectedOrder.orderNo} />
                <SubcontractFact label="Factory" value={selectedOrder.factoryName} />
                <SubcontractFact label="Product" value={selectedOrder.productName} />
                <SubcontractFact label="Qty" value={formatQuantity(selectedOrder.quantity)} />
                <SubcontractFact label="Spec" value={selectedOrder.specVersion} />
                <SubcontractFact label="Expected" value={selectedOrder.expectedDeliveryDate} />
                <SubcontractFact label="Deposit" value={formatSubcontractDepositStatus(selectedOrder.depositStatus)} />
                <SubcontractFact
                  label="Deposit amount"
                  value={selectedOrder.depositAmount ? formatMoney(selectedOrder.depositAmount) : "-"}
                />
              </div>

              <div className="erp-subcontract-status-grid" aria-label="Subcontract order status model">
                {subcontractOrderStatusOptions.map((option, index) => (
                  <button
                    aria-pressed={selectedOrder.status === option.value}
                    className="erp-subcontract-status-step"
                    data-active={selectedOrder.status === option.value ? "true" : "false"}
                    key={option.value}
                    type="button"
                    onClick={() => handleStatusChange(option.value)}
                  >
                    <span>{index + 1}</span>
                    <strong>{option.label}</strong>
                  </button>
                ))}
              </div>
            </>
          ) : (
            <div className="erp-subcontract-empty-state">No subcontract order selected</div>
          )}
        </div>
      </section>

      <section className="erp-card erp-card--padded erp-subcontract-audit-card">
        <div className="erp-section-header">
          <h2 className="erp-section-title">Status audit</h2>
          <StatusChip tone={lastAudit ? "success" : "normal"}>{lastAudit ? "Written" : "Waiting"}</StatusChip>
        </div>
        {lastAudit ? (
          <div className="erp-subcontract-fact-grid">
            <SubcontractFact label="Action" value={lastAudit.action} />
            <SubcontractFact label="Entity" value={lastAudit.entityType} />
            <SubcontractFact label="Before" value={String(lastAudit.beforeData?.status ?? "-")} />
            <SubcontractFact label="After" value={String(lastAudit.afterData?.status ?? "-")} />
            <SubcontractFact label="Actor" value={String(lastAudit.metadata.actor_name ?? lastAudit.actorId)} />
            <SubcontractFact label="Created at" value={lastAudit.createdAt} />
          </div>
        ) : (
          <div className="erp-subcontract-empty-state">Change a status to write the audit event</div>
        )}
      </section>

      <section className="erp-card erp-card--padded erp-module-table-card" id="subcontract-orders">
        <div className="erp-section-header">
          <h2 className="erp-section-title">External factory orders</h2>
          <StatusChip tone={orders.length === 0 ? "warning" : "info"}>{orders.length} rows</StatusChip>
        </div>
        <DataTable columns={orderColumns} rows={orders} getRowKey={(row) => row.id} />
      </section>
    </section>
  );
}

function SubcontractKPI({ label, value, tone }: { label: string; value: number; tone: StatusTone }) {
  return (
    <article className="erp-card erp-card--padded erp-kpi-card">
      <div className="erp-kpi-label">{label}</div>
      <strong className="erp-kpi-value">{formatQuantity(value)}</strong>
      <StatusChip tone={tone}>{label}</StatusChip>
    </article>
  );
}

function SubcontractFact({ label, value }: { label: string; value: string }) {
  return (
    <div className="erp-subcontract-fact">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function formatQuantity(value: number) {
  return new Intl.NumberFormat("en-US", { maximumFractionDigits: 0 }).format(value);
}

function formatMoney(value: number) {
  return `${new Intl.NumberFormat("en-US", { maximumFractionDigits: 0 }).format(value)} VND`;
}
