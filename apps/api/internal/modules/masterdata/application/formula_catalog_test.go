package application

import (
	"context"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestFormulaCatalogCreatesAndListsFormula(t *testing.T) {
	catalog := NewPrototypeFormulaCatalogAt(audit.NewInMemoryLogStore(), fixedFormulaCatalogTime())

	result, err := catalog.Create(context.Background(), formulaCreateInputForTest("V1", "3", "G", "3000", "0.003000"))
	if err != nil {
		t.Fatalf("create formula: %v", err)
	}
	if result.Formula.FormulaCode != "FORMULA-XFF-V1" {
		t.Fatalf("formula code = %q, want FORMULA-XFF-V1", result.Formula.FormulaCode)
	}

	formulas, err := catalog.List(context.Background(), domain.FormulaFilter{FinishedItemID: "item-xff"})
	if err != nil {
		t.Fatalf("list formulas: %v", err)
	}
	if len(formulas) != 1 {
		t.Fatalf("formulas = %d, want 1", len(formulas))
	}
	if formulas[0].Lines[0].CalcQty.String() != "3000.000000" {
		t.Fatalf("calc qty = %s, want 3000.000000", formulas[0].Lines[0].CalcQty)
	}
}

func TestFormulaCatalogActivationDeactivatesPreviousActiveVersion(t *testing.T) {
	catalog := NewPrototypeFormulaCatalogAt(audit.NewInMemoryLogStore(), fixedFormulaCatalogTime())
	first, err := catalog.Create(context.Background(), formulaCreateInputForTest("V1", "3", "G", "3000", "0.003000"))
	if err != nil {
		t.Fatalf("create first: %v", err)
	}
	if _, err := catalog.Activate(context.Background(), ActivateFormulaInput{
		ID:        first.Formula.ID,
		ActorID:   "user-admin",
		RequestID: "req-activate-v1",
	}); err != nil {
		t.Fatalf("activate first: %v", err)
	}
	second, err := catalog.Create(context.Background(), formulaCreateInputForTest("V2", "4", "G", "4000", "0.004000"))
	if err != nil {
		t.Fatalf("create second: %v", err)
	}
	if _, err := catalog.Activate(context.Background(), ActivateFormulaInput{
		ID:        second.Formula.ID,
		ActorID:   "user-admin",
		RequestID: "req-activate-v2",
	}); err != nil {
		t.Fatalf("activate second: %v", err)
	}

	reloadedFirst, err := catalog.Get(context.Background(), first.Formula.ID)
	if err != nil {
		t.Fatalf("get first: %v", err)
	}
	reloadedSecond, err := catalog.Get(context.Background(), second.Formula.ID)
	if err != nil {
		t.Fatalf("get second: %v", err)
	}
	if reloadedFirst.Status != domain.FormulaStatusInactive {
		t.Fatalf("first status = %s, want inactive", reloadedFirst.Status)
	}
	if reloadedSecond.Status != domain.FormulaStatusActive {
		t.Fatalf("second status = %s, want active", reloadedSecond.Status)
	}
}

func TestFormulaCatalogCalculatesRequirementPreview(t *testing.T) {
	catalog := NewPrototypeFormulaCatalogAt(audit.NewInMemoryLogStore(), fixedFormulaCatalogTime())
	result, err := catalog.Create(context.Background(), formulaCreateInputForTest("V1", "3", "G", "3000", "0.003000"))
	if err != nil {
		t.Fatalf("create formula: %v", err)
	}

	preview, err := catalog.CalculateRequirement(context.Background(), CalculateFormulaRequirementInput{
		ID:             result.Formula.ID,
		PlannedQty:     decimal.MustQuantity("162"),
		PlannedUOMCode: "PCS",
	})
	if err != nil {
		t.Fatalf("calculate requirement: %v", err)
	}
	if len(preview.Requirements) != 1 {
		t.Fatalf("requirements = %d, want 1", len(preview.Requirements))
	}
	if got := preview.Requirements[0].RequiredCalcQty.String(); got != "6000.000000" {
		t.Fatalf("required calc qty = %s, want 6000.000000", got)
	}
	if got := preview.Requirements[0].RequiredStockBaseQty.String(); got != "0.006000" {
		t.Fatalf("required stock base qty = %s, want 0.006000", got)
	}
}

func fixedFormulaCatalogTime() time.Time {
	return time.Date(2026, 5, 3, 10, 0, 0, 0, time.UTC)
}

func formulaCreateInputForTest(version string, enteredQty string, enteredUOM string, calcQty string, stockBaseQty string) CreateFormulaInput {
	return CreateFormulaInput{
		FormulaCode:      "FORMULA-XFF-" + version,
		FinishedItemID:   "item-xff",
		FinishedSKU:      "XFF",
		FinishedItemName: "Tinh chat buoi Fast & Furious 150ml",
		FinishedItemType: string(domain.ItemTypeFinishedGood),
		FormulaVersion:   version,
		BatchQty:         decimal.MustQuantity("81"),
		BatchUOMCode:     "PCS",
		BaseBatchQty:     decimal.MustQuantity("81"),
		BaseBatchUOMCode: "PCS",
		Lines: []CreateFormulaLineInput{
			{
				LineNo:           1,
				ComponentItemID:  "item-moi-pg",
				ComponentSKU:     "MOI_PG",
				ComponentName:    "PROPYLENE GLYCOL USP/EP",
				ComponentType:    string(domain.FormulaComponentRawMaterial),
				EnteredQty:       decimal.MustQuantity(enteredQty),
				EnteredUOMCode:   enteredUOM,
				CalcQty:          decimal.MustQuantity(calcQty),
				CalcUOMCode:      "MG",
				StockBaseQty:     decimal.MustQuantity(stockBaseQty),
				StockBaseUOMCode: "KG",
				IsRequired:       true,
				IsStockManaged:   true,
			},
		},
		ActorID:   "user-admin",
		RequestID: "req-create-" + version,
	}
}
