package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
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
	if payload.Data.Lines[0].ShortageQty != "0.161500" {
		t.Fatalf("shortage = %s, want 0.161500", payload.Data.Lines[0].ShortageQty)
	}
	if len(payload.Data.PurchaseRequestDraft.Lines) != 1 {
		t.Fatalf("purchase request draft = %+v, want one draft line", payload.Data.PurchaseRequestDraft)
	}
	if payload.Data.PurchaseRequestDraft.Lines[0].RequestedQty != "0.161500" {
		t.Fatalf("requested qty = %s, want 0.161500", payload.Data.PurchaseRequestDraft.Lines[0].RequestedQty)
	}

	authConfig := smokeAuthConfig()
	draftID := payload.Data.PurchaseRequestDraft.ID
	detailReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodGet, "/api/v1/purchase-requests/"+draftID, nil),
		authConfig,
		auth.RolePurchaseOps,
	)
	detailReq.SetPathValue("purchase_request_id", draftID)
	detailRec := httptest.NewRecorder()

	purchaseRequestDetailHandler(service).ServeHTTP(detailRec, detailReq)

	if detailRec.Code != http.StatusOK {
		t.Fatalf("purchase request detail status = %d, want %d: %s", detailRec.Code, http.StatusOK, detailRec.Body.String())
	}
	detail := decodeSmokeSuccess[purchaseRequestDraftResponse](t, detailRec).Data
	if detail.ID != draftID || detail.SourceProductionPlanNo != "PP-260504-0001" {
		t.Fatalf("purchase request detail = %+v, want source plan link", detail)
	}

	submitReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/purchase-requests/"+draftID+"/submit", bytes.NewBufferString(`{}`)),
		authConfig,
		auth.RolePurchaseOps,
	)
	submitReq.SetPathValue("purchase_request_id", draftID)
	submitReq.Header.Set(response.HeaderRequestID, "req-purchase-request-submit")
	submitRec := httptest.NewRecorder()

	purchaseRequestSubmitHandler(service).ServeHTTP(submitRec, submitReq)

	if submitRec.Code != http.StatusOK {
		t.Fatalf("purchase request submit status = %d, want %d: %s", submitRec.Code, http.StatusOK, submitRec.Body.String())
	}
	submitted := decodeSmokeSuccess[purchaseRequestActionResponse](t, submitRec).Data
	if submitted.PreviousStatus != "draft" || submitted.CurrentStatus != "submitted" {
		t.Fatalf("submitted purchase request = %+v, want draft->submitted", submitted)
	}

	approveReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/purchase-requests/"+draftID+"/approve", bytes.NewBufferString(`{}`)),
		authConfig,
		auth.RolePurchaseOps,
	)
	approveReq.SetPathValue("purchase_request_id", draftID)
	approveReq.Header.Set(response.HeaderRequestID, "req-purchase-request-approve")
	approveRec := httptest.NewRecorder()

	purchaseRequestApproveHandler(service).ServeHTTP(approveRec, approveReq)

	if approveRec.Code != http.StatusOK {
		t.Fatalf("purchase request approve status = %d, want %d: %s", approveRec.Code, http.StatusOK, approveRec.Body.String())
	}
	approved := decodeSmokeSuccess[purchaseRequestActionResponse](t, approveRec).Data
	if approved.PreviousStatus != "submitted" || approved.CurrentStatus != "approved" {
		t.Fatalf("approved purchase request = %+v, want submitted->approved", approved)
	}

	purchaseOrderService, _ := newTestPurchaseOrderAPIService()
	convertReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/purchase-requests/"+draftID+"/convert-to-po", bytes.NewBufferString(`{
			"supplier_id":"sup-rm-bioactive",
			"warehouse_id":"wh-hcm-rm",
			"expected_date":"2026-05-12",
			"currency_code":"VND",
			"unit_price":"0"
		}`)),
		authConfig,
		auth.RolePurchaseOps,
	)
	convertReq.SetPathValue("purchase_request_id", draftID)
	convertReq.Header.Set(response.HeaderRequestID, "req-purchase-request-convert")
	convertRec := httptest.NewRecorder()

	purchaseRequestConvertToPOHandler(service, purchaseOrderService).ServeHTTP(convertRec, convertReq)

	if convertRec.Code != http.StatusCreated {
		t.Fatalf("purchase request convert status = %d, want %d: %s", convertRec.Code, http.StatusCreated, convertRec.Body.String())
	}
	converted := decodeSmokeSuccess[convertPurchaseRequestToPurchaseOrderResponse](t, convertRec).Data
	if converted.PurchaseRequest.Status != "converted_to_po" || converted.PurchaseRequest.ConvertedPurchaseOrderID == "" {
		t.Fatalf("converted purchase request = %+v, want converted PO reference", converted.PurchaseRequest)
	}
	if converted.PurchaseOrder.Status != "draft" || converted.PurchaseOrder.PONo == "" {
		t.Fatalf("converted PO = %+v, want draft PO", converted.PurchaseOrder)
	}
}

func TestProductionPlanHandlersCreateWarehouseIssueFromPlanLine(t *testing.T) {
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
	issueService := inventoryapp.NewWarehouseIssueService(
		inventoryapp.NewPrototypeWarehouseIssueStore(),
		inventoryapp.NewInMemoryStockMovementStore(),
		auditLog,
	)
	service := productionapp.NewProductionPlanService(
		productionapp.NewPrototypeProductionPlanStore(auditLog),
		formulas,
		fakeProductionPlanHandlerAvailableStock{
			rows: []inventorydomain.AvailableStockSnapshot{
				{
					WarehouseID:   "wh-hcm-rm",
					WarehouseCode: "WH-HCM-RM",
					LocationID:    "bin-rm-a01",
					LocationCode:  "RM-A01",
					ItemID:        "item-act-baicapil",
					SKU:           "ACT_BAICAPIL",
					BatchID:       "batch-act-baicapil-001",
					BatchNo:       "LOT-ACT-001",
					BaseUOMCode:   decimal.MustUOMCode("KG"),
					AvailableQty:  decimal.MustQuantity("1.000000"),
				},
			},
		},
	).WithWarehouseIssueService(issueService)
	body := bytes.NewBufferString(`{
		"id":"plan-issue-api",
		"plan_no":"PP-260505-ISSUE-API",
		"output_item_id":"item-xff-150",
		"formula_id":"` + formulaResult.Formula.ID + `",
		"planned_qty":"162",
		"uom_code":"PCS"
	}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/production-plans", body)
	createReq = createReq.WithContext(auth.WithPrincipal(createReq.Context(), auth.MockPrincipalForRole(smokeAuthConfig(), auth.RoleERPAdmin)))
	createRec := httptest.NewRecorder()

	productionPlansHandler(service).ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("create plan status = %d, want %d: %s", createRec.Code, http.StatusCreated, createRec.Body.String())
	}
	plan := decodeSmokeSuccess[productionPlanResponse](t, createRec).Data
	lineID := plan.Lines[0].ID
	issueReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/production-plans/plan-issue-api/warehouse-issues", bytes.NewBufferString(`{
			"line_ids":["`+lineID+`"],
			"destination_type":"factory",
			"destination_name":"Factory A",
			"reason_code":"production_plan_issue"
		}`)),
		smokeAuthConfig(),
		auth.RoleERPAdmin,
	)
	issueReq.SetPathValue("production_plan_id", "plan-issue-api")
	issueReq.Header.Set(response.HeaderRequestID, "req-production-plan-issue")
	issueRec := httptest.NewRecorder()

	productionPlanWarehouseIssuesHandler(service).ServeHTTP(issueRec, issueReq)

	if issueRec.Code != http.StatusCreated {
		t.Fatalf("create issue status = %d, want %d: %s", issueRec.Code, http.StatusCreated, issueRec.Body.String())
	}
	issue := decodeSmokeSuccess[warehouseIssueResponse](t, issueRec).Data
	if issue.Lines[0].SourceDocumentType != "production_plan" ||
		issue.Lines[0].SourceDocumentID != "plan-issue-api" ||
		issue.Lines[0].SourceDocumentLineID != lineID {
		t.Fatalf("issue line source = %+v, want production plan line source", issue.Lines[0])
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
