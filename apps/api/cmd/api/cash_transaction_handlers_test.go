package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	financeapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
)

func TestCashTransactionHandlersCreateListAndDetail(t *testing.T) {
	service := newTestCashTransactionHandlerService()
	createBody := []byte(`{
		"id": "cash-handler-0001",
		"transaction_no": "CASH-IN-HANDLER-0001",
		"direction": "cash_in",
		"business_date": "2026-04-30",
		"counterparty_id": "carrier-ghn",
		"counterparty_name": "GHN COD",
		"payment_method": "bank_transfer",
		"reference_no": "BANK-COD-HANDLER-0001",
		"total_amount": "1250000.00",
		"currency_code": "VND",
		"memo": "COD bank receipt",
		"allocations": [
			{
				"id": "cash-handler-0001-line-1",
				"target_type": "customer_receivable",
				"target_id": "ar-cod-260430-0001",
				"target_no": "AR-COD-260430-0001",
				"amount": "1000000.00"
			},
			{
				"id": "cash-handler-0001-line-2",
				"target_type": "cod_remittance",
				"target_id": "cod-remit-260430-0001",
				"target_no": "COD-REMIT-260430-0001",
				"amount": "250000.00"
			}
		]
	}`)
	createReq := cashTransactionRequest(http.MethodPost, "/api/v1/cash-transactions", createBody, auth.RoleFinanceOps)
	createRec := httptest.NewRecorder()

	cashTransactionsHandler(service).ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", createRec.Code, createRec.Body.String())
	}
	var createPayload struct {
		Data cashTransactionResponse `json:"data"`
	}
	if err := json.NewDecoder(createRec.Body).Decode(&createPayload); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	if createPayload.Data.ID != "cash-handler-0001" ||
		createPayload.Data.Status != "posted" ||
		createPayload.Data.Direction != "cash_in" ||
		createPayload.Data.AuditLogID == "" {
		t.Fatalf("create payload = %+v", createPayload.Data)
	}

	listReq := cashTransactionRequest(http.MethodGet, "/api/v1/cash-transactions?direction=cash_in&status=posted&q=HANDLER", nil, auth.RoleFinanceOps)
	listRec := httptest.NewRecorder()
	cashTransactionsHandler(service).ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d, body = %s", listRec.Code, listRec.Body.String())
	}
	var listPayload struct {
		Data []cashTransactionListItemResponse `json:"data"`
	}
	if err := json.NewDecoder(listRec.Body).Decode(&listPayload); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listPayload.Data) != 1 || listPayload.Data[0].ID != "cash-handler-0001" {
		t.Fatalf("list payload = %+v", listPayload.Data)
	}

	detailReq := cashTransactionRequest(http.MethodGet, "/api/v1/cash-transactions/cash-handler-0001", nil, auth.RoleFinanceOps)
	detailReq.SetPathValue("cash_transaction_id", "cash-handler-0001")
	detailRec := httptest.NewRecorder()
	cashTransactionDetailHandler(service).ServeHTTP(detailRec, detailReq)
	if detailRec.Code != http.StatusOK {
		t.Fatalf("detail status = %d, body = %s", detailRec.Code, detailRec.Body.String())
	}
	var detailPayload struct {
		Data cashTransactionResponse `json:"data"`
	}
	if err := json.NewDecoder(detailRec.Body).Decode(&detailPayload); err != nil {
		t.Fatalf("decode detail: %v", err)
	}
	if len(detailPayload.Data.Allocations) != 2 || detailPayload.Data.TotalAmount != "1250000.00" {
		t.Fatalf("detail payload = %+v", detailPayload.Data)
	}
}

func TestCashTransactionCreateRequiresFinanceManagePermission(t *testing.T) {
	service := newTestCashTransactionHandlerService()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cash-transactions", bytes.NewReader([]byte(`{}`)))
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{
		UserID: "finance-viewer",
		Email:  "finance-viewer@example.local",
		Name:   "Finance Viewer",
		Role:   auth.RoleFinanceOps,
		Permissions: []auth.PermissionKey{
			auth.PermissionFinanceView,
		},
	}))
	rec := httptest.NewRecorder()

	cashTransactionsHandler(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want forbidden", rec.Code)
	}
}

func newTestCashTransactionHandlerService() financeapp.CashTransactionService {
	return financeapp.NewCashTransactionService(
		financeapp.NewPrototypeCashTransactionStore(),
		audit.NewInMemoryLogStore(),
	)
}

func cashTransactionRequest(
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
