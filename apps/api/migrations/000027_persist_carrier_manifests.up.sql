BEGIN;

ALTER TABLE shipping.carrier_manifests
  ADD COLUMN IF NOT EXISTS manifest_ref text,
  ADD COLUMN IF NOT EXISTS org_ref text,
  ADD COLUMN IF NOT EXISTS warehouse_ref text,
  ADD COLUMN IF NOT EXISTS warehouse_code text,
  ADD COLUMN IF NOT EXISTS carrier_ref text,
  ADD COLUMN IF NOT EXISTS carrier_code text,
  ADD COLUMN IF NOT EXISTS carrier_name text,
  ADD COLUMN IF NOT EXISTS owner_ref text,
  ADD COLUMN IF NOT EXISTS completed_by_ref text,
  ADD COLUMN IF NOT EXISTS handed_over_by_ref text,
  ADD COLUMN IF NOT EXISTS created_by_ref text,
  ADD COLUMN IF NOT EXISTS updated_by_ref text;

UPDATE shipping.carrier_manifests AS manifest
SET manifest_ref = COALESCE(NULLIF(btrim(manifest.manifest_ref), ''), manifest.manifest_no, manifest.id::text),
    org_ref = COALESCE(NULLIF(btrim(manifest.org_ref), ''), manifest.org_id::text),
    warehouse_ref = COALESCE(NULLIF(btrim(manifest.warehouse_ref), ''), manifest.warehouse_id::text),
    warehouse_code = COALESCE(NULLIF(btrim(manifest.warehouse_code), ''), warehouse.code),
    carrier_ref = COALESCE(NULLIF(btrim(manifest.carrier_ref), ''), manifest.carrier_id::text),
    carrier_code = COALESCE(NULLIF(btrim(manifest.carrier_code), ''), carrier.code),
    carrier_name = COALESCE(NULLIF(btrim(manifest.carrier_name), ''), carrier.name),
    owner_ref = COALESCE(NULLIF(btrim(manifest.owner_ref), ''), NULLIF(btrim(manifest.created_by_ref), ''), manifest.created_by::text),
    completed_by_ref = COALESCE(NULLIF(btrim(manifest.completed_by_ref), ''), manifest.completed_by::text),
    handed_over_by_ref = COALESCE(NULLIF(btrim(manifest.handed_over_by_ref), ''), manifest.handed_over_by::text),
    created_by_ref = COALESCE(NULLIF(btrim(manifest.created_by_ref), ''), manifest.created_by::text),
    updated_by_ref = COALESCE(NULLIF(btrim(manifest.updated_by_ref), ''), manifest.updated_by::text)
FROM mdm.warehouses AS warehouse,
     mdm.carriers AS carrier
WHERE warehouse.id = manifest.warehouse_id
  AND carrier.id = manifest.carrier_id;

UPDATE shipping.carrier_manifests
SET manifest_ref = COALESCE(NULLIF(btrim(manifest_ref), ''), manifest_no, id::text),
    org_ref = COALESCE(NULLIF(btrim(org_ref), ''), org_id::text),
    warehouse_ref = COALESCE(NULLIF(btrim(warehouse_ref), ''), warehouse_id::text),
    carrier_ref = COALESCE(NULLIF(btrim(carrier_ref), ''), carrier_id::text),
    owner_ref = COALESCE(NULLIF(btrim(owner_ref), ''), NULLIF(btrim(created_by_ref), ''), created_by::text),
    completed_by_ref = COALESCE(NULLIF(btrim(completed_by_ref), ''), completed_by::text),
    handed_over_by_ref = COALESCE(NULLIF(btrim(handed_over_by_ref), ''), handed_over_by::text),
    created_by_ref = COALESCE(NULLIF(btrim(created_by_ref), ''), created_by::text),
    updated_by_ref = COALESCE(NULLIF(btrim(updated_by_ref), ''), updated_by::text)
WHERE manifest_ref IS NULL
   OR org_ref IS NULL
   OR warehouse_ref IS NULL
   OR carrier_ref IS NULL
   OR owner_ref IS NULL
   OR completed_by_ref IS NULL
   OR handed_over_by_ref IS NULL
   OR created_by_ref IS NULL
   OR updated_by_ref IS NULL;

ALTER TABLE shipping.carrier_manifests
  ALTER COLUMN manifest_ref SET NOT NULL,
  ALTER COLUMN org_ref SET NOT NULL,
  ALTER COLUMN warehouse_ref SET NOT NULL,
  ALTER COLUMN carrier_ref SET NOT NULL;

ALTER TABLE shipping.carrier_manifest_orders
  ADD COLUMN IF NOT EXISTS line_ref text,
  ADD COLUMN IF NOT EXISTS manifest_ref text,
  ADD COLUMN IF NOT EXISTS shipment_ref text,
  ADD COLUMN IF NOT EXISTS sales_order_ref text,
  ADD COLUMN IF NOT EXISTS scanned_by_ref text,
  ADD COLUMN IF NOT EXISTS created_by_ref text,
  ADD COLUMN IF NOT EXISTS updated_by_ref text;

UPDATE shipping.carrier_manifest_orders AS manifest_order
SET line_ref = COALESCE(NULLIF(btrim(manifest_order.line_ref), ''), manifest_order.id::text),
    manifest_ref = COALESCE(NULLIF(btrim(manifest_order.manifest_ref), ''), manifest.manifest_ref, manifest.id::text),
    shipment_ref = COALESCE(
      NULLIF(btrim(manifest_order.shipment_ref), ''),
      (SELECT shipment.shipment_no FROM shipping.shipments AS shipment WHERE shipment.id = manifest_order.shipment_id),
      manifest_order.shipment_id::text
    ),
    sales_order_ref = COALESCE(
      NULLIF(btrim(manifest_order.sales_order_ref), ''),
      (
        SELECT COALESCE(sales_order.order_ref, sales_order.order_no)
        FROM sales.sales_orders AS sales_order
        WHERE sales_order.id = manifest_order.sales_order_id
      ),
      manifest_order.sales_order_id::text
    ),
    scanned_by_ref = COALESCE(NULLIF(btrim(manifest_order.scanned_by_ref), ''), manifest_order.scanned_by::text),
    created_by_ref = COALESCE(NULLIF(btrim(manifest_order.created_by_ref), ''), manifest_order.created_by::text),
    updated_by_ref = COALESCE(NULLIF(btrim(manifest_order.updated_by_ref), ''), manifest_order.updated_by::text)
FROM shipping.carrier_manifests AS manifest
WHERE manifest.id = manifest_order.carrier_manifest_id;

UPDATE shipping.carrier_manifest_orders
SET line_ref = COALESCE(NULLIF(btrim(line_ref), ''), id::text),
    shipment_ref = COALESCE(NULLIF(btrim(shipment_ref), ''), shipment_id::text),
    sales_order_ref = COALESCE(NULLIF(btrim(sales_order_ref), ''), sales_order_id::text),
    scanned_by_ref = COALESCE(NULLIF(btrim(scanned_by_ref), ''), scanned_by::text),
    created_by_ref = COALESCE(NULLIF(btrim(created_by_ref), ''), created_by::text),
    updated_by_ref = COALESCE(NULLIF(btrim(updated_by_ref), ''), updated_by::text)
WHERE line_ref IS NULL
   OR shipment_ref IS NULL
   OR sales_order_ref IS NULL
   OR scanned_by_ref IS NULL
   OR created_by_ref IS NULL
   OR updated_by_ref IS NULL;

ALTER TABLE shipping.carrier_manifest_orders
  ALTER COLUMN line_ref SET NOT NULL,
  ALTER COLUMN manifest_ref SET NOT NULL;

ALTER TABLE shipping.scan_events
  ADD COLUMN IF NOT EXISTS scan_ref text,
  ADD COLUMN IF NOT EXISTS manifest_ref text,
  ADD COLUMN IF NOT EXISTS expected_manifest_ref text,
  ADD COLUMN IF NOT EXISTS shipment_ref text,
  ADD COLUMN IF NOT EXISTS actor_ref text,
  ADD COLUMN IF NOT EXISTS warehouse_ref text,
  ADD COLUMN IF NOT EXISTS carrier_code text;

UPDATE shipping.scan_events AS event
SET scan_ref = COALESCE(NULLIF(btrim(event.scan_ref), ''), event.idempotency_key, event.id::text),
    manifest_ref = COALESCE(NULLIF(btrim(event.manifest_ref), ''), manifest.manifest_ref, event.carrier_manifest_id::text),
    shipment_ref = COALESCE(
      NULLIF(btrim(event.shipment_ref), ''),
      (SELECT shipment.shipment_no FROM shipping.shipments AS shipment WHERE shipment.id = event.shipment_id),
      event.shipment_id::text
    ),
    actor_ref = COALESCE(NULLIF(btrim(event.actor_ref), ''), event.scanned_by::text),
    warehouse_ref = COALESCE(NULLIF(btrim(event.warehouse_ref), ''), manifest.warehouse_ref, manifest.warehouse_id::text),
    carrier_code = COALESCE(NULLIF(btrim(event.carrier_code), ''), manifest.carrier_code)
FROM shipping.carrier_manifests AS manifest
WHERE manifest.id = event.carrier_manifest_id;

UPDATE shipping.scan_events
SET scan_ref = COALESCE(NULLIF(btrim(scan_ref), ''), idempotency_key, id::text),
    manifest_ref = COALESCE(NULLIF(btrim(manifest_ref), ''), carrier_manifest_id::text),
    shipment_ref = COALESCE(NULLIF(btrim(shipment_ref), ''), shipment_id::text),
    actor_ref = COALESCE(NULLIF(btrim(actor_ref), ''), scanned_by::text)
WHERE scan_ref IS NULL
   OR manifest_ref IS NULL
   OR shipment_ref IS NULL
   OR actor_ref IS NULL;

ALTER TABLE shipping.scan_events
  ALTER COLUMN scan_ref SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_carrier_manifests_org_ref
  ON shipping.carrier_manifests(org_id, manifest_ref);

CREATE INDEX IF NOT EXISTS ix_carrier_manifests_runtime_filters
  ON shipping.carrier_manifests(org_id, warehouse_ref, handover_date, carrier_code, status);

CREATE UNIQUE INDEX IF NOT EXISTS uq_carrier_manifest_orders_ref
  ON shipping.carrier_manifest_orders(org_id, carrier_manifest_id, line_ref);

CREATE INDEX IF NOT EXISTS ix_carrier_manifest_orders_runtime_scan
  ON shipping.carrier_manifest_orders(org_id, carrier_manifest_id, scan_status, tracking_no, shipment_ref);

CREATE UNIQUE INDEX IF NOT EXISTS uq_scan_events_scan_ref
  ON shipping.scan_events(org_id, scan_ref);

CREATE INDEX IF NOT EXISTS ix_scan_events_manifest_ref
  ON shipping.scan_events(org_id, manifest_ref, scanned_at DESC);

COMMIT;
