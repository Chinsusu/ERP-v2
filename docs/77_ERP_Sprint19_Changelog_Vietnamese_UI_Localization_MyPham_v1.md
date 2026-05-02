# 77_ERP_Sprint19_Changelog_Vietnamese_UI_Localization_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 19 - Vietnamese UI Localization
Document role: Sprint changelog and release evidence
Version: v1.0
Date: 2026-05-02
Status: Released as `v0.19.0-vietnamese-ui-localization` after cloud CI passed on the Sprint 19 release commit

---

## 1. Sprint 19 Scope

Sprint 19 localized the ERP frontend into Vietnamese-first UI while keeping technical contracts stable:

```text
Frontend display labels: Vietnamese-first
Routes: English technical routes unchanged
Backend/API/DB enum values: unchanged
Permission keys and audit event codes: unchanged
Locale: vi-VN
Timezone: Asia/Ho_Chi_Minh
Currency display: VND
```

Promoted scope:

```text
i18n foundation and dictionaries
Vietnamese navigation and app shell labels
Vietnamese dashboard, warehouse, sales, shipping, returns, purchase, QC, master data, inventory, auth, audit, and attachment UI
Shared ERP number/date/money/quantity/rate formatting
Status, error, validation, action, and unit label mapping
Localization smoke and hardcoded-English regression checks
README current status update
```

Out of scope:

```text
Route localization
Backend enum renaming
Database value renaming
OpenAPI schema renaming
Permission key renaming
Audit event code renaming
Full multi-language admin switching
```

---

## 2. Merged PRs

| Task | PR | Result |
| --- | --- | --- |
| S19-01/S19-04 localization foundation and formatters | earlier S19 PRs | Added shared i18n dictionaries, label helpers, Ant Design-ready Vietnamese UI foundation, and ERP display formatters |
| S19-05 dashboard and warehouse localization | earlier S19 PRs | Localized dashboard, Warehouse Daily Board, and shift closing surfaces |
| S19-06 sales, picking, packing, and handover localization | earlier S19 PRs | Localized order fulfillment, pick/pack, carrier manifest, scan handover, and exception copy |
| S19-07 returns localization | #518, #519 | Localized return inspection, disposition, stock count, and adjustment UI |
| S19-08 purchase and inbound QC localization | #520, #521, #522 | Localized purchase order, goods receiving, and inbound QC UI |
| S19-09 master data and inventory localization | #523, #524, #525, #526 | Localized product, party, warehouse/location master data, inventory, available stock, and batch QC audit UI |
| S19-10 auth, audit, and attachment localization | #527, #528, #529 | Localized login, audit log, attachment panels, return evidence, and supplier rejection evidence UI |
| S19-11 localization tests/checklist | #530 | Added i18n smoke tests, hardcoded-English checklist, and broader label mapping test coverage |
| S19-12 README and changelog | #531 | Updated README current status and created this Sprint 19 changelog |

All PRs used the manual review and merge flow. GitHub auto review and auto merge were not used.

---

## 3. UI Contract

Stable technical contracts:

```text
/dashboard
/master-data
/inventory
/sales
/shipping
/returns
/purchase
/qc
/finance
/audit-log
/settings
```

Display behavior:

```text
App shell/menu labels are Vietnamese.
Module page headers and operational copy are Vietnamese.
Status chips use centralized Vietnamese labels.
API error and validation messages use centralized Vietnamese mappings.
Audit/event codes remain English and traceable where needed.
Money/date/quantity/rate display follows vi-VN and Asia/Ho_Chi_Minh standards.
```

---

## 4. Verification Evidence

Cloud PR checks:

```text
#523 required-ci, web-ci, e2e-ci: success
#524 required-ci, web-ci, e2e-ci: success
#525 required-ci, web-ci, e2e-ci: success
#526 required-ci, web-ci, e2e-ci: success
#527 required-ci, web-ci, e2e-ci: success
#528 required-ci, web-ci, e2e-ci: success
#529 required-ci, web-ci, e2e-ci: success
#530 required-ci, web-ci, e2e-ci: success after import fix
```

Latest verified Sprint 19 release gate:

```text
release commit df9b9567
release tag v0.19.0-vietnamese-ui-localization
required-ci run 25259927794: success
required-api: success
required-web: success
required-openapi: success
required-migration: success
required-migration PostgreSQL 16 apply plus rollback: success
```

Local checks performed during S19 tasks:

```text
JSON dictionary parse checks with PowerShell ConvertFrom-Json
git diff --check and git diff --cached --check
targeted hardcoded-label searches in changed UI components
PowerShell mirror of the hardcoded-English checklist for #530
```

Local limitation:

```text
Host pnpm is not in PATH.
apps/web/node_modules/.bin/vitest.cmd returns Access is denied on this Windows host.
Frontend test execution was verified by GitHub Actions web-ci/required-web.
```

---

## 5. Release Recommendation

Sprint 19 tag:

```text
v0.19.0-vietnamese-ui-localization -> df9b9567
```

Do not treat this as a backend persistence release. Sprint 19 is a UI localization release on top of the Sprint 18 runtime persistence baseline.

---

## 6. Next Sprint

Proceed to:

```text
Sprint 20 - Release Hygiene + API Modularization + Fallback Cleanup
```

Priority items:

```text
README/release hygiene follow-through
Migration apply/down/reapply CI gate
Node.js 24 GitHub Actions compatibility check
cmd/api/main.go route registration modularization
Frontend fallback service production gating
Production runtime mode checklist
```
