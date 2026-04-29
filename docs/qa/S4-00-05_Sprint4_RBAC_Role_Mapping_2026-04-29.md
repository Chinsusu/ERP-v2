# S4-00-05 Sprint 4 RBAC Role Mapping

Task: S4-00-05
Date: 2026-04-29
Owner: BE / Tech Lead
Verifier: Codex

## Decision

Add explicit Sprint 4 roles instead of overloading `ERP_ADMIN` for purchase and finance workflows.

New roles:

```text
PURCHASE_OPS
FINANCE_OPS
```

Existing `QA` remains the inbound QC decision role.

New permission:

```text
finance:view
```

## Approval Matrix

| Sprint 4 action | Allowed role / permission | Notes |
|---|---|---|
| PO list/detail view | `purchase:view` | `PURCHASE_OPS`, `FINANCE_OPS`, `ERP_ADMIN` |
| PO draft/create/submit | `PURCHASE_OPS`, `ERP_ADMIN` with `record:create` | Purchase users can create operational PO records |
| PO approve/cancel/close | `ERP_ADMIN` initially; later explicit approval role if needed | Avoid granting broad approval rights before PO handler exists |
| Receiving against approved PO | `WAREHOUSE_STAFF`, `WAREHOUSE_LEAD`, `ERP_ADMIN` | Uses warehouse/record permissions in implementation |
| Receiving exception approval | `WAREHOUSE_LEAD`, `ERP_ADMIN` | Mirrors current warehouse lead pattern |
| Inbound QC view | `qc:view` | `QA`, `ERP_ADMIN` |
| Inbound QC pass/fail/hold/partial | `qc:decision` | `QA`, `ERP_ADMIN` only |
| Finance/cost visibility | `finance:view` | `FINANCE_OPS`, `ERP_ADMIN` |
| Master data for suppliers/items/UOM | `master-data:view` and `record:create` where creating | `PURCHASE_OPS` can see needed master data; writes remain handler-gated |
| Audit visibility | `audit-log:view` | `FINANCE_OPS`, `ERP_ADMIN`, `CEO` where already allowed |

## Implementation

Backend RBAC catalog:

```text
apps/api/internal/shared/auth/rbac.go
```

Frontend permission catalog:

```text
apps/web/src/shared/permissions/menu.ts
```

OpenAPI role/permission enum:

```text
packages/openapi/openapi.yaml
apps/web/src/shared/api/generated/schema.ts
```

## Guardrails

```text
PURCHASE_OPS does not receive qc:decision.
PURCHASE_OPS does not receive finance:view.
FINANCE_OPS does not receive record:create.
FINANCE_OPS does not receive qc:decision.
WAREHOUSE_LEAD still does not receive qc:decision.
ERP_ADMIN remains the break-glass/admin role.
```

## Verification

Backend tests updated:

```text
apps/api/internal/shared/auth/auth_test.go
```

Frontend tests updated:

```text
apps/web/src/shared/permissions/menu.test.ts
```
