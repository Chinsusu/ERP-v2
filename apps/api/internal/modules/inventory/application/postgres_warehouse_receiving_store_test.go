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

const testWarehouseReceivingOrgID = "00000000-0000-4000-8000-000000000001"

func TestPostgresWarehouseReceivingStorePersistsPostedLifecycle(t *testing.T) {
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

	if err := seedWarehouseReceivingSmokeOrg(ctx, db); err != nil {
		t.Fatalf("seed org: %v", err)
	}

	suffix := fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	store := NewPostgresWarehouseReceivingStore(db, PostgresWarehouseReceivingStoreConfig{DefaultOrgID: testWarehouseReceivingOrgID})
	createdAt := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	receipt, err := domain.NewWarehouseReceiving(domain.NewWarehouseReceivingInput{
		ID:               "grn-s10-04-01-" + suffix,
		OrgID:            "org-my-pham",
		ReceiptNo:        "GRN-S10-04-01-" + suffix,
		WarehouseID:      "wh-hcm-fg",
		WarehouseCode:    "WH-HCM-FG",
		LocationID:       "loc-hcm-fg-recv-01",
		LocationCode:     "FG-RECV-01",
		ReferenceDocType: "manual_receiving",
		ReferenceDocID:   "manual-s10-04-01-" + suffix,
		SupplierID:       "supplier-local",
		DeliveryNoteNo:   "DN-S10-04-01-" + suffix,
		Lines: []domain.NewWarehouseReceivingLineInput{{
			ID:              "grn-line-s10-04-01-" + suffix,
			ItemID:          "item-serum-30ml",
			SKU:             "SERUM-30ML",
			ItemName:        "Vitamin C Serum",
			BatchID:         "batch-serum-s10-04-01-" + suffix,
			BatchNo:         "LOT-S10-04-01",
			LotNo:           "LOT-S10-04-01",
			ExpiryDate:      time.Date(2027, 5, 1, 0, 0, 0, 0, time.UTC),
			Quantity:        decimal.MustQuantity("12"),
			UOMCode:         "EA",
			BaseUOMCode:     "EA",
			PackagingStatus: domain.ReceivingPackagingStatusIntact,
			QCStatus:        domain.QCStatusPass,
		}},
		CreatedBy: "user-erp-admin",
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	})
	if err != nil {
		t.Fatalf("new warehouse receiving: %v", err)
	}
	if err := store.Save(ctx, receipt); err != nil {
		t.Fatalf("save draft receiving: %v", err)
	}

	submitted, err := receipt.Submit("user-erp-admin", createdAt.Add(10*time.Minute))
	if err != nil {
		t.Fatalf("submit receiving: %v", err)
	}
	if err := store.Save(ctx, submitted); err != nil {
		t.Fatalf("save submitted receiving: %v", err)
	}
	inspectReady, err := submitted.MarkInspectReady("user-qa", createdAt.Add(20*time.Minute))
	if err != nil {
		t.Fatalf("mark inspect ready: %v", err)
	}
	if err := store.Save(ctx, inspectReady); err != nil {
		t.Fatalf("save inspect ready receiving: %v", err)
	}
	posted, err := inspectReady.Post("user-erp-admin", createdAt.Add(30*time.Minute))
	if err != nil {
		t.Fatalf("post receiving: %v", err)
	}
	line := posted.Lines[0]
	movement, err := domain.NewStockMovement(domain.NewStockMovementInput{
		MovementNo:       posted.ReceiptNo + "-MV-001",
		MovementType:     domain.MovementPurchaseReceipt,
		OrgID:            posted.OrgID,
		ItemID:           line.ItemID,
		BatchID:          line.BatchID,
		WarehouseID:      line.WarehouseID,
		BinID:            line.LocationID,
		Quantity:         line.Quantity,
		BaseUOMCode:      line.BaseUOMCode.String(),
		SourceQuantity:   line.Quantity,
		SourceUOMCode:    line.UOMCode.String(),
		ConversionFactor: decimal.MustQuantity("1"),
		StockStatus:      domain.StockStatusAvailable,
		SourceDocType:    receivingSourceDocType,
		SourceDocID:      posted.ID,
		SourceDocLineID:  line.ID,
		Reason:           "warehouse receiving posted",
		CreatedBy:        "user-erp-admin",
		MovementAt:       posted.PostedAt,
	})
	if err != nil {
		t.Fatalf("new receiving movement: %v", err)
	}
	posted = posted.AttachStockMovements([]domain.StockMovement{movement})
	if err := store.Save(ctx, posted); err != nil {
		t.Fatalf("save posted receiving: %v", err)
	}

	loaded, err := store.Get(ctx, receipt.ID)
	if err != nil {
		t.Fatalf("get persisted receiving: %v", err)
	}
	if loaded.Status != domain.WarehouseReceivingStatusPosted ||
		loaded.PostedBy != "user-erp-admin" ||
		len(loaded.Lines) != 1 ||
		len(loaded.StockMovements) != 1 ||
		loaded.StockMovements[0].SourceDocID != receipt.ID {
		t.Fatalf("loaded receiving = %+v, want posted receipt with line and movement", loaded)
	}

	rows, err := store.List(ctx, domain.NewWarehouseReceivingFilter("wh-hcm-fg", domain.WarehouseReceivingStatusPosted))
	if err != nil {
		t.Fatalf("list persisted receivings: %v", err)
	}
	if !containsWarehouseReceiving(rows, receipt.ID) {
		t.Fatalf("list receivings missing %s", receipt.ID)
	}
}

func containsWarehouseReceiving(rows []domain.WarehouseReceiving, id string) bool {
	for _, row := range rows {
		if row.ID == id {
			return true
		}
	}

	return false
}

func seedWarehouseReceivingSmokeOrg(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(
		ctx,
		`INSERT INTO core.organizations (id, code, name, status)
VALUES ($1, 'ERP_LOCAL', 'ERP Local Demo', 'active')
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
		testWarehouseReceivingOrgID,
	)

	return err
}
