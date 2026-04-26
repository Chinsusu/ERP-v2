package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestAvailableStockHandlerReturnsFilteredRows(t *testing.T) {
	service := inventoryapp.NewListAvailableStock(inventoryapp.NewPrototypeStockAvailabilityStore())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/inventory/available-stock?warehouse_id=wh-hn&sku=toner-100ml", nil)
	req.Header.Set(response.HeaderRequestID, "req-stock")
	rec := httptest.NewRecorder()

	availableStockHandler(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var payload response.SuccessEnvelope[[]availableStockResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Data) != 1 {
		t.Fatalf("rows = %d, want 1", len(payload.Data))
	}

	got := payload.Data[0]
	if got.WarehouseID != "wh-hn" || got.SKU != "TONER-100ML" || got.AvailableStock != 65 {
		t.Fatalf("available stock row = %+v, want HN TONER-100ML available 65", got)
	}
}

func TestAvailableStockHandlerRejectsUnsupportedMethod(t *testing.T) {
	service := inventoryapp.NewListAvailableStock(inventoryapp.NewPrototypeStockAvailabilityStore())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/inventory/available-stock", nil)
	rec := httptest.NewRecorder()

	availableStockHandler(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}
