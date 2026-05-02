# 72_ERP_Sprint17_Changelog_Master_Data_Runtime_Store_Persistence_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 17 - Master data runtime store persistence v1
Document role: Sprint changelog and release evidence
Version: v1.0
Date: 2026-05-02
Status: Release evidence complete except cloud CI; production tag is on hold while GitHub Actions minutes are exhausted

---

## 1. Sprint 17 Scope

Sprint 17 closed the highest remaining master data reset risk after Sprint 16:

```text
item catalog was prototype-only
warehouse/location catalog was prototype-only
party catalog was prototype-only
UOM catalog was prototype-only
runtime API selected prototype master data even when operational stores used PostgreSQL
full dev smoke did not prove master data state survived API restart
```

Promoted scope:

```text
Sprint 17 task board
master data persistence design
master data migration foundation
PostgreSQL-backed item catalog store
PostgreSQL-backed warehouse/location catalog store
PostgreSQL-backed party catalog store
PostgreSQL-backed UOM catalog store
service/store lifecycle tests for master data stores
package-level runtime master data store selector
startup seed compatibility for existing prototype baseline references
full dev master data persistence smoke through API restart
remaining prototype store ledger update
Sprint 17 release evidence
```

No new frontend master data screens, pricing policy redesign, supplier CRM workflows, barcode generation, warehouse bin optimization, multi-currency master data, or production auth/session persistence was introduced. Sprint 17 changed persistence behavior behind the existing master data APIs and downstream runtime references.

---

## 2. Merged PRs

| Task | PR | Result |
| --- | --- | --- |
| S17-00-00 Sprint 17 task board | #485 | Created Sprint 17 task board |
| S17-01-01 Master data persistence design | #486 | Documented schema mapping, selectors, fallback behavior, tests, smoke, and rollback |
| S17-01-02 Master data migration foundation | #487 | Added migration `000034_persist_master_data_runtime_foundation` |
| S17-02-01 / S17-02-02 Item catalog PostgreSQL store and tests | #488 | Added PostgreSQL-backed item lifecycle, duplicate checks, status, UOM fields, cost, and audit |
| S17-03-01 / S17-03-02 Warehouse/location PostgreSQL store and tests | #489 | Added PostgreSQL-backed warehouse and location lifecycle, hierarchy, status, and audit |
| S17-04-01 / S17-04-02 Party catalog PostgreSQL store and tests | #490 | Added PostgreSQL-backed supplier/customer lifecycle, group/type, metrics, terms, status, and audit |
| S17-05-01 / S17-05-02 UOM catalog PostgreSQL store and tests | #491 | Added PostgreSQL-backed UOM definitions and conversion factors without float/double |
| S17-06-01 / S17-06-02 Package runtime selectors and restart smoke | #492 | Wired DB-mode master data stores as one package and added full dev restart smoke |
| S17-07-01 Remaining prototype ledger update | #493 | Removed master data catalogs from remaining production persistence gaps |
| S17-08-01 Sprint 17 release evidence | #494 | Records release evidence, CI quota blocker, remaining gaps, and production tag hold |

All PRs used the manual review and merge flow. GitHub auto review and auto merge were not used.

---

## 3. Persistence Changes

### Runtime Selector

| Runtime path | DB mode | No-DB/local fallback |
| --- | --- | --- |
| `masterDataStores.items` | `PostgresItemCatalog` | `ItemCatalog` prototype |
| `masterDataStores.uoms` | `PostgresUOMCatalog` | `UOMCatalog` prototype |
| `masterDataStores.warehouses` | `PostgresWarehouseLocationCatalog` | `WarehouseLocationCatalog` prototype |
| `masterDataStores.parties` | `PostgresPartyCatalog` | `PartyCatalog` prototype |

DB mode selects all master data stores as one package. Prototype fallback remains intentional for no-DB/local mode and is not production persistence evidence.

### PostgreSQL Persistence

| Migration | Purpose |
| --- | --- |
| `000002_create_phase1_base_tables` | Existing master data baseline tables in the `mdm` schema |
| `000034_persist_master_data_runtime_foundation` | Extends master data tables with stable runtime refs, item/UOM/party/warehouse fields, indexes, constraints, and conversion refs |

Persisted behavior:

```text
GET   /api/v1/products
POST  /api/v1/products
GET   /api/v1/products/{product_id}
PATCH /api/v1/products/{product_id}
PATCH /api/v1/products/{product_id}/status
GET   /api/v1/warehouses
POST  /api/v1/warehouses
GET   /api/v1/warehouses/{warehouse_id}
PATCH /api/v1/warehouses/{warehouse_id}
PATCH /api/v1/warehouses/{warehouse_id}/status
GET   /api/v1/warehouse-locations
POST  /api/v1/warehouse-locations
GET   /api/v1/warehouse-locations/{location_id}
PATCH /api/v1/warehouse-locations/{location_id}
PATCH /api/v1/warehouse-locations/{location_id}/status
GET   /api/v1/suppliers
POST  /api/v1/suppliers
GET   /api/v1/suppliers/{supplier_id}
PATCH /api/v1/suppliers/{supplier_id}
PATCH /api/v1/suppliers/{supplier_id}/status
GET   /api/v1/customers
POST  /api/v1/customers
GET   /api/v1/customers/{customer_id}
PATCH /api/v1/customers/{customer_id}
PATCH /api/v1/customers/{customer_id}/status
```

Persisted evidence:

```text
mdm.items
mdm.units
mdm.uoms
mdm.uom_conversions
mdm.warehouses
mdm.warehouse_bins
mdm.suppliers
mdm.customers
audit.audit_logs master data lifecycle actions
```

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
After PR #492, main was synced/deployed to dev at commit 616ca760.
Migration `000034` was applied to the dev PostgreSQL database.
API was rebuilt and restarted from clean main.
Full dev smoke passed after the final master data runtime selector merge.
PR #493 was docs-only and synced to dev at commit 4b86e22b; no runtime rebuild was required.
This S17-08 task is documentation-only and does not require a runtime redeploy.
```

Latest Sprint 17 smoke evidence on dev:

```text
masterdata_item_create 201
masterdata_wh_create 201
masterdata_loc_create 201
masterdata_supplier_create 201
masterdata_customer_create 201
api_restart ok
masterdata_item_read 200
masterdata_wh_read 200
masterdata_loc_read 200
masterdata_supplier_read 200
masterdata_customer_read 200
persisted_masterdata ok 0002
Full ERP dev smoke passed
```

The same full smoke also passed the previously persisted audit, finance, sales reservation/order, stock adjustment/movement/count, purchase order, inbound QC, carrier manifest, pick task, pack task, return receipt, supplier rejection, and subcontract checks.

---

## 5. CI And Migration Evidence

GitHub Actions status:

```text
Cloud CI is blocked for Sprint 17 PRs because the GitHub Actions plan has used 100% of the included monthly minutes.
Quota message: 2,000 min used / 2,000 min included.
Do not treat Sprint 17 as production-tagged while this CI gate is blocked.
```

Local/dev verification highlights:

```text
S17-01-02: PostgreSQL 16 isolated migration apply/rollback passed.
S17-02-01: masterdata application tests passed, including Postgres item catalog integration.
S17-03-01: masterdata application tests passed, including Postgres warehouse/location catalog integration.
S17-04-01: masterdata application tests passed, including Postgres party catalog integration.
S17-05-01: masterdata application tests passed, including Postgres UOM catalog integration.
S17-06-01: dev server go test ./internal/modules/masterdata/application -count=1 passed.
S17-06-01: dev server go test ./cmd/api -count=1 passed.
S17-06-01: dev server go vet ./cmd/api ./internal/modules/masterdata/application passed.
S17-06-01: dev migration up to 000034 passed.
S17-06-01: dev API build/restart passed.
S17-06-02: full dev smoke passed on clean main commit 616ca760, including persisted_masterdata ok 0002.
S17-07-01: remaining prototype ledger updated; documentation-only change.
S17-08-01: changelog created with CI blocked and tag hold recorded.
```

Migration runtime gate:

```text
PostgreSQL 16 isolated container
PostgreSQL version: 16.13
Source: /opt/ERP-v2 during S17-01-02
Action: apply every *.up.sql in order, then apply every *.down.sql in reverse order
Result: passed
Applied migrations: 34
Rolled back migrations: 34
```

---

## 6. Remaining Prototype Stores

Current remaining-store ledger:

```text
docs/qa/S14-04-01_remaining_prototype_store_ledger.md
```

Highest remaining persistence candidates after Sprint 17:

```text
1. Auth/session hardening.
2. Remove or gate frontend fallback services where backend coverage is now available.
```

Item catalog, warehouse/location catalog, party catalog, and UOM catalog are no longer listed as production persistence gaps when DB config exists.

---

## 7. Release Status

Sprint 17 release gate status:

```text
Task PRs: merged through S17-08-01 PR #494
Main cloud CI: blocked by exhausted GitHub Actions minutes
Dev runtime smoke: green at runtime commit 616ca760
Docs-only main sync: green after prototype ledger merge at 4b86e22b
Migration apply/rollback: green on PostgreSQL 16 isolated instance
Production tag: HOLD
```

Recommended production tag after CI is available and green:

```text
v0.17.0-master-data-runtime-store-persistence
```

Do not create the tag until either:

```text
1. GitHub Actions minutes reset or billing is fixed, required checks run, and the checks are green; or
2. the team explicitly accepts a manual-only release gate for this sprint.
```
