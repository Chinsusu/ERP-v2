package domain

import (
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestNewProductionPlanSnapshotsFormulaDemandAndShortage(t *testing.T) {
	createdAt := time.Date(2026, 5, 4, 10, 0, 0, 0, time.UTC)

	plan, err := NewProductionPlanDocument(NewProductionPlanDocumentInput{
		ID:                  "plan-001",
		OrgID:               "org-my-pham",
		PlanNo:              "PP-260504-0001",
		OutputItemID:        "item-xff-150",
		OutputSKU:           "XFF",
		OutputItemName:      "Tinh chat buoi Fast & Furious 150ML",
		OutputItemType:      "finished_good",
		PlannedQty:          decimal.MustQuantity("162"),
		UOMCode:             "PCS",
		FormulaID:           "formula-xff-v1",
		FormulaCode:         "XFF-150ML",
		FormulaVersion:      "v1",
		FormulaBatchQty:     decimal.MustQuantity("81"),
		FormulaBatchUOMCode: "PCS",
		PlannedStartDate:    "2026-05-10",
		PlannedEndDate:      "2026-05-11",
		Lines: []NewProductionPlanLineInput{
			{
				ID:                   "plan-line-001",
				FormulaLineID:        "formula-line-001",
				LineNo:               1,
				ComponentItemID:      "item-act-baicapil",
				ComponentSKU:         "ACT_BAICAPIL",
				ComponentName:        "BAICAPIL",
				ComponentType:        "raw_material",
				FormulaQty:           decimal.MustQuantity("1"),
				FormulaUOMCode:       "G",
				RequiredQty:          decimal.MustQuantity("162"),
				RequiredUOMCode:      "G",
				RequiredStockBaseQty: decimal.MustQuantity("0.162"),
				StockBaseUOMCode:     "KG",
				AvailableQty:         decimal.MustQuantity("0.0005"),
				ShortageQty:          decimal.MustQuantity("0.1615"),
				PurchaseDraftQty:     decimal.MustQuantity("0.1615"),
				PurchaseDraftUOMCode: "KG",
				IsStockManaged:       true,
				NeedsPurchase:        true,
			},
		},
		CreatedAt: createdAt,
		CreatedBy: "user-production",
	})
	if err != nil {
		t.Fatalf("NewProductionPlanDocument() error = %v", err)
	}

	if plan.Status != ProductionPlanStatusPurchaseRequestDraftCreated {
		t.Fatalf("status = %q, want purchase_request_draft_created", plan.Status)
	}
	if plan.FormulaID != "formula-xff-v1" || plan.FormulaVersion != "v1" || plan.FormulaBatchQty != "81.000000" {
		t.Fatalf("formula snapshot = %+v, want immutable formula reference and batch", plan)
	}
	if len(plan.Lines) != 1 {
		t.Fatalf("lines = %d, want 1", len(plan.Lines))
	}
	line := plan.Lines[0]
	if line.RequiredStockBaseQty != "0.162000" || line.AvailableQty != "0.000500" || line.ShortageQty != "0.161500" {
		t.Fatalf("line quantities = %+v, want required 0.162 KG, available 0.0005 KG, shortage 0.1615 KG", line)
	}
	if !line.NeedsPurchase || line.PurchaseDraftQty != "0.161500" || line.PurchaseDraftUOMCode != "KG" {
		t.Fatalf("purchase draft fields = %+v, want shortage drafted in stock-base UOM", line)
	}
}

func TestNewProductionPlanRejectsPurchaseLineWithoutShortage(t *testing.T) {
	_, err := NewProductionPlanDocument(NewProductionPlanDocumentInput{
		ID:                  "plan-001",
		OrgID:               "org-my-pham",
		PlanNo:              "PP-260504-0001",
		OutputItemID:        "item-xff-150",
		OutputSKU:           "XFF",
		OutputItemName:      "Tinh chat buoi Fast & Furious 150ML",
		OutputItemType:      "finished_good",
		PlannedQty:          decimal.MustQuantity("162"),
		UOMCode:             "PCS",
		FormulaID:           "formula-xff-v1",
		FormulaCode:         "XFF-150ML",
		FormulaVersion:      "v1",
		FormulaBatchQty:     decimal.MustQuantity("81"),
		FormulaBatchUOMCode: "PCS",
		Lines: []NewProductionPlanLineInput{
			{
				ID:                   "plan-line-001",
				FormulaLineID:        "formula-line-001",
				LineNo:               1,
				ComponentItemID:      "item-act-baicapil",
				ComponentSKU:         "ACT_BAICAPIL",
				ComponentName:        "BAICAPIL",
				ComponentType:        "raw_material",
				FormulaQty:           decimal.MustQuantity("1"),
				FormulaUOMCode:       "G",
				RequiredQty:          decimal.MustQuantity("2"),
				RequiredUOMCode:      "G",
				RequiredStockBaseQty: decimal.MustQuantity("0.002"),
				StockBaseUOMCode:     "KG",
				AvailableQty:         decimal.MustQuantity("0.003"),
				ShortageQty:          decimal.MustQuantity("0"),
				PurchaseDraftQty:     decimal.MustQuantity("0.001"),
				PurchaseDraftUOMCode: "KG",
				IsStockManaged:       true,
				NeedsPurchase:        true,
			},
		},
		CreatedBy: "user-production",
	})
	if err != ErrProductionPlanInvalidShortage {
		t.Fatalf("error = %v, want ErrProductionPlanInvalidShortage", err)
	}
}

func TestPurchaseRequestDraftTransitionsToApprovedAndConvertedPO(t *testing.T) {
	plan := mustProductionPlanWithPurchaseDraft(t)
	submittedAt := time.Date(2026, 5, 5, 9, 0, 0, 0, time.UTC)
	approvedAt := submittedAt.Add(time.Hour)
	convertedAt := approvedAt.Add(time.Hour)

	submitted, err := plan.PurchaseDraft.Submit("user-purchase", submittedAt)
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	approved, err := submitted.Approve("user-manager", approvedAt)
	if err != nil {
		t.Fatalf("Approve() error = %v", err)
	}
	converted, err := approved.MarkConvertedToPO("user-purchase", convertedAt, "po-260505-0001", "PO-260505-0001")
	if err != nil {
		t.Fatalf("MarkConvertedToPO() error = %v", err)
	}

	if converted.Status != PurchaseRequestDraftStatusConvertedToPO {
		t.Fatalf("status = %q, want converted_to_po", converted.Status)
	}
	if converted.SubmittedBy != "user-purchase" || !converted.SubmittedAt.Equal(submittedAt) {
		t.Fatalf("submitted fields = %+v, want actor/time", converted)
	}
	if converted.ApprovedBy != "user-manager" || !converted.ApprovedAt.Equal(approvedAt) {
		t.Fatalf("approved fields = %+v, want actor/time", converted)
	}
	if converted.ConvertedPurchaseOrderID != "po-260505-0001" || converted.ConvertedPurchaseOrderNo != "PO-260505-0001" {
		t.Fatalf("converted PO fields = %+v, want PO reference", converted)
	}
}

func TestPurchaseRequestDraftRejectsConvertBeforeApproval(t *testing.T) {
	plan := mustProductionPlanWithPurchaseDraft(t)

	_, err := plan.PurchaseDraft.MarkConvertedToPO("user-purchase", time.Now().UTC(), "po-260505-0001", "PO-260505-0001")
	if err != ErrProductionPlanInvalidPurchaseRequestTransition {
		t.Fatalf("error = %v, want ErrProductionPlanInvalidPurchaseRequestTransition", err)
	}
}

func mustProductionPlanWithPurchaseDraft(t *testing.T) ProductionPlan {
	t.Helper()
	plan, err := NewProductionPlanDocument(NewProductionPlanDocumentInput{
		ID:                  "plan-001",
		OrgID:               "org-my-pham",
		PlanNo:              "PP-260504-0001",
		OutputItemID:        "item-xff-150",
		OutputSKU:           "XFF",
		OutputItemName:      "Tinh chat buoi Fast & Furious 150ML",
		OutputItemType:      "finished_good",
		PlannedQty:          decimal.MustQuantity("162"),
		UOMCode:             "PCS",
		FormulaID:           "formula-xff-v1",
		FormulaCode:         "XFF-150ML",
		FormulaVersion:      "v1",
		FormulaBatchQty:     decimal.MustQuantity("81"),
		FormulaBatchUOMCode: "PCS",
		Lines: []NewProductionPlanLineInput{
			{
				ID:                   "plan-line-001",
				FormulaLineID:        "formula-line-001",
				LineNo:               1,
				ComponentItemID:      "item-act-baicapil",
				ComponentSKU:         "ACT_BAICAPIL",
				ComponentName:        "BAICAPIL",
				ComponentType:        "raw_material",
				FormulaQty:           decimal.MustQuantity("1"),
				FormulaUOMCode:       "G",
				RequiredQty:          decimal.MustQuantity("162"),
				RequiredUOMCode:      "G",
				RequiredStockBaseQty: decimal.MustQuantity("0.162"),
				StockBaseUOMCode:     "KG",
				AvailableQty:         decimal.MustQuantity("0.0005"),
				ShortageQty:          decimal.MustQuantity("0.1615"),
				PurchaseDraftQty:     decimal.MustQuantity("0.1615"),
				PurchaseDraftUOMCode: "KG",
				IsStockManaged:       true,
				NeedsPurchase:        true,
			},
		},
		CreatedAt: time.Date(2026, 5, 4, 10, 0, 0, 0, time.UTC),
		CreatedBy: "user-production",
	})
	if err != nil {
		t.Fatalf("NewProductionPlanDocument() error = %v", err)
	}

	return plan
}
