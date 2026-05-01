BEGIN;

DROP INDEX IF EXISTS audit.ix_audit_logs_entity_ref;
DROP INDEX IF EXISTS audit.ix_audit_logs_actor_ref;
DROP INDEX IF EXISTS audit.uq_audit_logs_log_ref;

ALTER TABLE audit.audit_logs
  DROP COLUMN IF EXISTS entity_ref,
  DROP COLUMN IF EXISTS actor_ref,
  DROP COLUMN IF EXISTS org_ref,
  DROP COLUMN IF EXISTS log_ref;

COMMIT;
