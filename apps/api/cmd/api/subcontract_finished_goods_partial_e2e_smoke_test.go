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

func TestSubcontractFinishedGoodsPartialAcceptE2ESmoke(t *testing.T) {
	authConfig := smokeAuthConfig()
	service, movementStore, receiptStore, claimStore, paymentStore, auditStore := newTestSubcontractFinishedGoodsPartialAPIService()
	orderID := "sco-e2e-260429-fg-partial"
	receiptID := "sfgr-e2e-260429-partial"
	createAndSubmitSubcontractOrderForTest(t, service, authConfig, orderID)
	approveAndConfirmSubcontractOrderForTest(t, service, authConfig, orderID)
	issueSubcontractMaterialsForTest(t, service, authConfig, orderID, 4)
	submitAndApproveSubcontractSampleForTest(t, service, authConfig, orderID, 5)
	startMassProductionForTest(t, service, authConfig, orderID, 7)

	receiveReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/"+orderID+"/receive-finished-goods", bytes.NewBufferString(`{
			"expected_version": 8,
			"receipt_id": "`+receiptID+`",
			"receipt_no": "SFGR-E2E-260429-PARTIAL",
			"warehouse_id": "wh-hcm-fg",
			"warehouse_code": "WH-HCM-FG",
			"location_id": "loc-hcm-fg-qc",
			"location_code": "FG-QC-01",
			"delivery_note_no": "DN-FACTORY-260429-PARTIAL",
			"received_by": "warehouse-user",
			"received_at": "2026-04-29T14:00:00Z",
			"lines": [
				{
					"line_no": 1,
					"item_id": "item-serum-30ml",
					"receive_qty": "100",
					"uom_code": "EA",
					"base_receive_qty": "100",
					"base_uom_code": "EA",
					"conversion_factor": "1",
					"batch_id": "batch-fg-260429-partial",
					"batch_no": "LOT-FG-260429-PARTIAL",
					"lot_no": "LOT-FG-260429-PARTIAL",
					"expiry_date": "2028-04-29",
					"packaging_status": "mixed"
				}
			],
			"evidence": [
				{
					"id": "sfgr-e2e-260429-partial-photo",
					"evidence_type": "qc_photo",
					"file_name": "qc-partial.jpg",
					"object_key": "subcontract/sfgr-e2e-260429-partial/qc-partial.jpg"
				}
			]
		}`)),
		authConfig,
		auth.RoleProductionOps,
	)
	receiveReq.SetPathValue("subcontract_order_id", orderID)
	receiveReq.Header.Set(response.HeaderRequestID, "req-e2e-subcontract-fg-partial-receive")
	receiveRec := httptest.NewRecorder()

	subcontractOrderReceiveFinishedGoodsHandler(service).ServeHTTP(receiveRec, receiveReq)

	if receiveRec.Code != http.StatusOK {
		t.Fatalf("receive finished goods status = %d, want %d: %s", receiveRec.Code, http.StatusOK, receiveRec.Body.String())
	}
	received := decodeSmokeSuccess[receiveSubcontractFinishedGoodsResponse](t, receiveRec).Data
	if received.SubcontractOrder.Status != string(productiondomain.SubcontractOrderStatusFinishedGoodsReceived) ||
		received.SubcontractOrder.ReceivedQty != "100.000000" ||
		received.Receipt.Status != "qc_hold" ||
		received.AuditLogID == "" {
		t.Fatalf("received payload = %+v, want 100 finished goods in qc hold with audit", received)
	}
	if receiptStore.Count() != 1 || movementStore.Count() != 2 {
		t.Fatalf("side effects = receipts %d movements %d, want one receipt and material issue plus receipt movements", receiptStore.Count(), movementStore.Count())
	}

	partialReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/"+orderID+"/partial-accept", bytes.NewBufferString(`{
			"expected_version": 9,
			"accepted_qty": "80",
			"uom_code": "EA",
			"base_accepted_qty": "80",
			"base_uom_code": "EA",
			"rejected_qty": "20",
			"base_rejected_qty": "20",
			"claim_id": "sfc-e2e-260429-partial",
			"claim_no": "SFC-E2E-260429-PARTIAL",
			"receipt_id": "`+receiptID+`",
			"receipt_no": "SFGR-E2E-260429-PARTIAL",
			"reason_code": "PARTIAL_PACKAGING_DAMAGE",
			"reason": "20 units held due to packaging damage after subcontract finished goods QC",
			"severity": "P2",
			"owner_id": "qa-lead",
			"accepted_by": "qc-lead",
			"accepted_at": "2026-04-29T15:00:00Z",
			"opened_by": "qa-lead",
			"opened_at": "2026-04-29T15:05:00Z",
			"note": "QC pass 80 units and hold 20 units under factory claim",
			"evidence": [
				{
					"id": "sfc-e2e-260429-partial-photo",
					"evidence_type": "photo",
					"file_name": "partial-damage.jpg",
					"object_key": "subcontract/sfc-e2e-260429-partial/partial-damage.jpg"
				}
			]
		}`)),
		authConfig,
		auth.RoleProductionOps,
	)
	partialReq.SetPathValue("subcontract_order_id", orderID)
	partialReq.Header.Set(response.HeaderRequestID, "req-e2e-subcontract-fg-partial-accept")
	partialRec := httptest.NewRecorder()

	subcontractOrderPartialAcceptFinishedGoodsHandler(service).ServeHTTP(partialRec, partialReq)

	if partialRec.Code != http.StatusOK {
		t.Fatalf("partial accept status = %d, want %d: %s", partialRec.Code, http.StatusOK, partialRec.Body.String())
	}
	partial := decodeSmokeSuccess[partialAcceptSubcontractFinishedGoodsResponse](t, partialRec).Data
	if partial.PreviousStatus != string(productiondomain.SubcontractOrderStatusFinishedGoodsReceived) ||
		partial.CurrentStatus != string(productiondomain.SubcontractOrderStatusAccepted) ||
		partial.SubcontractOrder.AcceptedQty != "80.000000" ||
		partial.SubcontractOrder.RejectedQty != "20.000000" ||
		partial.Claim.Status != "open" ||
		partial.Claim.AffectedQty != "20.000000" ||
		!partial.Claim.BlocksFinalPayment ||
		partial.AcceptAuditLogID == "" ||
		partial.ClaimAuditLogID == "" {
		t.Fatalf("partial payload = %+v, want accepted 80, blocking claim 20, and audit", partial)
	}
	if claimStore.Count() != 1 {
		t.Fatalf("factory claim count = %d, want 1", claimStore.Count())
	}
	if len(partial.StockMovements) != 1 ||
		partial.StockMovements[0].MovementType != string(inventorydomain.MovementQCRelease) ||
		partial.StockMovements[0].StockStatus != string(inventorydomain.StockStatusAvailable) ||
		partial.StockMovements[0].Quantity != "80.000000" {
		t.Fatalf("partial movements = %+v, want one qc release for accepted 80 only", partial.StockMovements)
	}
	receiptMovement := smokeFindMovementByType(movementStore.Movements(), inventorydomain.MovementSubcontractReceipt)
	receiptDelta, err := receiptMovement.BalanceDelta()
	if err != nil {
		t.Fatalf("receipt movement balance delta: %v", err)
	}
	if receiptDelta.OnHand.String() != "100.000000" || !receiptDelta.Available.IsZero() {
		t.Fatalf("receipt delta = %+v, want on-hand 100 and no available increase", receiptDelta)
	}
	releaseMovement := smokeFindMovementByType(movementStore.Movements(), inventorydomain.MovementQCRelease)
	releaseDelta, err := releaseMovement.BalanceDelta()
	if err != nil {
		t.Fatalf("release movement balance delta: %v", err)
	}
	if !releaseDelta.OnHand.IsZero() || releaseDelta.Available.String() != "80.000000" {
		t.Fatalf("release delta = %+v, want available +80 without on-hand increase", releaseDelta)
	}

	finalReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/subcontract-orders/"+orderID+"/mark-final-payment-ready", bytes.NewBufferString(`{
			"expected_version": `+strconv.Itoa(partial.SubcontractOrder.Version)+`,
			"milestone_id": "spm-e2e-260429-fg-partial-final",
			"ready_by": "finance-user"
		}`)),
		authConfig,
		auth.RoleProductionOps,
	)
	finalReq.SetPathValue("subcontract_order_id", orderID)
	finalReq.Header.Set(response.HeaderRequestID, "req-e2e-subcontract-fg-partial-final-payment")
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
		t.Fatalf("payment milestone count = %d, want none while partial claim is open", paymentStore.Count())
	}

	assertSubcontractFinishedGoodsPartialE2EAuditAction(t, auditStore, "subcontract.finished_goods_received")
	assertSubcontractFinishedGoodsPartialE2EAuditAction(t, auditStore, "subcontract.finished_goods_accepted")
	assertSubcontractFinishedGoodsPartialE2EAuditAction(t, auditStore, "subcontract.factory_claim_opened")
	requireNoSmokeAuditAction(t, auditStore, finalReq.Context(), "subcontract.final_payment_ready")
}

func newTestSubcontractFinishedGoodsPartialAPIService() (
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

func assertSubcontractFinishedGoodsPartialE2EAuditAction(t *testing.T, auditStore audit.LogStore, action string) {
	t.Helper()

	logs, err := auditStore.List(context.Background(), audit.Query{Action: action})
	if err != nil {
		t.Fatalf("list audit logs for %s: %v", action, err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit action %s count = %d, want 1", action, len(logs))
	}
}
