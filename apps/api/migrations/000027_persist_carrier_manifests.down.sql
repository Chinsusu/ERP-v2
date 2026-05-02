BEGIN;

DROP INDEX IF EXISTS shipping.ix_scan_events_manifest_ref;
DROP INDEX IF EXISTS shipping.uq_scan_events_scan_ref;
DROP INDEX IF EXISTS shipping.ix_carrier_manifest_orders_runtime_scan;
DROP INDEX IF EXISTS shipping.uq_carrier_manifest_orders_ref;
DROP INDEX IF EXISTS shipping.ix_carrier_manifests_runtime_filters;
DROP INDEX IF EXISTS shipping.uq_carrier_manifests_org_ref;

ALTER TABLE shipping.scan_events
  DROP COLUMN IF EXISTS carrier_code,
  DROP COLUMN IF EXISTS warehouse_ref,
  DROP COLUMN IF EXISTS actor_ref,
  DROP COLUMN IF EXISTS shipment_ref,
  DROP COLUMN IF EXISTS expected_manifest_ref,
  DROP COLUMN IF EXISTS manifest_ref,
  DROP COLUMN IF EXISTS scan_ref;

ALTER TABLE shipping.carrier_manifest_orders
  DROP COLUMN IF EXISTS updated_by_ref,
  DROP COLUMN IF EXISTS created_by_ref,
  DROP COLUMN IF EXISTS scanned_by_ref,
  DROP COLUMN IF EXISTS sales_order_ref,
  DROP COLUMN IF EXISTS shipment_ref,
  DROP COLUMN IF EXISTS manifest_ref,
  DROP COLUMN IF EXISTS line_ref;

ALTER TABLE shipping.carrier_manifests
  DROP COLUMN IF EXISTS updated_by_ref,
  DROP COLUMN IF EXISTS created_by_ref,
  DROP COLUMN IF EXISTS handed_over_by_ref,
  DROP COLUMN IF EXISTS completed_by_ref,
  DROP COLUMN IF EXISTS owner_ref,
  DROP COLUMN IF EXISTS carrier_name,
  DROP COLUMN IF EXISTS carrier_code,
  DROP COLUMN IF EXISTS carrier_ref,
  DROP COLUMN IF EXISTS warehouse_code,
  DROP COLUMN IF EXISTS warehouse_ref,
  DROP COLUMN IF EXISTS org_ref,
  DROP COLUMN IF EXISTS manifest_ref;

COMMIT;
