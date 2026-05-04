package application

import (
	"context"
	"errors"
	"testing"
	"time"

	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	masterdatadomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestProductionPlanServiceCreatesFormulaSnapshotDemandAndPurchaseRequestDraft(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 4, 10, 0, 0, 0, time.UTC)
	formula := activeProductionPlanFormula(t)
	store := NewPrototypeProductionPlanStore(audit.NewInMemoryLogStore())
	service := NewProductionPlanService(
		store,
		fakeProductionPlanFormulaReader{formula: formula},
		fakeProductionPlanAvailableStock{
			rows: []inventorydomain.AvailableStockSnapshot{
				{
					ItemID:       "item-act-baicapil",
					SKU:          "ACT_BAICAPIL",
					BaseUOMCode:  decimal.MustUOMCode("KG"),
					AvailableQty: decimal.MustQuantity("0.000500"),
				},
			},
		},
	)
	service.clock = func() time.Time { return now }

	result, err := service.CreateProductionPlan(ctx, CreateProductionPlanInput{
		ID:               "plan-001",
		PlanNo:           "PP-260504-0001",
		OutputItemID:     "item-xff-150",
		PlannedQty:       "162",
		UOMCode:          "PCS",
		PlannedStartDate: "2026-05-10",
		PlannedEndDate:   "2026-05-11",
		ActorID:          "user-production",
		RequestID:        "req-production-plan",
	})
	if err != nil {
		t.Fatalf("CreateProductionPlan() error = %v", err)
	}

	plan := result.ProductionPlan
	if plan.FormulaID != formula.ID || plan.FormulaVersion != "v1" || len(plan.Lines) != 1 {
		t.Fatalf("plan snapshot = %+v, want active formula snapshot with one line", plan)
	}
	line := plan.Lines[0]
	if line.RequiredStockBaseQty != "0.002000" || line.AvailableQty != "0.000500" || line.ShortageQty != "0.001500" {
		t.Fatalf("line = %+v, want shortage from formula demand minus available stock", line)
	}
	if len(plan.PurchaseDraft.Lines) != 1 {
		t.Fatalf("purchase draft lines = %d, want 1", len(plan.PurchaseDraft.Lines))
	}
	prLine := plan.PurchaseDraft.Lines[0]
	if prLine.SKU != "ACT_BAICAPIL" || prLine.RequestedQty != "0.001500" || prLine.UOMCode != "KG" {
		t.Fatalf("purchase draft line = %+v, want shortage draft only", prLine)
	}
	if result.AuditLogID == "" {
		t.Fatalf("audit id is empty")
	}
	if store.Count() != 1 {
		t.Fatalf("stored plans = %d, want 1", store.Count())
	}
}

func TestProductionPlanServiceDoesNotCreatePurchaseDraftWhenEnoughStock(t *testing.T) {
	ctx := context.Background()
	formula := activeProductionPlanFormula(t)
	store := NewPrototypeProductionPlanStore(audit.NewInMemoryLogStore())
	service := NewProductionPlanService(
		store,
		fakeProductionPlanFormulaReader{formula: formula},
		fakeProductionPlanAvailableStock{
			rows: []inventorydomain.AvailableStockSnapshot{
				{
					ItemID:       "item-act-baicapil",
					SKU:          "ACT_BAICAPIL",
					BaseUOMCode:  decimal.MustUOMCode("KG"),
					AvailableQty: decimal.MustQuantity("0.010000"),
				},
			},
		},
	)

	result, err := service.CreateProductionPlan(ctx, CreateProductionPlanInput{
		ID:           "plan-002",
		PlanNo:       "PP-260504-0002",
		OutputItemID: "item-xff-150",
		PlannedQty:   "162",
		UOMCode:      "PCS",
		ActorID:      "user-production",
	})
	if err != nil {
		t.Fatalf("CreateProductionPlan() error = %v", err)
	}
	if len(result.ProductionPlan.PurchaseDraft.Lines) != 0 {
		t.Fatalf("purchase draft = %+v, want no draft when stock is enough", result.ProductionPlan.PurchaseDraft)
	}
	if result.ProductionPlan.Lines[0].NeedsPurchase || result.ProductionPlan.Lines[0].ShortageQty != "0.000000" {
		t.Fatalf("line = %+v, want no shortage", result.ProductionPlan.Lines[0])
	}
}

func activeProductionPlanFormula(t *testing.T) masterdatadomain.Formula {
	t.Helper()
	formula, err := masterdatadomain.NewFormula(masterdatadomain.NewFormulaInput{
		ID:               "formula-xff-v1",
		FormulaCode:      "XFF-150ML",
		FinishedItemID:   "item-xff-150",
		FinishedSKU:      "XFF",
		FinishedItemName: "Tinh chat buoi Fast & Furious 150ML",
		FinishedItemType: masterdatadomain.ItemTypeFinishedGood,
		FormulaVersion:   "v1",
		BatchQty:         decimal.MustQuantity("81"),
		BatchUOMCode:     "PCS",
		BaseBatchQty:     decimal.MustQuantity("81"),
		BaseBatchUOMCode: "PCS",
		Status:           masterdatadomain.FormulaStatusActive,
		ApprovalStatus:   masterdatadomain.FormulaApprovalApproved,
		Lines: []masterdatadomain.NewFormulaLineInput{
			{
				ID:               "formula-line-001",
				LineNo:           1,
				ComponentItemID:  "item-act-baicapil",
				ComponentSKU:     "ACT_BAICAPIL",
				ComponentName:    "BAICAPIL",
				ComponentType:    masterdatadomain.FormulaComponentRawMaterial,
				EnteredQty:       decimal.MustQuantity("0.001"),
				EnteredUOMCode:   "KG",
				CalcQty:          decimal.MustQuantity("1"),
				CalcUOMCode:      "G",
				StockBaseQty:     decimal.MustQuantity("0.001"),
				StockBaseUOMCode: "KG",
				WastePercent:     decimal.MustRate("0"),
				IsRequired:       true,
				IsStockManaged:   true,
				LineStatus:       masterdatadomain.FormulaLineStatusActive,
			},
		},
	})
	if err != nil {
		t.Fatalf("formula fixture: %v", err)
	}

	return formula
}

type fakeProductionPlanFormulaReader struct {
	formula masterdatadomain.Formula
	err     error
}

func (r fakeProductionPlanFormulaReader) Get(ctx context.Context, id string) (masterdatadomain.Formula, error) {
	if r.err != nil {
		return masterdatadomain.Formula{}, r.err
	}
	if id != "" && id != r.formula.ID {
		return masterdatadomain.Formula{}, errors.New("formula not found")
	}

	return r.formula.Clone(), nil
}

func (r fakeProductionPlanFormulaReader) List(ctx context.Context, filter masterdatadomain.FormulaFilter) ([]masterdatadomain.Formula, error) {
	if r.err != nil {
		return nil, r.err
	}
	if filter.FinishedItemID != "" && filter.FinishedItemID != r.formula.FinishedItemID {
		return nil, nil
	}
	if filter.Status != "" && filter.Status != r.formula.Status {
		return nil, nil
	}

	return []masterdatadomain.Formula{r.formula.Clone()}, nil
}

type fakeProductionPlanAvailableStock struct {
	rows []inventorydomain.AvailableStockSnapshot
}

func (s fakeProductionPlanAvailableStock) Execute(ctx context.Context, filter inventorydomain.AvailableStockFilter) ([]inventorydomain.AvailableStockSnapshot, error) {
	matches := make([]inventorydomain.AvailableStockSnapshot, 0, len(s.rows))
	for _, row := range s.rows {
		if filter.SKU != "" && row.SKU != filter.SKU {
			continue
		}
		matches = append(matches, row)
	}

	return matches, nil
}

var _ ProductionPlanStore = (*PrototypeProductionPlanStore)(nil)
