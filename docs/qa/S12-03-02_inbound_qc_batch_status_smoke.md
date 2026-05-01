# S12-03-02 Inbound QC Batch-Status Smoke

Project: Web ERP for cosmetics operations
Sprint: Sprint 12 - Batch QC status persistence
Task: S12-03-02 Inbound QC batch-status smoke
Date: 2026-05-01
Status: Pass on shared dev

---

## 1. Purpose

Prove that inbound QC decisions update the persisted batch QC status and remain consistent across batch detail, available stock, audit logs, and API restart.

This smoke covers the task-board acceptance:

```text
Inbound QC PASS/FAIL/PARTIAL status effects remain consistent across batch detail, available stock, and audit after restart/redeploy.
```

---

## 2. Environment

```text
Dev host: 10.1.1.120
Dev repo: /opt/ERP-v2
Observed repo commit: 8f834033
Runtime: shared dev Docker Compose stack
Database: PostgreSQL service erp-dev-postgres-1
API container restarted during smoke: erp-dev-api-1
Smoke run id: 005250
```

This task did not add or change runtime code.

---

## 3. Smoke Fixture

The smoke created three dev-only inbound receiving fixtures, each starting with a persisted batch in QC `hold`:

| Case | Batch ID | Inbound QC ID | Quantity |
| --- | --- | --- | --- |
| PASS | `00000000-0000-4000-8000-030211005250` | `00000000-0000-4000-8000-030241005250` | `6.000000` |
| FAIL | `00000000-0000-4000-8000-030212005250` | `00000000-0000-4000-8000-030242005250` | `5.000000` |
| PARTIAL | `00000000-0000-4000-8000-030213005250` | `00000000-0000-4000-8000-030243005250` | `9.000000` |

The fixtures used persisted `inventory.warehouse_receivings`, `inventory.warehouse_receiving_lines`, and `inventory.batches` rows. The QC decisions under test were performed through the HTTP API.

---

## 4. Verification Performed

### 4.1 API Flow

For each case, the smoke called:

```text
POST /api/v1/inbound-qc-inspections
POST /api/v1/inbound-qc-inspections/{inspection_id}/start
POST /api/v1/inbound-qc-inspections/{inspection_id}/pass|fail|partial
```

Observed result:

```text
pass_decision=ok
fail_decision=ok
partial_decision=ok
```

### 4.2 Pre-Restart Consistency

After the decisions, the smoke checked batch detail and available stock:

```text
pre_restart_consistency=pass_fail_partial_ok
```

Expected and observed effects:

| Case | Batch QC Status | Available-Stock Evidence |
| --- | --- | --- |
| PASS | `pass` | `physical_qty=6.000000`, `available_qty=6.000000` |
| FAIL | `fail` | `damaged_qty=5.000000`, `available_qty=0.000000` |
| PARTIAL | `pass` | `physical_qty=9.000000`, `damaged_qty=2.000000`, `qc_hold_qty=3.000000`, `available_qty=4.000000` |

### 4.3 Audit

Database audit checks:

```text
batch_transition_audit_count=3
inbound_decision_audit_count=3
```

This proves:

1. Each inbound QC decision wrote the expected `inventory.batch.qc_status_changed` audit event.
2. Each inbound QC decision also wrote its own inbound inspection decision audit event.

### 4.4 Restart Survival

The dev API container was restarted and health was checked:

```text
api_restart=healthy
```

The smoke then re-ran the same batch detail and available-stock checks:

```text
after_restart_consistency=pass_fail_partial_ok
```

---

## 5. Result

S12-03-02 passes on shared dev.

Verified behavior:

1. Inbound QC PASS changes persisted batch QC status to `pass` and releases available stock.
2. Inbound QC FAIL changes persisted batch QC status to `fail` and keeps available stock at `0.000000`.
3. Inbound QC PARTIAL changes persisted batch QC status to `pass` when passed quantity exists, while failed and held quantities remain blocked.
4. Batch detail and available-stock reads stay consistent after API restart.
5. Batch transition audit and inbound QC decision audit are both present.

Not verified in this task:

- Production release tagging. That belongs in S12-05-01 release evidence after all Sprint 12 tasks are complete.
