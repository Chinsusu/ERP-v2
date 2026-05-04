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
				RequiredQty:          decimal.MustQuantity("2"),
				RequiredUOMCode:      "G",
				RequiredStockBaseQty: decimal.MustQuantity("0.002"),
				StockBaseUOMCode:     "KG",
				AvailableQty:         decimal.MustQuantity("0.0005"),
				ShortageQty:          decimal.MustQuantity("0.0015"),
				PurchaseDraftQty:     decimal.MustQuantity("0.0015"),
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
	if line.RequiredStockBaseQty != "0.002000" || line.AvailableQty != "0.000500" || line.ShortageQty != "0.001500" {
		t.Fatalf("line quantities = %+v, want required 0.002 KG, available 0.0005 KG, shortage 0.0015 KG", line)
	}
	if !line.NeedsPurchase || line.PurchaseDraftQty != "0.001500" || line.PurchaseDraftUOMCode != "KG" {
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
