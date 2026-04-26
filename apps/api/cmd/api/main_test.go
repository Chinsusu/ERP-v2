package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	shippingapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/application"
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

func TestEndOfDayReconciliationsHandlerReturnsFilteredRows(t *testing.T) {
	store := inventoryapp.NewPrototypeEndOfDayReconciliationStore()
	service := inventoryapp.NewListEndOfDayReconciliations(store)
	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/warehouse/end-of-day-reconciliations?warehouse_id=wh-hcm&date=2026-04-26&status=in_review",
		nil,
	)
	req.Header.Set(response.HeaderRequestID, "req-reconciliation")
	rec := httptest.NewRecorder()

	endOfDayReconciliationsHandler(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var payload response.SuccessEnvelope[[]endOfDayReconciliationResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Data) != 1 {
		t.Fatalf("rows = %d, want 1", len(payload.Data))
	}
	if payload.Data[0].Summary.VarianceQuantity != -2 {
		t.Fatalf("variance = %d, want -2", payload.Data[0].Summary.VarianceQuantity)
	}
}

func TestCloseEndOfDayReconciliationHandlerWritesAudit(t *testing.T) {
	store := inventoryapp.NewPrototypeEndOfDayReconciliationStore()
	auditStore := audit.NewInMemoryLogStore()
	service := inventoryapp.NewCloseEndOfDayReconciliation(store, auditStore)
	body := bytes.NewBufferString(`{"exception_note":"variance accepted by lead"}`)
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/warehouse/end-of-day-reconciliations/rec-hcm-260426-day/close",
		body,
	)
	req.SetPathValue("reconciliation_id", "rec-hcm-260426-day")
	req.Header.Set(response.HeaderRequestID, "req-close")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead)))
	rec := httptest.NewRecorder()

	closeEndOfDayReconciliationHandler(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var payload response.SuccessEnvelope[endOfDayReconciliationResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.Status != "closed" {
		t.Fatalf("status = %q, want closed", payload.Data.Status)
	}
	if payload.Data.AuditLogID == "" {
		t.Fatal("audit log id is empty")
	}

	logs, err := auditStore.List(req.Context(), audit.Query{Action: "warehouse.shift.closed"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit logs = %d, want 1", len(logs))
	}
}

func TestCarrierManifestsHandlerReturnsFilteredRows(t *testing.T) {
	store := shippingapp.NewPrototypeCarrierManifestStore()
	service := shippingapp.NewListCarrierManifests(store)
	createService := shippingapp.NewCreateCarrierManifest(store, audit.NewInMemoryLogStore())
	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/shipping/manifests?warehouse_id=wh-hcm&date=2026-04-26&carrier_code=GHN&status=scanning",
		nil,
	)
	req.Header.Set(response.HeaderRequestID, "req-manifest-list")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead)))
	rec := httptest.NewRecorder()

	carrierManifestsHandler(service, createService).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var payload response.SuccessEnvelope[[]carrierManifestResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Data) != 1 {
		t.Fatalf("rows = %d, want 1", len(payload.Data))
	}
	if payload.Data[0].Summary.ExpectedCount != 3 || payload.Data[0].Summary.MissingCount != 1 {
		t.Fatalf("summary = %+v, want expected 3 missing 1", payload.Data[0].Summary)
	}
}

func TestCarrierManifestsHandlerCreatesManifestAndWritesAudit(t *testing.T) {
	store := shippingapp.NewPrototypeCarrierManifestStore()
	auditStore := audit.NewInMemoryLogStore()
	listService := shippingapp.NewListCarrierManifests(store)
	createService := shippingapp.NewCreateCarrierManifest(store, auditStore)
	body := bytes.NewBufferString(`{
		"carrier_code": "NJV",
		"carrier_name": "Ninja Van",
		"warehouse_id": "wh-hcm",
		"warehouse_code": "HCM",
		"date": "2026-04-26"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/shipping/manifests", body)
	req.Header.Set(response.HeaderRequestID, "req-manifest-create")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead)))
	rec := httptest.NewRecorder()

	carrierManifestsHandler(listService, createService).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var payload response.SuccessEnvelope[carrierManifestResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.Status != "draft" || payload.Data.AuditLogID == "" {
		t.Fatalf("manifest response = %+v, want draft with audit log", payload.Data)
	}

	logs, err := auditStore.List(req.Context(), audit.Query{Action: "shipping.manifest.created"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit logs = %d, want 1", len(logs))
	}
}

func TestAddShipmentToCarrierManifestHandlerUpdatesCounts(t *testing.T) {
	store := shippingapp.NewPrototypeCarrierManifestStore()
	auditStore := audit.NewInMemoryLogStore()
	service := shippingapp.NewAddShipmentToCarrierManifest(store, auditStore)
	body := bytes.NewBufferString(`{"shipment_id":"ship-hcm-260426-004"}`)
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/shipping/manifests/manifest-hcm-ghn-morning/shipments",
		body,
	)
	req.SetPathValue("manifest_id", "manifest-hcm-ghn-morning")
	req.Header.Set(response.HeaderRequestID, "req-manifest-add")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead)))
	rec := httptest.NewRecorder()

	addShipmentToCarrierManifestHandler(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var payload response.SuccessEnvelope[carrierManifestResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.Summary.ExpectedCount != 4 || payload.Data.Summary.MissingCount != 2 {
		t.Fatalf("summary = %+v, want expected 4 missing 2", payload.Data.Summary)
	}
}

func TestVerifyCarrierManifestScanHandlerMarksLineAndWritesAudit(t *testing.T) {
	store := shippingapp.NewPrototypeCarrierManifestStore()
	auditStore := audit.NewInMemoryLogStore()
	service := shippingapp.NewVerifyCarrierManifestScan(store, auditStore)
	body := bytes.NewBufferString(`{"code":"GHN260426003","station_id":"dock-a"}`)
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/shipping/manifests/manifest-hcm-ghn-morning/scan",
		body,
	)
	req.SetPathValue("manifest_id", "manifest-hcm-ghn-morning")
	req.Header.Set(response.HeaderRequestID, "req-manifest-scan")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseStaff)))
	rec := httptest.NewRecorder()

	verifyCarrierManifestScanHandler(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var payload response.SuccessEnvelope[carrierManifestScanResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.ResultCode != "MATCHED" || payload.Data.Manifest.Summary.MissingCount != 0 {
		t.Fatalf("scan response = %+v, want MATCHED and 0 missing", payload.Data)
	}
	if payload.Data.ScanEvent.ID == "" || payload.Data.AuditLogID == "" {
		t.Fatalf("scan event/audit response = %+v, want ids", payload.Data)
	}

	logs, err := auditStore.List(req.Context(), audit.Query{Action: "shipping.manifest.scan_recorded"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit logs = %d, want 1", len(logs))
	}
}

func TestVerifyCarrierManifestScanHandlerReturnsWarningCodes(t *testing.T) {
	cases := []struct {
		name string
		code string
		want string
	}{
		{name: "duplicate", code: "GHN260426001", want: "DUPLICATE_SCAN"},
		{name: "wrong manifest", code: "VTP260426011", want: "MANIFEST_MISMATCH"},
		{name: "unpacked", code: "GHN260426099", want: "INVALID_STATE"},
		{name: "unknown", code: "UNKNOWN-CODE", want: "NOT_FOUND"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := shippingapp.NewPrototypeCarrierManifestStore()
			service := shippingapp.NewVerifyCarrierManifestScan(store, audit.NewInMemoryLogStore())
			body := bytes.NewBufferString(`{"code":"` + tc.code + `","station_id":"dock-a"}`)
			req := httptest.NewRequest(
				http.MethodPost,
				"/api/v1/shipping/manifests/manifest-hcm-ghn-morning/scan",
				body,
			)
			req.SetPathValue("manifest_id", "manifest-hcm-ghn-morning")
			req.Header.Set(response.HeaderRequestID, "req-manifest-scan-warning")
			req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
				Email:       "admin@example.local",
				Password:    "local-only-mock-password",
				AccessToken: "local-dev-access-token",
			}, auth.RoleWarehouseStaff)))
			rec := httptest.NewRecorder()

			verifyCarrierManifestScanHandler(service).ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
			}

			var payload response.SuccessEnvelope[carrierManifestScanResponse]
			if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if payload.Data.ResultCode != tc.want {
				t.Fatalf("result code = %q, want %q", payload.Data.ResultCode, tc.want)
			}
			if payload.Data.ScanEvent.ID == "" {
				t.Fatal("scan event id is empty")
			}
		})
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
