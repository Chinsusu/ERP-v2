BEGIN;

ALTER TABLE mdm.items
  ADD COLUMN IF NOT EXISTS item_ref text,
  ADD COLUMN IF NOT EXISTS item_code text,
  ADD COLUMN IF NOT EXISTS item_group text,
  ADD COLUMN IF NOT EXISTS brand_code text,
  ADD COLUMN IF NOT EXISTS uom_base text,
  ADD COLUMN IF NOT EXISTS uom_purchase text,
  ADD COLUMN IF NOT EXISTS uom_issue text,
  ADD COLUMN IF NOT EXISTS lot_controlled boolean NOT NULL DEFAULT true,
  ADD COLUMN IF NOT EXISTS expiry_controlled boolean NOT NULL DEFAULT true,
  ADD COLUMN IF NOT EXISTS qc_required boolean NOT NULL DEFAULT true,
  ADD COLUMN IF NOT EXISTS standard_cost numeric(18,6) NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS is_sellable boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS is_purchasable boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS is_producible boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS spec_version text;

UPDATE mdm.items
SET item_ref = COALESCE(NULLIF(btrim(item_ref), ''), id::text),
    item_code = COALESCE(NULLIF(btrim(item_code), ''), sku),
    uom_base = COALESCE(NULLIF(btrim(uom_base), ''), 'PCS'),
    uom_purchase = COALESCE(NULLIF(btrim(uom_purchase), ''), COALESCE(NULLIF(btrim(uom_base), ''), 'PCS')),
    uom_issue = COALESCE(NULLIF(btrim(uom_issue), ''), COALESCE(NULLIF(btrim(uom_base), ''), 'PCS')),
    lot_controlled = requires_batch,
    expiry_controlled = requires_expiry
WHERE item_ref IS NULL
   OR item_code IS NULL
   OR uom_base IS NULL
   OR uom_purchase IS NULL
   OR uom_issue IS NULL;

ALTER TABLE mdm.items
  DROP CONSTRAINT IF EXISTS ck_items_status;

ALTER TABLE mdm.items
  ADD CONSTRAINT ck_items_status CHECK (
    status IN ('draft', 'active', 'inactive', 'obsolete', 'blocked')
  );

ALTER TABLE mdm.items
  DROP CONSTRAINT IF EXISTS ck_items_runtime_required;

ALTER TABLE mdm.items
  ADD CONSTRAINT ck_items_runtime_required CHECK (
    nullif(btrim(item_ref), '') IS NOT NULL
    AND nullif(btrim(item_code), '') IS NOT NULL
    AND nullif(btrim(uom_base), '') IS NOT NULL
    AND nullif(btrim(uom_purchase), '') IS NOT NULL
    AND nullif(btrim(uom_issue), '') IS NOT NULL
    AND standard_cost >= 0
  );

CREATE UNIQUE INDEX IF NOT EXISTS uq_items_org_item_ref
  ON mdm.items(org_id, lower(item_ref));

CREATE UNIQUE INDEX IF NOT EXISTS uq_items_org_item_code
  ON mdm.items(org_id, lower(item_code));

CREATE INDEX IF NOT EXISTS ix_items_runtime_filters
  ON mdm.items(org_id, status, item_type, updated_at DESC);

ALTER TABLE mdm.warehouses
  ADD COLUMN IF NOT EXISTS warehouse_ref text,
  ADD COLUMN IF NOT EXISTS warehouse_type text,
  ADD COLUMN IF NOT EXISTS site_code text,
  ADD COLUMN IF NOT EXISTS allow_sale_issue boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS allow_prod_issue boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS allow_quarantine boolean NOT NULL DEFAULT false;

UPDATE mdm.warehouses
SET warehouse_ref = COALESCE(NULLIF(btrim(warehouse_ref), ''), id::text),
    warehouse_type = COALESCE(NULLIF(btrim(warehouse_type), ''), 'finished_good'),
    site_code = COALESCE(NULLIF(btrim(site_code), ''), 'HCM')
WHERE warehouse_ref IS NULL
   OR warehouse_type IS NULL
   OR site_code IS NULL;

ALTER TABLE mdm.warehouses
  DROP CONSTRAINT IF EXISTS ck_warehouses_runtime_required;

ALTER TABLE mdm.warehouses
  ADD CONSTRAINT ck_warehouses_runtime_required CHECK (
    nullif(btrim(warehouse_ref), '') IS NOT NULL
    AND nullif(btrim(warehouse_type), '') IS NOT NULL
    AND nullif(btrim(site_code), '') IS NOT NULL
    AND warehouse_type IN (
      'raw_material',
      'packaging',
      'semi_finished',
      'finished_good',
      'quarantine',
      'sample',
      'defect',
      'retail_store'
    )
  );

CREATE UNIQUE INDEX IF NOT EXISTS uq_warehouses_org_ref
  ON mdm.warehouses(org_id, lower(warehouse_ref));

CREATE INDEX IF NOT EXISTS ix_warehouses_runtime_filters
  ON mdm.warehouses(org_id, status, warehouse_type, updated_at DESC);

ALTER TABLE mdm.warehouse_bins
  ADD COLUMN IF NOT EXISTS location_ref text,
  ADD COLUMN IF NOT EXISTS zone_code text,
  ADD COLUMN IF NOT EXISTS allow_receive boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS allow_pick boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS allow_store boolean NOT NULL DEFAULT true,
  ADD COLUMN IF NOT EXISTS is_default boolean NOT NULL DEFAULT false;

UPDATE mdm.warehouse_bins
SET location_ref = COALESCE(NULLIF(btrim(location_ref), ''), id::text),
    zone_code = COALESCE(NULLIF(btrim(zone_code), ''), COALESCE(NULLIF(btrim(code), ''), 'DEFAULT'))
WHERE location_ref IS NULL
   OR zone_code IS NULL;

ALTER TABLE mdm.warehouse_bins
  DROP CONSTRAINT IF EXISTS ck_warehouse_bins_runtime_required;

ALTER TABLE mdm.warehouse_bins
  ADD CONSTRAINT ck_warehouse_bins_runtime_required CHECK (
    nullif(btrim(location_ref), '') IS NOT NULL
    AND nullif(btrim(zone_code), '') IS NOT NULL
  );

CREATE UNIQUE INDEX IF NOT EXISTS uq_warehouse_bins_org_location_ref
  ON mdm.warehouse_bins(org_id, lower(location_ref));

CREATE INDEX IF NOT EXISTS ix_warehouse_bins_runtime_filters
  ON mdm.warehouse_bins(org_id, warehouse_id, status, bin_type, updated_at DESC);

ALTER TABLE mdm.suppliers
  ADD COLUMN IF NOT EXISTS supplier_ref text,
  ADD COLUMN IF NOT EXISTS supplier_group text,
  ADD COLUMN IF NOT EXISTS contact_name text,
  ADD COLUMN IF NOT EXISTS payment_terms text,
  ADD COLUMN IF NOT EXISTS lead_time_days integer NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS moq numeric(18,6) NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS quality_score numeric(9,4) NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS delivery_score numeric(9,4) NOT NULL DEFAULT 0;

UPDATE mdm.suppliers
SET supplier_ref = COALESCE(NULLIF(btrim(supplier_ref), ''), id::text),
    supplier_group = COALESCE(
      NULLIF(btrim(supplier_group), ''),
      CASE supplier_type
        WHEN 'factory' THEN 'outsource'
        WHEN 'carrier_partner' THEN 'logistics'
        ELSE 'service'
      END
    )
WHERE supplier_ref IS NULL
   OR supplier_group IS NULL;

ALTER TABLE mdm.suppliers
  DROP CONSTRAINT IF EXISTS ck_suppliers_status;

ALTER TABLE mdm.suppliers
  ADD CONSTRAINT ck_suppliers_status CHECK (
    status IN ('draft', 'active', 'inactive', 'blacklisted', 'blocked')
  );

ALTER TABLE mdm.suppliers
  DROP CONSTRAINT IF EXISTS ck_suppliers_runtime_required;

ALTER TABLE mdm.suppliers
  ADD CONSTRAINT ck_suppliers_runtime_required CHECK (
    nullif(btrim(supplier_ref), '') IS NOT NULL
    AND nullif(btrim(supplier_group), '') IS NOT NULL
    AND supplier_group IN ('raw_material', 'packaging', 'service', 'logistics', 'outsource')
    AND lead_time_days >= 0
    AND moq >= 0
    AND quality_score >= 0
    AND delivery_score >= 0
  );

CREATE UNIQUE INDEX IF NOT EXISTS uq_suppliers_org_ref
  ON mdm.suppliers(org_id, lower(supplier_ref));

CREATE INDEX IF NOT EXISTS ix_suppliers_runtime_filters
  ON mdm.suppliers(org_id, status, supplier_group, updated_at DESC);

ALTER TABLE mdm.customers
  ADD COLUMN IF NOT EXISTS customer_ref text,
  ADD COLUMN IF NOT EXISTS customer_type text,
  ADD COLUMN IF NOT EXISTS channel_code text,
  ADD COLUMN IF NOT EXISTS price_list_code text,
  ADD COLUMN IF NOT EXISTS discount_group text,
  ADD COLUMN IF NOT EXISTS credit_limit numeric(18,2) NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS payment_terms text,
  ADD COLUMN IF NOT EXISTS contact_name text,
  ADD COLUMN IF NOT EXISTS tax_code text;

UPDATE mdm.customers
SET customer_ref = COALESCE(NULLIF(btrim(customer_ref), ''), id::text),
    customer_type = COALESCE(NULLIF(btrim(customer_type), ''), 'distributor')
WHERE customer_ref IS NULL
   OR customer_type IS NULL;

ALTER TABLE mdm.customers
  DROP CONSTRAINT IF EXISTS ck_customers_status;

ALTER TABLE mdm.customers
  ADD CONSTRAINT ck_customers_status CHECK (
    status IN ('draft', 'active', 'inactive', 'blocked')
  );

ALTER TABLE mdm.customers
  DROP CONSTRAINT IF EXISTS ck_customers_runtime_required;

ALTER TABLE mdm.customers
  ADD CONSTRAINT ck_customers_runtime_required CHECK (
    nullif(btrim(customer_ref), '') IS NOT NULL
    AND nullif(btrim(customer_type), '') IS NOT NULL
    AND customer_type IN ('distributor', 'dealer', 'retail_customer', 'marketplace', 'internal_store')
    AND credit_limit >= 0
  );

CREATE UNIQUE INDEX IF NOT EXISTS uq_customers_org_ref
  ON mdm.customers(org_id, lower(customer_ref));

CREATE INDEX IF NOT EXISTS ix_customers_runtime_filters
  ON mdm.customers(org_id, status, customer_type, updated_at DESC);

ALTER TABLE mdm.uom_conversions
  ADD COLUMN IF NOT EXISTS conversion_ref text,
  ADD COLUMN IF NOT EXISTS item_ref text;

UPDATE mdm.uom_conversions
SET conversion_ref = COALESCE(NULLIF(btrim(conversion_ref), ''), id::text),
    item_ref = COALESCE(NULLIF(btrim(item_ref), ''), item_id::text)
WHERE conversion_ref IS NULL;

ALTER TABLE mdm.uom_conversions
  DROP CONSTRAINT IF EXISTS ck_uom_conversions_runtime_required;

ALTER TABLE mdm.uom_conversions
  ADD CONSTRAINT ck_uom_conversions_runtime_required CHECK (
    nullif(btrim(conversion_ref), '') IS NOT NULL
    AND factor > 0
  );

CREATE UNIQUE INDEX IF NOT EXISTS uq_uom_conversions_ref
  ON mdm.uom_conversions(lower(conversion_ref));

CREATE INDEX IF NOT EXISTS ix_uom_conversions_runtime_item_ref
  ON mdm.uom_conversions(lower(item_ref))
  WHERE nullif(btrim(item_ref), '') IS NOT NULL;

COMMIT;
