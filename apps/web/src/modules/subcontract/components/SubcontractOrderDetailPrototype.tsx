"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import {
  DataTable,
  EmptyState,
  ErrorState,
  LoadingState,
  StatusChip,
  type DataTableColumn
} from "@/shared/design-system/components";
import { formatProductionPlanQuantity } from "@/modules/production-planning/services/productionPlanService";
import {
  acceptSubcontractFinishedGoods,
  createSubcontractFactoryDispatch,
  formatSubcontractFactoryDispatchStatus,
  formatSubcontractDepositStatus,
  formatSubcontractOrderStatus,
  getSubcontractFactoryDispatches,
  getSubcontractOrder,
  issueSubcontractMaterials,
  markSubcontractFactoryDispatchReady,
  markSubcontractFactoryDispatchSent,
  partialAcceptSubcontractFinishedGoods,
  approveSubcontractSample,
  recordSubcontractFactoryDispatchResponse,
  rejectSubcontractSample,
  reportSubcontractFactoryDefect,
  receiveSubcontractFinishedGoods,
  startMassProductionSubcontractOrder,
  submitSubcontractSample,
  subcontractFactoryDispatchStatusTone,
  subcontractOrderStatusTone
} from "../services/subcontractOrderService";
import { subcontractTransferWarehouseOptions } from "../services/subcontractMaterialTransferService";
import {
  buildSubcontractOrderTimeline,
  productionFactoryOrderSourcePlanHref,
  subcontractOperationsHref,
  type SubcontractOrderTimelineItem
} from "../services/subcontractOrderTimeline";
import {
  buildSubcontractFactoryExecutionTracker,
  type FactoryExecutionWorkItem
} from "../services/subcontractFactoryExecutionTracker";
import {
  buildFactorySampleDecisionInput,
  buildFactorySampleSubmissionInput,
  buildSubcontractFactorySampleMassProduction,
  type FactoryMassProductionStatus,
  type FactorySampleStatus
} from "../services/subcontractFactorySampleMassProduction";
import {
  buildFactoryMaterialHandoverIssueInput,
  buildSubcontractFactoryMaterialHandover,
  type FactoryMaterialHandover,
  type FactoryMaterialHandoverLineDraft
} from "../services/subcontractFactoryMaterialHandover";
import {
  buildFactoryFinishedGoodsReceiptInput,
  buildSubcontractFactoryFinishedGoodsReceipt,
  type FactoryFinishedGoodsReceiptDraft
} from "../services/subcontractFactoryFinishedGoodsReceipt";
import {
  buildFactoryFinishedGoodsQcAcceptInput,
  buildFactoryFinishedGoodsQcPartialAcceptInput,
  buildFactoryFinishedGoodsQcRejectInput,
  buildSubcontractFactoryFinishedGoodsQcCloseout
} from "../services/subcontractFactoryFinishedGoodsQcCloseout";
import type {
  SubcontractFactoryDispatch,
  SubcontractFactoryClaimSeverity,
  SubcontractFinalPaymentStatus,
  SubcontractFinishedGoodsPackagingStatus,
  SubcontractFinishedGoodsReceipt,
  SubcontractOrderMaterialLine,
  SubcontractOrder,
  SubcontractMaterialTransfer,
  SubcontractSampleApproval
} from "../types";

type SubcontractOrderDetailPrototypeProps = {
  orderId: string;
};

const factoryFinishedGoodsLocationOptions = [{ label: "QC hold", value: "qc_hold", code: "QC-HOLD" }] as const;

export function SubcontractOrderDetailPrototype({ orderId }: SubcontractOrderDetailPrototypeProps) {
  const [order, setOrder] = useState<SubcontractOrder>();
  const [dispatches, setDispatches] = useState<SubcontractFactoryDispatch[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | undefined>();
  const [dispatchError, setDispatchError] = useState<string | undefined>();
  const [dispatchMessage, setDispatchMessage] = useState<string | undefined>();
  const [dispatchBusy, setDispatchBusy] = useState(false);
  const [responseNote, setResponseNote] = useState("");
  const [sourceWarehouseId, setSourceWarehouseId] = useState<string>(subcontractTransferWarehouseOptions[0].value);
  const [handoverReceiver, setHandoverReceiver] = useState("factory-receiver");
  const [handoverContact, setHandoverContact] = useState("");
  const [handoverVehicleNo, setHandoverVehicleNo] = useState("");
  const [handoverEvidenceName, setHandoverEvidenceName] = useState("");
  const [handoverNote, setHandoverNote] = useState("");
  const [signedHandover, setSignedHandover] = useState(true);
  const [materialDrafts, setMaterialDrafts] = useState<Record<string, FactoryMaterialHandoverLineDraft>>({});
  const [materialBusy, setMaterialBusy] = useState(false);
  const [materialError, setMaterialError] = useState<string | undefined>();
  const [materialMessage, setMaterialMessage] = useState<string | undefined>();
  const [latestMaterialTransfer, setLatestMaterialTransfer] = useState<SubcontractMaterialTransfer>();
  const [sampleCode, setSampleCode] = useState("");
  const [sampleFormulaVersion, setSampleFormulaVersion] = useState("");
  const [sampleEvidenceName, setSampleEvidenceName] = useState("");
  const [sampleDecisionReason, setSampleDecisionReason] = useState("Mẫu đạt theo tiêu chuẩn lưu");
  const [sampleStorageStatus, setSampleStorageStatus] = useState("retained_in_qa_cabinet");
  const [sampleNote, setSampleNote] = useState("");
  const [sampleBusy, setSampleBusy] = useState(false);
  const [sampleError, setSampleError] = useState<string | undefined>();
  const [sampleMessage, setSampleMessage] = useState<string | undefined>();
  const [latestSampleApproval, setLatestSampleApproval] = useState<SubcontractSampleApproval>();
  const [massBusy, setMassBusy] = useState(false);
  const [massError, setMassError] = useState<string | undefined>();
  const [massMessage, setMassMessage] = useState<string | undefined>();
  const [receiptWarehouseId, setReceiptWarehouseId] = useState<string>(subcontractTransferWarehouseOptions[0].value);
  const [receiptLocationId, setReceiptLocationId] = useState<string>(factoryFinishedGoodsLocationOptions[0].value);
  const [receiptDeliveryNoteNo, setReceiptDeliveryNoteNo] = useState("");
  const [receiptReceivedBy, setReceiptReceivedBy] = useState("warehouse-user");
  const [receiptEvidenceName, setReceiptEvidenceName] = useState("");
  const [receiptNote, setReceiptNote] = useState("");
  const [receiptDraft, setReceiptDraft] = useState<FactoryFinishedGoodsReceiptDraft>({
    receiveQty: "0",
    batchNo: "",
    lotNo: "",
    expiryDate: "",
    packagingStatus: "intact",
    note: ""
  });
  const [receiptBusy, setReceiptBusy] = useState(false);
  const [receiptError, setReceiptError] = useState<string | undefined>();
  const [receiptMessage, setReceiptMessage] = useState<string | undefined>();
  const [latestFinishedGoodsReceipt, setLatestFinishedGoodsReceipt] = useState<SubcontractFinishedGoodsReceipt>();
  const [qcAcceptedQty, setQcAcceptedQty] = useState("0");
  const [qcRejectedQty, setQcRejectedQty] = useState("0");
  const [qcAcceptedBy, setQcAcceptedBy] = useState("qc-lead");
  const [qcOpenedBy, setQcOpenedBy] = useState("qa-lead");
  const [qcOwnerId, setQcOwnerId] = useState("factory-owner");
  const [qcReasonCode, setQcReasonCode] = useState("QUALITY_FAIL");
  const [qcReason, setQcReason] = useState("Thành phẩm không đạt QC sau khi nhận từ nhà máy");
  const [qcSeverity, setQcSeverity] = useState<SubcontractFactoryClaimSeverity>("P2");
  const [qcEvidenceName, setQcEvidenceName] = useState("");
  const [qcEvidenceNote, setQcEvidenceNote] = useState("");
  const [qcNote, setQcNote] = useState("");
  const [qcBusy, setQcBusy] = useState(false);
  const [qcError, setQcError] = useState<string | undefined>();
  const [qcMessage, setQcMessage] = useState<string | undefined>();

  useEffect(() => {
    let active = true;

    setLoading(true);
    setError(undefined);
    Promise.all([getSubcontractOrder(orderId), getSubcontractFactoryDispatches(orderId)])
      .then(([nextOrder, nextDispatches]) => {
        if (active) {
          setOrder(nextOrder);
          setDispatches(nextDispatches);
        }
      })
      .catch((loadError) => {
        if (active) {
          setError(errorText(loadError));
        }
      })
      .finally(() => {
        if (active) {
          setLoading(false);
        }
      });

    return () => {
      active = false;
    };
  }, [orderId]);

  const latestDispatch = dispatches[0];
  const timeline = useMemo(
    () => (order ? buildSubcontractOrderTimeline(order, { dispatchStatus: latestDispatch?.status }) : []),
    [latestDispatch?.status, order]
  );
  const executionTracker = useMemo(
    () => (order ? buildSubcontractFactoryExecutionTracker(order, { dispatchStatus: latestDispatch?.status }) : undefined),
    [latestDispatch?.status, order]
  );
  const materialHandover = useMemo(() => (order ? buildSubcontractFactoryMaterialHandover(order) : undefined), [order]);
  const sampleMassGate = useMemo(
    () => (order ? buildSubcontractFactorySampleMassProduction(order, latestSampleApproval) : undefined),
    [latestSampleApproval, order]
  );
  const finishedGoodsReceiptGate = useMemo(
    () => (order ? buildSubcontractFactoryFinishedGoodsReceipt(order) : undefined),
    [order]
  );
  const finishedGoodsQcGate = useMemo(
    () => (order ? buildSubcontractFactoryFinishedGoodsQcCloseout(order, latestFinishedGoodsReceipt) : undefined),
    [latestFinishedGoodsReceipt, order]
  );
  const sourcePlanHref = useMemo(() => (order ? productionFactoryOrderSourcePlanHref(order) : undefined), [order]);
  const missingLotCount = useMemo(
    () =>
      materialHandover?.lines.filter((line) => {
        const draft = materialDrafts[line.id];
        return line.status === "ready" && line.lotTraceRequired && !draft?.batchNo?.trim();
      }).length ?? 0,
    [materialDrafts, materialHandover]
  );
  const canSubmitMaterialHandover =
    Boolean(materialHandover?.canIssue) &&
    signedHandover &&
    handoverReceiver.trim() !== "" &&
    missingLotCount === 0 &&
    !materialBusy;
  const canSubmitFactorySample =
    Boolean(sampleMassGate?.canSubmitSample) && sampleCode.trim() !== "" && sampleEvidenceName.trim() !== "" && !sampleBusy;
  const canApproveFactorySample =
    Boolean(sampleMassGate?.canApproveSample) && sampleDecisionReason.trim() !== "" && sampleStorageStatus.trim() !== "" && !sampleBusy;
  const canRejectFactorySample = Boolean(sampleMassGate?.canRejectSample) && sampleDecisionReason.trim() !== "" && !sampleBusy;
  const canStartFactoryMassProduction = Boolean(sampleMassGate?.canStartMassProduction) && !massBusy;
  const canReceiveFinishedGoods =
    Boolean(finishedGoodsReceiptGate?.canReceive) &&
    receiptDeliveryNoteNo.trim() !== "" &&
    receiptReceivedBy.trim() !== "" &&
    receiptDraft.batchNo.trim() !== "" &&
    receiptDraft.expiryDate.trim() !== "" &&
    isPositiveQuantity(receiptDraft.receiveQty) &&
    isWithinQuantity(receiptDraft.receiveQty, finishedGoodsReceiptGate?.remainingQty) &&
    !receiptBusy;
  const qcSplitQty = useMemo(() => quantitySum(qcAcceptedQty, qcRejectedQty), [qcAcceptedQty, qcRejectedQty]);
  const qcSplitMatchesRemaining = Boolean(
    finishedGoodsQcGate?.remainingQcQty && isQuantityEqual(qcSplitQty, finishedGoodsQcGate.remainingQcQty)
  );
  const canAcceptFinishedGoodsQc =
    Boolean(finishedGoodsQcGate?.canCloseout) && qcAcceptedBy.trim() !== "" && !qcBusy;
  const canPartialAcceptFinishedGoodsQc =
    Boolean(finishedGoodsQcGate?.canCloseout) &&
    Boolean(latestFinishedGoodsReceipt) &&
    isPositiveQuantity(qcAcceptedQty) &&
    isPositiveQuantity(qcRejectedQty) &&
    qcSplitMatchesRemaining &&
    qcAcceptedBy.trim() !== "" &&
    qcOpenedBy.trim() !== "" &&
    qcOwnerId.trim() !== "" &&
    qcReasonCode.trim() !== "" &&
    qcReason.trim() !== "" &&
    qcEvidenceName.trim() !== "" &&
    !qcBusy;
  const canRejectFinishedGoodsQc =
    Boolean(finishedGoodsQcGate?.canCloseout) &&
    Boolean(latestFinishedGoodsReceipt) &&
    qcOpenedBy.trim() !== "" &&
    qcOwnerId.trim() !== "" &&
    qcReasonCode.trim() !== "" &&
    qcReason.trim() !== "" &&
    qcEvidenceName.trim() !== "" &&
    !qcBusy;

  useEffect(() => {
    if (!order) {
      return;
    }
    const handover = buildSubcontractFactoryMaterialHandover(order);
    setMaterialDrafts(
      Object.fromEntries(
        handover.lines.map((line) => [
          line.id,
          {
            issueQty: line.remainingQty,
            batchNo: "",
            sourceBinId: "",
            note: ""
          }
        ])
      )
    );
  }, [order?.id, order?.version]);

  useEffect(() => {
    if (!order) {
      return;
    }
    setSampleCode(`${order.orderNo}-SAMPLE-A`);
    setSampleFormulaVersion(order.specVersion);
    setSampleEvidenceName("");
    setSampleNote("");
    setSampleDecisionReason(order.sampleRejectReason || "Mẫu đạt theo tiêu chuẩn lưu");
    setSampleStorageStatus("retained_in_qa_cabinet");
    setSampleError(undefined);
    setSampleMessage(undefined);
    setMassError(undefined);
    setMassMessage(undefined);
  }, [order?.id]);

  useEffect(() => {
    if (!order) {
      return;
    }
    const gate = buildSubcontractFactoryFinishedGoodsReceipt(order);
    setReceiptDraft({
      receiveQty: compactQuantity(gate.remainingQty),
      batchNo: "",
      lotNo: "",
      expiryDate: "",
      packagingStatus: "intact",
      note: ""
    });
  }, [order?.id, order?.version]);

  useEffect(() => {
    if (!finishedGoodsQcGate) {
      return;
    }
    setQcAcceptedQty(compactQuantity(finishedGoodsQcGate.remainingQcQty));
    setQcRejectedQty("0");
    setQcError(undefined);
    setQcMessage(undefined);
  }, [finishedGoodsQcGate?.receiptNo, finishedGoodsQcGate?.remainingQcQty, order?.id]);

  const reloadFactoryDispatches = async (nextOrder?: SubcontractOrder) => {
    const targetOrder = nextOrder ?? order;
    if (!targetOrder) {
      return;
    }
    const nextDispatches = await getSubcontractFactoryDispatches(targetOrder.id);
    setDispatches(nextDispatches);
  };

  const runDispatchAction = async (
    action: () => Promise<{ order: SubcontractOrder; dispatch: SubcontractFactoryDispatch }>,
    message: string
  ) => {
    setDispatchBusy(true);
    setDispatchError(undefined);
    setDispatchMessage(undefined);
    try {
      const result = await action();
      setOrder(result.order);
      await reloadFactoryDispatches(result.order);
      setDispatchMessage(message);
      if (result.dispatch.status === "confirmed") {
        setResponseNote("");
      }
    } catch (cause) {
      setDispatchError(errorText(cause));
    } finally {
      setDispatchBusy(false);
    }
  };

  const updateMaterialDraft = (lineId: string, field: keyof FactoryMaterialHandoverLineDraft, value: string) => {
    setMaterialDrafts((current) => ({
      ...current,
      [lineId]: {
        issueQty: current[lineId]?.issueQty ?? "",
        batchNo: current[lineId]?.batchNo ?? "",
        sourceBinId: current[lineId]?.sourceBinId ?? "",
        note: current[lineId]?.note ?? "",
        [field]: value
      }
    }));
  };

  const updateFinishedGoodsReceiptDraft = (field: keyof FactoryFinishedGoodsReceiptDraft, value: string) => {
    setReceiptDraft((current) => ({
      ...current,
      [field]: value
    }));
  };

  const handleFactoryMaterialHandover = async () => {
    if (!order) {
      return;
    }

    setMaterialBusy(true);
    setMaterialError(undefined);
    setMaterialMessage(undefined);
    try {
      const warehouse =
        subcontractTransferWarehouseOptions.find((option) => option.value === sourceWarehouseId) ??
        subcontractTransferWarehouseOptions[0];
      const result = await issueSubcontractMaterials(
        buildFactoryMaterialHandoverIssueInput({
          order,
          sourceWarehouseId: warehouse.value,
          sourceWarehouseCode: warehouse.code,
          handoverBy: "warehouse-user",
          receivedBy: handoverReceiver,
          receiverContact: handoverContact,
          vehicleNo: handoverVehicleNo,
          note: handoverNote,
          evidenceFileName: handoverEvidenceName,
          lineDrafts: materialDrafts
        })
      );

      setOrder(result.order);
      setLatestMaterialTransfer(result.transfer);
      setMaterialMessage(`${result.transfer.transferNo} đã ghi nhận ${result.stockMovements.length} dòng xuất vật tư.`);
    } catch (cause) {
      setMaterialError(errorText(cause));
    } finally {
      setMaterialBusy(false);
    }
  };

  const handleSubmitFactorySample = async () => {
    if (!order) {
      return;
    }

    setSampleBusy(true);
    setSampleError(undefined);
    setSampleMessage(undefined);
    try {
      const result = await submitSubcontractSample(
        buildFactorySampleSubmissionInput({
          order,
          sampleCode,
          formulaVersion: sampleFormulaVersion,
          evidenceFileName: sampleEvidenceName,
          note: sampleNote,
          submittedAt: new Date().toISOString()
        })
      );

      setOrder(result.order);
      setLatestSampleApproval(result.sampleApproval);
      setSampleMessage(`${result.sampleApproval.sampleCode} đã ghi nhận mẫu chờ duyệt.`);
    } catch (cause) {
      setSampleError(errorText(cause));
    } finally {
      setSampleBusy(false);
    }
  };

  const handleDecideFactorySample = async (decision: "approve" | "reject") => {
    if (!order) {
      return;
    }

    setSampleBusy(true);
    setSampleError(undefined);
    setSampleMessage(undefined);
    try {
      const input = buildFactorySampleDecisionInput({
        order,
        sampleApproval: latestSampleApproval,
        decision,
        reason: sampleDecisionReason,
        storageStatus: sampleStorageStatus,
        decisionAt: new Date().toISOString()
      });
      const result =
        decision === "approve" ? await approveSubcontractSample(input) : await rejectSubcontractSample(input);

      setOrder(result.order);
      setLatestSampleApproval(result.sampleApproval);
      setSampleMessage(decision === "approve" ? "Mẫu đã đạt; có thể bắt đầu sản xuất hàng loạt." : "Mẫu đã bị từ chối; cần gửi lại mẫu.");
    } catch (cause) {
      setSampleError(errorText(cause));
    } finally {
      setSampleBusy(false);
    }
  };

  const handleStartFactoryMassProduction = async () => {
    if (!order) {
      return;
    }

    setMassBusy(true);
    setMassError(undefined);
    setMassMessage(undefined);
    try {
      const result = await startMassProductionSubcontractOrder(order.id, order.version);
      setOrder(result.order);
      setMassMessage("Đã bắt đầu sản xuất hàng loạt.");
    } catch (cause) {
      setMassError(errorText(cause));
    } finally {
      setMassBusy(false);
    }
  };

  const handleFactoryFinishedGoodsReceipt = async () => {
    if (!order) {
      return;
    }

    setReceiptBusy(true);
    setReceiptError(undefined);
    setReceiptMessage(undefined);
    try {
      const warehouse =
        subcontractTransferWarehouseOptions.find((option) => option.value === receiptWarehouseId) ??
        subcontractTransferWarehouseOptions[0];
      const location =
        factoryFinishedGoodsLocationOptions.find((option) => option.value === receiptLocationId) ??
        factoryFinishedGoodsLocationOptions[0];
      const result = await receiveSubcontractFinishedGoods(
        buildFactoryFinishedGoodsReceiptInput({
          order,
          warehouseId: warehouse.value,
          warehouseCode: warehouse.code,
          locationId: location.value,
          locationCode: location.code,
          deliveryNoteNo: receiptDeliveryNoteNo,
          receivedBy: receiptReceivedBy,
          evidenceFileName: receiptEvidenceName,
          note: receiptNote,
          receivedAt: new Date().toISOString(),
          draft: receiptDraft
        })
      );

      setOrder(result.order);
      setLatestFinishedGoodsReceipt(result.receipt);
      setReceiptMessage(`${result.receipt.receiptNo} đã nhận ${result.stockMovements.length} dòng vào QC hold.`);
      setReceiptDeliveryNoteNo("");
      setReceiptEvidenceName("");
      setReceiptNote("");
    } catch (cause) {
      setReceiptError(errorText(cause));
    } finally {
      setReceiptBusy(false);
    }
  };

  const handleFactoryFinishedGoodsQcAccept = async () => {
    if (!order) {
      return;
    }

    setQcBusy(true);
    setQcError(undefined);
    setQcMessage(undefined);
    try {
      const result = await acceptSubcontractFinishedGoods(
        buildFactoryFinishedGoodsQcAcceptInput({
          order,
          latestReceipt: latestFinishedGoodsReceipt,
          acceptedBy: qcAcceptedBy,
          acceptedAt: new Date().toISOString(),
          note: qcNote
        })
      );

      setOrder(result.order);
      setQcMessage(`Đã QC đạt ${result.stockMovements.length} dòng và chuyển thành tồn khả dụng.`);
    } catch (cause) {
      setQcError(errorText(cause));
    } finally {
      setQcBusy(false);
    }
  };

  const handleFactoryFinishedGoodsQcPartialAccept = async () => {
    if (!order || !latestFinishedGoodsReceipt) {
      return;
    }

    setQcBusy(true);
    setQcError(undefined);
    setQcMessage(undefined);
    try {
      const result = await partialAcceptSubcontractFinishedGoods(
        buildFactoryFinishedGoodsQcPartialAcceptInput({
          order,
          latestReceipt: latestFinishedGoodsReceipt,
          acceptedQty: qcAcceptedQty,
          rejectedQty: qcRejectedQty,
          acceptedBy: qcAcceptedBy,
          acceptedAt: new Date().toISOString(),
          openedBy: qcOpenedBy,
          openedAt: new Date().toISOString(),
          ownerId: qcOwnerId,
          reasonCode: qcReasonCode,
          reason: qcReason,
          severity: qcSeverity,
          evidenceFileName: qcEvidenceName,
          evidenceNote: qcEvidenceNote,
          note: qcNote
        })
      );

      setOrder(result.order);
      setQcMessage(`${result.claim.claimNo} đã mở cho ${formatProductionPlanQuantity(result.claim.affectedQty, result.claim.uomCode)} không đạt; phần đạt đã vào tồn khả dụng.`);
    } catch (cause) {
      setQcError(errorText(cause));
    } finally {
      setQcBusy(false);
    }
  };

  const handleFactoryFinishedGoodsQcReject = async () => {
    if (!order || !latestFinishedGoodsReceipt) {
      return;
    }

    setQcBusy(true);
    setQcError(undefined);
    setQcMessage(undefined);
    try {
      const result = await reportSubcontractFactoryDefect(
        buildFactoryFinishedGoodsQcRejectInput({
          order,
          latestReceipt: latestFinishedGoodsReceipt,
          openedBy: qcOpenedBy,
          openedAt: new Date().toISOString(),
          ownerId: qcOwnerId,
          reasonCode: qcReasonCode,
          reason: qcReason,
          severity: qcSeverity,
          evidenceFileName: qcEvidenceName,
          evidenceNote: qcEvidenceNote
        })
      );

      setOrder(result.order);
      setQcMessage(`${result.claim.claimNo} đã mở; không chuyển thành phẩm lỗi vào tồn khả dụng.`);
    } catch (cause) {
      setQcError(errorText(cause));
    } finally {
      setQcBusy(false);
    }
  };

  if (loading) {
    return <LoadingState title="Đang tải lệnh nhà máy" />;
  }

  if (error || !order) {
    return (
      <ErrorState
        title="Không tải được lệnh nhà máy"
        description={error ?? "Không tìm thấy lệnh nhà máy."}
        action={
          <Link className="erp-button erp-button--secondary" href="/production">
            Quay lại sản xuất
          </Link>
        }
      />
    );
  }

  return (
    <main className="erp-masterdata-page erp-purchase-detail-page">
      <header className="erp-page-header">
        <div>
          <span className="erp-production-step-label">Lệnh nhà máy / Gia công ngoài</span>
          <h1 className="erp-page-title">{order.orderNo}</h1>
          <p className="erp-page-description">
            {order.factoryName} / {order.sku} / {formatProductionPlanQuantity(String(order.quantity), order.uomCode ?? "PCS")}
          </p>
        </div>
        <div className="erp-page-actions">
          {sourcePlanHref ? (
            <Link className="erp-button erp-button--secondary" href={sourcePlanHref}>
              Mở kế hoạch
            </Link>
          ) : null}
          <Link className="erp-button erp-button--secondary" href={subcontractOperationsHref(order)}>
            Mở xử lý lệnh
          </Link>
          <Link className="erp-button erp-button--secondary" href="/production">
            Quay lại sản xuất
          </Link>
        </div>
      </header>

      <section className="erp-production-selected-plan-card" aria-label="Tóm tắt lệnh nhà máy">
        <div className="erp-production-selected-plan-main">
          <span className="erp-production-step-label">Trạng thái lệnh</span>
          <h2>{formatSubcontractOrderStatus(order.status)}</h2>
          <p>
            Lệnh gửi nhà máy ngoài từ kế hoạch {order.sourceProductionPlanNo ?? "-"}; thành phẩm chỉ vào tồn khả dụng sau khi nhận về và QC đạt.
          </p>
          <div className="erp-production-selected-plan-badges">
            <StatusChip tone={subcontractOrderStatusTone(order.status)}>{formatSubcontractOrderStatus(order.status)}</StatusChip>
            <StatusChip tone={order.sampleRequired ? "warning" : "normal"}>
              {order.sampleRequired ? "Cần duyệt mẫu" : "Không yêu cầu mẫu"}
            </StatusChip>
            <StatusChip tone={closeoutTone(order)}>{closeoutLabel(order)}</StatusChip>
          </div>
        </div>
        <dl className="erp-production-selected-plan-meta">
          <div>
            <dt>Nhà máy</dt>
            <dd>{order.factoryName}</dd>
          </div>
          <div>
            <dt>Ngày nhận dự kiến</dt>
            <dd>{formatDate(order.expectedDeliveryDate)}</dd>
          </div>
          <div>
            <dt>Kế hoạch nguồn</dt>
            <dd>{order.sourceProductionPlanNo ?? "-"}</dd>
          </div>
          <div>
            <dt>Đã nhận / Đạt QC</dt>
            <dd>
              {formatProductionPlanQuantity(order.receivedQty ?? "0", order.uomCode ?? "PCS")} /{" "}
              {formatProductionPlanQuantity(order.acceptedQty ?? "0", order.uomCode ?? "PCS")}
            </dd>
          </div>
          <div>
            <dt>Đặt cọc</dt>
            <dd>{formatSubcontractDepositStatus(order.depositStatus)}</dd>
          </div>
          <div>
            <dt>Thanh toán cuối</dt>
            <dd>{formatFinalPaymentStatus(order.finalPaymentStatus)}</dd>
          </div>
        </dl>
      </section>

      {executionTracker ? (
        <section className="erp-masterdata-list-card" id="factory-execution-tracker">
          <header className="erp-section-header">
            <div>
              <h2 className="erp-section-title">Theo dõi thực thi nhà máy</h2>
              <p className="erp-page-description">
                Một danh sách công việc cho lệnh này, từ gửi nhà máy, đặt cọc, bàn giao vật tư, duyệt mẫu, sản xuất, nhận hàng đến QC và thanh toán.
              </p>
            </div>
            <StatusChip tone={executionTracker.currentGate.tone}>{factoryExecutionStatusLabel(executionTracker.currentGate.status)}</StatusChip>
          </header>
          <div className="erp-production-selected-plan-main">
            <span className="erp-production-step-label">Việc cần xử lý</span>
            <h3>{executionTracker.currentGate.title}</h3>
            <p>{executionTracker.currentGate.description}</p>
            <div className="erp-production-selected-plan-badges">
              <StatusChip tone={executionTracker.currentGate.tone}>{factoryExecutionStatusLabel(executionTracker.currentGate.status)}</StatusChip>
              <StatusChip tone="normal">{executionTracker.currentGate.metric}</StatusChip>
            </div>
            {executionTracker.currentGate.action ? (
              executionTracker.currentGate.action.disabled ? (
                <button className="erp-button erp-button--secondary erp-button--compact" type="button" disabled>
                  {executionTracker.currentGate.action.label}
                </button>
              ) : (
                <Link className="erp-button erp-button--secondary erp-button--compact" href={executionTracker.currentGate.action.href}>
                  {executionTracker.currentGate.action.label}
                </Link>
              )
            ) : null}
          </div>
          <DataTable
            columns={factoryExecutionColumns}
            rows={executionTracker.items}
            getRowKey={(item) => item.id}
            preserveColumnWidths
            emptyState={<EmptyState title="Chưa có công việc thực thi" />}
          />
        </section>
      ) : null}

      <FactoryMaterialHandoverSection
        canSubmit={canSubmitMaterialHandover}
        handover={materialHandover}
        handoverContact={handoverContact}
        handoverEvidenceName={handoverEvidenceName}
        handoverNote={handoverNote}
        handoverReceiver={handoverReceiver}
        handoverVehicleNo={handoverVehicleNo}
        latestTransfer={latestMaterialTransfer}
        materialBusy={materialBusy}
        materialDrafts={materialDrafts}
        materialError={materialError}
        materialMessage={materialMessage}
        missingLotCount={missingLotCount}
        setHandoverContact={setHandoverContact}
        setHandoverEvidenceName={setHandoverEvidenceName}
        setHandoverNote={setHandoverNote}
        setHandoverReceiver={setHandoverReceiver}
        setHandoverVehicleNo={setHandoverVehicleNo}
        setSignedHandover={setSignedHandover}
        setSourceWarehouseId={setSourceWarehouseId}
        signedHandover={signedHandover}
        sourceWarehouseId={sourceWarehouseId}
        updateMaterialDraft={updateMaterialDraft}
        onSubmit={handleFactoryMaterialHandover}
      />

      <FactorySampleApprovalSection
        canApprove={canApproveFactorySample}
        canReject={canRejectFactorySample}
        canSubmit={canSubmitFactorySample}
        decisionReason={sampleDecisionReason}
        evidenceName={sampleEvidenceName}
        formulaVersion={sampleFormulaVersion}
        gate={sampleMassGate}
        latestSample={latestSampleApproval}
        note={sampleNote}
        sampleBusy={sampleBusy}
        sampleCode={sampleCode}
        sampleError={sampleError}
        sampleMessage={sampleMessage}
        setDecisionReason={setSampleDecisionReason}
        setEvidenceName={setSampleEvidenceName}
        setFormulaVersion={setSampleFormulaVersion}
        setNote={setSampleNote}
        setSampleCode={setSampleCode}
        setStorageStatus={setSampleStorageStatus}
        storageStatus={sampleStorageStatus}
        onApprove={() => handleDecideFactorySample("approve")}
        onReject={() => handleDecideFactorySample("reject")}
        onSubmit={handleSubmitFactorySample}
      />

      <FactoryMassProductionSection
        canStart={canStartFactoryMassProduction}
        gate={sampleMassGate}
        massBusy={massBusy}
        massError={massError}
        massMessage={massMessage}
        order={order}
        onStart={handleStartFactoryMassProduction}
      />

      <FactoryFinishedGoodsReceiptSection
        canReceive={canReceiveFinishedGoods}
        deliveryNoteNo={receiptDeliveryNoteNo}
        draft={receiptDraft}
        evidenceName={receiptEvidenceName}
        gate={finishedGoodsReceiptGate}
        latestReceipt={latestFinishedGoodsReceipt}
        locationId={receiptLocationId}
        note={receiptNote}
        receiptBusy={receiptBusy}
        receiptError={receiptError}
        receiptMessage={receiptMessage}
        receivedBy={receiptReceivedBy}
        setDeliveryNoteNo={setReceiptDeliveryNoteNo}
        setEvidenceName={setReceiptEvidenceName}
        setLocationId={setReceiptLocationId}
        setNote={setReceiptNote}
        setReceivedBy={setReceiptReceivedBy}
        setWarehouseId={setReceiptWarehouseId}
        updateDraft={updateFinishedGoodsReceiptDraft}
        warehouseId={receiptWarehouseId}
        onReceive={handleFactoryFinishedGoodsReceipt}
      />

      <FactoryFinishedGoodsQcCloseoutSection
        acceptedBy={qcAcceptedBy}
        acceptedQty={qcAcceptedQty}
        canAccept={canAcceptFinishedGoodsQc}
        canPartialAccept={canPartialAcceptFinishedGoodsQc}
        canReject={canRejectFinishedGoodsQc}
        evidenceName={qcEvidenceName}
        evidenceNote={qcEvidenceNote}
        gate={finishedGoodsQcGate}
        latestReceipt={latestFinishedGoodsReceipt}
        note={qcNote}
        openedBy={qcOpenedBy}
        ownerId={qcOwnerId}
        qcBusy={qcBusy}
        qcError={qcError}
        qcMessage={qcMessage}
        reason={qcReason}
        reasonCode={qcReasonCode}
        rejectedQty={qcRejectedQty}
        setAcceptedBy={setQcAcceptedBy}
        setAcceptedQty={setQcAcceptedQty}
        setEvidenceName={setQcEvidenceName}
        setEvidenceNote={setQcEvidenceNote}
        setNote={setQcNote}
        setOpenedBy={setQcOpenedBy}
        setOwnerId={setQcOwnerId}
        setReason={setQcReason}
        setReasonCode={setQcReasonCode}
        setRejectedQty={setQcRejectedQty}
        setSeverity={setQcSeverity}
        severity={qcSeverity}
        splitMatchesRemaining={qcSplitMatchesRemaining}
        splitQty={qcSplitQty}
        onAccept={handleFactoryFinishedGoodsQcAccept}
        onPartialAccept={handleFactoryFinishedGoodsQcPartialAccept}
        onReject={handleFactoryFinishedGoodsQcReject}
      />

      <section className="erp-masterdata-list-card" id="factory-dispatch">
        <header className="erp-section-header">
          <div>
            <h2 className="erp-section-title">Gửi nhà máy</h2>
            <p className="erp-page-description">
              Tạo bộ gửi lệnh, ghi nhận đã gửi thủ công và phản hồi xác nhận từ nhà máy. Chưa gửi qua email/Zalo trong sprint này.
            </p>
          </div>
          <FactoryDispatchActions
            busy={dispatchBusy}
            dispatch={latestDispatch}
            order={order}
            responseNote={responseNote}
            setResponseNote={setResponseNote}
            onCreate={() =>
              runDispatchAction(
                () => createSubcontractFactoryDispatch({ order, note: "Tạo bộ gửi nhà máy từ chi tiết lệnh." }),
                "Đã tạo bộ gửi nhà máy"
              )
            }
            onReady={(dispatch) =>
              runDispatchAction(() => markSubcontractFactoryDispatchReady(order, dispatch), "Bộ gửi đã sẵn sàng")
            }
            onSent={(dispatch) =>
              runDispatchAction(
                () =>
                  markSubcontractFactoryDispatchSent({
                    order,
                    dispatch,
                    sentBy: "subcontract-user",
                    sentAt: new Date().toISOString(),
                    note: "Đã gửi thủ công ngoài hệ thống.",
                    evidence: [
                      {
                        id: `${dispatch.id}-manual-send`,
                        evidenceType: "manual_send",
                        objectKey: `manual-factory-dispatch/${dispatch.id}`,
                        note: "Ghi nhận gửi thủ công; chưa tích hợp email/Zalo."
                      }
                    ]
                  }),
                "Đã ghi nhận gửi nhà máy"
              )
            }
            onResponse={(dispatch, status) =>
              runDispatchAction(
                () =>
                  recordSubcontractFactoryDispatchResponse({
                    order,
                    dispatch,
                    responseStatus: status,
                    responseBy: "factory-user",
                    respondedAt: new Date().toISOString(),
                    responseNote: responseNote.trim()
                  }),
                status === "confirmed" ? "Nhà máy đã xác nhận" : "Đã ghi nhận phản hồi nhà máy"
              )
            }
          />
        </header>
        {dispatchError ? (
          <p className="erp-form-error" role="alert">
            {dispatchError}
          </p>
        ) : null}
        {dispatchMessage ? (
          <p className="erp-form-success" role="status">
            {dispatchMessage}
          </p>
        ) : null}
        {latestDispatch ? (
          <FactoryDispatchSummary dispatch={latestDispatch} />
        ) : (
          <EmptyState title="Chưa có bộ gửi nhà máy cho lệnh này" />
        )}
      </section>

      <section className="erp-masterdata-list-card">
        <header className="erp-section-header">
          <div>
            <h2 className="erp-section-title">Timeline lệnh nhà máy</h2>
            <p className="erp-page-description">
              Theo dõi lệnh từ kế hoạch sản xuất, xác nhận nhà máy, xuất vật tư, nhận thành phẩm, QC đến đóng lệnh.
            </p>
          </div>
        </header>
        <ol className="erp-document-timeline" aria-label="Timeline lệnh nhà máy">
          {timeline.map((item) => (
            <TimelineItem item={item} key={item.id} />
          ))}
        </ol>
      </section>

      <section className="erp-masterdata-list-card">
        <header className="erp-section-header">
          <div>
            <h2 className="erp-section-title">Vật tư xuất cho nhà máy</h2>
            <p className="erp-page-description">Các dòng nguyên liệu/bao bì cần bàn giao cho nhà máy theo lệnh này.</p>
          </div>
          <Link className="erp-button erp-button--secondary" href="#factory-material-handover">
            Mở xuất vật tư
          </Link>
        </header>
        <DataTable
          columns={materialLineColumns}
          rows={order.materialLines}
          getRowKey={(line) => line.id}
          pagination
          preserveColumnWidths
          emptyState={<EmptyState title="Lệnh này chưa có dòng vật tư" />}
        />
      </section>
    </main>
  );
}

function FactoryMaterialHandoverSection({
  canSubmit,
  handover,
  handoverContact,
  handoverEvidenceName,
  handoverNote,
  handoverReceiver,
  handoverVehicleNo,
  latestTransfer,
  materialBusy,
  materialDrafts,
  materialError,
  materialMessage,
  missingLotCount,
  setHandoverContact,
  setHandoverEvidenceName,
  setHandoverNote,
  setHandoverReceiver,
  setHandoverVehicleNo,
  setSignedHandover,
  setSourceWarehouseId,
  signedHandover,
  sourceWarehouseId,
  updateMaterialDraft,
  onSubmit
}: {
  canSubmit: boolean;
  handover?: FactoryMaterialHandover;
  handoverContact: string;
  handoverEvidenceName: string;
  handoverNote: string;
  handoverReceiver: string;
  handoverVehicleNo: string;
  latestTransfer?: SubcontractMaterialTransfer;
  materialBusy: boolean;
  materialDrafts: Record<string, FactoryMaterialHandoverLineDraft>;
  materialError?: string;
  materialMessage?: string;
  missingLotCount: number;
  setHandoverContact: (value: string) => void;
  setHandoverEvidenceName: (value: string) => void;
  setHandoverNote: (value: string) => void;
  setHandoverReceiver: (value: string) => void;
  setHandoverVehicleNo: (value: string) => void;
  setSignedHandover: (value: boolean) => void;
  setSourceWarehouseId: (value: string) => void;
  signedHandover: boolean;
  sourceWarehouseId: string;
  updateMaterialDraft: (lineId: string, field: keyof FactoryMaterialHandoverLineDraft, value: string) => void;
  onSubmit: () => void;
}) {
  return (
    <section className="erp-masterdata-list-card" id="factory-material-handover">
      <header className="erp-section-header">
        <div>
          <h2 className="erp-section-title">Bàn giao vật tư cho nhà máy</h2>
          <p className="erp-page-description">
            Chọn kho xuất, lô/bin và số lượng bàn giao cho nhà máy. Hệ thống ghi nhận transfer, stock movement và cập nhật trạng thái lệnh khi đủ vật tư.
          </p>
        </div>
        {handover ? (
          <StatusChip tone={materialHandoverStatusTone(handover.status)}>
            {materialHandoverStatusLabel(handover.status)}
          </StatusChip>
        ) : null}
      </header>
      {materialError ? (
        <p className="erp-form-error" role="alert">
          {materialError}
        </p>
      ) : null}
      {handover?.blockedReason ? (
        <p className="erp-form-error" role="status">
          {handover.blockedReason}
        </p>
      ) : null}
      {materialMessage ? (
        <p className="erp-form-success" role="status">
          {materialMessage}
        </p>
      ) : null}

      <div className="erp-subcontract-transfer-controls">
        <label className="erp-field">
          <span>Kho xuất</span>
          <select className="erp-input" value={sourceWarehouseId} onChange={(event) => setSourceWarehouseId(event.target.value)}>
            {subcontractTransferWarehouseOptions.map((warehouse) => (
              <option key={warehouse.value} value={warehouse.value}>
                {warehouse.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Người nhận nhà máy</span>
          <input className="erp-input" type="text" value={handoverReceiver} onChange={(event) => setHandoverReceiver(event.target.value)} />
        </label>
        <label className="erp-field">
          <span>Liên hệ</span>
          <input className="erp-input" type="text" value={handoverContact} onChange={(event) => setHandoverContact(event.target.value)} />
        </label>
        <label className="erp-field">
          <span>Xe / vận chuyển</span>
          <input className="erp-input" type="text" value={handoverVehicleNo} onChange={(event) => setHandoverVehicleNo(event.target.value)} />
        </label>
        <label className="erp-subcontract-sample-toggle">
          <input checked={signedHandover} type="checkbox" onChange={(event) => setSignedHandover(event.target.checked)} />
          <span>Đã có biên bản</span>
        </label>
      </div>
      <label className="erp-field erp-subcontract-note-field">
        <span>File / mã biên bản bàn giao</span>
        <input className="erp-input" type="text" value={handoverEvidenceName} onChange={(event) => setHandoverEvidenceName(event.target.value)} />
      </label>
      <label className="erp-field erp-subcontract-note-field">
        <span>Ghi chú bàn giao</span>
        <input className="erp-input" type="text" value={handoverNote} onChange={(event) => setHandoverNote(event.target.value)} />
      </label>

      <div className="erp-subcontract-line-list" aria-label="Dòng vật tư bàn giao cho nhà máy">
        {(handover?.lines ?? []).map((line) => {
          const draft = materialDrafts[line.id] ?? { issueQty: line.remainingQty, batchNo: "", sourceBinId: "", note: "" };
          const disabled = line.status === "complete" || !handover?.canIssue || materialBusy;

          return (
            <div className="erp-subcontract-line-item erp-subcontract-line-item--editable" key={line.id}>
              <span className="erp-masterdata-product-cell">
                <strong>{line.skuCode}</strong>
                <small>{line.itemName}</small>
              </span>
              <span>
                Cần {formatProductionPlanQuantity(line.plannedQty, line.uomCode)} / đã giao{" "}
                {formatProductionPlanQuantity(line.issuedQty, line.uomCode)} / còn{" "}
                {formatProductionPlanQuantity(line.remainingQty, line.uomCode)}
              </span>
              <label className="erp-field">
                <span>Số lượng giao</span>
                <input
                  className="erp-input"
                  disabled={disabled}
                  inputMode="decimal"
                  type="text"
                  value={draft.issueQty}
                  onChange={(event) => updateMaterialDraft(line.id, "issueQty", event.target.value)}
                />
              </label>
              <label className="erp-field">
                <span>Lô / batch</span>
                <input
                  className="erp-input"
                  disabled={disabled}
                  type="text"
                  value={draft.batchNo}
                  onChange={(event) => updateMaterialDraft(line.id, "batchNo", event.target.value)}
                />
              </label>
              <label className="erp-field">
                <span>Bin</span>
                <input
                  className="erp-input"
                  disabled={disabled}
                  type="text"
                  value={draft.sourceBinId}
                  onChange={(event) => updateMaterialDraft(line.id, "sourceBinId", event.target.value)}
                />
              </label>
              <StatusChip tone={line.status === "complete" ? "success" : line.lotTraceRequired ? "warning" : "normal"}>
                {materialLineStatusLabel(line.status, line.lotTraceRequired)}
              </StatusChip>
            </div>
          );
        })}
      </div>

      <div className="erp-subcontract-actions">
        <button className="erp-button erp-button--primary" type="button" disabled={!canSubmit} onClick={onSubmit}>
          Ghi nhận bàn giao vật tư
        </button>
        <StatusChip tone="normal">
          {handover?.completeLines ?? 0}/{handover?.totalLines ?? 0} dòng đủ
        </StatusChip>
        {missingLotCount > 0 ? <StatusChip tone="warning">Thiếu lô cho {missingLotCount} dòng</StatusChip> : null}
      </div>

      {latestTransfer ? (
        <div className="erp-production-selected-plan-main">
          <div className="erp-production-selected-plan-badges">
            <StatusChip tone={factoryMaterialTransferTone(latestTransfer.status)}>
              {factoryMaterialTransferStatusLabel(latestTransfer.status)}
            </StatusChip>
            <StatusChip tone="normal">{latestTransfer.transferNo}</StatusChip>
          </div>
          <dl className="erp-production-selected-plan-meta">
            <div>
              <dt>Kho xuất</dt>
              <dd>{latestTransfer.sourceWarehouseCode}</dd>
            </div>
            <div>
              <dt>Nhà máy</dt>
              <dd>{latestTransfer.factoryName}</dd>
            </div>
            <div>
              <dt>Dòng vật tư</dt>
              <dd>{latestTransfer.lines.length}</dd>
            </div>
            <div>
              <dt>Stock movement</dt>
              <dd>{latestTransfer.stockMovements.length}</dd>
            </div>
            <div>
              <dt>Bằng chứng cần kèm</dt>
              <dd>{latestTransfer.attachmentPlaceholders.length}</dd>
            </div>
            <div>
              <dt>Người tạo</dt>
              <dd>{latestTransfer.createdBy}</dd>
            </div>
          </dl>
        </div>
      ) : null}
    </section>
  );
}

function FactorySampleApprovalSection({
  canApprove,
  canReject,
  canSubmit,
  decisionReason,
  evidenceName,
  formulaVersion,
  gate,
  latestSample,
  note,
  sampleBusy,
  sampleCode,
  sampleError,
  sampleMessage,
  setDecisionReason,
  setEvidenceName,
  setFormulaVersion,
  setNote,
  setSampleCode,
  setStorageStatus,
  storageStatus,
  onApprove,
  onReject,
  onSubmit
}: {
  canApprove: boolean;
  canReject: boolean;
  canSubmit: boolean;
  decisionReason: string;
  evidenceName: string;
  formulaVersion: string;
  gate?: ReturnType<typeof buildSubcontractFactorySampleMassProduction>;
  latestSample?: SubcontractSampleApproval;
  note: string;
  sampleBusy: boolean;
  sampleCode: string;
  sampleError?: string;
  sampleMessage?: string;
  setDecisionReason: (value: string) => void;
  setEvidenceName: (value: string) => void;
  setFormulaVersion: (value: string) => void;
  setNote: (value: string) => void;
  setSampleCode: (value: string) => void;
  setStorageStatus: (value: string) => void;
  storageStatus: string;
  onApprove: () => void;
  onReject: () => void;
  onSubmit: () => void;
}) {
  return (
    <section className="erp-masterdata-list-card" id="factory-sample-approval">
      <header className="erp-section-header">
        <div>
          <h2 className="erp-section-title">Duyệt mẫu nhà máy</h2>
          <p className="erp-page-description">
            Ghi nhận mẫu nhà máy gửi, lưu bằng chứng mẫu và chốt đạt/không đạt trước khi mở sản xuất hàng loạt.
          </p>
        </div>
        {gate ? <StatusChip tone={sampleGateStatusTone(gate.sampleStatus)}>{sampleGateStatusLabel(gate.sampleStatus)}</StatusChip> : null}
      </header>
      {sampleError ? (
        <p className="erp-form-error" role="alert">
          {sampleError}
        </p>
      ) : null}
      {gate?.sampleBlockedReason ? (
        <p className="erp-form-error" role="status">
          {gate.sampleBlockedReason}
        </p>
      ) : null}
      {sampleMessage ? (
        <p className="erp-form-success" role="status">
          {sampleMessage}
        </p>
      ) : null}

      <div className="erp-subcontract-transfer-controls">
        <label className="erp-field">
          <span>Mã mẫu</span>
          <input className="erp-input" disabled={!gate?.canSubmitSample || sampleBusy} type="text" value={sampleCode} onChange={(event) => setSampleCode(event.target.value)} />
        </label>
        <label className="erp-field">
          <span>Phiên bản công thức</span>
          <input
            className="erp-input"
            disabled={!gate?.canSubmitSample || sampleBusy}
            type="text"
            value={formulaVersion}
            onChange={(event) => setFormulaVersion(event.target.value)}
          />
        </label>
        <label className="erp-field">
          <span>File / mã bằng chứng mẫu</span>
          <input
            className="erp-input"
            disabled={!gate?.canSubmitSample || sampleBusy}
            type="text"
            value={evidenceName}
            onChange={(event) => setEvidenceName(event.target.value)}
          />
        </label>
      </div>
      <label className="erp-field erp-subcontract-note-field">
        <span>Ghi chú mẫu</span>
        <input className="erp-input" disabled={!gate?.canSubmitSample || sampleBusy} type="text" value={note} onChange={(event) => setNote(event.target.value)} />
      </label>

      <div className="erp-subcontract-actions">
        <button className="erp-button erp-button--primary" type="button" disabled={!canSubmit} onClick={onSubmit}>
          Ghi nhận mẫu đã gửi
        </button>
        <StatusChip tone={gate ? sampleGateStatusTone(gate.sampleStatus) : "normal"}>
          {gate ? sampleGateStatusLabel(gate.sampleStatus) : "Chưa có dữ liệu"}
        </StatusChip>
      </div>

      {latestSample ? (
        <div className="erp-production-selected-plan-main">
          <div className="erp-production-selected-plan-badges">
            <StatusChip tone={sampleApprovalStatusTone(latestSample.status)}>{sampleApprovalStatusLabel(latestSample.status)}</StatusChip>
            <StatusChip tone="normal">{latestSample.sampleCode}</StatusChip>
          </div>
          <dl className="erp-production-selected-plan-meta">
            <div>
              <dt>Đã gửi</dt>
              <dd>
                {formatDate(latestSample.submittedAt)}
                {latestSample.submittedBy ? ` / ${latestSample.submittedBy}` : ""}
              </dd>
            </div>
            <div>
              <dt>Công thức</dt>
              <dd>{latestSample.formulaVersion ?? "-"}</dd>
            </div>
            <div>
              <dt>Bằng chứng</dt>
              <dd>{latestSample.evidence.length} dòng</dd>
            </div>
            <div>
              <dt>Lý do quyết định</dt>
              <dd>{latestSample.decisionReason ?? "-"}</dd>
            </div>
          </dl>
        </div>
      ) : null}

      <div className="erp-subcontract-transfer-controls">
        <label className="erp-field">
          <span>Lý do quyết định</span>
          <input
            className="erp-input"
            disabled={!gate?.canApproveSample || sampleBusy}
            type="text"
            value={decisionReason}
            onChange={(event) => setDecisionReason(event.target.value)}
          />
        </label>
        <label className="erp-field">
          <span>Trạng thái lưu mẫu</span>
          <input
            className="erp-input"
            disabled={!gate?.canApproveSample || sampleBusy}
            type="text"
            value={storageStatus}
            onChange={(event) => setStorageStatus(event.target.value)}
          />
        </label>
      </div>
      <div className="erp-subcontract-actions">
        <button className="erp-button erp-button--primary" type="button" disabled={!canApprove} onClick={onApprove}>
          Duyệt mẫu đạt
        </button>
        <button className="erp-button erp-button--secondary" type="button" disabled={!canReject} onClick={onReject}>
          Từ chối mẫu
        </button>
      </div>
    </section>
  );
}

function FactoryMassProductionSection({
  canStart,
  gate,
  massBusy,
  massError,
  massMessage,
  order,
  onStart
}: {
  canStart: boolean;
  gate?: ReturnType<typeof buildSubcontractFactorySampleMassProduction>;
  massBusy: boolean;
  massError?: string;
  massMessage?: string;
  order: SubcontractOrder;
  onStart: () => void;
}) {
  return (
    <section className="erp-masterdata-list-card" id="factory-mass-production">
      <header className="erp-section-header">
        <div>
          <h2 className="erp-section-title">Sản xuất hàng loạt</h2>
          <p className="erp-page-description">
            Chỉ bắt đầu khi vật tư đã bàn giao đủ và mẫu đạt hoặc lệnh không yêu cầu mẫu.
          </p>
        </div>
        {gate ? <StatusChip tone={massGateStatusTone(gate.massStatus)}>{massGateStatusLabel(gate.massStatus)}</StatusChip> : null}
      </header>
      {massError ? (
        <p className="erp-form-error" role="alert">
          {massError}
        </p>
      ) : null}
      {gate?.massBlockedReason ? (
        <p className="erp-form-error" role="status">
          {gate.massBlockedReason}
        </p>
      ) : null}
      {massMessage ? (
        <p className="erp-form-success" role="status">
          {massMessage}
        </p>
      ) : null}

      <dl className="erp-production-selected-plan-meta">
        <div>
          <dt>Lệnh</dt>
          <dd>{order.orderNo}</dd>
        </div>
        <div>
          <dt>Thành phẩm</dt>
          <dd>
            {order.sku} / {formatProductionPlanQuantity(String(order.quantity), order.uomCode ?? "PCS")}
          </dd>
        </div>
        <div>
          <dt>Nhà máy</dt>
          <dd>{order.factoryName}</dd>
        </div>
        <div>
          <dt>Cổng mẫu</dt>
          <dd>{gate ? sampleGateStatusLabel(gate.sampleStatus) : "-"}</dd>
        </div>
      </dl>
      <div className="erp-subcontract-actions">
        <button className="erp-button erp-button--primary" type="button" disabled={!canStart || massBusy} onClick={onStart}>
          Bắt đầu sản xuất hàng loạt
        </button>
        <StatusChip tone={gate ? massGateStatusTone(gate.massStatus) : "normal"}>
          {gate ? massGateStatusLabel(gate.massStatus) : "Chưa có dữ liệu"}
        </StatusChip>
      </div>
    </section>
  );
}

function FactoryFinishedGoodsReceiptSection({
  canReceive,
  deliveryNoteNo,
  draft,
  evidenceName,
  gate,
  latestReceipt,
  locationId,
  note,
  receiptBusy,
  receiptError,
  receiptMessage,
  receivedBy,
  setDeliveryNoteNo,
  setEvidenceName,
  setLocationId,
  setNote,
  setReceivedBy,
  setWarehouseId,
  updateDraft,
  warehouseId,
  onReceive
}: {
  canReceive: boolean;
  deliveryNoteNo: string;
  draft: FactoryFinishedGoodsReceiptDraft;
  evidenceName: string;
  gate?: ReturnType<typeof buildSubcontractFactoryFinishedGoodsReceipt>;
  latestReceipt?: SubcontractFinishedGoodsReceipt;
  locationId: string;
  note: string;
  receiptBusy: boolean;
  receiptError?: string;
  receiptMessage?: string;
  receivedBy: string;
  setDeliveryNoteNo: (value: string) => void;
  setEvidenceName: (value: string) => void;
  setLocationId: (value: string) => void;
  setNote: (value: string) => void;
  setReceivedBy: (value: string) => void;
  setWarehouseId: (value: string) => void;
  updateDraft: (field: keyof FactoryFinishedGoodsReceiptDraft, value: string) => void;
  warehouseId: string;
  onReceive: () => void;
}) {
  const overRemaining = gate?.remainingQty ? !isWithinQuantity(draft.receiveQty, gate.remainingQty) : false;

  return (
    <section className="erp-masterdata-list-card" id="factory-finished-goods-receipt">
      <header className="erp-section-header">
        <div>
          <h2 className="erp-section-title">Nhận thành phẩm từ nhà máy</h2>
          <p className="erp-page-description">
            Nhận thành phẩm vào khu QC hold. Thành phẩm chưa vào tồn khả dụng cho tới khi QC đạt.
          </p>
        </div>
        {gate ? (
          <StatusChip tone={finishedGoodsReceiptGateTone(gate.status)}>
            {finishedGoodsReceiptGateLabel(gate.status)}
          </StatusChip>
        ) : null}
      </header>

      {receiptError ? (
        <p className="erp-form-error" role="alert">
          {receiptError}
        </p>
      ) : null}
      {gate?.blockedReason ? (
        <p className="erp-form-error" role="status">
          {gate.blockedReason}
        </p>
      ) : null}
      {overRemaining ? (
        <p className="erp-form-error" role="status">
          Số lượng nhận không được vượt quá số lượng còn lại.
        </p>
      ) : null}
      {receiptMessage ? (
        <p className="erp-form-success" role="status">
          {receiptMessage}
        </p>
      ) : null}

      <dl className="erp-production-selected-plan-meta">
        <div>
          <dt>Cần nhận</dt>
          <dd>{gate ? formatProductionPlanQuantity(gate.remainingQty, gate.uomCode) : "-"}</dd>
        </div>
        <div>
          <dt>Đã nhận</dt>
          <dd>{gate ? formatProductionPlanQuantity(gate.receivedQty, gate.uomCode) : "-"}</dd>
        </div>
        <div>
          <dt>Tổng kế hoạch</dt>
          <dd>{gate ? formatProductionPlanQuantity(gate.plannedQty, gate.uomCode) : "-"}</dd>
        </div>
        <div>
          <dt>Đích nhận</dt>
          <dd>QC hold</dd>
        </div>
      </dl>

      <div className="erp-subcontract-line-item erp-subcontract-line-item--editable">
        <label className="erp-field">
          <span>Kho nhận</span>
          <select className="erp-input" value={warehouseId} onChange={(event) => setWarehouseId(event.target.value)}>
            {subcontractTransferWarehouseOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Vị trí QC</span>
          <select className="erp-input" value={locationId} onChange={(event) => setLocationId(event.target.value)}>
            {factoryFinishedGoodsLocationOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Số phiếu giao nhà máy</span>
          <input className="erp-input" value={deliveryNoteNo} onChange={(event) => setDeliveryNoteNo(event.target.value)} />
        </label>
        <label className="erp-field">
          <span>Người nhận</span>
          <input className="erp-input" value={receivedBy} onChange={(event) => setReceivedBy(event.target.value)} />
        </label>
      </div>

      <div className="erp-subcontract-line-item erp-subcontract-line-item--editable">
        <label className="erp-field">
          <span>Số lượng nhận</span>
          <input
            className="erp-input"
            min="0"
            max={gate?.remainingQty}
            step="1"
            type="number"
            value={draft.receiveQty}
            onChange={(event) => updateDraft("receiveQty", event.target.value)}
          />
        </label>
        <label className="erp-field">
          <span>Lô / batch</span>
          <input className="erp-input" value={draft.batchNo} onChange={(event) => updateDraft("batchNo", event.target.value)} />
        </label>
        <label className="erp-field">
          <span>Mã lot</span>
          <input className="erp-input" value={draft.lotNo} onChange={(event) => updateDraft("lotNo", event.target.value)} />
        </label>
        <label className="erp-field">
          <span>Hạn dùng</span>
          <input
            className="erp-input"
            type="date"
            value={draft.expiryDate}
            onChange={(event) => updateDraft("expiryDate", event.target.value)}
          />
        </label>
        <label className="erp-field">
          <span>Tình trạng kiện</span>
          <select
            className="erp-input"
            value={draft.packagingStatus}
            onChange={(event) => updateDraft("packagingStatus", event.target.value as SubcontractFinishedGoodsPackagingStatus)}
          >
            <option value="intact">Nguyên kiện</option>
            <option value="damaged">Móp / hư hỏng</option>
            <option value="mixed">Lẫn tình trạng</option>
          </select>
        </label>
      </div>

      <div className="erp-subcontract-line-item erp-subcontract-line-item--editable">
        <label className="erp-field">
          <span>Bằng chứng</span>
          <input className="erp-input" value={evidenceName} onChange={(event) => setEvidenceName(event.target.value)} />
        </label>
        <label className="erp-field">
          <span>Ghi chú dòng</span>
          <input className="erp-input" value={draft.note} onChange={(event) => updateDraft("note", event.target.value)} />
        </label>
        <label className="erp-field">
          <span>Ghi chú phiếu</span>
          <input className="erp-input" value={note} onChange={(event) => setNote(event.target.value)} />
        </label>
      </div>

      <div className="erp-subcontract-actions">
        <button className="erp-button erp-button--primary" type="button" disabled={!canReceive || receiptBusy} onClick={onReceive}>
          Ghi nhận vào QC hold
        </button>
        <StatusChip tone={gate ? finishedGoodsReceiptGateTone(gate.status) : "normal"}>
          {gate ? finishedGoodsReceiptGateLabel(gate.status) : "Chưa có dữ liệu"}
        </StatusChip>
      </div>

      {latestReceipt ? (
        <div className="erp-production-selected-plan-main">
          <span className="erp-production-step-label">Phiếu nhận gần nhất</span>
          <h3>{latestReceipt.receiptNo}</h3>
          <p>
            {latestReceipt.deliveryNoteNo} / {latestReceipt.warehouseCode}-{latestReceipt.locationCode} /{" "}
            {finishedGoodsReceiptStatusLabel(latestReceipt.status)}
          </p>
          <div className="erp-production-selected-plan-badges">
            <StatusChip tone="warning">{finishedGoodsReceiptStatusLabel(latestReceipt.status)}</StatusChip>
            <StatusChip tone="normal">
              {latestReceipt.lines.length} dòng / {latestReceipt.evidence.length} bằng chứng
            </StatusChip>
          </div>
          <dl className="erp-production-selected-plan-meta">
            {latestReceipt.lines.map((line) => (
              <div key={line.id}>
                <dt>{line.skuCode}</dt>
                <dd>
                  {formatProductionPlanQuantity(line.receiveQty, line.uomCode)} / {line.batchNo} /{" "}
                  {finishedGoodsPackagingStatusLabel(line.packagingStatus)}
                </dd>
              </div>
            ))}
          </dl>
        </div>
      ) : null}
    </section>
  );
}

function FactoryFinishedGoodsQcCloseoutSection({
  acceptedBy,
  acceptedQty,
  canAccept,
  canPartialAccept,
  canReject,
  evidenceName,
  evidenceNote,
  gate,
  latestReceipt,
  note,
  openedBy,
  ownerId,
  qcBusy,
  qcError,
  qcMessage,
  reason,
  reasonCode,
  rejectedQty,
  setAcceptedBy,
  setAcceptedQty,
  setEvidenceName,
  setEvidenceNote,
  setNote,
  setOpenedBy,
  setOwnerId,
  setReason,
  setReasonCode,
  setRejectedQty,
  setSeverity,
  severity,
  splitMatchesRemaining,
  splitQty,
  onAccept,
  onPartialAccept,
  onReject
}: {
  acceptedBy: string;
  acceptedQty: string;
  canAccept: boolean;
  canPartialAccept: boolean;
  canReject: boolean;
  evidenceName: string;
  evidenceNote: string;
  gate?: ReturnType<typeof buildSubcontractFactoryFinishedGoodsQcCloseout>;
  latestReceipt?: SubcontractFinishedGoodsReceipt;
  note: string;
  openedBy: string;
  ownerId: string;
  qcBusy: boolean;
  qcError?: string;
  qcMessage?: string;
  reason: string;
  reasonCode: string;
  rejectedQty: string;
  setAcceptedBy: (value: string) => void;
  setAcceptedQty: (value: string) => void;
  setEvidenceName: (value: string) => void;
  setEvidenceNote: (value: string) => void;
  setNote: (value: string) => void;
  setOpenedBy: (value: string) => void;
  setOwnerId: (value: string) => void;
  setReason: (value: string) => void;
  setReasonCode: (value: string) => void;
  setRejectedQty: (value: string) => void;
  setSeverity: (value: SubcontractFactoryClaimSeverity) => void;
  severity: SubcontractFactoryClaimSeverity;
  splitMatchesRemaining: boolean;
  splitQty: string;
  onAccept: () => void;
  onPartialAccept: () => void;
  onReject: () => void;
}) {
  return (
    <section className="erp-masterdata-list-card" id="factory-finished-goods-qc-closeout">
      <header className="erp-section-header">
        <div>
          <h2 className="erp-section-title">QC thành phẩm sau khi nhận từ nhà máy</h2>
          <p className="erp-page-description">
            Chốt QC cho thành phẩm đang ở QC hold. Phần đạt QC mới được chuyển vào tồn khả dụng; phần lỗi phải mở claim nhà máy.
          </p>
        </div>
        {gate ? <StatusChip tone={finishedGoodsQcGateTone(gate.status)}>{finishedGoodsQcGateLabel(gate.status)}</StatusChip> : null}
      </header>

      {qcError ? (
        <p className="erp-form-error" role="alert">
          {qcError}
        </p>
      ) : null}
      {gate?.blockedReason ? (
        <p className="erp-form-error" role="status">
          {gate.blockedReason}
        </p>
      ) : null}
      {!latestReceipt && gate?.canCloseout ? (
        <p className="erp-page-description" role="status">
          Phiếu nhận gần nhất chưa tải trong phiên này. QC đạt toàn bộ vẫn chốt theo lệnh; QC một phần/lỗi toàn bộ cần phiếu nhận để gắn claim.
        </p>
      ) : null}
      {!splitMatchesRemaining && isPositiveQuantity(rejectedQty) ? (
        <p className="erp-form-error" role="status">
          Tổng đạt + lỗi phải bằng số lượng còn trong QC hold.
        </p>
      ) : null}
      {qcMessage ? (
        <p className="erp-form-success" role="status">
          {qcMessage}
        </p>
      ) : null}

      <dl className="erp-production-selected-plan-meta">
        <div>
          <dt>Phiếu nhận</dt>
          <dd>{gate?.receiptNo ?? latestReceipt?.receiptNo ?? "-"}</dd>
        </div>
        <div>
          <dt>Đã nhận</dt>
          <dd>{gate ? formatProductionPlanQuantity(gate.receivedQty, gate.uomCode) : "-"}</dd>
        </div>
        <div>
          <dt>Đã đạt QC</dt>
          <dd>{gate ? formatProductionPlanQuantity(gate.acceptedQty, gate.uomCode) : "-"}</dd>
        </div>
        <div>
          <dt>Đã lỗi</dt>
          <dd>{gate ? formatProductionPlanQuantity(gate.rejectedQty, gate.uomCode) : "-"}</dd>
        </div>
        <div>
          <dt>Còn trong QC hold</dt>
          <dd>{gate ? formatProductionPlanQuantity(gate.remainingQcQty, gate.uomCode) : "-"}</dd>
        </div>
        <div>
          <dt>Split đang nhập</dt>
          <dd>{gate ? formatProductionPlanQuantity(splitQty, gate.uomCode) : "-"}</dd>
        </div>
      </dl>

      <div className="erp-subcontract-line-item erp-subcontract-line-item--editable">
        <label className="erp-field">
          <span>Số đạt QC</span>
          <input
            className="erp-input"
            disabled={!gate?.canCloseout || qcBusy}
            inputMode="decimal"
            type="text"
            value={acceptedQty}
            onChange={(event) => setAcceptedQty(event.target.value)}
          />
        </label>
        <label className="erp-field">
          <span>Số lỗi / claim</span>
          <input
            className="erp-input"
            disabled={!gate?.canCloseout || qcBusy}
            inputMode="decimal"
            type="text"
            value={rejectedQty}
            onChange={(event) => setRejectedQty(event.target.value)}
          />
        </label>
        <label className="erp-field">
          <span>Người QC</span>
          <input className="erp-input" disabled={!gate?.canCloseout || qcBusy} value={acceptedBy} onChange={(event) => setAcceptedBy(event.target.value)} />
        </label>
        <label className="erp-field">
          <span>Ghi chú QC</span>
          <input className="erp-input" disabled={!gate?.canCloseout || qcBusy} value={note} onChange={(event) => setNote(event.target.value)} />
        </label>
      </div>

      <div className="erp-subcontract-line-item erp-subcontract-line-item--editable">
        <label className="erp-field">
          <span>Mã lỗi</span>
          <input className="erp-input" disabled={!gate?.canCloseout || qcBusy} value={reasonCode} onChange={(event) => setReasonCode(event.target.value)} />
        </label>
        <label className="erp-field">
          <span>Mức độ</span>
          <select className="erp-input" disabled={!gate?.canCloseout || qcBusy} value={severity} onChange={(event) => setSeverity(event.target.value as SubcontractFactoryClaimSeverity)}>
            <option value="P1">P1</option>
            <option value="P2">P2</option>
            <option value="P3">P3</option>
          </select>
        </label>
        <label className="erp-field">
          <span>Người mở claim</span>
          <input className="erp-input" disabled={!gate?.canCloseout || qcBusy} value={openedBy} onChange={(event) => setOpenedBy(event.target.value)} />
        </label>
        <label className="erp-field">
          <span>Owner claim</span>
          <input className="erp-input" disabled={!gate?.canCloseout || qcBusy} value={ownerId} onChange={(event) => setOwnerId(event.target.value)} />
        </label>
      </div>

      <div className="erp-subcontract-line-item erp-subcontract-line-item--editable">
        <label className="erp-field">
          <span>Lý do lỗi</span>
          <input className="erp-input" disabled={!gate?.canCloseout || qcBusy} value={reason} onChange={(event) => setReason(event.target.value)} />
        </label>
        <label className="erp-field">
          <span>Bằng chứng lỗi</span>
          <input className="erp-input" disabled={!gate?.canCloseout || qcBusy} value={evidenceName} onChange={(event) => setEvidenceName(event.target.value)} />
        </label>
        <label className="erp-field">
          <span>Ghi chú bằng chứng</span>
          <input className="erp-input" disabled={!gate?.canCloseout || qcBusy} value={evidenceNote} onChange={(event) => setEvidenceNote(event.target.value)} />
        </label>
      </div>

      <div className="erp-subcontract-actions">
        <button className="erp-button erp-button--primary" type="button" disabled={!canAccept} onClick={onAccept}>
          QC đạt toàn bộ
        </button>
        <button className="erp-button erp-button--secondary" type="button" disabled={!canPartialAccept} onClick={onPartialAccept}>
          QC đạt một phần
        </button>
        <button className="erp-button erp-button--danger" type="button" disabled={!canReject} onClick={onReject}>
          QC lỗi toàn bộ
        </button>
        <StatusChip tone={gate ? finishedGoodsQcGateTone(gate.status) : "normal"}>
          {gate ? finishedGoodsQcGateLabel(gate.status) : "Chưa có dữ liệu"}
        </StatusChip>
      </div>
    </section>
  );
}

const materialLineColumns: DataTableColumn<SubcontractOrderMaterialLine>[] = [
  {
    key: "sku",
    header: "Vật tư",
    render: (line) => (
      <span className="erp-masterdata-product-cell">
        <strong>{line.skuCode}</strong>
        <small>{line.itemName}</small>
      </span>
    ),
    width: "260px"
  },
  {
    key: "planned",
    header: "Cần xuất",
    render: (line) => formatProductionPlanQuantity(line.plannedQty, line.uomCode),
    align: "right",
    width: "140px"
  },
  {
    key: "issued",
    header: "Đã xuất",
    render: (line) => formatProductionPlanQuantity(line.issuedQty, line.uomCode),
    align: "right",
    width: "140px"
  },
  {
    key: "qc",
    header: "Kiểm soát",
    render: (line) => (line.lotTraceRequired ? "Lô, QC" : "Không kiểm lô"),
    width: "140px"
  },
  {
    key: "note",
    header: "Ghi chú",
    render: (line) => line.note ?? "-",
    width: "220px"
  }
];

const factoryExecutionColumns: DataTableColumn<FactoryExecutionWorkItem>[] = [
  {
    key: "task",
    header: "Công việc",
    render: (item) => (
      <span className="erp-masterdata-product-cell">
        <strong>{item.title}</strong>
        <small>{item.description}</small>
      </span>
    ),
    width: "380px"
  },
  {
    key: "status",
    header: "Trạng thái",
    render: (item) => <StatusChip tone={item.tone}>{factoryExecutionStatusLabel(item.status)}</StatusChip>,
    width: "140px"
  },
  {
    key: "metric",
    header: "Số liệu",
    render: (item) => item.metric,
    width: "180px"
  },
  {
    key: "action",
    header: "Thao tác",
    render: (item) =>
      item.action ? (
        item.action.disabled ? (
          <button className="erp-button erp-button--secondary erp-button--compact" type="button" disabled>
            {item.action.label}
          </button>
        ) : (
          <Link className="erp-button erp-button--secondary erp-button--compact" href={item.action.href}>
            {item.action.label}
          </Link>
        )
      ) : (
        "-"
      ),
    width: "160px"
  }
];

function FactoryDispatchActions({
  busy,
  dispatch,
  order,
  responseNote,
  setResponseNote,
  onCreate,
  onReady,
  onSent,
  onResponse
}: {
  busy: boolean;
  dispatch?: SubcontractFactoryDispatch;
  order: SubcontractOrder;
  responseNote: string;
  setResponseNote: (value: string) => void;
  onCreate: () => void;
  onReady: (dispatch: SubcontractFactoryDispatch) => void;
  onSent: (dispatch: SubcontractFactoryDispatch) => void;
  onResponse: (dispatch: SubcontractFactoryDispatch, status: "confirmed" | "revision_requested" | "rejected") => void;
}) {
  if (!dispatch) {
    return (
      <button className="erp-button erp-button--primary" type="button" disabled={busy || order.status !== "approved"} onClick={onCreate}>
        Tạo bộ gửi
      </button>
    );
  }
  if (dispatch.status === "draft" || dispatch.status === "revision_requested") {
    return (
      <button className="erp-button erp-button--primary" type="button" disabled={busy} onClick={() => onReady(dispatch)}>
        Đánh dấu sẵn sàng gửi
      </button>
    );
  }
  if (dispatch.status === "ready") {
    return (
      <button className="erp-button erp-button--primary" type="button" disabled={busy} onClick={() => onSent(dispatch)}>
        Đánh dấu đã gửi
      </button>
    );
  }
  if (dispatch.status === "sent") {
    return (
      <div className="erp-form-inline-actions">
        <label className="erp-form-field erp-form-field--inline">
          <span>Ghi chú phản hồi</span>
          <input value={responseNote} onChange={(event) => setResponseNote(event.target.value)} />
        </label>
        <button className="erp-button erp-button--primary" type="button" disabled={busy} onClick={() => onResponse(dispatch, "confirmed")}>
          Xác nhận
        </button>
        <button
          className="erp-button erp-button--secondary"
          type="button"
          disabled={busy || responseNote.trim() === ""}
          onClick={() => onResponse(dispatch, "revision_requested")}
        >
          Cần chỉnh
        </button>
        <button
          className="erp-button erp-button--secondary"
          type="button"
          disabled={busy || responseNote.trim() === ""}
          onClick={() => onResponse(dispatch, "rejected")}
        >
          Từ chối
        </button>
      </div>
    );
  }

  return null;
}

function FactoryDispatchSummary({ dispatch }: { dispatch: SubcontractFactoryDispatch }) {
  return (
    <div className="erp-production-selected-plan-main">
      <div className="erp-production-selected-plan-badges">
        <StatusChip tone={subcontractFactoryDispatchStatusTone(dispatch.status)}>
          {formatSubcontractFactoryDispatchStatus(dispatch.status)}
        </StatusChip>
        <StatusChip tone="normal">{dispatch.dispatchNo}</StatusChip>
      </div>
      <dl className="erp-production-selected-plan-meta">
        <div>
          <dt>Nhà máy</dt>
          <dd>{dispatch.factoryName}</dd>
        </div>
        <div>
          <dt>Thành phẩm</dt>
          <dd>
            {dispatch.sku} / {formatProductionPlanQuantity(dispatch.plannedQty, dispatch.uomCode)}
          </dd>
        </div>
        <div>
          <dt>Đã gửi</dt>
          <dd>
            {dispatch.sentAt ? formatDate(dispatch.sentAt) : "-"}
            {dispatch.sentBy ? ` / ${dispatch.sentBy}` : ""}
          </dd>
        </div>
        <div>
          <dt>Phản hồi</dt>
          <dd>
            {dispatch.respondedAt ? formatDate(dispatch.respondedAt) : "-"}
            {dispatch.responseBy ? ` / ${dispatch.responseBy}` : ""}
          </dd>
        </div>
        <div>
          <dt>Sẵn sàng</dt>
          <dd>{dispatch.readyAt ? formatDate(dispatch.readyAt) : "-"}</dd>
        </div>
        <div>
          <dt>Bằng chứng gửi</dt>
          <dd>{dispatch.evidence.length} dòng</dd>
        </div>
      </dl>
      {dispatch.factoryResponseNote ? <p className="erp-page-description">{dispatch.factoryResponseNote}</p> : null}
      {dispatch.evidence.length > 0 ? (
        <div className="erp-page-description">
          <strong>Bằng chứng:</strong>{" "}
          {dispatch.evidence
            .map((evidence) => evidence.fileName || evidence.objectKey || evidence.externalURL || evidence.note)
            .filter(Boolean)
            .join("; ")}
        </div>
      ) : null}
      <DataTable
        columns={factoryDispatchLineColumns}
        rows={dispatch.lines}
        getRowKey={(line) => line.id}
        pagination
        preserveColumnWidths
        emptyState={<EmptyState title="Bộ gửi chưa có dòng vật tư" />}
      />
    </div>
  );
}

const factoryDispatchLineColumns: DataTableColumn<SubcontractFactoryDispatch["lines"][number]>[] = [
  {
    key: "sku",
    header: "Vật tư",
    render: (line) => (
      <span className="erp-masterdata-product-cell">
        <strong>{line.skuCode}</strong>
        <small>{line.itemName}</small>
      </span>
    ),
    width: "260px"
  },
  {
    key: "planned",
    header: "Cần gửi",
    render: (line) => formatProductionPlanQuantity(line.plannedQty, line.uomCode),
    align: "right",
    width: "140px"
  },
  {
    key: "qc",
    header: "Kiểm soát",
    render: (line) => (line.lotTraceRequired ? "Lô, QC" : "Không kiểm lô"),
    width: "140px"
  },
  {
    key: "note",
    header: "Ghi chú",
    render: (line) => line.note ?? "-",
    width: "220px"
  }
];

function TimelineItem({ item }: { item: SubcontractOrderTimelineItem }) {
  return (
    <li className="erp-document-timeline-step" data-status={item.status}>
      <span className="erp-document-timeline-marker" aria-hidden="true" />
      <div className="erp-document-timeline-content">
        <div className="erp-document-timeline-heading">
          <strong>{item.label}</strong>
          <StatusChip tone={item.tone}>{timelineStatusLabel(item.status)}</StatusChip>
        </div>
        <p>{item.description}</p>
        {item.action ? (
          item.action.disabled ? (
            <button className="erp-button erp-button--secondary erp-button--compact erp-document-timeline-action" type="button" disabled>
              {item.action.label}
            </button>
          ) : (
            <Link className="erp-button erp-button--secondary erp-button--compact erp-document-timeline-action" href={item.action.href}>
              {item.action.label}
            </Link>
          )
        ) : null}
        {item.occurredAt ? <small>{formatDate(item.occurredAt)}</small> : null}
      </div>
    </li>
  );
}

function timelineStatusLabel(status: SubcontractOrderTimelineItem["status"]) {
  switch (status) {
    case "complete":
      return "Đã xong";
    case "current":
      return "Đang xử lý";
    case "blocked":
      return "Dừng";
    case "pending":
    default:
      return "Chờ";
  }
}

function factoryExecutionStatusLabel(status: FactoryExecutionWorkItem["status"]) {
  switch (status) {
    case "complete":
      return "Đã xong";
    case "current":
      return "Đang xử lý";
    case "blocked":
      return "Đang chặn";
    case "pending":
    default:
      return "Chờ";
  }
}

function materialHandoverStatusTone(status: FactoryMaterialHandover["status"]) {
  switch (status) {
    case "complete":
      return "success" as const;
    case "ready":
      return "info" as const;
    case "blocked":
    default:
      return "warning" as const;
  }
}

function materialHandoverStatusLabel(status: FactoryMaterialHandover["status"]) {
  switch (status) {
    case "complete":
      return "Đã bàn giao đủ";
    case "ready":
      return "Sẵn sàng bàn giao";
    case "blocked":
    default:
      return "Chưa thể bàn giao";
  }
}

function materialLineStatusLabel(status: FactoryMaterialHandover["lines"][number]["status"], lotTraceRequired: boolean) {
  if (status === "complete") {
    return "Đã đủ";
  }

  return lotTraceRequired ? "Cần lô" : "Sẵn sàng";
}

function factoryMaterialTransferTone(status: SubcontractMaterialTransfer["status"]) {
  switch (status) {
    case "SENT":
      return "success" as const;
    case "READY_TO_SEND":
      return "info" as const;
    case "DRAFT":
    default:
      return "warning" as const;
  }
}

function factoryMaterialTransferStatusLabel(status: SubcontractMaterialTransfer["status"]) {
  switch (status) {
    case "SENT":
      return "Đã bàn giao";
    case "READY_TO_SEND":
      return "Sẵn sàng gửi";
    case "DRAFT":
    default:
      return "Nháp";
  }
}

function sampleGateStatusTone(status: FactorySampleStatus) {
  switch (status) {
    case "approved":
    case "not_required":
      return "success" as const;
    case "ready_to_submit":
    case "submitted":
      return "info" as const;
    case "rejected":
    case "blocked":
      return "danger" as const;
    case "pending":
    default:
      return "warning" as const;
  }
}

function sampleGateStatusLabel(status: FactorySampleStatus) {
  switch (status) {
    case "approved":
      return "Mẫu đạt";
    case "not_required":
      return "Không yêu cầu mẫu";
    case "ready_to_submit":
      return "Sẵn sàng gửi mẫu";
    case "submitted":
      return "Chờ duyệt mẫu";
    case "rejected":
      return "Mẫu không đạt";
    case "blocked":
      return "Đang chặn";
    case "pending":
    default:
      return "Chờ vật tư";
  }
}

function massGateStatusTone(status: FactoryMassProductionStatus) {
  switch (status) {
    case "started":
      return "success" as const;
    case "ready_to_start":
      return "info" as const;
    case "blocked":
      return "danger" as const;
    case "pending":
    default:
      return "warning" as const;
  }
}

function massGateStatusLabel(status: FactoryMassProductionStatus) {
  switch (status) {
    case "started":
      return "Đã bắt đầu";
    case "ready_to_start":
      return "Sẵn sàng bắt đầu";
    case "blocked":
      return "Đang chặn";
    case "pending":
    default:
      return "Chờ đủ điều kiện";
  }
}

function sampleApprovalStatusTone(status: SubcontractSampleApproval["status"]) {
  switch (status) {
    case "approved":
      return "success" as const;
    case "rejected":
      return "danger" as const;
    case "submitted":
    default:
      return "info" as const;
  }
}

function sampleApprovalStatusLabel(status: SubcontractSampleApproval["status"]) {
  switch (status) {
    case "approved":
      return "Mẫu đạt";
    case "rejected":
      return "Mẫu không đạt";
    case "submitted":
    default:
      return "Đã gửi mẫu";
  }
}

function finishedGoodsReceiptGateTone(status: ReturnType<typeof buildSubcontractFactoryFinishedGoodsReceipt>["status"]) {
  switch (status) {
    case "complete":
      return "success" as const;
    case "ready_to_receive":
    case "partial":
      return "info" as const;
    case "blocked":
    default:
      return "warning" as const;
  }
}

function finishedGoodsReceiptGateLabel(status: ReturnType<typeof buildSubcontractFactoryFinishedGoodsReceipt>["status"]) {
  switch (status) {
    case "complete":
      return "Đã nhận đủ";
    case "partial":
      return "Đã nhận một phần";
    case "ready_to_receive":
      return "Sẵn sàng nhận";
    case "blocked":
    default:
      return "Chưa thể nhận";
  }
}

function finishedGoodsReceiptStatusLabel(status: SubcontractFinishedGoodsReceipt["status"]) {
  switch (status) {
    case "qc_hold":
    default:
      return "QC hold";
  }
}

function finishedGoodsQcGateTone(status: ReturnType<typeof buildSubcontractFactoryFinishedGoodsQcCloseout>["status"]) {
  switch (status) {
    case "passed":
      return "success" as const;
    case "ready_for_qc":
    case "partially_closed":
      return "info" as const;
    case "failed":
      return "danger" as const;
    case "blocked":
    default:
      return "warning" as const;
  }
}

function finishedGoodsQcGateLabel(status: ReturnType<typeof buildSubcontractFactoryFinishedGoodsQcCloseout>["status"]) {
  switch (status) {
    case "passed":
      return "QC đạt";
    case "failed":
      return "Claim nhà máy";
    case "partially_closed":
      return "Đã QC một phần";
    case "ready_for_qc":
      return "Sẵn sàng QC";
    case "blocked":
    default:
      return "Chưa thể QC";
  }
}

function finishedGoodsPackagingStatusLabel(status?: string) {
  switch (status) {
    case "intact":
      return "Nguyên kiện";
    case "damaged":
      return "Móp / hư hỏng";
    case "mixed":
      return "Lẫn tình trạng";
    default:
      return "-";
  }
}

function closeoutLabel(order: SubcontractOrder) {
  if (order.status === "closed") {
    return "Đã đóng";
  }
  if (["accepted", "final_payment_ready"].includes(order.status)) {
    return "QC đạt";
  }
  if (order.status === "rejected_with_factory_issue") {
    return "Claim nhà máy";
  }
  if (["finished_goods_received", "qc_in_progress"].includes(order.status)) {
    return "Chờ QC";
  }

  return "Đang xử lý";
}

function closeoutTone(order: SubcontractOrder) {
  if (order.status === "closed" || ["accepted", "final_payment_ready"].includes(order.status)) {
    return "success" as const;
  }
  if (order.status === "rejected_with_factory_issue") {
    return "danger" as const;
  }
  if (["finished_goods_received", "qc_in_progress"].includes(order.status)) {
    return "warning" as const;
  }

  return "info" as const;
}

function formatFinalPaymentStatus(status: SubcontractFinalPaymentStatus) {
  switch (status) {
    case "released":
      return "Đã thanh toán";
    case "hold":
      return "Đang giữ";
    case "pending":
    default:
      return "Chờ xử lý";
  }
}

function compactQuantity(value: string) {
  const numericValue = Number.parseFloat(value.replace(",", "."));
  if (!Number.isFinite(numericValue)) {
    return "0";
  }

  return String(numericValue);
}

function isPositiveQuantity(value: string) {
  const numericValue = Number.parseFloat(value.replace(",", "."));
  return Number.isFinite(numericValue) && numericValue > 0;
}

function isWithinQuantity(value: string, maxValue?: string) {
  if (!maxValue) {
    return true;
  }
  const numericValue = Number.parseFloat(value.replace(",", "."));
  const numericMax = Number.parseFloat(maxValue.replace(",", "."));

  return Number.isFinite(numericValue) && Number.isFinite(numericMax) && numericValue <= numericMax + 0.000001;
}

function quantitySum(left: string, right: string) {
  return (quantityNumber(left) + quantityNumber(right)).toFixed(6);
}

function isQuantityEqual(left: string, right: string) {
  return Math.abs(quantityNumber(left) - quantityNumber(right)) <= 0.000001;
}

function quantityNumber(value?: string) {
  if (!value) {
    return 0;
  }

  const parsed = Number.parseFloat(value.replace(",", "."));
  return Number.isFinite(parsed) ? parsed : 0;
}

function formatDate(value?: string) {
  if (!value) {
    return "-";
  }

  return new Intl.DateTimeFormat("vi-VN", { day: "2-digit", month: "2-digit", year: "numeric" }).format(new Date(value));
}

function errorText(error: unknown) {
  return error instanceof Error ? error.message : "Request failed";
}
