package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	financeapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
)

func TestFinanceDashboardHandlerReturnsMetrics(t *testing.T) {
	service := newTestFinanceDashboardHandlerService()
	req := cashTransactionRequest(http.MethodGet, "/api/v1/finance/dashboard?business_date=2026-04-30", nil, auth.RoleFinanceOps)
	rec := httptest.NewRecorder()

	financeDashboardHandler(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Data financeDashboardResponse `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload.Data.BusinessDate != "2026-04-30" ||
		payload.Data.AR.OpenCount != 1 ||
		payload.Data.AP.OpenCount != 1 ||
		payload.Data.COD.PendingCount != 1 ||
		payload.Data.Cash.NetCashToday != "-3000000.00" {
		t.Fatalf("payload = %+v", payload.Data)
	}
}

func TestFinanceDashboardHandlerRejectsInvalidBusinessDate(t *testing.T) {
	service := newTestFinanceDashboardHandlerService()
	req := cashTransactionRequest(http.MethodGet, "/api/v1/finance/dashboard?business_date=30/04/2026", nil, auth.RoleFinanceOps)
	rec := httptest.NewRecorder()

	financeDashboardHandler(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want bad request, body = %s", rec.Code, rec.Body.String())
	}
}

func TestFinanceDashboardHandlerRequiresFinanceViewPermission(t *testing.T) {
	service := newTestFinanceDashboardHandlerService()
	req := cashTransactionRequest(http.MethodGet, "/api/v1/finance/dashboard", nil, auth.RoleWarehouseStaff)
	rec := httptest.NewRecorder()

	financeDashboardHandler(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want forbidden", rec.Code)
	}
}

func newTestFinanceDashboardHandlerService() financeapp.FinanceDashboardService {
	return financeapp.NewFinanceDashboardService(
		financeapp.NewPrototypeCustomerReceivableStore(),
		financeapp.NewPrototypeSupplierPayableStore(),
		financeapp.NewPrototypeCODRemittanceStore(),
		financeapp.NewPrototypeCashTransactionStore(),
	)
}
