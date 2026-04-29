package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	masterdataapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/application"
	productionapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/application"
	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	purchaseapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/purchase/application"
	purchasedomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/purchase/domain"
	qcapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/qc/application"
	qcdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/qc/domain"
	returnsapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/returns/application"
	returnsdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/returns/domain"
	salesapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/application"
	salesdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/domain"
	shippingapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/application"
	shippingdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
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
		"supplier_id": "supplier-local",
		"delivery_note_no": "dn-260427-api",
		"lines": [
			{
				"id": "line-api-test",
				"purchase_order_line_id": "po-line-260427-api-001",
				"batch_id": "batch-cream-2603b",
				"quantity": "6",
				"uom_code": "EA",
				"base_uom_code": "EA",
				"packaging_status": "intact"
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
	if createPayload.Data.Status != "draft" ||
		createPayload.Data.DeliveryNoteNo != "DN-260427-API" ||
		createPayload.Data.Lines[0].SKU != "CREAM-50G" ||
		createPayload.Data.Lines[0].PurchaseOrderLineID != "po-line-260427-api-001" ||
		createPayload.Data.Lines[0].LotNo != "LOT-2603B" ||
		createPayload.Data.Lines[0].ExpiryDate != "2028-03-01" ||
		createPayload.Data.Lines[0].PackagingStatus != "intact" {
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
	service, auditStore := newTestGoodsReceiptService()
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
	logs, err := auditStore.List(req.Context(), audit.Query{Action: "inventory.receiving.posted"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 0 {
		t.Fatalf("posted audit log count = %d, want 0 for denied action", len(logs))
	}
}

func TestEndOfDayReconciliationsHandlerReturnsFilteredRows(t *testing.T) {
	store := inventoryapp.NewPrototypeEndOfDayReconciliationStore()
	service := inventoryapp.NewListEndOfDayReconciliations(store)
	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/warehouse/end-of-day-reconciliations?warehouse_id=wh-hcm&date=2026-04-26&shift_code=day&status=in_review",
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
	if payload.Data[0].Operations.HandoverOrderCount != 27 {
		t.Fatalf("handover order count = %d, want 27", payload.Data[0].Operations.HandoverOrderCount)
	}
}

func TestCloseEndOfDayReconciliationHandlerWritesAudit(t *testing.T) {
	store := inventoryapp.NewPrototypeEndOfDayReconciliationStore()
	auditStore := audit.NewInMemoryLogStore()
	service := inventoryapp.NewCloseEndOfDayReconciliation(store, auditStore)
	body := bytes.NewBufferString(`{"exception_note":""}`)
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/warehouse/end-of-day-reconciliations/rec-hn-260426-day/close",
		body,
	)
	req.SetPathValue("reconciliation_id", "rec-hn-260426-day")
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
	if payload.Data.Operations.PendingIssueCount != 0 {
		t.Fatalf("pending issue count = %d, want 0", payload.Data.Operations.PendingIssueCount)
	}

	logs, err := auditStore.List(req.Context(), audit.Query{Action: "warehouse.shift.closed"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit logs = %d, want 1", len(logs))
	}
}

func TestCloseEndOfDayReconciliationHandlerBlocksUnresolvedIssue(t *testing.T) {
	store := inventoryapp.NewPrototypeEndOfDayReconciliationStore()
	service := inventoryapp.NewCloseEndOfDayReconciliation(store, audit.NewInMemoryLogStore())
	body := bytes.NewBufferString(`{"exception_note":"variance accepted by lead"}`)
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/warehouse/end-of-day-reconciliations/rec-hcm-260426-day/close",
		body,
	)
	req.SetPathValue("reconciliation_id", "rec-hcm-260426-day")
	req.Header.Set(response.HeaderRequestID, "req-close-blocked")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead)))
	rec := httptest.NewRecorder()

	closeEndOfDayReconciliationHandler(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
}

func TestCloseEndOfDayReconciliationHandlerRequiresCreatePermission(t *testing.T) {
	store := inventoryapp.NewPrototypeEndOfDayReconciliationStore()
	auditStore := audit.NewInMemoryLogStore()
	service := inventoryapp.NewCloseEndOfDayReconciliation(store, auditStore)
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/warehouse/end-of-day-reconciliations/rec-hn-260426-day/close",
		bytes.NewBufferString(`{"exception_note":""}`),
	)
	req.SetPathValue("reconciliation_id", "rec-hn-260426-day")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseStaff)))
	rec := httptest.NewRecorder()

	closeEndOfDayReconciliationHandler(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
	logs, err := auditStore.List(req.Context(), audit.Query{Action: "warehouse.shift.closed"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 0 {
		t.Fatalf("audit logs = %d, want 0", len(logs))
	}
}

func TestCloseEndOfDayReconciliationHandlerAllowsVarianceExceptionForWarehouseLead(t *testing.T) {
	store := inventoryapp.NewPrototypeEndOfDayReconciliationStore()
	if err := store.Save(context.Background(), inventorydomain.EndOfDayReconciliation{
		ID:            "rec-hcm-variance-only",
		WarehouseID:   "wh-hcm",
		WarehouseCode: "HCM",
		Date:          "2026-04-26",
		ShiftCode:     "day",
		Status:        inventorydomain.ReconciliationStatusInReview,
		Owner:         "Warehouse Lead",
		Checklist: []inventorydomain.ReconciliationChecklistItem{
			{Key: "variance", Label: "Stock variance reviewed", Complete: false, Blocking: true},
		},
		Lines: []inventorydomain.ReconciliationLine{
			{
				ID:              "line-variance-only",
				SKU:             "SERUM-30ML",
				BatchNo:         "LOT-2604A",
				BinCode:         "A-01",
				SystemQuantity:  120,
				CountedQuantity: 118,
				Owner:           "Warehouse Lead",
			},
		},
	}); err != nil {
		t.Fatalf("save reconciliation: %v", err)
	}
	auditStore := audit.NewInMemoryLogStore()
	service := inventoryapp.NewCloseEndOfDayReconciliation(store, auditStore)
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/warehouse/end-of-day-reconciliations/rec-hcm-variance-only/close",
		bytes.NewBufferString(`{"exception_note":"variance accepted by lead"}`),
	)
	req.SetPathValue("reconciliation_id", "rec-hcm-variance-only")
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

func TestWarehouseDailyBoardFulfillmentMetricsHandlerSummarizesOrderStates(t *testing.T) {
	salesService, _ := newTestSalesOrderAPIService()
	manifestStore := shippingapp.NewPrototypeCarrierManifestStore()
	handler := warehouseDailyBoardFulfillmentMetricsHandler(
		salesService,
		shippingapp.NewListCarrierManifests(manifestStore),
	)
	authConfig := auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}

	cases := []struct {
		name            string
		path            string
		wantTotal       int
		wantNew         int
		wantPicking     int
		wantWaiting     int
		wantMissing     int
		wantCarrierCode string
	}{
		{
			name:            "carrier handover day",
			path:            "/api/v1/warehouse/daily-board/fulfillment-metrics?warehouse_id=wh-hcm&date=2026-04-26&shift_code=day&carrier_code=GHN",
			wantTotal:       3,
			wantWaiting:     3,
			wantMissing:     1,
			wantCarrierCode: "GHN",
		},
		{
			name:        "sales operation day",
			path:        "/api/v1/warehouse/daily-board/fulfillment-metrics?warehouse_id=wh-hcm-fg&date=2026-04-28",
			wantTotal:   3,
			wantNew:     2,
			wantPicking: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			req.Header.Set(response.HeaderRequestID, "req-warehouse-fulfillment-metrics")
			req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(authConfig, auth.RoleWarehouseLead)))
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
			}
			var payload response.SuccessEnvelope[warehouseFulfillmentMetricsResponse]
			if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if payload.Data.TotalOrders != tc.wantTotal ||
				payload.Data.NewOrders != tc.wantNew ||
				payload.Data.PickingOrders != tc.wantPicking ||
				payload.Data.WaitingHandoverOrders != tc.wantWaiting ||
				payload.Data.MissingOrders != tc.wantMissing ||
				payload.Data.CarrierCode != tc.wantCarrierCode {
				t.Fatalf("metrics = %+v, want total/new/picking/waiting/missing/carrier = %d/%d/%d/%d/%d/%s",
					payload.Data,
					tc.wantTotal,
					tc.wantNew,
					tc.wantPicking,
					tc.wantWaiting,
					tc.wantMissing,
					tc.wantCarrierCode,
				)
			}
			if payload.Data.GeneratedAt == "" {
				t.Fatal("generated_at is empty")
			}
		})
	}
}

func TestWarehouseDailyBoardFulfillmentMetricsMatchSalesAndManifestState(t *testing.T) {
	salesService, _ := newTestSalesOrderAPIService()
	manifestStore := shippingapp.NewPrototypeCarrierManifestStore()
	listManifests := shippingapp.NewListCarrierManifests(manifestStore)
	ctx := context.Background()
	orders, err := salesService.ListSalesOrders(ctx, salesapp.SalesOrderFilter{
		WarehouseID: "wh-hcm",
		DateFrom:    "2026-04-26",
		DateTo:      "2026-04-26",
	})
	if err != nil {
		t.Fatalf("list sales orders: %v", err)
	}
	manifests, err := listManifests.Execute(
		ctx,
		shippingdomain.NewCarrierManifestFilter("wh-hcm", "2026-04-26", "GHN", ""),
	)
	if err != nil {
		t.Fatalf("list carrier manifests: %v", err)
	}
	expected := expectedFulfillmentMetricsFromSources(orders, manifests, "GHN")
	authConfig := auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}
	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/warehouse/daily-board/fulfillment-metrics?warehouse_id=wh-hcm&date=2026-04-26&carrier_code=GHN",
		nil,
	)
	req.Header.Set(response.HeaderRequestID, "req-warehouse-fulfillment-consistency")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(authConfig, auth.RoleWarehouseLead)))
	rec := httptest.NewRecorder()

	warehouseDailyBoardFulfillmentMetricsHandler(salesService, listManifests).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var payload response.SuccessEnvelope[warehouseFulfillmentMetricsResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.TotalOrders != expected.total ||
		payload.Data.NewOrders != expected.newOrders ||
		payload.Data.ReservedOrders != expected.reserved ||
		payload.Data.PickingOrders != expected.picking ||
		payload.Data.PackedOrders != expected.packed ||
		payload.Data.WaitingHandoverOrders != expected.waitingHandover ||
		payload.Data.MissingOrders != expected.missing ||
		payload.Data.HandoverOrders != expected.handover {
		t.Fatalf("metrics = %+v, want consistency = %+v", payload.Data, expected)
	}
}

func TestWarehouseDailyBoardInboundMetricsHandlerSummarizesInboundState(t *testing.T) {
	businessDay := "2026-04-29"
	businessDayTimeUTC := time.Date(2026, 4, 28, 18, 0, 0, 0, time.UTC)
	authConfig := auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}
	handler := warehouseDailyBoardInboundMetricsHandler(
		testWarehouseDailyBoardPurchaseOrderLister{orders: []purchasedomain.PurchaseOrder{
			mustWarehouseDailyBoardPurchaseOrder(t, "po-approved", purchasedomain.PurchaseOrderStatusApproved, "wh-hcm-fg", businessDay),
			mustWarehouseDailyBoardPurchaseOrder(t, "po-partial", purchasedomain.PurchaseOrderStatusPartiallyReceived, "wh-hcm-fg", businessDay),
			mustWarehouseDailyBoardPurchaseOrder(t, "po-received", purchasedomain.PurchaseOrderStatusReceived, "wh-hcm-fg", businessDay),
			mustWarehouseDailyBoardPurchaseOrder(t, "po-other-warehouse", purchasedomain.PurchaseOrderStatusApproved, "wh-da-nang", businessDay),
			mustWarehouseDailyBoardPurchaseOrder(t, "po-other-day", purchasedomain.PurchaseOrderStatusApproved, "wh-hcm-fg", "2026-04-30"),
		}},
		testWarehouseDailyBoardReceivingLister{receipts: []inventorydomain.WarehouseReceiving{
			mustWarehouseDailyBoardReceiving(t, "grn-draft", inventorydomain.WarehouseReceivingStatusDraft, "wh-hcm-fg", businessDayTimeUTC),
			mustWarehouseDailyBoardReceiving(t, "grn-submitted", inventorydomain.WarehouseReceivingStatusSubmitted, "wh-hcm-fg", businessDayTimeUTC),
			mustWarehouseDailyBoardReceiving(t, "grn-inspect", inventorydomain.WarehouseReceivingStatusInspectReady, "wh-hcm-fg", businessDayTimeUTC),
			mustWarehouseDailyBoardReceiving(t, "grn-posted", inventorydomain.WarehouseReceivingStatusPosted, "wh-hcm-fg", businessDayTimeUTC),
			mustWarehouseDailyBoardReceiving(t, "grn-other-warehouse", inventorydomain.WarehouseReceivingStatusDraft, "wh-da-nang", businessDayTimeUTC),
		}},
		testWarehouseDailyBoardInboundQCLister{inspections: []qcdomain.InboundQCInspection{
			{ID: "iqc-hold", WarehouseID: "wh-hcm-fg", Status: qcdomain.InboundQCInspectionStatusCompleted, Result: qcdomain.InboundQCResultHold, CreatedAt: businessDayTimeUTC},
			{ID: "iqc-fail", WarehouseID: "wh-hcm-fg", Status: qcdomain.InboundQCInspectionStatusCompleted, Result: qcdomain.InboundQCResultFail, CreatedAt: businessDayTimeUTC},
			{ID: "iqc-pass", WarehouseID: "wh-hcm-fg", Status: qcdomain.InboundQCInspectionStatusCompleted, Result: qcdomain.InboundQCResultPass, CreatedAt: businessDayTimeUTC},
			{ID: "iqc-partial", WarehouseID: "wh-hcm-fg", Status: qcdomain.InboundQCInspectionStatusCompleted, Result: qcdomain.InboundQCResultPartial, CreatedAt: businessDayTimeUTC},
			{ID: "iqc-pending", WarehouseID: "wh-hcm-fg", Status: qcdomain.InboundQCInspectionStatusPending, CreatedAt: businessDayTimeUTC},
			{ID: "iqc-other-warehouse", WarehouseID: "wh-da-nang", Status: qcdomain.InboundQCInspectionStatusCompleted, Result: qcdomain.InboundQCResultFail, CreatedAt: businessDayTimeUTC},
		}},
		testWarehouseDailyBoardSupplierRejectionLister{rejections: []inventorydomain.SupplierRejection{
			{ID: "sr-draft", WarehouseID: "wh-hcm-fg", Status: inventorydomain.SupplierRejectionStatusDraft, CreatedAt: businessDayTimeUTC},
			{ID: "sr-submitted", WarehouseID: "wh-hcm-fg", Status: inventorydomain.SupplierRejectionStatusSubmitted, CreatedAt: businessDayTimeUTC},
			{ID: "sr-confirmed", WarehouseID: "wh-hcm-fg", Status: inventorydomain.SupplierRejectionStatusConfirmed, CreatedAt: businessDayTimeUTC},
			{ID: "sr-cancelled", WarehouseID: "wh-hcm-fg", Status: inventorydomain.SupplierRejectionStatusCancelled, CreatedAt: businessDayTimeUTC},
			{ID: "sr-other-day", WarehouseID: "wh-hcm-fg", Status: inventorydomain.SupplierRejectionStatusDraft, CreatedAt: time.Date(2026, 4, 30, 2, 0, 0, 0, time.UTC)},
		}},
	)
	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/warehouse/daily-board/inbound-metrics?warehouse_id=wh-hcm-fg&date=2026-04-29&shift_code=day",
		nil,
	)
	req.Header.Set(response.HeaderRequestID, "req-warehouse-inbound-metrics")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(authConfig, auth.RoleWarehouseLead)))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var payload response.SuccessEnvelope[warehouseInboundMetricsResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.PurchaseOrdersIncoming != 2 ||
		payload.Data.ReceivingPending != 3 ||
		payload.Data.ReceivingDraft != 1 ||
		payload.Data.ReceivingSubmitted != 1 ||
		payload.Data.ReceivingInspectReady != 1 ||
		payload.Data.QCHold != 1 ||
		payload.Data.QCFail != 1 ||
		payload.Data.QCPass != 1 ||
		payload.Data.QCPartial != 1 ||
		payload.Data.SupplierRejections != 3 ||
		payload.Data.SupplierRejectionDraft != 1 ||
		payload.Data.SupplierRejectionSubmitted != 1 ||
		payload.Data.SupplierRejectionConfirmed != 1 ||
		payload.Data.SupplierRejectionCancelled != 1 {
		t.Fatalf("metrics = %+v, want inbound source counts", payload.Data)
	}
	if payload.Data.WarehouseID != "wh-hcm-fg" || payload.Data.Date != businessDay || payload.Data.ShiftCode != "day" {
		t.Fatalf("scope = %+v, want wh-hcm-fg/%s/day", payload.Data, businessDay)
	}
	if payload.Data.GeneratedAt == "" {
		t.Fatal("generated_at is empty")
	}
}

func TestWarehouseDailyBoardInboundMetricsHandlerRequiresWarehousePermission(t *testing.T) {
	handler := warehouseDailyBoardInboundMetricsHandler(
		testWarehouseDailyBoardPurchaseOrderLister{},
		testWarehouseDailyBoardReceivingLister{},
		testWarehouseDailyBoardInboundQCLister{},
		testWarehouseDailyBoardSupplierRejectionLister{},
	)
	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/warehouse/daily-board/inbound-metrics?warehouse_id=wh-hcm-fg&date=2026-04-29",
		nil,
	)
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "sales@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleSalesOps)))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestWarehouseDailyBoardInboundMetricsMatchSourceState(t *testing.T) {
	purchaseService := newTestWarehouseDailyBoardPurchaseOrderService(t)
	receivingService, _ := newTestGoodsReceiptService()
	businessDayTimeUTC := time.Date(2026, 4, 27, 9, 0, 0, 0, time.UTC)
	inboundQCService := qcapp.NewInboundQCInspectionService(
		qcapp.NewPrototypeInboundQCInspectionStore(
			qcdomain.InboundQCInspection{
				ID: "iqc-board-pass", WarehouseID: "wh-hcm-fg", Status: qcdomain.InboundQCInspectionStatusCompleted,
				Result: qcdomain.InboundQCResultPass, CreatedAt: businessDayTimeUTC,
			},
			qcdomain.InboundQCInspection{
				ID: "iqc-board-fail", WarehouseID: "wh-hcm-fg", Status: qcdomain.InboundQCInspectionStatusCompleted,
				Result: qcdomain.InboundQCResultFail, CreatedAt: businessDayTimeUTC,
			},
			qcdomain.InboundQCInspection{
				ID: "iqc-board-hold", WarehouseID: "wh-hcm-fg", Status: qcdomain.InboundQCInspectionStatusCompleted,
				Result: qcdomain.InboundQCResultHold, CreatedAt: businessDayTimeUTC,
			},
			qcdomain.InboundQCInspection{
				ID: "iqc-board-partial", WarehouseID: "wh-hcm-fg", Status: qcdomain.InboundQCInspectionStatusCompleted,
				Result: qcdomain.InboundQCResultPartial, CreatedAt: businessDayTimeUTC,
			},
		),
		inventoryapp.NewPrototypeWarehouseReceivingStore(),
		audit.NewInMemoryLogStore(),
	)
	listRejections := inventoryapp.NewListSupplierRejections(inventoryapp.NewPrototypeSupplierRejectionStore(
		inventorydomain.SupplierRejection{
			ID: "sr-board-draft", WarehouseID: "wh-hcm-fg", Status: inventorydomain.SupplierRejectionStatusDraft,
			CreatedAt: businessDayTimeUTC,
		},
		inventorydomain.SupplierRejection{
			ID: "sr-board-submitted", WarehouseID: "wh-hcm-fg", Status: inventorydomain.SupplierRejectionStatusSubmitted,
			CreatedAt: businessDayTimeUTC,
		},
		inventorydomain.SupplierRejection{
			ID: "sr-board-cancelled", WarehouseID: "wh-hcm-fg", Status: inventorydomain.SupplierRejectionStatusCancelled,
			CreatedAt: businessDayTimeUTC,
		},
	))
	handler := warehouseDailyBoardInboundMetricsHandler(
		purchaseService,
		receivingService,
		inboundQCService,
		listRejections,
	)
	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/warehouse/daily-board/inbound-metrics?warehouse_id=wh-hcm-fg&date=2026-04-27&shift_code=day",
		nil,
	)
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "warehouse@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead)))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var payload response.SuccessEnvelope[warehouseInboundMetricsResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	ctx := context.Background()
	orders, err := purchaseService.ListPurchaseOrders(ctx, purchaseapp.PurchaseOrderFilter{
		WarehouseID:  "wh-hcm-fg",
		ExpectedFrom: "2026-04-27",
		ExpectedTo:   "2026-04-27",
	})
	if err != nil {
		t.Fatalf("list purchase orders: %v", err)
	}
	receipts, err := receivingService.ListWarehouseReceivings(ctx, inventorydomain.NewWarehouseReceivingFilter("wh-hcm-fg", ""))
	if err != nil {
		t.Fatalf("list receivings: %v", err)
	}
	inspections, err := inboundQCService.ListInboundQCInspections(ctx, qcapp.NewInboundQCInspectionFilter("", "", "", "wh-hcm-fg"))
	if err != nil {
		t.Fatalf("list inbound qc inspections: %v", err)
	}
	rejections, err := listRejections.Execute(ctx, inventorydomain.NewSupplierRejectionFilter("", "wh-hcm-fg", ""))
	if err != nil {
		t.Fatalf("list supplier rejections: %v", err)
	}
	expected := expectedWarehouseInboundMetricsFromSources(orders, receipts, inspections, rejections, "wh-hcm-fg", "2026-04-27")

	if payload.Data.PurchaseOrdersIncoming != expected.purchaseOrdersIncoming ||
		payload.Data.ReceivingPending != expected.receivingPending ||
		payload.Data.ReceivingDraft != expected.receivingDraft ||
		payload.Data.ReceivingSubmitted != expected.receivingSubmitted ||
		payload.Data.ReceivingInspectReady != expected.receivingInspectReady ||
		payload.Data.QCHold != expected.qcHold ||
		payload.Data.QCFail != expected.qcFail ||
		payload.Data.QCPass != expected.qcPass ||
		payload.Data.QCPartial != expected.qcPartial ||
		payload.Data.SupplierRejections != expected.supplierRejections ||
		payload.Data.SupplierRejectionDraft != expected.supplierRejectionDraft ||
		payload.Data.SupplierRejectionSubmitted != expected.supplierRejectionSubmitted ||
		payload.Data.SupplierRejectionConfirmed != expected.supplierRejectionConfirmed ||
		payload.Data.SupplierRejectionCancelled != expected.supplierRejectionCancelled {
		t.Fatalf("metrics = %+v, want source-derived counts = %+v", payload.Data, expected)
	}
}

func TestWarehouseDailyBoardSubcontractMetricsHandlerSummarizesSubcontractState(t *testing.T) {
	businessDay := "2026-04-29"
	businessDayTimeUTC := time.Date(2026, 4, 29, 3, 0, 0, 0, time.UTC)
	previousDayTimeUTC := time.Date(2026, 4, 28, 3, 0, 0, 0, time.UTC)
	authConfig := auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}
	handler := warehouseDailyBoardSubcontractMetricsHandler(
		testWarehouseDailyBoardSubcontractOrderLister{orders: []productiondomain.SubcontractOrder{
			{ID: "sco-open", Status: productiondomain.SubcontractOrderStatusApproved, ExpectedReceiptDate: businessDay},
			{
				ID:                  "sco-material",
				Status:              productiondomain.SubcontractOrderStatusMaterialsIssued,
				ExpectedReceiptDate: businessDay,
				MaterialsIssuedAt:   businessDayTimeUTC,
			},
			{
				ID:                  "sco-sample",
				Status:              productiondomain.SubcontractOrderStatusSampleSubmitted,
				ExpectedReceiptDate: businessDay,
				SampleRequired:      true,
				SampleSubmittedAt:   businessDayTimeUTC,
			},
			{
				ID:                     "sco-claim",
				Status:                 productiondomain.SubcontractOrderStatusRejectedFactoryIssue,
				ExpectedReceiptDate:    "2026-04-28",
				RejectedFactoryIssueAt: previousDayTimeUTC,
			},
			{
				ID:                  "sco-payment",
				Status:              productiondomain.SubcontractOrderStatusFinalPaymentReady,
				ExpectedReceiptDate: businessDay,
				FinalPaymentReadyAt: businessDayTimeUTC,
			},
			{
				ID:                  "sco-closed",
				Status:              productiondomain.SubcontractOrderStatusClosed,
				ExpectedReceiptDate: businessDay,
				ClosedAt:            businessDayTimeUTC,
			},
		}},
		testWarehouseDailyBoardSubcontractMaterialTransferLister{transfers: map[string][]productiondomain.SubcontractMaterialTransfer{
			"sco-material": {
				{ID: "smt-match", SubcontractOrderID: "sco-material", SourceWarehouseID: "wh-hcm", HandoverAt: businessDayTimeUTC},
				{ID: "smt-other-warehouse", SubcontractOrderID: "sco-material", SourceWarehouseID: "wh-hn", HandoverAt: businessDayTimeUTC},
				{ID: "smt-other-day", SubcontractOrderID: "sco-material", SourceWarehouseID: "wh-hcm", HandoverAt: previousDayTimeUTC},
			},
		}},
		testWarehouseDailyBoardSubcontractFactoryClaimLister{claims: map[string][]productiondomain.SubcontractFactoryClaim{
			"sco-claim": {
				{
					ID:                 "sfc-open",
					SubcontractOrderID: "sco-claim",
					Status:             productiondomain.SubcontractFactoryClaimStatusOpen,
					OpenedAt:           previousDayTimeUTC,
					DueAt:              previousDayTimeUTC,
				},
				{
					ID:                 "sfc-ack",
					SubcontractOrderID: "sco-claim",
					Status:             productiondomain.SubcontractFactoryClaimStatusAcknowledged,
					OpenedAt:           businessDayTimeUTC,
					DueAt:              previousDayTimeUTC,
				},
				{
					ID:                 "sfc-resolved",
					SubcontractOrderID: "sco-claim",
					Status:             productiondomain.SubcontractFactoryClaimStatusResolved,
					OpenedAt:           businessDayTimeUTC,
					DueAt:              previousDayTimeUTC,
				},
			},
		}},
		testWarehouseDailyBoardSubcontractPaymentMilestoneLister{milestones: map[string][]productiondomain.SubcontractPaymentMilestone{
			"sco-payment": {
				{
					ID:                 "spm-final-ready",
					SubcontractOrderID: "sco-payment",
					Kind:               productiondomain.SubcontractPaymentMilestoneKindFinalPayment,
					Status:             productiondomain.SubcontractPaymentMilestoneStatusReady,
					ReadyAt:            businessDayTimeUTC,
				},
				{
					ID:                 "spm-deposit",
					SubcontractOrderID: "sco-payment",
					Kind:               productiondomain.SubcontractPaymentMilestoneKindDeposit,
					Status:             productiondomain.SubcontractPaymentMilestoneStatusRecorded,
					RecordedAt:         businessDayTimeUTC,
				},
			},
		}},
	)
	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/warehouse/daily-board/subcontract-metrics?warehouse_id=wh-hcm&date=2026-04-29&shift_code=day",
		nil,
	)
	req.Header.Set(response.HeaderRequestID, "req-warehouse-subcontract-metrics")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(authConfig, auth.RoleWarehouseLead)))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var payload response.SuccessEnvelope[warehouseSubcontractMetricsResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.OpenOrders != 5 ||
		payload.Data.MaterialIssuedOrders != 1 ||
		payload.Data.MaterialTransferCount != 1 ||
		payload.Data.SamplePending != 1 ||
		payload.Data.FactoryClaims != 2 ||
		payload.Data.FactoryClaimsOverdue != 2 ||
		payload.Data.FinalPaymentReadyOrders != 1 {
		t.Fatalf("metrics = %+v, want subcontract source counts", payload.Data)
	}
	if payload.Data.WarehouseID != "wh-hcm" || payload.Data.Date != businessDay || payload.Data.ShiftCode != "day" {
		t.Fatalf("scope = %+v, want wh-hcm/%s/day", payload.Data, businessDay)
	}
	if payload.Data.GeneratedAt == "" {
		t.Fatal("generated_at is empty")
	}
}

func TestWarehouseDailyBoardSubcontractMetricsHandlerRequiresWarehousePermission(t *testing.T) {
	handler := warehouseDailyBoardSubcontractMetricsHandler(
		testWarehouseDailyBoardSubcontractOrderLister{},
		testWarehouseDailyBoardSubcontractMaterialTransferLister{},
		testWarehouseDailyBoardSubcontractFactoryClaimLister{},
		testWarehouseDailyBoardSubcontractPaymentMilestoneLister{},
	)
	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/warehouse/daily-board/subcontract-metrics?warehouse_id=wh-hcm&date=2026-04-29",
		nil,
	)
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "sales@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleSalesOps)))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

type expectedFulfillmentMetrics struct {
	total           int
	newOrders       int
	reserved        int
	picking         int
	packed          int
	waitingHandover int
	missing         int
	handover        int
}

func expectedFulfillmentMetricsFromSources(
	orders []salesdomain.SalesOrder,
	manifests []shippingdomain.CarrierManifest,
	carrierCode string,
) expectedFulfillmentMetrics {
	manifestOrderNos := make(map[string]struct{})
	missingOrderNos := make(map[string]struct{})
	for _, manifest := range manifests {
		if strings.TrimSpace(carrierCode) != "" && manifest.CarrierCode != carrierCode {
			continue
		}
		for _, line := range manifest.Lines {
			orderNo := strings.TrimSpace(line.OrderNo)
			if orderNo == "" {
				continue
			}
			manifestOrderNos[orderNo] = struct{}{}
			if !line.Scanned {
				missingOrderNos[orderNo] = struct{}{}
			}
		}
	}

	expected := expectedFulfillmentMetrics{missing: len(missingOrderNos)}
	for _, order := range orders {
		if strings.TrimSpace(carrierCode) != "" {
			if _, ok := manifestOrderNos[order.OrderNo]; !ok {
				continue
			}
		}

		expected.total++
		switch salesdomain.NormalizeSalesOrderStatus(order.Status) {
		case salesdomain.SalesOrderStatusDraft, salesdomain.SalesOrderStatusConfirmed:
			expected.newOrders++
		case salesdomain.SalesOrderStatusReserved:
			expected.reserved++
		case salesdomain.SalesOrderStatusPicking, salesdomain.SalesOrderStatusPicked, salesdomain.SalesOrderStatusPacking:
			expected.picking++
		case salesdomain.SalesOrderStatusPacked:
			expected.packed++
		case salesdomain.SalesOrderStatusWaitingHandover:
			expected.waitingHandover++
		case salesdomain.SalesOrderStatusHandedOver:
			expected.handover++
		case salesdomain.SalesOrderStatusHandoverException:
			if orderNo := strings.TrimSpace(order.OrderNo); orderNo != "" {
				missingOrderNos[orderNo] = struct{}{}
			}
		}
	}
	expected.missing = len(missingOrderNos)

	return expected
}

type testWarehouseDailyBoardPurchaseOrderLister struct {
	orders []purchasedomain.PurchaseOrder
}

func (l testWarehouseDailyBoardPurchaseOrderLister) ListPurchaseOrders(
	_ context.Context,
	_ purchaseapp.PurchaseOrderFilter,
) ([]purchasedomain.PurchaseOrder, error) {
	return append([]purchasedomain.PurchaseOrder(nil), l.orders...), nil
}

type testWarehouseDailyBoardReceivingLister struct {
	receipts []inventorydomain.WarehouseReceiving
}

func (l testWarehouseDailyBoardReceivingLister) ListWarehouseReceivings(
	_ context.Context,
	_ inventorydomain.WarehouseReceivingFilter,
) ([]inventorydomain.WarehouseReceiving, error) {
	return append([]inventorydomain.WarehouseReceiving(nil), l.receipts...), nil
}

type testWarehouseDailyBoardInboundQCLister struct {
	inspections []qcdomain.InboundQCInspection
}

func (l testWarehouseDailyBoardInboundQCLister) ListInboundQCInspections(
	_ context.Context,
	_ qcapp.InboundQCInspectionFilter,
) ([]qcdomain.InboundQCInspection, error) {
	return append([]qcdomain.InboundQCInspection(nil), l.inspections...), nil
}

type testWarehouseDailyBoardSupplierRejectionLister struct {
	rejections []inventorydomain.SupplierRejection
}

func (l testWarehouseDailyBoardSupplierRejectionLister) Execute(
	_ context.Context,
	_ inventorydomain.SupplierRejectionFilter,
) ([]inventorydomain.SupplierRejection, error) {
	return append([]inventorydomain.SupplierRejection(nil), l.rejections...), nil
}

type testWarehouseDailyBoardSubcontractOrderLister struct {
	orders []productiondomain.SubcontractOrder
}

func (l testWarehouseDailyBoardSubcontractOrderLister) ListSubcontractOrders(
	_ context.Context,
	_ productionapp.SubcontractOrderFilter,
) ([]productiondomain.SubcontractOrder, error) {
	return append([]productiondomain.SubcontractOrder(nil), l.orders...), nil
}

type testWarehouseDailyBoardSubcontractMaterialTransferLister struct {
	transfers map[string][]productiondomain.SubcontractMaterialTransfer
}

func (l testWarehouseDailyBoardSubcontractMaterialTransferLister) ListBySubcontractOrder(
	_ context.Context,
	subcontractOrderID string,
) ([]productiondomain.SubcontractMaterialTransfer, error) {
	return append([]productiondomain.SubcontractMaterialTransfer(nil), l.transfers[subcontractOrderID]...), nil
}

type testWarehouseDailyBoardSubcontractFactoryClaimLister struct {
	claims map[string][]productiondomain.SubcontractFactoryClaim
}

func (l testWarehouseDailyBoardSubcontractFactoryClaimLister) ListBySubcontractOrder(
	_ context.Context,
	subcontractOrderID string,
) ([]productiondomain.SubcontractFactoryClaim, error) {
	return append([]productiondomain.SubcontractFactoryClaim(nil), l.claims[subcontractOrderID]...), nil
}

type testWarehouseDailyBoardSubcontractPaymentMilestoneLister struct {
	milestones map[string][]productiondomain.SubcontractPaymentMilestone
}

func (l testWarehouseDailyBoardSubcontractPaymentMilestoneLister) ListBySubcontractOrder(
	_ context.Context,
	subcontractOrderID string,
) ([]productiondomain.SubcontractPaymentMilestone, error) {
	return append([]productiondomain.SubcontractPaymentMilestone(nil), l.milestones[subcontractOrderID]...), nil
}

type expectedWarehouseInboundMetrics struct {
	purchaseOrdersIncoming     int
	receivingPending           int
	receivingDraft             int
	receivingSubmitted         int
	receivingInspectReady      int
	qcHold                     int
	qcFail                     int
	qcPass                     int
	qcPartial                  int
	supplierRejections         int
	supplierRejectionDraft     int
	supplierRejectionSubmitted int
	supplierRejectionConfirmed int
	supplierRejectionCancelled int
}

func expectedWarehouseInboundMetricsFromSources(
	orders []purchasedomain.PurchaseOrder,
	receipts []inventorydomain.WarehouseReceiving,
	inspections []qcdomain.InboundQCInspection,
	rejections []inventorydomain.SupplierRejection,
	warehouseID string,
	date string,
) expectedWarehouseInboundMetrics {
	expected := expectedWarehouseInboundMetrics{}
	for _, order := range orders {
		if order.WarehouseID != warehouseID || order.ExpectedDate != date {
			continue
		}
		switch purchasedomain.NormalizePurchaseOrderStatus(order.Status) {
		case purchasedomain.PurchaseOrderStatusApproved, purchasedomain.PurchaseOrderStatusPartiallyReceived:
			expected.purchaseOrdersIncoming++
		}
	}
	for _, receipt := range receipts {
		if receipt.WarehouseID != warehouseID || businessDate(receipt.CreatedAt) != date {
			continue
		}
		switch inventorydomain.NormalizeWarehouseReceivingStatus(receipt.Status) {
		case inventorydomain.WarehouseReceivingStatusDraft:
			expected.receivingDraft++
			expected.receivingPending++
		case inventorydomain.WarehouseReceivingStatusSubmitted:
			expected.receivingSubmitted++
			expected.receivingPending++
		case inventorydomain.WarehouseReceivingStatusInspectReady:
			expected.receivingInspectReady++
			expected.receivingPending++
		}
	}
	for _, inspection := range inspections {
		if inspection.WarehouseID != warehouseID || businessDate(inspection.CreatedAt) != date {
			continue
		}
		if qcdomain.NormalizeInboundQCInspectionStatus(inspection.Status) != qcdomain.InboundQCInspectionStatusCompleted {
			continue
		}
		switch qcdomain.NormalizeInboundQCResult(inspection.Result) {
		case qcdomain.InboundQCResultHold:
			expected.qcHold++
		case qcdomain.InboundQCResultFail:
			expected.qcFail++
		case qcdomain.InboundQCResultPass:
			expected.qcPass++
		case qcdomain.InboundQCResultPartial:
			expected.qcPartial++
		}
	}
	for _, rejection := range rejections {
		if rejection.WarehouseID != warehouseID || businessDate(rejection.CreatedAt) != date {
			continue
		}
		switch inventorydomain.NormalizeSupplierRejectionStatus(rejection.Status) {
		case inventorydomain.SupplierRejectionStatusDraft:
			expected.supplierRejectionDraft++
			expected.supplierRejections++
		case inventorydomain.SupplierRejectionStatusSubmitted:
			expected.supplierRejectionSubmitted++
			expected.supplierRejections++
		case inventorydomain.SupplierRejectionStatusConfirmed:
			expected.supplierRejectionConfirmed++
			expected.supplierRejections++
		case inventorydomain.SupplierRejectionStatusCancelled:
			expected.supplierRejectionCancelled++
		}
	}

	return expected
}

func newTestWarehouseDailyBoardPurchaseOrderService(t *testing.T) purchaseapp.PurchaseOrderService {
	t.Helper()

	auditStore := audit.NewInMemoryLogStore()
	service := purchaseapp.NewPurchaseOrderService(
		purchaseapp.NewPrototypePurchaseOrderStore(auditStore),
		masterdataapp.NewPrototypePartyCatalog(auditStore),
		masterdataapp.NewPrototypeItemCatalog(auditStore),
		masterdataapp.NewPrototypeWarehouseLocationCatalog(auditStore),
		purchaseOrderUOMConverterAdapter{catalog: masterdataapp.NewPrototypeUOMCatalog()},
	)
	createApprovedWarehouseDailyBoardPurchaseOrder(t, service, "po-board-approved", "2026-04-27")
	createApprovedWarehouseDailyBoardPurchaseOrder(t, service, "po-board-other-day", "2026-04-28")

	return service
}

func createApprovedWarehouseDailyBoardPurchaseOrder(
	t *testing.T,
	service purchaseapp.PurchaseOrderService,
	id string,
	expectedDate string,
) {
	t.Helper()

	if _, err := service.CreatePurchaseOrder(context.Background(), purchaseapp.CreatePurchaseOrderInput{
		ID:           id,
		OrgID:        "org-my-pham",
		PONo:         strings.ToUpper(id),
		SupplierID:   "sup-rm-bioactive",
		WarehouseID:  "wh-hcm-fg",
		ExpectedDate: expectedDate,
		CurrencyCode: "VND",
		Lines: []purchaseapp.PurchaseOrderLineInput{
			{
				ID:         id + "-line-1",
				LineNo:     1,
				ItemID:     "item-cream-50g",
				OrderedQty: "12.000000",
				UOMCode:    "EA",
				UnitPrice:  "1.0000",
			},
		},
		ActorID:   "user-purchase-ops",
		RequestID: "req-" + id + "-create",
	}); err != nil {
		t.Fatalf("create purchase order: %v", err)
	}
	if _, err := service.SubmitPurchaseOrder(context.Background(), purchaseapp.PurchaseOrderActionInput{
		ID:        id,
		ActorID:   "user-purchase-ops",
		RequestID: "req-" + id + "-submit",
	}); err != nil {
		t.Fatalf("submit purchase order: %v", err)
	}
	if _, err := service.ApprovePurchaseOrder(context.Background(), purchaseapp.PurchaseOrderActionInput{
		ID:        id,
		ActorID:   "user-purchase-lead",
		RequestID: "req-" + id + "-approve",
	}); err != nil {
		t.Fatalf("approve purchase order: %v", err)
	}
}

func mustWarehouseDailyBoardPurchaseOrder(
	t *testing.T,
	id string,
	status purchasedomain.PurchaseOrderStatus,
	warehouseID string,
	expectedDate string,
) purchasedomain.PurchaseOrder {
	t.Helper()

	order := approvedTestGoodsReceiptPurchaseOrder()
	order.ID = id
	order.PONo = strings.ToUpper(id)
	order.WarehouseID = warehouseID
	order.ExpectedDate = expectedDate

	switch status {
	case purchasedomain.PurchaseOrderStatusApproved:
		return order
	case purchasedomain.PurchaseOrderStatusPartiallyReceived:
		partial, err := order.MarkPartiallyReceived("user-warehouse-lead", time.Date(2026, 4, 29, 9, 0, 0, 0, time.UTC))
		if err != nil {
			t.Fatalf("mark purchase order partially received: %v", err)
		}
		return partial
	case purchasedomain.PurchaseOrderStatusReceived:
		partial, err := order.MarkPartiallyReceived("user-warehouse-lead", time.Date(2026, 4, 29, 9, 0, 0, 0, time.UTC))
		if err != nil {
			t.Fatalf("mark purchase order partially received: %v", err)
		}
		received, err := partial.MarkReceived("user-warehouse-lead", time.Date(2026, 4, 29, 10, 0, 0, 0, time.UTC))
		if err != nil {
			t.Fatalf("mark purchase order received: %v", err)
		}
		return received
	default:
		order.Status = status
		return order
	}
}

func mustWarehouseDailyBoardReceiving(
	t *testing.T,
	id string,
	status inventorydomain.WarehouseReceivingStatus,
	warehouseID string,
	createdAt time.Time,
) inventorydomain.WarehouseReceiving {
	t.Helper()

	receipt, err := inventorydomain.NewWarehouseReceiving(inventorydomain.NewWarehouseReceivingInput{
		ID:               id,
		OrgID:            "org-my-pham",
		ReceiptNo:        strings.ToUpper(id),
		WarehouseID:      warehouseID,
		WarehouseCode:    strings.ToUpper(warehouseID),
		LocationID:       "loc-hcm-fg-recv-01",
		LocationCode:     "FG-RECV-01",
		ReferenceDocType: "purchase_order",
		ReferenceDocID:   "po-approved",
		SupplierID:       "supplier-local",
		DeliveryNoteNo:   "DN-" + strings.ToUpper(id),
		Lines: []inventorydomain.NewWarehouseReceivingLineInput{
			{
				ID:                  id + "-line-1",
				PurchaseOrderLineID: "po-line-260427-api-001",
				ItemID:              "item-cream-50g",
				SKU:                 "CREAM-50G",
				ItemName:            "Moisturizing Cream",
				BatchID:             "batch-cream-2603b",
				BatchNo:             "LOT-2603B",
				LotNo:               "LOT-2603B",
				ExpiryDate:          time.Date(2028, 3, 1, 0, 0, 0, 0, time.UTC),
				Quantity:            decimal.MustQuantity("12"),
				UOMCode:             "EA",
				BaseUOMCode:         "EA",
				PackagingStatus:     inventorydomain.ReceivingPackagingStatusIntact,
				QCStatus:            inventorydomain.QCStatusPass,
			},
		},
		CreatedBy: "user-warehouse-lead",
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	})
	if err != nil {
		t.Fatalf("new warehouse receiving: %v", err)
	}

	switch status {
	case inventorydomain.WarehouseReceivingStatusDraft:
		return receipt
	case inventorydomain.WarehouseReceivingStatusSubmitted:
		submitted, err := receipt.Submit("user-warehouse-lead", createdAt.Add(10*time.Minute))
		if err != nil {
			t.Fatalf("submit receiving: %v", err)
		}
		return submitted
	case inventorydomain.WarehouseReceivingStatusInspectReady:
		submitted, err := receipt.Submit("user-warehouse-lead", createdAt.Add(10*time.Minute))
		if err != nil {
			t.Fatalf("submit receiving: %v", err)
		}
		ready, err := submitted.MarkInspectReady("user-qa", createdAt.Add(20*time.Minute))
		if err != nil {
			t.Fatalf("mark receiving inspect ready: %v", err)
		}
		return ready
	case inventorydomain.WarehouseReceivingStatusPosted:
		submitted, err := receipt.Submit("user-warehouse-lead", createdAt.Add(10*time.Minute))
		if err != nil {
			t.Fatalf("submit receiving: %v", err)
		}
		ready, err := submitted.MarkInspectReady("user-qa", createdAt.Add(20*time.Minute))
		if err != nil {
			t.Fatalf("mark receiving inspect ready: %v", err)
		}
		posted, err := ready.Post("user-warehouse-lead", createdAt.Add(30*time.Minute))
		if err != nil {
			t.Fatalf("post receiving: %v", err)
		}
		return posted
	default:
		receipt.Status = status
		return receipt
	}
}

func TestConfirmCarrierManifestHandoverHandlerMarksManifestHandedOver(t *testing.T) {
	store := shippingapp.NewPrototypeCarrierManifestStore()
	auditStore := audit.NewInMemoryLogStore()
	handover := &recordingCarrierManifestSalesOrderHandover{}
	if _, err := shippingapp.NewVerifyCarrierManifestScan(store, auditStore).Execute(context.Background(), shippingapp.VerifyCarrierManifestScanInput{
		ManifestID: "manifest-hcm-ghn-morning",
		Code:       "GHN260426003",
		ActorID:    "user-handover-operator",
		RequestID:  "req-confirm-scan",
	}); err != nil {
		t.Fatalf("scan missing line: %v", err)
	}
	service := shippingapp.NewConfirmCarrierManifestHandover(store, auditStore, handover)
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/shipping/manifests/manifest-hcm-ghn-morning/confirm-handover",
		nil,
	)
	req.SetPathValue("manifest_id", "manifest-hcm-ghn-morning")
	req.Header.Set(response.HeaderRequestID, "req-manifest-confirm-handover")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead)))
	rec := httptest.NewRecorder()

	confirmCarrierManifestHandoverHandler(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var payload response.SuccessEnvelope[carrierManifestResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.Status != "handed_over" || payload.Data.AuditLogID == "" {
		t.Fatalf("payload = %+v, want handed_over manifest with audit", payload.Data)
	}
	if len(handover.orderNos) != 3 || handover.orderNos[0] != "SO-260426-001" {
		t.Fatalf("handover calls = %+v, want all manifest orders", handover.orderNos)
	}

	logs, err := auditStore.List(req.Context(), audit.Query{Action: "shipping.manifest.handed_over"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 || logs[0].Metadata["handed_over_order_count"] != 3 {
		t.Fatalf("audit logs = %+v, want handed over count", logs)
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

func TestReturnMasterDataHandlerListsReasonConditionAndDispositionSeed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/return-reasons", nil)
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseStaff)))
	rec := httptest.NewRecorder()

	returnMasterDataHandler(returnsapp.NewListReturnMasterData()).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var payload response.SuccessEnvelope[returnMasterDataResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Data.Reasons) == 0 {
		t.Fatal("return reasons are empty")
	}
	if len(payload.Data.Conditions) != 5 {
		t.Fatalf("conditions = %d, want 5", len(payload.Data.Conditions))
	}
	if len(payload.Data.Dispositions) != 3 {
		t.Fatalf("dispositions = %d, want 3", len(payload.Data.Dispositions))
	}
	if payload.Data.Conditions[0].Code != "sealed_good" {
		t.Fatalf("first condition = %q, want sealed_good", payload.Data.Conditions[0].Code)
	}
}

func TestReturnScanHandlerCreatesPendingInspectionWithDefaultDisposition(t *testing.T) {
	store := returnsapp.NewPrototypeReturnReceiptStore()
	receiveService := returnsapp.NewReceiveReturn(store, audit.NewInMemoryLogStore())
	body := bytes.NewBufferString(`{
		"warehouse_id": "wh-hcm",
		"warehouse_code": "HCM",
		"source": "CARRIER",
		"code": "GHN260426001",
		"package_condition": "sealed"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/returns/scan", body)
	req.Header.Set(response.HeaderRequestID, "req-return-scan")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead)))
	rec := httptest.NewRecorder()

	returnScanHandler(receiveService).ServeHTTP(rec, req)

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
	if payload.Data.Disposition != "needs_inspection" || payload.Data.Status != "pending_inspection" {
		t.Fatalf("payload = %+v, want pending inspection with needs_inspection", payload.Data)
	}
}

func TestReturnScanHandlerCreatesDeliveredOrderReceipt(t *testing.T) {
	store := returnsapp.NewPrototypeReturnReceiptStore()
	receiveService := returnsapp.NewReceiveReturn(store, audit.NewInMemoryLogStore())
	body := bytes.NewBufferString(`{
		"warehouse_id": "wh-hcm",
		"code": "SO-260426-004"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/returns/scan", body)
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead)))
	rec := httptest.NewRecorder()

	returnScanHandler(receiveService).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var payload response.SuccessEnvelope[returnReceiptResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.OriginalOrderNo != "SO-260426-004" {
		t.Fatalf("order no = %q, want SO-260426-004", payload.Data.OriginalOrderNo)
	}
	if payload.Data.TrackingNo != "GHN260426004" || payload.Data.UnknownCase {
		t.Fatalf("payload = %+v, want known delivered order linked to tracking GHN260426004", payload.Data)
	}
}

func TestReturnScanHandlerCreatesUnknownCaseForUnmatchedCode(t *testing.T) {
	store := returnsapp.NewPrototypeReturnReceiptStore()
	receiveService := returnsapp.NewReceiveReturn(store, audit.NewInMemoryLogStore())
	body := bytes.NewBufferString(`{
		"warehouse_id": "wh-hcm",
		"source": "SHIPPER",
		"code": "UNKNOWN-RETURN-SCAN",
		"package_condition": "dented box"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/returns/scan", body)
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead)))
	rec := httptest.NewRecorder()

	returnScanHandler(receiveService).ServeHTTP(rec, req)

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
	if payload.Data.TrackingNo != "UNKNOWN-RETURN-SCAN" || payload.Data.TargetLocation != "return-inspection-queue" {
		t.Fatalf("payload = %+v, want unknown scan routed to inspection queue", payload.Data)
	}
}

func TestReturnScanHandlerRejectsBlankScanCode(t *testing.T) {
	store := returnsapp.NewPrototypeReturnReceiptStore()
	receiveService := returnsapp.NewReceiveReturn(store, audit.NewInMemoryLogStore())
	body := bytes.NewBufferString(`{
		"warehouse_id": "wh-hcm",
		"code": " "
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/returns/scan", body)
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead)))
	rec := httptest.NewRecorder()

	returnScanHandler(receiveService).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestReturnScanHandlerRejectsOrderBeforeHandoverOrDelivery(t *testing.T) {
	store := returnsapp.NewPrototypeReturnReceiptStore()
	receiveService := returnsapp.NewReceiveReturn(store, audit.NewInMemoryLogStore())
	body := bytes.NewBufferString(`{
		"warehouse_id": "wh-hcm",
		"code": "GHN260426009"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/returns/scan", body)
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead)))
	rec := httptest.NewRecorder()

	returnScanHandler(receiveService).ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
}

func TestReturnScanHandlerRejectsDuplicateReturn(t *testing.T) {
	store := returnsapp.NewPrototypeReturnReceiptStore()
	receiveService := returnsapp.NewReceiveReturn(store, audit.NewInMemoryLogStore())
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/returns/scan",
		bytes.NewBufferString(`{"warehouse_id":"wh-hcm","code":"GHN260426001"}`),
	)
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead)))
	rec := httptest.NewRecorder()

	returnScanHandler(receiveService).ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("first status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	duplicate := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/returns/scan",
		bytes.NewBufferString(`{"warehouse_id":"wh-hcm","code":"SO-260426-001"}`),
	)
	duplicate = duplicate.WithContext(req.Context())
	duplicateRec := httptest.NewRecorder()
	returnScanHandler(receiveService).ServeHTTP(duplicateRec, duplicate)
	if duplicateRec.Code != http.StatusConflict {
		t.Fatalf("duplicate status = %d, want %d: %s", duplicateRec.Code, http.StatusConflict, duplicateRec.Body.String())
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

func TestReturnInspectionHandlerRecordsReusableInspection(t *testing.T) {
	store := returnsapp.NewPrototypeReturnReceiptStore()
	auditStore := audit.NewInMemoryLogStore()
	inspectService := returnsapp.NewInspectReturn(store, auditStore)
	body := bytes.NewBufferString(`{
		"condition": "intact",
		"disposition": "reusable",
		"note": "seal and box intact",
		"evidence_label": "photo-001"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/returns/rr-260426-0001/inspect", body)
	req.SetPathValue("return_receipt_id", "rr-260426-0001")
	req.Header.Set(response.HeaderRequestID, "req-return-inspect")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead)))
	rec := httptest.NewRecorder()

	returnInspectionHandler(inspectService).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var payload response.SuccessEnvelope[returnInspectionResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.Status != "inspection_recorded" || payload.Data.TargetLocation != "return-area-qc-release" {
		t.Fatalf("payload = %+v, want reusable inspection recorded", payload.Data)
	}
	if payload.Data.AuditLogID == "" {
		t.Fatal("audit log id is empty")
	}

	rows, err := returnsapp.NewListReturnReceipts(store).Execute(
		req.Context(),
		returnsdomain.NewReturnReceiptFilter("wh-hcm", returnsdomain.ReturnStatusInspected),
	)
	if err != nil {
		t.Fatalf("list return receipts: %v", err)
	}
	if len(rows) != 1 || rows[0].TargetLocation != "return-area-qc-release" {
		t.Fatalf("rows = %+v, want inspected receipt routed to qc release", rows)
	}
}

func TestReturnInspectionHandlerRoutesNeedQAInspection(t *testing.T) {
	store := returnsapp.NewPrototypeReturnReceiptStore()
	inspectService := returnsapp.NewInspectReturn(store, audit.NewInMemoryLogStore())
	body := bytes.NewBufferString(`{
		"condition": "missing_accessory",
		"disposition": "needs_inspection"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/returns/rr-260426-0001/inspect", body)
	req.SetPathValue("return_receipt_id", "rr-260426-0001")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead)))
	rec := httptest.NewRecorder()

	returnInspectionHandler(inspectService).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var payload response.SuccessEnvelope[returnInspectionResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.Status != "return_qa_hold" || payload.Data.RiskLevel != "high" {
		t.Fatalf("payload = %+v, want high risk QA hold", payload.Data)
	}
	if payload.Data.TargetLocation != "return-qa-hold" {
		t.Fatalf("target location = %q, want return-qa-hold", payload.Data.TargetLocation)
	}
}

func TestReturnInspectionHandlerRejectsInvalidConditionAndMissingReceipt(t *testing.T) {
	store := returnsapp.NewPrototypeReturnReceiptStore()
	inspectService := returnsapp.NewInspectReturn(store, audit.NewInMemoryLogStore())
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/returns/rr-260426-0001/inspect",
		bytes.NewBufferString(`{"condition":"sealed_good","disposition":"reusable"}`),
	)
	req.SetPathValue("return_receipt_id", "rr-260426-0001")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead)))
	rec := httptest.NewRecorder()

	returnInspectionHandler(inspectService).ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid condition status = %d, want %d: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}

	missing := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/returns/missing/inspect",
		bytes.NewBufferString(`{"condition":"intact","disposition":"reusable"}`),
	)
	missing.SetPathValue("return_receipt_id", "missing")
	missing = missing.WithContext(req.Context())
	missingRec := httptest.NewRecorder()

	returnInspectionHandler(inspectService).ServeHTTP(missingRec, missing)
	if missingRec.Code != http.StatusNotFound {
		t.Fatalf("missing receipt status = %d, want %d: %s", missingRec.Code, http.StatusNotFound, missingRec.Body.String())
	}
}

func TestReturnDispositionHandlerRoutesReusableAfterInspection(t *testing.T) {
	store := returnsapp.NewPrototypeReturnReceiptStore()
	auditStore := audit.NewInMemoryLogStore()
	inspectService := returnsapp.NewInspectReturn(store, auditStore)
	_, err := inspectService.Execute(context.Background(), returnsapp.InspectReturnInput{
		ReceiptID:   "rr-260426-0001",
		Condition:   "intact",
		Disposition: "reusable",
		ActorID:     "user-return-inspector",
		RequestID:   "req-return-inspect",
	})
	if err != nil {
		t.Fatalf("inspect return: %v", err)
	}

	movementStore := inventoryapp.NewInMemoryStockMovementStore()
	applyService := returnsapp.NewApplyReturnDisposition(store, movementStore, auditStore)
	body := bytes.NewBufferString(`{
		"disposition": "reusable",
		"note": "ready for putaway"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/returns/rr-260426-0001/disposition", body)
	req.SetPathValue("return_receipt_id", "rr-260426-0001")
	req.Header.Set(response.HeaderRequestID, "req-return-disposition")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead)))
	rec := httptest.NewRecorder()

	returnDispositionHandler(applyService).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var payload response.SuccessEnvelope[returnDispositionActionResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.ActionCode != "route_to_putaway" || payload.Data.TargetLocation != "return-putaway-ready" {
		t.Fatalf("payload = %+v, want putaway disposition", payload.Data)
	}
	if payload.Data.AuditLogID == "" {
		t.Fatal("audit log id is empty")
	}

	rows, err := returnsapp.NewListReturnReceipts(store).Execute(
		req.Context(),
		returnsdomain.NewReturnReceiptFilter("wh-hcm", returnsdomain.ReturnStatusDispositioned),
	)
	if err != nil {
		t.Fatalf("list return receipts: %v", err)
	}
	if len(rows) != 1 || rows[0].TargetLocation != "return-putaway-ready" || rows[0].StockMovement == nil {
		t.Fatalf("rows = %+v, want dispositioned receipt with stock movement", rows)
	}
	if rows[0].StockMovement.MovementType != "return_restock" || rows[0].StockMovement.TargetStockStatus != "available" {
		t.Fatalf("stock movement = %+v, want reusable available restock", rows[0].StockMovement)
	}
	if movementStore.Count() != 1 {
		t.Fatalf("stock movement count = %d, want 1", movementStore.Count())
	}
}

func TestReturnDispositionHandlerRoutesQAHold(t *testing.T) {
	store := returnsapp.NewPrototypeReturnReceiptStore()
	auditStore := audit.NewInMemoryLogStore()
	_, err := returnsapp.NewInspectReturn(store, auditStore).Execute(context.Background(), returnsapp.InspectReturnInput{
		ReceiptID:   "rr-260426-0001",
		Condition:   "seal_torn",
		Disposition: "needs_inspection",
		ActorID:     "user-return-inspector",
		RequestID:   "req-return-inspect",
	})
	if err != nil {
		t.Fatalf("inspect return: %v", err)
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/returns/rr-260426-0001/disposition",
		bytes.NewBufferString(`{"disposition":"needs_inspection"}`),
	)
	req.SetPathValue("return_receipt_id", "rr-260426-0001")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead)))
	rec := httptest.NewRecorder()

	movementStore := inventoryapp.NewInMemoryStockMovementStore()
	returnDispositionHandler(returnsapp.NewApplyReturnDisposition(store, movementStore, auditStore)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var payload response.SuccessEnvelope[returnDispositionActionResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.TargetLocation != "return-quarantine-hold" || payload.Data.TargetStockStatus != "qc_hold" {
		t.Fatalf("payload = %+v, want quarantine hold", payload.Data)
	}
	if movementStore.Count() != 1 {
		t.Fatalf("stock movement count = %d, want 1 for qa hold", movementStore.Count())
	}
	movements := movementStore.Movements()
	if movements[0].MovementType != inventorydomain.MovementReturnReceipt ||
		movements[0].StockStatus != inventorydomain.StockStatusQCHold {
		t.Fatalf("movement = %+v, want return receipt qc hold", movements[0])
	}
}

func TestReturnDispositionHandlerRejectsPendingReceiptAndInvalidDisposition(t *testing.T) {
	store := returnsapp.NewPrototypeReturnReceiptStore()
	auditStore := audit.NewInMemoryLogStore()
	applyService := returnsapp.NewApplyReturnDisposition(store, inventoryapp.NewInMemoryStockMovementStore(), auditStore)
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/returns/rr-260426-0001/disposition",
		bytes.NewBufferString(`{"disposition":"reusable"}`),
	)
	req.SetPathValue("return_receipt_id", "rr-260426-0001")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead)))
	rec := httptest.NewRecorder()

	returnDispositionHandler(applyService).ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("pending receipt status = %d, want %d: %s", rec.Code, http.StatusConflict, rec.Body.String())
	}

	_, err := returnsapp.NewInspectReturn(store, auditStore).Execute(context.Background(), returnsapp.InspectReturnInput{
		ReceiptID:   "rr-260426-0001",
		Condition:   "intact",
		Disposition: "reusable",
		ActorID:     "user-return-inspector",
		RequestID:   "req-return-inspect",
	})
	if err != nil {
		t.Fatalf("inspect return: %v", err)
	}
	invalid := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/returns/rr-260426-0001/disposition",
		bytes.NewBufferString(`{"disposition":"usable"}`),
	)
	invalid.SetPathValue("return_receipt_id", "rr-260426-0001")
	invalid = invalid.WithContext(req.Context())
	invalidRec := httptest.NewRecorder()

	returnDispositionHandler(applyService).ServeHTTP(invalidRec, invalid)
	if invalidRec.Code != http.StatusBadRequest {
		t.Fatalf("invalid disposition status = %d, want %d: %s", invalidRec.Code, http.StatusBadRequest, invalidRec.Body.String())
	}
}

func TestReturnAttachmentHandlerUploadsInspectionEvidence(t *testing.T) {
	store := returnsapp.NewPrototypeReturnReceiptStore()
	auditStore := audit.NewInMemoryLogStore()
	_, err := returnsapp.NewInspectReturn(store, auditStore).Execute(context.Background(), returnsapp.InspectReturnInput{
		ReceiptID:   "rr-260426-0001",
		Condition:   "intact",
		Disposition: "reusable",
		ActorID:     "user-return-inspector",
		RequestID:   "req-return-inspect",
	})
	if err != nil {
		t.Fatalf("inspect return: %v", err)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("inspection_id", "inspect-rr-260426-0001-intact"); err != nil {
		t.Fatalf("write inspection field: %v", err)
	}
	if err := writer.WriteField("note", "front photo before putaway"); err != nil {
		t.Fatalf("write note field: %v", err)
	}
	filePart, err := writer.CreateFormFile("file", "return-photo.png")
	if err != nil {
		t.Fatalf("create file part: %v", err)
	}
	if _, err := filePart.Write([]byte("fake image bytes")); err != nil {
		t.Fatalf("write file part: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/returns/rr-260426-0001/attachments", &body)
	req.SetPathValue("return_receipt_id", "rr-260426-0001")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set(response.HeaderRequestID, "req-return-attachment")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead)))
	rec := httptest.NewRecorder()
	objectStore := returnsapp.NewInMemoryReturnAttachmentObjectStore()

	returnAttachmentHandler(
		returnsapp.NewUploadReturnAttachment(store, auditStore).WithObjectStore(objectStore),
	).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var payload response.SuccessEnvelope[returnAttachmentResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.InspectionID != "inspect-rr-260426-0001-intact" ||
		payload.Data.MIMEType != "image/png" ||
		payload.Data.FileSizeBytes == 0 ||
		payload.Data.AuditLogID == "" {
		t.Fatalf("payload = %+v, want png attachment metadata and audit", payload.Data)
	}
	if _, ok := objectStore.Get(payload.Data.StorageBucket, payload.Data.StorageKey); !ok {
		t.Fatalf("object %s/%s was not stored", payload.Data.StorageBucket, payload.Data.StorageKey)
	}
	if payload.Data.StorageBucket == "" ||
		payload.Data.StorageKey != "returns/rr-260426-0001/inspections/inspect-rr-260426-0001-intact/return-photo.png" {
		t.Fatalf("storage metadata = %s/%s, want deterministic bucket and key", payload.Data.StorageBucket, payload.Data.StorageKey)
	}

	logs, err := auditStore.List(req.Context(), audit.Query{Action: "returns.inspection.attachment_uploaded"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit logs = %d, want 1", len(logs))
	}
	if logs[0].AfterData["storage_bucket"] != payload.Data.StorageBucket ||
		logs[0].AfterData["storage_key"] != payload.Data.StorageKey {
		t.Fatalf("audit after data = %+v, want storage metadata", logs[0].AfterData)
	}
	if _, ok := logs[0].AfterData["file_content"]; ok {
		t.Fatalf("audit after data = %+v, must not contain file content", logs[0].AfterData)
	}
}

func TestReturnAttachmentHandlerBlocksUnauthorizedUpload(t *testing.T) {
	store := returnsapp.NewPrototypeReturnReceiptStore()
	auditStore := audit.NewInMemoryLogStore()
	if _, err := returnsapp.NewInspectReturn(store, auditStore).Execute(context.Background(), returnsapp.InspectReturnInput{
		ReceiptID:   "rr-260426-0001",
		Condition:   "intact",
		Disposition: "reusable",
		ActorID:     "user-return-inspector",
		RequestID:   "req-return-inspect",
	}); err != nil {
		t.Fatalf("inspect return: %v", err)
	}
	objectStore := returnsapp.NewInMemoryReturnAttachmentObjectStore()
	handler := returnAttachmentHandler(
		returnsapp.NewUploadReturnAttachment(store, auditStore).WithObjectStore(objectStore),
	)

	unauthenticated := newReturnAttachmentRequest(t, "rr-260426-0001", "inspect-rr-260426-0001-intact", "return-photo.png", "fake image bytes")
	unauthenticated = unauthenticated.WithContext(context.Background())
	unauthenticatedRec := httptest.NewRecorder()
	handler.ServeHTTP(unauthenticatedRec, unauthenticated)
	if unauthenticatedRec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d, want %d: %s", unauthenticatedRec.Code, http.StatusUnauthorized, unauthenticatedRec.Body.String())
	}

	forbidden := newReturnAttachmentRequest(t, "rr-260426-0001", "inspect-rr-260426-0001-intact", "return-photo.png", "fake image bytes")
	forbidden = forbidden.WithContext(auth.WithPrincipal(context.Background(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "staff@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseStaff)))
	forbiddenRec := httptest.NewRecorder()
	handler.ServeHTTP(forbiddenRec, forbidden)
	if forbiddenRec.Code != http.StatusForbidden {
		t.Fatalf("forbidden status = %d, want %d: %s", forbiddenRec.Code, http.StatusForbidden, forbiddenRec.Body.String())
	}

	logs, err := auditStore.List(context.Background(), audit.Query{Action: "returns.inspection.attachment_uploaded"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 0 || objectStore.Len() != 0 {
		t.Fatalf("blocked upload wrote logs=%d objects=%d, want none", len(logs), objectStore.Len())
	}
}

func TestReturnAttachmentHandlerRejectsPendingReceiptAndInvalidFile(t *testing.T) {
	store := returnsapp.NewPrototypeReturnReceiptStore()
	auditStore := audit.NewInMemoryLogStore()
	req := newReturnAttachmentRequest(t, "rr-260426-0001", "inspect-rr-260426-0001-intact", "return-photo.png", "fake image bytes")
	rec := httptest.NewRecorder()
	objectStore := returnsapp.NewInMemoryReturnAttachmentObjectStore()

	returnAttachmentHandler(
		returnsapp.NewUploadReturnAttachment(store, auditStore).WithObjectStore(objectStore),
	).ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("pending status = %d, want %d: %s", rec.Code, http.StatusConflict, rec.Body.String())
	}

	_, err := returnsapp.NewInspectReturn(store, auditStore).Execute(context.Background(), returnsapp.InspectReturnInput{
		ReceiptID:   "rr-260426-0001",
		Condition:   "intact",
		Disposition: "reusable",
		ActorID:     "user-return-inspector",
		RequestID:   "req-return-inspect",
	})
	if err != nil {
		t.Fatalf("inspect return: %v", err)
	}

	invalid := newReturnAttachmentRequest(t, "rr-260426-0001", "inspect-rr-260426-0001-intact", "return-photo.exe", "binary")
	invalidRec := httptest.NewRecorder()

	returnAttachmentHandler(
		returnsapp.NewUploadReturnAttachment(store, auditStore).WithObjectStore(objectStore),
	).ServeHTTP(invalidRec, invalid)
	if invalidRec.Code != http.StatusBadRequest {
		t.Fatalf("invalid file status = %d, want %d: %s", invalidRec.Code, http.StatusBadRequest, invalidRec.Body.String())
	}
}

func newReturnAttachmentRequest(
	t *testing.T,
	receiptID string,
	inspectionID string,
	fileName string,
	content string,
) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("inspection_id", inspectionID); err != nil {
		t.Fatalf("write inspection field: %v", err)
	}
	filePart, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		t.Fatalf("create file part: %v", err)
	}
	if _, err := filePart.Write([]byte(content)); err != nil {
		t.Fatalf("write file part: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/returns/"+receiptID+"/attachments", &body)
	req.SetPathValue("return_receipt_id", receiptID)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead)))

	return req
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

func TestStockAdjustmentsHandlerCreatesDraftRequestWithoutStockMovement(t *testing.T) {
	adjustmentStore := inventoryapp.NewPrototypeStockAdjustmentStore()
	auditStore := audit.NewInMemoryLogStore()
	handler := stockAdjustmentsHandler(
		inventoryapp.NewListStockAdjustments(adjustmentStore),
		inventoryapp.NewCreateStockAdjustment(adjustmentStore, auditStore),
	)
	body := bytes.NewBufferString(`{
		"adjustment_no": "ADJ-HCM-260428-001",
		"warehouse_id": "wh-hcm",
		"warehouse_code": "HCM",
		"source_type": "stock_count",
		"source_id": "count-hcm-260428-001",
		"reason": "cycle count variance",
		"lines": [
			{
				"sku": "serum-30ml",
				"expected_qty": "20",
				"counted_qty": "18",
				"base_uom_code": "EA",
				"reason": "short count"
			}
		]
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/stock-adjustments", body)
	req.Header.Set(response.HeaderRequestID, "req-stock-adjustment")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead)))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var payload response.SuccessEnvelope[stockAdjustmentResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode adjustment response: %v", err)
	}
	if payload.Data.Status != "draft" ||
		payload.Data.Lines[0].DeltaQty != "-2.000000" ||
		payload.Data.AuditLogID == "" {
		t.Fatalf("payload = %+v, want draft adjustment request with variance", payload.Data)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/stock-adjustments", nil).WithContext(req.Context())
	listRec := httptest.NewRecorder()
	handler.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d: %s", listRec.Code, http.StatusOK, listRec.Body.String())
	}
	var listPayload response.SuccessEnvelope[[]stockAdjustmentResponse]
	if err := json.NewDecoder(listRec.Body).Decode(&listPayload); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listPayload.Data) != 1 {
		t.Fatalf("list payload = %+v, want one adjustment", listPayload.Data)
	}

	movementLogs, err := auditStore.List(req.Context(), audit.Query{Action: "inventory.stock_movement.adjusted"})
	if err != nil {
		t.Fatalf("list stock movement audit logs: %v", err)
	}
	if len(movementLogs) != 0 {
		t.Fatalf("stock movement logs = %+v, want no direct stock movement", movementLogs)
	}
}

func TestStockAdjustmentsHandlerRejectsNoVariance(t *testing.T) {
	adjustmentStore := inventoryapp.NewPrototypeStockAdjustmentStore()
	auditStore := audit.NewInMemoryLogStore()
	handler := stockAdjustmentsHandler(
		inventoryapp.NewListStockAdjustments(adjustmentStore),
		inventoryapp.NewCreateStockAdjustment(adjustmentStore, auditStore),
	)
	body := bytes.NewBufferString(`{
		"warehouse_id": "wh-hcm",
		"reason": "no variance",
		"lines": [
			{
				"sku": "SERUM-30ML",
				"expected_qty": "20",
				"counted_qty": "20",
				"base_uom_code": "EA"
			}
		]
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/stock-adjustments", body)
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead)))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestStockAdjustmentActionHandlerApprovesAndPosts(t *testing.T) {
	adjustmentStore := inventoryapp.NewPrototypeStockAdjustmentStore()
	auditStore := audit.NewInMemoryLogStore()
	movementStore := inventoryapp.NewInMemoryStockMovementStore()
	createService := inventoryapp.NewCreateStockAdjustment(adjustmentStore, auditStore)
	created, err := createService.Execute(context.Background(), inventoryapp.CreateStockAdjustmentInput{
		ID:          "adj-hcm-action",
		WarehouseID: "wh-hcm",
		Reason:      "cycle count variance",
		RequestedBy: "user-warehouse-lead",
		Lines: []inventoryapp.CreateStockAdjustmentLineInput{
			{ID: "line-serum", SKU: "SERUM-30ML", ExpectedQty: "20", CountedQty: "18", BaseUOMCode: "EA"},
		},
	})
	if err != nil {
		t.Fatalf("create adjustment: %v", err)
	}
	service := inventoryapp.NewTransitionStockAdjustment(adjustmentStore, movementStore, auditStore)
	warehouseLeadContext := auth.WithPrincipal(context.Background(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "lead@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead))

	submitReq := httptest.NewRequest(http.MethodPost, "/api/v1/stock-adjustments/adj-hcm-action/submit", nil).WithContext(warehouseLeadContext)
	submitReq.SetPathValue("stock_adjustment_id", created.Adjustment.ID)
	submitRec := httptest.NewRecorder()
	stockAdjustmentActionHandler(service, "submit").ServeHTTP(submitRec, submitReq)
	if submitRec.Code != http.StatusOK {
		t.Fatalf("submit status = %d, want %d: %s", submitRec.Code, http.StatusOK, submitRec.Body.String())
	}

	approveReq := httptest.NewRequest(http.MethodPost, "/api/v1/stock-adjustments/adj-hcm-action/approve", nil).WithContext(warehouseLeadContext)
	approveReq.SetPathValue("stock_adjustment_id", created.Adjustment.ID)
	approveRec := httptest.NewRecorder()
	stockAdjustmentActionHandler(service, "approve").ServeHTTP(approveRec, approveReq)
	if approveRec.Code != http.StatusOK {
		t.Fatalf("approve status = %d, want %d: %s", approveRec.Code, http.StatusOK, approveRec.Body.String())
	}

	postReq := httptest.NewRequest(http.MethodPost, "/api/v1/stock-adjustments/adj-hcm-action/post", nil).WithContext(warehouseLeadContext)
	postReq.SetPathValue("stock_adjustment_id", created.Adjustment.ID)
	postRec := httptest.NewRecorder()
	stockAdjustmentActionHandler(service, "post").ServeHTTP(postRec, postReq)
	if postRec.Code != http.StatusOK {
		t.Fatalf("post status = %d, want %d: %s", postRec.Code, http.StatusOK, postRec.Body.String())
	}
	var payload response.SuccessEnvelope[stockAdjustmentResponse]
	if err := json.NewDecoder(postRec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode post response: %v", err)
	}
	if payload.Data.Status != "posted" || payload.Data.PostedBy == "" {
		t.Fatalf("post payload = %+v, want posted adjustment", payload.Data)
	}
	if movementStore.Count() != 1 || movementStore.Movements()[0].MovementType != inventorydomain.MovementAdjustmentOut {
		t.Fatalf("movements = %+v, want one adjustment out", movementStore.Movements())
	}
}

func TestStockCountsHandlerCreatesAndSubmitsVarianceReview(t *testing.T) {
	stockCountStore := inventoryapp.NewPrototypeStockCountStore()
	auditStore := audit.NewInMemoryLogStore()
	listService := inventoryapp.NewListStockCounts(stockCountStore)
	createService := inventoryapp.NewCreateStockCount(stockCountStore, auditStore)
	submitService := inventoryapp.NewSubmitStockCount(stockCountStore, auditStore)
	principalContext := auth.WithPrincipal(context.Background(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead))

	createReq := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/stock-counts",
		bytes.NewBufferString(`{
			"id": "count-hcm-001",
			"count_no": "CNT-HCM-001",
			"warehouse_id": "wh-hcm",
			"scope": "cycle_count",
			"lines": [
				{"id": "line-serum", "sku": "SERUM-30ML", "expected_qty": "20", "base_uom_code": "EA"}
			]
		}`),
	).WithContext(principalContext)
	createReq.Header.Set(response.HeaderRequestID, "req-stock-count-create")
	createRec := httptest.NewRecorder()

	stockCountsHandler(listService, createService).ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d: %s", createRec.Code, http.StatusCreated, createRec.Body.String())
	}
	var createPayload response.SuccessEnvelope[stockCountResponse]
	if err := json.NewDecoder(createRec.Body).Decode(&createPayload); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if createPayload.Data.Status != "open" || createPayload.Data.AuditLogID == "" {
		t.Fatalf("create payload = %+v, want open stock count with audit", createPayload.Data)
	}

	submitReq := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/stock-counts/count-hcm-001/submit",
		bytes.NewBufferString(`{"lines":[{"id":"line-serum","counted_qty":"18","note":"short count"}]}`),
	).WithContext(principalContext)
	submitReq.SetPathValue("stock_count_id", "count-hcm-001")
	submitReq.Header.Set(response.HeaderRequestID, "req-stock-count-submit")
	submitRec := httptest.NewRecorder()

	stockCountSubmitHandler(submitService).ServeHTTP(submitRec, submitReq)

	if submitRec.Code != http.StatusOK {
		t.Fatalf("submit status = %d, want %d: %s", submitRec.Code, http.StatusOK, submitRec.Body.String())
	}
	var submitPayload response.SuccessEnvelope[stockCountResponse]
	if err := json.NewDecoder(submitRec.Body).Decode(&submitPayload); err != nil {
		t.Fatalf("decode submit response: %v", err)
	}
	if submitPayload.Data.Status != "variance_review" ||
		submitPayload.Data.Lines[0].DeltaQty != "-2.000000" ||
		submitPayload.Data.AuditLogID == "" {
		t.Fatalf("submit payload = %+v, want variance review", submitPayload.Data)
	}

	logs, err := auditStore.List(context.Background(), audit.Query{EntityID: "count-hcm-001"})
	if err != nil {
		t.Fatalf("list stock count audit logs: %v", err)
	}
	if len(logs) != 2 {
		t.Fatalf("audit logs = %+v, want create and submit logs", logs)
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
	).WithPurchaseOrderReader(testGoodsReceiptPurchaseOrderReader{
		order: approvedTestGoodsReceiptPurchaseOrder(),
	})

	return service, auditStore
}

type testGoodsReceiptPurchaseOrderReader struct {
	order purchasedomain.PurchaseOrder
}

func (r testGoodsReceiptPurchaseOrderReader) GetPurchaseOrder(
	_ context.Context,
	id string,
) (purchasedomain.PurchaseOrder, error) {
	if r.order.ID != id {
		return purchasedomain.PurchaseOrder{}, inventoryapp.ErrReceivingPurchaseOrderMismatch
	}

	return r.order.Clone(), nil
}

func approvedTestGoodsReceiptPurchaseOrder() purchasedomain.PurchaseOrder {
	order, err := purchasedomain.NewPurchaseOrderDocument(purchasedomain.NewPurchaseOrderDocumentInput{
		ID:            "PO-260427-API",
		OrgID:         "org-my-pham",
		PONo:          "PO-260427-API",
		SupplierID:    "supplier-local",
		SupplierCode:  "SUP-LOCAL",
		SupplierName:  "Local Supplier",
		WarehouseID:   "wh-hcm-fg",
		WarehouseCode: "WH-HCM-FG",
		ExpectedDate:  "2026-04-29",
		CurrencyCode:  "VND",
		Lines: []purchasedomain.NewPurchaseOrderLineInput{
			{
				ID:           "po-line-260427-api-001",
				LineNo:       1,
				ItemID:       "item-cream-50g",
				SKUCode:      "CREAM-50G",
				ItemName:     "Moisturizing Cream",
				OrderedQty:   decimal.MustQuantity("12"),
				UOMCode:      "EA",
				BaseUOMCode:  "EA",
				UnitPrice:    decimal.MustUnitPrice("1"),
				CurrencyCode: "VND",
			},
		},
		CreatedAt: time.Date(2026, 4, 27, 9, 0, 0, 0, time.UTC),
		CreatedBy: "user-purchase-ops",
	})
	if err != nil {
		panic(err)
	}
	submitted, err := order.Submit("user-purchase-ops", time.Date(2026, 4, 27, 9, 30, 0, 0, time.UTC))
	if err != nil {
		panic(err)
	}
	approved, err := submitted.Approve("user-purchase-ops", time.Date(2026, 4, 27, 10, 0, 0, 0, time.UTC))
	if err != nil {
		panic(err)
	}

	return approved
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

type recordingCarrierManifestSalesOrderHandover struct {
	orderNos []string
}

func (r *recordingCarrierManifestSalesOrderHandover) MarkSalesOrderHandedOver(
	_ context.Context,
	input shippingapp.CarrierManifestSalesOrderHandoverInput,
) (salesdomain.SalesOrder, error) {
	r.orderNos = append(r.orderNos, input.OrderNo)
	return salesdomain.SalesOrder{
		OrderNo: input.OrderNo,
		Status:  salesdomain.SalesOrderStatusHandedOver,
	}, nil
}
