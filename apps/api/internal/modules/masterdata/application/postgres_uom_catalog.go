package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

type PostgresUOMCatalog struct {
	db *sql.DB
}

func NewPostgresUOMCatalog(db *sql.DB) *PostgresUOMCatalog {
	return &PostgresUOMCatalog{db: db}
}

const upsertPostgresUOMSQL = `
INSERT INTO mdm.uoms (
  uom_code,
  name_vi,
  name_en,
  uom_group,
  decimal_scale,
  allow_decimal,
  is_global_convertible,
  is_active,
  description
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9
)
ON CONFLICT (uom_code) DO UPDATE SET
  name_vi = EXCLUDED.name_vi,
  name_en = EXCLUDED.name_en,
  uom_group = EXCLUDED.uom_group,
  decimal_scale = EXCLUDED.decimal_scale,
  allow_decimal = EXCLUDED.allow_decimal,
  is_global_convertible = EXCLUDED.is_global_convertible,
  is_active = EXCLUDED.is_active,
  description = EXCLUDED.description,
  updated_at = now()`

const upsertPostgresUOMConversionSQL = `
INSERT INTO mdm.uom_conversions (
  id,
  conversion_ref,
  item_id,
  item_ref,
  from_uom_code,
  to_uom_code,
  factor,
  conversion_type,
  is_active
) VALUES (
  COALESCE($1::uuid, gen_random_uuid()),
  $2,
  $3::uuid,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9
)
ON CONFLICT ON CONSTRAINT uq_uom_conversions_scope DO UPDATE SET
  conversion_ref = EXCLUDED.conversion_ref,
  item_ref = EXCLUDED.item_ref,
  factor = EXCLUDED.factor,
  conversion_type = EXCLUDED.conversion_type,
  is_active = EXCLUDED.is_active,
  updated_at = now()`

const selectPostgresUOMActiveSQL = `
SELECT is_active
FROM mdm.uoms
WHERE uom_code = $1
LIMIT 1`

const selectPostgresItemSpecificUOMConversionSQL = `
SELECT
  COALESCE(conversion_ref, id::text),
  COALESCE(item_ref, ''),
  from_uom_code,
  to_uom_code,
  factor::text,
  conversion_type,
  is_active
FROM mdm.uom_conversions
WHERE lower(COALESCE(item_ref, '')) = lower($1)
  AND from_uom_code = $2
  AND to_uom_code = $3
LIMIT 1`

const selectPostgresGlobalUOMConversionSQL = `
SELECT
  COALESCE(conversion_ref, id::text),
  '',
  from_uom_code,
  to_uom_code,
  factor::text,
  conversion_type,
  is_active
FROM mdm.uom_conversions
WHERE item_id IS NULL
  AND nullif(btrim(COALESCE(item_ref, '')), '') IS NULL
  AND from_uom_code = $1
  AND to_uom_code = $2
LIMIT 1`

const selectPostgresUOMItemIDSQL = `
SELECT id::text
FROM mdm.items
WHERE lower(COALESCE(item_ref, id::text)) = lower($1)
   OR id::text = $1
   OR lower(sku) = lower($1)
LIMIT 1`

func (c *PostgresUOMCatalog) ConvertToBase(ctx context.Context, input ConvertToBaseInput) (ConvertToBaseResult, error) {
	if c == nil || c.db == nil {
		return ConvertToBaseResult{}, errors.New("database connection is required")
	}
	if err := c.ensureSeed(ctx); err != nil {
		return ConvertToBaseResult{}, err
	}

	quantity, err := decimal.ParseQuantity(input.Quantity.String())
	if err != nil || quantity.IsNegative() || quantity.IsZero() {
		return ConvertToBaseResult{}, domain.ErrUOMInvalid
	}
	fromUOM, err := decimal.NormalizeUOMCode(input.FromUOMCode)
	if err != nil {
		return ConvertToBaseResult{}, domain.ErrUOMInvalid
	}
	baseUOM, err := decimal.NormalizeUOMCode(input.BaseUOMCode)
	if err != nil {
		return ConvertToBaseResult{}, domain.ErrUOMInvalid
	}
	if err := c.ensureActiveUOM(ctx, fromUOM); err != nil {
		return ConvertToBaseResult{}, c.conversionError(err, input, fromUOM, baseUOM)
	}
	if err := c.ensureActiveUOM(ctx, baseUOM); err != nil {
		return ConvertToBaseResult{}, c.conversionError(err, input, fromUOM, baseUOM)
	}
	if fromUOM == baseUOM {
		return ConvertToBaseResult{
			Quantity:          quantity,
			SourceUOMCode:     fromUOM,
			BaseQuantity:      quantity,
			BaseUOMCode:       baseUOM,
			ConversionFactor:  decimal.MustQuantity("1"),
			ConversionType:    domain.UOMConversionDirect,
			IsBasePassthrough: true,
		}, nil
	}

	conversion, err := c.findConversion(ctx, strings.TrimSpace(input.ItemID), fromUOM, baseUOM)
	if err != nil {
		return ConvertToBaseResult{}, c.conversionError(err, input, fromUOM, baseUOM)
	}
	baseQuantity, err := decimal.MultiplyQuantityByFactor(quantity, conversion.Factor)
	if err != nil {
		return ConvertToBaseResult{}, err
	}

	return ConvertToBaseResult{
		Quantity:         quantity,
		SourceUOMCode:    fromUOM,
		BaseQuantity:     baseQuantity,
		BaseUOMCode:      baseUOM,
		ConversionFactor: conversion.Factor,
		ConversionType:   conversion.ConversionType,
		ConversionItemID: conversion.ItemID,
	}, nil
}

func (c *PostgresUOMCatalog) UpsertConversion(conversion domain.UOMConversion) {
	if c == nil || c.db == nil {
		return
	}
	_ = c.upsertConversion(context.Background(), conversion)
}

func (c *PostgresUOMCatalog) ensureActiveUOM(ctx context.Context, code decimal.UOMCode) error {
	var active bool
	err := c.db.QueryRowContext(ctx, selectPostgresUOMActiveSQL, code.String()).Scan(&active)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrUOMInvalid
	}
	if err != nil {
		return err
	}
	if !active {
		return domain.ErrUOMInactive
	}

	return nil
}

func (c *PostgresUOMCatalog) findConversion(ctx context.Context, itemID string, fromUOM decimal.UOMCode, toUOM decimal.UOMCode) (domain.UOMConversion, error) {
	if strings.TrimSpace(itemID) != "" {
		conversion, err := scanPostgresUOMConversion(
			c.db.QueryRowContext(ctx, selectPostgresItemSpecificUOMConversionSQL, itemID, fromUOM.String(), toUOM.String()),
		)
		if err == nil {
			if !conversion.IsActive {
				return domain.UOMConversion{}, domain.ErrUOMConversionInactive
			}
			return conversion, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return domain.UOMConversion{}, err
		}
	}

	conversion, err := scanPostgresUOMConversion(
		c.db.QueryRowContext(ctx, selectPostgresGlobalUOMConversionSQL, fromUOM.String(), toUOM.String()),
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.UOMConversion{}, domain.ErrUOMConversionMissing
	}
	if err != nil {
		return domain.UOMConversion{}, err
	}
	if !conversion.IsActive {
		return domain.UOMConversion{}, domain.ErrUOMConversionInactive
	}

	return conversion, nil
}

func (c *PostgresUOMCatalog) conversionError(cause error, input ConvertToBaseInput, fromUOM decimal.UOMCode, toUOM decimal.UOMCode) error {
	return domain.UOMConversionError{
		Cause:       cause,
		SKU:         strings.TrimSpace(input.SKU),
		ItemID:      strings.TrimSpace(input.ItemID),
		FromUOMCode: fromUOM.String(),
		ToUOMCode:   toUOM.String(),
	}
}

func (c *PostgresUOMCatalog) ensureSeed(ctx context.Context) error {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin uom seed transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	for _, uom := range phase1UOMs() {
		if _, err := tx.ExecContext(
			ctx,
			upsertPostgresUOMSQL,
			uom.Code.String(),
			uom.NameVI,
			uom.NameEN,
			string(uom.Group),
			uom.DecimalScale,
			uom.AllowDecimal,
			uom.IsGlobalConvertible,
			uom.IsActive,
			uom.Description,
		); err != nil {
			return fmt.Errorf("upsert uom %s: %w", uom.Code, err)
		}
	}
	for _, conversion := range phase1Conversions() {
		if conversion.ConversionType != domain.UOMConversionGlobal {
			continue
		}
		if err := upsertPostgresUOMConversion(ctx, tx, "", conversion); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit uom seed transaction: %w", err)
	}
	committed = true

	return nil
}

func (c *PostgresUOMCatalog) upsertConversion(ctx context.Context, conversion domain.UOMConversion) error {
	if err := c.ensureSeed(ctx); err != nil {
		return err
	}
	itemPersistedID := ""
	if strings.TrimSpace(conversion.ItemID) != "" {
		err := c.db.QueryRowContext(ctx, selectPostgresUOMItemIDSQL, conversion.ItemID).Scan(&itemPersistedID)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrUOMConversionMissing
		}
		if err != nil {
			return fmt.Errorf("resolve uom conversion item %q: %w", conversion.ItemID, err)
		}
	}

	return upsertPostgresUOMConversion(ctx, c.db, itemPersistedID, conversion)
}

func upsertPostgresUOMConversion(ctx context.Context, queryer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}, itemPersistedID string, conversion domain.UOMConversion) error {
	if _, err := queryer.ExecContext(
		ctx,
		upsertPostgresUOMConversionSQL,
		nullablePostgresUOMUUID(conversion.ID),
		conversion.ID,
		nullablePostgresUOMUUID(itemPersistedID),
		nullablePostgresUOMText(conversion.ItemID),
		conversion.FromUOMCode.String(),
		conversion.ToUOMCode.String(),
		conversion.Factor.String(),
		string(conversion.ConversionType),
		conversion.IsActive,
	); err != nil {
		return fmt.Errorf("upsert uom conversion %q: %w", conversion.ID, err)
	}

	return nil
}

func scanPostgresUOMConversion(scanner interface{ Scan(dest ...any) error }) (domain.UOMConversion, error) {
	var (
		id             string
		itemID         string
		fromUOMCode    string
		toUOMCode      string
		factor         string
		conversionType string
		active         bool
	)
	if err := scanner.Scan(
		&id,
		&itemID,
		&fromUOMCode,
		&toUOMCode,
		&factor,
		&conversionType,
		&active,
	); err != nil {
		return domain.UOMConversion{}, err
	}

	return domain.NewUOMConversion(id, itemID, fromUOMCode, toUOMCode, factor, domain.UOMConversionType(conversionType), active)
}

func nullablePostgresUOMText(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return value
}

func nullablePostgresUOMUUID(value string) any {
	value = strings.TrimSpace(value)
	if !isPostgresUOMUUIDText(value) {
		return nil
	}

	return value
}

func isPostgresUOMUUIDText(value string) bool {
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
