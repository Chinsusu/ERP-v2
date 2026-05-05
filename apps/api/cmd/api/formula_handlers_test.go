package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	masterdataapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestFormulaHandlersCreateListAndCalculateRequirement(t *testing.T) {
	authConfig := smokeAuthConfig()
	catalog := masterdataapp.NewPrototypeFormulaCatalog(audit.NewInMemoryLogStore())
	createBody := bytes.NewBufferString(`{
		"formula_code":"XFF-150ML",
		"finished_item_id":"item-xff-150",
		"finished_sku":"XFF",
		"finished_item_name":"Tính chất bưởi Fast & Furious 150ML",
		"finished_item_type":"finished_good",
		"formula_version":"v1",
		"batch_qty":"81",
		"batch_uom_code":"PCS",
		"base_batch_qty":"81",
		"base_batch_uom_code":"PCS",
		"lines":[
			{
				"line_no":1,
				"component_item_id":"item-act-baicapil",
				"component_sku":"ACT_BAICAPIL",
				"component_name":"BAICAPIL",
				"component_type":"raw_material",
				"entered_qty":"0,001",
				"entered_uom_code":"KG",
				"calc_qty":"1",
				"calc_uom_code":"G",
				"stock_base_qty":"0,001",
				"stock_base_uom_code":"KG",
				"waste_percent":"0",
				"is_required":true,
				"is_stock_managed":true
			}
		]
	}`)
	createReq := masterDataFormulaRequest(httptest.NewRequest(http.MethodPost, "/api/v1/formulas", createBody), authConfig)
	createReq.Header.Set(response.HeaderRequestID, "req-formula-create")
	createRec := httptest.NewRecorder()

	formulasHandler(catalog).ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d: %s", createRec.Code, http.StatusCreated, createRec.Body.String())
	}
	createPayload := decodeSmokeSuccess[formulaResponse](t, createRec)
	if createPayload.Data.FinishedSKU != "XFF" || len(createPayload.Data.Lines) != 1 {
		t.Fatalf("create payload = %+v, want XFF formula with one line", createPayload.Data)
	}
	if createPayload.Data.Lines[0].EnteredQty != "0.001000" || createPayload.Data.Lines[0].CalcUOMCode != "G" {
		t.Fatalf("line quantity = %+v, want decimal-normalized KG entered and gram display basis", createPayload.Data.Lines[0])
	}

	listReq := masterDataFormulaRequest(httptest.NewRequest(http.MethodGet, "/api/v1/formulas?q=xff", nil), authConfig)
	listRec := httptest.NewRecorder()

	formulasHandler(catalog).ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d: %s", listRec.Code, http.StatusOK, listRec.Body.String())
	}
	listPayload := decodeSmokeSuccess[[]formulaResponse](t, listRec)
	if len(listPayload.Data) != 1 || listPayload.Data[0].ID != createPayload.Data.ID {
		t.Fatalf("list payload = %+v, want created formula", listPayload.Data)
	}

	calcBody := bytes.NewBufferString(`{"planned_qty":"162","planned_uom_code":"PCS"}`)
	calcReq := masterDataFormulaRequest(httptest.NewRequest(http.MethodPost, "/api/v1/formulas/"+createPayload.Data.ID+"/calculate-requirement", calcBody), authConfig)
	calcReq.SetPathValue("formula_id", createPayload.Data.ID)
	calcRec := httptest.NewRecorder()

	calculateFormulaRequirementHandler(catalog).ServeHTTP(calcRec, calcReq)

	if calcRec.Code != http.StatusOK {
		t.Fatalf("calculate status = %d, want %d: %s", calcRec.Code, http.StatusOK, calcRec.Body.String())
	}
	calcPayload := decodeSmokeSuccess[formulaRequirementResponse](t, calcRec)
	if len(calcPayload.Data.Requirements) != 1 || calcPayload.Data.Requirements[0].RequiredStockBaseQty != "0.162000" {
		t.Fatalf("requirement payload = %+v, want 0.162000 KG stock-base requirement", calcPayload.Data)
	}
}

func masterDataFormulaRequest(req *http.Request, authConfig auth.MockConfig) *http.Request {
	return req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(authConfig, auth.RoleERPAdmin)))
}
