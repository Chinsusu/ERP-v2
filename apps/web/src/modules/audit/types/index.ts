export type AuditLogItem = {
  id: string;
  actorId: string;
  action: string;
  entityType: string;
  entityId: string;
  requestId?: string;
  beforeData?: Record<string, unknown>;
  afterData?: Record<string, unknown>;
  metadata: Record<string, unknown>;
  createdAt: string;
};

export type AuditLogQuery = {
  actorId?: string;
  action?: string;
  entityType?: string;
  entityId?: string;
  limit?: number;
};

export type AuditLogSummary = {
  total: number;
  adjustments: number;
  securityEvents: number;
  latestEventAt?: string;
};
