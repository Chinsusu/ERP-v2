package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewReturnDispositionActionRoutesReusableToPutawayReady(t *testing.T) {
	action, err := NewReturnDispositionAction(NewReturnDispositionActionInput{
		ReceiptID:   "rr-260426-0001",
		ReceiptNo:   "RR-260426-0001",
		Disposition: ReturnDispositionReusable,
		ActorID:     "user-return-inspector",
		Note:        "usable after inspection",
		DecidedAt:   time.Date(2026, 4, 26, 11, 30, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("new return disposition action: %v", err)
	}

	if action.TargetLocation != "return-putaway-ready" {
		t.Fatalf("target location = %q, want return-putaway-ready", action.TargetLocation)
	}
	if action.TargetStockStatus != "return_pending" {
		t.Fatalf("target stock status = %q, want return_pending", action.TargetStockStatus)
	}
	if action.ActionCode != "route_to_putaway" {
		t.Fatalf("action code = %q, want route_to_putaway", action.ActionCode)
	}
}

func TestNewReturnDispositionActionRoutesNotReusableAndQAHold(t *testing.T) {
	tests := []struct {
		name            string
		disposition     ReturnDisposition
		wantLocation    string
		wantStockStatus string
		wantActionCode  string
	}{
		{
			name:            "not reusable",
			disposition:     ReturnDispositionNotReusable,
			wantLocation:    "lab-damaged-placeholder",
			wantStockStatus: "damaged",
			wantActionCode:  "route_to_lab_or_damaged",
		},
		{
			name:            "needs inspection",
			disposition:     ReturnDispositionNeedsInspection,
			wantLocation:    "return-quarantine-hold",
			wantStockStatus: "qc_hold",
			wantActionCode:  "route_to_quarantine_hold",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action, err := NewReturnDispositionAction(NewReturnDispositionActionInput{
				ReceiptID:   "rr-260426-0001",
				Disposition: tt.disposition,
				ActorID:     "user-return-inspector",
			})
			if err != nil {
				t.Fatalf("new return disposition action: %v", err)
			}
			if action.TargetLocation != tt.wantLocation {
				t.Fatalf("target location = %q, want %s", action.TargetLocation, tt.wantLocation)
			}
			if action.TargetStockStatus != tt.wantStockStatus {
				t.Fatalf("target stock status = %q, want %s", action.TargetStockStatus, tt.wantStockStatus)
			}
			if action.ActionCode != tt.wantActionCode {
				t.Fatalf("action code = %q, want %s", action.ActionCode, tt.wantActionCode)
			}
		})
	}
}

func TestReturnReceiptApplyDispositionMarksDispositionedWithoutStockMovement(t *testing.T) {
	receipt, err := NewReturnReceipt(NewReturnReceiptInput{
		WarehouseID:      "wh-hcm",
		ScanCode:         "RET-260426-001",
		PackageCondition: "sealed bag",
		Disposition:      ReturnDispositionNeedsInspection,
	})
	if err != nil {
		t.Fatalf("new return receipt: %v", err)
	}
	inspection, err := NewReturnInspection(NewReturnInspectionInput{
		ReceiptID:   receipt.ID,
		ReceiptNo:   receipt.ReceiptNo,
		Condition:   ReturnInspectionConditionIntact,
		Disposition: ReturnDispositionReusable,
		InspectorID: "user-return-inspector",
	})
	if err != nil {
		t.Fatalf("new return inspection: %v", err)
	}
	action, err := NewReturnDispositionAction(NewReturnDispositionActionInput{
		ReceiptID:   receipt.ID,
		ReceiptNo:   receipt.ReceiptNo,
		Disposition: ReturnDispositionReusable,
		ActorID:     "user-return-inspector",
	})
	if err != nil {
		t.Fatalf("new return disposition action: %v", err)
	}

	updated := receipt.ApplyInspection(inspection).ApplyDisposition(action)
	if updated.Status != ReturnStatusDispositioned {
		t.Fatalf("status = %q, want dispositioned", updated.Status)
	}
	if updated.TargetLocation != "return-putaway-ready" {
		t.Fatalf("target location = %q, want return-putaway-ready", updated.TargetLocation)
	}
	if updated.StockMovement != nil {
		t.Fatalf("stock movement = %+v, want nil", updated.StockMovement)
	}
}

func TestNewReturnDispositionActionValidatesRequiredFieldsAndDisposition(t *testing.T) {
	if _, err := NewReturnDispositionAction(NewReturnDispositionActionInput{
		Disposition: ReturnDispositionReusable,
		ActorID:     "user-return-inspector",
	}); !errors.Is(err, ErrReturnDispositionRequiredField) {
		t.Fatalf("err = %v, want required field", err)
	}

	if _, err := NewReturnDispositionAction(NewReturnDispositionActionInput{
		ReceiptID:   "rr-260426-0001",
		Disposition: "usable",
		ActorID:     "user-return-inspector",
	}); !errors.Is(err, ErrReturnReceiptInvalidDisposition) {
		t.Fatalf("err = %v, want invalid disposition", err)
	}
}
