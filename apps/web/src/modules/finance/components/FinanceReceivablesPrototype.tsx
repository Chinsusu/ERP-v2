"use client";

import { useEffect, useMemo, useState, type FormEvent } from "react";
import { DataTable, EmptyState, StatusChip, type DataTableColumn, type StatusTone } from "@/shared/design-system/components";
import { CODReconciliationPanel } from "./CODReconciliationPanel";
import { CashTransactionsPanel } from "./CashTransactionsPanel";
import { SupplierPayablesPanel } from "./SupplierPayablesPanel";
import { useCustomerReceivables } from "../hooks/useCustomerReceivables";
import {
  canRecordCustomerReceipt,
  customerReceivableStatusOptions,
  customerReceivableStatusTone,
  formatCustomerReceivableStatus,
  formatFinanceDate,
  formatFinanceMoney,
  getCustomerReceivable,
  markCustomerReceivableDisputed,
  recordCustomerReceivableReceipt,
  voidCustomerReceivable
} from "../services/customerReceivableService";
import type { CustomerReceivable, CustomerReceivableLine, CustomerReceivableQuery, CustomerReceivableStatus } from "../types";

type StatusFilter = "" | CustomerReceivableStatus;

const receivableColumns = (onSelect: (receivable: CustomerReceivable) => void): DataTableColumn<CustomerReceivable>[] => [
  {
    key: "receivable",
    header: "AR",
    render: (row) => (
      <span className="erp-finance-receivable-cell">
        <strong>{row.receivableNo}</strong>
        <small>{row.customerName}</small>
      </span>
    ),
    width: "230px"
  },
  {
    key: "status",
    header: "Status",
    render: (row) => (
      <StatusChip tone={customerReceivableStatusTone(row.status)}>{formatCustomerReceivableStatus(row.status)}</StatusChip>
    ),
    width: "150px"
  },
  {
    key: "due",
    header: "Due",
    render: (row) => formatFinanceDate(row.dueDate),
    width: "120px"
  },
  {
    key: "total",
    header: "Total",
    render: (row) => formatFinanceMoney(row.totalAmount, row.currencyCode),
    align: "right",
    width: "150px"
  },
  {
    key: "paid",
    header: "Paid",
    render: (row) => formatFinanceMoney(row.paidAmount, row.currencyCode),
    align: "right",
    width: "150px"
  },
  {
    key: "outstanding",
    header: "Outstanding",
    render: (row) => formatFinanceMoney(row.outstandingAmount, row.currencyCode),
    align: "right",
    width: "160px"
  },
  {
    key: "action",
    header: "Action",
    render: (row) => (
      <button className="erp-button erp-button--secondary" type="button" onClick={() => onSelect(row)}>
        Open
      </button>
    ),
    width: "96px",
    sticky: true
  }
];

const lineColumns: DataTableColumn<CustomerReceivableLine>[] = [
  {
    key: "description",
    header: "Line",
    render: (row) => (
      <span className="erp-finance-receivable-cell">
        <strong>{row.description}</strong>
        <small>{sourceDocumentLabel(row.sourceDocument)}</small>
      </span>
    )
  },
  {
    key: "amount",
    header: "Amount",
    render: (row) => formatFinanceMoney(row.amount),
    align: "right",
    width: "160px"
  }
];

export function FinanceReceivablesPrototype() {
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState<StatusFilter>("");
  const [selectedReceivableId, setSelectedReceivableId] = useState("ar-cod-260430-0001");
  const [localReceivables, setLocalReceivables] = useState<CustomerReceivable[]>([]);
  const [detailById, setDetailById] = useState<Record<string, CustomerReceivable>>({});
  const [detailLoadingId, setDetailLoadingId] = useState("");
  const [receiptAmount, setReceiptAmount] = useState("");
  const [disputeReason, setDisputeReason] = useState("COD remittance mismatch");
  const [voidReason, setVoidReason] = useState("Duplicate receivable");
  const [busyAction, setBusyAction] = useState("");
  const [feedback, setFeedback] = useState<{ tone: StatusTone; message: string } | null>(null);
  const query = useMemo<CustomerReceivableQuery>(
    () => ({
      search: search || undefined,
      status: status || undefined
    }),
    [search, status]
  );
  const { receivables, loading, error } = useCustomerReceivables(query);
  const visibleReceivables = useMemo(
    () => mergeReceivables(localReceivables, receivables, query),
    [localReceivables, query, receivables]
  );
  const selectedListReceivable =
    visibleReceivables.find((receivable) => receivable.id === selectedReceivableId) ?? visibleReceivables[0] ?? null;
  const selectedReceivable = selectedListReceivable ? detailById[selectedListReceivable.id] ?? selectedListReceivable : null;
  const totals = useMemo(() => summarizeReceivables(visibleReceivables), [visibleReceivables]);

  useEffect(() => {
    if (visibleReceivables.length > 0 && !visibleReceivables.some((receivable) => receivable.id === selectedReceivableId)) {
      setSelectedReceivableId(visibleReceivables[0].id);
    }
  }, [selectedReceivableId, visibleReceivables]);

  useEffect(() => {
    if (!selectedListReceivable || detailById[selectedListReceivable.id] || selectedListReceivable.lines.length > 0) {
      return;
    }

    let active = true;
    setDetailLoadingId(selectedListReceivable.id);
    getCustomerReceivable(selectedListReceivable.id)
      .then((detail) => {
        if (active) {
          setDetailById((current) => ({ ...current, [detail.id]: detail }));
        }
      })
      .catch((reason) => {
        if (active) {
          setFeedback({
            tone: "danger",
            message: reason instanceof Error ? reason.message : "AR detail could not be loaded"
          });
        }
      })
      .finally(() => {
        if (active) {
          setDetailLoadingId("");
        }
      });

    return () => {
      active = false;
    };
  }, [detailById, selectedListReceivable]);

  useEffect(() => {
    if (selectedReceivable && receiptAmount === "") {
      setReceiptAmount(selectedReceivable.outstandingAmount);
    }
  }, [receiptAmount, selectedReceivable]);

  function handleSelect(receivable: CustomerReceivable) {
    setSelectedReceivableId(receivable.id);
    setReceiptAmount(receivable.outstandingAmount);
    setFeedback(null);
  }

  async function handleRecordReceipt(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedReceivable || busyAction) {
      return;
    }

    await runAction("receipt", async () => {
      const result = await recordCustomerReceivableReceipt(selectedReceivable.id, receiptAmount);
      applyActionResult(result.customerReceivable);
      setReceiptAmount(result.customerReceivable.outstandingAmount);
      setFeedback({
        tone: "success",
        message: `${result.customerReceivable.receivableNo} receipt recorded`
      });
    });
  }

  async function handleMarkDisputed(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedReceivable || busyAction) {
      return;
    }

    await runAction("dispute", async () => {
      const result = await markCustomerReceivableDisputed(selectedReceivable.id, disputeReason);
      applyActionResult(result.customerReceivable);
      setFeedback({
        tone: "warning",
        message: `${result.customerReceivable.receivableNo} marked disputed`
      });
    });
  }

  async function handleVoid(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedReceivable || busyAction) {
      return;
    }

    await runAction("void", async () => {
      const result = await voidCustomerReceivable(selectedReceivable.id, voidReason);
      applyActionResult(result.customerReceivable);
      setFeedback({
        tone: "danger",
        message: `${result.customerReceivable.receivableNo} voided`
      });
    });
  }

  async function runAction(action: string, callback: () => Promise<void>) {
    setBusyAction(action);
    setFeedback(null);
    try {
      await callback();
    } catch (reason) {
      setFeedback({
        tone: "danger",
        message: reason instanceof Error ? reason.message : "Finance action failed"
      });
    } finally {
      setBusyAction("");
    }
  }

  function applyActionResult(receivable: CustomerReceivable) {
    setLocalReceivables((current) => [receivable, ...current.filter((candidate) => candidate.id !== receivable.id)]);
    setDetailById((current) => ({ ...current, [receivable.id]: receivable }));
    setSelectedReceivableId(receivable.id);
  }

  return (
    <section className="erp-module-page erp-finance-page">
      <header className="erp-page-header">
        <div>
          <p className="erp-module-eyebrow">FI</p>
          <h1 className="erp-page-title">Finance</h1>
          <p className="erp-page-description">COD reconciliation, customer receivables, supplier payables, collection, and payment tracking</p>
        </div>
        <div className="erp-page-actions">
          <a className="erp-button erp-button--primary" href="#cod-reconciliation">
            COD reconciliation
          </a>
          <a className="erp-button erp-button--secondary" href="#supplier-payables">
            AP list
          </a>
          <a className="erp-button erp-button--secondary" href="#cash-transactions">
            Cash
          </a>
          <a className="erp-button erp-button--secondary" href="#customer-receivables">
            AR list
          </a>
          <a className="erp-button erp-button--secondary" href="#ar-actions">
            Receipt
          </a>
        </div>
      </header>

      <CODReconciliationPanel />

      <SupplierPayablesPanel />

      <CashTransactionsPanel />

      <section className="erp-kpi-grid erp-finance-kpis">
        <FinanceKPI label="Open AR" value={totals.openCount} tone={totals.openCount > 0 ? "warning" : "normal"} />
        <FinanceKPI label="Outstanding" value={formatFinanceMoney(totals.outstandingAmount)} tone="warning" />
        <FinanceKPI label="Paid" value={formatFinanceMoney(totals.paidAmount)} tone="success" />
        <FinanceKPI label="Disputed" value={totals.disputedCount} tone={totals.disputedCount > 0 ? "danger" : "normal"} />
      </section>

      <section className="erp-finance-toolbar" aria-label="AR filters">
        <label className="erp-field">
          <span>Search</span>
          <input
            className="erp-input"
            type="search"
            value={search}
            placeholder="AR no, customer, source document"
            onChange={(event) => setSearch(event.target.value)}
          />
        </label>
        <label className="erp-field">
          <span>Status</span>
          <select className="erp-input" value={status} onChange={(event) => setStatus(event.target.value as StatusFilter)}>
            {customerReceivableStatusOptions.map((option) => (
              <option key={option.value || "all"} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
      </section>

      {feedback ? (
        <div className={`erp-finance-feedback erp-finance-feedback--${feedback.tone}`} role="status">
          {feedback.message}
        </div>
      ) : null}

      <section className="erp-finance-layout">
        <section className="erp-card erp-card--padded" id="customer-receivables">
          <div className="erp-section-header">
            <h2 className="erp-section-title">Customer receivables</h2>
            <StatusChip tone={loading ? "warning" : "info"}>{visibleReceivables.length} records</StatusChip>
          </div>
          <DataTable
            columns={receivableColumns(handleSelect)}
            rows={visibleReceivables}
            getRowKey={(row) => row.id}
            loading={loading}
            error={error?.message}
            emptyState={<EmptyState title="No receivables" description="Try another status or search term." />}
          />
        </section>

        <aside className="erp-finance-side">
          <section className="erp-card erp-card--padded">
            <div className="erp-section-header">
              <h2 className="erp-section-title">AR detail</h2>
              {selectedReceivable ? (
                <StatusChip tone={customerReceivableStatusTone(selectedReceivable.status)}>
                  {formatCustomerReceivableStatus(selectedReceivable.status)}
                </StatusChip>
              ) : null}
            </div>

            {selectedReceivable ? (
              <div className="erp-finance-detail">
                <strong>{selectedReceivable.receivableNo}</strong>
                <span>{selectedReceivable.customerName}</span>
                <dl className="erp-finance-detail-list">
                  <div>
                    <dt>Customer</dt>
                    <dd>{selectedReceivable.customerCode ?? selectedReceivable.customerId}</dd>
                  </div>
                  <div>
                    <dt>Source</dt>
                    <dd>{selectedReceivable.sourceDocument ? sourceDocumentLabel(selectedReceivable.sourceDocument) : "-"}</dd>
                  </div>
                  <div>
                    <dt>Due date</dt>
                    <dd>{formatFinanceDate(selectedReceivable.dueDate)}</dd>
                  </div>
                  <div>
                    <dt>Total</dt>
                    <dd>{formatFinanceMoney(selectedReceivable.totalAmount, selectedReceivable.currencyCode)}</dd>
                  </div>
                  <div>
                    <dt>Paid</dt>
                    <dd>{formatFinanceMoney(selectedReceivable.paidAmount, selectedReceivable.currencyCode)}</dd>
                  </div>
                  <div>
                    <dt>Outstanding</dt>
                    <dd>{formatFinanceMoney(selectedReceivable.outstandingAmount, selectedReceivable.currencyCode)}</dd>
                  </div>
                </dl>
                {selectedReceivable.disputeReason ? <p className="erp-finance-note">{selectedReceivable.disputeReason}</p> : null}
                {selectedReceivable.voidReason ? <p className="erp-finance-note">{selectedReceivable.voidReason}</p> : null}
              </div>
            ) : (
              <EmptyState title="No AR selected" />
            )}
          </section>

          <section className="erp-card erp-card--padded">
            <div className="erp-section-header">
              <h2 className="erp-section-title">Lines</h2>
              <StatusChip tone={detailLoadingId ? "warning" : "normal"}>
                {detailLoadingId ? "Loading" : `${selectedReceivable?.lines.length ?? 0} lines`}
              </StatusChip>
            </div>
            <DataTable
              columns={lineColumns}
              rows={selectedReceivable?.lines ?? []}
              getRowKey={(row) => row.id}
              emptyState={<EmptyState title="No line detail" description="Open AR detail could not load line allocation." />}
            />
          </section>

          <section className="erp-card erp-card--padded" id="ar-actions">
            <div className="erp-section-header">
              <h2 className="erp-section-title">Receipt action</h2>
              <StatusChip tone={canActOnReceipt(selectedReceivable) ? "success" : "normal"}>Finance</StatusChip>
            </div>
            <form className="erp-finance-action-form" onSubmit={handleRecordReceipt}>
              <label className="erp-field">
                <span>Receipt amount</span>
                <input
                  className="erp-input"
                  type="text"
                  inputMode="decimal"
                  value={receiptAmount}
                  disabled={!canActOnReceipt(selectedReceivable) || busyAction !== ""}
                  onChange={(event) => setReceiptAmount(event.target.value)}
                />
              </label>
              <button
                className="erp-button erp-button--primary"
                type="submit"
                disabled={!canActOnReceipt(selectedReceivable) || busyAction !== ""}
              >
                Record receipt
              </button>
            </form>

            <form className="erp-finance-action-form" onSubmit={handleMarkDisputed}>
              <label className="erp-field">
                <span>Dispute reason</span>
                <input
                  className="erp-input"
                  type="text"
                  value={disputeReason}
                  disabled={!canMarkDisputed(selectedReceivable) || busyAction !== ""}
                  onChange={(event) => setDisputeReason(event.target.value)}
                />
              </label>
              <button
                className="erp-button erp-button--secondary"
                type="submit"
                disabled={!canMarkDisputed(selectedReceivable) || busyAction !== ""}
              >
                Mark disputed
              </button>
            </form>

            <form className="erp-finance-action-form" onSubmit={handleVoid}>
              <label className="erp-field">
                <span>Void reason</span>
                <input
                  className="erp-input"
                  type="text"
                  value={voidReason}
                  disabled={!canVoid(selectedReceivable) || busyAction !== ""}
                  onChange={(event) => setVoidReason(event.target.value)}
                />
              </label>
              <button
                className="erp-button erp-button--danger"
                type="submit"
                disabled={!canVoid(selectedReceivable) || busyAction !== ""}
              >
                Void AR
              </button>
            </form>
          </section>
        </aside>
      </section>
    </section>
  );
}

function FinanceKPI({ label, value, tone }: { label: string; value: string | number; tone: StatusTone }) {
  return (
    <div className="erp-card erp-card--padded erp-kpi-card">
      <span className="erp-kpi-label">{label}</span>
      <strong className="erp-kpi-value">{value}</strong>
      <StatusChip tone={tone}>{label}</StatusChip>
    </div>
  );
}

function mergeReceivables(
  localReceivables: CustomerReceivable[],
  receivables: CustomerReceivable[],
  query: CustomerReceivableQuery
) {
  const localMatches = localReceivables.filter((receivable) => matchesReceivableQuery(receivable, query));
  const localIds = new Set(localMatches.map((receivable) => receivable.id));

  return [...localMatches, ...receivables.filter((receivable) => !localIds.has(receivable.id))];
}

function matchesReceivableQuery(receivable: CustomerReceivable, query: CustomerReceivableQuery) {
  const search = query.search?.trim().toLowerCase();
  return (
    (!query.status || receivable.status === query.status) &&
    (!search ||
      [receivable.receivableNo, receivable.customerCode, receivable.customerName, receivable.sourceDocument?.no]
        .filter(Boolean)
        .some((value) => String(value).toLowerCase().includes(search)))
  );
}

function summarizeReceivables(receivables: CustomerReceivable[]) {
  return receivables.reduce(
    (summary, receivable) => ({
      openCount: summary.openCount + (["open", "partially_paid"].includes(receivable.status) ? 1 : 0),
      disputedCount: summary.disputedCount + (receivable.status === "disputed" ? 1 : 0),
      paidAmount: summary.paidAmount + Number(receivable.paidAmount),
      outstandingAmount: summary.outstandingAmount + Number(receivable.outstandingAmount)
    }),
    { openCount: 0, disputedCount: 0, paidAmount: 0, outstandingAmount: 0 }
  );
}

function canActOnReceipt(receivable: CustomerReceivable | null) {
  return Boolean(receivable && canRecordCustomerReceipt(receivable));
}

function canMarkDisputed(receivable: CustomerReceivable | null) {
  return Boolean(receivable && ["open", "partially_paid"].includes(receivable.status));
}

function canVoid(receivable: CustomerReceivable | null) {
  return Boolean(receivable && !["paid", "void"].includes(receivable.status));
}

function sourceDocumentLabel(source: { type: string; no?: string; id?: string }) {
  return `${source.type.replaceAll("_", " ")} / ${source.no ?? source.id ?? "-"}`;
}
