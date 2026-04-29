"use client";

import { useEffect, useMemo, useState } from "react";
import type { AuditLogItem } from "@/modules/audit/types";
import { DataTable, StatusChip, type DataTableColumn, type StatusTone } from "@/shared/design-system/components";
import { decimalScales, normalizeDecimalInput } from "@/shared/format/numberFormat";
import {
  formatSubcontractAttachmentType,
  formatSubcontractTransferStatus,
  subcontractTransferStatusTone,
  subcontractTransferWarehouseOptions,
  summarizeSubcontractMaterialTransfers
} from "../services/subcontractMaterialTransferService";
import {
  approveSubcontractOrder,
  approveSubcontractSample,
  availableSubcontractOrderActions,
  cancelSubcontractOrder,
  changeSubcontractOrderStatus,
  closeSubcontractOrder,
  confirmFactorySubcontractOrder,
  createSubcontractOrder,
  formatSubcontractDepositStatus,
  formatSubcontractOrderStatus,
  getSubcontractOrder,
  getSubcontractOrders,
  issueSubcontractMaterials,
  prototypeSubcontractOrders,
  rejectSubcontractSample,
  subcontractDepositStatusOptions,
  subcontractFactoryOptions,
  subcontractMaterialItemOptions,
  subcontractOrderActionOptions,
  subcontractOrderStatusOptions,
  subcontractOrderStatusTone,
  subcontractProductOptions,
  submitSubcontractSample,
  submitSubcontractOrder,
  summarizeSubcontractOrders,
  updateSubcontractOrder
} from "../services/subcontractOrderService";
import type {
  SubcontractDepositStatus,
  SubcontractFinalPaymentStatus,
  SubcontractMaterialTransfer,
  SubcontractOrder,
  SubcontractOrderStatus,
  SubcontractSampleApproval
} from "../types";

function subcontractOrderColumns(onOpen: (order: SubcontractOrder) => void): DataTableColumn<SubcontractOrder>[] {
  return [
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
    render: (row) => (
      <button className="erp-button erp-button--secondary erp-button--compact" type="button" onClick={() => onOpen(row)}>
        Open
      </button>
    ),
    width: "110px"
  }
  ];
}

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
  const [materialItemId, setMaterialItemId] = useState(subcontractMaterialItemOptions[0].id);
  const [materialQty, setMaterialQty] = useState("20");
  const [materialUnitCost, setMaterialUnitCost] = useState(subcontractMaterialItemOptions[0].defaultUnitCost);
  const [orders, setOrders] = useState<SubcontractOrder[]>([]);
  const [selectedOrderId, setSelectedOrderId] = useState("");
  const [isLoadingOrders, setIsLoadingOrders] = useState(true);
  const [isSavingOrder, setIsSavingOrder] = useState(false);
  const [search, setSearch] = useState("");
  const [factoryFilter, setFactoryFilter] = useState("all");
  const [productFilter, setProductFilter] = useState("all");
  const [statusFilter, setStatusFilter] = useState<SubcontractOrderStatus | "all">("all");
  const [etaFilter, setEtaFilter] = useState("");
  const [feedback, setFeedback] = useState<{ tone: StatusTone; message: string } | null>(null);
  const [lastAudit, setLastAudit] = useState<AuditLogItem | null>(null);
  const [sourceWarehouseId, setSourceWarehouseId] = useState<string>(subcontractTransferWarehouseOptions[0].value);
  const [signedHandover, setSignedHandover] = useState(false);
  const [handoverReceiver, setHandoverReceiver] = useState("Factory receiver");
  const [handoverContact, setHandoverContact] = useState("");
  const [handoverEvidenceName, setHandoverEvidenceName] = useState("handover.pdf");
  const [materialIssueDrafts, setMaterialIssueDrafts] = useState<Record<string, { issueQty: string; batchNo: string; sourceBinId: string }>>({});
  const [isCreatingTransfer, setIsCreatingTransfer] = useState(false);
  const [transfers, setTransfers] = useState<SubcontractMaterialTransfer[]>([]);
  const [transferFeedback, setTransferFeedback] = useState<{ tone: StatusTone; message: string } | null>(null);
  const [sampleApprovals, setSampleApprovals] = useState<SubcontractSampleApproval[]>([]);
  const [sampleCode, setSampleCode] = useState("SAMPLE-A");
  const [sampleFormulaVersion, setSampleFormulaVersion] = useState("FORMULA-2026.04");
  const [sampleEvidenceName, setSampleEvidenceName] = useState("sample-front.jpg");
  const [sampleDecisionReason, setSampleDecisionReason] = useState("Approved against retained standard");
  const [sampleStorageStatus, setSampleStorageStatus] = useState("retained_in_qa_cabinet");
  const [sampleFeedback, setSampleFeedback] = useState<{ tone: StatusTone; message: string } | null>(null);
  const [isSavingSample, setIsSavingSample] = useState(false);
  const summary = useMemo(() => summarizeSubcontractOrders(orders), [orders]);
  const transferSummary = useMemo(() => summarizeSubcontractMaterialTransfers(transfers), [transfers]);
  const orderColumns = useMemo(() => subcontractOrderColumns(handleOpenOrder), []);
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
  const latestSampleApproval =
    sampleApprovals.find((sampleApproval) => sampleApproval.orderId === selectedOrder?.id) ?? null;
  const canIssueSelectedOrder =
    selectedOrder?.status === "factory_confirmed" || selectedOrder?.status === "deposit_recorded";
  const canSubmitSample =
    selectedOrder?.sampleRequired &&
    (selectedOrder.status === "materials_issued_to_factory" || selectedOrder.status === "sample_rejected");
  const canDecideSample = selectedOrder?.status === "sample_submitted" && latestSampleApproval?.status === "submitted";

  useEffect(() => {
    if (!selectedOrder) {
      return;
    }
    const firstLine = selectedOrder.materialLines[0];
    setFactoryId(selectedOrder.factoryId);
    setProductId(selectedOrder.productId);
    setQuantity(String(selectedOrder.quantity));
    setSpecVersion(selectedOrder.specVersion);
    setSampleRequired(selectedOrder.sampleRequired);
    setExpectedDeliveryDate(selectedOrder.expectedDeliveryDate);
    setDepositStatus(selectedOrder.depositStatus);
    setDepositAmount(selectedOrder.depositAmount ? String(selectedOrder.depositAmount) : "");
    if (firstLine) {
      setMaterialItemId(firstLine.itemId);
      setMaterialQty(firstLine.plannedQty);
      setMaterialUnitCost(firstLine.unitCost);
    }
  }, [selectedOrder]);

  useEffect(() => {
    if (!selectedOrder) {
      setMaterialIssueDrafts({});
      return;
    }

    setMaterialIssueDrafts(
      Object.fromEntries(
        selectedOrder.materialLines.map((line) => [
          line.id,
          {
            issueQty: remainingMaterialQty(line.plannedQty, line.issuedQty),
            batchNo: line.lotTraceRequired ? `${line.skuCode}-LOT-001` : "",
            sourceBinId: ""
          }
        ])
      )
    );
    setSignedHandover(false);
    setHandoverReceiver("Factory receiver");
  }, [selectedOrder?.id, selectedOrder?.version]);

  useEffect(() => {
    if (!selectedOrder) {
      return;
    }

    setSampleCode(`${selectedOrder.orderNo}-SAMPLE-A`);
    setSampleFormulaVersion("FORMULA-2026.04");
    setSampleEvidenceName("sample-front.jpg");
    setSampleDecisionReason(
      selectedOrder.status === "sample_rejected" && selectedOrder.sampleRejectReason
        ? selectedOrder.sampleRejectReason
        : "Approved against retained standard"
    );
    setSampleStorageStatus("retained_in_qa_cabinet");
  }, [selectedOrder?.id, selectedOrder?.version]);

  useEffect(() => {
    let active = true;

    async function loadOrders() {
      try {
        const loadedOrders = await getSubcontractOrders();
        if (!active) {
          return;
        }
        setOrders(loadedOrders);
        setSelectedOrderId(loadedOrders[0]?.id ?? "");
        setFeedback(null);
      } catch (error) {
        if (!active) {
          return;
        }
        setOrders(prototypeSubcontractOrders);
        setSelectedOrderId(prototypeSubcontractOrders[0]?.id ?? "");
        setFeedback({
          tone: "warning",
          message: error instanceof Error ? error.message : "Loaded prototype subcontract orders"
        });
      } finally {
        if (active) {
          setIsLoadingOrders(false);
        }
      }
    }

    loadOrders();

    return () => {
      active = false;
    };
  }, []);

  async function handleOpenOrder(order: SubcontractOrder) {
    try {
      const detail = await getSubcontractOrder(order.id);
      setOrders((current) => [detail, ...current.filter((candidate) => candidate.id !== detail.id)]);
      setSelectedOrderId(detail.id);
      setFeedback(null);
    } catch (error) {
      setSelectedOrderId(order.id);
      setFeedback({
        tone: "danger",
        message: error instanceof Error ? error.message : "Subcontract order detail could not be loaded"
      });
    }
  }

  async function handleCreateOrder() {
    try {
      setIsSavingOrder(true);
      const order = await createSubcontractOrder({
        factoryId,
        productId,
        quantity: Number(quantity),
        specVersion,
        sampleRequired,
        expectedDeliveryDate,
        depositStatus,
        depositAmount: depositAmount.trim() === "" ? undefined : Number(depositAmount),
        materialItemId,
        materialQty,
        materialUnitCost
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
    } finally {
      setIsSavingOrder(false);
    }
  }

  async function handleUpdateDraft() {
    if (!selectedOrder) {
      return;
    }
    try {
      setIsSavingOrder(true);
      const order = await updateSubcontractOrder(selectedOrder.id, {
        factoryId,
        productId,
        quantity: Number(quantity),
        specVersion,
        sampleRequired,
        expectedDeliveryDate,
        depositStatus,
        depositAmount: depositAmount.trim() === "" ? undefined : Number(depositAmount),
        materialItemId,
        materialQty,
        materialUnitCost,
        expectedVersion: selectedOrder.version
      });

      setOrders((current) => [order, ...current.filter((candidate) => candidate.id !== order.id)]);
      setSelectedOrderId(order.id);
      setFeedback({
        tone: "success",
        message: `${order.orderNo} updated as draft`
      });
    } catch (error) {
      setFeedback({
        tone: "danger",
        message: error instanceof Error ? error.message : "Subcontract order could not be updated"
      });
    } finally {
      setIsSavingOrder(false);
    }
  }

  async function handleStatusAction(action: (typeof subcontractOrderActionOptions)[number]["value"]) {
    if (!selectedOrder) {
      return;
    }

    try {
      const result =
        action === "submit"
          ? await submitSubcontractOrder(selectedOrder.id, selectedOrder.version)
          : action === "approve"
            ? await approveSubcontractOrder(selectedOrder.id, selectedOrder.version)
            : action === "confirm-factory"
              ? await confirmFactorySubcontractOrder(selectedOrder.id, selectedOrder.version)
              : action === "cancel"
                ? await cancelSubcontractOrder(selectedOrder.id, "Cancelled from subcontract UI", selectedOrder.version)
                : await closeSubcontractOrder(selectedOrder.id, selectedOrder.version);

      setOrders((current) => [result.order, ...current.filter((order) => order.id !== result.order.id)]);
      setSelectedOrderId(result.order.id);
      setLastAudit(result.auditLog);
      setFeedback({
        tone: subcontractOrderStatusTone(result.order.status),
        message: `${result.order.orderNo} moved to ${formatSubcontractOrderStatus(result.order.status)}`
      });
    } catch (error) {
      setFeedback({
        tone: "danger",
        message: error instanceof Error ? error.message : "Subcontract order action failed"
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

  function updateMaterialIssueDraft(lineId: string, field: "issueQty" | "batchNo" | "sourceBinId", value: string) {
    setMaterialIssueDrafts((current) => ({
      ...current,
      [lineId]: {
        issueQty: current[lineId]?.issueQty ?? "",
        batchNo: current[lineId]?.batchNo ?? "",
        sourceBinId: current[lineId]?.sourceBinId ?? "",
        [field]: value
      }
    }));
  }

  async function handleCreateTransfer() {
    if (!selectedOrder) {
      return;
    }

    const warehouse =
      subcontractTransferWarehouseOptions.find((option) => option.value === sourceWarehouseId) ??
      subcontractTransferWarehouseOptions[0];

    try {
      setIsCreatingTransfer(true);
      const result = await issueSubcontractMaterials({
        order: selectedOrder,
        sourceWarehouseId: warehouse.value,
        sourceWarehouseCode: warehouse.code,
        handoverBy: "warehouse-user",
        handoverAt: new Date().toISOString(),
        receivedBy: handoverReceiver,
        receiverContact: handoverContact,
        lines: selectedOrder.materialLines.map((line) => {
          const draft = materialIssueDrafts[line.id];

          return {
            orderMaterialLineId: line.id,
            issueQty: draft?.issueQty || remainingMaterialQty(line.plannedQty, line.issuedQty),
            uomCode: line.uomCode,
            batchNo: draft?.batchNo,
            sourceBinId: draft?.sourceBinId
          };
        }),
        evidence: handoverEvidenceName.trim()
          ? [
              {
                id: `${selectedOrder.id}-handover`,
                evidenceType: "handover",
                fileName: handoverEvidenceName.trim(),
                objectKey: `subcontract/${selectedOrder.id}/${handoverEvidenceName.trim()}`
              }
            ]
          : undefined
      });

      setTransfers((current) => [result.transfer, ...current]);
      setOrders((current) => [result.order, ...current.filter((order) => order.id !== result.order.id)]);
      setSelectedOrderId(result.order.id);
      setLastAudit(result.auditLog);
      setTransferFeedback({
        tone: subcontractTransferStatusTone(result.transfer.status),
        message: `${result.transfer.transferNo} created with ${result.stockMovements.length} SUBCONTRACT_ISSUE movements`
      });
    } catch (error) {
      setTransferFeedback({
        tone: "danger",
        message: error instanceof Error ? error.message : "Material transfer could not be created"
      });
    } finally {
      setIsCreatingTransfer(false);
    }
  }

  async function handleSubmitSample() {
    if (!selectedOrder) {
      return;
    }

    try {
      setIsSavingSample(true);
      const result = await submitSubcontractSample({
        order: selectedOrder,
        sampleApprovalId: `${selectedOrder.id}-sample-${Date.now()}`,
        sampleCode,
        formulaVersion: sampleFormulaVersion,
        specVersion: selectedOrder.specVersion,
        submittedBy: "factory-user",
        submittedAt: new Date().toISOString(),
        evidence: [
          {
            evidenceType: "photo",
            fileName: sampleEvidenceName,
            objectKey: `subcontract/${selectedOrder.id}/${sampleEvidenceName || "sample-front.jpg"}`
          }
        ]
      });

      setOrders((current) => [result.order, ...current.filter((order) => order.id !== result.order.id)]);
      setSelectedOrderId(result.order.id);
      setSampleApprovals((current) => [
        result.sampleApproval,
        ...current.filter((sampleApproval) => sampleApproval.id !== result.sampleApproval.id)
      ]);
      setLastAudit(result.auditLog);
      setSampleFeedback({
        tone: "info",
        message: `${result.sampleApproval.sampleCode} submitted`
      });
    } catch (error) {
      setSampleFeedback({
        tone: "danger",
        message: error instanceof Error ? error.message : "Sample could not be submitted"
      });
    } finally {
      setIsSavingSample(false);
    }
  }

  async function handleDecideSample(decision: "approve" | "reject") {
    if (!selectedOrder || !latestSampleApproval) {
      return;
    }

    try {
      setIsSavingSample(true);
      const input = {
        order: selectedOrder,
        sampleApprovalId: latestSampleApproval.id,
        reason: sampleDecisionReason,
        decisionAt: new Date().toISOString()
      };
      const result =
        decision === "approve"
          ? await approveSubcontractSample({ ...input, storageStatus: sampleStorageStatus })
          : await rejectSubcontractSample(input);

      setOrders((current) => [result.order, ...current.filter((order) => order.id !== result.order.id)]);
      setSelectedOrderId(result.order.id);
      setSampleApprovals((current) => [
        result.sampleApproval,
        ...current.filter((sampleApproval) => sampleApproval.id !== result.sampleApproval.id)
      ]);
      setLastAudit(result.auditLog);
      setSampleFeedback({
        tone: decision === "approve" ? "success" : "danger",
        message: `${result.sampleApproval.sampleCode} ${result.sampleApproval.status}`
      });
    } catch (error) {
      setSampleFeedback({
        tone: "danger",
        message: error instanceof Error ? error.message : "Sample decision could not be saved"
      });
    } finally {
      setIsSavingSample(false);
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
              <span>Finished item</span>
              <select className="erp-input" value={productId} onChange={(event) => setProductId(event.target.value)}>
                {subcontractProductOptions.map((product) => (
                  <option key={product.id} value={product.id}>
                    {product.name}
                  </option>
                ))}
              </select>
            </label>
            <label className="erp-field">
              <span>Planned quantity</span>
              <input
                className="erp-input"
                min="1"
                type="number"
                value={quantity}
                onChange={(event) => setQuantity(event.target.value)}
              />
            </label>
            <label className="erp-field">
              <span>Spec summary</span>
              <input
                className="erp-input"
                type="text"
                value={specVersion}
                onChange={(event) => setSpecVersion(event.target.value)}
              />
            </label>
            <label className="erp-field">
              <span>Expected receipt date</span>
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
            <label className="erp-field">
              <span>Material line</span>
              <select
                className="erp-input"
                value={materialItemId}
                onChange={(event) => {
                  const nextItem = subcontractMaterialItemOptions.find((item) => item.id === event.target.value);
                  setMaterialItemId(event.target.value);
                  setMaterialUnitCost(nextItem?.defaultUnitCost ?? materialUnitCost);
                }}
              >
                {subcontractMaterialItemOptions.map((item) => (
                  <option key={item.id} value={item.id}>
                    {item.sku} / {item.name}
                  </option>
                ))}
              </select>
            </label>
            <label className="erp-field">
              <span>Material qty</span>
              <input
                className="erp-input"
                inputMode="decimal"
                type="text"
                value={materialQty}
                onChange={(event) => setMaterialQty(event.target.value)}
              />
            </label>
            <label className="erp-field">
              <span>Unit cost</span>
              <input
                className="erp-input"
                inputMode="decimal"
                type="text"
                value={materialUnitCost}
                onChange={(event) => setMaterialUnitCost(event.target.value)}
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
            <button className="erp-button erp-button--primary" type="button" disabled={isSavingOrder} onClick={handleCreateOrder}>
              Create order
            </button>
            <button
              className="erp-button erp-button--secondary"
              type="button"
              disabled={isSavingOrder || selectedOrder?.status !== "draft"}
              onClick={handleUpdateDraft}
            >
              Update draft
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
                  <span
                    aria-current={selectedOrder.status === option.value ? "step" : undefined}
                    className="erp-subcontract-status-step"
                    data-active={selectedOrder.status === option.value ? "true" : "false"}
                    key={option.value}
                  >
                    <span>{index + 1}</span>
                    <strong>{option.label}</strong>
                  </span>
                ))}
              </div>
              <div className="erp-subcontract-actions">
                {availableSubcontractOrderActions(selectedOrder.status).map((action) => {
                  const option = subcontractOrderActionOptions.find((candidate) => candidate.value === action);

                  return (
                    <button
                      className={`erp-button erp-button--${action === "cancel" ? "danger" : "primary"}`}
                      key={action}
                      type="button"
                      onClick={() => handleStatusAction(action)}
                    >
                      {option?.label ?? action}
                    </button>
                  );
                })}
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
            <label className="erp-field">
              <span>Factory receiver</span>
              <input
                className="erp-input"
                type="text"
                value={handoverReceiver}
                onChange={(event) => setHandoverReceiver(event.target.value)}
              />
            </label>
            <label className="erp-field">
              <span>Receiver contact</span>
              <input
                className="erp-input"
                type="text"
                value={handoverContact}
                onChange={(event) => setHandoverContact(event.target.value)}
              />
            </label>
            <label className="erp-field">
              <span>Handover file</span>
              <input
                className="erp-input"
                type="text"
                value={handoverEvidenceName}
                onChange={(event) => setHandoverEvidenceName(event.target.value)}
              />
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
            {(selectedOrder?.materialLines ?? []).map((line) => {
              const draft = materialIssueDrafts[line.id] ?? {
                issueQty: remainingMaterialQty(line.plannedQty, line.issuedQty),
                batchNo: "",
                sourceBinId: ""
              };

              return (
                <div className="erp-subcontract-line-item erp-subcontract-line-item--editable" key={line.id}>
                  <strong>{line.itemName}</strong>
                  <span>
                    {line.skuCode} / remaining {remainingMaterialQty(line.plannedQty, line.issuedQty)} {line.uomCode}
                  </span>
                  <label className="erp-field">
                    <span>Issue qty</span>
                    <input
                      className="erp-input"
                      inputMode="decimal"
                      type="text"
                      value={draft.issueQty}
                      onChange={(event) => updateMaterialIssueDraft(line.id, "issueQty", event.target.value)}
                    />
                  </label>
                  <label className="erp-field">
                    <span>Batch / lot</span>
                    <input
                      className="erp-input"
                      type="text"
                      value={draft.batchNo}
                      onChange={(event) => updateMaterialIssueDraft(line.id, "batchNo", event.target.value)}
                    />
                  </label>
                  <label className="erp-field">
                    <span>Bin</span>
                    <input
                      className="erp-input"
                      type="text"
                      value={draft.sourceBinId}
                      onChange={(event) => updateMaterialIssueDraft(line.id, "sourceBinId", event.target.value)}
                    />
                  </label>
                  <StatusChip tone={line.lotTraceRequired ? "warning" : "normal"}>
                    {line.lotTraceRequired ? "Lot required" : "No lot"}
                  </StatusChip>
                </div>
              );
            })}
          </div>

          <div className="erp-subcontract-actions">
            <button
              className="erp-button erp-button--primary"
              type="button"
              disabled={!selectedOrder || !canIssueSelectedOrder || !signedHandover || handoverReceiver.trim() === "" || isCreatingTransfer}
              onClick={handleCreateTransfer}
            >
              Issue materials
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
            <StatusChip tone={latestSampleApproval ? sampleApprovalTone(latestSampleApproval.status) : "normal"}>
              {latestSampleApproval ? formatSampleApprovalStatus(latestSampleApproval.status) : "No sample"}
            </StatusChip>
          </div>

          <div className="erp-subcontract-sample-controls">
            <label className="erp-field">
              <span>Sample code</span>
              <input
                className="erp-input"
                type="text"
                value={sampleCode}
                onChange={(event) => setSampleCode(event.target.value)}
              />
            </label>
            <label className="erp-field">
              <span>Formula</span>
              <input
                className="erp-input"
                type="text"
                value={sampleFormulaVersion}
                onChange={(event) => setSampleFormulaVersion(event.target.value)}
              />
            </label>
            <label className="erp-field">
              <span>Evidence file</span>
              <input
                className="erp-input"
                type="text"
                value={sampleEvidenceName}
                onChange={(event) => setSampleEvidenceName(event.target.value)}
              />
            </label>
            <label className="erp-field">
              <span>Decision note</span>
              <input
                className="erp-input"
                type="text"
                value={sampleDecisionReason}
                onChange={(event) => setSampleDecisionReason(event.target.value)}
              />
            </label>
            <label className="erp-field">
              <span>Storage status</span>
              <input
                className="erp-input"
                type="text"
                value={sampleStorageStatus}
                onChange={(event) => setSampleStorageStatus(event.target.value)}
              />
            </label>
          </div>

          {latestSampleApproval ? (
            <div className="erp-subcontract-sample-list">
              <div className="erp-subcontract-sample-item">
                <strong>{latestSampleApproval.sampleCode}</strong>
                <span>{latestSampleApproval.submittedAt.slice(0, 10)}</span>
                <span>{latestSampleApproval.submittedBy}</span>
                <StatusChip tone={sampleApprovalTone(latestSampleApproval.status)}>
                  {formatSampleApprovalStatus(latestSampleApproval.status)}
                </StatusChip>
                <span>{latestSampleApproval.evidence[0]?.fileName ?? latestSampleApproval.evidence[0]?.objectKey ?? "-"}</span>
                <span>
                  {latestSampleApproval.decisionReason || latestSampleApproval.storageStatus || latestSampleApproval.note || "-"}
                </span>
              </div>
            </div>
          ) : (
            <div className="erp-subcontract-empty-state">No sample decision recorded</div>
          )}

          <div className="erp-subcontract-actions">
            <button
              className="erp-button erp-button--secondary"
              type="button"
              disabled={!selectedOrder || !canSubmitSample || isSavingSample || sampleCode.trim() === ""}
              onClick={handleSubmitSample}
            >
              Submit sample
            </button>
            <button
              className="erp-button erp-button--primary"
              type="button"
              disabled={!canDecideSample || isSavingSample || sampleStorageStatus.trim() === ""}
              onClick={() => handleDecideSample("approve")}
            >
              Approve sample
            </button>
            <button
              className="erp-button erp-button--danger"
              type="button"
              disabled={!canDecideSample || isSavingSample || sampleDecisionReason.trim() === ""}
              onClick={() => handleDecideSample("reject")}
            >
              Reject sample
            </button>
            {sampleFeedback ? <StatusChip tone={sampleFeedback.tone}>{sampleFeedback.message}</StatusChip> : null}
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
        <DataTable columns={orderColumns} rows={displayedOrders} getRowKey={(row) => row.id} loading={isLoadingOrders} />
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

function remainingMaterialQty(plannedQty: string, issuedQty: string) {
  const remaining =
    toScaledBigInt(plannedQty, decimalScales.quantity) - toScaledBigInt(issuedQty, decimalScales.quantity);
  if (remaining < BigInt(0)) {
    return "0.000000";
  }

  return fromScaledBigInt(remaining, decimalScales.quantity);
}

function toScaledBigInt(value: string, scale: number) {
  const normalized = normalizeDecimalInput(value, scale);
  return BigInt(normalized.replace(".", ""));
}

function fromScaledBigInt(value: bigint, scale: number) {
  const digits = value.toString().padStart(scale + 1, "0");
  const integer = digits.slice(0, -scale);
  const fraction = scale > 0 ? `.${digits.slice(-scale)}` : "";

  return `${integer}${fraction}`;
}

function formatMoney(value: number) {
  return `${new Intl.NumberFormat("en-US", { maximumFractionDigits: 0 }).format(value)} VND`;
}

function subcontractQcLabel(order: SubcontractOrder) {
  if (order.status === "accepted" || order.status === "closed") {
    return "QC pass";
  }
  if (order.status === "rejected_with_factory_issue") {
    return "Claim hold";
  }
  if (order.status === "qc_in_progress") {
    return "QC review";
  }
  if (order.sampleRequired) {
    return "Sample pending";
  }

  return "Not required";
}

function subcontractQcTone(order: SubcontractOrder): StatusTone {
  if (order.status === "accepted" || order.status === "closed") {
    return "success";
  }
  if (order.status === "rejected_with_factory_issue") {
    return "danger";
  }
  if (order.status === "qc_in_progress" || order.sampleRequired) {
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

function formatSampleApprovalStatus(status: SubcontractSampleApproval["status"]) {
  switch (status) {
    case "approved":
      return "Approved";
    case "rejected":
      return "Rejected";
    case "submitted":
    default:
      return "Submitted";
  }
}

function sampleApprovalTone(status: SubcontractSampleApproval["status"]): StatusTone {
  switch (status) {
    case "approved":
      return "success";
    case "rejected":
      return "danger";
    case "submitted":
    default:
      return "warning";
  }
}

function timelineIndexForStatus(status: SubcontractOrderStatus) {
  switch (status) {
    case "draft":
      return 0;
    case "submitted":
    case "approved":
    case "factory_confirmed":
    case "deposit_recorded":
      return 1;
    case "materials_issued_to_factory":
      return 2;
    case "sample_submitted":
    case "sample_approved":
    case "sample_rejected":
      return 3;
    case "mass_production_started":
      return 4;
    case "finished_goods_received":
      return 5;
    case "qc_in_progress":
    case "accepted":
    case "rejected_with_factory_issue":
      return 6;
    case "final_payment_ready":
    case "closed":
    case "cancelled":
    default:
      return 7;
  }
}
