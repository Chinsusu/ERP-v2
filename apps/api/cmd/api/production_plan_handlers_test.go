package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	masterdataapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/application"
	productionapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestProductionPlanHandlersCreateDemandAndPurchaseRequestDraft(t *testing.T) {
	ctx := context.Background()
	auditLog := audit.NewInMemoryLogStore()
	formulas := masterdataapp.NewPrototypeFormulaCatalog(auditLog)
	formulaResult, err := formulas.Create(ctx, masterdataapp.CreateFormulaInput{
		FormulaCode:      "XFF-150ML",
		FinishedItemID:   "item-xff-150",
		FinishedSKU:      "XFF",
		FinishedItemName: "Tinh chat buoi Fast & Furious 150ML",
		FinishedItemType: "finished_good",
		FormulaVersion:   "v1",
		BatchQty:         decimal.MustQuantity("81"),
		BatchUOMCode:     "PCS",
		BaseBatchQty:     decimal.MustQuantity("81"),
		BaseBatchUOMCode: "PCS",
		Lines: []masterdataapp.CreateFormulaLineInput{
			{
				LineNo:           1,
				ComponentItemID:  "item-act-baicapil",
				ComponentSKU:     "ACT_BAICAPIL",
				ComponentName:    "BAICAPIL",
				ComponentType:    "raw_material",
				EnteredQty:       decimal.MustQuantity("0.001"),
				EnteredUOMCode:   "KG",
				CalcQty:          decimal.MustQuantity("1"),
				CalcUOMCode:      "G",
				StockBaseQty:     decimal.MustQuantity("0.001"),
				StockBaseUOMCode: "KG",
				WastePercent:     decimal.MustRate("0"),
				IsRequired:       true,
				IsStockManaged:   true,
			},
		},
		ActorID:   "user-erp-admin",
		RequestID: "req-formula-create",
	})
	if err != nil {
		t.Fatalf("create formula: %v", err)
	}
	if _, err := formulas.Activate(ctx, masterdataapp.ActivateFormulaInput{
		ID:        formulaResult.Formula.ID,
		ActorID:   "user-erp-admin",
		RequestID: "req-formula-activate",
	}); err != nil {
		t.Fatalf("activate formula: %v", err)
	}
	service := productionapp.NewProductionPlanService(
		productionapp.NewPrototypeProductionPlanStore(auditLog),
		formulas,
		fakeProductionPlanHandlerAvailableStock{
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
	body := bytes.NewBufferString(`{
		"id":"plan-001",
		"plan_no":"PP-260504-0001",
		"output_item_id":"item-xff-150",
		"formula_id":"` + formulaResult.Formula.ID + `",
		"planned_qty":"162",
		"uom_code":"PCS",
		"planned_start_date":"2026-05-10",
		"planned_end_date":"2026-05-11"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/production-plans", body)
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(smokeAuthConfig(), auth.RoleERPAdmin)))
	req.Header.Set(response.HeaderRequestID, "req-production-plan")
	rec := httptest.NewRecorder()

	productionPlansHandler(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	payload := decodeSmokeSuccess[productionPlanResponse](t, rec)
	if payload.Data.FormulaID != formulaResult.Formula.ID || len(payload.Data.Lines) != 1 {
		t.Fatalf("plan response = %+v, want formula snapshot and one line", payload.Data)
	}
	if payload.Data.Lines[0].ShortageQty != "0.001500" {
		t.Fatalf("shortage = %s, want 0.001500", payload.Data.Lines[0].ShortageQty)
	}
	if len(payload.Data.PurchaseRequestDraft.Lines) != 1 {
		t.Fatalf("purchase request draft = %+v, want one draft line", payload.Data.PurchaseRequestDraft)
	}
	if payload.Data.PurchaseRequestDraft.Lines[0].RequestedQty != "0.001500" {
		t.Fatalf("requested qty = %s, want 0.001500", payload.Data.PurchaseRequestDraft.Lines[0].RequestedQty)
	}
}

type fakeProductionPlanHandlerAvailableStock struct {
	rows []inventorydomain.AvailableStockSnapshot
}

func (s fakeProductionPlanHandlerAvailableStock) Execute(_ context.Context, filter inventorydomain.AvailableStockFilter) ([]inventorydomain.AvailableStockSnapshot, error) {
	matches := make([]inventorydomain.AvailableStockSnapshot, 0, len(s.rows))
	for _, row := range s.rows {
		if filter.SKU != "" && row.SKU != filter.SKU {
			continue
		}
		matches = append(matches, row)
	}

	return matches, nil
}
