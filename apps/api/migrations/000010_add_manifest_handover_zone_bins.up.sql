BEGIN;

ALTER TABLE shipping.carrier_manifests
  ADD COLUMN IF NOT EXISTS handover_zone_id uuid REFERENCES mdm.warehouse_zones(id),
  ADD COLUMN IF NOT EXISTS handover_zone_code text,
  ADD COLUMN IF NOT EXISTS handover_bin_id uuid REFERENCES mdm.warehouse_bins(id),
  ADD COLUMN IF NOT EXISTS handover_bin_code text;

UPDATE shipping.carrier_manifests
SET handover_zone_code = COALESCE(handover_zone_code, handover_zone)
WHERE handover_zone IS NOT NULL;

ALTER TABLE shipping.carrier_manifest_lines
  ADD COLUMN IF NOT EXISTS handover_zone_id uuid REFERENCES mdm.warehouse_zones(id),
  ADD COLUMN IF NOT EXISTS handover_zone_code text,
  ADD COLUMN IF NOT EXISTS handover_bin_id uuid REFERENCES mdm.warehouse_bins(id),
  ADD COLUMN IF NOT EXISTS handover_bin_code text;

UPDATE shipping.carrier_manifest_lines
SET handover_zone_code = COALESCE(handover_zone_code, staging_zone)
WHERE staging_zone IS NOT NULL;

ALTER TABLE shipping.carrier_manifest_orders
  ADD COLUMN IF NOT EXISTS handover_zone_id uuid REFERENCES mdm.warehouse_zones(id),
  ADD COLUMN IF NOT EXISTS handover_zone_code text,
  ADD COLUMN IF NOT EXISTS handover_bin_id uuid REFERENCES mdm.warehouse_bins(id),
  ADD COLUMN IF NOT EXISTS handover_bin_code text;

UPDATE shipping.carrier_manifest_orders
SET handover_zone_code = COALESCE(handover_zone_code, staging_zone)
WHERE staging_zone IS NOT NULL;

CREATE INDEX IF NOT EXISTS ix_carrier_manifests_handover_zone
  ON shipping.carrier_manifests(org_id, warehouse_id, handover_zone_code, status);

CREATE INDEX IF NOT EXISTS ix_carrier_manifest_orders_handover_bin
  ON shipping.carrier_manifest_orders(org_id, handover_zone_code, handover_bin_code, scan_status);

COMMIT;
