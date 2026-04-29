package application

import (
	"context"
	"errors"
	"testing"
	"time"

	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
)

func TestSubcontractFactoryClaimServiceBuildClaimRejectsOrderAndBlocksFinalPayment(t *testing.T) {
	fixedNow := time.Date(2026, 4, 29, 15, 0, 0, 0, time.UTC)
	service := NewSubcontractFactoryClaimService()
	service.clock = func() time.Time { return fixedNow }
	order := subcontractFactoryClaimQCOrder(t, fixedNow)

	result, err := service.BuildClaim(context.Background(), BuildSubcontractFactoryClaimInput{
		ID:              "sfc_001",
		ClaimNo:         "SFC-20260429-001",
		Order:           order,
		ReceiptID:       "sfgr_001",
		ReceiptNo:       "SFGR-20260429-001",
		ReasonCode:      "packaging_damaged",
		Reason:          "Outer cartons crushed and bottle caps scratched",
		Severity:        "p1",
		AffectedQty:     "12",
		UOMCode:         "PCS",
		BaseAffectedQty: "12",
		BaseUOMCode:     "PCS",
		OwnerID:         "factory-owner",
		OpenedBy:        "qa-user",
		OpenedAt:        fixedNow,
		ActorID:         "qa-user",
		Evidence: []BuildSubcontractFactoryClaimEvidenceInput{
			{
				EvidenceType: "qc_photo",
				ObjectKey:    "subcontract/sfc_001/damaged-cartons.jpg",
			},
		},
	})
	if err != nil {
		t.Fatalf("build factory claim: %v", err)
	}

	if result.PreviousStatus != productiondomain.SubcontractOrderStatusQCInProgress ||
		result.CurrentStatus != productiondomain.SubcontractOrderStatusRejectedFactoryIssue ||
		result.UpdatedOrder.FactoryIssueReason != "Outer cartons crushed and bottle caps scratched" {
		t.Fatalf("result = %+v, want qc order rejected with factory issue", result)
	}
	if !result.Claim.BlocksFinalPayment() ||
		!result.Claim.DueAt.Equal(fixedNow.AddDate(0, 0, order.ClaimWindowDays)) {
		t.Fatalf("claim = %+v, want final payment block and claim window due date", result.Claim)
	}
	if _, err := result.UpdatedOrder.MarkFinalPaymentReady("finance-user", fixedNow.Add(time.Hour)); !errors.Is(err, productiondomain.ErrSubcontractOrderInvalidTransition) {
		t.Fatalf("final payment ready error = %v, want invalid transition from rejected factory issue", err)
	}
}

func TestSubcontractFactoryClaimServiceRejectsNonQCOrder(t *testing.T) {
	service := NewSubcontractFactoryClaimService()
	order := subcontractFinishedGoodsReceiptMassProductionOrder(t)

	_, err := service.BuildClaim(context.Background(), BuildSubcontractFactoryClaimInput{
		Order:       order,
		Reason:      "Wrong shade",
		AffectedQty: "10",
		ActorID:     "qa-user",
		Evidence: []BuildSubcontractFactoryClaimEvidenceInput{
			{EvidenceType: "qc_photo", ObjectKey: "subcontract/sfc_001/wrong-shade.jpg"},
		},
	})
	if !errors.Is(err, productiondomain.ErrSubcontractOrderInvalidTransition) {
		t.Fatalf("error = %v, want invalid transition before qc", err)
	}
}

func TestPrototypeSubcontractFactoryClaimStoreSavesAndListsByOrder(t *testing.T) {
	store := NewPrototypeSubcontractFactoryClaimStore()
	service := NewSubcontractFactoryClaimService()
	openedAt := time.Date(2026, 4, 29, 15, 0, 0, 0, time.UTC)
	result, err := service.BuildClaim(context.Background(), BuildSubcontractFactoryClaimInput{
		ID:          "sfc_store_001",
		Order:       subcontractFactoryClaimQCOrder(t, openedAt),
		Reason:      "Wrong inner packaging",
		AffectedQty: "4",
		ActorID:     "qa-user",
		OpenedAt:    openedAt,
		Evidence: []BuildSubcontractFactoryClaimEvidenceInput{
			{EvidenceType: "qc_photo", ObjectKey: "subcontract/sfc_store_001/wrong-packaging.jpg"},
		},
	})
	if err != nil {
		t.Fatalf("build claim: %v", err)
	}
	if err := store.Save(context.Background(), result.Claim); err != nil {
		t.Fatalf("save claim: %v", err)
	}

	saved, err := store.Get(context.Background(), result.Claim.ID)
	if err != nil {
		t.Fatalf("get claim: %v", err)
	}
	rows, err := store.ListBySubcontractOrder(context.Background(), result.Claim.SubcontractOrderID)
	if err != nil {
		t.Fatalf("list claims: %v", err)
	}
	if store.Count() != 1 || saved.ID != result.Claim.ID || len(rows) != 1 || rows[0].ID != result.Claim.ID {
		t.Fatalf("store count/saved/rows = %d/%+v/%+v, want saved claim", store.Count(), saved, rows)
	}
}

func subcontractFactoryClaimQCOrder(t *testing.T, changedAt time.Time) productiondomain.SubcontractOrder {
	t.Helper()

	order := subcontractFinishedGoodsReceiptMassProductionOrder(t)
	received, err := order.MarkFinishedGoodsReceived("warehouse-user", changedAt.Add(-2*time.Hour))
	if err != nil {
		t.Fatalf("mark finished goods received: %v", err)
	}
	qc, err := received.StartQC("qa-user", changedAt.Add(-time.Hour))
	if err != nil {
		t.Fatalf("start qc: %v", err)
	}

	return qc
}
