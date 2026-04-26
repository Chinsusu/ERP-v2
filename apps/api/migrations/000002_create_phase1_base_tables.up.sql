BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE SCHEMA IF NOT EXISTS core;
CREATE SCHEMA IF NOT EXISTS mdm;
CREATE SCHEMA IF NOT EXISTS purchase;
CREATE SCHEMA IF NOT EXISTS inventory;
CREATE SCHEMA IF NOT EXISTS qc;
CREATE SCHEMA IF NOT EXISTS sales;
CREATE SCHEMA IF NOT EXISTS shipping;
CREATE SCHEMA IF NOT EXISTS returns;
CREATE SCHEMA IF NOT EXISTS subcontract;
CREATE SCHEMA IF NOT EXISTS file;
CREATE SCHEMA IF NOT EXISTS integration;
CREATE SCHEMA IF NOT EXISTS audit;

ALTER TABLE IF EXISTS audit.audit_logs RENAME TO audit_logs_legacy_000001;
ALTER TABLE IF EXISTS inventory.stock_ledger RENAME TO stock_ledger_legacy_000001;

CREATE TABLE core.organizations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  code text NOT NULL,
  name text NOT NULL,
  status text NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_organizations_code UNIQUE (code),
  CONSTRAINT ck_organizations_status CHECK (status IN ('active', 'inactive'))
);

CREATE TABLE core.users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  email text NOT NULL,
  username text NOT NULL,
  display_name text NOT NULL,
  password_hash text,
  status text NOT NULL DEFAULT 'invited',
  last_login_at timestamptz,
  password_updated_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT ck_users_status CHECK (status IN ('invited', 'active', 'locked', 'disabled'))
);

CREATE UNIQUE INDEX uq_users_org_email_lower ON core.users(org_id, lower(email));
CREATE UNIQUE INDEX uq_users_org_username_lower ON core.users(org_id, lower(username));
CREATE INDEX ix_users_org_status ON core.users(org_id, status);

CREATE TABLE core.roles (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  code text NOT NULL,
  name text NOT NULL,
  description text,
  status text NOT NULL DEFAULT 'active',
  is_system boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_roles_org_code UNIQUE (org_id, code),
  CONSTRAINT ck_roles_status CHECK (status IN ('active', 'inactive'))
);

CREATE TABLE core.permissions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  code text NOT NULL,
  module text NOT NULL,
  resource text NOT NULL,
  action text NOT NULL,
  description text,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uq_permissions_code UNIQUE (code)
);

CREATE TABLE core.role_permissions (
  role_id uuid NOT NULL REFERENCES core.roles(id) ON DELETE CASCADE,
  permission_id uuid NOT NULL REFERENCES core.permissions(id) ON DELETE CASCADE,
  granted_at timestamptz NOT NULL DEFAULT now(),
  granted_by uuid REFERENCES core.users(id),
  PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE core.user_roles (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES core.users(id) ON DELETE CASCADE,
  role_id uuid NOT NULL REFERENCES core.roles(id) ON DELETE CASCADE,
  scope_type text NOT NULL DEFAULT 'company',
  scope_id uuid,
  valid_from timestamptz NOT NULL DEFAULT now(),
  valid_until timestamptz,
  assigned_at timestamptz NOT NULL DEFAULT now(),
  assigned_by uuid REFERENCES core.users(id),
  CONSTRAINT ck_user_roles_scope_type CHECK (
    scope_type IN ('company', 'warehouse', 'department', 'channel', 'own', 'assigned')
  )
);

CREATE UNIQUE INDEX uq_user_roles_assignment
  ON core.user_roles(user_id, role_id, scope_type, scope_id) NULLS NOT DISTINCT;

CREATE TABLE core.idempotency_keys (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  idempotency_key text NOT NULL,
  request_hash text NOT NULL,
  status text NOT NULL DEFAULT 'processing',
  response_status integer,
  response_body jsonb,
  locked_until timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uq_idempotency_keys_org_key UNIQUE (org_id, idempotency_key),
  CONSTRAINT ck_idempotency_keys_status CHECK (status IN ('processing', 'completed', 'failed'))
);

CREATE TABLE core.document_sequences (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  doc_type text NOT NULL,
  prefix text NOT NULL,
  current_value bigint NOT NULL DEFAULT 0,
  reset_policy text NOT NULL DEFAULT 'yearly',
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uq_document_sequences_org_type_prefix UNIQUE (org_id, doc_type, prefix),
  CONSTRAINT ck_document_sequences_reset CHECK (reset_policy IN ('none', 'daily', 'monthly', 'yearly'))
);

CREATE TABLE mdm.units (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  code text NOT NULL,
  name text NOT NULL,
  precision_scale integer NOT NULL DEFAULT 4,
  status text NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_units_org_code UNIQUE (org_id, code),
  CONSTRAINT ck_units_status CHECK (status IN ('active', 'inactive')),
  CONSTRAINT ck_units_precision CHECK (precision_scale BETWEEN 0 AND 6)
);

CREATE TABLE mdm.suppliers (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  code text NOT NULL,
  name text NOT NULL,
  supplier_type text NOT NULL DEFAULT 'supplier',
  tax_code text,
  email text,
  phone text,
  address text,
  status text NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_suppliers_org_code UNIQUE (org_id, code),
  CONSTRAINT ck_suppliers_status CHECK (status IN ('active', 'inactive', 'blocked')),
  CONSTRAINT ck_suppliers_type CHECK (supplier_type IN ('supplier', 'factory', 'carrier_partner'))
);

CREATE TABLE mdm.customers (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  code text NOT NULL,
  name text NOT NULL,
  email text,
  phone text,
  address text,
  status text NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_customers_org_code UNIQUE (org_id, code),
  CONSTRAINT ck_customers_status CHECK (status IN ('active', 'inactive', 'blocked'))
);

CREATE TABLE mdm.items (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  sku text NOT NULL,
  name text NOT NULL,
  item_type text NOT NULL DEFAULT 'finished_good',
  base_unit_id uuid NOT NULL REFERENCES mdm.units(id),
  barcode text,
  shelf_life_days integer,
  requires_batch boolean NOT NULL DEFAULT true,
  requires_expiry boolean NOT NULL DEFAULT true,
  status text NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_items_org_sku UNIQUE (org_id, sku),
  CONSTRAINT ck_items_type CHECK (
    item_type IN ('raw_material', 'packaging', 'semi_finished', 'finished_good', 'service')
  ),
  CONSTRAINT ck_items_status CHECK (status IN ('active', 'inactive', 'blocked')),
  CONSTRAINT ck_items_shelf_life CHECK (shelf_life_days IS NULL OR shelf_life_days >= 0)
);

CREATE INDEX ix_items_sku_lower ON mdm.items(lower(sku));
CREATE INDEX ix_items_org_status ON mdm.items(org_id, status);

CREATE TABLE mdm.warehouses (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  code text NOT NULL,
  name text NOT NULL,
  address text,
  status text NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_warehouses_org_code UNIQUE (org_id, code),
  CONSTRAINT ck_warehouses_status CHECK (status IN ('active', 'inactive'))
);

CREATE TABLE mdm.warehouse_zones (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  warehouse_id uuid NOT NULL REFERENCES mdm.warehouses(id),
  code text NOT NULL,
  name text NOT NULL,
  zone_type text NOT NULL DEFAULT 'storage',
  status text NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_warehouse_zones_warehouse_code UNIQUE (warehouse_id, code),
  CONSTRAINT ck_warehouse_zones_type CHECK (
    zone_type IN ('receiving', 'qc_hold', 'storage', 'pick', 'pack', 'handover', 'return', 'lab', 'scrap')
  ),
  CONSTRAINT ck_warehouse_zones_status CHECK (status IN ('active', 'inactive'))
);

CREATE TABLE mdm.warehouse_bins (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  warehouse_id uuid NOT NULL REFERENCES mdm.warehouses(id),
  zone_id uuid REFERENCES mdm.warehouse_zones(id),
  code text NOT NULL,
  name text,
  bin_type text NOT NULL DEFAULT 'storage',
  status text NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_warehouse_bins_warehouse_code UNIQUE (warehouse_id, code),
  CONSTRAINT ck_warehouse_bins_type CHECK (
    bin_type IN ('receiving', 'qc_hold', 'storage', 'pick', 'pack', 'handover', 'return', 'lab', 'scrap')
  ),
  CONSTRAINT ck_warehouse_bins_status CHECK (status IN ('active', 'inactive', 'blocked'))
);

CREATE TABLE mdm.carriers (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  code text NOT NULL,
  name text NOT NULL,
  contact_name text,
  phone text,
  status text NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_carriers_org_code UNIQUE (org_id, code),
  CONSTRAINT ck_carriers_status CHECK (status IN ('active', 'inactive', 'blocked'))
);

CREATE TABLE purchase.purchase_orders (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  po_no text NOT NULL,
  supplier_id uuid NOT NULL REFERENCES mdm.suppliers(id),
  order_date date NOT NULL,
  expected_receipt_date date,
  status text NOT NULL DEFAULT 'draft',
  currency text NOT NULL DEFAULT 'VND',
  total_amount numeric(18, 2) NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  submitted_at timestamptz,
  submitted_by uuid REFERENCES core.users(id),
  approved_at timestamptz,
  approved_by uuid REFERENCES core.users(id),
  cancelled_at timestamptz,
  cancelled_by uuid REFERENCES core.users(id),
  cancel_reason text,
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_purchase_orders_org_no UNIQUE (org_id, po_no),
  CONSTRAINT ck_purchase_orders_status CHECK (
    status IN ('draft', 'submitted', 'approved', 'partially_received', 'received', 'closed', 'cancelled')
  ),
  CONSTRAINT ck_purchase_orders_total CHECK (total_amount >= 0)
);

CREATE TABLE purchase.purchase_order_lines (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  purchase_order_id uuid NOT NULL REFERENCES purchase.purchase_orders(id) ON DELETE RESTRICT,
  line_no integer NOT NULL,
  item_id uuid NOT NULL REFERENCES mdm.items(id),
  unit_id uuid NOT NULL REFERENCES mdm.units(id),
  ordered_qty numeric(18, 4) NOT NULL,
  received_qty numeric(18, 4) NOT NULL DEFAULT 0,
  unit_price numeric(18, 2) NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  CONSTRAINT uq_purchase_order_lines_order_line UNIQUE (purchase_order_id, line_no),
  CONSTRAINT ck_purchase_order_lines_qty CHECK (ordered_qty > 0 AND received_qty >= 0),
  CONSTRAINT ck_purchase_order_lines_price CHECK (unit_price >= 0)
);

CREATE TABLE inventory.batches (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  item_id uuid NOT NULL REFERENCES mdm.items(id),
  batch_no text NOT NULL,
  supplier_id uuid REFERENCES mdm.suppliers(id),
  mfg_date date,
  expiry_date date,
  qc_status text NOT NULL DEFAULT 'hold',
  status text NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_batches_item_batch UNIQUE (item_id, batch_no),
  CONSTRAINT ck_batches_qc_status CHECK (
    qc_status IN ('hold', 'pass', 'fail', 'quarantine', 'retest_required')
  ),
  CONSTRAINT ck_batches_status CHECK (status IN ('active', 'inactive', 'blocked')),
  CONSTRAINT ck_batches_expiry_after_mfg CHECK (
    mfg_date IS NULL OR expiry_date IS NULL OR expiry_date >= mfg_date
  )
);

CREATE INDEX ix_batches_org_qc_expiry ON inventory.batches(org_id, qc_status, expiry_date);

CREATE TABLE inventory.goods_receipts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  grn_no text NOT NULL,
  warehouse_id uuid NOT NULL REFERENCES mdm.warehouses(id),
  supplier_id uuid REFERENCES mdm.suppliers(id),
  purchase_order_id uuid REFERENCES purchase.purchase_orders(id),
  receipt_date date NOT NULL,
  status text NOT NULL DEFAULT 'draft',
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  submitted_at timestamptz,
  submitted_by uuid REFERENCES core.users(id),
  approved_at timestamptz,
  approved_by uuid REFERENCES core.users(id),
  cancelled_at timestamptz,
  cancelled_by uuid REFERENCES core.users(id),
  cancel_reason text,
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_goods_receipts_org_no UNIQUE (org_id, grn_no),
  CONSTRAINT ck_goods_receipts_status CHECK (
    status IN ('draft', 'submitted', 'received', 'qc_pending', 'closed', 'cancelled')
  )
);

CREATE TABLE inventory.goods_receipt_lines (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  goods_receipt_id uuid NOT NULL REFERENCES inventory.goods_receipts(id) ON DELETE RESTRICT,
  line_no integer NOT NULL,
  item_id uuid NOT NULL REFERENCES mdm.items(id),
  batch_id uuid REFERENCES inventory.batches(id),
  unit_id uuid NOT NULL REFERENCES mdm.units(id),
  received_qty numeric(18, 4) NOT NULL,
  accepted_qty numeric(18, 4) NOT NULL DEFAULT 0,
  rejected_qty numeric(18, 4) NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  CONSTRAINT uq_goods_receipt_lines_receipt_line UNIQUE (goods_receipt_id, line_no),
  CONSTRAINT ck_goods_receipt_lines_qty CHECK (
    received_qty > 0 AND accepted_qty >= 0 AND rejected_qty >= 0
  )
);

CREATE TABLE inventory.stock_ledger (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  movement_no text NOT NULL,
  movement_type text NOT NULL,
  movement_at timestamptz NOT NULL DEFAULT now(),
  direction text NOT NULL,
  item_id uuid NOT NULL REFERENCES mdm.items(id),
  batch_id uuid REFERENCES inventory.batches(id),
  warehouse_id uuid NOT NULL REFERENCES mdm.warehouses(id),
  bin_id uuid REFERENCES mdm.warehouse_bins(id),
  unit_id uuid NOT NULL REFERENCES mdm.units(id),
  quantity numeric(18, 4) NOT NULL,
  stock_status text NOT NULL DEFAULT 'available',
  source_doc_type text NOT NULL,
  source_doc_id uuid NOT NULL,
  source_doc_line_id uuid,
  reason text,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  CONSTRAINT uq_stock_ledger_org_movement_no UNIQUE (org_id, movement_no),
  CONSTRAINT ck_stock_ledger_direction CHECK (direction IN ('in', 'out', 'transfer', 'adjustment')),
  CONSTRAINT ck_stock_ledger_quantity CHECK (quantity > 0),
  CONSTRAINT ck_stock_ledger_stock_status CHECK (
    stock_status IN ('available', 'reserved', 'qc_hold', 'return_pending', 'damaged', 'subcontract_issued')
  )
);

CREATE INDEX ix_stock_ledger_item_batch_warehouse
  ON inventory.stock_ledger(item_id, batch_id, warehouse_id, movement_at DESC);
CREATE INDEX ix_stock_ledger_source_doc
  ON inventory.stock_ledger(source_doc_type, source_doc_id);

CREATE FUNCTION inventory.prevent_stock_ledger_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  RAISE EXCEPTION 'inventory.stock_ledger rows are immutable';
END;
$$;

CREATE TRIGGER trg_stock_ledger_immutable
  BEFORE UPDATE OR DELETE ON inventory.stock_ledger
  FOR EACH ROW
  EXECUTE FUNCTION inventory.prevent_stock_ledger_mutation();

CREATE TABLE inventory.stock_balances (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  item_id uuid NOT NULL REFERENCES mdm.items(id),
  batch_id uuid REFERENCES inventory.batches(id),
  warehouse_id uuid NOT NULL REFERENCES mdm.warehouses(id),
  bin_id uuid REFERENCES mdm.warehouse_bins(id),
  stock_status text NOT NULL DEFAULT 'available',
  qty_on_hand numeric(18, 4) NOT NULL DEFAULT 0,
  qty_reserved numeric(18, 4) NOT NULL DEFAULT 0,
  qty_available numeric(18, 4) NOT NULL DEFAULT 0,
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT ck_stock_balances_status CHECK (
    stock_status IN ('available', 'reserved', 'qc_hold', 'return_pending', 'damaged', 'subcontract_issued')
  ),
  CONSTRAINT ck_stock_balances_non_negative CHECK (
    qty_on_hand >= 0 AND qty_reserved >= 0 AND qty_available >= 0
  )
);

CREATE UNIQUE INDEX uq_stock_balances_key
  ON inventory.stock_balances(org_id, item_id, batch_id, warehouse_id, bin_id, stock_status)
  NULLS NOT DISTINCT;
CREATE INDEX ix_stock_balances_lookup
  ON inventory.stock_balances(org_id, warehouse_id, item_id, batch_id, stock_status);
CREATE INDEX ix_stock_balances_available_positive
  ON inventory.stock_balances(org_id, warehouse_id, item_id)
  WHERE qty_available > 0;

CREATE FUNCTION inventory.require_stock_balance_write_context()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  IF current_setting('erp.allow_stock_balance_write', true) IS DISTINCT FROM 'on' THEN
    RAISE EXCEPTION 'inventory.stock_balances can only be changed through the stock movement service';
  END IF;

  IF TG_OP = 'DELETE' THEN
    RETURN OLD;
  END IF;

  RETURN NEW;
END;
$$;

CREATE TRIGGER trg_stock_balances_write_guard
  BEFORE INSERT OR UPDATE OR DELETE ON inventory.stock_balances
  FOR EACH ROW
  EXECUTE FUNCTION inventory.require_stock_balance_write_context();

CREATE TABLE inventory.stock_reservations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  reservation_no text NOT NULL,
  item_id uuid NOT NULL REFERENCES mdm.items(id),
  batch_id uuid REFERENCES inventory.batches(id),
  warehouse_id uuid NOT NULL REFERENCES mdm.warehouses(id),
  quantity numeric(18, 4) NOT NULL,
  source_doc_type text NOT NULL,
  source_doc_id uuid NOT NULL,
  status text NOT NULL DEFAULT 'active',
  expires_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  released_at timestamptz,
  released_by uuid REFERENCES core.users(id),
  CONSTRAINT uq_stock_reservations_org_no UNIQUE (org_id, reservation_no),
  CONSTRAINT ck_stock_reservations_qty CHECK (quantity > 0),
  CONSTRAINT ck_stock_reservations_status CHECK (status IN ('active', 'released', 'consumed', 'expired', 'cancelled'))
);

CREATE INDEX ix_stock_reservations_active
  ON inventory.stock_reservations(source_doc_type, source_doc_id)
  WHERE status = 'active';

CREATE TABLE inventory.stock_counts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  count_no text NOT NULL,
  warehouse_id uuid NOT NULL REFERENCES mdm.warehouses(id),
  count_date date NOT NULL,
  status text NOT NULL DEFAULT 'draft',
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  approved_at timestamptz,
  approved_by uuid REFERENCES core.users(id),
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_stock_counts_org_no UNIQUE (org_id, count_no),
  CONSTRAINT ck_stock_counts_status CHECK (status IN ('draft', 'counting', 'review', 'approved', 'cancelled'))
);

CREATE TABLE inventory.warehouse_daily_closings (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  closing_no text NOT NULL,
  warehouse_id uuid NOT NULL REFERENCES mdm.warehouses(id),
  business_date date NOT NULL,
  shift_code text NOT NULL DEFAULT 'day',
  status text NOT NULL DEFAULT 'open',
  orders_processed_count integer NOT NULL DEFAULT 0,
  pending_task_count integer NOT NULL DEFAULT 0,
  variance_count integer NOT NULL DEFAULT 0,
  exception_note text,
  closed_at timestamptz,
  closed_by uuid REFERENCES core.users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_warehouse_daily_closings_org_no UNIQUE (org_id, closing_no),
  CONSTRAINT uq_warehouse_daily_closings_day_shift UNIQUE (warehouse_id, business_date, shift_code),
  CONSTRAINT ck_warehouse_daily_closings_status CHECK (status IN ('open', 'in_review', 'closed', 'reopened')),
  CONSTRAINT ck_warehouse_daily_closings_counts CHECK (
    orders_processed_count >= 0 AND pending_task_count >= 0 AND variance_count >= 0
  )
);

CREATE TABLE qc.inspections (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  inspection_no text NOT NULL,
  item_id uuid NOT NULL REFERENCES mdm.items(id),
  batch_id uuid REFERENCES inventory.batches(id),
  source_doc_type text NOT NULL,
  source_doc_id uuid NOT NULL,
  status text NOT NULL DEFAULT 'pending',
  result text,
  inspected_at timestamptz,
  inspected_by uuid REFERENCES core.users(id),
  approved_at timestamptz,
  approved_by uuid REFERENCES core.users(id),
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_inspections_org_no UNIQUE (org_id, inspection_no),
  CONSTRAINT ck_inspections_status CHECK (status IN ('pending', 'in_progress', 'approved', 'rejected', 'cancelled')),
  CONSTRAINT ck_inspections_result CHECK (result IS NULL OR result IN ('pass', 'fail', 'hold', 'retest_required'))
);

CREATE TABLE qc.batch_quality_statuses (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  batch_id uuid NOT NULL REFERENCES inventory.batches(id),
  from_status text,
  to_status text NOT NULL,
  changed_at timestamptz NOT NULL DEFAULT now(),
  changed_by uuid REFERENCES core.users(id),
  inspection_id uuid REFERENCES qc.inspections(id),
  reason text,
  CONSTRAINT ck_batch_quality_statuses_from CHECK (
    from_status IS NULL OR from_status IN ('hold', 'pass', 'fail', 'quarantine', 'retest_required')
  ),
  CONSTRAINT ck_batch_quality_statuses_to CHECK (
    to_status IN ('hold', 'pass', 'fail', 'quarantine', 'retest_required')
  )
);

CREATE TABLE sales.sales_orders (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  order_no text NOT NULL,
  customer_id uuid REFERENCES mdm.customers(id),
  order_date date NOT NULL,
  channel text NOT NULL DEFAULT 'internal',
  status text NOT NULL DEFAULT 'draft',
  currency text NOT NULL DEFAULT 'VND',
  total_amount numeric(18, 2) NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  submitted_at timestamptz,
  submitted_by uuid REFERENCES core.users(id),
  approved_at timestamptz,
  approved_by uuid REFERENCES core.users(id),
  cancelled_at timestamptz,
  cancelled_by uuid REFERENCES core.users(id),
  cancel_reason text,
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_sales_orders_org_no UNIQUE (org_id, order_no),
  CONSTRAINT ck_sales_orders_status CHECK (
    status IN ('draft', 'confirmed', 'reserved', 'picking', 'packed', 'handed_over', 'delivered', 'closed', 'cancelled', 'returned')
  ),
  CONSTRAINT ck_sales_orders_total CHECK (total_amount >= 0)
);

CREATE INDEX ix_sales_orders_status_order_date
  ON sales.sales_orders(org_id, status, order_date DESC);

CREATE TABLE sales.sales_order_lines (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  sales_order_id uuid NOT NULL REFERENCES sales.sales_orders(id) ON DELETE RESTRICT,
  line_no integer NOT NULL,
  item_id uuid NOT NULL REFERENCES mdm.items(id),
  unit_id uuid NOT NULL REFERENCES mdm.units(id),
  ordered_qty numeric(18, 4) NOT NULL,
  reserved_qty numeric(18, 4) NOT NULL DEFAULT 0,
  shipped_qty numeric(18, 4) NOT NULL DEFAULT 0,
  unit_price numeric(18, 2) NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  CONSTRAINT uq_sales_order_lines_order_line UNIQUE (sales_order_id, line_no),
  CONSTRAINT ck_sales_order_lines_qty CHECK (
    ordered_qty > 0 AND reserved_qty >= 0 AND shipped_qty >= 0
  ),
  CONSTRAINT ck_sales_order_lines_price CHECK (unit_price >= 0)
);

CREATE INDEX ix_sales_order_lines_sales_order_id
  ON sales.sales_order_lines(sales_order_id);

CREATE TABLE shipping.shipments (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  shipment_no text NOT NULL,
  sales_order_id uuid REFERENCES sales.sales_orders(id),
  warehouse_id uuid NOT NULL REFERENCES mdm.warehouses(id),
  carrier_id uuid REFERENCES mdm.carriers(id),
  tracking_no text,
  status text NOT NULL DEFAULT 'draft',
  packed_at timestamptz,
  packed_by uuid REFERENCES core.users(id),
  handed_over_at timestamptz,
  handed_over_by uuid REFERENCES core.users(id),
  delivered_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_shipments_org_no UNIQUE (org_id, shipment_no),
  CONSTRAINT uq_shipments_org_tracking UNIQUE (org_id, tracking_no),
  CONSTRAINT ck_shipments_status CHECK (
    status IN ('draft', 'picking', 'packed', 'ready_for_handover', 'handed_over', 'delivered', 'exception', 'cancelled')
  )
);

CREATE INDEX ix_shipments_status_carrier
  ON shipping.shipments(org_id, status, carrier_id, created_at DESC);

CREATE TABLE shipping.carrier_manifests (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  manifest_no text NOT NULL,
  warehouse_id uuid NOT NULL REFERENCES mdm.warehouses(id),
  carrier_id uuid NOT NULL REFERENCES mdm.carriers(id),
  handover_date date NOT NULL,
  status text NOT NULL DEFAULT 'draft',
  expected_count integer NOT NULL DEFAULT 0,
  scanned_count integer NOT NULL DEFAULT 0,
  missing_count integer NOT NULL DEFAULT 0,
  exception_note text,
  completed_at timestamptz,
  completed_by uuid REFERENCES core.users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_carrier_manifests_org_no UNIQUE (org_id, manifest_no),
  CONSTRAINT ck_carrier_manifests_status CHECK (
    status IN ('draft', 'ready', 'scanning', 'completed', 'exception', 'cancelled')
  ),
  CONSTRAINT ck_carrier_manifests_counts CHECK (
    expected_count >= 0 AND scanned_count >= 0 AND missing_count >= 0
  )
);

CREATE INDEX ix_carrier_manifests_carrier_date
  ON shipping.carrier_manifests(carrier_id, handover_date, status);

CREATE TABLE shipping.carrier_manifest_lines (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  carrier_manifest_id uuid NOT NULL REFERENCES shipping.carrier_manifests(id) ON DELETE RESTRICT,
  shipment_id uuid NOT NULL REFERENCES shipping.shipments(id),
  sales_order_id uuid REFERENCES sales.sales_orders(id),
  line_no integer NOT NULL,
  tracking_no text,
  staging_zone text,
  status text NOT NULL DEFAULT 'expected',
  scanned_at timestamptz,
  scanned_by uuid REFERENCES core.users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  CONSTRAINT uq_carrier_manifest_lines_manifest_line UNIQUE (carrier_manifest_id, line_no),
  CONSTRAINT uq_carrier_manifest_lines_manifest_shipment UNIQUE (carrier_manifest_id, shipment_id),
  CONSTRAINT ck_carrier_manifest_lines_status CHECK (
    status IN ('expected', 'scanned', 'missing', 'exception', 'removed')
  )
);

CREATE TABLE shipping.scan_events (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  carrier_manifest_id uuid REFERENCES shipping.carrier_manifests(id),
  shipment_id uuid REFERENCES shipping.shipments(id),
  barcode text NOT NULL,
  scan_context text NOT NULL,
  scan_result text NOT NULL,
  error_code text,
  scan_station text,
  scanned_at timestamptz NOT NULL DEFAULT now(),
  scanned_by uuid REFERENCES core.users(id),
  idempotency_key text,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  CONSTRAINT ck_scan_events_context CHECK (scan_context IN ('handover', 'packing', 'return', 'stock_count')),
  CONSTRAINT ck_scan_events_result CHECK (
    scan_result IN ('matched', 'wrong_manifest', 'duplicate', 'not_found', 'invalid_state', 'error')
  )
);

CREATE INDEX ix_scan_events_context_barcode
  ON shipping.scan_events(scan_context, barcode, scanned_at DESC);

CREATE TABLE returns.return_orders (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  return_no text NOT NULL,
  sales_order_id uuid REFERENCES sales.sales_orders(id),
  shipment_id uuid REFERENCES shipping.shipments(id),
  warehouse_id uuid NOT NULL REFERENCES mdm.warehouses(id),
  status text NOT NULL DEFAULT 'received',
  return_reason text,
  received_at timestamptz NOT NULL DEFAULT now(),
  received_by uuid REFERENCES core.users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_return_orders_org_no UNIQUE (org_id, return_no),
  CONSTRAINT ck_return_orders_status CHECK (
    status IN ('received', 'pending_inspection', 'inspected', 'disposed', 'closed', 'cancelled')
  )
);

CREATE INDEX ix_return_orders_status_created
  ON returns.return_orders(org_id, status, created_at DESC);

CREATE TABLE returns.return_order_lines (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  return_order_id uuid NOT NULL REFERENCES returns.return_orders(id) ON DELETE RESTRICT,
  line_no integer NOT NULL,
  item_id uuid NOT NULL REFERENCES mdm.items(id),
  batch_id uuid REFERENCES inventory.batches(id),
  unit_id uuid NOT NULL REFERENCES mdm.units(id),
  returned_qty numeric(18, 4) NOT NULL,
  condition_note text,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  CONSTRAINT uq_return_order_lines_order_line UNIQUE (return_order_id, line_no),
  CONSTRAINT ck_return_order_lines_qty CHECK (returned_qty > 0)
);

CREATE TABLE returns.return_dispositions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  return_order_line_id uuid NOT NULL REFERENCES returns.return_order_lines(id),
  disposition text NOT NULL,
  target_warehouse_id uuid REFERENCES mdm.warehouses(id),
  target_bin_id uuid REFERENCES mdm.warehouse_bins(id),
  inspected_at timestamptz NOT NULL DEFAULT now(),
  inspected_by uuid REFERENCES core.users(id),
  approved_at timestamptz,
  approved_by uuid REFERENCES core.users(id),
  reason text,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT ck_return_dispositions_disposition CHECK (
    disposition IN ('reusable', 'qc_required', 'damaged', 'missing_item', 'wrong_item', 'scrap', 'unknown')
  )
);

CREATE TABLE subcontract.subcontract_orders (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  subcontract_no text NOT NULL,
  factory_supplier_id uuid NOT NULL REFERENCES mdm.suppliers(id),
  product_item_id uuid REFERENCES mdm.items(id),
  order_qty numeric(18, 4) NOT NULL,
  unit_id uuid REFERENCES mdm.units(id),
  spec text,
  expected_delivery_date date,
  deposit_amount numeric(18, 2) NOT NULL DEFAULT 0,
  final_payment_status text NOT NULL DEFAULT 'not_due',
  status text NOT NULL DEFAULT 'draft',
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  submitted_at timestamptz,
  submitted_by uuid REFERENCES core.users(id),
  approved_at timestamptz,
  approved_by uuid REFERENCES core.users(id),
  cancelled_at timestamptz,
  cancelled_by uuid REFERENCES core.users(id),
  cancel_reason text,
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_subcontract_orders_org_no UNIQUE (org_id, subcontract_no),
  CONSTRAINT ck_subcontract_orders_qty CHECK (order_qty > 0),
  CONSTRAINT ck_subcontract_orders_deposit CHECK (deposit_amount >= 0),
  CONSTRAINT ck_subcontract_orders_payment_status CHECK (
    final_payment_status IN ('not_due', 'pending', 'paid', 'held')
  ),
  CONSTRAINT ck_subcontract_orders_status CHECK (
    status IN (
      'draft',
      'confirmed',
      'deposit_paid',
      'materials_prepared',
      'materials_sent',
      'sample_pending',
      'sample_approved',
      'sample_rejected',
      'mass_production',
      'inbound_pending',
      'qc_checking',
      'accepted',
      'claimed',
      'closed',
      'cancelled'
    )
  )
);

CREATE INDEX ix_subcontract_orders_factory_status
  ON subcontract.subcontract_orders(factory_supplier_id, status);

CREATE TABLE subcontract.material_issues (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  issue_no text NOT NULL,
  subcontract_order_id uuid NOT NULL REFERENCES subcontract.subcontract_orders(id),
  source_warehouse_id uuid NOT NULL REFERENCES mdm.warehouses(id),
  status text NOT NULL DEFAULT 'draft',
  issued_at timestamptz,
  issued_by uuid REFERENCES core.users(id),
  handover_note text,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_material_issues_org_no UNIQUE (org_id, issue_no),
  CONSTRAINT ck_material_issues_status CHECK (status IN ('draft', 'issued', 'received_by_factory', 'cancelled'))
);

CREATE TABLE subcontract.material_issue_lines (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  material_issue_id uuid NOT NULL REFERENCES subcontract.material_issues(id) ON DELETE RESTRICT,
  line_no integer NOT NULL,
  item_id uuid NOT NULL REFERENCES mdm.items(id),
  batch_id uuid REFERENCES inventory.batches(id),
  unit_id uuid NOT NULL REFERENCES mdm.units(id),
  issued_qty numeric(18, 4) NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  CONSTRAINT uq_material_issue_lines_issue_line UNIQUE (material_issue_id, line_no),
  CONSTRAINT ck_material_issue_lines_qty CHECK (issued_qty > 0)
);

CREATE TABLE subcontract.sample_approvals (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  subcontract_order_id uuid NOT NULL REFERENCES subcontract.subcontract_orders(id),
  sample_version text NOT NULL,
  status text NOT NULL DEFAULT 'pending',
  approved_at timestamptz,
  approved_by uuid REFERENCES core.users(id),
  rejected_reason text,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  CONSTRAINT uq_sample_approvals_order_version UNIQUE (subcontract_order_id, sample_version),
  CONSTRAINT ck_sample_approvals_status CHECK (status IN ('pending', 'approved', 'rejected', 'rework'))
);

CREATE TABLE subcontract.factory_claims (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  claim_no text NOT NULL,
  subcontract_order_id uuid NOT NULL REFERENCES subcontract.subcontract_orders(id),
  issue_type text NOT NULL,
  affected_qty numeric(18, 4) NOT NULL DEFAULT 0,
  requested_action text,
  response_deadline_at timestamptz NOT NULL,
  status text NOT NULL DEFAULT 'open',
  resolved_at timestamptz,
  resolved_by uuid REFERENCES core.users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_factory_claims_org_no UNIQUE (org_id, claim_no),
  CONSTRAINT ck_factory_claims_qty CHECK (affected_qty >= 0),
  CONSTRAINT ck_factory_claims_status CHECK (
    status IN ('open', 'sent', 'factory_responded', 'resolved', 'rejected', 'overdue', 'cancelled')
  )
);

CREATE INDEX ix_factory_claims_deadline_status
  ON subcontract.factory_claims(response_deadline_at, status);

CREATE TABLE file.attachments (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  entity_type text NOT NULL,
  entity_id uuid NOT NULL,
  file_name text NOT NULL,
  file_ext text,
  mime_type text,
  file_size_bytes bigint,
  storage_bucket text NOT NULL,
  storage_key text NOT NULL,
  checksum text,
  uploaded_at timestamptz NOT NULL DEFAULT now(),
  uploaded_by uuid REFERENCES core.users(id),
  status text NOT NULL DEFAULT 'active',
  CONSTRAINT ck_attachments_status CHECK (status IN ('active', 'deleted', 'quarantined')),
  CONSTRAINT ck_attachments_size CHECK (file_size_bytes IS NULL OR file_size_bytes >= 0)
);

CREATE INDEX ix_attachments_entity ON file.attachments(entity_type, entity_id);

CREATE TABLE integration.outbox_events (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  aggregate_type text NOT NULL,
  aggregate_id uuid NOT NULL,
  event_type text NOT NULL,
  payload jsonb NOT NULL,
  status text NOT NULL DEFAULT 'pending',
  idempotency_key text,
  available_at timestamptz NOT NULL DEFAULT now(),
  published_at timestamptz,
  retry_count integer NOT NULL DEFAULT 0,
  last_error text,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uq_outbox_events_idempotency UNIQUE (org_id, idempotency_key),
  CONSTRAINT ck_outbox_events_status CHECK (status IN ('pending', 'processing', 'published', 'failed', 'dead')),
  CONSTRAINT ck_outbox_events_retry CHECK (retry_count >= 0)
);

CREATE INDEX ix_outbox_events_pending
  ON integration.outbox_events(available_at)
  WHERE status = 'pending';

CREATE TABLE audit.audit_logs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  actor_id uuid REFERENCES core.users(id),
  action text NOT NULL,
  entity_type text NOT NULL,
  entity_id uuid,
  request_id text,
  before_data jsonb,
  after_data jsonb,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX ix_audit_logs_entity ON audit.audit_logs(entity_type, entity_id, created_at DESC);
CREATE INDEX ix_audit_logs_actor ON audit.audit_logs(actor_id, created_at DESC);

COMMIT;
