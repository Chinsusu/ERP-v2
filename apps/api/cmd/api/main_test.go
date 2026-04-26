package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
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

func TestAuditLogsHandlerReturnsFilteredRows(t *testing.T) {
	log, err := audit.NewLog(audit.NewLogInput{
		ID:         "audit-test",
		ActorID:    "user-erp-admin",
		Action:     "inventory.stock_movement.adjusted",
		EntityType: "inventory.stock_movement",
		EntityID:   "mov-test",
		Metadata:   map[string]any{"reason": "cycle count"},
	})
	if err != nil {
		t.Fatalf("new log: %v", err)
	}
	store := audit.NewInMemoryLogStore(log)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit-logs?action=inventory.stock_movement.adjusted", nil)
	req.Header.Set(response.HeaderRequestID, "req-audit")
	rec := httptest.NewRecorder()

	auditLogsHandler(store).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var payload response.SuccessEnvelope[[]auditLogResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Data) != 1 {
		t.Fatalf("rows = %d, want 1", len(payload.Data))
	}
	if payload.Data[0].ActorID != "user-erp-admin" || payload.Data[0].EntityID != "mov-test" {
		t.Fatalf("audit row = %+v, want admin mov-test", payload.Data[0])
	}
}

func TestAuditLogsHandlerRejectsDelete(t *testing.T) {
	store := audit.NewInMemoryLogStore()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/audit-logs", nil)
	rec := httptest.NewRecorder()

	auditLogsHandler(store).ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestStockMovementHandlerWritesAuditForAdjustment(t *testing.T) {
	store := audit.NewInMemoryLogStore()
	body := bytes.NewBufferString(`{
		"movementId": "mov-adjust-test",
		"sku": "serum-30ml",
		"warehouseId": "wh-hcm",
		"movementType": "ADJUST",
		"quantity": 8,
		"reason": "cycle count"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/inventory/stock-movements", body)
	req.Header.Set(response.HeaderRequestID, "req-adjust")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleERPAdmin)))
	rec := httptest.NewRecorder()

	stockMovementHandler(store).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	logs, err := store.List(req.Context(), audit.Query{EntityID: "mov-adjust-test"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit logs = %d, want 1", len(logs))
	}
	got := logs[0]
	if got.ActorID != "user-erp-admin" || got.Action != "inventory.stock_movement.adjusted" {
		t.Fatalf("audit log = %+v, want admin adjustment action", got)
	}
	if got.RequestID != "req-adjust" {
		t.Fatalf("request id = %q, want req-adjust", got.RequestID)
	}
}
