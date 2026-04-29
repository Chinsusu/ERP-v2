package application

import (
	"context"
	"errors"
	"testing"
	"time"

	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
)

func TestSubcontractFinishedGoodsReceiptServiceBuildReceiptCreatesQCHoldDocumentAndMovements(t *testing.T) {
	fixedNow := time.Date(2026, 4, 29, 13, 0, 0, 0, time.UTC)
	service := NewSubcontractFinishedGoodsReceiptService()
	service.clock = func() time.Time { return fixedNow }
	order := subcontractFinishedGoodsReceiptMassProductionOrder(t)

	result, err := service.BuildReceipt(context.Background(), BuildSubcontractFinishedGoodsReceiptInput{
		ID:             "sfgr_001",
		ReceiptNo:      "SFGR-20260429-001",
		Order:          order,
		WarehouseID:    "wh_hcm_fg",
		WarehouseCode:  "WH-HCM-FG",
		LocationID:     "loc_hcm_fg_qc",
		LocationCode:   "FG-QC-01",
		DeliveryNoteNo: "DN-FACTORY-001",
		ReceivedBy:     "warehouse-user",
		ReceivedAt:     fixedNow.Add(30 * time.Minute),
		ActorID:        "warehouse-user",
		Lines: []BuildSubcontractFinishedGoodsReceiptLineInput{
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
		Evidence: []BuildSubcontractFinishedGoodsReceiptEvidenceInput{
			{
				EvidenceType: "delivery_note",
				FileName:     "factory-delivery.pdf",
				ObjectKey:    "subcontract/sfgr_001/factory-delivery.pdf",
			},
		},
	})
	if err != nil {
		t.Fatalf("build finished goods receipt: %v", err)
	}

	if result.Receipt.Status != productiondomain.SubcontractFinishedGoodsReceiptStatusQCHold {
		t.Fatalf("receipt status = %q, want qc hold", result.Receipt.Status)
	}
	if result.UpdatedOrder.Status != productiondomain.SubcontractOrderStatusFinishedGoodsReceived {
		t.Fatalf("order status = %q, want finished goods received", result.UpdatedOrder.Status)
	}
	if got, want := result.UpdatedOrder.ReceivedQty.String(), "80.000000"; got != want {
		t.Fatalf("updated order received qty = %s, want %s", got, want)
	}
	if len(result.StockMovements) != 1 {
		t.Fatalf("stock movement count = %d, want 1", len(result.StockMovements))
	}
	movement := result.StockMovements[0]
	if movement.MovementType != inventorydomain.MovementSubcontractReceipt ||
		movement.StockStatus != inventorydomain.StockStatusQCHold ||
		movement.SourceDocType != subcontractFinishedGoodsReceiptSourceDoc ||
		movement.SourceDocID != result.Receipt.ID ||
		movement.SourceDocLineID != result.Receipt.Lines[0].ID {
		t.Fatalf("movement = %+v, want subcontract receipt into qc hold", movement)
	}
	delta, err := movement.BalanceDelta()
	if err != nil {
		t.Fatalf("movement balance delta: %v", err)
	}
	if delta.OnHand.String() != "80.000000" || !delta.Available.IsZero() {
		t.Fatalf("movement delta = %+v, want on hand 80 and available 0", delta)
	}
}

func TestSubcontractFinishedGoodsReceiptServiceRejectsReceiptBeforeMassProduction(t *testing.T) {
	service := NewSubcontractFinishedGoodsReceiptService()
	order := subcontractMaterialTransferTestOrder(t)

	_, err := service.BuildReceipt(context.Background(), BuildSubcontractFinishedGoodsReceiptInput{
		Order:          order,
		WarehouseID:    "wh_hcm_fg",
		LocationID:     "loc_hcm_fg_qc",
		DeliveryNoteNo: "DN-FACTORY-001",
		ReceivedBy:     "warehouse-user",
		ActorID:        "warehouse-user",
		Lines: []BuildSubcontractFinishedGoodsReceiptLineInput{
			{
				ReceiveQty: "80",
				UOMCode:    "PCS",
				BatchID:    "batch_fg_260429",
				BatchNo:    "LOT-FG-260429",
				LotNo:      "LOT-FG-260429",
				ExpiryDate: "2028-04-29",
			},
		},
	})
	if !errors.Is(err, productiondomain.ErrSubcontractOrderInvalidTransition) {
		t.Fatalf("error = %v, want invalid transition", err)
	}
}

func TestSubcontractFinishedGoodsReceiptServiceRejectsWrongFinishedItem(t *testing.T) {
	service := NewSubcontractFinishedGoodsReceiptService()
	order := subcontractFinishedGoodsReceiptMassProductionOrder(t)

	_, err := service.BuildReceipt(context.Background(), BuildSubcontractFinishedGoodsReceiptInput{
		Order:          order,
		WarehouseID:    "wh_hcm_fg",
		LocationID:     "loc_hcm_fg_qc",
		DeliveryNoteNo: "DN-FACTORY-001",
		ReceivedBy:     "warehouse-user",
		ActorID:        "warehouse-user",
		Lines: []BuildSubcontractFinishedGoodsReceiptLineInput{
			{
				ItemID:     "item_other",
				SKUCode:    "FG-OTHER-001",
				ItemName:   "Other Finished Goods",
				ReceiveQty: "80",
				UOMCode:    "PCS",
				BatchID:    "batch_fg_260429",
				BatchNo:    "LOT-FG-260429",
				LotNo:      "LOT-FG-260429",
				ExpiryDate: "2028-04-29",
			},
		},
	})
	if !errors.Is(err, productiondomain.ErrSubcontractFinishedGoodsReceiptInvalidQuantity) {
		t.Fatalf("error = %v, want invalid quantity for finished item mismatch", err)
	}
}

func TestPrototypeSubcontractFinishedGoodsReceiptStoreSavesAndListsByOrder(t *testing.T) {
	store := NewPrototypeSubcontractFinishedGoodsReceiptStore()
	service := NewSubcontractFinishedGoodsReceiptService()
	result, err := service.BuildReceipt(context.Background(), BuildSubcontractFinishedGoodsReceiptInput{
		ID:             "sfgr_store_001",
		Order:          subcontractFinishedGoodsReceiptMassProductionOrder(t),
		WarehouseID:    "wh_hcm_fg",
		LocationID:     "loc_hcm_fg_qc",
		DeliveryNoteNo: "DN-FACTORY-001",
		ReceivedBy:     "warehouse-user",
		ActorID:        "warehouse-user",
		Lines: []BuildSubcontractFinishedGoodsReceiptLineInput{
			{
				ReceiveQty: "80",
				UOMCode:    "PCS",
				BatchID:    "batch_fg_260429",
				BatchNo:    "LOT-FG-260429",
				LotNo:      "LOT-FG-260429",
				ExpiryDate: "2028-04-29",
			},
		},
	})
	if err != nil {
		t.Fatalf("build receipt: %v", err)
	}
	if err := store.Save(context.Background(), result.Receipt); err != nil {
		t.Fatalf("save receipt: %v", err)
	}

	rows, err := store.ListBySubcontractOrder(context.Background(), result.Receipt.SubcontractOrderID)
	if err != nil {
		t.Fatalf("list receipts: %v", err)
	}
	if len(rows) != 1 || rows[0].ID != result.Receipt.ID {
		t.Fatalf("rows = %+v, want saved receipt for subcontract order", rows)
	}
}

func subcontractFinishedGoodsReceiptMassProductionOrder(t *testing.T) productiondomain.SubcontractOrder {
	t.Helper()

	order := subcontractMaterialTransferTestOrder(t)
	order.SampleRequired = false
	issued, err := order.MarkMaterialsIssued("warehouse-user", time.Now())
	if err != nil {
		t.Fatalf("mark materials issued: %v", err)
	}
	started, err := issued.StartMassProduction("operations-lead", time.Now())
	if err != nil {
		t.Fatalf("start mass production: %v", err)
	}

	return started
}
