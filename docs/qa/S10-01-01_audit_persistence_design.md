# S10-01-01 Audit Persistence Design

Project: Web ERP for cosmetics operations
Sprint: Sprint 10 - Persist operational runtime stores v1
Task: S10-01-01 Audit persistence design
Date: 2026-05-01
Status: Design complete; ready for S10-01-02 implementation

---

## 1. Purpose

This document maps the current runtime audit behavior to a PostgreSQL-backed store.

The goal is to persist audit evidence without changing audit semantics, API response envelopes, or existing action behavior.

---

## 2. Current Runtime Behavior

Current store wiring:

```text
apps/api/cmd/api/main.go
auditLogStore := audit.NewPrototypeLogStore()
```

Current interface:

```text
apps/api/internal/shared/audit/audit.go

type LogStore interface {
  Record(ctx context.Context, log Log) error
  List(ctx context.Context, query Query) ([]Log, error)
}
```

Current HTTP read endpoint:

```text
GET /api/v1/audit-logs
```

Current query filters:

```text
actor_id
action
entity_type
entity_id
limit
```

Current API response fields:

```text
id
actor_id
action
entity_type
entity_id
request_id
before_data
after_data
metadata
created_at
```

Runtime risk:

```text
The prototype audit store resets on API restart or redeploy, so auth/action evidence disappears.
```

---

## 3. Existing PostgreSQL Schema

Existing table:

```text
audit.audit_logs
```

Current columns from `000002_create_phase1_base_tables.up.sql`:

```text
id uuid PRIMARY KEY DEFAULT gen_random_uuid()
org_id uuid NOT NULL REFERENCES core.organizations(id)
actor_id uuid REFERENCES core.users(id)
action text NOT NULL
entity_type text NOT NULL
entity_id uuid
request_id text
before_data jsonb
after_data jsonb
metadata jsonb NOT NULL DEFAULT '{}'::jsonb
created_at timestamptz NOT NULL DEFAULT now()
```

Current indexes:

```text
ix_audit_logs_entity ON audit.audit_logs(entity_type, entity_id, created_at DESC)
ix_audit_logs_actor  ON audit.audit_logs(actor_id, created_at DESC)
```

Important existing writer:

```text
apps/api/internal/modules/inventory/application/postgres_stock_movement_store.go
```

That store already inserts stock movement audit rows directly into `audit.audit_logs`.

---

## 4. Compatibility Gap

The Go `audit.Log` model uses string references:

```text
Log.ID
Log.OrgID
Log.ActorID
Log.EntityID
```

Many current runtime events use stable text references, not UUIDs:

```text
org-my-pham
user-erp-admin
anonymous
sales order / purchase order / receiving / return / shipping prototype ids
```

The existing database table uses UUID/FK columns for `org_id`, `actor_id`, and `entity_id`.

Do not solve this by forcing text values into UUID columns or dropping those text references. That would either fail writes or break traceability.

---

## 5. Design Decision

Keep `audit.LogStore` unchanged and add a PostgreSQL implementation behind the existing interface.

Use the existing `audit.audit_logs` table as the canonical audit table, but add compatibility reference columns so current API semantics survive:

```text
log_ref text
org_ref text
actor_ref text
entity_ref text
```

Mapping:

| Go audit.Log field | PostgreSQL write |
| --- | --- |
| `Log.ID` | `log_ref`; also `id` only when the value is a UUID |
| `Log.OrgID` | `org_ref`; `org_id` resolved to a UUID for FK compatibility |
| `Log.ActorID` | `actor_ref`; `actor_id` only when the value is a known UUID user id |
| `Log.EntityID` | `entity_ref`; `entity_id` only when the value is a UUID |
| `Log.Action` | `action` |
| `Log.EntityType` | `entity_type` |
| `Log.RequestID` | `request_id` |
| `Log.BeforeData` | `before_data` |
| `Log.AfterData` | `after_data` |
| `Log.Metadata` | `metadata` |
| `Log.CreatedAt` | `created_at` |

Read mapping:

```text
Log.ID       = COALESCE(log_ref, id::text)
Log.OrgID    = COALESCE(org_ref, org_id::text)
Log.ActorID  = COALESCE(actor_ref, actor_id::text, '')
Log.EntityID = COALESCE(entity_ref, entity_id::text, '')
```

Query mapping:

```text
actor_id    -> actor_ref, with fallback to actor_id::text for older rows
entity_id   -> entity_ref, with fallback to entity_id::text for older rows
entity_type -> entity_type
action      -> action
limit       -> same 1..100 normalization as current in-memory store
```

This preserves existing API behavior while keeping the UUID FK columns useful for future fully persisted entities.

---

## 6. Migration Plan For S10-01-02

Add a small migration after current migration head:

```text
ALTER TABLE audit.audit_logs
  ADD COLUMN IF NOT EXISTS log_ref text,
  ADD COLUMN IF NOT EXISTS org_ref text,
  ADD COLUMN IF NOT EXISTS actor_ref text,
  ADD COLUMN IF NOT EXISTS entity_ref text;
```

Backfill compatibility references for existing rows:

```text
UPDATE audit.audit_logs
SET
  log_ref = COALESCE(log_ref, id::text),
  org_ref = COALESCE(org_ref, org_id::text),
  actor_ref = COALESCE(actor_ref, actor_id::text),
  entity_ref = COALESCE(entity_ref, entity_id::text);
```

Add indexes:

```text
CREATE UNIQUE INDEX IF NOT EXISTS uq_audit_logs_log_ref
  ON audit.audit_logs(log_ref)
  WHERE log_ref IS NOT NULL;

CREATE INDEX IF NOT EXISTS ix_audit_logs_actor_ref
  ON audit.audit_logs(actor_ref, created_at DESC)
  WHERE actor_ref IS NOT NULL;

CREATE INDEX IF NOT EXISTS ix_audit_logs_entity_ref
  ON audit.audit_logs(entity_type, entity_ref, created_at DESC)
  WHERE entity_ref IS NOT NULL;
```

Keep these compatibility columns nullable so existing direct SQL writers, especially stock movement audit insertion, do not break before they are updated.

Down migration should drop only the new indexes and columns.

---

## 7. Store Selection Plan For S10-01-02

Add a runtime selector similar to stock movement store selection:

```text
if DATABASE_URL is empty:
  audit.NewPrototypeLogStore()
else:
  audit.NewPostgresLogStore(...)
```

The store should remain safe in local/no-DB mode.

The PostgreSQL store needs an organization UUID for `audit.audit_logs.org_id`. For Sprint 10 implementation:

```text
1. If Log.OrgID is a UUID, use it as org_id.
2. Else try to resolve Log.OrgID against core.organizations.code.
3. Else fall back to the dev seed organization only in local/dev environments:
   00000000-0000-4000-8000-000000000001
4. Preserve the original Log.OrgID in org_ref either way.
```

If later production multi-organization behavior needs stricter org resolution, that should be a separate auth/tenant hardening task. Do not block S10 audit persistence on a full tenant model.

---

## 8. Required Tests For S10-01-02

Backend unit tests:

```text
PostgresLogStore.Record inserts log_ref/org_ref/actor_ref/entity_ref.
PostgresLogStore.List filters by actor_ref.
PostgresLogStore.List filters by entity_type + entity_ref.
PostgresLogStore.List returns newest first and respects limit <= 100.
PostgresLogStore preserves before_data, after_data, metadata, request_id, and created_at.
Runtime selector uses prototype store without DATABASE_URL.
Runtime selector uses PostgreSQL store with DATABASE_URL.
```

Migration checks:

```text
Apply all migrations on PostgreSQL 16.
Roll back all migrations on PostgreSQL 16.
```

Regression checks:

```text
go test ./internal/shared/audit ./cmd/api -run 'Test.*Audit|Test.*Permission' -count=1
make audit-permission-regression
```

---

## 9. Smoke Plan For S10-01-03

Add or extend dev smoke to prove persistence:

```text
1. Trigger login or another audited action through API.
2. Capture returned/request audit evidence.
3. Query /api/v1/audit-logs by actor_id or entity_type/entity_id.
4. Query PostgreSQL audit.audit_logs for the same log_ref/entity_ref.
5. Restart or redeploy API.
6. Query /api/v1/audit-logs again and confirm the event remains present.
```

The smoke should be deterministic enough to run repeatedly without relying on production data.

---

## 10. Out Of Scope For S10-01-01

This design task does not:

```text
- Implement the PostgreSQL audit store.
- Add migrations.
- Change runtime wiring.
- Change audit API response shape.
- Update OpenAPI.
- Backfill historical production data.
- Solve full tenant/organization identity modeling.
```

---

## 11. Acceptance For S10-01-01

S10-01-01 is complete when:

```text
1. Current audit interface, endpoint, DB schema, and direct DB writer are documented.
2. UUID/text compatibility gap is explicit.
3. S10-01-02 has a concrete migration/store wiring plan.
4. S10-01-03 has a concrete persistence smoke plan.
5. No runtime behavior changes in this task.
```
