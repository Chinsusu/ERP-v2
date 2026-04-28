package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	shippingapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestPickTaskAPIActionsSmoke(t *testing.T) {
	authConfig := smokeAuthConfig()
	store := shippingapp.NewPrototypePickTaskStore(mustPrototypePickTask())
	auditStore := audit.NewInMemoryLogStore()

	listRec := httptest.NewRecorder()
	listReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodGet, "/api/v1/pick-tasks?status=created", nil),
		authConfig,
		auth.RoleWarehouseStaff,
	)
	pickTasksHandler(shippingapp.NewListPickTasks(store)).ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d: %s", listRec.Code, http.StatusOK, listRec.Body.String())
	}
	listPayload := decodeSmokeSuccess[[]pickTaskResponse](t, listRec)
	if len(listPayload.Data) != 1 || listPayload.Data[0].ID != "pick-so-260428-0001" {
		t.Fatalf("list payload = %+v, want seeded pick task", listPayload.Data)
	}

	startRec := httptest.NewRecorder()
	startReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/pick-tasks/pick-so-260428-0001/start", nil),
		authConfig,
		auth.RoleWarehouseLead,
	)
	startReq.SetPathValue("pick_task_id", "pick-so-260428-0001")
	startReq.Header.Set(response.HeaderRequestID, "req-pick-start-smoke")
	startPickTaskHandler(shippingapp.NewStartPickTask(store, auditStore)).ServeHTTP(startRec, startReq)
	if startRec.Code != http.StatusOK {
		t.Fatalf("start status = %d, want %d: %s", startRec.Code, http.StatusOK, startRec.Body.String())
	}
	startPayload := decodeSmokeSuccess[pickTaskResponse](t, startRec)
	if startPayload.Data.Status != "in_progress" || startPayload.Data.AssignedTo == "" || startPayload.Data.AuditLogID == "" {
		t.Fatalf("start payload = %+v, want in-progress task with audit", startPayload.Data)
	}

	confirmBody := bytes.NewBufferString(`{"line_id":"pick-so-260428-0001-line-01","picked_qty":"3.000000"}`)
	confirmRec := httptest.NewRecorder()
	confirmReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/pick-tasks/pick-so-260428-0001/confirm-line", confirmBody),
		authConfig,
		auth.RoleWarehouseLead,
	)
	confirmReq.SetPathValue("pick_task_id", "pick-so-260428-0001")
	confirmReq.Header.Set(response.HeaderRequestID, "req-pick-confirm-smoke")
	confirmPickTaskLineHandler(shippingapp.NewConfirmPickTaskLine(store, auditStore)).ServeHTTP(confirmRec, confirmReq)
	if confirmRec.Code != http.StatusOK {
		t.Fatalf("confirm status = %d, want %d: %s", confirmRec.Code, http.StatusOK, confirmRec.Body.String())
	}
	confirmPayload := decodeSmokeSuccess[pickTaskResponse](t, confirmRec)
	if confirmPayload.Data.Lines[0].Status != "picked" || confirmPayload.Data.Lines[0].QtyPicked != "3.000000" || confirmPayload.Data.AuditLogID == "" {
		t.Fatalf("confirm payload = %+v, want picked line with audit", confirmPayload.Data)
	}

	completeRec := httptest.NewRecorder()
	completeReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/pick-tasks/pick-so-260428-0001/complete", nil),
		authConfig,
		auth.RoleWarehouseLead,
	)
	completeReq.SetPathValue("pick_task_id", "pick-so-260428-0001")
	completeReq.Header.Set(response.HeaderRequestID, "req-pick-complete-smoke")
	completePickTaskHandler(shippingapp.NewCompletePickTask(store, auditStore)).ServeHTTP(completeRec, completeReq)
	if completeRec.Code != http.StatusOK {
		t.Fatalf("complete status = %d, want %d: %s", completeRec.Code, http.StatusOK, completeRec.Body.String())
	}
	completePayload := decodeSmokeSuccess[pickTaskResponse](t, completeRec)
	if completePayload.Data.Status != "completed" || completePayload.Data.AuditLogID == "" {
		t.Fatalf("complete payload = %+v, want completed task with audit", completePayload.Data)
	}
}

func TestPickTaskAPIExceptionSmoke(t *testing.T) {
	authConfig := smokeAuthConfig()
	store := shippingapp.NewPrototypePickTaskStore(mustPrototypePickTask())
	auditStore := audit.NewInMemoryLogStore()
	body := bytes.NewBufferString(`{"line_id":"pick-so-260428-0001-line-01","exception_code":"wrong_location","investigation":"Scanner reported a different bin"}`)

	rec := httptest.NewRecorder()
	req := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/pick-tasks/pick-so-260428-0001/exception", body),
		authConfig,
		auth.RoleWarehouseLead,
	)
	req.SetPathValue("pick_task_id", "pick-so-260428-0001")
	req.Header.Set(response.HeaderRequestID, "req-pick-exception-smoke")

	reportPickTaskExceptionHandler(shippingapp.NewReportPickTaskException(store, auditStore)).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("exception status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	payload := decodeSmokeSuccess[pickTaskResponse](t, rec)
	if payload.Data.Status != "wrong_location" ||
		payload.Data.Lines[0].Status != "wrong_location" ||
		payload.Data.Lines[0].QtyPicked != "0.000000" ||
		payload.Data.AuditLogID == "" {
		t.Fatalf("exception payload = %+v, want wrong-location task and line with audit", payload.Data)
	}
}
