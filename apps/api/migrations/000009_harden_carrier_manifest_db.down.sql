BEGIN;

DROP INDEX IF EXISTS shipping.ix_carrier_manifest_orders_scan_status;
DROP INDEX IF EXISTS shipping.uq_carrier_manifest_orders_manifest_sales_order;
DROP INDEX IF EXISTS shipping.uq_carrier_manifest_orders_manifest_shipment;
DROP TABLE IF EXISTS shipping.carrier_manifest_orders;

DROP INDEX IF EXISTS shipping.ix_carrier_manifests_status_date;

ALTER TABLE shipping.carrier_manifests
  DROP CONSTRAINT IF EXISTS ck_carrier_manifests_status,
  DROP CONSTRAINT IF EXISTS ck_carrier_manifests_counts,
  DROP COLUMN IF EXISTS note,
  DROP COLUMN IF EXISTS carrier_signature_ref,
  DROP COLUMN IF EXISTS carrier_receiver_name,
  DROP COLUMN IF EXISTS handed_over_by,
  DROP COLUMN IF EXISTS handed_over_at,
  DROP COLUMN IF EXISTS handover_zone,
  DROP COLUMN IF EXISTS handover_batch,
  ADD CONSTRAINT ck_carrier_manifests_status CHECK (
    status IN (
      'draft',
      'ready',
      'scanning',
      'completed',
      'exception',
      'cancelled'
    )
  ),
  ADD CONSTRAINT ck_carrier_manifests_counts CHECK (
    expected_count >= 0
    AND scanned_count >= 0
    AND missing_count >= 0
  );

ALTER TABLE mdm.carriers
  DROP CONSTRAINT IF EXISTS ck_carriers_status,
  DROP COLUMN IF EXISTS sla_profile,
  DROP COLUMN IF EXISTS handover_zone,
  ADD CONSTRAINT ck_carriers_status CHECK (status IN ('active', 'inactive', 'blocked'));

COMMIT;
