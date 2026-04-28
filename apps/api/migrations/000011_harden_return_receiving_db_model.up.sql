BEGIN;

ALTER TABLE returns.return_orders
  ADD COLUMN IF NOT EXISTS customer_id uuid REFERENCES mdm.customers(id),
  ADD COLUMN IF NOT EXISTS carrier_id uuid REFERENCES mdm.carriers(id),
  ADD COLUMN IF NOT EXISTS source text NOT NULL DEFAULT 'UNKNOWN',
  ADD COLUMN IF NOT EXISTS tracking_no text,
  ADD COLUMN IF NOT EXISTS return_code text,
  ADD COLUMN IF NOT EXISTS return_reason_code text,
  ADD COLUMN IF NOT EXISTS package_condition text,
  ADD COLUMN IF NOT EXISTS initial_disposition text NOT NULL DEFAULT 'needs_inspection',
  ADD COLUMN IF NOT EXISTS unknown_case boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS investigation_note text;

ALTER TABLE returns.return_orders
  DROP CONSTRAINT IF EXISTS ck_return_orders_source;

ALTER TABLE returns.return_orders
  ADD CONSTRAINT ck_return_orders_source CHECK (
    source IN ('SHIPPER', 'CARRIER', 'CUSTOMER', 'MARKETPLACE', 'UNKNOWN')
  );

ALTER TABLE returns.return_orders
  DROP CONSTRAINT IF EXISTS ck_return_orders_initial_disposition;

ALTER TABLE returns.return_orders
  ADD CONSTRAINT ck_return_orders_initial_disposition CHECK (
    initial_disposition IN ('reusable', 'not_reusable', 'needs_inspection')
  );

ALTER TABLE returns.return_orders
  DROP CONSTRAINT IF EXISTS ck_return_orders_receiving_identity;

ALTER TABLE returns.return_orders
  ADD CONSTRAINT ck_return_orders_receiving_identity CHECK (
    sales_order_id IS NOT NULL
    OR shipment_id IS NOT NULL
    OR nullif(btrim(coalesce(tracking_no, '')), '') IS NOT NULL
    OR nullif(btrim(coalesce(return_code, '')), '') IS NOT NULL
    OR unknown_case
  );

CREATE INDEX IF NOT EXISTS ix_return_orders_tracking_no
  ON returns.return_orders(org_id, tracking_no)
  WHERE tracking_no IS NOT NULL;

CREATE INDEX IF NOT EXISTS ix_return_orders_customer_created
  ON returns.return_orders(org_id, customer_id, created_at DESC)
  WHERE customer_id IS NOT NULL;

ALTER TABLE returns.return_order_lines
  ADD COLUMN IF NOT EXISTS source_order_line_id uuid REFERENCES sales.sales_order_lines(id),
  ADD COLUMN IF NOT EXISTS return_reason_code text,
  ADD COLUMN IF NOT EXISTS condition_code text NOT NULL DEFAULT 'unknown';

ALTER TABLE returns.return_order_lines
  ALTER COLUMN returned_qty TYPE numeric(18,6) USING returned_qty::numeric(18,6);

ALTER TABLE returns.return_order_lines
  DROP CONSTRAINT IF EXISTS ck_return_order_lines_condition_code;

ALTER TABLE returns.return_order_lines
  ADD CONSTRAINT ck_return_order_lines_condition_code CHECK (
    condition_code IN ('unknown', 'sealed_good', 'opened_good', 'damaged', 'expired', 'suspected_quality_issue')
  );

CREATE INDEX IF NOT EXISTS ix_return_order_lines_source_order_line
  ON returns.return_order_lines(org_id, source_order_line_id)
  WHERE source_order_line_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS ix_return_order_lines_item_batch
  ON returns.return_order_lines(org_id, item_id, batch_id);

COMMIT;
