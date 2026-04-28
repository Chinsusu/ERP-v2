package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewReturnInspectionRecordsReusableItem(t *testing.T) {
	inspection, err := NewReturnInspection(NewReturnInspectionInput{
		ReceiptID:   "rr-260426-0001",
		ReceiptNo:   "RR-260426-0001",
		Condition:   ReturnInspectionConditionIntact,
		Disposition: ReturnDispositionReusable,
		InspectorID: "user-return-inspector",
		InspectedAt: time.Date(2026, 4, 26, 11, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("new return inspection: %v", err)
	}

	if inspection.Status != ReturnInspectionStatusRecorded {
		t.Fatalf("status = %q, want inspection_recorded", inspection.Status)
	}
	if inspection.TargetLocation != "return-area-qc-release" {
		t.Fatalf("target location = %q, want return-area-qc-release", inspection.TargetLocation)
	}
	if inspection.RiskLevel != "low" {
		t.Fatalf("risk level = %q, want low", inspection.RiskLevel)
	}
}

func TestNewReturnInspectionRoutesDamagedAndMissingAccessoryAsHighRisk(t *testing.T) {
	for _, condition := range []ReturnInspectionCondition{
		ReturnInspectionConditionDamaged,
		ReturnInspectionConditionMissingAccessory,
	} {
		inspection, err := NewReturnInspection(NewReturnInspectionInput{
			ReceiptID:   "rr-260426-0001",
			Condition:   condition,
			Disposition: ReturnDispositionNotReusable,
			InspectorID: "user-return-inspector",
		})
		if err != nil {
			t.Fatalf("new return inspection for %q: %v", condition, err)
		}
		if inspection.TargetLocation != "lab-damaged-placeholder" {
			t.Fatalf("target location = %q, want lab-damaged-placeholder", inspection.TargetLocation)
		}
		if inspection.RiskLevel != "high" {
			t.Fatalf("risk level = %q, want high", inspection.RiskLevel)
		}
	}
}

func TestNewReturnInspectionRoutesQAHold(t *testing.T) {
	inspection, err := NewReturnInspection(NewReturnInspectionInput{
		ReceiptID:     "rr-260426-0001",
		Condition:     ReturnInspectionConditionSealTorn,
		Disposition:   ReturnDispositionNeedsInspection,
		InspectorID:   "user-return-inspector",
		Note:          "seal torn during carrier handling",
		EvidenceLabel: "photo-001",
		InspectedAt:   time.Date(2026, 4, 26, 11, 5, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("new return inspection: %v", err)
	}

	if inspection.Status != ReturnInspectionStatusQAHold {
		t.Fatalf("status = %q, want return_qa_hold", inspection.Status)
	}
	if inspection.TargetLocation != "return-qa-hold" {
		t.Fatalf("target location = %q, want return-qa-hold", inspection.TargetLocation)
	}
	if inspection.RiskLevel != "medium" {
		t.Fatalf("risk level = %q, want medium", inspection.RiskLevel)
	}
	if inspection.Note == "" || inspection.EvidenceLabel == "" {
		t.Fatalf("note/evidence = %q/%q, want preserved values", inspection.Note, inspection.EvidenceLabel)
	}
}

func TestReturnReceiptApplyInspectionUpdatesStatusAndLineCondition(t *testing.T) {
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
		Condition:   ReturnInspectionConditionUsed,
		Disposition: ReturnDispositionNeedsInspection,
		InspectorID: "user-return-inspector",
	})
	if err != nil {
		t.Fatalf("new return inspection: %v", err)
	}

	updated := receipt.ApplyInspection(inspection)
	if updated.Status != ReturnStatusInspected {
		t.Fatalf("status = %q, want inspected", updated.Status)
	}
	if updated.PackageCondition != "used" || updated.Lines[0].Condition != "used" {
		t.Fatalf("condition = %q/%q, want used", updated.PackageCondition, updated.Lines[0].Condition)
	}
	if updated.TargetLocation != "return-qa-hold" {
		t.Fatalf("target location = %q, want return-qa-hold", updated.TargetLocation)
	}
	if updated.StockMovement != nil {
		t.Fatalf("stock movement = %+v, want nil", updated.StockMovement)
	}
}

func TestNewReturnInspectionValidatesRequiredFieldsConditionAndDisposition(t *testing.T) {
	if _, err := NewReturnInspection(NewReturnInspectionInput{
		Condition:   ReturnInspectionConditionIntact,
		Disposition: ReturnDispositionReusable,
		InspectorID: "user-return-inspector",
	}); !errors.Is(err, ErrReturnInspectionRequiredField) {
		t.Fatalf("err = %v, want required field", err)
	}

	if _, err := NewReturnInspection(NewReturnInspectionInput{
		ReceiptID:   "rr-260426-0001",
		Condition:   "opened_good",
		Disposition: ReturnDispositionReusable,
		InspectorID: "user-return-inspector",
	}); !errors.Is(err, ErrReturnInspectionInvalidCondition) {
		t.Fatalf("err = %v, want invalid condition", err)
	}

	if _, err := NewReturnInspection(NewReturnInspectionInput{
		ReceiptID:   "rr-260426-0001",
		Condition:   ReturnInspectionConditionIntact,
		Disposition: "usable",
		InspectorID: "user-return-inspector",
	}); !errors.Is(err, ErrReturnReceiptInvalidDisposition) {
		t.Fatalf("err = %v, want invalid disposition", err)
	}
}
