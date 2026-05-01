# S12-02-02 Available-Stock QC Consistency Smoke

Project: Web ERP for cosmetics operations
Sprint: Sprint 12 - Batch QC status persistence
Task: S12-02-02 Available-stock QC consistency smoke
Date: 2026-05-01
Status: Pass on shared dev

---

## 1. Purpose

Prove that available-stock reads use the persisted batch QC status after a QC transition.

This smoke covers the task-board acceptance:

```text
Available stock and reservation-facing reads respect persisted batch QC status after transition.
```

---

## 2. Environment

```text
Dev host: 10.1.1.120
Dev repo: /opt/ERP-v2
Observed repo commit: 3e9952bc
Runtime: shared dev Docker Compose stack
Database: PostgreSQL service erp-dev-postgres-1
```

This task did not add or change runtime code.

---

## 3. Smoke Fixture

The smoke used an existing seeded batch with a persisted stock balance:

```text
batch_id: 00000000-0000-4000-8000-000000001202
batch_no: BATCH-SER-LOCAL
stock_status: available
qty_on_hand: 80.000000
qty_reserved: 5.000000
```

The batch QC status was reset to `hold` as fixture setup before the API transition.

---

## 4. Verification Performed

### 4.1 Available Stock While Batch Is HOLD

Request:

```text
GET /api/v1/inventory/available-stock?batch_id=00000000-0000-4000-8000-000000001202
```

Expected behavior:

```text
on-hand stock exists, reserved stock exists, but QC hold blocks availability.
```

Observed result:

```text
available_before=hold_reserved5_available0
```

The response contained:

```text
batch_qc_status=hold
physical_qty=80.000000
reserved_qty=5.000000
qc_hold_qty=80.000000
available_qty=0.000000
```

### 4.2 API Transition

Request:

```text
POST /api/v1/inventory/batches/00000000-0000-4000-8000-000000001202/qc-transitions
body: qc_status=pass, business_ref=S12-02-02-SMOKE
```

Observed result:

```text
transition_api=success_pass
```

### 4.3 Available Stock After Batch PASS

Request:

```text
GET /api/v1/inventory/available-stock?batch_id=00000000-0000-4000-8000-000000001202
```

Observed result:

```text
available_after=pass_reserved5_available75
```

The response contained:

```text
batch_qc_status=pass
physical_qty=80.000000
reserved_qty=5.000000
qc_hold_qty=0.000000
available_qty=75.000000
```

### 4.4 Database Cross-Check

Batch status after the API transition:

```text
db_batch_after=00000000-0000-4000-8000-000000001202|pass|active|user-erp-admin
```

Stock balance stayed unchanged:

```text
db_balance_after=80.000000|5.000000|available
```

---

## 5. Result

S12-02-02 passes on shared dev.

Verified behavior:

1. Persisted QC `hold` blocks available stock even when stock balance status is `available`.
2. Reserved quantity remains visible in the read model while availability is blocked.
3. API QC transition changes the persisted batch status to `pass`.
4. Available stock recalculates from the same persisted stock balance as `80 - 5 = 75`.

Not verified in this task:

- Inbound QC status updater integration. That is covered by S12-03-01 and S12-03-02.
