package application

import (
	"context"
	"errors"
	"testing"
	"time"

	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestSubcontractSampleApprovalServiceBuildsSubmitApproveFlow(t *testing.T) {
	ctx := context.Background()
	service := NewSubcontractSampleApprovalService()
	submittedAt := time.Date(2026, 4, 29, 11, 0, 0, 0, time.UTC)
	approvedAt := submittedAt.Add(time.Hour)
	order := subcontractSampleOrderReadyForSample(t)

	submitted, err := service.BuildSubmission(ctx, BuildSubcontractSampleSubmissionInput{
		ID:             "sample-001",
		Order:          order,
		SampleCode:     "SCO-260429-001-SAMPLE-A",
		FormulaVersion: "formula-2026.04",
		SpecVersion:    "spec-2026.04",
		SubmittedBy:    "factory-user",
		SubmittedAt:    submittedAt,
		ActorID:        "qa-user",
		Evidence: []BuildSubcontractSampleEvidenceInput{
			{
				EvidenceType: "photo",
				FileName:     "sample-front.jpg",
				ObjectKey:    "subcontract/sco-001/sample-front.jpg",
			},
		},
	})
	if err != nil {
		t.Fatalf("submit sample: %v", err)
	}

	if submitted.UpdatedOrder.Status != productiondomain.SubcontractOrderStatusSampleSubmitted ||
		submitted.SampleApproval.Status != productiondomain.SubcontractSampleApprovalStatusSubmitted ||
		submitted.SampleApproval.SubmittedAt != submittedAt ||
		len(submitted.SampleApproval.Evidence) != 1 {
		t.Fatalf("submitted = %+v, want sample submission and order transition", submitted)
	}

	approved, err := service.BuildApproval(ctx, BuildSubcontractSampleDecisionInput{
		Order:          submitted.UpdatedOrder,
		SampleApproval: submitted.SampleApproval,
		DecisionBy:     "qa-lead",
		DecisionAt:     approvedAt,
		Reason:         "approved against spec",
		StorageStatus:  "retained_in_qa_cabinet",
	})
	if err != nil {
		t.Fatalf("approve sample: %v", err)
	}

	if approved.UpdatedOrder.Status != productiondomain.SubcontractOrderStatusSampleApproved ||
		approved.UpdatedOrder.SampleApprovedBy != "qa-lead" ||
		approved.SampleApproval.Status != productiondomain.SubcontractSampleApprovalStatusApproved ||
		approved.SampleApproval.StorageStatus != "retained_in_qa_cabinet" {
		t.Fatalf("approved = %+v, want approved sample and order transition", approved)
	}
}

func TestSubcontractSampleApprovalServiceBuildsRejectionAndKeepsMassProductionBlocked(t *testing.T) {
	ctx := context.Background()
	service := NewSubcontractSampleApprovalService()
	order := subcontractSampleOrderReadyForSample(t)
	submitted, err := service.BuildSubmission(ctx, BuildSubcontractSampleSubmissionInput{
		ID:          "sample-001",
		Order:       order,
		SampleCode:  "SCO-260429-001-SAMPLE-A",
		SubmittedBy: "factory-user",
		SubmittedAt: time.Date(2026, 4, 29, 11, 0, 0, 0, time.UTC),
		ActorID:     "qa-user",
		Evidence: []BuildSubcontractSampleEvidenceInput{
			{
				EvidenceType: "photo",
				FileName:     "sample-front.jpg",
				ObjectKey:    "subcontract/sco-001/sample-front.jpg",
			},
		},
	})
	if err != nil {
		t.Fatalf("submit sample: %v", err)
	}

	rejected, err := service.BuildRejection(ctx, BuildSubcontractSampleDecisionInput{
		Order:          submitted.UpdatedOrder,
		SampleApproval: submitted.SampleApproval,
		DecisionBy:     "qa-lead",
		DecisionAt:     time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC),
		Reason:         "shade does not match approved formula",
	})
	if err != nil {
		t.Fatalf("reject sample: %v", err)
	}

	if rejected.UpdatedOrder.Status != productiondomain.SubcontractOrderStatusSampleRejected ||
		rejected.UpdatedOrder.SampleRejectReason != "shade does not match approved formula" ||
		rejected.SampleApproval.Status != productiondomain.SubcontractSampleApprovalStatusRejected {
		t.Fatalf("rejected = %+v, want rejected sample and order transition", rejected)
	}
	_, err = rejected.UpdatedOrder.StartMassProduction("operations-lead", time.Now())
	if !errors.Is(err, productiondomain.ErrSubcontractOrderInvalidTransition) {
		t.Fatalf("start mass production error = %v, want invalid transition after sample rejection", err)
	}
}

func TestSubcontractSampleApprovalServiceRejectsMissingEvidence(t *testing.T) {
	service := NewSubcontractSampleApprovalService()
	order := subcontractSampleOrderReadyForSample(t)

	_, err := service.BuildSubmission(context.Background(), BuildSubcontractSampleSubmissionInput{
		Order:       order,
		SampleCode:  "SCO-260429-001-SAMPLE-A",
		SubmittedBy: "factory-user",
		ActorID:     "qa-user",
	})
	if !errors.Is(err, productiondomain.ErrSubcontractSampleApprovalRequiredField) {
		t.Fatalf("error = %v, want required evidence", err)
	}
}

func subcontractSampleOrderReadyForSample(t *testing.T) productiondomain.SubcontractOrder {
	t.Helper()

	order, err := productiondomain.NewSubcontractOrderDocument(productiondomain.NewSubcontractOrderDocumentInput{
		ID:                  "sco-001",
		OrgID:               "org-my-pham",
		OrderNo:             "SCO-260429-001",
		FactoryID:           "fac-001",
		FactoryCode:         "FAC-HCM-01",
		FactoryName:         "HCM Cosmetics Factory",
		FinishedItemID:      "item-serum",
		FinishedSKUCode:     "FG-SERUM-001",
		FinishedItemName:    "Brightening Serum",
		PlannedQty:          decimal.MustQuantity("1000"),
		UOMCode:             "PCS",
		BasePlannedQty:      decimal.MustQuantity("1000"),
		BaseUOMCode:         "PCS",
		ConversionFactor:    decimal.MustQuantity("1"),
		CurrencyCode:        "VND",
		SpecSummary:         "spec-2026.04",
		SampleRequired:      true,
		ExpectedReceiptDate: "2026-05-12",
		CreatedBy:           "subcontract-user",
		MaterialLines: []productiondomain.NewSubcontractMaterialLineInput{
			{
				ID:               "sco-mat-001",
				LineNo:           1,
				ItemID:           "item-base",
				SKUCode:          "RM-BASE-001",
				ItemName:         "Serum Base",
				PlannedQty:       decimal.MustQuantity("10"),
				IssuedQty:        decimal.MustQuantity("10"),
				UOMCode:          "KG",
				BasePlannedQty:   decimal.MustQuantity("10000"),
				BaseIssuedQty:    decimal.MustQuantity("10000"),
				BaseUOMCode:      "G",
				ConversionFactor: decimal.MustQuantity("1000"),
				UnitCost:         decimal.MustUnitCost("150000"),
				CurrencyCode:     "VND",
				LotTraceRequired: true,
			},
		},
	})
	if err != nil {
		t.Fatalf("new subcontract order: %v", err)
	}
	order, err = order.Submit("subcontract-user", time.Now())
	if err != nil {
		t.Fatalf("submit order: %v", err)
	}
	order, err = order.Approve("operations-lead", time.Now())
	if err != nil {
		t.Fatalf("approve order: %v", err)
	}
	order, err = order.ConfirmFactory("factory-user", time.Now())
	if err != nil {
		t.Fatalf("confirm factory: %v", err)
	}
	order, err = order.MarkMaterialsIssued("warehouse-user", time.Now())
	if err != nil {
		t.Fatalf("mark materials issued: %v", err)
	}

	return order
}
