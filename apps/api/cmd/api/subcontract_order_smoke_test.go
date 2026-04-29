package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	masterdataapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/application"
	productionapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/application"
	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestSubcontractOrderAPISmoke(t *testing.T) {
	authConfig := smokeAuthConfig()

	t.Run("create update submit approve and confirm factory with audit", func(t *testing.T) {
		service, auditStore := newTestSubcontractOrderAPIService()

		createBody := bytes.NewBufferString(`{
			"id": "sco-smoke-260429-0001",
			"order_no": "SCO-SMOKE-260429-0001",
			"factory_id": "sup-out-lotus",
			"finished_item_id": "item-serum-30ml",
			"planned_qty": "100",
			"uom_code": "EA",
			"currency_code": "VND",
			"spec_summary": "Hydrating serum outsource batch",
			"sample_required": true,
			"claim_window_days": 7,
			"target_start_date": "2026-05-04",
			"expected_receipt_date": "2026-05-20",
			"material_lines": [
				{
					"item_id": "item-cream-50g",
					"planned_qty": "20",
					"uom_code": "EA",
					"unit_cost": "58000",
					"lot_trace_required": true
				}
			]
		}`)
		createReq := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders", createBody),
			authConfig,
			auth.RoleProductionOps,
		)
		createReq.Header.Set(response.HeaderRequestID, "req-subcontract-create")
		createRec := httptest.NewRecorder()

		subcontractOrdersHandler(service).ServeHTTP(createRec, createReq)

		if createRec.Code != http.StatusCreated {
			t.Fatalf("create status = %d, want %d: %s", createRec.Code, http.StatusCreated, createRec.Body.String())
		}
		created := decodeSmokeSuccess[subcontractOrderResponse](t, createRec).Data
		if created.Status != "draft" || created.EstimatedCostAmount != "1160000.00" || created.Version != 1 || created.AuditLogID == "" {
			t.Fatalf("created order = %+v, want draft VND estimate with audit", created)
		}

		updateBody := bytes.NewBufferString(`{
			"expected_version": 1,
			"planned_qty": "120",
			"expected_receipt_date": "2026-05-22",
			"material_lines": [
				{
					"item_id": "item-cream-50g",
					"planned_qty": "25",
					"uom_code": "EA",
					"unit_cost": "58000",
					"lot_trace_required": true
				}
			]
		}`)
		updateReq := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPatch, "/api/v1/subcontract-orders/sco-smoke-260429-0001", updateBody),
			authConfig,
			auth.RoleProductionOps,
		)
		updateReq.SetPathValue("subcontract_order_id", "sco-smoke-260429-0001")
		updateReq.Header.Set(response.HeaderRequestID, "req-subcontract-update")
		updateRec := httptest.NewRecorder()

		subcontractOrderDetailHandler(service).ServeHTTP(updateRec, updateReq)

		if updateRec.Code != http.StatusOK {
			t.Fatalf("update status = %d, want %d: %s", updateRec.Code, http.StatusOK, updateRec.Body.String())
		}
		updated := decodeSmokeSuccess[subcontractOrderResponse](t, updateRec).Data
		if updated.Version != 2 || updated.PlannedQty != "120.000000" || updated.EstimatedCostAmount != "1450000.00" || updated.ExpectedReceiptDate != "2026-05-22" {
			t.Fatalf("updated order = %+v, want replaced planned qty and material estimate", updated)
		}

		submitReq := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/sco-smoke-260429-0001/submit", bytes.NewBufferString(`{"expected_version":2}`)),
			authConfig,
			auth.RoleProductionOps,
		)
		submitReq.SetPathValue("subcontract_order_id", "sco-smoke-260429-0001")
		submitReq.Header.Set(response.HeaderRequestID, "req-subcontract-submit")
		submitRec := httptest.NewRecorder()

		subcontractOrderSubmitHandler(service).ServeHTTP(submitRec, submitReq)

		if submitRec.Code != http.StatusOK {
			t.Fatalf("submit status = %d, want %d: %s", submitRec.Code, http.StatusOK, submitRec.Body.String())
		}
		submitted := decodeSmokeSuccess[subcontractOrderActionResultResponse](t, submitRec).Data
		if submitted.PreviousStatus != "draft" || submitted.CurrentStatus != "submitted" || submitted.SubcontractOrder.Version != 3 {
			t.Fatalf("submitted result = %+v, want submitted transition", submitted)
		}

		approveReq := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/sco-smoke-260429-0001/approve", bytes.NewBufferString(`{"expected_version":3}`)),
			authConfig,
			auth.RoleProductionOps,
		)
		approveReq.SetPathValue("subcontract_order_id", "sco-smoke-260429-0001")
		approveReq.Header.Set(response.HeaderRequestID, "req-subcontract-approve")
		approveRec := httptest.NewRecorder()

		subcontractOrderApproveHandler(service).ServeHTTP(approveRec, approveReq)

		if approveRec.Code != http.StatusOK {
			t.Fatalf("approve status = %d, want %d: %s", approveRec.Code, http.StatusOK, approveRec.Body.String())
		}
		approved := decodeSmokeSuccess[subcontractOrderActionResultResponse](t, approveRec).Data
		if approved.PreviousStatus != "submitted" || approved.CurrentStatus != "approved" || approved.SubcontractOrder.Version != 4 {
			t.Fatalf("approved result = %+v, want approved transition", approved)
		}

		confirmReq := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/sco-smoke-260429-0001/confirm-factory", bytes.NewBufferString(`{"expected_version":4}`)),
			authConfig,
			auth.RoleProductionOps,
		)
		confirmReq.SetPathValue("subcontract_order_id", "sco-smoke-260429-0001")
		confirmReq.Header.Set(response.HeaderRequestID, "req-subcontract-confirm")
		confirmRec := httptest.NewRecorder()

		subcontractOrderConfirmFactoryHandler(service).ServeHTTP(confirmRec, confirmReq)

		if confirmRec.Code != http.StatusOK {
			t.Fatalf("confirm status = %d, want %d: %s", confirmRec.Code, http.StatusOK, confirmRec.Body.String())
		}
		confirmed := decodeSmokeSuccess[subcontractOrderActionResultResponse](t, confirmRec).Data
		if confirmed.PreviousStatus != "approved" || confirmed.CurrentStatus != "factory_confirmed" || confirmed.SubcontractOrder.Version != 5 {
			t.Fatalf("confirmed result = %+v, want factory_confirmed transition", confirmed)
		}

		logs, err := auditStore.List(confirmReq.Context(), audit.Query{EntityID: "sco-smoke-260429-0001"})
		if err != nil {
			t.Fatalf("list audit logs: %v", err)
		}
		if len(logs) != 5 {
			t.Fatalf("audit log count = %d, want 5", len(logs))
		}
	})

	t.Run("validates required material lines", func(t *testing.T) {
		service, _ := newTestSubcontractOrderAPIService()
		req := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders", bytes.NewBufferString(`{
				"factory_id": "sup-out-lotus",
				"finished_item_id": "item-serum-30ml",
				"planned_qty": "100",
				"uom_code": "EA",
				"currency_code": "VND",
				"claim_window_days": 7,
				"expected_receipt_date": "2026-05-20",
				"material_lines": []
			}`)),
			authConfig,
			auth.RoleProductionOps,
		)
		rec := httptest.NewRecorder()

		subcontractOrdersHandler(service).ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		payload := decodeSmokeError(t, rec)
		if payload.Error.Code != productionapp.ErrorCodeSubcontractOrderValidation {
			t.Fatalf("code = %s, want %s", payload.Error.Code, productionapp.ErrorCodeSubcontractOrderValidation)
		}
	})

	t.Run("issues materials with transfer stock movement and audit", func(t *testing.T) {
		service, movementStore, transferStore, auditStore := newTestSubcontractMaterialIssueAPIService()
		orderID := "sco-smoke-260429-issue"
		createAndSubmitSubcontractOrderForTest(t, service, authConfig, orderID)
		approveAndConfirmSubcontractOrderForTest(t, service, authConfig, orderID)

		issueReq := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/"+orderID+"/issue-materials", bytes.NewBufferString(`{
				"expected_version": 4,
				"transfer_id": "smt-smoke-260429-issue",
				"transfer_no": "SMT-SMOKE-260429-ISSUE",
				"source_warehouse_id": "wh-hcm-rm",
				"source_warehouse_code": "WH-HCM-RM",
				"handover_by": "warehouse-user",
				"handover_at": "2026-04-29T09:30:00Z",
				"received_by": "factory-receiver",
				"receiver_contact": "0988000111",
				"vehicle_no": "51A-12345",
				"lines": [
					{
						"order_material_line_id": "sco-smoke-260429-issue-material-01",
						"issue_qty": "20",
						"uom_code": "EA",
						"batch_id": "batch-cream-2603b",
						"source_bin_id": "rm-a01"
					}
				],
				"evidence": [
					{
						"id": "smt-smoke-260429-issue-handover",
						"evidence_type": "handover",
						"file_name": "handover.pdf",
						"object_key": "subcontract/smt-smoke-260429-issue/handover.pdf"
					}
				]
			}`)),
			authConfig,
			auth.RoleProductionOps,
		)
		issueReq.SetPathValue("subcontract_order_id", orderID)
		issueReq.Header.Set(response.HeaderRequestID, "req-subcontract-issue-materials")
		issueRec := httptest.NewRecorder()

		subcontractOrderIssueMaterialsHandler(service).ServeHTTP(issueRec, issueReq)

		if issueRec.Code != http.StatusOK {
			t.Fatalf("issue materials status = %d, want %d: %s", issueRec.Code, http.StatusOK, issueRec.Body.String())
		}
		payload := decodeSmokeSuccess[issueSubcontractMaterialsResponse](t, issueRec).Data
		if payload.SubcontractOrder.Status != "materials_issued_to_factory" ||
			payload.Transfer.Status != "sent_to_factory" ||
			payload.AuditLogID == "" {
			t.Fatalf("issue materials payload = %+v, want issued order, sent transfer, and audit", payload)
		}
		if len(payload.StockMovements) != 1 ||
			payload.StockMovements[0].MovementType != string(inventorydomain.MovementSubcontractIssue) ||
			payload.StockMovements[0].StockStatus != string(inventorydomain.StockStatusSubcontractIssued) ||
			payload.StockMovements[0].SourceDocID != payload.Transfer.ID {
			t.Fatalf("stock movements = %+v, want subcontract issue from transfer", payload.StockMovements)
		}
		if movementStore.Count() != 1 {
			t.Fatalf("stock movement count = %d, want 1", movementStore.Count())
		}
		if transferStore.Count() != 1 {
			t.Fatalf("transfer count = %d, want 1", transferStore.Count())
		}
		logs, err := auditStore.List(issueReq.Context(), audit.Query{Action: "subcontract.materials_issued"})
		if err != nil {
			t.Fatalf("list material issue audit logs: %v", err)
		}
		if len(logs) != 1 {
			t.Fatalf("material issue audit log count = %d, want 1", len(logs))
		}
	})

	t.Run("denies material issue without side effects", func(t *testing.T) {
		service, movementStore, transferStore, auditStore := newTestSubcontractMaterialIssueAPIService()
		req := smokeSubcontractViewOnlyRequest(
			httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/sco-smoke-denied/issue-materials", bytes.NewBufferString(`{
				"source_warehouse_id": "wh-hcm-rm",
				"received_by": "factory-receiver",
				"lines": []
			}`)),
			authConfig,
		)
		req.SetPathValue("subcontract_order_id", "sco-smoke-denied")
		rec := httptest.NewRecorder()

		subcontractOrderIssueMaterialsHandler(service).ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusForbidden, rec.Body.String())
		}
		if movementStore.Count() != 0 || transferStore.Count() != 0 {
			t.Fatalf("side effects = movements %d transfers %d, want none", movementStore.Count(), transferStore.Count())
		}
		requireNoSmokeAuditAction(t, auditStore, req.Context(), "subcontract.materials_issued")
	})

	t.Run("submits and approves subcontract samples with audit", func(t *testing.T) {
		service, sampleStore, auditStore := newTestSubcontractSampleAPIService()
		orderID := "sco-smoke-260429-sample"
		createAndSubmitSubcontractOrderForTest(t, service, authConfig, orderID)
		approveAndConfirmSubcontractOrderForTest(t, service, authConfig, orderID)
		issueSubcontractMaterialsForTest(t, service, authConfig, orderID, 4)

		submitReq := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/"+orderID+"/submit-sample", bytes.NewBufferString(`{
				"expected_version": 5,
				"sample_approval_id": "sample-smoke-260429-a",
				"sample_code": "SCO-SMOKE-260429-SAMPLE-A",
				"formula_version": "FORMULA-2026.04",
				"spec_version": "SPEC-2026.04",
				"submitted_by": "factory-user",
				"submitted_at": "2026-04-29T10:30:00Z",
				"evidence": [
					{
						"evidence_type": "photo",
						"file_name": "sample-front.jpg",
						"object_key": "subcontract/sample-smoke-260429-a/sample-front.jpg"
					}
				]
			}`)),
			authConfig,
			auth.RoleProductionOps,
		)
		submitReq.SetPathValue("subcontract_order_id", orderID)
		submitReq.Header.Set(response.HeaderRequestID, "req-subcontract-submit-sample")
		submitRec := httptest.NewRecorder()

		subcontractOrderSubmitSampleHandler(service).ServeHTTP(submitRec, submitReq)

		if submitRec.Code != http.StatusOK {
			t.Fatalf("submit sample status = %d, want %d: %s", submitRec.Code, http.StatusOK, submitRec.Body.String())
		}
		submitted := decodeSmokeSuccess[subcontractSampleApprovalResultResponse](t, submitRec).Data
		if submitted.CurrentStatus != "sample_submitted" ||
			submitted.SampleApproval.Status != "submitted" ||
			submitted.SampleApproval.SampleCode != "SCO-SMOKE-260429-SAMPLE-A" ||
			submitted.AuditLogID == "" {
			t.Fatalf("submitted sample = %+v, want submitted sample and audit", submitted)
		}

		approveReq := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/"+orderID+"/approve-sample", bytes.NewBufferString(`{
				"expected_version": 6,
				"sample_approval_id": "sample-smoke-260429-a",
				"reason": "Approved against retained standard",
				"storage_status": "retained_in_qa_cabinet",
				"decision_at": "2026-04-29T11:00:00Z"
			}`)),
			authConfig,
			auth.RoleProductionOps,
		)
		approveReq.SetPathValue("subcontract_order_id", orderID)
		approveReq.Header.Set(response.HeaderRequestID, "req-subcontract-approve-sample")
		approveRec := httptest.NewRecorder()

		subcontractOrderApproveSampleHandler(service).ServeHTTP(approveRec, approveReq)

		if approveRec.Code != http.StatusOK {
			t.Fatalf("approve sample status = %d, want %d: %s", approveRec.Code, http.StatusOK, approveRec.Body.String())
		}
		approved := decodeSmokeSuccess[subcontractSampleApprovalResultResponse](t, approveRec).Data
		if approved.CurrentStatus != "sample_approved" ||
			approved.SampleApproval.Status != "approved" ||
			approved.SampleApproval.StorageStatus != "retained_in_qa_cabinet" {
			t.Fatalf("approved sample = %+v, want approved sample and storage status", approved)
		}
		if sampleStore.Count() != 1 {
			t.Fatalf("sample record count = %d, want 1 updated record", sampleStore.Count())
		}
		logs, err := auditStore.List(approveReq.Context(), audit.Query{EntityID: orderID})
		if err != nil {
			t.Fatalf("list sample audit logs: %v", err)
		}
		if !smokeAuditActionsContain(logs, "subcontract.sample_submitted") ||
			!smokeAuditActionsContain(logs, "subcontract.sample_approved") {
			t.Fatalf("audit logs = %+v, want sample submit and approve actions", logs)
		}
	})

	t.Run("rejects subcontract sample with required reason and audit", func(t *testing.T) {
		service, sampleStore, auditStore := newTestSubcontractSampleAPIService()
		orderID := "sco-smoke-260429-sample-reject"
		createAndSubmitSubcontractOrderForTest(t, service, authConfig, orderID)
		approveAndConfirmSubcontractOrderForTest(t, service, authConfig, orderID)
		issueSubcontractMaterialsForTest(t, service, authConfig, orderID, 4)

		submitReq := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/"+orderID+"/submit-sample", bytes.NewBufferString(`{
				"expected_version": 5,
				"sample_approval_id": "sample-smoke-260429-reject",
				"sample_code": "SCO-SMOKE-260429-SAMPLE-R",
				"submitted_by": "factory-user",
				"evidence": [
					{
						"evidence_type": "photo",
						"object_key": "subcontract/sample-smoke-260429-reject/sample-front.jpg"
					}
				]
			}`)),
			authConfig,
			auth.RoleProductionOps,
		)
		submitReq.SetPathValue("subcontract_order_id", orderID)
		submitRec := httptest.NewRecorder()

		subcontractOrderSubmitSampleHandler(service).ServeHTTP(submitRec, submitReq)

		if submitRec.Code != http.StatusOK {
			t.Fatalf("submit sample status = %d, want %d: %s", submitRec.Code, http.StatusOK, submitRec.Body.String())
		}
		rejectReq := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/"+orderID+"/reject-sample", bytes.NewBufferString(`{
				"expected_version": 6,
				"sample_approval_id": "sample-smoke-260429-reject",
				"reason": "Label color does not match approved spec"
			}`)),
			authConfig,
			auth.RoleProductionOps,
		)
		rejectReq.SetPathValue("subcontract_order_id", orderID)
		rejectReq.Header.Set(response.HeaderRequestID, "req-subcontract-reject-sample")
		rejectRec := httptest.NewRecorder()

		subcontractOrderRejectSampleHandler(service).ServeHTTP(rejectRec, rejectReq)

		if rejectRec.Code != http.StatusOK {
			t.Fatalf("reject sample status = %d, want %d: %s", rejectRec.Code, http.StatusOK, rejectRec.Body.String())
		}
		rejected := decodeSmokeSuccess[subcontractSampleApprovalResultResponse](t, rejectRec).Data
		if rejected.CurrentStatus != "sample_rejected" ||
			rejected.SampleApproval.Status != "rejected" ||
			rejected.SubcontractOrder.SampleRejectReason != "Label color does not match approved spec" {
			t.Fatalf("rejected sample = %+v, want rejected order/sample with reason", rejected)
		}
		if sampleStore.Count() != 1 {
			t.Fatalf("sample record count = %d, want 1 updated record", sampleStore.Count())
		}
		logs, err := auditStore.List(rejectReq.Context(), audit.Query{Action: "subcontract.sample_rejected"})
		if err != nil {
			t.Fatalf("list reject sample audit logs: %v", err)
		}
		if len(logs) != 1 {
			t.Fatalf("reject sample audit count = %d, want 1", len(logs))
		}
	})

	t.Run("receives finished goods into qc hold with movement and audit", func(t *testing.T) {
		service, movementStore, receiptStore, auditStore := newTestSubcontractFinishedGoodsAPIService()
		orderID := "sco-smoke-260429-fg-receipt"
		createAndSubmitSubcontractOrderForTest(t, service, authConfig, orderID)
		approveAndConfirmSubcontractOrderForTest(t, service, authConfig, orderID)
		issueSubcontractMaterialsForTest(t, service, authConfig, orderID, 4)
		submitAndApproveSubcontractSampleForTest(t, service, authConfig, orderID, 5)
		startMassProductionForTest(t, service, authConfig, orderID, 7)

		receiveReq := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/"+orderID+"/receive-finished-goods", bytes.NewBufferString(`{
				"expected_version": 8,
				"receipt_id": "sfgr-smoke-260429-001",
				"receipt_no": "SFGR-SMOKE-260429-001",
				"warehouse_id": "wh-hcm-fg",
				"warehouse_code": "WH-HCM-FG",
				"location_id": "loc-hcm-fg-qc",
				"location_code": "FG-QC-01",
				"delivery_note_no": "DN-FACTORY-260429-001",
				"received_by": "warehouse-user",
				"received_at": "2026-04-29T14:00:00Z",
				"lines": [
					{
						"receive_qty": "80",
						"uom_code": "EA",
						"batch_id": "batch-fg-260429-a",
						"batch_no": "LOT-FG-260429-A",
						"lot_no": "LOT-FG-260429-A",
						"expiry_date": "2028-04-29",
						"packaging_status": "intact"
					}
				],
				"evidence": [
					{
						"evidence_type": "delivery_note",
						"file_name": "factory-delivery.pdf",
						"object_key": "subcontract/sfgr-smoke-260429-001/factory-delivery.pdf"
					}
				]
			}`)),
			authConfig,
			auth.RoleProductionOps,
		)
		receiveReq.SetPathValue("subcontract_order_id", orderID)
		receiveReq.Header.Set(response.HeaderRequestID, "req-subcontract-receive-fg")
		receiveRec := httptest.NewRecorder()

		subcontractOrderReceiveFinishedGoodsHandler(service).ServeHTTP(receiveRec, receiveReq)

		if receiveRec.Code != http.StatusOK {
			t.Fatalf("receive finished goods status = %d, want %d: %s", receiveRec.Code, http.StatusOK, receiveRec.Body.String())
		}
		payload := decodeSmokeSuccess[receiveSubcontractFinishedGoodsResponse](t, receiveRec).Data
		if payload.SubcontractOrder.Status != "finished_goods_received" ||
			payload.SubcontractOrder.ReceivedQty != "80.000000" ||
			payload.Receipt.Status != "qc_hold" ||
			payload.AuditLogID == "" {
			t.Fatalf("receive payload = %+v, want finished goods received into qc hold with audit", payload)
		}
		if len(payload.StockMovements) != 1 ||
			payload.StockMovements[0].MovementType != string(inventorydomain.MovementSubcontractReceipt) ||
			payload.StockMovements[0].StockStatus != string(inventorydomain.StockStatusQCHold) ||
			payload.StockMovements[0].SourceDocID != payload.Receipt.ID {
			t.Fatalf("stock movements = %+v, want subcontract receipt qc hold movement", payload.StockMovements)
		}
		if receiptStore.Count() != 1 {
			t.Fatalf("receipt count = %d, want 1", receiptStore.Count())
		}
		receiptMovement := smokeFindMovementByType(movementStore.Movements(), inventorydomain.MovementSubcontractReceipt)
		delta, err := receiptMovement.BalanceDelta()
		if err != nil {
			t.Fatalf("receipt movement balance delta: %v", err)
		}
		if delta.OnHand.String() != "80.000000" || !delta.Available.IsZero() {
			t.Fatalf("receipt delta = %+v, want on hand 80 and no available increase", delta)
		}
		logs, err := auditStore.List(receiveReq.Context(), audit.Query{Action: "subcontract.finished_goods_received"})
		if err != nil {
			t.Fatalf("list finished goods receipt audit logs: %v", err)
		}
		if len(logs) != 1 || logs[0].AfterData["receipt_status"] != "qc_hold" {
			t.Fatalf("finished goods audit logs = %+v, want qc hold receipt audit", logs)
		}
	})

	t.Run("denies finished goods receipt without side effects", func(t *testing.T) {
		service, movementStore, receiptStore, auditStore := newTestSubcontractFinishedGoodsAPIService()
		req := smokeSubcontractViewOnlyRequest(
			httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/sco-smoke-denied/receive-finished-goods", bytes.NewBufferString(`{
				"warehouse_id": "wh-hcm-fg",
				"location_id": "loc-hcm-fg-qc",
				"delivery_note_no": "DN-DENIED",
				"received_by": "warehouse-user",
				"lines": [
					{
						"receive_qty": "1",
						"uom_code": "EA",
						"batch_id": "batch-denied",
						"batch_no": "LOT-DENIED",
						"lot_no": "LOT-DENIED",
						"expiry_date": "2028-04-29"
					}
				]
			}`)),
			authConfig,
		)
		req.SetPathValue("subcontract_order_id", "sco-smoke-denied")
		rec := httptest.NewRecorder()

		subcontractOrderReceiveFinishedGoodsHandler(service).ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusForbidden, rec.Body.String())
		}
		if movementStore.Count() != 0 || receiptStore.Count() != 0 {
			t.Fatalf("side effects = movements %d receipts %d, want none", movementStore.Count(), receiptStore.Count())
		}
		requireNoSmokeAuditAction(t, auditStore, req.Context(), "subcontract.finished_goods_received")
	})

	t.Run("denies sample submit without side effects", func(t *testing.T) {
		service, sampleStore, auditStore := newTestSubcontractSampleAPIService()
		req := smokeSubcontractViewOnlyRequest(
			httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/sco-smoke-denied/submit-sample", bytes.NewBufferString(`{
				"sample_approval_id": "sample-denied",
				"sample_code": "SAMPLE-DENIED",
				"evidence": [
					{
						"evidence_type": "photo",
						"object_key": "subcontract/sample-denied/photo.jpg"
					}
				]
			}`)),
			authConfig,
		)
		req.SetPathValue("subcontract_order_id", "sco-smoke-denied")
		rec := httptest.NewRecorder()

		subcontractOrderSubmitSampleHandler(service).ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusForbidden, rec.Body.String())
		}
		if sampleStore.Count() != 0 {
			t.Fatalf("sample record count = %d, want no side effect", sampleStore.Count())
		}
		requireNoSmokeAuditAction(t, auditStore, req.Context(), "subcontract.sample_submitted")
	})

	t.Run("records subcontract deposit with milestone and audit", func(t *testing.T) {
		service, paymentStore, auditStore, _ := newTestSubcontractPaymentAPIService()
		orderID := "sco-smoke-260429-deposit"
		createAndSubmitSubcontractOrderForTest(t, service, authConfig, orderID)
		approveAndConfirmSubcontractOrderForTest(t, service, authConfig, orderID)

		req := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/"+orderID+"/record-deposit", bytes.NewBufferString(`{
				"expected_version": 4,
				"milestone_id": "spm-smoke-deposit-001",
				"milestone_no": "SPM-SMOKE-DEPOSIT-001",
				"amount": "250000",
				"recorded_by": "finance-user",
				"recorded_at": "2026-04-29T16:30:00Z",
				"note": "Deposit transfer confirmed"
			}`)),
			authConfig,
			auth.RoleProductionOps,
		)
		req.SetPathValue("subcontract_order_id", orderID)
		req.Header.Set(response.HeaderRequestID, "req-subcontract-record-deposit")
		rec := httptest.NewRecorder()

		subcontractOrderRecordDepositHandler(service).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("record deposit status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
		payload := decodeSmokeSuccess[subcontractPaymentMilestoneResultResponse](t, rec).Data
		if payload.CurrentStatus != "deposit_recorded" ||
			payload.SubcontractOrder.DepositAmount != "250000.00" ||
			payload.Milestone.Status != "recorded" ||
			payload.Milestone.Kind != "deposit" ||
			payload.AuditLogID == "" {
			t.Fatalf("record deposit payload = %+v, want deposit milestone and audit", payload)
		}
		if paymentStore.Count() != 1 {
			t.Fatalf("payment milestone count = %d, want 1", paymentStore.Count())
		}
		logs, err := auditStore.List(req.Context(), audit.Query{Action: "subcontract.deposit_recorded"})
		if err != nil {
			t.Fatalf("list deposit audit logs: %v", err)
		}
		if len(logs) != 1 {
			t.Fatalf("deposit audit logs = %+v, want one", logs)
		}
	})

	t.Run("marks final payment ready with milestone and audit", func(t *testing.T) {
		service, paymentStore, auditStore, orderStore := newTestSubcontractPaymentAPIService()
		order := subcontractPaymentAcceptedSmokeOrder(t)
		if err := orderStore.WithinTx(context.Background(), func(txCtx context.Context, tx productionapp.SubcontractOrderTx) error {
			return tx.Save(txCtx, order)
		}); err != nil {
			t.Fatalf("seed accepted order: %v", err)
		}

		req := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/"+order.ID+"/mark-final-payment-ready", bytes.NewBufferString(`{
				"expected_version": `+strconv.Itoa(order.Version)+`,
				"milestone_id": "spm-smoke-final-001",
				"milestone_no": "SPM-SMOKE-FINAL-001",
				"ready_by": "finance-user",
				"ready_at": "2026-04-29T18:00:00Z",
				"note": "Accepted goods cleared"
			}`)),
			authConfig,
			auth.RoleProductionOps,
		)
		req.SetPathValue("subcontract_order_id", order.ID)
		req.Header.Set(response.HeaderRequestID, "req-subcontract-final-payment-ready")
		rec := httptest.NewRecorder()

		subcontractOrderMarkFinalPaymentReadyHandler(service).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("final payment status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
		payload := decodeSmokeSuccess[subcontractPaymentMilestoneResultResponse](t, rec).Data
		if payload.CurrentStatus != "final_payment_ready" ||
			payload.Milestone.Status != "ready" ||
			payload.Milestone.Kind != "final_payment" ||
			payload.AuditLogID == "" {
			t.Fatalf("final payment payload = %+v, want final payment milestone and audit", payload)
		}
		if paymentStore.Count() != 1 {
			t.Fatalf("payment milestone count = %d, want 1", paymentStore.Count())
		}
		logs, err := auditStore.List(req.Context(), audit.Query{Action: "subcontract.final_payment_ready"})
		if err != nil {
			t.Fatalf("list final payment audit logs: %v", err)
		}
		if len(logs) != 1 {
			t.Fatalf("final payment audit logs = %+v, want one", logs)
		}
	})

	t.Run("denies payment milestones without side effects", func(t *testing.T) {
		service, paymentStore, auditStore, orderStore := newTestSubcontractPaymentAPIService()
		orderID := "sco-smoke-260429-payment-denied"
		createAndSubmitSubcontractOrderForTest(t, service, authConfig, orderID)
		approveAndConfirmSubcontractOrderForTest(t, service, authConfig, orderID)

		depositReq := smokeSubcontractViewOnlyRequest(
			httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/"+orderID+"/record-deposit", bytes.NewBufferString(`{
				"expected_version": 4,
				"milestone_id": "spm-smoke-deposit-denied",
				"amount": "250000",
				"recorded_by": "finance-user"
			}`)),
			authConfig,
		)
		depositReq.SetPathValue("subcontract_order_id", orderID)
		depositRec := httptest.NewRecorder()

		subcontractOrderRecordDepositHandler(service).ServeHTTP(depositRec, depositReq)

		if depositRec.Code != http.StatusForbidden {
			t.Fatalf("deposit status = %d, want %d: %s", depositRec.Code, http.StatusForbidden, depositRec.Body.String())
		}
		if paymentStore.Count() != 0 {
			t.Fatalf("payment milestone count = %d, want none after denied deposit", paymentStore.Count())
		}
		depositOrder, err := service.GetSubcontractOrder(depositReq.Context(), orderID)
		if err != nil {
			t.Fatalf("get denied deposit order: %v", err)
		}
		if depositOrder.Status != productiondomain.SubcontractOrderStatusFactoryConfirmed || depositOrder.Version != 4 {
			t.Fatalf("deposit order = %+v, want factory_confirmed version 4", depositOrder)
		}
		requireNoSmokeAuditAction(t, auditStore, depositReq.Context(), "subcontract.deposit_recorded")

		acceptedOrder := subcontractPaymentAcceptedSmokeOrder(t)
		if err := orderStore.WithinTx(context.Background(), func(txCtx context.Context, tx productionapp.SubcontractOrderTx) error {
			return tx.Save(txCtx, acceptedOrder)
		}); err != nil {
			t.Fatalf("seed denied final payment order: %v", err)
		}
		finalReq := smokeSubcontractViewOnlyRequest(
			httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/"+acceptedOrder.ID+"/mark-final-payment-ready", bytes.NewBufferString(`{
				"expected_version": `+strconv.Itoa(acceptedOrder.Version)+`,
				"milestone_id": "spm-smoke-final-denied",
				"ready_by": "finance-user"
			}`)),
			authConfig,
		)
		finalReq.SetPathValue("subcontract_order_id", acceptedOrder.ID)
		finalRec := httptest.NewRecorder()

		subcontractOrderMarkFinalPaymentReadyHandler(service).ServeHTTP(finalRec, finalReq)

		if finalRec.Code != http.StatusForbidden {
			t.Fatalf("final payment status = %d, want %d: %s", finalRec.Code, http.StatusForbidden, finalRec.Body.String())
		}
		if paymentStore.Count() != 0 {
			t.Fatalf("payment milestone count = %d, want none after denied final payment", paymentStore.Count())
		}
		finalOrder, err := service.GetSubcontractOrder(finalReq.Context(), acceptedOrder.ID)
		if err != nil {
			t.Fatalf("get denied final payment order: %v", err)
		}
		if finalOrder.Status != productiondomain.SubcontractOrderStatusAccepted || finalOrder.Version != acceptedOrder.Version {
			t.Fatalf("final payment order = %+v, want accepted version %d", finalOrder, acceptedOrder.Version)
		}
		requireNoSmokeAuditAction(t, auditStore, finalReq.Context(), "subcontract.final_payment_ready")
	})

	t.Run("denies sample decisions without side effects", func(t *testing.T) {
		service, sampleStore, auditStore := newTestSubcontractSampleAPIService()
		orderID := "sco-smoke-260429-sample-decision-denied"
		sampleID := orderID + "-sample"
		createAndSubmitSubcontractOrderForTest(t, service, authConfig, orderID)
		approveAndConfirmSubcontractOrderForTest(t, service, authConfig, orderID)
		issueSubcontractMaterialsForTest(t, service, authConfig, orderID, 4)
		submitSubcontractSampleForTest(t, service, authConfig, orderID, 5)

		for _, tc := range []struct {
			name        string
			path        string
			handler     http.HandlerFunc
			body        string
			auditAction string
		}{
			{
				name:        "approve",
				path:        "/api/v1/subcontract-orders/" + orderID + "/approve-sample",
				handler:     subcontractOrderApproveSampleHandler(service),
				body:        `{"expected_version":6,"sample_approval_id":"` + sampleID + `","reason":"Approved","storage_status":"retained_in_qa_cabinet"}`,
				auditAction: "subcontract.sample_approved",
			},
			{
				name:        "reject",
				path:        "/api/v1/subcontract-orders/" + orderID + "/reject-sample",
				handler:     subcontractOrderRejectSampleHandler(service),
				body:        `{"expected_version":6,"sample_approval_id":"` + sampleID + `","reason":"Rejected"}`,
				auditAction: "subcontract.sample_rejected",
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				req := smokeSubcontractViewOnlyRequest(
					httptest.NewRequest(http.MethodPost, tc.path, bytes.NewBufferString(tc.body)),
					authConfig,
				)
				req.SetPathValue("subcontract_order_id", orderID)
				rec := httptest.NewRecorder()

				tc.handler.ServeHTTP(rec, req)

				if rec.Code != http.StatusForbidden {
					t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusForbidden, rec.Body.String())
				}
				if sampleStore.Count() != 1 {
					t.Fatalf("sample record count = %d, want original submitted sample only", sampleStore.Count())
				}
				sample, err := sampleStore.Get(req.Context(), sampleID)
				if err != nil {
					t.Fatalf("get submitted sample: %v", err)
				}
				if sample.Status != productiondomain.SubcontractSampleApprovalStatusSubmitted || sample.Version != 1 {
					t.Fatalf("sample = %+v, want submitted version 1", sample)
				}
				order, err := service.GetSubcontractOrder(req.Context(), orderID)
				if err != nil {
					t.Fatalf("get sample decision denied order: %v", err)
				}
				if order.Status != productiondomain.SubcontractOrderStatusSampleSubmitted || order.Version != 6 {
					t.Fatalf("order = %+v, want sample_submitted version 6", order)
				}
				requireNoSmokeAuditAction(t, auditStore, req.Context(), tc.auditAction)
			})
		}
	})

	t.Run("denies missing record create from approval without side effects", func(t *testing.T) {
		service, auditStore := newTestSubcontractOrderAPIService()
		orderID := "sco-smoke-260429-denied"
		createAndSubmitSubcontractOrderForTest(t, service, authConfig, orderID)
		req := smokeSubcontractViewOnlyRequest(
			httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/"+orderID+"/approve", bytes.NewBufferString(`{"expected_version":2}`)),
			authConfig,
		)
		req.SetPathValue("subcontract_order_id", orderID)
		rec := httptest.NewRecorder()

		subcontractOrderApproveHandler(service).ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusForbidden, rec.Body.String())
		}
		payload := decodeSmokeError(t, rec)
		if payload.Error.Code != response.ErrorCodeForbidden {
			t.Fatalf("code = %s, want %s", payload.Error.Code, response.ErrorCodeForbidden)
		}
		order, err := service.GetSubcontractOrder(req.Context(), orderID)
		if err != nil {
			t.Fatalf("get denied approval order: %v", err)
		}
		if order.Status != productiondomain.SubcontractOrderStatusSubmitted || order.Version != 2 {
			t.Fatalf("order = %+v, want submitted version 2", order)
		}
		requireNoSmokeAuditAction(t, auditStore, req.Context(), "subcontract.order.approved")
	})
}

func newTestSubcontractOrderAPIService() (productionapp.SubcontractOrderService, audit.LogStore) {
	auditStore := audit.NewInMemoryLogStore()
	itemCatalog := masterdataapp.NewPrototypeItemCatalog(auditStore)
	partyCatalog := masterdataapp.NewPrototypePartyCatalog(auditStore)
	uomCatalog := masterdataapp.NewPrototypeUOMCatalog()
	subcontractOrderStore := productionapp.NewPrototypeSubcontractOrderStore(auditStore)

	return productionapp.NewSubcontractOrderService(
		subcontractOrderStore,
		partyCatalog,
		itemCatalog,
		subcontractOrderUOMConverterAdapter{catalog: uomCatalog},
	), auditStore
}

func newTestSubcontractMaterialIssueAPIService() (
	productionapp.SubcontractOrderService,
	*inventoryapp.InMemoryStockMovementStore,
	*productionapp.PrototypeSubcontractMaterialTransferStore,
	audit.LogStore,
) {
	auditStore := audit.NewInMemoryLogStore()
	itemCatalog := masterdataapp.NewPrototypeItemCatalog(auditStore)
	partyCatalog := masterdataapp.NewPrototypePartyCatalog(auditStore)
	uomCatalog := masterdataapp.NewPrototypeUOMCatalog()
	subcontractOrderStore := productionapp.NewPrototypeSubcontractOrderStore(auditStore)
	movementStore := inventoryapp.NewInMemoryStockMovementStore()
	transferStore := productionapp.NewPrototypeSubcontractMaterialTransferStore()
	service := productionapp.NewSubcontractOrderService(
		subcontractOrderStore,
		partyCatalog,
		itemCatalog,
		subcontractOrderUOMConverterAdapter{catalog: uomCatalog},
	).WithMaterialIssueStores(transferStore, movementStore)

	return service, movementStore, transferStore, auditStore
}

func newTestSubcontractSampleAPIService() (
	productionapp.SubcontractOrderService,
	*productionapp.PrototypeSubcontractSampleApprovalStore,
	audit.LogStore,
) {
	auditStore := audit.NewInMemoryLogStore()
	itemCatalog := masterdataapp.NewPrototypeItemCatalog(auditStore)
	partyCatalog := masterdataapp.NewPrototypePartyCatalog(auditStore)
	uomCatalog := masterdataapp.NewPrototypeUOMCatalog()
	subcontractOrderStore := productionapp.NewPrototypeSubcontractOrderStore(auditStore)
	movementStore := inventoryapp.NewInMemoryStockMovementStore()
	transferStore := productionapp.NewPrototypeSubcontractMaterialTransferStore()
	sampleStore := productionapp.NewPrototypeSubcontractSampleApprovalStore()
	service := productionapp.NewSubcontractOrderService(
		subcontractOrderStore,
		partyCatalog,
		itemCatalog,
		subcontractOrderUOMConverterAdapter{catalog: uomCatalog},
	).
		WithMaterialIssueStores(transferStore, movementStore).
		WithSampleApprovalStore(sampleStore)

	return service, sampleStore, auditStore
}

func newTestSubcontractFinishedGoodsAPIService() (
	productionapp.SubcontractOrderService,
	*inventoryapp.InMemoryStockMovementStore,
	*productionapp.PrototypeSubcontractFinishedGoodsReceiptStore,
	audit.LogStore,
) {
	auditStore := audit.NewInMemoryLogStore()
	itemCatalog := masterdataapp.NewPrototypeItemCatalog(auditStore)
	partyCatalog := masterdataapp.NewPrototypePartyCatalog(auditStore)
	uomCatalog := masterdataapp.NewPrototypeUOMCatalog()
	subcontractOrderStore := productionapp.NewPrototypeSubcontractOrderStore(auditStore)
	movementStore := inventoryapp.NewInMemoryStockMovementStore()
	transferStore := productionapp.NewPrototypeSubcontractMaterialTransferStore()
	sampleStore := productionapp.NewPrototypeSubcontractSampleApprovalStore()
	receiptStore := productionapp.NewPrototypeSubcontractFinishedGoodsReceiptStore()
	service := productionapp.NewSubcontractOrderService(
		subcontractOrderStore,
		partyCatalog,
		itemCatalog,
		subcontractOrderUOMConverterAdapter{catalog: uomCatalog},
	).
		WithMaterialIssueStores(transferStore, movementStore).
		WithSampleApprovalStore(sampleStore).
		WithFinishedGoodsReceiptStores(receiptStore, movementStore)

	return service, movementStore, receiptStore, auditStore
}

func newTestSubcontractPaymentAPIService() (
	productionapp.SubcontractOrderService,
	*productionapp.PrototypeSubcontractPaymentMilestoneStore,
	audit.LogStore,
	*productionapp.PrototypeSubcontractOrderStore,
) {
	auditStore := audit.NewInMemoryLogStore()
	itemCatalog := masterdataapp.NewPrototypeItemCatalog(auditStore)
	partyCatalog := masterdataapp.NewPrototypePartyCatalog(auditStore)
	uomCatalog := masterdataapp.NewPrototypeUOMCatalog()
	subcontractOrderStore := productionapp.NewPrototypeSubcontractOrderStore(auditStore)
	claimStore := productionapp.NewPrototypeSubcontractFactoryClaimStore()
	paymentStore := productionapp.NewPrototypeSubcontractPaymentMilestoneStore()
	service := productionapp.NewSubcontractOrderService(
		subcontractOrderStore,
		partyCatalog,
		itemCatalog,
		subcontractOrderUOMConverterAdapter{catalog: uomCatalog},
	).
		WithFactoryClaimStore(claimStore).
		WithPaymentMilestoneStore(paymentStore)

	return service, paymentStore, auditStore, subcontractOrderStore
}

func createAndSubmitSubcontractOrderForTest(
	t *testing.T,
	service productionapp.SubcontractOrderService,
	authConfig auth.MockConfig,
	id string,
) {
	t.Helper()
	createBody := bytes.NewBufferString(`{
		"id": "` + id + `",
		"order_no": "` + id + `",
		"factory_id": "sup-out-lotus",
		"finished_item_id": "item-serum-30ml",
		"planned_qty": "100",
		"uom_code": "EA",
		"currency_code": "VND",
		"sample_required": true,
		"claim_window_days": 7,
		"expected_receipt_date": "2026-05-20",
		"material_lines": [
			{
				"item_id": "item-cream-50g",
				"planned_qty": "20",
				"uom_code": "EA",
				"unit_cost": "58000"
			}
		]
	}`)
	createReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders", createBody),
		authConfig,
		auth.RoleProductionOps,
	)
	createRec := httptest.NewRecorder()

	subcontractOrdersHandler(service).ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d: %s", createRec.Code, http.StatusCreated, createRec.Body.String())
	}
	submitReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/"+id+"/submit", bytes.NewBufferString(`{"expected_version":1}`)),
		authConfig,
		auth.RoleProductionOps,
	)
	submitReq.SetPathValue("subcontract_order_id", id)
	submitRec := httptest.NewRecorder()

	subcontractOrderSubmitHandler(service).ServeHTTP(submitRec, submitReq)

	if submitRec.Code != http.StatusOK {
		t.Fatalf("submit status = %d, want %d: %s", submitRec.Code, http.StatusOK, submitRec.Body.String())
	}
}

func approveAndConfirmSubcontractOrderForTest(
	t *testing.T,
	service productionapp.SubcontractOrderService,
	authConfig auth.MockConfig,
	id string,
) {
	t.Helper()

	approveReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/"+id+"/approve", bytes.NewBufferString(`{"expected_version":2}`)),
		authConfig,
		auth.RoleProductionOps,
	)
	approveReq.SetPathValue("subcontract_order_id", id)
	approveRec := httptest.NewRecorder()

	subcontractOrderApproveHandler(service).ServeHTTP(approveRec, approveReq)

	if approveRec.Code != http.StatusOK {
		t.Fatalf("approve status = %d, want %d: %s", approveRec.Code, http.StatusOK, approveRec.Body.String())
	}

	confirmReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/"+id+"/confirm-factory", bytes.NewBufferString(`{"expected_version":3}`)),
		authConfig,
		auth.RoleProductionOps,
	)
	confirmReq.SetPathValue("subcontract_order_id", id)
	confirmRec := httptest.NewRecorder()

	subcontractOrderConfirmFactoryHandler(service).ServeHTTP(confirmRec, confirmReq)

	if confirmRec.Code != http.StatusOK {
		t.Fatalf("confirm factory status = %d, want %d: %s", confirmRec.Code, http.StatusOK, confirmRec.Body.String())
	}
}

func issueSubcontractMaterialsForTest(
	t *testing.T,
	service productionapp.SubcontractOrderService,
	authConfig auth.MockConfig,
	id string,
	expectedVersion int,
) {
	t.Helper()

	issueReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/"+id+"/issue-materials", bytes.NewBufferString(`{
			"expected_version": `+strconv.Itoa(expectedVersion)+`,
			"transfer_id": "`+id+`-transfer",
			"transfer_no": "`+id+`-TRANSFER",
			"source_warehouse_id": "wh-hcm-rm",
			"source_warehouse_code": "WH-HCM-RM",
			"handover_by": "warehouse-user",
			"received_by": "factory-receiver",
			"lines": [
				{
					"order_material_line_id": "`+id+`-material-01",
					"issue_qty": "20",
					"uom_code": "EA",
					"batch_id": "batch-cream-2603b"
				}
			],
			"evidence": [
				{
					"id": "`+id+`-handover",
					"evidence_type": "handover",
					"object_key": "subcontract/`+id+`/handover.pdf"
				}
			]
		}`)),
		authConfig,
		auth.RoleProductionOps,
	)
	issueReq.SetPathValue("subcontract_order_id", id)
	issueRec := httptest.NewRecorder()

	subcontractOrderIssueMaterialsHandler(service).ServeHTTP(issueRec, issueReq)

	if issueRec.Code != http.StatusOK {
		t.Fatalf("issue materials status = %d, want %d: %s", issueRec.Code, http.StatusOK, issueRec.Body.String())
	}
}

func submitSubcontractSampleForTest(
	t *testing.T,
	service productionapp.SubcontractOrderService,
	authConfig auth.MockConfig,
	id string,
	expectedVersion int,
) {
	t.Helper()

	submitReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/"+id+"/submit-sample", bytes.NewBufferString(`{
			"expected_version": `+strconv.Itoa(expectedVersion)+`,
			"sample_approval_id": "`+id+`-sample",
			"sample_code": "`+id+`-SAMPLE",
			"submitted_by": "factory-user",
			"evidence": [
				{
					"evidence_type": "photo",
					"object_key": "subcontract/`+id+`/sample.jpg"
				}
			]
		}`)),
		authConfig,
		auth.RoleProductionOps,
	)
	submitReq.SetPathValue("subcontract_order_id", id)
	submitRec := httptest.NewRecorder()

	subcontractOrderSubmitSampleHandler(service).ServeHTTP(submitRec, submitReq)

	if submitRec.Code != http.StatusOK {
		t.Fatalf("submit sample status = %d, want %d: %s", submitRec.Code, http.StatusOK, submitRec.Body.String())
	}
}

func submitAndApproveSubcontractSampleForTest(
	t *testing.T,
	service productionapp.SubcontractOrderService,
	authConfig auth.MockConfig,
	id string,
	expectedVersion int,
) {
	t.Helper()

	submitReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/"+id+"/submit-sample", bytes.NewBufferString(`{
			"expected_version": `+strconv.Itoa(expectedVersion)+`,
			"sample_approval_id": "`+id+`-sample",
			"sample_code": "`+id+`-SAMPLE",
			"submitted_by": "factory-user",
			"evidence": [
				{
					"evidence_type": "photo",
					"object_key": "subcontract/`+id+`/sample.jpg"
				}
			]
		}`)),
		authConfig,
		auth.RoleProductionOps,
	)
	submitReq.SetPathValue("subcontract_order_id", id)
	submitRec := httptest.NewRecorder()

	subcontractOrderSubmitSampleHandler(service).ServeHTTP(submitRec, submitReq)

	if submitRec.Code != http.StatusOK {
		t.Fatalf("submit sample status = %d, want %d: %s", submitRec.Code, http.StatusOK, submitRec.Body.String())
	}

	approveReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/"+id+"/approve-sample", bytes.NewBufferString(`{
			"expected_version": `+strconv.Itoa(expectedVersion+1)+`,
			"sample_approval_id": "`+id+`-sample",
			"reason": "Approved for mass production",
			"storage_status": "retained_in_qa_cabinet"
		}`)),
		authConfig,
		auth.RoleProductionOps,
	)
	approveReq.SetPathValue("subcontract_order_id", id)
	approveRec := httptest.NewRecorder()

	subcontractOrderApproveSampleHandler(service).ServeHTTP(approveRec, approveReq)

	if approveRec.Code != http.StatusOK {
		t.Fatalf("approve sample status = %d, want %d: %s", approveRec.Code, http.StatusOK, approveRec.Body.String())
	}
}

func startMassProductionForTest(
	t *testing.T,
	service productionapp.SubcontractOrderService,
	authConfig auth.MockConfig,
	id string,
	expectedVersion int,
) {
	t.Helper()

	req := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/"+id+"/start-mass-production", bytes.NewBufferString(`{"expected_version":`+strconv.Itoa(expectedVersion)+`}`)),
		authConfig,
		auth.RoleProductionOps,
	)
	req.SetPathValue("subcontract_order_id", id)
	rec := httptest.NewRecorder()

	subcontractOrderStartMassProductionHandler(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("start mass production status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	payload := decodeSmokeSuccess[subcontractOrderActionResultResponse](t, rec).Data
	if payload.CurrentStatus != "mass_production_started" {
		t.Fatalf("start mass production payload = %+v, want mass_production_started", payload)
	}
}

func subcontractPaymentAcceptedSmokeOrder(t *testing.T) productiondomain.SubcontractOrder {
	t.Helper()

	order, err := productiondomain.NewSubcontractOrderDocument(productiondomain.NewSubcontractOrderDocumentInput{
		ID:                  "sco-smoke-260429-final-payment",
		OrgID:               "org-my-pham",
		OrderNo:             "SCO-SMOKE-260429-FINAL",
		FactoryID:           "sup-out-lotus",
		FactoryCode:         "SUP-OUT-LOTUS",
		FactoryName:         "Lotus Filling Partner",
		FinishedItemID:      "item-serum-30ml",
		FinishedSKUCode:     "SERUM-30ML",
		FinishedItemName:    "Hydrating Serum 30ml",
		PlannedQty:          decimal.MustQuantity("100"),
		UOMCode:             "EA",
		BasePlannedQty:      decimal.MustQuantity("100"),
		BaseUOMCode:         "EA",
		ConversionFactor:    decimal.MustQuantity("1"),
		CurrencyCode:        "VND",
		SpecSummary:         "Smoke accepted final payment batch",
		SampleRequired:      false,
		ClaimWindowDays:     7,
		TargetStartDate:     "2026-05-04",
		ExpectedReceiptDate: "2026-05-20",
		CreatedAt:           timeNowForSubcontractSmoke(),
		CreatedBy:           "subcontract-user",
		MaterialLines: []productiondomain.NewSubcontractMaterialLineInput{
			{
				ID:               "sco-smoke-260429-final-material-01",
				LineNo:           1,
				ItemID:           "item-cream-50g",
				SKUCode:          "CREAM-50G",
				ItemName:         "Repair Cream 50g",
				PlannedQty:       decimal.MustQuantity("20"),
				UOMCode:          "EA",
				BasePlannedQty:   decimal.MustQuantity("20"),
				BaseUOMCode:      "EA",
				ConversionFactor: decimal.MustQuantity("1"),
				UnitCost:         decimal.MustUnitCost("58000"),
				CurrencyCode:     "VND",
				LotTraceRequired: true,
			},
		},
	})
	if err != nil {
		t.Fatalf("new subcontract order: %v", err)
	}
	for _, step := range []struct {
		name string
		fn   func(productiondomain.SubcontractOrder, string, time.Time) (productiondomain.SubcontractOrder, error)
	}{
		{"submit", productiondomain.SubcontractOrder.Submit},
		{"approve", productiondomain.SubcontractOrder.Approve},
		{"confirm factory", productiondomain.SubcontractOrder.ConfirmFactory},
		{"materials issued", productiondomain.SubcontractOrder.MarkMaterialsIssued},
		{"mass production", productiondomain.SubcontractOrder.StartMassProduction},
		{"finished goods received", productiondomain.SubcontractOrder.MarkFinishedGoodsReceived},
		{"qc started", productiondomain.SubcontractOrder.StartQC},
		{"accepted", productiondomain.SubcontractOrder.Accept},
	} {
		order, err = step.fn(order, "subcontract-user", timeNowForSubcontractSmoke())
		if err != nil {
			t.Fatalf("%s: %v", step.name, err)
		}
	}

	return order
}

func timeNowForSubcontractSmoke() time.Time {
	return time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC)
}

func smokeFindMovementByType(
	movements []inventorydomain.StockMovement,
	movementType inventorydomain.MovementType,
) inventorydomain.StockMovement {
	for _, movement := range movements {
		if movement.MovementType == movementType {
			return movement
		}
	}

	return inventorydomain.StockMovement{}
}

func smokeAuditActionsContain(logs []audit.Log, action string) bool {
	for _, log := range logs {
		if log.Action == action {
			return true
		}
	}

	return false
}

func smokeSubcontractViewOnlyRequest(req *http.Request, authConfig auth.MockConfig) *http.Request {
	principal := auth.Principal{
		UserID:      "user-subcontract-view-only",
		Email:       authConfig.Email,
		Name:        "Subcontract View Only",
		Role:        auth.RoleKey("SUBCONTRACT_VIEW_ONLY"),
		Permissions: []auth.PermissionKey{auth.PermissionSubcontractView},
	}

	return req.WithContext(auth.WithPrincipal(req.Context(), principal))
}

func requireNoSmokeAuditAction(
	t *testing.T,
	auditStore audit.LogStore,
	ctx context.Context,
	action string,
) {
	t.Helper()

	logs, err := auditStore.List(ctx, audit.Query{Action: action})
	if err != nil {
		t.Fatalf("list %s audit logs: %v", action, err)
	}
	if len(logs) != 0 {
		t.Fatalf("%s audit logs = %+v, want none", action, logs)
	}
}
