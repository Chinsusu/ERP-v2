import { apiGet } from "../../../shared/api/client";
import { shouldUsePrototypeFallback } from "../../../shared/api/prototypeFallback";
import type { components, operations } from "../../../shared/api/generated/schema";
import type { AuditLogItem, AuditLogQuery, AuditLogSummary } from "../types";

type AuditLogApiItem = components["schemas"]["AuditLogItem"];
type AuditLogApiQuery = operations["listAuditLogs"]["parameters"]["query"];

const defaultAccessToken = "local-dev-access-token";

export const prototypeAuditLogs: AuditLogItem[] = [
  {
    id: "audit-adjust-260426-0001",
    actorId: "user-erp-admin",
    action: "inventory.stock_movement.adjusted",
    entityType: "inventory.stock_movement",
    entityId: "mov-adjust-260426-0001",
    requestId: "req_adjust_260426",
    afterData: {
      movement_type: "ADJUST",
      quantity: 8,
      warehouse_id: "wh-hcm",
      sku: "SERUM-30ML",
      available_delta: -8
    },
    metadata: {
      reason: "cycle count variance",
      source: "inventory stock movement"
    },
    createdAt: "2026-04-26T08:30:00Z"
  },
  {
    id: "audit-rbac-260426-0002",
    actorId: "user-erp-admin",
    action: "security.role.assigned",
    entityType: "core.user_role",
    entityId: "role-assignment-warehouse-lead",
    requestId: "req_rbac_260426",
    afterData: {
      user_id: "user-warehouse-lead",
      role: "WAREHOUSE_LEAD"
    },
    metadata: {
      scope_type: "warehouse",
      warehouse: "HCM"
    },
    createdAt: "2026-04-26T07:50:00Z"
  },
  {
    id: "audit-qc-260426-0003",
    actorId: "user-qa",
    action: "qc.lot.released",
    entityType: "qc.inspection",
    entityId: "qc-inspection-260426-0007",
    requestId: "req_qc_260426",
    beforeData: {
      status: "hold"
    },
    afterData: {
      status: "released"
    },
    metadata: {
      lot_no: "LOT-2604A",
      sku: "SERUM-30ML"
    },
    createdAt: "2026-04-26T06:55:00Z"
  }
];

export async function getAuditLogs(query: AuditLogQuery = {}): Promise<AuditLogItem[]> {
  try {
    const items = await apiGet("/audit-logs", {
      accessToken: defaultAccessToken,
      query: toApiQuery(query)
    });

    return items.map(fromApiItem);
  } catch (reason) {
    if (!shouldUsePrototypeFallback(reason)) {
      throw reason;
    }

    return filterPrototypeAuditLogs(query);
  }
}

export function summarizeAuditLogs(items: AuditLogItem[]): AuditLogSummary {
  return {
    total: items.length,
    adjustments: items.filter((item) => item.action.includes("adjusted")).length,
    securityEvents: items.filter((item) => item.action.startsWith("security.")).length,
    latestEventAt: items[0]?.createdAt
  };
}

export function auditActionTone(action: string): "normal" | "success" | "warning" | "danger" | "info" {
  if (action.includes("adjusted")) {
    return "warning";
  }
  if (action.startsWith("security.")) {
    return "danger";
  }
  if (action.startsWith("qc.")) {
    return "success";
  }

  return "info";
}

export function compactAuditPayload(value?: Record<string, unknown>) {
  if (!value || Object.keys(value).length === 0) {
    return "-";
  }

  return Object.entries(value)
    .slice(0, 3)
    .map(([key, item]) => `${key}: ${String(item)}`)
    .join(", ");
}

function fromApiItem(item: AuditLogApiItem): AuditLogItem {
  return {
    id: item.id,
    actorId: item.actor_id,
    action: item.action,
    entityType: item.entity_type,
    entityId: item.entity_id,
    requestId: item.request_id,
    beforeData: item.before_data,
    afterData: item.after_data,
    metadata: item.metadata,
    createdAt: item.created_at
  };
}

function toApiQuery(query: AuditLogQuery): AuditLogApiQuery {
  return {
    actor_id: query.actorId,
    action: query.action,
    entity_type: query.entityType,
    entity_id: query.entityId,
    limit: query.limit
  };
}

function filterPrototypeAuditLogs(query: AuditLogQuery): AuditLogItem[] {
  const action = query.action?.trim().toLowerCase();
  const actorId = query.actorId?.trim().toLowerCase();
  const entityType = query.entityType?.trim().toLowerCase();
  const entityId = query.entityId?.trim().toLowerCase();
  const limit = query.limit && query.limit > 0 ? query.limit : prototypeAuditLogs.length;

  return prototypeAuditLogs
    .filter((item) => {
      if (action && item.action.toLowerCase() !== action) {
        return false;
      }
      if (actorId && item.actorId.toLowerCase() !== actorId) {
        return false;
      }
      if (entityType && item.entityType.toLowerCase() !== entityType) {
        return false;
      }
      if (entityId && item.entityId.toLowerCase() !== entityId) {
        return false;
      }

      return true;
    })
    .slice(0, limit);
}
