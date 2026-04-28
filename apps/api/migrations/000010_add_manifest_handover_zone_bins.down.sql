BEGIN;

DROP INDEX IF EXISTS shipping.ix_carrier_manifest_orders_handover_bin;
DROP INDEX IF EXISTS shipping.ix_carrier_manifests_handover_zone;

ALTER TABLE shipping.carrier_manifest_orders
  DROP COLUMN IF EXISTS handover_bin_code,
  DROP COLUMN IF EXISTS handover_bin_id,
  DROP COLUMN IF EXISTS handover_zone_code,
  DROP COLUMN IF EXISTS handover_zone_id;

ALTER TABLE shipping.carrier_manifest_lines
  DROP COLUMN IF EXISTS handover_bin_code,
  DROP COLUMN IF EXISTS handover_bin_id,
  DROP COLUMN IF EXISTS handover_zone_code,
  DROP COLUMN IF EXISTS handover_zone_id;

ALTER TABLE shipping.carrier_manifests
  DROP COLUMN IF EXISTS handover_bin_code,
  DROP COLUMN IF EXISTS handover_bin_id,
  DROP COLUMN IF EXISTS handover_zone_code,
  DROP COLUMN IF EXISTS handover_zone_id;

COMMIT;
