package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/domain"
)

type PostgresCarrierManifestStoreConfig struct {
	DefaultOrgID string
}

type PostgresCarrierManifestStore struct {
	db           *sql.DB
	defaultOrgID string
}

type postgresCarrierManifestQueryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func NewPostgresCarrierManifestStore(
	db *sql.DB,
	cfg PostgresCarrierManifestStoreConfig,
) PostgresCarrierManifestStore {
	return PostgresCarrierManifestStore{
		db:           db,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
	}
}

const selectCarrierManifestOrgIDSQL = `
SELECT id::text
FROM core.organizations
WHERE code = $1
LIMIT 1`

const selectCarrierManifestWarehouseSQL = `
SELECT id::text, code
FROM mdm.warehouses
WHERE org_id = $1::uuid
  AND (
    id::text = $2
    OR lower(code) = lower($2)
    OR lower(code) = lower($3)
  )
LIMIT 1`

const selectCarrierManifestCarrierSQL = `
SELECT id::text, code, name
FROM mdm.carriers
WHERE org_id = $1::uuid
  AND (
    id::text = $2
    OR lower(code) = lower($2)
    OR lower(code) = lower($3)
  )
LIMIT 1`

const selectCarrierManifestHeadersBaseSQL = `
SELECT
  manifest.id::text,
  COALESCE(manifest.manifest_ref, manifest.manifest_no, manifest.id::text),
  COALESCE(manifest.carrier_code, carrier.code, ''),
  COALESCE(manifest.carrier_name, carrier.name, manifest.carrier_code, carrier.code, ''),
  COALESCE(manifest.warehouse_ref, manifest.warehouse_id::text, ''),
  COALESCE(manifest.warehouse_code, warehouse.code, ''),
  manifest.handover_date::text,
  COALESCE(manifest.handover_batch, ''),
  COALESCE(manifest.handover_zone, ''),
  COALESCE(manifest.handover_zone_id::text, ''),
  COALESCE(manifest.handover_zone_code, ''),
  COALESCE(manifest.handover_bin_id::text, ''),
  COALESCE(manifest.handover_bin_code, ''),
  manifest.status,
  COALESCE(manifest.owner_ref, manifest.created_by_ref, manifest.created_by::text, ''),
  manifest.created_at
FROM shipping.carrier_manifests AS manifest
LEFT JOIN mdm.carriers AS carrier ON carrier.id = manifest.carrier_id
LEFT JOIN mdm.warehouses AS warehouse ON warehouse.id = manifest.warehouse_id`

const selectCarrierManifestHeadersSQL = selectCarrierManifestHeadersBaseSQL + `
ORDER BY manifest.handover_date DESC, COALESCE(manifest.warehouse_code, warehouse.code, ''), COALESCE(manifest.carrier_code, carrier.code, ''), manifest.handover_batch`

const findCarrierManifestHeaderSQL = selectCarrierManifestHeadersBaseSQL + `
WHERE lower(COALESCE(manifest.manifest_ref, manifest.manifest_no, manifest.id::text)) = lower($1)
   OR manifest.id::text = $1
LIMIT 1`

const findCarrierManifestPersistedIDSQL = `
SELECT id::text, org_id::text
FROM shipping.carrier_manifests
WHERE lower(COALESCE(manifest_ref, manifest_no, id::text)) = lower($1)
   OR id::text = $1
LIMIT 1
FOR UPDATE`

const selectCarrierManifestLinesSQL = `
SELECT
  COALESCE(line.line_ref, line.id::text),
  COALESCE(line.shipment_ref, line.shipment_id::text, ''),
  line.order_no,
  COALESCE(line.tracking_no, ''),
  COALESCE(line.package_code, ''),
  COALESCE(line.staging_zone, ''),
  COALESCE(line.handover_zone_id::text, ''),
  COALESCE(line.handover_zone_code, ''),
  COALESCE(line.handover_bin_id::text, ''),
  COALESCE(line.handover_bin_code, ''),
  line.scan_status
FROM shipping.carrier_manifest_orders AS line
WHERE line.carrier_manifest_id = $1::uuid
  AND line.scan_status <> 'removed'
ORDER BY line.line_no, line.created_at, COALESCE(line.line_ref, line.id::text)`

const upsertCarrierManifestSQL = `
INSERT INTO shipping.carrier_manifests (
  id,
  org_id,
  manifest_ref,
  org_ref,
  manifest_no,
  warehouse_id,
  warehouse_ref,
  warehouse_code,
  carrier_id,
  carrier_ref,
  carrier_code,
  carrier_name,
  handover_date,
  handover_batch,
  handover_zone,
  handover_zone_id,
  handover_zone_code,
  handover_bin_id,
  handover_bin_code,
  status,
  expected_count,
  scanned_count,
  missing_count,
  completed_at,
  completed_by,
  completed_by_ref,
  handed_over_at,
  handed_over_by,
  handed_over_by_ref,
  owner_ref,
  created_at,
  created_by,
  created_by_ref,
  updated_at,
  updated_by,
  updated_by_ref
) VALUES (
  COALESCE($1::uuid, gen_random_uuid()),
  $2::uuid,
  $3,
  $4,
  $5,
  $6::uuid,
  $7,
  $8,
  $9::uuid,
  $10,
  $11,
  $12,
  $13::date,
  $14,
  $15,
  $16::uuid,
  $17,
  $18::uuid,
  $19,
  $20,
  $21,
  $22,
  $23,
  $24,
  $25::uuid,
  $26,
  $27,
  $28::uuid,
  $29,
  $30,
  $31,
  $32::uuid,
  $33,
  $34,
  $35::uuid,
  $36
)
ON CONFLICT (org_id, manifest_ref)
DO UPDATE SET
  org_ref = EXCLUDED.org_ref,
  manifest_no = EXCLUDED.manifest_no,
  warehouse_id = EXCLUDED.warehouse_id,
  warehouse_ref = EXCLUDED.warehouse_ref,
  warehouse_code = EXCLUDED.warehouse_code,
  carrier_id = EXCLUDED.carrier_id,
  carrier_ref = EXCLUDED.carrier_ref,
  carrier_code = EXCLUDED.carrier_code,
  carrier_name = EXCLUDED.carrier_name,
  handover_date = EXCLUDED.handover_date,
  handover_batch = EXCLUDED.handover_batch,
  handover_zone = EXCLUDED.handover_zone,
  handover_zone_id = EXCLUDED.handover_zone_id,
  handover_zone_code = EXCLUDED.handover_zone_code,
  handover_bin_id = EXCLUDED.handover_bin_id,
  handover_bin_code = EXCLUDED.handover_bin_code,
  status = EXCLUDED.status,
  expected_count = EXCLUDED.expected_count,
  scanned_count = EXCLUDED.scanned_count,
  missing_count = EXCLUDED.missing_count,
  completed_at = EXCLUDED.completed_at,
  completed_by = EXCLUDED.completed_by,
  completed_by_ref = EXCLUDED.completed_by_ref,
  handed_over_at = EXCLUDED.handed_over_at,
  handed_over_by = EXCLUDED.handed_over_by,
  handed_over_by_ref = EXCLUDED.handed_over_by_ref,
  owner_ref = EXCLUDED.owner_ref,
  updated_at = EXCLUDED.updated_at,
  updated_by = EXCLUDED.updated_by,
  updated_by_ref = EXCLUDED.updated_by_ref,
  version = shipping.carrier_manifests.version + 1
RETURNING id::text`

const deleteCarrierManifestOrdersSQL = `
DELETE FROM shipping.carrier_manifest_orders
WHERE carrier_manifest_id = $1::uuid`

const insertCarrierManifestOrderSQL = `
INSERT INTO shipping.carrier_manifest_orders (
  id,
  org_id,
  carrier_manifest_id,
  line_ref,
  manifest_ref,
  line_no,
  shipment_id,
  shipment_ref,
  sales_order_id,
  sales_order_ref,
  order_no,
  tracking_no,
  package_code,
  staging_zone,
  handover_zone_id,
  handover_zone_code,
  handover_bin_id,
  handover_bin_code,
  scan_status,
  scanned_at,
  scanned_by,
  scanned_by_ref,
  created_at,
  created_by,
  created_by_ref,
  updated_at,
  updated_by,
  updated_by_ref
) VALUES (
  COALESCE($1::uuid, gen_random_uuid()),
  $2::uuid,
  $3::uuid,
  $4,
  $5,
  $6,
  $7::uuid,
  $8,
  $9::uuid,
  $10,
  $11,
  $12,
  $13,
  $14,
  $15::uuid,
  $16,
  $17::uuid,
  $18,
  $19,
  $20,
  $21::uuid,
  $22,
  $23,
  $24::uuid,
  $25,
  $26,
  $27::uuid,
  $28
)`

const selectPackedShipmentByIDSQL = `
SELECT
  COALESCE(shipment.shipment_no, shipment.id::text),
  COALESCE(sales_order.order_no, ''),
  COALESCE(shipment.tracking_no, ''),
  COALESCE(carrier.code, ''),
  COALESCE(carrier.name, ''),
  COALESCE(warehouse.id::text, shipment.warehouse_id::text, ''),
  COALESCE(warehouse.code, ''),
  COALESCE(shipment.shipment_no, shipment.tracking_no, ''),
  shipment.status
FROM shipping.shipments AS shipment
LEFT JOIN sales.sales_orders AS sales_order ON sales_order.id = shipment.sales_order_id
LEFT JOIN mdm.carriers AS carrier ON carrier.id = shipment.carrier_id
LEFT JOIN mdm.warehouses AS warehouse ON warehouse.id = shipment.warehouse_id
WHERE lower(COALESCE(shipment.shipment_no, shipment.id::text)) = lower($1)
   OR shipment.id::text = $1
LIMIT 1`

const selectPackedShipmentByCodeSQL = `
SELECT
  COALESCE(shipment.shipment_no, shipment.id::text),
  COALESCE(sales_order.order_no, ''),
  COALESCE(shipment.tracking_no, ''),
  COALESCE(carrier.code, ''),
  COALESCE(carrier.name, ''),
  COALESCE(warehouse.id::text, shipment.warehouse_id::text, ''),
  COALESCE(warehouse.code, ''),
  COALESCE(shipment.shipment_no, shipment.tracking_no, ''),
  shipment.status
FROM shipping.shipments AS shipment
LEFT JOIN sales.sales_orders AS sales_order ON sales_order.id = shipment.sales_order_id
LEFT JOIN mdm.carriers AS carrier ON carrier.id = shipment.carrier_id
LEFT JOIN mdm.warehouses AS warehouse ON warehouse.id = shipment.warehouse_id
WHERE upper(shipment.id::text) = upper($1)
   OR upper(COALESCE(shipment.shipment_no, '')) = upper($1)
   OR upper(COALESCE(shipment.tracking_no, '')) = upper($1)
   OR upper(COALESCE(sales_order.order_no, '')) = upper($1)
ORDER BY shipment.created_at DESC
LIMIT 1`

const findCarrierManifestByLineCodeSQL = `
SELECT COALESCE(manifest.manifest_ref, manifest.manifest_no, manifest.id::text)
FROM shipping.carrier_manifest_orders AS line
JOIN shipping.carrier_manifests AS manifest ON manifest.id = line.carrier_manifest_id
WHERE upper(COALESCE(line.shipment_ref, line.shipment_id::text, '')) = upper($1)
   OR upper(line.order_no) = upper($1)
   OR upper(COALESCE(line.tracking_no, '')) = upper($1)
   OR upper(COALESCE(line.package_code, '')) = upper($1)
ORDER BY line.updated_at DESC, line.created_at DESC
LIMIT 1`

const findCarrierManifestForScanEventSQL = `
SELECT id::text, org_id::text, COALESCE(manifest_ref, manifest_no, id::text)
FROM shipping.carrier_manifests
WHERE lower(COALESCE(manifest_ref, manifest_no, id::text)) = lower($1)
   OR id::text = $1
LIMIT 1`

const insertCarrierManifestScanEventSQL = `
INSERT INTO shipping.scan_events (
  id,
  org_id,
  carrier_manifest_id,
  manifest_ref,
  expected_manifest_ref,
  shipment_id,
  shipment_ref,
  barcode,
  scan_context,
  scan_result,
  error_code,
  scan_station,
  scanned_at,
  scanned_by,
  actor_ref,
  idempotency_key,
  scan_ref,
  warehouse_ref,
  carrier_code,
  metadata
) VALUES (
  COALESCE($1::uuid, gen_random_uuid()),
  $2::uuid,
  $3::uuid,
  $4,
  $5,
  $6::uuid,
  $7,
  $8,
  'handover',
  $9,
  $10,
  $11,
  $12,
  $13::uuid,
  $14,
  $15,
  $16,
  $17,
  $18,
  $19::jsonb
)
ON CONFLICT (org_id, scan_ref)
DO UPDATE SET
  carrier_manifest_id = EXCLUDED.carrier_manifest_id,
  manifest_ref = EXCLUDED.manifest_ref,
  expected_manifest_ref = EXCLUDED.expected_manifest_ref,
  shipment_id = EXCLUDED.shipment_id,
  shipment_ref = EXCLUDED.shipment_ref,
  barcode = EXCLUDED.barcode,
  scan_result = EXCLUDED.scan_result,
  error_code = EXCLUDED.error_code,
  scan_station = EXCLUDED.scan_station,
  scanned_at = EXCLUDED.scanned_at,
  scanned_by = EXCLUDED.scanned_by,
  actor_ref = EXCLUDED.actor_ref,
  warehouse_ref = EXCLUDED.warehouse_ref,
  carrier_code = EXCLUDED.carrier_code,
  metadata = EXCLUDED.metadata`

const listCarrierManifestScanEventsSQL = `
SELECT
  COALESCE(scan_ref, idempotency_key, id::text),
  COALESCE(manifest_ref, carrier_manifest_id::text, ''),
  COALESCE(expected_manifest_ref, ''),
  barcode,
  scan_result,
  COALESCE(metadata->>'severity', ''),
  COALESCE(metadata->>'message', ''),
  COALESCE(shipment_ref, shipment_id::text, ''),
  COALESCE(metadata->>'order_no', ''),
  COALESCE(metadata->>'tracking_no', ''),
  COALESCE(actor_ref, scanned_by::text, ''),
  COALESCE(scan_station, ''),
  COALESCE(metadata->>'device_id', ''),
  COALESCE(metadata->>'source', ''),
  COALESCE(warehouse_ref, ''),
  COALESCE(carrier_code, ''),
  scanned_at
FROM shipping.scan_events
WHERE scan_context = 'handover'
  AND (NULLIF($1, '') IS NULL OR manifest_ref = $1 OR carrier_manifest_id::text = $1)
ORDER BY scanned_at DESC, COALESCE(scan_ref, idempotency_key, id::text)`

func (s PostgresCarrierManifestStore) List(
	ctx context.Context,
	filter domain.CarrierManifestFilter,
) ([]domain.CarrierManifest, error) {
	if s.db == nil {
		return nil, errors.New("database connection is required")
	}
	rows, err := s.db.QueryContext(ctx, selectCarrierManifestHeadersSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	manifests := make([]domain.CarrierManifest, 0)
	for rows.Next() {
		manifest, err := scanPostgresCarrierManifest(ctx, s.db, rows)
		if err != nil {
			return nil, err
		}
		if carrierManifestMatchesFilter(manifest, filter) {
			manifests = append(manifests, manifest)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	domain.SortCarrierManifests(manifests)

	return manifests, nil
}

func (s PostgresCarrierManifestStore) Get(ctx context.Context, id string) (domain.CarrierManifest, error) {
	if s.db == nil {
		return domain.CarrierManifest{}, errors.New("database connection is required")
	}
	row := s.db.QueryRowContext(ctx, findCarrierManifestHeaderSQL, strings.TrimSpace(id))
	manifest, err := scanPostgresCarrierManifest(ctx, s.db, row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.CarrierManifest{}, ErrCarrierManifestNotFound
	}
	if err != nil {
		return domain.CarrierManifest{}, err
	}

	return manifest, nil
}

func (s PostgresCarrierManifestStore) Save(ctx context.Context, manifest domain.CarrierManifest) error {
	if s.db == nil {
		return errors.New("database connection is required")
	}
	if strings.TrimSpace(manifest.ID) == "" {
		return errors.New("carrier manifest id is required")
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin carrier manifest transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	persistedID, orgID, err := findPostgresCarrierManifest(ctx, tx, manifest.ID)
	if err != nil {
		return err
	}
	if orgID == "" {
		orgID, err = s.resolveOrgID(ctx, tx, "")
		if err != nil {
			return err
		}
	}
	warehouse, err := resolvePostgresCarrierManifestWarehouse(ctx, tx, orgID, manifest)
	if err != nil {
		return err
	}
	carrier, err := resolvePostgresCarrierManifestCarrier(ctx, tx, orgID, manifest)
	if err != nil {
		return err
	}
	persistedID, err = upsertPostgresCarrierManifest(ctx, tx, orgID, persistedID, warehouse, carrier, manifest)
	if err != nil {
		return err
	}
	if err := replacePostgresCarrierManifestOrders(ctx, tx, orgID, persistedID, manifest); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit carrier manifest transaction: %w", err)
	}
	committed = true

	return nil
}

func (s PostgresCarrierManifestStore) GetPackedShipment(
	ctx context.Context,
	id string,
) (domain.PackedShipment, error) {
	if s.db == nil {
		return domain.PackedShipment{}, errors.New("database connection is required")
	}
	row := s.db.QueryRowContext(ctx, selectPackedShipmentByIDSQL, strings.TrimSpace(id))
	shipment, err := scanPostgresPackedShipment(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PackedShipment{}, ErrPackedShipmentNotFound
	}
	if err != nil {
		return domain.PackedShipment{}, err
	}

	return shipment, nil
}

func (s PostgresCarrierManifestStore) FindPackedShipmentByCode(
	ctx context.Context,
	code string,
) (domain.PackedShipment, error) {
	if s.db == nil {
		return domain.PackedShipment{}, errors.New("database connection is required")
	}
	normalizedCode := domain.NormalizeManifestScanCode(code)
	if normalizedCode == "" {
		return domain.PackedShipment{}, ErrPackedShipmentNotFound
	}
	row := s.db.QueryRowContext(ctx, selectPackedShipmentByCodeSQL, normalizedCode)
	shipment, err := scanPostgresPackedShipment(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PackedShipment{}, ErrPackedShipmentNotFound
	}
	if err != nil {
		return domain.PackedShipment{}, err
	}

	return shipment, nil
}

func (s PostgresCarrierManifestStore) FindCarrierManifestLineByCode(
	ctx context.Context,
	code string,
) (domain.CarrierManifest, domain.CarrierManifestLine, error) {
	if s.db == nil {
		return domain.CarrierManifest{}, domain.CarrierManifestLine{}, errors.New("database connection is required")
	}
	normalizedCode := domain.NormalizeManifestScanCode(code)
	if normalizedCode == "" {
		return domain.CarrierManifest{}, domain.CarrierManifestLine{}, ErrPackedShipmentNotFound
	}
	var manifestID string
	err := s.db.QueryRowContext(ctx, findCarrierManifestByLineCodeSQL, normalizedCode).Scan(&manifestID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.CarrierManifest{}, domain.CarrierManifestLine{}, ErrPackedShipmentNotFound
	}
	if err != nil {
		return domain.CarrierManifest{}, domain.CarrierManifestLine{}, err
	}
	manifest, err := s.Get(ctx, manifestID)
	if err != nil {
		return domain.CarrierManifest{}, domain.CarrierManifestLine{}, err
	}
	_, line, ok := manifest.FindLineByScanCode(normalizedCode)
	if !ok {
		return domain.CarrierManifest{}, domain.CarrierManifestLine{}, ErrPackedShipmentNotFound
	}

	return manifest, line, nil
}

func (s PostgresCarrierManifestStore) RecordScanEvent(
	ctx context.Context,
	event CarrierManifestScanEvent,
) error {
	if s.db == nil {
		return errors.New("database connection is required")
	}
	if strings.TrimSpace(event.ID) == "" {
		return errors.New("carrier manifest scan event id is required")
	}
	orgID := ""
	persistedManifestID := ""
	manifestRef := strings.TrimSpace(event.ManifestID)
	if manifestRef != "" {
		row := s.db.QueryRowContext(ctx, findCarrierManifestForScanEventSQL, manifestRef)
		if err := row.Scan(&persistedManifestID, &orgID, &manifestRef); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("find carrier manifest scan event manifest: %w", err)
		}
	}
	if orgID == "" {
		var err error
		orgID, err = s.resolveOrgID(ctx, s.db, "")
		if err != nil {
			return err
		}
	}
	metadata, err := carrierManifestScanEventMetadata(event)
	if err != nil {
		return err
	}
	createdAt := event.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	_, err = s.db.ExecContext(
		ctx,
		insertCarrierManifestScanEventSQL,
		nullablePostgresCarrierManifestUUID(event.ID),
		orgID,
		nullablePostgresCarrierManifestUUID(persistedManifestID),
		nullablePostgresCarrierManifestText(manifestRef),
		nullablePostgresCarrierManifestText(event.ExpectedManifestID),
		nullablePostgresCarrierManifestUUID(event.ShipmentID),
		nullablePostgresCarrierManifestText(event.ShipmentID),
		domain.NormalizeManifestScanCode(event.Code),
		postgresCarrierManifestScanResult(event.ResultCode),
		nullablePostgresCarrierManifestText(string(event.ResultCode)),
		nullablePostgresCarrierManifestText(event.StationID),
		createdAt.UTC(),
		nullablePostgresCarrierManifestUUID(event.ActorID),
		nullablePostgresCarrierManifestText(event.ActorID),
		nullablePostgresCarrierManifestText(event.ID),
		nullablePostgresCarrierManifestText(event.ID),
		nullablePostgresCarrierManifestText(event.WarehouseID),
		nullablePostgresCarrierManifestText(event.CarrierCode),
		metadata,
	)
	if err != nil {
		return fmt.Errorf("insert carrier manifest scan event: %w", err)
	}

	return nil
}

func (s PostgresCarrierManifestStore) ListScanEvents(
	ctx context.Context,
	manifestID string,
) ([]CarrierManifestScanEvent, error) {
	if s.db == nil {
		return nil, errors.New("database connection is required")
	}
	rows, err := s.db.QueryContext(ctx, listCarrierManifestScanEventsSQL, strings.TrimSpace(manifestID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]CarrierManifestScanEvent, 0)
	for rows.Next() {
		event, err := scanPostgresCarrierManifestScanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (s PostgresCarrierManifestStore) resolveOrgID(
	ctx context.Context,
	queryer postgresCarrierManifestQueryer,
	orgRef string,
) (string, error) {
	orgRef = strings.TrimSpace(orgRef)
	if isPostgresCarrierManifestUUIDText(orgRef) {
		return orgRef, nil
	}
	if orgRef != "" {
		var orgID string
		err := queryer.QueryRowContext(ctx, selectCarrierManifestOrgIDSQL, orgRef).Scan(&orgID)
		if err == nil && isPostgresCarrierManifestUUIDText(orgID) {
			return orgID, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("resolve carrier manifest org %q: %w", orgRef, err)
		}
	}
	if isPostgresCarrierManifestUUIDText(s.defaultOrgID) {
		return s.defaultOrgID, nil
	}

	return "", fmt.Errorf("carrier manifest org %q cannot be resolved", orgRef)
}

func scanPostgresCarrierManifest(
	ctx context.Context,
	queryer postgresCarrierManifestQueryer,
	row interface{ Scan(dest ...any) error },
) (domain.CarrierManifest, error) {
	var (
		persistedID string
		status      string
		manifest    domain.CarrierManifest
	)
	if err := row.Scan(
		&persistedID,
		&manifest.ID,
		&manifest.CarrierCode,
		&manifest.CarrierName,
		&manifest.WarehouseID,
		&manifest.WarehouseCode,
		&manifest.Date,
		&manifest.HandoverBatch,
		&manifest.StagingZone,
		&manifest.HandoverZoneID,
		&manifest.HandoverZoneCode,
		&manifest.HandoverBinID,
		&manifest.HandoverBinCode,
		&status,
		&manifest.Owner,
		&manifest.CreatedAt,
	); err != nil {
		return domain.CarrierManifest{}, err
	}
	manifest.CarrierCode = strings.ToUpper(strings.TrimSpace(manifest.CarrierCode))
	manifest.WarehouseCode = strings.ToUpper(strings.TrimSpace(manifest.WarehouseCode))
	manifest.HandoverZoneCode = strings.ToUpper(strings.TrimSpace(manifest.HandoverZoneCode))
	manifest.HandoverBinCode = strings.ToUpper(strings.TrimSpace(manifest.HandoverBinCode))
	manifest.Status = domain.NormalizeManifestStatus(domain.CarrierManifestStatus(status))
	if manifest.Status == "" {
		manifest.Status = domain.ManifestStatusDraft
	}
	lines, err := listPostgresCarrierManifestLines(ctx, queryer, persistedID)
	if err != nil {
		return domain.CarrierManifest{}, err
	}
	manifest.Lines = lines

	return manifest, nil
}

func listPostgresCarrierManifestLines(
	ctx context.Context,
	queryer postgresCarrierManifestQueryer,
	persistedID string,
) ([]domain.CarrierManifestLine, error) {
	rows, err := queryer.QueryContext(ctx, selectCarrierManifestLinesSQL, persistedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lines := make([]domain.CarrierManifestLine, 0)
	for rows.Next() {
		line, err := scanPostgresCarrierManifestLine(rows)
		if err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func scanPostgresCarrierManifestLine(row interface{ Scan(dest ...any) error }) (domain.CarrierManifestLine, error) {
	var line domain.CarrierManifestLine
	var scanStatus string
	if err := row.Scan(
		&line.ID,
		&line.ShipmentID,
		&line.OrderNo,
		&line.TrackingNo,
		&line.PackageCode,
		&line.StagingZone,
		&line.HandoverZoneID,
		&line.HandoverZoneCode,
		&line.HandoverBinID,
		&line.HandoverBinCode,
		&scanStatus,
	); err != nil {
		return domain.CarrierManifestLine{}, err
	}
	line.HandoverZoneCode = strings.ToUpper(strings.TrimSpace(line.HandoverZoneCode))
	line.HandoverBinCode = strings.ToUpper(strings.TrimSpace(line.HandoverBinCode))
	line.Scanned = strings.EqualFold(scanStatus, "scanned")

	return line, nil
}

type postgresCarrierManifestWarehouse struct {
	ID   string
	Code string
}

type postgresCarrierManifestCarrier struct {
	ID   string
	Code string
	Name string
}

func resolvePostgresCarrierManifestWarehouse(
	ctx context.Context,
	queryer postgresCarrierManifestQueryer,
	orgID string,
	manifest domain.CarrierManifest,
) (postgresCarrierManifestWarehouse, error) {
	warehouseRef := strings.TrimSpace(manifest.WarehouseID)
	warehouseCode := strings.TrimSpace(manifest.WarehouseCode)
	if isPostgresCarrierManifestUUIDText(warehouseRef) && warehouseCode != "" {
		return postgresCarrierManifestWarehouse{ID: warehouseRef, Code: warehouseCode}, nil
	}

	var warehouse postgresCarrierManifestWarehouse
	err := queryer.QueryRowContext(ctx, selectCarrierManifestWarehouseSQL, orgID, warehouseRef, warehouseCode).
		Scan(&warehouse.ID, &warehouse.Code)
	if err != nil {
		return postgresCarrierManifestWarehouse{}, fmt.Errorf("resolve carrier manifest warehouse %q: %w", warehouseRef, err)
	}

	return warehouse, nil
}

func resolvePostgresCarrierManifestCarrier(
	ctx context.Context,
	queryer postgresCarrierManifestQueryer,
	orgID string,
	manifest domain.CarrierManifest,
) (postgresCarrierManifestCarrier, error) {
	carrierRef := strings.TrimSpace(manifest.CarrierCode)
	carrierName := strings.TrimSpace(manifest.CarrierName)
	if isPostgresCarrierManifestUUIDText(carrierRef) && carrierName != "" {
		return postgresCarrierManifestCarrier{ID: carrierRef, Code: carrierRef, Name: carrierName}, nil
	}

	var carrier postgresCarrierManifestCarrier
	err := queryer.QueryRowContext(ctx, selectCarrierManifestCarrierSQL, orgID, carrierRef, carrierName).
		Scan(&carrier.ID, &carrier.Code, &carrier.Name)
	if err != nil {
		return postgresCarrierManifestCarrier{}, fmt.Errorf("resolve carrier manifest carrier %q: %w", carrierRef, err)
	}

	return carrier, nil
}

func findPostgresCarrierManifest(
	ctx context.Context,
	queryer postgresCarrierManifestQueryer,
	id string,
) (string, string, error) {
	var persistedID string
	var orgID string
	err := queryer.QueryRowContext(ctx, findCarrierManifestPersistedIDSQL, strings.TrimSpace(id)).
		Scan(&persistedID, &orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", nil
	}
	if err != nil {
		return "", "", fmt.Errorf("find carrier manifest %q: %w", id, err)
	}

	return persistedID, orgID, nil
}

func upsertPostgresCarrierManifest(
	ctx context.Context,
	queryer postgresCarrierManifestQueryer,
	orgID string,
	persistedID string,
	warehouse postgresCarrierManifestWarehouse,
	carrier postgresCarrierManifestCarrier,
	manifest domain.CarrierManifest,
) (string, error) {
	summary := manifest.Summary()
	createdAt := postgresCarrierManifestTime(manifest.CreatedAt)
	completedAt := nullablePostgresCarrierManifestStatusTime(manifest.Status, domain.ManifestStatusCompleted, createdAt)
	handedOverAt := nullablePostgresCarrierManifestStatusTime(manifest.Status, domain.ManifestStatusHandedOver, createdAt)
	var savedID string
	err := queryer.QueryRowContext(
		ctx,
		upsertCarrierManifestSQL,
		nullablePostgresCarrierManifestUUID(firstNonBlankPostgresCarrierManifest(persistedID, manifest.ID)),
		orgID,
		nullablePostgresCarrierManifestText(manifest.ID),
		nullablePostgresCarrierManifestText(orgID),
		manifest.ID,
		warehouse.ID,
		nullablePostgresCarrierManifestText(manifest.WarehouseID),
		nullablePostgresCarrierManifestText(firstNonBlankPostgresCarrierManifest(manifest.WarehouseCode, warehouse.Code)),
		carrier.ID,
		nullablePostgresCarrierManifestText(carrier.ID),
		nullablePostgresCarrierManifestText(firstNonBlankPostgresCarrierManifest(manifest.CarrierCode, carrier.Code)),
		nullablePostgresCarrierManifestText(firstNonBlankPostgresCarrierManifest(manifest.CarrierName, carrier.Name)),
		manifest.Date,
		nullablePostgresCarrierManifestText(firstNonBlankPostgresCarrierManifest(manifest.HandoverBatch, "day")),
		nullablePostgresCarrierManifestText(firstNonBlankPostgresCarrierManifest(manifest.StagingZone, "handover")),
		nullablePostgresCarrierManifestUUID(manifest.HandoverZoneID),
		nullablePostgresCarrierManifestText(manifest.HandoverZoneCode),
		nullablePostgresCarrierManifestUUID(manifest.HandoverBinID),
		nullablePostgresCarrierManifestText(manifest.HandoverBinCode),
		string(manifest.Status),
		summary.ExpectedCount,
		summary.ScannedCount,
		summary.MissingCount,
		completedAt,
		nullablePostgresCarrierManifestUUID(manifest.Owner),
		nullablePostgresCarrierManifestText(manifest.Owner),
		handedOverAt,
		nullablePostgresCarrierManifestUUID(manifest.Owner),
		nullablePostgresCarrierManifestText(manifest.Owner),
		nullablePostgresCarrierManifestText(manifest.Owner),
		createdAt,
		nullablePostgresCarrierManifestUUID(manifest.Owner),
		nullablePostgresCarrierManifestText(manifest.Owner),
		time.Now().UTC(),
		nullablePostgresCarrierManifestUUID(manifest.Owner),
		nullablePostgresCarrierManifestText(manifest.Owner),
	).Scan(&savedID)
	if err != nil {
		return "", fmt.Errorf("upsert carrier manifest %q: %w", manifest.ID, err)
	}

	return savedID, nil
}

func replacePostgresCarrierManifestOrders(
	ctx context.Context,
	queryer postgresCarrierManifestQueryer,
	orgID string,
	persistedID string,
	manifest domain.CarrierManifest,
) error {
	if _, err := queryer.ExecContext(ctx, deleteCarrierManifestOrdersSQL, persistedID); err != nil {
		return fmt.Errorf("delete carrier manifest orders: %w", err)
	}
	now := time.Now().UTC()
	for index, line := range manifest.Lines {
		orderNo := firstNonBlankPostgresCarrierManifest(line.OrderNo, line.ShipmentID, line.TrackingNo, line.PackageCode)
		if strings.TrimSpace(orderNo) == "" {
			return domain.ErrManifestRequiredField
		}
		scanStatus := "pending"
		var scannedAt any
		if line.Scanned {
			scanStatus = "scanned"
			scannedAt = now
		}
		if _, err := queryer.ExecContext(
			ctx,
			insertCarrierManifestOrderSQL,
			nullablePostgresCarrierManifestUUID(line.ID),
			orgID,
			persistedID,
			nullablePostgresCarrierManifestText(line.ID),
			nullablePostgresCarrierManifestText(manifest.ID),
			index+1,
			nullablePostgresCarrierManifestUUID(line.ShipmentID),
			nullablePostgresCarrierManifestText(line.ShipmentID),
			nullablePostgresCarrierManifestUUID(line.OrderNo),
			nullablePostgresCarrierManifestText(line.OrderNo),
			orderNo,
			nullablePostgresCarrierManifestText(line.TrackingNo),
			nullablePostgresCarrierManifestText(line.PackageCode),
			nullablePostgresCarrierManifestText(line.StagingZone),
			nullablePostgresCarrierManifestUUID(line.HandoverZoneID),
			nullablePostgresCarrierManifestText(line.HandoverZoneCode),
			nullablePostgresCarrierManifestUUID(line.HandoverBinID),
			nullablePostgresCarrierManifestText(line.HandoverBinCode),
			scanStatus,
			scannedAt,
			nil,
			nil,
			now,
			nil,
			nil,
			now,
			nil,
			nil,
		); err != nil {
			return fmt.Errorf("insert carrier manifest order %q: %w", line.ID, err)
		}
	}

	return nil
}

func scanPostgresPackedShipment(row interface{ Scan(dest ...any) error }) (domain.PackedShipment, error) {
	var shipment domain.PackedShipment
	var status string
	if err := row.Scan(
		&shipment.ID,
		&shipment.OrderNo,
		&shipment.TrackingNo,
		&shipment.CarrierCode,
		&shipment.CarrierName,
		&shipment.WarehouseID,
		&shipment.WarehouseCode,
		&shipment.PackageCode,
		&status,
	); err != nil {
		return domain.PackedShipment{}, err
	}
	shipment.CarrierCode = strings.ToUpper(strings.TrimSpace(shipment.CarrierCode))
	shipment.WarehouseCode = strings.ToUpper(strings.TrimSpace(shipment.WarehouseCode))
	shipment.PackageCode = firstNonBlankPostgresCarrierManifest(shipment.PackageCode, shipment.TrackingNo, shipment.ID)
	shipment.StagingZone = "handover"
	shipment.HandoverZoneCode = "HANDOVER"
	shipment.Packed = postgresShipmentIsPacked(status)

	return shipment, nil
}

func postgresShipmentIsPacked(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "packed", "ready_for_handover", "handed_over", "delivered":
		return true
	default:
		return false
	}
}

func carrierManifestMatchesFilter(manifest domain.CarrierManifest, filter domain.CarrierManifestFilter) bool {
	if strings.TrimSpace(filter.WarehouseID) != "" && manifest.WarehouseID != strings.TrimSpace(filter.WarehouseID) {
		return false
	}
	if strings.TrimSpace(filter.Date) != "" && manifest.Date != strings.TrimSpace(filter.Date) {
		return false
	}
	if strings.TrimSpace(filter.CarrierCode) != "" &&
		manifest.CarrierCode != strings.ToUpper(strings.TrimSpace(filter.CarrierCode)) {
		return false
	}
	if status := domain.NormalizeManifestStatus(filter.Status); status != "" && manifest.Status != status {
		return false
	}

	return true
}

func carrierManifestScanEventMetadata(event CarrierManifestScanEvent) (string, error) {
	metadata := map[string]any{
		"severity":             strings.TrimSpace(event.Severity),
		"message":              strings.TrimSpace(event.Message),
		"order_no":             strings.TrimSpace(event.OrderNo),
		"tracking_no":          strings.TrimSpace(event.TrackingNo),
		"station_id":           strings.TrimSpace(event.StationID),
		"device_id":            strings.TrimSpace(event.DeviceID),
		"source":               strings.TrimSpace(event.Source),
		"expected_manifest_id": strings.TrimSpace(event.ExpectedManifestID),
	}
	data, err := json.Marshal(metadata)
	if err != nil {
		return "", fmt.Errorf("encode carrier manifest scan event metadata: %w", err)
	}

	return string(data), nil
}

func scanPostgresCarrierManifestScanEvent(
	row interface{ Scan(dest ...any) error },
) (CarrierManifestScanEvent, error) {
	var event CarrierManifestScanEvent
	var resultCode string
	if err := row.Scan(
		&event.ID,
		&event.ManifestID,
		&event.ExpectedManifestID,
		&event.Code,
		&resultCode,
		&event.Severity,
		&event.Message,
		&event.ShipmentID,
		&event.OrderNo,
		&event.TrackingNo,
		&event.ActorID,
		&event.StationID,
		&event.DeviceID,
		&event.Source,
		&event.WarehouseID,
		&event.CarrierCode,
		&event.CreatedAt,
	); err != nil {
		return CarrierManifestScanEvent{}, err
	}
	event.ResultCode = postgresCarrierManifestScanResultCode(resultCode)

	return event, nil
}

func postgresCarrierManifestScanResult(result domain.CarrierManifestScanResultCode) string {
	switch result {
	case domain.ScanResultMatched:
		return "matched"
	case domain.ScanResultManifestMismatch:
		return "wrong_manifest"
	case domain.ScanResultDuplicate:
		return "duplicate"
	case domain.ScanResultInvalidState:
		return "invalid_state"
	case domain.ScanResultNotFound:
		return "not_found"
	default:
		return "error"
	}
}

func postgresCarrierManifestScanResultCode(value string) domain.CarrierManifestScanResultCode {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "matched":
		return domain.ScanResultMatched
	case "wrong_manifest":
		return domain.ScanResultManifestMismatch
	case "duplicate":
		return domain.ScanResultDuplicate
	case "invalid_state":
		return domain.ScanResultInvalidState
	case "not_found":
		return domain.ScanResultNotFound
	default:
		return domain.ScanResultNotFound
	}
}

func nullablePostgresCarrierManifestText(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return value
}

func nullablePostgresCarrierManifestUUID(value string) any {
	value = strings.TrimSpace(value)
	if !isPostgresCarrierManifestUUIDText(value) {
		return nil
	}

	return value
}

func postgresCarrierManifestTime(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}

	return value.UTC()
}

func nullablePostgresCarrierManifestStatusTime(
	status domain.CarrierManifestStatus,
	target domain.CarrierManifestStatus,
	fallback time.Time,
) any {
	if status != target {
		return nil
	}
	if fallback.IsZero() {
		return time.Now().UTC()
	}

	return fallback.UTC()
}

func firstNonBlankPostgresCarrierManifest(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}

	return ""
}

func isPostgresCarrierManifestUUIDText(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) != 36 {
		return false
	}
	for index, char := range value {
		switch index {
		case 8, 13, 18, 23:
			if char != '-' {
				return false
			}
		default:
			if !isPostgresCarrierManifestHexText(char) {
				return false
			}
		}
	}

	return true
}

func isPostgresCarrierManifestHexText(char rune) bool {
	return (char >= '0' && char <= '9') ||
		(char >= 'a' && char <= 'f') ||
		(char >= 'A' && char <= 'F')
}

var _ CarrierManifestStore = PostgresCarrierManifestStore{}
