package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestSubcontractMaterialIssueE2ESmoke(t *testing.T) {
	authConfig := smokeAuthConfig()
	service, movementStore, transferStore, auditStore := newTestSubcontractMaterialIssueAPIService()
	orderID := "sco-e2e-260429-material-issue"
	createAndSubmitSubcontractOrderForTest(t, service, authConfig, orderID)
	approveAndConfirmSubcontractOrderForTest(t, service, authConfig, orderID)

	issueReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/"+orderID+"/issue-materials", bytes.NewBufferString(`{
			"expected_version": 4,
			"transfer_id": "smt-e2e-260429-material-issue",
			"transfer_no": "SMT-E2E-260429-MATERIAL-ISSUE",
			"source_warehouse_id": "wh-hcm-rm",
			"source_warehouse_code": "WH-HCM-RM",
			"handover_by": "warehouse-lead",
			"handover_at": "2026-04-29T09:30:00Z",
			"received_by": "factory-receiver",
			"receiver_contact": "0988000222",
			"vehicle_no": "51A-54321",
			"lines": [
				{
					"order_material_line_id": "`+orderID+`-material-01",
					"issue_qty": "20",
					"uom_code": "EA",
					"batch_id": "batch-cream-2603b",
					"batch_no": "LOT-CREAM-2603B",
					"lot_no": "LOT-CREAM-2603B",
					"source_bin_id": "rm-a01"
				}
			],
			"evidence": [
				{
					"id": "smt-e2e-260429-material-issue-handover",
					"evidence_type": "handover",
					"file_name": "handover.pdf",
					"object_key": "subcontract/smt-e2e-260429-material-issue/handover.pdf"
				}
			]
		}`)),
		authConfig,
		auth.RoleProductionOps,
	)
	issueReq.SetPathValue("subcontract_order_id", orderID)
	issueReq.Header.Set(response.HeaderRequestID, "req-e2e-subcontract-material-issue")
	issueRec := httptest.NewRecorder()

	subcontractOrderIssueMaterialsHandler(service).ServeHTTP(issueRec, issueReq)

	if issueRec.Code != http.StatusOK {
		t.Fatalf("issue materials status = %d, want %d: %s", issueRec.Code, http.StatusOK, issueRec.Body.String())
	}
	payload := decodeSmokeSuccess[issueSubcontractMaterialsResponse](t, issueRec).Data
	if payload.SubcontractOrder.Status != string(productiondomain.SubcontractOrderStatusMaterialsIssued) ||
		payload.SubcontractOrder.Version != 5 ||
		payload.Transfer.Status != "sent_to_factory" ||
		len(payload.Transfer.Evidence) != 1 ||
		payload.Transfer.Evidence[0].ObjectKey != "subcontract/smt-e2e-260429-material-issue/handover.pdf" ||
		payload.AuditLogID == "" {
		t.Fatalf("issue payload = %+v, want issued order, transfer evidence, and audit", payload)
	}
	if transferStore.Count() != 1 {
		t.Fatalf("transfer count = %d, want 1", transferStore.Count())
	}
	if movementStore.Count() != 1 || len(payload.StockMovements) != 1 {
		t.Fatalf("stock movements = store %d response %+v, want one movement", movementStore.Count(), payload.StockMovements)
	}
	movement := movementStore.Movements()[0]
	if movement.MovementType != inventorydomain.MovementSubcontractIssue ||
		movement.StockStatus != inventorydomain.StockStatusSubcontractIssued ||
		movement.SourceDocType != "subcontract_material_transfer" ||
		movement.SourceDocID != payload.Transfer.ID ||
		movement.BatchID != "batch-cream-2603b" {
		t.Fatalf("movement = %+v, want subcontract issue movement linked to transfer and batch", movement)
	}
	delta, err := movement.BalanceDelta()
	if err != nil {
		t.Fatalf("material issue balance delta: %v", err)
	}
	if delta.OnHand.String() != "-20.000000" ||
		delta.Available.String() != "-20.000000" ||
		!delta.Reserved.IsZero() {
		t.Fatalf("movement delta = %+v, want issued qty removed from on-hand/available and not reserved", delta)
	}
	order, err := service.GetSubcontractOrder(issueReq.Context(), orderID)
	if err != nil {
		t.Fatalf("get issued subcontract order: %v", err)
	}
	if order.Status != productiondomain.SubcontractOrderStatusMaterialsIssued || order.Version != 5 {
		t.Fatalf("order = %+v, want materials_issued_to_factory version 5", order)
	}

	assertSubcontractMaterialIssueE2EAuditAction(t, auditStore, "subcontract.order.created")
	assertSubcontractMaterialIssueE2EAuditAction(t, auditStore, "subcontract.order.submitted")
	assertSubcontractMaterialIssueE2EAuditAction(t, auditStore, "subcontract.order.approved")
	assertSubcontractMaterialIssueE2EAuditAction(t, auditStore, "subcontract.order.factory_confirmed")
	assertSubcontractMaterialIssueE2EAuditAction(t, auditStore, "subcontract.materials_issued")
}

func assertSubcontractMaterialIssueE2EAuditAction(t *testing.T, auditStore audit.LogStore, action string) {
	t.Helper()

	logs, err := auditStore.List(context.Background(), audit.Query{Action: action})
	if err != nil {
		t.Fatalf("list audit logs for %s: %v", action, err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit action %s count = %d, want 1", action, len(logs))
	}
}
