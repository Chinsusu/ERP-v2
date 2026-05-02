"use client";

import { useMemo, useState } from "react";
import { DataTable, StatusChip, type DataTableColumn } from "@/shared/design-system/components";
import { erpNumberLocale, erpTimezone, formatDateTimeVI } from "@/shared/format/numberFormat";
import { t } from "@/shared/i18n";
import { useAuditLogs } from "../hooks/useAuditLogs";
import { auditActionTone, compactAuditPayload, prototypeAuditLogs } from "../services/auditLogService";
import type { AuditLogItem, AuditLogQuery } from "../types";

const actionOptions = [
  { label: auditCopy("filters.allActions"), value: "" },
  ...Array.from(new Set(prototypeAuditLogs.map((item) => item.action))).map((value) => ({
    label: getAuditActionLabel(value),
    value
  }))
];

const entityTypeOptions = [
  { label: auditCopy("filters.allEntityTypes"), value: "" },
  ...Array.from(new Set(prototypeAuditLogs.map((item) => item.entityType))).map((value) => ({
    label: getAuditEntityTypeLabel(value),
    value
  }))
];

const columns: DataTableColumn<AuditLogItem>[] = [
  {
    key: "created",
    header: auditCopy("columns.time"),
    render: (row) => formatDateTime(row.createdAt),
    width: "160px"
  },
  {
    key: "actor",
    header: auditCopy("columns.actor"),
    render: (row) => row.actorId,
    width: "170px"
  },
  {
    key: "action",
    header: auditCopy("columns.action"),
    render: (row) => (
      <span className="erp-audit-entity">
        <strong>
          <StatusChip tone={auditActionTone(row.action)}>{getAuditActionLabel(row.action)}</StatusChip>
        </strong>
        <small>{row.action}</small>
      </span>
    ),
    width: "250px"
  },
  {
    key: "entity",
    header: auditCopy("columns.entity"),
    render: (row) => (
      <span className="erp-audit-entity">
        <strong>{getAuditEntityTypeLabel(row.entityType)}</strong>
        <small>{row.entityType}</small>
        <small>{row.entityId}</small>
      </span>
    ),
    width: "260px"
  },
  {
    key: "metadata",
    header: auditCopy("columns.metadata"),
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
          <h1 className="erp-page-title">{auditCopy("title")}</h1>
          <p className="erp-page-description">{auditCopy("description")}</p>
        </div>
      </header>

      <section className="erp-audit-toolbar" aria-label={auditCopy("filters.ariaLabel")}>
        <label className="erp-field">
          <span>{auditCopy("filters.action")}</span>
          <select className="erp-input" value={action} onChange={(event) => setAction(event.target.value)}>
            {actionOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>{auditCopy("filters.entityType")}</span>
          <select className="erp-input" value={entityType} onChange={(event) => setEntityType(event.target.value)}>
            {entityTypeOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>{auditCopy("filters.entityId")}</span>
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
        <AuditKPI label={auditCopy("kpi.events")} value={summary.total} tone="info" />
        <AuditKPI label={auditCopy("kpi.adjustments")} value={summary.adjustments} tone="warning" />
        <AuditKPI label={auditCopy("kpi.security")} value={summary.securityEvents} tone="danger" />
        <AuditKPI
          label={auditCopy("kpi.latest")}
          value={summary.latestEventAt ? formatShortTime(summary.latestEventAt) : "-"}
          tone="normal"
        />
      </section>

      <section className="erp-card erp-card--padded erp-module-table-card">
        <div className="erp-section-header">
          <h2 className="erp-section-title">{auditCopy("eventsTitle")}</h2>
          <StatusChip tone={items.length === 0 ? "warning" : "info"}>
            {auditCopy("rows", { count: items.length })}
          </StatusChip>
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
  return formatDateTimeVI(value);
}

function formatShortTime(value: string) {
  return new Intl.DateTimeFormat(erpNumberLocale, {
    timeZone: erpTimezone,
    hour: "2-digit",
    minute: "2-digit"
  }).format(new Date(value));
}

function getAuditActionLabel(action: string) {
  return auditCopy(`actions.${action}`, undefined, action);
}

function getAuditEntityTypeLabel(entityType: string) {
  return auditCopy(`entityTypes.${entityType}`, undefined, entityType);
}

function auditCopy(key: string, values?: Record<string, string | number>, fallback?: string) {
  return t(`audit.${key}`, { values, fallback });
}
