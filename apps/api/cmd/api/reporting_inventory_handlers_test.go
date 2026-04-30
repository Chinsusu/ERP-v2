package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestInventorySnapshotReportHandlerReturnsFilteredReport(t *testing.T) {
	service := inventoryapp.NewListAvailableStock(inventoryapp.NewPrototypeStockAvailabilityStore())
	req := cashTransactionRequest(
		http.MethodGet,
		"/api/v1/reports/inventory-snapshot?business_date=2026-04-30&warehouse_id=wh-hcm&item_id=item-serum-30ml&status=quarantine&low_stock_threshold=10&expiry_warning_days=45",
		nil,
		auth.RoleWarehouseLead,
	)
	req.Header.Set(response.HeaderRequestID, "req-report-inventory")
	rec := httptest.NewRecorder()

	inventorySnapshotReportHandler(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var payload response.SuccessEnvelope[inventorySnapshotReportResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.RequestID != "req-report-inventory" {
		t.Fatalf("request id = %q, want req-report-inventory", payload.RequestID)
	}
	if payload.Data.Metadata.Timezone != "Asia/Ho_Chi_Minh" ||
		payload.Data.Metadata.SourceVersion != "reporting-v1" ||
		payload.Data.Metadata.Filters.BusinessDate != "2026-04-30" ||
		payload.Data.Metadata.Filters.WarehouseID != "wh-hcm" ||
		payload.Data.Metadata.Filters.ItemID != "item-serum-30ml" {
		t.Fatalf("metadata = %+v", payload.Data.Metadata)
	}
	if payload.Data.Summary.RowCount != 1 || len(payload.Data.Rows) != 1 {
		t.Fatalf("row count = %d rows = %d, want one filtered row", payload.Data.Summary.RowCount, len(payload.Data.Rows))
	}

	row := payload.Data.Rows[0]
	if row.ItemID != "item-serum-30ml" ||
		row.SKU != "SERUM-30ML" ||
		row.SourceStockState != "quarantine" ||
		row.AvailableQty != "110.000000" ||
		row.QuarantineQty != "8.000000" {
		t.Fatalf("row = %+v, want serum quarantine snapshot with decimal quantity strings", row)
	}
	if len(payload.Data.Summary.TotalsByUOM) != 1 ||
		payload.Data.Summary.TotalsByUOM[0].AvailableQty != "110.000000" {
		t.Fatalf("totals = %+v, want filtered PCS total", payload.Data.Summary.TotalsByUOM)
	}
	if len(row.SourceReferences) < 4 {
		t.Fatalf("source references = %+v, want warehouse/item/batch/stock refs", row.SourceReferences)
	}
	if row.SourceReferences[0].EntityType != "warehouse" ||
		row.SourceReferences[0].ID != "wh-hcm" ||
		row.SourceReferences[0].Href != "/master-data?source_id=wh-hcm&source_type=warehouse" {
		t.Fatalf("warehouse source reference = %+v", row.SourceReferences[0])
	}
}

func TestInventorySnapshotReportHandlerFiltersStatusByQuantityBucket(t *testing.T) {
	service := inventoryapp.NewListAvailableStock(inventoryapp.NewPrototypeStockAvailabilityStore())
	tests := []struct {
		name         string
		status       string
		itemID       string
		wantSKU      string
		wantStateRef string
	}{
		{
			name:         "available stock context",
			status:       "available",
			itemID:       "item-serum-30ml",
			wantSKU:      "SERUM-30ML",
			wantStateRef: "wh-hcm:bin-hcm-a01:SERUM-30ML:batch-serum-2604a:available",
		},
		{
			name:         "reserved stock context",
			status:       "reserved",
			itemID:       "item-serum-30ml",
			wantSKU:      "SERUM-30ML",
			wantStateRef: "wh-hcm:bin-hcm-a01:SERUM-30ML:batch-serum-2604a:reserved",
		},
		{
			name:         "quarantine stock context",
			status:       "quarantine",
			itemID:       "item-serum-30ml",
			wantSKU:      "SERUM-30ML",
			wantStateRef: "wh-hcm:bin-hcm-a01:SERUM-30ML:batch-serum-2604a:quarantine",
		},
		{
			name:         "blocked stock context",
			status:       "blocked",
			itemID:       "item-cream-50g",
			wantSKU:      "CREAM-50G",
			wantStateRef: "wh-hcm:bin-hcm-a01:CREAM-50G:batch-cream-2603b:blocked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := cashTransactionRequest(
				http.MethodGet,
				"/api/v1/reports/inventory-snapshot?business_date=2026-04-30&warehouse_id=wh-hcm&item_id="+tt.itemID+"&status="+tt.status,
				nil,
				auth.RoleWarehouseLead,
			)
			rec := httptest.NewRecorder()

			inventorySnapshotReportHandler(service).ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
			}
			var payload response.SuccessEnvelope[inventorySnapshotReportResponse]
			if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if payload.Data.Summary.RowCount != 1 || len(payload.Data.Rows) != 1 {
				t.Fatalf("row count = %d rows = %d, want one row", payload.Data.Summary.RowCount, len(payload.Data.Rows))
			}
			row := payload.Data.Rows[0]
			if row.SKU != tt.wantSKU {
				t.Fatalf("row = %+v, want sku %s", row, tt.wantSKU)
			}
			if !hasInventorySnapshotSourceReference(row.SourceReferences, "stock_state", tt.wantStateRef) {
				t.Fatalf("source references = %+v, missing stock state %s", row.SourceReferences, tt.wantStateRef)
			}
		})
	}
}

func hasInventorySnapshotSourceReference(
	references []reportSourceReferenceResponse,
	entityType string,
	id string,
) bool {
	for _, reference := range references {
		if reference.EntityType == entityType && reference.ID == id && reference.Href != "" && !reference.Unavailable {
			return true
		}
	}

	return false
}

func TestInventorySnapshotCSVExportHandlerReturnsCSV(t *testing.T) {
	service := inventoryapp.NewListAvailableStock(inventoryapp.NewPrototypeStockAvailabilityStore())
	req := cashTransactionRequest(
		http.MethodGet,
		"/api/v1/reports/inventory-snapshot/export.csv?business_date=2026-04-30&warehouse_id=wh-hcm&item_id=item-serum-30ml&status=quarantine&low_stock_threshold=10&expiry_warning_days=45",
		nil,
		auth.RoleWarehouseLead,
	)
	req.Header.Set(response.HeaderRequestID, "req-report-inventory-csv")
	rec := httptest.NewRecorder()

	inventorySnapshotCSVExportHandler(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "text/csv; charset=utf-8" {
		t.Fatalf("content type = %q, want text/csv", got)
	}
	if got := rec.Header().Get("Content-Disposition"); got != `attachment; filename="inventory-snapshot-2026-04-30.csv"` {
		t.Fatalf("content disposition = %q", got)
	}
	if got := rec.Header().Get(response.HeaderRequestID); got != "req-report-inventory-csv" {
		t.Fatalf("request id = %q, want req-report-inventory-csv", got)
	}

	lines := strings.Split(strings.TrimSpace(rec.Body.String()), "\n")
	const header = "warehouse_id,warehouse_code,location_id,location_code,item_id,sku,batch_id,batch_no,batch_expiry,base_uom_code,physical_qty,reserved_qty,quarantine_qty,blocked_qty,available_qty,low_stock,expiry_warning,expired,batch_qc_status,batch_status,source_stock_state"
	if len(lines) != 2 || lines[0] != header {
		t.Fatalf("csv lines = %q, want header and one row", lines)
	}
	row := lines[1]
	for _, want := range []string{
		"wh-hcm",
		"SERUM-30ML",
		"LOT-2604A",
		"110.000000",
		"false,false,false",
		"quarantine",
	} {
		if !strings.Contains(row, want) {
			t.Fatalf("csv row = %q, want %q", row, want)
		}
	}
}

func TestInventorySnapshotReportHandlerRequiresReportsPermission(t *testing.T) {
	service := inventoryapp.NewListAvailableStock(inventoryapp.NewPrototypeStockAvailabilityStore())
	req := cashTransactionRequest(http.MethodGet, "/api/v1/reports/inventory-snapshot", nil, auth.RoleWarehouseStaff)
	rec := httptest.NewRecorder()

	inventorySnapshotReportHandler(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestInventorySnapshotCSVExportHandlerRequiresExportPermission(t *testing.T) {
	service := inventoryapp.NewListAvailableStock(inventoryapp.NewPrototypeStockAvailabilityStore())
	req := cashTransactionRequest(http.MethodGet, "/api/v1/reports/inventory-snapshot/export.csv", nil, auth.RoleQA)
	rec := httptest.NewRecorder()

	inventorySnapshotCSVExportHandler(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestInventorySnapshotReportHandlerRejectsInvalidFilters(t *testing.T) {
	service := inventoryapp.NewListAvailableStock(inventoryapp.NewPrototypeStockAvailabilityStore())

	tests := []string{
		"/api/v1/reports/inventory-snapshot?from_date=2026-05-01&to_date=2026-04-30",
		"/api/v1/reports/inventory-snapshot?status=missing",
		"/api/v1/reports/inventory-snapshot?expiry_warning_days=0",
	}

	for _, target := range tests {
		req := cashTransactionRequest(http.MethodGet, target, nil, auth.RoleWarehouseLead)
		rec := httptest.NewRecorder()

		inventorySnapshotReportHandler(service).ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("%s status = %d, want %d", target, rec.Code, http.StatusBadRequest)
		}
	}
}

func TestInventorySnapshotCSVExportHandlerRejectsInvalidFilters(t *testing.T) {
	service := inventoryapp.NewListAvailableStock(inventoryapp.NewPrototypeStockAvailabilityStore())
	req := cashTransactionRequest(
		http.MethodGet,
		"/api/v1/reports/inventory-snapshot/export.csv?business_date=not-a-date",
		nil,
		auth.RoleWarehouseLead,
	)
	rec := httptest.NewRecorder()

	inventorySnapshotCSVExportHandler(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestInventorySnapshotReportHandlerRejectsUnsupportedMethod(t *testing.T) {
	service := inventoryapp.NewListAvailableStock(inventoryapp.NewPrototypeStockAvailabilityStore())
	req := cashTransactionRequest(http.MethodPost, "/api/v1/reports/inventory-snapshot", nil, auth.RoleWarehouseLead)
	rec := httptest.NewRecorder()

	inventorySnapshotReportHandler(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestInventorySnapshotCSVExportHandlerRejectsUnsupportedMethod(t *testing.T) {
	service := inventoryapp.NewListAvailableStock(inventoryapp.NewPrototypeStockAvailabilityStore())
	req := cashTransactionRequest(http.MethodPost, "/api/v1/reports/inventory-snapshot/export.csv", nil, auth.RoleWarehouseLead)
	rec := httptest.NewRecorder()

	inventorySnapshotCSVExportHandler(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}
