package application

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	testPostgresFormulaCatalogOrgID  = "00000000-0000-4000-8000-000000220001"
	testPostgresFormulaCatalogUnitID = "00000000-0000-4000-8000-000000220002"
	testPostgresFormulaCatalogItemID = "00000000-0000-4000-8000-000000220003"
)

func TestPostgresFormulaCatalogRequiresDatabase(t *testing.T) {
	store := NewPostgresFormulaCatalog(nil, nil, PostgresFormulaCatalogConfig{})

	if _, err := store.List(context.Background(), domain.FormulaFilter{}); err == nil {
		t.Fatal("List() error = nil, want database required error")
	}
	if _, err := store.Get(context.Background(), "formula-xff-v1"); err == nil {
		t.Fatal("Get() error = nil, want database required error")
	}
	if _, err := store.Create(context.Background(), CreateFormulaInput{}); err == nil {
		t.Fatal("Create() error = nil, want database required error")
	}
	if _, err := store.Activate(context.Background(), ActivateFormulaInput{}); err == nil {
		t.Fatal("Activate() error = nil, want database required error")
	}
	if _, err := store.CalculateRequirement(context.Background(), CalculateFormulaRequirementInput{}); err == nil {
		t.Fatal("CalculateRequirement() error = nil, want database required error")
	}
}

func TestPostgresFormulaCatalogListsByParentItemReference(t *testing.T) {
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

	if err := seedPostgresFormulaCatalogFixture(ctx, db); err != nil {
		t.Fatalf("seed fixture: %v", err)
	}

	auditStore := audit.NewPostgresLogStore(db, audit.PostgresLogStoreConfig{DefaultOrgID: testPostgresFormulaCatalogOrgID})
	store := NewPostgresFormulaCatalog(db, auditStore, PostgresFormulaCatalogConfig{
		DefaultOrgID: testPostgresFormulaCatalogOrgID,
		Clock:        fixedPostgresFormulaClock(time.Date(2026, 5, 4, 9, 0, 0, 0, time.UTC)),
	})

	created, err := store.Create(ctx, CreateFormulaInput{
		FormulaCode:      "S22-FORMULA",
		FinishedItemID:   "item-s22-finished",
		FinishedSKU:      "CLIENT-SKU-IGNORED",
		FinishedItemName: "Client name ignored",
		FinishedItemType: "raw_material",
		FormulaVersion:   "v1",
		BatchQty:         decimal.MustQuantity("81"),
		BatchUOMCode:     "PCS",
		BaseBatchQty:     decimal.MustQuantity("81"),
		BaseBatchUOMCode: "PCS",
		Lines: []CreateFormulaLineInput{
			{
				LineNo:           1,
				ComponentSKU:     "ACT_BAICAPIL",
				ComponentName:    "BAICAPIL",
				ComponentType:    "raw_material",
				EnteredQty:       decimal.MustQuantity("0.001"),
				EnteredUOMCode:   "KG",
				CalcQty:          decimal.MustQuantity("1"),
				CalcUOMCode:      "G",
				StockBaseQty:     decimal.MustQuantity("0.001"),
				StockBaseUOMCode: "KG",
				WastePercent:     decimal.MustRate("0"),
				IsRequired:       true,
				IsStockManaged:   true,
				LineStatus:       "active",
			},
		},
		ActorID:   "user-erp-admin",
		RequestID: "req-s22-formula-create",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.Formula.FinishedItemID != testPostgresFormulaCatalogItemID {
		t.Fatalf("finished item id = %s, want %s", created.Formula.FinishedItemID, testPostgresFormulaCatalogItemID)
	}

	byItemRef, err := store.List(ctx, domain.FormulaFilter{FinishedItemID: "item-s22-finished"})
	if err != nil {
		t.Fatalf("List() by item_ref error = %v", err)
	}
	if len(byItemRef) != 1 || byItemRef[0].FormulaCode != "S22-FORMULA" {
		t.Fatalf("List() by item_ref = %+v, want created formula", byItemRef)
	}

	bySKU, err := store.List(ctx, domain.FormulaFilter{FinishedItemID: "S22-FG"})
	if err != nil {
		t.Fatalf("List() by SKU error = %v", err)
	}
	if len(bySKU) != 1 || bySKU[0].FormulaCode != "S22-FORMULA" {
		t.Fatalf("List() by SKU = %+v, want created formula", bySKU)
	}
}

func seedPostgresFormulaCatalogFixture(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
INSERT INTO core.organizations (id, code, name, status)
VALUES ($1::uuid, 'S22_FORMULA_ORG', 'S22 Formula Catalog Test Org', 'active')
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
		testPostgresFormulaCatalogOrgID,
	); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO mdm.units (id, org_id, code, name, precision_scale, status)
VALUES ($1::uuid, $2::uuid, 'PCS', 'Piece', 6, 'active')
ON CONFLICT (org_id, code) DO UPDATE
SET name = EXCLUDED.name,
    precision_scale = EXCLUDED.precision_scale,
    status = EXCLUDED.status,
    updated_at = now()`,
		testPostgresFormulaCatalogUnitID,
		testPostgresFormulaCatalogOrgID,
	); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `
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
  requires_batch,
  requires_expiry,
  lot_controlled,
  expiry_controlled,
  qc_required,
  status,
  is_sellable,
  is_purchasable,
  is_producible
)
VALUES (
  $1::uuid,
  $2::uuid,
  'item-s22-finished',
  'S22-FG',
  'S22-FG',
  'S22 Finished Good',
  'finished_good',
  $3::uuid,
  'PCS',
  true,
  true,
  true,
  true,
  true,
  'active',
  true,
  false,
  true
)
ON CONFLICT (org_id, sku) DO UPDATE
SET item_ref = EXCLUDED.item_ref,
    item_code = EXCLUDED.item_code,
    name = EXCLUDED.name,
    item_type = EXCLUDED.item_type,
    base_unit_id = EXCLUDED.base_unit_id,
    uom_base = EXCLUDED.uom_base,
	status = EXCLUDED.status,
	updated_at = now()`,
		testPostgresFormulaCatalogItemID,
		testPostgresFormulaCatalogOrgID,
		testPostgresFormulaCatalogUnitID,
	); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `
DELETE FROM mdm.item_formula_lines AS line
USING mdm.item_formulas AS formula
WHERE line.formula_id = formula.id
  AND formula.org_id = $1::uuid
  AND formula.finished_item_id = $2::uuid`,
		testPostgresFormulaCatalogOrgID,
		testPostgresFormulaCatalogItemID,
	); err != nil {
		return err
	}
	_, err := db.ExecContext(ctx, `
DELETE FROM mdm.item_formulas
WHERE org_id = $1::uuid
  AND finished_item_id = $2::uuid`,
		testPostgresFormulaCatalogOrgID,
		testPostgresFormulaCatalogItemID,
	)

	return err
}

func fixedPostgresFormulaClock(base time.Time) func() time.Time {
	current := base

	return func() time.Time {
		current = current.Add(time.Minute)
		return current
	}
}
