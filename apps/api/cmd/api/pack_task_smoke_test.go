package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	masterdataapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/application"
	salesapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/application"
	salesdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/domain"
	shippingapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestPackTaskAPIActionsSmoke(t *testing.T) {
	authConfig := smokeAuthConfig()
	packStore := shippingapp.NewPrototypePackTaskStore(mustPrototypePackTask())
	auditStore := audit.NewInMemoryLogStore()
	salesService := newTestPackTaskSalesOrderService(auditStore)

	listRec := httptest.NewRecorder()
	listReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodGet, "/api/v1/pack-tasks?status=created", nil),
		authConfig,
		auth.RoleWarehouseStaff,
	)
	packTasksHandler(shippingapp.NewListPackTasks(packStore)).ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d: %s", listRec.Code, http.StatusOK, listRec.Body.String())
	}
	listPayload := decodeSmokeSuccess[[]packTaskResponse](t, listRec)
	if len(listPayload.Data) != 1 || listPayload.Data[0].ID != "pack-so-260428-0003" {
		t.Fatalf("list payload = %+v, want seeded pack task", listPayload.Data)
	}

	startRec := httptest.NewRecorder()
	startReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/pack-tasks/pack-so-260428-0003/start", nil),
		authConfig,
		auth.RoleWarehouseLead,
	)
	startReq.SetPathValue("pack_task_id", "pack-so-260428-0003")
	startReq.Header.Set(response.HeaderRequestID, "req-pack-start-smoke")
	startPackTaskHandler(shippingapp.NewStartPackTask(packStore, auditStore)).ServeHTTP(startRec, startReq)
	if startRec.Code != http.StatusOK {
		t.Fatalf("start status = %d, want %d: %s", startRec.Code, http.StatusOK, startRec.Body.String())
	}
	startPayload := decodeSmokeSuccess[packTaskResponse](t, startRec)
	if startPayload.Data.Status != "in_progress" || startPayload.Data.AuditLogID == "" {
		t.Fatalf("start payload = %+v, want in-progress task with audit", startPayload.Data)
	}

	confirmBody := bytes.NewBufferString(`{"lines":[{"line_id":"pack-so-260428-0003-line-01","packed_qty":"3"}]}`)
	confirmRec := httptest.NewRecorder()
	confirmReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/pack-tasks/pack-so-260428-0003/confirm", confirmBody),
		authConfig,
		auth.RoleWarehouseLead,
	)
	confirmReq.SetPathValue("pack_task_id", "pack-so-260428-0003")
	confirmReq.Header.Set(response.HeaderRequestID, "req-pack-confirm-smoke")
	confirmPackTaskHandler(
		shippingapp.NewConfirmPackTask(packStore, auditStore, salesOrderPackerAdapter{service: salesService}),
	).ServeHTTP(confirmRec, confirmReq)
	if confirmRec.Code != http.StatusOK {
		t.Fatalf("confirm status = %d, want %d: %s", confirmRec.Code, http.StatusOK, confirmRec.Body.String())
	}
	confirmPayload := decodeSmokeSuccess[packTaskResponse](t, confirmRec)
	if confirmPayload.Data.Status != "packed" ||
		confirmPayload.Data.Lines[0].Status != "packed" ||
		confirmPayload.Data.Lines[0].QtyPacked != "3.000000" ||
		confirmPayload.Data.SalesOrderStatus != "packed" ||
		confirmPayload.Data.AuditLogID == "" {
		t.Fatalf("confirm payload = %+v, want packed task, line, order, and audit", confirmPayload.Data)
	}
	order, err := salesService.GetSalesOrder(confirmReq.Context(), "so-260428-0003")
	if err != nil {
		t.Fatalf("get sales order: %v", err)
	}
	if order.Status != salesdomain.SalesOrderStatusPacked {
		t.Fatalf("sales order status = %s, want packed", order.Status)
	}
}

func TestPackTaskAPIExceptionSmoke(t *testing.T) {
	authConfig := smokeAuthConfig()
	cases := []struct {
		name          string
		exceptionCode string
		investigation string
	}{
		{
			name:          "missing stock",
			exceptionCode: "missing_stock",
			investigation: "Packed quantity did not match the order line",
		},
		{
			name:          "wrong SKU",
			exceptionCode: "wrong_sku",
			investigation: "Scanner reported a different SKU",
		},
		{
			name:          "wrong batch",
			exceptionCode: "wrong_batch",
			investigation: "Scanner reported a different batch",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			packStore := shippingapp.NewPrototypePackTaskStore(mustPrototypePackTask())
			auditStore := audit.NewInMemoryLogStore()

			body := bytes.NewBufferString(fmt.Sprintf(
				`{"line_id":"pack-so-260428-0003-line-01","exception_code":%q,"investigation":%q}`,
				tc.exceptionCode,
				tc.investigation,
			))
			rec := httptest.NewRecorder()
			req := smokeRequestAsRole(
				httptest.NewRequest(http.MethodPost, "/api/v1/pack-tasks/pack-so-260428-0003/exception", body),
				authConfig,
				auth.RoleWarehouseLead,
			)
			req.SetPathValue("pack_task_id", "pack-so-260428-0003")
			req.Header.Set(response.HeaderRequestID, "req-pack-exception-smoke")
			reportPackTaskExceptionHandler(shippingapp.NewReportPackTaskException(packStore, auditStore)).ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("exception status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
			}
			payload := decodeSmokeSuccess[packTaskResponse](t, rec)
			if payload.Data.Status != "pack_exception" ||
				payload.Data.Lines[0].Status != "pack_exception" ||
				payload.Data.Lines[0].QtyPacked != "0.000000" ||
				payload.Data.AuditLogID == "" {
				t.Fatalf("exception payload = %+v, want pack exception task and line with audit", payload.Data)
			}

			confirmRec := httptest.NewRecorder()
			confirmReq := smokeRequestAsRole(
				httptest.NewRequest(http.MethodPost, "/api/v1/pack-tasks/pack-so-260428-0003/confirm", nil),
				authConfig,
				auth.RoleWarehouseLead,
			)
			confirmReq.SetPathValue("pack_task_id", "pack-so-260428-0003")
			confirmPackTaskHandler(
				shippingapp.NewConfirmPackTask(packStore, auditStore, salesOrderPackerAdapter{service: newTestPackTaskSalesOrderService(auditStore)}),
			).ServeHTTP(confirmRec, confirmReq)
			if confirmRec.Code != http.StatusConflict {
				t.Fatalf("confirm exception status = %d, want %d: %s", confirmRec.Code, http.StatusConflict, confirmRec.Body.String())
			}
			errorPayload := decodeSmokeError(t, confirmRec)
			if errorPayload.Error.Code != response.ErrorCodeConflict {
				t.Fatalf("confirm exception code = %s, want %s", errorPayload.Error.Code, response.ErrorCodeConflict)
			}
		})
	}
}

func TestPackTaskAPIPermissions(t *testing.T) {
	authConfig := smokeAuthConfig()
	packStore := shippingapp.NewPrototypePackTaskStore(mustPrototypePackTask())
	auditStore := audit.NewInMemoryLogStore()

	req := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/pack-tasks/pack-so-260428-0003/start", nil),
		authConfig,
		auth.RoleWarehouseStaff,
	)
	req.SetPathValue("pack_task_id", "pack-so-260428-0003")
	rec := httptest.NewRecorder()

	startPackTaskHandler(shippingapp.NewStartPackTask(packStore, auditStore)).ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("start status = %d, want %d: %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
	payload := decodeSmokeError(t, rec)
	if payload.Error.Code != response.ErrorCodeForbidden {
		t.Fatalf("code = %s, want %s", payload.Error.Code, response.ErrorCodeForbidden)
	}
}

func newTestPackTaskSalesOrderService(auditStore audit.LogStore) salesapp.SalesOrderService {
	return salesapp.NewSalesOrderService(
		salesapp.NewPrototypeSalesOrderStore(auditStore),
		masterdataapp.NewPrototypePartyCatalog(auditStore),
		masterdataapp.NewPrototypeItemCatalog(auditStore),
		masterdataapp.NewPrototypeWarehouseLocationCatalog(auditStore),
	)
}
