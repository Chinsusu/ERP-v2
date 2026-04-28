package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	masterdataapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/application"
	salesapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestSalesOrderAPISmokePack(t *testing.T) {
	authConfig := smokeAuthConfig()

	t.Run("create update confirm and cancel with audit", func(t *testing.T) {
		service, auditStore := newTestSalesOrderAPIService()

		createBody := bytes.NewBufferString(`{
			"id": "so-smoke-260428-0001",
			"order_no": "SO-SMOKE-260428-0001",
			"customer_id": "cus-dl-minh-anh",
			"channel": "B2B",
			"warehouse_id": "wh-hcm-fg",
			"order_date": "2026-04-28",
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
			httptest.NewRequest(http.MethodPost, "/api/v1/sales-orders", createBody),
			authConfig,
			auth.RoleSalesOps,
		)
		createReq.Header.Set(response.HeaderRequestID, "req-sales-create")
		createRec := httptest.NewRecorder()

		salesOrdersHandler(service).ServeHTTP(createRec, createReq)

		if createRec.Code != http.StatusCreated {
			t.Fatalf("create status = %d, want %d: %s", createRec.Code, http.StatusCreated, createRec.Body.String())
		}
		created := decodeSmokeSuccess[salesOrderResponse](t, createRec).Data
		if created.Status != "draft" || created.TotalAmount != "250000.00" || created.Version != 1 || created.AuditLogID == "" {
			t.Fatalf("created order = %+v, want draft VND total with audit", created)
		}

		updateBody := bytes.NewBufferString(`{
			"expected_version": 1,
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
			httptest.NewRequest(http.MethodPatch, "/api/v1/sales-orders/so-smoke-260428-0001", updateBody),
			authConfig,
			auth.RoleSalesOps,
		)
		updateReq.SetPathValue("sales_order_id", "so-smoke-260428-0001")
		updateReq.Header.Set(response.HeaderRequestID, "req-sales-update")
		updateRec := httptest.NewRecorder()

		salesOrderDetailHandler(service).ServeHTTP(updateRec, updateReq)

		if updateRec.Code != http.StatusOK {
			t.Fatalf("update status = %d, want %d: %s", updateRec.Code, http.StatusOK, updateRec.Body.String())
		}
		updated := decodeSmokeSuccess[salesOrderResponse](t, updateRec).Data
		if updated.Version != 2 || updated.TotalAmount != "285000.00" || updated.Lines[0].SKUCode != "CREAM-50G" {
			t.Fatalf("updated order = %+v, want replaced cream line", updated)
		}

		confirmReq := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/sales-orders/so-smoke-260428-0001/confirm", bytes.NewBufferString(`{"expected_version":2}`)),
			authConfig,
			auth.RoleSalesOps,
		)
		confirmReq.SetPathValue("sales_order_id", "so-smoke-260428-0001")
		confirmReq.Header.Set(response.HeaderRequestID, "req-sales-confirm")
		confirmRec := httptest.NewRecorder()

		salesOrderConfirmHandler(service).ServeHTTP(confirmRec, confirmReq)

		if confirmRec.Code != http.StatusOK {
			t.Fatalf("confirm status = %d, want %d: %s", confirmRec.Code, http.StatusOK, confirmRec.Body.String())
		}
		confirmed := decodeSmokeSuccess[salesOrderActionResultResponse](t, confirmRec).Data
		if confirmed.PreviousStatus != "draft" || confirmed.CurrentStatus != "confirmed" || confirmed.SalesOrder.Version != 3 {
			t.Fatalf("confirmed result = %+v, want confirmed transition", confirmed)
		}

		cancelReq := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/sales-orders/so-smoke-260428-0001/cancel", bytes.NewBufferString(`{"expected_version":3,"reason":"customer changed order"}`)),
			authConfig,
			auth.RoleSalesOps,
		)
		cancelReq.SetPathValue("sales_order_id", "so-smoke-260428-0001")
		cancelReq.Header.Set(response.HeaderRequestID, "req-sales-cancel")
		cancelRec := httptest.NewRecorder()

		salesOrderCancelHandler(service).ServeHTTP(cancelRec, cancelReq)

		if cancelRec.Code != http.StatusOK {
			t.Fatalf("cancel status = %d, want %d: %s", cancelRec.Code, http.StatusOK, cancelRec.Body.String())
		}
		cancelled := decodeSmokeSuccess[salesOrderActionResultResponse](t, cancelRec).Data
		if cancelled.PreviousStatus != "confirmed" || cancelled.CurrentStatus != "cancelled" || cancelled.SalesOrder.CancelReason != "customer changed order" {
			t.Fatalf("cancelled result = %+v, want cancelled transition with reason", cancelled)
		}

		logs, err := auditStore.List(cancelReq.Context(), audit.Query{EntityID: "so-smoke-260428-0001"})
		if err != nil {
			t.Fatalf("list audit logs: %v", err)
		}
		if len(logs) != 4 {
			t.Fatalf("audit log count = %d, want 4", len(logs))
		}
	})

	t.Run("validates required lines", func(t *testing.T) {
		service, _ := newTestSalesOrderAPIService()
		req := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/sales-orders", bytes.NewBufferString(`{
				"customer_id": "cus-dl-minh-anh",
				"channel": "B2B",
				"warehouse_id": "wh-hcm-fg",
				"order_date": "2026-04-28",
				"currency_code": "VND",
				"lines": []
			}`)),
			authConfig,
			auth.RoleSalesOps,
		)
		rec := httptest.NewRecorder()

		salesOrdersHandler(service).ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		payload := decodeSmokeError(t, rec)
		if payload.Error.Code != salesapp.ErrorCodeSalesOrderValidation {
			t.Fatalf("code = %s, want %s", payload.Error.Code, salesapp.ErrorCodeSalesOrderValidation)
		}
	})

	t.Run("denies warehouse role without sales permission", func(t *testing.T) {
		service, _ := newTestSalesOrderAPIService()
		req := smokeRequestAsRole(
			httptest.NewRequest(http.MethodGet, "/api/v1/sales-orders", nil),
			authConfig,
			auth.RoleWarehouseStaff,
		)
		rec := httptest.NewRecorder()

		salesOrdersHandler(service).ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusForbidden, rec.Body.String())
		}
		payload := decodeSmokeError(t, rec)
		if payload.Error.Code != response.ErrorCodeForbidden {
			t.Fatalf("code = %s, want %s", payload.Error.Code, response.ErrorCodeForbidden)
		}
	})

	t.Run("returns not found for missing detail", func(t *testing.T) {
		service, _ := newTestSalesOrderAPIService()
		req := smokeRequestAsRole(
			httptest.NewRequest(http.MethodGet, "/api/v1/sales-orders/so-missing", nil),
			authConfig,
			auth.RoleSalesOps,
		)
		req.SetPathValue("sales_order_id", "so-missing")
		rec := httptest.NewRecorder()

		salesOrderDetailHandler(service).ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusNotFound, rec.Body.String())
		}
		payload := decodeSmokeError(t, rec)
		if payload.Error.Code != salesapp.ErrorCodeSalesOrderNotFound {
			t.Fatalf("code = %s, want %s", payload.Error.Code, salesapp.ErrorCodeSalesOrderNotFound)
		}
	})
}

func newTestSalesOrderAPIService() (salesapp.SalesOrderService, *audit.InMemoryLogStore) {
	auditStore := audit.NewInMemoryLogStore()
	service := salesapp.NewSalesOrderService(
		salesapp.NewPrototypeSalesOrderStore(auditStore),
		masterdataapp.NewPrototypePartyCatalog(auditStore),
		masterdataapp.NewPrototypeItemCatalog(auditStore),
		masterdataapp.NewPrototypeWarehouseLocationCatalog(auditStore),
	)

	return service, auditStore
}

func decodeSmokeError(t *testing.T, rec *httptest.ResponseRecorder) response.ErrorEnvelope {
	t.Helper()

	var payload response.ErrorEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode error response: %v", err)
	}

	return payload
}
