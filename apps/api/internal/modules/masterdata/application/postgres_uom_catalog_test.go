package application

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	testPostgresUOMOrgID  = "00000000-0000-4000-8000-000000170501"
	testPostgresUOMUnitID = "00000000-0000-4000-8000-000000170502"
	testPostgresUOMItemID = "00000000-0000-4000-8000-000000170503"
)

func TestPostgresUOMCatalogRequiresDatabase(t *testing.T) {
	catalog := NewPostgresUOMCatalog(nil)

	if _, err := catalog.ConvertToBase(context.Background(), ConvertToBaseInput{}); err == nil {
		t.Fatal("ConvertToBase() error = nil, want database required error")
	}
}

func TestPostgresUOMCatalogPersistsConversionsAndReloads(t *testing.T) {
	databaseURL := os.Getenv("ERP_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("ERP_TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if err := seedPostgresUOMCatalogFixture(ctx, db); err != nil {
		t.Fatalf("seed fixture: %v", err)
	}
	catalog := NewPostgresUOMCatalog(db)

	global, err := catalog.ConvertToBase(ctx, ConvertToBaseInput{
		SKU:         "S17-POWDER",
		Quantity:    decimal.MustQuantity("0.125"),
		FromUOMCode: "KG",
		BaseUOMCode: "G",
	})
	if err != nil {
		t.Fatalf("global ConvertToBase() error = %v", err)
	}
	if global.BaseQuantity.String() != "125.000000" ||
		global.ConversionFactor.String() != "1000.000000" ||
		global.ConversionType != domain.UOMConversionGlobal {
		t.Fatalf("global result = %+v", global)
	}

	itemSpecific, err := domain.NewUOMConversion(
		"item-s17-carton-pcs",
		"item-s17-uom",
		"CARTON",
		"PCS",
		"24",
		domain.UOMConversionItemSpecific,
		true,
	)
	if err != nil {
		t.Fatalf("new item conversion: %v", err)
	}
	if err := catalog.upsertConversion(ctx, itemSpecific); err != nil {
		t.Fatalf("upsert item conversion: %v", err)
	}

	reloaded := NewPostgresUOMCatalog(db)
	converted, err := reloaded.ConvertToBase(ctx, ConvertToBaseInput{
		ItemID:      "item-s17-uom",
		SKU:         "S17-UOM",
		Quantity:    decimal.MustQuantity("2"),
		FromUOMCode: "CARTON",
		BaseUOMCode: "PCS",
	})
	if err != nil {
		t.Fatalf("item ConvertToBase() error = %v", err)
	}
	if converted.BaseQuantity.String() != "48.000000" ||
		converted.ConversionFactor.String() != "24.000000" ||
		converted.ConversionType != domain.UOMConversionItemSpecific ||
		converted.ConversionItemID != "item-s17-uom" {
		t.Fatalf("item result = %+v", converted)
	}

	inactive, err := domain.NewUOMConversion(
		"item-s17-carton-pcs",
		"item-s17-uom",
		"CARTON",
		"PCS",
		"24",
		domain.UOMConversionItemSpecific,
		false,
	)
	if err != nil {
		t.Fatalf("new inactive conversion: %v", err)
	}
	if err := catalog.upsertConversion(ctx, inactive); err != nil {
		t.Fatalf("upsert inactive conversion: %v", err)
	}
	_, err = reloaded.ConvertToBase(ctx, ConvertToBaseInput{
		ItemID:      "item-s17-uom",
		SKU:         "S17-UOM",
		Quantity:    decimal.MustQuantity("2"),
		FromUOMCode: "CARTON",
		BaseUOMCode: "PCS",
	})
	if !errors.Is(err, domain.ErrUOMConversionInactive) {
		t.Fatalf("inactive conversion err = %v, want ErrUOMConversionInactive", err)
	}

	_, err = reloaded.ConvertToBase(ctx, ConvertToBaseInput{
		ItemID:      "item-s17-uom",
		SKU:         "S17-UOM",
		Quantity:    decimal.MustQuantity("2"),
		FromUOMCode: "BOX",
		BaseUOMCode: "PCS",
	})
	if !errors.Is(err, domain.ErrUOMConversionMissing) {
		t.Fatalf("missing conversion err = %v, want ErrUOMConversionMissing", err)
	}
}

func seedPostgresUOMCatalogFixture(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
INSERT INTO core.organizations (id, code, name, status)
VALUES ($1::uuid, 'S17_UOM_ORG', 'S17 UOM Test Org', 'active')
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
		testPostgresUOMOrgID,
	); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO mdm.units (id, org_id, code, name, precision_scale, status)
VALUES ($1::uuid, $2::uuid, 'PCS', 'Piece', 0, 'active')
ON CONFLICT (org_id, code) DO UPDATE
SET name = EXCLUDED.name,
    precision_scale = EXCLUDED.precision_scale,
    status = EXCLUDED.status,
    updated_at = now()`,
		testPostgresUOMUnitID,
		testPostgresUOMOrgID,
	); err != nil {
		return err
	}
	_, err := db.ExecContext(ctx, `
INSERT INTO mdm.items (
  id,
  org_id,
  item_ref,
  item_code,
  sku,
  name,
  item_type,
  base_unit_id,
  uom_base,
  uom_purchase,
  uom_issue,
  requires_batch,
  requires_expiry,
  lot_controlled,
  expiry_controlled,
  qc_required,
  status,
  standard_cost,
  is_sellable,
  is_producible
) VALUES (
  $1::uuid,
  $2::uuid,
  'item-s17-uom',
  'ITEM-S17-UOM',
  'S17-UOM',
  'S17 UOM Item',
  'finished_good',
  $3::uuid,
  'PCS',
  'PCS',
  'PCS',
  true,
  true,
  true,
  true,
  true,
  'active',
  0,
  true,
  true
)
ON CONFLICT (org_id, sku) DO UPDATE
SET item_ref = EXCLUDED.item_ref,
    item_code = EXCLUDED.item_code,
    name = EXCLUDED.name,
    base_unit_id = EXCLUDED.base_unit_id,
    status = EXCLUDED.status,
    updated_at = now()`,
		testPostgresUOMItemID,
		testPostgresUOMOrgID,
		testPostgresUOMUnitID,
	)

	return err
}
