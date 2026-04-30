package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	financeapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/application"
	financedomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
)

func TestCODRemittanceHandlersCreateListAndMatch(t *testing.T) {
	service := newTestCODRemittanceHandlerService()
	createBody := []byte(`{
		"id": "cod-handler-0001",
		"remittance_no": "COD-HANDLER-0001",
		"carrier_id": "carrier-ghn",
		"carrier_code": "GHN",
		"carrier_name": "GHN Express",
		"business_date": "2026-04-30",
		"expected_amount": "1250000.00",
		"remitted_amount": "1250000.00",
		"currency_code": "VND",
		"lines": [
			{
				"id": "cod-handler-line-1",
				"receivable_id": "ar-handler-0001",
				"receivable_no": "AR-HANDLER-0001",
				"shipment_id": "shipment-handler-0001",
				"tracking_no": "GHN-HANDLER-0001",
				"customer_name": "My Pham HCM Retail",
				"expected_amount": "1250000.00",
				"remitted_amount": "1250000.00"
			}
		]
	}`)
	createReq := codRemittanceRequest(http.MethodPost, "/api/v1/cod-remittances", createBody, auth.RoleFinanceOps)
	createRec := httptest.NewRecorder()

	codRemittancesHandler(service).ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", createRec.Code, createRec.Body.String())
	}
	var createPayload struct {
		Data codRemittanceResponse `json:"data"`
	}
	if err := json.NewDecoder(createRec.Body).Decode(&createPayload); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	if createPayload.Data.ID != "cod-handler-0001" || createPayload.Data.AuditLogID == "" {
		t.Fatalf("create payload = %+v", createPayload.Data)
	}

	listReq := codRemittanceRequest(http.MethodGet, "/api/v1/cod-remittances?status=draft&search=HANDLER", nil, auth.RoleFinanceOps)
	listRec := httptest.NewRecorder()
	codRemittancesHandler(service).ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d, body = %s", listRec.Code, listRec.Body.String())
	}
	var listPayload struct {
		Data []codRemittanceListItemResponse `json:"data"`
	}
	if err := json.NewDecoder(listRec.Body).Decode(&listPayload); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listPayload.Data) != 1 || listPayload.Data[0].ID != "cod-handler-0001" {
		t.Fatalf("list payload = %+v", listPayload.Data)
	}

	matchReq := codRemittanceRequest(http.MethodPost, "/api/v1/cod-remittances/cod-handler-0001/match", nil, auth.RoleFinanceOps)
	matchReq.SetPathValue("cod_remittance_id", "cod-handler-0001")
	matchRec := httptest.NewRecorder()
	codRemittanceMatchHandler(service).ServeHTTP(matchRec, matchReq)
	if matchRec.Code != http.StatusOK {
		t.Fatalf("match status = %d, body = %s", matchRec.Code, matchRec.Body.String())
	}
	var matchPayload struct {
		Data codRemittanceActionResultResponse `json:"data"`
	}
	if err := json.NewDecoder(matchRec.Body).Decode(&matchPayload); err != nil {
		t.Fatalf("decode match: %v", err)
	}
	if matchPayload.Data.CurrentStatus != string(financedomain.CODRemittanceStatusMatching) {
		t.Fatalf("match payload = %+v", matchPayload.Data)
	}
}

func TestCODRemittanceHandlersRecordDiscrepancyAndSubmit(t *testing.T) {
	service := newTestCODRemittanceHandlerService()
	discrepancyReq := codRemittanceRequest(
		http.MethodPost,
		"/api/v1/cod-remittances/cod-remit-260430-0001/record-discrepancy",
		[]byte(`{"id":"disc-handler-1","line_id":"cod-remit-260430-0001-line-1","reason":"carrier short","owner_id":"finance-user"}`),
		auth.RoleFinanceOps,
	)
	discrepancyReq.SetPathValue("cod_remittance_id", "cod-remit-260430-0001")
	discrepancyRec := httptest.NewRecorder()
	codRemittanceDiscrepancyHandler(service).ServeHTTP(discrepancyRec, discrepancyReq)
	if discrepancyRec.Code != http.StatusOK {
		t.Fatalf("discrepancy status = %d, body = %s", discrepancyRec.Code, discrepancyRec.Body.String())
	}

	submitReq := codRemittanceRequest(
		http.MethodPost,
		"/api/v1/cod-remittances/cod-remit-260430-0001/submit",
		nil,
		auth.RoleFinanceOps,
	)
	submitReq.SetPathValue("cod_remittance_id", "cod-remit-260430-0001")
	submitRec := httptest.NewRecorder()
	codRemittanceSubmitHandler(service).ServeHTTP(submitRec, submitReq)
	if submitRec.Code != http.StatusOK {
		t.Fatalf("submit status = %d, body = %s", submitRec.Code, submitRec.Body.String())
	}
	var submitPayload struct {
		Data codRemittanceActionResultResponse `json:"data"`
	}
	if err := json.NewDecoder(submitRec.Body).Decode(&submitPayload); err != nil {
		t.Fatalf("decode submit: %v", err)
	}
	if submitPayload.Data.CurrentStatus != string(financedomain.CODRemittanceStatusSubmitted) ||
		len(submitPayload.Data.CODRemittance.Discrepancies) != 1 {
		t.Fatalf("submit payload = %+v", submitPayload.Data)
	}
}

func TestCODRemittanceCreateRequiresCODReconcilePermission(t *testing.T) {
	service := newTestCODRemittanceHandlerService()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cod-remittances", bytes.NewReader([]byte(`{}`)))
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{
		UserID:      "finance-viewer",
		Email:       "finance-viewer@example.local",
		Name:        "Finance Viewer",
		Role:        auth.RoleFinanceOps,
		Permissions: []auth.PermissionKey{auth.PermissionFinanceView},
	}))
	rec := httptest.NewRecorder()

	codRemittancesHandler(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want forbidden", rec.Code)
	}
}

func newTestCODRemittanceHandlerService() financeapp.CODRemittanceService {
	return financeapp.NewCODRemittanceService(
		financeapp.NewPrototypeCODRemittanceStore(),
		audit.NewInMemoryLogStore(),
	)
}

func codRemittanceRequest(method string, target string, body []byte, role auth.RoleKey) *http.Request {
	req := httptest.NewRequest(method, target, bytes.NewReader(body))
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, role)))

	return req
}
