package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/returns/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

func TestInspectReturnRecordsReusableInspectionAndAudit(t *testing.T) {
	store := NewPrototypeReturnReceiptStore()
	auditStore := audit.NewInMemoryLogStore()
	service := NewPrototypeInspectReturnAt(
		store,
		auditStore,
		time.Date(2026, 4, 26, 11, 0, 0, 0, time.UTC),
	)

	result, err := service.Execute(context.Background(), InspectReturnInput{
		ReceiptID:   "rr-260426-0001",
		Condition:   "intact",
		Disposition: "reusable",
		ActorID:     "user-return-inspector",
		RequestID:   "req-return-inspect",
	})
	if err != nil {
		t.Fatalf("inspect return: %v", err)
	}

	if result.Receipt.Status != domain.ReturnStatusInspected {
		t.Fatalf("receipt status = %q, want inspected", result.Receipt.Status)
	}
	if result.Inspection.Status != domain.ReturnInspectionStatusRecorded {
		t.Fatalf("inspection status = %q, want inspection_recorded", result.Inspection.Status)
	}
	if result.Receipt.TargetLocation != "return-area-qc-release" {
		t.Fatalf("target location = %q, want return-area-qc-release", result.Receipt.TargetLocation)
	}
	if result.Receipt.StockMovement != nil {
		t.Fatalf("stock movement = %+v, want nil until disposition action", result.Receipt.StockMovement)
	}
	if result.AuditLogID == "" {
		t.Fatal("audit log id is empty")
	}

	logs, err := auditStore.List(context.Background(), audit.Query{Action: "returns.receipt.inspected"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit logs = %d, want 1", len(logs))
	}
	if logs[0].AfterData["condition"] != "intact" || logs[0].AfterData["disposition"] != "reusable" {
		t.Fatalf("audit after data = %+v, want condition and disposition", logs[0].AfterData)
	}
}

func TestInspectReturnRoutesDamagedItemToLabPlaceholder(t *testing.T) {
	store := NewPrototypeReturnReceiptStore()
	service := NewPrototypeInspectReturnAt(
		store,
		audit.NewInMemoryLogStore(),
		time.Date(2026, 4, 26, 11, 5, 0, 0, time.UTC),
	)

	result, err := service.Execute(context.Background(), InspectReturnInput{
		ReceiptID:   "rr-260426-0001",
		Condition:   "damaged",
		Disposition: "not_reusable",
		ActorID:     "user-return-inspector",
		RequestID:   "req-return-inspect-damaged",
	})
	if err != nil {
		t.Fatalf("inspect return: %v", err)
	}

	if result.Inspection.RiskLevel != "high" {
		t.Fatalf("risk level = %q, want high", result.Inspection.RiskLevel)
	}
	if result.Receipt.TargetLocation != "lab-damaged-placeholder" {
		t.Fatalf("target location = %q, want lab-damaged-placeholder", result.Receipt.TargetLocation)
	}
}

func TestInspectReturnRoutesNeedQAItemToHold(t *testing.T) {
	store := NewPrototypeReturnReceiptStore()
	service := NewPrototypeInspectReturnAt(
		store,
		audit.NewInMemoryLogStore(),
		time.Date(2026, 4, 26, 11, 10, 0, 0, time.UTC),
	)

	result, err := service.Execute(context.Background(), InspectReturnInput{
		ReceiptID:     "rr-260426-0001",
		Condition:     "seal_torn",
		Disposition:   "needs_inspection",
		Note:          "seal torn",
		EvidenceLabel: "photo-001",
		ActorID:       "user-return-inspector",
		RequestID:     "req-return-inspect-qa",
	})
	if err != nil {
		t.Fatalf("inspect return: %v", err)
	}

	if result.Inspection.Status != domain.ReturnInspectionStatusQAHold {
		t.Fatalf("inspection status = %q, want return_qa_hold", result.Inspection.Status)
	}
	if result.Receipt.TargetLocation != "return-qa-hold" {
		t.Fatalf("target location = %q, want return-qa-hold", result.Receipt.TargetLocation)
	}
	if result.Inspection.Note == "" || result.Inspection.EvidenceLabel == "" {
		t.Fatalf("inspection = %+v, want note and evidence label", result.Inspection)
	}
}

func TestInspectReturnRejectsUnknownInvalidAndAlreadyInspectedReceipts(t *testing.T) {
	store := NewPrototypeReturnReceiptStore()
	service := NewPrototypeInspectReturnAt(
		store,
		audit.NewInMemoryLogStore(),
		time.Date(2026, 4, 26, 11, 15, 0, 0, time.UTC),
	)

	_, err := service.Execute(context.Background(), InspectReturnInput{
		ReceiptID:   "missing-receipt",
		Condition:   "intact",
		Disposition: "reusable",
		ActorID:     "user-return-inspector",
	})
	if !errors.Is(err, ErrReturnReceiptNotFound) {
		t.Fatalf("error = %v, want receipt not found", err)
	}

	_, err = service.Execute(context.Background(), InspectReturnInput{
		ReceiptID:   "rr-260426-0001",
		Condition:   "sealed_good",
		Disposition: "reusable",
		ActorID:     "user-return-inspector",
	})
	if !errors.Is(err, domain.ErrReturnInspectionInvalidCondition) {
		t.Fatalf("error = %v, want invalid condition", err)
	}

	if _, err := service.Execute(context.Background(), InspectReturnInput{
		ReceiptID:   "rr-260426-0001",
		Condition:   "intact",
		Disposition: "reusable",
		ActorID:     "user-return-inspector",
	}); err != nil {
		t.Fatalf("first inspection failed: %v", err)
	}
	_, err = service.Execute(context.Background(), InspectReturnInput{
		ReceiptID:   "rr-260426-0001",
		Condition:   "damaged",
		Disposition: "not_reusable",
		ActorID:     "user-return-inspector",
	})
	if !errors.Is(err, ErrReturnReceiptNotInspectable) {
		t.Fatalf("error = %v, want not inspectable", err)
	}
}
