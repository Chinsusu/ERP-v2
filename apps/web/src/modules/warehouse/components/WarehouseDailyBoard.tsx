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
  { label: "All active queues", value: "" },
  { label: "New orders", value: "waiting" },
  { label: "Picking", value: "picking" },
  { label: "Packed", value: "packed" },
  { label: "Handover", value: "handover" },
  { label: "Returns", value: "returns" },
  { label: "QA hold", value: "qa_hold" },
  { label: "Adjustment", value: "adjustment" },
  { label: "Stock variance", value: "mismatch" },
  { label: "Closing", value: "closing" },
  { label: "P0 exceptions", value: "overdue" }
];

const columns: DataTableColumn<WarehouseDailyTask>[] = [
  {
    key: "type",
    header: "Type",
    render: (row) => (
      <span className="erp-warehouse-task-type">
        <strong>{taskTypeLabel(row.status)}</strong>
        <small>{row.title}</small>
      </span>
    ),
    width: "250px"
  },
  {
    key: "reference",
    header: "Reference",
    render: (row) => (
      <a className="erp-warehouse-task-link" href={row.href}>
        {row.reference}
      </a>
    ),
    width: "170px"
  },
  {
    key: "status",
    header: "Status",
    render: (row) => <StatusChip tone={warehouseTaskTone(row.status)}>{statusLabel(row.status)}</StatusChip>,
    width: "130px"
  },
  {
    key: "source",
    header: "Source",
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
    header: "Owner",
    render: (row) => row.owner,
    width: "120px"
  },
  {
    key: "sla",
    header: "SLA",
    render: (row) => <StatusChip tone={priorityTone(row.priority)}>{formatSla(row)}</StatusChip>,
    width: "130px"
  },
  {
    key: "action",
    header: "Action",
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
    warehouseOptions.find((option) => option.value === warehouseId)?.label ?? board?.warehouseCode ?? "All warehouses";
  const selectedCarrierLabel = carrierOptions.find((option) => option.value === carrierCode)?.label ?? "All carriers";
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
      label: "Inventory report",
      value: inventoryReportStatusLabel(inventoryReportStatus),
      helper: `${boardDateLabel} / ${selectedWarehouseLabel}`,
      chip: "Inventory",
      tone: inventoryReportTone(inventoryReportStatus),
      href: buildWarehouseInventoryReportHref(query, inventoryReportStatus)
    },
    {
      key: "operations",
      label: "Operations report",
      value: operationsReportStatus ? operationsReportStatusLabel(operationsReportStatus) : "All statuses",
      helper: `${boardDateLabel} / ${activeQueueLabel}`,
      chip: "Operations",
      tone: operationsReportStatus ? operationsReportTone(operationsReportStatus) : "info",
      href: buildWarehouseOperationsReportHref(query, operationsReportStatus)
    },
    {
      key: "finance",
      label: "Finance report",
      value: "AR / AP / COD",
      helper: boardDateLabel,
      chip: "Finance",
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
      label: "Inbound",
      value: inboundSignalCount,
      status: inboundSignalStatus,
      tone: operationsSignalTone(inboundSignalStatus, inboundSignalCount),
      href: buildWarehouseOperationsSignalReportHref("inbound", query)
    },
    {
      key: "outbound",
      label: "Outbound",
      value: outboundSignalCount,
      status: outboundSignalStatus,
      tone: operationsSignalTone(outboundSignalStatus, outboundSignalCount),
      href: buildWarehouseOperationsSignalReportHref("outbound", query, { hasException: outboundHasException })
    },
    {
      key: "returns",
      label: "Returns",
      value: returnSignalCount,
      status: returnSignalStatus,
      tone: operationsSignalTone(returnSignalStatus, returnSignalCount),
      href: buildWarehouseOperationsSignalReportHref("returns", query)
    },
    {
      key: "stock_count",
      label: "Stock count",
      value: stockCountSignalCount,
      status: stockCountSignalStatus,
      tone: operationsSignalTone(stockCountSignalStatus, stockCountSignalCount),
      href: buildWarehouseOperationsSignalReportHref("stock_count", query)
    },
    {
      key: "qc",
      label: "QC",
      value: qcSignalCount,
      status: qcSignalStatus,
      tone: operationsSignalTone(qcSignalStatus, qcSignalCount),
      href: buildWarehouseOperationsSignalReportHref("qc", query, { hasException: qcHasException })
    },
    {
      key: "subcontract",
      label: "Subcontract",
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
      label: "New",
      value: fulfillment?.newOrders ?? 0,
      tone: "normal",
      href: buildWarehouseFulfillmentDrillDownHref("new", query)
    },
    {
      key: "reserved",
      label: "Reserved",
      value: fulfillment?.reservedOrders ?? 0,
      tone: "info",
      href: buildWarehouseFulfillmentDrillDownHref("reserved", query)
    },
    {
      key: "picking",
      label: "Picking",
      value: fulfillment?.pickingOrders ?? 0,
      tone: "warning",
      href: buildWarehouseFulfillmentDrillDownHref("picking", query)
    },
    {
      key: "packed",
      label: "Packed",
      value: fulfillment?.packedOrders ?? 0,
      tone: "success",
      href: buildWarehouseFulfillmentDrillDownHref("packed", query)
    },
    {
      key: "waiting_handover",
      label: "Waiting handover",
      value: fulfillment?.waitingHandoverOrders ?? 0,
      tone: "info",
      href: buildWarehouseFulfillmentDrillDownHref("waiting_handover", query)
    },
    {
      key: "missing",
      label: "Missing",
      value: fulfillment?.missingOrders ?? 0,
      tone: "danger",
      href: buildWarehouseFulfillmentDrillDownHref("missing", query)
    },
    {
      key: "handover",
      label: "Handed over",
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
      label: "PO incoming",
      value: inbound?.purchaseOrdersIncoming ?? 0,
      tone: "info",
      href: buildWarehouseInboundDrillDownHref("purchase_orders_incoming", query),
      chip: "PO"
    },
    {
      key: "receiving_pending",
      label: "Receiving pending",
      value: inbound?.receivingPending ?? 0,
      tone: "warning",
      href: buildWarehouseInboundDrillDownHref("receiving_pending", query),
      chip: "GRN"
    },
    {
      key: "qc_hold",
      label: "QC hold",
      value: inbound?.qcHold ?? 0,
      tone: "warning",
      href: buildWarehouseInboundDrillDownHref("qc_hold", query),
      chip: "QC"
    },
    {
      key: "qc_fail",
      label: "QC fail",
      value: inbound?.qcFail ?? 0,
      tone: "danger",
      href: buildWarehouseInboundDrillDownHref("qc_fail", query),
      chip: "QC"
    },
    {
      key: "qc_pass",
      label: "QC pass",
      value: inbound?.qcPass ?? 0,
      tone: "success",
      href: buildWarehouseInboundDrillDownHref("qc_pass", query),
      chip: "QC"
    },
    {
      key: "qc_partial",
      label: "QC partial",
      value: inbound?.qcPartial ?? 0,
      tone: "warning",
      href: buildWarehouseInboundDrillDownHref("qc_partial", query),
      chip: "QC"
    },
    {
      key: "supplier_rejections",
      label: "Supplier reject",
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
      label: "Open orders",
      value: subcontract?.openOrders ?? 0,
      tone: "info",
      href: buildWarehouseSubcontractDrillDownHref("open_orders", query),
      chip: "Order"
    },
    {
      key: "material_issued_orders",
      label: "Material issued",
      value: subcontract?.materialIssuedOrders ?? 0,
      tone: "warning",
      href: buildWarehouseSubcontractDrillDownHref("material_issued_orders", query),
      chip: `${subcontract?.materialTransferCount ?? 0} transfers`
    },
    {
      key: "sample_pending",
      label: "Sample pending",
      value: subcontract?.samplePending ?? 0,
      tone: "warning",
      href: buildWarehouseSubcontractDrillDownHref("sample_pending", query),
      chip: "Sample"
    },
    {
      key: "factory_claims",
      label: "Factory claims",
      value: subcontract?.factoryClaims ?? 0,
      tone: (subcontract?.factoryClaims ?? 0) > 0 ? "danger" : "success",
      href: buildWarehouseSubcontractDrillDownHref("factory_claims", query),
      chip: `${subcontract?.factoryClaimsOverdue ?? 0} overdue`
    },
    {
      key: "final_payment_ready_orders",
      label: "Final ready",
      value: subcontract?.finalPaymentReadyOrders ?? 0,
      tone: "success",
      href: buildWarehouseSubcontractDrillDownHref("final_payment_ready_orders", query),
      chip: "Pay"
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
      label: "New orders",
      value: board?.summary.waiting ?? 0,
      tone: "normal",
      helper: "Receive",
      href: buildWarehouseQueueDrillDownHref("waiting", query),
      actionLabel: "Open queue",
      target: "task-board"
    },
    {
      key: "picking",
      label: "Picking",
      value: board?.summary.picking ?? 0,
      tone: "warning",
      helper: "Pick",
      href: buildWarehouseQueueDrillDownHref("picking", query),
      actionLabel: "Open queue",
      target: "task-board"
    },
    {
      key: "packed",
      label: "Packed",
      value: board?.summary.packed ?? 0,
      tone: "success",
      helper: "Pack",
      href: buildWarehouseQueueDrillDownHref("packed", query),
      actionLabel: "Open queue",
      target: "task-board"
    },
    {
      key: "handover",
      label: "Handover",
      value: board?.summary.handover ?? 0,
      tone: "info",
      helper: "Scan",
      href: buildWarehouseQueueDrillDownHref("handover", query),
      actionLabel: "Open queue",
      target: "task-board"
    },
    {
      key: "returns",
      label: "Return pending",
      value: board?.summary.returnPending ?? 0,
      tone: "warning",
      helper: "Inspect",
      href: buildWarehouseQueueDrillDownHref("returns", query),
      actionLabel: "Open returns",
      target: "drill-down"
    },
    {
      key: "qa_hold",
      label: "QA hold",
      value: board?.summary.qaHold ?? 0,
      tone: "danger",
      helper: "Release",
      href: buildWarehouseQueueDrillDownHref("qa_hold", query),
      actionLabel: "Open returns",
      target: "drill-down"
    },
    {
      key: "adjustment",
      label: "Adjustment pending",
      value: board?.summary.adjustmentPending ?? 0,
      tone: "danger",
      helper: "Approve",
      href: buildWarehouseQueueDrillDownHref("adjustment", query),
      actionLabel: "Open inventory",
      target: "drill-down"
    },
    {
      key: "mismatch",
      label: "Stock count variance",
      value: board?.summary.stockCountVariance ?? 0,
      tone: "danger",
      helper: "Reconcile",
      href: buildWarehouseQueueDrillDownHref("mismatch", query),
      actionLabel: "Open counts",
      target: "drill-down"
    },
    {
      key: "closing",
      label: "Closing blocked",
      value: board?.summary.closingBlocked ?? 0,
      tone: (board?.summary.closingBlocked ?? 0) > 0 ? "danger" : "success",
      helper: "Close",
      href: buildWarehouseQueueDrillDownHref("closing", query),
      actionLabel: "Open closing",
      target: "drill-down"
    },
    {
      key: "overdue",
      label: "P0 exceptions",
      value: board?.summary.overdue ?? 0,
      tone: "danger",
      helper: "Resolve",
      href: buildWarehouseQueueDrillDownHref("overdue", query),
      actionLabel: "Open queue",
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
          <h1 className="erp-page-title">Warehouse Daily Board</h1>
          <p className="erp-page-description">Daily warehouse work queue, handover, returns, and variance control</p>
        </div>
        <div className="erp-page-actions">
          <a className="erp-button erp-button--secondary" href="#scan-station">
            Scan
          </a>
          <a className="erp-button erp-button--primary" href={buildWarehouseShiftClosingDrillDownHref(query)}>
            Shift closing
          </a>
        </div>
      </header>

      <section className="erp-warehouse-context" aria-label="Warehouse shift context">
        <WarehouseContext label="Date" value={formatBoardDate(date)} />
        <WarehouseContext label="Shift" value={shiftLabel(shiftCode)} />
        <WarehouseContext label="Warehouse" value={board?.warehouseCode ?? selectedWarehouseLabel} />
        <WarehouseContext label="Carrier" value={selectedCarrierLabel} />
        <WarehouseContext label="Owner" value={board?.owner ?? "Warehouse Lead"} />
        <WarehouseContext label="Shift status" value={statusLabel(board?.shiftStatus ?? "open")} />
      </section>

      <section className="erp-warehouse-toolbar" aria-label="Warehouse daily board filters">
        <label className="erp-field">
          <span>Warehouse</span>
          <select className="erp-input" value={warehouseId} onChange={(event) => setWarehouseId(event.target.value)}>
            {warehouseOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Date</span>
          <input className="erp-input" type="date" value={date} onChange={(event) => setDate(event.target.value)} />
        </label>
        <label className="erp-field">
          <span>Shift</span>
          <select
            className="erp-input"
            value={shiftCode}
            onChange={(event) => setShiftCode(event.target.value as WarehouseDailyShiftCode)}
          >
            {warehouseShiftOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Carrier</span>
          <select className="erp-input" value={carrierCode} onChange={(event) => setCarrierCode(event.target.value)}>
            {carrierOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Queue</span>
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

      <section className="erp-warehouse-report-grid" aria-label="Dashboard report entry points">
        {reportCards.map((card) => (
          <a className="erp-warehouse-report-card" href={card.href} key={card.key}>
            <span>{card.label}</span>
            <strong>{card.value}</strong>
            <small>{card.helper}</small>
            <StatusChip tone={card.tone}>{card.chip}</StatusChip>
          </a>
        ))}
      </section>

      <section className="erp-warehouse-signal-report-grid" aria-label="Warehouse signal report drilldowns">
        {signalReportCards.map((card) => (
          <a className="erp-warehouse-signal-report-card" href={card.href} key={card.key}>
            <span>{card.label}</span>
            <strong>{card.value}</strong>
            <small>{operationsReportStatusLabel(card.status)}</small>
            <StatusChip tone={card.tone}>Report</StatusChip>
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

      <section className="erp-warehouse-fulfillment" aria-label="Inbound status metrics">
        <div className="erp-section-header">
          <div>
            <h2 className="erp-section-title">Inbound control</h2>
            <p className="erp-section-description">
              {inbound?.purchaseOrdersIncoming ?? 0} incoming PO / {inbound?.receivingPending ?? 0} receiving pending
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

      <section className="erp-warehouse-fulfillment" aria-label="Subcontract manufacturing metrics">
        <div className="erp-section-header">
          <div>
            <h2 className="erp-section-title">Subcontract manufacturing</h2>
            <p className="erp-section-description">
              {subcontract?.openOrders ?? 0} open orders / {subcontract?.factoryClaims ?? 0} factory claims
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

      <section className="erp-warehouse-fulfillment" aria-label="Fulfillment status metrics">
        <div className="erp-section-header">
          <div>
            <h2 className="erp-section-title">Fulfillment status</h2>
            <p className="erp-section-description">
              {fulfillment?.totalOrders ?? 0} orders / {selectedCarrierLabel}
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
              <StatusChip tone={card.tone}>{card.key === "missing" ? "Exception" : "Order"}</StatusChip>
            </a>
          ))}
        </div>
      </section>

      <section className="erp-warehouse-source-strip" aria-label="Counter source fields">
        {(board?.sourceFields ?? []).map((source) => (
          <span className="erp-warehouse-source-item" key={source.counter}>
            <strong>{source.label}</strong>
            <small>{source.fields.join(" / ")}</small>
          </span>
        ))}
      </section>

      <section className="erp-warehouse-ops">
        <div className="erp-card erp-card--padded erp-warehouse-scan-card" id="scan-station">
          <div className="erp-section-header">
            <h2 className="erp-section-title">Scan station</h2>
            <StatusChip tone={board?.shiftStatus === "open" ? "warning" : "success"}>
              {board?.shiftStatus ?? "open"}
            </StatusChip>
          </div>
          <ScanInput
            label="Warehouse scan"
            placeholder="Order, manifest, return, or variance code"
            feedback={scanFeedback}
            onScan={(value) =>
              setScanFeedback({
                id: value,
                title: `Queued ${value.toUpperCase()}`,
                tone: "info"
              })
            }
          />
        </div>

        <div className="erp-card erp-card--padded erp-warehouse-exception-card">
          <div className="erp-section-header">
            <h2 className="erp-section-title">Exceptions</h2>
            <StatusChip tone={exceptions.length > 0 ? "danger" : "success"}>{exceptions.length} open</StatusChip>
          </div>
          <div className="erp-warehouse-exception-list">
            {exceptions.length > 0 ? (
              exceptions.map((task) => (
                <a className="erp-warehouse-exception" href={task.href} key={task.id}>
                  <strong>{task.reference}</strong>
                  <span>
                    {task.title} / {formatSla(task)}
                  </span>
                </a>
              ))
            ) : (
              <span className="erp-warehouse-empty">No P0 exceptions</span>
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
            <h2 className="erp-section-title">Task board</h2>
            <p className="erp-section-description">{activeQueueLabel}</p>
          </div>
          <StatusChip tone={visibleTasks.length === 0 ? "warning" : "info"}>{visibleTasks.length} rows</StatusChip>
        </div>
        <DataTable
          columns={columns}
          rows={visibleTasks}
          getRowKey={(row) => row.id}
          loading={loading}
          emptyState={<EmptyState title="No tasks in this queue" description="Change the warehouse, date, or queue filter." />}
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
  return queueOptions.find((option) => option.value === queueFilter)?.label ?? "All active queues";
}

function inventoryReportStatusLabel(status: WarehouseInventoryReportStatus) {
  switch (status) {
    case "reserved":
      return "Reserved stock";
    case "quarantine":
      return "Quarantine stock";
    case "blocked":
      return "Blocked stock";
    case "available":
    default:
      return "Available stock";
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
      return "In progress";
    case "completed":
      return "Completed";
    case "blocked":
      return "Blocked";
    case "exception":
      return "Exception";
    case "pending":
    default:
      return "Pending";
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
      return "Receiving";
    case "shipping":
      return "Shipping";
    case "returns":
      return "Returns";
    case "adjustment":
      return "Adjustment";
    case "closing":
      return "Closing";
    case "stock_movement":
      return "Stock movement";
    case "reconciliation":
      return "Reconciliation";
    case "order_queue":
    default:
      return "Order queue";
  }
}

function shiftLabel(shiftCode: WarehouseDailyShiftCode) {
  return warehouseShiftOptions.find((option) => option.value === shiftCode)?.label ?? shiftCode;
}

function taskTypeLabel(status: WarehouseDailyTaskStatus) {
  switch (status) {
    case "handover":
      return "Handover";
    case "mismatch":
      return "Variance";
    case "picking":
      return "Picking";
    case "packed":
      return "Packed";
    case "returns":
      return "Return";
    case "qa_hold":
      return "QA hold";
    case "adjustment":
      return "Adjustment";
    case "closing":
      return "Closing";
    case "waiting":
    default:
      return "New order";
  }
}

function statusLabel(status: WarehouseDailyTaskStatus | "open" | "closing" | "closed") {
  switch (status) {
    case "handover":
      return "Handover";
    case "mismatch":
      return "Mismatch";
    case "picking":
      return "Picking";
    case "packed":
      return "Packed";
    case "returns":
      return "Returns";
    case "qa_hold":
      return "QA hold";
    case "adjustment":
      return "Adjustment";
    case "closing":
      return "Closing";
    case "closing":
      return "Closing";
    case "closed":
      return "Closed";
    case "open":
      return "Open";
    case "waiting":
    default:
      return "Waiting";
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
      return "Scan";
    case "mismatch":
      return "Reconcile";
    case "picking":
      return "Continue";
    case "packed":
      return "Review";
    case "returns":
      return "Inspect";
    case "qa_hold":
      return "Review";
    case "adjustment":
      return "Approve";
    case "closing":
      return "Close";
    case "waiting":
    default:
      return "Start";
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
  return new Intl.DateTimeFormat("en-GB", {
    hour: "2-digit",
    minute: "2-digit"
  }).format(new Date(value));
}

function formatBoardDate(value: string) {
  return new Intl.DateTimeFormat("en-GB", {
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
    return "Loading";
  }

  return new Intl.DateTimeFormat("en-GB", {
    hour: "2-digit",
    minute: "2-digit"
  }).format(new Date(value));
}
