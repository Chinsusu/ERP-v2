# S22 Session 0 Rerun Evidence - Role Auth Unblock - 2026-05-03

Project: Web ERP for cosmetics operations
Sprint: Sprint 22 - UAT Pilot Pack for Warehouse + Sales + QC
Session: S22-SESSION-00 rerun after S22-ISSUE-001 fix
Environment: `http://10.1.1.120:8088`
Status: Readiness passed; business UAT not started

---

## 1. Change Under Verification

Issue:

```text
S22-ISSUE-001 - Role-based UAT users cannot authenticate
```

Fix evidence:

```text
PR: #546 Fix S22-ISSUE-001 role-based UAT user authentication
Merge commit: db894ddb0066a404f102ff82504b5b930b261e7c
Branch: codex/fix-S22-ISSUE-001-role-uat-auth
GitHub checks: api, web, openapi, e2e, required-api, required-web, required-openapi, required-migration passed
```

Root cause recorded in PR #546:

```text
Runtime login accepted only the configured admin mock login.
Warehouse/Sales seed users existed but could not authenticate.
QC UAT user was missing from dev seed/runtime auth.
Sales and QA permissions were missing the menu/route access needed by Sprint 22 Session 0.
```

---

## 2. Dev Deploy And Smoke

Deploy command:

```text
./infra/scripts/deploy-dev-staging.sh dev
```

Result:

```text
API, worker, and web images built from source.
PostgreSQL migrations and dev seed ran.
Reverse proxy was recreated after app deploy.
Deploy script completed successfully.
```

Full dev smoke result:

```text
./infra/scripts/smoke-dev-full.sh passed during deploy.
```

Role auth checks added to full dev smoke and verified:

```text
auth_role_warehouse_login: 200
auth_role_warehouse_me: 200, role WAREHOUSE_STAFF
auth_role_warehouse_route_1: 200
auth_role_warehouse_logout: 200

auth_role_sales_login: 200
auth_role_sales_me: 200, role SALES_OPS
auth_role_sales_route_1: 200
auth_role_sales_route_2: 200
auth_role_sales_logout: 200

auth_role_qc_login: 200
auth_role_qc_me: 200, role QA
auth_role_qc_route_1: 200
auth_role_qc_route_2: 200
auth_role_qc_logout: 200
```

Existing auth/session checks also passed:

```text
admin login
/me before and after API restart
refresh rotation
old refresh rejection
logout invalidation
lockout persistence after repeated failures
```

---

## 3. Browser UI Smoke

Tooling:

```text
Chrome headless CDP fallback was used because local npx/Playwright CLI was unavailable.
```

Verified browser flows:

```text
warehouse_user@example.local:
- login passed
- Warehouse Daily Board, Receiving, and Inventory menu entries visible
- Sales, QC, Finance, and Settings menu entries hidden
- /warehouse route accessible
- /finance route blocked
- logout returned to /login

sales_user@example.local:
- login passed
- Sales Orders and Inventory menu entries visible
- Warehouse, Receiving, QC, Finance, and Settings menu entries hidden
- /sales and /inventory routes accessible
- /qc route blocked
- logout returned to /login

qc_user@example.local:
- login passed
- Warehouse Daily Board, Receiving, Inventory, and QC menu entries visible
- Sales, Finance, and Settings menu entries hidden
- /qc and /receiving routes accessible
- /sales route blocked
- logout returned to /login

Invalid login:
- invalid credentials redirected to /login?error=invalid_credentials
- Vietnamese error copy rendered on the login form
```

Screenshot evidence:

```text
output/playwright/s22-role-auth-warehouse-menu.png
output/playwright/s22-role-auth-sales-menu.png
output/playwright/s22-role-auth-qc-menu.png
output/playwright/s22-role-auth-invalid-login.png
```

---

## 4. Session 0 Decision

```text
S22-ISSUE-001: resolved
Session 0 readiness: rerun passed
Warehouse/Sales/QC role auth: passed
RBAC menu/route guard smoke: passed
Invalid-login Vietnamese copy: passed
Business UAT: ready to schedule, not started
Go/No-Go: pending
v0.22 tag: hold
```

This note does not claim business UAT passed. It only records that the P0 readiness blocker preventing business UAT has been removed on the dev target.
