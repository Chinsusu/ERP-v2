BEGIN;

ALTER TABLE audit.audit_logs
  ADD COLUMN IF NOT EXISTS log_ref text,
  ADD COLUMN IF NOT EXISTS org_ref text,
  ADD COLUMN IF NOT EXISTS actor_ref text,
  ADD COLUMN IF NOT EXISTS entity_ref text;

UPDATE audit.audit_logs
SET
  log_ref = COALESCE(log_ref, id::text),
  org_ref = COALESCE(org_ref, org_id::text),
  actor_ref = COALESCE(actor_ref, actor_id::text),
  entity_ref = COALESCE(entity_ref, entity_id::text);

CREATE UNIQUE INDEX IF NOT EXISTS uq_audit_logs_log_ref
  ON audit.audit_logs(log_ref)
  WHERE log_ref IS NOT NULL;

CREATE INDEX IF NOT EXISTS ix_audit_logs_actor_ref
  ON audit.audit_logs(actor_ref, created_at DESC)
  WHERE actor_ref IS NOT NULL;

CREATE INDEX IF NOT EXISTS ix_audit_logs_entity_ref
  ON audit.audit_logs(entity_type, entity_ref, created_at DESC)
  WHERE entity_ref IS NOT NULL;

COMMIT;
