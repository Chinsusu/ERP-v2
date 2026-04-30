package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestOperationsDailyReportSmoke(t *testing.T) {
	authConfig := smokeAuthConfig()

	t.Run("JSON report returns blocked operations with exceptions", func(t *testing.T) {
		req := smokeRequestAsRole(
			httptest.NewRequest(
				http.MethodGet,
				"/api/v1/reports/operations-daily?business_date=2026-04-30&warehouse_id=wh-hcm&status=blocked",
				nil,
			),
			authConfig,
			auth.RoleWarehouseLead,
		)
		req.Header.Set(response.HeaderRequestID, "req-smoke-operations-report")
		rec := httptest.NewRecorder()

		operationsDailyReportHandler(prototypeOperationsDailySignalSource{}).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
		payload := decodeSmokeSuccess[operationsDailyReportResponse](t, rec)
		if payload.Data.Metadata.Filters.BusinessDate != "2026-04-30" ||
			payload.Data.Metadata.Filters.WarehouseID != "wh-hcm" ||
			payload.Data.Metadata.Filters.Status != "blocked" {
			t.Fatalf("filters = %+v, want HCM blocked operations", payload.Data.Metadata.Filters)
		}
		if payload.Data.Summary.SignalCount != 2 || payload.Data.Summary.BlockedCount != 2 {
			t.Fatalf("summary = %+v, want two blocked signals", payload.Data.Summary)
		}
		if len(payload.Data.Rows) != 2 {
			t.Fatalf("rows length = %d, want 2", len(payload.Data.Rows))
		}
		if payload.Data.Rows[0].ExceptionCode != "MISSING_HANDOVER_SCAN" ||
			payload.Data.Rows[1].ExceptionCode != "VARIANCE_REVIEW" {
			t.Fatalf("rows = %+v, want handover and cycle-count exceptions", payload.Data.Rows)
		}
		for _, row := range payload.Data.Rows {
			if row.SourceReference.EntityType != row.SourceType ||
				row.SourceReference.ID != row.SourceID ||
				row.SourceReference.Href == "" ||
				row.SourceReference.Unavailable {
				t.Fatalf("row source reference = %+v, want linked %s/%s", row.SourceReference, row.SourceType, row.SourceID)
			}
		}
	})

	t.Run("CSV export returns stable operations daily file", func(t *testing.T) {
		req := smokeRequestAsRole(
			httptest.NewRequest(
				http.MethodGet,
				"/api/v1/reports/operations-daily/export.csv?business_date=2026-04-30&warehouse_id=wh-hcm&status=pending",
				nil,
			),
			authConfig,
			auth.RoleWarehouseLead,
		)
		req.Header.Set(response.HeaderRequestID, "req-smoke-operations-report-csv")
		rec := httptest.NewRecorder()

		operationsDailyCSVExportHandler(prototypeOperationsDailySignalSource{}).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
		if got := rec.Header().Get("Content-Type"); got != "text/csv; charset=utf-8" {
			t.Fatalf("content type = %q, want text/csv", got)
		}
		if got := rec.Header().Get("Content-Disposition"); got != `attachment; filename="operations-daily-2026-04-30-to-2026-04-30.csv"` {
			t.Fatalf("content disposition = %q", got)
		}

		lines := strings.Split(strings.TrimSpace(rec.Body.String()), "\n")
		const header = "id,area,source_type,source_id,ref_no,title,warehouse_id,warehouse_code,business_date,status,severity,quantity,uom_code,exception_code,owner"
		if len(lines) != 3 || lines[0] != header {
			t.Fatalf("csv lines = %q, want stable header and two pending rows", lines)
		}
		if !strings.Contains(lines[0], "source_type,source_id") {
			t.Fatalf("csv header = %q, want source_type and source_id source columns", lines[0])
		}
		body := rec.Body.String()
		for _, want := range []string{"GR-260430-0001", "RET-260430-0001", "12.000000", "3.000000", "pending"} {
			if !strings.Contains(body, want) {
				t.Fatalf("csv body = %q, want %q", body, want)
			}
		}
	})

	t.Run("report endpoint remains read only", func(t *testing.T) {
		req := smokeRequestAsRole(
			httptest.NewRequest(
				http.MethodPost,
				"/api/v1/reports/operations-daily?business_date=2026-04-30&warehouse_id=wh-hcm&status=blocked",
				nil,
			),
			authConfig,
			auth.RoleWarehouseLead,
		)
		rec := httptest.NewRecorder()

		operationsDailyReportHandler(prototypeOperationsDailySignalSource{}).ServeHTTP(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
		}
	})
}
