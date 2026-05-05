package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestSubcontractFactoryDispatchLifecycle(t *testing.T) {
	createdAt := time.Date(2026, 5, 6, 9, 0, 0, 0, time.UTC)
	dispatch, err := NewSubcontractFactoryDispatch(NewSubcontractFactoryDispatchInput{
		ID:                     "dispatch-001",
		OrgID:                  "org-my-pham",
		DispatchNo:             "FDP-260506-001",
		SubcontractOrderID:     "sco-001",
		SubcontractOrderNo:     "SCO-260506-001",
		SourceProductionPlanID: "plan-001",
		SourceProductionPlanNo: "PP-260506-001",
		FactoryID:              "fac-001",
		FactoryCode:            "FAC-HCM",
		FactoryName:            "HCM Cosmetics Factory",
		FinishedItemID:         "fg-serum",
		FinishedSKUCode:        "FG-SERUM-001",
		FinishedItemName:       "Brightening Serum",
		PlannedQty:             decimal.MustQuantity("1000"),
		UOMCode:                "PCS",
		SpecSummary:            "formula v2026.05, box spec approved",
		SampleRequired:         true,
		TargetStartDate:        "2026-05-08",
		ExpectedReceiptDate:    "2026-05-20",
		Note:                   "Send manual dispatch pack to factory coordinator.",
		CreatedAt:              createdAt,
		CreatedBy:              "production-user",
		Lines: []NewSubcontractFactoryDispatchLineInput{
			{
				ID:                  "dispatch-line-001",
				LineNo:              1,
				OrderMaterialLineID: "sco-mat-001",
				ItemID:              "rm-base",
				SKUCode:             "RM-BASE",
				ItemName:            "Serum Base",
				PlannedQty:          decimal.MustQuantity("10"),
				UOMCode:             "KG",
				LotTraceRequired:    true,
				Note:                "Use approved lot.",
			},
		},
	})
	if err != nil {
		t.Fatalf("new dispatch: %v", err)
	}
	if dispatch.Status != SubcontractFactoryDispatchStatusDraft || dispatch.Version != 1 || len(dispatch.Lines) != 1 {
		t.Fatalf("dispatch = %+v, want draft version 1 with material snapshot", dispatch)
	}

	readyAt := createdAt.Add(time.Hour)
	ready, err := dispatch.MarkReady("production-lead", readyAt)
	if err != nil {
		t.Fatalf("mark ready: %v", err)
	}
	if ready.Status != SubcontractFactoryDispatchStatusReady || ready.ReadyBy != "production-lead" || ready.ReadyAt != readyAt {
		t.Fatalf("ready dispatch = %+v, want ready metadata", ready)
	}

	sentAt := readyAt.Add(time.Hour)
	sent, err := ready.MarkSent(MarkSubcontractFactoryDispatchSentInput{
		SentBy: "production-user",
		SentAt: sentAt,
		Evidence: []NewSubcontractFactoryDispatchEvidenceInput{
			{
				ID:           "dispatch-evidence-001",
				EvidenceType: "manual_send",
				FileName:     "factory-dispatch-screenshot.png",
				ObjectKey:    "subcontract/dispatch-001/factory-dispatch-screenshot.png",
				Note:         "Manual send evidence.",
			},
		},
		Note: "Sent through manual business channel.",
	})
	if err != nil {
		t.Fatalf("mark sent: %v", err)
	}
	if sent.Status != SubcontractFactoryDispatchStatusSent || sent.SentBy != "production-user" || len(sent.Evidence) != 1 {
		t.Fatalf("sent dispatch = %+v, want sent metadata and evidence", sent)
	}

	respondedAt := sentAt.Add(time.Hour)
	confirmed, err := sent.RecordResponse(RecordSubcontractFactoryDispatchResponseInput{
		ResponseStatus: SubcontractFactoryDispatchStatusConfirmed,
		ResponseBy:     "factory-coordinator",
		RespondedAt:    respondedAt,
		ResponseNote:   "Factory confirmed quantity, spec, and delivery date.",
	})
	if err != nil {
		t.Fatalf("record response: %v", err)
	}
	if confirmed.Status != SubcontractFactoryDispatchStatusConfirmed ||
		confirmed.ResponseBy != "factory-coordinator" ||
		confirmed.RespondedAt != respondedAt ||
		confirmed.FactoryResponseNote == "" {
		t.Fatalf("confirmed dispatch = %+v, want confirmed response metadata", confirmed)
	}
}

func TestSubcontractFactoryDispatchRejectedResponseRequiresReason(t *testing.T) {
	dispatch := subcontractFactoryDispatchReadyToSend(t)
	sent, err := dispatch.MarkSent(MarkSubcontractFactoryDispatchSentInput{
		SentBy: "production-user",
		Evidence: []NewSubcontractFactoryDispatchEvidenceInput{
			{
				ID:           "dispatch-evidence-001",
				EvidenceType: "manual_send",
				ObjectKey:    "subcontract/dispatch-001/factory-dispatch-screenshot.png",
			},
		},
	})
	if err != nil {
		t.Fatalf("mark sent: %v", err)
	}

	_, err = sent.RecordResponse(RecordSubcontractFactoryDispatchResponseInput{
		ResponseStatus: SubcontractFactoryDispatchStatusRevisionRequested,
		ResponseBy:     "factory-coordinator",
		ResponseNote:   " ",
	})
	if !errors.Is(err, ErrSubcontractFactoryDispatchRequiredField) {
		t.Fatalf("error = %v, want required response reason", err)
	}
}

func subcontractFactoryDispatchReadyToSend(t *testing.T) SubcontractFactoryDispatch {
	t.Helper()

	dispatch, err := NewSubcontractFactoryDispatch(NewSubcontractFactoryDispatchInput{
		ID:                  "dispatch-001",
		OrgID:               "org-my-pham",
		DispatchNo:          "FDP-260506-001",
		SubcontractOrderID:  "sco-001",
		SubcontractOrderNo:  "SCO-260506-001",
		FactoryID:           "fac-001",
		FactoryName:         "HCM Cosmetics Factory",
		FinishedItemID:      "fg-serum",
		FinishedSKUCode:     "FG-SERUM-001",
		FinishedItemName:    "Brightening Serum",
		PlannedQty:          decimal.MustQuantity("1000"),
		UOMCode:             "PCS",
		ExpectedReceiptDate: "2026-05-20",
		CreatedBy:           "production-user",
		Lines: []NewSubcontractFactoryDispatchLineInput{
			{
				ID:                  "dispatch-line-001",
				LineNo:              1,
				OrderMaterialLineID: "sco-mat-001",
				ItemID:              "rm-base",
				SKUCode:             "RM-BASE",
				ItemName:            "Serum Base",
				PlannedQty:          decimal.MustQuantity("10"),
				UOMCode:             "KG",
				LotTraceRequired:    true,
			},
		},
	})
	if err != nil {
		t.Fatalf("new dispatch: %v", err)
	}
	ready, err := dispatch.MarkReady("production-lead", time.Now())
	if err != nil {
		t.Fatalf("mark ready: %v", err)
	}

	return ready
}
