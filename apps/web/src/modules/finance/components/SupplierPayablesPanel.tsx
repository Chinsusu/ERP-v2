"use client";

import { useEffect, useMemo, useState, type FormEvent } from "react";
import { DataTable, EmptyState, StatusChip, type DataTableColumn, type StatusTone } from "@/shared/design-system/components";
import { useSupplierInvoices } from "../hooks/useSupplierInvoices";
import { useSupplierPayables } from "../hooks/useSupplierPayables";
import {
  createSupplierInvoice,
  formatSupplierInvoiceStatus,
  supplierInvoiceStatusTone
} from "../services/supplierInvoiceService";
import {
  approveSupplierPayablePayment,
  canApproveSupplierPayablePayment,
  canRejectSupplierPayablePayment,
  canRecordSupplierPayablePayment,
  canRequestSupplierPayablePayment,
  canVoidSupplierPayable,
  formatSupplierPayableStatus,
  getSupplierPayable,
  recordSupplierPayablePayment,
  rejectSupplierPayablePayment,
  requestSupplierPayablePayment,
  supplierPayableStatusOptions,
  supplierPayableStatusTone,
  voidSupplierPayable
} from "../services/supplierPayableService";
import { formatFinanceDate, formatFinanceMoney } from "../services/customerReceivableService";
import type {
  SupplierInvoice,
  SupplierInvoiceLine,
  SupplierInvoiceQuery,
  SupplierPayable,
  SupplierPayableLine,
  SupplierPayableQuery,
  SupplierPayableStatus
} from "../types";

type SupplierPayableStatusFilter = "" | SupplierPayableStatus;

const payableColumns = (onSelect: (payable: SupplierPayable) => void): DataTableColumn<SupplierPayable>[] => [
  {
    key: "payable",
    header: "AP",
    render: (row) => (
      <span className="erp-finance-receivable-cell">
        <strong>{row.payableNo}</strong>
        <small>{row.supplierCode ? `${row.supplierCode} / ${row.supplierName}` : row.supplierName}</small>
      </span>
    ),
    width: "260px"
  },
  {
    key: "status",
    header: "Status",
    render: (row) => (
      <StatusChip tone={supplierPayableStatusTone(row.status)}>{formatSupplierPayableStatus(row.status)}</StatusChip>
    ),
    width: "170px"
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

const lineColumns: DataTableColumn<SupplierPayableLine>[] = [
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

const invoiceLineColumns: DataTableColumn<SupplierInvoiceLine>[] = [
  {
    key: "description",
    header: "Nguồn",
    render: (row) => (
      <span className="erp-finance-receivable-cell">
        <strong>{row.description}</strong>
        <small>{sourceDocumentLabel(row.sourceDocument)}</small>
      </span>
    )
  },
  {
    key: "amount",
    header: "Số tiền",
    render: (row) => formatFinanceMoney(row.amount),
    align: "right",
    width: "140px"
  }
];

export function SupplierPayablesPanel() {
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState<SupplierPayableStatusFilter>("");
  const [selectedPayableId, setSelectedPayableId] = useState("ap-supplier-260430-0001");
  const [localPayables, setLocalPayables] = useState<SupplierPayable[]>([]);
  const [detailById, setDetailById] = useState<Record<string, SupplierPayable>>({});
  const [detailLoadingId, setDetailLoadingId] = useState("");
  const [paymentAmount, setPaymentAmount] = useState("");
  const [invoiceNo, setInvoiceNo] = useState("");
  const [invoiceAmount, setInvoiceAmount] = useState("");
  const [invoiceDate, setInvoiceDate] = useState(defaultFinanceDate());
  const [localInvoices, setLocalInvoices] = useState<SupplierInvoice[]>([]);
  const [rejectReason, setRejectReason] = useState("Supplier invoice mismatch");
  const [voidReason, setVoidReason] = useState("Duplicate supplier invoice");
  const [busyAction, setBusyAction] = useState("");
  const [feedback, setFeedback] = useState<{ tone: StatusTone; message: string } | null>(null);
  const query = useMemo<SupplierPayableQuery>(
    () => ({
      search: search || undefined,
      status: status || undefined
    }),
    [search, status]
  );
  const { payables, loading, error } = useSupplierPayables(query);
  const visiblePayables = useMemo(() => mergePayables(localPayables, payables, query), [localPayables, payables, query]);
  const selectedListPayable =
    visiblePayables.find((payable) => payable.id === selectedPayableId) ?? visiblePayables[0] ?? null;
  const selectedPayable = selectedListPayable ? detailById[selectedListPayable.id] ?? selectedListPayable : null;
  const invoiceQuery = useMemo<SupplierInvoiceQuery>(
    () => ({
      payableId: selectedPayable?.id
    }),
    [selectedPayable?.id]
  );
  const { invoices, loading: invoicesLoading, error: invoicesError } = useSupplierInvoices(invoiceQuery);
  const visibleInvoices = useMemo(
    () => mergeSupplierInvoices(localInvoices, invoices, invoiceQuery),
    [invoiceQuery, invoices, localInvoices]
  );
  const selectedInvoice = visibleInvoices[0] ?? null;
  const totals = useMemo(() => summarizePayables(visiblePayables), [visiblePayables]);

  useEffect(() => {
    const initialSearch = initialSupplierPayableSearch();
    if (initialSearch) {
      setSearch(initialSearch);
    }
  }, []);

  useEffect(() => {
    if (visiblePayables.length > 0 && !visiblePayables.some((payable) => payable.id === selectedPayableId)) {
      setSelectedPayableId(visiblePayables[0].id);
    }
  }, [selectedPayableId, visiblePayables]);

  useEffect(() => {
    if (!selectedListPayable || detailById[selectedListPayable.id] || selectedListPayable.lines.length > 0) {
      return;
    }

    let active = true;
    setDetailLoadingId(selectedListPayable.id);
    getSupplierPayable(selectedListPayable.id)
      .then((detail) => {
        if (active) {
          setDetailById((current) => ({ ...current, [detail.id]: detail }));
        }
      })
      .catch((reason) => {
        if (active) {
          setFeedback({
            tone: "danger",
            message: reason instanceof Error ? reason.message : "AP detail could not be loaded"
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
  }, [detailById, selectedListPayable]);

  useEffect(() => {
    if (selectedPayable && paymentAmount === "") {
      setPaymentAmount(selectedPayable.outstandingAmount);
    }
  }, [paymentAmount, selectedPayable]);

  useEffect(() => {
    if (!selectedPayable) {
      return;
    }
    setInvoiceAmount(selectedPayable.totalAmount);
    setInvoiceNo((current) => current || `INV-${selectedPayable.payableNo}`);
  }, [selectedPayable?.id]);

  function handleSelect(payable: SupplierPayable) {
    setSelectedPayableId(payable.id);
    setPaymentAmount(payable.outstandingAmount);
    setInvoiceAmount(payable.totalAmount);
    setInvoiceNo(`INV-${payable.payableNo}`);
    setFeedback(null);
  }

  async function handleApprovePayment() {
    if (!selectedPayable || busyAction) {
      return;
    }

    await runAction("approve", async () => {
      const result = await approveSupplierPayablePayment(selectedPayable.id);
      applyActionResult(result.supplierPayable);
      setPaymentAmount(result.supplierPayable.outstandingAmount);
      setFeedback({
        tone: "success",
        message: `${result.supplierPayable.payableNo} payment approved`
      });
    });
  }

  async function handleRequestPayment() {
    if (!selectedPayable || busyAction) {
      return;
    }

    await runAction("request", async () => {
      const result = await requestSupplierPayablePayment(selectedPayable.id);
      applyActionResult(result.supplierPayable);
      setFeedback({
        tone: "info",
        message: `${result.supplierPayable.payableNo} payment requested`
      });
    });
  }

  async function handleRejectPayment(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedPayable || busyAction) {
      return;
    }

    await runAction("reject", async () => {
      const result = await rejectSupplierPayablePayment(selectedPayable.id, rejectReason);
      applyActionResult(result.supplierPayable);
      setFeedback({
        tone: "warning",
        message: `${result.supplierPayable.payableNo} payment rejected`
      });
    });
  }

  async function handleRecordPayment(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedPayable || busyAction) {
      return;
    }

    await runAction("payment", async () => {
      const result = await recordSupplierPayablePayment(selectedPayable.id, paymentAmount);
      applyActionResult(result.supplierPayable);
      setPaymentAmount(result.supplierPayable.outstandingAmount);
      setFeedback({
        tone: "success",
        message: `${result.supplierPayable.payableNo} payment recorded`
      });
    });
  }

  async function handleVoid(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedPayable || busyAction) {
      return;
    }

    await runAction("void", async () => {
      const result = await voidSupplierPayable(selectedPayable.id, voidReason);
      applyActionResult(result.supplierPayable);
      setFeedback({
        tone: "danger",
        message: `${result.supplierPayable.payableNo} voided`
      });
    });
  }

  async function handleCreateSupplierInvoice(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedPayable || busyAction) {
      return;
    }

    await runAction("invoice", async () => {
      const invoice = await createSupplierInvoice({
        invoiceNo: invoiceNo.trim(),
        payableId: selectedPayable.id,
        invoiceDate,
        invoiceAmount,
        currencyCode: selectedPayable.currencyCode
      });
      setLocalInvoices((current) => [invoice, ...current.filter((candidate) => candidate.id !== invoice.id)]);
      setInvoiceNo(`INV-${selectedPayable.payableNo}`);
      setInvoiceAmount(selectedPayable.totalAmount);
      setFeedback({
        tone: invoice.status === "matched" ? "success" : "warning",
        message:
          invoice.status === "matched"
            ? `${invoice.invoiceNo} đã khớp với ${invoice.payableNo}`
            : `${invoice.invoiceNo} lệch ${formatFinanceMoney(invoice.varianceAmount, invoice.currencyCode)}`
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
        message: reason instanceof Error ? reason.message : "AP action failed"
      });
    } finally {
      setBusyAction("");
    }
  }

  function applyActionResult(payable: SupplierPayable) {
    setLocalPayables((current) => [payable, ...current.filter((candidate) => candidate.id !== payable.id)]);
    setDetailById((current) => ({ ...current, [payable.id]: payable }));
    setSelectedPayableId(payable.id);
  }

  return (
    <section className="erp-finance-section" id="supplier-payables">
      <div className="erp-section-header">
        <div>
          <h2 className="erp-section-title">Supplier payables</h2>
          <p className="erp-section-description">Supplier AP detail, payment approval, payment recording, and void control</p>
        </div>
        <StatusChip tone={loading ? "warning" : "info"}>{visiblePayables.length} records</StatusChip>
      </div>

      <section className="erp-kpi-grid erp-finance-kpis">
        <FinanceAPKPI label="Open AP" value={totals.openCount} tone={totals.openCount > 0 ? "warning" : "normal"} />
        <FinanceAPKPI label="Outstanding" value={formatFinanceMoney(totals.outstandingAmount)} tone="warning" />
        <FinanceAPKPI label="Paid" value={formatFinanceMoney(totals.paidAmount)} tone="success" />
        <FinanceAPKPI label="Approved" value={totals.approvedCount} tone={totals.approvedCount > 0 ? "info" : "normal"} />
      </section>

      <section className="erp-finance-toolbar" aria-label="AP filters">
        <label className="erp-field">
          <span>Search</span>
          <input
            className="erp-input"
            type="search"
            value={search}
            placeholder="AP no, supplier, source document"
            onChange={(event) => setSearch(event.target.value)}
          />
        </label>
        <label className="erp-field">
          <span>Status</span>
          <select className="erp-input" value={status} onChange={(event) => setStatus(event.target.value as SupplierPayableStatusFilter)}>
            {supplierPayableStatusOptions.map((option) => (
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
            columns={payableColumns(handleSelect)}
            rows={visiblePayables}
            getRowKey={(row) => row.id}
            loading={loading}
            error={error?.message}
            emptyState={<EmptyState title="No supplier payables" description="Try another status or search term." />}
          />
        </section>

        <aside className="erp-finance-side">
          <section className="erp-card erp-card--padded">
            <div className="erp-section-header">
              <h2 className="erp-section-title">AP detail</h2>
              {selectedPayable ? (
                <StatusChip tone={supplierPayableStatusTone(selectedPayable.status)}>
                  {formatSupplierPayableStatus(selectedPayable.status)}
                </StatusChip>
              ) : null}
            </div>

            {selectedPayable ? (
              <div className="erp-finance-detail">
                <strong>{selectedPayable.payableNo}</strong>
                <span>{selectedPayable.supplierName}</span>
                <dl className="erp-finance-detail-list">
                  <div>
                    <dt>Supplier</dt>
                    <dd>{selectedPayable.supplierCode ?? selectedPayable.supplierId}</dd>
                  </div>
                  <div>
                    <dt>Source</dt>
                    <dd>{selectedPayable.sourceDocument ? sourceDocumentLabel(selectedPayable.sourceDocument) : "-"}</dd>
                  </div>
                  <div>
                    <dt>Due date</dt>
                    <dd>{formatFinanceDate(selectedPayable.dueDate)}</dd>
                  </div>
                  <div>
                    <dt>Total</dt>
                    <dd>{formatFinanceMoney(selectedPayable.totalAmount, selectedPayable.currencyCode)}</dd>
                  </div>
                  <div>
                    <dt>Paid</dt>
                    <dd>{formatFinanceMoney(selectedPayable.paidAmount, selectedPayable.currencyCode)}</dd>
                  </div>
                  <div>
                    <dt>Outstanding</dt>
                    <dd>{formatFinanceMoney(selectedPayable.outstandingAmount, selectedPayable.currencyCode)}</dd>
                  </div>
                </dl>
                {selectedPayable.paymentApprovedBy ? (
                  <p className="erp-finance-note">
                    Payment approved by {selectedPayable.paymentApprovedBy} on {formatFinanceDate(selectedPayable.paymentApprovedAt)}
                  </p>
                ) : null}
                {selectedPayable.paymentRequestedBy && selectedPayable.status === "payment_requested" ? (
                  <p className="erp-finance-note">
                    Payment requested by {selectedPayable.paymentRequestedBy} on {formatFinanceDate(selectedPayable.paymentRequestedAt)}
                  </p>
                ) : null}
                {selectedPayable.paymentRejectReason ? (
                  <p className="erp-finance-note">
                    Payment rejected by {selectedPayable.paymentRejectedBy ?? "finance"}: {selectedPayable.paymentRejectReason}
                  </p>
                ) : null}
                {selectedPayable.voidReason ? <p className="erp-finance-note">{selectedPayable.voidReason}</p> : null}
              </div>
            ) : (
              <EmptyState title="No AP selected" />
            )}
          </section>

          <section className="erp-card erp-card--padded">
            <div className="erp-section-header">
              <h2 className="erp-section-title">Lines</h2>
              <StatusChip tone={detailLoadingId ? "warning" : "normal"}>
                {detailLoadingId ? "Loading" : `${selectedPayable?.lines.length ?? 0} lines`}
              </StatusChip>
            </div>
            <DataTable
              columns={lineColumns}
              rows={selectedPayable?.lines ?? []}
              getRowKey={(row) => row.id}
              emptyState={<EmptyState title="No AP line detail" description="Open AP detail could not load line allocation." />}
            />
          </section>

          <section className="erp-card erp-card--padded" id="supplier-invoice-match">
            <div className="erp-section-header">
              <div>
                <h2 className="erp-section-title">Hóa đơn NCC</h2>
                <p className="erp-section-description">Đối chiếu hóa đơn với AP trước khi thanh toán</p>
              </div>
              <StatusChip tone={invoicesLoading ? "warning" : selectedInvoice ? supplierInvoiceStatusTone(selectedInvoice.status) : "normal"}>
                {invoicesLoading ? "Đang tải" : selectedInvoice ? formatSupplierInvoiceStatus(selectedInvoice.status) : "Chưa có"}
              </StatusChip>
            </div>

            {invoicesError ? <p className="erp-finance-note">{invoicesError.message}</p> : null}

            {selectedInvoice ? (
              <div className="erp-finance-detail">
                <strong>{selectedInvoice.invoiceNo}</strong>
                <span>{selectedInvoice.payableNo}</span>
                <dl className="erp-finance-detail-list">
                  <div>
                    <dt>Ngày hóa đơn</dt>
                    <dd>{formatFinanceDate(selectedInvoice.invoiceDate)}</dd>
                  </div>
                  <div>
                    <dt>Số tiền hóa đơn</dt>
                    <dd>{formatFinanceMoney(selectedInvoice.invoiceAmount, selectedInvoice.currencyCode)}</dd>
                  </div>
                  <div>
                    <dt>Số tiền AP</dt>
                    <dd>{formatFinanceMoney(selectedInvoice.expectedAmount, selectedInvoice.currencyCode)}</dd>
                  </div>
                  <div>
                    <dt>Chênh lệch</dt>
                    <dd>{formatFinanceMoney(selectedInvoice.varianceAmount, selectedInvoice.currencyCode)}</dd>
                  </div>
                  <div>
                    <dt>Nguồn nhận hàng</dt>
                    <dd>{sourceDocumentLabel(selectedInvoice.sourceDocument)}</dd>
                  </div>
                </dl>
                <DataTable
                  columns={invoiceLineColumns}
                  rows={selectedInvoice.lines}
                  getRowKey={(row) => row.id}
                  emptyState={<EmptyState title="Chưa có dòng hóa đơn" />}
                />
              </div>
            ) : (
              <EmptyState title="Chưa có hóa đơn NCC" description="Tạo hóa đơn từ AP đang chọn để kiểm tra khớp/lệch." />
            )}

            <form className="erp-finance-action-form" onSubmit={handleCreateSupplierInvoice}>
              <label className="erp-field">
                <span>Số hóa đơn</span>
                <input
                  className="erp-input"
                  type="text"
                  value={invoiceNo}
                  disabled={!selectedPayable || busyAction !== ""}
                  onChange={(event) => setInvoiceNo(event.target.value)}
                />
              </label>
              <label className="erp-field">
                <span>Ngày hóa đơn</span>
                <input
                  className="erp-input"
                  type="date"
                  value={invoiceDate}
                  disabled={!selectedPayable || busyAction !== ""}
                  onChange={(event) => setInvoiceDate(event.target.value)}
                />
              </label>
              <label className="erp-field">
                <span>Số tiền hóa đơn</span>
                <input
                  className="erp-input"
                  type="text"
                  inputMode="decimal"
                  value={invoiceAmount}
                  disabled={!selectedPayable || busyAction !== ""}
                  onChange={(event) => setInvoiceAmount(event.target.value)}
                />
              </label>
              <button
                className="erp-button erp-button--secondary"
                type="submit"
                disabled={!selectedPayable || invoiceNo.trim() === "" || invoiceAmount.trim() === "" || busyAction !== ""}
              >
                Tạo hóa đơn NCC
              </button>
            </form>
          </section>

          <section className="erp-card erp-card--padded" id="ap-actions">
            <div className="erp-section-header">
              <h2 className="erp-section-title">Payment action</h2>
              <StatusChip tone={canRecordSupplierPayablePayment(selectedPayable) ? "success" : "normal"}>AP</StatusChip>
            </div>

            <div className="erp-finance-action-row">
              <button
                className="erp-button erp-button--secondary"
                type="button"
                disabled={!canRequestSupplierPayablePayment(selectedPayable) || busyAction !== ""}
                onClick={() => void handleRequestPayment()}
              >
                Request payment
              </button>
              <button
                className="erp-button erp-button--primary"
                type="button"
                disabled={!canApproveSupplierPayablePayment(selectedPayable) || busyAction !== ""}
                onClick={() => void handleApprovePayment()}
              >
                Approve payment
              </button>
            </div>

            <form className="erp-finance-action-form" onSubmit={handleRejectPayment}>
              <label className="erp-field">
                <span>Reject reason</span>
                <input
                  className="erp-input"
                  type="text"
                  value={rejectReason}
                  disabled={!canRejectSupplierPayablePayment(selectedPayable) || busyAction !== ""}
                  onChange={(event) => setRejectReason(event.target.value)}
                />
              </label>
              <button
                className="erp-button erp-button--danger"
                type="submit"
                disabled={!canRejectSupplierPayablePayment(selectedPayable) || rejectReason.trim() === "" || busyAction !== ""}
              >
                Reject payment
              </button>
            </form>

            <form className="erp-finance-action-form" onSubmit={handleRecordPayment}>
              <label className="erp-field">
                <span>Payment amount</span>
                <input
                  className="erp-input"
                  type="text"
                  inputMode="decimal"
                  value={paymentAmount}
                  disabled={!canRecordSupplierPayablePayment(selectedPayable) || busyAction !== ""}
                  onChange={(event) => setPaymentAmount(event.target.value)}
                />
              </label>
              <button
                className="erp-button erp-button--secondary"
                type="submit"
                disabled={!canRecordSupplierPayablePayment(selectedPayable) || busyAction !== ""}
              >
                Record payment
              </button>
            </form>

            <form className="erp-finance-action-form" onSubmit={handleVoid}>
              <label className="erp-field">
                <span>Void reason</span>
                <input
                  className="erp-input"
                  type="text"
                  value={voidReason}
                  disabled={!canVoidSupplierPayable(selectedPayable) || busyAction !== ""}
                  onChange={(event) => setVoidReason(event.target.value)}
                />
              </label>
              <button
                className="erp-button erp-button--danger"
                type="submit"
                disabled={!canVoidSupplierPayable(selectedPayable) || busyAction !== ""}
              >
                Void AP
              </button>
            </form>
          </section>
        </aside>
      </section>
    </section>
  );
}

function FinanceAPKPI({ label, value, tone }: { label: string; value: string | number; tone: StatusTone }) {
  return (
    <div className="erp-card erp-card--padded erp-kpi-card">
      <span className="erp-kpi-label">{label}</span>
      <strong className="erp-kpi-value">{value}</strong>
      <StatusChip tone={tone}>{label}</StatusChip>
    </div>
  );
}

function mergePayables(localPayables: SupplierPayable[], payables: SupplierPayable[], query: SupplierPayableQuery) {
  const localMatches = localPayables.filter((payable) => matchesPayableQuery(payable, query));
  const localIds = new Set(localMatches.map((payable) => payable.id));

  return [...localMatches, ...payables.filter((payable) => !localIds.has(payable.id))];
}

function mergeSupplierInvoices(
  localInvoices: SupplierInvoice[],
  invoices: SupplierInvoice[],
  query: SupplierInvoiceQuery
) {
  const localMatches = localInvoices.filter((invoice) => matchesSupplierInvoiceQuery(invoice, query));
  const localIds = new Set(localMatches.map((invoice) => invoice.id));

  return [...localMatches, ...invoices.filter((invoice) => !localIds.has(invoice.id))];
}

function matchesPayableQuery(payable: SupplierPayable, query: SupplierPayableQuery) {
  const search = query.search?.trim().toLowerCase();
  return (
    (!query.status || payable.status === query.status) &&
    (!search ||
      [
        payable.payableNo,
        payable.supplierCode,
        payable.supplierName,
        payable.sourceDocument?.id,
        payable.sourceDocument?.no,
        ...payable.lines.flatMap((line) => [line.description, line.sourceDocument.id, line.sourceDocument.no])
      ]
        .filter(Boolean)
        .some((value) => String(value).toLowerCase().includes(search)))
  );
}

function matchesSupplierInvoiceQuery(invoice: SupplierInvoice, query: SupplierInvoiceQuery) {
  const search = query.search?.trim().toLowerCase();
  return (
    (!query.status || invoice.status === query.status) &&
    (!query.supplierId || invoice.supplierId === query.supplierId) &&
    (!query.payableId || invoice.payableId === query.payableId) &&
    (!search ||
      [
        invoice.invoiceNo,
        invoice.supplierCode,
        invoice.supplierName,
        invoice.payableNo,
        invoice.sourceDocument.id,
        invoice.sourceDocument.no,
        ...invoice.lines.flatMap((line) => [line.description, line.sourceDocument.id, line.sourceDocument.no])
      ]
        .filter(Boolean)
        .some((value) => String(value).toLowerCase().includes(search)))
  );
}

function initialSupplierPayableSearch() {
  if (typeof window === "undefined") {
    return "";
  }

  return new URLSearchParams(window.location.search).get("ap_q") ?? "";
}

function summarizePayables(payables: SupplierPayable[]) {
  return payables.reduce(
    (summary, payable) => ({
      openCount: summary.openCount + (["open", "payment_requested", "payment_approved", "partially_paid"].includes(payable.status) ? 1 : 0),
      approvedCount: summary.approvedCount + (payable.status === "payment_approved" ? 1 : 0),
      paidAmount: summary.paidAmount + Number(payable.paidAmount),
      outstandingAmount: summary.outstandingAmount + Number(payable.outstandingAmount)
    }),
    { openCount: 0, approvedCount: 0, paidAmount: 0, outstandingAmount: 0 }
  );
}

function sourceDocumentLabel(source: { type: string; no?: string; id?: string }) {
  return `${source.type.replaceAll("_", " ")} / ${source.no ?? source.id ?? "-"}`;
}

function defaultFinanceDate() {
  return new Date().toISOString().slice(0, 10);
}
