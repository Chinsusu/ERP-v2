package application

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const testStockCountOrgID = "00000000-0000-4000-8000-000000000001"

func TestPostgresStockCountStorePersistsSubmittedVariance(t *testing.T) {
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

	if err := seedStockCountSmokeOrg(ctx, db); err != nil {
		t.Fatalf("seed org: %v", err)
	}

	suffix := fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	store := NewPostgresStockCountStore(db, PostgresStockCountStoreConfig{DefaultOrgID: testStockCountOrgID})
	auditStore := audit.NewPostgresLogStore(db, audit.PostgresLogStoreConfig{DefaultOrgID: testStockCountOrgID})
	create := NewCreateStockCount(store, auditStore)
	submit := NewSubmitStockCount(store, auditStore)

	created, err := create.Execute(ctx, CreateStockCountInput{
		ID:            "count-s10-03-02-" + suffix,
		CountNo:       "CNT-S10-03-02-" + suffix,
		OrgID:         "org-my-pham",
		WarehouseID:   "wh-hcm-fg",
		WarehouseCode: "WH-HCM-FG",
		Scope:         "cycle_count",
		CreatedBy:     "user-erp-admin",
		RequestID:     "req-s10-03-02-create-" + suffix,
		Lines: []CreateStockCountLineInput{{
			ID:           "count-line-s10-03-02-" + suffix,
			ItemID:       "item-serum-30ml",
			SKU:          "SERUM-30ML",
			BatchID:      "batch-serum-30ml",
			BatchNo:      "LOT-S10-03-02",
			LocationID:   "bin-hcm-pick-a01",
			LocationCode: "PICK-A-01",
			ExpectedQty:  "20",
			BaseUOMCode:  "EA",
		}},
	})
	if err != nil {
		t.Fatalf("create stock count: %v", err)
	}
	if created.Session.Status != domain.StockCountStatusOpen {
		t.Fatalf("created status = %s, want open", created.Session.Status)
	}

	submitted, err := submit.Execute(ctx, SubmitStockCountInput{
		ID:          created.Session.ID,
		SubmittedBy: "user-erp-admin",
		RequestID:   "req-s10-03-02-submit-" + suffix,
		Lines: []SubmitStockCountLineInput{{
			ID:         "count-line-s10-03-02-" + suffix,
			CountedQty: "18",
			Note:       "integration short count",
		}},
	})
	if err != nil {
		t.Fatalf("submit stock count: %v", err)
	}
	if submitted.Session.Status != domain.StockCountStatusVarianceReview {
		t.Fatalf("submitted status = %s, want variance_review", submitted.Session.Status)
	}

	loaded, err := store.FindStockCountByID(ctx, created.Session.ID)
	if err != nil {
		t.Fatalf("find persisted stock count: %v", err)
	}
	if loaded.Status != domain.StockCountStatusVarianceReview ||
		loaded.SubmittedBy != "user-erp-admin" ||
		len(loaded.Lines) != 1 ||
		!loaded.Lines[0].Counted ||
		loaded.Lines[0].DeltaQty.String() != "-2.000000" {
		t.Fatalf("loaded stock count = %+v, want submitted variance line", loaded)
	}

	rows, err := store.ListStockCounts(ctx)
	if err != nil {
		t.Fatalf("list stock counts: %v", err)
	}
	if !containsStockCount(rows, created.Session.ID) {
		t.Fatalf("list stock counts missing %s", created.Session.ID)
	}
}

func containsStockCount(rows []domain.StockCountSession, id string) bool {
	for _, row := range rows {
		if row.ID == id {
			return true
		}
	}

	return false
}

func seedStockCountSmokeOrg(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(
		ctx,
		`INSERT INTO core.organizations (id, code, name, status)
VALUES ($1, 'ERP_LOCAL', 'ERP Local Demo', 'active')
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
		testStockCountOrgID,
	)

	return err
}
