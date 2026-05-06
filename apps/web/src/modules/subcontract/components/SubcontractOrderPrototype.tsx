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
  addSubcontractFactoryClaimEvidence,
  approveSubcontractOrder,
  approveSubcontractSample,
  availableSubcontractOrderActions,
  cancelSubcontractOrder,
  changeSubcontractOrderStatus,
  closeSubcontractOrder,
  confirmFactorySubcontractOrder,
  createSubcontractFactoryClaim,
  createSubcontractOrder,
  formatSubcontractFactoryClaimStatus,
  formatSubcontractDepositStatus,
  formatSubcontractOrderStatus,
  getSubcontractOrder,
  getSubcontractOrders,
  issueSubcontractMaterials,
  markSubcontractFinalPaymentReady,
  prototypeSubcontractOrders,
  receiveSubcontractFinishedGoods,
  recordSubcontractDeposit,
  rejectSubcontractSample,
  startMassProductionSubcontractOrder,
  subcontractFactoryClaimSlaLabel,
  subcontractFactoryClaimStatusTone,
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
  SubcontractFactoryClaim,
  SubcontractFactoryClaimSeverity,
  SubcontractFinishedGoodsPackagingStatus,
  SubcontractFinishedGoodsReceipt,
  SubcontractFinalPaymentStatus,
  SubcontractMaterialTransfer,
  SubcontractOrder,
  SubcontractOrderStatus,
  SubcontractPaymentMilestone,
  SubcontractSampleApproval,
  SubcontractStockMovement
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
  const [initialSearchFilter, setInitialSearchFilter] = useState("");
  const [sourceProductionPlanFilter, setSourceProductionPlanFilter] = useState("");
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
  const [receiptWarehouseId, setReceiptWarehouseId] = useState<string>(subcontractTransferWarehouseOptions[0].value);
  const [receiptLocationId, setReceiptLocationId] = useState("loc-hcm-fg-qc");
  const [receiptLocationCode, setReceiptLocationCode] = useState("FG-QC-01");
  const [deliveryNoteNo, setDeliveryNoteNo] = useState("DN-FACTORY-001");
  const [receivedBy, setReceivedBy] = useState("warehouse-user");
  const [finishedGoodsQty, setFinishedGoodsQty] = useState("100.000000");
  const [finishedGoodsBatchNo, setFinishedGoodsBatchNo] = useState("FG-LOT-001");
  const [finishedGoodsLotNo, setFinishedGoodsLotNo] = useState("FG-LOT-001");
  const [finishedGoodsExpiryDate, setFinishedGoodsExpiryDate] = useState("2028-04-29");
  const [packagingStatus, setPackagingStatus] = useState<SubcontractFinishedGoodsPackagingStatus>("intact");
  const [receiptEvidenceName, setReceiptEvidenceName] = useState("factory-delivery.pdf");
  const [receiptNote, setReceiptNote] = useState("");
  const [finishedGoodsReceipts, setFinishedGoodsReceipts] = useState<SubcontractFinishedGoodsReceipt[]>([]);
  const [receiptMovements, setReceiptMovements] = useState<SubcontractStockMovement[]>([]);
  const [receiptFeedback, setReceiptFeedback] = useState<{ tone: StatusTone; message: string } | null>(null);
  const [isReceivingFinishedGoods, setIsReceivingFinishedGoods] = useState(false);
  const [factoryClaims, setFactoryClaims] = useState<SubcontractFactoryClaim[]>([]);
  const [claimReasonCode, setClaimReasonCode] = useState("PACKAGING_DAMAGED");
  const [claimReason, setClaimReason] = useState("Outer cartons damaged during factory delivery");
  const [claimSeverity, setClaimSeverity] = useState<SubcontractFactoryClaimSeverity>("P1");
  const [claimAffectedQty, setClaimAffectedQty] = useState("0.000000");
  const [claimOwner, setClaimOwner] = useState("factory-owner");
  const [claimEvidenceName, setClaimEvidenceName] = useState("factory-claim-photo.jpg");
  const [claimEvidenceNote, setClaimEvidenceNote] = useState("Inbound QC fail evidence");
  const [claimFeedback, setClaimFeedback] = useState<{ tone: StatusTone; message: string } | null>(null);
  const [isSavingClaim, setIsSavingClaim] = useState(false);
  const [isAddingClaimEvidence, setIsAddingClaimEvidence] = useState(false);
  const [paymentMilestones, setPaymentMilestones] = useState<SubcontractPaymentMilestone[]>([]);
  const [paymentDepositAmount, setPaymentDepositAmount] = useState("5000000");
  const [finalPaymentAmount, setFinalPaymentAmount] = useState("");
  const [paymentApprovedExceptionId, setPaymentApprovedExceptionId] = useState("");
  const [paymentFeedback, setPaymentFeedback] = useState<{ tone: StatusTone; message: string } | null>(null);
  const [isRecordingDeposit, setIsRecordingDeposit] = useState(false);
  const [isMarkingFinalPayment, setIsMarkingFinalPayment] = useState(false);
  const summary = useMemo(() => summarizeSubcontractOrders(orders), [orders]);
  const transferSummary = useMemo(() => summarizeSubcontractMaterialTransfers(transfers), [transfers]);
  const orderColumns = useMemo(() => subcontractOrderColumns(handleOpenOrder), []);
  const displayedOrders = useMemo(
    () =>
      orders.filter((order) => {
        const searchValue = search.trim().toLowerCase();
        if (
          searchValue &&
          ![order.orderNo, order.factoryName, order.productName, order.sku, order.sourceProductionPlanNo ?? ""].some((value) =>
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
  const latestFinishedGoodsReceipt =
    finishedGoodsReceipts.find((receipt) => receipt.orderId === selectedOrder?.id) ?? null;
  const latestReceiptMovements = latestFinishedGoodsReceipt
    ? receiptMovements.filter((movement) => movement.sourceDocId === latestFinishedGoodsReceipt.id)
    : [];
  const selectedFactoryClaims = selectedOrder
    ? factoryClaims.filter((claim) => claim.orderId === selectedOrder.id)
    : [];
  const latestFactoryClaim = selectedFactoryClaims[0] ?? null;
  const selectedPaymentMilestones = selectedOrder
    ? paymentMilestones.filter((milestone) => milestone.orderId === selectedOrder.id)
    : [];
  const latestPaymentMilestone = selectedPaymentMilestones[0] ?? null;
  const hasBlockingPaymentClaim = selectedFactoryClaims.some(
    (claim) => claim.blocksFinalPayment && ["open", "acknowledged"].includes(claim.status)
  );
  const canCreateFactoryClaim =
    !!selectedOrder &&
    !!latestFinishedGoodsReceipt &&
    ["finished_goods_received", "qc_in_progress"].includes(selectedOrder.status);
  const canRecordDeposit = selectedOrder?.status === "factory_confirmed";
  const canMarkFinalPaymentReady = selectedOrder?.status === "accepted";
  const canIssueSelectedOrder =
    selectedOrder?.status === "factory_confirmed" || selectedOrder?.status === "deposit_recorded";
  const canSubmitSample =
    selectedOrder?.sampleRequired &&
    (selectedOrder.status === "materials_issued_to_factory" || selectedOrder.status === "sample_rejected");
  const canDecideSample = selectedOrder?.status === "sample_submitted" && latestSampleApproval?.status === "submitted";
  const canReceiveFinishedGoods =
    selectedOrder?.status === "mass_production_started" || selectedOrder?.status === "finished_goods_received";

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
    setPaymentDepositAmount(selectedOrder.depositAmount ? String(selectedOrder.depositAmount) : "5000000");
    setFinalPaymentAmount(selectedOrder.estimatedCostAmount || "");
    setPaymentApprovedExceptionId("");
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
    if (!selectedOrder) {
      return;
    }

    setDeliveryNoteNo(`${selectedOrder.orderNo}-DN-001`);
    setReceivedBy("warehouse-user");
    setFinishedGoodsQty(remainingFinishedGoodsQty(selectedOrder));
    setFinishedGoodsBatchNo(`${selectedOrder.sku}-LOT-001`);
    setFinishedGoodsLotNo(`${selectedOrder.sku}-LOT-001`);
    setFinishedGoodsExpiryDate("2028-04-29");
    setPackagingStatus("intact");
    setReceiptEvidenceName("factory-delivery.pdf");
    setReceiptNote("");
    setClaimReasonCode("PACKAGING_DAMAGED");
    setClaimReason("Outer cartons damaged during factory delivery");
    setClaimSeverity("P1");
    setClaimAffectedQty(remainingFinishedGoodsQty(selectedOrder));
    setClaimOwner("factory-owner");
    setClaimEvidenceName("factory-claim-photo.jpg");
    setClaimEvidenceNote("Inbound QC fail evidence");
  }, [selectedOrder?.id, selectedOrder?.version]);

  useEffect(() => {
    const latestLine = latestFinishedGoodsReceipt?.lines[0];
    if (latestLine) {
      setClaimAffectedQty(latestLine.receiveQty);
    }
  }, [latestFinishedGoodsReceipt?.id]);

  useEffect(() => {
    if (typeof window === "undefined") {
      return;
    }
    const params = new URLSearchParams(window.location.search);
    const sourceProductionPlanId = params.get("source_production_plan_id")?.trim() ?? "";
    const searchValue = params.get("search")?.trim() ?? "";
    setSourceProductionPlanFilter(sourceProductionPlanId);
    setInitialSearchFilter(searchValue);
    setSearch(searchValue);
  }, []);

  useEffect(() => {
    let active = true;
    setIsLoadingOrders(true);

    async function loadOrders() {
      try {
        let loadedOrders = await getSubcontractOrders({
          sourceProductionPlanId: sourceProductionPlanFilter || undefined,
          search: initialSearchFilter || undefined
        });
        if (loadedOrders.length === 0 && sourceProductionPlanFilter && initialSearchFilter) {
          loadedOrders = await getSubcontractOrders({ search: initialSearchFilter });
        }
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
  }, [initialSearchFilter, sourceProductionPlanFilter]);

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
              : action === "start-mass-production"
                ? await startMassProductionSubcontractOrder(selectedOrder.id, selectedOrder.version)
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

  async function handleReceiveFinishedGoods() {
    if (!selectedOrder) {
      return;
    }

    const warehouse =
      subcontractTransferWarehouseOptions.find((option) => option.value === receiptWarehouseId) ??
      subcontractTransferWarehouseOptions[0];

    try {
      setIsReceivingFinishedGoods(true);
      const result = await receiveSubcontractFinishedGoods({
        order: selectedOrder,
        receiptId: `${selectedOrder.id}-fg-receipt-${Date.now()}`,
        warehouseId: warehouse.value,
        warehouseCode: warehouse.code,
        locationId: receiptLocationId,
        locationCode: receiptLocationCode,
        deliveryNoteNo,
        receivedBy,
        receivedAt: new Date().toISOString(),
        note: receiptNote,
        lines: [
          {
            itemId: selectedOrder.productId,
            skuCode: selectedOrder.sku,
            itemName: selectedOrder.productName,
            receiveQty: finishedGoodsQty,
            uomCode: "EA",
            baseReceiveQty: finishedGoodsQty,
            baseUOMCode: "EA",
            conversionFactor: "1",
            batchNo: finishedGoodsBatchNo,
            lotNo: finishedGoodsLotNo,
            expiryDate: finishedGoodsExpiryDate,
            packagingStatus
          }
        ],
        evidence: receiptEvidenceName.trim()
          ? [
              {
                id: `${selectedOrder.id}-fg-delivery-note`,
                evidenceType: "delivery_note",
                fileName: receiptEvidenceName.trim(),
                objectKey: `subcontract/${selectedOrder.id}/${receiptEvidenceName.trim()}`
              }
            ]
          : undefined
      });

      setOrders((current) => [result.order, ...current.filter((order) => order.id !== result.order.id)]);
      setSelectedOrderId(result.order.id);
      setFinishedGoodsReceipts((current) => [
        result.receipt,
        ...current.filter((receipt) => receipt.id !== result.receipt.id)
      ]);
      setReceiptMovements((current) => [
        ...result.stockMovements,
        ...current.filter((movement) => movement.sourceDocId !== result.receipt.id)
      ]);
      setLastAudit(result.auditLog);
      setReceiptFeedback({
        tone: "success",
        message: `${result.receipt.receiptNo} received into QC hold`
      });
    } catch (error) {
      setReceiptFeedback({
        tone: "danger",
        message: error instanceof Error ? error.message : "Finished goods receipt could not be created"
      });
    } finally {
      setIsReceivingFinishedGoods(false);
    }
  }

  async function handleCreateFactoryClaim() {
    if (!selectedOrder || !latestFinishedGoodsReceipt) {
      return;
    }

    try {
      setIsSavingClaim(true);
      const result = await createSubcontractFactoryClaim({
        order: selectedOrder,
        claimId: `${selectedOrder.id}-factory-claim-${Date.now()}`,
        receiptId: latestFinishedGoodsReceipt.id,
        receiptNo: latestFinishedGoodsReceipt.receiptNo,
        reasonCode: claimReasonCode,
        reason: claimReason,
        severity: claimSeverity,
        affectedQty: claimAffectedQty,
        uomCode: latestFinishedGoodsReceipt.lines[0]?.uomCode ?? "EA",
        baseAffectedQty: claimAffectedQty,
        baseUOMCode: latestFinishedGoodsReceipt.lines[0]?.baseUOMCode ?? "EA",
        ownerId: claimOwner,
        openedBy: "qa-user",
        openedAt: new Date().toISOString(),
        evidence: [
          {
            evidenceType: "qc_photo",
            fileName: claimEvidenceName,
            objectKey: `subcontract/${selectedOrder.id}/${claimEvidenceName || "factory-claim-photo.jpg"}`,
            note: claimEvidenceNote
          }
        ]
      });

      setOrders((current) => [result.order, ...current.filter((order) => order.id !== result.order.id)]);
      setSelectedOrderId(result.order.id);
      setFactoryClaims((current) => [result.claim, ...current.filter((claim) => claim.id !== result.claim.id)]);
      setLastAudit(result.auditLog);
      setClaimFeedback({
        tone: "danger",
        message: `${result.claim.claimNo} opened; final payment on hold`
      });
    } catch (error) {
      setClaimFeedback({
        tone: "danger",
        message: error instanceof Error ? error.message : "Factory claim could not be created"
      });
    } finally {
      setIsSavingClaim(false);
    }
  }

  async function handleAddClaimEvidence() {
    if (!latestFactoryClaim) {
      return;
    }

    try {
      setIsAddingClaimEvidence(true);
      const updated = await addSubcontractFactoryClaimEvidence(latestFactoryClaim.id, {
        evidenceType: "inspection_note",
        fileName: claimEvidenceName,
        objectKey: `subcontract/${latestFactoryClaim.orderId}/${claimEvidenceName || "factory-claim-note.pdf"}`,
        note: claimEvidenceNote
      });

      setFactoryClaims((current) => [updated, ...current.filter((claim) => claim.id !== updated.id)]);
      setClaimFeedback({
        tone: "info",
        message: `${updated.claimNo} evidence updated`
      });
    } catch (error) {
      setClaimFeedback({
        tone: "danger",
        message: error instanceof Error ? error.message : "Factory claim evidence could not be updated"
      });
    } finally {
      setIsAddingClaimEvidence(false);
    }
  }

  async function handleRecordDeposit() {
    if (!selectedOrder) {
      return;
    }

    try {
      setIsRecordingDeposit(true);
      const result = await recordSubcontractDeposit({
        order: selectedOrder,
        amount: paymentDepositAmount,
        recordedBy: "finance-user",
        recordedAt: new Date().toISOString(),
        note: "Deposit transfer confirmed"
      });

      setOrders((current) => [result.order, ...current.filter((order) => order.id !== result.order.id)]);
      setSelectedOrderId(result.order.id);
      setPaymentMilestones((current) => [
        result.milestone,
        ...current.filter((milestone) => milestone.id !== result.milestone.id)
      ]);
      setLastAudit(result.auditLog);
      setPaymentFeedback({
        tone: "success",
        message: `${result.milestone.milestoneNo} recorded`
      });
    } catch (error) {
      setPaymentFeedback({
        tone: "danger",
        message: error instanceof Error ? error.message : "Deposit could not be recorded"
      });
    } finally {
      setIsRecordingDeposit(false);
    }
  }

  async function handleMarkFinalPaymentReady() {
    if (!selectedOrder) {
      return;
    }

    try {
      setIsMarkingFinalPayment(true);
      const result = await markSubcontractFinalPaymentReady({
        order: selectedOrder,
        amount: finalPaymentAmount,
        readyBy: "finance-user",
        readyAt: new Date().toISOString(),
        approvedExceptionId: paymentApprovedExceptionId,
        note: "Accepted goods cleared for final payment"
      });

      setOrders((current) => [result.order, ...current.filter((order) => order.id !== result.order.id)]);
      setSelectedOrderId(result.order.id);
      setPaymentMilestones((current) => [
        result.milestone,
        ...current.filter((milestone) => milestone.id !== result.milestone.id)
      ]);
      setLastAudit(result.auditLog);
      setPaymentFeedback({
        tone: "success",
        message: `${result.milestone.milestoneNo} ready`
      });
    } catch (error) {
      setPaymentFeedback({
        tone: "danger",
        message: error instanceof Error ? error.message : "Final payment could not be marked ready"
      });
    } finally {
      setIsMarkingFinalPayment(false);
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
          <a className="erp-button erp-button--secondary" href="#subcontract-inbound">
            Inbound
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
                {availableSubcontractOrderActions(selectedOrder.status, selectedOrder.sampleRequired).map((action) => {
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

      <section className="erp-subcontract-quality-grid" id="subcontract-inbound">
        <div className="erp-card erp-card--padded erp-subcontract-card">
          <div className="erp-section-header">
            <h2 className="erp-section-title">Finished goods receipt</h2>
            <StatusChip tone={canReceiveFinishedGoods ? "info" : "normal"}>
              {selectedOrder ? formatSubcontractOrderStatus(selectedOrder.status) : "No order"}
            </StatusChip>
          </div>

          <div className="erp-subcontract-receipt-controls">
            <label className="erp-field">
              <span>Warehouse</span>
              <select
                className="erp-input"
                value={receiptWarehouseId}
                onChange={(event) => setReceiptWarehouseId(event.target.value)}
              >
                {subcontractTransferWarehouseOptions.map((warehouse) => (
                  <option key={warehouse.value} value={warehouse.value}>
                    {warehouse.label}
                  </option>
                ))}
              </select>
            </label>
            <label className="erp-field">
              <span>QC location</span>
              <input
                className="erp-input"
                type="text"
                value={receiptLocationCode}
                onChange={(event) => {
                  setReceiptLocationCode(event.target.value);
                  setReceiptLocationId(event.target.value.trim() ? `loc-${event.target.value.toLowerCase()}` : "");
                }}
              />
            </label>
            <label className="erp-field">
              <span>Delivery note</span>
              <input
                className="erp-input"
                type="text"
                value={deliveryNoteNo}
                onChange={(event) => setDeliveryNoteNo(event.target.value)}
              />
            </label>
            <label className="erp-field">
              <span>Received by</span>
              <input
                className="erp-input"
                type="text"
                value={receivedBy}
                onChange={(event) => setReceivedBy(event.target.value)}
              />
            </label>
            <label className="erp-field">
              <span>Evidence file</span>
              <input
                className="erp-input"
                type="text"
                value={receiptEvidenceName}
                onChange={(event) => setReceiptEvidenceName(event.target.value)}
              />
            </label>
          </div>

          <div className="erp-subcontract-line-list" aria-label="Finished goods receipt line">
            <div className="erp-subcontract-line-item erp-subcontract-line-item--receipt">
              <strong>{selectedOrder?.productName ?? "Finished item"}</strong>
              <span>{selectedOrder ? `${selectedOrder.sku} / remaining ${remainingFinishedGoodsQty(selectedOrder)} EA` : "-"}</span>
              <label className="erp-field">
                <span>Receive qty</span>
                <input
                  className="erp-input"
                  inputMode="decimal"
                  type="text"
                  value={finishedGoodsQty}
                  onChange={(event) => setFinishedGoodsQty(event.target.value)}
                />
              </label>
              <label className="erp-field">
                <span>Batch / lot</span>
                <input
                  className="erp-input"
                  type="text"
                  value={finishedGoodsBatchNo}
                  onChange={(event) => {
                    setFinishedGoodsBatchNo(event.target.value);
                    setFinishedGoodsLotNo(event.target.value);
                  }}
                />
              </label>
              <label className="erp-field">
                <span>Expiry</span>
                <input
                  className="erp-input"
                  type="date"
                  value={finishedGoodsExpiryDate}
                  onChange={(event) => setFinishedGoodsExpiryDate(event.target.value)}
                />
              </label>
              <label className="erp-field">
                <span>Packaging</span>
                <select
                  className="erp-input"
                  value={packagingStatus}
                  onChange={(event) => setPackagingStatus(event.target.value as SubcontractFinishedGoodsPackagingStatus)}
                >
                  <option value="intact">Intact</option>
                  <option value="damaged">Damaged</option>
                  <option value="mixed">Mixed</option>
                </select>
              </label>
            </div>
          </div>

          <label className="erp-field erp-subcontract-note-field">
            <span>Receipt note</span>
            <input
              className="erp-input"
              type="text"
              value={receiptNote}
              onChange={(event) => setReceiptNote(event.target.value)}
            />
          </label>

          <div className="erp-subcontract-actions">
            <button
              className="erp-button erp-button--primary"
              type="button"
              disabled={
                !selectedOrder ||
                !canReceiveFinishedGoods ||
                isReceivingFinishedGoods ||
                deliveryNoteNo.trim() === "" ||
                receivedBy.trim() === "" ||
                finishedGoodsQty.trim() === "" ||
                finishedGoodsBatchNo.trim() === "" ||
                finishedGoodsExpiryDate.trim() === ""
              }
              onClick={handleReceiveFinishedGoods}
            >
              Receive to QC hold
            </button>
            {receiptFeedback ? <StatusChip tone={receiptFeedback.tone}>{receiptFeedback.message}</StatusChip> : null}
          </div>
        </div>

        <div className="erp-card erp-card--padded erp-subcontract-card">
          <div className="erp-section-header">
            <h2 className="erp-section-title">Receipt result</h2>
            <StatusChip tone={latestFinishedGoodsReceipt ? "warning" : "normal"}>
              {latestFinishedGoodsReceipt ? "QC hold" : "No receipt"}
            </StatusChip>
          </div>

          {latestFinishedGoodsReceipt ? (
            <>
              <div className="erp-subcontract-fact-grid">
                <SubcontractFact label="Receipt" value={latestFinishedGoodsReceipt.receiptNo} />
                <SubcontractFact label="Delivery note" value={latestFinishedGoodsReceipt.deliveryNoteNo} />
                <SubcontractFact label="Warehouse" value={latestFinishedGoodsReceipt.warehouseCode} />
                <SubcontractFact label="Location" value={latestFinishedGoodsReceipt.locationCode} />
                <SubcontractFact label="Received by" value={latestFinishedGoodsReceipt.receivedBy} />
                <SubcontractFact label="Received at" value={latestFinishedGoodsReceipt.receivedAt.slice(0, 10)} />
              </div>
              <div className="erp-subcontract-line-list" aria-label="Finished goods receipt lines">
                {latestFinishedGoodsReceipt.lines.map((line) => (
                  <div className="erp-subcontract-line-item" key={line.id}>
                    <strong>{line.itemName}</strong>
                    <span>{line.skuCode}</span>
                    <span>
                      {Number(line.receiveQty).toLocaleString("en-US")} {line.uomCode} / {line.batchNo}
                    </span>
                    <StatusChip tone="warning">{line.packagingStatus ?? "qc_hold"}</StatusChip>
                  </div>
                ))}
              </div>
              <div className="erp-subcontract-movement-list" aria-label="Finished goods receipt movements">
                {latestReceiptMovements.map((movement) => (
                  <div className="erp-subcontract-movement-item" key={movement.id}>
                    <strong>{movement.movementType}</strong>
                    <span>
                      {movement.itemCode} / {formatQuantity(movement.quantity)} {movement.unit} / {movement.targetLocation}
                    </span>
                  </div>
                ))}
              </div>
              <div className="erp-subcontract-attachment-list" aria-label="Finished goods receipt evidence">
                {latestFinishedGoodsReceipt.evidence.map((evidence) => (
                  <StatusChip key={evidence.id} tone="info">
                    {evidence.fileName ?? evidence.objectKey ?? evidence.evidenceType}
                  </StatusChip>
                ))}
              </div>
            </>
          ) : (
            <div className="erp-subcontract-empty-state">No factory delivery received into QC hold</div>
          )}
        </div>
      </section>

      <section className="erp-subcontract-quality-grid" id="subcontract-sample">
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

        <div className="erp-card erp-card--padded erp-subcontract-card" id="subcontract-claim">
          <div className="erp-section-header">
            <h2 className="erp-section-title">Factory claim</h2>
            <StatusChip tone={latestFactoryClaim ? subcontractFactoryClaimStatusTone(latestFactoryClaim.status) : "normal"}>
              {latestFactoryClaim ? formatSubcontractFactoryClaimStatus(latestFactoryClaim.status) : "No claim"}
            </StatusChip>
          </div>

          <div className="erp-subcontract-claim-controls">
            <label className="erp-field">
              <span>Reason code</span>
              <select
                className="erp-input"
                value={claimReasonCode}
                onChange={(event) => setClaimReasonCode(event.target.value)}
              >
                <option value="PACKAGING_DAMAGED">Packaging damaged</option>
                <option value="SPEC_MISMATCH">Spec mismatch</option>
                <option value="QTY_SHORT">Quantity short</option>
                <option value="QUALITY_FAIL">Quality fail</option>
              </select>
            </label>
            <label className="erp-field">
              <span>Severity</span>
              <select
                className="erp-input"
                value={claimSeverity}
                onChange={(event) => setClaimSeverity(event.target.value as SubcontractFactoryClaimSeverity)}
              >
                <option value="P1">P1</option>
                <option value="P2">P2</option>
                <option value="P3">P3</option>
              </select>
            </label>
            <label className="erp-field">
              <span>Affected qty</span>
              <input
                className="erp-input"
                inputMode="decimal"
                type="text"
                value={claimAffectedQty}
                onChange={(event) => setClaimAffectedQty(event.target.value)}
              />
            </label>
            <label className="erp-field">
              <span>Owner</span>
              <input
                className="erp-input"
                type="text"
                value={claimOwner}
                onChange={(event) => setClaimOwner(event.target.value)}
              />
            </label>
            <label className="erp-field">
              <span>Evidence file</span>
              <input
                className="erp-input"
                type="text"
                value={claimEvidenceName}
                onChange={(event) => setClaimEvidenceName(event.target.value)}
              />
            </label>
          </div>

          <label className="erp-field erp-subcontract-note-field">
            <span>Claim reason</span>
            <input
              className="erp-input"
              type="text"
              value={claimReason}
              onChange={(event) => setClaimReason(event.target.value)}
            />
          </label>
          <label className="erp-field erp-subcontract-note-field">
            <span>Evidence note</span>
            <input
              className="erp-input"
              type="text"
              value={claimEvidenceNote}
              onChange={(event) => setClaimEvidenceNote(event.target.value)}
            />
          </label>

          {selectedFactoryClaims.length > 0 ? (
            <>
              <div className="erp-subcontract-fact-grid erp-subcontract-claim-facts">
                <SubcontractFact label="Claim" value={latestFactoryClaim?.claimNo ?? "-"} />
                <SubcontractFact label="Receipt" value={latestFactoryClaim?.receiptNo ?? "-"} />
                <SubcontractFact label="Owner" value={latestFactoryClaim?.ownerId ?? "-"} />
                <SubcontractFact label="SLA" value={latestFactoryClaim ? subcontractFactoryClaimSlaLabel(latestFactoryClaim) : "-"} />
              </div>
              <div className="erp-subcontract-claim-list">
                {selectedFactoryClaims.map((claim) => (
                  <div className="erp-subcontract-claim-item" key={claim.id}>
                    <strong>{claim.reasonCode}</strong>
                    <span>
                      {Number(claim.affectedQty).toLocaleString("en-US")} {claim.uomCode}
                    </span>
                    <span>Opened {claim.openedAt.slice(0, 10)}</span>
                    <span>Due {claim.dueAt.slice(0, 10)}</span>
                    <StatusChip tone={claim.severity === "P1" ? "danger" : "warning"}>{claim.severity}</StatusChip>
                    <StatusChip tone={subcontractFactoryClaimStatusTone(claim.status)}>
                      {formatSubcontractFactoryClaimStatus(claim.status)}
                    </StatusChip>
                    <StatusChip tone={claim.blocksFinalPayment ? "danger" : "success"}>
                      {claim.blocksFinalPayment ? "Payment hold" : "Payment clear"}
                    </StatusChip>
                  </div>
                ))}
              </div>
              <div className="erp-subcontract-attachment-list" aria-label="Factory claim evidence">
                {latestFactoryClaim?.evidence.map((evidence) => (
                  <StatusChip key={evidence.id} tone="info">
                    {evidence.fileName ?? evidence.objectKey ?? evidence.evidenceType}
                  </StatusChip>
                ))}
              </div>
            </>
          ) : (
            <div className="erp-subcontract-empty-state">No factory claim opened for this order</div>
          )}

          <div className="erp-subcontract-actions">
            <button
              className="erp-button erp-button--danger"
              type="button"
              disabled={
                !canCreateFactoryClaim ||
                isSavingClaim ||
                claimReason.trim() === "" ||
                claimAffectedQty.trim() === "" ||
                claimOwner.trim() === "" ||
                claimEvidenceName.trim() === ""
              }
              onClick={handleCreateFactoryClaim}
            >
              Create claim
            </button>
            <button
              className="erp-button erp-button--secondary"
              type="button"
              disabled={!latestFactoryClaim || isAddingClaimEvidence || claimEvidenceName.trim() === ""}
              onClick={handleAddClaimEvidence}
            >
              Add evidence
            </button>
            {claimFeedback ? <StatusChip tone={claimFeedback.tone}>{claimFeedback.message}</StatusChip> : null}
          </div>
        </div>
      </section>

      <section className="erp-subcontract-quality-grid" id="subcontract-payment">
        <div className="erp-card erp-card--padded erp-subcontract-card">
          <div className="erp-section-header">
            <h2 className="erp-section-title">Payment milestone</h2>
            <StatusChip tone={latestPaymentMilestone ? paymentMilestoneTone(latestPaymentMilestone.status) : "normal"}>
              {latestPaymentMilestone ? formatPaymentMilestoneStatus(latestPaymentMilestone.status) : "No milestone"}
            </StatusChip>
          </div>

          <div className="erp-subcontract-fact-grid">
            <SubcontractFact
              label="Deposit"
              value={selectedOrder ? formatSubcontractDepositStatus(selectedOrder.depositStatus) : "-"}
            />
            <SubcontractFact
              label="Deposit amount"
              value={selectedOrder?.depositAmount ? formatMoney(selectedOrder.depositAmount) : "-"}
            />
            <SubcontractFact
              label="Final payment"
              value={selectedOrder ? formatFinalPaymentStatus(selectedOrder.finalPaymentStatus) : "-"}
            />
            <SubcontractFact label="Factory claim" value={hasBlockingPaymentClaim ? "Payment hold" : "Clear"} />
          </div>

          <div className="erp-subcontract-claim-controls">
            <label className="erp-field">
              <span>Deposit amount</span>
              <input
                className="erp-input"
                inputMode="decimal"
                type="text"
                value={paymentDepositAmount}
                onChange={(event) => setPaymentDepositAmount(event.target.value)}
              />
            </label>
            <label className="erp-field">
              <span>Final amount</span>
              <input
                className="erp-input"
                inputMode="decimal"
                type="text"
                value={finalPaymentAmount}
                onChange={(event) => setFinalPaymentAmount(event.target.value)}
              />
            </label>
            <label className="erp-field">
              <span>Approved exception</span>
              <input
                className="erp-input"
                type="text"
                value={paymentApprovedExceptionId}
                onChange={(event) => setPaymentApprovedExceptionId(event.target.value)}
              />
            </label>
          </div>

          <div className="erp-subcontract-actions">
            <button
              className="erp-button erp-button--primary"
              type="button"
              disabled={!selectedOrder || !canRecordDeposit || isRecordingDeposit || paymentDepositAmount.trim() === ""}
              onClick={handleRecordDeposit}
            >
              Record deposit
            </button>
            <button
              className="erp-button erp-button--secondary"
              type="button"
              disabled={
                !selectedOrder ||
                !canMarkFinalPaymentReady ||
                isMarkingFinalPayment ||
                finalPaymentAmount.trim() === "" ||
                (hasBlockingPaymentClaim && paymentApprovedExceptionId.trim() === "")
              }
              onClick={handleMarkFinalPaymentReady}
            >
              Mark final ready
            </button>
            {paymentFeedback ? <StatusChip tone={paymentFeedback.tone}>{paymentFeedback.message}</StatusChip> : null}
          </div>
        </div>

        <div className="erp-card erp-card--padded erp-subcontract-card">
          <div className="erp-section-header">
            <h2 className="erp-section-title">Payment result</h2>
            <StatusChip tone={selectedPaymentMilestones.length > 0 ? "info" : "normal"}>
              {selectedPaymentMilestones.length} milestones
            </StatusChip>
          </div>

          {selectedPaymentMilestones.length > 0 ? (
            <div className="erp-subcontract-claim-list">
              {selectedPaymentMilestones.map((milestone) => (
                <div className="erp-subcontract-claim-item" key={milestone.id}>
                  <strong>{milestone.milestoneNo}</strong>
                  <span>{milestone.kind === "deposit" ? "Deposit" : "Final payment"}</span>
                  <span>{formatMoney(milestone.amount)}</span>
                  <span>{(milestone.recordedAt || milestone.readyAt || milestone.updatedAt).slice(0, 10)}</span>
                  <StatusChip tone={paymentMilestoneTone(milestone.status)}>
                    {formatPaymentMilestoneStatus(milestone.status)}
                  </StatusChip>
                </div>
              ))}
            </div>
          ) : (
            <div className="erp-subcontract-empty-state">No payment milestone recorded</div>
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

function remainingFinishedGoodsQty(order: SubcontractOrder) {
  const plannedQty = normalizeDecimalInput(String(order.quantity), decimalScales.quantity);
  const receivedQty = normalizeDecimalInput(order.receivedQty ?? "0", decimalScales.quantity);
  const remaining =
    toScaledBigInt(plannedQty, decimalScales.quantity) - toScaledBigInt(receivedQty, decimalScales.quantity);
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

function formatMoney(value: number | string) {
  return `${new Intl.NumberFormat("en-US", { maximumFractionDigits: 0 }).format(Number(value))} VND`;
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

function formatPaymentMilestoneStatus(status: SubcontractPaymentMilestone["status"]) {
  switch (status) {
    case "recorded":
      return "Recorded";
    case "ready":
      return "Ready";
    case "blocked":
      return "Blocked";
    case "cancelled":
      return "Cancelled";
    case "pending":
    default:
      return "Pending";
  }
}

function paymentMilestoneTone(status: SubcontractPaymentMilestone["status"]): StatusTone {
  switch (status) {
    case "recorded":
    case "ready":
      return "success";
    case "blocked":
      return "danger";
    case "cancelled":
      return "normal";
    case "pending":
    default:
      return "warning";
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
