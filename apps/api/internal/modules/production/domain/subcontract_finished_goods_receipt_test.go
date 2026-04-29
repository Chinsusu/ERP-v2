package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestNewSubcontractFinishedGoodsReceiptCreatesQCHoldDocument(t *testing.T) {
	receipt, err := NewSubcontractFinishedGoodsReceipt(validSubcontractFinishedGoodsReceiptInput())
	if err != nil {
		t.Fatalf("new subcontract finished goods receipt: %v", err)
	}

	if receipt.Status != SubcontractFinishedGoodsReceiptStatusQCHold {
		t.Fatalf("status = %q, want qc hold", receipt.Status)
	}
	if receipt.ReceiptNo != "SFGR-20260429-001" || receipt.FactoryCode != "FAC-HCM-01" || receipt.Lines[0].SKUCode != "FG-SERUM-001" {
		t.Fatalf("normalized receipt = %+v, want uppercase receipt/factory/sku codes", receipt)
	}
	if got, want := receipt.Lines[0].BaseReceiveQty.String(), "80.000000"; got != want {
		t.Fatalf("base receive qty = %s, want %s", got, want)
	}
	if len(receipt.Evidence) != 1 || receipt.Evidence[0].EvidenceType != "delivery_note" {
		t.Fatalf("evidence = %+v, want delivery note evidence", receipt.Evidence)
	}
}

func TestSubcontractFinishedGoodsReceiptRequiresBatchTraceability(t *testing.T) {
	input := validSubcontractFinishedGoodsReceiptInput()
	input.Lines[0].BatchID = ""

	_, err := NewSubcontractFinishedGoodsReceipt(input)
	if !errors.Is(err, ErrSubcontractFinishedGoodsReceiptRequiredField) {
		t.Fatalf("error = %v, want required field", err)
	}

	input = validSubcontractFinishedGoodsReceiptInput()
	input.Lines[0].ExpiryDate = time.Time{}
	_, err = NewSubcontractFinishedGoodsReceipt(input)
	if !errors.Is(err, ErrSubcontractFinishedGoodsReceiptRequiredField) {
		t.Fatalf("error = %v, want required field", err)
	}
}

func TestSubcontractOrderReceiveFinishedGoodsUpdatesProgressAndAllowsPartialReceipt(t *testing.T) {
	receivedAt := time.Date(2026, 4, 29, 13, 0, 0, 0, time.UTC)
	order := subcontractOrderReadyForMassProduction(t)

	updated, err := order.ReceiveFinishedGoods(ReceiveSubcontractFinishedGoodsInput{
		ReceiptQty:       decimal.MustQuantity("80"),
		UOMCode:          "PCS",
		BaseReceiptQty:   decimal.MustQuantity("80"),
		BaseUOMCode:      "PCS",
		ConversionFactor: decimal.MustQuantity("1"),
		ActorID:          "warehouse-user",
		ChangedAt:        receivedAt,
	})
	if err != nil {
		t.Fatalf("receive finished goods: %v", err)
	}
	if updated.Status != SubcontractOrderStatusFinishedGoodsReceived {
		t.Fatalf("status = %q, want finished goods received", updated.Status)
	}
	if updated.FinishedGoodsReceivedBy != "warehouse-user" || !updated.FinishedGoodsReceivedAt.Equal(receivedAt) {
		t.Fatalf("finished goods metadata = %s/%s, want actor and timestamp", updated.FinishedGoodsReceivedBy, updated.FinishedGoodsReceivedAt)
	}
	if got, want := updated.ReceivedQty.String(), "80.000000"; got != want {
		t.Fatalf("received qty = %s, want %s", got, want)
	}

	updated, err = updated.ReceiveFinishedGoods(ReceiveSubcontractFinishedGoodsInput{
		ReceiptQty:       decimal.MustQuantity("20"),
		UOMCode:          "PCS",
		BaseReceiptQty:   decimal.MustQuantity("20"),
		BaseUOMCode:      "PCS",
		ConversionFactor: decimal.MustQuantity("1"),
		ActorID:          "warehouse-user",
		ChangedAt:        receivedAt.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("receive second partial finished goods: %v", err)
	}
	if got, want := updated.ReceivedQty.String(), "100.000000"; got != want {
		t.Fatalf("received qty after second receipt = %s, want %s", got, want)
	}
}

func TestSubcontractOrderReceiveFinishedGoodsRejectsInvalidStateAndOverReceipt(t *testing.T) {
	order, err := subcontractOrderReadyForFactoryConfirmation(t)
	if err != nil {
		t.Fatalf("factory confirmed order: %v", err)
	}

	_, err = order.ReceiveFinishedGoods(ReceiveSubcontractFinishedGoodsInput{
		ReceiptQty:     decimal.MustQuantity("1"),
		UOMCode:        "PCS",
		BaseReceiptQty: decimal.MustQuantity("1"),
		BaseUOMCode:    "PCS",
		ActorID:        "warehouse-user",
	})
	if !errors.Is(err, ErrSubcontractOrderInvalidTransition) {
		t.Fatalf("invalid state error = %v, want invalid transition", err)
	}

	order = subcontractOrderReadyForMassProduction(t)
	_, err = order.ReceiveFinishedGoods(ReceiveSubcontractFinishedGoodsInput{
		ReceiptQty:       decimal.MustQuantity("1001"),
		UOMCode:          "PCS",
		BaseReceiptQty:   decimal.MustQuantity("1001"),
		BaseUOMCode:      "PCS",
		ConversionFactor: decimal.MustQuantity("1"),
		ActorID:          "warehouse-user",
	})
	if !errors.Is(err, ErrSubcontractOrderInvalidQuantity) {
		t.Fatalf("over receipt error = %v, want invalid quantity", err)
	}
}

func subcontractOrderReadyForMassProduction(t *testing.T) SubcontractOrder {
	t.Helper()

	order := subcontractOrderReadyForMaterials(t, false)
	started, err := order.StartMassProduction("operations-lead", time.Now())
	if err != nil {
		t.Fatalf("start mass production: %v", err)
	}

	return started
}

func validSubcontractFinishedGoodsReceiptInput() NewSubcontractFinishedGoodsReceiptInput {
	return NewSubcontractFinishedGoodsReceiptInput{
		ID:                 "sfgr_001",
		OrgID:              "org-my-pham",
		ReceiptNo:          " sfgr-20260429-001 ",
		SubcontractOrderID: "sco_001",
		SubcontractOrderNo: "sco-20260429-001",
		FactoryID:          "fac_001",
		FactoryCode:        " fac-hcm-01 ",
		FactoryName:        "HCM Cosmetics Factory",
		WarehouseID:        "wh_hcm_fg",
		WarehouseCode:      " wh-hcm-fg ",
		LocationID:         "loc_hcm_fg_qc",
		LocationCode:       " fg-qc-01 ",
		DeliveryNoteNo:     " dn-factory-001 ",
		ReceivedBy:         "warehouse-user",
		ReceivedAt:         time.Date(2026, 4, 29, 13, 0, 0, 0, time.UTC),
		CreatedAt:          time.Date(2026, 4, 29, 12, 30, 0, 0, time.UTC),
		CreatedBy:          "warehouse-user",
		Lines: []NewSubcontractFinishedGoodsReceiptLineInput{
			{
				ID:               "sfgr_line_001",
				LineNo:           1,
				ItemID:           "item_serum",
				SKUCode:          " fg-serum-001 ",
				ItemName:         "Brightening Serum",
				BatchID:          "batch_fg_260429",
				BatchNo:          " lot-fg-260429 ",
				LotNo:            " lot-fg-260429 ",
				ExpiryDate:       time.Date(2028, 4, 29, 0, 0, 0, 0, time.UTC),
				ReceiveQty:       decimal.MustQuantity("80"),
				UOMCode:          " pcs ",
				BaseReceiveQty:   decimal.MustQuantity("80"),
				BaseUOMCode:      " pcs ",
				ConversionFactor: decimal.MustQuantity("1"),
				PackagingStatus:  "intact",
			},
		},
		Evidence: []NewSubcontractFinishedGoodsReceiptEvidenceInput{
			{
				ID:           "evidence_001",
				EvidenceType: "DELIVERY_NOTE",
				FileName:     "factory-delivery.pdf",
				ObjectKey:    "subcontract/sfgr_001/factory-delivery.pdf",
			},
		},
	}
}
