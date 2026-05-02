"use client";

import { useEffect, useMemo, useState, type MouseEvent } from "react";
import {
  DataTable,
  EmptyState,
  ScanInput,
  StatusChip,
  type DataTableColumn,
  type StatusTone,
  type ToastMessage
} from "@/shared/design-system/components";
import { t } from "@/shared/i18n";
import { carrierOptions } from "../../shipping/services/carrierManifestService";
import { useWarehouseDailyBoard } from "../hooks/useWarehouseDailyBoard";
import { ShiftClosingPanel } from "./ShiftClosingPanel";
import {
  buildWarehouseFinanceReportHref,
  buildWarehouseFulfillmentDrillDownHref,
  buildWarehouseInboundDrillDownHref,
  buildWarehouseInventoryReportHref,
  buildWarehouseOperationsReportHref,
  buildWarehouseOperationsSignalReportHref,
  buildWarehouseSubcontractDrillDownHref,
  buildWarehouseQueueDrillDownHref,
  buildWarehouseShiftClosingDrillDownHref,
  defaultWarehouseDailyBoardShiftCode,
  defaultWarehouseDailyBoardDate,
  warehouseShiftOptions,
  warehouseInventoryReportStatusFromQueue,
  warehouseOperationsReportStatusFromSignal,
  warehouseOperationsReportStatusFromQueue,
  warehouseOptions,
  warehouseTaskTone
} from "../services/warehouseDailyBoardService";
import type {
  WarehouseFulfillmentDrillDownKey,
  WarehouseInboundDrillDownKey,
  WarehouseInventoryReportStatus,
  WarehouseOperationsReportSignal,
  WarehouseOperationsReportStatus,
  WarehouseSubcontractDrillDownKey
} from "../services/warehouseDailyBoardService";
import type {
  WarehouseDailyBoardQuery,
  WarehouseDailyShiftCode,
  WarehouseDailyTask,
  WarehouseDailyTaskStatus,
  WarehouseInboundMetrics,
  WarehouseSubcontractMetrics
} from "../types";

type QueueFilter = "" | WarehouseDailyTaskStatus | "overdue";
type QueueCardTarget = "task-board" | "drill-down";

const queueOptions: { label: string; value: QueueFilter }[] = [
  { label: warehouseCopy("queue.allActiveQueues"), value: "" },
  { label: warehouseCopy("queue.newOrders"), value: "waiting" },
  { label: warehouseCopy("queue.picking"), value: "picking" },
  { label: warehouseCopy("queue.packed"), value: "packed" },
  { label: warehouseCopy("queue.handover"), value: "handover" },
  { label: warehouseCopy("queue.returns"), value: "returns" },
  { label: warehouseCopy("queue.qaHold"), value: "qa_hold" },
  { label: warehouseCopy("queue.adjustment"), value: "adjustment" },
  { label: warehouseCopy("queue.stockVariance"), value: "mismatch" },
  { label: warehouseCopy("queue.closing"), value: "closing" },
  { label: warehouseCopy("queue.p0Exceptions"), value: "overdue" }
];

const columns: DataTableColumn<WarehouseDailyTask>[] = [
  {
    key: "type",
    header: warehouseCopy("columns.type"),
    render: (row) => (
      <span className="erp-warehouse-task-type">
        <strong>{taskTypeLabel(row.status)}</strong>
        <small>{taskSummaryLabel(row)}</small>
      </span>
    ),
    width: "250px"
  },
  {
    key: "reference",
    header: warehouseCopy("columns.reference"),
    render: (row) => (
      <a className="erp-warehouse-task-link" href={row.href}>
        {row.reference}
      </a>
    ),
    width: "170px"
  },
  {
    key: "status",
    header: warehouseCopy("columns.status"),
    render: (row) => <StatusChip tone={warehouseTaskTone(row.status)}>{statusLabel(row.status)}</StatusChip>,
    width: "130px"
  },
  {
    key: "source",
    header: warehouseCopy("columns.source"),
    render: (row) => (
      <span className="erp-warehouse-source-cell">
        <strong>{sourceLabel(row.source)}</strong>
        <small>{row.sourceField}</small>
      </span>
    ),
    width: "170px"
  },
  {
    key: "owner",
    header: warehouseCopy("columns.owner"),
    render: (row) => row.owner,
    width: "120px"
  },
  {
    key: "sla",
    header: warehouseCopy("columns.sla"),
    render: (row) => <StatusChip tone={priorityTone(row.priority)}>{formatSla(row)}</StatusChip>,
    width: "130px"
  },
  {
    key: "action",
    header: warehouseCopy("columns.action"),
    render: (row) => (
      <a className="erp-button erp-button--secondary erp-button--compact" href={row.href}>
        {taskActionLabel(row.status)}
      </a>
    ),
    width: "150px"
  }
];

export default function WarehouseDailyBoard() {
  const [warehouseId, setWarehouseId] = useState("");
  const [date, setDate] = useState(defaultWarehouseDailyBoardDate);
  const [shiftCode, setShiftCode] = useState<WarehouseDailyShiftCode>(defaultWarehouseDailyBoardShiftCode);
  const [carrierCode, setCarrierCode] = useState("");
  const [queueFilter, setQueueFilter] = useState<QueueFilter>("");
  const [scanFeedback, setScanFeedback] = useState<ToastMessage | undefined>();
  const query = useMemo<WarehouseDailyBoardQuery>(
    () => ({
      warehouseId: warehouseId || undefined,
      date,
      shiftCode,
      carrierCode: carrierCode || undefined
    }),
    [carrierCode, date, shiftCode, warehouseId]
  );
  const { board, loading } = useWarehouseDailyBoard(query);
  const allTasks = board?.tasks ?? [];
  const visibleTasks = useMemo(() => filterTasksByQueue(allTasks, queueFilter), [allTasks, queueFilter]);
  const exceptions = allTasks.filter((task) => task.priority === "P0" || task.status === "mismatch");
  const selectedWarehouseLabel =
    warehouseOptionLabel(warehouseId) ?? board?.warehouseCode ?? warehouseCopy("allWarehouses");
  const selectedCarrierLabel = carrierOptionLabel(carrierCode) ?? warehouseCopy("allCarriers");
  const activeQueueLabel = queueLabel(queueFilter);
  const inventoryReportStatus = warehouseInventoryReportStatusFromQueue(queueFilter);
  const operationsReportStatus = warehouseOperationsReportStatusFromQueue(queueFilter);
  const boardDateLabel = formatBoardDate(date);
  const reportCards: {
    key: "inventory" | "operations" | "finance";
    label: string;
    value: string;
    helper: string;
    chip: string;
    tone: StatusTone;
    href: string;
  }[] = [
    {
      key: "inventory",
      label: warehouseCopy("reports.inventoryReport"),
      value: inventoryReportStatusLabel(inventoryReportStatus),
      helper: `${boardDateLabel} / ${selectedWarehouseLabel}`,
      chip: warehouseCopy("reports.inventory"),
      tone: inventoryReportTone(inventoryReportStatus),
      href: buildWarehouseInventoryReportHref(query, inventoryReportStatus)
    },
    {
      key: "operations",
      label: warehouseCopy("reports.operationsReport"),
      value: operationsReportStatus ? operationsReportStatusLabel(operationsReportStatus) : warehouseCopy("reports.allStatuses"),
      helper: `${boardDateLabel} / ${activeQueueLabel}`,
      chip: warehouseCopy("reports.operations"),
      tone: operationsReportStatus ? operationsReportTone(operationsReportStatus) : "info",
      href: buildWarehouseOperationsReportHref(query, operationsReportStatus)
    },
    {
      key: "finance",
      label: warehouseCopy("reports.financeReport"),
      value: warehouseCopy("reports.arApCod"),
      helper: boardDateLabel,
      chip: warehouseCopy("reports.finance"),
      tone: "info",
      href: buildWarehouseFinanceReportHref(query)
    }
  ];
  const fulfillment = board?.fulfillment;
  const inbound = board?.inbound;
  const subcontract = board?.subcontract;
  const inboundSignalCount = inbound?.receivingPending ?? 0;
  const outboundSignalCount =
    (fulfillment?.pickingOrders ?? 0) +
    (fulfillment?.packedOrders ?? 0) +
    (fulfillment?.waitingHandoverOrders ?? 0) +
    (fulfillment?.missingOrders ?? 0);
  const returnSignalCount = (board?.summary.returnPending ?? 0) + (board?.summary.qaHold ?? 0);
  const stockCountSignalCount = board?.summary.stockCountVariance ?? 0;
  const qcSignalCount = (inbound?.qcHold ?? 0) + (inbound?.qcFail ?? 0) + (inbound?.qcPartial ?? 0);
  const subcontractSignalCount = (subcontract?.openOrders ?? 0) + (subcontract?.factoryClaims ?? 0);
  const outboundHasException = (fulfillment?.missingOrders ?? 0) > 0;
  const qcHasException = (inbound?.qcFail ?? 0) > 0;
  const subcontractHasException = (subcontract?.factoryClaims ?? 0) > 0;
  const inboundSignalStatus = warehouseOperationsReportStatusFromSignal("inbound");
  const outboundSignalStatus = warehouseOperationsReportStatusFromSignal("outbound", {
    hasException: outboundHasException
  });
  const returnSignalStatus = warehouseOperationsReportStatusFromSignal("returns");
  const stockCountSignalStatus = warehouseOperationsReportStatusFromSignal("stock_count");
  const qcSignalStatus = warehouseOperationsReportStatusFromSignal("qc", { hasException: qcHasException });
  const subcontractSignalStatus = warehouseOperationsReportStatusFromSignal("subcontract", {
    hasException: subcontractHasException
  });
  const signalReportCards: {
    key: WarehouseOperationsReportSignal;
    label: string;
    value: number;
    status: WarehouseOperationsReportStatus;
    tone: StatusTone;
    href: string;
  }[] = [
    {
      key: "inbound",
      label: warehouseCopy("signals.inbound"),
      value: inboundSignalCount,
      status: inboundSignalStatus,
      tone: operationsSignalTone(inboundSignalStatus, inboundSignalCount),
      href: buildWarehouseOperationsSignalReportHref("inbound", query)
    },
    {
      key: "outbound",
      label: warehouseCopy("signals.outbound"),
      value: outboundSignalCount,
      status: outboundSignalStatus,
      tone: operationsSignalTone(outboundSignalStatus, outboundSignalCount),
      href: buildWarehouseOperationsSignalReportHref("outbound", query, { hasException: outboundHasException })
    },
    {
      key: "returns",
      label: warehouseCopy("signals.returns"),
      value: returnSignalCount,
      status: returnSignalStatus,
      tone: operationsSignalTone(returnSignalStatus, returnSignalCount),
      href: buildWarehouseOperationsSignalReportHref("returns", query)
    },
    {
      key: "stock_count",
      label: warehouseCopy("signals.stockCount"),
      value: stockCountSignalCount,
      status: stockCountSignalStatus,
      tone: operationsSignalTone(stockCountSignalStatus, stockCountSignalCount),
      href: buildWarehouseOperationsSignalReportHref("stock_count", query)
    },
    {
      key: "qc",
      label: warehouseCopy("signals.qc"),
      value: qcSignalCount,
      status: qcSignalStatus,
      tone: operationsSignalTone(qcSignalStatus, qcSignalCount),
      href: buildWarehouseOperationsSignalReportHref("qc", query, { hasException: qcHasException })
    },
    {
      key: "subcontract",
      label: warehouseCopy("signals.subcontract"),
      value: subcontractSignalCount,
      status: subcontractSignalStatus,
      tone: operationsSignalTone(subcontractSignalStatus, subcontractSignalCount),
      href: buildWarehouseOperationsSignalReportHref("subcontract", query, { hasException: subcontractHasException })
    }
  ];
  const fulfillmentCards: {
    key: WarehouseFulfillmentDrillDownKey;
    label: string;
    value: number;
    tone: "normal" | "success" | "warning" | "danger" | "info";
    href: string;
  }[] = [
    {
      key: "new",
      label: warehouseCopy("fulfillment.new"),
      value: fulfillment?.newOrders ?? 0,
      tone: "normal",
      href: buildWarehouseFulfillmentDrillDownHref("new", query)
    },
    {
      key: "reserved",
      label: warehouseCopy("fulfillment.reserved"),
      value: fulfillment?.reservedOrders ?? 0,
      tone: "info",
      href: buildWarehouseFulfillmentDrillDownHref("reserved", query)
    },
    {
      key: "picking",
      label: warehouseCopy("fulfillment.picking"),
      value: fulfillment?.pickingOrders ?? 0,
      tone: "warning",
      href: buildWarehouseFulfillmentDrillDownHref("picking", query)
    },
    {
      key: "packed",
      label: warehouseCopy("fulfillment.packed"),
      value: fulfillment?.packedOrders ?? 0,
      tone: "success",
      href: buildWarehouseFulfillmentDrillDownHref("packed", query)
    },
    {
      key: "waiting_handover",
      label: warehouseCopy("fulfillment.waitingHandover"),
      value: fulfillment?.waitingHandoverOrders ?? 0,
      tone: "info",
      href: buildWarehouseFulfillmentDrillDownHref("waiting_handover", query)
    },
    {
      key: "missing",
      label: warehouseCopy("fulfillment.missing"),
      value: fulfillment?.missingOrders ?? 0,
      tone: "danger",
      href: buildWarehouseFulfillmentDrillDownHref("missing", query)
    },
    {
      key: "handover",
      label: warehouseCopy("fulfillment.handover"),
      value: fulfillment?.handoverOrders ?? 0,
      tone: "success",
      href: buildWarehouseFulfillmentDrillDownHref("handover", query)
    }
  ];
  const inboundCards: {
    key: WarehouseInboundDrillDownKey;
    label: string;
    value: number;
    tone: "normal" | "success" | "warning" | "danger" | "info";
    href: string;
    chip: string;
  }[] = [
    {
      key: "purchase_orders_incoming",
      label: warehouseCopy("inbound.purchaseOrdersIncoming"),
      value: inbound?.purchaseOrdersIncoming ?? 0,
      tone: "info",
      href: buildWarehouseInboundDrillDownHref("purchase_orders_incoming", query),
      chip: "PO"
    },
    {
      key: "receiving_pending",
      label: warehouseCopy("inbound.receivingPending"),
      value: inbound?.receivingPending ?? 0,
      tone: "warning",
      href: buildWarehouseInboundDrillDownHref("receiving_pending", query),
      chip: "GRN"
    },
    {
      key: "qc_hold",
      label: warehouseCopy("inbound.qcHold"),
      value: inbound?.qcHold ?? 0,
      tone: "warning",
      href: buildWarehouseInboundDrillDownHref("qc_hold", query),
      chip: "QC"
    },
    {
      key: "qc_fail",
      label: warehouseCopy("inbound.qcFail"),
      value: inbound?.qcFail ?? 0,
      tone: "danger",
      href: buildWarehouseInboundDrillDownHref("qc_fail", query),
      chip: "QC"
    },
    {
      key: "qc_pass",
      label: warehouseCopy("inbound.qcPass"),
      value: inbound?.qcPass ?? 0,
      tone: "success",
      href: buildWarehouseInboundDrillDownHref("qc_pass", query),
      chip: "QC"
    },
    {
      key: "qc_partial",
      label: warehouseCopy("inbound.qcPartial"),
      value: inbound?.qcPartial ?? 0,
      tone: "warning",
      href: buildWarehouseInboundDrillDownHref("qc_partial", query),
      chip: "QC"
    },
    {
      key: "supplier_rejections",
      label: warehouseCopy("inbound.supplierRejections"),
      value: inbound?.supplierRejections ?? 0,
      tone: "danger",
      href: buildWarehouseInboundDrillDownHref("supplier_rejections", query),
      chip: "RTS"
    }
  ];
  const subcontractCards: {
    key: WarehouseSubcontractDrillDownKey;
    label: string;
    value: number;
    tone: "normal" | "success" | "warning" | "danger" | "info";
    href: string;
    chip: string;
  }[] = [
    {
      key: "open_orders",
      label: warehouseCopy("subcontract.openOrders"),
      value: subcontract?.openOrders ?? 0,
      tone: "info",
      href: buildWarehouseSubcontractDrillDownHref("open_orders", query),
      chip: warehouseCopy("subcontract.orders")
    },
    {
      key: "material_issued_orders",
      label: warehouseCopy("subcontract.materialIssued"),
      value: subcontract?.materialIssuedOrders ?? 0,
      tone: "warning",
      href: buildWarehouseSubcontractDrillDownHref("material_issued_orders", query),
      chip: warehouseCopy("subcontract.transfers", { count: subcontract?.materialTransferCount ?? 0 })
    },
    {
      key: "sample_pending",
      label: warehouseCopy("subcontract.samplePending"),
      value: subcontract?.samplePending ?? 0,
      tone: "warning",
      href: buildWarehouseSubcontractDrillDownHref("sample_pending", query),
      chip: warehouseCopy("subcontract.sample")
    },
    {
      key: "factory_claims",
      label: warehouseCopy("subcontract.factoryClaims"),
      value: subcontract?.factoryClaims ?? 0,
      tone: (subcontract?.factoryClaims ?? 0) > 0 ? "danger" : "success",
      href: buildWarehouseSubcontractDrillDownHref("factory_claims", query),
      chip: warehouseCopy("subcontract.overdue", { count: subcontract?.factoryClaimsOverdue ?? 0 })
    },
    {
      key: "final_payment_ready_orders",
      label: warehouseCopy("subcontract.finalReady"),
      value: subcontract?.finalPaymentReadyOrders ?? 0,
      tone: "success",
      href: buildWarehouseSubcontractDrillDownHref("final_payment_ready_orders", query),
      chip: warehouseCopy("subcontract.pay")
    }
  ];
  const queueCards: {
    key: QueueFilter;
    label: string;
    value: number;
    tone: "normal" | "success" | "warning" | "danger" | "info";
    helper: string;
    href: string;
    actionLabel: string;
    target: QueueCardTarget;
  }[] = [
    {
      key: "waiting",
      label: warehouseCopy("queue.newOrders"),
      value: board?.summary.waiting ?? 0,
      tone: "normal",
      helper: warehouseCopy("queueCards.receive"),
      href: buildWarehouseQueueDrillDownHref("waiting", query),
      actionLabel: warehouseCopy("queueCards.openQueue"),
      target: "task-board"
    },
    {
      key: "picking",
      label: warehouseCopy("queue.picking"),
      value: board?.summary.picking ?? 0,
      tone: "warning",
      helper: warehouseCopy("queueCards.pick"),
      href: buildWarehouseQueueDrillDownHref("picking", query),
      actionLabel: warehouseCopy("queueCards.openQueue"),
      target: "task-board"
    },
    {
      key: "packed",
      label: warehouseCopy("queue.packed"),
      value: board?.summary.packed ?? 0,
      tone: "success",
      helper: warehouseCopy("queueCards.pack"),
      href: buildWarehouseQueueDrillDownHref("packed", query),
      actionLabel: warehouseCopy("queueCards.openQueue"),
      target: "task-board"
    },
    {
      key: "handover",
      label: warehouseCopy("queue.handover"),
      value: board?.summary.handover ?? 0,
      tone: "info",
      helper: warehouseCopy("queueCards.scan"),
      href: buildWarehouseQueueDrillDownHref("handover", query),
      actionLabel: warehouseCopy("queueCards.openQueue"),
      target: "task-board"
    },
    {
      key: "returns",
      label: warehouseCopy("queue.returns"),
      value: board?.summary.returnPending ?? 0,
      tone: "warning",
      helper: warehouseCopy("queueCards.inspect"),
      href: buildWarehouseQueueDrillDownHref("returns", query),
      actionLabel: warehouseCopy("queueCards.openReturns"),
      target: "drill-down"
    },
    {
      key: "qa_hold",
      label: warehouseCopy("queue.qaHold"),
      value: board?.summary.qaHold ?? 0,
      tone: "danger",
      helper: warehouseCopy("queueCards.release"),
      href: buildWarehouseQueueDrillDownHref("qa_hold", query),
      actionLabel: warehouseCopy("queueCards.openReturns"),
      target: "drill-down"
    },
    {
      key: "adjustment",
      label: warehouseCopy("queue.adjustment"),
      value: board?.summary.adjustmentPending ?? 0,
      tone: "danger",
      helper: warehouseCopy("queueCards.approve"),
      href: buildWarehouseQueueDrillDownHref("adjustment", query),
      actionLabel: warehouseCopy("queueCards.openInventory"),
      target: "drill-down"
    },
    {
      key: "mismatch",
      label: warehouseCopy("queue.stockVariance"),
      value: board?.summary.stockCountVariance ?? 0,
      tone: "danger",
      helper: warehouseCopy("queueCards.reconcile"),
      href: buildWarehouseQueueDrillDownHref("mismatch", query),
      actionLabel: warehouseCopy("queueCards.openCounts"),
      target: "drill-down"
    },
    {
      key: "closing",
      label: warehouseCopy("queue.closing"),
      value: board?.summary.closingBlocked ?? 0,
      tone: (board?.summary.closingBlocked ?? 0) > 0 ? "danger" : "success",
      helper: warehouseCopy("queueCards.close"),
      href: buildWarehouseQueueDrillDownHref("closing", query),
      actionLabel: warehouseCopy("queueCards.openClosing"),
      target: "drill-down"
    },
    {
      key: "overdue",
      label: warehouseCopy("queue.p0Exceptions"),
      value: board?.summary.overdue ?? 0,
      tone: "danger",
      helper: warehouseCopy("queueCards.resolve"),
      href: buildWarehouseQueueDrillDownHref("overdue", query),
      actionLabel: warehouseCopy("queueCards.openQueue"),
      target: "task-board"
    }
  ];

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const nextWarehouseId = filterOptionValue(params.get("warehouse_id"), warehouseOptions);
    const nextShiftCode = filterOptionValue(params.get("shift_code"), warehouseShiftOptions);
    const nextCarrierCode = filterOptionValue(params.get("carrier_code"), carrierOptions);
    const nextQueue = queueFilterFromParam(params.get("queue"));
    const nextDate = params.get("date");

    if (nextWarehouseId !== null) {
      setWarehouseId(nextWarehouseId);
    }
    if (nextDate) {
      setDate(nextDate);
    }
    if (nextShiftCode !== null) {
      setShiftCode(nextShiftCode as WarehouseDailyShiftCode);
    }
    if (nextCarrierCode !== null) {
      setCarrierCode(nextCarrierCode);
    }
    if (nextQueue !== null) {
      setQueueFilter(nextQueue);
      window.setTimeout(scrollTaskBoardIntoView, 0);
    }
  }, []);

  function handleQueueCardClick(
    event: MouseEvent<HTMLAnchorElement>,
    nextQueue: QueueFilter,
    href: string,
    target: QueueCardTarget
  ) {
    if (target !== "task-board" || !shouldHandleClientNavigation(event)) {
      return;
    }

    event.preventDefault();
    setQueueFilter(nextQueue);
    window.history.replaceState(null, "", href);
    window.setTimeout(scrollTaskBoardIntoView, 0);
  }

  return (
    <section className="erp-module-page erp-warehouse-page">
      <header className="erp-page-header">
        <div>
          <p className="erp-module-eyebrow">WH</p>
          <h1 className="erp-page-title">{warehouseCopy("dailyBoard")}</h1>
          <p className="erp-page-description">{warehouseCopy("dailyBoardDescription")}</p>
        </div>
        <div className="erp-page-actions">
          <a className="erp-button erp-button--secondary" href="#scan-station">
            {warehouseCopy("scan")}
          </a>
          <a className="erp-button erp-button--primary" href={buildWarehouseShiftClosingDrillDownHref(query)}>
            {warehouseCopy("shiftClosing")}
          </a>
        </div>
      </header>

      <section className="erp-warehouse-context" aria-label={warehouseCopy("filters.label")}>
        <WarehouseContext label={warehouseCopy("context.date")} value={formatBoardDate(date)} />
        <WarehouseContext label={warehouseCopy("context.shift")} value={shiftLabel(shiftCode)} />
        <WarehouseContext label={warehouseCopy("context.warehouse")} value={board?.warehouseCode ?? selectedWarehouseLabel} />
        <WarehouseContext label={warehouseCopy("context.carrier")} value={selectedCarrierLabel} />
        <WarehouseContext label={warehouseCopy("context.owner")} value={board?.owner ?? warehouseCopy("warehouseLead")} />
        <WarehouseContext label={warehouseCopy("context.shiftStatus")} value={statusLabel(board?.shiftStatus ?? "open")} />
      </section>

      <section className="erp-warehouse-toolbar" aria-label={warehouseCopy("filters.label")}>
        <label className="erp-field">
          <span>{warehouseCopy("filters.warehouse")}</span>
          <select className="erp-input" value={warehouseId} onChange={(event) => setWarehouseId(event.target.value)}>
            {warehouseOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {warehouseOptionLabel(option.value) ?? option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>{warehouseCopy("filters.date")}</span>
          <input className="erp-input" type="date" value={date} onChange={(event) => setDate(event.target.value)} />
        </label>
        <label className="erp-field">
          <span>{warehouseCopy("filters.shift")}</span>
          <select
            className="erp-input"
            value={shiftCode}
            onChange={(event) => setShiftCode(event.target.value as WarehouseDailyShiftCode)}
          >
            {warehouseShiftOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {shiftLabel(option.value)}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>{warehouseCopy("filters.carrier")}</span>
          <select className="erp-input" value={carrierCode} onChange={(event) => setCarrierCode(event.target.value)}>
            {carrierOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {carrierOptionLabel(option.value) ?? option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>{warehouseCopy("filters.queue")}</span>
          <select
            className="erp-input"
            value={queueFilter}
            onChange={(event) => setQueueFilter(event.target.value as QueueFilter)}
          >
            {queueOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
      </section>

      <section className="erp-warehouse-report-grid" aria-label={warehouseCopy("reports.operationsReport")}>
        {reportCards.map((card) => (
          <a className="erp-warehouse-report-card" href={card.href} key={card.key}>
            <span>{card.label}</span>
            <strong>{card.value}</strong>
            <small>{card.helper}</small>
            <StatusChip tone={card.tone}>{card.chip}</StatusChip>
          </a>
        ))}
      </section>

      <section className="erp-warehouse-signal-report-grid" aria-label={warehouseCopy("reports.operationsReport")}>
        {signalReportCards.map((card) => (
          <a className="erp-warehouse-signal-report-card" href={card.href} key={card.key}>
            <span>{card.label}</span>
            <strong>{card.value}</strong>
            <small>{operationsReportStatusLabel(card.status)}</small>
            <StatusChip tone={card.tone}>{warehouseCopy("reports.report")}</StatusChip>
          </a>
        ))}
      </section>

      <section className="erp-kpi-grid erp-warehouse-kpis">
        {queueCards.map((card) => (
          <WarehouseKPI
            active={queueFilter === card.key}
            helper={card.helper}
            key={card.key}
            label={card.label}
            tone={card.tone}
            value={card.value}
            href={card.href}
            actionLabel={card.actionLabel}
            onSelect={(event) => handleQueueCardClick(event, card.key, card.href, card.target)}
          />
        ))}
      </section>

      <section className="erp-warehouse-fulfillment" aria-label={warehouseCopy("metrics.inboundControl")}>
        <div className="erp-section-header">
          <div>
            <h2 className="erp-section-title">{warehouseCopy("metrics.inboundControl")}</h2>
            <p className="erp-section-description">
              {warehouseCopy("metrics.incomingPO", { count: inbound?.purchaseOrdersIncoming ?? 0 })} /{" "}
              {warehouseCopy("metrics.receivingPending", { count: inbound?.receivingPending ?? 0 })}
            </p>
          </div>
          <StatusChip tone={inboundStatusTone(inbound)}>
            {formatMetricTimestamp(inbound?.generatedAt)}
          </StatusChip>
        </div>
        <div className="erp-warehouse-fulfillment-grid">
          {inboundCards.map((card) => (
            <a className="erp-warehouse-fulfillment-card" href={card.href} key={card.key}>
              <span>{card.label}</span>
              <strong>{card.value}</strong>
              <StatusChip tone={card.tone}>{card.chip}</StatusChip>
            </a>
          ))}
        </div>
      </section>

      <section className="erp-warehouse-fulfillment" aria-label={warehouseCopy("metrics.subcontractManufacturing")}>
        <div className="erp-section-header">
          <div>
            <h2 className="erp-section-title">{warehouseCopy("metrics.subcontractManufacturing")}</h2>
            <p className="erp-section-description">
              {warehouseCopy("metrics.openOrders", { count: subcontract?.openOrders ?? 0 })} /{" "}
              {warehouseCopy("metrics.factoryClaims", { count: subcontract?.factoryClaims ?? 0 })}
            </p>
          </div>
          <StatusChip tone={subcontractStatusTone(subcontract)}>
            {formatMetricTimestamp(subcontract?.generatedAt)}
          </StatusChip>
        </div>
        <div className="erp-warehouse-fulfillment-grid erp-warehouse-subcontract-grid">
          {subcontractCards.map((card) => (
            <a className="erp-warehouse-fulfillment-card" href={card.href} key={card.key}>
              <span>{card.label}</span>
              <strong>{card.value}</strong>
              <StatusChip tone={card.tone}>{card.chip}</StatusChip>
            </a>
          ))}
        </div>
      </section>

      <section className="erp-warehouse-fulfillment" aria-label={warehouseCopy("metrics.fulfillmentStatus")}>
        <div className="erp-section-header">
          <div>
            <h2 className="erp-section-title">{warehouseCopy("metrics.fulfillmentStatus")}</h2>
            <p className="erp-section-description">
              {warehouseCopy("metrics.ordersCarrier", { count: fulfillment?.totalOrders ?? 0, carrier: selectedCarrierLabel })}
            </p>
          </div>
          <StatusChip tone={!fulfillment ? "info" : fulfillment.missingOrders > 0 ? "danger" : "success"}>
            {formatMetricTimestamp(fulfillment?.generatedAt)}
          </StatusChip>
        </div>
        <div className="erp-warehouse-fulfillment-grid">
          {fulfillmentCards.map((card) => (
            <a className="erp-warehouse-fulfillment-card" href={card.href} key={card.key}>
              <span>{card.label}</span>
              <strong>{card.value}</strong>
              <StatusChip tone={card.tone}>
                {card.key === "missing" ? warehouseCopy("fulfillment.exception") : warehouseCopy("fulfillment.order")}
              </StatusChip>
            </a>
          ))}
        </div>
      </section>

      <section className="erp-warehouse-source-strip" aria-label={warehouseCopy("columns.source")}>
        {(board?.sourceFields ?? []).map((source) => (
          <span className="erp-warehouse-source-item" key={source.counter}>
            <strong>{counterSourceLabel(source.counter, source.label)}</strong>
            <small>{source.fields.join(" / ")}</small>
          </span>
        ))}
      </section>

      <section className="erp-warehouse-ops">
        <div className="erp-card erp-card--padded erp-warehouse-scan-card" id="scan-station">
          <div className="erp-section-header">
            <h2 className="erp-section-title">{warehouseCopy("scanStation.title")}</h2>
            <StatusChip tone={board?.shiftStatus === "open" ? "warning" : "success"}>
              {statusLabel(board?.shiftStatus ?? "open")}
            </StatusChip>
          </div>
          <ScanInput
            label={warehouseCopy("scanStation.label")}
            placeholder={warehouseCopy("scanStation.placeholder")}
            feedback={scanFeedback}
            onScan={(value) =>
              setScanFeedback({
                id: value,
                title: warehouseCopy("scanStation.queued", { code: value.toUpperCase() }),
                tone: "info"
              })
            }
          />
        </div>

        <div className="erp-card erp-card--padded erp-warehouse-exception-card">
          <div className="erp-section-header">
            <h2 className="erp-section-title">{warehouseCopy("exceptions.title")}</h2>
            <StatusChip tone={exceptions.length > 0 ? "danger" : "success"}>
              {warehouseCopy("exceptions.open", { count: exceptions.length })}
            </StatusChip>
          </div>
          <div className="erp-warehouse-exception-list">
            {exceptions.length > 0 ? (
              exceptions.map((task) => (
                <a className="erp-warehouse-exception" href={task.href} key={task.id}>
                  <strong>{task.reference}</strong>
                  <span>
                    {taskSummaryLabel(task)} / {formatSla(task)}
                  </span>
                </a>
              ))
            ) : (
              <span className="erp-warehouse-empty">{warehouseCopy("exceptions.none")}</span>
            )}
          </div>
        </div>
      </section>

      <div id="shift-closing">
        <ShiftClosingPanel query={{ warehouseId: warehouseId || undefined, date, shiftCode }} />
      </div>

      <section className="erp-card erp-card--padded erp-module-table-card" id="task-board" tabIndex={-1}>
        <div className="erp-section-header">
          <div>
            <h2 className="erp-section-title">{warehouseCopy("taskBoard.title")}</h2>
            <p className="erp-section-description">{activeQueueLabel}</p>
          </div>
          <StatusChip tone={visibleTasks.length === 0 ? "warning" : "info"}>
            {warehouseCopy("taskBoard.rows", { count: visibleTasks.length })}
          </StatusChip>
        </div>
        <DataTable
          columns={columns}
          rows={visibleTasks}
          getRowKey={(row) => row.id}
          loading={loading}
          emptyState={
            <EmptyState
              title={warehouseCopy("taskBoard.emptyTitle")}
              description={warehouseCopy("taskBoard.emptyDescription")}
            />
          }
        />
      </section>
    </section>
  );
}

function WarehouseContext({ label, value }: { label: string; value: string }) {
  return (
    <div className="erp-warehouse-context-item">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function WarehouseKPI({
  actionLabel,
  active,
  helper,
  href,
  label,
  onSelect,
  value,
  tone
}: {
  actionLabel: string;
  active: boolean;
  helper: string;
  href: string;
  label: string;
  onSelect: (event: MouseEvent<HTMLAnchorElement>) => void;
  value: number;
  tone: "normal" | "success" | "warning" | "danger" | "info";
}) {
  return (
    <a
      aria-current={active ? "true" : undefined}
      className={warehouseKpiClassName(active, value, tone)}
      href={href}
      onClick={onSelect}
    >
      <div className="erp-kpi-label">{label}</div>
      <strong className="erp-kpi-value">{value}</strong>
      <span className="erp-warehouse-kpi-footer">
        <StatusChip tone={tone}>{helper}</StatusChip>
        <span className="erp-warehouse-kpi-action">{actionLabel}</span>
      </span>
    </a>
  );
}

function warehouseKpiClassName(active: boolean, value: number, tone: "normal" | "success" | "warning" | "danger" | "info") {
  const classes = ["erp-card", "erp-card--padded", "erp-kpi-card", "erp-warehouse-kpi-button"];
  if (active) {
    classes.push("is-active");
  }
  if (value === 0) {
    classes.push("is-muted");
  } else if (tone === "danger") {
    classes.push("is-critical");
  } else if (tone === "warning") {
    classes.push("is-warning");
  }

  return classes.join(" ");
}

function shouldHandleClientNavigation(event: MouseEvent<HTMLAnchorElement>) {
  return event.button === 0 && !event.metaKey && !event.altKey && !event.ctrlKey && !event.shiftKey;
}

function filterTasksByQueue(tasks: WarehouseDailyTask[], queueFilter: QueueFilter) {
  if (queueFilter === "overdue") {
    return tasks.filter((task) => task.priority === "P0");
  }
  if (queueFilter) {
    return tasks.filter((task) => task.status === queueFilter);
  }

  return tasks;
}

function queueFilterFromParam(value: string | null): QueueFilter | null {
  if (value === "overdue" || statusOptions().includes(value as WarehouseDailyTaskStatus)) {
    return value as QueueFilter;
  }

  return null;
}

function statusOptions() {
  return queueOptions
    .map((option) => option.value)
    .filter((value): value is WarehouseDailyTaskStatus => value !== "" && value !== "overdue");
}

function filterOptionValue<TValue extends string>(
  value: string | null,
  options: readonly { value: TValue; label: string }[]
): TValue | null {
  if (value === null) {
    return null;
  }

  return options.some((option) => option.value === value) ? (value as TValue) : null;
}

function scrollTaskBoardIntoView() {
  document.getElementById("task-board")?.scrollIntoView({ block: "start" });
  document.getElementById("task-board")?.focus({ preventScroll: true });
}

function queueLabel(queueFilter: QueueFilter) {
  return queueOptions.find((option) => option.value === queueFilter)?.label ?? warehouseCopy("queue.allActiveQueues");
}

function inventoryReportStatusLabel(status: WarehouseInventoryReportStatus) {
  switch (status) {
    case "reserved":
      return warehouseCopy("inventoryStatus.reserved");
    case "quarantine":
      return warehouseCopy("inventoryStatus.quarantine");
    case "blocked":
      return warehouseCopy("inventoryStatus.blocked");
    case "available":
    default:
      return warehouseCopy("inventoryStatus.available");
  }
}

function inventoryReportTone(status: WarehouseInventoryReportStatus): StatusTone {
  switch (status) {
    case "reserved":
      return "info";
    case "quarantine":
      return "warning";
    case "blocked":
      return "danger";
    case "available":
    default:
      return "success";
  }
}

function operationsReportStatusLabel(status: WarehouseOperationsReportStatus) {
  switch (status) {
    case "in_progress":
      return warehouseCopy("operationStatus.inProgress");
    case "completed":
      return warehouseCopy("operationStatus.completed");
    case "blocked":
      return warehouseCopy("operationStatus.blocked");
    case "exception":
      return warehouseCopy("operationStatus.exception");
    case "pending":
    default:
      return warehouseCopy("operationStatus.pending");
  }
}

function operationsReportTone(status: WarehouseOperationsReportStatus): StatusTone {
  switch (status) {
    case "completed":
      return "success";
    case "blocked":
    case "exception":
      return "danger";
    case "in_progress":
      return "warning";
    case "pending":
    default:
      return "info";
  }
}

function operationsSignalTone(status: WarehouseOperationsReportStatus, value: number): StatusTone {
  if (value === 0) {
    return "normal";
  }

  return operationsReportTone(status);
}

function sourceLabel(source: WarehouseDailyTask["source"]) {
  switch (source) {
    case "receiving":
      return warehouseCopy("source.receiving");
    case "shipping":
      return warehouseCopy("source.shipping");
    case "returns":
      return warehouseCopy("source.returns");
    case "adjustment":
      return warehouseCopy("source.adjustment");
    case "closing":
      return warehouseCopy("source.closing");
    case "stock_movement":
      return warehouseCopy("source.stockMovement");
    case "reconciliation":
      return warehouseCopy("source.reconciliation");
    case "order_queue":
    default:
      return warehouseCopy("source.orderQueue");
  }
}

function shiftLabel(shiftCode: WarehouseDailyShiftCode) {
  return warehouseCopy(`shift.${shiftCode}`, { fallback: shiftCode });
}

function taskTypeLabel(status: WarehouseDailyTaskStatus) {
  switch (status) {
    case "handover":
      return warehouseCopy("taskType.handover");
    case "mismatch":
      return warehouseCopy("taskType.mismatch");
    case "picking":
      return warehouseCopy("taskType.picking");
    case "packed":
      return warehouseCopy("taskType.packed");
    case "returns":
      return warehouseCopy("taskType.returns");
    case "qa_hold":
      return warehouseCopy("taskType.qaHold");
    case "adjustment":
      return warehouseCopy("taskType.adjustment");
    case "closing":
      return warehouseCopy("taskType.closing");
    case "waiting":
    default:
      return warehouseCopy("taskType.waiting");
  }
}

function taskSummaryLabel(task: WarehouseDailyTask) {
  switch (task.status) {
    case "handover":
      return warehouseCopy("taskSummary.handover");
    case "mismatch":
      return warehouseCopy("taskSummary.mismatch");
    case "picking":
      return warehouseCopy("taskSummary.picking");
    case "packed":
      return warehouseCopy("taskSummary.packed");
    case "returns":
      return warehouseCopy("taskSummary.returns");
    case "qa_hold":
      return warehouseCopy("taskSummary.qaHold");
    case "adjustment":
      return warehouseCopy("taskSummary.adjustment");
    case "closing":
      return task.priority === "P0" ? warehouseCopy("taskSummary.closingBlocked") : warehouseCopy("taskSummary.closingReady");
    case "waiting":
    default:
      return warehouseCopy("taskSummary.waiting");
  }
}

function statusLabel(status: WarehouseDailyTaskStatus | "open" | "closing" | "closed") {
  switch (status) {
    case "handover":
      return warehouseCopy("status.handover");
    case "mismatch":
      return warehouseCopy("status.mismatch");
    case "picking":
      return warehouseCopy("status.picking");
    case "packed":
      return warehouseCopy("status.packed");
    case "returns":
      return warehouseCopy("status.returns");
    case "qa_hold":
      return warehouseCopy("status.qaHold");
    case "adjustment":
      return warehouseCopy("status.adjustment");
    case "closing":
      return warehouseCopy("status.closing");
    case "closed":
      return warehouseCopy("status.closed");
    case "open":
      return warehouseCopy("status.open");
    case "waiting":
    default:
      return warehouseCopy("status.waiting");
  }
}

function priorityTone(priority: WarehouseDailyTask["priority"]) {
  if (priority === "P0") {
    return "danger";
  }
  if (priority === "P1") {
    return "warning";
  }

  return "normal";
}

function taskActionLabel(status: WarehouseDailyTaskStatus) {
  switch (status) {
    case "handover":
      return warehouseCopy("taskAction.scan");
    case "mismatch":
      return warehouseCopy("taskAction.reconcile");
    case "picking":
      return warehouseCopy("taskAction.continue");
    case "packed":
      return warehouseCopy("taskAction.review");
    case "returns":
      return warehouseCopy("taskAction.inspect");
    case "qa_hold":
      return warehouseCopy("taskAction.review");
    case "adjustment":
      return warehouseCopy("taskAction.approve");
    case "closing":
      return warehouseCopy("taskAction.close");
    case "waiting":
    default:
      return warehouseCopy("taskAction.start");
  }
}

function formatSla(row: WarehouseDailyTask) {
  const due = formatDueTime(row.dueAt);
  if (row.priority === "P0") {
    return `P0 / ${due}`;
  }

  return `${row.priority} / ${due}`;
}

function formatDueTime(value: string) {
  return new Intl.DateTimeFormat("vi-VN", {
    timeZone: "Asia/Ho_Chi_Minh",
    hour: "2-digit",
    minute: "2-digit"
  }).format(new Date(value));
}

function formatBoardDate(value: string) {
  return new Intl.DateTimeFormat("vi-VN", {
    timeZone: "Asia/Ho_Chi_Minh",
    day: "2-digit",
    month: "2-digit",
    year: "numeric"
  }).format(new Date(`${value}T00:00:00Z`));
}

function inboundStatusTone(metrics?: WarehouseInboundMetrics) {
  if (!metrics) {
    return "info";
  }
  if (metrics.qcFail > 0 || metrics.supplierRejections > 0) {
    return "danger";
  }
  if (metrics.qcHold > 0 || metrics.qcPartial > 0 || metrics.receivingPending > 0) {
    return "warning";
  }

  return "success";
}

function subcontractStatusTone(metrics?: WarehouseSubcontractMetrics) {
  if (!metrics) {
    return "info";
  }
  if (metrics.factoryClaimsOverdue > 0 || metrics.factoryClaims > 0) {
    return "danger";
  }
  if (metrics.samplePending > 0 || metrics.materialIssuedOrders > 0) {
    return "warning";
  }
  if (metrics.openOrders > 0) {
    return "info";
  }

  return "success";
}

function formatMetricTimestamp(value?: string) {
  if (!value) {
    return warehouseCopy("status.loading");
  }

  return new Intl.DateTimeFormat("vi-VN", {
    timeZone: "Asia/Ho_Chi_Minh",
    hour: "2-digit",
    minute: "2-digit"
  }).format(new Date(value));
}

function warehouseCopy(key: string, options: { fallback?: string; count?: number; carrier?: string; code?: string } = {}) {
  const values: Record<string, string | number> = {};
  if (options.count !== undefined) {
    values.count = options.count;
  }
  if (options.carrier !== undefined) {
    values.carrier = options.carrier;
  }
  if (options.code !== undefined) {
    values.code = options.code;
  }

  return t(`warehouse.${key}`, { fallback: options.fallback, values });
}

function warehouseOptionLabel(value: string) {
  if (value === "") {
    return warehouseCopy("allWarehouses");
  }

  return warehouseOptions.find((option) => option.value === value)?.label;
}

function carrierOptionLabel(value: string) {
  if (value === "") {
    return warehouseCopy("allCarriers");
  }

  return carrierOptions.find((option) => option.value === value)?.label;
}

function counterSourceLabel(counter: string, fallback: string) {
  switch (counter) {
    case "waiting":
      return warehouseCopy("queue.newOrders");
    case "picking":
      return warehouseCopy("queue.picking");
    case "packed":
      return warehouseCopy("queue.packed");
    case "handover":
      return warehouseCopy("queue.handover");
    case "returnPending":
      return warehouseCopy("queue.returns");
    case "qaHold":
      return warehouseCopy("queue.qaHold");
    case "adjustmentPending":
      return warehouseCopy("queue.adjustment");
    case "reconciliationMismatch":
      return warehouseCopy("queue.stockVariance");
    case "closingBlocked":
      return warehouseCopy("queue.closing");
    case "overdue":
      return warehouseCopy("queue.p0Exceptions");
    default:
      return fallback;
  }
}
