package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestWriteCSVWritesHeadersRowsAndResponseHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/inventory-snapshot/export.csv", nil)
	req.Header.Set(response.HeaderRequestID, "req-report-csv")
	rec := httptest.NewRecorder()

	err := WriteCSV(rec, req, CSVExport{
		Filename: `inventory/"snapshot"`,
		Headers:  []string{"item_id", "available_qty"},
		Rows: [][]string{
			{"sku-001", "12.000000"},
			{"sku-002", "0.500000"},
		},
	})
	if err != nil {
		t.Fatalf("WriteCSV returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "text/csv; charset=utf-8" {
		t.Fatalf("content type = %q, want text/csv", got)
	}
	if got := rec.Header().Get("Content-Disposition"); got != `attachment; filename="inventory__snapshot_.csv"` {
		t.Fatalf("content disposition = %q", got)
	}
	if got := rec.Header().Get(response.HeaderRequestID); got != "req-report-csv" {
		t.Fatalf("request id = %q, want req-report-csv", got)
	}
	if body := rec.Body.String(); body != "item_id,available_qty\nsku-001,12.000000\nsku-002,0.500000\n" {
		t.Fatalf("body = %q", body)
	}
}

func TestWriteCSVQuotesCells(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/operations-daily/export.csv", nil)
	rec := httptest.NewRecorder()

	err := WriteCSV(rec, req, CSVExport{
		Filename: "operations.csv",
		Headers:  []string{"status", "note"},
		Rows:     [][]string{{"blocked", "needs, review"}},
	})
	if err != nil {
		t.Fatalf("WriteCSV returned error: %v", err)
	}

	if !strings.Contains(rec.Body.String(), `"needs, review"`) {
		t.Fatalf("body = %q, want quoted csv cell", rec.Body.String())
	}
}

func TestWriteCSVRejectsInvalidExport(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/finance-summary/export.csv", nil)
	rec := httptest.NewRecorder()

	err := WriteCSV(rec, req, CSVExport{
		Filename: "finance.csv",
		Headers:  []string{"metric", "amount"},
		Rows:     [][]string{{"ar_open"}},
	})
	if !errors.Is(err, ErrInvalidCSVExport) {
		t.Fatalf("error = %v, want ErrInvalidCSVExport", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want default recorder status before write", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("body = %q, want empty body", rec.Body.String())
	}
}
