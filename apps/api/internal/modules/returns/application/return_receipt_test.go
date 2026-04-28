package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/returns/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

func TestListReturnReceiptsFiltersByWarehouseAndStatus(t *testing.T) {
	service := NewListReturnReceipts(NewPrototypeReturnReceiptStore())

	rows, err := service.Execute(context.Background(), domain.NewReturnReceiptFilter("wh-hcm", domain.ReturnStatusPendingInspection))
	if err != nil {
		t.Fatalf("list return receipts: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(rows))
	}
	if rows[0].ReceiptNo != "RR-260426-0001" {
		t.Fatalf("receipt no = %q, want RR-260426-0001", rows[0].ReceiptNo)
	}
}

func TestReceiveReturnCreatesKnownReceiptMovementAndAudit(t *testing.T) {
	store := NewPrototypeReturnReceiptStore()
	auditStore := audit.NewInMemoryLogStore()
	service := NewPrototypeReceiveReturnAt(
		store,
		auditStore,
		time.Date(2026, 4, 26, 10, 0, 0, 0, time.UTC),
	)

	result, err := service.Execute(context.Background(), ReceiveReturnInput{
		WarehouseID:      "wh-hcm",
		WarehouseCode:    "HCM",
		Source:           "CARRIER",
		ScanCode:         "GHN260426001",
		PackageCondition: "sealed",
		Disposition:      "reusable",
		ActorID:          "user-warehouse-lead",
		RequestID:        "req-return-receive",
	})
	if err != nil {
		t.Fatalf("receive return: %v", err)
	}
	if result.Receipt.OriginalOrderNo != "SO-260426-001" {
		t.Fatalf("order no = %q, want SO-260426-001", result.Receipt.OriginalOrderNo)
	}
	if result.Receipt.StockMovement == nil {
		t.Fatal("stock movement is nil")
	}
	if result.Receipt.StockMovement.MovementType != domain.ReturnReceiptMovementType {
		t.Fatalf("movement type = %q, want RETURN_RECEIPT", result.Receipt.StockMovement.MovementType)
	}
	if result.AuditLogID == "" {
		t.Fatal("audit log id is empty")
	}

	logs, err := auditStore.List(context.Background(), audit.Query{Action: "returns.receipt.created"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit logs = %d, want 1", len(logs))
	}
	if logs[0].AfterData["movement_type"] != domain.ReturnReceiptMovementType {
		t.Fatalf("audit after data = %+v, want RETURN_RECEIPT movement", logs[0].AfterData)
	}
}

func TestReceiveReturnCreatesUnknownCase(t *testing.T) {
	store := NewPrototypeReturnReceiptStore()
	service := NewPrototypeReceiveReturnAt(
		store,
		audit.NewInMemoryLogStore(),
		time.Date(2026, 4, 26, 10, 30, 0, 0, time.UTC),
	)

	result, err := service.Execute(context.Background(), ReceiveReturnInput{
		WarehouseID:      "wh-hcm",
		Source:           "SHIPPER",
		ScanCode:         "UNKNOWN-RETURN",
		PackageCondition: "damaged box",
		Disposition:      "needs_inspection",
		ActorID:          "user-warehouse-lead",
		RequestID:        "req-return-unknown",
	})
	if err != nil {
		t.Fatalf("receive return: %v", err)
	}

	if !result.Receipt.UnknownCase {
		t.Fatal("unknown case = false, want true")
	}
	if result.Receipt.TargetLocation != "return-inspection-queue" {
		t.Fatalf("target location = %q, want return-inspection-queue", result.Receipt.TargetLocation)
	}
	if result.Receipt.StockMovement != nil {
		t.Fatalf("stock movement = %+v, want nil", result.Receipt.StockMovement)
	}
}

func TestReceiveReturnRoutesNotReusableToLabPlaceholder(t *testing.T) {
	store := NewPrototypeReturnReceiptStore()
	service := NewPrototypeReceiveReturnAt(
		store,
		audit.NewInMemoryLogStore(),
		time.Date(2026, 4, 26, 11, 0, 0, 0, time.UTC),
	)

	result, err := service.Execute(context.Background(), ReceiveReturnInput{
		WarehouseID:      "wh-hcm",
		ScanCode:         "SO-260426-004",
		PackageCondition: "leaking",
		Disposition:      "not_reusable",
		ActorID:          "user-warehouse-lead",
		RequestID:        "req-return-damaged",
	})
	if err != nil {
		t.Fatalf("receive return: %v", err)
	}

	if result.Receipt.TargetLocation != "lab-damaged-placeholder" {
		t.Fatalf("target location = %q, want lab-damaged-placeholder", result.Receipt.TargetLocation)
	}
	if result.Receipt.StockMovement != nil {
		t.Fatalf("stock movement = %+v, want nil", result.Receipt.StockMovement)
	}
}

func TestReceiveReturnRejectsOrderBeforeHandoverOrDelivery(t *testing.T) {
	store := NewPrototypeReturnReceiptStore()
	service := NewPrototypeReceiveReturnAt(
		store,
		audit.NewInMemoryLogStore(),
		time.Date(2026, 4, 26, 11, 30, 0, 0, time.UTC),
	)

	_, err := service.Execute(context.Background(), ReceiveReturnInput{
		WarehouseID:      "wh-hcm",
		ScanCode:         "GHN260426009",
		PackageCondition: "sealed",
		Disposition:      "needs_inspection",
		ActorID:          "user-warehouse-lead",
		RequestID:        "req-return-not-eligible",
	})
	if !errors.Is(err, ErrExpectedReturnOrderNotReturnable) {
		t.Fatalf("error = %v, want order status not returnable", err)
	}
}

func TestReceiveReturnRejectsDuplicateKnownReturn(t *testing.T) {
	store := NewPrototypeReturnReceiptStore()
	service := NewPrototypeReceiveReturnAt(
		store,
		audit.NewInMemoryLogStore(),
		time.Date(2026, 4, 26, 11, 30, 0, 0, time.UTC),
	)

	input := ReceiveReturnInput{
		WarehouseID:      "wh-hcm",
		ScanCode:         "GHN260426001",
		PackageCondition: "sealed",
		Disposition:      "needs_inspection",
		ActorID:          "user-warehouse-lead",
		RequestID:        "req-return-first-scan",
	}
	if _, err := service.Execute(context.Background(), input); err != nil {
		t.Fatalf("first scan failed: %v", err)
	}

	input.ScanCode = "SO-260426-001"
	input.RequestID = "req-return-duplicate-scan"
	_, err := service.Execute(context.Background(), input)
	if !errors.Is(err, ErrReturnReceiptDuplicate) {
		t.Fatalf("error = %v, want duplicate return receipt", err)
	}
}
