package application

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestPostgresSubcontractPaymentMilestoneStorePersistsMilestones(t *testing.T) {
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
	order.ID = "sco-s16-07-01-" + suffix
	order.OrderNo = "SCO-S16-07-01-" + suffix
	order.MaterialLines[0].ID = "sco-mat-s16-07-01-a-" + suffix
	order.MaterialLines[1].ID = "sco-mat-s16-07-01-b-" + suffix
	if err := order.Validate(); err != nil {
		t.Fatalf("validate deposit test order: %v", err)
	}
	orderStore := NewPostgresSubcontractOrderStore(
		db,
		PostgresSubcontractOrderStoreConfig{DefaultOrgID: testSubcontractOrderOrgID},
	)
	if err := orderStore.WithinTx(ctx, func(txCtx context.Context, tx SubcontractOrderTx) error {
		return tx.Save(txCtx, order)
	}); err != nil {
		t.Fatalf("seed subcontract order: %v", err)
	}

	fixedNow := time.Date(2026, 5, 2, 16, 0, 0, 0, time.UTC)
	builder := NewSubcontractPaymentMilestoneService()
	builder.clock = func() time.Time { return fixedNow }
	deposit, err := builder.BuildDepositMilestone(ctx, BuildSubcontractDepositMilestoneInput{
		ID:          "spm-s16-07-01-deposit-" + suffix,
		MilestoneNo: "SPM-S16-07-01-DEPOSIT-" + suffix,
		Order:       order,
		Amount:      "1000000",
		RecordedBy:  "finance-user",
		RecordedAt:  fixedNow.Add(15 * time.Minute),
		ActorID:     "finance-user",
		Note:        "Deposit transfer confirmed by finance",
	})
	if err != nil {
		t.Fatalf("build deposit milestone: %v", err)
	}

	acceptedOrder := buildAcceptedSubcontractOrderForPaymentMilestoneStore(t, suffix+"-final", fixedNow)
	final, err := builder.BuildFinalPaymentMilestone(ctx, BuildSubcontractFinalPaymentMilestoneInput{
		ID:          "spm-s16-07-01-final-" + suffix,
		MilestoneNo: "SPM-S16-07-01-FINAL-" + suffix,
		Order:       acceptedOrder,
		ReadyBy:     "finance-user",
		ReadyAt:     fixedNow.Add(30 * time.Minute),
		ActorID:     "finance-user",
		Note:        "QC accepted and final payment can be prepared",
	})
	if err != nil {
		t.Fatalf("build final payment milestone: %v", err)
	}

	store := NewPostgresSubcontractPaymentMilestoneStore(
		db,
		PostgresSubcontractPaymentMilestoneStoreConfig{DefaultOrgID: testSubcontractOrderOrgID},
	)
	if err := store.Save(ctx, deposit.Milestone); err != nil {
		t.Fatalf("save deposit milestone: %v", err)
	}
	if err := store.Save(ctx, final.Milestone); err != nil {
		t.Fatalf("save final payment milestone: %v", err)
	}

	loadedDeposit, err := store.Get(ctx, deposit.Milestone.ID)
	if err != nil {
		t.Fatalf("get deposit milestone: %v", err)
	}
	if loadedDeposit.Status != productiondomain.SubcontractPaymentMilestoneStatusRecorded ||
		loadedDeposit.Kind != productiondomain.SubcontractPaymentMilestoneKindDeposit ||
		loadedDeposit.Amount.String() != "1000000.00" ||
		loadedDeposit.RecordedBy != "finance-user" ||
		!loadedDeposit.RecordedAt.Equal(fixedNow.Add(15*time.Minute)) {
		t.Fatalf("loaded deposit = %+v, want recorded deposit milestone", loadedDeposit)
	}
	loadedFinal, err := store.Get(ctx, final.Milestone.MilestoneNo)
	if err != nil {
		t.Fatalf("get final payment milestone: %v", err)
	}
	if loadedFinal.Status != productiondomain.SubcontractPaymentMilestoneStatusReady ||
		loadedFinal.Kind != productiondomain.SubcontractPaymentMilestoneKindFinalPayment ||
		loadedFinal.Amount != acceptedOrder.EstimatedCostAmount ||
		loadedFinal.ReadyBy != "finance-user" ||
		loadedFinal.BlocksFinalPayment() {
		t.Fatalf("loaded final milestone = %+v, want ready final payment milestone", loadedFinal)
	}
	milestones, err := store.ListBySubcontractOrder(ctx, order.OrderNo)
	if err != nil {
		t.Fatalf("list milestones: %v", err)
	}
	if len(milestones) != 1 || milestones[0].ID != deposit.Milestone.ID {
		t.Fatalf("deposit milestones = %+v, want milestone by order", milestones)
	}

	blocked, err := productiondomain.NewSubcontractPaymentMilestone(productiondomain.NewSubcontractPaymentMilestoneInput{
		ID:                 "spm-s16-07-01-blocked-" + suffix,
		OrgID:              acceptedOrder.OrgID,
		MilestoneNo:        "SPM-S16-07-01-BLOCKED-" + suffix,
		SubcontractOrderID: acceptedOrder.ID,
		SubcontractOrderNo: acceptedOrder.OrderNo,
		FactoryID:          acceptedOrder.FactoryID,
		FactoryCode:        acceptedOrder.FactoryCode,
		FactoryName:        acceptedOrder.FactoryName,
		Kind:               productiondomain.SubcontractPaymentMilestoneKindFinalPayment,
		Amount:             decimal.MustMoneyAmount("500000"),
		CurrencyCode:       "VND",
		CreatedAt:          fixedNow,
		CreatedBy:          "finance-user",
	})
	if err != nil {
		t.Fatalf("build pending blocked milestone: %v", err)
	}
	blocked, err = blocked.Block("finance-user", "open factory claim blocks final payment", fixedNow.Add(time.Hour))
	if err != nil {
		t.Fatalf("block milestone: %v", err)
	}
	if err := store.Save(ctx, blocked); err != nil {
		t.Fatalf("save blocked milestone: %v", err)
	}
	loadedBlocked, err := store.Get(ctx, blocked.ID)
	if err != nil {
		t.Fatalf("get blocked milestone: %v", err)
	}
	if loadedBlocked.Status != productiondomain.SubcontractPaymentMilestoneStatusBlocked ||
		loadedBlocked.BlockedBy != "finance-user" ||
		loadedBlocked.BlockReason != "open factory claim blocks final payment" ||
		!loadedBlocked.BlocksFinalPayment() {
		t.Fatalf("loaded blocked milestone = %+v, want persisted block state", loadedBlocked)
	}
}

func TestPostgresSubcontractPaymentMilestoneStoreRequiresDatabase(t *testing.T) {
	store := NewPostgresSubcontractPaymentMilestoneStore(nil, PostgresSubcontractPaymentMilestoneStoreConfig{})

	if err := store.Save(context.Background(), productiondomain.SubcontractPaymentMilestone{}); err == nil {
		t.Fatal("Save() error = nil, want database required error")
	}
	if _, err := store.Get(context.Background(), "spm-missing"); err == nil {
		t.Fatal("Get() error = nil, want database required error")
	}
	if _, err := store.ListBySubcontractOrder(context.Background(), "sco-missing"); err == nil {
		t.Fatal("ListBySubcontractOrder() error = nil, want database required error")
	}
}

func buildAcceptedSubcontractOrderForPaymentMilestoneStore(
	t *testing.T,
	suffix string,
	changedAt time.Time,
) productiondomain.SubcontractOrder {
	t.Helper()

	order := buildMassProductionSubcontractOrderForFinishedGoodsReceiptStore(t, suffix)
	received, err := order.MarkFinishedGoodsReceived("warehouse-user", changedAt.Add(-3*time.Hour))
	if err != nil {
		t.Fatalf("mark finished goods received: %v", err)
	}
	qc, err := received.StartQC("qa-user", changedAt.Add(-2*time.Hour))
	if err != nil {
		t.Fatalf("start qc: %v", err)
	}
	accepted, err := qc.Accept("qa-lead", changedAt.Add(-time.Hour))
	if err != nil {
		t.Fatalf("accept: %v", err)
	}

	return accepted
}
