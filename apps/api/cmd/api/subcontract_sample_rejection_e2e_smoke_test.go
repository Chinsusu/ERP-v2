package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	productionapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/application"
	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestSubcontractSampleRejectionE2ESmoke(t *testing.T) {
	authConfig := smokeAuthConfig()
	service, sampleStore, auditStore := newTestSubcontractSampleAPIService()
	orderID := "sco-e2e-260429-sample-rejection"
	sampleID := orderID + "-sample"
	createAndSubmitSubcontractOrderForTest(t, service, authConfig, orderID)
	approveAndConfirmSubcontractOrderForTest(t, service, authConfig, orderID)
	issueSubcontractMaterialsForTest(t, service, authConfig, orderID, 4)
	submitSubcontractSampleForTest(t, service, authConfig, orderID, 5)

	rejectReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/"+orderID+"/reject-sample", bytes.NewBufferString(`{
			"expected_version": 6,
			"sample_approval_id": "`+sampleID+`",
			"reason": "Texture and label color do not match approved spec",
			"decision_at": "2026-04-29T11:30:00Z"
		}`)),
		authConfig,
		auth.RoleProductionOps,
	)
	rejectReq.SetPathValue("subcontract_order_id", orderID)
	rejectReq.Header.Set(response.HeaderRequestID, "req-e2e-subcontract-sample-reject")
	rejectRec := httptest.NewRecorder()

	subcontractOrderRejectSampleHandler(service).ServeHTTP(rejectRec, rejectReq)

	if rejectRec.Code != http.StatusOK {
		t.Fatalf("reject sample status = %d, want %d: %s", rejectRec.Code, http.StatusOK, rejectRec.Body.String())
	}
	rejected := decodeSmokeSuccess[subcontractSampleApprovalResultResponse](t, rejectRec).Data
	if rejected.CurrentStatus != string(productiondomain.SubcontractOrderStatusSampleRejected) ||
		rejected.SampleApproval.Status != string(productiondomain.SubcontractSampleApprovalStatusRejected) ||
		rejected.SampleApproval.DecisionReason != "Texture and label color do not match approved spec" ||
		rejected.AuditLogID == "" {
		t.Fatalf("rejected sample = %+v, want rejected order/sample with audit", rejected)
	}
	if sampleStore.Count() != 1 {
		t.Fatalf("sample record count = %d, want one rejected sample", sampleStore.Count())
	}
	sample, err := sampleStore.Get(rejectReq.Context(), sampleID)
	if err != nil {
		t.Fatalf("get rejected sample: %v", err)
	}
	if sample.Status != productiondomain.SubcontractSampleApprovalStatusRejected || sample.Version != 2 {
		t.Fatalf("sample = %+v, want rejected version 2", sample)
	}

	startReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/"+orderID+"/start-mass-production", bytes.NewBufferString(`{"expected_version":7}`)),
		authConfig,
		auth.RoleProductionOps,
	)
	startReq.SetPathValue("subcontract_order_id", orderID)
	startReq.Header.Set(response.HeaderRequestID, "req-e2e-subcontract-sample-reject-start-mass")
	startRec := httptest.NewRecorder()

	subcontractOrderStartMassProductionHandler(service).ServeHTTP(startRec, startReq)

	if startRec.Code != http.StatusConflict {
		t.Fatalf("start mass production status = %d, want %d: %s", startRec.Code, http.StatusConflict, startRec.Body.String())
	}
	errorPayload := decodeSmokeError(t, startRec)
	if errorPayload.Error.Code != productionapp.ErrorCodeSubcontractOrderInvalidState {
		t.Fatalf("start mass production error code = %s, want %s", errorPayload.Error.Code, productionapp.ErrorCodeSubcontractOrderInvalidState)
	}
	order, err := service.GetSubcontractOrder(startReq.Context(), orderID)
	if err != nil {
		t.Fatalf("get sample rejected order: %v", err)
	}
	if order.Status != productiondomain.SubcontractOrderStatusSampleRejected || order.Version != 7 {
		t.Fatalf("order = %+v, want sample_rejected version 7", order)
	}

	assertSubcontractSampleRejectionE2EAuditAction(t, auditStore, "subcontract.sample_submitted")
	assertSubcontractSampleRejectionE2EAuditAction(t, auditStore, "subcontract.sample_rejected")
	requireNoSmokeAuditAction(t, auditStore, startReq.Context(), "subcontract.order.mass_production_started")
}

func assertSubcontractSampleRejectionE2EAuditAction(t *testing.T, auditStore audit.LogStore, action string) {
	t.Helper()

	logs, err := auditStore.List(context.Background(), audit.Query{Action: action})
	if err != nil {
		t.Fatalf("list audit logs for %s: %v", action, err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit action %s count = %d, want 1", action, len(logs))
	}
}
