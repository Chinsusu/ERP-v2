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

func TestPostgresSubcontractMaterialTransferStorePersistsTransfer(t *testing.T) {
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
	order := subcontractMaterialTransferTestOrder(t)
	order.ID = "sco-s16-03-01-" + suffix
	order.OrderNo = "SCO-S16-03-01-" + suffix
	order.MaterialLines[0].ID = "sco-mat-s16-03-01-a-" + suffix
	order.MaterialLines[1].ID = "sco-mat-s16-03-01-b-" + suffix
	if err := order.Validate(); err != nil {
		t.Fatalf("validate test order: %v", err)
	}

	fixedNow := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)
	builder := NewSubcontractMaterialTransferService()
	builder.clock = func() time.Time { return fixedNow }
	buildResult, err := builder.BuildIssue(ctx, BuildSubcontractMaterialTransferInput{
		ID:                  "smt-s16-03-01-" + suffix,
		TransferNo:          "SMT-S16-03-01-" + suffix,
		Order:               order,
		SourceWarehouseID:   "wh-hcm-rm",
		SourceWarehouseCode: "WH-HCM-RM",
		HandoverBy:          "warehouse-user",
		HandoverAt:          fixedNow.Add(15 * time.Minute),
		ReceivedBy:          "factory-receiver",
		ReceiverContact:     "0988000111",
		VehicleNo:           "51A-12345",
		ActorID:             "warehouse-user",
		Lines: []BuildSubcontractMaterialTransferLineInput{
			{
				ID:                  "smt-line-s16-03-01-a-" + suffix,
				OrderMaterialLineID: order.MaterialLines[0].ID,
				IssueQty:            "10",
				UOMCode:             "KG",
				BatchID:             "batch-base-s16-03-01",
				BatchNo:             "BASE-S16-03-01",
				SourceBinID:         "bin-rm-a01",
			},
			{
				ID:                  "smt-line-s16-03-01-b-" + suffix,
				OrderMaterialLineID: order.MaterialLines[1].ID,
				IssueQty:            "1000",
				UOMCode:             "PCS",
				SourceBinID:         "bin-pk-b01",
			},
		},
		Evidence: []BuildSubcontractMaterialTransferEvidenceInput{{
			ID:           "smt-evidence-s16-03-01-" + suffix,
			EvidenceType: "handover",
			FileName:     "handover.pdf",
			ObjectKey:    "subcontract/smt-s16-03-01/handover.pdf",
		}},
	})
	if err != nil {
		t.Fatalf("build transfer: %v", err)
	}

	store := NewPostgresSubcontractMaterialTransferStore(
		db,
		PostgresSubcontractMaterialTransferStoreConfig{DefaultOrgID: testSubcontractOrderOrgID},
	)
	if err := store.Save(ctx, buildResult.Transfer); err != nil {
		t.Fatalf("save transfer: %v", err)
	}

	transfers, err := store.ListBySubcontractOrder(ctx, order.ID)
	if err != nil {
		t.Fatalf("list transfers: %v", err)
	}
	if len(transfers) != 1 {
		t.Fatalf("transfers count = %d, want 1", len(transfers))
	}
	loaded := transfers[0]
	if loaded.ID != buildResult.Transfer.ID ||
		loaded.SubcontractOrderID != order.ID ||
		loaded.Status != productiondomain.SubcontractMaterialTransferStatusSentToFactory ||
		loaded.FactoryName != order.FactoryName ||
		len(loaded.Lines) != 2 ||
		loaded.Lines[0].BatchID != "batch-base-s16-03-01" ||
		loaded.Lines[0].BaseIssueQty.String() != "10000.000000" ||
		len(loaded.Evidence) != 1 ||
		loaded.Evidence[0].ObjectKey != "subcontract/smt-s16-03-01/handover.pdf" {
		t.Fatalf("loaded transfer = %+v, want persisted transfer document", loaded)
	}

	loaded.Note = "updated handover note"
	loaded.Version = 2
	loaded.UpdatedAt = fixedNow.Add(30 * time.Minute)
	loaded.UpdatedBy = "warehouse-lead"
	if err := store.Save(ctx, loaded); err != nil {
		t.Fatalf("save updated transfer: %v", err)
	}
	transfers, err = store.ListBySubcontractOrder(ctx, order.OrderNo)
	if err != nil {
		t.Fatalf("list updated transfers: %v", err)
	}
	if len(transfers) != 1 || transfers[0].Note != "updated handover note" || transfers[0].Version != 2 {
		t.Fatalf("updated transfers = %+v, want updated transfer state", transfers)
	}
}

func TestPostgresSubcontractMaterialTransferStoreRequiresDatabase(t *testing.T) {
	store := NewPostgresSubcontractMaterialTransferStore(nil, PostgresSubcontractMaterialTransferStoreConfig{})

	if err := store.Save(context.Background(), productiondomain.SubcontractMaterialTransfer{}); err == nil {
		t.Fatal("Save() error = nil, want database required error")
	}
	if _, err := store.ListBySubcontractOrder(context.Background(), "sco-missing"); err == nil {
		t.Fatal("ListBySubcontractOrder() error = nil, want database required error")
	}
}
