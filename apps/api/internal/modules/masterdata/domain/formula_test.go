package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestNewFormulaAcceptsFormulaWithMilligramLine(t *testing.T) {
	formula, err := NewFormula(NewFormulaInput{
		ID:               "formula-xff150-v1",
		FormulaCode:      "FORMULA-XFF150-V1",
		FinishedItemID:   "item-xff150",
		FinishedSKU:      "XFF",
		FinishedItemName: "Tinh chat buoi Fast & Furious 150ml",
		FinishedItemType: ItemTypeFinishedGood,
		FormulaVersion:   "V1",
		BatchQty:         decimal.MustQuantity("81"),
		BatchUOMCode:     "PCS",
		BaseBatchQty:     decimal.MustQuantity("81"),
		BaseBatchUOMCode: "PCS",
		Status:           FormulaStatusDraft,
		ApprovalStatus:   FormulaApprovalDraft,
		Lines: []NewFormulaLineInput{
			{
				ID:               "formula-xff150-v1-line-1",
				LineNo:           1,
				ComponentItemID:  "item-act-baicapil",
				ComponentSKU:     "ACT_BAICAPIL",
				ComponentName:    "BAICAPIL",
				ComponentType:    FormulaComponentRawMaterial,
				EnteredQty:       decimal.MustQuantity("1"),
				EnteredUOMCode:   "MG",
				CalcQty:          decimal.MustQuantity("1"),
				CalcUOMCode:      "MG",
				StockBaseQty:     decimal.MustQuantity("0.000001"),
				StockBaseUOMCode: "KG",
				IsRequired:       true,
				IsStockManaged:   true,
				LineStatus:       FormulaLineStatusActive,
			},
		},
		CreatedAt: time.Date(2026, 5, 3, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 3, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("new formula: %v", err)
	}
	if formula.BatchQty.String() != "81.000000" {
		t.Fatalf("batch qty = %s, want 81.000000", formula.BatchQty)
	}
	if got := formula.Lines[0].StockBaseQty.String(); got != "0.000001" {
		t.Fatalf("stock base qty = %s, want 0.000001", got)
	}
	if got := formula.Lines[0].EnteredUOMCode.String(); got != "MG" {
		t.Fatalf("entered uom = %s, want MG", got)
	}
}

func TestFormulaActivationValidationRejectsRequiredZeroQuantity(t *testing.T) {
	formula := validFormulaForTest(t)
	formula.Lines[0].EnteredQty = decimal.MustQuantity("0")
	formula.Lines[0].CalcQty = decimal.MustQuantity("0")
	formula.Lines[0].StockBaseQty = decimal.MustQuantity("0")

	if err := formula.ValidateForActivation(); !errors.Is(err, ErrFormulaInvalidLineQuantity) {
		t.Fatalf("ValidateForActivation() error = %v, want invalid line quantity", err)
	}
}

func TestFormulaRequirementScalesLineQuantityPerFinishedUnit(t *testing.T) {
	formula := validFormulaForTest(t)
	formula.Lines[0].EnteredQty = decimal.MustQuantity("3")
	formula.Lines[0].EnteredUOMCode = decimal.MustUOMCode("G")
	formula.Lines[0].CalcQty = decimal.MustQuantity("3000")
	formula.Lines[0].CalcUOMCode = decimal.MustUOMCode("MG")
	formula.Lines[0].StockBaseQty = decimal.MustQuantity("0.003000")
	formula.Lines[0].StockBaseUOMCode = decimal.MustUOMCode("KG")

	requirements, err := formula.CalculateRequirement(decimal.MustQuantity("162"), "PCS")
	if err != nil {
		t.Fatalf("calculate requirement: %v", err)
	}
	if len(requirements) != 1 {
		t.Fatalf("requirements = %d, want 1", len(requirements))
	}
	if got := requirements[0].RequiredCalcQty.String(); got != "486000.000000" {
		t.Fatalf("required calc qty = %s, want 486000.000000", got)
	}
	if got := requirements[0].RequiredStockBaseQty.String(); got != "0.486000" {
		t.Fatalf("required stock base qty = %s, want 0.486000", got)
	}
}

func TestFormulaRequirementUsesLineQuantityPerFinishedUnit(t *testing.T) {
	formula := validFormulaForTest(t)
	formula.BatchQty = decimal.MustQuantity("10")
	formula.BaseBatchQty = decimal.MustQuantity("10")
	formula.Lines[0].EnteredQty = decimal.MustQuantity("0.001")
	formula.Lines[0].EnteredUOMCode = decimal.MustUOMCode("KG")
	formula.Lines[0].CalcQty = decimal.MustQuantity("1")
	formula.Lines[0].CalcUOMCode = decimal.MustUOMCode("G")
	formula.Lines[0].StockBaseQty = decimal.MustQuantity("0.001000")
	formula.Lines[0].StockBaseUOMCode = decimal.MustUOMCode("KG")

	requirements, err := formula.CalculateRequirement(decimal.MustQuantity("999"), "PCS")
	if err != nil {
		t.Fatalf("calculate requirement: %v", err)
	}
	if len(requirements) != 1 {
		t.Fatalf("requirements = %d, want 1", len(requirements))
	}
	if got := requirements[0].RequiredCalcQty.String(); got != "999.000000" {
		t.Fatalf("required calc qty = %s, want 999.000000", got)
	}
	if got := requirements[0].RequiredStockBaseQty.String(); got != "0.999000" {
		t.Fatalf("required stock base qty = %s, want 0.999000", got)
	}
}

func TestFormulaAllowsOnlyFinishedOrSemiFinishedParent(t *testing.T) {
	input := validFormulaInputForTest()
	input.FinishedItemType = ItemTypeRawMaterial

	if _, err := NewFormula(input); !errors.Is(err, ErrFormulaInvalidFinishedItemType) {
		t.Fatalf("NewFormula() error = %v, want invalid finished item type", err)
	}
}

func validFormulaForTest(t *testing.T) Formula {
	t.Helper()

	formula, err := NewFormula(validFormulaInputForTest())
	if err != nil {
		t.Fatalf("new formula: %v", err)
	}

	return formula
}

func validFormulaInputForTest() NewFormulaInput {
	now := time.Date(2026, 5, 3, 10, 0, 0, 0, time.UTC)

	return NewFormulaInput{
		ID:               "formula-xff150-v1",
		FormulaCode:      "FORMULA-XFF150-V1",
		FinishedItemID:   "item-xff150",
		FinishedSKU:      "XFF",
		FinishedItemName: "Tinh chat buoi Fast & Furious 150ml",
		FinishedItemType: ItemTypeFinishedGood,
		FormulaVersion:   "V1",
		BatchQty:         decimal.MustQuantity("81"),
		BatchUOMCode:     "PCS",
		BaseBatchQty:     decimal.MustQuantity("81"),
		BaseBatchUOMCode: "PCS",
		Status:           FormulaStatusDraft,
		ApprovalStatus:   FormulaApprovalDraft,
		Lines: []NewFormulaLineInput{
			{
				ID:               "formula-xff150-v1-line-1",
				LineNo:           1,
				ComponentItemID:  "item-moi-pg",
				ComponentSKU:     "MOI_PG",
				ComponentName:    "PROPYLENE GLYCOL USP/EP",
				ComponentType:    FormulaComponentRawMaterial,
				EnteredQty:       decimal.MustQuantity("1"),
				EnteredUOMCode:   "MG",
				CalcQty:          decimal.MustQuantity("1"),
				CalcUOMCode:      "MG",
				StockBaseQty:     decimal.MustQuantity("0.000001"),
				StockBaseUOMCode: "KG",
				IsRequired:       true,
				IsStockManaged:   true,
				LineStatus:       FormulaLineStatusActive,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}
