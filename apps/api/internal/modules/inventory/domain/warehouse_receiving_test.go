package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestNewWarehouseReceivingCapturesPurchaseLinkedInboundFields(t *testing.T) {
	receipt, err := NewWarehouseReceiving(NewWarehouseReceivingInput{
		ID:               "grn-hcm-260427-0001",
		OrgID:            "org-my-pham",
		ReceiptNo:        " grn-260427-0001 ",
		WarehouseID:      "wh-hcm-fg",
		WarehouseCode:    "wh-hcm-fg",
		LocationID:       "loc-hcm-fg-recv-01",
		LocationCode:     "fg-recv-01",
		ReferenceDocType: " purchase_order ",
		ReferenceDocID:   "PO-260427-0001",
		SupplierID:       "supplier-local",
		DeliveryNoteNo:   " dn-260427-0001 ",
		Lines: []NewWarehouseReceivingLineInput{
			{
				ID:                  "line-001",
				PurchaseOrderLineID: "po-line-001",
				ItemID:              "item-serum-30ml",
				SKU:                 " serum-30ml ",
				ItemName:            "Vitamin C Serum",
				BatchID:             "batch-serum-2604a",
				BatchNo:             " lot-2604a ",
				LotNo:               " lot-2604a ",
				ExpiryDate:          time.Date(2027, 4, 1, 14, 30, 0, 0, time.FixedZone("ICT", 7*60*60)),
				WarehouseID:         "wh-hcm-fg",
				LocationID:          "loc-hcm-fg-recv-01",
				Quantity:            decimal.MustQuantity("24"),
				UOMCode:             " ea ",
				BaseUOMCode:         " ea ",
				PackagingStatus:     " intact ",
				QCStatus:            QCStatusHold,
			},
		},
		CreatedBy: "user-warehouse-lead",
		CreatedAt: time.Date(2026, 4, 27, 9, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("new receiving: %v", err)
	}

	if receipt.ReceiptNo != "GRN-260427-0001" ||
		receipt.ReferenceDocType != ReceivingReferenceDocTypePurchaseOrder ||
		receipt.DeliveryNoteNo != "DN-260427-0001" {
		t.Fatalf("receipt header = %+v, want normalized PO-linked delivery note", receipt)
	}
	line := receipt.Lines[0]
	if line.PurchaseOrderLineID != "po-line-001" ||
		line.SKU != "SERUM-30ML" ||
		line.BatchNo != "LOT-2604A" ||
		line.LotNo != "LOT-2604A" ||
		line.ExpiryDate.Format("2006-01-02") != "2027-04-01" ||
		line.UOMCode.String() != "EA" ||
		line.BaseUOMCode.String() != "EA" ||
		line.PackagingStatus != ReceivingPackagingStatusIntact {
		t.Fatalf("line = %+v, want PO line, lot, expiry, UOM, packaging captured", line)
	}
}

func TestWarehouseReceivingRequiresPurchaseLinkAndPackaging(t *testing.T) {
	input := validWarehouseReceivingInputForTest()
	input.Lines[0].PurchaseOrderLineID = ""
	if _, err := NewWarehouseReceiving(input); !errors.Is(err, ErrReceivingRequiredField) {
		t.Fatalf("missing PO line error = %v, want required field", err)
	}

	input = validWarehouseReceivingInputForTest()
	input.Lines[0].PackagingStatus = ""
	if _, err := NewWarehouseReceiving(input); !errors.Is(err, ErrReceivingInvalidPackagingStatus) {
		t.Fatalf("missing packaging error = %v, want invalid packaging", err)
	}
}

func validWarehouseReceivingInputForTest() NewWarehouseReceivingInput {
	return NewWarehouseReceivingInput{
		ID:               "grn-hcm-260427-0001",
		OrgID:            "org-my-pham",
		ReceiptNo:        "GRN-260427-0001",
		WarehouseID:      "wh-hcm-fg",
		LocationID:       "loc-hcm-fg-recv-01",
		ReferenceDocType: ReceivingReferenceDocTypePurchaseOrder,
		ReferenceDocID:   "PO-260427-0001",
		SupplierID:       "supplier-local",
		DeliveryNoteNo:   "DN-260427-0001",
		Lines: []NewWarehouseReceivingLineInput{
			{
				ID:                  "line-001",
				PurchaseOrderLineID: "po-line-001",
				ItemID:              "item-serum-30ml",
				SKU:                 "SERUM-30ML",
				BatchID:             "batch-serum-2604a",
				BatchNo:             "LOT-2604A",
				ExpiryDate:          time.Date(2027, 4, 1, 0, 0, 0, 0, time.UTC),
				Quantity:            decimal.MustQuantity("24"),
				BaseUOMCode:         "EA",
				PackagingStatus:     ReceivingPackagingStatusIntact,
			},
		},
		CreatedBy: "user-warehouse-lead",
	}
}
