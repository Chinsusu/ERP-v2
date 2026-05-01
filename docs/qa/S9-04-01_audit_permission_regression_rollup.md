# S9-04-01 Audit / Permission Regression Rollup

Project: Web ERP for cosmetics operations
Sprint: Sprint 9 - System hardening / production readiness core
Task: S9-04-01 Audit/permission regression rollup
Date: 2026-05-01
Status: Checkpoint defined

---

## 1. Purpose

This checkpoint groups the existing audit and permission regression tests so release checks can run them deliberately instead of relying on memory or ad hoc test selection.

It does not change runtime behavior.

---

## 2. Make Target

Run:

```text
make audit-permission-regression
```

The target runs the backend packages that contain regression coverage for:

```text
- Finance mutation denial and audit write behavior.
- Order fulfillment permission and audit regression.
- Returns inspection and shift closing permission/audit regression.
- Reporting and finance drilldown permission boundaries.
- Bearer token, role catalog, permission catalog, and permission middleware behavior.
- Shared audit log filtering, sorting, and defensive-copy behavior.
```

The target also runs focused frontend tests for:

```text
- Reporting access regression.
- Permission menu visibility.
- Audit log service behavior.
```

---

## 3. Source Coverage

Backend packages run in full:

```text
apps/api/cmd/api
apps/api/internal/shared/auth
apps/api/internal/shared/audit
```

Frontend test files:

```text
apps/web/src/modules/reporting/services/reportAccessRegression.test.ts
apps/web/src/shared/permissions/menu.test.ts
apps/web/src/modules/audit/services/auditLogService.test.ts
```

---

## 4. Release Gate Use

Use this checkpoint before Sprint 9 release evidence is finalized and whenever a task changes:

```text
- Auth/session behavior.
- RBAC roles or permission keys.
- Protected route wiring.
- Finance/reporting access boundaries.
- Audit log write or read behavior.
```

This checkpoint complements the full `ci-check`; it is not a replacement for full CI when GitHub Actions is available.

---

## 5. Verification Notes

Expected local/dev verification:

```text
make audit-permission-regression
```

If the local machine lacks Go or pnpm, run equivalent backend and frontend commands in the dev verification environment and record the executed commands in the task changelog.
