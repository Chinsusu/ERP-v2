package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

type PostgresWarehouseLocationCatalogConfig struct {
	DefaultOrgID string
	Clock        func() time.Time
}

type PostgresWarehouseLocationCatalog struct {
	db           *sql.DB
	auditLog     audit.LogStore
	defaultOrgID string
	clock        func() time.Time
}

func NewPostgresWarehouseLocationCatalog(
	db *sql.DB,
	auditLog audit.LogStore,
	cfg PostgresWarehouseLocationCatalogConfig,
) *PostgresWarehouseLocationCatalog {
	clock := cfg.Clock
	if clock == nil {
		clock = func() time.Time { return time.Now().UTC() }
	}

	return &PostgresWarehouseLocationCatalog{
		db:           db,
		auditLog:     auditLog,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
		clock:        clock,
	}
}

const selectPostgresWarehousesSQL = `
SELECT
  COALESCE(warehouse.warehouse_ref, warehouse.id::text),
  warehouse.code,
  warehouse.name,
  COALESCE(warehouse.warehouse_type, 'finished_good'),
  COALESCE(warehouse.site_code, ''),
  COALESCE(warehouse.address, ''),
  warehouse.allow_sale_issue,
  warehouse.allow_prod_issue,
  warehouse.allow_quarantine,
  warehouse.status,
  warehouse.created_at,
  warehouse.updated_at
FROM mdm.warehouses AS warehouse
ORDER BY warehouse.status, warehouse.site_code, warehouse.code`

const selectPostgresWarehouseSQL = `
SELECT
  COALESCE(warehouse.warehouse_ref, warehouse.id::text),
  warehouse.code,
  warehouse.name,
  COALESCE(warehouse.warehouse_type, 'finished_good'),
  COALESCE(warehouse.site_code, ''),
  COALESCE(warehouse.address, ''),
  warehouse.allow_sale_issue,
  warehouse.allow_prod_issue,
  warehouse.allow_quarantine,
  warehouse.status,
  warehouse.created_at,
  warehouse.updated_at
FROM mdm.warehouses AS warehouse
WHERE lower(COALESCE(warehouse.warehouse_ref, warehouse.id::text)) = lower($1)
   OR warehouse.id::text = $1
   OR lower(warehouse.code) = lower($1)
LIMIT 1`

const insertPostgresWarehouseSQL = `
INSERT INTO mdm.warehouses (
  id,
  org_id,
  warehouse_ref,
  code,
  name,
  warehouse_type,
  site_code,
  address,
  allow_sale_issue,
  allow_prod_issue,
  allow_quarantine,
  status,
  created_at,
  updated_at
) VALUES (
  COALESCE($1::uuid, gen_random_uuid()),
  $2::uuid,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11,
  $12,
  $13,
  $14
)`

const updatePostgresWarehouseSQL = `
UPDATE mdm.warehouses
SET code = $3,
    name = $4,
    warehouse_type = $5,
    site_code = $6,
    address = $7,
    allow_sale_issue = $8,
    allow_prod_issue = $9,
    allow_quarantine = $10,
    status = $11,
    updated_at = $12,
    version = version + 1
WHERE org_id = $1::uuid
  AND lower(COALESCE(warehouse_ref, id::text)) = lower($2)`

const updatePostgresWarehouseStatusSQL = `
UPDATE mdm.warehouses
SET status = $3,
    updated_at = $4,
    version = version + 1
WHERE org_id = $1::uuid
  AND lower(COALESCE(warehouse_ref, id::text)) = lower($2)`

const selectPostgresWarehousePersistedSQL = `
SELECT id::text, COALESCE(warehouse_ref, id::text), code
FROM mdm.warehouses
WHERE org_id = $1::uuid
  AND (
    lower(COALESCE(warehouse_ref, id::text)) = lower($2)
    OR id::text = $2
    OR lower(code) = lower($2)
  )
LIMIT 1`

const selectPostgresWarehouseDuplicateCodeSQL = `
SELECT COALESCE(warehouse_ref, id::text)
FROM mdm.warehouses
WHERE org_id = $1::uuid
  AND lower(code) = lower($2)
  AND lower(COALESCE(warehouse_ref, id::text)) <> lower($3)
LIMIT 1`

const selectPostgresLocationsSQL = `
SELECT
  COALESCE(bin.location_ref, bin.id::text),
  COALESCE(warehouse.warehouse_ref, warehouse.id::text),
  warehouse.code,
  bin.code,
  COALESCE(bin.name, bin.code),
  bin.bin_type,
  COALESCE(bin.zone_code, bin.code),
  bin.allow_receive,
  bin.allow_pick,
  bin.allow_store,
  bin.is_default,
  CASE WHEN bin.status = 'blocked' THEN 'inactive' ELSE bin.status END,
  bin.created_at,
  bin.updated_at
FROM mdm.warehouse_bins AS bin
JOIN mdm.warehouses AS warehouse ON warehouse.id = bin.warehouse_id
ORDER BY bin.status, warehouse.code, bin.zone_code, bin.code`

const selectPostgresLocationSQL = `
SELECT
  COALESCE(bin.location_ref, bin.id::text),
  COALESCE(warehouse.warehouse_ref, warehouse.id::text),
  warehouse.code,
  bin.code,
  COALESCE(bin.name, bin.code),
  bin.bin_type,
  COALESCE(bin.zone_code, bin.code),
  bin.allow_receive,
  bin.allow_pick,
  bin.allow_store,
  bin.is_default,
  CASE WHEN bin.status = 'blocked' THEN 'inactive' ELSE bin.status END,
  bin.created_at,
  bin.updated_at
FROM mdm.warehouse_bins AS bin
JOIN mdm.warehouses AS warehouse ON warehouse.id = bin.warehouse_id
WHERE lower(COALESCE(bin.location_ref, bin.id::text)) = lower($1)
   OR bin.id::text = $1
   OR lower(bin.code) = lower($1)
LIMIT 1`

const insertPostgresLocationSQL = `
INSERT INTO mdm.warehouse_bins (
  id,
  org_id,
  warehouse_id,
  location_ref,
  code,
  name,
  bin_type,
  zone_code,
  allow_receive,
  allow_pick,
  allow_store,
  is_default,
  status,
  created_at,
  updated_at
) VALUES (
  COALESCE($1::uuid, gen_random_uuid()),
  $2::uuid,
  $3::uuid,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11,
  $12,
  $13,
  $14,
  $15
)`

const updatePostgresLocationSQL = `
UPDATE mdm.warehouse_bins
SET warehouse_id = $3::uuid,
    code = $4,
    name = $5,
    bin_type = $6,
    zone_code = $7,
    allow_receive = $8,
    allow_pick = $9,
    allow_store = $10,
    is_default = $11,
    status = $12,
    updated_at = $13,
    version = version + 1
WHERE org_id = $1::uuid
  AND lower(COALESCE(location_ref, id::text)) = lower($2)`

const updatePostgresLocationStatusSQL = `
UPDATE mdm.warehouse_bins
SET status = $3,
    updated_at = $4,
    version = version + 1
WHERE org_id = $1::uuid
  AND lower(COALESCE(location_ref, id::text)) = lower($2)`

const selectPostgresLocationDuplicateCodeSQL = `
SELECT COALESCE(location_ref, id::text)
FROM mdm.warehouse_bins
WHERE org_id = $1::uuid
  AND warehouse_id = $2::uuid
  AND lower(code) = lower($3)
  AND lower(COALESCE(location_ref, id::text)) <> lower($4)
LIMIT 1`

func (s *PostgresWarehouseLocationCatalog) ListWarehouses(ctx context.Context, filter domain.WarehouseFilter) ([]domain.Warehouse, response.Pagination, error) {
	if s == nil || s.db == nil {
		return nil, response.Pagination{}, errors.New("database connection is required")
	}
	if filter.Status != "" && !domain.IsValidWarehouseStatus(filter.Status) {
		return nil, response.Pagination{}, domain.ErrWarehouseInvalidStatus
	}
	if filter.Type != "" && !domain.IsValidWarehouseType(filter.Type) {
		return nil, response.Pagination{}, domain.ErrWarehouseInvalidType
	}

	rows, err := s.db.QueryContext(ctx, selectPostgresWarehousesSQL)
	if err != nil {
		return nil, response.Pagination{}, err
	}
	defer rows.Close()

	warehouses := make([]domain.Warehouse, 0)
	for rows.Next() {
		warehouse, err := scanPostgresWarehouse(rows)
		if err != nil {
			return nil, response.Pagination{}, err
		}
		if filter.Matches(warehouse) {
			warehouses = append(warehouses, warehouse)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, response.Pagination{}, err
	}
	domain.SortWarehouses(warehouses)
	pageRows, pagination := paginateWarehouses(warehouses, filter.Page, filter.PageSize)

	return pageRows, pagination, nil
}

func (s *PostgresWarehouseLocationCatalog) GetWarehouse(ctx context.Context, id string) (domain.Warehouse, error) {
	if s == nil || s.db == nil {
		return domain.Warehouse{}, errors.New("database connection is required")
	}

	warehouse, err := scanPostgresWarehouse(s.db.QueryRowContext(ctx, selectPostgresWarehouseSQL, strings.TrimSpace(id)))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Warehouse{}, ErrWarehouseNotFound
	}
	if err != nil {
		return domain.Warehouse{}, err
	}

	return warehouse, nil
}

func (s *PostgresWarehouseLocationCatalog) CreateWarehouse(ctx context.Context, input CreateWarehouseInput) (WarehouseResult, error) {
	if s == nil || s.db == nil {
		return WarehouseResult{}, errors.New("database connection is required")
	}
	if s.auditLog == nil {
		return WarehouseResult{}, errors.New("audit log store is required")
	}

	now := s.clock().UTC()
	warehouse, err := domain.NewWarehouse(domain.NewWarehouseInput{
		ID:              newWarehouseID(input.Code, now),
		Code:            input.Code,
		Name:            input.Name,
		Type:            domain.WarehouseType(input.Type),
		SiteCode:        input.SiteCode,
		Address:         input.Address,
		AllowSaleIssue:  input.AllowSaleIssue,
		AllowProdIssue:  input.AllowProdIssue,
		AllowQuarantine: input.AllowQuarantine,
		Status:          domain.WarehouseStatus(input.Status),
		CreatedAt:       now,
		UpdatedAt:       now,
	})
	if err != nil {
		return WarehouseResult{}, err
	}
	if err := s.insertWarehouse(ctx, warehouse); err != nil {
		return WarehouseResult{}, err
	}
	log, err := newWarehouseAuditLog(input.ActorID, input.RequestID, "masterdata.warehouse.created", warehouse, nil, warehouseToAuditMap(warehouse), now)
	if err != nil {
		return WarehouseResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return WarehouseResult{}, err
	}

	return WarehouseResult{Warehouse: warehouse, AuditLogID: log.ID}, nil
}

func (s *PostgresWarehouseLocationCatalog) UpdateWarehouse(ctx context.Context, input UpdateWarehouseInput) (WarehouseResult, error) {
	if s == nil || s.db == nil {
		return WarehouseResult{}, errors.New("database connection is required")
	}
	if s.auditLog == nil {
		return WarehouseResult{}, errors.New("audit log store is required")
	}

	current, err := s.GetWarehouse(ctx, input.ID)
	if err != nil {
		return WarehouseResult{}, err
	}
	now := s.clock().UTC()
	updated, err := current.Update(domain.UpdateWarehouseInput{
		Code:            input.Code,
		Name:            input.Name,
		Type:            domain.WarehouseType(input.Type),
		SiteCode:        input.SiteCode,
		Address:         input.Address,
		AllowSaleIssue:  input.AllowSaleIssue,
		AllowProdIssue:  input.AllowProdIssue,
		AllowQuarantine: input.AllowQuarantine,
		Status:          domain.WarehouseStatus(input.Status),
		UpdatedAt:       now,
	})
	if err != nil {
		return WarehouseResult{}, err
	}
	if err := s.updateWarehouse(ctx, updated); err != nil {
		return WarehouseResult{}, err
	}
	log, err := newWarehouseAuditLog(input.ActorID, input.RequestID, "masterdata.warehouse.updated", updated, warehouseToAuditMap(current), warehouseToAuditMap(updated), now)
	if err != nil {
		return WarehouseResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return WarehouseResult{}, err
	}

	return WarehouseResult{Warehouse: updated, AuditLogID: log.ID}, nil
}

func (s *PostgresWarehouseLocationCatalog) ChangeWarehouseStatus(ctx context.Context, input ChangeWarehouseStatusInput) (WarehouseResult, error) {
	if s == nil || s.db == nil {
		return WarehouseResult{}, errors.New("database connection is required")
	}
	if s.auditLog == nil {
		return WarehouseResult{}, errors.New("audit log store is required")
	}
	current, err := s.GetWarehouse(ctx, input.ID)
	if err != nil {
		return WarehouseResult{}, err
	}
	now := s.clock().UTC()
	updated, err := current.ChangeStatus(domain.WarehouseStatus(input.Status), now)
	if err != nil {
		return WarehouseResult{}, err
	}
	orgID, err := s.resolveOrgID()
	if err != nil {
		return WarehouseResult{}, err
	}
	result, err := s.db.ExecContext(ctx, updatePostgresWarehouseStatusSQL, orgID, updated.ID, string(updated.Status), updated.UpdatedAt)
	if err != nil {
		return WarehouseResult{}, fmt.Errorf("update warehouse status %q: %w", updated.ID, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return WarehouseResult{}, err
	}
	if affected == 0 {
		return WarehouseResult{}, ErrWarehouseNotFound
	}
	log, err := newWarehouseAuditLog(input.ActorID, input.RequestID, "masterdata.warehouse.status_changed", updated, map[string]any{"status": string(current.Status)}, map[string]any{"status": string(updated.Status)}, now)
	if err != nil {
		return WarehouseResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return WarehouseResult{}, err
	}

	return WarehouseResult{Warehouse: updated, AuditLogID: log.ID}, nil
}

func (s *PostgresWarehouseLocationCatalog) ListLocations(ctx context.Context, filter domain.LocationFilter) ([]domain.Location, response.Pagination, error) {
	if s == nil || s.db == nil {
		return nil, response.Pagination{}, errors.New("database connection is required")
	}
	if filter.Status != "" && !domain.IsValidLocationStatus(filter.Status) {
		return nil, response.Pagination{}, domain.ErrLocationInvalidStatus
	}
	if filter.Type != "" && !domain.IsValidLocationType(filter.Type) {
		return nil, response.Pagination{}, domain.ErrLocationInvalidType
	}

	rows, err := s.db.QueryContext(ctx, selectPostgresLocationsSQL)
	if err != nil {
		return nil, response.Pagination{}, err
	}
	defer rows.Close()

	locations := make([]domain.Location, 0)
	for rows.Next() {
		location, err := scanPostgresLocation(rows)
		if err != nil {
			return nil, response.Pagination{}, err
		}
		if filter.Matches(location) {
			locations = append(locations, location)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, response.Pagination{}, err
	}
	domain.SortLocations(locations)
	pageRows, pagination := paginateLocations(locations, filter.Page, filter.PageSize)

	return pageRows, pagination, nil
}

func (s *PostgresWarehouseLocationCatalog) GetLocation(ctx context.Context, id string) (domain.Location, error) {
	if s == nil || s.db == nil {
		return domain.Location{}, errors.New("database connection is required")
	}
	location, err := scanPostgresLocation(s.db.QueryRowContext(ctx, selectPostgresLocationSQL, strings.TrimSpace(id)))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Location{}, ErrLocationNotFound
	}
	if err != nil {
		return domain.Location{}, err
	}

	return location, nil
}

func (s *PostgresWarehouseLocationCatalog) CreateLocation(ctx context.Context, input CreateLocationInput) (LocationResult, error) {
	if s == nil || s.db == nil {
		return LocationResult{}, errors.New("database connection is required")
	}
	if s.auditLog == nil {
		return LocationResult{}, errors.New("audit log store is required")
	}
	now := s.clock().UTC()
	orgID, err := s.resolveOrgID()
	if err != nil {
		return LocationResult{}, err
	}
	persistedWarehouseID, warehouse, err := s.findPersistedWarehouse(ctx, orgID, input.WarehouseID)
	if err != nil {
		return LocationResult{}, err
	}
	location, err := domain.NewLocation(domain.NewLocationInput{
		ID:            newLocationID(warehouse.Code, input.Code, now),
		WarehouseID:   warehouse.ID,
		WarehouseCode: warehouse.Code,
		Code:          input.Code,
		Name:          input.Name,
		Type:          domain.LocationType(input.Type),
		ZoneCode:      input.ZoneCode,
		AllowReceive:  input.AllowReceive,
		AllowPick:     input.AllowPick,
		AllowStore:    input.AllowStore,
		IsDefault:     input.IsDefault,
		Status:        domain.LocationStatus(input.Status),
		CreatedAt:     now,
		UpdatedAt:     now,
	})
	if err != nil {
		return LocationResult{}, err
	}
	if err := s.insertLocation(ctx, orgID, persistedWarehouseID, location); err != nil {
		return LocationResult{}, err
	}
	log, err := newLocationAuditLog(input.ActorID, input.RequestID, "masterdata.location.created", location, nil, locationToAuditMap(location), now)
	if err != nil {
		return LocationResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return LocationResult{}, err
	}

	return LocationResult{Location: location, AuditLogID: log.ID}, nil
}

func (s *PostgresWarehouseLocationCatalog) UpdateLocation(ctx context.Context, input UpdateLocationInput) (LocationResult, error) {
	if s == nil || s.db == nil {
		return LocationResult{}, errors.New("database connection is required")
	}
	if s.auditLog == nil {
		return LocationResult{}, errors.New("audit log store is required")
	}
	current, err := s.GetLocation(ctx, input.ID)
	if err != nil {
		return LocationResult{}, err
	}
	if current.Status != domain.LocationStatusActive && domain.NormalizeLocationStatus(domain.LocationStatus(input.Status)) != domain.LocationStatusActive {
		return LocationResult{}, ErrInactiveLocation
	}
	now := s.clock().UTC()
	orgID, err := s.resolveOrgID()
	if err != nil {
		return LocationResult{}, err
	}
	persistedWarehouseID, warehouse, err := s.findPersistedWarehouse(ctx, orgID, input.WarehouseID)
	if err != nil {
		return LocationResult{}, err
	}
	updated, err := current.Update(domain.UpdateLocationInput{
		WarehouseID:   warehouse.ID,
		WarehouseCode: warehouse.Code,
		Code:          input.Code,
		Name:          input.Name,
		Type:          domain.LocationType(input.Type),
		ZoneCode:      input.ZoneCode,
		AllowReceive:  input.AllowReceive,
		AllowPick:     input.AllowPick,
		AllowStore:    input.AllowStore,
		IsDefault:     input.IsDefault,
		Status:        domain.LocationStatus(input.Status),
		UpdatedAt:     now,
	})
	if err != nil {
		return LocationResult{}, err
	}
	if err := s.updateLocation(ctx, orgID, persistedWarehouseID, updated); err != nil {
		return LocationResult{}, err
	}
	log, err := newLocationAuditLog(input.ActorID, input.RequestID, "masterdata.location.updated", updated, locationToAuditMap(current), locationToAuditMap(updated), now)
	if err != nil {
		return LocationResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return LocationResult{}, err
	}

	return LocationResult{Location: updated, AuditLogID: log.ID}, nil
}

func (s *PostgresWarehouseLocationCatalog) ChangeLocationStatus(ctx context.Context, input ChangeLocationStatusInput) (LocationResult, error) {
	if s == nil || s.db == nil {
		return LocationResult{}, errors.New("database connection is required")
	}
	if s.auditLog == nil {
		return LocationResult{}, errors.New("audit log store is required")
	}
	current, err := s.GetLocation(ctx, input.ID)
	if err != nil {
		return LocationResult{}, err
	}
	now := s.clock().UTC()
	updated, err := current.ChangeStatus(domain.LocationStatus(input.Status), now)
	if err != nil {
		return LocationResult{}, err
	}
	orgID, err := s.resolveOrgID()
	if err != nil {
		return LocationResult{}, err
	}
	result, err := s.db.ExecContext(ctx, updatePostgresLocationStatusSQL, orgID, updated.ID, string(updated.Status), updated.UpdatedAt)
	if err != nil {
		return LocationResult{}, fmt.Errorf("update location status %q: %w", updated.ID, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return LocationResult{}, err
	}
	if affected == 0 {
		return LocationResult{}, ErrLocationNotFound
	}
	log, err := newLocationAuditLog(input.ActorID, input.RequestID, "masterdata.location.status_changed", updated, map[string]any{"status": string(current.Status)}, map[string]any{"status": string(updated.Status)}, now)
	if err != nil {
		return LocationResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return LocationResult{}, err
	}

	return LocationResult{Location: updated, AuditLogID: log.ID}, nil
}

func (s *PostgresWarehouseLocationCatalog) insertWarehouse(ctx context.Context, warehouse domain.Warehouse) error {
	orgID, err := s.resolveOrgID()
	if err != nil {
		return err
	}
	if err := s.ensureUniqueWarehouse(ctx, orgID, warehouse, ""); err != nil {
		return err
	}
	_, err = s.db.ExecContext(
		ctx,
		insertPostgresWarehouseSQL,
		nullablePostgresWarehouseUUID(warehouse.ID),
		orgID,
		warehouse.ID,
		warehouse.Code,
		warehouse.Name,
		string(warehouse.Type),
		warehouse.SiteCode,
		nullablePostgresWarehouseText(warehouse.Address),
		warehouse.AllowSaleIssue,
		warehouse.AllowProdIssue,
		warehouse.AllowQuarantine,
		string(warehouse.Status),
		warehouse.CreatedAt,
		warehouse.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert warehouse %q: %w", warehouse.ID, err)
	}

	return nil
}

func (s *PostgresWarehouseLocationCatalog) updateWarehouse(ctx context.Context, warehouse domain.Warehouse) error {
	orgID, err := s.resolveOrgID()
	if err != nil {
		return err
	}
	if err := s.ensureUniqueWarehouse(ctx, orgID, warehouse, warehouse.ID); err != nil {
		return err
	}
	result, err := s.db.ExecContext(
		ctx,
		updatePostgresWarehouseSQL,
		orgID,
		warehouse.ID,
		warehouse.Code,
		warehouse.Name,
		string(warehouse.Type),
		warehouse.SiteCode,
		nullablePostgresWarehouseText(warehouse.Address),
		warehouse.AllowSaleIssue,
		warehouse.AllowProdIssue,
		warehouse.AllowQuarantine,
		string(warehouse.Status),
		warehouse.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update warehouse %q: %w", warehouse.ID, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrWarehouseNotFound
	}

	return nil
}

func (s *PostgresWarehouseLocationCatalog) insertLocation(ctx context.Context, orgID string, persistedWarehouseID string, location domain.Location) error {
	if err := s.ensureUniqueLocation(ctx, orgID, persistedWarehouseID, location, ""); err != nil {
		return err
	}
	_, err := s.db.ExecContext(
		ctx,
		insertPostgresLocationSQL,
		nullablePostgresWarehouseUUID(location.ID),
		orgID,
		persistedWarehouseID,
		location.ID,
		location.Code,
		location.Name,
		string(location.Type),
		location.ZoneCode,
		location.AllowReceive,
		location.AllowPick,
		location.AllowStore,
		location.IsDefault,
		string(location.Status),
		location.CreatedAt,
		location.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert location %q: %w", location.ID, err)
	}

	return nil
}

func (s *PostgresWarehouseLocationCatalog) updateLocation(ctx context.Context, orgID string, persistedWarehouseID string, location domain.Location) error {
	if err := s.ensureUniqueLocation(ctx, orgID, persistedWarehouseID, location, location.ID); err != nil {
		return err
	}
	result, err := s.db.ExecContext(
		ctx,
		updatePostgresLocationSQL,
		orgID,
		location.ID,
		persistedWarehouseID,
		location.Code,
		location.Name,
		string(location.Type),
		location.ZoneCode,
		location.AllowReceive,
		location.AllowPick,
		location.AllowStore,
		location.IsDefault,
		string(location.Status),
		location.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update location %q: %w", location.ID, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrLocationNotFound
	}

	return nil
}

func (s *PostgresWarehouseLocationCatalog) findPersistedWarehouse(ctx context.Context, orgID string, id string) (string, domain.Warehouse, error) {
	var persistedID string
	var warehouseRef string
	var code string
	err := s.db.QueryRowContext(ctx, selectPostgresWarehousePersistedSQL, orgID, strings.TrimSpace(id)).Scan(&persistedID, &warehouseRef, &code)
	if errors.Is(err, sql.ErrNoRows) {
		return "", domain.Warehouse{}, ErrInvalidLocationWarehouse
	}
	if err != nil {
		return "", domain.Warehouse{}, fmt.Errorf("find warehouse %q: %w", id, err)
	}
	warehouse, err := s.GetWarehouse(ctx, warehouseRef)
	if err != nil {
		return "", domain.Warehouse{}, err
	}
	warehouse.Code = code

	return persistedID, warehouse, nil
}

func (s *PostgresWarehouseLocationCatalog) ensureUniqueWarehouse(ctx context.Context, orgID string, warehouse domain.Warehouse, currentID string) error {
	var duplicate string
	err := s.db.QueryRowContext(ctx, selectPostgresWarehouseDuplicateCodeSQL, orgID, warehouse.Code, currentID).Scan(&duplicate)
	if err == nil {
		return ErrDuplicateWarehouseCode
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("check duplicate warehouse code: %w", err)
	}

	return nil
}

func (s *PostgresWarehouseLocationCatalog) ensureUniqueLocation(ctx context.Context, orgID string, persistedWarehouseID string, location domain.Location, currentID string) error {
	var duplicate string
	err := s.db.QueryRowContext(ctx, selectPostgresLocationDuplicateCodeSQL, orgID, persistedWarehouseID, location.Code, currentID).Scan(&duplicate)
	if err == nil {
		return ErrDuplicateLocationCode
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("check duplicate location code: %w", err)
	}

	return nil
}

func (s *PostgresWarehouseLocationCatalog) resolveOrgID() (string, error) {
	if isPostgresWarehouseUUIDText(s.defaultOrgID) {
		return s.defaultOrgID, nil
	}

	return "", errors.New("warehouse catalog default org id is required")
}

func scanPostgresWarehouse(scanner interface{ Scan(dest ...any) error }) (domain.Warehouse, error) {
	var (
		warehouse     domain.Warehouse
		warehouseType string
		status        string
	)
	if err := scanner.Scan(
		&warehouse.ID,
		&warehouse.Code,
		&warehouse.Name,
		&warehouseType,
		&warehouse.SiteCode,
		&warehouse.Address,
		&warehouse.AllowSaleIssue,
		&warehouse.AllowProdIssue,
		&warehouse.AllowQuarantine,
		&status,
		&warehouse.CreatedAt,
		&warehouse.UpdatedAt,
	); err != nil {
		return domain.Warehouse{}, err
	}

	return domain.NewWarehouse(domain.NewWarehouseInput{
		ID:              warehouse.ID,
		Code:            warehouse.Code,
		Name:            warehouse.Name,
		Type:            domain.WarehouseType(warehouseType),
		SiteCode:        warehouse.SiteCode,
		Address:         warehouse.Address,
		AllowSaleIssue:  warehouse.AllowSaleIssue,
		AllowProdIssue:  warehouse.AllowProdIssue,
		AllowQuarantine: warehouse.AllowQuarantine,
		Status:          domain.WarehouseStatus(status),
		CreatedAt:       warehouse.CreatedAt,
		UpdatedAt:       warehouse.UpdatedAt,
	})
}

func scanPostgresLocation(scanner interface{ Scan(dest ...any) error }) (domain.Location, error) {
	var (
		location     domain.Location
		locationType string
		status       string
	)
	if err := scanner.Scan(
		&location.ID,
		&location.WarehouseID,
		&location.WarehouseCode,
		&location.Code,
		&location.Name,
		&locationType,
		&location.ZoneCode,
		&location.AllowReceive,
		&location.AllowPick,
		&location.AllowStore,
		&location.IsDefault,
		&status,
		&location.CreatedAt,
		&location.UpdatedAt,
	); err != nil {
		return domain.Location{}, err
	}

	return domain.NewLocation(domain.NewLocationInput{
		ID:            location.ID,
		WarehouseID:   location.WarehouseID,
		WarehouseCode: location.WarehouseCode,
		Code:          location.Code,
		Name:          location.Name,
		Type:          domain.LocationType(locationType),
		ZoneCode:      location.ZoneCode,
		AllowReceive:  location.AllowReceive,
		AllowPick:     location.AllowPick,
		AllowStore:    location.AllowStore,
		IsDefault:     location.IsDefault,
		Status:        domain.LocationStatus(status),
		CreatedAt:     location.CreatedAt,
		UpdatedAt:     location.UpdatedAt,
	})
}

func nullablePostgresWarehouseText(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return value
}

func nullablePostgresWarehouseUUID(value string) any {
	value = strings.TrimSpace(value)
	if !isPostgresWarehouseUUIDText(value) {
		return nil
	}

	return value
}

func isPostgresWarehouseUUIDText(value string) bool {
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
			if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
				return false
			}
		}
	}

	return true
}
