"use client";

import { useEffect, useMemo, useState, type FormEvent } from "react";
import { DataTable, EmptyState, StatusChip, type DataTableColumn, type StatusTone } from "@/shared/design-system/components";
import { useCashTransactions } from "../hooks/useCashTransactions";
import {
  cashAllocationTargetAllowedForDirection,
  cashAllocationTargetOptions,
  cashTransactionDirectionOptions,
  cashTransactionDirectionTone,
  cashTransactionStatusOptions,
  cashTransactionStatusTone,
  createCashTransaction,
  formatCashTransactionDirection,
  formatCashTransactionStatus,
  getCashTransaction
} from "../services/cashTransactionService";
import { formatFinanceDate, formatFinanceMoney } from "../services/customerReceivableService";
import type {
  CashAllocationTargetType,
  CashTransaction,
  CashTransactionAllocation,
  CashTransactionDirection,
  CashTransactionQuery,
  CashTransactionStatus
} from "../types";

type CashTransactionStatusFilter = "" | CashTransactionStatus;
type CashTransactionDirectionFilter = "" | CashTransactionDirection;

type CashTransactionFormState = {
  direction: CashTransactionDirection;
  businessDate: string;
  counterpartyId: string;
  counterpartyName: string;
  paymentMethod: string;
  referenceNo: string;
  amount: string;
  targetType: CashAllocationTargetType;
  targetId: string;
  targetNo: string;
  memo: string;
};

const transactionColumns = (onSelect: (transaction: CashTransaction) => void): DataTableColumn<CashTransaction>[] => [
  {
    key: "transaction",
    header: "Cash no",
    render: (row) => (
      <span className="erp-finance-receivable-cell">
        <strong>{row.transactionNo}</strong>
        <small>{row.counterpartyName}</small>
      </span>
    ),
    width: "260px"
  },
  {
    key: "direction",
    header: "Direction",
    render: (row) => (
      <StatusChip tone={cashTransactionDirectionTone(row.direction)}>{formatCashTransactionDirection(row.direction)}</StatusChip>
    ),
    width: "130px"
  },
  {
    key: "status",
    header: "Status",
    render: (row) => (
      <StatusChip tone={cashTransactionStatusTone(row.status)}>{formatCashTransactionStatus(row.status)}</StatusChip>
    ),
    width: "120px"
  },
  {
    key: "business_date",
    header: "Date",
    render: (row) => formatFinanceDate(row.businessDate),
    width: "120px"
  },
  {
    key: "amount",
    header: "Amount",
    render: (row) => formatFinanceMoney(row.totalAmount, row.currencyCode),
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

const allocationColumns: DataTableColumn<CashTransactionAllocation>[] = [
  {
    key: "target",
    header: "Allocation",
    render: (row) => (
      <span className="erp-finance-receivable-cell">
        <strong>{formatTargetType(row.targetType)}</strong>
        <small>{row.targetNo}</small>
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

export function CashTransactionsPanel() {
  const [search, setSearch] = useState("");
  const [direction, setDirection] = useState<CashTransactionDirectionFilter>("");
  const [status, setStatus] = useState<CashTransactionStatusFilter>("");
  const [selectedTransactionId, setSelectedTransactionId] = useState("cash-in-260430-0001");
  const [localTransactions, setLocalTransactions] = useState<CashTransaction[]>([]);
  const [detailById, setDetailById] = useState<Record<string, CashTransaction>>({});
  const [detailLoadingId, setDetailLoadingId] = useState("");
  const [busy, setBusy] = useState(false);
  const [feedback, setFeedback] = useState<{ tone: StatusTone; message: string } | null>(null);
  const [form, setForm] = useState<CashTransactionFormState>({
    direction: "cash_in",
    businessDate: "2026-04-30",
    counterpartyId: "carrier-ghn",
    counterpartyName: "GHN COD",
    paymentMethod: "bank_transfer",
    referenceNo: "BANK-COD-260430-LOCAL",
    amount: "1250000",
    targetType: "customer_receivable",
    targetId: "ar-cod-260430-0001",
    targetNo: "AR-COD-260430-0001",
    memo: "COD receipt"
  });
  const query = useMemo<CashTransactionQuery>(
    () => ({
      search: search || undefined,
      direction: direction || undefined,
      status: status || undefined
    }),
    [direction, search, status]
  );
  const { transactions, loading, error } = useCashTransactions(query);
  const visibleTransactions = useMemo(
    () => mergeTransactions(localTransactions, transactions, query),
    [localTransactions, query, transactions]
  );
  const selectedListTransaction =
    visibleTransactions.find((transaction) => transaction.id === selectedTransactionId) ?? visibleTransactions[0] ?? null;
  const selectedTransaction = selectedListTransaction ? detailById[selectedListTransaction.id] ?? selectedListTransaction : null;
  const totals = useMemo(() => summarizeTransactions(visibleTransactions), [visibleTransactions]);
  const allowedTargets = cashAllocationTargetOptions.filter((option) =>
    cashAllocationTargetAllowedForDirection(form.direction, option.value)
  );

  useEffect(() => {
    if (visibleTransactions.length > 0 && !visibleTransactions.some((transaction) => transaction.id === selectedTransactionId)) {
      setSelectedTransactionId(visibleTransactions[0].id);
    }
  }, [selectedTransactionId, visibleTransactions]);

  useEffect(() => {
    if (!selectedListTransaction || detailById[selectedListTransaction.id] || selectedListTransaction.allocations.length > 0) {
      return;
    }

    let active = true;
    setDetailLoadingId(selectedListTransaction.id);
    getCashTransaction(selectedListTransaction.id)
      .then((detail) => {
        if (active) {
          setDetailById((current) => ({ ...current, [detail.id]: detail }));
        }
      })
      .catch((reason) => {
        if (active) {
          setFeedback({
            tone: "danger",
            message: reason instanceof Error ? reason.message : "Cash transaction detail could not be loaded"
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
  }, [detailById, selectedListTransaction]);

  function handleSelect(transaction: CashTransaction) {
    setSelectedTransactionId(transaction.id);
    setFeedback(null);
  }

  function setFormValue<TKey extends keyof CashTransactionFormState>(key: TKey, value: CashTransactionFormState[TKey]) {
    setForm((current) => ({ ...current, [key]: value }));
  }

  function handleDirectionChange(value: CashTransactionDirection) {
    setForm((current) => ({
      ...current,
      direction: value,
      targetType: value === "cash_in" ? "customer_receivable" : "supplier_payable",
      counterpartyId: value === "cash_in" ? "carrier-ghn" : "supplier-hcm-001",
      counterpartyName: value === "cash_in" ? "GHN COD" : "Nguyen Lieu HCM",
      referenceNo: value === "cash_in" ? "BANK-COD-260430-LOCAL" : "BANK-AP-260430-LOCAL",
      targetId: value === "cash_in" ? "ar-cod-260430-0001" : "ap-supplier-260430-0001",
      targetNo: value === "cash_in" ? "AR-COD-260430-0001" : "AP-SUP-260430-0001",
      memo: value === "cash_in" ? "COD receipt" : "Supplier payment"
    }));
  }

  async function handleCreate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (busy) {
      return;
    }
    if (!cashAllocationTargetAllowedForDirection(form.direction, form.targetType)) {
      setFeedback({ tone: "danger", message: "Allocation target does not match cash direction" });
      return;
    }

    setBusy(true);
    setFeedback(null);
    try {
      const allocationId = `cash-line-${Date.now()}`;
      const transaction = await createCashTransaction({
        direction: form.direction,
        businessDate: form.businessDate,
        counterpartyId: form.counterpartyId,
        counterpartyName: form.counterpartyName,
        paymentMethod: form.paymentMethod,
        referenceNo: form.referenceNo,
        totalAmount: form.amount,
        currencyCode: "VND",
        memo: form.memo,
        allocations: [
          {
            id: allocationId,
            targetType: form.targetType,
            targetId: form.targetId,
            targetNo: form.targetNo,
            amount: form.amount
          }
        ]
      });
      setLocalTransactions((current) => [transaction, ...current.filter((candidate) => candidate.id !== transaction.id)]);
      setDetailById((current) => ({ ...current, [transaction.id]: transaction }));
      setSelectedTransactionId(transaction.id);
      setFeedback({ tone: "success", message: `${transaction.transactionNo} posted` });
    } catch (reason) {
      setFeedback({
        tone: "danger",
        message: reason instanceof Error ? reason.message : "Cash transaction could not be recorded"
      });
    } finally {
      setBusy(false);
    }
  }

  return (
    <section className="erp-finance-section" id="cash-transactions">
      <div className="erp-section-header">
        <div>
          <h2 className="erp-section-title">Cash transactions</h2>
          <p className="erp-section-description">Cash in, cash out, allocation reference, and posted receipt/payment trace</p>
        </div>
        <StatusChip tone={loading ? "warning" : "info"}>{visibleTransactions.length} records</StatusChip>
      </div>

      <section className="erp-kpi-grid erp-finance-kpis">
        <FinanceCashKPI label="Cash in" value={formatFinanceMoney(totals.cashIn)} tone="success" />
        <FinanceCashKPI label="Cash out" value={formatFinanceMoney(totals.cashOut)} tone="warning" />
        <FinanceCashKPI label="Net" value={formatFinanceMoney(totals.net)} tone={totals.net >= 0 ? "success" : "danger"} />
        <FinanceCashKPI label="Posted" value={totals.postedCount} tone="info" />
      </section>

      <section className="erp-finance-toolbar" aria-label="Cash transaction filters">
        <label className="erp-field">
          <span>Search</span>
          <input
            className="erp-input"
            type="search"
            value={search}
            placeholder="Cash no, counterparty, reference"
            onChange={(event) => setSearch(event.target.value)}
          />
        </label>
        <label className="erp-field">
          <span>Direction</span>
          <select className="erp-input" value={direction} onChange={(event) => setDirection(event.target.value as CashTransactionDirectionFilter)}>
            {cashTransactionDirectionOptions.map((option) => (
              <option key={option.value || "all"} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Status</span>
          <select className="erp-input" value={status} onChange={(event) => setStatus(event.target.value as CashTransactionStatusFilter)}>
            {cashTransactionStatusOptions.map((option) => (
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
        <section className="erp-card erp-card--padded">
          <DataTable
            columns={transactionColumns(handleSelect)}
            rows={visibleTransactions}
            getRowKey={(row) => row.id}
            loading={loading}
            error={error?.message}
            emptyState={<EmptyState title="No cash transactions" description="Try another direction or search term." />}
          />
        </section>

        <aside className="erp-finance-side">
          <section className="erp-card erp-card--padded">
            <div className="erp-section-header">
              <h2 className="erp-section-title">Cash detail</h2>
              {selectedTransaction ? (
                <StatusChip tone={cashTransactionStatusTone(selectedTransaction.status)}>
                  {formatCashTransactionStatus(selectedTransaction.status)}
                </StatusChip>
              ) : null}
            </div>

            {selectedTransaction ? (
              <div className="erp-finance-detail">
                <strong>{selectedTransaction.transactionNo}</strong>
                <span>{selectedTransaction.counterpartyName}</span>
                <dl className="erp-finance-detail-list">
                  <div>
                    <dt>Direction</dt>
                    <dd>{formatCashTransactionDirection(selectedTransaction.direction)}</dd>
                  </div>
                  <div>
                    <dt>Business date</dt>
                    <dd>{formatFinanceDate(selectedTransaction.businessDate)}</dd>
                  </div>
                  <div>
                    <dt>Method</dt>
                    <dd>{selectedTransaction.paymentMethod}</dd>
                  </div>
                  <div>
                    <dt>Reference</dt>
                    <dd>{selectedTransaction.referenceNo ?? "-"}</dd>
                  </div>
                  <div>
                    <dt>Total</dt>
                    <dd>{formatFinanceMoney(selectedTransaction.totalAmount, selectedTransaction.currencyCode)}</dd>
                  </div>
                  <div>
                    <dt>Posted by</dt>
                    <dd>{selectedTransaction.postedBy ?? "-"}</dd>
                  </div>
                </dl>
                {selectedTransaction.memo ? <p className="erp-finance-note">{selectedTransaction.memo}</p> : null}
              </div>
            ) : (
              <EmptyState title="No cash transaction selected" />
            )}
          </section>

          <section className="erp-card erp-card--padded">
            <div className="erp-section-header">
              <h2 className="erp-section-title">Allocations</h2>
              <StatusChip tone={detailLoadingId ? "warning" : "normal"}>
                {detailLoadingId ? "Loading" : `${selectedTransaction?.allocations.length ?? 0} lines`}
              </StatusChip>
            </div>
            <DataTable
              columns={allocationColumns}
              rows={selectedTransaction?.allocations ?? []}
              getRowKey={(row) => row.id}
              emptyState={<EmptyState title="No allocation detail" description="Open cash detail could not load allocation lines." />}
            />
          </section>

          <section className="erp-card erp-card--padded">
            <div className="erp-section-header">
              <h2 className="erp-section-title">Record cash</h2>
              <StatusChip tone="success">Finance</StatusChip>
            </div>
            <form className="erp-finance-action-form" onSubmit={handleCreate}>
              <label className="erp-field">
                <span>Direction</span>
                <select className="erp-input" value={form.direction} onChange={(event) => handleDirectionChange(event.target.value as CashTransactionDirection)}>
                  <option value="cash_in">Cash in</option>
                  <option value="cash_out">Cash out</option>
                </select>
              </label>
              <label className="erp-field">
                <span>Business date</span>
                <input className="erp-input" type="date" value={form.businessDate} onChange={(event) => setFormValue("businessDate", event.target.value)} />
              </label>
              <label className="erp-field">
                <span>Counterparty</span>
                <input
                  className="erp-input"
                  type="text"
                  value={form.counterpartyName}
                  onChange={(event) => setFormValue("counterpartyName", event.target.value)}
                />
              </label>
              <label className="erp-field">
                <span>Payment method</span>
                <input className="erp-input" type="text" value={form.paymentMethod} onChange={(event) => setFormValue("paymentMethod", event.target.value)} />
              </label>
              <label className="erp-field">
                <span>Reference</span>
                <input className="erp-input" type="text" value={form.referenceNo} onChange={(event) => setFormValue("referenceNo", event.target.value)} />
              </label>
              <label className="erp-field">
                <span>Amount</span>
                <input
                  className="erp-input"
                  type="text"
                  inputMode="decimal"
                  value={form.amount}
                  onChange={(event) => setFormValue("amount", event.target.value)}
                />
              </label>
              <label className="erp-field">
                <span>Allocation target</span>
                <select className="erp-input" value={form.targetType} onChange={(event) => setFormValue("targetType", event.target.value as CashAllocationTargetType)}>
                  {allowedTargets.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              </label>
              <label className="erp-field">
                <span>Target no</span>
                <input className="erp-input" type="text" value={form.targetNo} onChange={(event) => setFormValue("targetNo", event.target.value)} />
              </label>
              <button className="erp-button erp-button--primary" type="submit" disabled={busy}>
                Record cash
              </button>
            </form>
          </section>
        </aside>
      </section>
    </section>
  );
}

function FinanceCashKPI({ label, value, tone }: { label: string; value: string | number; tone: StatusTone }) {
  return (
    <div className="erp-card erp-card--padded erp-kpi-card">
      <span className="erp-kpi-label">{label}</span>
      <strong className="erp-kpi-value">{value}</strong>
      <StatusChip tone={tone}>{label}</StatusChip>
    </div>
  );
}

function mergeTransactions(
  localTransactions: CashTransaction[],
  transactions: CashTransaction[],
  query: CashTransactionQuery
) {
  const localMatches = localTransactions.filter((transaction) => matchesTransactionQuery(transaction, query));
  const localIds = new Set(localMatches.map((transaction) => transaction.id));

  return [...localMatches, ...transactions.filter((transaction) => !localIds.has(transaction.id))];
}

function matchesTransactionQuery(transaction: CashTransaction, query: CashTransactionQuery) {
  const search = query.search?.trim().toLowerCase();
  return (
    (!query.status || transaction.status === query.status) &&
    (!query.direction || transaction.direction === query.direction) &&
    (!search ||
      [transaction.transactionNo, transaction.counterpartyName, transaction.referenceNo]
        .filter(Boolean)
        .some((value) => String(value).toLowerCase().includes(search)))
  );
}

function summarizeTransactions(transactions: CashTransaction[]) {
  return transactions.reduce(
    (summary, transaction) => {
      const amount = Number(transaction.totalAmount);
      return {
        cashIn: summary.cashIn + (transaction.direction === "cash_in" ? amount : 0),
        cashOut: summary.cashOut + (transaction.direction === "cash_out" ? amount : 0),
        net: summary.net + (transaction.direction === "cash_in" ? amount : -amount),
        postedCount: summary.postedCount + (transaction.status === "posted" ? 1 : 0)
      };
    },
    { cashIn: 0, cashOut: 0, net: 0, postedCount: 0 }
  );
}

function formatTargetType(targetType: CashAllocationTargetType) {
  return targetType
    .split("_")
    .map((part) => `${part.slice(0, 1).toUpperCase()}${part.slice(1)}`)
    .join(" ");
}
