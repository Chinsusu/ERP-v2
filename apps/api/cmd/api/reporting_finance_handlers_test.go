package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	financeapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestFinanceSummaryReportHandlerReturnsReport(t *testing.T) {
	req := cashTransactionRequest(
		http.MethodGet,
		"/api/v1/reports/finance-summary?from_date=2026-04-30&to_date=2026-05-08&business_date=2026-05-08",
		nil,
		auth.RoleFinanceOps,
	)
	req.Header.Set(response.HeaderRequestID, "req-report-finance")
	rec := httptest.NewRecorder()

	newTestFinanceSummaryReportHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var payload response.SuccessEnvelope[financeSummaryReportResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.RequestID != "req-report-finance" {
		t.Fatalf("request id = %q, want req-report-finance", payload.RequestID)
	}
	if payload.Data.Metadata.Timezone != "Asia/Ho_Chi_Minh" ||
		payload.Data.Metadata.SourceVersion != "reporting-v1" ||
		payload.Data.Metadata.Filters.FromDate != "2026-04-30" ||
		payload.Data.Metadata.Filters.ToDate != "2026-05-08" ||
		payload.Data.Metadata.Filters.BusinessDate != "2026-05-08" ||
		payload.Data.CurrencyCode != "VND" {
		t.Fatalf("metadata = %+v currency = %q", payload.Data.Metadata, payload.Data.CurrencyCode)
	}
	if payload.Data.AR.OpenCount != 1 ||
		payload.Data.AR.OverdueCount != 1 ||
		payload.Data.AR.OpenAmount != "1250000.00" ||
		payload.Data.AR.OutstandingAmount != "1250000.00" {
		t.Fatalf("ar = %+v", payload.Data.AR)
	}
	if payload.Data.AP.OpenCount != 1 ||
		payload.Data.AP.DueCount != 1 ||
		payload.Data.AP.OpenAmount != "4250000.00" ||
		payload.Data.AP.OutstandingAmount != "4250000.00" {
		t.Fatalf("ap = %+v", payload.Data.AP)
	}
	if payload.Data.COD.PendingCount != 1 ||
		payload.Data.COD.DiscrepancyCount != 1 ||
		payload.Data.COD.PendingAmount != "2000000.00" ||
		payload.Data.COD.DiscrepancyAmount != "-50000.00" {
		t.Fatalf("cod = %+v", payload.Data.COD)
	}
	if payload.Data.Cash.TransactionCount != 2 ||
		payload.Data.Cash.CashInAmount != "1250000.00" ||
		payload.Data.Cash.CashOutAmount != "4250000.00" ||
		payload.Data.Cash.NetCashAmount != "-3000000.00" {
		t.Fatalf("cash = %+v", payload.Data.Cash)
	}
}

func TestFinanceSummaryReportHandlerRequiresFinanceReportPermission(t *testing.T) {
	req := cashTransactionRequest(
		http.MethodGet,
		"/api/v1/reports/finance-summary?business_date=2026-04-30",
		nil,
		auth.RoleWarehouseLead,
	)
	rec := httptest.NewRecorder()

	newTestFinanceSummaryReportHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestFinanceSummaryReportHandlerRejectsInvalidFilters(t *testing.T) {
	req := cashTransactionRequest(
		http.MethodGet,
		"/api/v1/reports/finance-summary?from_date=2026-05-01&to_date=2026-04-30",
		nil,
		auth.RoleFinanceOps,
	)
	rec := httptest.NewRecorder()

	newTestFinanceSummaryReportHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestFinanceSummaryReportHandlerRejectsUnsupportedMethod(t *testing.T) {
	req := cashTransactionRequest(http.MethodPost, "/api/v1/reports/finance-summary", nil, auth.RoleFinanceOps)
	rec := httptest.NewRecorder()

	newTestFinanceSummaryReportHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func newTestFinanceSummaryReportHandler() http.HandlerFunc {
	return financeSummaryReportHandler(
		financeapp.NewPrototypeCustomerReceivableStore(),
		financeapp.NewPrototypeSupplierPayableStore(),
		financeapp.NewPrototypeCODRemittanceStore(),
		financeapp.NewPrototypeCashTransactionStore(),
	)
}
