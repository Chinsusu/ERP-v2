package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestSupplierRejectionModelRequiresLinkedRejectedGoods(t *testing.T) {
	rejection, err := NewSupplierRejection(validSupplierRejectionInput())
	if err != nil {
		t.Fatalf("new supplier rejection: %v", err)
	}

	if rejection.Status != SupplierRejectionStatusDraft ||
		rejection.SupplierID != "supplier-local" ||
		rejection.GoodsReceiptID != "grn-hcm-260427-inspect" ||
		rejection.InboundQCInspectionID != "iqc-fail-001" ||
		len(rejection.Lines) != 1 ||
		len(rejection.Attachments) != 1 {
		t.Fatalf("rejection = %+v, want linked draft rejection with evidence", rejection)
	}
	if rejection.Lines[0].RejectedQuantity.String() != "6.000000" ||
		rejection.Lines[0].BatchNo != "LOT-2604A" ||
		rejection.Lines[0].UOMCode.String() != "EA" {
		t.Fatalf("line = %+v, want normalized rejected goods line", rejection.Lines[0])
	}
}

func TestSupplierRejectionRejectsMissingTraceability(t *testing.T) {
	input := validSupplierRejectionInput()
	input.InboundQCInspectionID = ""

	_, err := NewSupplierRejection(input)
	if !errors.Is(err, ErrSupplierRejectionRequiredField) {
		t.Fatalf("err = %v, want required traceability field", err)
	}
}

func TestSupplierRejectionRejectsInvalidQuantity(t *testing.T) {
	input := validSupplierRejectionInput()
	input.Lines[0].RejectedQuantity = decimal.MustQuantity("0")

	_, err := NewSupplierRejection(input)
	if !errors.Is(err, ErrSupplierRejectionInvalidQuantity) {
		t.Fatalf("err = %v, want invalid rejected quantity", err)
	}
}

func TestSupplierRejectionTransitions(t *testing.T) {
	rejection, err := NewSupplierRejection(validSupplierRejectionInput())
	if err != nil {
		t.Fatalf("new supplier rejection: %v", err)
	}

	submitted, err := rejection.Submit("user-warehouse-lead", time.Date(2026, 4, 29, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("submit rejection: %v", err)
	}
	confirmed, err := submitted.Confirm("user-warehouse-lead", time.Date(2026, 4, 29, 11, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("confirm rejection: %v", err)
	}
	if confirmed.Status != SupplierRejectionStatusConfirmed ||
		confirmed.SubmittedBy != "user-warehouse-lead" ||
		confirmed.ConfirmedBy != "user-warehouse-lead" {
		t.Fatalf("confirmed rejection = %+v, want submitted then confirmed", confirmed)
	}

	_, err = confirmed.Cancel("user-warehouse-lead", "wrong supplier", time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC))
	if !errors.Is(err, ErrSupplierRejectionInvalidTransition) {
		t.Fatalf("err = %v, want invalid cancel after confirmed", err)
	}
}

func validSupplierRejectionInput() NewSupplierRejectionInput {
	return NewSupplierRejectionInput{
		ID:                    "srj-260429-0001",
		OrgID:                 "org-my-pham",
		RejectionNo:           "SRJ-260429-0001",
		SupplierID:            "supplier-local",
		SupplierCode:          "SUP-LOCAL",
		SupplierName:          "Local Supplier",
		PurchaseOrderID:       "po-260427-0003",
		PurchaseOrderNo:       "PO-260427-0003",
		GoodsReceiptID:        "grn-hcm-260427-inspect",
		GoodsReceiptNo:        "GRN-260427-0003",
		InboundQCInspectionID: "iqc-fail-001",
		WarehouseID:           "wh-hcm-fg",
		WarehouseCode:         "WH-HCM-FG",
		Reason:                "damaged packaging",
		CreatedAt:             time.Date(2026, 4, 29, 9, 0, 0, 0, time.UTC),
		CreatedBy:             "user-qa",
		Lines: []NewSupplierRejectionLineInput{
			{
				ID:                    "srj-line-001",
				PurchaseOrderLineID:   "po-line-260427-0003-001",
				GoodsReceiptLineID:    "grn-line-draft-001",
				InboundQCInspectionID: "iqc-fail-001",
				ItemID:                "item-serum-30ml",
				SKU:                   "serum-30ml",
				ItemName:              "Vitamin C Serum",
				BatchID:               "batch-serum-2604a",
				BatchNo:               "lot-2604a",
				LotNo:                 "lot-2604a",
				ExpiryDate:            time.Date(2027, 4, 1, 0, 0, 0, 0, time.UTC),
				RejectedQuantity:      decimal.MustQuantity("6"),
				UOMCode:               "ea",
				BaseUOMCode:           "EA",
				Reason:                "damaged packaging",
			},
		},
		Attachments: []NewSupplierRejectionAttachmentInput{
			{
				ID:          "srj-att-001",
				LineID:      "srj-line-001",
				FileName:    "damage-photo.jpg",
				ObjectKey:   "supplier-rejections/srj-260429-0001/damage-photo.jpg",
				ContentType: "image/jpeg",
				Source:      "inbound_qc",
			},
		},
	}
}
