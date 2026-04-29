package application

import (
	"context"
	"testing"
	"time"

	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

func TestSubcontractOrderServiceCreateFactoryClaimPersistsClaimOrderAndAudit(t *testing.T) {
	ctx := context.Background()
	auditStore := audit.NewInMemoryLogStore()
	orderStore := NewPrototypeSubcontractOrderStore(auditStore)
	claimStore := NewPrototypeSubcontractFactoryClaimStore()
	service := SubcontractOrderService{
		store:             orderStore,
		factoryClaimStore: claimStore,
		factoryClaimBuild: NewSubcontractFactoryClaimService(),
	}
	openedAt := time.Date(2026, 4, 29, 15, 0, 0, 0, time.UTC)
	order := subcontractFactoryClaimQCOrder(t, openedAt)
	if err := orderStore.WithinTx(ctx, func(txCtx context.Context, tx SubcontractOrderTx) error {
		return tx.Save(txCtx, order)
	}); err != nil {
		t.Fatalf("seed order: %v", err)
	}

	result, err := service.CreateSubcontractFactoryClaim(ctx, CreateSubcontractFactoryClaimInput{
		ID:              order.ID,
		ExpectedVersion: order.Version,
		ClaimID:         "sfc_service_001",
		ClaimNo:         "SFC-SERVICE-001",
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
		OpenedAt:        openedAt,
		ActorID:         "qa-user",
		RequestID:       "req-factory-claim",
		Evidence: []CreateSubcontractFactoryClaimEvidenceInput{
			{
				EvidenceType: "qc_photo",
				ObjectKey:    "subcontract/sfc_service_001/damaged-cartons.jpg",
			},
		},
	})
	if err != nil {
		t.Fatalf("create factory claim: %v", err)
	}

	if result.SubcontractOrder.Status != productiondomain.SubcontractOrderStatusRejectedFactoryIssue ||
		result.Claim.Status != productiondomain.SubcontractFactoryClaimStatusOpen ||
		!result.Claim.BlocksFinalPayment() ||
		result.AuditLogID == "" {
		t.Fatalf("result = %+v, want rejected order, open blocking claim, audit", result)
	}
	if claimStore.Count() != 1 {
		t.Fatalf("claim count = %d, want 1", claimStore.Count())
	}
	savedOrder, err := orderStore.Get(ctx, order.ID)
	if err != nil {
		t.Fatalf("get order: %v", err)
	}
	if savedOrder.Status != productiondomain.SubcontractOrderStatusRejectedFactoryIssue ||
		savedOrder.FactoryIssueReason != "Outer cartons crushed and bottle caps scratched" {
		t.Fatalf("saved order = %+v, want rejected with factory issue reason", savedOrder)
	}
	logs, err := auditStore.List(ctx, audit.Query{Action: subcontractFactoryClaimAction})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 ||
		logs[0].AfterData["factory_claim_status"] != "open" ||
		logs[0].AfterData["blocks_final_payment"] != true {
		t.Fatalf("audit logs = %+v, want factory claim audit", logs)
	}
}
