BEGIN;

DROP INDEX IF EXISTS mdm.ix_uom_conversions_runtime_item_ref;
DROP INDEX IF EXISTS mdm.uq_uom_conversions_ref;
DROP INDEX IF EXISTS mdm.ix_customers_runtime_filters;
DROP INDEX IF EXISTS mdm.uq_customers_org_ref;
DROP INDEX IF EXISTS mdm.ix_suppliers_runtime_filters;
DROP INDEX IF EXISTS mdm.uq_suppliers_org_ref;
DROP INDEX IF EXISTS mdm.ix_warehouse_bins_runtime_filters;
DROP INDEX IF EXISTS mdm.uq_warehouse_bins_org_location_ref;
DROP INDEX IF EXISTS mdm.ix_warehouses_runtime_filters;
DROP INDEX IF EXISTS mdm.uq_warehouses_org_ref;
DROP INDEX IF EXISTS mdm.ix_items_runtime_filters;
DROP INDEX IF EXISTS mdm.uq_items_org_item_code;
DROP INDEX IF EXISTS mdm.uq_items_org_item_ref;

ALTER TABLE mdm.uom_conversions
  DROP CONSTRAINT IF EXISTS ck_uom_conversions_runtime_required,
  DROP COLUMN IF EXISTS item_ref,
  DROP COLUMN IF EXISTS conversion_ref;

ALTER TABLE mdm.customers
  DROP CONSTRAINT IF EXISTS ck_customers_runtime_required,
  DROP CONSTRAINT IF EXISTS ck_customers_status,
  ADD CONSTRAINT ck_customers_status CHECK (status IN ('active', 'inactive', 'blocked')),
  DROP COLUMN IF EXISTS tax_code,
  DROP COLUMN IF EXISTS contact_name,
  DROP COLUMN IF EXISTS payment_terms,
  DROP COLUMN IF EXISTS credit_limit,
  DROP COLUMN IF EXISTS discount_group,
  DROP COLUMN IF EXISTS price_list_code,
  DROP COLUMN IF EXISTS channel_code,
  DROP COLUMN IF EXISTS customer_type,
  DROP COLUMN IF EXISTS customer_ref;

ALTER TABLE mdm.suppliers
  DROP CONSTRAINT IF EXISTS ck_suppliers_runtime_required,
  DROP CONSTRAINT IF EXISTS ck_suppliers_status,
  ADD CONSTRAINT ck_suppliers_status CHECK (status IN ('active', 'inactive', 'blocked')),
  DROP COLUMN IF EXISTS delivery_score,
  DROP COLUMN IF EXISTS quality_score,
  DROP COLUMN IF EXISTS moq,
  DROP COLUMN IF EXISTS lead_time_days,
  DROP COLUMN IF EXISTS payment_terms,
  DROP COLUMN IF EXISTS contact_name,
  DROP COLUMN IF EXISTS supplier_group,
  DROP COLUMN IF EXISTS supplier_ref;

ALTER TABLE mdm.warehouse_bins
  DROP CONSTRAINT IF EXISTS ck_warehouse_bins_runtime_required,
  DROP COLUMN IF EXISTS is_default,
  DROP COLUMN IF EXISTS allow_store,
  DROP COLUMN IF EXISTS allow_pick,
  DROP COLUMN IF EXISTS allow_receive,
  DROP COLUMN IF EXISTS zone_code,
  DROP COLUMN IF EXISTS location_ref;

ALTER TABLE mdm.warehouses
  DROP CONSTRAINT IF EXISTS ck_warehouses_runtime_required,
  DROP COLUMN IF EXISTS allow_quarantine,
  DROP COLUMN IF EXISTS allow_prod_issue,
  DROP COLUMN IF EXISTS allow_sale_issue,
  DROP COLUMN IF EXISTS site_code,
  DROP COLUMN IF EXISTS warehouse_type,
  DROP COLUMN IF EXISTS warehouse_ref;

ALTER TABLE mdm.items
  DROP CONSTRAINT IF EXISTS ck_items_runtime_required,
  DROP CONSTRAINT IF EXISTS ck_items_status,
  ADD CONSTRAINT ck_items_status CHECK (status IN ('active', 'inactive', 'blocked')),
  DROP COLUMN IF EXISTS spec_version,
  DROP COLUMN IF EXISTS is_producible,
  DROP COLUMN IF EXISTS is_purchasable,
  DROP COLUMN IF EXISTS is_sellable,
  DROP COLUMN IF EXISTS standard_cost,
  DROP COLUMN IF EXISTS qc_required,
  DROP COLUMN IF EXISTS expiry_controlled,
  DROP COLUMN IF EXISTS lot_controlled,
  DROP COLUMN IF EXISTS uom_issue,
  DROP COLUMN IF EXISTS uom_purchase,
  DROP COLUMN IF EXISTS uom_base,
  DROP COLUMN IF EXISTS brand_code,
  DROP COLUMN IF EXISTS item_group,
  DROP COLUMN IF EXISTS item_code,
  DROP COLUMN IF EXISTS item_ref;

COMMIT;
