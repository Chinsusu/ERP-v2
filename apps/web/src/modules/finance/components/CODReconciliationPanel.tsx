"use client";

import { useEffect, useMemo, useState, type FormEvent } from "react";
import { DataTable, EmptyState, StatusChip, type DataTableColumn, type StatusTone } from "@/shared/design-system/components";
import { useCODRemittances } from "../hooks/useCODRemittances";
import {
  approveCODRemittance,
  canApproveCODRemittance,
  canCloseCODRemittance,
  canMatchCODRemittance,
  canRecordCODDiscrepancy,
  canSubmitCODRemittance,
  closeCODRemittance,
  codLineMatchStatusTone,
  codRemittanceStatusOptions,
  codRemittanceStatusTone,
  formatCODStatus,
  getCODRemittance,
  matchCODRemittance,
  recordCODRemittanceDiscrepancy,
  submitCODRemittance
} from "../services/codRemittanceService";
import { formatFinanceDate, formatFinanceMoney } from "../services/customerReceivableService";
import type { CODDiscrepancy, CODRemittance, CODRemittanceLine, CODRemittanceQuery, CODRemittanceStatus } from "../types";

type CODStatusFilter = "" | CODRemittanceStatus;

const remittanceColumns = (onSelect: (remittance: CODRemittance) => void): DataTableColumn<CODRemittance>[] => [
  {
    key: "remittance",
    header: "COD batch",
    render: (row) => (
      <span className="erp-finance-receivable-cell">
        <strong>{row.remittanceNo}</strong>
        <small>{row.carrierCode ? `${row.carrierCode} / ${row.carrierName}` : row.carrierName}</small>
      </span>
    ),
    width: "240px"
  },
  {
    key: "status",
    header: "Status",
    render: (row) => <StatusChip tone={codRemittanceStatusTone(row.status)}>{formatCODStatus(row.status)}</StatusChip>,
    width: "140px"
  },
  {
    key: "date",
    header: "Date",
    render: (row) => formatFinanceDate(row.businessDate),
    width: "120px"
  },
  {
    key: "expected",
    header: "Expected",
    render: (row) => formatFinanceMoney(row.expectedAmount, row.currencyCode),
    align: "right",
    width: "150px"
  },
  {
    key: "remitted",
    header: "Remitted",
    render: (row) => formatFinanceMoney(row.remittedAmount, row.currencyCode),
    align: "right",
    width: "150px"
  },
  {
    key: "discrepancy",
    header: "Diff",
    render: (row) => formatFinanceMoney(row.discrepancyAmount, row.currencyCode),
    align: "right",
    width: "140px"
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

const lineColumns: DataTableColumn<CODRemittanceLine>[] = [
  {
    key: "shipment",
    header: "Shipment",
    render: (row) => (
      <span className="erp-finance-receivable-cell">
        <strong>{row.trackingNo}</strong>
        <small>{row.customerName ?? row.receivableNo}</small>
      </span>
    )
  },
  {
    key: "status",
    header: "Match",
    render: (row) => <StatusChip tone={codLineMatchStatusTone(row.matchStatus)}>{formatCODStatus(row.matchStatus)}</StatusChip>,
    width: "130px"
  },
  {
    key: "expected",
    header: "Expected",
    render: (row) => formatFinanceMoney(row.expectedAmount),
    align: "right",
    width: "140px"
  },
  {
    key: "remitted",
    header: "Remitted",
    render: (row) => formatFinanceMoney(row.remittedAmount),
    align: "right",
    width: "140px"
  },
  {
    key: "discrepancy",
    header: "Diff",
    render: (row) => formatFinanceMoney(row.discrepancyAmount),
    align: "right",
    width: "130px"
  }
];

const discrepancyColumns: DataTableColumn<CODDiscrepancy>[] = [
  {
    key: "reason",
    header: "Trace",
    render: (row) => (
      <span className="erp-finance-receivable-cell">
        <strong>{formatCODStatus(row.type)}</strong>
        <small>{row.reason}</small>
      </span>
    )
  },
  {
    key: "status",
    header: "Status",
    render: (row) => <StatusChip tone={row.status === "resolved" ? "success" : "warning"}>{formatCODStatus(row.status)}</StatusChip>,
    width: "120px"
  },
  {
    key: "amount",
    header: "Amount",
    render: (row) => formatFinanceMoney(row.amount),
    align: "right",
    width: "140px"
  }
];

export function CODReconciliationPanel() {
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState<CODStatusFilter>("");
  const [selectedRemittanceId, setSelectedRemittanceId] = useState("cod-remit-260430-0001");
  const [localRemittances, setLocalRemittances] = useState<CODRemittance[]>([]);
  const [detailById, setDetailById] = useState<Record<string, CODRemittance>>({});
  const [detailLoadingId, setDetailLoadingId] = useState("");
  const [discrepancyLineId, setDiscrepancyLineId] = useState("");
  const [discrepancyReason, setDiscrepancyReason] = useState("Carrier short remitted COD amount");
  const [discrepancyOwnerId, setDiscrepancyOwnerId] = useState("finance-user");
  const [busyAction, setBusyAction] = useState("");
  const [feedback, setFeedback] = useState<{ tone: StatusTone; message: string } | null>(null);
  const query = useMemo<CODRemittanceQuery>(
    () => ({
      search: search || undefined,
      status: status || undefined
    }),
    [search, status]
  );
  const { remittances, loading, error } = useCODRemittances(query);
  const visibleRemittances = useMemo(
    () => mergeRemittances(localRemittances, remittances, query),
    [localRemittances, query, remittances]
  );
  const selectedListRemittance =
    visibleRemittances.find((remittance) => remittance.id === selectedRemittanceId) ?? visibleRemittances[0] ?? null;
  const selectedRemittance = selectedListRemittance ? detailById[selectedListRemittance.id] ?? selectedListRemittance : null;
  const discrepantLines = selectedRemittance?.lines.filter((line) => Number(line.discrepancyAmount) !== 0) ?? [];
  const totals = useMemo(() => summarizeRemittances(visibleRemittances), [visibleRemittances]);

  useEffect(() => {
    if (
      visibleRemittances.length > 0 &&
      !visibleRemittances.some((remittance) => remittance.id === selectedRemittanceId)
    ) {
      setSelectedRemittanceId(visibleRemittances[0].id);
    }
  }, [selectedRemittanceId, visibleRemittances]);

  useEffect(() => {
    if (!selectedListRemittance || detailById[selectedListRemittance.id] || selectedListRemittance.lines.length > 0) {
      return;
    }

    let active = true;
    setDetailLoadingId(selectedListRemittance.id);
    getCODRemittance(selectedListRemittance.id)
      .then((detail) => {
        if (active) {
          setDetailById((current) => ({ ...current, [detail.id]: detail }));
        }
      })
      .catch((reason) => {
        if (active) {
          setFeedback({
            tone: "danger",
            message: reason instanceof Error ? reason.message : "COD detail could not be loaded"
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
  }, [detailById, selectedListRemittance]);

  useEffect(() => {
    if (discrepantLines.length > 0 && !discrepantLines.some((line) => line.id === discrepancyLineId)) {
      setDiscrepancyLineId(discrepantLines[0].id);
    }
  }, [discrepancyLineId, discrepantLines]);

  function handleSelect(remittance: CODRemittance) {
    setSelectedRemittanceId(remittance.id);
    setFeedback(null);
  }

  async function handleRecordDiscrepancy(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedRemittance || busyAction || !discrepancyLineId) {
      return;
    }

    await runAction("record-discrepancy", async () => {
      const result = await recordCODRemittanceDiscrepancy(selectedRemittance.id, {
        lineId: discrepancyLineId,
        reason: discrepancyReason,
        ownerId: discrepancyOwnerId
      });
      applyActionResult(result.codRemittance);
      setFeedback({
        tone: "warning",
        message: `${result.codRemittance.remittanceNo} discrepancy traced`
      });
    });
  }

  async function handleStatusAction(action: "match" | "submit" | "approve" | "close") {
    if (!selectedRemittance || busyAction) {
      return;
    }

    await runAction(action, async () => {
      const result = await runStatusAction(selectedRemittance.id, action);
      applyActionResult(result.codRemittance);
      setFeedback({
        tone: "success",
        message: `${result.codRemittance.remittanceNo} moved to ${formatCODStatus(result.currentStatus)}`
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
        message: reason instanceof Error ? reason.message : "COD action failed"
      });
    } finally {
      setBusyAction("");
    }
  }

  function applyActionResult(remittance: CODRemittance) {
    setLocalRemittances((current) => [remittance, ...current.filter((candidate) => candidate.id !== remittance.id)]);
    setDetailById((current) => ({ ...current, [remittance.id]: remittance }));
    setSelectedRemittanceId(remittance.id);
  }

  return (
    <section className="erp-finance-section" id="cod-reconciliation">
      <div className="erp-section-header">
        <div>
          <h2 className="erp-section-title">COD reconciliation</h2>
          <p className="erp-section-description">Expected COD, carrier remittance, discrepancy trace, approval, and close</p>
        </div>
        <StatusChip tone={loading ? "warning" : "info"}>{visibleRemittances.length} batches</StatusChip>
      </div>

      <section className="erp-kpi-grid erp-finance-kpis">
        <FinanceCODKPI label="Expected" value={formatFinanceMoney(totals.expectedAmount)} tone="info" />
        <FinanceCODKPI label="Remitted" value={formatFinanceMoney(totals.remittedAmount)} tone="success" />
        <FinanceCODKPI label="Diff" value={formatFinanceMoney(totals.discrepancyAmount)} tone={totals.discrepancyAmount === 0 ? "normal" : "warning"} />
        <FinanceCODKPI label="Open traces" value={totals.openDiscrepancies} tone={totals.openDiscrepancies > 0 ? "warning" : "normal"} />
      </section>

      <section className="erp-finance-toolbar" aria-label="COD filters">
        <label className="erp-field">
          <span>Search</span>
          <input
            className="erp-input"
            type="search"
            value={search}
            placeholder="COD batch, carrier, tracking, AR"
            onChange={(event) => setSearch(event.target.value)}
          />
        </label>
        <label className="erp-field">
          <span>Status</span>
          <select className="erp-input" value={status} onChange={(event) => setStatus(event.target.value as CODStatusFilter)}>
            {codRemittanceStatusOptions.map((option) => (
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
            columns={remittanceColumns(handleSelect)}
            rows={visibleRemittances}
            getRowKey={(row) => row.id}
            loading={loading}
            error={error?.message}
            emptyState={<EmptyState title="No COD remittances" description="Try another status or search term." />}
          />
        </section>

        <aside className="erp-finance-side">
          <section className="erp-card erp-card--padded">
            <div className="erp-section-header">
              <h2 className="erp-section-title">Batch detail</h2>
              {selectedRemittance ? (
                <StatusChip tone={codRemittanceStatusTone(selectedRemittance.status)}>
                  {formatCODStatus(selectedRemittance.status)}
                </StatusChip>
              ) : null}
            </div>

            {selectedRemittance ? (
              <div className="erp-finance-detail">
                <strong>{selectedRemittance.remittanceNo}</strong>
                <span>{selectedRemittance.carrierCode ? `${selectedRemittance.carrierCode} / ${selectedRemittance.carrierName}` : selectedRemittance.carrierName}</span>
                <dl className="erp-finance-detail-list">
                  <div>
                    <dt>Business date</dt>
                    <dd>{formatFinanceDate(selectedRemittance.businessDate)}</dd>
                  </div>
                  <div>
                    <dt>Lines</dt>
                    <dd>{selectedRemittance.lineCount}</dd>
                  </div>
                  <div>
                    <dt>Expected</dt>
                    <dd>{formatFinanceMoney(selectedRemittance.expectedAmount, selectedRemittance.currencyCode)}</dd>
                  </div>
                  <div>
                    <dt>Remitted</dt>
                    <dd>{formatFinanceMoney(selectedRemittance.remittedAmount, selectedRemittance.currencyCode)}</dd>
                  </div>
                  <div>
                    <dt>Discrepancy</dt>
                    <dd>{formatFinanceMoney(selectedRemittance.discrepancyAmount, selectedRemittance.currencyCode)}</dd>
                  </div>
                  <div>
                    <dt>Trace records</dt>
                    <dd>{selectedRemittance.discrepancyCount}</dd>
                  </div>
                </dl>
              </div>
            ) : (
              <EmptyState title="No COD batch selected" />
            )}
          </section>

          <section className="erp-card erp-card--padded">
            <div className="erp-section-header">
              <h2 className="erp-section-title">Shipment lines</h2>
              <StatusChip tone={detailLoadingId ? "warning" : "normal"}>
                {detailLoadingId ? "Loading" : `${selectedRemittance?.lines.length ?? 0} lines`}
              </StatusChip>
            </div>
            <DataTable
              columns={lineColumns}
              rows={selectedRemittance?.lines ?? []}
              getRowKey={(row) => row.id}
              emptyState={<EmptyState title="No COD line detail" description="Open batch detail could not load line allocation." />}
            />
          </section>

          <section className="erp-card erp-card--padded" id="cod-actions">
            <div className="erp-section-header">
              <h2 className="erp-section-title">Reconciliation action</h2>
              <StatusChip tone={canSubmitCODRemittance(selectedRemittance) ? "success" : "normal"}>COD</StatusChip>
            </div>

            <form className="erp-finance-action-form" onSubmit={handleRecordDiscrepancy}>
              <label className="erp-field">
                <span>Discrepancy line</span>
                <select
                  className="erp-input"
                  value={discrepancyLineId}
                  disabled={!canRecordCODDiscrepancy(selectedRemittance) || discrepantLines.length === 0 || busyAction !== ""}
                  onChange={(event) => setDiscrepancyLineId(event.target.value)}
                >
                  {discrepantLines.map((line) => (
                    <option key={line.id} value={line.id}>
                      {line.trackingNo} / {formatFinanceMoney(line.discrepancyAmount)}
                    </option>
                  ))}
                </select>
              </label>
              <label className="erp-field">
                <span>Reason</span>
                <input
                  className="erp-input"
                  type="text"
                  value={discrepancyReason}
                  disabled={!canRecordCODDiscrepancy(selectedRemittance) || discrepantLines.length === 0 || busyAction !== ""}
                  onChange={(event) => setDiscrepancyReason(event.target.value)}
                />
              </label>
              <label className="erp-field">
                <span>Owner</span>
                <input
                  className="erp-input"
                  type="text"
                  value={discrepancyOwnerId}
                  disabled={!canRecordCODDiscrepancy(selectedRemittance) || discrepantLines.length === 0 || busyAction !== ""}
                  onChange={(event) => setDiscrepancyOwnerId(event.target.value)}
                />
              </label>
              <button
                className="erp-button erp-button--secondary"
                type="submit"
                disabled={!canRecordCODDiscrepancy(selectedRemittance) || discrepantLines.length === 0 || busyAction !== ""}
              >
                Record discrepancy
              </button>
            </form>

            <div className="erp-finance-action-row">
              <button
                className="erp-button erp-button--secondary"
                type="button"
                disabled={!canMatchCODRemittance(selectedRemittance) || busyAction !== ""}
                onClick={() => void handleStatusAction("match")}
              >
                Match
              </button>
              <button
                className="erp-button erp-button--primary"
                type="button"
                disabled={!canSubmitCODRemittance(selectedRemittance) || busyAction !== ""}
                onClick={() => void handleStatusAction("submit")}
              >
                Submit
              </button>
              <button
                className="erp-button erp-button--secondary"
                type="button"
                disabled={!canApproveCODRemittance(selectedRemittance) || busyAction !== ""}
                onClick={() => void handleStatusAction("approve")}
              >
                Approve
              </button>
              <button
                className="erp-button erp-button--secondary"
                type="button"
                disabled={!canCloseCODRemittance(selectedRemittance) || busyAction !== ""}
                onClick={() => void handleStatusAction("close")}
              >
                Close
              </button>
            </div>
          </section>

          <section className="erp-card erp-card--padded">
            <div className="erp-section-header">
              <h2 className="erp-section-title">Discrepancy trace</h2>
              <StatusChip tone={(selectedRemittance?.discrepancyCount ?? 0) > 0 ? "warning" : "normal"}>
                {selectedRemittance?.discrepancyCount ?? 0} traces
              </StatusChip>
            </div>
            <DataTable
              columns={discrepancyColumns}
              rows={selectedRemittance?.discrepancies ?? []}
              getRowKey={(row) => row.id}
              emptyState={<EmptyState title="No discrepancy trace" description="Matched batches do not need trace records." />}
            />
          </section>
        </aside>
      </section>
    </section>
  );
}

function FinanceCODKPI({ label, value, tone }: { label: string; value: string | number; tone: StatusTone }) {
  return (
    <div className="erp-card erp-card--padded erp-kpi-card">
      <span className="erp-kpi-label">{label}</span>
      <strong className="erp-kpi-value">{value}</strong>
      <StatusChip tone={tone}>{label}</StatusChip>
    </div>
  );
}

function mergeRemittances(
  localRemittances: CODRemittance[],
  remittances: CODRemittance[],
  query: CODRemittanceQuery
) {
  const localMatches = localRemittances.filter((remittance) => matchesRemittanceQuery(remittance, query));
  const localIds = new Set(localMatches.map((remittance) => remittance.id));

  return [...localMatches, ...remittances.filter((remittance) => !localIds.has(remittance.id))];
}

function matchesRemittanceQuery(remittance: CODRemittance, query: CODRemittanceQuery) {
  const search = query.search?.trim().toLowerCase();
  return (
    (!query.status || remittance.status === query.status) &&
    (!search ||
      [remittance.remittanceNo, remittance.carrierCode, remittance.carrierName]
        .filter(Boolean)
        .some((value) => String(value).toLowerCase().includes(search)))
  );
}

function summarizeRemittances(remittances: CODRemittance[]) {
  return remittances.reduce(
    (summary, remittance) => ({
      expectedAmount: summary.expectedAmount + Number(remittance.expectedAmount),
      remittedAmount: summary.remittedAmount + Number(remittance.remittedAmount),
      discrepancyAmount: summary.discrepancyAmount + Number(remittance.discrepancyAmount),
      openDiscrepancies:
        summary.openDiscrepancies +
        remittance.discrepancies.filter((discrepancy) => discrepancy.status === "open").length
    }),
    { expectedAmount: 0, remittedAmount: 0, discrepancyAmount: 0, openDiscrepancies: 0 }
  );
}

async function runStatusAction(id: string, action: "match" | "submit" | "approve" | "close") {
  if (action === "match") {
    return matchCODRemittance(id);
  }
  if (action === "submit") {
    return submitCODRemittance(id);
  }
  if (action === "approve") {
    return approveCODRemittance(id);
  }

  return closeCODRemittance(id);
}
