package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	masterdataapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/application"
	returnsapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/returns/application"
	shippingapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/application"
	shippingdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestReadinessHandlerReturnsReady(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	readinessHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var payload response.SuccessEnvelope[healthResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.Status != "ready" {
		t.Fatalf("readiness status = %q, want ready", payload.Data.Status)
	}
}

func TestLoginHandlerIssuesSessionContract(t *testing.T) {
	authConfig := auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}
	sessions := auth.NewSessionManager(authConfig, nil)
	store := audit.NewInMemoryLogStore()
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/login",
		bytes.NewBufferString(`{"email":"admin@example.local","password":"local-only-mock-password"}`),
	)
	rec := httptest.NewRecorder()

	loginHandler(sessions, store).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var payload response.SuccessEnvelope[loginResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.AccessToken == "" || payload.Data.RefreshToken == "" {
		t.Fatalf("tokens are empty: %+v", payload.Data)
	}
	if payload.Data.TokenType != "Bearer" || payload.Data.ExpiresIn <= 0 || payload.Data.RefreshExpiresIn <= 0 {
		t.Fatalf("session contract = %+v, want bearer with positive expiries", payload.Data)
	}

	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+payload.Data.AccessToken)
	meRec := httptest.NewRecorder()
	auth.RequireSessionToken(sessions, http.HandlerFunc(meHandler)).ServeHTTP(meRec, meReq)
	if meRec.Code != http.StatusOK {
		t.Fatalf("me status = %d, want %d: %s", meRec.Code, http.StatusOK, meRec.Body.String())
	}

	logs, err := store.List(req.Context(), audit.Query{Action: "auth.login_succeeded"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 || logs[0].ActorID != "user-erp-admin" {
		t.Fatalf("auth audit logs = %+v, want one admin success event", logs)
	}
}

func TestRefreshHandlerRotatesSessionContract(t *testing.T) {
	sessions := auth.NewSessionManager(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, nil)
	session, failure, ok := sessions.Login("admin@example.local", "local-only-mock-password")
	if !ok {
		t.Fatalf("login rejected: %+v", failure)
	}
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/refresh",
		bytes.NewBufferString(`{"refresh_token":"`+session.RefreshToken+`"}`),
	)
	rec := httptest.NewRecorder()

	refreshHandler(sessions).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var payload response.SuccessEnvelope[loginResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.AccessToken == session.AccessToken || payload.Data.RefreshToken == session.RefreshToken {
		t.Fatalf("tokens were not rotated: old=%+v new=%+v", session, payload.Data)
	}
}

func TestAuthPolicyHandlerDocumentsPasswordAndLockoutPolicy(t *testing.T) {
	sessions := auth.NewSessionManager(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/policy", nil)
	rec := httptest.NewRecorder()

	authPolicyHandler(sessions).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var payload response.SuccessEnvelope[authPolicyResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.PasswordMinLength != 10 || payload.Data.MaxFailedAttempts != 5 {
		t.Fatalf("policy = %+v, want password min 10 and max attempts 5", payload.Data)
	}
}

func TestRbacPermissionsHandlerReturnsCatalog(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rbac/permissions", nil)
	rec := httptest.NewRecorder()

	rbacPermissionsHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var payload response.SuccessEnvelope[[]permissionResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	var hasSubcontract bool
	for _, permission := range payload.Data {
		if permission.Key == string(auth.PermissionSubcontractView) && permission.Group == "operations" {
			hasSubcontract = true
		}
	}
	if !hasSubcontract {
		t.Fatalf("permissions = %+v, want subcontract operations permission", payload.Data)
	}
}

func TestLoginHandlerLocksAfterRepeatedFailures(t *testing.T) {
	sessions := auth.NewSessionManager(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, nil)

	var rec *httptest.ResponseRecorder
	for range 5 {
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/auth/login",
			bytes.NewBufferString(`{"email":"admin@example.local","password":"wrong-password!"}`),
		)
		rec = httptest.NewRecorder()
		loginHandler(sessions).ServeHTTP(rec, req)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	var payload response.ErrorEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Error.Details["reason"] != string(auth.LoginFailureLocked) {
		t.Fatalf("details = %+v, want locked reason", payload.Error.Details)
	}
}

func TestAccessLogMiddlewareWritesRequestSummary(t *testing.T) {
	var logs bytes.Buffer
	logger := log.New(&logs, "", 0)
	handler := accessLogMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	}), logger)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/test", nil)
	req.Header.Set(response.HeaderRequestID, "req-access")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	got := logs.String()
	for _, want := range []string{
		"access method=POST",
		"path=/api/v1/test",
		"status=201",
		"bytes=2",
		"request_id=req-access",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("access log %q missing %q", got, want)
		}
	}
}

func TestAvailableStockHandlerReturnsFilteredRows(t *testing.T) {
	service := inventoryapp.NewListAvailableStock(inventoryapp.NewPrototypeStockAvailabilityStore())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/inventory/available-stock?warehouse_id=wh-hn&location_id=bin-hn-r01&sku=toner-100ml", nil)
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
	if got.WarehouseID != "wh-hn" || got.LocationID != "bin-hn-r01" || got.SKU != "TONER-100ML" || got.AvailableQty != "65.000000" {
		t.Fatalf("available stock row = %+v, want HN TONER-100ML available 65.000000", got)
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

func TestBatchesHandlerReturnsFilteredRows(t *testing.T) {
	catalog := inventoryapp.NewPrototypeBatchCatalog()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/inventory/batches?sku=serum-30ml&qc_status=hold", nil)
	req.Header.Set(response.HeaderRequestID, "req-batch")
	rec := httptest.NewRecorder()

	batchesHandler(catalog).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var payload response.SuccessEnvelope[[]batchResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Data) != 1 {
		t.Fatalf("rows = %d, want 1", len(payload.Data))
	}
	got := payload.Data[0]
	if got.ID != "batch-serum-2604a" || got.QCStatus != "hold" || got.ExpiryDate != "2027-04-01" {
		t.Fatalf("batch row = %+v, want hold serum batch", got)
	}
}

func TestBatchDetailHandlerReturnsNotFound(t *testing.T) {
	catalog := inventoryapp.NewPrototypeBatchCatalog()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/inventory/batches/missing", nil)
	req.SetPathValue("batch_id", "missing")
	rec := httptest.NewRecorder()

	batchDetailHandler(catalog).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestBatchQCTransitionsHandlerRequiresDecisionPermission(t *testing.T) {
	catalog := inventoryapp.NewPrototypeBatchCatalog(audit.NewInMemoryLogStore())
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/inventory/batches/batch-serum-2604a/qc-transitions",
		bytes.NewBufferString(`{"qc_status":"pass","reason":"inspection passed"}`),
	)
	req.SetPathValue("batch_id", "batch-serum-2604a")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead)))
	rec := httptest.NewRecorder()

	batchQCTransitionsHandler(catalog).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestBatchQCTransitionsHandlerChangesStatusAndListsHistory(t *testing.T) {
	auditStore := audit.NewInMemoryLogStore()
	catalog := inventoryapp.NewPrototypeBatchCatalog(auditStore)
	principal := auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleQA)
	body := bytes.NewBufferString(`{"qc_status":"pass","reason":"COA and visual inspection passed","business_ref":"QC-260427-0001"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/inventory/batches/batch-serum-2604a/qc-transitions", body)
	req.SetPathValue("batch_id", "batch-serum-2604a")
	req.Header.Set(response.HeaderRequestID, "req-batch-qc")
	req = req.WithContext(auth.WithPrincipal(req.Context(), principal))
	rec := httptest.NewRecorder()

	batchQCTransitionsHandler(catalog).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var payload response.SuccessEnvelope[batchQCTransitionResultResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.Batch.QCStatus != "pass" || payload.Data.Transition.FromQCStatus != "hold" {
		t.Fatalf("transition payload = %+v, want hold -> pass", payload.Data)
	}

	historyReq := httptest.NewRequest(http.MethodGet, "/api/v1/inventory/batches/batch-serum-2604a/qc-transitions", nil)
	historyReq.SetPathValue("batch_id", "batch-serum-2604a")
	historyReq = historyReq.WithContext(auth.WithPrincipal(historyReq.Context(), principal))
	historyRec := httptest.NewRecorder()
	batchQCTransitionsHandler(catalog).ServeHTTP(historyRec, historyReq)

	if historyRec.Code != http.StatusOK {
		t.Fatalf("history status = %d, want %d: %s", historyRec.Code, http.StatusOK, historyRec.Body.String())
	}
	var historyPayload response.SuccessEnvelope[[]batchQCTransitionResponse]
	if err := json.NewDecoder(historyRec.Body).Decode(&historyPayload); err != nil {
		t.Fatalf("decode history response: %v", err)
	}
	if len(historyPayload.Data) != 1 || historyPayload.Data[0].AuditLogID == "" {
		t.Fatalf("history = %+v, want one audited transition", historyPayload.Data)
	}

	logs, err := auditStore.List(req.Context(), audit.Query{Action: "inventory.batch.qc_status_changed"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 || logs[0].Metadata["business_ref"] != "QC-260427-0001" {
		t.Fatalf("audit logs = %+v, want business reference", logs)
	}
}

func TestGoodsReceiptsHandlerCreatesAndPostsReceipt(t *testing.T) {
	service, auditStore := newTestGoodsReceiptService()
	principal := auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead)
	body := bytes.NewBufferString(`{
		"id": "grn-api-test",
		"receipt_no": "GRN-260427-API",
		"warehouse_id": "wh-hcm-fg",
		"location_id": "loc-hcm-fg-recv-01",
		"reference_doc_type": "purchase_order",
		"reference_doc_id": "PO-260427-API",
		"lines": [
			{
				"id": "line-api-test",
				"batch_id": "batch-cream-2603b",
				"quantity": "6",
				"base_uom_code": "EA"
			}
		]
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/goods-receipts", body)
	req.Header.Set(response.HeaderRequestID, "req-goods-receipt-create")
	req = req.WithContext(auth.WithPrincipal(req.Context(), principal))
	rec := httptest.NewRecorder()

	goodsReceiptsHandler(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var createPayload response.SuccessEnvelope[warehouseReceivingResponse]
	if err := json.NewDecoder(rec.Body).Decode(&createPayload); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if createPayload.Data.Status != "draft" || createPayload.Data.Lines[0].SKU != "CREAM-50G" {
		t.Fatalf("create payload = %+v, want hydrated draft cream receipt", createPayload.Data)
	}

	for _, step := range []struct {
		name    string
		handler http.HandlerFunc
	}{
		{name: "submit", handler: submitGoodsReceiptHandler(service)},
		{name: "inspect", handler: markGoodsReceiptInspectReadyHandler(service)},
		{name: "post", handler: postGoodsReceiptHandler(service)},
	} {
		actionReq := httptest.NewRequest(http.MethodPost, "/api/v1/goods-receipts/grn-api-test/"+step.name, nil)
		actionReq.SetPathValue("receipt_id", "grn-api-test")
		actionReq = actionReq.WithContext(auth.WithPrincipal(actionReq.Context(), principal))
		actionRec := httptest.NewRecorder()

		step.handler.ServeHTTP(actionRec, actionReq)

		if actionRec.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want %d: %s", step.name, actionRec.Code, http.StatusOK, actionRec.Body.String())
		}
		if step.name != "post" {
			continue
		}
		var postPayload response.SuccessEnvelope[warehouseReceivingResponse]
		if err := json.NewDecoder(actionRec.Body).Decode(&postPayload); err != nil {
			t.Fatalf("decode post response: %v", err)
		}
		if postPayload.Data.Status != "posted" || len(postPayload.Data.StockMovements) != 1 {
			t.Fatalf("post payload = %+v, want posted with one stock movement", postPayload.Data)
		}
	}

	logs, err := auditStore.List(context.Background(), audit.Query{Action: "inventory.receiving.posted"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("posted audit logs = %d, want 1", len(logs))
	}
}

func TestPostGoodsReceiptHandlerRequiresRecordCreate(t *testing.T) {
	service, _ := newTestGoodsReceiptService()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/goods-receipts/grn-hcm-260427-inspect/post", nil)
	req.SetPathValue("receipt_id", "grn-hcm-260427-inspect")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseStaff)))
	rec := httptest.NewRecorder()

	postGoodsReceiptHandler(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
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
	if len(payload.Data[0].MissingLines) != 1 || payload.Data[0].MissingLines[0].TrackingNo != "GHN260426003" {
		t.Fatalf("missing lines = %+v, want GHN260426003", payload.Data[0].MissingLines)
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
		"date": "2026-04-26",
		"handover_bin_code": "tote-c01"
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
	if payload.Data.Status != "draft" ||
		payload.Data.CarrierName != "Ninja Van" ||
		payload.Data.StagingZone != "handover-c" ||
		payload.Data.HandoverZoneCode != "HANDOVER-C" ||
		payload.Data.HandoverBinCode != "TOTE-C01" ||
		payload.Data.AuditLogID == "" {
		t.Fatalf("manifest response = %+v, want draft with carrier master defaults and audit log", payload.Data)
	}

	logs, err := auditStore.List(req.Context(), audit.Query{Action: "shipping.manifest.created"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit logs = %d, want 1", len(logs))
	}
}

func TestCarrierManifestsHandlerRejectsInvalidCarrierMaster(t *testing.T) {
	cases := []struct {
		name        string
		carrierCode string
		wantStatus  int
		wantCode    response.ErrorCode
	}{
		{
			name:        "inactive",
			carrierCode: "GHTK",
			wantStatus:  http.StatusConflict,
			wantCode:    response.ErrorCodeConflict,
		},
		{
			name:        "unknown",
			carrierCode: "UNKNOWN",
			wantStatus:  http.StatusNotFound,
			wantCode:    response.ErrorCodeNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := shippingapp.NewPrototypeCarrierManifestStore()
			listService := shippingapp.NewListCarrierManifests(store)
			createService := shippingapp.NewCreateCarrierManifest(store, audit.NewInMemoryLogStore())
			body := bytes.NewBufferString(`{
				"carrier_code": "` + tc.carrierCode + `",
				"warehouse_id": "wh-hcm",
				"warehouse_code": "HCM",
				"date": "2026-04-26"
			}`)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/shipping/manifests", body)
			req.Header.Set(response.HeaderRequestID, "req-manifest-create-invalid-carrier")
			req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
				Email:       "admin@example.local",
				Password:    "local-only-mock-password",
				AccessToken: "local-dev-access-token",
			}, auth.RoleWarehouseLead)))
			rec := httptest.NewRecorder()

			carrierManifestsHandler(listService, createService).ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d: %s", rec.Code, tc.wantStatus, rec.Body.String())
			}
			var payload response.ErrorEnvelope
			if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if payload.Error.Code != tc.wantCode {
				t.Fatalf("code = %s, want %s", payload.Error.Code, tc.wantCode)
			}
		})
	}
}

func TestAddShipmentToCarrierManifestHandlerUpdatesCounts(t *testing.T) {
	store := shippingapp.NewPrototypeCarrierManifestStore()
	auditStore := audit.NewInMemoryLogStore()
	manifest := mustDraftCarrierManifestForHandler(t)
	if err := store.Save(context.Background(), manifest); err != nil {
		t.Fatalf("save manifest: %v", err)
	}
	service := shippingapp.NewAddShipmentToCarrierManifest(store, auditStore)
	body := bytes.NewBufferString(`{"shipment_id":"ship-hcm-260426-004"}`)
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/shipping/manifests/manifest-hcm-ghn-handler/shipments",
		body,
	)
	req.SetPathValue("manifest_id", manifest.ID)
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
	if payload.Data.Status != "draft" ||
		payload.Data.Summary.ExpectedCount != 1 ||
		payload.Data.Summary.MissingCount != 1 ||
		payload.Data.Lines[0].HandoverZoneCode != "HANDOVER-A" ||
		payload.Data.Lines[0].HandoverBinCode != "TOTE-A03" {
		t.Fatalf("manifest = %+v, want draft with expected 1 missing 1", payload.Data)
	}
}

func TestAddShipmentToCarrierManifestHandlerRejectsUnpackedAndWrongCarrier(t *testing.T) {
	cases := []struct {
		name       string
		shipmentID string
		want       string
	}{
		{name: "unpacked", shipmentID: "ship-hcm-260426-099", want: "Shipment must be packed"},
		{name: "wrong carrier", shipmentID: "ship-hcm-vtp-260426-001", want: "carrier does not match"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := shippingapp.NewPrototypeCarrierManifestStore()
			auditStore := audit.NewInMemoryLogStore()
			manifest := mustDraftCarrierManifestForHandler(t)
			if err := store.Save(context.Background(), manifest); err != nil {
				t.Fatalf("save manifest: %v", err)
			}
			service := shippingapp.NewAddShipmentToCarrierManifest(store, auditStore)
			body := bytes.NewBufferString(`{"shipment_id":"` + tc.shipmentID + `"}`)
			req := httptest.NewRequest(
				http.MethodPost,
				"/api/v1/shipping/manifests/manifest-hcm-ghn-handler/shipments",
				body,
			)
			req.SetPathValue("manifest_id", manifest.ID)
			req.Header.Set(response.HeaderRequestID, "req-manifest-add-reject")
			req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
				Email:       "admin@example.local",
				Password:    "local-only-mock-password",
				AccessToken: "local-dev-access-token",
			}, auth.RoleWarehouseLead)))
			rec := httptest.NewRecorder()

			addShipmentToCarrierManifestHandler(service).ServeHTTP(rec, req)

			if rec.Code != http.StatusConflict {
				t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusConflict, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), tc.want) {
				t.Fatalf("body = %s, want message containing %q", rec.Body.String(), tc.want)
			}
		})
	}
}

func TestCarrierManifestActionHandlersReadyRemoveAndCancel(t *testing.T) {
	store := shippingapp.NewPrototypeCarrierManifestStore()
	auditStore := audit.NewInMemoryLogStore()
	manifest := mustDraftCarrierManifestForHandler(t)
	if err := store.Save(context.Background(), manifest); err != nil {
		t.Fatalf("save manifest: %v", err)
	}
	if _, err := shippingapp.NewAddShipmentToCarrierManifest(store, auditStore).Execute(context.Background(), shippingapp.AddShipmentToCarrierManifestInput{
		ManifestID: manifest.ID,
		ShipmentID: "ship-hcm-260426-004",
		ActorID:    "user-warehouse-lead",
		RequestID:  "req-seed-manifest-line",
	}); err != nil {
		t.Fatalf("seed manifest line: %v", err)
	}

	readyReq := httptest.NewRequest(http.MethodPost, "/api/v1/shipping/manifests/manifest-hcm-ghn-handler/ready", nil)
	readyReq.SetPathValue("manifest_id", manifest.ID)
	readyReq.Header.Set(response.HeaderRequestID, "req-manifest-ready")
	readyReq = readyReq.WithContext(auth.WithPrincipal(readyReq.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead)))
	readyRec := httptest.NewRecorder()

	markCarrierManifestReadyToScanHandler(shippingapp.NewMarkCarrierManifestReadyToScan(store, auditStore)).ServeHTTP(readyRec, readyReq)
	if readyRec.Code != http.StatusOK {
		t.Fatalf("ready status = %d, want %d: %s", readyRec.Code, http.StatusOK, readyRec.Body.String())
	}
	var readyPayload response.SuccessEnvelope[carrierManifestResponse]
	if err := json.NewDecoder(readyRec.Body).Decode(&readyPayload); err != nil {
		t.Fatalf("decode ready response: %v", err)
	}
	if readyPayload.Data.Status != "ready" || readyPayload.Data.AuditLogID == "" {
		t.Fatalf("ready payload = %+v, want ready with audit", readyPayload.Data)
	}

	removeReq := httptest.NewRequest(http.MethodDelete, "/api/v1/shipping/manifests/manifest-hcm-ghn-handler/shipments/ship-hcm-260426-004", nil)
	removeReq.SetPathValue("manifest_id", manifest.ID)
	removeReq.SetPathValue("shipment_id", "ship-hcm-260426-004")
	removeReq.Header.Set(response.HeaderRequestID, "req-manifest-remove")
	removeReq = removeReq.WithContext(readyReq.Context())
	removeRec := httptest.NewRecorder()

	removeShipmentFromCarrierManifestHandler(shippingapp.NewRemoveShipmentFromCarrierManifest(store, auditStore)).ServeHTTP(removeRec, removeReq)
	if removeRec.Code != http.StatusOK {
		t.Fatalf("remove status = %d, want %d: %s", removeRec.Code, http.StatusOK, removeRec.Body.String())
	}
	var removePayload response.SuccessEnvelope[carrierManifestResponse]
	if err := json.NewDecoder(removeRec.Body).Decode(&removePayload); err != nil {
		t.Fatalf("decode remove response: %v", err)
	}
	if removePayload.Data.Status != "draft" || removePayload.Data.Summary.ExpectedCount != 0 || removePayload.Data.AuditLogID == "" {
		t.Fatalf("remove payload = %+v, want empty draft with audit", removePayload.Data)
	}

	cancelReq := httptest.NewRequest(http.MethodPost, "/api/v1/shipping/manifests/manifest-hcm-ghn-handler/cancel", bytes.NewBufferString(`{"reason":"carrier pickup moved"}`))
	cancelReq.SetPathValue("manifest_id", manifest.ID)
	cancelReq.Header.Set(response.HeaderRequestID, "req-manifest-cancel")
	cancelReq = cancelReq.WithContext(readyReq.Context())
	cancelRec := httptest.NewRecorder()

	cancelCarrierManifestHandler(shippingapp.NewCancelCarrierManifest(store, auditStore)).ServeHTTP(cancelRec, cancelReq)
	if cancelRec.Code != http.StatusOK {
		t.Fatalf("cancel status = %d, want %d: %s", cancelRec.Code, http.StatusOK, cancelRec.Body.String())
	}
	var cancelPayload response.SuccessEnvelope[carrierManifestResponse]
	if err := json.NewDecoder(cancelRec.Body).Decode(&cancelPayload); err != nil {
		t.Fatalf("decode cancel response: %v", err)
	}
	if cancelPayload.Data.Status != "cancelled" || cancelPayload.Data.AuditLogID == "" {
		t.Fatalf("cancel payload = %+v, want cancelled with audit", cancelPayload.Data)
	}
}

func TestReportCarrierManifestMissingOrdersHandlerMarksException(t *testing.T) {
	store := shippingapp.NewPrototypeCarrierManifestStore()
	auditStore := audit.NewInMemoryLogStore()
	service := shippingapp.NewReportCarrierManifestMissingOrders(store, auditStore)
	body := bytes.NewBufferString(`{"reason":"physical tote missing one order"}`)
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/shipping/manifests/manifest-hcm-ghn-morning/exceptions",
		body,
	)
	req.SetPathValue("manifest_id", "manifest-hcm-ghn-morning")
	req.Header.Set(response.HeaderRequestID, "req-manifest-missing")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead)))
	rec := httptest.NewRecorder()

	reportCarrierManifestMissingOrdersHandler(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var payload response.SuccessEnvelope[carrierManifestResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.Status != "exception" || payload.Data.AuditLogID == "" {
		t.Fatalf("payload = %+v, want exception manifest with audit", payload.Data)
	}
	if len(payload.Data.MissingLines) != 1 || payload.Data.MissingLines[0].OrderNo != "SO-260426-003" {
		t.Fatalf("missing lines = %+v, want SO-260426-003", payload.Data.MissingLines)
	}

	logs, err := auditStore.List(req.Context(), audit.Query{Action: "shipping.manifest.missing_exception_reported"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 || logs[0].Metadata["missing_count"] != 1 {
		t.Fatalf("audit logs = %+v, want missing exception count", logs)
	}
}

func TestVerifyCarrierManifestScanHandlerMarksLineAndWritesAudit(t *testing.T) {
	store := shippingapp.NewPrototypeCarrierManifestStore()
	auditStore := audit.NewInMemoryLogStore()
	service := shippingapp.NewVerifyCarrierManifestScan(store, auditStore)
	body := bytes.NewBufferString(`{"code":"GHN260426003","station_id":"dock-a","device_id":"scanner-01","source":"handheld_scanner"}`)
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
	if payload.Data.ScanEvent.DeviceID != "scanner-01" || payload.Data.ScanEvent.Source != "handheld_scanner" {
		t.Fatalf("scan event = %+v, want device/source retained", payload.Data.ScanEvent)
	}
	if payload.Data.ScanEvent.ManifestID != "manifest-hcm-ghn-morning" || payload.Data.ScanEvent.OrderNo != "SO-260426-003" || payload.Data.ScanEvent.TrackingNo != "GHN260426003" {
		t.Fatalf("scan event = %+v, want manifest/order/tracking retained", payload.Data.ScanEvent)
	}

	logs, err := auditStore.List(req.Context(), audit.Query{Action: "shipping.manifest.scan_recorded"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit logs = %d, want 1", len(logs))
	}
	if logs[0].AfterData["device_id"] != "scanner-01" || logs[0].Metadata["scan_source"] != "handheld_scanner" {
		t.Fatalf("audit log = %+v, want scanner device/source retained", logs[0])
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
		{name: "wrong carrier", code: "VTP260426012", want: "MANIFEST_MISMATCH"},
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

func TestReturnReceiptsHandlerReturnsFilteredRows(t *testing.T) {
	store := returnsapp.NewPrototypeReturnReceiptStore()
	listService := returnsapp.NewListReturnReceipts(store)
	receiveService := returnsapp.NewReceiveReturn(store, audit.NewInMemoryLogStore())
	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/returns/receipts?warehouse_id=wh-hcm&status=pending_inspection",
		nil,
	)
	req.Header.Set(response.HeaderRequestID, "req-return-list")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead)))
	rec := httptest.NewRecorder()

	returnReceiptsHandler(listService, receiveService).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var payload response.SuccessEnvelope[[]returnReceiptResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Data) != 1 {
		t.Fatalf("rows = %d, want 1", len(payload.Data))
	}
	if payload.Data[0].ReceiptNo != "RR-260426-0001" {
		t.Fatalf("receipt no = %q, want RR-260426-0001", payload.Data[0].ReceiptNo)
	}
}

func TestReturnReceiptsHandlerCreatesReusableReceiptMovementAndAudit(t *testing.T) {
	store := returnsapp.NewPrototypeReturnReceiptStore()
	auditStore := audit.NewInMemoryLogStore()
	listService := returnsapp.NewListReturnReceipts(store)
	receiveService := returnsapp.NewReceiveReturn(store, auditStore)
	body := bytes.NewBufferString(`{
		"warehouse_id": "wh-hcm",
		"warehouse_code": "HCM",
		"source": "CARRIER",
		"code": "GHN260426001",
		"package_condition": "sealed",
		"disposition": "reusable"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/returns/receipts", body)
	req.Header.Set(response.HeaderRequestID, "req-return-create")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead)))
	rec := httptest.NewRecorder()

	returnReceiptsHandler(listService, receiveService).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var payload response.SuccessEnvelope[returnReceiptResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.OriginalOrderNo != "SO-260426-001" {
		t.Fatalf("order no = %q, want SO-260426-001", payload.Data.OriginalOrderNo)
	}
	if payload.Data.StockMovement == nil || payload.Data.StockMovement.MovementType != "RETURN_RECEIPT" {
		t.Fatalf("stock movement = %+v, want RETURN_RECEIPT", payload.Data.StockMovement)
	}
	if payload.Data.AuditLogID == "" {
		t.Fatal("audit log id is empty")
	}

	logs, err := auditStore.List(req.Context(), audit.Query{Action: "returns.receipt.created"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit logs = %d, want 1", len(logs))
	}
}

func TestReturnReceiptsHandlerCreatesUnknownCase(t *testing.T) {
	store := returnsapp.NewPrototypeReturnReceiptStore()
	listService := returnsapp.NewListReturnReceipts(store)
	receiveService := returnsapp.NewReceiveReturn(store, audit.NewInMemoryLogStore())
	body := bytes.NewBufferString(`{
		"warehouse_id": "wh-hcm",
		"source": "SHIPPER",
		"code": "UNKNOWN-RETURN",
		"package_condition": "damaged box",
		"disposition": "needs_inspection"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/returns/receipts", body)
	req.Header.Set(response.HeaderRequestID, "req-return-unknown")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead)))
	rec := httptest.NewRecorder()

	returnReceiptsHandler(listService, receiveService).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var payload response.SuccessEnvelope[returnReceiptResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !payload.Data.UnknownCase {
		t.Fatal("unknown case = false, want true")
	}
	if payload.Data.TargetLocation != "return-inspection-queue" {
		t.Fatalf("target location = %q, want return-inspection-queue", payload.Data.TargetLocation)
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

func TestProductsHandlerListsFilteredMasterData(t *testing.T) {
	catalog := masterdataapp.NewPrototypeItemCatalog(audit.NewInMemoryLogStore())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/products?q=serum&status=active", nil)
	req.Header.Set(response.HeaderRequestID, "req-product-list")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleERPAdmin)))
	rec := httptest.NewRecorder()

	productsHandler(catalog).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var payload response.PaginatedSuccessEnvelope[[]productResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Data) != 1 {
		t.Fatalf("products = %d, want 1", len(payload.Data))
	}
	if payload.Data[0].SKUCode != "SERUM-30ML" || payload.Pagination.TotalItems != 1 {
		t.Fatalf("payload = %+v, want SERUM-30ML with one total item", payload)
	}
}

func TestProductsHandlerCreatesBlocksDuplicateAndWritesAudit(t *testing.T) {
	auditStore := audit.NewInMemoryLogStore()
	catalog := masterdataapp.NewPrototypeItemCatalog(auditStore)
	body := bytes.NewBufferString(`{
		"item_code": "ITEM-MASK-SET",
		"sku_code": "MASK-SET-05",
		"name": "Sheet Mask Set",
		"item_type": "finished_good",
		"item_group": "mask",
		"brand_code": "MYH",
		"uom_base": "EA",
		"lot_controlled": true,
		"expiry_controlled": true,
		"shelf_life_days": 540,
		"qc_required": true,
		"status": "draft",
		"is_sellable": true,
		"is_producible": true
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/products", body)
	req.Header.Set(response.HeaderRequestID, "req-product-create")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleERPAdmin)))
	rec := httptest.NewRecorder()

	productsHandler(catalog).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var payload response.SuccessEnvelope[productResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.ItemCode != "ITEM-MASK-SET" || payload.Data.AuditLogID == "" {
		t.Fatalf("product = %+v, want normalized item with audit id", payload.Data)
	}

	duplicate := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/products",
		bytes.NewBufferString(`{"item_code":"ITEM-MASK-SET","sku_code":"MASK-SET-99","name":"Duplicate","item_type":"finished_good","uom_base":"EA"}`),
	)
	duplicate = duplicate.WithContext(req.Context())
	duplicateRec := httptest.NewRecorder()
	productsHandler(catalog).ServeHTTP(duplicateRec, duplicate)
	if duplicateRec.Code != http.StatusConflict {
		t.Fatalf("duplicate status = %d, want %d: %s", duplicateRec.Code, http.StatusConflict, duplicateRec.Body.String())
	}

	logs, err := auditStore.List(req.Context(), audit.Query{Action: "masterdata.item.created"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit logs = %d, want 1", len(logs))
	}
}

func TestProductDetailHandlerUpdatesAndStatusChange(t *testing.T) {
	auditStore := audit.NewInMemoryLogStore()
	catalog := masterdataapp.NewPrototypeItemCatalog(auditStore)
	principalContext := auth.WithPrincipal(context.Background(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleERPAdmin))

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/products/item-serum-30ml", nil).WithContext(principalContext)
	getReq.SetPathValue("product_id", "item-serum-30ml")
	getRec := httptest.NewRecorder()
	productDetailHandler(catalog).ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d: %s", getRec.Code, http.StatusOK, getRec.Body.String())
	}

	updateBody := bytes.NewBufferString(`{
		"item_code": "ITEM-SERUM-HYDRA",
		"sku_code": "SERUM-30ML",
		"name": "Hydrating Serum 30ml v2",
		"item_type": "finished_good",
		"item_group": "serum",
		"brand_code": "MYH",
		"uom_base": "EA",
		"lot_controlled": true,
		"expiry_controlled": true,
		"shelf_life_days": 730,
		"qc_required": true,
		"status": "active",
		"is_sellable": true,
		"is_producible": true
	}`)
	updateReq := httptest.NewRequest(http.MethodPatch, "/api/v1/products/item-serum-30ml", updateBody).WithContext(principalContext)
	updateReq.SetPathValue("product_id", "item-serum-30ml")
	updateRec := httptest.NewRecorder()

	productDetailHandler(catalog).ServeHTTP(updateRec, updateReq)

	if updateRec.Code != http.StatusOK {
		t.Fatalf("update status = %d, want %d: %s", updateRec.Code, http.StatusOK, updateRec.Body.String())
	}
	var updatePayload response.SuccessEnvelope[productResponse]
	if err := json.NewDecoder(updateRec.Body).Decode(&updatePayload); err != nil {
		t.Fatalf("decode update response: %v", err)
	}
	if updatePayload.Data.Name != "Hydrating Serum 30ml v2" || updatePayload.Data.AuditLogID == "" {
		t.Fatalf("updated product = %+v, want changed name with audit", updatePayload.Data)
	}

	statusReq := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/products/item-serum-30ml/status",
		bytes.NewBufferString(`{"status":"inactive"}`),
	).WithContext(principalContext)
	statusReq.SetPathValue("product_id", "item-serum-30ml")
	statusRec := httptest.NewRecorder()

	changeProductStatusHandler(catalog).ServeHTTP(statusRec, statusReq)

	if statusRec.Code != http.StatusOK {
		t.Fatalf("status change = %d, want %d: %s", statusRec.Code, http.StatusOK, statusRec.Body.String())
	}
	var statusPayload response.SuccessEnvelope[productResponse]
	if err := json.NewDecoder(statusRec.Body).Decode(&statusPayload); err != nil {
		t.Fatalf("decode status response: %v", err)
	}
	if statusPayload.Data.Status != "inactive" || statusPayload.Data.AuditLogID == "" {
		t.Fatalf("status product = %+v, want inactive with audit", statusPayload.Data)
	}
}

func TestWarehousesHandlerCreatesBlocksDuplicateAndWritesAudit(t *testing.T) {
	auditStore := audit.NewInMemoryLogStore()
	catalog := masterdataapp.NewPrototypeWarehouseLocationCatalog(auditStore)
	body := bytes.NewBufferString(`{
		"warehouse_code": "WH-HN-FG",
		"warehouse_name": "Finished Goods Warehouse HN",
		"warehouse_type": "finished_good",
		"site_code": "HN",
		"address": "Ha Noi distribution center",
		"allow_sale_issue": true,
		"allow_prod_issue": false,
		"allow_quarantine": false,
		"status": "active"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/warehouses", body)
	req.Header.Set(response.HeaderRequestID, "req-warehouse-create")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleERPAdmin)))
	rec := httptest.NewRecorder()

	warehousesHandler(catalog).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var payload response.SuccessEnvelope[warehouseResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.WarehouseCode != "WH-HN-FG" || payload.Data.AuditLogID == "" {
		t.Fatalf("warehouse = %+v, want normalized warehouse with audit id", payload.Data)
	}

	duplicate := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/warehouses",
		bytes.NewBufferString(`{"warehouse_code":"WH-HN-FG","warehouse_name":"Duplicate","warehouse_type":"finished_good","site_code":"HN","status":"active"}`),
	)
	duplicate = duplicate.WithContext(req.Context())
	duplicateRec := httptest.NewRecorder()
	warehousesHandler(catalog).ServeHTTP(duplicateRec, duplicate)
	if duplicateRec.Code != http.StatusConflict {
		t.Fatalf("duplicate status = %d, want %d: %s", duplicateRec.Code, http.StatusConflict, duplicateRec.Body.String())
	}

	logs, err := auditStore.List(req.Context(), audit.Query{Action: "masterdata.warehouse.created"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit logs = %d, want 1", len(logs))
	}
}

func TestWarehouseLocationsHandlerBlocksInvalidWarehouseAndInactiveLocation(t *testing.T) {
	auditStore := audit.NewInMemoryLogStore()
	catalog := masterdataapp.NewPrototypeWarehouseLocationCatalog(auditStore)
	principalContext := auth.WithPrincipal(context.Background(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleERPAdmin))

	invalidWarehouse := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/warehouse-locations",
		bytes.NewBufferString(`{"warehouse_id":"missing-warehouse","location_code":"FG-PACK-02","location_name":"Packing Bay 02","location_type":"pack","zone_code":"PACK","allow_store":true,"status":"active"}`),
	).WithContext(principalContext)
	invalidRec := httptest.NewRecorder()
	warehouseLocationsHandler(catalog).ServeHTTP(invalidRec, invalidWarehouse)
	if invalidRec.Code != http.StatusBadRequest {
		t.Fatalf("invalid warehouse status = %d, want %d: %s", invalidRec.Code, http.StatusBadRequest, invalidRec.Body.String())
	}

	createReq := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/warehouse-locations",
		bytes.NewBufferString(`{"warehouse_id":"wh-hcm-fg","location_code":"FG-PACK-02","location_name":"Packing Bay 02","location_type":"pack","zone_code":"PACK","allow_store":true,"status":"active"}`),
	).WithContext(principalContext)
	createReq.Header.Set(response.HeaderRequestID, "req-location-create")
	createRec := httptest.NewRecorder()
	warehouseLocationsHandler(catalog).ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d: %s", createRec.Code, http.StatusCreated, createRec.Body.String())
	}

	var created response.SuccessEnvelope[warehouseLocationResponse]
	if err := json.NewDecoder(createRec.Body).Decode(&created); err != nil {
		t.Fatalf("decode created location: %v", err)
	}

	statusReq := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/warehouse-locations/"+created.Data.ID+"/status",
		bytes.NewBufferString(`{"status":"inactive"}`),
	).WithContext(principalContext)
	statusReq.SetPathValue("location_id", created.Data.ID)
	statusRec := httptest.NewRecorder()
	changeWarehouseLocationStatusHandler(catalog).ServeHTTP(statusRec, statusReq)
	if statusRec.Code != http.StatusOK {
		t.Fatalf("status change = %d, want %d: %s", statusRec.Code, http.StatusOK, statusRec.Body.String())
	}

	activeReq := httptest.NewRequest(http.MethodGet, "/api/v1/warehouse-locations?warehouse_id=wh-hcm-fg&status=active", nil).WithContext(principalContext)
	activeRec := httptest.NewRecorder()
	warehouseLocationsHandler(catalog).ServeHTTP(activeRec, activeReq)
	if activeRec.Code != http.StatusOK {
		t.Fatalf("active list = %d, want %d: %s", activeRec.Code, http.StatusOK, activeRec.Body.String())
	}
	var activePayload response.PaginatedSuccessEnvelope[[]warehouseLocationResponse]
	if err := json.NewDecoder(activeRec.Body).Decode(&activePayload); err != nil {
		t.Fatalf("decode active locations: %v", err)
	}
	for _, location := range activePayload.Data {
		if location.ID == created.Data.ID {
			t.Fatalf("inactive location %q was returned in active list", location.ID)
		}
	}
}

func TestSuppliersHandlerCreatesBlocksDuplicateAndWritesAudit(t *testing.T) {
	auditStore := audit.NewInMemoryLogStore()
	catalog := masterdataapp.NewPrototypePartyCatalog(auditStore)
	body := bytes.NewBufferString(`{
		"supplier_code": "SUP-SVC-LAB",
		"supplier_name": "Lab Services Partner",
		"supplier_group": "service",
		"contact_name": "Nguyen Lab",
		"email": "Lab@Partner.Example",
		"tax_code": "0319999001",
		"address": "Ho Chi Minh lab site",
		"payment_terms": "NET15",
		"lead_time_days": 5,
		"quality_score": "90.0000",
		"delivery_score": "92.0000",
		"status": "draft"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/suppliers", body)
	req.Header.Set(response.HeaderRequestID, "req-supplier-create")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleERPAdmin)))
	rec := httptest.NewRecorder()

	suppliersHandler(catalog).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var payload response.SuccessEnvelope[supplierResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.SupplierCode != "SUP-SVC-LAB" || payload.Data.Email != "lab@partner.example" || payload.Data.AuditLogID == "" {
		t.Fatalf("supplier = %+v, want normalized supplier with audit id", payload.Data)
	}

	duplicate := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/suppliers",
		bytes.NewBufferString(`{"supplier_code":"SUP-SVC-LAB","supplier_name":"Duplicate","supplier_group":"service","status":"active"}`),
	).WithContext(req.Context())
	duplicateRec := httptest.NewRecorder()
	suppliersHandler(catalog).ServeHTTP(duplicateRec, duplicate)
	if duplicateRec.Code != http.StatusConflict {
		t.Fatalf("duplicate status = %d, want %d: %s", duplicateRec.Code, http.StatusConflict, duplicateRec.Body.String())
	}

	logs, err := auditStore.List(req.Context(), audit.Query{Action: "masterdata.supplier.created"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit logs = %d, want 1", len(logs))
	}
}

func TestSupplierDetailHandlerUpdatesAndStatusChange(t *testing.T) {
	auditStore := audit.NewInMemoryLogStore()
	catalog := masterdataapp.NewPrototypePartyCatalog(auditStore)
	principalContext := auth.WithPrincipal(context.Background(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleERPAdmin))

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/suppliers/sup-rm-bioactive", nil).WithContext(principalContext)
	getReq.SetPathValue("supplier_id", "sup-rm-bioactive")
	getRec := httptest.NewRecorder()
	supplierDetailHandler(catalog).ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d: %s", getRec.Code, http.StatusOK, getRec.Body.String())
	}

	updateBody := bytes.NewBufferString(`{
		"supplier_code": "SUP-RM-BIO",
		"supplier_name": "BioActive Raw Materials v2",
		"supplier_group": "raw_material",
		"contact_name": "Nguyen Van An",
		"email": "purchasing@bioactive.example",
		"tax_code": "0312345001",
		"address": "Binh Duong",
		"payment_terms": "NET45",
		"lead_time_days": 10,
		"moq": "60.000000",
		"quality_score": "95.0000",
		"delivery_score": "92.0000",
		"status": "active"
	}`)
	updateReq := httptest.NewRequest(http.MethodPatch, "/api/v1/suppliers/sup-rm-bioactive", updateBody).WithContext(principalContext)
	updateReq.SetPathValue("supplier_id", "sup-rm-bioactive")
	updateReq.Header.Set(response.HeaderRequestID, "req-supplier-update")
	updateRec := httptest.NewRecorder()

	supplierDetailHandler(catalog).ServeHTTP(updateRec, updateReq)

	if updateRec.Code != http.StatusOK {
		t.Fatalf("update status = %d, want %d: %s", updateRec.Code, http.StatusOK, updateRec.Body.String())
	}
	var updatePayload response.SuccessEnvelope[supplierResponse]
	if err := json.NewDecoder(updateRec.Body).Decode(&updatePayload); err != nil {
		t.Fatalf("decode update response: %v", err)
	}
	if updatePayload.Data.SupplierName != "BioActive Raw Materials v2" || updatePayload.Data.AuditLogID == "" {
		t.Fatalf("updated supplier = %+v, want changed name with audit", updatePayload.Data)
	}

	statusReq := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/suppliers/sup-rm-bioactive/status",
		bytes.NewBufferString(`{"status":"inactive"}`),
	).WithContext(principalContext)
	statusReq.SetPathValue("supplier_id", "sup-rm-bioactive")
	statusReq.Header.Set(response.HeaderRequestID, "req-supplier-status")
	statusRec := httptest.NewRecorder()

	changeSupplierStatusHandler(catalog).ServeHTTP(statusRec, statusReq)

	if statusRec.Code != http.StatusOK {
		t.Fatalf("status change = %d, want %d: %s", statusRec.Code, http.StatusOK, statusRec.Body.String())
	}
	var statusPayload response.SuccessEnvelope[supplierResponse]
	if err := json.NewDecoder(statusRec.Body).Decode(&statusPayload); err != nil {
		t.Fatalf("decode status response: %v", err)
	}
	if statusPayload.Data.Status != "inactive" || statusPayload.Data.AuditLogID == "" {
		t.Fatalf("status supplier = %+v, want inactive with audit", statusPayload.Data)
	}
}

func TestCustomersHandlerCreatesBlocksDuplicateAndWritesAudit(t *testing.T) {
	auditStore := audit.NewInMemoryLogStore()
	catalog := masterdataapp.NewPrototypePartyCatalog(auditStore)
	body := bytes.NewBufferString(`{
		"customer_code": "CUS-DL-HANOI",
		"customer_name": "Ha Noi Dealer",
		"customer_type": "dealer",
		"channel_code": "dealer",
		"price_list_code": "pl-dealer-2026",
		"discount_group": "tier_2",
		"credit_limit": "200000000.00",
		"payment_terms": "NET15",
		"contact_name": "Tran Ha Noi",
		"email": "Buyer@HaNoiDealer.Example",
		"tax_code": "0319999002",
		"address": "Ha Noi",
		"status": "draft"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/customers", body)
	req.Header.Set(response.HeaderRequestID, "req-customer-create")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleERPAdmin)))
	rec := httptest.NewRecorder()

	customersHandler(catalog).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var payload response.SuccessEnvelope[customerResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.CustomerCode != "CUS-DL-HANOI" || payload.Data.ChannelCode != "DEALER" || payload.Data.AuditLogID == "" {
		t.Fatalf("customer = %+v, want normalized customer with audit id", payload.Data)
	}

	duplicate := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/customers",
		bytes.NewBufferString(`{"customer_code":"CUS-DL-HANOI","customer_name":"Duplicate","customer_type":"dealer","status":"active"}`),
	).WithContext(req.Context())
	duplicateRec := httptest.NewRecorder()
	customersHandler(catalog).ServeHTTP(duplicateRec, duplicate)
	if duplicateRec.Code != http.StatusConflict {
		t.Fatalf("duplicate status = %d, want %d: %s", duplicateRec.Code, http.StatusConflict, duplicateRec.Body.String())
	}

	logs, err := auditStore.List(req.Context(), audit.Query{Action: "masterdata.customer.created"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit logs = %d, want 1", len(logs))
	}
}

func TestCustomerDetailHandlerUpdatesAndStatusChange(t *testing.T) {
	auditStore := audit.NewInMemoryLogStore()
	catalog := masterdataapp.NewPrototypePartyCatalog(auditStore)
	principalContext := auth.WithPrincipal(context.Background(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleERPAdmin))

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/customers/cus-dl-minh-anh", nil).WithContext(principalContext)
	getReq.SetPathValue("customer_id", "cus-dl-minh-anh")
	getRec := httptest.NewRecorder()
	customerDetailHandler(catalog).ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d: %s", getRec.Code, http.StatusOK, getRec.Body.String())
	}

	updateBody := bytes.NewBufferString(`{
		"customer_code": "CUS-DL-MINHANH",
		"customer_name": "Minh Anh Distributor v2",
		"customer_type": "distributor",
		"channel_code": "B2B",
		"price_list_code": "PL-B2B-2026",
		"discount_group": "tier_1",
		"credit_limit": "550000000.00",
		"payment_terms": "NET30",
		"contact_name": "Do Minh Anh",
		"email": "orders@minhanh.example",
		"tax_code": "0315678001",
		"address": "District 7, Ho Chi Minh City",
		"status": "active"
	}`)
	updateReq := httptest.NewRequest(http.MethodPatch, "/api/v1/customers/cus-dl-minh-anh", updateBody).WithContext(principalContext)
	updateReq.SetPathValue("customer_id", "cus-dl-minh-anh")
	updateReq.Header.Set(response.HeaderRequestID, "req-customer-update")
	updateRec := httptest.NewRecorder()

	customerDetailHandler(catalog).ServeHTTP(updateRec, updateReq)

	if updateRec.Code != http.StatusOK {
		t.Fatalf("update status = %d, want %d: %s", updateRec.Code, http.StatusOK, updateRec.Body.String())
	}
	var updatePayload response.SuccessEnvelope[customerResponse]
	if err := json.NewDecoder(updateRec.Body).Decode(&updatePayload); err != nil {
		t.Fatalf("decode update response: %v", err)
	}
	if updatePayload.Data.CustomerName != "Minh Anh Distributor v2" || updatePayload.Data.AuditLogID == "" {
		t.Fatalf("updated customer = %+v, want changed name with audit", updatePayload.Data)
	}

	statusReq := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/customers/cus-dl-minh-anh/status",
		bytes.NewBufferString(`{"status":"inactive"}`),
	).WithContext(principalContext)
	statusReq.SetPathValue("customer_id", "cus-dl-minh-anh")
	statusReq.Header.Set(response.HeaderRequestID, "req-customer-status")
	statusRec := httptest.NewRecorder()

	changeCustomerStatusHandler(catalog).ServeHTTP(statusRec, statusReq)

	if statusRec.Code != http.StatusOK {
		t.Fatalf("status change = %d, want %d: %s", statusRec.Code, http.StatusOK, statusRec.Body.String())
	}
	var statusPayload response.SuccessEnvelope[customerResponse]
	if err := json.NewDecoder(statusRec.Body).Decode(&statusPayload); err != nil {
		t.Fatalf("decode status response: %v", err)
	}
	if statusPayload.Data.Status != "inactive" || statusPayload.Data.AuditLogID == "" {
		t.Fatalf("status customer = %+v, want inactive with audit", statusPayload.Data)
	}
}

func TestStockMovementHandlerWritesAuditForAdjustment(t *testing.T) {
	store := audit.NewInMemoryLogStore()
	body := bytes.NewBufferString(`{
		"movementId": "mov-adjust-test",
		"sku": "serum-30ml",
		"warehouseId": "wh-hcm",
		"movementType": "ADJUST",
		"quantity": "8.000000",
		"baseUomCode": "PCS",
		"sourceQuantity": "2.000000",
		"sourceUomCode": "CARTON",
		"conversionFactor": "4.000000",
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
	var payload response.SuccessEnvelope[stockMovementResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode stock movement response: %v", err)
	}
	if payload.Data.MovementQuantity != "8.000000" || payload.Data.BaseUOMCode != "PCS" || payload.Data.SourceUOMCode != "CARTON" || payload.Data.ConversionFactor != "4.000000" {
		t.Fatalf("stock movement response = %+v, want decimal/base/source UOM fields", payload.Data)
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
	if got.AfterData["movement_qty"] != "8.000000" || got.AfterData["base_uom_code"] != "PCS" || got.AfterData["source_uom_code"] != "CARTON" {
		t.Fatalf("audit after data = %+v, want decimal/base/source UOM metadata", got.AfterData)
	}
}

func newTestGoodsReceiptService() (inventoryapp.WarehouseReceivingService, *audit.InMemoryLogStore) {
	auditStore := audit.NewInMemoryLogStore()
	service := inventoryapp.NewWarehouseReceivingService(
		inventoryapp.NewPrototypeWarehouseReceivingStore(),
		masterdataapp.NewPrototypeWarehouseLocationCatalog(auditStore),
		inventoryapp.NewPrototypeBatchCatalog(auditStore),
		inventoryapp.NewInMemoryStockMovementStore(),
		auditStore,
	)

	return service, auditStore
}

func mustDraftCarrierManifestForHandler(t *testing.T) shippingdomain.CarrierManifest {
	t.Helper()

	manifest, err := shippingdomain.NewCarrierManifest(shippingdomain.NewCarrierManifestInput{
		ID:               "manifest-hcm-ghn-handler",
		CarrierCode:      "GHN",
		CarrierName:      "GHN Express",
		WarehouseID:      "wh-hcm",
		WarehouseCode:    "HCM",
		Date:             "2026-04-28",
		HandoverBatch:    "afternoon",
		StagingZone:      "handover-a",
		HandoverZoneCode: "HANDOVER-A",
		HandoverBinCode:  "TOTE-A01",
		Owner:            "Warehouse Lead",
	})
	if err != nil {
		t.Fatalf("new carrier manifest: %v", err)
	}

	return manifest
}
