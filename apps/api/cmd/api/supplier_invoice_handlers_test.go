package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	financeapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/application"
	financedomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
)

func TestSupplierInvoiceHandlersCreateListGetAndVoid(t *testing.T) {
	service := newTestSupplierInvoiceHandlerService()
	createBody := []byte(`{
		"id": "si-handler-0001",
		"invoice_no": "inv-handler-0001",
		"payable_id": "ap-supplier-260430-0001",
		"invoice_date": "2026-05-05",
		"invoice_amount": "4250000.00",
		"currency_code": "VND"
	}`)
	createReq := supplierPayableRequest(http.MethodPost, "/api/v1/supplier-invoices", createBody, auth.RoleFinanceOps)
	createRec := httptest.NewRecorder()

	supplierInvoicesHandler(service).ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", createRec.Code, createRec.Body.String())
	}
	var createPayload struct {
		Data supplierInvoiceResponse `json:"data"`
	}
	if err := json.NewDecoder(createRec.Body).Decode(&createPayload); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	if createPayload.Data.ID != "si-handler-0001" ||
		createPayload.Data.Status != string(financedomain.SupplierInvoiceStatusMatched) ||
		createPayload.Data.AuditLogID == "" {
		t.Fatalf("create payload = %+v", createPayload.Data)
	}

	listReq := supplierPayableRequest(http.MethodGet, "/api/v1/supplier-invoices?payable_id=ap-supplier-260430-0001&q=INV-HANDLER", nil, auth.RoleFinanceOps)
	listRec := httptest.NewRecorder()
	supplierInvoicesHandler(service).ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d, body = %s", listRec.Code, listRec.Body.String())
	}
	var listPayload struct {
		Data []supplierInvoiceListItemResponse `json:"data"`
	}
	if err := json.NewDecoder(listRec.Body).Decode(&listPayload); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listPayload.Data) != 1 || listPayload.Data[0].ID != "si-handler-0001" {
		t.Fatalf("list payload = %+v", listPayload.Data)
	}

	detailReq := supplierPayableRequest(http.MethodGet, "/api/v1/supplier-invoices/si-handler-0001", nil, auth.RoleFinanceOps)
	detailReq.SetPathValue("supplier_invoice_id", "si-handler-0001")
	detailRec := httptest.NewRecorder()
	supplierInvoiceDetailHandler(service).ServeHTTP(detailRec, detailReq)
	if detailRec.Code != http.StatusOK {
		t.Fatalf("detail status = %d, body = %s", detailRec.Code, detailRec.Body.String())
	}
	var detailPayload struct {
		Data supplierInvoiceResponse `json:"data"`
	}
	if err := json.NewDecoder(detailRec.Body).Decode(&detailPayload); err != nil {
		t.Fatalf("decode detail: %v", err)
	}
	if detailPayload.Data.PayableNo != "AP-SUP-260430-0001" || len(detailPayload.Data.Lines) == 0 {
		t.Fatalf("detail payload = %+v", detailPayload.Data)
	}

	voidReq := supplierPayableRequest(
		http.MethodPost,
		"/api/v1/supplier-invoices/si-handler-0001/void",
		[]byte(`{"reason":"duplicate invoice"}`),
		auth.RoleFinanceOps,
	)
	voidReq.SetPathValue("supplier_invoice_id", "si-handler-0001")
	voidRec := httptest.NewRecorder()
	supplierInvoiceVoidHandler(service).ServeHTTP(voidRec, voidReq)
	if voidRec.Code != http.StatusOK {
		t.Fatalf("void status = %d, body = %s", voidRec.Code, voidRec.Body.String())
	}
	var voidPayload struct {
		Data supplierInvoiceActionResultResponse `json:"data"`
	}
	if err := json.NewDecoder(voidRec.Body).Decode(&voidPayload); err != nil {
		t.Fatalf("decode void: %v", err)
	}
	if voidPayload.Data.CurrentStatus != string(financedomain.SupplierInvoiceStatusVoid) ||
		voidPayload.Data.SupplierInvoice.VoidReason != "duplicate invoice" {
		t.Fatalf("void payload = %+v", voidPayload.Data)
	}
}

func TestSupplierInvoiceCreateRequiresFinanceManagePermission(t *testing.T) {
	service := newTestSupplierInvoiceHandlerService()
	req := supplierPayableRequest(http.MethodPost, "/api/v1/supplier-invoices", []byte(`{}`), auth.RoleWarehouseStaff)
	rec := httptest.NewRecorder()

	supplierInvoicesHandler(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want forbidden", rec.Code)
	}
}

func newTestSupplierInvoiceHandlerService() financeapp.SupplierInvoiceService {
	return financeapp.NewSupplierInvoiceService(
		financeapp.NewPrototypeSupplierInvoiceStore(),
		financeapp.NewPrototypeSupplierPayableStore(),
		audit.NewInMemoryLogStore(),
	)
}
