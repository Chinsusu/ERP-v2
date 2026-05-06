package application

import (
	"context"
	"errors"
	"testing"
	"time"

	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestSubcontractOrderServiceRecordDepositPersistsMilestoneOrderAndAudit(t *testing.T) {
	ctx := context.Background()
	auditStore := audit.NewInMemoryLogStore()
	orderStore := NewPrototypeSubcontractOrderStore(auditStore)
	milestoneStore := NewPrototypeSubcontractPaymentMilestoneStore()
	service := SubcontractOrderService{
		store:                 orderStore,
		paymentMilestoneStore: milestoneStore,
		paymentMilestoneBuild: NewSubcontractPaymentMilestoneService(),
	}
	recordedAt := time.Date(2026, 4, 29, 16, 30, 0, 0, time.UTC)
	order := subcontractMaterialTransferTestOrder(t)
	if err := orderStore.WithinTx(ctx, func(txCtx context.Context, tx SubcontractOrderTx) error {
		return tx.Save(txCtx, order)
	}); err != nil {
		t.Fatalf("seed order: %v", err)
	}

	result, err := service.RecordSubcontractDeposit(ctx, RecordSubcontractDepositInput{
		ID:              order.ID,
		ExpectedVersion: order.Version,
		MilestoneID:     "spm_service_deposit_001",
		MilestoneNo:     "SPM-SERVICE-DEPOSIT-001",
		Amount:          "1000000",
		RecordedBy:      "finance-user",
		RecordedAt:      recordedAt,
		ActorID:         "finance-user",
		RequestID:       "req-record-deposit",
		Note:            "Deposit transfer confirmed",
	})
	if err != nil {
		t.Fatalf("record deposit: %v", err)
	}

	if result.PreviousStatus != productiondomain.SubcontractOrderStatusFactoryConfirmed ||
		result.CurrentStatus != productiondomain.SubcontractOrderStatusDepositRecorded ||
		result.SubcontractOrder.DepositAmount.String() != "1000000.00" ||
		result.Milestone.Status != productiondomain.SubcontractPaymentMilestoneStatusRecorded ||
		result.AuditLogID == "" {
		t.Fatalf("result = %+v, want recorded deposit milestone and order audit", result)
	}
	if milestoneStore.Count() != 1 {
		t.Fatalf("payment milestone count = %d, want 1", milestoneStore.Count())
	}
	logs, err := auditStore.List(ctx, audit.Query{Action: subcontractDepositRecordedAction})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 ||
		logs[0].AfterData["payment_milestone_status"] != "recorded" ||
		logs[0].AfterData["deposit_amount"] != "1000000.00" {
		t.Fatalf("audit logs = %+v, want deposit milestone audit", logs)
	}
}

func TestSubcontractOrderServiceMarkFinalPaymentReadyPersistsMilestoneOrderAndAudit(t *testing.T) {
	ctx := context.Background()
	auditStore := audit.NewInMemoryLogStore()
	orderStore := NewPrototypeSubcontractOrderStore(auditStore)
	milestoneStore := NewPrototypeSubcontractPaymentMilestoneStore()
	claimStore := NewPrototypeSubcontractFactoryClaimStore()
	service := SubcontractOrderService{
		store:                 orderStore,
		factoryClaimStore:     claimStore,
		paymentMilestoneStore: milestoneStore,
		paymentMilestoneBuild: NewSubcontractPaymentMilestoneService(),
	}
	readyAt := time.Date(2026, 4, 29, 18, 0, 0, 0, time.UTC)
	order := subcontractPaymentAcceptedOrder(t)
	if err := orderStore.WithinTx(ctx, func(txCtx context.Context, tx SubcontractOrderTx) error {
		return tx.Save(txCtx, order)
	}); err != nil {
		t.Fatalf("seed order: %v", err)
	}

	result, err := service.MarkSubcontractFinalPaymentReady(ctx, MarkSubcontractFinalPaymentReadyInput{
		ID:              order.ID,
		ExpectedVersion: order.Version,
		MilestoneID:     "spm_service_final_001",
		MilestoneNo:     "SPM-SERVICE-FINAL-001",
		ReadyBy:         "finance-user",
		ReadyAt:         readyAt,
		ActorID:         "finance-user",
		RequestID:       "req-final-payment-ready",
		Note:            "Accepted goods cleared for final payment",
	})
	if err != nil {
		t.Fatalf("mark final payment ready: %v", err)
	}

	if result.PreviousStatus != productiondomain.SubcontractOrderStatusAccepted ||
		result.CurrentStatus != productiondomain.SubcontractOrderStatusFinalPaymentReady ||
		result.Milestone.Status != productiondomain.SubcontractPaymentMilestoneStatusReady ||
		result.Milestone.Amount.String() != order.EstimatedCostAmount.String() ||
		result.AuditLogID == "" {
		t.Fatalf("result = %+v, want final payment ready milestone and audit", result)
	}
	if milestoneStore.Count() != 1 {
		t.Fatalf("payment milestone count = %d, want 1", milestoneStore.Count())
	}
	logs, err := auditStore.List(ctx, audit.Query{Action: subcontractFinalPaymentAction})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 ||
		logs[0].AfterData["payment_milestone_status"] != "ready" ||
		logs[0].AfterData["blocking_claim_count"] != 0 {
		t.Fatalf("audit logs = %+v, want final payment milestone audit", logs)
	}
}

func TestSubcontractOrderServiceMarkFinalPaymentReadyCreatesSupplierPayable(t *testing.T) {
	ctx := context.Background()
	auditStore := audit.NewInMemoryLogStore()
	orderStore := NewPrototypeSubcontractOrderStore(auditStore)
	milestoneStore := NewPrototypeSubcontractPaymentMilestoneStore()
	claimStore := NewPrototypeSubcontractFactoryClaimStore()
	payableCreator := &recordingSubcontractPayableCreator{
		result: SubcontractPayableCreationResult{
			PayableID:  "ap-spm-service-final-002",
			PayableNo:  "AP-SPM-SERVICE-FINAL-002",
			AuditLogID: "audit-ap-spm-service-final-002",
		},
	}
	service := SubcontractOrderService{
		store:                 orderStore,
		factoryClaimStore:     claimStore,
		paymentMilestoneStore: milestoneStore,
		paymentMilestoneBuild: NewSubcontractPaymentMilestoneService(),
		payableCreator:        payableCreator,
	}
	order := subcontractPaymentAcceptedOrder(t)
	if err := orderStore.WithinTx(ctx, func(txCtx context.Context, tx SubcontractOrderTx) error {
		return tx.Save(txCtx, order)
	}); err != nil {
		t.Fatalf("seed order: %v", err)
	}

	result, err := service.MarkSubcontractFinalPaymentReady(ctx, MarkSubcontractFinalPaymentReadyInput{
		ID:              order.ID,
		ExpectedVersion: order.Version,
		MilestoneID:     "spm_service_final_002",
		MilestoneNo:     "SPM-SERVICE-FINAL-002",
		ReadyBy:         "finance-user",
		ActorID:         "finance-user",
		RequestID:       "req-final-payment-payable",
	})
	if err != nil {
		t.Fatalf("mark final payment ready: %v", err)
	}

	if payableCreator.count != 1 {
		t.Fatalf("payable creator count = %d, want 1", payableCreator.count)
	}
	input := payableCreator.input
	if input.SubcontractOrder.ID != order.ID ||
		input.Milestone.ID != "spm_service_final_002" ||
		input.Milestone.Amount.String() != order.EstimatedCostAmount.String() ||
		input.ActorID != "finance-user" ||
		input.RequestID != "req-final-payment-payable" {
		t.Fatalf("payable input = %+v, want subcontract final payment milestone source", input)
	}
	if result.SupplierPayable.PayableID != "ap-spm-service-final-002" ||
		result.SupplierPayable.AuditLogID != "audit-ap-spm-service-final-002" {
		t.Fatalf("supplier payable result = %+v, want payable creation details", result.SupplierPayable)
	}
	logs, err := auditStore.List(ctx, audit.Query{Action: subcontractFinalPaymentAction})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 ||
		logs[0].AfterData["supplier_payable_id"] != "ap-spm-service-final-002" ||
		logs[0].AfterData["supplier_payable_no"] != "AP-SPM-SERVICE-FINAL-002" {
		t.Fatalf("audit logs = %+v, want final payment audit linked to supplier payable", logs)
	}
}

func TestSubcontractOrderServiceBlocksFinalPaymentWithOpenClaim(t *testing.T) {
	ctx := context.Background()
	auditStore := audit.NewInMemoryLogStore()
	orderStore := NewPrototypeSubcontractOrderStore(auditStore)
	milestoneStore := NewPrototypeSubcontractPaymentMilestoneStore()
	claimStore := NewPrototypeSubcontractFactoryClaimStore()
	service := SubcontractOrderService{
		store:                 orderStore,
		factoryClaimStore:     claimStore,
		paymentMilestoneStore: milestoneStore,
		paymentMilestoneBuild: NewSubcontractPaymentMilestoneService(),
	}
	order := subcontractPaymentAcceptedOrder(t)
	if err := orderStore.WithinTx(ctx, func(txCtx context.Context, tx SubcontractOrderTx) error {
		return tx.Save(txCtx, order)
	}); err != nil {
		t.Fatalf("seed order: %v", err)
	}
	claim := subcontractPaymentBlockingClaim(t, order)
	if err := claimStore.Save(ctx, claim); err != nil {
		t.Fatalf("seed claim: %v", err)
	}

	_, err := service.MarkSubcontractFinalPaymentReady(ctx, MarkSubcontractFinalPaymentReadyInput{
		ID:              order.ID,
		ExpectedVersion: order.Version,
		ActorID:         "finance-user",
		RequestID:       "req-final-payment-blocked",
	})
	if !errors.Is(err, productiondomain.ErrSubcontractPaymentMilestoneBlocked) {
		t.Fatalf("error = %v, want payment milestone blocked", err)
	}
	if milestoneStore.Count() != 0 {
		t.Fatalf("payment milestone count = %d, want 0", milestoneStore.Count())
	}
	logs, err := auditStore.List(ctx, audit.Query{Action: subcontractFinalPaymentAction})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 0 {
		t.Fatalf("final payment audit logs = %+v, want none", logs)
	}
}

func TestSubcontractOrderServiceResolvesFactoryClaimBeforeFinalPayment(t *testing.T) {
	ctx := context.Background()
	auditStore := audit.NewInMemoryLogStore()
	orderStore := NewPrototypeSubcontractOrderStore(auditStore)
	milestoneStore := NewPrototypeSubcontractPaymentMilestoneStore()
	claimStore := NewPrototypeSubcontractFactoryClaimStore()
	service := SubcontractOrderService{
		store:                 orderStore,
		factoryClaimStore:     claimStore,
		paymentMilestoneStore: milestoneStore,
		paymentMilestoneBuild: NewSubcontractPaymentMilestoneService(),
	}
	order := subcontractPaymentAcceptedOrder(t)
	if err := orderStore.WithinTx(ctx, func(txCtx context.Context, tx SubcontractOrderTx) error {
		return tx.Save(txCtx, order)
	}); err != nil {
		t.Fatalf("seed order: %v", err)
	}
	claim := subcontractPaymentBlockingClaim(t, order)
	if err := claimStore.Save(ctx, claim); err != nil {
		t.Fatalf("seed claim: %v", err)
	}
	resolvedAt := time.Date(2026, 4, 30, 9, 0, 0, 0, time.UTC)

	claimResult, err := service.ResolveSubcontractFactoryClaim(ctx, ResolveSubcontractFactoryClaimInput{
		ID:              claim.ID,
		ExpectedVersion: claim.Version,
		ResolvedBy:      "factory-owner",
		ResolvedAt:      resolvedAt,
		ResolutionNote:  "Factory accepted a credit memo for rejected goods",
		ActorID:         "qa-user",
		RequestID:       "req-factory-claim-resolved",
	})
	if err != nil {
		t.Fatalf("resolve factory claim: %v", err)
	}
	if claimResult.Claim.Status != productiondomain.SubcontractFactoryClaimStatusResolved ||
		claimResult.Claim.BlocksFinalPayment() ||
		claimResult.AuditLogID == "" {
		t.Fatalf("claim result = %+v, want resolved non-blocking claim and audit", claimResult)
	}

	result, err := service.MarkSubcontractFinalPaymentReady(ctx, MarkSubcontractFinalPaymentReadyInput{
		ID:              order.ID,
		ExpectedVersion: order.Version,
		MilestoneID:     "spm_service_final_after_claim_001",
		MilestoneNo:     "SPM-SERVICE-FINAL-AFTER-CLAIM-001",
		ReadyBy:         "finance-user",
		ActorID:         "finance-user",
		RequestID:       "req-final-payment-after-claim",
	})
	if err != nil {
		t.Fatalf("mark final payment after resolved claim: %v", err)
	}
	if result.CurrentStatus != productiondomain.SubcontractOrderStatusFinalPaymentReady ||
		result.Milestone.Status != productiondomain.SubcontractPaymentMilestoneStatusReady {
		t.Fatalf("result = %+v, want final payment ready after resolved claim", result)
	}
	logs, err := auditStore.List(ctx, audit.Query{Action: subcontractFactoryClaimResolvedAction})
	if err != nil {
		t.Fatalf("list claim audit logs: %v", err)
	}
	if len(logs) != 1 ||
		logs[0].AfterData["factory_claim_status"] != "resolved" ||
		logs[0].AfterData["resolution_note"] != "Factory accepted a credit memo for rejected goods" ||
		logs[0].AfterData["blocks_final_payment"] != false {
		t.Fatalf("claim audit logs = %+v, want resolved claim audit", logs)
	}
}

type recordingSubcontractPayableCreator struct {
	input  CreateSubcontractPayableInput
	result SubcontractPayableCreationResult
	count  int
}

func (c *recordingSubcontractPayableCreator) CreateSubcontractPayable(
	_ context.Context,
	input CreateSubcontractPayableInput,
) (SubcontractPayableCreationResult, error) {
	c.input = input
	c.count++

	return c.result, nil
}

func subcontractPaymentBlockingClaim(
	t *testing.T,
	order productiondomain.SubcontractOrder,
) productiondomain.SubcontractFactoryClaim {
	t.Helper()

	openedAt := time.Date(2026, 4, 29, 17, 0, 0, 0, time.UTC)
	claim, err := productiondomain.NewSubcontractFactoryClaim(productiondomain.NewSubcontractFactoryClaimInput{
		ID:                 "sfc_payment_block_001",
		OrgID:              order.OrgID,
		ClaimNo:            "SFC-PAYMENT-BLOCK-001",
		SubcontractOrderID: order.ID,
		SubcontractOrderNo: order.OrderNo,
		FactoryID:          order.FactoryID,
		FactoryCode:        order.FactoryCode,
		FactoryName:        order.FactoryName,
		ReceiptID:          "sfgr-payment-block-001",
		ReceiptNo:          "SFGR-PAYMENT-BLOCK-001",
		ReasonCode:         "QUALITY_FAIL",
		Reason:             "Open quality issue before final payment",
		Severity:           "P1",
		AffectedQty:        decimal.MustQuantity("5"),
		UOMCode:            order.UOMCode.String(),
		BaseAffectedQty:    decimal.MustQuantity("5"),
		BaseUOMCode:        order.BaseUOMCode.String(),
		Evidence: []productiondomain.NewSubcontractFactoryClaimEvidenceInput{
			{
				ID:           "sfc-payment-block-evidence-001",
				EvidenceType: "qc_photo",
				ObjectKey:    "subcontract/sfc_payment_block_001/qc-photo.jpg",
			},
		},
		OwnerID:   "factory-owner",
		OpenedBy:  "qa-user",
		OpenedAt:  openedAt,
		DueAt:     openedAt.AddDate(0, 0, 7),
		CreatedAt: openedAt,
		CreatedBy: "qa-user",
		UpdatedAt: openedAt,
		UpdatedBy: "qa-user",
	})
	if err != nil {
		t.Fatalf("new blocking claim: %v", err)
	}

	return claim
}
