BEGIN;

ALTER TABLE mdm.carriers
  ADD COLUMN IF NOT EXISTS handover_zone text NOT NULL DEFAULT 'handover',
  ADD COLUMN IF NOT EXISTS sla_profile text NOT NULL DEFAULT 'standard';

ALTER TABLE mdm.carriers
  DROP CONSTRAINT IF EXISTS ck_carriers_status,
  ADD CONSTRAINT ck_carriers_status CHECK (status IN ('active', 'inactive'));

ALTER TABLE shipping.carrier_manifests
  ADD COLUMN IF NOT EXISTS handover_batch text NOT NULL DEFAULT 'day',
  ADD COLUMN IF NOT EXISTS handover_zone text NOT NULL DEFAULT 'handover',
  ADD COLUMN IF NOT EXISTS handed_over_at timestamptz,
  ADD COLUMN IF NOT EXISTS handed_over_by uuid REFERENCES core.users(id),
  ADD COLUMN IF NOT EXISTS carrier_receiver_name text,
  ADD COLUMN IF NOT EXISTS carrier_signature_ref text,
  ADD COLUMN IF NOT EXISTS note text;

ALTER TABLE shipping.carrier_manifests
  DROP CONSTRAINT IF EXISTS ck_carrier_manifests_status,
  DROP CONSTRAINT IF EXISTS ck_carrier_manifests_counts,
  ADD CONSTRAINT ck_carrier_manifests_status CHECK (
    status IN (
      'draft',
      'ready',
      'scanning',
      'completed',
      'handed_over',
      'exception',
      'cancelled'
    )
  ),
  ADD CONSTRAINT ck_carrier_manifests_counts CHECK (
    expected_count >= 0
    AND scanned_count >= 0
    AND missing_count >= 0
  );

CREATE INDEX IF NOT EXISTS ix_carrier_manifests_status_date
  ON shipping.carrier_manifests(org_id, status, handover_date DESC);

CREATE TABLE IF NOT EXISTS shipping.carrier_manifest_orders (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  carrier_manifest_id uuid NOT NULL REFERENCES shipping.carrier_manifests(id) ON DELETE RESTRICT,
  line_no integer NOT NULL,
  shipment_id uuid REFERENCES shipping.shipments(id),
  sales_order_id uuid REFERENCES sales.sales_orders(id),
  order_no text NOT NULL,
  tracking_no text,
  package_code text,
  staging_zone text,
  scan_status text NOT NULL DEFAULT 'pending',
  scanned_at timestamptz,
  scanned_by uuid REFERENCES core.users(id),
  issue_code text,
  note text,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  CONSTRAINT uq_carrier_manifest_orders_manifest_line UNIQUE (carrier_manifest_id, line_no),
  CONSTRAINT ck_carrier_manifest_orders_ref CHECK (
    shipment_id IS NOT NULL OR sales_order_id IS NOT NULL OR btrim(order_no) <> ''
  ),
  CONSTRAINT ck_carrier_manifest_orders_scan_status CHECK (
    scan_status IN (
      'pending',
      'scanned',
      'missing',
      'extra',
      'exception',
      'removed',
      'cancelled'
    )
  )
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_carrier_manifest_orders_manifest_shipment
  ON shipping.carrier_manifest_orders(carrier_manifest_id, shipment_id)
  WHERE shipment_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_carrier_manifest_orders_manifest_sales_order
  ON shipping.carrier_manifest_orders(carrier_manifest_id, sales_order_id)
  WHERE sales_order_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS ix_carrier_manifest_orders_scan_status
  ON shipping.carrier_manifest_orders(org_id, scan_status, carrier_manifest_id);

COMMIT;
