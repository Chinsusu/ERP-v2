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

func TestPostgresSubcontractSampleApprovalStorePersistsSampleApproval(t *testing.T) {
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
	order := buildIssuedSubcontractOrderForSampleStore(t, suffix)
	orderStore := NewPostgresSubcontractOrderStore(
		db,
		PostgresSubcontractOrderStoreConfig{DefaultOrgID: testSubcontractOrderOrgID},
	)
	if err := orderStore.WithinTx(ctx, func(txCtx context.Context, tx SubcontractOrderTx) error {
		return tx.Save(txCtx, order)
	}); err != nil {
		t.Fatalf("seed issued subcontract order: %v", err)
	}

	sampleStore := NewPostgresSubcontractSampleApprovalStore(
		db,
		PostgresSubcontractSampleApprovalStoreConfig{DefaultOrgID: testSubcontractOrderOrgID},
	)
	submittedAt := time.Date(2026, 5, 2, 11, 0, 0, 0, time.UTC)
	sampleBuilder := NewSubcontractSampleApprovalService()
	submitted, err := sampleBuilder.BuildSubmission(ctx, BuildSubcontractSampleSubmissionInput{
		ID:             "ssa-s16-04-01-" + suffix,
		Order:          order,
		SampleCode:     "SCO-S16-04-01-SAMPLE-A-" + suffix,
		FormulaVersion: "formula-2026.05",
		SpecVersion:    "spec-2026.05",
		SubmittedBy:    "factory-user",
		SubmittedAt:    submittedAt,
		ActorID:        "qa-user",
		Evidence: []BuildSubcontractSampleEvidenceInput{{
			ID:           "ssa-evidence-s16-04-01-" + suffix,
			EvidenceType: "photo",
			FileName:     "sample-front.jpg",
			ObjectKey:    "subcontract/ssa-s16-04-01/sample-front.jpg",
		}},
	})
	if err != nil {
		t.Fatalf("build sample submission: %v", err)
	}
	if err := sampleStore.Save(ctx, submitted.SampleApproval); err != nil {
		t.Fatalf("save submitted sample approval: %v", err)
	}

	loaded, err := sampleStore.Get(ctx, submitted.SampleApproval.ID)
	if err != nil {
		t.Fatalf("get submitted sample approval: %v", err)
	}
	if loaded.ID != submitted.SampleApproval.ID ||
		loaded.Status != productiondomain.SubcontractSampleApprovalStatusSubmitted ||
		loaded.SampleCode != submitted.SampleApproval.SampleCode ||
		loaded.FormulaVersion != "formula-2026.05" ||
		len(loaded.Evidence) != 1 ||
		loaded.Evidence[0].ObjectKey != "subcontract/ssa-s16-04-01/sample-front.jpg" {
		t.Fatalf("loaded sample approval = %+v, want persisted submitted sample with evidence", loaded)
	}
	latest, err := sampleStore.GetLatestBySubcontractOrder(ctx, order.ID)
	if err != nil {
		t.Fatalf("get latest sample approval by order: %v", err)
	}
	if latest.ID != submitted.SampleApproval.ID {
		t.Fatalf("latest sample approval = %+v, want %s", latest, submitted.SampleApproval.ID)
	}

	approvedAt := submittedAt.Add(time.Hour)
	approved, err := sampleBuilder.BuildApproval(ctx, BuildSubcontractSampleDecisionInput{
		Order:          submitted.UpdatedOrder,
		SampleApproval: loaded,
		DecisionBy:     "qa-lead",
		DecisionAt:     approvedAt,
		Reason:         "approved against spec",
		StorageStatus:  "retained_in_qa_cabinet",
	})
	if err != nil {
		t.Fatalf("build sample approval: %v", err)
	}
	if err := sampleStore.Save(ctx, approved.SampleApproval); err != nil {
		t.Fatalf("save approved sample approval: %v", err)
	}
	loaded, err = sampleStore.Get(ctx, approved.SampleApproval.SampleCode)
	if err != nil {
		t.Fatalf("get approved sample approval by sample code: %v", err)
	}
	if loaded.Status != productiondomain.SubcontractSampleApprovalStatusApproved ||
		loaded.DecisionBy != "qa-lead" ||
		!loaded.DecisionAt.Equal(approvedAt) ||
		loaded.StorageStatus != "retained_in_qa_cabinet" ||
		loaded.Version != 2 {
		t.Fatalf("approved sample approval = %+v, want persisted decision actor/status/version", loaded)
	}

	rejectedOrder := buildIssuedSubcontractOrderForSampleStore(t, suffix+"-reject")
	rejected, err := sampleBuilder.BuildSubmission(ctx, BuildSubcontractSampleSubmissionInput{
		ID:          "ssa-s16-04-01-reject-" + suffix,
		Order:       rejectedOrder,
		SampleCode:  "SCO-S16-04-01-SAMPLE-R-" + suffix,
		SubmittedBy: "factory-user",
		SubmittedAt: submittedAt.Add(2 * time.Hour),
		ActorID:     "qa-user",
		Evidence: []BuildSubcontractSampleEvidenceInput{{
			ID:           "ssa-evidence-s16-04-01-r-" + suffix,
			EvidenceType: "photo",
			ObjectKey:    "subcontract/ssa-s16-04-01/sample-reject.jpg",
		}},
	})
	if err != nil {
		t.Fatalf("build rejected sample submission: %v", err)
	}
	rejected, err = sampleBuilder.BuildRejection(ctx, BuildSubcontractSampleDecisionInput{
		Order:          rejected.UpdatedOrder,
		SampleApproval: rejected.SampleApproval,
		DecisionBy:     "qa-lead",
		DecisionAt:     submittedAt.Add(3 * time.Hour),
		Reason:         "shade does not match spec",
	})
	if err != nil {
		t.Fatalf("build sample rejection: %v", err)
	}
	if err := sampleStore.Save(ctx, rejected.SampleApproval); err != nil {
		t.Fatalf("save rejected sample approval: %v", err)
	}
	loaded, err = sampleStore.Get(ctx, rejected.SampleApproval.ID)
	if err != nil {
		t.Fatalf("get rejected sample approval: %v", err)
	}
	if loaded.Status != productiondomain.SubcontractSampleApprovalStatusRejected ||
		loaded.DecisionBy != "qa-lead" ||
		loaded.DecisionReason != "shade does not match spec" ||
		loaded.Version != 2 {
		t.Fatalf("rejected sample approval = %+v, want persisted reject reason and actor", loaded)
	}
}

func TestPostgresSubcontractSampleApprovalStoreRequiresDatabase(t *testing.T) {
	store := NewPostgresSubcontractSampleApprovalStore(nil, PostgresSubcontractSampleApprovalStoreConfig{})

	if err := store.Save(context.Background(), productiondomain.SubcontractSampleApproval{}); err == nil {
		t.Fatal("Save() error = nil, want database required error")
	}
	if _, err := store.Get(context.Background(), "sample-missing"); err == nil {
		t.Fatal("Get() error = nil, want database required error")
	}
	if _, err := store.GetLatestBySubcontractOrder(context.Background(), "sco-missing"); err == nil {
		t.Fatal("GetLatestBySubcontractOrder() error = nil, want database required error")
	}
}

func buildIssuedSubcontractOrderForSampleStore(t *testing.T, suffix string) productiondomain.SubcontractOrder {
	t.Helper()

	order := subcontractMaterialTransferTestOrder(t)
	order.ID = "sco-s16-04-01-" + suffix
	order.OrderNo = "SCO-S16-04-01-" + suffix
	order.MaterialLines[0].ID = "sco-mat-s16-04-01-a-" + suffix
	order.MaterialLines[1].ID = "sco-mat-s16-04-01-b-" + suffix
	if err := order.Validate(); err != nil {
		t.Fatalf("validate sample store test order: %v", err)
	}

	fixedNow := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)
	builder := NewSubcontractMaterialTransferService()
	builder.clock = func() time.Time { return fixedNow }
	buildResult, err := builder.BuildIssue(context.Background(), BuildSubcontractMaterialTransferInput{
		ID:                  "smt-s16-04-01-" + suffix,
		TransferNo:          "SMT-S16-04-01-" + suffix,
		Order:               order,
		SourceWarehouseID:   "wh-hcm-rm",
		SourceWarehouseCode: "WH-HCM-RM",
		HandoverBy:          "warehouse-user",
		HandoverAt:          fixedNow.Add(15 * time.Minute),
		ReceivedBy:          "factory-receiver",
		ReceiverContact:     "0988000111",
		ActorID:             "warehouse-user",
		Lines: []BuildSubcontractMaterialTransferLineInput{
			{
				ID:                  "smt-line-s16-04-01-a-" + suffix,
				OrderMaterialLineID: order.MaterialLines[0].ID,
				IssueQty:            "10",
				UOMCode:             "KG",
				BatchID:             "batch-base-s16-04-01",
				BatchNo:             "BASE-S16-04-01",
				SourceBinID:         "bin-rm-a01",
			},
			{
				ID:                  "smt-line-s16-04-01-b-" + suffix,
				OrderMaterialLineID: order.MaterialLines[1].ID,
				IssueQty:            "1000",
				UOMCode:             "PCS",
				SourceBinID:         "bin-pk-b01",
			},
		},
	})
	if err != nil {
		t.Fatalf("build issued sample store test order: %v", err)
	}

	return buildResult.UpdatedOrder
}
