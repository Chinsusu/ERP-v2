package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	qcapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/qc/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestInboundQCInspectionHandlersCreateStartPass(t *testing.T) {
	service, auditStore := newTestInboundQCHandlerService()
	principalContext := auth.WithPrincipal(context.Background(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "qa@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleQA))

	createReq := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/inbound-qc-inspections",
		bytes.NewBufferString(`{
			"id": "iqc-handler-flow",
			"goods_receipt_id": "grn-hcm-260427-inspect",
			"goods_receipt_line_id": "grn-line-draft-001"
		}`),
	).WithContext(principalContext)
	createReq.Header.Set(response.HeaderRequestID, "req-iqc-create")
	createRec := httptest.NewRecorder()

	inboundQCInspectionsHandler(service).ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d: %s", createRec.Code, http.StatusCreated, createRec.Body.String())
	}
	var createPayload response.SuccessEnvelope[inboundQCActionResultResponse]
	if err := json.NewDecoder(createRec.Body).Decode(&createPayload); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if createPayload.Data.Inspection.Status != "pending" ||
		createPayload.Data.Inspection.GoodsReceiptNo != "GRN-260427-0003" ||
		createPayload.Data.Inspection.SKU != "CREAM-50G" ||
		createPayload.Data.Inspection.UOMCode != "EA" {
		t.Fatalf("create payload = %+v, want pending cream inspection", createPayload.Data)
	}

	startReq := httptest.NewRequest(http.MethodPost, "/api/v1/inbound-qc-inspections/iqc-handler-flow/start", nil).
		WithContext(principalContext)
	startReq.SetPathValue("inspection_id", "iqc-handler-flow")
	startRec := httptest.NewRecorder()
	inboundQCInspectionStartHandler(service).ServeHTTP(startRec, startReq)
	if startRec.Code != http.StatusOK {
		t.Fatalf("start status = %d, want %d: %s", startRec.Code, http.StatusOK, startRec.Body.String())
	}

	passReq := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/inbound-qc-inspections/iqc-handler-flow/pass",
		bytes.NewBufferString(`{
			"checklist": [
				{"id": "check-packaging", "code": "PACKAGING", "label": "Packaging condition", "required": true, "status": "pass"},
				{"id": "check-lot-expiry", "code": "LOT_EXPIRY", "label": "Lot and expiry match delivery", "required": true, "status": "pass"},
				{"id": "check-sample", "code": "SAMPLE", "label": "Sample retained when required", "status": "not_applicable"}
			]
		}`),
	).WithContext(principalContext)
	passReq.SetPathValue("inspection_id", "iqc-handler-flow")
	passReq.Header.Set(response.HeaderRequestID, "req-iqc-pass")
	passRec := httptest.NewRecorder()

	inboundQCInspectionPassHandler(service).ServeHTTP(passRec, passReq)

	if passRec.Code != http.StatusOK {
		t.Fatalf("pass status = %d, want %d: %s", passRec.Code, http.StatusOK, passRec.Body.String())
	}
	var passPayload response.SuccessEnvelope[inboundQCActionResultResponse]
	if err := json.NewDecoder(passRec.Body).Decode(&passPayload); err != nil {
		t.Fatalf("decode pass response: %v", err)
	}
	if passPayload.Data.PreviousStatus != "in_progress" ||
		passPayload.Data.CurrentStatus != "completed" ||
		passPayload.Data.CurrentResult != "pass" ||
		passPayload.Data.Inspection.PassedQuantity != passPayload.Data.Inspection.Quantity ||
		passPayload.Data.AuditLogID == "" {
		t.Fatalf("pass payload = %+v, want audited full pass", passPayload.Data)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/inbound-qc-inspections?status=completed", nil).
		WithContext(principalContext)
	listRec := httptest.NewRecorder()
	inboundQCInspectionsHandler(service).ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d: %s", listRec.Code, http.StatusOK, listRec.Body.String())
	}
	var listPayload response.SuccessEnvelope[[]inboundQCInspectionResponse]
	if err := json.NewDecoder(listRec.Body).Decode(&listPayload); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listPayload.Data) != 1 || listPayload.Data[0].ID != "iqc-handler-flow" {
		t.Fatalf("list payload = %+v, want completed handler flow inspection", listPayload.Data)
	}

	logs, err := auditStore.List(context.Background(), audit.Query{Action: "qc.inbound_inspection.passed"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 || logs[0].RequestID != "req-iqc-pass" {
		t.Fatalf("audit logs = %+v, want passed log with request id", logs)
	}
}

func TestInboundQCInspectionActionRequiresQCDecisionPermission(t *testing.T) {
	service, _ := newTestInboundQCHandlerService()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/inbound-qc-inspections/iqc-handler-flow/pass", nil)
	req.SetPathValue("inspection_id", "iqc-handler-flow")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "lead@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead)))
	rec := httptest.NewRecorder()

	inboundQCInspectionPassHandler(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func newTestInboundQCHandlerService() (qcapp.InboundQCInspectionService, *audit.InMemoryLogStore) {
	auditStore := audit.NewInMemoryLogStore()
	service := qcapp.NewInboundQCInspectionService(
		qcapp.NewPrototypeInboundQCInspectionStore(),
		inventoryapp.NewPrototypeWarehouseReceivingStore(),
		auditStore,
	)

	return service, auditStore
}
