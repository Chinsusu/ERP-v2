package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestInventorySnapshotReportHandlerRejectsUnsupportedMethod(t *testing.T) {
	service := inventoryapp.NewListAvailableStock(inventoryapp.NewPrototypeStockAvailabilityStore())
	req := cashTransactionRequest(http.MethodPost, "/api/v1/reports/inventory-snapshot", nil, auth.RoleWarehouseLead)
	rec := httptest.NewRecorder()

	inventorySnapshotReportHandler(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}
