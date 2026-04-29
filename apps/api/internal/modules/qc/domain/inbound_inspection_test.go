package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestNewInboundQCInspectionNormalizesReceivingContext(t *testing.T) {
	inspection, err := NewInboundQCInspection(validInboundQCInspectionInput())
	if err != nil {
		t.Fatalf("new inbound qc inspection: %v", err)
	}

	if inspection.Status != InboundQCInspectionStatusPending {
		t.Fatalf("status = %q, want pending", inspection.Status)
	}
	if inspection.GoodsReceiptNo != "GRN-260429-0001" || inspection.SKU != "CREAM-50G" || inspection.BatchNo != "LOT-2603B" {
		t.Fatalf("normalized context = %+v, want uppercase receipt/SKU/batch", inspection)
	}
	if inspection.LotNo != "LOT-2603B" || inspection.ExpiryDate.Hour() != 0 {
		t.Fatalf("lot/expiry = %s/%s, want normalized date-only lot", inspection.LotNo, inspection.ExpiryDate)
	}
	if got, want := inspection.Quantity.String(), "12.000000"; got != want {
		t.Fatalf("quantity = %s, want %s", got, want)
	}
	if len(inspection.Checklist) != 3 || inspection.Checklist[0].Code != "PACKAGING" {
		t.Fatalf("checklist = %+v, want normalized checklist", inspection.Checklist)
	}
}

func TestInboundQCInspectionStartAndPassDecision(t *testing.T) {
	startedAt := time.Date(2026, 4, 29, 10, 0, 0, 0, time.UTC)
	inspection, err := NewInboundQCInspection(validInboundQCInspectionInput())
	if err != nil {
		t.Fatalf("new inbound qc inspection: %v", err)
	}

	started, err := inspection.Start("qa-user", startedAt)
	if err != nil {
		t.Fatalf("start inspection: %v", err)
	}
	if started.Status != InboundQCInspectionStatusInProgress || started.StartedBy != "qa-user" || !started.StartedAt.Equal(startedAt) {
		t.Fatalf("started inspection = %+v, want in-progress metadata", started)
	}

	decidedAt := startedAt.Add(15 * time.Minute)
	completed, err := started.RecordDecision(InboundQCDecisionInput{
		Result:         " pass ",
		PassedQuantity: decimal.MustQuantity("12"),
		Checklist:      completedChecklist(InboundQCChecklistStatusPass),
		ActorID:        "qa-user",
		ChangedAt:      decidedAt,
	})
	if err != nil {
		t.Fatalf("record pass decision: %v", err)
	}
	if completed.Status != InboundQCInspectionStatusCompleted ||
		completed.Result != InboundQCResultPass ||
		completed.PassedQuantity.String() != "12.000000" ||
		completed.DecidedBy != "qa-user" ||
		!completed.DecidedAt.Equal(decidedAt) {
		t.Fatalf("completed inspection = %+v, want pass decision metadata", completed)
	}
}

func TestInboundQCInspectionRejectsIncompleteChecklistAndMissingReason(t *testing.T) {
	started := mustStartedInboundQCInspection(t)

	_, err := started.RecordDecision(InboundQCDecisionInput{
		Result:         InboundQCResultFail,
		FailedQuantity: decimal.MustQuantity("12"),
		Checklist:      completedChecklist(InboundQCChecklistStatusPending),
		Reason:         "damaged cartons",
		ActorID:        "qa-user",
		ChangedAt:      time.Now(),
	})
	if !errors.Is(err, ErrInboundQCChecklistIncomplete) {
		t.Fatalf("error = %v, want incomplete checklist", err)
	}

	_, err = started.RecordDecision(InboundQCDecisionInput{
		Result:         InboundQCResultFail,
		FailedQuantity: decimal.MustQuantity("12"),
		Checklist:      completedChecklist(InboundQCChecklistStatusFail),
		ActorID:        "qa-user",
		ChangedAt:      time.Now(),
	})
	if !errors.Is(err, ErrInboundQCInspectionRequiredField) {
		t.Fatalf("error = %v, want reason required", err)
	}
}

func TestInboundQCInspectionPartialDecisionRequiresQuantitySplit(t *testing.T) {
	started := mustStartedInboundQCInspection(t)

	_, err := started.RecordDecision(InboundQCDecisionInput{
		Result:         InboundQCResultPartial,
		PassedQuantity: decimal.MustQuantity("8"),
		HoldQuantity:   decimal.MustQuantity("3"),
		Checklist:      completedChecklist(InboundQCChecklistStatusPass),
		Reason:         "sample hold",
		ActorID:        "qa-user",
		ChangedAt:      time.Now(),
	})
	if !errors.Is(err, ErrInboundQCInspectionInvalidQuantity) {
		t.Fatalf("error = %v, want invalid quantity", err)
	}

	completed, err := started.RecordDecision(InboundQCDecisionInput{
		Result:         InboundQCResultPartial,
		PassedQuantity: decimal.MustQuantity("8"),
		HoldQuantity:   decimal.MustQuantity("4"),
		Checklist:      completedChecklist(InboundQCChecklistStatusPass),
		Reason:         "sample hold",
		ActorID:        "qa-user",
		ChangedAt:      time.Now(),
	})
	if err != nil {
		t.Fatalf("record partial decision: %v", err)
	}
	if completed.Result != InboundQCResultPartial ||
		completed.PassedQuantity.String() != "8.000000" ||
		completed.HoldQuantity.String() != "4.000000" {
		t.Fatalf("completed partial = %+v, want 8 pass and 4 hold", completed)
	}
}

func TestInboundQCInspectionRejectsInvalidTransitions(t *testing.T) {
	inspection, err := NewInboundQCInspection(validInboundQCInspectionInput())
	if err != nil {
		t.Fatalf("new inbound qc inspection: %v", err)
	}

	_, err = inspection.RecordDecision(InboundQCDecisionInput{
		Result:         InboundQCResultPass,
		PassedQuantity: decimal.MustQuantity("12"),
		Checklist:      completedChecklist(InboundQCChecklistStatusPass),
		ActorID:        "qa-user",
	})
	if !errors.Is(err, ErrInboundQCInspectionInvalidTransition) {
		t.Fatalf("error = %v, want invalid transition", err)
	}

	started, err := inspection.Start("qa-user", time.Now())
	if err != nil {
		t.Fatalf("start inspection: %v", err)
	}
	cancelled, err := started.Cancel("qa-user", "duplicate inspection", time.Now())
	if err != nil {
		t.Fatalf("cancel inspection: %v", err)
	}
	if cancelled.Status != InboundQCInspectionStatusCancelled || cancelled.Reason != "duplicate inspection" {
		t.Fatalf("cancelled = %+v, want cancelled reason", cancelled)
	}
	_, err = cancelled.Start("qa-user", time.Now())
	if !errors.Is(err, ErrInboundQCInspectionInvalidTransition) {
		t.Fatalf("error = %v, want terminal cancelled transition", err)
	}
}

func validInboundQCInspectionInput() NewInboundQCInspectionInput {
	return NewInboundQCInspectionInput{
		ID:                  "iqc-grn-260429-0001-line-01",
		OrgID:               "org-my-pham",
		GoodsReceiptID:      "grn-260429-0001",
		GoodsReceiptNo:      " grn-260429-0001 ",
		GoodsReceiptLineID:  "grn-260429-0001-line-01",
		PurchaseOrderID:     "po-260429-0003",
		PurchaseOrderLineID: "po-260429-0003-line-01",
		ItemID:              "item-cream-50g",
		SKU:                 " cream-50g ",
		ItemName:            "Repair Cream 50g",
		BatchID:             "batch-cream-2603b",
		BatchNo:             " lot-2603b ",
		LotNo:               " lot-2603b ",
		ExpiryDate:          time.Date(2028, 3, 1, 15, 30, 0, 0, time.FixedZone("ICT", 7*60*60)),
		WarehouseID:         "wh-hcm-fg",
		LocationID:          "loc-hcm-fg-recv-01",
		Quantity:            decimal.MustQuantity("12"),
		UOMCode:             " ea ",
		InspectorID:         "qa-user",
		Checklist: []NewInboundQCChecklistItemInput{
			{ID: "check-packaging", Code: " packaging ", Label: "Packaging", Required: true},
			{ID: "check-lot", Code: "lot", Label: "Lot / expiry", Required: true},
			{ID: "check-sample", Code: "sample", Label: "Sample retained", Required: false},
		},
		CreatedAt: time.Date(2026, 4, 29, 9, 0, 0, 0, time.UTC),
		CreatedBy: "warehouse-user",
	}
}

func mustStartedInboundQCInspection(t *testing.T) InboundQCInspection {
	t.Helper()
	inspection, err := NewInboundQCInspection(validInboundQCInspectionInput())
	if err != nil {
		t.Fatalf("new inbound qc inspection: %v", err)
	}
	started, err := inspection.Start("qa-user", time.Now())
	if err != nil {
		t.Fatalf("start inspection: %v", err)
	}

	return started
}

func completedChecklist(status InboundQCChecklistStatus) []NewInboundQCChecklistItemInput {
	return []NewInboundQCChecklistItemInput{
		{ID: "check-packaging", Code: "PACKAGING", Label: "Packaging", Required: true, Status: status},
		{ID: "check-lot", Code: "LOT", Label: "Lot / expiry", Required: true, Status: status},
		{ID: "check-sample", Code: "SAMPLE", Label: "Sample retained", Required: false, Status: InboundQCChecklistStatusNotApplicable},
	}
}
