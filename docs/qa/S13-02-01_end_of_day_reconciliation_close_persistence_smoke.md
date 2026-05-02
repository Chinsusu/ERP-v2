# S13-02-01 End-of-Day Reconciliation Close Persistence Smoke

Project: Web ERP for cosmetics operations
Sprint: Sprint 13 - End-of-day reconciliation persistence v1
Task: S13-02-01 Close persistence smoke
Date: 2026-05-02
Status: Passed on dev

---

## 1. Environment

```text
Host: 10.1.1.120
Repo: /opt/ERP-v2
Runtime URL: http://10.1.1.120:8088
Main commit before smoke: 839d98b3
API deployment commit: 637ecb58 plus test-only sync through 839d98b3
Database: erp-dev-postgres-1, PostgreSQL 16
```

The runtime persistence change was deployed after PR #435. PR #436 was test-only and synced to the dev repo without redeploying the API image.

---

## 2. Smoke Record

```text
Reconciliation ID: rec-s13-02-01-smoke-1777684750
Warehouse ref: warehouse_main
Business date: 2026-05-02
Shift code: s13-smoke-1777684750
Initial status: in_review
Exception note: S13 smoke variance accepted
Variance line: SERUM-30ML / LOT-S13-SMOKE / system 20 / counted 18
```

The smoke inserted a dedicated reconciliation row, checklist rows, and a variance line directly into PostgreSQL, then closed it only through the public API.

---

## 3. Execution

API close request:

```text
POST /api/v1/warehouse/end-of-day-reconciliations/rec-s13-02-01-smoke-1777684750/close
Payload: {"exception_note":"S13 smoke variance accepted"}
Actor: dev admin session
```

Result before restart:

```text
before_status: in_review
before_ready: false
close_http: 200
close_status: closed
audit_log_id: audit_1777684750627765060_27
```

The API container was restarted after close. Because current mock auth sessions are in-memory, the post-restart read used a fresh login session.

Result after restart:

```text
after_success: true
after_status: closed
after_closed_by: user-erp-admin
after_variance: -2
```

Database verification after restart:

```text
inventory.warehouse_daily_closings:
closed|user-erp-admin|S13 smoke variance accepted

audit.audit_logs:
warehouse.shift.closed rows for entity_ref rec-s13-02-01-smoke-1777684750 = 1
```

---

## 4. Result

```text
PASS
```

The close status, close actor, exception note, variance evidence, and audit row survived API restart. This proves S13-01-02/S13-01-03 moved end-of-day reconciliation close evidence out of runtime memory in DB mode.

---

## 5. Notes

```text
Mock auth sessions still reset on API restart.
This is expected under the current dev auth model and is not part of Sprint 13 reconciliation persistence.
```
