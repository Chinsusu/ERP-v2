# S17-01-01 Master Data Persistence Design

Project: Web ERP for cosmetics operations
Sprint: Sprint 17 - Master data runtime store persistence
Task: S17-01-01 Master data persistence design
Date: 2026-05-02
Status: Ready for S17-01-02 migration foundation

---

## 1. Purpose

Persist the runtime master data catalogs that are still prototype-only after Sprint 16:

```text
Item catalog
Warehouse/location catalog
Party catalog
UOM catalog
```

The design keeps existing domain validation, API response shapes, audit actions, and no-DB/local prototype fallback behavior. Sprint 17 should only change DB-mode durability.

---

## 2. Current Runtime Gap

Current API startup creates master data catalogs directly in `apps/api/cmd/api/main.go`:

```text
masterdataapp.NewPrototypeItemCatalog(auditLogStore)
masterdataapp.NewPrototypeUOMCatalog()
masterdataapp.NewPrototypeWarehouseLocationCatalog(auditLogStore)
masterdataapp.NewPrototypePartyCatalog(auditLogStore)
```

Risk:

```text
Operational documents now persist, but edited item, warehouse/location, supplier/customer, and UOM conversion data can reset after API restart.
```

---

## 3. Runtime Selection Rule

Add a package-level runtime selector:

```text
apps/api/cmd/api/masterdata_store_selection.go
```

Expected behavior:

```text
if DATABASE_URL is empty:
  use existing prototype item, warehouse/location, party, and UOM catalogs

if DATABASE_URL exists:
  open one PostgreSQL connection
  select PostgreSQL-backed item, warehouse/location, party, and UOM catalogs together
  return one close function for the shared DB handle
```

This avoids partial master data truth. DB mode must not wire one master data catalog to PostgreSQL while the other master data catalogs remain memory-backed.

---

## 4. PostgreSQL Schema Plan

Migration:

```text
apps/api/migrations/000034_persist_master_data_runtime_foundation.up.sql
apps/api/migrations/000034_persist_master_data_runtime_foundation.down.sql
```

Schema:

```text
masterdata
```

Tables:

| Table | Purpose |
| --- | --- |
| `masterdata.items` | Item/SKU catalog state |
| `masterdata.warehouses` | Warehouse header state |
| `masterdata.warehouse_locations` | Location/bin state under warehouse |
| `masterdata.suppliers` | Supplier master data |
| `masterdata.customers` | Customer master data |
| `masterdata.uoms` | UOM definitions |
| `masterdata.uom_conversions` | Global and item-specific UOM conversion factors |

Constraints:

```text
items.item_code unique
items.sku_code unique
warehouses.code unique
warehouse_locations unique by warehouse_id + code
suppliers.code unique
customers.code unique
uoms.code primary key
uom_conversions unique by item_id + from_uom_code + to_uom_code
```

Decimal columns follow file 40:

```text
items.standard_cost numeric(18,6)
suppliers.moq numeric(18,6)
suppliers.quality_score numeric(9,4)
suppliers.delivery_score numeric(9,4)
customers.credit_limit numeric(18,2)
uom_conversions.factor numeric(18,6)
```

---

## 5. Store Plan

### Item Catalog

File:

```text
apps/api/internal/modules/masterdata/application/postgres_item_catalog.go
```

Must preserve:

```text
List
Get
Create
Update
ChangeStatus
ErrItemNotFound
ErrDuplicateItemCode
ErrDuplicateSKUCode
domain validation errors
audit actions: masterdata.item.created, masterdata.item.updated, masterdata.item.status_changed
```

### Warehouse/Location Catalog

File:

```text
apps/api/internal/modules/masterdata/application/postgres_warehouse_location_catalog.go
```

Must preserve:

```text
ListWarehouses
GetWarehouse
CreateWarehouse
UpdateWarehouse
ChangeWarehouseStatus
ListLocations
GetLocation
CreateLocation
UpdateLocation
ChangeLocationStatus
ErrWarehouseNotFound
ErrLocationNotFound
ErrDuplicateWarehouseCode
ErrDuplicateLocationCode
ErrInvalidLocationWarehouse
ErrInactiveLocation
warehouse code propagation to locations on warehouse update
audit actions for warehouse and location lifecycle
```

### Party Catalog

File:

```text
apps/api/internal/modules/masterdata/application/postgres_party_catalog.go
```

Must preserve:

```text
ListSuppliers
GetSupplier
CreateSupplier
UpdateSupplier
ChangeSupplierStatus
ListCustomers
GetCustomer
CreateCustomer
UpdateCustomer
ChangeCustomerStatus
ErrSupplierNotFound
ErrCustomerNotFound
ErrDuplicateSupplierCode
ErrDuplicateCustomerCode
domain validation and status-transition errors
audit actions for supplier and customer lifecycle
```

### UOM Catalog

File:

```text
apps/api/internal/modules/masterdata/application/postgres_uom_catalog.go
```

Must preserve:

```text
ConvertToBase
UpsertConversion
global conversions
item-specific conversion precedence over global conversion
inactive conversion error behavior
no float/double usage
```

UOM catalog has no public CRUD handlers today. Sprint 17 persistence should seed Phase 1 UOM definitions and conversions into PostgreSQL when missing, then read/write conversions through PostgreSQL in DB mode.

---

## 6. Seeding Rule

PostgreSQL stores should preserve current prototype seed availability for DB-mode dev/test startup.

Recommended approach:

```text
constructor ensures seed rows exist with INSERT ... ON CONFLICT DO NOTHING
seed source reuses existing prototype seed helpers or equivalent values
seed insertion is idempotent
created_at/updated_at values match existing prototype seed intent where relevant
```

This avoids breaking current dev smoke and E2E tests that rely on known master data such as:

```text
item-serum-30ml
wh-hcm-fg
loc-hcm-fg-recv-01
sup-rm-bioactive
cus-dl-minhanh
KG -> G
CARTON -> PCS for item-serum-30ml
```

---

## 7. Test Plan

Focused PostgreSQL lifecycle tests:

| Test file | Coverage |
| --- | --- |
| `postgres_item_catalog_test.go` | create/update/status, duplicate item code/SKU, fresh store reload |
| `postgres_warehouse_location_catalog_test.go` | warehouse/location create/update/status, duplicate checks, invalid warehouse, warehouse-code propagation, fresh reload |
| `postgres_party_catalog_test.go` | supplier/customer create/update/status, duplicate checks, status-transition errors, fresh reload |
| `postgres_uom_catalog_test.go` | seeded UOMs, global conversion, item-specific conversion, upsert conversion, fresh reload |
| `masterdata_store_selection_test.go` | no-DB selects prototype package, DB config selects PostgreSQL package |

Dev smoke:

```text
Create or update one item.
Create or update one warehouse and location.
Create or update one supplier and customer.
Upsert or verify one UOM conversion.
Restart API.
Read the same records back.
Check audit rows for item/warehouse/location/supplier/customer actions.
```

---

## 8. Migration Gate

When migration `000034` exists, release evidence must include:

```text
PostgreSQL 16 isolated instance
apply all up migrations through 000034
rollback all down migrations through 000034
record applied and rolled-back migration count
```

GitHub Actions cloud CI is currently blocked by exhausted monthly minutes. Do not claim cloud CI verification until Actions minutes are available again.

---

## 9. Rollout Notes

Sprint 17 should merge in dependency order:

```text
S17-01-02 migration foundation
S17-02-01/S17-02-02 item catalog store and tests
S17-03-01/S17-03-02 warehouse/location store and tests
S17-04-01/S17-04-02 party store and tests
S17-05-01/S17-05-02 UOM store and tests
S17-06-01/S17-06-02 runtime selector and smoke
S17-07-01/S17-08-01 ledger and changelog
```

Production tag remains on hold while cloud CI is unavailable unless the team explicitly accepts a manual-only release gate.
