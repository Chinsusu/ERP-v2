"use client";

import { useMemo, useState } from "react";
import type { AuditLogItem } from "@/modules/audit/types";
import { DataTable, StatusChip, type DataTableColumn, type StatusTone } from "@/shared/design-system/components";
import {
  createSubcontractMaterialTransfer,
  formatSubcontractAttachmentType,
  formatSubcontractTransferStatus,
  prototypeTransferLines,
  subcontractTransferStatusTone,
  subcontractTransferWarehouseOptions,
  summarizeSubcontractMaterialTransfers
} from "../services/subcontractMaterialTransferService";
import {
  changeSubcontractOrderStatus,
  createSubcontractOrder,
  formatSubcontractDepositStatus,
  formatSubcontractOrderStatus,
  prototypeSubcontractOrders,
  subcontractDepositStatusOptions,
  subcontractFactoryOptions,
  subcontractOrderStatusOptions,
  subcontractOrderStatusTone,
  subcontractProductOptions,
  summarizeSubcontractOrders
} from "../services/subcontractOrderService";
import type {
  SubcontractDepositStatus,
  SubcontractFinalPaymentStatus,
  SubcontractMaterialTransfer,
  SubcontractOrder,
  SubcontractOrderStatus
} from "../types";

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
    key: "factory",
    header: "Factory",
    render: (row) => row.factoryName,
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
    key: "status",
    header: "Status",
    render: (row) => (
      <StatusChip tone={subcontractOrderStatusTone(row.status)}>{formatSubcontractOrderStatus(row.status)}</StatusChip>
    ),
    width: "170px"
  },
  {
    key: "eta",
    header: "ETA",
    render: (row) => row.expectedDeliveryDate,
    width: "130px"
  },
  {
    key: "qc",
    header: "QC",
    render: (row) => <StatusChip tone={subcontractQcTone(row)}>{subcontractQcLabel(row)}</StatusChip>,
    width: "150px"
  },
  {
    key: "payment",
    header: "Payment",
    render: (row) => formatFinalPaymentStatus(row.finalPaymentStatus),
    width: "140px"
  },
  {
    key: "action",
    header: "Action",
    render: () => <span className="erp-button erp-button--secondary erp-button--compact">Open</span>,
    width: "110px"
  }
];

const subcontractTimelineSteps = [
  "Order",
  "Deposit",
  "Transfer",
  "Sample",
  "Mass production",
  "Inbound",
  "QC",
  "Close"
];

const subcontractDetailTabs = [
  "Overview",
  "Transfer",
  "Sample approval",
  "Mass production",
  "Inbound & QC",
  "Factory claim",
  "Payment"
];

const prototypeSampleApprovals = [
  {
    sampleNo: "M01",
    receivedAt: "2026-04-14",
    reviewer: "QA A",
    result: "Fail",
    fileLabel: "photo-m01",
    note: "Scent is not within approved range"
  },
  {
    sampleNo: "M02",
    receivedAt: "2026-04-16",
    reviewer: "QA A",
    result: "Pass",
    fileLabel: "photo-m02",
    note: "Approved retained sample"
  }
] as const;

const prototypeFactoryClaims = [
  {
    issueType: "Packaging scuff",
    affectedQty: 120,
    detectedAt: "2026-04-26",
    responseDeadline: "2026-04-30",
    severity: "P1",
    status: "Open",
    slaLabel: "4 days"
  }
] as const;

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
  const [search, setSearch] = useState("");
  const [factoryFilter, setFactoryFilter] = useState("all");
  const [productFilter, setProductFilter] = useState("all");
  const [statusFilter, setStatusFilter] = useState<SubcontractOrderStatus | "all">("all");
  const [etaFilter, setEtaFilter] = useState("");
  const [feedback, setFeedback] = useState<{ tone: StatusTone; message: string } | null>(null);
  const [lastAudit, setLastAudit] = useState<AuditLogItem | null>(null);
  const [sourceWarehouseId, setSourceWarehouseId] = useState<string>(subcontractTransferWarehouseOptions[0].value);
  const [signedHandover, setSignedHandover] = useState(false);
  const [transfers, setTransfers] = useState<SubcontractMaterialTransfer[]>([]);
  const [transferFeedback, setTransferFeedback] = useState<{ tone: StatusTone; message: string } | null>(null);
  const summary = useMemo(() => summarizeSubcontractOrders(orders), [orders]);
  const transferSummary = useMemo(() => summarizeSubcontractMaterialTransfers(transfers), [transfers]);
  const displayedOrders = useMemo(
    () =>
      orders.filter((order) => {
        const searchValue = search.trim().toLowerCase();
        if (
          searchValue &&
          ![order.orderNo, order.factoryName, order.productName, order.sku].some((value) =>
            value.toLowerCase().includes(searchValue)
          )
        ) {
          return false;
        }
        if (factoryFilter !== "all" && order.factoryId !== factoryFilter) {
          return false;
        }
        if (productFilter !== "all" && order.productId !== productFilter) {
          return false;
        }
        if (statusFilter !== "all" && order.status !== statusFilter) {
          return false;
        }
        if (etaFilter && order.expectedDeliveryDate !== etaFilter) {
          return false;
        }

        return true;
      }),
    [etaFilter, factoryFilter, orders, productFilter, search, statusFilter]
  );
  const selectedOrder = orders.find((order) => order.id === selectedOrderId) ?? orders[0] ?? null;
  const latestTransfer = transfers[0] ?? null;

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

  function handleCreateTransfer() {
    if (!selectedOrder) {
      return;
    }

    const warehouse =
      subcontractTransferWarehouseOptions.find((option) => option.value === sourceWarehouseId) ??
      subcontractTransferWarehouseOptions[0];

    try {
      const transfer = createSubcontractMaterialTransfer({
        order: selectedOrder,
        sourceWarehouseId: warehouse.value,
        sourceWarehouseCode: warehouse.code,
        signedHandover,
        lines: prototypeTransferLines
      });
      const statusResult = changeSubcontractOrderStatus({
        order: selectedOrder,
        nextStatus: "MATERIAL_TRANSFERRED",
        actorName: "Subcontract Coordinator",
        note: `${transfer.transferNo} created with SUBCONTRACT_ISSUE movement`
      });

      setTransfers((current) => [transfer, ...current]);
      setOrders((current) => current.map((order) => (order.id === statusResult.order.id ? statusResult.order : order)));
      setSelectedOrderId(statusResult.order.id);
      setLastAudit(statusResult.auditLog);
      setTransferFeedback({
        tone: subcontractTransferStatusTone(transfer.status),
        message: `${transfer.transferNo} created with ${transfer.stockMovements.length} SUBCONTRACT_ISSUE movements`
      });
    } catch (error) {
      setTransferFeedback({
        tone: "danger",
        message: error instanceof Error ? error.message : "Material transfer could not be created"
      });
    }
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
          <a className="erp-button erp-button--secondary" href="#subcontract-transfer">
            Transfer
          </a>
          <a className="erp-button erp-button--primary" href="#subcontract-orders">
            Orders
          </a>
        </div>
      </header>

      <section className="erp-subcontract-list-toolbar" aria-label="Subcontract order filters">
        <label className="erp-field">
          <span>Search</span>
          <input
            className="erp-input"
            type="search"
            value={search}
            placeholder="SUB-260426-0001 / factory / SKU"
            onChange={(event) => setSearch(event.target.value)}
          />
        </label>
        <label className="erp-field">
          <span>Factory</span>
          <select className="erp-input" value={factoryFilter} onChange={(event) => setFactoryFilter(event.target.value)}>
            <option value="all">All factories</option>
            {subcontractFactoryOptions.map((factory) => (
              <option key={factory.id} value={factory.id}>
                {factory.name}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Product</span>
          <select className="erp-input" value={productFilter} onChange={(event) => setProductFilter(event.target.value)}>
            <option value="all">All products</option>
            {subcontractProductOptions.map((product) => (
              <option key={product.id} value={product.id}>
                {product.name}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Status</span>
          <select
            className="erp-input"
            value={statusFilter}
            onChange={(event) => setStatusFilter(event.target.value as SubcontractOrderStatus | "all")}
          >
            <option value="all">All statuses</option>
            {subcontractOrderStatusOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>ETA</span>
          <input className="erp-input" type="date" value={etaFilter} onChange={(event) => setEtaFilter(event.target.value)} />
        </label>
      </section>

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

              <div className="erp-subcontract-timeline" aria-label="Subcontract manufacturing timeline">
                {subcontractTimelineSteps.map((step, index) => (
                  <span
                    className="erp-subcontract-timeline-step"
                    data-active={index <= timelineIndexForStatus(selectedOrder.status) ? "true" : "false"}
                    key={step}
                  >
                    {step}
                  </span>
                ))}
              </div>

              <div className="erp-subcontract-tabs" aria-label="Subcontract detail tabs">
                {subcontractDetailTabs.map((tab, index) => (
                  <button className="erp-subcontract-tab" data-active={index === 0 ? "true" : "false"} key={tab} type="button">
                    {tab}
                  </button>
                ))}
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

      <section className="erp-subcontract-transfer-grid" id="subcontract-transfer">
        <div className="erp-card erp-card--padded erp-subcontract-card">
          <div className="erp-section-header">
            <h2 className="erp-section-title">Material transfer</h2>
            <StatusChip tone="info">{selectedOrder?.orderNo ?? "No order"}</StatusChip>
          </div>

          <div className="erp-subcontract-transfer-controls">
            <label className="erp-field">
              <span>Source warehouse</span>
              <select
                className="erp-input"
                value={sourceWarehouseId}
                onChange={(event) => setSourceWarehouseId(event.target.value)}
              >
                {subcontractTransferWarehouseOptions.map((warehouse) => (
                  <option key={warehouse.value} value={warehouse.value}>
                    {warehouse.label}
                  </option>
                ))}
              </select>
            </label>
            <label className="erp-subcontract-sample-toggle">
              <input
                checked={signedHandover}
                type="checkbox"
                onChange={(event) => setSignedHandover(event.target.checked)}
              />
              <span>Signed handover</span>
            </label>
          </div>

          <div className="erp-subcontract-line-list" aria-label="Material and packaging lines">
            {prototypeTransferLines.map((line) => (
              <div className="erp-subcontract-line-item" key={line.id}>
                <strong>{line.itemName}</strong>
                <span>
                  {line.itemCode} / {formatQuantity(line.quantity)} {line.unit}
                </span>
                <StatusChip tone={line.lotControlled ? "warning" : "normal"}>
                  {line.lotControlled ? line.batchNo : "No lot"}
                </StatusChip>
                <StatusChip tone={line.qcStatus === "passed" ? "success" : "warning"}>{line.qcStatus}</StatusChip>
              </div>
            ))}
          </div>

          <div className="erp-subcontract-actions">
            <button className="erp-button erp-button--primary" type="button" onClick={handleCreateTransfer}>
              Create transfer
            </button>
            {transferFeedback ? <StatusChip tone={transferFeedback.tone}>{transferFeedback.message}</StatusChip> : null}
          </div>
        </div>

        <div className="erp-card erp-card--padded erp-subcontract-card">
          <div className="erp-section-header">
            <h2 className="erp-section-title">Transfer result</h2>
            <StatusChip tone={latestTransfer ? subcontractTransferStatusTone(latestTransfer.status) : "normal"}>
              {latestTransfer ? formatSubcontractTransferStatus(latestTransfer.status) : `${transferSummary.total} docs`}
            </StatusChip>
          </div>

          {latestTransfer ? (
            <>
              <div className="erp-subcontract-fact-grid">
                <SubcontractFact label="Transfer" value={latestTransfer.transferNo} />
                <SubcontractFact label="Order" value={latestTransfer.orderNo} />
                <SubcontractFact label="Source" value={latestTransfer.sourceWarehouseCode} />
                <SubcontractFact label="Factory" value={latestTransfer.factoryName} />
                <SubcontractFact label="Signed" value={latestTransfer.signedHandover ? "Yes" : "No"} />
                <SubcontractFact label="Movements" value={String(latestTransfer.stockMovements.length)} />
              </div>
              <div className="erp-subcontract-attachment-list" aria-label="Attachment placeholders">
                {latestTransfer.attachmentPlaceholders.map((placeholder) => (
                  <StatusChip key={placeholder.type} tone={placeholder.required ? "warning" : "normal"}>
                    {formatSubcontractAttachmentType(placeholder.type)}
                  </StatusChip>
                ))}
              </div>
              <div className="erp-subcontract-movement-list" aria-label="Stock movements">
                {latestTransfer.stockMovements.map((movement) => (
                  <div className="erp-subcontract-movement-item" key={movement.id}>
                    <strong>{movement.movementType}</strong>
                    <span>
                      {movement.itemCode} / {formatQuantity(movement.quantity)} {movement.unit} /{" "}
                      {movement.targetLocation}
                    </span>
                  </div>
                ))}
              </div>
            </>
          ) : (
            <div className="erp-subcontract-empty-state">No material transfer created</div>
          )}
        </div>
      </section>

      <section className="erp-subcontract-quality-grid">
        <div className="erp-card erp-card--padded erp-subcontract-card">
          <div className="erp-section-header">
            <h2 className="erp-section-title">Sample approval</h2>
            <StatusChip tone="success">Sample approved</StatusChip>
          </div>
          <div className="erp-subcontract-sample-list">
            {prototypeSampleApprovals.map((sample) => (
              <div className="erp-subcontract-sample-item" key={sample.sampleNo}>
                <strong>{sample.sampleNo}</strong>
                <span>{sample.receivedAt}</span>
                <span>{sample.reviewer}</span>
                <StatusChip tone={sample.result === "Pass" ? "success" : "danger"}>{sample.result}</StatusChip>
                <span>{sample.fileLabel}</span>
                <span>{sample.note}</span>
              </div>
            ))}
          </div>
          <div className="erp-subcontract-actions">
            <button className="erp-button erp-button--secondary" type="button">
              Approve sample
            </button>
          </div>
        </div>

        <div className="erp-card erp-card--padded erp-subcontract-card">
          <div className="erp-section-header">
            <h2 className="erp-section-title">Factory claim</h2>
            <StatusChip tone="danger">P1 open</StatusChip>
          </div>
          <div className="erp-subcontract-claim-list">
            {prototypeFactoryClaims.map((claim) => (
              <div className="erp-subcontract-claim-item" key={`${claim.issueType}-${claim.detectedAt}`}>
                <strong>{claim.issueType}</strong>
                <span>{formatQuantity(claim.affectedQty)} affected</span>
                <span>Detected {claim.detectedAt}</span>
                <span>Deadline {claim.responseDeadline}</span>
                <StatusChip tone="danger">{claim.severity}</StatusChip>
                <StatusChip tone="warning">{claim.status}</StatusChip>
                <StatusChip tone="warning">SLA {claim.slaLabel}</StatusChip>
              </div>
            ))}
          </div>
          <div className="erp-subcontract-actions">
            <button className="erp-button erp-button--danger" type="button">
              Create claim
            </button>
          </div>
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
          <StatusChip tone={displayedOrders.length === 0 ? "warning" : "info"}>{displayedOrders.length} rows</StatusChip>
        </div>
        <DataTable columns={orderColumns} rows={displayedOrders} getRowKey={(row) => row.id} />
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

function subcontractQcLabel(order: SubcontractOrder) {
  if (order.status === "ACCEPTED" || order.status === "CLOSED") {
    return "QC pass";
  }
  if (order.status === "REJECTED") {
    return "Claim hold";
  }
  if (order.status === "QC_REVIEW") {
    return "QC review";
  }
  if (order.sampleRequired) {
    return "Sample pending";
  }

  return "Not required";
}

function subcontractQcTone(order: SubcontractOrder): StatusTone {
  if (order.status === "ACCEPTED" || order.status === "CLOSED") {
    return "success";
  }
  if (order.status === "REJECTED") {
    return "danger";
  }
  if (order.status === "QC_REVIEW" || order.sampleRequired) {
    return "warning";
  }

  return "normal";
}

function formatFinalPaymentStatus(status: SubcontractFinalPaymentStatus) {
  switch (status) {
    case "released":
      return "Released";
    case "hold":
      return "Hold";
    case "pending":
    default:
      return "Pending";
  }
}

function timelineIndexForStatus(status: SubcontractOrderStatus) {
  switch (status) {
    case "DRAFT":
      return 0;
    case "CONFIRMED":
      return 1;
    case "MATERIAL_TRANSFERRED":
      return 2;
    case "SAMPLE_APPROVED":
      return 3;
    case "IN_PRODUCTION":
      return 4;
    case "DELIVERED":
      return 5;
    case "QC_REVIEW":
    case "ACCEPTED":
    case "REJECTED":
      return 6;
    case "CLOSED":
    default:
      return 7;
  }
}
