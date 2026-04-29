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

func TestSubcontractFinishedGoodsPassE2ESmoke(t *testing.T) {
	authConfig := smokeAuthConfig()
	service, movementStore, receiptStore, paymentStore, auditStore := newTestSubcontractFinishedGoodsPassAPIService()
	orderID := "sco-e2e-260429-fg-pass"
	createAndSubmitSubcontractOrderForTest(t, service, authConfig, orderID)
	approveAndConfirmSubcontractOrderForTest(t, service, authConfig, orderID)
	issueSubcontractMaterialsForTest(t, service, authConfig, orderID, 4)
	submitAndApproveSubcontractSampleForTest(t, service, authConfig, orderID, 5)
	startMassProductionForTest(t, service, authConfig, orderID, 7)

	receiveReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/"+orderID+"/receive-finished-goods", bytes.NewBufferString(`{
			"expected_version": 8,
			"receipt_id": "sfgr-e2e-260429-pass",
			"receipt_no": "SFGR-E2E-260429-PASS",
			"warehouse_id": "wh-hcm-fg",
			"warehouse_code": "WH-HCM-FG",
			"location_id": "loc-hcm-fg-qc",
			"location_code": "FG-QC-01",
			"delivery_note_no": "DN-FACTORY-260429-PASS",
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
					"batch_id": "batch-fg-260429-pass",
					"batch_no": "LOT-FG-260429-PASS",
					"lot_no": "LOT-FG-260429-PASS",
					"expiry_date": "2028-04-29",
					"packaging_status": "intact"
				}
			],
			"evidence": [
				{
					"id": "sfgr-e2e-260429-pass-photo",
					"evidence_type": "qc_photo",
					"file_name": "qc-pass.jpg",
					"object_key": "subcontract/sfgr-e2e-260429-pass/qc-pass.jpg"
				}
			]
		}`)),
		authConfig,
		auth.RoleProductionOps,
	)
	receiveReq.SetPathValue("subcontract_order_id", orderID)
	receiveReq.Header.Set(response.HeaderRequestID, "req-e2e-subcontract-fg-pass-receive")
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

	acceptReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/"+orderID+"/accept", bytes.NewBufferString(`{
			"expected_version": 9,
			"accepted_by": "qc-lead",
			"accepted_at": "2026-04-29T15:00:00Z",
			"note": "QC pass, release to available stock"
		}`)),
		authConfig,
		auth.RoleProductionOps,
	)
	acceptReq.SetPathValue("subcontract_order_id", orderID)
	acceptReq.Header.Set(response.HeaderRequestID, "req-e2e-subcontract-fg-pass-accept")
	acceptRec := httptest.NewRecorder()

	subcontractOrderAcceptFinishedGoodsHandler(service).ServeHTTP(acceptRec, acceptReq)

	if acceptRec.Code != http.StatusOK {
		t.Fatalf("accept finished goods status = %d, want %d: %s", acceptRec.Code, http.StatusOK, acceptRec.Body.String())
	}
	accepted := decodeSmokeSuccess[acceptSubcontractFinishedGoodsResponse](t, acceptRec).Data
	if accepted.PreviousStatus != string(productiondomain.SubcontractOrderStatusFinishedGoodsReceived) ||
		accepted.CurrentStatus != string(productiondomain.SubcontractOrderStatusAccepted) ||
		accepted.SubcontractOrder.Status != string(productiondomain.SubcontractOrderStatusAccepted) ||
		accepted.SubcontractOrder.AcceptedQty != "80.000000" ||
		accepted.AuditLogID == "" {
		t.Fatalf("accepted payload = %+v, want accepted order and audit", accepted)
	}
	if len(accepted.StockMovements) != 1 ||
		accepted.StockMovements[0].MovementType != string(inventorydomain.MovementQCRelease) ||
		accepted.StockMovements[0].StockStatus != string(inventorydomain.StockStatusAvailable) ||
		accepted.StockMovements[0].SourceDocID != received.Receipt.ID {
		t.Fatalf("accepted movements = %+v, want qc release into available linked to receipt", accepted.StockMovements)
	}
	releaseMovement := smokeFindMovementByType(movementStore.Movements(), inventorydomain.MovementQCRelease)
	delta, err := releaseMovement.BalanceDelta()
	if err != nil {
		t.Fatalf("release movement balance delta: %v", err)
	}
	if !delta.OnHand.IsZero() ||
		delta.Available.String() != "80.000000" ||
		!delta.Reserved.IsZero() {
		t.Fatalf("release delta = %+v, want available +80 without on-hand/reserved change", delta)
	}

	finalReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/"+orderID+"/mark-final-payment-ready", bytes.NewBufferString(`{
			"expected_version": `+strconv.Itoa(accepted.SubcontractOrder.Version)+`,
			"milestone_id": "spm-e2e-260429-fg-pass-final",
			"milestone_no": "SPM-E2E-260429-FG-PASS-FINAL",
			"ready_by": "finance-user",
			"ready_at": "2026-04-29T16:00:00Z",
			"note": "Finished goods accepted by QC"
		}`)),
		authConfig,
		auth.RoleProductionOps,
	)
	finalReq.SetPathValue("subcontract_order_id", orderID)
	finalReq.Header.Set(response.HeaderRequestID, "req-e2e-subcontract-fg-pass-final-payment")
	finalRec := httptest.NewRecorder()

	subcontractOrderMarkFinalPaymentReadyHandler(service).ServeHTTP(finalRec, finalReq)

	if finalRec.Code != http.StatusOK {
		t.Fatalf("final payment status = %d, want %d: %s", finalRec.Code, http.StatusOK, finalRec.Body.String())
	}
	finalPayment := decodeSmokeSuccess[subcontractPaymentMilestoneResultResponse](t, finalRec).Data
	if finalPayment.CurrentStatus != string(productiondomain.SubcontractOrderStatusFinalPaymentReady) ||
		finalPayment.Milestone.Status != "ready" ||
		finalPayment.Milestone.Kind != "final_payment" ||
		finalPayment.AuditLogID == "" {
		t.Fatalf("final payment payload = %+v, want final payment ready milestone and audit", finalPayment)
	}
	if paymentStore.Count() != 1 {
		t.Fatalf("payment milestone count = %d, want 1", paymentStore.Count())
	}

	assertSubcontractFinishedGoodsPassE2EAuditAction(t, auditStore, "subcontract.finished_goods_received")
	assertSubcontractFinishedGoodsPassE2EAuditAction(t, auditStore, "subcontract.finished_goods_accepted")
	assertSubcontractFinishedGoodsPassE2EAuditAction(t, auditStore, "subcontract.final_payment_ready")
}

func newTestSubcontractFinishedGoodsPassAPIService() (
	productionapp.SubcontractOrderService,
	*inventoryapp.InMemoryStockMovementStore,
	*productionapp.PrototypeSubcontractFinishedGoodsReceiptStore,
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

	return service, movementStore, receiptStore, paymentStore, auditStore
}

func assertSubcontractFinishedGoodsPassE2EAuditAction(t *testing.T, auditStore audit.LogStore, action string) {
	t.Helper()

	logs, err := auditStore.List(context.Background(), audit.Query{Action: action})
	if err != nil {
		t.Fatalf("list audit logs for %s: %v", action, err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit action %s count = %d, want 1", action, len(logs))
	}
}
