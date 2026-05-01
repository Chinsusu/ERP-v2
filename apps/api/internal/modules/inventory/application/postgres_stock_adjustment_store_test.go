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

const testStockAdjustmentOrgID = "00000000-0000-4000-8000-000000000001"

func TestPostgresStockAdjustmentStorePersistsLifecycle(t *testing.T) {
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

	if err := seedStockAdjustmentSmokeOrg(ctx, db); err != nil {
		t.Fatalf("seed org: %v", err)
	}

	suffix := fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	store := NewPostgresStockAdjustmentStore(db, PostgresStockAdjustmentStoreConfig{DefaultOrgID: testStockAdjustmentOrgID})
	auditStore := audit.NewPostgresLogStore(db, audit.PostgresLogStoreConfig{DefaultOrgID: testStockAdjustmentOrgID})
	create := NewCreateStockAdjustment(store, auditStore)
	transition := NewTransitionStockAdjustment(store, NewInMemoryStockMovementStore(), auditStore)

	created, err := create.Execute(ctx, CreateStockAdjustmentInput{
		ID:            "adj-s10-03-01-" + suffix,
		AdjustmentNo:  "ADJ-S10-03-01-" + suffix,
		OrgID:         "org-my-pham",
		WarehouseID:   "wh-hcm-fg",
		WarehouseCode: "WH-HCM-FG",
		SourceType:    "smoke",
		SourceID:      "source-s10-03-01-" + suffix,
		Reason:        "S10-03-01 stock adjustment persistence",
		RequestedBy:   "user-erp-admin",
		RequestID:     "req-s10-03-01-create-" + suffix,
		Lines: []CreateStockAdjustmentLineInput{{
			ID:           "line-s10-03-01-" + suffix,
			ItemID:       "item-serum-30ml",
			SKU:          "SERUM-30ML",
			LocationID:   "bin-hcm-pick-a01",
			LocationCode: "PICK-A-01",
			ExpectedQty:  "20",
			CountedQty:   "21",
			BaseUOMCode:  "EA",
			Reason:       "integration variance",
		}},
	})
	if err != nil {
		t.Fatalf("create stock adjustment: %v", err)
	}
	if created.Adjustment.Status != domain.StockAdjustmentStatusDraft {
		t.Fatalf("created status = %s, want draft", created.Adjustment.Status)
	}

	if _, err := transition.Submit(ctx, created.Adjustment.ID, "user-erp-admin", "req-s10-03-01-submit-"+suffix); err != nil {
		t.Fatalf("submit stock adjustment: %v", err)
	}
	if _, err := transition.Approve(ctx, created.Adjustment.ID, "user-erp-admin", "req-s10-03-01-approve-"+suffix); err != nil {
		t.Fatalf("approve stock adjustment: %v", err)
	}
	posted, err := transition.Post(ctx, created.Adjustment.ID, "user-erp-admin", "req-s10-03-01-post-"+suffix)
	if err != nil {
		t.Fatalf("post stock adjustment: %v", err)
	}
	if posted.Adjustment.Status != domain.StockAdjustmentStatusPosted {
		t.Fatalf("posted status = %s, want posted", posted.Adjustment.Status)
	}

	loaded, err := store.FindStockAdjustmentByID(ctx, created.Adjustment.ID)
	if err != nil {
		t.Fatalf("find persisted stock adjustment: %v", err)
	}
	if loaded.Status != domain.StockAdjustmentStatusPosted ||
		loaded.PostedBy != "user-erp-admin" ||
		len(loaded.Lines) != 1 ||
		loaded.Lines[0].DeltaQty.String() != "1.000000" {
		t.Fatalf("loaded adjustment = %+v, want posted lifecycle and delta line", loaded)
	}

	rows, err := store.ListStockAdjustments(ctx)
	if err != nil {
		t.Fatalf("list stock adjustments: %v", err)
	}
	if !containsStockAdjustment(rows, created.Adjustment.ID) {
		t.Fatalf("list stock adjustments missing %s", created.Adjustment.ID)
	}
}

func containsStockAdjustment(rows []domain.StockAdjustment, id string) bool {
	for _, row := range rows {
		if row.ID == id {
			return true
		}
	}

	return false
}

func seedStockAdjustmentSmokeOrg(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(
		ctx,
		`INSERT INTO core.organizations (id, code, name, status)
VALUES ($1, 'ERP_LOCAL', 'ERP Local Demo', 'active')
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
		testStockAdjustmentOrgID,
	)

	return err
}
