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

func TestSupplierPayableHandlersCreateListApproveAndRecordPayment(t *testing.T) {
	service := newTestSupplierPayableHandlerService()
	createBody := []byte(`{
		"id": "ap-handler-0001",
		"payable_no": "AP-HANDLER-0001",
		"supplier_id": "supplier-hcm-001",
		"supplier_code": "SUP-HCM-001",
		"supplier_name": "Nguyen Lieu HCM",
		"source_document": {"type": "qc_inspection", "id": "qc-handler-0001", "no": "QC-HANDLER-0001"},
		"total_amount": "4250000.00",
		"currency_code": "VND",
		"due_date": "2026-05-07",
		"lines": [
			{
				"id": "ap-handler-line-1",
				"description": "Accepted goods after inbound QC",
				"source_document": {"type": "warehouse_receipt", "id": "gr-handler-0001", "no": "GR-HANDLER-0001"},
				"amount": "4250000.00"
			}
		]
	}`)
	createReq := supplierPayableRequest(http.MethodPost, "/api/v1/supplier-payables", createBody, auth.RoleFinanceOps)
	createRec := httptest.NewRecorder()

	supplierPayablesHandler(service).ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", createRec.Code, createRec.Body.String())
	}
	var createPayload struct {
		Data supplierPayableResponse `json:"data"`
	}
	if err := json.NewDecoder(createRec.Body).Decode(&createPayload); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	if createPayload.Data.ID != "ap-handler-0001" || createPayload.Data.AuditLogID == "" {
		t.Fatalf("create payload = %+v", createPayload.Data)
	}

	listReq := supplierPayableRequest(http.MethodGet, "/api/v1/supplier-payables?status=open&q=AP-HANDLER", nil, auth.RoleFinanceOps)
	listRec := httptest.NewRecorder()
	supplierPayablesHandler(service).ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d, body = %s", listRec.Code, listRec.Body.String())
	}
	var listPayload struct {
		Data []supplierPayableListItemResponse `json:"data"`
	}
	if err := json.NewDecoder(listRec.Body).Decode(&listPayload); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listPayload.Data) != 1 || listPayload.Data[0].ID != "ap-handler-0001" {
		t.Fatalf("list payload = %+v", listPayload.Data)
	}

	approveReq := supplierPayableRequest(
		http.MethodPost,
		"/api/v1/supplier-payables/ap-handler-0001/approve-payment",
		[]byte(`{}`),
		auth.RoleFinanceOps,
	)
	approveReq.SetPathValue("supplier_payable_id", "ap-handler-0001")
	approveRec := httptest.NewRecorder()
	supplierPayableApprovePaymentHandler(service).ServeHTTP(approveRec, approveReq)
	if approveRec.Code != http.StatusOK {
		t.Fatalf("approve status = %d, body = %s", approveRec.Code, approveRec.Body.String())
	}
	var approvePayload struct {
		Data supplierPayableActionResultResponse `json:"data"`
	}
	if err := json.NewDecoder(approveRec.Body).Decode(&approvePayload); err != nil {
		t.Fatalf("decode approve: %v", err)
	}
	if approvePayload.Data.CurrentStatus != string(financedomain.PayableStatusPaymentApproved) {
		t.Fatalf("approve payload = %+v", approvePayload.Data)
	}

	paymentReq := supplierPayableRequest(
		http.MethodPost,
		"/api/v1/supplier-payables/ap-handler-0001/record-payment",
		[]byte(`{"amount":"1250000.00"}`),
		auth.RoleFinanceOps,
	)
	paymentReq.SetPathValue("supplier_payable_id", "ap-handler-0001")
	paymentRec := httptest.NewRecorder()
	supplierPayableRecordPaymentHandler(service).ServeHTTP(paymentRec, paymentReq)
	if paymentRec.Code != http.StatusOK {
		t.Fatalf("payment status = %d, body = %s", paymentRec.Code, paymentRec.Body.String())
	}
	var paymentPayload struct {
		Data supplierPayableActionResultResponse `json:"data"`
	}
	if err := json.NewDecoder(paymentRec.Body).Decode(&paymentPayload); err != nil {
		t.Fatalf("decode payment: %v", err)
	}
	if paymentPayload.Data.CurrentStatus != string(financedomain.PayableStatusPartiallyPaid) ||
		paymentPayload.Data.SupplierPayable.PaidAmount != "1250000.00" {
		t.Fatalf("payment payload = %+v", paymentPayload.Data)
	}
}

func TestSupplierPayableApproveRequiresPaymentApprovePermission(t *testing.T) {
	service := newTestSupplierPayableHandlerService()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/supplier-payables/ap-supplier-260430-0001/approve-payment", bytes.NewReader([]byte(`{}`)))
	req.SetPathValue("supplier_payable_id", "ap-supplier-260430-0001")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{
		UserID: "finance-manager-without-approval",
		Email:  "finance-manager@example.local",
		Name:   "Finance Manager",
		Role:   auth.RoleFinanceOps,
		Permissions: []auth.PermissionKey{
			auth.PermissionFinanceView,
			auth.PermissionFinanceManage,
		},
	}))
	rec := httptest.NewRecorder()

	supplierPayableApprovePaymentHandler(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want forbidden", rec.Code)
	}
}

func newTestSupplierPayableHandlerService() financeapp.SupplierPayableService {
	return financeapp.NewSupplierPayableService(
		financeapp.NewPrototypeSupplierPayableStore(),
		audit.NewInMemoryLogStore(),
	)
}

func supplierPayableRequest(
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
