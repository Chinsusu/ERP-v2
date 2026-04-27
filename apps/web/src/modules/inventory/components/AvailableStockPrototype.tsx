"use client";

import { useEffect, useMemo, useState, type FormEvent } from "react";
import { DataTable, StatusChip, type DataTableColumn } from "@/shared/design-system/components";
import { useAvailableStock } from "../hooks/useAvailableStock";
import {
  availabilityTone,
  createBatchQCTransition,
  formatQuantity,
  getBatchQCTransitions
} from "../services/stockAvailabilityService";
import type { AvailableStockItem, AvailableStockQuery, BatchQCStatus, BatchQCTransition } from "../types";

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
