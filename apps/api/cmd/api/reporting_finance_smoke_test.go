package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestFinanceSummaryReportSmoke(t *testing.T) {
	authConfig := smokeAuthConfig()

	t.Run("JSON report returns finance summary buckets", func(t *testing.T) {
		req := smokeRequestAsRole(
			httptest.NewRequest(
				http.MethodGet,
				"/api/v1/reports/finance-summary?from_date=2026-04-30&to_date=2026-05-08&business_date=2026-05-08",
				nil,
			),
			authConfig,
			auth.RoleFinanceOps,
		)
		req.Header.Set(response.HeaderRequestID, "req-smoke-finance-report")
		rec := httptest.NewRecorder()

		newTestFinanceSummaryReportHandler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
		payload := decodeSmokeSuccess[financeSummaryReportResponse](t, rec)
		if payload.Data.Metadata.Filters.FromDate != "2026-04-30" ||
			payload.Data.Metadata.Filters.ToDate != "2026-05-08" ||
			payload.Data.Metadata.Filters.BusinessDate != "2026-05-08" ||
			payload.Data.CurrencyCode != "VND" {
			t.Fatalf("metadata = %+v currency = %q", payload.Data.Metadata, payload.Data.CurrencyCode)
		}
		if payload.Data.AR.OpenCount != 1 || payload.Data.AR.OverdueAmount != "1250000.00" {
			t.Fatalf("ar = %+v, want one overdue receivable", payload.Data.AR)
		}
		if payload.Data.AP.DueCount != 1 || payload.Data.AP.OutstandingAmount != "4250000.00" {
			t.Fatalf("ap = %+v, want due payable", payload.Data.AP)
		}
		if payload.Data.COD.PendingCount != 1 ||
			payload.Data.COD.DiscrepancyCount != 1 ||
			payload.Data.COD.DiscrepancyAmount != "-50000.00" {
			t.Fatalf("cod = %+v, want pending COD discrepancy", payload.Data.COD)
		}
		if len(payload.Data.COD.DiscrepancyBuckets) != 1 ||
			payload.Data.COD.DiscrepancyBuckets[0].SourceReference.EntityType != "cod_discrepancy" ||
			!strings.Contains(payload.Data.COD.DiscrepancyBuckets[0].SourceReference.ID, "cod-remit-260430-0001-line-1") {
			t.Fatalf("cod discrepancy buckets = %+v, want auditable source reference", payload.Data.COD.DiscrepancyBuckets)
		}
		if payload.Data.Cash.TransactionCount != 2 || payload.Data.Cash.NetCashAmount != "-3000000.00" {
			t.Fatalf("cash = %+v, want net cash summary", payload.Data.Cash)
		}
	})

	t.Run("CSV export returns stable finance summary file", func(t *testing.T) {
		req := smokeRequestAsRole(
			httptest.NewRequest(
				http.MethodGet,
				"/api/v1/reports/finance-summary/export.csv?from_date=2026-04-30&to_date=2026-05-08&business_date=2026-05-08",
				nil,
			),
			authConfig,
			auth.RoleFinanceOps,
		)
		req.Header.Set(response.HeaderRequestID, "req-smoke-finance-report-csv")
		rec := httptest.NewRecorder()

		newTestFinanceSummaryCSVExportHandler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
		if got := rec.Header().Get("Content-Type"); got != "text/csv; charset=utf-8" {
			t.Fatalf("content type = %q, want text/csv", got)
		}
		if got := rec.Header().Get("Content-Disposition"); got != `attachment; filename="finance-summary-2026-04-30-to-2026-05-08.csv"` {
			t.Fatalf("content disposition = %q", got)
		}

		lines := strings.Split(strings.TrimSpace(rec.Body.String()), "\n")
		const header = "section,metric,bucket,type,status,count,amount,currency_code"
		if len(lines) < 10 || lines[0] != header {
			t.Fatalf("csv lines = %q, want stable header and finance rows", lines)
		}
		body := rec.Body.String()
		for _, want := range []string{
			"ar,open,,,,1,1250000.00,VND",
			"ap,due,,,,1,4250000.00,VND",
			"cod,discrepancy,,,,1,-50000.00,VND",
			"cash,net_cash,,,,2,-3000000.00,VND",
		} {
			if !strings.Contains(body, want) {
				t.Fatalf("csv body = %q, want %q", body, want)
			}
		}
	})
}
