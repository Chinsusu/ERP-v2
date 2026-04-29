package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	masterdataapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/application"
	productionapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/application"
	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestSubcontractFinishedGoodsFailE2ESmoke(t *testing.T) {
	authConfig := smokeAuthConfig()
	service, movementStore, receiptStore, claimStore, paymentStore, auditStore := newTestSubcontractFinishedGoodsFailAPIService()
	orderID := "sco-e2e-260429-fg-fail"
	receiptID := "sfgr-e2e-260429-fail"
	createAndSubmitSubcontractOrderForTest(t, service, authConfig, orderID)
	approveAndConfirmSubcontractOrderForTest(t, service, authConfig, orderID)
	issueSubcontractMaterialsForTest(t, service, authConfig, orderID, 4)
	submitAndApproveSubcontractSampleForTest(t, service, authConfig, orderID, 5)
	startMassProductionForTest(t, service, authConfig, orderID, 7)

	receiveReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/"+orderID+"/receive-finished-goods", bytes.NewBufferString(`{
			"expected_version": 8,
			"receipt_id": "`+receiptID+`",
			"receipt_no": "SFGR-E2E-260429-FAIL",
			"warehouse_id": "wh-hcm-fg",
			"warehouse_code": "WH-HCM-FG",
			"location_id": "loc-hcm-fg-qc",
			"location_code": "FG-QC-01",
			"delivery_note_no": "DN-FACTORY-260429-FAIL",
			"received_by": "warehouse-user",
			"received_at": "2026-04-29T14:00:00Z",
			"lines": [
				{
					"line_no": 1,
					"item_id": "item-serum-30ml",
					"receive_qty": "80",
					"uom_code": "EA",
					"base_receive_qty": "80",
					"base_uom_code": "EA",
					"conversion_factor": "1",
					"batch_id": "batch-fg-260429-fail",
					"batch_no": "LOT-FG-260429-FAIL",
					"lot_no": "LOT-FG-260429-FAIL",
					"expiry_date": "2028-04-29",
					"packaging_status": "carton_damaged"
				}
			],
			"evidence": [
				{
					"id": "sfgr-e2e-260429-fail-photo",
					"evidence_type": "qc_photo",
					"file_name": "qc-fail.jpg",
					"object_key": "subcontract/sfgr-e2e-260429-fail/qc-fail.jpg"
				}
			]
		}`)),
		authConfig,
		auth.RoleProductionOps,
	)
	receiveReq.SetPathValue("subcontract_order_id", orderID)
	receiveReq.Header.Set(response.HeaderRequestID, "req-e2e-subcontract-fg-fail-receive")
	receiveRec := httptest.NewRecorder()

	subcontractOrderReceiveFinishedGoodsHandler(service).ServeHTTP(receiveRec, receiveReq)

	if receiveRec.Code != http.StatusOK {
		t.Fatalf("receive finished goods status = %d, want %d: %s", receiveRec.Code, http.StatusOK, receiveRec.Body.String())
	}
	received := decodeSmokeSuccess[receiveSubcontractFinishedGoodsResponse](t, receiveRec).Data
	if received.SubcontractOrder.Status != string(productiondomain.SubcontractOrderStatusFinishedGoodsReceived) ||
		received.SubcontractOrder.ReceivedQty != "80.000000" ||
		received.Receipt.Status != "qc_hold" ||
		received.AuditLogID == "" {
		t.Fatalf("received payload = %+v, want received finished goods in qc hold with audit", received)
	}
	if receiptStore.Count() != 1 || movementStore.Count() != 2 {
		t.Fatalf("side effects = receipts %d movements %d, want one receipt and material issue plus receipt movements", receiptStore.Count(), movementStore.Count())
	}

	reportReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/"+orderID+"/report-factory-defect", bytes.NewBufferString(`{
			"expected_version": 9,
			"claim_id": "sfc-e2e-260429-fail",
			"claim_no": "SFC-E2E-260429-FAIL",
			"receipt_id": "`+receiptID+`",
			"receipt_no": "SFGR-E2E-260429-FAIL",
			"reason_code": "PACKAGING_DAMAGED",
			"reason": "Outer carton crushed and seal broken on received finished goods",
			"severity": "P1",
			"affected_qty": "80",
			"uom_code": "EA",
			"base_affected_qty": "80",
			"base_uom_code": "EA",
			"owner_id": "qa-lead",
			"opened_by": "qa-lead",
			"opened_at": "2026-04-29T15:00:00Z",
			"evidence": [
				{
					"id": "sfc-e2e-260429-fail-photo",
					"evidence_type": "photo",
					"file_name": "damaged-carton.jpg",
					"object_key": "subcontract/sfc-e2e-260429-fail/damaged-carton.jpg"
				}
			]
		}`)),
		authConfig,
		auth.RoleProductionOps,
	)
	reportReq.SetPathValue("subcontract_order_id", orderID)
	reportReq.Header.Set(response.HeaderRequestID, "req-e2e-subcontract-fg-fail-claim")
	reportRec := httptest.NewRecorder()

	subcontractOrderReportFactoryDefectHandler(service).ServeHTTP(reportRec, reportReq)

	if reportRec.Code != http.StatusOK {
		t.Fatalf("report factory defect status = %d, want %d: %s", reportRec.Code, http.StatusOK, reportRec.Body.String())
	}
	reported := decodeSmokeSuccess[reportSubcontractFactoryDefectResponse](t, reportRec).Data
	if reported.PreviousStatus != string(productiondomain.SubcontractOrderStatusFinishedGoodsReceived) ||
		reported.CurrentStatus != string(productiondomain.SubcontractOrderStatusRejectedFactoryIssue) ||
		reported.SubcontractOrder.Status != string(productiondomain.SubcontractOrderStatusRejectedFactoryIssue) ||
		reported.Claim.Status != "open" ||
		!reported.Claim.BlocksFinalPayment ||
		reported.Claim.AffectedQty != "80.000000" ||
		reported.Claim.ReceiptID != receiptID ||
		reported.AuditLogID == "" {
		t.Fatalf("reported payload = %+v, want rejected order and blocking factory claim", reported)
	}
	if claimStore.Count() != 1 {
		t.Fatalf("factory claim count = %d, want 1", claimStore.Count())
	}
	receiptMovement := smokeFindMovementByType(movementStore.Movements(), inventorydomain.MovementSubcontractReceipt)
	delta, err := receiptMovement.BalanceDelta()
	if err != nil {
		t.Fatalf("receipt movement balance delta: %v", err)
	}
	if delta.OnHand.String() != "80.000000" || !delta.Available.IsZero() {
		t.Fatalf("receipt delta = %+v, want on-hand 80 and no available increase", delta)
	}
	if smokeMovementsContainType(movementStore.Movements(), inventorydomain.MovementQCRelease) {
		t.Fatalf("movements = %+v, want no qc release for failed finished goods", movementStore.Movements())
	}

	finalReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/"+orderID+"/mark-final-payment-ready", bytes.NewBufferString(`{
			"expected_version": `+strconv.Itoa(reported.SubcontractOrder.Version)+`,
			"milestone_id": "spm-e2e-260429-fg-fail-final",
			"ready_by": "finance-user"
		}`)),
		authConfig,
		auth.RoleProductionOps,
	)
	finalReq.SetPathValue("subcontract_order_id", orderID)
	finalReq.Header.Set(response.HeaderRequestID, "req-e2e-subcontract-fg-fail-final-payment")
	finalRec := httptest.NewRecorder()

	subcontractOrderMarkFinalPaymentReadyHandler(service).ServeHTTP(finalRec, finalReq)

	if finalRec.Code != http.StatusConflict {
		t.Fatalf("final payment status = %d, want %d: %s", finalRec.Code, http.StatusConflict, finalRec.Body.String())
	}
	finalError := decodeSmokeError(t, finalRec)
	if finalError.Error.Code != productionapp.ErrorCodeSubcontractOrderInvalidState {
		t.Fatalf("final payment error code = %s, want %s", finalError.Error.Code, productionapp.ErrorCodeSubcontractOrderInvalidState)
	}
	if paymentStore.Count() != 0 {
		t.Fatalf("payment milestone count = %d, want none for failed finished goods", paymentStore.Count())
	}

	assertSubcontractFinishedGoodsFailE2EAuditAction(t, auditStore, "subcontract.finished_goods_received")
	assertSubcontractFinishedGoodsFailE2EAuditAction(t, auditStore, "subcontract.factory_claim_opened")
	requireNoSmokeAuditAction(t, auditStore, finalReq.Context(), "subcontract.finished_goods_accepted")
	requireNoSmokeAuditAction(t, auditStore, finalReq.Context(), "subcontract.final_payment_ready")
}

func newTestSubcontractFinishedGoodsFailAPIService() (
	productionapp.SubcontractOrderService,
	*inventoryapp.InMemoryStockMovementStore,
	*productionapp.PrototypeSubcontractFinishedGoodsReceiptStore,
	*productionapp.PrototypeSubcontractFactoryClaimStore,
	*productionapp.PrototypeSubcontractPaymentMilestoneStore,
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
	claimStore := productionapp.NewPrototypeSubcontractFactoryClaimStore()
	paymentStore := productionapp.NewPrototypeSubcontractPaymentMilestoneStore()
	service := productionapp.NewSubcontractOrderService(
		subcontractOrderStore,
		partyCatalog,
		itemCatalog,
		subcontractOrderUOMConverterAdapter{catalog: uomCatalog},
	).
		WithMaterialIssueStores(transferStore, movementStore).
		WithSampleApprovalStore(sampleStore).
		WithFinishedGoodsReceiptStores(receiptStore, movementStore).
		WithFactoryClaimStore(claimStore).
		WithPaymentMilestoneStore(paymentStore)

	return service, movementStore, receiptStore, claimStore, paymentStore, auditStore
}

func smokeMovementsContainType(
	movements []inventorydomain.StockMovement,
	movementType inventorydomain.MovementType,
) bool {
	for _, movement := range movements {
		if movement.MovementType == movementType {
			return true
		}
	}

	return false
}

func assertSubcontractFinishedGoodsFailE2EAuditAction(t *testing.T, auditStore audit.LogStore, action string) {
	t.Helper()

	logs, err := auditStore.List(context.Background(), audit.Query{Action: action})
	if err != nil {
		t.Fatalf("list audit logs for %s: %v", action, err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit action %s count = %d, want 1", action, len(logs))
	}
}
