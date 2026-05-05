package application

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const testWarehouseDocumentOrgID = "00000000-0000-4000-8000-000000000001"

func TestPostgresStockTransferStorePersistsDocument(t *testing.T) {
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

	if err := seedWarehouseDocumentSmokeOrg(ctx, db); err != nil {
		t.Fatalf("seed org: %v", err)
	}

	suffix := fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	transfer, err := domain.NewStockTransfer(domain.NewStockTransferInput{
		ID:                       "st-persist-" + suffix,
		TransferNo:               "ST-PERSIST-" + suffix,
		OrgID:                    "org-my-pham",
		SourceWarehouseID:        "wh-main",
		SourceWarehouseCode:      "MAIN",
		DestinationWarehouseID:   "wh-stage",
		DestinationWarehouseCode: "STAGE",
		ReasonCode:               "uat_replenishment",
		RequestedBy:              "user-warehouse",
		CreatedAt:                time.Now().UTC(),
		Lines: []domain.NewStockTransferLineInput{{
			ID:                      "st-line-" + suffix,
			ItemID:                  "item-serum",
			SKU:                     "SERUM-30ML",
			SourceLocationCode:      "A-01",
			DestinationLocationCode: "STAGE-01",
			Quantity:                decimal.MustQuantity("3"),
			BaseUOMCode:             "PCS",
		}},
	})
	if err != nil {
		t.Fatalf("new transfer: %v", err)
	}

	store := NewPostgresStockTransferStore(db, PostgresWarehouseDocumentStoreConfig{DefaultOrgID: testWarehouseDocumentOrgID})
	if err := store.SaveStockTransfer(ctx, transfer); err != nil {
		t.Fatalf("save transfer: %v", err)
	}
	loaded, err := store.FindStockTransferByID(ctx, transfer.ID)
	if err != nil {
		t.Fatalf("find transfer: %v", err)
	}
	if loaded.ID != transfer.ID || loaded.Lines[0].Quantity.String() != "3.000000" {
		t.Fatalf("loaded transfer = %+v, want persisted payload", loaded)
	}
	rows, err := store.ListStockTransfers(ctx)
	if err != nil {
		t.Fatalf("list transfers: %v", err)
	}
	if !containsStockTransfer(rows, transfer.ID) {
		t.Fatalf("list transfers missing %s", transfer.ID)
	}
}

func TestPostgresWarehouseIssueStorePersistsDocument(t *testing.T) {
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

	if err := seedWarehouseDocumentSmokeOrg(ctx, db); err != nil {
		t.Fatalf("seed org: %v", err)
	}

	suffix := fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	issue, err := domain.NewWarehouseIssue(domain.NewWarehouseIssueInput{
		ID:              "wi-persist-" + suffix,
		IssueNo:         "WI-PERSIST-" + suffix,
		OrgID:           "org-my-pham",
		WarehouseID:     "wh-main",
		WarehouseCode:   "MAIN",
		DestinationType: "factory",
		DestinationName: "Factory A",
		ReasonCode:      "production_plan_issue",
		RequestedBy:     "user-warehouse",
		CreatedAt:       time.Now().UTC(),
		Lines: []domain.NewWarehouseIssueLineInput{{
			ID:                 "wi-line-" + suffix,
			ItemID:             "item-aci-bha",
			SKU:                "ACI_BHA",
			ItemName:           "ACID SALICYLIC",
			Quantity:           decimal.MustQuantity("0.125"),
			BaseUOMCode:        "KG",
			SourceDocumentType: "production_plan",
			SourceDocumentID:   "plan-0001",
		}},
	})
	if err != nil {
		t.Fatalf("new issue: %v", err)
	}

	store := NewPostgresWarehouseIssueStore(db, PostgresWarehouseDocumentStoreConfig{DefaultOrgID: testWarehouseDocumentOrgID})
	if err := store.SaveWarehouseIssue(ctx, issue); err != nil {
		t.Fatalf("save issue: %v", err)
	}
	loaded, err := store.FindWarehouseIssueByID(ctx, issue.ID)
	if err != nil {
		t.Fatalf("find issue: %v", err)
	}
	if loaded.ID != issue.ID || loaded.Lines[0].Quantity.String() != "0.125000" {
		t.Fatalf("loaded issue = %+v, want persisted payload", loaded)
	}
	rows, err := store.ListWarehouseIssues(ctx)
	if err != nil {
		t.Fatalf("list issues: %v", err)
	}
	if !containsWarehouseIssue(rows, issue.ID) {
		t.Fatalf("list issues missing %s", issue.ID)
	}
}

func containsStockTransfer(rows []domain.StockTransfer, id string) bool {
	for _, row := range rows {
		if row.ID == id {
			return true
		}
	}

	return false
}

func containsWarehouseIssue(rows []domain.WarehouseIssue, id string) bool {
	for _, row := range rows {
		if row.ID == id {
			return true
		}
	}

	return false
}

func seedWarehouseDocumentSmokeOrg(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(
		ctx,
		`INSERT INTO core.organizations (id, code, name, status)
VALUES ($1, 'ERP_LOCAL', 'ERP Local Demo', 'active')
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
		testWarehouseDocumentOrgID,
	)

	return err
}
