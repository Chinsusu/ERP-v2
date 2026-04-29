package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestNewPurchaseOrderDocumentCreatesDraftWithLinesAndAmounts(t *testing.T) {
	order, err := NewPurchaseOrderDocument(validPurchaseOrderInput())
	if err != nil {
		t.Fatalf("new purchase order: %v", err)
	}

	if order.Status != PurchaseOrderStatusDraft {
		t.Fatalf("status = %q, want draft", order.Status)
	}
	if order.PONo != "PO-20260429-001" || order.SupplierCode != "SUP-MYPHAM" || order.WarehouseCode != "WH-HCM" {
		t.Fatalf("normalized header = %+v, want uppercase PO/supplier/warehouse codes", order)
	}
	if order.CurrencyCode != decimal.CurrencyVND {
		t.Fatalf("currency = %q, want VND", order.CurrencyCode)
	}
	if got, want := order.TotalAmount.String(), "1250000.00"; got != want {
		t.Fatalf("total amount = %s, want %s", got, want)
	}
	if len(order.Lines) != 2 {
		t.Fatalf("line count = %d, want 2", len(order.Lines))
	}
	first := order.Lines[0]
	if first.LineNo != 1 || first.SKUCode != "RM-ROSE-001" || first.UOMCode != "KG" || first.BaseUOMCode != "G" {
		t.Fatalf("first line = %+v, want normalized line fields", first)
	}
	if first.ExpectedDate != order.ExpectedDate {
		t.Fatalf("line expected date = %q, want inherited %q", first.ExpectedDate, order.ExpectedDate)
	}
	if first.ReceivedQty.String() != "0.000000" || first.BaseReceivedQty.String() != "0.000000" {
		t.Fatalf("received qty = %s/%s, want zero defaults", first.ReceivedQty, first.BaseReceivedQty)
	}
}

func TestPurchaseOrderHappyPathTransitions(t *testing.T) {
	changedAt := time.Date(2026, 4, 29, 9, 0, 0, 0, time.UTC)
	order, err := NewPurchaseOrderDocument(validPurchaseOrderInput())
	if err != nil {
		t.Fatalf("new purchase order: %v", err)
	}

	steps := []struct {
		name string
		run  func(PurchaseOrder) (PurchaseOrder, error)
		want PurchaseOrderStatus
	}{
		{"submit", func(o PurchaseOrder) (PurchaseOrder, error) {
			return o.Submit("purchase-user", changedAt)
		}, PurchaseOrderStatusSubmitted},
		{"approve", func(o PurchaseOrder) (PurchaseOrder, error) {
			return o.Approve("finance-user", changedAt.Add(time.Minute))
		}, PurchaseOrderStatusApproved},
		{"partially received", func(o PurchaseOrder) (PurchaseOrder, error) {
			return o.MarkPartiallyReceived("warehouse-user", changedAt.Add(2*time.Minute))
		}, PurchaseOrderStatusPartiallyReceived},
		{"received", func(o PurchaseOrder) (PurchaseOrder, error) {
			return o.MarkReceived("warehouse-user", changedAt.Add(3*time.Minute))
		}, PurchaseOrderStatusReceived},
		{"close", func(o PurchaseOrder) (PurchaseOrder, error) {
			return o.Close("purchase-user", changedAt.Add(4*time.Minute))
		}, PurchaseOrderStatusClosed},
	}
	for _, step := range steps {
		order, err = step.run(order)
		if err != nil {
			t.Fatalf("%s: %v", step.name, err)
		}
		if order.Status != step.want {
			t.Fatalf("%s status = %q, want %q", step.name, order.Status, step.want)
		}
	}
	if order.Version != 6 {
		t.Fatalf("version = %d, want 6", order.Version)
	}
	if order.SubmittedBy == "" || order.ApprovedBy == "" || order.PartiallyReceivedBy == "" || order.ReceivedBy == "" || order.ClosedBy == "" {
		t.Fatalf("transition metadata = %+v, want actors", order)
	}
}

func TestPurchaseOrderRejectsInvalidTransitionsAndMissingActor(t *testing.T) {
	order, err := NewPurchaseOrder(PurchaseOrderStatusDraft)
	if err != nil {
		t.Fatalf("new purchase order: %v", err)
	}

	_, err = order.Approve("purchase-user", time.Now())
	if !errors.Is(err, ErrPurchaseOrderInvalidTransition) {
		t.Fatalf("error = %v, want invalid transition", err)
	}
	_, err = order.Submit(" ", time.Now())
	if !errors.Is(err, ErrPurchaseOrderTransitionActorRequired) {
		t.Fatalf("error = %v, want actor required", err)
	}

	cancelled, err := order.Cancel("purchase-user", time.Now())
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}
	_, err = cancelled.Submit("purchase-user", time.Now())
	if !errors.Is(err, ErrPurchaseOrderInvalidTransition) {
		t.Fatalf("error = %v, want terminal cancelled status", err)
	}
}

func TestPurchaseOrderRejectsInvalidQuantityCurrencyAndRequiredFields(t *testing.T) {
	input := validPurchaseOrderInput()
	input.CurrencyCode = "USD"
	_, err := NewPurchaseOrderDocument(input)
	if !errors.Is(err, ErrPurchaseOrderInvalidCurrency) {
		t.Fatalf("error = %v, want invalid currency", err)
	}

	input = validPurchaseOrderInput()
	input.SupplierID = " "
	_, err = NewPurchaseOrderDocument(input)
	if !errors.Is(err, ErrPurchaseOrderRequiredField) {
		t.Fatalf("error = %v, want required field", err)
	}

	input = validPurchaseOrderInput()
	input.Lines[0].OrderedQty = decimal.MustQuantity("0")
	_, err = NewPurchaseOrderDocument(input)
	if !errors.Is(err, ErrPurchaseOrderInvalidQuantity) {
		t.Fatalf("error = %v, want invalid quantity", err)
	}

	input = validPurchaseOrderInput()
	input.Lines[0].UnitPrice = decimal.Decimal("-1.0000")
	_, err = NewPurchaseOrderDocument(input)
	if !errors.Is(err, ErrPurchaseOrderInvalidAmount) {
		t.Fatalf("error = %v, want invalid amount", err)
	}
}

func TestPurchaseOrderReceivedQuantityCannotExceedOrdered(t *testing.T) {
	input := validPurchaseOrderInput()
	input.Lines[0].ReceivedQty = decimal.MustQuantity("11")
	input.Lines[0].BaseReceivedQty = decimal.MustQuantity("11000")
	_, err := NewPurchaseOrderDocument(input)
	if !errors.Is(err, ErrPurchaseOrderInvalidQuantity) {
		t.Fatalf("error = %v, want invalid quantity", err)
	}

	input = validPurchaseOrderInput()
	input.Lines[0].ReceivedQty = decimal.MustQuantity("10")
	input.Lines[0].BaseReceivedQty = decimal.MustQuantity("10001")
	_, err = NewPurchaseOrderDocument(input)
	if !errors.Is(err, ErrPurchaseOrderInvalidQuantity) {
		t.Fatalf("error = %v, want invalid base quantity", err)
	}
}

func TestPurchaseOrderLineDefaultsBaseQuantitiesFromConversionFactor(t *testing.T) {
	line, err := NewPurchaseOrderLine(NewPurchaseOrderLineInput{
		ID:               "po_line_003",
		LineNo:           1,
		ItemID:           "item_bulk",
		SKUCode:          "rm-bulk-001",
		ItemName:         "Bulk Ingredient",
		OrderedQty:       decimal.MustQuantity("2.5"),
		ReceivedQty:      decimal.MustQuantity("1"),
		UOMCode:          "KG",
		BaseUOMCode:      "G",
		ConversionFactor: decimal.MustQuantity("1000"),
		UnitPrice:        decimal.MustUnitPrice("20000"),
		CurrencyCode:     "VND",
	})
	if err != nil {
		t.Fatalf("new purchase order line: %v", err)
	}
	if got, want := line.BaseOrderedQty.String(), "2500.000000"; got != want {
		t.Fatalf("base ordered qty = %s, want %s", got, want)
	}
	if got, want := line.BaseReceivedQty.String(), "1000.000000"; got != want {
		t.Fatalf("base received qty = %s, want %s", got, want)
	}
}

func TestPurchaseOrderSubmittedCanBeRejected(t *testing.T) {
	order, err := NewPurchaseOrder(PurchaseOrderStatusDraft)
	if err != nil {
		t.Fatalf("new purchase order: %v", err)
	}
	submitted, err := order.Submit("purchase-user", time.Now())
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	rejected, err := submitted.RejectWithReason("finance-user", "wrong supplier", time.Now())
	if err != nil {
		t.Fatalf("reject: %v", err)
	}
	if rejected.Status != PurchaseOrderStatusRejected || rejected.RejectReason != "wrong supplier" || rejected.RejectedBy == "" {
		t.Fatalf("rejected order = %+v, want rejected metadata and reason", rejected)
	}
}

func TestCanTransitionPurchaseOrderStatus(t *testing.T) {
	if !CanTransitionPurchaseOrderStatus(PurchaseOrderStatusDraft, PurchaseOrderStatusSubmitted) {
		t.Fatal("draft should transition to submitted")
	}
	if CanTransitionPurchaseOrderStatus(PurchaseOrderStatusDraft, PurchaseOrderStatusApproved) {
		t.Fatal("draft should not skip approval")
	}
	if CanTransitionPurchaseOrderStatus(PurchaseOrderStatusClosed, PurchaseOrderStatusCancelled) {
		t.Fatal("closed should be terminal")
	}
	if CanTransitionPurchaseOrderStatus(PurchaseOrderStatusDraft, PurchaseOrderStatusDraft) {
		t.Fatal("same-status transition should not be accepted")
	}
}

func validPurchaseOrderInput() NewPurchaseOrderDocumentInput {
	return NewPurchaseOrderDocumentInput{
		ID:            "po_001",
		OrgID:         "org-my-pham",
		PONo:          " po-20260429-001 ",
		SupplierID:    "sup_001",
		SupplierCode:  " sup-mypham ",
		SupplierName:  "My Pham Supplier",
		WarehouseID:   "wh_hcm",
		WarehouseCode: " wh-hcm ",
		ExpectedDate:  "2026-05-02",
		CurrencyCode:  "VND",
		CreatedAt:     time.Date(2026, 4, 29, 8, 0, 0, 0, time.UTC),
		CreatedBy:     "purchase-user",
		Lines: []NewPurchaseOrderLineInput{
			{
				ID:               "po_line_001",
				LineNo:           1,
				ItemID:           "item_rose",
				SKUCode:          " rm-rose-001 ",
				ItemName:         "Rose Extract",
				OrderedQty:       decimal.MustQuantity("10"),
				UOMCode:          " kg ",
				BaseOrderedQty:   decimal.MustQuantity("10000"),
				BaseUOMCode:      " g ",
				ConversionFactor: decimal.MustQuantity("1000"),
				UnitPrice:        decimal.MustUnitPrice("75000"),
			},
			{
				ID:               "po_line_002",
				ItemID:           "item_bottle",
				SKUCode:          " pk-bottle-100 ",
				ItemName:         "Bottle 100ml",
				OrderedQty:       decimal.MustQuantity("1000"),
				UOMCode:          " pcs ",
				BaseOrderedQty:   decimal.MustQuantity("1000"),
				BaseUOMCode:      " pcs ",
				ConversionFactor: decimal.MustQuantity("1"),
				UnitPrice:        decimal.MustUnitPrice("500"),
			},
		},
	}
}
