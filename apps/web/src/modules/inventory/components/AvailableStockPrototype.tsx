"use client";

import { useEffect, useMemo, useState, type FormEvent } from "react";
import { DataTable, StatusChip, type DataTableColumn, type StatusTone } from "@/shared/design-system/components";
import { decimalScales, normalizeDecimalInput } from "@/shared/format/numberFormat";
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
import type {
  AvailableStockItem,
  AvailableStockQuery,
  BatchQCStatus,
  BatchQCTransition,
  StockCountSession,
  StockCountStatus
} from "../types";

const warehouseOptions = [
  { label: "All warehouses", value: "" },
  { label: "HCM", value: "wh-hcm" },
  { label: "HN", value: "wh-hn" }
];

const qcStatusOptions: { label: string; value: BatchQCStatus }[] = [
  { label: "Pass", value: "pass" },
  { label: "Fail", value: "fail" },
  { label: "Quarantine", value: "quarantine" },
  { label: "Retest", value: "retest_required" }
];

const columns: DataTableColumn<AvailableStockItem>[] = [
  {
    key: "warehouse",
    header: "Warehouse",
    render: (row) => row.warehouseCode,
    width: "110px"
  },
  {
    key: "location",
    header: "Location",
    render: (row) => row.locationCode ?? "-",
    width: "110px"
  },
  {
    key: "sku",
    header: "SKU",
    render: (row) => row.sku,
    width: "180px"
  },
  {
    key: "batch",
    header: "Batch",
    render: (row) => row.batchNo ?? "-",
    width: "150px"
  },
  {
    key: "physical",
    header: "Physical",
    render: (row) => formatQuantity(row.physicalQty, row.baseUomCode),
    align: "right",
    width: "130px"
  },
  {
    key: "reserved",
    header: "Reserved",
    render: (row) => formatQuantity(row.reservedQty, row.baseUomCode),
    align: "right",
    width: "130px"
  },
  {
    key: "qcHold",
    header: "QC Hold",
    render: (row) => formatQuantity(row.qcHoldQty, row.baseUomCode),
    align: "right",
    width: "130px"
  },
  {
    key: "blocked",
    header: "Blocked",
    render: (row) => formatQuantity(row.blockedQty, row.baseUomCode),
    align: "right",
    width: "130px"
  },
  {
    key: "available",
    header: "Available",
    render: (row) => formatQuantity(row.availableQty, row.baseUomCode),
    align: "right",
    width: "130px"
  },
  {
    key: "state",
    header: "State",
    render: (row) => <StatusChip tone={availabilityTone(row)}>{statusLabel(row)}</StatusChip>,
    width: "130px"
  }
];

const transitionColumns: DataTableColumn<BatchQCTransition>[] = [
  {
    key: "createdAt",
    header: "Time",
    render: (row) => formatDateTime(row.createdAt),
    width: "180px"
  },
  {
    key: "status",
    header: "QC status",
    render: (row) => (
      <span className="erp-stock-qc-status-flow">
        <StatusChip tone={qcStatusTone(row.fromQcStatus)}>{qcStatusLabel(row.fromQcStatus)}</StatusChip>
        <span>to</span>
        <StatusChip tone={qcStatusTone(row.toQcStatus)}>{qcStatusLabel(row.toQcStatus)}</StatusChip>
      </span>
    ),
    width: "230px"
  },
  {
    key: "actor",
    header: "Actor",
    render: (row) => row.actorId,
    width: "150px"
  },
  {
    key: "businessRef",
    header: "Ref",
    render: (row) => row.businessRef || "-",
    width: "150px"
  },
  {
    key: "reason",
    header: "Reason",
    render: (row) => row.reason
  },
  {
    key: "audit",
    header: "Audit",
    render: (row) => row.auditLogId,
    width: "190px"
  }
];

const stockCountColumns: DataTableColumn<StockCountSession>[] = [
  {
    key: "countNo",
    header: "Count",
    render: (row) => row.countNo,
    width: "150px"
  },
  {
    key: "warehouse",
    header: "Warehouse",
    render: (row) => row.warehouseCode || row.warehouseId,
    width: "110px"
  },
  {
    key: "status",
    header: "Status",
    render: (row) => <StatusChip tone={stockCountStatusTone(row.status)}>{stockCountStatusLabel(row.status)}</StatusChip>,
    width: "150px"
  },
  {
    key: "line",
    header: "Line",
    render: (row) => stockCountLineLabel(row)
  },
  {
    key: "expected",
    header: "Expected",
    render: (row) => stockCountLineQuantity(row, "expectedQty"),
    align: "right",
    width: "130px"
  },
  {
    key: "counted",
    header: "Counted",
    render: (row) => stockCountLineQuantity(row, "countedQty"),
    align: "right",
    width: "130px"
  },
  {
    key: "delta",
    header: "Delta",
    render: (row) => stockCountLineQuantity(row, "deltaQty"),
    align: "right",
    width: "130px"
  },
  {
    key: "updated",
    header: "Updated",
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
  const [stockCounts, setStockCounts] = useState<StockCountSession[]>([]);
  const [stockCountsLoading, setStockCountsLoading] = useState(false);
  const [selectedStockKey, setSelectedStockKey] = useState("");
  const [selectedStockCountId, setSelectedStockCountId] = useState("");
  const [countedQty, setCountedQty] = useState("");
  const [stockCountSubmitting, setStockCountSubmitting] = useState(false);
  const [stockCountMessage, setStockCountMessage] = useState("");
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
      setTransitionMessage("Recorded");
    } catch {
      setTransitionMessage("Could not record");
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
      setStockCountMessage("Count opened");
    } catch {
      setStockCountMessage("Could not open count");
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
      setStockCountMessage("Invalid counted qty");
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
      setStockCountMessage(submitted.status === "variance_review" ? "Variance review" : "Submitted");
    } catch {
      setStockCountMessage("Could not submit count");
    } finally {
      setStockCountSubmitting(false);
    }
  }

  return (
    <section className="erp-module-page erp-stock-page">
      <header className="erp-page-header">
        <div>
          <p className="erp-module-eyebrow">IV</p>
          <h1 className="erp-page-title">Inventory</h1>
          <p className="erp-page-description">Available stock by warehouse, SKU, and batch</p>
        </div>
      </header>

      <section className="erp-stock-toolbar" aria-label="Inventory filters">
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
          <span>Location</span>
          <input
            className="erp-input"
            type="search"
            value={locationId}
            placeholder="bin-hcm-a01"
            onChange={(event) => setLocationId(event.target.value)}
          />
        </label>
        <label className="erp-field">
          <span>SKU</span>
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
        <StockKPI label="Physical" value={summary.physicalQty} uomCode={summary.baseUomCode} tone="normal" />
        <StockKPI label="Reserved" value={summary.reservedQty} uomCode={summary.baseUomCode} tone="warning" />
        <StockKPI label="QC Hold" value={summary.qcHoldQty} uomCode={summary.baseUomCode} tone="danger" />
        <StockKPI label="Blocked" value={summary.blockedQty} uomCode={summary.baseUomCode} tone="danger" />
        <StockKPI label="Available" value={summary.availableQty} uomCode={summary.baseUomCode} tone="success" />
      </section>

      <section className="erp-card erp-card--padded erp-module-table-card">
        <div className="erp-section-header">
          <h2 className="erp-section-title">Available stock</h2>
          <StatusChip tone={items.length === 0 ? "warning" : "info"}>{items.length} rows</StatusChip>
        </div>
        <DataTable
          columns={columns}
          rows={items}
          getRowKey={(row) => `${row.warehouseId}:${row.locationId ?? "-"}:${row.sku}:${row.batchId ?? "-"}:${row.baseUomCode}`}
          loading={loading}
        />
      </section>

      <section className="erp-card erp-card--padded erp-module-table-card">
        <div className="erp-section-header">
          <h2 className="erp-section-title">Stock counts</h2>
          <StatusChip tone={activeOpenStockCount ? "warning" : "info"}>
            {activeOpenStockCount ? activeOpenStockCount.countNo : `${stockCounts.length} sessions`}
          </StatusChip>
        </div>

        <form className="erp-stock-count-form" onSubmit={handleSubmitStockCount}>
          <label className="erp-field">
            <span>Stock row</span>
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
            <span>Counted qty</span>
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
            Open count
          </button>
          <button
            className="erp-button erp-button--primary"
            type="submit"
            disabled={!activeOpenStockCount || countedQty.trim() === "" || stockCountSubmitting}
          >
            Submit count
          </button>
          {stockCountMessage ? (
            <StatusChip tone={stockCountMessage.includes("Could") || stockCountMessage.includes("Invalid") ? "danger" : "success"}>
              {stockCountMessage}
            </StatusChip>
          ) : null}
        </form>

        <DataTable
          columns={stockCountColumns}
          rows={stockCounts}
          getRowKey={(row) => row.id}
          loading={stockCountsLoading}
        />
      </section>

      <section className="erp-card erp-card--padded erp-module-table-card erp-stock-qc-audit">
        <div className="erp-section-header">
          <h2 className="erp-section-title">Batch QC audit</h2>
          <StatusChip tone={selectedBatchQCStatus ? qcStatusTone(selectedBatchQCStatus) : "normal"}>
            {selectedBatchQCStatus ? qcStatusLabel(selectedBatchQCStatus) : "No batch"}
          </StatusChip>
        </div>

        <form className="erp-stock-qc-form" onSubmit={handleTransitionSubmit}>
          <label className="erp-field">
            <span>Batch</span>
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
            <span>Next QC</span>
            <select
              className="erp-input"
              value={nextQCStatus}
              onChange={(event) => setNextQCStatus(event.target.value as BatchQCStatus)}
            >
              {qcStatusOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>
          <label className="erp-field">
            <span>Reference</span>
            <input
              className="erp-input"
              type="text"
              value={businessRef}
              placeholder="QC-260427-0001"
              onChange={(event) => setBusinessRef(event.target.value)}
            />
          </label>
          <label className="erp-field erp-stock-qc-reason">
            <span>Reason</span>
            <input
              className="erp-input"
              type="text"
              value={transitionReason}
              placeholder="COA and visual inspection passed"
              required
              onChange={(event) => setTransitionReason(event.target.value)}
            />
          </label>
          <button
            className="erp-button erp-button--primary"
            type="submit"
            disabled={!selectedBatchId || transitionReason.trim() === "" || transitionSubmitting}
          >
            {transitionSubmitting ? "Recording" : "Record"}
          </button>
          {transitionMessage ? <StatusChip tone={transitionMessage === "Recorded" ? "success" : "danger"}>{transitionMessage}</StatusChip> : null}
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
    return "Blocked";
  }
  if (item.qcHoldQty !== "0.000000") {
    return "QC Hold";
  }
  if (item.blockedQty !== "0.000000") {
    return "Blocked";
  }

  return "Available";
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
    return `${count.lines.length} lines`;
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
    return `${count.lines.length} lines`;
  }

  return formatQuantity(firstLine[field], firstLine.baseUomCode);
}

function stockCountStatusLabel(status: StockCountStatus) {
  switch (status) {
    case "open":
      return "Open";
    case "submitted":
      return "Submitted";
    case "variance_review":
    default:
      return "Variance review";
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

function qcStatusLabel(status: BatchQCStatus) {
  switch (status) {
    case "pass":
      return "Pass";
    case "fail":
      return "Fail";
    case "quarantine":
      return "Quarantine";
    case "retest_required":
      return "Retest";
    case "hold":
    default:
      return "Hold";
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
