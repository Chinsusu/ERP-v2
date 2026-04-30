package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestInventorySnapshotReportSmoke(t *testing.T) {
	authConfig := smokeAuthConfig()
	service := inventoryapp.NewListAvailableStock(inventoryapp.NewPrototypeStockAvailabilityStore())

	t.Run("JSON report returns inventory snapshot rows", func(t *testing.T) {
		req := smokeRequestAsRole(
			httptest.NewRequest(
				http.MethodGet,
				"/api/v1/reports/inventory-snapshot?business_date=2026-04-30&warehouse_id=wh-hcm&item_id=item-serum-30ml&status=quarantine&expiry_warning_days=45",
				nil,
			),
			authConfig,
			auth.RoleWarehouseLead,
		)
		req.Header.Set(response.HeaderRequestID, "req-smoke-inventory-report")
		rec := httptest.NewRecorder()

		inventorySnapshotReportHandler(service).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
		payload := decodeSmokeSuccess[inventorySnapshotReportResponse](t, rec)
		if payload.Data.Metadata.Filters.BusinessDate != "2026-04-30" ||
			payload.Data.Metadata.Filters.WarehouseID != "wh-hcm" ||
			payload.Data.Metadata.Filters.ItemID != "item-serum-30ml" ||
			payload.Data.Metadata.Filters.Status != "quarantine" {
			t.Fatalf("filters = %+v, want HCM quarantine snapshot", payload.Data.Metadata.Filters)
		}
		if payload.Data.Summary.RowCount == 0 || payload.Data.Summary.TotalsByUOM[0].AvailableQty == "" {
			t.Fatalf("summary = %+v, want inventory totals", payload.Data.Summary)
		}
		if len(payload.Data.Rows) != 1 || payload.Data.Rows[0].SKU != "SERUM-30ML" {
			t.Fatalf("rows = %+v, want SERUM-30ML quarantine row", payload.Data.Rows)
		}
		if payload.Data.Rows[0].AvailableQty != "110.000000" || payload.Data.Rows[0].SourceStockState != "quarantine" {
			t.Fatalf("row = %+v, want decimal available quantity and quarantine state", payload.Data.Rows[0])
		}
		if !hasInventorySnapshotSourceReference(
			payload.Data.Rows[0].SourceReferences,
			"stock_state",
			"wh-hcm:bin-hcm-a01:SERUM-30ML:batch-serum-2604a:quarantine",
		) {
			t.Fatalf("source references = %+v, missing quarantine stock source", payload.Data.Rows[0].SourceReferences)
		}
		if !hasInventorySnapshotSourceReference(
			payload.Data.Rows[0].SourceReferences,
			"inventory_batch",
			"batch-serum-2604a",
		) {
			t.Fatalf("source references = %+v, missing inventory batch source", payload.Data.Rows[0].SourceReferences)
		}
	})

	t.Run("CSV export returns stable inventory snapshot file", func(t *testing.T) {
		req := smokeRequestAsRole(
			httptest.NewRequest(
				http.MethodGet,
				"/api/v1/reports/inventory-snapshot/export.csv?business_date=2026-04-30&warehouse_id=wh-hcm&item_id=item-serum-30ml&status=quarantine&expiry_warning_days=45",
				nil,
			),
			authConfig,
			auth.RoleWarehouseLead,
		)
		req.Header.Set(response.HeaderRequestID, "req-smoke-inventory-report-csv")
		rec := httptest.NewRecorder()

		inventorySnapshotCSVExportHandler(service).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
		if got := rec.Header().Get("Content-Type"); got != "text/csv; charset=utf-8" {
			t.Fatalf("content type = %q, want text/csv", got)
		}
		if got := rec.Header().Get("Content-Disposition"); got != `attachment; filename="inventory-snapshot-2026-04-30.csv"` {
			t.Fatalf("content disposition = %q", got)
		}

		lines := strings.Split(strings.TrimSpace(rec.Body.String()), "\n")
		const header = "warehouse_id,warehouse_code,location_id,location_code,item_id,sku,batch_id,batch_no,batch_expiry,base_uom_code,physical_qty,reserved_qty,quarantine_qty,blocked_qty,available_qty,low_stock,expiry_warning,expired,batch_qc_status,batch_status,source_stock_state"
		if len(lines) != 2 || lines[0] != header {
			t.Fatalf("csv lines = %q, want stable header and one row", lines)
		}
		if !strings.Contains(lines[0], "source_stock_state") {
			t.Fatalf("csv header = %q, want source_stock_state source column", lines[0])
		}
		for _, want := range []string{"SERUM-30ML", "128.000000", "110.000000", "quarantine"} {
			if !strings.Contains(lines[1], want) {
				t.Fatalf("csv row = %q, want %q", lines[1], want)
			}
		}
	})
}
