package application

import (
	"context"
	"errors"
	"testing"
	"time"

	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
)

func TestSubcontractPaymentMilestoneServiceBuildDepositRecordsOrderAndMilestone(t *testing.T) {
	fixedNow := time.Date(2026, 4, 29, 16, 0, 0, 0, time.UTC)
	service := NewSubcontractPaymentMilestoneService()
	service.clock = func() time.Time { return fixedNow }
	order := subcontractMaterialTransferTestOrder(t)

	result, err := service.BuildDepositMilestone(context.Background(), BuildSubcontractDepositMilestoneInput{
		ID:          "spm_deposit_001",
		MilestoneNo: "SPM-DEPOSIT-001",
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

	if result.PreviousStatus != productiondomain.SubcontractOrderStatusFactoryConfirmed ||
		result.CurrentStatus != productiondomain.SubcontractOrderStatusDepositRecorded ||
		result.UpdatedOrder.DepositAmount.String() != "1000000.00" ||
		result.Milestone.Status != productiondomain.SubcontractPaymentMilestoneStatusRecorded ||
		result.Milestone.Kind != productiondomain.SubcontractPaymentMilestoneKindDeposit {
		t.Fatalf("result = %+v, want recorded deposit milestone and order deposit status", result)
	}
}

func TestSubcontractPaymentMilestoneServiceBuildFinalPaymentReadyAfterAcceptance(t *testing.T) {
	fixedNow := time.Date(2026, 4, 29, 17, 0, 0, 0, time.UTC)
	service := NewSubcontractPaymentMilestoneService()
	service.clock = func() time.Time { return fixedNow }
	order := subcontractPaymentAcceptedOrder(t)

	result, err := service.BuildFinalPaymentMilestone(context.Background(), BuildSubcontractFinalPaymentMilestoneInput{
		ID:          "spm_final_001",
		MilestoneNo: "SPM-FINAL-001",
		Order:       order,
		ReadyBy:     "finance-user",
		ReadyAt:     fixedNow.Add(30 * time.Minute),
		ActorID:     "finance-user",
		Note:        "QC accepted and final payment can be prepared",
	})
	if err != nil {
		t.Fatalf("build final payment milestone: %v", err)
	}

	if result.PreviousStatus != productiondomain.SubcontractOrderStatusAccepted ||
		result.CurrentStatus != productiondomain.SubcontractOrderStatusFinalPaymentReady ||
		result.UpdatedOrder.FinalPaymentReadyBy != "finance-user" ||
		result.Milestone.Status != productiondomain.SubcontractPaymentMilestoneStatusReady ||
		result.Milestone.Kind != productiondomain.SubcontractPaymentMilestoneKindFinalPayment ||
		result.Milestone.Amount != order.EstimatedCostAmount ||
		result.Milestone.BlocksFinalPayment() {
		t.Fatalf("result = %+v, want ready final payment milestone after accepted order", result)
	}
}

func TestSubcontractPaymentMilestoneServiceRejectsFinalPaymentWithBlockingClaim(t *testing.T) {
	service := NewSubcontractPaymentMilestoneService()
	order := subcontractPaymentAcceptedOrder(t)

	_, err := service.BuildFinalPaymentMilestone(context.Background(), BuildSubcontractFinalPaymentMilestoneInput{
		Order:   order,
		ActorID: "finance-user",
		BlockingClaims: []productiondomain.SubcontractFactoryClaim{
			{Status: productiondomain.SubcontractFactoryClaimStatusOpen},
		},
	})
	if !errors.Is(err, productiondomain.ErrSubcontractPaymentMilestoneBlocked) {
		t.Fatalf("error = %v, want payment milestone blocked", err)
	}
}

func TestSubcontractPaymentMilestoneServiceRejectsFinalPaymentBeforeAcceptance(t *testing.T) {
	service := NewSubcontractPaymentMilestoneService()
	order := subcontractFinishedGoodsReceiptMassProductionOrder(t)

	_, err := service.BuildFinalPaymentMilestone(context.Background(), BuildSubcontractFinalPaymentMilestoneInput{
		Order:   order,
		ActorID: "finance-user",
	})
	if !errors.Is(err, productiondomain.ErrSubcontractOrderInvalidTransition) {
		t.Fatalf("error = %v, want invalid transition before acceptance", err)
	}
}

func TestPrototypeSubcontractPaymentMilestoneStoreSavesAndListsByOrder(t *testing.T) {
	store := NewPrototypeSubcontractPaymentMilestoneStore()
	service := NewSubcontractPaymentMilestoneService()
	order := subcontractMaterialTransferTestOrder(t)
	result, err := service.BuildDepositMilestone(context.Background(), BuildSubcontractDepositMilestoneInput{
		ID:      "spm_store_001",
		Order:   order,
		Amount:  "1000000",
		ActorID: "finance-user",
	})
	if err != nil {
		t.Fatalf("build deposit milestone: %v", err)
	}
	if err := store.Save(context.Background(), result.Milestone); err != nil {
		t.Fatalf("save milestone: %v", err)
	}

	rows, err := store.ListBySubcontractOrder(context.Background(), order.ID)
	if err != nil {
		t.Fatalf("list milestones: %v", err)
	}
	if len(rows) != 1 || rows[0].ID != result.Milestone.ID {
		t.Fatalf("rows = %+v, want saved milestone for subcontract order", rows)
	}
	if store.Count() != 1 {
		t.Fatalf("count = %d, want 1", store.Count())
	}
}

func subcontractPaymentAcceptedOrder(t *testing.T) productiondomain.SubcontractOrder {
	t.Helper()

	order := subcontractFinishedGoodsReceiptMassProductionOrder(t)
	received, err := order.MarkFinishedGoodsReceived("warehouse-user", time.Now())
	if err != nil {
		t.Fatalf("mark finished goods received: %v", err)
	}
	qc, err := received.StartQC("qa-user", time.Now())
	if err != nil {
		t.Fatalf("start qc: %v", err)
	}
	accepted, err := qc.Accept("qa-lead", time.Now())
	if err != nil {
		t.Fatalf("accept: %v", err)
	}

	return accepted
}
