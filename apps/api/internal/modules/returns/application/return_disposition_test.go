package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/returns/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

func TestApplyReturnDispositionRoutesReusableToPutawayAndAudit(t *testing.T) {
	store := NewPrototypeReturnReceiptStore()
	auditStore := audit.NewInMemoryLogStore()
	inspectPrototypeReceipt(t, store, auditStore, "intact", "reusable")
	service := NewPrototypeApplyReturnDispositionAt(
		store,
		auditStore,
		time.Date(2026, 4, 26, 11, 30, 0, 0, time.UTC),
	)

	result, err := service.Execute(context.Background(), ApplyReturnDispositionInput{
		ReceiptID:   "rr-260426-0001",
		Disposition: "reusable",
		Note:        "ready for putaway after final disposition",
		ActorID:     "user-return-inspector",
		RequestID:   "req-return-disposition",
	})
	if err != nil {
		t.Fatalf("apply return disposition: %v", err)
	}

	if result.Receipt.Status != domain.ReturnStatusDispositioned {
		t.Fatalf("receipt status = %q, want dispositioned", result.Receipt.Status)
	}
	if result.Action.TargetLocation != "return-putaway-ready" {
		t.Fatalf("target location = %q, want return-putaway-ready", result.Action.TargetLocation)
	}
	if result.Receipt.StockMovement == nil {
		t.Fatal("stock movement = nil, want reusable restock movement")
	}
	if result.Receipt.StockMovement.MovementType != "return_restock" ||
		result.Receipt.StockMovement.TargetStockStatus != "available" ||
		result.Receipt.StockMovement.SourceDocID != result.Receipt.ID {
		t.Fatalf("stock movement = %+v, want reusable available restock", result.Receipt.StockMovement)
	}
	if result.AuditLogID == "" {
		t.Fatal("audit log id is empty")
	}

	logs, err := auditStore.List(context.Background(), audit.Query{Action: "returns.inspection.disposition"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit logs = %d, want 1", len(logs))
	}
	if logs[0].AfterData["action_code"] != "route_to_putaway" {
		t.Fatalf("audit after data = %+v, want route_to_putaway", logs[0].AfterData)
	}
	if logs[0].AfterData["stock_movement_type"] != "return_restock" {
		t.Fatalf("audit after data = %+v, want stock movement type", logs[0].AfterData)
	}
}

func TestApplyReturnDispositionRoutesNotReusableAndQAHold(t *testing.T) {
	tests := []struct {
		name            string
		condition       string
		disposition     string
		wantLocation    string
		wantStockStatus string
		wantMovement    *domain.ReturnStockMovement
	}{
		{
			name:            "not reusable",
			condition:       "damaged",
			disposition:     "not_reusable",
			wantLocation:    "lab-damaged-placeholder",
			wantStockStatus: "damaged",
		},
		{
			name:            "needs qa",
			condition:       "seal_torn",
			disposition:     "needs_inspection",
			wantLocation:    "return-quarantine-hold",
			wantStockStatus: "qc_hold",
			wantMovement: &domain.ReturnStockMovement{
				MovementType:      "return_receipt",
				TargetStockStatus: "qc_hold",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewPrototypeReturnReceiptStore()
			auditStore := audit.NewInMemoryLogStore()
			inspectPrototypeReceipt(t, store, auditStore, tt.condition, tt.disposition)
			service := NewPrototypeApplyReturnDispositionAt(
				store,
				auditStore,
				time.Date(2026, 4, 26, 11, 35, 0, 0, time.UTC),
			)

			result, err := service.Execute(context.Background(), ApplyReturnDispositionInput{
				ReceiptID:   "rr-260426-0001",
				Disposition: tt.disposition,
				ActorID:     "user-return-inspector",
				RequestID:   "req-return-disposition",
			})
			if err != nil {
				t.Fatalf("apply return disposition: %v", err)
			}
			if result.Action.TargetLocation != tt.wantLocation {
				t.Fatalf("target location = %q, want %s", result.Action.TargetLocation, tt.wantLocation)
			}
			if result.Action.TargetStockStatus != tt.wantStockStatus {
				t.Fatalf("target stock status = %q, want %s", result.Action.TargetStockStatus, tt.wantStockStatus)
			}
			if tt.wantMovement == nil && result.Receipt.StockMovement != nil {
				t.Fatalf("stock movement = %+v, want nil", result.Receipt.StockMovement)
			}
			if tt.wantMovement != nil {
				if result.Receipt.StockMovement == nil {
					t.Fatal("stock movement = nil, want quarantine movement")
				}
				if result.Receipt.StockMovement.MovementType != tt.wantMovement.MovementType ||
					result.Receipt.StockMovement.TargetStockStatus != tt.wantMovement.TargetStockStatus {
					t.Fatalf("stock movement = %+v, want %+v", result.Receipt.StockMovement, tt.wantMovement)
				}
			}
		})
	}
}

func TestApplyReturnDispositionRejectsPendingInvalidAndAlreadyDispositionedReceipts(t *testing.T) {
	store := NewPrototypeReturnReceiptStore()
	auditStore := audit.NewInMemoryLogStore()
	service := NewPrototypeApplyReturnDispositionAt(
		store,
		auditStore,
		time.Date(2026, 4, 26, 11, 40, 0, 0, time.UTC),
	)

	_, err := service.Execute(context.Background(), ApplyReturnDispositionInput{
		ReceiptID:   "rr-260426-0001",
		Disposition: "reusable",
		ActorID:     "user-return-inspector",
	})
	if !errors.Is(err, ErrReturnReceiptDispositionNotAllowed) {
		t.Fatalf("error = %v, want disposition not allowed", err)
	}

	inspectPrototypeReceipt(t, store, auditStore, "intact", "reusable")
	_, err = service.Execute(context.Background(), ApplyReturnDispositionInput{
		ReceiptID:   "rr-260426-0001",
		Disposition: "usable",
		ActorID:     "user-return-inspector",
	})
	if !errors.Is(err, domain.ErrReturnReceiptInvalidDisposition) {
		t.Fatalf("error = %v, want invalid disposition", err)
	}

	if _, err := service.Execute(context.Background(), ApplyReturnDispositionInput{
		ReceiptID:   "rr-260426-0001",
		Disposition: "reusable",
		ActorID:     "user-return-inspector",
	}); err != nil {
		t.Fatalf("first disposition failed: %v", err)
	}
	_, err = service.Execute(context.Background(), ApplyReturnDispositionInput{
		ReceiptID:   "rr-260426-0001",
		Disposition: "not_reusable",
		ActorID:     "user-return-inspector",
	})
	if !errors.Is(err, ErrReturnReceiptDispositionNotAllowed) {
		t.Fatalf("error = %v, want disposition not allowed", err)
	}
}

func inspectPrototypeReceipt(
	t *testing.T,
	store *PrototypeReturnReceiptStore,
	auditStore audit.LogStore,
	condition string,
	disposition string,
) {
	t.Helper()

	inspectService := NewPrototypeInspectReturnAt(
		store,
		auditStore,
		time.Date(2026, 4, 26, 11, 0, 0, 0, time.UTC),
	)
	if _, err := inspectService.Execute(context.Background(), InspectReturnInput{
		ReceiptID:   "rr-260426-0001",
		Condition:   condition,
		Disposition: disposition,
		ActorID:     "user-return-inspector",
		RequestID:   "req-return-inspect",
	}); err != nil {
		t.Fatalf("inspect return: %v", err)
	}
}
