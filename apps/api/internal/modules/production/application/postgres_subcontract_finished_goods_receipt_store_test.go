package application

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestPostgresSubcontractFinishedGoodsReceiptStorePersistsReceipt(t *testing.T) {
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

	if err := seedSubcontractOrderSmokeOrg(ctx, db); err != nil {
		t.Fatalf("seed org: %v", err)
	}

	suffix := fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	order := buildMassProductionSubcontractOrderForFinishedGoodsReceiptStore(t, suffix)
	orderStore := NewPostgresSubcontractOrderStore(
		db,
		PostgresSubcontractOrderStoreConfig{DefaultOrgID: testSubcontractOrderOrgID},
	)
	if err := orderStore.WithinTx(ctx, func(txCtx context.Context, tx SubcontractOrderTx) error {
		return tx.Save(txCtx, order)
	}); err != nil {
		t.Fatalf("seed mass production subcontract order: %v", err)
	}

	fixedNow := time.Date(2026, 5, 2, 13, 0, 0, 0, time.UTC)
	builder := NewSubcontractFinishedGoodsReceiptService()
	builder.clock = func() time.Time { return fixedNow }
	buildResult, err := builder.BuildReceipt(ctx, BuildSubcontractFinishedGoodsReceiptInput{
		ID:             "sfgr-s16-05-01-" + suffix,
		ReceiptNo:      "SFGR-S16-05-01-" + suffix,
		Order:          order,
		WarehouseID:    "wh-hcm-fg",
		WarehouseCode:  "WH-HCM-FG",
		LocationID:     "loc-hcm-fg-qc",
		LocationCode:   "FG-QC-01",
		DeliveryNoteNo: "DN-S16-05-01-" + suffix,
		ReceivedBy:     "warehouse-user",
		ReceivedAt:     fixedNow.Add(30 * time.Minute),
		ActorID:        "warehouse-user",
		Lines: []BuildSubcontractFinishedGoodsReceiptLineInput{{
			ID:              "sfgr-line-s16-05-01-" + suffix,
			ReceiveQty:      "80",
			UOMCode:         "PCS",
			BatchID:         "batch-fg-s16-05-01",
			BatchNo:         "LOT-FG-S16-05-01",
			LotNo:           "LOT-FG-S16-05-01",
			ExpiryDate:      "2028-05-02",
			PackagingStatus: "intact",
		}},
		Evidence: []BuildSubcontractFinishedGoodsReceiptEvidenceInput{{
			ID:           "sfgr-evidence-s16-05-01-" + suffix,
			EvidenceType: "delivery_note",
			FileName:     "factory-delivery.pdf",
			ObjectKey:    "subcontract/sfgr-s16-05-01/factory-delivery.pdf",
		}},
	})
	if err != nil {
		t.Fatalf("build finished goods receipt: %v", err)
	}

	store := NewPostgresSubcontractFinishedGoodsReceiptStore(
		db,
		PostgresSubcontractFinishedGoodsReceiptStoreConfig{DefaultOrgID: testSubcontractOrderOrgID},
	)
	if err := store.Save(ctx, buildResult.Receipt); err != nil {
		t.Fatalf("save finished goods receipt: %v", err)
	}

	receipts, err := store.ListBySubcontractOrder(ctx, order.ID)
	if err != nil {
		t.Fatalf("list finished goods receipts: %v", err)
	}
	if len(receipts) != 1 {
		t.Fatalf("receipt count = %d, want 1", len(receipts))
	}
	loaded := receipts[0]
	if loaded.ID != buildResult.Receipt.ID ||
		loaded.SubcontractOrderID != order.ID ||
		loaded.Status != productiondomain.SubcontractFinishedGoodsReceiptStatusQCHold ||
		loaded.FactoryName != order.FactoryName ||
		loaded.DeliveryNoteNo != "DN-S16-05-01-"+suffix ||
		len(loaded.Lines) != 1 ||
		loaded.Lines[0].BatchID != "batch-fg-s16-05-01" ||
		loaded.Lines[0].ReceiveQty.String() != "80.000000" ||
		loaded.Lines[0].BaseReceiveQty.String() != "80.000000" ||
		loaded.Lines[0].PackagingStatus != "intact" ||
		len(loaded.Evidence) != 1 ||
		loaded.Evidence[0].ObjectKey != "subcontract/sfgr-s16-05-01/factory-delivery.pdf" {
		t.Fatalf("loaded receipt = %+v, want persisted receipt document", loaded)
	}
	if got, want := loaded.Lines[0].ExpiryDate.Format("2006-01-02"), "2028-05-02"; got != want {
		t.Fatalf("expiry date = %s, want %s", got, want)
	}

	loaded.Note = "updated finished goods receipt note"
	loaded.Version = 2
	loaded.UpdatedAt = fixedNow.Add(time.Hour)
	loaded.UpdatedBy = "warehouse-lead"
	if err := store.Save(ctx, loaded); err != nil {
		t.Fatalf("save updated finished goods receipt: %v", err)
	}
	receipts, err = store.ListBySubcontractOrder(ctx, order.OrderNo)
	if err != nil {
		t.Fatalf("list updated finished goods receipts: %v", err)
	}
	if len(receipts) != 1 || receipts[0].Note != "updated finished goods receipt note" || receipts[0].Version != 2 {
		t.Fatalf("updated receipts = %+v, want updated note/version", receipts)
	}
}

func TestPostgresSubcontractFinishedGoodsReceiptStoreRequiresDatabase(t *testing.T) {
	store := NewPostgresSubcontractFinishedGoodsReceiptStore(nil, PostgresSubcontractFinishedGoodsReceiptStoreConfig{})

	if err := store.Save(context.Background(), productiondomain.SubcontractFinishedGoodsReceipt{}); err == nil {
		t.Fatal("Save() error = nil, want database required error")
	}
	if _, err := store.ListBySubcontractOrder(context.Background(), "sco-missing"); err == nil {
		t.Fatal("ListBySubcontractOrder() error = nil, want database required error")
	}
}

func buildMassProductionSubcontractOrderForFinishedGoodsReceiptStore(t *testing.T, suffix string) productiondomain.SubcontractOrder {
	t.Helper()

	order := subcontractMaterialTransferTestOrder(t)
	order.ID = "sco-s16-05-01-" + suffix
	order.OrderNo = "SCO-S16-05-01-" + suffix
	order.SampleRequired = false
	order.MaterialLines[0].ID = "sco-mat-s16-05-01-a-" + suffix
	order.MaterialLines[1].ID = "sco-mat-s16-05-01-b-" + suffix
	if err := order.Validate(); err != nil {
		t.Fatalf("validate finished goods receipt test order: %v", err)
	}

	fixedNow := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)
	materialBuilder := NewSubcontractMaterialTransferService()
	materialBuilder.clock = func() time.Time { return fixedNow }
	issued, err := materialBuilder.BuildIssue(context.Background(), BuildSubcontractMaterialTransferInput{
		ID:                  "smt-s16-05-01-" + suffix,
		TransferNo:          "SMT-S16-05-01-" + suffix,
		Order:               order,
		SourceWarehouseID:   "wh-hcm-rm",
		SourceWarehouseCode: "WH-HCM-RM",
		HandoverBy:          "warehouse-user",
		HandoverAt:          fixedNow.Add(15 * time.Minute),
		ReceivedBy:          "factory-receiver",
		ActorID:             "warehouse-user",
		Lines: []BuildSubcontractMaterialTransferLineInput{
			{
				ID:                  "smt-line-s16-05-01-a-" + suffix,
				OrderMaterialLineID: order.MaterialLines[0].ID,
				IssueQty:            "10",
				UOMCode:             "KG",
				BatchID:             "batch-base-s16-05-01",
				BatchNo:             "BASE-S16-05-01",
				SourceBinID:         "bin-rm-a01",
			},
			{
				ID:                  "smt-line-s16-05-01-b-" + suffix,
				OrderMaterialLineID: order.MaterialLines[1].ID,
				IssueQty:            "1000",
				UOMCode:             "PCS",
				SourceBinID:         "bin-pk-b01",
			},
		},
	})
	if err != nil {
		t.Fatalf("build issued finished goods receipt test order: %v", err)
	}
	started, err := issued.UpdatedOrder.StartMassProduction("operations-lead", fixedNow.Add(time.Hour))
	if err != nil {
		t.Fatalf("start mass production: %v", err)
	}

	return started
}
