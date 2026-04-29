package application

import (
	"context"
	"testing"
	"time"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

func TestSubcontractOrderServiceReceiveFinishedGoodsPersistsReceiptMovementAndAudit(t *testing.T) {
	ctx := context.Background()
	auditStore := audit.NewInMemoryLogStore()
	orderStore := NewPrototypeSubcontractOrderStore(auditStore)
	receiptStore := NewPrototypeSubcontractFinishedGoodsReceiptStore()
	movementStore := inventoryapp.NewInMemoryStockMovementStore()
	service := SubcontractOrderService{
		store:                 orderStore,
		finishedGoodsStore:    receiptStore,
		finishedGoodsRecorder: movementStore,
		finishedGoodsBuild:    NewSubcontractFinishedGoodsReceiptService(),
	}
	order := subcontractFinishedGoodsReceiptMassProductionOrder(t)
	if err := orderStore.WithinTx(ctx, func(txCtx context.Context, tx SubcontractOrderTx) error {
		return tx.Save(txCtx, order)
	}); err != nil {
		t.Fatalf("seed order: %v", err)
	}

	result, err := service.ReceiveSubcontractFinishedGoods(ctx, ReceiveSubcontractFinishedGoodsInput{
		ID:              order.ID,
		ExpectedVersion: order.Version,
		ReceiptID:       "sfgr-service-001",
		ReceiptNo:       "SFGR-SERVICE-001",
		WarehouseID:     "wh_hcm_fg",
		WarehouseCode:   "WH-HCM-FG",
		LocationID:      "loc_hcm_fg_qc",
		LocationCode:    "FG-QC-01",
		DeliveryNoteNo:  "DN-FACTORY-001",
		ReceivedBy:      "warehouse-user",
		ReceivedAt:      time.Date(2026, 4, 29, 14, 0, 0, 0, time.UTC),
		ActorID:         "warehouse-user",
		RequestID:       "req-fg-receive",
		Lines: []ReceiveSubcontractFinishedGoodsLineInput{
			{
				ReceiveQty:      "80",
				UOMCode:         "PCS",
				BatchID:         "batch_fg_260429",
				BatchNo:         "LOT-FG-260429",
				LotNo:           "LOT-FG-260429",
				ExpiryDate:      "2028-04-29",
				PackagingStatus: "intact",
			},
		},
		Evidence: []ReceiveSubcontractFinishedGoodsEvidenceInput{
			{
				EvidenceType: "delivery_note",
				ObjectKey:    "subcontract/sfgr-service-001/factory-delivery.pdf",
			},
		},
	})
	if err != nil {
		t.Fatalf("receive finished goods: %v", err)
	}

	if result.SubcontractOrder.Status != productiondomain.SubcontractOrderStatusFinishedGoodsReceived ||
		result.Receipt.Status != productiondomain.SubcontractFinishedGoodsReceiptStatusQCHold ||
		result.AuditLogID == "" {
		t.Fatalf("result = %+v, want finished goods received order, qc hold receipt, audit", result)
	}
	if receiptStore.Count() != 1 {
		t.Fatalf("receipt count = %d, want 1", receiptStore.Count())
	}
	if movementStore.Count() != 1 {
		t.Fatalf("movement count = %d, want 1", movementStore.Count())
	}
	movement := movementStore.Movements()[0]
	if movement.MovementType != inventorydomain.MovementSubcontractReceipt ||
		movement.StockStatus != inventorydomain.StockStatusQCHold ||
		movement.SourceDocID != result.Receipt.ID {
		t.Fatalf("movement = %+v, want subcontract receipt into qc hold", movement)
	}
	delta, err := movement.BalanceDelta()
	if err != nil {
		t.Fatalf("movement balance delta: %v", err)
	}
	if delta.OnHand.String() != "80.000000" || !delta.Available.IsZero() {
		t.Fatalf("movement delta = %+v, want on hand 80 and no available change", delta)
	}
	logs, err := auditStore.List(ctx, audit.Query{Action: subcontractFinishedGoodsAction})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 || logs[0].AfterData["receipt_status"] != "qc_hold" {
		t.Fatalf("audit logs = %+v, want finished goods receipt audit", logs)
	}
}
