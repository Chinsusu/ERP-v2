"use client";

import { useEffect, useMemo, useState, type FormEvent } from "react";
import { DataTable, StatusChip, type DataTableColumn, type StatusTone } from "@/shared/design-system/components";
import { decimalScales, normalizeDecimalInput } from "@/shared/format/numberFormat";
import { t } from "@/shared/i18n";
import { useAvailableStock } from "../hooks/useAvailableStock";
import {
  availabilityTone,
  createStockCount,
  createBatchQCTransition,
  formatQuantity,
  getBatchQCTransitions,
  getStockCounts,
  submitStockCount
} from "../services/stockAvailabilityService";
import {
  getStockAdjustments,
  summarizeStockAdjustmentDelta,
  transitionStockAdjustment
} from "../services/stockAdjustmentService";
import {
  createStockTransfer,
  createWarehouseIssue,
  getStockTransfers,
  getWarehouseIssues,
  transitionStockTransfer,
  transitionWarehouseIssue
} from "../services/warehouseDocumentService";
import type {
  AvailableStockItem,
  AvailableStockQuery,
  BatchQCStatus,
  BatchQCTransition,
  StockAdjustment,
  StockAdjustmentAction,
  StockAdjustmentStatus,
  StockTransfer,
  StockTransferStatus,
  StockCountSession,
  StockCountStatus,
  WarehouseDocumentAction,
  WarehouseIssue,
  WarehouseIssueStatus
} from "../types";

const warehouseOptions = [
  { label: inventoryCopy("filters.allWarehouses"), value: "" },
  { label: "HCM", value: "wh-hcm" },
  { label: "HN", value: "wh-hn" }
];

const qcStatusOptions: { label: string; value: BatchQCStatus }[] = [
  { label: inventoryCopy("qc.status.pass"), value: "pass" },
  { label: inventoryCopy("qc.status.fail"), value: "fail" },
  { label: inventoryCopy("qc.status.quarantine"), value: "quarantine" },
  { label: inventoryCopy("qc.status.retest_required"), value: "retest_required" }
];

const columns: DataTableColumn<AvailableStockItem>[] = [
  {
    key: "warehouse",
    header: inventoryCopy("availableStock.columns.warehouse"),
    render: (row) => row.warehouseCode,
    width: "110px"
  },
  {
    key: "location",
    header: inventoryCopy("availableStock.columns.location"),
    render: (row) => row.locationCode ?? "-",
    width: "110px"
  },
  {
    key: "sku",
    header: inventoryCopy("availableStock.columns.sku"),
    render: (row) => row.sku,
    width: "180px"
  },
  {
    key: "batch",
    header: inventoryCopy("availableStock.columns.batch"),
    render: (row) => row.batchNo ?? "-",
    width: "150px"
  },
  {
    key: "physical",
    header: inventoryCopy("availableStock.columns.physical"),
    render: (row) => formatQuantity(row.physicalQty, row.baseUomCode),
    align: "right",
    width: "130px"
  },
  {
    key: "reserved",
    header: inventoryCopy("availableStock.columns.reserved"),
    render: (row) => formatQuantity(row.reservedQty, row.baseUomCode),
    align: "right",
    width: "130px"
  },
  {
    key: "qcHold",
    header: inventoryCopy("availableStock.columns.qcHold"),
    render: (row) => formatQuantity(row.qcHoldQty, row.baseUomCode),
    align: "right",
    width: "130px"
  },
  {
    key: "blocked",
    header: inventoryCopy("availableStock.columns.blocked"),
    render: (row) => formatQuantity(row.blockedQty, row.baseUomCode),
    align: "right",
    width: "130px"
  },
  {
    key: "available",
    header: inventoryCopy("availableStock.columns.available"),
    render: (row) => formatQuantity(row.availableQty, row.baseUomCode),
    align: "right",
    width: "130px"
  },
  {
    key: "state",
    header: inventoryCopy("availableStock.columns.state"),
    render: (row) => <StatusChip tone={availabilityTone(row)}>{statusLabel(row)}</StatusChip>,
    width: "130px"
  }
];

const transitionColumns: DataTableColumn<BatchQCTransition>[] = [
  {
    key: "createdAt",
    header: inventoryCopy("qc.columns.time"),
    render: (row) => formatDateTime(row.createdAt),
    width: "180px"
  },
  {
    key: "status",
    header: inventoryCopy("qc.columns.status"),
    render: (row) => (
      <span className="erp-stock-qc-status-flow">
        <StatusChip tone={qcStatusTone(row.fromQcStatus)}>{qcStatusLabel(row.fromQcStatus)}</StatusChip>
        <span>{inventoryCopy("qc.to")}</span>
        <StatusChip tone={qcStatusTone(row.toQcStatus)}>{qcStatusLabel(row.toQcStatus)}</StatusChip>
      </span>
    ),
    width: "230px"
  },
  {
    key: "actor",
    header: inventoryCopy("qc.columns.actor"),
    render: (row) => row.actorId,
    width: "150px"
  },
  {
    key: "businessRef",
    header: inventoryCopy("qc.columns.ref"),
    render: (row) => row.businessRef || "-",
    width: "150px"
  },
  {
    key: "reason",
    header: inventoryCopy("qc.columns.reason"),
    render: (row) => row.reason
  },
  {
    key: "audit",
    header: inventoryCopy("qc.columns.audit"),
    render: (row) => row.auditLogId,
    width: "190px"
  }
];

const stockCountColumns: DataTableColumn<StockCountSession>[] = [
  {
    key: "countNo",
    header: inventoryCopy("stockCount.columns.count"),
    render: (row) => row.countNo,
    width: "150px"
  },
  {
    key: "warehouse",
    header: inventoryCopy("stockCount.columns.warehouse"),
    render: (row) => row.warehouseCode || row.warehouseId,
    width: "110px"
  },
  {
    key: "status",
    header: inventoryCopy("stockCount.columns.status"),
    render: (row) => <StatusChip tone={stockCountStatusTone(row.status)}>{stockCountStatusLabel(row.status)}</StatusChip>,
    width: "150px"
  },
  {
    key: "line",
    header: inventoryCopy("stockCount.columns.line"),
    render: (row) => stockCountLineLabel(row)
  },
  {
    key: "expected",
    header: inventoryCopy("stockCount.columns.expected"),
    render: (row) => stockCountLineQuantity(row, "expectedQty"),
    align: "right",
    width: "130px"
  },
  {
    key: "counted",
    header: inventoryCopy("stockCount.columns.counted"),
    render: (row) => stockCountLineQuantity(row, "countedQty"),
    align: "right",
    width: "130px"
  },
  {
    key: "delta",
    header: inventoryCopy("stockCount.columns.delta"),
    render: (row) => stockCountLineQuantity(row, "deltaQty"),
    align: "right",
    width: "130px"
  },
  {
    key: "updated",
    header: inventoryCopy("stockCount.columns.updated"),
    render: (row) => formatDateTime(row.updatedAt),
    width: "170px"
  }
];

const stockAdjustmentColumns: DataTableColumn<StockAdjustment>[] = [
  {
    key: "adjustmentNo",
    header: inventoryCopy("stockAdjustment.columns.adjustment"),
    render: (row) => row.adjustmentNo,
    width: "160px"
  },
  {
    key: "warehouse",
    header: inventoryCopy("stockAdjustment.columns.warehouse"),
    render: (row) => row.warehouseCode || row.warehouseId,
    width: "110px"
  },
  {
    key: "status",
    header: inventoryCopy("stockAdjustment.columns.status"),
    render: (row) => (
      <StatusChip tone={stockAdjustmentStatusTone(row.status)}>{stockAdjustmentStatusLabel(row.status)}</StatusChip>
    ),
    width: "140px"
  },
  {
    key: "beforeAfter",
    header: inventoryCopy("stockAdjustment.columns.beforeAfter"),
    render: (row) => stockAdjustmentBeforeAfter(row),
    width: "210px"
  },
  {
    key: "delta",
    header: inventoryCopy("stockAdjustment.columns.delta"),
    render: (row) => stockAdjustmentDelta(row),
    align: "right",
    width: "130px"
  },
  {
    key: "reason",
    header: inventoryCopy("stockAdjustment.columns.reason"),
    render: (row) => row.reason
  },
  {
    key: "requestedBy",
    header: inventoryCopy("stockAdjustment.columns.requestedBy"),
    render: (row) => row.requestedBy,
    width: "150px"
  },
  {
    key: "audit",
    header: inventoryCopy("stockAdjustment.columns.audit"),
    render: (row) => row.auditLogId || "-",
    width: "210px"
  }
];

const stockTransferColumns: DataTableColumn<StockTransfer>[] = [
  {
    key: "transferNo",
    header: inventoryCopy("stockTransfer.columns.transfer"),
    render: (row) => row.transferNo,
    width: "160px"
  },
  {
    key: "route",
    header: inventoryCopy("stockTransfer.columns.route"),
    render: (row) => `${row.sourceWarehouseCode || row.sourceWarehouseId} -> ${row.destinationWarehouseCode || row.destinationWarehouseId}`,
    width: "170px"
  },
  {
    key: "status",
    header: inventoryCopy("stockTransfer.columns.status"),
    render: (row) => <StatusChip tone={warehouseDocumentStatusTone(row.status)}>{stockTransferStatusLabel(row.status)}</StatusChip>,
    width: "140px"
  },
  {
    key: "line",
    header: inventoryCopy("stockTransfer.columns.line"),
    render: (row) => stockTransferLineLabel(row)
  },
  {
    key: "quantity",
    header: inventoryCopy("stockTransfer.columns.quantity"),
    render: (row) => stockTransferQuantity(row),
    align: "right",
    width: "130px"
  },
  {
    key: "updated",
    header: inventoryCopy("stockTransfer.columns.updated"),
    render: (row) => formatDateTime(row.updatedAt),
    width: "170px"
  }
];

const warehouseIssueColumns: DataTableColumn<WarehouseIssue>[] = [
  {
    key: "issueNo",
    header: inventoryCopy("warehouseIssue.columns.issue"),
    render: (row) => row.issueNo,
    width: "160px"
  },
  {
    key: "destination",
    header: inventoryCopy("warehouseIssue.columns.destination"),
    render: (row) => row.destinationName,
    width: "180px"
  },
  {
    key: "status",
    header: inventoryCopy("warehouseIssue.columns.status"),
    render: (row) => <StatusChip tone={warehouseDocumentStatusTone(row.status)}>{warehouseIssueStatusLabel(row.status)}</StatusChip>,
    width: "140px"
  },
  {
    key: "line",
    header: inventoryCopy("warehouseIssue.columns.line"),
    render: (row) => warehouseIssueLineLabel(row)
  },
  {
    key: "quantity",
    header: inventoryCopy("warehouseIssue.columns.quantity"),
    render: (row) => warehouseIssueQuantity(row),
    align: "right",
    width: "130px"
  },
  {
    key: "updated",
    header: inventoryCopy("warehouseIssue.columns.updated"),
    render: (row) => formatDateTime(row.updatedAt),
    width: "170px"
  }
];

export function AvailableStockPrototype() {
  const [warehouseId, setWarehouseId] = useState("");
  const [locationId, setLocationId] = useState("");
  const [sku, setSKU] = useState("");
  const [selectedBatchId, setSelectedBatchId] = useState("");
  const [nextQCStatus, setNextQCStatus] = useState<BatchQCStatus>("pass");
  const [transitionReason, setTransitionReason] = useState("");
  const [businessRef, setBusinessRef] = useState("");
  const [transitions, setTransitions] = useState<BatchQCTransition[]>([]);
  const [historyLoading, setHistoryLoading] = useState(false);
  const [transitionSubmitting, setTransitionSubmitting] = useState(false);
  const [transitionMessage, setTransitionMessage] = useState("");
  const [transitionMessageTone, setTransitionMessageTone] = useState<StatusTone>("info");
  const [stockCounts, setStockCounts] = useState<StockCountSession[]>([]);
  const [stockCountsLoading, setStockCountsLoading] = useState(false);
  const [selectedStockKey, setSelectedStockKey] = useState("");
  const [selectedStockCountId, setSelectedStockCountId] = useState("");
  const [countedQty, setCountedQty] = useState("");
  const [stockCountSubmitting, setStockCountSubmitting] = useState(false);
  const [stockCountMessage, setStockCountMessage] = useState("");
  const [stockCountMessageTone, setStockCountMessageTone] = useState<StatusTone>("info");
  const [stockAdjustments, setStockAdjustments] = useState<StockAdjustment[]>([]);
  const [stockAdjustmentsLoading, setStockAdjustmentsLoading] = useState(false);
  const [selectedAdjustmentId, setSelectedAdjustmentId] = useState("");
  const [adjustmentSubmitting, setAdjustmentSubmitting] = useState(false);
  const [adjustmentMessage, setAdjustmentMessage] = useState("");
  const [adjustmentMessageTone, setAdjustmentMessageTone] = useState<StatusTone>("info");
  const [stockTransfers, setStockTransfers] = useState<StockTransfer[]>([]);
  const [stockTransfersLoading, setStockTransfersLoading] = useState(false);
  const [selectedTransferId, setSelectedTransferId] = useState("");
  const [transferDestinationWarehouseId, setTransferDestinationWarehouseId] = useState("wh-hn");
  const [transferDestinationWarehouseCode, setTransferDestinationWarehouseCode] = useState("HN");
  const [transferQty, setTransferQty] = useState("");
  const [transferSubmitting, setTransferSubmitting] = useState(false);
  const [transferMessage, setTransferMessage] = useState("");
  const [transferMessageTone, setTransferMessageTone] = useState<StatusTone>("info");
  const [warehouseIssues, setWarehouseIssues] = useState<WarehouseIssue[]>([]);
  const [warehouseIssuesLoading, setWarehouseIssuesLoading] = useState(false);
  const [selectedIssueId, setSelectedIssueId] = useState("");
  const [issueDestinationName, setIssueDestinationName] = useState(inventoryCopy("warehouseIssue.defaultDestination"));
  const [issueQty, setIssueQty] = useState("");
  const [issueSubmitting, setIssueSubmitting] = useState(false);
  const [issueMessage, setIssueMessage] = useState("");
  const [issueMessageTone, setIssueMessageTone] = useState<StatusTone>("info");
  const query = useMemo<AvailableStockQuery>(
    () => ({
      warehouseId: warehouseId || undefined,
      locationId: locationId || undefined,
      sku: sku || undefined
    }),
    [locationId, warehouseId, sku]
  );
  const { items, loading, summary } = useAvailableStock(query);
  const batchOptions = useMemo(() => uniqueBatchOptions(items), [items]);
  const stockOptions = useMemo(() => items.filter((item) => item.physicalQty !== "0.000000"), [items]);
  const selectedStock = useMemo(
    () => stockOptions.find((item) => stockRowKey(item) === selectedStockKey),
    [selectedStockKey, stockOptions]
  );
  const openStockCounts = useMemo(() => stockCounts.filter((count) => count.status === "open"), [stockCounts]);
  const activeOpenStockCount = useMemo(
    () => openStockCounts.find((count) => count.id === selectedStockCountId) ?? openStockCounts[0],
    [openStockCounts, selectedStockCountId]
  );
  const actionableAdjustment = useMemo(() => firstActionableAdjustment(stockAdjustments), [stockAdjustments]);
  const selectedAdjustment = useMemo(
    () =>
      stockAdjustments.find((adjustment) => adjustment.id === selectedAdjustmentId) ??
      actionableAdjustment ??
      stockAdjustments[0],
    [actionableAdjustment, selectedAdjustmentId, stockAdjustments]
  );
  const actionableTransfer = useMemo(() => firstActionableWarehouseDocument(stockTransfers), [stockTransfers]);
  const selectedTransfer = useMemo(
    () => stockTransfers.find((transfer) => transfer.id === selectedTransferId) ?? actionableTransfer ?? stockTransfers[0],
    [actionableTransfer, selectedTransferId, stockTransfers]
  );
  const actionableIssue = useMemo(() => firstActionableWarehouseDocument(warehouseIssues), [warehouseIssues]);
  const selectedIssue = useMemo(
    () => warehouseIssues.find((issue) => issue.id === selectedIssueId) ?? actionableIssue ?? warehouseIssues[0],
    [actionableIssue, selectedIssueId, warehouseIssues]
  );
  const selectedBatch = useMemo(
    () => batchOptions.find((batch) => batch.id === selectedBatchId),
    [batchOptions, selectedBatchId]
  );
  const selectedBatchQCStatus = transitions[0]?.toQcStatus ?? selectedBatch?.qcStatus;

  useEffect(() => {
    if (selectedBatchId && batchOptions.some((batch) => batch.id === selectedBatchId)) {
      return;
    }
    setSelectedBatchId(batchOptions[0]?.id ?? "");
  }, [batchOptions, selectedBatchId]);

  useEffect(() => {
    if (selectedStockKey && stockOptions.some((item) => stockRowKey(item) === selectedStockKey)) {
      return;
    }
    setSelectedStockKey(stockOptions[0] ? stockRowKey(stockOptions[0]) : "");
  }, [selectedStockKey, stockOptions]);

  useEffect(() => {
    setCountedQty(selectedStock?.physicalQty ?? "");
  }, [selectedStock?.physicalQty, selectedStockKey]);

  useEffect(() => {
    const defaultQty = selectedStock?.availableQty !== "0.000000" ? selectedStock?.availableQty : selectedStock?.physicalQty;
    setTransferQty(defaultQty ?? "");
    setIssueQty(defaultQty ?? "");
  }, [selectedStock?.availableQty, selectedStock?.physicalQty, selectedStockKey]);

  useEffect(() => {
    if (!selectedStock || transferDestinationWarehouseId !== selectedStock.warehouseId) {
      return;
    }
    const nextDestination = warehouseOptions.find((option) => option.value && option.value !== selectedStock.warehouseId);
    if (!nextDestination) {
      return;
    }
    setTransferDestinationWarehouseId(nextDestination.value);
    setTransferDestinationWarehouseCode(nextDestination.label);
  }, [selectedStock, transferDestinationWarehouseId]);

  useEffect(() => {
    let active = true;
    if (!selectedBatchId) {
      setTransitions([]);
      return;
    }

    setHistoryLoading(true);
    getBatchQCTransitions(selectedBatchId)
      .then((items) => {
        if (active) {
          setTransitions(items);
        }
      })
      .finally(() => {
        if (active) {
          setHistoryLoading(false);
        }
      });

    return () => {
      active = false;
    };
  }, [selectedBatchId]);

  useEffect(() => {
    let active = true;
    setStockCountsLoading(true);
    getStockCounts()
      .then((rows) => {
        if (active) {
          setStockCounts(rows);
          setSelectedStockCountId((current) => current || rows.find((row) => row.status === "open")?.id || "");
        }
      })
      .finally(() => {
        if (active) {
          setStockCountsLoading(false);
        }
      });

    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    let active = true;
    setStockAdjustmentsLoading(true);
    getStockAdjustments()
      .then((rows) => {
        if (active) {
          setStockAdjustments(rows);
          setSelectedAdjustmentId((current) => current || firstActionableAdjustment(rows)?.id || rows[0]?.id || "");
        }
      })
      .finally(() => {
        if (active) {
          setStockAdjustmentsLoading(false);
        }
      });

    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    let active = true;
    setStockTransfersLoading(true);
    getStockTransfers()
      .then((rows) => {
        if (active) {
          setStockTransfers(rows);
          setSelectedTransferId((current) => current || firstActionableWarehouseDocument(rows)?.id || rows[0]?.id || "");
        }
      })
      .finally(() => {
        if (active) {
          setStockTransfersLoading(false);
        }
      });

    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    let active = true;
    setWarehouseIssuesLoading(true);
    getWarehouseIssues()
      .then((rows) => {
        if (active) {
          setWarehouseIssues(rows);
          setSelectedIssueId((current) => current || firstActionableWarehouseDocument(rows)?.id || rows[0]?.id || "");
        }
      })
      .finally(() => {
        if (active) {
          setWarehouseIssuesLoading(false);
        }
      });

    return () => {
      active = false;
    };
  }, []);

  async function handleTransitionSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedBatchId || transitionReason.trim() === "") {
      return;
    }

    setTransitionSubmitting(true);
    setTransitionMessage("");
    try {
      await createBatchQCTransition(selectedBatchId, {
        qcStatus: nextQCStatus,
        reason: transitionReason,
        businessRef
      });
      setTransitionReason("");
      setBusinessRef("");
      setTransitions(await getBatchQCTransitions(selectedBatchId));
      setTransitionMessageTone("success");
      setTransitionMessage(inventoryCopy("qc.messages.recorded"));
    } catch {
      setTransitionMessageTone("danger");
      setTransitionMessage(inventoryCopy("qc.messages.recordError"));
    } finally {
      setTransitionSubmitting(false);
    }
  }

  async function handleCreateStockCount() {
    if (!selectedStock) {
      return;
    }

    setStockCountSubmitting(true);
    setStockCountMessage("");
    try {
      const created = await createStockCount({
        warehouseId: selectedStock.warehouseId,
        warehouseCode: selectedStock.warehouseCode,
        scope: "cycle-count",
        lines: [
          {
            sku: selectedStock.sku,
            batchId: selectedStock.batchId,
            batchNo: selectedStock.batchNo,
            locationId: selectedStock.locationId,
            locationCode: selectedStock.locationCode,
            expectedQty: selectedStock.physicalQty,
            baseUomCode: selectedStock.baseUomCode
          }
        ]
      });
      setSelectedStockCountId(created.id);
      setCountedQty(selectedStock.physicalQty);
      setStockCounts(await getStockCounts());
      setStockCountMessageTone("success");
      setStockCountMessage(inventoryCopy("stockCount.messages.opened"));
    } catch {
      setStockCountMessageTone("danger");
      setStockCountMessage(inventoryCopy("stockCount.messages.openError"));
    } finally {
      setStockCountSubmitting(false);
    }
  }

  async function handleSubmitStockCount(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!activeOpenStockCount) {
      return;
    }

    let normalizedCountedQty: string;
    try {
      normalizedCountedQty = normalizeDecimalInput(countedQty, decimalScales.quantity);
    } catch {
      setStockCountMessageTone("danger");
      setStockCountMessage(inventoryCopy("stockCount.messages.invalidCountedQty"));
      return;
    }

    setStockCountSubmitting(true);
    setStockCountMessage("");
    try {
      const submitted = await submitStockCount(activeOpenStockCount.id, {
        lines: activeOpenStockCount.lines.map((line, index) => ({
          id: line.id,
          countedQty: index === 0 ? normalizedCountedQty : line.expectedQty,
          note: "cycle count"
        }))
      });
      setSelectedStockCountId(submitted.id);
      setStockCounts(await getStockCounts());
      setStockCountMessageTone("success");
      setStockCountMessage(
        submitted.status === "variance_review"
          ? inventoryCopy("stockCount.messages.varianceReview")
          : inventoryCopy("stockCount.messages.submitted")
      );
    } catch {
      setStockCountMessageTone("danger");
      setStockCountMessage(inventoryCopy("stockCount.messages.submitError"));
    } finally {
      setStockCountSubmitting(false);
    }
  }

  async function handleAdjustmentAction(action: StockAdjustmentAction) {
    if (!selectedAdjustment) {
      return;
    }

    setAdjustmentSubmitting(true);
    setAdjustmentMessage("");
    try {
      const updated = await transitionStockAdjustment(selectedAdjustment.id, action);
      setSelectedAdjustmentId(updated.id);
      setStockAdjustments((current) => replaceStockAdjustment(current, updated));
      setAdjustmentMessageTone("success");
      setAdjustmentMessage(stockAdjustmentActionResultLabel(action));
    } catch {
      setAdjustmentMessageTone("danger");
      setAdjustmentMessage(inventoryCopy("stockAdjustment.messages.actionError", { action: stockAdjustmentActionLabel(action) }));
    } finally {
      setAdjustmentSubmitting(false);
    }
  }

  async function handleCreateStockTransfer(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedStock || transferQty.trim() === "" || transferDestinationWarehouseId === selectedStock.warehouseId) {
      return;
    }

    let normalizedQty: string;
    try {
      normalizedQty = normalizeDecimalInput(transferQty, decimalScales.quantity);
    } catch {
      setTransferMessageTone("danger");
      setTransferMessage(inventoryCopy("stockTransfer.messages.invalidQty"));
      return;
    }

    setTransferSubmitting(true);
    setTransferMessage("");
    try {
      const created = await createStockTransfer({
        sourceWarehouseId: selectedStock.warehouseId,
        sourceWarehouseCode: selectedStock.warehouseCode,
        destinationWarehouseId: transferDestinationWarehouseId,
        destinationWarehouseCode: transferDestinationWarehouseCode,
        reasonCode: "warehouse_replenishment",
        lines: [
          {
            sku: selectedStock.sku,
            batchId: selectedStock.batchId,
            batchNo: selectedStock.batchNo,
            sourceLocationId: selectedStock.locationId,
            sourceLocationCode: selectedStock.locationCode,
            quantity: normalizedQty,
            baseUomCode: selectedStock.baseUomCode
          }
        ]
      });
      setSelectedTransferId(created.id);
      setStockTransfers((current) => replaceWarehouseDocument(current, created));
      setTransferMessageTone("success");
      setTransferMessage(inventoryCopy("stockTransfer.messages.created"));
    } catch {
      setTransferMessageTone("danger");
      setTransferMessage(inventoryCopy("stockTransfer.messages.createError"));
    } finally {
      setTransferSubmitting(false);
    }
  }

  async function handleTransferAction(action: WarehouseDocumentAction) {
    if (!selectedTransfer) {
      return;
    }

    setTransferSubmitting(true);
    setTransferMessage("");
    try {
      const updated = await transitionStockTransfer(selectedTransfer.id, action);
      setSelectedTransferId(updated.id);
      setStockTransfers((current) => replaceWarehouseDocument(current, updated));
      setTransferMessageTone("success");
      setTransferMessage(warehouseDocumentActionResultLabel(action));
    } catch {
      setTransferMessageTone("danger");
      setTransferMessage(inventoryCopy("stockTransfer.messages.actionError", { action: warehouseDocumentActionLabel(action) }));
    } finally {
      setTransferSubmitting(false);
    }
  }

  async function handleCreateWarehouseIssue(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedStock || issueQty.trim() === "" || issueDestinationName.trim() === "") {
      return;
    }

    let normalizedQty: string;
    try {
      normalizedQty = normalizeDecimalInput(issueQty, decimalScales.quantity);
    } catch {
      setIssueMessageTone("danger");
      setIssueMessage(inventoryCopy("warehouseIssue.messages.invalidQty"));
      return;
    }

    setIssueSubmitting(true);
    setIssueMessage("");
    try {
      const created = await createWarehouseIssue({
        warehouseId: selectedStock.warehouseId,
        warehouseCode: selectedStock.warehouseCode,
        destinationType: "factory",
        destinationName: issueDestinationName,
        reasonCode: "production_plan_issue",
        lines: [
          {
            sku: selectedStock.sku,
            batchId: selectedStock.batchId,
            batchNo: selectedStock.batchNo,
            locationId: selectedStock.locationId,
            locationCode: selectedStock.locationCode,
            quantity: normalizedQty,
            baseUomCode: selectedStock.baseUomCode,
            sourceDocumentType: "manual_inventory_issue"
          }
        ]
      });
      setSelectedIssueId(created.id);
      setWarehouseIssues((current) => replaceWarehouseDocument(current, created));
      setIssueMessageTone("success");
      setIssueMessage(inventoryCopy("warehouseIssue.messages.created"));
    } catch {
      setIssueMessageTone("danger");
      setIssueMessage(inventoryCopy("warehouseIssue.messages.createError"));
    } finally {
      setIssueSubmitting(false);
    }
  }

  async function handleIssueAction(action: WarehouseDocumentAction) {
    if (!selectedIssue) {
      return;
    }

    setIssueSubmitting(true);
    setIssueMessage("");
    try {
      const updated = await transitionWarehouseIssue(selectedIssue.id, action);
      setSelectedIssueId(updated.id);
      setWarehouseIssues((current) => replaceWarehouseDocument(current, updated));
      setIssueMessageTone("success");
      setIssueMessage(warehouseDocumentActionResultLabel(action));
    } catch {
      setIssueMessageTone("danger");
      setIssueMessage(inventoryCopy("warehouseIssue.messages.actionError", { action: warehouseDocumentActionLabel(action) }));
    } finally {
      setIssueSubmitting(false);
    }
  }

  return (
    <section className="erp-module-page erp-stock-page">
      <header className="erp-page-header">
        <div>
          <p className="erp-module-eyebrow">IV</p>
          <h1 className="erp-page-title">{inventoryCopy("title")}</h1>
          <p className="erp-page-description">{inventoryCopy("description")}</p>
        </div>
      </header>

      <section className="erp-stock-toolbar" aria-label={inventoryCopy("filters.label")}>
        <label className="erp-field">
          <span>{inventoryCopy("filters.warehouse")}</span>
          <select className="erp-input" value={warehouseId} onChange={(event) => setWarehouseId(event.target.value)}>
            {warehouseOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>{inventoryCopy("filters.location")}</span>
          <input
            className="erp-input"
            type="search"
            value={locationId}
            placeholder="bin-hcm-a01"
            onChange={(event) => setLocationId(event.target.value)}
          />
        </label>
        <label className="erp-field">
          <span>{inventoryCopy("filters.sku")}</span>
          <input
            className="erp-input"
            type="search"
            value={sku}
            placeholder="SERUM-30ML"
            onChange={(event) => setSKU(event.target.value.toUpperCase())}
          />
        </label>
      </section>

      <section className="erp-kpi-grid erp-stock-kpis">
        <StockKPI label={inventoryCopy("kpi.physical")} value={summary.physicalQty} uomCode={summary.baseUomCode} tone="normal" />
        <StockKPI label={inventoryCopy("kpi.reserved")} value={summary.reservedQty} uomCode={summary.baseUomCode} tone="warning" />
        <StockKPI label={inventoryCopy("kpi.qcHold")} value={summary.qcHoldQty} uomCode={summary.baseUomCode} tone="danger" />
        <StockKPI label={inventoryCopy("kpi.blocked")} value={summary.blockedQty} uomCode={summary.baseUomCode} tone="danger" />
        <StockKPI label={inventoryCopy("kpi.available")} value={summary.availableQty} uomCode={summary.baseUomCode} tone="success" />
      </section>

      <section className="erp-card erp-card--padded erp-module-table-card">
        <div className="erp-section-header">
          <h2 className="erp-section-title">{inventoryCopy("availableStock.title")}</h2>
          <StatusChip tone={items.length === 0 ? "warning" : "info"}>{inventoryCopy("availableStock.rows", { count: items.length })}</StatusChip>
        </div>
        <DataTable
          columns={columns}
          rows={items}
          getRowKey={(row) => `${row.warehouseId}:${row.locationId ?? "-"}:${row.sku}:${row.batchId ?? "-"}:${row.baseUomCode}`}
          loading={loading}
        />
      </section>

      <section className="erp-card erp-card--padded erp-module-table-card" id="stock-counts">
        <div className="erp-section-header">
          <h2 className="erp-section-title">{inventoryCopy("stockCount.title")}</h2>
          <StatusChip tone={activeOpenStockCount ? "warning" : "info"}>
            {activeOpenStockCount
              ? activeOpenStockCount.countNo
              : inventoryCopy("stockCount.sessions", { count: stockCounts.length })}
          </StatusChip>
        </div>

        <form className="erp-stock-count-form" onSubmit={handleSubmitStockCount}>
          <label className="erp-field">
            <span>{inventoryCopy("stockCount.stockRow")}</span>
            <select
              className="erp-input"
              value={selectedStockKey}
              onChange={(event) => setSelectedStockKey(event.target.value)}
            >
              {stockOptions.map((item) => (
                <option key={stockRowKey(item)} value={stockRowKey(item)}>
                  {stockRowLabel(item)}
                </option>
              ))}
            </select>
          </label>
          <label className="erp-field">
            <span>{inventoryCopy("stockCount.countedQty")}</span>
            <input
              className="erp-input"
              inputMode="decimal"
              type="text"
              value={countedQty}
              placeholder="128.000000"
              onChange={(event) => setCountedQty(event.target.value)}
            />
          </label>
          <button
            className="erp-button erp-button--secondary"
            type="button"
            disabled={!selectedStock || stockCountSubmitting}
            onClick={handleCreateStockCount}
          >
            {inventoryCopy("stockCount.openCount")}
          </button>
          <button
            className="erp-button erp-button--primary"
            type="submit"
            disabled={!activeOpenStockCount || countedQty.trim() === "" || stockCountSubmitting}
          >
            {inventoryCopy("stockCount.submitCount")}
          </button>
          {stockCountMessage ? (
            <StatusChip tone={stockCountMessageTone}>{stockCountMessage}</StatusChip>
          ) : null}
        </form>

        <DataTable
          columns={stockCountColumns}
          rows={stockCounts}
          getRowKey={(row) => row.id}
          loading={stockCountsLoading}
        />
      </section>

      <section className="erp-card erp-card--padded erp-module-table-card" id="stock-adjustments">
        <div className="erp-section-header">
          <h2 className="erp-section-title">{inventoryCopy("stockAdjustment.title")}</h2>
          <StatusChip tone={selectedAdjustment ? stockAdjustmentStatusTone(selectedAdjustment.status) : "info"}>
            {selectedAdjustment
              ? selectedAdjustment.adjustmentNo
              : inventoryCopy("stockAdjustment.requests", { count: stockAdjustments.length })}
          </StatusChip>
        </div>

        <div className="erp-stock-adjustment-actions">
          <label className="erp-field">
            <span>{inventoryCopy("stockAdjustment.request")}</span>
            <select
              className="erp-input"
              value={selectedAdjustment?.id ?? ""}
              onChange={(event) => setSelectedAdjustmentId(event.target.value)}
            >
              {stockAdjustments.map((adjustment) => (
                <option key={adjustment.id} value={adjustment.id}>
                  {adjustment.adjustmentNo} / {stockAdjustmentStatusLabel(adjustment.status)}
                </option>
              ))}
            </select>
          </label>
          <button
            className="erp-button erp-button--secondary"
            type="button"
            disabled={!canTransitionAdjustment(selectedAdjustment, "submit") || adjustmentSubmitting}
            onClick={() => handleAdjustmentAction("submit")}
          >
            {stockAdjustmentActionLabel("submit")}
          </button>
          <button
            className="erp-button erp-button--primary"
            type="button"
            disabled={!canTransitionAdjustment(selectedAdjustment, "approve") || adjustmentSubmitting}
            onClick={() => handleAdjustmentAction("approve")}
          >
            {stockAdjustmentActionLabel("approve")}
          </button>
          <button
            className="erp-button erp-button--danger"
            type="button"
            disabled={!canTransitionAdjustment(selectedAdjustment, "reject") || adjustmentSubmitting}
            onClick={() => handleAdjustmentAction("reject")}
          >
            {stockAdjustmentActionLabel("reject")}
          </button>
          <button
            className="erp-button erp-button--secondary"
            type="button"
            disabled={!canTransitionAdjustment(selectedAdjustment, "post") || adjustmentSubmitting}
            onClick={() => handleAdjustmentAction("post")}
          >
            {stockAdjustmentActionLabel("post")}
          </button>
          {adjustmentMessage ? (
            <StatusChip tone={adjustmentMessageTone}>{adjustmentMessage}</StatusChip>
          ) : null}
        </div>

        <DataTable
          columns={stockAdjustmentColumns}
          rows={stockAdjustments}
          getRowKey={(row) => row.id}
          loading={stockAdjustmentsLoading}
        />
      </section>

      <section className="erp-card erp-card--padded erp-module-table-card" id="stock-transfers">
        <div className="erp-section-header">
          <h2 className="erp-section-title">{inventoryCopy("stockTransfer.title")}</h2>
          <StatusChip tone={selectedTransfer ? warehouseDocumentStatusTone(selectedTransfer.status) : "info"}>
            {selectedTransfer ? selectedTransfer.transferNo : inventoryCopy("stockTransfer.documents", { count: stockTransfers.length })}
          </StatusChip>
        </div>

        <form className="erp-warehouse-document-form" onSubmit={handleCreateStockTransfer}>
          <label className="erp-field">
            <span>{inventoryCopy("stockTransfer.stockRow")}</span>
            <select
              className="erp-input"
              value={selectedStockKey}
              onChange={(event) => setSelectedStockKey(event.target.value)}
            >
              {stockOptions.map((item) => (
                <option key={stockRowKey(item)} value={stockRowKey(item)}>
                  {stockRowLabel(item)}
                </option>
              ))}
            </select>
          </label>
          <label className="erp-field">
            <span>{inventoryCopy("stockTransfer.destinationWarehouse")}</span>
            <select
              className="erp-input"
              value={transferDestinationWarehouseId}
              onChange={(event) => {
                const option = warehouseOptions.find((item) => item.value === event.target.value);
                setTransferDestinationWarehouseId(event.target.value);
                setTransferDestinationWarehouseCode(option?.label || event.target.value);
              }}
            >
              {warehouseOptions
                .filter((option) => option.value !== "" && option.value !== selectedStock?.warehouseId)
                .map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
            </select>
          </label>
          <label className="erp-field">
            <span>{inventoryCopy("stockTransfer.quantity")}</span>
            <input
              className="erp-input"
              inputMode="decimal"
              type="text"
              value={transferQty}
              placeholder="1.000000"
              onChange={(event) => setTransferQty(event.target.value)}
            />
          </label>
          <button
            className="erp-button erp-button--primary"
            type="submit"
            disabled={
              !selectedStock ||
              transferQty.trim() === "" ||
              transferDestinationWarehouseId === selectedStock.warehouseId ||
              transferSubmitting
            }
          >
            {inventoryCopy("stockTransfer.create")}
          </button>
          {transferMessage ? <StatusChip tone={transferMessageTone}>{transferMessage}</StatusChip> : null}
        </form>

        <div className="erp-warehouse-document-actions">
          <label className="erp-field">
            <span>{inventoryCopy("stockTransfer.document")}</span>
            <select
              className="erp-input"
              value={selectedTransfer?.id ?? ""}
              onChange={(event) => setSelectedTransferId(event.target.value)}
            >
              {stockTransfers.map((transfer) => (
                <option key={transfer.id} value={transfer.id}>
                  {transfer.transferNo} / {stockTransferStatusLabel(transfer.status)}
                </option>
              ))}
            </select>
          </label>
          <button
            className="erp-button erp-button--secondary"
            type="button"
            disabled={!canTransitionWarehouseDocument(selectedTransfer, "submit") || transferSubmitting}
            onClick={() => handleTransferAction("submit")}
          >
            {warehouseDocumentActionLabel("submit")}
          </button>
          <button
            className="erp-button erp-button--primary"
            type="button"
            disabled={!canTransitionWarehouseDocument(selectedTransfer, "approve") || transferSubmitting}
            onClick={() => handleTransferAction("approve")}
          >
            {warehouseDocumentActionLabel("approve")}
          </button>
          <button
            className="erp-button erp-button--secondary"
            type="button"
            disabled={!canTransitionWarehouseDocument(selectedTransfer, "post") || transferSubmitting}
            onClick={() => handleTransferAction("post")}
          >
            {warehouseDocumentActionLabel("post")}
          </button>
        </div>

        <DataTable
          columns={stockTransferColumns}
          rows={stockTransfers}
          getRowKey={(row) => row.id}
          loading={stockTransfersLoading}
        />
      </section>

      <section className="erp-card erp-card--padded erp-module-table-card" id="warehouse-issues">
        <div className="erp-section-header">
          <h2 className="erp-section-title">{inventoryCopy("warehouseIssue.title")}</h2>
          <StatusChip tone={selectedIssue ? warehouseDocumentStatusTone(selectedIssue.status) : "info"}>
            {selectedIssue ? selectedIssue.issueNo : inventoryCopy("warehouseIssue.documents", { count: warehouseIssues.length })}
          </StatusChip>
        </div>

        <form className="erp-warehouse-document-form" onSubmit={handleCreateWarehouseIssue}>
          <label className="erp-field">
            <span>{inventoryCopy("warehouseIssue.stockRow")}</span>
            <select
              className="erp-input"
              value={selectedStockKey}
              onChange={(event) => setSelectedStockKey(event.target.value)}
            >
              {stockOptions.map((item) => (
                <option key={stockRowKey(item)} value={stockRowKey(item)}>
                  {stockRowLabel(item)}
                </option>
              ))}
            </select>
          </label>
          <label className="erp-field">
            <span>{inventoryCopy("warehouseIssue.destination")}</span>
            <input
              className="erp-input"
              type="text"
              value={issueDestinationName}
              onChange={(event) => setIssueDestinationName(event.target.value)}
            />
          </label>
          <label className="erp-field">
            <span>{inventoryCopy("warehouseIssue.quantity")}</span>
            <input
              className="erp-input"
              inputMode="decimal"
              type="text"
              value={issueQty}
              placeholder="1.000000"
              onChange={(event) => setIssueQty(event.target.value)}
            />
          </label>
          <button
            className="erp-button erp-button--primary"
            type="submit"
            disabled={!selectedStock || issueQty.trim() === "" || issueDestinationName.trim() === "" || issueSubmitting}
          >
            {inventoryCopy("warehouseIssue.create")}
          </button>
          {issueMessage ? <StatusChip tone={issueMessageTone}>{issueMessage}</StatusChip> : null}
        </form>

        <div className="erp-warehouse-document-actions">
          <label className="erp-field">
            <span>{inventoryCopy("warehouseIssue.document")}</span>
            <select
              className="erp-input"
              value={selectedIssue?.id ?? ""}
              onChange={(event) => setSelectedIssueId(event.target.value)}
            >
              {warehouseIssues.map((issue) => (
                <option key={issue.id} value={issue.id}>
                  {issue.issueNo} / {warehouseIssueStatusLabel(issue.status)}
                </option>
              ))}
            </select>
          </label>
          <button
            className="erp-button erp-button--secondary"
            type="button"
            disabled={!canTransitionWarehouseDocument(selectedIssue, "submit") || issueSubmitting}
            onClick={() => handleIssueAction("submit")}
          >
            {warehouseDocumentActionLabel("submit")}
          </button>
          <button
            className="erp-button erp-button--primary"
            type="button"
            disabled={!canTransitionWarehouseDocument(selectedIssue, "approve") || issueSubmitting}
            onClick={() => handleIssueAction("approve")}
          >
            {warehouseDocumentActionLabel("approve")}
          </button>
          <button
            className="erp-button erp-button--secondary"
            type="button"
            disabled={!canTransitionWarehouseDocument(selectedIssue, "post") || issueSubmitting}
            onClick={() => handleIssueAction("post")}
          >
            {warehouseDocumentActionLabel("post")}
          </button>
        </div>

        <DataTable
          columns={warehouseIssueColumns}
          rows={warehouseIssues}
          getRowKey={(row) => row.id}
          loading={warehouseIssuesLoading}
        />
      </section>

      <section className="erp-card erp-card--padded erp-module-table-card erp-stock-qc-audit">
        <div className="erp-section-header">
          <h2 className="erp-section-title">{inventoryCopy("qc.title")}</h2>
          <StatusChip tone={selectedBatchQCStatus ? qcStatusTone(selectedBatchQCStatus) : "normal"}>
            {selectedBatchQCStatus ? qcStatusLabel(selectedBatchQCStatus) : inventoryCopy("qc.noBatch")}
          </StatusChip>
        </div>

        <form className="erp-stock-qc-form" onSubmit={handleTransitionSubmit}>
          <label className="erp-field">
            <span>{inventoryCopy("qc.batch")}</span>
            <select
              className="erp-input"
              value={selectedBatchId}
              onChange={(event) => setSelectedBatchId(event.target.value)}
            >
              {batchOptions.map((batch) => (
                <option key={batch.id} value={batch.id}>
                  {batch.batchNo} / {batch.sku}
                </option>
              ))}
            </select>
          </label>
          <label className="erp-field">
            <span>{inventoryCopy("qc.nextQC")}</span>
            <select
              className="erp-input"
              value={nextQCStatus}
              onChange={(event) => setNextQCStatus(event.target.value as BatchQCStatus)}
            >
              {qcStatusOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {qcStatusLabel(option.value)}
                </option>
              ))}
            </select>
          </label>
          <label className="erp-field">
            <span>{inventoryCopy("qc.reference")}</span>
            <input
              className="erp-input"
              type="text"
              value={businessRef}
              placeholder="QC-260427-0001"
              onChange={(event) => setBusinessRef(event.target.value)}
            />
          </label>
          <label className="erp-field erp-stock-qc-reason">
            <span>{inventoryCopy("qc.reason")}</span>
            <input
              className="erp-input"
              type="text"
              value={transitionReason}
              placeholder={inventoryCopy("qc.reasonPlaceholder")}
              required
              onChange={(event) => setTransitionReason(event.target.value)}
            />
          </label>
          <button
            className="erp-button erp-button--primary"
            type="submit"
            disabled={!selectedBatchId || transitionReason.trim() === "" || transitionSubmitting}
          >
            {transitionSubmitting ? inventoryCopy("qc.recording") : inventoryCopy("qc.record")}
          </button>
          {transitionMessage ? <StatusChip tone={transitionMessageTone}>{transitionMessage}</StatusChip> : null}
        </form>

        <DataTable
          columns={transitionColumns}
          rows={transitions}
          getRowKey={(row) => row.id}
          loading={historyLoading}
        />
      </section>
    </section>
  );
}

function StockKPI({
  label,
  value,
  uomCode,
  tone
}: {
  label: string;
  value: string;
  uomCode?: string;
  tone: "normal" | "success" | "warning" | "danger";
}) {
  return (
    <article className="erp-card erp-card--padded erp-kpi-card">
      <div className="erp-kpi-label">{label}</div>
      <strong className="erp-kpi-value">{formatQuantity(value, uomCode)}</strong>
      <StatusChip tone={tone}>{label}</StatusChip>
    </article>
  );
}

function statusLabel(item: AvailableStockItem) {
  if (item.availableQty === "0.000000") {
    return inventoryCopy("state.blocked");
  }
  if (item.qcHoldQty !== "0.000000") {
    return inventoryCopy("state.qcHold");
  }
  if (item.blockedQty !== "0.000000") {
    return inventoryCopy("state.blocked");
  }

  return inventoryCopy("state.available");
}

function uniqueBatchOptions(items: AvailableStockItem[]) {
  const batches = new Map<
    string,
    { id: string; batchNo: string; sku: string; qcStatus?: BatchQCStatus }
  >();
  for (const item of items) {
    if (!item.batchId) {
      continue;
    }
    batches.set(item.batchId, {
      id: item.batchId,
      batchNo: item.batchNo ?? item.batchId,
      sku: item.sku,
      qcStatus: item.batchQcStatus
    });
  }

  return Array.from(batches.values()).sort((left, right) => left.batchNo.localeCompare(right.batchNo));
}

function stockRowKey(item: AvailableStockItem) {
  return `${item.warehouseId}:${item.locationId ?? "-"}:${item.sku}:${item.batchId ?? "-"}:${item.baseUomCode}`;
}

function stockRowLabel(item: AvailableStockItem) {
  return `${item.warehouseCode} / ${item.locationCode ?? "-"} / ${item.sku} / ${item.batchNo ?? "-"}`;
}

function stockCountLineLabel(count: StockCountSession) {
  const firstLine = count.lines[0];
  if (!firstLine) {
    return "-";
  }
  if (count.lines.length > 1) {
    return inventoryCopy("stockCount.lines", { count: count.lines.length });
  }

  return `${firstLine.sku} / ${firstLine.batchNo ?? "-"} / ${firstLine.locationCode ?? "-"}`;
}

function stockCountLineQuantity(
  count: StockCountSession,
  field: "expectedQty" | "countedQty" | "deltaQty"
) {
  const firstLine = count.lines[0];
  if (!firstLine) {
    return "-";
  }
  if (count.lines.length > 1) {
    return inventoryCopy("stockCount.lines", { count: count.lines.length });
  }

  return formatQuantity(firstLine[field], firstLine.baseUomCode);
}

function stockCountStatusLabel(status: StockCountStatus) {
  switch (status) {
    case "open":
      return inventoryCopy("stockCount.status.open");
    case "submitted":
      return inventoryCopy("stockCount.status.submitted");
    case "variance_review":
    default:
      return inventoryCopy("stockCount.status.variance_review");
  }
}

function stockCountStatusTone(status: StockCountStatus): StatusTone {
  switch (status) {
    case "submitted":
      return "success";
    case "variance_review":
      return "warning";
    case "open":
    default:
      return "info";
  }
}

function firstActionableAdjustment(adjustments: StockAdjustment[]) {
  return (
    adjustments.find((adjustment) => adjustment.status === "submitted") ??
    adjustments.find((adjustment) => adjustment.status === "approved") ??
    adjustments.find((adjustment) => adjustment.status === "draft")
  );
}

function replaceStockAdjustment(rows: StockAdjustment[], updated: StockAdjustment) {
  if (rows.some((row) => row.id === updated.id)) {
    return rows.map((row) => (row.id === updated.id ? updated : row));
  }

  return [updated, ...rows];
}

function replaceWarehouseDocument<TDocument extends { id: string }>(rows: TDocument[], updated: TDocument) {
  if (rows.some((row) => row.id === updated.id)) {
    return rows.map((row) => (row.id === updated.id ? updated : row));
  }

  return [updated, ...rows];
}

function firstActionableWarehouseDocument<TDocument extends { status: StockTransferStatus | WarehouseIssueStatus }>(
  documents: TDocument[]
) {
  return (
    documents.find((document) => document.status === "submitted") ??
    documents.find((document) => document.status === "approved") ??
    documents.find((document) => document.status === "draft")
  );
}

function canTransitionAdjustment(adjustment: StockAdjustment | undefined, action: StockAdjustmentAction) {
  if (!adjustment) {
    return false;
  }
  if (action === "submit") {
    return adjustment.status === "draft";
  }
  if (action === "approve" || action === "reject") {
    return adjustment.status === "submitted";
  }
  if (action === "post") {
    return adjustment.status === "approved";
  }

  return false;
}

function canTransitionWarehouseDocument(
  document: { status: StockTransferStatus | WarehouseIssueStatus } | undefined,
  action: WarehouseDocumentAction
) {
  if (!document) {
    return false;
  }
  if (action === "submit") {
    return document.status === "draft";
  }
  if (action === "approve") {
    return document.status === "submitted";
  }
  if (action === "post") {
    return document.status === "approved";
  }

  return false;
}

function warehouseDocumentActionLabel(action: WarehouseDocumentAction) {
  switch (action) {
    case "submit":
      return inventoryCopy("warehouseDocument.actions.submit");
    case "approve":
      return inventoryCopy("warehouseDocument.actions.approve");
    case "post":
    default:
      return inventoryCopy("warehouseDocument.actions.post");
  }
}

function warehouseDocumentActionResultLabel(action: WarehouseDocumentAction) {
  switch (action) {
    case "submit":
      return inventoryCopy("warehouseDocument.messages.submitted");
    case "approve":
      return inventoryCopy("warehouseDocument.messages.approved");
    case "post":
    default:
      return inventoryCopy("warehouseDocument.messages.posted");
  }
}

function stockAdjustmentActionLabel(action: StockAdjustmentAction) {
  switch (action) {
    case "submit":
      return inventoryCopy("stockAdjustment.actions.submit");
    case "approve":
      return inventoryCopy("stockAdjustment.actions.approve");
    case "reject":
      return inventoryCopy("stockAdjustment.actions.reject");
    case "post":
    default:
      return inventoryCopy("stockAdjustment.actions.post");
  }
}

function stockAdjustmentActionResultLabel(action: StockAdjustmentAction) {
  switch (action) {
    case "submit":
      return inventoryCopy("stockAdjustment.messages.submitted");
    case "approve":
      return inventoryCopy("stockAdjustment.messages.approved");
    case "reject":
      return inventoryCopy("stockAdjustment.messages.rejected");
    case "post":
    default:
      return inventoryCopy("stockAdjustment.messages.posted");
  }
}

function stockAdjustmentBeforeAfter(adjustment: StockAdjustment) {
  const firstLine = adjustment.lines[0];
  if (!firstLine) {
    return "-";
  }
  if (adjustment.lines.length > 1) {
    return inventoryCopy("stockAdjustment.lines", { count: adjustment.lines.length });
  }

  return (
    <span className="erp-stock-qc-status-flow">
      <span>{formatQuantity(firstLine.expectedQty, firstLine.baseUomCode)}</span>
      <span>{inventoryCopy("stockAdjustment.to")}</span>
      <span>{formatQuantity(firstLine.countedQty, firstLine.baseUomCode)}</span>
    </span>
  );
}

function stockAdjustmentDelta(adjustment: StockAdjustment) {
  if (adjustment.lines.length === 0) {
    return "-";
  }
  if (adjustment.lines.length === 1) {
    const line = adjustment.lines[0];
    return formatQuantity(line.deltaQty, line.baseUomCode);
  }

  const summary = summarizeStockAdjustmentDelta(adjustment);
  return summary.baseUomCode
    ? formatQuantity(summary.deltaQty, summary.baseUomCode)
    : inventoryCopy("stockAdjustment.lines", { count: adjustment.lines.length });
}

function stockAdjustmentStatusLabel(status: StockAdjustmentStatus) {
  switch (status) {
    case "draft":
      return inventoryCopy("stockAdjustment.status.draft");
    case "submitted":
      return inventoryCopy("stockAdjustment.status.submitted");
    case "approved":
      return inventoryCopy("stockAdjustment.status.approved");
    case "rejected":
      return inventoryCopy("stockAdjustment.status.rejected");
    case "posted":
      return inventoryCopy("stockAdjustment.status.posted");
    case "cancelled":
    default:
      return inventoryCopy("stockAdjustment.status.cancelled");
  }
}

function stockTransferStatusLabel(status: StockTransferStatus) {
  return warehouseDocumentStatusLabel(status);
}

function warehouseIssueStatusLabel(status: WarehouseIssueStatus) {
  return warehouseDocumentStatusLabel(status);
}

function warehouseDocumentStatusLabel(status: StockTransferStatus | WarehouseIssueStatus) {
  switch (status) {
    case "draft":
      return inventoryCopy("warehouseDocument.status.draft");
    case "submitted":
      return inventoryCopy("warehouseDocument.status.submitted");
    case "approved":
      return inventoryCopy("warehouseDocument.status.approved");
    case "posted":
      return inventoryCopy("warehouseDocument.status.posted");
    case "cancelled":
    default:
      return inventoryCopy("warehouseDocument.status.cancelled");
  }
}

function warehouseDocumentStatusTone(status: StockTransferStatus | WarehouseIssueStatus): StatusTone {
  switch (status) {
    case "posted":
      return "success";
    case "approved":
      return "info";
    case "submitted":
    case "draft":
      return "warning";
    case "cancelled":
    default:
      return "danger";
  }
}

function stockTransferLineLabel(transfer: StockTransfer) {
  const firstLine = transfer.lines[0];
  if (!firstLine) {
    return "-";
  }
  if (transfer.lines.length > 1) {
    return inventoryCopy("stockTransfer.lines", { count: transfer.lines.length });
  }

  return `${firstLine.sku} / ${firstLine.batchNo ?? "-"} / ${firstLine.sourceLocationCode ?? "-"}`;
}

function stockTransferQuantity(transfer: StockTransfer) {
  const firstLine = transfer.lines[0];
  if (!firstLine) {
    return "-";
  }
  if (transfer.lines.length > 1) {
    return inventoryCopy("stockTransfer.lines", { count: transfer.lines.length });
  }

  return formatQuantity(firstLine.quantity, firstLine.baseUomCode);
}

function warehouseIssueLineLabel(issue: WarehouseIssue) {
  const firstLine = issue.lines[0];
  if (!firstLine) {
    return "-";
  }
  if (issue.lines.length > 1) {
    return inventoryCopy("warehouseIssue.lines", { count: issue.lines.length });
  }

  return `${firstLine.sku} / ${firstLine.batchNo ?? "-"} / ${firstLine.locationCode ?? "-"}`;
}

function warehouseIssueQuantity(issue: WarehouseIssue) {
  const firstLine = issue.lines[0];
  if (!firstLine) {
    return "-";
  }
  if (issue.lines.length > 1) {
    return inventoryCopy("warehouseIssue.lines", { count: issue.lines.length });
  }

  return formatQuantity(firstLine.quantity, firstLine.baseUomCode);
}

function inventoryCopy(key: string, values?: Record<string, string | number>) {
  return t(`inventory.${key}`, { values });
}

function stockAdjustmentStatusTone(status: StockAdjustmentStatus): StatusTone {
  switch (status) {
    case "approved":
    case "posted":
      return "success";
    case "submitted":
    case "draft":
      return "warning";
    case "rejected":
    case "cancelled":
      return "danger";
    default:
      return "normal";
  }
}

function qcStatusLabel(status: BatchQCStatus) {
  switch (status) {
    case "pass":
      return inventoryCopy("qc.status.pass");
    case "fail":
      return inventoryCopy("qc.status.fail");
    case "quarantine":
      return inventoryCopy("qc.status.quarantine");
    case "retest_required":
      return inventoryCopy("qc.status.retest_required");
    case "hold":
    default:
      return inventoryCopy("qc.status.hold");
  }
}

function qcStatusTone(status: BatchQCStatus): "success" | "warning" | "danger" | "info" {
  switch (status) {
    case "pass":
      return "success";
    case "fail":
      return "danger";
    case "quarantine":
    case "retest_required":
      return "warning";
    case "hold":
    default:
      return "info";
  }
}

function formatDateTime(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat("vi-VN", {
    dateStyle: "short",
    timeStyle: "short",
    timeZone: "Asia/Ho_Chi_Minh"
  }).format(date);
}
