# 66_ERP_Sprint14_Changelog_Shipping_Pick_Pack_Persistence_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 14 - Shipping pick/pack persistence v1
Document role: Sprint changelog and release evidence
Version: v1.0
Date: 2026-05-02
Status: Release evidence prepared; production tag pending after this changelog merges and main CI is green

---

## 1. Sprint 14 Scope

Sprint 14 closed the highest remaining warehouse execution reset risk after Sprint 13:

```text
carrier manifests were prototype-only
pick task progress was prototype-only
pack task progress was prototype-only
shipping state could reset while sales orders, reservations, returns, and daily close evidence persisted
```

Promoted scope:

```text
Sprint 14 task board
shipping persistence design
runtime selectors for carrier manifests, pick tasks, and pack tasks
PostgreSQL-backed carrier manifest, pick task, and pack task stores
shipping runtime ref migrations and actor-reference hardening
focused shipping store tests
dev persistence smoke for manifest handover/scan and pick/pack progress
sales order pack integration hardening
remaining prototype store ledger update
```

No new shipping frontend screen, carrier integration API, route planning, rate shopping, or label-purchase workflow was introduced. Sprint 14 changed persistence behavior behind the existing shipping APIs.

---

## 2. Merged PRs

| Task | PR | Result |
| --- | --- | --- |
| S14-00-00 Sprint 14 task board | #440 | Created Sprint 14 task board |
| S14-01-01 Shipping persistence design | #441 | Documented schema, selectors, store behavior, migration, tests, smoke, and guardrails |
| S14-01-02 Runtime contracts and selectors | #442 | Added DB-mode runtime selectors with explicit prototype fallback |
| S14-01-03 Carrier manifest PostgreSQL store | #443 | Added migration 000027 and PostgreSQL-backed carrier manifest persistence |
| S14-01-04 Pick task PostgreSQL store | #444 | Added migration 000028 and PostgreSQL-backed pick task persistence |
| S14-01-05 Pack task PostgreSQL store | #445 | Added migration 000029 and PostgreSQL-backed pack task persistence |
| S14-01-06 Store and handler tests | #446 | Added focused shipping persistence coverage |
| S14-02-01 Manifest handover/scan persistence smoke | #447 | Proved manifest handover/scan state through full dev smoke |
| S14-02-02 Pick/pack persistence smoke | #448 | Added pick and pack persistence checks to full dev smoke and migration 000030 actor-reference hardening |
| S14-03-01 Shipping package integration check | #449 | Preserved sales order line row IDs during pack confirm and verified sales order packed integration |
| S14-04-01 Remaining prototype ledger update | #450 | Superseded Sprint 13 remaining-store ledger after shipping persistence |

All PRs used the manual review and merge flow.

---

## 3. Persistence Changes

### Runtime Selectors

| Runtime path | DB mode | No-DB/local fallback |
| --- | --- | --- |
| `newRuntimeCarrierManifestStore` | `PostgresCarrierManifestStore` | `PrototypeCarrierManifestStore` |
| `newRuntimePickTaskStore` | `PostgresPickTaskStore` | `PrototypePickTaskStore` |
| `newRuntimePackTaskStore` | `PostgresPackTaskStore` | `PrototypePackTaskStore` |

Prototype fallback remains intentional for no-DB/local mode and is not production persistence evidence.

### PostgreSQL Persistence

| Migration | Purpose |
| --- | --- |
| `000027_persist_carrier_manifests` | Adds runtime refs and indexes for carrier manifests, manifest orders, scan events, and shipping exceptions |
| `000028_persist_pick_tasks` | Adds runtime refs and indexes for pick task headers, lines, and exceptions |
| `000029_persist_pack_tasks` | Adds runtime refs and indexes for pack task headers, lines, and exceptions |
| `000030_harden_pick_pack_actor_refs` | Allows persisted pick/pack lifecycle state to keep actor refs when UUID actor columns are unavailable |

Persisted behavior:

```text
GET  /api/v1/shipping/manifests
POST /api/v1/shipping/manifests
POST /api/v1/shipping/manifests/{id}/shipments
POST /api/v1/shipping/manifests/{id}/ready
POST /api/v1/shipping/manifests/{id}/scan
POST /api/v1/shipping/manifests/{id}/confirm-handover
GET  /api/v1/pick-tasks
POST /api/v1/pick-tasks/{id}/start
POST /api/v1/pick-tasks/{id}/confirm-line
POST /api/v1/pick-tasks/{id}/complete
GET  /api/v1/pack-tasks
POST /api/v1/pack-tasks/{id}/start
POST /api/v1/pack-tasks/{id}/confirm
```

Persisted evidence:

```text
shipping.carrier_manifests
shipping.carrier_manifest_orders
shipping.scan_events
shipping.shipping_exceptions
shipping.pick_tasks
shipping.pick_task_lines
shipping.pack_tasks
shipping.pack_task_lines
audit.audit_logs shipping manifest/pick/pack actions
sales.sales_orders packed and handed-over integration state
```

S14-03-01 also preserved existing `sales.sales_order_lines.id` values when pack confirmation saves sales order state. That keeps shipping pick/pack evidence references valid while still deleting intentionally removed stale lines.

---

## 4. Dev Release Evidence

Dev server:

```text
Host: 10.1.1.120
Repo: /opt/ERP-v2
Runtime dev URL: http://10.1.1.120:8088
```

Runtime deploy evidence:

```text
After PR #450, main was deployed to dev at commit 264f3af.
Deploy built API, worker, and web images from source after GHCR dev-image pull warnings.
Migrations through 000030 were present on dev.
Deploy smoke passed.
Full host smoke passed.
```

Release smoke at current Sprint 14 main:

```text
Commit: 264f3af
Command: infra/scripts/smoke-dev-full.sh
Result: Full ERP dev smoke passed
```

Full dev smoke included these persisted checks:

```text
persisted_audit_login
persisted_sales_reservation
persisted_sales_order
persisted_stock_adjustment
persisted_stock_movement
persisted_available_stock
persisted_stock_count
persisted_purchase_order
persisted_inbound_qc
persisted_carrier_manifest
persisted_pick_task
persisted_pack_task
persisted_return_receipt
persisted_supplier_rejection
```

Latest Sprint 14 smoke evidence on dev:

```text
persisted_carrier_manifest ok manifest-s14-02-01-smoke-0013
persisted_pick_task        ok PICK-S14-02-02-SMOKE-0010
persisted_pack_task        ok PACK-S14-02-02-SMOKE-0010
```

---

## 5. CI And Migration Evidence

GitHub checks:

```text
PR #440: required-api, required-migration, required-openapi, required-web passed.
PR #441: e2e, required-api, required-migration, required-openapi, required-web passed.
PR #442: api, e2e, openapi, required-api, required-migration, required-openapi, required-web passed.
PR #443: api, e2e, migration, openapi, required-api, required-migration, required-openapi, required-web passed.
PR #444: api, e2e, migration, openapi, required-api, required-migration, required-openapi, required-web passed.
PR #445: api, e2e, migration, openapi, required-api, required-migration, required-openapi, required-web passed.
PR #446: api, e2e, required-api, required-migration, required-openapi, required-web passed.
PR #447: required-api, required-migration, required-openapi, required-web passed.
PR #448: api, e2e, migration, required-api, required-migration, required-openapi, required-web passed.
PR #449: api, e2e, required-api, required-migration, required-openapi, required-web passed.
PR #450: e2e, required-api, required-migration, required-openapi, required-web passed.
```

Local/dev verification:

```text
S14-03-01: git diff --check passed.
S14-03-01: sh -n infra/scripts/smoke-dev-full.sh passed on dev server.
S14-03-01: dev branch deploy passed.
S14-03-01: full dev smoke passed.
S14-03-01: API package tests via Go 1.23 container passed with go test ./...
S14-04-01: git diff --cached --check passed.
S14-04-01: documentation-only change; no runtime build test required.
```

Migration runtime gate:

```text
PostgreSQL 16 isolated container
Source: /opt/ERP-v2 at main commit 264f3af
Action: apply every *.up.sql in order, then apply every *.down.sql in reverse order
Result: passed
Applied migrations: 30
Rolled back migrations: 30
```

Migration notices observed during rollback were idempotent schema/constraint notices from earlier migrations. They did not fail the gate.

---

## 6. Remaining Prototype Stores

Current remaining-store ledger:

```text
docs/qa/S14-04-01_remaining_prototype_store_ledger.md
```

Highest remaining persistence candidates after Sprint 14:

```text
1. Finance AR/AP/COD/cash runtime stores.
2. Subcontract runtime stores.
3. Master data catalogs.
4. Auth/session hardening.
```

Shipping manifest, pick, and pack task stores are no longer listed as production persistence gaps when DB config exists.

---

## 7. Release Status

Sprint 14 release gate status:

```text
Task PRs: merged through S14-04-01
Current changelog PR: pending merge
Main CI: green through PR #450; rerun required for this changelog PR
Dev runtime smoke: green at commit 264f3af
Migration apply/rollback: green on PostgreSQL 16 isolated instance
Production tag: pending
```

Recommended tag after this changelog merges and main CI is green:

```text
v0.14.0-shipping-pick-pack-persistence
```

Do not move the tag once pushed. If a post-tag fix is needed, create a new patch tag instead.
