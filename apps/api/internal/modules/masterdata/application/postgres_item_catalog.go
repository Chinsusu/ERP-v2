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
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

type PostgresItemCatalogConfig struct {
	DefaultOrgID string
	Clock        func() time.Time
}

type PostgresItemCatalog struct {
	db           *sql.DB
	auditLog     audit.LogStore
	defaultOrgID string
	clock        func() time.Time
}

type postgresItemQueryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func NewPostgresItemCatalog(db *sql.DB, auditLog audit.LogStore, cfg PostgresItemCatalogConfig) *PostgresItemCatalog {
	clock := cfg.Clock
	if clock == nil {
		clock = func() time.Time { return time.Now().UTC() }
	}

	return &PostgresItemCatalog{
		db:           db,
		auditLog:     auditLog,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
		clock:        clock,
	}
}

const selectPostgresItemsSQL = `
SELECT
  COALESCE(item.item_ref, item.id::text),
  COALESCE(item.item_code, item.sku),
  item.sku,
  item.name,
  item.item_type,
  COALESCE(item.item_group, ''),
  COALESCE(item.brand_code, ''),
  COALESCE(item.uom_base, unit.code, 'PCS'),
  COALESCE(item.uom_purchase, item.uom_base, unit.code, 'PCS'),
  COALESCE(item.uom_issue, item.uom_base, unit.code, 'PCS'),
  item.lot_controlled,
  item.expiry_controlled,
  COALESCE(item.shelf_life_days, 0),
  item.qc_required,
  CASE WHEN item.status = 'blocked' THEN 'inactive' ELSE item.status END,
  item.standard_cost::text,
  item.is_sellable,
  item.is_purchasable,
  item.is_producible,
  COALESCE(item.spec_version, ''),
  item.created_at,
  item.updated_at
FROM mdm.items AS item
LEFT JOIN mdm.units AS unit ON unit.id = item.base_unit_id
ORDER BY item.status, item.sku, item.item_code`

const selectPostgresItemSQL = `
SELECT
  COALESCE(item.item_ref, item.id::text),
  COALESCE(item.item_code, item.sku),
  item.sku,
  item.name,
  item.item_type,
  COALESCE(item.item_group, ''),
  COALESCE(item.brand_code, ''),
  COALESCE(item.uom_base, unit.code, 'PCS'),
  COALESCE(item.uom_purchase, item.uom_base, unit.code, 'PCS'),
  COALESCE(item.uom_issue, item.uom_base, unit.code, 'PCS'),
  item.lot_controlled,
  item.expiry_controlled,
  COALESCE(item.shelf_life_days, 0),
  item.qc_required,
  CASE WHEN item.status = 'blocked' THEN 'inactive' ELSE item.status END,
  item.standard_cost::text,
  item.is_sellable,
  item.is_purchasable,
  item.is_producible,
  COALESCE(item.spec_version, ''),
  item.created_at,
  item.updated_at
FROM mdm.items AS item
LEFT JOIN mdm.units AS unit ON unit.id = item.base_unit_id
WHERE lower(COALESCE(item.item_ref, item.id::text)) = lower($1)
   OR item.id::text = $1
   OR lower(item.sku) = lower($1)
   OR lower(COALESCE(item.item_code, '')) = lower($1)
LIMIT 1`

const insertPostgresItemSQL = `
INSERT INTO mdm.items (
  id,
  org_id,
  item_ref,
  item_code,
  sku,
  name,
  item_type,
  item_group,
  brand_code,
  base_unit_id,
  uom_base,
  uom_purchase,
  uom_issue,
  requires_batch,
  requires_expiry,
  lot_controlled,
  expiry_controlled,
  shelf_life_days,
  qc_required,
  status,
  standard_cost,
  is_sellable,
  is_purchasable,
  is_producible,
  spec_version,
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
  $10::uuid,
  $11,
  $12,
  $13,
  $14,
  $15,
  $16,
  $17,
  $18,
  $19,
  $20,
  $21,
  $22,
  $23,
  $24,
  $25,
  $26,
  $27
)`

const updatePostgresItemSQL = `
UPDATE mdm.items
SET item_code = $3,
    sku = $4,
    name = $5,
    item_type = $6,
    item_group = $7,
    brand_code = $8,
    base_unit_id = $9::uuid,
    uom_base = $10,
    uom_purchase = $11,
    uom_issue = $12,
    requires_batch = $13,
    requires_expiry = $14,
    lot_controlled = $15,
    expiry_controlled = $16,
    shelf_life_days = $17,
    qc_required = $18,
    status = $19,
    standard_cost = $20,
    is_sellable = $21,
    is_purchasable = $22,
    is_producible = $23,
    spec_version = $24,
    updated_at = $25,
    version = version + 1
WHERE org_id = $1::uuid
  AND lower(COALESCE(item_ref, id::text)) = lower($2)`

const updatePostgresItemStatusSQL = `
UPDATE mdm.items
SET status = $3,
    updated_at = $4,
    version = version + 1
WHERE org_id = $1::uuid
  AND lower(COALESCE(item_ref, id::text)) = lower($2)`

const selectPostgresItemPersistedSQL = `
SELECT id::text
FROM mdm.items
WHERE org_id = $1::uuid
  AND lower(COALESCE(item_ref, id::text)) = lower($2)
LIMIT 1`

const selectPostgresItemDuplicateCodeSQL = `
SELECT COALESCE(item_ref, id::text)
FROM mdm.items
WHERE org_id = $1::uuid
  AND lower(item_code) = lower($2)
  AND lower(COALESCE(item_ref, id::text)) <> lower($3)
LIMIT 1`

const selectPostgresItemDuplicateSKUSQL = `
SELECT COALESCE(item_ref, id::text)
FROM mdm.items
WHERE org_id = $1::uuid
  AND lower(sku) = lower($2)
  AND lower(COALESCE(item_ref, id::text)) <> lower($3)
LIMIT 1`

const upsertPostgresItemUnitSQL = `
INSERT INTO mdm.units (org_id, code, name, precision_scale, status)
VALUES ($1::uuid, $2, $2, $3, 'active')
ON CONFLICT (org_id, code) DO UPDATE
SET precision_scale = EXCLUDED.precision_scale,
    status = 'active',
    updated_at = now()
RETURNING id::text`

func (s *PostgresItemCatalog) List(ctx context.Context, filter domain.ItemFilter) ([]domain.Item, response.Pagination, error) {
	if s == nil || s.db == nil {
		return nil, response.Pagination{}, errors.New("database connection is required")
	}
	if filter.Status != "" && !domain.IsValidItemStatus(filter.Status) {
		return nil, response.Pagination{}, domain.ErrItemInvalidStatus
	}
	if filter.Type != "" && !domain.IsValidItemType(filter.Type) {
		return nil, response.Pagination{}, domain.ErrItemInvalidType
	}

	rows, err := s.db.QueryContext(ctx, selectPostgresItemsSQL)
	if err != nil {
		return nil, response.Pagination{}, err
	}
	defer rows.Close()

	items := make([]domain.Item, 0)
	for rows.Next() {
		item, err := scanPostgresItem(rows)
		if err != nil {
			return nil, response.Pagination{}, err
		}
		if filter.Matches(item) {
			items = append(items, item)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, response.Pagination{}, err
	}
	domain.SortItems(items)
	pageRows, pagination := paginateItems(items, filter.Page, filter.PageSize)

	return pageRows, pagination, nil
}

func (s *PostgresItemCatalog) Get(ctx context.Context, id string) (domain.Item, error) {
	if s == nil || s.db == nil {
		return domain.Item{}, errors.New("database connection is required")
	}

	item, err := scanPostgresItem(s.db.QueryRowContext(ctx, selectPostgresItemSQL, strings.TrimSpace(id)))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Item{}, ErrItemNotFound
	}
	if err != nil {
		return domain.Item{}, err
	}

	return item, nil
}

func (s *PostgresItemCatalog) Create(ctx context.Context, input CreateItemInput) (ItemResult, error) {
	if s == nil || s.db == nil {
		return ItemResult{}, errors.New("database connection is required")
	}
	if s.auditLog == nil {
		return ItemResult{}, errors.New("audit log store is required")
	}

	now := s.clock().UTC()
	item, err := domain.NewItem(domain.NewItemInput{
		ID:               newItemID(input.SKUCode, now),
		ItemCode:         input.ItemCode,
		SKUCode:          input.SKUCode,
		Name:             input.Name,
		Type:             domain.ItemType(input.Type),
		Group:            input.Group,
		BrandCode:        input.BrandCode,
		UOMBase:          input.UOMBase,
		UOMPurchase:      input.UOMPurchase,
		UOMIssue:         input.UOMIssue,
		LotControlled:    input.LotControlled,
		ExpiryControlled: input.ExpiryControlled,
		ShelfLifeDays:    input.ShelfLifeDays,
		QCRequired:       input.QCRequired,
		Status:           domain.ItemStatus(input.Status),
		StandardCost:     input.StandardCost,
		IsSellable:       input.IsSellable,
		IsPurchasable:    input.IsPurchasable,
		IsProducible:     input.IsProducible,
		SpecVersion:      input.SpecVersion,
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	if err != nil {
		return ItemResult{}, err
	}

	if err := s.saveNewItem(ctx, item); err != nil {
		return ItemResult{}, err
	}
	log, err := newItemAuditLog(input.ActorID, input.RequestID, "masterdata.item.created", item, nil, itemToAuditMap(item), now)
	if err != nil {
		return ItemResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return ItemResult{}, err
	}

	return ItemResult{Item: item, AuditLogID: log.ID}, nil
}

func (s *PostgresItemCatalog) Update(ctx context.Context, input UpdateItemInput) (ItemResult, error) {
	if s == nil || s.db == nil {
		return ItemResult{}, errors.New("database connection is required")
	}
	if s.auditLog == nil {
		return ItemResult{}, errors.New("audit log store is required")
	}

	current, err := s.Get(ctx, input.ID)
	if err != nil {
		return ItemResult{}, err
	}
	now := s.clock().UTC()
	updated, err := current.Update(domain.UpdateItemInput{
		ItemCode:         input.ItemCode,
		SKUCode:          input.SKUCode,
		Name:             input.Name,
		Type:             domain.ItemType(input.Type),
		Group:            input.Group,
		BrandCode:        input.BrandCode,
		UOMBase:          input.UOMBase,
		UOMPurchase:      input.UOMPurchase,
		UOMIssue:         input.UOMIssue,
		LotControlled:    input.LotControlled,
		ExpiryControlled: input.ExpiryControlled,
		ShelfLifeDays:    input.ShelfLifeDays,
		QCRequired:       input.QCRequired,
		Status:           domain.ItemStatus(input.Status),
		StandardCost:     input.StandardCost,
		IsSellable:       input.IsSellable,
		IsPurchasable:    input.IsPurchasable,
		IsProducible:     input.IsProducible,
		SpecVersion:      input.SpecVersion,
		UpdatedAt:        now,
	})
	if err != nil {
		return ItemResult{}, err
	}
	if err := s.saveExistingItem(ctx, updated); err != nil {
		return ItemResult{}, err
	}
	log, err := newItemAuditLog(input.ActorID, input.RequestID, "masterdata.item.updated", updated, itemToAuditMap(current), itemToAuditMap(updated), now)
	if err != nil {
		return ItemResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return ItemResult{}, err
	}

	return ItemResult{Item: updated, AuditLogID: log.ID}, nil
}

func (s *PostgresItemCatalog) ChangeStatus(ctx context.Context, input ChangeItemStatusInput) (ItemResult, error) {
	if s == nil || s.db == nil {
		return ItemResult{}, errors.New("database connection is required")
	}
	if s.auditLog == nil {
		return ItemResult{}, errors.New("audit log store is required")
	}

	current, err := s.Get(ctx, input.ID)
	if err != nil {
		return ItemResult{}, err
	}
	now := s.clock().UTC()
	updated, err := current.ChangeStatus(domain.ItemStatus(input.Status), now)
	if err != nil {
		return ItemResult{}, err
	}
	orgID, err := s.resolveOrgID()
	if err != nil {
		return ItemResult{}, err
	}
	result, err := s.db.ExecContext(ctx, updatePostgresItemStatusSQL, orgID, updated.ID, string(updated.Status), updated.UpdatedAt)
	if err != nil {
		return ItemResult{}, fmt.Errorf("update item status %q: %w", updated.ID, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return ItemResult{}, err
	}
	if affected == 0 {
		return ItemResult{}, ErrItemNotFound
	}

	log, err := newItemAuditLog(
		input.ActorID,
		input.RequestID,
		"masterdata.item.status_changed",
		updated,
		map[string]any{"status": string(current.Status)},
		map[string]any{"status": string(updated.Status)},
		now,
	)
	if err != nil {
		return ItemResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return ItemResult{}, err
	}

	return ItemResult{Item: updated, AuditLogID: log.ID}, nil
}

func (s *PostgresItemCatalog) saveNewItem(ctx context.Context, item domain.Item) error {
	orgID, err := s.resolveOrgID()
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin item transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	if err := ensurePostgresItemUnique(ctx, tx, orgID, item, ""); err != nil {
		return err
	}
	unitID, err := ensurePostgresItemUnit(ctx, tx, orgID, item.UOMBase)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(
		ctx,
		insertPostgresItemSQL,
		nullablePostgresItemUUID(item.ID),
		orgID,
		item.ID,
		item.ItemCode,
		item.SKUCode,
		item.Name,
		string(item.Type),
		nullablePostgresItemText(item.Group),
		nullablePostgresItemText(item.BrandCode),
		unitID,
		item.UOMBase,
		item.UOMPurchase,
		item.UOMIssue,
		item.LotControlled,
		item.ExpiryControlled,
		item.LotControlled,
		item.ExpiryControlled,
		nullablePostgresItemInt(item.ShelfLifeDays),
		item.QCRequired,
		string(item.Status),
		item.StandardCost.String(),
		item.IsSellable,
		item.IsPurchasable,
		item.IsProducible,
		nullablePostgresItemText(item.SpecVersion),
		item.CreatedAt,
		item.UpdatedAt,
	); err != nil {
		return fmt.Errorf("insert item %q: %w", item.ID, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit item transaction: %w", err)
	}
	committed = true

	return nil
}

func (s *PostgresItemCatalog) saveExistingItem(ctx context.Context, item domain.Item) error {
	orgID, err := s.resolveOrgID()
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin item transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	if err := ensurePostgresItemUnique(ctx, tx, orgID, item, item.ID); err != nil {
		return err
	}
	unitID, err := ensurePostgresItemUnit(ctx, tx, orgID, item.UOMBase)
	if err != nil {
		return err
	}
	result, err := tx.ExecContext(
		ctx,
		updatePostgresItemSQL,
		orgID,
		item.ID,
		item.ItemCode,
		item.SKUCode,
		item.Name,
		string(item.Type),
		nullablePostgresItemText(item.Group),
		nullablePostgresItemText(item.BrandCode),
		unitID,
		item.UOMBase,
		item.UOMPurchase,
		item.UOMIssue,
		item.LotControlled,
		item.ExpiryControlled,
		item.LotControlled,
		item.ExpiryControlled,
		nullablePostgresItemInt(item.ShelfLifeDays),
		item.QCRequired,
		string(item.Status),
		item.StandardCost.String(),
		item.IsSellable,
		item.IsPurchasable,
		item.IsProducible,
		nullablePostgresItemText(item.SpecVersion),
		item.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update item %q: %w", item.ID, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrItemNotFound
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit item transaction: %w", err)
	}
	committed = true

	return nil
}

func (s *PostgresItemCatalog) resolveOrgID() (string, error) {
	if isPostgresItemUUIDText(s.defaultOrgID) {
		return s.defaultOrgID, nil
	}

	return "", errors.New("item catalog default org id is required")
}

func scanPostgresItem(scanner interface{ Scan(dest ...any) error }) (domain.Item, error) {
	var (
		item         domain.Item
		itemType     string
		status       string
		standardCost string
	)
	if err := scanner.Scan(
		&item.ID,
		&item.ItemCode,
		&item.SKUCode,
		&item.Name,
		&itemType,
		&item.Group,
		&item.BrandCode,
		&item.UOMBase,
		&item.UOMPurchase,
		&item.UOMIssue,
		&item.LotControlled,
		&item.ExpiryControlled,
		&item.ShelfLifeDays,
		&item.QCRequired,
		&status,
		&standardCost,
		&item.IsSellable,
		&item.IsPurchasable,
		&item.IsProducible,
		&item.SpecVersion,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return domain.Item{}, err
	}
	cost, err := decimal.ParseUnitCost(standardCost)
	if err != nil {
		return domain.Item{}, err
	}

	return domain.NewItem(domain.NewItemInput{
		ID:               item.ID,
		ItemCode:         item.ItemCode,
		SKUCode:          item.SKUCode,
		Name:             item.Name,
		Type:             domain.ItemType(itemType),
		Group:            item.Group,
		BrandCode:        item.BrandCode,
		UOMBase:          item.UOMBase,
		UOMPurchase:      item.UOMPurchase,
		UOMIssue:         item.UOMIssue,
		LotControlled:    item.LotControlled,
		ExpiryControlled: item.ExpiryControlled,
		ShelfLifeDays:    item.ShelfLifeDays,
		QCRequired:       item.QCRequired,
		Status:           domain.ItemStatus(status),
		StandardCost:     cost,
		IsSellable:       item.IsSellable,
		IsPurchasable:    item.IsPurchasable,
		IsProducible:     item.IsProducible,
		SpecVersion:      item.SpecVersion,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
	})
}

func ensurePostgresItemUnique(ctx context.Context, queryer postgresItemQueryer, orgID string, item domain.Item, currentID string) error {
	var duplicate string
	err := queryer.QueryRowContext(ctx, selectPostgresItemDuplicateCodeSQL, orgID, item.ItemCode, currentID).Scan(&duplicate)
	if err == nil {
		return ErrDuplicateItemCode
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("check duplicate item code: %w", err)
	}
	err = queryer.QueryRowContext(ctx, selectPostgresItemDuplicateSKUSQL, orgID, item.SKUCode, currentID).Scan(&duplicate)
	if err == nil {
		return ErrDuplicateSKUCode
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("check duplicate item sku: %w", err)
	}

	return nil
}

func ensurePostgresItemUnit(ctx context.Context, queryer postgresItemQueryer, orgID string, uomCode string) (string, error) {
	uomCode = domain.NormalizeUOM(uomCode)
	if uomCode == "" {
		return "", domain.ErrItemRequiredField
	}
	scale := 0
	switch uomCode {
	case "MG", "G", "KG", "ML", "L":
		scale = decimal.QuantityScale
	}
	var unitID string
	if err := queryer.QueryRowContext(ctx, upsertPostgresItemUnitSQL, orgID, uomCode, scale).Scan(&unitID); err != nil {
		return "", fmt.Errorf("ensure item base unit %q: %w", uomCode, err)
	}

	return unitID, nil
}

func nullablePostgresItemText(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return value
}

func nullablePostgresItemInt(value int) any {
	if value == 0 {
		return nil
	}

	return value
}

func nullablePostgresItemUUID(value string) any {
	value = strings.TrimSpace(value)
	if !isPostgresItemUUIDText(value) {
		return nil
	}

	return value
}

func isPostgresItemUUIDText(value string) bool {
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
