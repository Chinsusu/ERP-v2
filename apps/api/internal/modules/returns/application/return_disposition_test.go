package application

import (
	"context"
	"errors"
	"testing"
	"time"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
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

	movementLogs, err := auditStore.List(context.Background(), audit.Query{Action: returnStockMovementRecordedAction})
	if err != nil {
		t.Fatalf("list movement audit logs: %v", err)
	}
	if len(movementLogs) != 1 {
		t.Fatalf("movement audit logs = %d, want 1", len(movementLogs))
	}
	if movementLogs[0].EntityType != returnStockMovementEntityType ||
		movementLogs[0].AfterData["movement_type"] != "return_restock" ||
		movementLogs[0].AfterData["delta_available"] != "1.000000" {
		t.Fatalf("movement audit log = %+v, want reusable available movement", movementLogs[0])
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
			quarantineLogs, err := auditStore.List(context.Background(), audit.Query{Action: returnStockQuarantinedAction})
			if err != nil {
				t.Fatalf("list quarantine audit logs: %v", err)
			}
			if tt.wantMovement == nil && len(quarantineLogs) != 0 {
				t.Fatalf("quarantine logs = %+v, want none", quarantineLogs)
			}
			if tt.wantMovement != nil {
				if len(quarantineLogs) != 1 {
					t.Fatalf("quarantine logs = %d, want 1", len(quarantineLogs))
				}
				if quarantineLogs[0].AfterData["stock_status"] != "qc_hold" ||
					quarantineLogs[0].AfterData["delta_available"] != "0.000000" {
					t.Fatalf("quarantine audit log = %+v, want qc hold with zero available delta", quarantineLogs[0])
				}
			}
		})
	}
}

func TestApplyReturnDispositionStockRegressionByDisposition(t *testing.T) {
	tests := []struct {
		name                    string
		condition               string
		disposition             string
		wantMovementCount       int
		wantMovementType        string
		wantStockStatus         string
		wantDeltaOnHand         string
		wantDeltaReserved       string
		wantDeltaAvailable      string
		wantRecordedAuditAction string
	}{
		{
			name:                    "reusable moves to available",
			condition:               "intact",
			disposition:             "reusable",
			wantMovementCount:       1,
			wantMovementType:        "return_restock",
			wantStockStatus:         "available",
			wantDeltaOnHand:         "1.000000",
			wantDeltaReserved:       "0.000000",
			wantDeltaAvailable:      "1.000000",
			wantRecordedAuditAction: returnStockMovementRecordedAction,
		},
		{
			name:                    "needs inspection moves to quarantine only",
			condition:               "seal_torn",
			disposition:             "needs_inspection",
			wantMovementCount:       1,
			wantMovementType:        "return_receipt",
			wantStockStatus:         "qc_hold",
			wantDeltaOnHand:         "1.000000",
			wantDeltaReserved:       "0.000000",
			wantDeltaAvailable:      "0.000000",
			wantRecordedAuditAction: returnStockQuarantinedAction,
		},
		{
			name:              "not reusable does not move stock",
			condition:         "damaged",
			disposition:       "not_reusable",
			wantMovementCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewPrototypeReturnReceiptStore()
			auditStore := audit.NewInMemoryLogStore()
			inspectPrototypeReceipt(t, store, auditStore, tt.condition, tt.disposition)
			movementStore := inventoryapp.NewInMemoryStockMovementStore()
			service := NewApplyReturnDisposition(store, movementStore, auditStore)
			service.clock = func() time.Time { return time.Date(2026, 4, 26, 12, 0, 0, 0, time.UTC) }

			if _, err := service.Execute(context.Background(), ApplyReturnDispositionInput{
				ReceiptID:   "rr-260426-0001",
				Disposition: tt.disposition,
				ActorID:     "user-return-inspector",
				RequestID:   "req-return-regression",
			}); err != nil {
				t.Fatalf("apply return disposition: %v", err)
			}

			movements := movementStore.Movements()
			if len(movements) != tt.wantMovementCount {
				t.Fatalf("stock movements = %d, want %d", len(movements), tt.wantMovementCount)
			}
			if tt.wantMovementCount == 0 {
				logs, err := auditStore.List(context.Background(), audit.Query{EntityType: returnStockMovementEntityType})
				if err != nil {
					t.Fatalf("list stock movement audit logs: %v", err)
				}
				if len(logs) != 0 {
					t.Fatalf("stock movement audit logs = %+v, want none", logs)
				}
				return
			}

			movement := movements[0]
			if string(movement.MovementType) != tt.wantMovementType || string(movement.StockStatus) != tt.wantStockStatus {
				t.Fatalf("movement = %+v, want %s/%s", movement, tt.wantMovementType, tt.wantStockStatus)
			}
			delta, err := movement.BalanceDelta()
			if err != nil {
				t.Fatalf("balance delta: %v", err)
			}
			if delta.OnHand.String() != tt.wantDeltaOnHand ||
				delta.Reserved.String() != tt.wantDeltaReserved ||
				delta.Available.String() != tt.wantDeltaAvailable {
				t.Fatalf("delta = %+v, want on_hand %s reserved %s available %s", delta, tt.wantDeltaOnHand, tt.wantDeltaReserved, tt.wantDeltaAvailable)
			}

			logs, err := auditStore.List(context.Background(), audit.Query{Action: tt.wantRecordedAuditAction})
			if err != nil {
				t.Fatalf("list stock movement audit logs: %v", err)
			}
			if len(logs) != 1 {
				t.Fatalf("stock movement audit logs = %d, want 1", len(logs))
			}
			if logs[0].AfterData["delta_on_hand"] != tt.wantDeltaOnHand ||
				logs[0].AfterData["delta_reserved"] != tt.wantDeltaReserved ||
				logs[0].AfterData["delta_available"] != tt.wantDeltaAvailable {
				t.Fatalf("stock movement audit log = %+v, want movement deltas", logs[0])
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
