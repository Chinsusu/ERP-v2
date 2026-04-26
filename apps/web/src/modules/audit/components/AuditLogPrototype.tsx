"use client";

import { useMemo, useState } from "react";
import { DataTable, StatusChip, type DataTableColumn } from "@/shared/design-system/components";
import { useAuditLogs } from "../hooks/useAuditLogs";
import { auditActionTone, compactAuditPayload, prototypeAuditLogs } from "../services/auditLogService";
import type { AuditLogItem, AuditLogQuery } from "../types";

const actionOptions = [
  { label: "All actions", value: "" },
  ...Array.from(new Set(prototypeAuditLogs.map((item) => item.action))).map((value) => ({
    label: value,
    value
  }))
];

const entityTypeOptions = [
  { label: "All entity types", value: "" },
  ...Array.from(new Set(prototypeAuditLogs.map((item) => item.entityType))).map((value) => ({
    label: value,
    value
  }))
];

const columns: DataTableColumn<AuditLogItem>[] = [
  {
    key: "created",
    header: "Time",
    render: (row) => formatDateTime(row.createdAt),
    width: "160px"
  },
  {
    key: "actor",
    header: "Actor",
    render: (row) => row.actorId,
    width: "170px"
  },
  {
    key: "action",
    header: "Action",
    render: (row) => <StatusChip tone={auditActionTone(row.action)}>{row.action}</StatusChip>,
    width: "250px"
  },
  {
    key: "entity",
    header: "Entity",
    render: (row) => (
      <span className="erp-audit-entity">
        <strong>{row.entityType}</strong>
        <small>{row.entityId}</small>
      </span>
    ),
    width: "260px"
  },
  {
    key: "metadata",
    header: "Metadata",
    render: (row) => <span className="erp-audit-metadata">{compactAuditPayload(row.metadata)}</span>
  }
];

export function AuditLogPrototype() {
  const [action, setAction] = useState("");
  const [entityType, setEntityType] = useState("");
  const [entityId, setEntityId] = useState("");
  const query = useMemo<AuditLogQuery>(
    () => ({
      action: action || undefined,
      entityType: entityType || undefined,
      entityId: entityId || undefined,
      limit: 50
    }),
    [action, entityId, entityType]
  );
  const { items, loading, summary } = useAuditLogs(query);

  return (
    <section className="erp-module-page erp-audit-page">
      <header className="erp-page-header">
        <div>
          <p className="erp-module-eyebrow">AU</p>
          <h1 className="erp-page-title">Audit Log</h1>
          <p className="erp-page-description">Immutable trace of sensitive ERP actions</p>
        </div>
      </header>

      <section className="erp-audit-toolbar" aria-label="Audit log filters">
        <label className="erp-field">
          <span>Action</span>
          <select className="erp-input" value={action} onChange={(event) => setAction(event.target.value)}>
            {actionOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Entity type</span>
          <select className="erp-input" value={entityType} onChange={(event) => setEntityType(event.target.value)}>
            {entityTypeOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Entity ID</span>
          <input
            className="erp-input"
            type="search"
            value={entityId}
            placeholder="mov-adjust-260426-0001"
            onChange={(event) => setEntityId(event.target.value)}
          />
        </label>
      </section>

      <section className="erp-kpi-grid erp-audit-kpis">
        <AuditKPI label="Events" value={summary.total} tone="info" />
        <AuditKPI label="Adjustments" value={summary.adjustments} tone="warning" />
        <AuditKPI label="Security" value={summary.securityEvents} tone="danger" />
        <AuditKPI label="Latest" value={summary.latestEventAt ? formatShortTime(summary.latestEventAt) : "-"} tone="normal" />
      </section>

      <section className="erp-card erp-card--padded erp-module-table-card">
        <div className="erp-section-header">
          <h2 className="erp-section-title">Audit events</h2>
          <StatusChip tone={items.length === 0 ? "warning" : "info"}>{items.length} rows</StatusChip>
        </div>
        <DataTable columns={columns} rows={items} getRowKey={(row) => row.id} loading={loading} />
      </section>
    </section>
  );
}

function AuditKPI({
  label,
  value,
  tone
}: {
  label: string;
  value: number | string;
  tone: "normal" | "success" | "warning" | "danger" | "info";
}) {
  return (
    <article className="erp-card erp-card--padded erp-kpi-card">
      <div className="erp-kpi-label">{label}</div>
      <strong className="erp-kpi-value">{value}</strong>
      <StatusChip tone={tone}>{label}</StatusChip>
    </article>
  );
}

function formatDateTime(value: string) {
  return new Intl.DateTimeFormat("en-GB", {
    dateStyle: "medium",
    timeStyle: "short"
  }).format(new Date(value));
}

function formatShortTime(value: string) {
  return new Intl.DateTimeFormat("en-GB", {
    hour: "2-digit",
    minute: "2-digit"
  }).format(new Date(value));
}
