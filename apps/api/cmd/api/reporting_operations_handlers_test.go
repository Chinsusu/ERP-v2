package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestOperationsDailyReportHandlerReturnsFilteredReport(t *testing.T) {
	req := cashTransactionRequest(
		http.MethodGet,
		"/api/v1/reports/operations-daily?business_date=2026-04-30&warehouse_id=wh-hcm&status=blocked",
		nil,
		auth.RoleWarehouseLead,
	)
	req.Header.Set(response.HeaderRequestID, "req-report-operations")
	rec := httptest.NewRecorder()

	operationsDailyReportHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var payload response.SuccessEnvelope[operationsDailyReportResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.RequestID != "req-report-operations" {
		t.Fatalf("request id = %q, want req-report-operations", payload.RequestID)
	}
	if payload.Data.Metadata.Timezone != "Asia/Ho_Chi_Minh" ||
		payload.Data.Metadata.SourceVersion != "reporting-v1" ||
		payload.Data.Metadata.Filters.BusinessDate != "2026-04-30" ||
		payload.Data.Metadata.Filters.WarehouseID != "wh-hcm" ||
		payload.Data.Metadata.Filters.Status != "blocked" {
		t.Fatalf("metadata = %+v", payload.Data.Metadata)
	}
	if payload.Data.Summary.SignalCount != 2 || payload.Data.Summary.BlockedCount != 2 {
		t.Fatalf("summary = %+v, want two blocked operations", payload.Data.Summary)
	}
	if len(payload.Data.Areas) != 2 ||
		payload.Data.Areas[0].Area != "outbound" ||
		payload.Data.Areas[1].Area != "stock_count" {
		t.Fatalf("areas = %+v, want outbound and stock_count summaries", payload.Data.Areas)
	}
	if len(payload.Data.Rows) != 2 {
		t.Fatalf("rows length = %d, want 2", len(payload.Data.Rows))
	}
	for _, row := range payload.Data.Rows {
		if row.WarehouseID != "wh-hcm" ||
			row.BusinessDate != "2026-04-30" ||
			row.Status != "blocked" {
			t.Fatalf("row = %+v, want filtered HCM blocked row", row)
		}
	}
	if payload.Data.Rows[0].ExceptionCode != "MISSING_HANDOVER_SCAN" ||
		payload.Data.Rows[1].ExceptionCode != "VARIANCE_REVIEW" {
		t.Fatalf("rows = %+v, want blocked exception codes sorted by area", payload.Data.Rows)
	}
	if payload.Data.Rows[0].SourceReference.EntityType != "carrier_manifest" ||
		payload.Data.Rows[0].SourceReference.ID != "manifest-260430-ghn" ||
		payload.Data.Rows[0].SourceReference.Label != "MAN-260430-GHN" ||
		payload.Data.Rows[0].SourceReference.Href != "/shipping?source_id=manifest-260430-ghn&source_type=carrier_manifest" ||
		payload.Data.Rows[0].SourceReference.Unavailable {
		t.Fatalf("source reference = %+v, want shipping manifest reference", payload.Data.Rows[0].SourceReference)
	}
}

func TestOperationsDailyCSVExportHandlerReturnsCSV(t *testing.T) {
	req := cashTransactionRequest(
		http.MethodGet,
		"/api/v1/reports/operations-daily/export.csv?business_date=2026-04-30&warehouse_id=wh-hcm&status=pending",
		nil,
		auth.RoleWarehouseLead,
	)
	req.Header.Set(response.HeaderRequestID, "req-report-operations-csv")
	rec := httptest.NewRecorder()

	operationsDailyCSVExportHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "text/csv; charset=utf-8" {
		t.Fatalf("content type = %q, want text/csv", got)
	}
	if got := rec.Header().Get("Content-Disposition"); got != `attachment; filename="operations-daily-2026-04-30-to-2026-04-30.csv"` {
		t.Fatalf("content disposition = %q", got)
	}
	if got := rec.Header().Get(response.HeaderRequestID); got != "req-report-operations-csv" {
		t.Fatalf("request id = %q, want req-report-operations-csv", got)
	}

	lines := strings.Split(strings.TrimSpace(rec.Body.String()), "\n")
	const header = "id,area,source_type,source_id,ref_no,title,warehouse_id,warehouse_code,business_date,status,severity,quantity,uom_code,exception_code,owner"
	if len(lines) != 3 || lines[0] != header {
		t.Fatalf("csv lines = %q, want header and two pending rows", lines)
	}
	body := rec.Body.String()
	for _, want := range []string{
		"GR-260430-0001",
		"12.000000",
		"RET-260430-0001",
		"3.000000",
		"pending",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("csv body = %q, want %q", body, want)
		}
	}
}

func TestOperationsDailyReportHandlerRequiresReportsPermission(t *testing.T) {
	req := cashTransactionRequest(http.MethodGet, "/api/v1/reports/operations-daily", nil, auth.RoleWarehouseStaff)
	rec := httptest.NewRecorder()

	operationsDailyReportHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestOperationsDailyCSVExportHandlerRequiresExportPermission(t *testing.T) {
	req := cashTransactionRequest(http.MethodGet, "/api/v1/reports/operations-daily/export.csv", nil, auth.RoleQA)
	rec := httptest.NewRecorder()

	operationsDailyCSVExportHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestOperationsDailyReportHandlerRejectsInvalidFilters(t *testing.T) {
	tests := []string{
		"/api/v1/reports/operations-daily?from_date=2026-05-01&to_date=2026-04-30",
		"/api/v1/reports/operations-daily?status=missing",
	}

	for _, target := range tests {
		req := cashTransactionRequest(http.MethodGet, target, nil, auth.RoleWarehouseLead)
		rec := httptest.NewRecorder()

		operationsDailyReportHandler().ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("%s status = %d, want %d", target, rec.Code, http.StatusBadRequest)
		}
	}
}

func TestOperationsDailyCSVExportHandlerRejectsInvalidFilters(t *testing.T) {
	req := cashTransactionRequest(
		http.MethodGet,
		"/api/v1/reports/operations-daily/export.csv?from_date=2026-05-01&to_date=2026-04-30",
		nil,
		auth.RoleWarehouseLead,
	)
	rec := httptest.NewRecorder()

	operationsDailyCSVExportHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestOperationsDailyReportHandlerRejectsUnsupportedMethod(t *testing.T) {
	req := cashTransactionRequest(http.MethodPost, "/api/v1/reports/operations-daily", nil, auth.RoleWarehouseLead)
	rec := httptest.NewRecorder()

	operationsDailyReportHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestOperationsDailyCSVExportHandlerRejectsUnsupportedMethod(t *testing.T) {
	req := cashTransactionRequest(http.MethodPost, "/api/v1/reports/operations-daily/export.csv", nil, auth.RoleWarehouseLead)
	rec := httptest.NewRecorder()

	operationsDailyCSVExportHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}
