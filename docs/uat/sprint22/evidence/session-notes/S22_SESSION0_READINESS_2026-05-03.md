# S22 Session 0 Readiness Evidence - 2026-05-03

Project: Web ERP for cosmetics operations
Sprint: Sprint 22 - UAT Pilot Pack for Warehouse + Sales + QC
Session: S22-SESSION-00 - Readiness Check
Environment: `http://10.1.1.120:8088`
Status: Blocked before business UAT

---

## 1. What Was Verified

Repository state:

```text
Local branch: codex/s22-uat-session0-readiness
Base main before Session 0 evidence: 183f7447
Remote dev repo status during smoke: main...origin/main
```

Environment health:

```text
GET /health: 200
GET /healthz: 200
GET /api/v1/health: success=true
```

Backend auth/session:

```text
admin@example.local login: passed
/me with issued access token: passed
auth refresh rotation: passed in full dev smoke
logout invalidates session: passed in full dev smoke
lockout after repeated failed login attempts: passed in full dev smoke
```

Full dev smoke:

```text
./infra/scripts/smoke-dev-full.sh passed on the dev server.
Covered auth/session, warehouse daily board metrics, reporting, master data persistence, finance, sales reservation, stock adjustment, stock count, inbound QC, carrier manifest, pick/pack, returns, supplier rejection, and subcontract flow.
```

API seed/data readiness snapshot:

```text
/products: 15 rows
/warehouses: 14 rows
/warehouse-locations: 14 rows
/suppliers: 12 rows
/customers: 13 rows
/inventory/batches: 10 rows
/sales-orders: 170 rows
/returns/receipts: 62 rows
/purchase-orders: 67 rows
/inbound-qc-inspections: 81 rows
/shipping/manifests: 51 rows
/pick-tasks: 49 rows
/pack-tasks: 49 rows
/rbac/roles: 9 rows
/rbac/permissions: 24 rows
```

Browser UI smoke:

```text
Chrome headless CDP fallback used because local npx/Playwright CLI was unavailable.
Admin login to dashboard: passed
Warehouse Daily Board page: passed
Vietnamese warehouse copy detected: passed
Invalid login Vietnamese error: passed
```

Screenshot evidence:

```text
output/playwright/s22-session0-dashboard.png
output/playwright/s22-session0-warehouse-daily-board.png
output/playwright/s22-session0-invalid-login.png
```

---

## 2. Blocker Found

Issue:

```text
S22-ISSUE-001
Role-based UAT users cannot authenticate.
```

Observed:

```text
admin@example.local login passed.
warehouse_user@example.local login returned 401.
sales_user@example.local login returned 401.
qc_user@example.local login returned 401.
```

Impact:

```text
Business UAT cannot start because Sprint 22 requires Warehouse, Sales, and QC users to validate role-specific menus and workflows.
Admin-only testing does not prove RBAC, menu visibility, or role-specific workflow readiness.
```

Required next action:

```text
Enable or create backend-authenticated Warehouse, Sales, and QC UAT users.
Re-run S22-UAT-001 before scheduling business users.
```

---

## 3. Session 0 Decision

```text
Environment smoke: passed
Admin auth/UI smoke: passed
Technical operational smoke: passed
Seed data availability: partial/dev-smoke available; S22-specific UAT data still needs business approval
Role-based UAT users: blocked
Business UAT: do not start yet
Release tag v0.22: hold
```
