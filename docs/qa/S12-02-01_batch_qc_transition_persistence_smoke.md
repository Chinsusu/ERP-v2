# S12-02-01 Batch QC Transition Persistence Smoke

Project: Web ERP for cosmetics operations
Sprint: Sprint 12 - Batch QC status persistence
Task: S12-02-01 Batch QC transition persistence smoke
Date: 2026-05-01
Status: Pass on shared dev

---

## 1. Purpose

Prove that a batch QC status transition writes through the persisted PostgreSQL batch catalog path, can be read back through the API, and survives an API restart.

This smoke covers the task-board acceptance:

```text
POST QC transition updates inventory.batches and survives restart/redeploy.
```

---

## 2. Environment

```text
Dev host: 10.1.1.120
Dev repo: /opt/ERP-v2
Observed repo commit: 9363093e
Runtime: shared dev Docker Compose stack
Database: PostgreSQL service erp-dev-postgres-1
API container restarted during smoke: erp-dev-api-1
```

Notes:

- `9363093e` includes the S12-01-04 test-only merge.
- Runtime code was previously deployed after S12-01-03, which introduced the PostgreSQL-backed batch catalog store.
- This task did not add or change runtime code.

---

## 3. Smoke Fixture

A dedicated dev-only batch was inserted or reset to `hold` before calling the API:

```text
batch_ref: batch-s12-0201-smoke
batch_no: LOT-S12-0201
item_ref: item-serum-30ml
supplier_ref: sup-rm-bioactive
initial qc_status: hold
```

The direct SQL setup is only fixture preparation. The QC transition under test was performed through the API.

---

## 4. Verification Performed

### 4.1 API Transition

Request:

```text
POST /api/v1/inventory/batches/batch-s12-0201-smoke/qc-transitions
body: qc_status=pass, business_ref=S12-02-01-SMOKE
```

Observed result:

```text
transition_api=success_pass
```

### 4.2 Database Persistence

Read directly from `inventory.batches` after the API transition:

```text
db_after_transition=batch-s12-0201-smoke|pass|active|user-erp-admin
```

### 4.3 Transition History

Request:

```text
GET /api/v1/inventory/batches/batch-s12-0201-smoke/qc-transitions
```

Observed result:

```text
history_api=success_contains_business_ref
```

The response contained `S12-02-01-SMOKE`, proving the transition history remained connected to the persisted audit path.

### 4.4 Restart Survival

The dev API container was restarted and health was checked before reading the batch again:

```text
api_restart=healthy
detail_after_restart=success_pass
db_after_restart=batch-s12-0201-smoke|pass|active|user-erp-admin
```

---

## 5. Result

S12-02-01 passes on shared dev.

Verified behavior:

1. API QC transition changes `hold -> pass`.
2. `inventory.batches.qc_status` persists as `pass`.
3. Transition history remains queryable through the API.
4. The persisted status survives API restart.

Not verified in this task:

- Available-stock and reservation-facing read consistency. That is covered by S12-02-02.
- Inbound QC status updater integration. That is covered by S12-03-01 and S12-03-02.
