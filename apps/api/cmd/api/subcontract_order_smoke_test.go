package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	masterdataapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/application"
	productionapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
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
		req := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/sco-smoke-denied/issue-materials", bytes.NewBufferString(`{
				"source_warehouse_id": "wh-hcm-rm",
				"received_by": "factory-receiver",
				"lines": []
			}`)),
			authConfig,
			auth.RoleFinanceOps,
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
		logs, err := auditStore.List(req.Context(), audit.Query{Action: "subcontract.materials_issued"})
		if err != nil {
			t.Fatalf("list material issue audit logs: %v", err)
		}
		if len(logs) != 0 {
			t.Fatalf("material issue audit logs = %+v, want none", logs)
		}
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

	t.Run("denies sample submit without side effects", func(t *testing.T) {
		service, sampleStore, auditStore := newTestSubcontractSampleAPIService()
		req := smokeRequestAsRole(
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
			auth.RoleFinanceOps,
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
		logs, err := auditStore.List(req.Context(), audit.Query{Action: "subcontract.sample_submitted"})
		if err != nil {
			t.Fatalf("list sample audit logs: %v", err)
		}
		if len(logs) != 0 {
			t.Fatalf("sample audit logs = %+v, want none", logs)
		}
	})

	t.Run("denies finance role from approval action without audit", func(t *testing.T) {
		service, auditStore := newTestSubcontractOrderAPIService()
		createAndSubmitSubcontractOrderForTest(t, service, authConfig, "sco-smoke-260429-denied")
		req := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/sco-smoke-260429-denied/approve", bytes.NewBufferString(`{"expected_version":2}`)),
			authConfig,
			auth.RoleFinanceOps,
		)
		req.SetPathValue("subcontract_order_id", "sco-smoke-260429-denied")
		rec := httptest.NewRecorder()

		subcontractOrderApproveHandler(service).ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusForbidden, rec.Body.String())
		}
		payload := decodeSmokeError(t, rec)
		if payload.Error.Code != response.ErrorCodeForbidden {
			t.Fatalf("code = %s, want %s", payload.Error.Code, response.ErrorCodeForbidden)
		}
		logs, err := auditStore.List(req.Context(), audit.Query{Action: "subcontract.order.approved"})
		if err != nil {
			t.Fatalf("list audit logs: %v", err)
		}
		if len(logs) != 0 {
			t.Fatalf("approval audit log count = %d, want 0 for denied action", len(logs))
		}
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

func smokeAuditActionsContain(logs []audit.Log, action string) bool {
	for _, log := range logs {
		if log.Action == action {
			return true
		}
	}

	return false
}
