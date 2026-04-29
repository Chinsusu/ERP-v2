package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	masterdataapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/application"
	purchaseapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/purchase/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestPurchaseOrderAPISmoke(t *testing.T) {
	authConfig := smokeAuthConfig()

	t.Run("create update submit approve and close with audit", func(t *testing.T) {
		service, auditStore := newTestPurchaseOrderAPIService()

		createBody := bytes.NewBufferString(`{
			"id": "po-smoke-260429-0001",
			"po_no": "PO-SMOKE-260429-0001",
			"supplier_id": "sup-rm-bioactive",
			"warehouse_id": "wh-hcm-rm",
			"expected_date": "2026-05-02",
			"currency_code": "VND",
			"lines": [
				{
					"item_id": "item-serum-30ml",
					"ordered_qty": "2",
					"uom_code": "EA",
					"unit_price": "125000"
				}
			]
		}`)
		createReq := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/purchase-orders", createBody),
			authConfig,
			auth.RolePurchaseOps,
		)
		createReq.Header.Set(response.HeaderRequestID, "req-purchase-create")
		createRec := httptest.NewRecorder()

		purchaseOrdersHandler(service).ServeHTTP(createRec, createReq)

		if createRec.Code != http.StatusCreated {
			t.Fatalf("create status = %d, want %d: %s", createRec.Code, http.StatusCreated, createRec.Body.String())
		}
		created := decodeSmokeSuccess[purchaseOrderResponse](t, createRec).Data
		if created.Status != "draft" || created.TotalAmount != "250000.00" || created.Version != 1 || created.AuditLogID == "" {
			t.Fatalf("created order = %+v, want draft VND total with audit", created)
		}

		updateBody := bytes.NewBufferString(`{
			"expected_version": 1,
			"expected_date": "2026-05-03",
			"lines": [
				{
					"item_id": "item-cream-50g",
					"ordered_qty": "3",
					"uom_code": "EA",
					"unit_price": "95000"
				}
			]
		}`)
		updateReq := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPatch, "/api/v1/purchase-orders/po-smoke-260429-0001", updateBody),
			authConfig,
			auth.RolePurchaseOps,
		)
		updateReq.SetPathValue("purchase_order_id", "po-smoke-260429-0001")
		updateReq.Header.Set(response.HeaderRequestID, "req-purchase-update")
		updateRec := httptest.NewRecorder()

		purchaseOrderDetailHandler(service).ServeHTTP(updateRec, updateReq)

		if updateRec.Code != http.StatusOK {
			t.Fatalf("update status = %d, want %d: %s", updateRec.Code, http.StatusOK, updateRec.Body.String())
		}
		updated := decodeSmokeSuccess[purchaseOrderResponse](t, updateRec).Data
		if updated.Version != 2 || updated.TotalAmount != "285000.00" || updated.ExpectedDate != "2026-05-03" || updated.Lines[0].SKUCode != "CREAM-50G" {
			t.Fatalf("updated order = %+v, want replaced cream line", updated)
		}

		submitReq := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/purchase-orders/po-smoke-260429-0001/submit", bytes.NewBufferString(`{"expected_version":2}`)),
			authConfig,
			auth.RolePurchaseOps,
		)
		submitReq.SetPathValue("purchase_order_id", "po-smoke-260429-0001")
		submitReq.Header.Set(response.HeaderRequestID, "req-purchase-submit")
		submitRec := httptest.NewRecorder()

		purchaseOrderSubmitHandler(service).ServeHTTP(submitRec, submitReq)

		if submitRec.Code != http.StatusOK {
			t.Fatalf("submit status = %d, want %d: %s", submitRec.Code, http.StatusOK, submitRec.Body.String())
		}
		submitted := decodeSmokeSuccess[purchaseOrderActionResultResponse](t, submitRec).Data
		if submitted.PreviousStatus != "draft" || submitted.CurrentStatus != "submitted" || submitted.PurchaseOrder.Version != 3 {
			t.Fatalf("submitted result = %+v, want submitted transition", submitted)
		}

		approveReq := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/purchase-orders/po-smoke-260429-0001/approve", bytes.NewBufferString(`{"expected_version":3}`)),
			authConfig,
			auth.RolePurchaseOps,
		)
		approveReq.SetPathValue("purchase_order_id", "po-smoke-260429-0001")
		approveReq.Header.Set(response.HeaderRequestID, "req-purchase-approve")
		approveRec := httptest.NewRecorder()

		purchaseOrderApproveHandler(service).ServeHTTP(approveRec, approveReq)

		if approveRec.Code != http.StatusOK {
			t.Fatalf("approve status = %d, want %d: %s", approveRec.Code, http.StatusOK, approveRec.Body.String())
		}
		approved := decodeSmokeSuccess[purchaseOrderActionResultResponse](t, approveRec).Data
		if approved.PreviousStatus != "submitted" || approved.CurrentStatus != "approved" || approved.PurchaseOrder.Version != 4 {
			t.Fatalf("approved result = %+v, want approved transition", approved)
		}

		closeReq := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/purchase-orders/po-smoke-260429-0001/close", bytes.NewBufferString(`{"expected_version":4}`)),
			authConfig,
			auth.RolePurchaseOps,
		)
		closeReq.SetPathValue("purchase_order_id", "po-smoke-260429-0001")
		closeReq.Header.Set(response.HeaderRequestID, "req-purchase-close")
		closeRec := httptest.NewRecorder()

		purchaseOrderCloseHandler(service).ServeHTTP(closeRec, closeReq)

		if closeRec.Code != http.StatusOK {
			t.Fatalf("close status = %d, want %d: %s", closeRec.Code, http.StatusOK, closeRec.Body.String())
		}
		closed := decodeSmokeSuccess[purchaseOrderActionResultResponse](t, closeRec).Data
		if closed.PreviousStatus != "approved" || closed.CurrentStatus != "closed" || closed.PurchaseOrder.Version != 5 {
			t.Fatalf("closed result = %+v, want closed transition", closed)
		}

		logs, err := auditStore.List(closeReq.Context(), audit.Query{EntityID: "po-smoke-260429-0001"})
		if err != nil {
			t.Fatalf("list audit logs: %v", err)
		}
		if len(logs) != 5 {
			t.Fatalf("audit log count = %d, want 5", len(logs))
		}
	})

	t.Run("validates required lines", func(t *testing.T) {
		service, _ := newTestPurchaseOrderAPIService()
		req := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/purchase-orders", bytes.NewBufferString(`{
				"supplier_id": "sup-rm-bioactive",
				"warehouse_id": "wh-hcm-rm",
				"expected_date": "2026-05-02",
				"currency_code": "VND",
				"lines": []
			}`)),
			authConfig,
			auth.RolePurchaseOps,
		)
		rec := httptest.NewRecorder()

		purchaseOrdersHandler(service).ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		payload := decodeSmokeError(t, rec)
		if payload.Error.Code != purchaseapp.ErrorCodePurchaseOrderValidation {
			t.Fatalf("code = %s, want %s", payload.Error.Code, purchaseapp.ErrorCodePurchaseOrderValidation)
		}
	})

	t.Run("denies finance role from approval action without audit", func(t *testing.T) {
		service, auditStore := newTestPurchaseOrderAPIService()
		createAndSubmitPurchaseOrderForTest(t, service, authConfig, "po-smoke-260429-denied")
		req := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/purchase-orders/po-smoke-260429-denied/approve", bytes.NewBufferString(`{"expected_version":2}`)),
			authConfig,
			auth.RoleFinanceOps,
		)
		req.SetPathValue("purchase_order_id", "po-smoke-260429-denied")
		rec := httptest.NewRecorder()

		purchaseOrderApproveHandler(service).ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusForbidden, rec.Body.String())
		}
		payload := decodeSmokeError(t, rec)
		if payload.Error.Code != response.ErrorCodeForbidden {
			t.Fatalf("code = %s, want %s", payload.Error.Code, response.ErrorCodeForbidden)
		}
		logs, err := auditStore.List(req.Context(), audit.Query{Action: "purchase.order.approved"})
		if err != nil {
			t.Fatalf("list audit logs: %v", err)
		}
		if len(logs) != 0 {
			t.Fatalf("approval audit log count = %d, want 0 for denied action", len(logs))
		}
	})

	t.Run("approval audit captures actor timestamp and status delta", func(t *testing.T) {
		service, auditStore := newTestPurchaseOrderAPIService()
		createAndSubmitPurchaseOrderForTest(t, service, authConfig, "po-smoke-260429-audit")
		req := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/purchase-orders/po-smoke-260429-audit/approve", bytes.NewBufferString(`{"expected_version":2}`)),
			authConfig,
			auth.RolePurchaseOps,
		)
		req.SetPathValue("purchase_order_id", "po-smoke-260429-audit")
		req.Header.Set(response.HeaderRequestID, "req-purchase-approve-audit")
		rec := httptest.NewRecorder()

		purchaseOrderApproveHandler(service).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
		logs, err := auditStore.List(req.Context(), audit.Query{Action: "purchase.order.approved", EntityID: "po-smoke-260429-audit"})
		if err != nil {
			t.Fatalf("list audit logs: %v", err)
		}
		if len(logs) != 1 {
			t.Fatalf("approval audit log count = %d, want 1", len(logs))
		}
		log := logs[0]
		if log.ActorID != "user-purchase-ops" || log.RequestID != "req-purchase-approve-audit" || log.CreatedAt.IsZero() {
			t.Fatalf("approval audit identity = %+v, want purchase actor, request id, timestamp", log)
		}
		if log.BeforeData["status"] != "submitted" || log.AfterData["status"] != "approved" {
			t.Fatalf("approval audit delta before=%v after=%v, want submitted->approved", log.BeforeData, log.AfterData)
		}
	})
}

func newTestPurchaseOrderAPIService() (purchaseapp.PurchaseOrderService, audit.LogStore) {
	auditStore := audit.NewInMemoryLogStore()
	itemCatalog := masterdataapp.NewPrototypeItemCatalog(auditStore)
	warehouseCatalog := masterdataapp.NewPrototypeWarehouseLocationCatalog(auditStore)
	partyCatalog := masterdataapp.NewPrototypePartyCatalog(auditStore)
	uomCatalog := masterdataapp.NewPrototypeUOMCatalog()
	purchaseOrderStore := purchaseapp.NewPrototypePurchaseOrderStore(auditStore)

	return purchaseapp.NewPurchaseOrderService(
		purchaseOrderStore,
		partyCatalog,
		itemCatalog,
		warehouseCatalog,
		purchaseOrderUOMConverterAdapter{catalog: uomCatalog},
	), auditStore
}

func createAndSubmitPurchaseOrderForTest(
	t *testing.T,
	service purchaseapp.PurchaseOrderService,
	authConfig auth.MockConfig,
	id string,
) {
	t.Helper()
	createBody := bytes.NewBufferString(`{
		"id": "` + id + `",
		"po_no": "` + id + `",
		"supplier_id": "sup-rm-bioactive",
		"warehouse_id": "wh-hcm-rm",
		"expected_date": "2026-05-02",
		"currency_code": "VND",
		"lines": [
			{
				"item_id": "item-serum-30ml",
				"ordered_qty": "2",
				"uom_code": "EA",
				"unit_price": "125000"
			}
		]
	}`)
	createReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/purchase-orders", createBody),
		authConfig,
		auth.RolePurchaseOps,
	)
	createRec := httptest.NewRecorder()

	purchaseOrdersHandler(service).ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d: %s", createRec.Code, http.StatusCreated, createRec.Body.String())
	}
	submitReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/purchase-orders/"+id+"/submit", bytes.NewBufferString(`{"expected_version":1}`)),
		authConfig,
		auth.RolePurchaseOps,
	)
	submitReq.SetPathValue("purchase_order_id", id)
	submitRec := httptest.NewRecorder()

	purchaseOrderSubmitHandler(service).ServeHTTP(submitRec, submitReq)

	if submitRec.Code != http.StatusOK {
		t.Fatalf("submit status = %d, want %d: %s", submitRec.Code, http.StatusOK, submitRec.Body.String())
	}
}
