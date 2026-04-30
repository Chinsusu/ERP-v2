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

func TestCustomerReceivableHandlersCreateListAndRecordReceipt(t *testing.T) {
	service := newTestCustomerReceivableHandlerService()
	createBody := []byte(`{
		"id": "ar-handler-0001",
		"receivable_no": "AR-HANDLER-0001",
		"customer_id": "customer-hcm-001",
		"customer_code": "KH-HCM-001",
		"customer_name": "My Pham HCM Retail",
		"source_document": {"type": "shipment", "id": "shipment-handler-0001", "no": "SHP-HANDLER-0001"},
		"total_amount": "1250000.00",
		"currency_code": "VND",
		"due_date": "2026-05-03",
		"lines": [
			{
				"id": "ar-handler-line-1",
				"description": "COD delivered goods",
				"source_document": {"type": "shipment", "id": "shipment-handler-0001", "no": "SHP-HANDLER-0001"},
				"amount": "1250000.00"
			}
		]
	}`)
	createReq := customerReceivableRequest(http.MethodPost, "/api/v1/customer-receivables", createBody, auth.RoleFinanceOps)
	createRec := httptest.NewRecorder()

	customerReceivablesHandler(service).ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", createRec.Code, createRec.Body.String())
	}
	var createPayload struct {
		Data customerReceivableResponse `json:"data"`
	}
	if err := json.NewDecoder(createRec.Body).Decode(&createPayload); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	if createPayload.Data.ID != "ar-handler-0001" || createPayload.Data.AuditLogID == "" {
		t.Fatalf("create payload = %+v", createPayload.Data)
	}

	listReq := customerReceivableRequest(http.MethodGet, "/api/v1/customer-receivables?status=open&search=AR-HANDLER", nil, auth.RoleFinanceOps)
	listRec := httptest.NewRecorder()
	customerReceivablesHandler(service).ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d, body = %s", listRec.Code, listRec.Body.String())
	}
	var listPayload struct {
		Data []customerReceivableListItemResponse `json:"data"`
	}
	if err := json.NewDecoder(listRec.Body).Decode(&listPayload); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listPayload.Data) != 1 || listPayload.Data[0].ID != "ar-handler-0001" {
		t.Fatalf("list payload = %+v", listPayload.Data)
	}

	receiptReq := customerReceivableRequest(
		http.MethodPost,
		"/api/v1/customer-receivables/ar-handler-0001/record-receipt",
		[]byte(`{"amount":"250000.00"}`),
		auth.RoleFinanceOps,
	)
	receiptReq.SetPathValue("customer_receivable_id", "ar-handler-0001")
	receiptRec := httptest.NewRecorder()
	customerReceivableRecordReceiptHandler(service).ServeHTTP(receiptRec, receiptReq)
	if receiptRec.Code != http.StatusOK {
		t.Fatalf("receipt status = %d, body = %s", receiptRec.Code, receiptRec.Body.String())
	}
	var receiptPayload struct {
		Data customerReceivableActionResultResponse `json:"data"`
	}
	if err := json.NewDecoder(receiptRec.Body).Decode(&receiptPayload); err != nil {
		t.Fatalf("decode receipt: %v", err)
	}
	if receiptPayload.Data.CurrentStatus != string(financedomain.ReceivableStatusPartiallyPaid) ||
		receiptPayload.Data.CustomerReceivable.PaidAmount != "250000.00" {
		t.Fatalf("receipt payload = %+v", receiptPayload.Data)
	}
}

func TestCustomerReceivableCreateRequiresFinanceManagePermission(t *testing.T) {
	service := newTestCustomerReceivableHandlerService()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/customer-receivables", bytes.NewReader([]byte(`{}`)))
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{
		UserID:      "finance-viewer",
		Email:       "finance-viewer@example.local",
		Name:        "Finance Viewer",
		Role:        auth.RoleFinanceOps,
		Permissions: []auth.PermissionKey{auth.PermissionFinanceView},
	}))
	rec := httptest.NewRecorder()

	customerReceivablesHandler(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want forbidden", rec.Code)
	}
}

func newTestCustomerReceivableHandlerService() financeapp.CustomerReceivableService {
	return financeapp.NewCustomerReceivableService(
		financeapp.NewPrototypeCustomerReceivableStore(),
		audit.NewInMemoryLogStore(),
	)
}

func customerReceivableRequest(
	method string,
	target string,
	body []byte,
	role auth.RoleKey,
) *http.Request {
	req := httptest.NewRequest(method, target, bytes.NewReader(body))
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, role)))

	return req
}
